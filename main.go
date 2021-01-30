package main

import (
	"flag"
	"log"
)

const (
	defaultVersionURL = "https://stable.release.flatcar-linux.net/amd64-usr/current/version.txt"
)

var (
	flagVersionURL = flag.String("version-url", defaultVersionURL, "Remote location of a version.txt file")
)

func main() {
	flag.Parse()

	ue, err := newUpdateEngine(*flagVersionURL)
	if err != nil {
		log.Fatalln(err)
	}

	ue.run()
}
