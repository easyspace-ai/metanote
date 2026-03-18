package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"youmind-backend-v2/internal/config"
)

// W6Client is a thin Go wrapper around the IECube W6 HTTP APIs used in the
// existing Java service (genChat / usePageMaker / getJsonRes / computeTokenUsed).
type W6Client struct {
	baseURL    string
	authField  string
	authValue  string
	procedure  string
	llm        string
	llmShort   string
	moduleName string

	httpClient *http.Client
}

// NewW6Client constructs a client from loaded configuration. If the required
// fields (base URL / auth) are empty, it returns nil to signal that W6 is
// disabled.
func NewW6Client(cfg config.W6Config) *W6Client {
	if cfg.BaseURL == "" {
		return nil
	}
	return &W6Client{
		baseURL:    strings.TrimRight(cfg.BaseURL, "/"),
		authField:  cfg.AuthHeaderKey,
		authValue:  cfg.AuthHeaderVal,
		procedure:  cfg.ModelProcedure,
		llm:        cfg.ModelLLM,
		llmShort:   cfg.ModelLLMShort,
		moduleName: cfg.ModuleName,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// StartChat mirrors W6ApiService.genChat(): it creates a new chat session and
// returns the chat_id used for subsequent WebSocket connections and agent
// calls.
func (c *W6Client) StartChat(ctx context.Context) (string, error) {
	if c == nil {
		return "", fmt.Errorf("w6 client is not configured")
	}

	u, err := url.Parse(c.baseURL + "/interact/chat")
	if err != nil {
		return "", fmt.Errorf("build url: %w", err)
	}

	// procedure is required; llm / llm_short are set when procedure != default
	q := u.Query()
	if c.procedure != "" {
		q.Set("procedure", c.procedure)
	}
	if c.procedure != "default" {
		if c.llm != "" {
			q.Set("llm", c.llm)
		}
		if c.llmShort != "" {
			q.Set("llm_short", c.llmShort)
		}
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), nil)
	if err != nil {
		return "", fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	if c.authField != "" && c.authValue != "" {
		req.Header.Set(c.authField, c.authValue)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("call w6 chat: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var body bytes.Buffer
		_, _ = body.ReadFrom(resp.Body)
		return "", fmt.Errorf("w6 chat status %d: %s", resp.StatusCode, body.String())
	}

	var parsed struct {
		ChatID string `json:"chat_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return "", fmt.Errorf("decode w6 chat response: %w", err)
	}
	if parsed.ChatID == "" {
		return "", fmt.Errorf("w6 chat response missing chat_id")
	}
	return parsed.ChatID, nil
}

// CallAgent mirrors W6ApiService.usePageMaker but is generic over the agent
// name and payload. Caller is responsible for constructing the payload fields
// expected by the specific agent (e.g. pagemaker).
func (c *W6Client) CallAgent(
	ctx context.Context,
	chatID string,
	agentName string,
	payload map[string]any,
	llmOverride string,
) error {
	if c == nil {
		return fmt.Errorf("w6 client is not configured")
	}

	u := c.baseURL + "/interact/agent"

	body := map[string]any{
		"payload":        payload,
		"agent_name":     agentName,
		"chat_id":        chatID,
		"llm_model_override": llmOverride,
		"module_name":    c.moduleName,
		"module_source":  nil,
	}

	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal agent body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	if c.authField != "" && c.authValue != "" {
		req.Header.Set(c.authField, c.authValue)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("call w6 agent: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var body bytes.Buffer
		_, _ = body.ReadFrom(resp.Body)
		return fmt.Errorf("w6 agent status %d: %s", resp.StatusCode, body.String())
	}
	return nil
}

// Artefact represents the minimal data we care about from /interact/artefact.
type Artefact struct {
	ID      string          `json:"id"`
	Type    string          `json:"type"`
	Content json.RawMessage `json:"content"`
	Raw     json.RawMessage `json:"-"`
}

// GetArtefact mirrors W6ApiService.getJsonRes: it loads the artefact JSON so
// callers can extract the generated HTML / outline content.
func (c *W6Client) GetArtefact(ctx context.Context, artefactID string) (*Artefact, error) {
	if c == nil {
		return nil, fmt.Errorf("w6 client is not configured")
	}

	u, err := url.Parse(fmt.Sprintf("%s/interact/artefact/%s", c.baseURL, artefactID))
	if err != nil {
		return nil, fmt.Errorf("build url: %w", err)
	}
	q := u.Query()
	q.Set("include_content", "true")
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	if c.authField != "" && c.authValue != "" {
		req.Header.Set(c.authField, c.authValue)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call w6 artefact: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var body bytes.Buffer
		_, _ = body.ReadFrom(resp.Body)
		return nil, fmt.Errorf("w6 artefact status %d: %s", resp.StatusCode, body.String())
	}

	var raw json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decode artefact: %w", err)
	}

	var parsed struct {
		ID      string          `json:"id"`
		Type    string          `json:"type"`
		Content json.RawMessage `json:"content"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		// Fall back to returning raw if shape changed.
		return &Artefact{ID: artefactID, Raw: raw}, nil
	}

	return &Artefact{
		ID:      parsed.ID,
		Type:    parsed.Type,
		Content: parsed.Content,
		Raw:     raw,
	}, nil
}

