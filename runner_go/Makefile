# build go exec
runner-build:
	@go build -o ./bin/runner ./cmd/main.go

# run build
runner-run-build: runner-build
	@./bin/runner

# Run the application
runner-run:
	@go run ./cmd/main.go