package server

import (
	"log"
	"net"
	"net/http"
)

//Server, handles the opening and closing of an HTTP server using the net/http and gorilla/handlers packages. 
//it holds the TradeHandler type which 
type Server struct {
	Listener net.Listener
	// Handler to serve http.
	HttpHandler TradeHandler
	// Bind address to open for http.
	Port string
}

func NewServer(port string, th TradeHandler) *Server {
	return &Server{
		HttpHandler:  th,
		Port:        ":" + port,
	}
}

//Open method starts the HTTP server using the provided TradeHandler and logs requests using CombinedLoggingHandler 
//from the gorilla/handlers package. 
func (s *Server) Open() (err error) {	
    log.Println("Opening server...")
	s.Listener, err = net.Listen("tcp", s.Port)
    if err != nil {
        log.Fatalf("Error while opening listener: %v", err)
        return err
    }

    log.Printf("Server listening on %s", s.Port)

    // Start serving
	// log.Fatal(http.Serve(s.Listener, handlers.CombinedLoggingHandler(os.Stderr, s.HttpHandler)))
	log.Fatal(http.Serve(s.Listener, s.HttpHandler))
	return nil
}

//Close method is responsible for closing the server's socket.
func (s *Server) Close() error {
	if s.Listener != nil {
		s.Listener.Close()
	}
	return nil
}