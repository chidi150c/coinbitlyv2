package openaiapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"coinbitly.com/model"
)

// Define structs to match the response structure
type OpenAIResponse struct {
	ID       string  `json:"id"`
	Object   string  `json:"object"`
	Created  int64   `json:"created"`
	Model    string  `json:"model"`
	Choices  []Choice `json:"choices"`
	Usage    Usage   `json:"usage"`
}

type Choice struct {
	Index         int    `json:"index"`
	Message       Message `json:"message"`
	Logprobs      interface{} `json:"logprobs"`
	FinishReason  string  `json:"finish_reason"`
}

type Usage struct {
	PromptTokens    int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens     int `json:"total_tokens"`
}

// Define a struct to hold the request payload
type CompletionRequest struct {
	Model    string `json:"model"`
	Messages []Message `json:"messages"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAI struct{
	ApiKey string
	Url string
	Model string 
	Messages []Message 
}

func NewOpenAI(apiKey, url, model string, messages []Message)*OpenAI{
	return &OpenAI{
		ApiKey: apiKey,
		Url: url, 
		Model: model,
		Messages: messages, 
	}
}
var _ model.AIAgent = &OpenAI{}

func (a *OpenAI)CourseIntroAgent(userInput string)(string, error){
	messages := []Message{
		{"system", "You are an informative and concise assistant tasked with creating educational content. Generate a comprehensive tutorial on " + userInput + "."},
		{"user", "Please start with the Title."},
		{"system", "The title should be engaging and descriptive of the tutorial's content."},
		{"user", "Next, provide the Introduction."},
		{"system", "The introduction should set the context for " + userInput + " and outline the objectives of the tutorial."},
		{"user", "Proceed with the Main Body."},
		{"system", "In the main body, elaborate on the following sections: Overview of " + userInput + ", Practical Applications, Step-by-Step Instructions, Best Practices, Common Pitfalls, and Troubleshooting Tips. Each section should be comprehensive yet concise."},
		{"user", "Finally, wrap up with the Conclusion."},
		{"system", "The conclusion should recap the key takeaways of the tutorial and suggest next steps for the reader to continue learning about " + userInput + "."},
	}
	return apiFetch(messages, a.Model, a.Url, a.ApiKey)
}

func apiFetch(messages []Message, Model, Url, ApiKey string)(string, error){
	requestBody := CompletionRequest{
		Model:    Model,
		Messages: messages,
	}

	// Marshal the request body to JSON
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("error marshalling the request body: %v", err)
	}

	// Create a new HTTP request
	req, err := http.NewRequest("POST", Url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("error creating the request: %v", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+ApiKey)
	req.Header.Set("Content-Type", "application/json")

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error making the request: %v", err)
	}
	defer resp.Body.Close()

	// Read the response
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading the response body: %v", err)
	}
	// fmt.Println("yes=", string(responseBody))
	// Parse the JSON response
	var response OpenAIResponse
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return "", fmt.Errorf("error parsing the JSON response: %v", err)
	}

	// Assuming you're interested in the content of the first choice
	if len(response.Choices) > 0 {
		generatedContent := response.Choices[0].Message.Content
		return generatedContent, nil
	} else {
		return "", fmt.Errorf("no content generated")
	}
}
func (a *OpenAI)CourseDetailsAgent(course *model.Course) (*model.Carousal, error){
	var carousal = &model.Carousal{
		MainBody:   make([]model.CourseBody, 6),
	}
	var messages []Message
	carousal.Title = course.Title
	for k,v := range course.MainBody{
		messages = []Message{
			{"system", "You are an informative and concise assistant tasked with creating educational content. Generate a comprehensive tutorial on " + course.Title + "."},
			{"user", "Please elaborate on "+ v +" in one slide"},
		}
		content, err := apiFetch(messages, a.Model,a.Url, a.ApiKey)
		if err != nil {
			log.Fatalf("error fetching API content: %v", err)
		}		
		carousal.MainBody[k] = model.CourseBody{
			SubTitle: strings.Split(v, ":")[0],
			Body:     content,
		}                  
	}
	carousal.Conclusion = course.Conclusion	
	return carousal, nil
}
