package parser

type (
	ParseResult struct {
		Variables map[string]interface{}
		Tests     []*Test
		Facit     []*Response
	}

	Test struct {
		Name    string
		Method  string
		Path    string
		Headers map[string]string
		Body    string
	}

	Response struct {
		Name    string
		Status  int
		Headers map[string]string
		Body    string
	}
)
