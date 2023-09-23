package server

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"coinbitly.com/strategies"
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
func NewSocketService(HostSite string) SocketService {
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
	return SocketService{
		Upgrader: upgrader,
	}
}

//TradeHandler is the main request handler for different endpoints, that render your application's Logics and pages
// using the Chi router and a custom webclient package to handle requests and generate responses.
//TradeHandler contains the chi mux that implements the ServeMux method
type TradeHandler struct {
	mux       *chi.Mux
	RESTAPI   *webclient.WebService
	WebSocket SocketService
	HostSite  string
	ts        *strategies.TradingSystem
}

//NewTradeHandler returns a new instance of *TradeHandler
func NewTradeHandler(ts *strategies.TradingSystem, HostSite string) TradeHandler {
	h := TradeHandler{
		mux:       chi.NewRouter(),
		RESTAPI:   webclient.NewWebService(),
		WebSocket: NewSocketService(HostSite),
		HostSite:  os.Getenv("HOSTSITE"),
		ts:        ts,
	}
	h.mux.Get("/ImageReceiver/ws", h.ImageReceiverHandler)
	h.mux.Get("/FeedsTradingSystem/ws", h.realTimeTradingSystemFeed)
	h.mux.Get("/FeedsAppData/ws", h.realTimeAppDataFeed)
	h.mux.Post("/updateZoom", h.updateZoomHandler)
	// Add an OPTIONS route for /updateZoom
	h.mux.Options("/updateZoom", func(w http.ResponseWriter, r *http.Request) {
		// Add an OPTIONS route for /updateZoom
    	w.Header().Set("Access-Control-Allow-Origin", "http://resoledge.com") //Adjust 
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.WriteHeader(http.StatusOK)
	})
	return h
}

//TradeHandler implements ServeHTTP method making it a Handler
func (h TradeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    log.Printf("Received request: %s %s", r.Method, r.URL.Path)
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
func (h TradeHandler) updateZoomHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the zoom value from the request body
	var zoomValue struct {
		Zoom int `json:"zoom"`
	}
	if err := json.NewDecoder(r.Body).Decode(&zoomValue); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Update the zoom parameter in your TradingSystem
	h.ts.Mu.Lock()
	if zoomValue.Zoom < 499 && zoomValue.Zoom > 0{
		h.ts.Zoom = zoomValue.Zoom
	}
	h.ts.Mu.Unlock()
	// log.Printf("Zoom = %d\n", zoomValue.Zoom)

	// Regenerate the image with the updated zoom
	err := h.ts.Reporting(<-h.ts.MDChan, "zoomimg")
	if err != nil {
		log.Printf("Error repoting zoom at handler %v", err)
		http.Error(w, "Failed to generate zoomed image", http.StatusInternalServerError)
		return
	}

	// You can send a success response if needed
	w.WriteHeader(http.StatusOK)
}

func (h TradeHandler) realTimeAppDataFeed(w http.ResponseWriter, r *http.Request) {
	conn, err := h.WebSocket.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Failed to upgrade connection to WebSocket", http.StatusInternalServerError)
		log.Println("realTimeAppDataFeed: WebSocket upgrade error:", err)
		return
	}
	// log.Println("realTimeAppDataFeed: WebSocket Connected!!!")
	defer conn.Close()
	var AppDataJSON []byte
	// h.ts.ADataChan = make(chan []byte)
Loop:
	for {
		select {
		case <-time.After(time.Second * 10):
			break Loop
		case AppDataJSON = <-h.ts.ADataChan:
			// Send the JSON data to the WebSocket client
			if err = conn.WriteMessage(websocket.TextMessage, AppDataJSON); err != nil {
				log.Println("realTimeAppDataFeed1: WebSocket write error:", err)
				break Loop
			}
		}
	}
	log.Println("realTimeAppDataFeed: going away!!!")
	return
}
func (h TradeHandler) realTimeTradingSystemFeed(w http.ResponseWriter, r *http.Request) {
	conn, err := h.WebSocket.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Failed to upgrade connection to WebSocket", http.StatusInternalServerError)
		log.Println("realTimeTradingSystemFeed: WebSocket upgrade error:", err)
		return
	}
	// log.Println("realTimeTradingSystemFeed: WebSocket Connected!!!")
	defer conn.Close()
	var tradingSystemJSON []byte
	// h.ts.TSDataChan = make(chan []byte)
Loop:
	for {
		select {
		case <-time.After(time.Second * 10):
			break Loop
		case tradingSystemJSON = <-h.ts.TSDataChan:
			// Send the JSON data to the WebSocket client
			if err = conn.WriteMessage(websocket.TextMessage, tradingSystemJSON); err != nil {
				log.Println("realTimeTradingSystemFeed1: WebSocket write error:", err)
				break Loop
			}
		}
	}
	log.Println("realTimeTradingSystemFeed: going away!!!")
	return
}

func (h TradeHandler) ImageReceiverHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := h.WebSocket.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Failed to upgrade connection to WebSocket", http.StatusInternalServerError)
		log.Println("ImageReceiverHandler: WebSocket upgrade error:", err)
		return
	}
	defer conn.Close()

	for {
		// Load an image
		imagePath := "./webclient/assets/line_chart_with_signals.png"
		imageData, err := loadImageData(imagePath)
		if err != nil {
			fmt.Println(err)
			return
		}

		// Encode image data to base64
		base64ImageData := base64.StdEncoding.EncodeToString(imageData)

		// Send the encoded image data over WebSocket
		if err = conn.WriteMessage(websocket.TextMessage, []byte(base64ImageData)); err != nil {
			fmt.Printf("Unable to write to socket: %v", err)
			return
		}
		time.Sleep(h.ts.EpochTime / 2)
	}
}

func loadImageData(imagePath string) ([]byte, error) {
	// Load image from file
	imageFile, err := os.Open(imagePath)
	if err != nil {
		return nil, err
	}
	defer imageFile.Close()

	// Decode image
	img, _, err := image.Decode(imageFile)
	if err != nil {
		return nil, err
	}

	// Encode image as PNG
	var imageData []byte
	pngBuffer := &bytes.Buffer{}
	err = png.Encode(pngBuffer, img)
	if err != nil {
		return nil, err
	}
	imageData = pngBuffer.Bytes()

	return imageData, nil
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
