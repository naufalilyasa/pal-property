package gemini

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/naufalilyasa/pal-property-backend/pkg/config"
	"google.golang.org/genai"
)

const (
	defaultEmbeddingModel    = "gemini-embedding-001"
	defaultGenerationModel   = "gemini-2.5-flash-lite"
	defaultTimeout           = 20 * time.Second
	defaultRetryAttempts     = 2
	defaultMaxOutputTokens   = 512
	defaultTemperature       = 0.2
	defaultSystemInstruction = "You are PAL Property's Indonesian customer-service assistant. Answer naturally, warmly, and clearly like a helpful real-estate CS agent. Use only the provided property context, never mention internal document processing, never say 'berdasarkan dokumen', and do not expose raw IDs unless the user explicitly asks. If recommendations exist, summarize the best matches conversationally and let the UI handle property cards/links."
)

// EmbeddingTask differentiates between query and document workloads.
type EmbeddingTask string

const (
	EmbeddingTaskQuery    EmbeddingTask = "query"
	EmbeddingTaskDocument EmbeddingTask = "document"
)

// PromptSafetyHook allows callers to vet requests before they reach Gemini.
type PromptSafetyHook func(ctx context.Context, req GroundedAnswerRequest) error

type clientOptions struct {
	embeddingModel  string
	generationModel string
	timeout         time.Duration
	maxRetries      int
	safetyHook      PromptSafetyHook
	httpClient      *http.Client
	models          modelsAPI
}

// Option configures the Gemini client factory.
type Option func(*clientOptions)

// WithEmbeddingModel overrides the model used when computing embeddings.
func WithEmbeddingModel(model string) Option {
	return func(o *clientOptions) {
		if trimmed := strings.TrimSpace(model); trimmed != "" {
			o.embeddingModel = trimmed
		}
	}
}

// WithGenerationModel overrides the model used for grounded answer generation.
func WithGenerationModel(model string) Option {
	return func(o *clientOptions) {
		if trimmed := strings.TrimSpace(model); trimmed != "" {
			o.generationModel = trimmed
		}
	}
}

// WithTimeout customizes the HTTP client timeout used by the GenAI SDK.
func WithTimeout(timeout time.Duration) Option {
	return func(o *clientOptions) {
		if timeout > 0 {
			o.timeout = timeout
		}
	}
}

// WithRetryCount adjusts the number of retry attempts.
func WithRetryCount(count int) Option {
	return func(o *clientOptions) {
		if count >= 0 {
			o.maxRetries = count
		}
	}
}

// WithSafetyHook registers a callback to analyze grounded requests before sending.
func WithSafetyHook(hook PromptSafetyHook) Option {
	return func(o *clientOptions) {
		o.safetyHook = hook
	}
}

// WithHTTPClient allows injecting a pre-configured http.Client.
func WithHTTPClient(client *http.Client) Option {
	return func(o *clientOptions) {
		if client != nil {
			o.httpClient = client
		}
	}
}

// WithModelsAPI injects a fake models implementation (used in tests).
func WithModelsAPI(api modelsAPI) Option {
	return func(o *clientOptions) {
		if api != nil {
			o.models = api
		}
	}
}

func newDefaultOptions() clientOptions {
	return clientOptions{
		embeddingModel:  defaultEmbeddingModel,
		generationModel: defaultGenerationModel,
		timeout:         defaultTimeout,
		maxRetries:      defaultRetryAttempts,
	}
}

// Client wraps the Google GenAI SDK for Gemini-specific embeddings and generation.
type Client struct {
	models modelsAPI
	opts   clientOptions
}

// NewClient instantiates a Gemini client backed by the official genai SDK.
func NewClient(ctx context.Context, apiKey string, opts ...Option) (*Client, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, errors.New("gemini: API key is required")
	}

	options := newDefaultOptions()
	for _, opt := range opts {
		opt(&options)
	}

	httpClient := options.httpClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: options.timeout}
	} else if options.timeout > 0 {
		httpClient.Timeout = options.timeout
	}

	if options.models == nil {
		cfg := &genai.ClientConfig{
			APIKey:     apiKey,
			Backend:    genai.BackendGeminiAPI,
			HTTPClient: httpClient,
		}
		client, err := genai.NewClient(ctx, cfg)
		if err != nil {
			return nil, err
		}
		options.models = &genaiModelsAdapter{models: client.Models}
	}

	options.httpClient = httpClient
	return &Client{models: options.models, opts: options}, nil
}

