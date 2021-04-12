package types

type Repo struct {
	Id          uint   `json:"id"`
	Name        string `json:"full_name"`
	Description string `json:"description"`
	Url         string `json:"html_url"`
	Language    string `json:"language"`
}
