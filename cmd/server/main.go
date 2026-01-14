package main

import (
	"log"
	"net/http"

	"github.com/ankitbahl/comic-compiler-backend/internal/router"
)

func main() {
	r := router.New()

	log.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
