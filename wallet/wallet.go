package wallet

import "sync"

type Entry struct {
	UserMsg      string  `json:"user_msg"`
	AssistantMsg string  `json:"assistant_msg"`
	InputTokens  int     `json:"input_tokens"`
	OutputTokens int     `json:"output_tokens"`
	CostBDT      float64 `json:"cost_bdt"`
	Provider     string  `json:"provider"`
	Timestamp    string  `json:"timestamp"`
}

type Wallet struct {
	mu      sync.Mutex
	balance float64
	history []Entry
}

func New(initial float64) *Wallet {
	return &Wallet{balance: initial}
}

func (w *Wallet) Debit(cost float64, entry Entry) (float64, []Entry) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.balance -= cost
	w.history = append(w.history, entry)
	h := make([]Entry, len(w.history))
	copy(h, w.history)
	return w.balance, h
}

func (w *Wallet) Reset(initial float64) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.balance = initial
	w.history = nil
}
