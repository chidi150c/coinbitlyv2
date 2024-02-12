package aiagents

import (
	"fmt"
	"coinbitly.com/model"
)

type CourseDBService struct{
	DB []*model.Course
}

func NewCourseDBService ()*CourseDBService{
	return &CourseDBService{
		DB: make([]*model.Course,0),
	}
}

var _ model.CourseDBServicer = &CourseDBService{}

func (c *CourseDBService)CreateDB()*model.Course{
	course := &model.Course{
		Id:       0,
		Title:    "",
		Intro:    "",
		ImageUrl: "",
	}
	c.DB = append(c.DB, course)
	course.Id = len(c.DB)-1
	return course
}

func (c *CourseDBService)ReadDB(id int)(*model.Course, error){
	var (
		course *model.Course
		err error
	)
	course = c.DB[id]
	return course, err
}

func (c *CourseDBService)UpdateDB(course *model.Course)error{
	var (
		err error
	)
	if ok := c.DB[course.Id] ; ok != nil{
		c.DB[course.Id] = course
	}else{
		return fmt.Errorf("course with ID: %d does not exist", course.Id)
	}
	return err
}

func (c *CourseDBService)DeleteDB(id int)(error){
	var (
		err error
	)
	return err
}
