# build go exec
build:
	@go build -o ./bin/replc ./cmd/main.go

# run build
runb: build
	@./bin/replc

# Run the application
run:
	@go run cmd/main.go

# Run all tests
test:
	@go test -count=1 -p 1 ./...