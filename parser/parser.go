package parser

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

type (
	Test struct {
		Name    string
		Method  string
		Path    string
		Headers map[string]string
		Body    string
	}
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
func ParseSource(bs []byte) ([]*Test, error) {
	rows := toRows(bs)

	position := name
	tests := make([]*Test, 0)
	test := &Test{
		Headers: make(map[string]string),
	}

	for _, row := range rows {
		if strings.HasPrefix(row, "#") {
			if position == body {
				tests = append(tests, test)
				test = &Test{
					Headers: make(map[string]string),
				}
				position = name
			}

			if position == name {
				test.Name = strings.SplitN(row, " ", 2)[1]
				position++
				continue
			}
		}

		if !strings.HasPrefix(row, "#") {
			if position == method {
				methodPath := strings.SplitN(row, " ", 2)
				test.Method = methodPath[0]
				position++
				test.Path = methodPath[1]
				position++
				continue
			}

			if position == headers {
				if row != "" {
					headerSplit := strings.SplitN(row, ":", 2)
					test.Headers[headerSplit[0]] = strings.TrimSpace(headerSplit[1])
				} else {
					position = body
					continue
				}
			}

			if position == body {
				if row != "" {
					if test.Body == "" {
						test.Body = row
					} else {
						test.Body = fmt.Sprintf("%s\n%s", test.Body, row)
					}
				} else {
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
		tests = append(tests, test)
	}

	return tests, nil
}

// ParseFacit parses a facit file (.result) returning any responses found or an error
func ParseFacit(bs []byte) ([]*Response, error) {
	rows := toRows(bs)

	position := resName
	responses := make([]*Response, 0)
	response := &Response{
		Headers: make(map[string]string),
	}

	for _, row := range rows {
		if strings.HasPrefix(row, "#") {
			if position == resBody {
				responses = append(responses, response)
				response = &Response{
					Headers: make(map[string]string),
				}
				position = resName
			}

			if position == resName {
				response.Name = strings.SplitN(row, " ", 2)[1]
				position++
				continue
			}
		}

		if !strings.HasPrefix(row, "#") {
			if position == status {
				s, err := strconv.Atoi(row)

				if err != nil {
					return nil, err
				}

				response.Status = s
				position++
				continue
			}

			if position == resHeaders {
				if row != "" {
					headerSplit := strings.SplitN(row, ":", 2)
					response.Headers[headerSplit[0]] = strings.TrimSpace(headerSplit[1])
				} else {
					position = resBody
					continue
				}
			}

			if position == resBody {
				if row != "" {
					if response.Body == "" {
						response.Body = row
					} else {
						response.Body = fmt.Sprintf("%s\n%s", response.Body, row)
					}
				} else {
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
		responses = append(responses, response)
	}

	return responses, nil
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
