package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// TestOpenRouterTurn: with the built-in defaults and a user's file that
// allows calls to an external provider and puts the profiles on haiku, a
// turn goes to the OpenAI-compatible endpoint as OpenRouter is called: the
// model anthropic/claude-haiku-5.5, the key from OPENROUTER_API_KEY as a
// bearer token, and the engine's cache marks; its answer is printed. A
// synthetic endpoint stands in for OpenRouter, the user's file pointing
// the base URL at it.
func TestOpenRouterTurn(t *testing.T) {
	const answer = "The worker could not reach its database: synthetic."
	var mu sync.Mutex
	var models, keys []string
	marked := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		var body struct {
			Model string `json:"model"`
		}
		if err := json.Unmarshal(b, &body); err != nil || r.URL.Path != "/api/v1/chat/completions" {
			t.Errorf("request %s: %v", r.URL.Path, err)
		}
		mu.Lock()
		models = append(models, body.Model)
		keys = append(keys, r.Header.Get("Authorization"))
		marked = marked || strings.Contains(string(b), "cache_control")
		mu.Unlock()
		w.Header().Set("Content-Type", "text/event-stream")
		content, _ := json.Marshal(answer)
		fmt.Fprintf(w, "data: {\"choices\":[{\"index\":0,\"delta\":{\"role\":\"assistant\",\"content\":%s}}]}\n\n", content)
		fmt.Fprint(w, "data: {\"choices\":[{\"index\":0,\"delta\":{},\"finish_reason\":\"stop\"}],\"usage\":{\"prompt_tokens\":10,\"completion_tokens\":5}}\n\n")
		fmt.Fprint(w, "data: [DONE]\n\n")
	}))
	t.Cleanup(srv.Close)
	t.Setenv("OPENROUTER_API_KEY", "synthetic-key")

	c := newCompa(t)
	user := filepath.Join(t.TempDir(), "config.yaml")
	config := "apiVersion: compa.lowmag.cc/config/v1alpha1\nkind: Config\nreasoning:\n" +
		"  maxEgressClass: external_provider\n" +
		"  openai: {baseURL: " + srv.URL + "/api/v1}\n" +
		"  profiles: {companion: {model: haiku}, ledger: {model: haiku}, editor: {model: haiku}, reviewer: {model: haiku}}\n"
	if err := os.WriteFile(user, []byte(config), 0o600); err != nil {
		t.Fatal(err)
	}
	out, errs, code := c.run(nil, "--config", user, "ask", "demo", "what failed")
	if code != 0 || !strings.Contains(out, answer) {
		t.Fatalf("exit %d, the answer:\n%s\n%s", code, out, errs)
	}
	mu.Lock()
	defer mu.Unlock()
	if len(models) == 0 {
		t.Fatal("no call reached the endpoint")
	}
	for i := range models {
		if models[i] != "anthropic/claude-haiku-5.5" || keys[i] != "Bearer synthetic-key" {
			t.Errorf("call %d: model %q, Authorization %q", i, models[i], keys[i])
		}
	}
	if !marked {
		t.Error("no call carried the engine's cache marks")
	}
}
