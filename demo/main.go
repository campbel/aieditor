package main

import (
	"fmt"
	"net/http"

	"github.com/charmbracelet/log"
)

func main() {
	log.Info("starting...")
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello, world")
}
