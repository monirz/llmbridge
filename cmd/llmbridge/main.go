package main

import (
	"log"
	"net/http"
	"os"

	"llmbridge/config"
	"llmbridge/handler"
	"llmbridge/wallet"
)

const (
	initialBalanceBDT = 100.0
	listenAddr        = ":9000"
)

func main() {
	cfg := config.Load()

	indexHTML, err := os.ReadFile("templates/index.html")
	if err != nil {
		log.Fatalf("cannot read templates/index.html: %v", err)
	}

	w := wallet.New(initialBalanceBDT)
	h := handler.New(cfg, w, initialBalanceBDT, indexHTML)

	http.HandleFunc("/", h.Index)
	http.HandleFunc("/chat", h.Chat)
	http.HandleFunc("/reset", h.Reset)
	http.HandleFunc("/config", h.Config)

	log.Printf("LLMBridge running at http://localhost%s  (model: %s)", listenAddr, cfg.DefaultModel)
	log.Fatal(http.ListenAndServe(listenAddr, nil))
}
