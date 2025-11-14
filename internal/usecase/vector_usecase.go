package usecase

import (
	"context"
	"fmt"

	"github.com/sawalreverr/cv-reviewer/internal/domain"
	"github.com/sawalreverr/cv-reviewer/internal/service"
)

type VectorUsecase interface {
	IngestDocument(ctx context.Context, filePath string, docType domain.DocumentType, metadata map[string]interface{}) error
	SearchSimilar(ctx context.Context, query string, docType domain.DocumentType, topK int) ([]*domain.VectorDocument, error)
	DeleteDocumentsByType(ctx context.Context, docType domain.DocumentType) error
	GetDocumentCount(ctx context.Context, docType domain.DocumentType) (int64, error)
}

type vectorUsecase struct {
	repo domain.VectorRepository
	pdfService service.PDFService 
	chunkingService service.ChunkingService
	embeddingService service.EmbeddingService 
}

func NewVectorUsecase(repo domain.VectorRepository, pdf service.PDFService, chunk service.ChunkingService, embed service.EmbeddingService) VectorUsecase {
	return &vectorUsecase{repo, pdf, chunk, embed}
}

func (uc *vectorUsecase) IngestDocument(ctx context.Context, filePath string, docType domain.DocumentType, metadata map[string]interface{}) error {

	// extract text from pdf
	text, err := uc.pdfService.ExtractText(filePath)
	if err != nil {
		return fmt.Errorf("failed to extract text: %w", err)
	}	

	if text == "" {
		return fmt.Errorf("no text extracted from document")
	}

	// chunk text
	chunks := uc.chunkingService.ChunkBySentence(text, 1000)
	if len(chunks) == 0 {
		return fmt.Errorf("no chunks created from text")
	}

	// generate embeddings
	chunkContents := make([]string, len(chunks))
	for i, chunk := range chunks {
		chunkContents[i] = chunk.Content
	}

	embeddings, err := uc.embeddingService.GenerateBatchEmbeddings(ctx, chunkContents)
	if err != nil {
		return fmt.Errorf("failed to generate batch embeddings: %w", err)
	}

	if len(embeddings) != len(chunks) {
		return fmt.Errorf("embeddings count mismatch: got %d, expected %d", len(embeddings), len(chunks))
	}

	// store chunks
	successCount := 0
	for i, chunk := range chunks {

		// merge metadata
		chunkMetadata := make(map[string]interface{})
		for k, v := range metadata {
			chunkMetadata[k] = v
		}
		chunkMetadata["chunk_index"] = chunk.Index
		chunkMetadata["chunk_length"] = len(chunk.Content)

		// create vector document with pregenerated embedding
		vectorDoc := domain.NewVectorDocument(docType, chunk.Content, embeddings[i], chunkMetadata)

		// store in db
		if err := uc.repo.Create(ctx, vectorDoc); err != nil {
			continue
		}

		successCount++
	}

	if successCount == 0 {
		return fmt.Errorf("failed to store any chunks")
	}

	return nil
}

func (uc *vectorUsecase) SearchSimilar(ctx context.Context, query string, docType domain.DocumentType, topK int) ([]*domain.VectorDocument, error) {
	// generate embedding for query
	queryEmbedding, err := uc.embeddingService.GenerateEmbedding(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// search similar
	docs, err := uc.repo.SearchSimilar(ctx, queryEmbedding, docType, topK)
	if err != nil {
		return nil, fmt.Errorf("failed to search similar documents: %w", err)
	}

	return docs, nil
}

func (uc *vectorUsecase) DeleteDocumentsByType(ctx context.Context, docType domain.DocumentType) error {
	return uc.repo.DeleteByDocType(ctx, docType)
}

func (uc *vectorUsecase) GetDocumentCount(ctx context.Context, docType domain.DocumentType) (int64, error) {
	return uc.repo.Count(ctx, docType)
}