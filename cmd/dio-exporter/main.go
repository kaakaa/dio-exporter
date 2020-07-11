package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/kaakaa/dio-exporter/pkg"
)

var (
	in          = flag.String("in", "", "input dir")
	out         = flag.String("out", "", "output dir")
	debug       = flag.Bool("debug", false, "set debug flag")
	format      = flag.String("format", "png", "format [png, svg]")
	bgcolor     = flag.String("bg", "white", "Background color (e.g.: white, red, #cccccc)")
	debugServer = flag.Bool("debug-server", false, "Run drawio server for debug")
)

func main() {
	flag.Parse()
	pkg.DebugMode = *debug

	// Run drawio server
	wg := &sync.WaitGroup{}
	srv, addr := pkg.StartDrawioServer(wg)

	if *debugServer {
		log.Println("Enter 'Ctrl + c' to stop debug server")
		wg.Wait()
		return
	}

	if err := validation(); err != nil {
		log.Fatalf("Parameter is not valid: %v", err)
	}

	// Convert diagrams
	diagrams, err := pkg.ReadDir(*in, []string{".dio", ".drawio"})
	if err != nil {
		log.Fatalf("failed to read diagrams from dir. %v", err)
	}
	for _, d := range diagrams {
		d.Export(addr, *format, *bgcolor, *out)
	}

	// Shutdown drawio server
	if err := srv.Shutdown(context.TODO()); err != nil {
		panic(err)
	}
	wg.Wait()

	log.Println("Succsess to shutdown server.")
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
