package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Application Go executee dans Docker !")
	})

	log.Println("Serveur demarre sur le port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
