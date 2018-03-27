package main

import (
	"errors"
	"os"
)

func main() {
	url, err := parseURL()
	if err != nil {
		panic(err.Error())
	}
	jmx := NewJMX(url)
	cfg := NewConf(url)
	Start(jmx, cfg)
}

func parseURL() (string, error) {
	if len(os.Args) != 2 {
		return "", errors.New("host url not specified")
	}
	return os.Args[1], nil
}
