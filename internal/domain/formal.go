package domain

type FormalResult struct {
	Format string `json:"format"`
	Data   string `json:"data"`
}

type FormalResponse struct {
	Result []FormalResult `json:"result"`
}
