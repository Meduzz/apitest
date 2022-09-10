package parser

type (
	ParseResult struct {
		Variables map[string]interface{}
		Tests     []*Test
		Facit     []*Response
	}

	Test struct {
		Name    string            `json:"-"`
		Method  string            `json:"method"`
		Path    string            `json:"path"`
		Headers map[string]string `json:"headers"`
		Body    string            `json:"body"`
	}

	Response struct {
		Name    string            `json:"-"`
		Status  int               `json:"-"`
		Headers map[string]string `json:"headers"`
		Body    string            `json:"body"`
	}
)
