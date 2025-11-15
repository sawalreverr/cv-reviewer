.PHONY: run migrate ingest

run:
	go run cmd/api/main.go

migrate:
	go run scripts/migration/migrate.go $(flag)

ingest:
	go run scripts/ingestion/ingest_docs.go