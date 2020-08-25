package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/bep/gotfmt/formatter"
)

func main() {
	if len(os.Args) == 2 && os.Args[1] == "-h" {
		fmt.Println("usage: cat template.html | gotfmt")
		return
	}

	log.SetPrefix("gotfmt: error: ")
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))

	stat, _ := os.Stdin.Stat()
	isPipe := (stat.Mode() & os.ModeCharDevice) == 0
	if !isPipe {
		log.Fatal("Nothing to read")
	}
	b, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	f := formatter.Formatter{}
	s, err := f.Format(string(b))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Fprint(os.Stdout, s)
}
