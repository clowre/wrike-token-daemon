package wrikehttp

import (
	"fmt"
	"log"
	"net"
	"net/http"

	wrikedaemon "github.com/clowre/wrike-token-daemon"
)

func StartServer(port int, daemon *wrikedaemon.Daemon) error {
	handler := createHandler(daemon)

	httpListener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("cannot create tcp/http listener: %v", err)
	}

	log.Printf("starting http server on [%s]", httpListener.Addr().String())

	httpServer := &http.Server{Handler: handler}
	return httpServer.Serve(httpListener)
}
