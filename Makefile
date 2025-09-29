format:
	go fmt .

test:
	go test ./...

lint:
	golangci-lint run ./...

deps:
	go mod tidy