package aiagents

import (
	"fmt"
	"log"
	"strings"

	"coinbitly.com/model"
)

type AgentWorker struct {
	Agent model.AIAgent
	Store *CourseDBService
}

func NewAgentWorker(ag model.AIAgent, cdb *CourseDBService) *AgentWorker {
	return &AgentWorker{
		Agent: ag,
		Store: cdb,
	}
}

var _ model.CourseServicer = &AgentWorker{}

func (g *AgentWorker) GenerateCourseContent(input string)(*model.Course, error){ 
	content, err := g.Agent.CourseIntroAgent(input)
	if err != nil{
		log.Fatalln(err)
	}
	course := g.Store.CreateDB()

	fmt.Printf("\n %s", content)

	course, err = processContent(content, course)
	if err != nil{
		return nil, err
	}
	g.Store.UpdateDB(course)
	return course, nil
}


// processContent takes the raw string from OpenAI and structures it into the desired format.
func processContent(generated string, course *model.Course) (*model.Course, error) {
	firstSplit := strings.Split(generated, "\n\n")
	course.Title = firstSplit[0]
	course.Intro = firstSplit[1]
	for i := 3; i <= 8; i++{
		course.MainBody = append(course.MainBody, firstSplit[i])
	}
	course.Conclusion = firstSplit[9]
	  
	return course, nil
}