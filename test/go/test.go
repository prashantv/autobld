package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	fmt.Println("Started go app!")
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":8080", nil))
	fmt.Println("Ended go app")
}

func handler(w http.ResponseWriter, r *http.Request) {
	data := "Hello v1\n"
	w.Write([]byte(data))
}
