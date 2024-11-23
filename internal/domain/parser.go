package domain

type ParserRequest struct {
	TRS            string `json:"TRS"`
	Interpretation string `json:"Interpretation"`
}

type ParserResponse struct {
	ErrorTrs            []string `json:"error_trs"`
	ErrorInterpretation []string `json:"error_interpretation"`
}
