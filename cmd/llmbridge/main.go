package main

import (
	"log"
	"net/http"
	"os"

	"llmbridge/config"
	"llmbridge/handler"
	"llmbridge/lago"
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

	lagoClient := lago.New(cfg.LagoURL, cfg.LagoAPIKey)
	if lagoClient.Enabled() {
		if err := lagoClient.EnsureMetric(); err != nil {
			log.Printf("[Lago] setup warning: %v", err)
		}
	} else {
		log.Printf("[Lago] no API key set — billing disabled")
	}

	w := wallet.New(initialBalanceBDT)
	h := handler.New(cfg, w, lagoClient, initialBalanceBDT, indexHTML)

	http.HandleFunc("/", h.Index)
	http.HandleFunc("/chat", h.Chat)
	http.HandleFunc("/reset", h.Reset)
	http.HandleFunc("/config", h.Config)

	log.Printf("LLMBridge running at http://localhost%s  (model: %s)", listenAddr, cfg.DefaultModel)
	log.Fatal(http.ListenAndServe(listenAddr, nil))
}
