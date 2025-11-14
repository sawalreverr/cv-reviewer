package service

import (
	"context"
	"fmt"

	"github.com/sawalreverr/cv-reviewer/config"
	"google.golang.org/genai"
)

type EmbeddingService interface {
	GenerateEmbedding(ctx context.Context, text string) ([]float32, error)
	GenerateBatchEmbeddings(ctx context.Context, texts []string) ([][]float32, error)
}

type embeddingService struct {
	client *genai.Client
	model string
	dimension *int32
}

func NewEmbeddingService(cfg *config.GeminiConfig) (EmbeddingService, error) {
	ctx := context.Background()

	client, err := genai.NewClient(ctx, &genai.ClientConfig{APIKey: cfg.APIKey, Backend: genai.BackendGeminiAPI})
	if err != nil {
		return nil, fmt.Errorf("failed to create gemini client: %w", err)
	}

	return &embeddingService{client, cfg.EmbeddingModel, cfg.Dimension}, nil
}

func (es *embeddingService) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	if text == "" {
		return nil, fmt.Errorf("text cannot be empty")
	}

	res, err := es.client.Models.EmbedContent(ctx, es.model, genai.Text(text), &genai.EmbedContentConfig{OutputDimensionality: es.dimension})
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding: %w", err)
	}

	if len(res.Embeddings) == 0 {
		return nil, fmt.Errorf("no embeddings received")
	}

	if res.Embeddings == nil || len(res.Embeddings[0].Values) == 0 {
		return nil, fmt.Errorf("empty embedding values received")
	}

	return res.Embeddings[0].Values, nil
}

func (es *embeddingService) GenerateBatchEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, fmt.Errorf("texts cannot be empty")
	}

	contents := make([]*genai.Content, len(texts))
	for i, text := range texts {
		contents[i] = genai.NewContentFromText(text, genai.RoleUser)
	}

	res, err := es.client.Models.EmbedContent(ctx, es.model, contents, &genai.EmbedContentConfig{OutputDimensionality: es.dimension})
	if err != nil {
		return nil, fmt.Errorf("failed to generate batch embeddings: %w", err)
	}

	if len(res.Embeddings) == 0 {
		return nil, fmt.Errorf("no embeddings received")
	}

	if len(res.Embeddings) != len(texts) {
		return nil, fmt.Errorf("expected %d embeddings but got %d", len(texts), len(res.Embeddings))
	}

	// extract all embedding values
	embeddings := make([][]float32, len(res.Embeddings))
	for i, emb := range res.Embeddings {
		if emb == nil || len(emb.Values) == 0 {
			return nil, fmt.Errorf("empty embbeding at index %d", i)
		}
		embeddings[i] = emb.Values
	}

	return embeddings, nil
}