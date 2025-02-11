.PHONY: run
S3C_DIR="./cmd/s3c/"

run: ## Run s3c locally
	go run -ldflags="-X 'main.VERSION=x.x.dev'" $(S3C_DIR)
