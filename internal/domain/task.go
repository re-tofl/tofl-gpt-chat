package domain

type UnionProblemResponse struct {
	Success FormalResponse
	Error   ParserErrorResponse
}
