package parser

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/aymerick/raymond"
	"github.com/oliveagle/jsonpath"
)

const (
	name = iota
	method
	path
	headers
	body
)

const (
	resName = iota
	status
	resHeaders
	resBody
)

// ParseSource parse a source file (.http|.rest) returning any requests found or an error
func ParseSource(bs []byte) (*ParseResult, error) {
	result := &ParseResult{}
	variables := make(map[string]interface{})
	result.Variables = variables

	rows := toRows(bs)

	position := name
	tests := make([]*Test, 0)
	test := &Test{
		Headers: make(map[string]string),
	}

	for _, row := range rows {
		// divider and request naming
		if strings.HasPrefix(row, "#") {
			if position == body {
				if test.Body != "" {
					// do body templating
					body, err := raymond.Render(test.Body, variables)

					if err != nil {
						return nil, err
					}

					test.Body = body
				}

				tests = append(tests, test)
				test = &Test{
					Headers: make(map[string]string),
				}
				position = name
			}

			if strings.TrimSpace(row) == "###" {
				// dividier
				continue
			}

			if position == name {
				if idx := strings.Index(row, "@name"); idx > 0 {
					// named request
					idx += len("@name")
					test.Name = strings.TrimSpace(row[idx:])
				} else {
					// normal request
					test.Name = strings.SplitN(row, " ", 2)[1]
				}
				position++
				continue
			}
		}

		// request stuff
		if !strings.HasPrefix(row, "#") {
			// file variable
			if strings.HasPrefix(row, "@") {
				v := strings.SplitN(row, "=", 2)
				key := strings.TrimSpace(v[0][1:])
				value := strings.TrimSpace(v[1])

				variables[key] = value
				continue
			}

			// request line
			if position == method {
				methodPath := strings.SplitN(row, " ", 2)
				test.Method = methodPath[0]
				position++

				// do path templating
				path, err := raymond.Render(methodPath[1], variables)

				if err != nil {
					return nil, err
				}

				test.Path = path
				position++
				continue
			}

			// headers
			if position == headers {
				if row != "" {
					headerSplit := strings.SplitN(row, ":", 2)

					// do header templating
					value, err := raymond.Render(strings.TrimSpace(headerSplit[1]), variables)

					if err != nil {
						return nil, err
					}

					test.Headers[headerSplit[0]] = value
				} else {
					position = body
					continue
				}
			}

			// request body
			if position == body {
				if row != "" {
					if test.Body == "" {
						test.Body = row
					} else {
						test.Body = fmt.Sprintf("%s\n%s", test.Body, row)
					}
				} else {
					if test.Body != "" {
						// do body templating
						body, err := raymond.Render(test.Body, variables)

						if err != nil {
							return nil, err
						}

						test.Body = body
					}

					tests = append(tests, test)
					test = &Test{
						Headers: make(map[string]string),
					}
					position = name
				}
			}
		}
	}

	if test.Name != "" {
		if test.Body != "" {
			// do body templating
			body, err := raymond.Render(test.Body, variables)

			if err != nil {
				return nil, err
			}

			test.Body = body
		}

		tests = append(tests, test)
	}

	result.Tests = tests

	return result, nil
}

