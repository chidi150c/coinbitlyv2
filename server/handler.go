package server

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/pkg/errors"
	"coinbitly.com/webclient"
	"coinbitly.com/model"
	"github.com/go-chi/chi"
)

//TradeHandler is the main request handler for different endpoints, that render your application's Logics and pages
// using the Chi router and a custom webclient package to handle requests and generate responses.
//TradeHandler contains the chi mux that implements the ServeMux method
type TradeHandler struct {
	mux              *chi.Mux
	Response *webclient.WebService
	
}

//NewTradeHandler returns a new instance of *TradeHandler
func NewTradeHandler() TradeHandler {
	h := TradeHandler{
		mux:              chi.NewRouter(),
		Response: webclient.NewWebService(),
	}
	h.mux.Get("/", h.indexHandler)
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

	err := h.Response.Execute(w, r, data, usr, msg, code)
	if err != nil {
		// Handle error if needed
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
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