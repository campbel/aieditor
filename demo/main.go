package main

import (
    "net/http"
    "math"
    "strconv"
    "github.com/charmbracelet/log"
)

func main() {
    log.Println("Starting...")
    http.HandleFunc("/", handler)
    http.HandleFunc("/prime", primeHandler)
    err := http.ListenAndServe(":8080", nil)
    if err != nil {
        log.Println(err)
    }
}

func handler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintln(w, "Hello World")
}

func primeHandler(w http.ResponseWriter, r *http.Request) {
    q := r.URL.Query().Get("number")
    num, err := strconv.Atoi(q)
    if err != nil {
        log.Println(err)
    }
    for i := 2; i < int(math.Sqrt(float64(num))+1); i++ {
        if num % i == 0 {
            fmt.Fprintln(w, "Number is not prime")
            return
        }
    }
    fmt.Fprintln(w, "Number is prime")
}