// ParseFacit parses a facit file (.result) returning any responses found or an error
func ParseFacit(bs []byte, result *ParseResult) (*ParseResult, error) {
	rows := toRows(bs)

	position := resName
	responses := make([]*Response, 0)
	response := &Response{
		Headers: make(map[string]string),
	}

	for _, row := range rows {
		// divider and naming
		if strings.HasPrefix(row, "#") {
			if position == resBody {
				if response.Body != "" {
					// do body templating
					body, err := raymond.Render(response.Body, result.Variables)

					if err != nil {
						return nil, err
					}

					response.Body = body
				}

				responses = append(responses, response)
				response = &Response{
					Headers: make(map[string]string),
				}
				position = resName
			}

			if strings.TrimSpace(row) == "###" {
				// dividier
				continue
			}

			if position == resName {
				if idx := strings.Index(row, "@name"); idx > 0 {
					// named request
					response.Name = strings.TrimSpace(row[idx:])
				} else {
					// normal request
					response.Name = strings.SplitN(row, " ", 2)[1]
				}
				position++
				continue
			}
		}

		// request body
		if !strings.HasPrefix(row, "#") {
			// variables
			if strings.HasPrefix(row, "@") {
				v := strings.SplitN(row, "=", 2)
				key := strings.TrimSpace(v[0][1:])
				value := strings.TrimSpace(v[1])

				result.Variables[key] = value
				continue
			}

			// status line
			if position == status {
				s, err := strconv.Atoi(row)

				if err != nil {
					return nil, err
				}

				response.Status = s
				position++
				continue
			}

			// response headers
			if position == resHeaders {
				if row != "" {
					headerSplit := strings.SplitN(row, ":", 2)

					// do header templating
					value := runtimeResponseTemplating(strings.TrimSpace(headerSplit[1]), result.Variables)
					value, err := raymond.Render(value, result.Variables)

					if err != nil {
						return nil, err
					}

					response.Headers[headerSplit[0]] = value
				} else {
					position = resBody
					continue
				}
			}

			// response body
			if position == resBody {
				if row != "" {
					if response.Body == "" {
						response.Body = row
					} else {
						response.Body = fmt.Sprintf("%s\n%s", response.Body, row)
					}
				} else {
					if response.Body != "" {
						// do body templating
						body := runtimeResponseTemplating(response.Body, result.Variables)
						body, err := raymond.Render(body, result.Variables)

						if err != nil {
							return nil, err
						}

						response.Body = body
					}

					responses = append(responses, response)
					response = &Response{
						Headers: make(map[string]string),
					}
					position = resName
				}
			}
		}
	}

	if response.Name != "" {
		if response.Body != "" {
			// do body templating
			body, err := raymond.Render(response.Body, result.Variables)

			if err != nil {
				return nil, err
			}

			response.Body = body
		}

		responses = append(responses, response)
	}

	result.Facit = responses

	return result, nil
}

// splits bs into rows and returns them as strings
func toRows(bs []byte) []string {
	rows := bytes.Split(bs, []byte("\n"))

	ret := make([]string, 0)

	for _, row := range rows {
		ret = append(ret, string(row))
	}

	return ret
}

func dig(data map[string]interface{}, keys []string) interface{} {
	value, ok := data[keys[0]]

	if !ok {
		return nil
	}

	if len(keys) == 1 {
		return value
	}

	workish, ok := value.(map[string]interface{})

	if !ok {
		return nil
	}

	return dig(workish, keys[1:])
}

func executeJsonpath(match string, variables map[string]interface{}) {
	idx := strings.Index(match, "$")
	key := match[2 : idx-1]
	path := match[idx : len(match)-2]

	keys := strings.Split(key, ".")
	_, ok := variables[keys[0]]

	if !ok {
		// the test has not been ran yet
		return
	}

	value := dig(variables, keys)

	if value != nil {
		var obj interface{}
		json.Unmarshal([]byte(value.(string)), &obj)
		data, err := jsonpath.JsonPathLookup(obj, path)

		if err != nil {
			fmt.Printf("jsonpath threw error: %v\n", err)
			return
		}

		// TODO better clean up the data
		variables[match[2:len(match)-2]] = data.(string)
	}
}

func runtimeResponseTemplating(data string, variables map[string]interface{}) string {
	rx := regexp.MustCompile("{{[a-z.@?$]+}}")

	if strings.Contains(data, "{{") {
		matches := rx.FindAllString(data, -1)

		for _, i := range matches {
			if strings.Contains(i, "$") {
				it, exists := variables[i[2:len(i)-2]]

				if exists {
					// the variable is already set and will be templated
					data = strings.ReplaceAll(data, i, it.(string))
					continue
				}

				executeJsonpath(i, variables)
				it, exists = variables[i[2:len(i)-2]]

				if exists {
					data = strings.ReplaceAll(data, i, it.(string))
				}
			}
		}
	}

	return data
}
