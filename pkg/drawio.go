package pkg

import (
	"log"
	"net"
	"net/http"
	"sync"

	"github.com/markbates/pkger"
)

func StartDrawioServer(wg *sync.WaitGroup) (*http.Server, string) {
	// Get free port
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatal(err)
	}
	logf("Run drawio server at %s", l.Addr().String())

	// Run server
	srv := &http.Server{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		fs := http.FileServer(pkger.Dir("/drawio/src/main/webapp"))
		http.Handle("/", fs)

		if err := srv.Serve(l); err != http.ErrServerClosed {
			logf("Unexpecte close: %v", err)
		}
	}()
	return srv, l.Addr().String()
}
