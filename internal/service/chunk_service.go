package service

import (
	"regexp"
	"strings"
)

type TextChunk struct {
	Content  string
	Index    int
	Metadata map[string]interface{}
}

type ChunkingService interface {
	ChunkText(text string, chunkSize, overlap int) []TextChunk
	ChunkBySentence(text string, maxChunkSize int) []TextChunk
}

type chunkingService struct{}

func NewChunkingService() ChunkingService {
	return &chunkingService{}
}

func (s *chunkingService) ChunkText(text string, chunkSize, overlap int) []TextChunk {
	if chunkSize <= 0 {
		chunkSize = 1000
	}
	if overlap < 0 || overlap >= chunkSize {
		overlap = 200
	}

	var chunks []TextChunk
	runes := []rune(text)
	totalLen := len(runes)

	if totalLen == 0 {
		return chunks
	}

	start := 0
	index := 0

	for start < totalLen {
		end := start + chunkSize
		if end > totalLen {
			end = totalLen
		}

		chunk := string(runes[start:end])
		chunks = append(chunks, TextChunk{
			Content: strings.TrimSpace(chunk),
			Index:   index,
			Metadata: map[string]interface{}{
				"start": start,
				"end": end,
			},
		})

		start += chunkSize - overlap
		index++
	}

	return chunks
}

func (s *chunkingService) ChunkBySentence(text string, maxChunkSize int) []TextChunk {
	if maxChunkSize <= 0 {
		maxChunkSize = 1000
	}

	sentences := s.splitIntoSentences(text)
	if len(sentences) == 0 {
		return []TextChunk{}
	}

	var chunks []TextChunk
	var currentChunk strings.Builder
	var currentChunkSentences []string
	index := 0
	const sentenceOverlap = 1

	for _, sentence := range sentences {
		sentence = strings.TrimSpace(sentence)
		if sentence == "" {
			continue
		}

		// check if adding this sentence exceeds max chunk size
		potentialLength := currentChunk.Len() + len(sentence) + 1
		if currentChunk.Len() > 0 && potentialLength > maxChunkSize {
			chunks = append(chunks, TextChunk{
				Content: strings.TrimSpace(currentChunk.String()),
				Index: index,
			})
			
			currentChunk.Reset()
			index++

			// get last N sentences for overlap
			var overlapContent []string
			if len(currentChunkSentences) > sentenceOverlap {
				overlapContent = currentChunkSentences[len(currentChunkSentences)-sentenceOverlap:]
			} else {
				overlapContent = currentChunkSentences
			}

			// new chunk with overlapped sentences
			if len(overlapContent) > 0 {
				currentChunk.WriteString(strings.Join(overlapContent, " "))
				currentChunkSentences = overlapContent
			} else {
				currentChunkSentences = []string{}
			}
		}

		if currentChunk.Len() > 0 {
			currentChunk.WriteString(" ")
		}
		currentChunk.WriteString(sentence)
		currentChunkSentences = append(currentChunkSentences, sentence)
	}

	if currentChunk.Len() > 0 {
		chunks = append(chunks, TextChunk{
			Content: strings.TrimSpace(currentChunk.String()),
			Index: index,
		})
	}

	return chunks
}

var sentenceRegex = regexp.MustCompile(`[^.!?]+[.!?]?`)

// uses regex to split text into sentences
func (s *chunkingService) splitIntoSentences(text string) []string {
	text = strings.ReplaceAll(text, "\n", " ")

	matches := sentenceRegex.FindAllString(text, -1)
	
    var sentences []string
    for _, s := range matches {
        trimmed := strings.TrimSpace(s)
        if trimmed != "" {
            sentences = append(sentences, trimmed)
        }
    }
	return sentences
}