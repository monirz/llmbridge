package lago

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type Client struct {
	url    string
	apiKey string
	http   *http.Client
}

func New(url, apiKey string) *Client {
	return &Client{
		url:    url,
		apiKey: apiKey,
		http:   &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) Enabled() bool {
	return c.apiKey != ""
}

// EnsureMetric creates the tokens_used billable metric if it doesn't exist.
func (c *Client) EnsureMetric() error {
	if !c.Enabled() {
		return nil
	}

	req, _ := http.NewRequest(http.MethodGet, c.url+"/api/v1/billable_metrics/tokens_used", nil)
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("lago unreachable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		log.Printf("[Lago] billable metric 'tokens_used' already exists")
		return nil
	}

	metric := map[string]any{
		"billable_metric": map[string]any{
			"name":             "Tokens Used",
			"code":             "tokens_used",
			"aggregation_type": "sum_agg",
			"field_name":       "total_tokens",
		},
	}
	body, _ := json.Marshal(metric)
	req, _ = http.NewRequest(http.MethodPost, c.url+"/api/v1/billable_metrics", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	resp, err = c.http.Do(req)
	if err != nil {
		return fmt.Errorf("create metric failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("create metric rejected: %s", resp.Status)
	}
	log.Printf("[Lago] billable metric 'tokens_used' created")
	return nil
}

// SendEvent fires a usage event to Lago asynchronously.
func (c *Client) SendEvent(customerID, transactionID, model string, inputTokens, outputTokens int) {
	if !c.Enabled() {
		return
	}
	go func() {
		payload := map[string]any{
			"event": map[string]any{
				"transaction_id":       transactionID,
				"external_customer_id": customerID,
				"code":                 "tokens_used",
				"timestamp":            time.Now().Unix(),
				"properties": map[string]any{
					"input_tokens":  inputTokens,
					"output_tokens": outputTokens,
					"total_tokens":  inputTokens + outputTokens,
					"model":         model,
				},
			},
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest(http.MethodPost, c.url+"/api/v1/events", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
		resp, err := c.http.Do(req)
		if err != nil {
			log.Printf("[Lago] event send failed: %v", err)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 300 {
			log.Printf("[Lago] event rejected: %s", resp.Status)
			return
		}
		log.Printf("[Lago] event sent: customer=%s model=%s tokens=%d", customerID, model, inputTokens+outputTokens)
	}()
}
