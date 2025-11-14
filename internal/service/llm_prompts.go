package service

import (
	"fmt"
	"strings"
)

func (s *llmService) CVEvaluationPrompt(cvText string, jobDescContext, rubricContext []string) string {
	prompt := fmt.Sprintf(`
You are an expert technical recruiter with 8+ years of experience.

TASK: Evaluate the candidate CV against the provided job description and scoring rubric.

JOB DESCRIPTION:
%s

CV SCORING RUBRIC:
%s

CANDIDATE CV:
%s

EVALUATION CRITERIA (1-5 scale):
1. Technical skills match (weight 40%%)
2. Experience level (weight 25%%)
3. Relevant achievements (weight 20%%)
4. Cultural fit (weight 15%%)

REQUIRED OUTPUT (JSON format):
{
  "cv_match_rate": <0.00-1.00>,
  "cv_feedback": "<3-5 sentences: technical strengths, skill gaps, specific recommendations>"
}

RULES:
- Score each criterion 1-5 according to rubric
- Calculate weighted average, convert to decimal
- Feedback must be specific, objective, actionable
- OUTPUT ONLY JSON, No additional text`, strings.Join(jobDescContext, "\n"), strings.Join(rubricContext, "\n"), cvText)

	return prompt
}

func (s *llmService) ProjectEvaluationPrompt(projectText string, caseStudyContext, rubricContext []string) string {
	prompt := fmt.Sprintf(`
You are a senior backend engineer reviewing a technical case study submission.

TASK: Evaluate the candidate Project Report against the case study brief and scoring rubric.

CASE STUDY BRIEF:
%s

PROJECT SCORING RUBRIC:
%s

PROJECT REPORT:
%s

EVALUATION CRITERIA (1-5 scale):
1. Correctness (weight 30%%):
2. Code Quality (weight 25%%)
3. Resilience (weight 20%%)
4. Documentation (weight 15%%)
5. Creativity (weight 10%%)

REQUIRED OUTPUT (JSON format):
{
  "project_score": <1.0-5.0>,
  "project_feedback": "<3-5 sentences: best aspects, technical gaps, improvement suggestions>"
}

RULES:
- Score each criterion 1-5 according to rubric
- Calculate weighted average for final score
- Feedback must be technical, constructive, evidence-based
- OUTPUT ONLY JSON, No additional text`, strings.Join(caseStudyContext, "\n"), strings.Join(rubricContext, "\n"), projectText)


	return prompt
}

func (s *llmService) FinalSummaryPrompt(cvEval *CVEvaluation, projectEval *ProjectEvaluation) string {
	prompt := fmt.Sprintf(`
You are a senior engineering hiring manager synthesizing candidate evaluation results.

CV EVALUATION RESULTS:
- Match Rate: %.2f (0-1 scale)
- Feedback: %s

PROJECT EVALUATION RESULTS:
- Overall Score: %.1f (1-5 scale)
- Feedback: %s

REQUIRED OUTPUT (JSON format):
{
  "overall_summary": "<3-5 sentences: holistic assessment, key strengths, critical gaps, hiring recommendation>"
}

RULES:
- Synthesize both evaluations into coherent narrative
- Balance technical skills (CV) with practical execution (Project)
- Provide clear hiring recommendation (strong hire/hire/maybe/pass)
- Be honest but professional
- OUTPUT ONLY JSON, No additional text`, cvEval.CVMatchRate, cvEval.CVFeedback, projectEval.ProjectScore, projectEval.ProjectFeedback)

	return prompt
}