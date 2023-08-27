package server

import (
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"coinbitly.com/model"
	"coinbitly.com/webclient"
	"github.com/go-chi/chi"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

// WebService is a user login-aware wrapper for a html/template.
type SocketService struct {	
	Upgrader websocket.Upgrader
}

// parseTemplate applies a given file to the body of the base template.
func NewSocketService(HostSite string) *SocketService {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {			
	// 		if r.Header.Get("Origin") != h.HostSite {
	// 			return false
	// 		}
			return true 
		},
	}
	return &SocketService{
		Upgrader: upgrader,
	}
}
//TradeHandler is the main request handler for different endpoints, that render your application's Logics and pages
// using the Chi router and a custom webclient package to handle requests and generate responses.
//TradeHandler contains the chi mux that implements the ServeMux method
type TradeHandler struct {
	mux              *chi.Mux
	RESTAPI *webclient.WebService
	WebSocket *SocketService
	HostSite string
}

//NewTradeHandler returns a new instance of *TradeHandler
func NewTradeHandler(HostSite string) TradeHandler {
	h := TradeHandler{
		mux:              chi.NewRouter(),
		RESTAPI: webclient.NewWebService(),		
		WebSocket: NewSocketService(HostSite),
		HostSite:  os.Getenv("HOSTSITE"),
	}
	h.mux.Get("/", h.indexHandler)
	h.mux.Get("/margins/ws", h.realTimeChartHandler)
	return h
}

//TradeHandler implements ServeHTTP method making it a Handler
func (h TradeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/favicon.ico") {
		http.NotFoundHandler().ServeHTTP(w, r)
	} else if strings.HasPrefix(r.URL.Path, "/images/") {
		// Serve images from the images subdirectory of webclient/assets
		http.StripPrefix("/images/", http.FileServer(http.Dir("./webclient/assets/images/"))).ServeHTTP(w, r)
	} else if strings.HasPrefix(r.URL.Path, "/webclient/assets/") {
		http.StripPrefix("/webclient/assets/", http.FileServer(http.Dir("./webclient/assets/"))).ServeHTTP(w, r)
	} else {
		h.mux.ServeHTTP(w, r)
	}
}

func (h TradeHandler) indexHandler(w http.ResponseWriter, r *http.Request) {
	// Replace the following with your actual logic
	// You might need to fetch data, populate the template, and execute it using your webclient
	// Here, we're using dummy data for illustration purposes
	data := struct {
		Message string
	}{
		Message: "Welcome to the index page",
	}

	usr := struct{}{} // Placeholder for user data
	msg := struct{}{} // Placeholder for messages
	code := model.ErrInternal // Placeholder for error code

	err := h.RESTAPI.Execute(w, r, data, usr, msg, code)
	if err != nil {
		// Handle error if needed
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (h TradeHandler) realTimeChartHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := h.WebSocket.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Failed to upgrade connection to WebSocket", http.StatusInternalServerError)
		log.Println("WebSocket upgrade error:", err)
		return
	}
	defer conn.Close()

	for {
		// Simulate generating real-time data
		// Replace this with your actual data fetching logic
		// For illustration, we're generating random data here
		data := struct {
			ClosingPrices float64 `json:"ClosingPrices"`
			Timestamps    int64   `json:"Timestamps"`
			Signals       string  `json:"Signals"`
			ShortEMA      float64 `json:"ShortEMA"`
			LongEMA       float64 `json:"LongEMA"`
		}{
			ClosingPrices:  100 + rand.Float64()*50, // Generate random ClosingPrices
			Timestamps:     time.Now().Unix(),       // Current timestamp
			Signals:        "Buy",                   // Buy signal for illustration
			ShortEMA:       110 + rand.Float64()*10, // Generate random ShortEMA
			LongEMA:        105 + rand.Float64()*10, // Generate random LongEMA
		}

		err := conn.WriteJSON(data)
		if err != nil {
			log.Println("WebSocket write error:", err)
			break
		}

		// Simulate real-time updates every second
		time.Sleep(time.Second)
	}
}


// ValidateRedirectURL checks that the URL provided is valid.
// If the URL is missing, redirect the user to the application's root.
// The URL must not be absolute (i.e., the URL must refer to a path within this
// application).
func validateRedirectURL(path string) (string, error) {
	if path == "" {
		return "/", nil
	}

	// Ensure redirect URL is valid and not pointing to a different server.
	parsedURL, err := url.Parse(path)
	if err != nil {
		return "/", err
	}
	if parsedURL.IsAbs() {
		return "/", errors.New("URL invalid: URL must be absolute")
	}
	if strings.Contains(path, "/signup?redirect=") {
		return path[17:], nil
	}
	return path, nil
}