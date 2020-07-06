package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"sync"

	"github.com/markbates/pkger"
)

var (
	in      = flag.String("in", "", "input dir")
	out     = flag.String("out", "", "output dir")
	debug   = flag.Bool("debug", false, "set debug flag")
	format  = flag.String("format", "png", "format [png, svg]")
	bgcolor = flag.String("bg", "white", "Background color (e.g.: white, red, #cccccc)")
)

func startDrawioServer(wg *sync.WaitGroup) (*http.Server, string) {
	// Get free port
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatal(err)
	}
	logf("Run drawio server at %s", l.Addr().String())

	// Run server
	srv := &http.Server{}
	go func() {
		defer wg.Done()
		fs := http.FileServer(pkger.Dir("/drawio/src/main/webapp"))
		http.Handle("/", fs)

		wg.Add(1)

		if err := srv.Serve(l); err != http.ErrServerClosed {
			logf("Unexpecte close: %v", err)
		}
	}()
	return srv, l.Addr().String()
}

func main() {
	flag.Parse()
	if err := validation(); err != nil {
		fatalf("Parameter is not valid: %v", err)
	}

	// Run drawio server
	wg := &sync.WaitGroup{}
	srv, addr := startDrawioServer(wg)

	// Convert diagrams
	logf("Start exporting.")
	diagrams, err := ReadDir(*in, []string{".dio", ".drawio"})
	if err != nil {
		fatalf("failed to read diagrams from dir. %v", err)
	}
	for _, d := range diagrams {
		d.Export(addr, *format, *bgcolor, *out)
	}
	logf("Exporting is done.")

	// Shutdown drawio server
	if err := srv.Shutdown(context.TODO()); err != nil {
		panic(err)
	}
	wg.Wait()

	logf("Succsess to shutdown server.")
}

func validation() error {
	if len(*in) == 0 {
		return fmt.Errorf("-in must not empty")
	}
	if len(*out) == 0 {
		return fmt.Errorf("-out must not empty")
	}

	if _, err := os.Stat(*in); os.IsNotExist(err) {
		return fmt.Errorf("input dir is not exist: %s", *in)
	}

	if *format != "png" && *format != "svg" {
		return fmt.Errorf(`Invalid format is specified. "png" or "svg" is valid, but you specify [%s]`, *format)
	}
	return nil
}
