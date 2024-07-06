package models

type PageData struct {
	Jobs  []Job `json:"items"`
	Pages int   `json:"pages"`
}

type Job struct {
	Id     string `json:"id"`
	Name   string `json:"name"`
	URL    string `json:"alternate_url"`
	Salary struct {
		SalaryFrom int    `json:"from"`
		SalaryTo   int    `json:"to"`
		Currency   string `json:"currency"`
	}
	Area struct {
		Name string `json:"name"`
	} `json:"area"`
	Employer struct {
		Name string `json:"name"`
	} `json:"employer"`
	Experience struct {
		Id string `json:"id"`
	} `json:"experience"`
	ExperienceInt int `json:"experience_int"`
}
