package aiagents

import (
	"fmt"
	"log"

	"coinbitly.com/model"
)

type CourseDBService struct{
	nextId int
	DB map[int]*model.Course
}

func NewCourseDBService ()*CourseDBService{
	return &CourseDBService{
		DB: make(map[int]*model.Course,0),
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
	c.DB[c.nextId] = course
	course.Id = c.nextId
	c.nextId++
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
	cse, ok := c.DB[course.Id] 
	if (!ok){
		return fmt.Errorf("course with ID: %d does not exist", course.Id)
	}else if cse.Id != course.Id{
		log.Fatalf("course with ID: %d does not exist", course.Id)
	}else{		
		c.DB[course.Id] = course
	}
	return err
}

func (c *CourseDBService)DeleteDB(id int)(error){
	var (
		err error
	)
	return err
}
