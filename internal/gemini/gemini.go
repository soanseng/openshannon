// Package gemini implements a Gemini API executor for text and image generation.
package gemini

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/scipio/claude-channels/internal/claude"
)

const baseURL = "https://generativelanguage.googleapis.com/v1beta"

// Executor calls the Gemini REST API directly.
type Executor struct {
	apiKey string
	model  string
	client *http.Client
}

// NewExecutor creates a Gemini executor with the given API key and model.
func NewExecutor(apiKey, model string) *Executor {
	if model == "" {
		model = "gemini-2.5-flash"
	}
	return &Executor{
		apiKey: apiKey,
		model:  model,
		client: &http.Client{Timeout: 5 * time.Minute},
	}
}

// request/response types for Gemini API
type generateRequest struct {
	Contents         []content        `json:"contents"`
	GenerationConfig *generationConfig `json:"generationConfig,omitempty"`
}

type content struct {
	Parts []part `json:"parts"`
}

type part struct {
	Text       string      `json:"text,omitempty"`
	InlineData *inlineData `json:"inlineData,omitempty"`
}

type inlineData struct {
	MimeType string `json:"mimeType"`
	Data     string `json:"data"`
}

type generationConfig struct {
	ResponseModalities []string `json:"responseModalities,omitempty"`
}

type generateResponse struct {
	Candidates []candidate `json:"candidates"`
	Error      *apiError   `json:"error,omitempty"`
}

type candidate struct {
	Content content `json:"content"`
}

type apiError struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// Run calls the Gemini API and returns the complete response.
func (e *Executor) Run(ctx context.Context, _, _, _, prompt string, _ claude.RunOpts) (*claude.Result, error) {
	text, _, err := e.call(ctx, prompt, false)
	if err != nil {
		return nil, err
	}
	return &claude.Result{Text: text}, nil
}

// RunWithStream calls the Gemini API then delivers the response via callback.
func (e *Executor) RunWithStream(ctx context.Context, _, _, _, prompt string, _ claude.RunOpts, cb claude.StreamCallback) (*claude.Result, error) {
	text, _, err := e.call(ctx, prompt, false)
	if err != nil {
		return nil, err
	}
	if cb != nil {
		cb(text)
	}
	return &claude.Result{Text: text}, nil
}

// Cancel is a no-op for the HTTP-based Gemini executor.
func (e *Executor) Cancel(_ string) error {
	return nil
}

// GenerateImage uses Gemini's native image generation (Imagen via Gemini 2.0 Flash).
// Returns the text response and path to the generated image file (if any).
func (e *Executor) GenerateImage(ctx context.Context, prompt, outputDir string) (text string, imagePath string, err error) {
	return e.call(ctx, prompt, true)
}

func (e *Executor) call(ctx context.Context, prompt string, wantImage bool) (text string, imagePath string, err error) {
	req := generateRequest{
		Contents: []content{
			{Parts: []part{{Text: prompt}}},
		},
	}

	// For image generation, use the best available image model
	model := e.model
	if wantImage {
		model = "gemini-3.1-flash-image-preview"
		req.GenerationConfig = &generationConfig{
			ResponseModalities: []string{"TEXT", "IMAGE"},
		}
	}

	data, err := json.Marshal(req)
	if err != nil {
		return "", "", fmt.Errorf("marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", baseURL, model, e.apiKey)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return "", "", fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := e.client.Do(httpReq)
	if err != nil {
		return "", "", fmt.Errorf("gemini API call failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		slog.Error("gemini API error", "status", resp.StatusCode, "body", string(body))
		return "", "", fmt.Errorf("gemini API returned %d: %s", resp.StatusCode, string(body))
	}

	var genResp generateResponse
	if err := json.Unmarshal(body, &genResp); err != nil {
		return "", "", fmt.Errorf("parse response: %w", err)
	}

	if genResp.Error != nil {
		return "", "", fmt.Errorf("gemini error: %s", genResp.Error.Message)
	}

	if len(genResp.Candidates) == 0 {
		return "", "", fmt.Errorf("gemini returned no candidates")
	}

	// Extract text and images from response
	var textResult string
	for _, p := range genResp.Candidates[0].Content.Parts {
		if p.Text != "" {
			textResult += p.Text
		}
		if p.InlineData != nil && p.InlineData.MimeType != "" {
			// Save image to temp file
			imgData, decErr := base64.StdEncoding.DecodeString(p.InlineData.Data)
			if decErr != nil {
				slog.Warn("failed to decode image data", "err", decErr)
				continue
			}
			ext := ".png"
			if p.InlineData.MimeType == "image/jpeg" {
				ext = ".jpg"
			}
			tmpDir := filepath.Join(os.TempDir(), "claude-channels")
			_ = os.MkdirAll(tmpDir, 0o700)
			tmpFile := filepath.Join(tmpDir, fmt.Sprintf("gemini-%d%s", time.Now().UnixNano(), ext))
			if writeErr := os.WriteFile(tmpFile, imgData, 0o600); writeErr != nil {
				slog.Warn("failed to save generated image", "err", writeErr)
				continue
			}
			imagePath = tmpFile
		}
	}

	if textResult == "" && imagePath == "" {
		return "", "", fmt.Errorf("gemini returned empty response")
	}

	return textResult, imagePath, nil
}
