package server

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"coinbitly.com/model"
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
	mux              *chi.Mux
	RESTAPI *webclient.WebService
	WebSocket SocketService
	HostSite string
	ts *strategies.TradingSystem
}

//NewTradeHandler returns a new instance of *TradeHandler
func NewTradeHandler(ts *strategies.TradingSystem, HostSite string) TradeHandler {
	h := TradeHandler{
		mux:              chi.NewRouter(),
		RESTAPI: webclient.NewWebService(),		
		WebSocket: NewSocketService(HostSite),
		HostSite:  os.Getenv("HOSTSITE"),
		ts: ts,
	}
	h.mux.Get("/ImageReceiver/ws", h.ImageReceiverHandler)
	h.mux.Get("/FeedsTradingSystem/ws", h.realTimeTradingSystemFeed)
	h.mux.Get("/FeedsAppData/ws", h.realTimeAppDataFeed)
	h.mux.Get("/margins/ws", h.realTimeChartHandler)//this.socket = new WebSocket.w3cwebsocket("ws://localhost:35260/ImageReceiver/ws");
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

func (h TradeHandler)realTimeAppDataFeed(w http.ResponseWriter, r *http.Request){
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
		select{
		case <- time.After(time.Second * 10):
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
func (h TradeHandler)realTimeTradingSystemFeed(w http.ResponseWriter, r *http.Request){
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
		select{
		case <- time.After(time.Second * 10):
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
func (h TradeHandler) realTimeChartHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := h.WebSocket.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Failed to upgrade connection to WebSocket", http.StatusInternalServerError)
		log.Println("realTimeChartHandler: WebSocket upgrade error:", err)
		return
	}
	defer conn.Close()
	var cd model.ChartData
	for {
		select{
		case cd = <-h.ts.ChartChan:
			fmt.Println(cd)
		default:
		}
		data := struct {
			ClosingPrices float64 `json:"ClosingPrices"`
			Timestamps    int64   `json:"Timestamps"`
			Signals       string  `json:"Signals"`
			ShortEMA      float64 `json:"ShortEMA"`
			LongEMA       float64 `json:"LongEMA"`
		}{
			ClosingPrices:  cd.ClosingPrices,
			Timestamps:     cd.Timestamps,
			Signals:        cd.Signals,
			ShortEMA:       cd.ShortEMA,
			LongEMA:        cd.LongEMA,
		}
		if data.ClosingPrices > 110.0 {
			data.Signals = "Sell"
		}		

		if err := conn.WriteJSON(data); err != nil {
			log.Println("realTimeChartHandler: WebSocket write error:", err)
			break
		}

		// Simulate real-time updates every second
		time.Sleep(time.Second)
	}
}

func (h TradeHandler) ImageReceiverHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := h.WebSocket.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Failed to upgrade connection to WebSocket", http.StatusInternalServerError)
		log.Println("ImageReceiverHandler: WebSocket upgrade error:", err)
		return
	}
	defer conn.Close()

	for{		
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
		time.Sleep(h.ts.EpochTime)
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