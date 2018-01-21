package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/howdy", howdy)
	log.Fatal(http.ListenAndServe(":8000", nil))
}

// howdy echoes howdy
func howdy(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Howdy, this is Heihei\n")
}
