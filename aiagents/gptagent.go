package aiagents

import (
	"fmt"
	"log"

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
	// content := input
	content, err := g.Agent.AIModelAPI(input)
	if err != nil{
		log.Fatalln(err)
	}
	log.Println(content)
	course := g.Store.CreateDB()
	course, err = processContent(content, course)
	if err != nil{
		return nil, err
	}
	g.Store.UpdateDB(course)
	return course, nil
}


// processContent takes the raw string from OpenAI and structures it into the desired format.
func processContent(generated string, course *model.Course) (*model.Course, error) {
	// firstSplit := strings.Split(generated, "\n\n")
	// parts := make([][]string, len(firstSplit))
	// for k, v := range firstSplit{
	// 	parts[k] = strings.Split(v, ":")
	// }

    // Splitting the main body into subsections using ". " as a separator
    // mainBodySubs := strings.Split(parts[2][1], ". ")

	course.Title = "Introduction to Flutter "+fmt.Sprintf("%d",course.Id)
	course.Intro = "Learn the fundamentals of artificial intelligence (AI), including machine learning, deep learning, and neural networks."
	course.ImageUrl = "path/to/ai-course-image.jpg"
	  
	return course, nil
}