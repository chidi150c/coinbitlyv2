package model


type Course struct {
	Id       int    `json:"id"`
	Title    string `json:"title"`
	Intro    string `json:"intro"`
	ImageUrl string `json:"imageUrl"`
}

type CourseDBServicer interface{
	CreateDB()*Course
	ReadDB (id int)(*Course, error)
	UpdateDB(*Course)error
	DeleteDB(id int)error
} 
// CourseRequest is the structure that represents the incoming POST request
type CourseRequest struct {
	Subject string `json:"subject"`
}

type CourseServicer interface{
	GenerateCourseContent(subject string)(*Course, error)
}

type AIAgent interface{
	AIModelAPI(string)(string, error)
}

