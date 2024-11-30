package domain

type UserAchievement struct {
	ChatId          int64         `json:"chat_id"`
	CountOfRequests int           `json:"count_of_req"`
	Achievements    []Achievement `json:"achievements"`
}

type Achievement struct {
	Title string `json:"ach_title"`
	Desc  string `json:"ach_desc"`
	Grade int    `json:"ach_grade"`
}
