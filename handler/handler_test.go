package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"llmbridge/config"
	"llmbridge/wallet"
)

func newTestHandler(bifrostURL string) *Handler {
	cfg := &config.Config{
		BifrostURL:   bifrostURL,
		DefaultModel: "ollama/qwen2.5:3b",
		Providers: []config.Provider{
			{ID: "ollama", Name: "Ollama", Model: "ollama/qwen2.5:3b"},
		},
	}
	w := wallet.New(100.0)
	return New(cfg, w, 100.0, []byte("<html>test</html>"))
}

func mockBifrost(reply string, inputTokens, outputTokens int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := bifrostResponse{}
		resp.Choices = []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		}{{}}
		resp.Choices[0].Message.Content = reply
		resp.Usage.PromptTokens = inputTokens
		resp.Usage.CompletionTokens = outputTokens
		json.NewEncoder(w).Encode(resp)
	}))
}

func TestIndex(t *testing.T) {
	h := newTestHandler("http://localhost")
	rr := httptest.NewRecorder()
	h.Index(rr, httptest.NewRequest(http.MethodGet, "/", nil))
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "<html>") {
		t.Error("expected HTML in response body")
	}
}

func TestChat_MethodNotAllowed(t *testing.T) {
	h := newTestHandler("http://localhost")
	rr := httptest.NewRecorder()
	h.Chat(rr, httptest.NewRequest(http.MethodGet, "/chat", nil))
	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rr.Code)
	}
}

func TestChat_EmptyMessage(t *testing.T) {
	h := newTestHandler("http://localhost")
	rr := httptest.NewRecorder()
	body := strings.NewReader(`{"message":""}`)
	h.Chat(rr, httptest.NewRequest(http.MethodPost, "/chat", body))
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestChat_InvalidJSON(t *testing.T) {
	h := newTestHandler("http://localhost")
	rr := httptest.NewRecorder()
	body := strings.NewReader(`not json`)
	h.Chat(rr, httptest.NewRequest(http.MethodPost, "/chat", body))
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestChat_BifrostUnreachable(t *testing.T) {
	h := newTestHandler("http://127.0.0.1:1")
	rr := httptest.NewRecorder()
	body := strings.NewReader(`{"message":"hello"}`)
	h.Chat(rr, httptest.NewRequest(http.MethodPost, "/chat", body))
	if rr.Code != http.StatusBadGateway {
		t.Errorf("expected 502, got %d", rr.Code)
	}
}

func TestChat_Success(t *testing.T) {
	srv := mockBifrost("hi there", 10, 5)
	defer srv.Close()

	h := newTestHandler(srv.URL)
	rr := httptest.NewRecorder()
	body := strings.NewReader(`{"message":"hello","model":"ollama/qwen2.5:3b"}`)
	h.Chat(rr, httptest.NewRequest(http.MethodPost, "/chat", body))

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	var resp chatResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Reply != "hi there" {
		t.Errorf("expected reply 'hi there', got '%s'", resp.Reply)
	}
	if resp.InputTokens != 10 {
		t.Errorf("expected 10 input tokens, got %d", resp.InputTokens)
	}
	if resp.OutputTokens != 5 {
		t.Errorf("expected 5 output tokens, got %d", resp.OutputTokens)
	}
	expectedCost := 10*inputPriceBDT + 5*outputPriceBDT
	if resp.CostBDT != expectedCost {
		t.Errorf("expected cost %f, got %f", expectedCost, resp.CostBDT)
	}
	if resp.Balance != 100.0-expectedCost {
		t.Errorf("expected balance %f, got %f", 100.0-expectedCost, resp.Balance)
	}
	if len(resp.History) != 1 {
		t.Errorf("expected 1 history entry, got %d", len(resp.History))
	}
}

func TestChat_UsesDefaultModel(t *testing.T) {
	srv := mockBifrost("ok", 1, 1)
	defer srv.Close()

	h := newTestHandler(srv.URL)
	rr := httptest.NewRecorder()
	body := strings.NewReader(`{"message":"hello"}`)
	h.Chat(rr, httptest.NewRequest(http.MethodPost, "/chat", body))

	var resp chatResponse
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp.Provider != "ollama/qwen2.5:3b" {
		t.Errorf("expected default model 'ollama/qwen2.5:3b', got '%s'", resp.Provider)
	}
}

func TestReset(t *testing.T) {
	srv := mockBifrost("hi", 10, 10)
	defer srv.Close()

	h := newTestHandler(srv.URL)

	// spend some balance first
	rr := httptest.NewRecorder()
	h.Chat(rr, httptest.NewRequest(http.MethodPost, "/chat", strings.NewReader(`{"message":"hello"}`)))

	// reset
	rr = httptest.NewRecorder()
	h.Reset(rr, httptest.NewRequest(http.MethodPost, "/reset", nil))

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "100.00") {
		t.Errorf("expected balance 100.00 after reset, got: %s", rr.Body.String())
	}
}

func TestConfig(t *testing.T) {
	h := newTestHandler("http://localhost")
	rr := httptest.NewRecorder()
	h.Config(rr, httptest.NewRequest(http.MethodGet, "/config", nil))

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	var resp map[string]any
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode config response: %v", err)
	}
	if resp["default_model"] != "ollama/qwen2.5:3b" {
		t.Errorf("unexpected default_model: %v", resp["default_model"])
	}
	providers, ok := resp["providers"].([]any)
	if !ok || len(providers) != 1 {
		t.Errorf("expected 1 provider, got: %v", resp["providers"])
	}
}
