package openaiapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

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

func (a *OpenAI)AIModelAPI(userInput string)(string, error){

	// Construct the request payload
	messages := []Message{
		{"system", "You are an informative and concise assistant tasked with creating educational content. Generate a tutorial on " + userInput + " with the following structure: Title, Introduction, Main Body, and Conclusion."},
		{"user", "Please start with the Title."},
		{"system", "The title should be catchy and reflect the content of the tutorial."},
		{"user", "Next, provide the Introduction."},
		{"system", "The introduction should briefly introduce " + userInput + " and mention what the tutorial will cover."},
		{"user", "Now, let's move on to the Main Body."},
		{"system", "The main body should contain detailed content on " + userInput + " setup, key concepts, and fundamental widgets each not more than 500 characters. Include examples where possible."},
		{"user", "Finally, conclude the tutorial."},
		{"system", "The conclusion, in less than 500 characters, should summarize what was learned and encourage further exploration of " + userInput + "."},
	}
	requestBody := CompletionRequest{
		Model:    a.Model,
		Messages: messages,
	}

	// Marshal the request body to JSON
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("Error marshalling the request body: %v", err)
	}

	// Create a new HTTP request
	req, err := http.NewRequest("POST", a.Url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("Error creating the request: %v", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+a.ApiKey)
	req.Header.Set("Content-Type", "application/json")

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Error making the request: %v", err)
	}
	defer resp.Body.Close()

	// Read the response
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("Error reading the response body: %v", err)
	}
	// fmt.Println("yes=", string(responseBody))
	// Parse the JSON response
	var response OpenAIResponse
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return "", fmt.Errorf("Error parsing the JSON response: %v", err)
	}

	// Assuming you're interested in the content of the first choice
	if len(response.Choices) > 0 {
		generatedContent := response.Choices[0].Message.Content
		return generatedContent, nil
	} else {
		return "", fmt.Errorf("No content generated.")
	}
}
