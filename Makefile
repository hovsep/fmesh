test:
	go test ./...

lint:
	golangci-lint run ./...

deps:
	go mod tidy