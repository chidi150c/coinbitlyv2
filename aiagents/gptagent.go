package aiagents

import (
	"log"
	"coinbitly.com/model"
)

type AgentWorker struct {
	GenContentChan chan string
	UserInputChan chan string
	Agent model.AIAgent
}

func NewAgentWorker(ag model.AIAgent) *AgentWorker {
	return &AgentWorker{
		GenContentChan : make(chan string),
		UserInputChan : make(chan string),
		Agent: ag,
	}
}

func (g *AgentWorker) LiveChat(input string){ 
	content, err := g.Agent.AIModelAPI(input)
	if err != nil{
		log.Fatalln(err)
	}
	// log.Println(content)
	g.GenContentChan <-content
}