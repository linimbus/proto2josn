package main

import (
	"flag"
	"fmt"
	"os"
)

var (
	Help         bool
	Version      bool
	ImportPaths  string
	OutputFormat string
)

func init() {
	flag.BoolVar(&Help, "help", false, "show help")
	flag.BoolVar(&Version, "version", false, "show version")
	flag.StringVar(&ImportPaths, "import", "", "import paths")
	flag.StringVar(&OutputFormat, "format", "compact", "output format (compart or readable)")

}

func main() {

	flag.Parse()
	if Help || Version {
		os.Args[0] += "version: 0.1.0"
		flag.Usage()
		return
	}

	parser := NewProtoParser(WithImportPaths([]string{"./proto2"}))

	err := parser.ParseProtoFile("person.proto")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	parser.PrintProtoInfo()
}
