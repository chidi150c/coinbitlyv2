package model

type Course struct {
	Id         int    `json:"id"`
	Title      string `json:"title"`
	Intro      string `json:"intro"`
	MainBody  []string `json:"mainbody"`
	Conclusion string `json:"conclusion"`
	ImageUrl   string `json:"imageUrl"`
}

type CourseDBServicer interface {
	CreateDB() *Course
	ReadDB(id int) (*Course, error)
	UpdateDB(*Course) error
	DeleteDB(id int) error
}

// CourseRequest is the structure that represents the incoming POST request
type CourseRequest struct {
	Subject string `json:"subject"`
}

type CourseServicer interface {
	GenerateCourseContent(subject string) (*Course, error)
}

type Slide struct {
	Title      string `json:"title"`
	SubTitle1  string `json:"subtitle1"`
	Body1      string `json:"body1"`
	SubTitle2  string `json:"subtitle2"`
	Body2      string `json:"body2"`
	SubTitle3  string `json:"subtitle3"`
	Body3      string `json:"body3"`
	SubTitle4  string `json:"subtitle4"`
	Body4      string `json:"body4"`
	SubTitle5  string `json:"subtitle5"`
	Body5      string `json:"body5"`
	SubTitle6  string `json:"subtitle6"`
	Body6      string `json:"body6"`
	Conclusion string `json:"conclusion"`
}
type CourseBody struct{
	SubTitle  string `json:"subtitle"`
	Body      string `json:"body"`
}
type Carousal struct {
	Title      string `json:"title"`
	MainBody  []CourseBody `json:"mainbody"`
	Conclusion string `json:"conclusion"`
}
type AIAgent interface {
	CourseIntroAgent(string) (string, error)
	CourseDetailsAgent(*Course) (*Carousal, error)
}
