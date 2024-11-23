package domain

const (
	StartState = iota
	ProblemInputState
	ProblemLLMState
	ProblemParserState
	ProblemParserApprovalState
	ProblemFormalState
	ProblemRateState
)
