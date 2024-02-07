package openaiapi

import (
	// "fmt"
	"testing"

	"coinbitly.com/config"
)

func TestAIModelAPI(t *testing.T) {
	userInput := "Flutter"
	c := config.NewModelConfigs()["gpt3"]
	o := NewOpenAI(c.ApiKey, c.Url, c.Model, []Message{})
	resp, err := o.AIModelAPI(userInput)
	// fmt.Println("Response =", resp)
	if err != nil {
		t.Fatal(err)
	}else{
		t.Fatal("Response", resp)
	}
}