// NewClientFromConfig wires the Gemini client using the backend contract defaults.
func NewClientFromConfig(ctx context.Context, opts ...Option) (*Client, error) {
	apiKey := strings.TrimSpace(config.Env.ChatGeminiAPIKey)
	if apiKey == "" {
		return nil, errors.New("gemini: CHAT_GEMINI_API_KEY must be configured")
	}

	base := []Option{
		WithEmbeddingModel(config.Env.ChatEmbeddingModel),
		WithGenerationModel(config.Env.ChatGeminiModel),
		WithTimeout(time.Duration(config.Env.ChatGeminiTimeoutSeconds) * time.Second),
	}
	base = append(base, opts...)
	return NewClient(ctx, apiKey, base...)
}

// GroundingDocument represents a source that can be cited in responses.
type GroundingDocument struct {
	ID      string
	Title   string
	Source  string
	Excerpt string
}

// GroundedAnswerRequest contains a user question plus documents for grounding.
type GroundedAnswerRequest struct {
	Question        string
	Documents       []GroundingDocument
	CandidateCount  int32
	MaxOutputTokens int32
	Temperature     *float32
	StopSequences   []string
}

// GroundedAnswerResponse captures the generated answer and raw SDK candidate.
type GroundedAnswerResponse struct {
	Answer        string
	Model         string
	Candidate     *genai.Candidate
	Candidates    []*genai.Candidate
	UsageMetadata *genai.GenerateContentResponseUsageMetadata
}

// EmbeddingResult exposes the vector returned for a specific task.
type EmbeddingResult struct {
	Task   EmbeddingTask
	Values []float64
	Model  string
}

// EmbedQuery returns embeddings tagged as a query workload.
func (c *Client) EmbedQuery(ctx context.Context, inputs ...string) ([]EmbeddingResult, error) {
	return c.embed(ctx, EmbeddingTaskQuery, inputs...)
}

// EmbedDocument returns embeddings tagged as a document workload.
func (c *Client) EmbedDocument(ctx context.Context, inputs ...string) ([]EmbeddingResult, error) {
	return c.embed(ctx, EmbeddingTaskDocument, inputs...)
}

func (c *Client) embed(ctx context.Context, task EmbeddingTask, inputs ...string) ([]EmbeddingResult, error) {
	if len(inputs) == 0 {
		return nil, errors.New("gemini: at least one input is required")
	}
	contents := make([]*genai.Content, 0, len(inputs))
	for _, raw := range inputs {
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" {
			return nil, errors.New("gemini: blank input is not allowed")
		}
		contents = append(contents, genai.NewContentFromText(trimmed, genai.RoleUser))
	}

	var resp *genai.EmbedContentResponse
	embedConfig := buildEmbedConfig(task)
	err := c.withRetry(ctx, func(inner context.Context) error {
		var callErr error
		resp, callErr = c.models.EmbedContent(inner, c.opts.embeddingModel, contents, embedConfig)
		return callErr
	})
	if err != nil {
		return nil, err
	}
	if resp == nil || len(resp.Embeddings) == 0 {
		return nil, errors.New("gemini: no embeddings returned")
	}
	results := make([]EmbeddingResult, 0, len(resp.Embeddings))
	for _, emb := range resp.Embeddings {
		vector := make([]float64, len(emb.Values))
		for i, v := range emb.Values {
			vector[i] = float64(v)
		}
		results = append(results, EmbeddingResult{
			Task:   task,
			Model:  c.opts.embeddingModel,
			Values: vector,
		})
	}
	return results, nil
}

func buildEmbedConfig(task EmbeddingTask) *genai.EmbedContentConfig {
	config := &genai.EmbedContentConfig{}
	outputDimensionality := int32(768)
	config.OutputDimensionality = &outputDimensionality
	switch task {
	case EmbeddingTaskQuery:
		config.TaskType = "RETRIEVAL_QUERY"
	case EmbeddingTaskDocument:
		config.TaskType = "RETRIEVAL_DOCUMENT"
	}
	return config
}

// GenerateGroundedAnswer composes a grounded prompt and returns Gemini's response.
func (c *Client) GenerateGroundedAnswer(ctx context.Context, req GroundedAnswerRequest) (*GroundedAnswerResponse, error) {
	if strings.TrimSpace(req.Question) == "" {
		return nil, errors.New("gemini: question is required")
	}
	if c.opts.safetyHook != nil {
		if err := c.opts.safetyHook(ctx, req); err != nil {
			return nil, err
		}
	}

	contents := buildGroundedContents(req)
	config := buildGenerateConfig(req)
	var resp *genai.GenerateContentResponse
	err := c.withRetry(ctx, func(inner context.Context) error {
		var callErr error
		resp, callErr = c.models.GenerateContent(inner, c.opts.generationModel, contents, config)
		return callErr
	})
	if err != nil {
		return nil, err
	}
	if resp == nil || len(resp.Candidates) == 0 {
		return nil, errors.New("gemini: no generated candidate")
	}
	candidate := resp.Candidates[0]
	answer := flattenContent(candidate.Content)
	return &GroundedAnswerResponse{
		Answer:        strings.TrimSpace(answer),
		Model:         resp.ModelVersion,
		Candidate:     candidate,
		Candidates:    resp.Candidates,
		UsageMetadata: resp.UsageMetadata,
	}, nil
}

