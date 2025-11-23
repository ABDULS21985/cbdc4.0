package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/offline/purse", func(w http.ResponseWriter, r *http.Request) {
		// Issue a new offline purse certificate
		fmt.Fprintf(w, `{"status": "issued", "certificate": "signed-blob"}`)
	})

	http.HandleFunc("/offline/reconcile", func(w http.ResponseWriter, r *http.Request) {
		// Process offline transactions
		fmt.Fprintf(w, `{"status": "processed", "count": 5}`)
	})

	log.Println("Offline Service running on :8084")
	log.Fatal(http.ListenAndServe(":8084", nil))
}
