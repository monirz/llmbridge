package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"llmbridge/config"
	"llmbridge/wallet"
)

const (
	inputPriceBDT  = 0.000018
	outputPriceBDT = 0.000072
)

type Handler struct {
	cfg       *config.Config
	wallet    *wallet.Wallet
	initial   float64
	indexHTML []byte
}

func New(cfg *config.Config, w *wallet.Wallet, initial float64, indexHTML []byte) *Handler {
	return &Handler{cfg: cfg, wallet: w, initial: initial, indexHTML: indexHTML}
}

type chatRequest struct {
	Message string `json:"message"`
	Model   string `json:"model"`
}

type chatResponse struct {
	Reply        string         `json:"reply"`
	InputTokens  int            `json:"input_tokens"`
	OutputTokens int            `json:"output_tokens"`
	CostBDT      float64        `json:"cost_bdt"`
	Balance      float64        `json:"balance"`
	History      []wallet.Entry `json:"history"`
	Provider     string         `json:"provider"`
}

type bifrostRequest struct {
	Model    string    `json:"model"`
	Messages []message `json:"messages"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type bifrostResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
}

func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write(h.indexHTML)
}

func (h *Handler) Chat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req chatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Message == "" {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	model := req.Model
	if model == "" {
		model = h.cfg.DefaultModel
	}

	payload, _ := json.Marshal(bifrostRequest{
		Model:    model,
		Messages: []message{{Role: "user", Content: req.Message}},
	})

	log.Printf("[Step 2] → Bifrost  model=%s", model)

	httpClient := &http.Client{Timeout: 60 * time.Second}
	httpReq, _ := http.NewRequest(http.MethodPost, h.cfg.BifrostURL, bytes.NewReader(payload))
	httpReq.Header.Set("Content-Type", "application/json")
	if apiKey := os.Getenv("API_KEY"); apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	}

	log.Printf("[Step 3] Bifrost → %s", model)
	resp, err := httpClient.Do(httpReq)
	if err != nil {
		http.Error(w, "bifrost unreachable: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	var br bifrostResponse
	rawBody, _ := io.ReadAll(resp.Body)
	log.Printf("[Bifrost raw response] %s", string(rawBody))
	if err := json.Unmarshal(rawBody, &br); err != nil {
		http.Error(w, "bad response from bifrost", http.StatusInternalServerError)
		return
	}

	reply := ""
	if len(br.Choices) > 0 {
		reply = br.Choices[0].Message.Content
	}

	inputTokens := br.Usage.PromptTokens
	outputTokens := br.Usage.CompletionTokens
	cost := float64(inputTokens)*inputPriceBDT + float64(outputTokens)*outputPriceBDT

	log.Printf("[Step 4] ← %d input + %d output tokens", inputTokens, outputTokens)
	log.Printf("[Step 5] → Lago  input=%d output=%d cost=%.6f BDT", inputTokens, outputTokens, cost)

	balance, history := h.wallet.Debit(cost, wallet.Entry{
		UserMsg:      req.Message,
		AssistantMsg: reply,
		InputTokens:  inputTokens,
		OutputTokens: outputTokens,
		CostBDT:      cost,
		Provider:     model,
		Timestamp:    time.Now().Format("15:04:05"),
	})

	log.Printf("[Step 6] Lago debited %.6f BDT  balance=%.4f BDT", cost, balance)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(chatResponse{
		Reply:        reply,
		InputTokens:  inputTokens,
		OutputTokens: outputTokens,
		CostBDT:      cost,
		Balance:      balance,
		History:      history,
		Provider:     model,
	})
}

func (h *Handler) Reset(w http.ResponseWriter, r *http.Request) {
	h.wallet.Reset(h.initial)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"balance":%.2f}`, h.initial)
}

func (h *Handler) Config(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"default_model": h.cfg.DefaultModel,
		"providers":     h.cfg.Providers,
	})
}
