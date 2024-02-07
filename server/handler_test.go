package server_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"coinbitly.com/aiagents"
	"coinbitly.com/config"
	"coinbitly.com/openaiapi"
	"coinbitly.com/server"
	"coinbitly.com/strategies"
)
type AIService interface{
	LiveChat(input string)
}

func TestHandler(t *testing.T) {
	ts := &strategies.TradingSystem{}
	hn := os.Getenv("HOSTSITE")
	c := config.NewModelConfigs()["gpt3"]
	a := openaiapi.NewOpenAI(c.ApiKey, c.Url, c.Model, []openaiapi.Message{})
	g := &aiagents.AgentWorker{
		GenContentChan: make(chan string),
		Agent:          a,
	}
	hr := server.NewTradeHandler(ts, hn, g)
	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll pass 'nil' as the third parameter.
    body := strings.NewReader(`{"user_input": "Flutter"}`)
    req, err := http.NewRequest("POST", "/generate_content", body)
    if err != nil {
        t.Fatal(err)
    }
    req.Header.Set("Content-Type", "application/json")

    // We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
    rr := httptest.NewRecorder()

    // Our handlers satisfy http.Handler, so we can call their ServeHTTP method directly and pass in our Request and ResponseRecorder.
    hr.GenerateContent(rr, req)

    // Check the status code is what we expect.
    if status := rr.Code; status != http.StatusOK {
        t.Errorf("hr returned wrong status code: got %v want %v", status, http.StatusOK)
    }

    // Check the response body is what we expect.
    expected := `Generated content for Flutter`
    if rr.Body.String() != expected {
        t.Errorf("hr returned unexpected body: got %v want %v", rr.Body.String(), expected)
    }
}
