package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"github.com/elankath/zlook"
	"github.com/pkg/errors"
)

//flag.IntVar(&flagvar, "flagname", 1234, "help message for flagname")
//var depthPtr = flag.Int("d", 1, "Maximum Depth to which to inspect archive")
//var archiveTypes = flag.Var

func main() {
	var input zlook.Input
	flag.IntVar(&input.MaxDepth, "d", 1, "Maximum Depth to which to inspect archive")
	flag.Var(&input.ArchiveTypes, "t", fmt.Sprintf("Comma Separated Archive Types (default %s)", zlook.DefaultArchiveTypes))
	flag.BoolVar(&input.PrefixPath, "p", true, "Prefix Parent Path")
	flag.StringVar(&input.ExtractEntry, "x", "", "Extract matching entry (suffix match) to stdout")
	flag.Parse()
	log.SetFlags(0)
	input.Paths = flag.Args()
	if len(input.Paths) == 0 {
		log.Fatal(errors.New("zlook: No files to look at"))
	}
	for _, f := range input.Paths {
		if fi, err := os.Stat(f); err != nil {
			log.Fatal(errors.Wrap(err, "zlook: cant get file info"))
		} else if fi.IsDir() {
			log.Fatal(fmt.Errorf("zlook: %s is a directory not archive", fi.Name()))
		}
	}
	inspector := zlook.NewInspector(input)
	if input.ExtractEntry != "" {
		inspector.Extract()
	} else {
		inspector.List()
	}
}
