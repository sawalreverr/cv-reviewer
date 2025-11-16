# CV Reviewer

A backend service that automates the initial screening of job applications using AI evaluation. The system evaluates candidate CV and project reports against job descriptions and case study briefs, producing structured evaluation reports through LLM driven analysis.

## Overview

This service provides automated candidate evaluation by:

-   Accepting candidate CV and project report submissions
-   Evaluating documents against job requirements and scoring rubrics
-   Using Retrieval Augmented Generation (RAG) for context evaluation
-   Generating structured feedback with match scores and recommendations
-   Processing evaluations asynchronously with a job queue system

## Architecture

The backend follows clean architecture principles with clear separation of concerns:

```
internal/
├── domain/         # entities and repository interfaces
├── repository/     # database access impl
├── usecase/        # business logic
├── service/        # external service integrations (LLM, PDF, embeddings)
└── handler/        # HTTP request handlers
```

## TechStacks

-   [Echo](https://github.com/labstack/echo) (Web Framework Go)
-   [PostgreSQL](https://hub.docker.com/r/pgvector/pgvector) (DB with pgvector extension)
-   [GORM](https://gorm.io/docs/) (ORM)
-   [Gemini API](https://ai.google.dev/gemini-api/docs) (LLM Provider)
-   [PDF](https://github.com/ledongthuc/pdf) (PDF Processing)
-   [Viper](https://github.com/spf13/viper) (Configuration)
-   [Docker Compose](https://docs.docker.com/compose/) (Containerization)

## Prerequisites

-   Go 1.25 or higher
-   PostgreSQL 16 with pgvector extension
-   Google Gemini API key
-   Make (optional)

## Installation

Clone the repository:

```bash
git clone https://github.com/sawalreverr/cv-reviewer.git
cd cv-reviewer
```

Install dependencies:

```bash
go mod download
```

Set up PostgreSQL with pgvector:

```bash
docker-compose up -d
```

Configure environment variables:

```bash
cp .env.example .env
```

Run database migrations:

```bash
make migrate

# drop all tables
make migrate flag=-drop

# or run manually (if doesnt have make)
go run scripts/migration/ migrate.go # -drop, if needed
```

Prepare system documents in `/docs/` directory:

-   `job_description.pdf` - job descriptions
-   `case_study_brief.pdf` - case study brief document
-   `cv_scoring_rubric.pdf` - cv evaluation criteria
-   `project_scoring_rubric.pdf` - project evaluation criteria

Ingest system documents into vector database:

```bash
make ingest

# or run manually (if doesnt have make)
go run scripts/ingestion/ingest_docs.go
```

## Running the app

Start the server:

```bash
make run

# or run manually (if doesnt have make)
go run cmd/api/main.go
```

The API will be available at `http://localhost:8080`

## API Endpoints

### Health Check

Ensuring the API is running

```
GET /health
```

Response:

```json
{
    "success": true,
    "data": {
        "status": "ok",
        "message": "api is running"
    }
}
```

### Upload Documents

```
POST /upload
Content-Type: multipart/form-data
```

Form fields:

-   `cv` (file): candidate CV in PDF
-   `project_report` (file): candidate project report in PDF

Response:

```json
{
    "success": true,
    "message": "documents uploaded successfully",
    "data": {
        "cv_id": "uuid",
        "project_report_id": "uuid"
    }
}
```

### Create Evaluation Job

```
POST /evaluate
Content-Type: application/json
```

Request body:

```json
{
    "job_title": "Backend Developer",
    "cv_id": "uuid",
    "project_report_id": "uuid"
}
```

Response:

```json
{
    "success": true,
    "message": "evaluation job created",
    "data": {
        "id": "uuid",
        "status": "queued"
    }
}
```

### Get Evaluation Result

```
GET /result/{job_id}
```

Response (while processing):

```json
{
    "success": true,
    "data": {
        "id": "uuid",
        "status": "processing"
    }
}
```

Response (completed):

```json
{
    "success": true,
    "data": {
        "id": "uuid",
        "status": "completed",
        "result": {
            "cv_match_rate": 0.82,
            "cv_feedback": "Strong in backend and cloud, limited AI integration experience...",
            "project_score": 4.5,
            "project_feedback": "Meets prompt chaining requirements, lacks error handling robustness...",
            "overall_summary": "Good candidate fit, would benefit from deeper RAG knowledge..."
        }
    }
}
```

## Evaluation Pipeline

The evaluation process consists of three main stages:

### CV Evaluation

-   Extracts text from candidate CV
-   Retrieves relevant job description context using vector similarity search
-   Retrieves CV scoring rubric criteria
-   Uses LLM to generate match rate (0-1 scale) and feedback

### Project Report Evaluation

-   Extracts text from project report
-   Retrieves relevant case study brief context
-   Retrieves project scoring rubric criteria
-   Uses LLM to generate score (1-5 scale) and feedback

### Final Summary

-   Synthesizes CV and project evaluations
-   Generates holistic candidate assessment
-   Provides hiring recommendation

## Testing

Example workflow:

#### Upload test documents:

```bash
curl -X POST http://localhost:8080/upload \
  -F "cv=@/path/to/cv.pdf" \
  -F "project_report=@/path/to/report.pdf"
```

#### Create evaluation job:

```bash
curl -X POST http://localhost:8080/evaluate \
  -H "Content-Type: application/json" \
  -d '{
    "job_title": "Backend Developer",
    "cv_id": "cv-uuid",
    "project_report_id": "report-uuid"
  }'
```

#### Check evaluation status:

```bash
curl http://localhost:8080/result/{job_id}
```

## Error Handling Examples

#### Invalid File Type

```json
{
    "success": false,
    "message": "invalid file type, only pdf allowed",
    "error": "invalid file type, only PDF allowed"
}
```

#### Document Not Found

```json
{
    "success": false,
    "message": "cv or project report document not found",
    "error": "resource not found"
}
```

#### Queue Full

```json
{
    "success": false,
    "message": "job queue is full, please try again later",
    "error": "job queue is full"
}
```
