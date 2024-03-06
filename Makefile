# build go exec
.PHONY: build
build:
	@go build -o ./bin/replc ./cmd/main.go

# run build
.PHONY: runb
runb: build
	@./bin/replc

# Run the application
.PHONY: run
run:
	@go run cmd/main.go

# Run all tests
.PHONY: test
test:
	@go test -count=1 -p 1 ./...