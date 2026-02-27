package cmd

type Course struct {
	ID          string   `json:"course_id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Language    string   `json:"language"`
	Lessons     []Lesson `json:"lessons"`
}

type Lesson struct {
	ID              string   `json:"lesson_id"`
	Title           string   `json:"title"`
	Slides          []string `json:"slides"`
	TaskDescription string   `json:"task_description"`
	InitialCode     string   `json:"initial_code"`
	FileName        string   `json:"file_name"`
}
