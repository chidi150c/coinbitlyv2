package aiagents

import (
	"testing"

	"coinbitly.com/config"
	// "coinbitly.com/model"
	"coinbitly.com/openaiapi"
)

func TestLiveChat(t *testing.T){
	c := config.NewModelConfigs()["gpt3"]
	a := openaiapi.NewOpenAI(c.ApiKey, c.Url, c.Model, []openaiapi.Message{})
	g := &AgentWorker{
		GenContentChan: make(chan string),
		Agent  :      a,
	}
	go g.LiveChat("Flutter")
	t.Fatalf("answer = %s", <-g.GenContentChan)
}