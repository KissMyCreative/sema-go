package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/jsnfwlr/vexil-go"
)

type Logger struct{}

func (l Logger) Log(message string) {
	fmt.Printf("%s\n", message)
}

func main() {
	l := Logger{}
	client := vexil.NewClient("http://localhost:9765", &http.Client{}, l)

	go func() {
		client.EventsToFunc("dev", func(flag vexil.Flag) {
			// Handle the received message
			fmt.Printf("Got flag from dev-env:\n\t%s (%s) = %q\n", flag.Name, flag.Type, flag.Value)
		})
	}()

	time.Sleep(10 * time.Minute)
}