func buildGroundedContents(req GroundedAnswerRequest) []*genai.Content {
	contents := make([]*genai.Content, 0, 2)
	summary := buildDocumentSummary(req.Documents)
	if summary != "" {
		contents = append(contents, genai.NewContentFromText(summary, genai.RoleUser))
	}
	question := fmt.Sprintf("Question: %s", strings.TrimSpace(req.Question))
	contents = append(contents, genai.NewContentFromText(question, genai.RoleUser))
	return contents
}

func buildDocumentSummary(docs []GroundingDocument) string {
	if len(docs) == 0 {
		return "Konteks properti: tidak ada data properti yang cocok untuk pertanyaan ini."
	}
	var sb strings.Builder
	sb.WriteString("Ringkasan properti kandidat:\n")
	for idx, doc := range docs {
		if idx > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(fmt.Sprintf("Properti %d", idx+1))
		if doc.Title != "" {
			sb.WriteString(fmt.Sprintf(" - %s", doc.Title))
		}
		sb.WriteString("\n")
		if doc.Source != "" {
			sb.WriteString(fmt.Sprintf("Slug: %s\n", doc.Source))
		}
		if doc.Excerpt != "" {
			sb.WriteString(fmt.Sprintf("Deskripsi singkat: %s\n", doc.Excerpt))
		}
	}
	return strings.TrimSpace(sb.String())
}

func buildGenerateConfig(req GroundedAnswerRequest) *genai.GenerateContentConfig {
	cfg := &genai.GenerateContentConfig{
		CandidateCount:    1,
		MaxOutputTokens:   defaultMaxOutputTokens,
		SystemInstruction: genai.NewContentFromText(defaultSystemInstruction, genai.RoleModel),
	}
	if req.CandidateCount > 0 {
		cfg.CandidateCount = req.CandidateCount
	}
	if req.MaxOutputTokens > 0 {
		cfg.MaxOutputTokens = req.MaxOutputTokens
	}
	if req.Temperature != nil {
		cfg.Temperature = req.Temperature
	} else {
		cfg.Temperature = float32Ptr(defaultTemperature)
	}
	if len(req.StopSequences) > 0 {
		cfg.StopSequences = req.StopSequences
	}
	return cfg
}

func float32Ptr(v float32) *float32 {
	return &v
}

func flattenContent(content *genai.Content) string {
	if content == nil {
		return ""
	}
	var parts []string
	for _, part := range content.Parts {
		if part == nil {
			continue
		}
		if trimmed := strings.TrimSpace(part.Text); trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return strings.Join(parts, "\n")
}

func (c *Client) withRetry(ctx context.Context, call func(context.Context) error) error {
	maxAttempts := c.opts.maxRetries
	if maxAttempts < 0 {
		maxAttempts = 0
	}
	var lastErr error
	for attempt := 0; attempt <= maxAttempts; attempt++ {
		if err := call(ctx); err != nil {
			lastErr = err
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return err
			}
			if attempt == maxAttempts {
				break
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(retryDelay(attempt)):
			}
			continue
		}
		return nil
	}
	return lastErr
}

func retryDelay(attempt int) time.Duration {
	base := time.Duration(attempt+1) * 150 * time.Millisecond
	if base > 2*time.Second {
		return 2 * time.Second
	}
	return base
}

type modelsAPI interface {
	GenerateContent(context.Context, string, []*genai.Content, *genai.GenerateContentConfig) (*genai.GenerateContentResponse, error)
	EmbedContent(context.Context, string, []*genai.Content, *genai.EmbedContentConfig) (*genai.EmbedContentResponse, error)
}

type genaiModelsAdapter struct {
	models *genai.Models
}

func (a *genaiModelsAdapter) GenerateContent(ctx context.Context, model string, contents []*genai.Content, config *genai.GenerateContentConfig) (*genai.GenerateContentResponse, error) {
	return a.models.GenerateContent(ctx, model, contents, config)
}

func (a *genaiModelsAdapter) EmbedContent(ctx context.Context, model string, contents []*genai.Content, config *genai.EmbedContentConfig) (*genai.EmbedContentResponse, error) {
	return a.models.EmbedContent(ctx, model, contents, config)
}
