package commands

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/Meduzz/helper/http/client"
	"meduzz.github.com/apitest/parser"

	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{}
	cmd.Use = "test"
	cmd.Short = "infile"
	cmd.Long = "test <infile>"
	cmd.Args = cobra.ExactArgs(1)
	cmd.RunE = handleTest

	Root.AddCommand(cmd)
}

const (
	post   = "POST"
	put    = "PUT"
	get    = "GET"
	delete = "DELETE"
)

type (
	result struct {
		Field    string
		Expected interface{}
		Actual   interface{}
	}
)

func handleTest(cmd *cobra.Command, args []string) error {
	infile := args[0]

	result, err := readAndParseInFile(infile)

	if err != nil {
		return err
	}

	outfile := fmt.Sprintf("%s.%s", infile, "facit")

	actual := make([]*parser.Response, 0)

	for _, test := range result.Tests {
		switch test.Method {
		case post:
			res, err := doPost(test)

			if err != nil {
				return err
			}

			actual = append(actual, res)
		}
	}

	file, err := os.OpenFile(outfile, os.O_RDWR|os.O_APPEND, 0644)

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			buf := bytes.NewBufferString("")

			for _, test := range actual {
				printResponse(buf, test)
			}

			err = os.WriteFile(outfile, buf.Bytes(), 0644)

			if err != nil {
				return err
			}

			return nil
		} else {
			return err
		}
	}

	defer file.Close()
	bs, err := os.ReadFile(outfile)

	if err != nil {
		return err
	}

	result, err = parser.ParseFacit(bs, result)

	if err != nil {
		return err
	}

	for _, res := range actual {
		facit := findResponseByName(res.Name, result.Facit)

		if facit != nil {
			diff := compareResponses(facit, res)

			fmt.Printf("### %s\n", facit.Name)
			if len(diff) > 0 {
				printDiffs(diff)
			} else {
				fmt.Println("Success!")
			}
			fmt.Println()
		} else {
			buf := bytes.NewBufferString("\n")
			printResponse(buf, res)

			_, err = file.Write(buf.Bytes())

			if err != nil {
				return err
			}

			fmt.Printf("### %s\n", res.Name)
			fmt.Println("New test found. Result was added to facit.")
		}
	}

	return nil
}

func readAndParseInFile(name string) (*parser.ParseResult, error) {
	bs, err := os.ReadFile(name)

	if err != nil {
		return nil, err
	}

	return parser.ParseSource(bs)
}

func setHeaders(req *client.HttpRequest, headers map[string]string) *client.HttpRequest {
	for k, v := range headers {
		req.Request().Header.Add(k, v)
	}

	return req
}

func flatternHeader(header []string) string {
	if len(header) == 1 {
		return header[0]
	}

	return strings.Join(header, ";")
}

func address(test *parser.Test) error {
	if !strings.HasPrefix(test.Path, "http") {
		host, ok := test.Headers["Host"]

		if !ok {
			return fmt.Errorf("no address found")
		}

		if !strings.HasPrefix(host, "http") {
			test.Path = fmt.Sprintf("http://%s%s", host, test.Path)
		} else {
			test.Path = fmt.Sprintf("%s%s", host, test.Path)
		}
	}

	return nil
}

func doPost(test *parser.Test) (*parser.Response, error) {
	err := address(test)

	if err != nil {
		return nil, err
	}

	req, err := client.POSTBytes(test.Path, []byte(test.Body), "application/json")

	if err != nil {
		return nil, err
	}

	req = setHeaders(req, test.Headers)

	res, err := req.DoDefault()

	if err != nil {
		return nil, err
	}

	body, err := res.AsText()

	if err != nil {
		return nil, err
	}

	response := &parser.Response{}
	response.Name = test.Name
	response.Status = res.Code()
	response.Headers = make(map[string]string)
	response.Body = body

	for k, v := range res.Response().Header {
		response.Headers[k] = flatternHeader(v)
	}

	return response, nil
}

func printResponse(buf *bytes.Buffer, res *parser.Response) {
	fmt.Fprintf(buf, "### %s\n", res.Name)
	fmt.Fprintf(buf, "%d\n", res.Status)

	for k, v := range res.Headers {
		fmt.Fprintf(buf, "%s: %s\n", k, v)
	}

	fmt.Fprintln(buf)

	if res.Body != "" {
		fmt.Fprintf(buf, "%s\n", res.Body)
		fmt.Fprintln(buf)
	}
}

func findResponseByName(name string, responses []*parser.Response) *parser.Response {
	for _, response := range responses {
		if response.Name == name {
			return response
		}
	}

	return nil
}

func compareResponses(facit, real *parser.Response) []*result {
	errs := make([]*result, 0)

	if facit.Status != real.Status {
		r := &result{
			Field:    "status",
			Expected: facit.Status,
			Actual:   real.Status,
		}
		errs = append(errs, r)
	}

	for k, v := range facit.Headers {
		if real.Headers[k] != v {
			r := &result{
				Field:    fmt.Sprintf("headers.%s", k),
				Expected: v,
				Actual:   real.Headers[k],
			}
			errs = append(errs, r)
		}
	}

	if facit.Body != real.Body {
		r := &result{
			Field:    "body",
			Expected: facit.Body,
			Actual:   real.Body,
		}
		errs = append(errs, r)
	}

	return errs
}

func printDiffs(diff []*result) {
	for _, d := range diff {
		fmt.Printf("%s expected %v but was %v\n", d.Field, d.Expected, d.Actual)
	}
}
