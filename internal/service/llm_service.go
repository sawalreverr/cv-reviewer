package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sawalreverr/cv-reviewer/config"
	"google.golang.org/genai"
)

type CVEvaluation struct {
	CVMatchRate float64 `json:"cv_match_rate"`
	CVFeedback string `json:"cv_feedback"`
}

type ProjectEvaluation struct {
	ProjectScore float64 `json:"project_score"`
	ProjectFeedback string `json:"project_feedback"`
}

type FinalSummary struct {
	OveralSummary string `json:"overall_summary"`
}

type LLMService interface {
	EvaluateCV(ctx context.Context, cvText string, jobDescContext, rubricContext []string) (*CVEvaluation, error)
	EvaluateProject(ctx context.Context, projectText string, caseStudyContext, rubricContext []string) (*ProjectEvaluation, error)
	FinalSummary(ctx context.Context, cvEval *CVEvaluation, projectEval *ProjectEvaluation) (string, error)
}

type llmService struct {
	client *genai.Client
	model string
	temperature float32
	maxTokens int32
}

func NewLLMService(cfg *config.GeminiConfig) (LLMService, error) {
	ctx := context.Background()

	client, err := genai.NewClient(ctx, &genai.ClientConfig{APIKey: cfg.APIKey, Backend: genai.BackendGeminiAPI})
	if err != nil {
		return nil, fmt.Errorf("failed to create gemini client: %w", err)
	}

	return &llmService{client, cfg.Model, cfg.Temperature, cfg.MaxTokens}, nil
}

func (s *llmService) EvaluateCV(ctx context.Context, cvText string, jobDescContext, rubricContext []string) (*CVEvaluation, error) {
	prompt := s.CVEvaluationPrompt(cvText, jobDescContext, rubricContext)

	response, err := s.generateContent(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate cv evaluation: %w", err)
	}

	// parse json response
	var eval CVEvaluation
	cleanedResponse := s.cleanJSONResponse(response)
	if err := json.Unmarshal([]byte(cleanedResponse), &eval); err != nil {
		return nil, fmt.Errorf("failed to parse cv evaluation: %w (response: %s)", err, cleanedResponse)
	}

	// validate cv_match_rate is between 0 and 1
	if eval.CVMatchRate < 0 || eval.CVMatchRate > 1 {
		return nil, fmt.Errorf("invalid cv_match_rate: %f (must be between 0 and 1)", eval.CVMatchRate)
	}

	return &eval, nil
}

func (s *llmService) EvaluateProject(ctx context.Context, projectText string, caseStudyContext, rubricContext []string) (*ProjectEvaluation, error) {
	prompt := s.ProjectEvaluationPrompt(projectText, caseStudyContext, rubricContext)

	response, err := s.generateContent(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate project evaluation: %w", err)
	}

	// parse json response
	var eval ProjectEvaluation
	cleanedResponse := s.cleanJSONResponse(response)
	if err := json.Unmarshal([]byte(cleanedResponse), &eval); err != nil {
		return nil, fmt.Errorf("failed to parse project evaluation: %w (response: %s)", err, cleanedResponse)
	}

	// validate project_score is between 1 and 5
	if eval.ProjectScore < 1 || eval.ProjectScore > 5 {
		return nil, fmt.Errorf("invalid project_score: %f (must be between 1 and 5)", eval.ProjectScore)
	}

	return &eval, nil
}

func (s *llmService) FinalSummary(ctx context.Context, cvEval *CVEvaluation, projectEval *ProjectEvaluation) (string, error) {
	prompt := s.FinalSummaryPrompt(cvEval, projectEval)

	response, err := s.generateContent(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to generate final summary: %w", err)
	} 

	var summary FinalSummary
	cleanedResponse := s.cleanJSONResponse(response)
	if err := json.Unmarshal([]byte(cleanedResponse), &summary); err != nil {
		return "", fmt.Errorf("failed to parse finaly summary: %w (response: %s)", err, cleanedResponse)
	}

	return strings.TrimSpace(summary.OveralSummary), nil
}

func (s *llmService) generateContent(ctx context.Context, prompt string) (string, error) {
	response, err := s.client.Models.GenerateContent(ctx, s.model, genai.Text(prompt), &genai.GenerateContentConfig{Temperature: &s.temperature, MaxOutputTokens: s.maxTokens})
	if err != nil {
		return "", err
	}

	if len(response.Candidates) == 0 {
		return "", fmt.Errorf("no candidates in response")
	}

	candidate := response.Candidates[0]
	if candidate.Content == nil || len(candidate.Content.Parts) == 0 {
		return "", fmt.Errorf("empty content in response")
	}

	return response.Text(), nil
}

func (s *llmService) cleanJSONResponse(response string) string {
	response = strings.ReplaceAll(response, "```json", "")
	response = strings.ReplaceAll(response, "```", "")
	return strings.TrimSpace(response)
}