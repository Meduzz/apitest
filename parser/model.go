package parser

type (
	Response struct {
		Name    string
		Status  int
		Headers map[string]string
		Body    string
	}
)
