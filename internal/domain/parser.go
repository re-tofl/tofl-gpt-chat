package domain

type ParserErrorResponse struct {
	ErrorTrs            []string `json:"error_trs"`
	ErrorInterpretation []string `json:"error_interpretation"`
}

type Child struct {
	Value  string  `json:"value"`
	Childs []Child `json:"childs"`
}

type ParserResponse struct {
	JsonTRS []struct {
		Left struct {
			Value  string  `json:"value"`
			Childs []Child `json:"childs"`
		} `json:"left"`
		Right struct {
			Value  string  `json:"value"`
			Childs []Child `json:"childs"`
		} `json:"right"`
	} `json:"json_TRS"`
	JsonInterpret struct {
		Functions []struct {
			Name       string   `json:"name"`
			Variables  []string `json:"variables"`
			Expression string   `json:"expression"`
		} `json:"functions"`
	} `json:"json_interpret"`
}

type UnionParserResponse struct {
	Success ParserResponse
	Error   ParserErrorResponse
}
