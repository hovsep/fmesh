fmt:
	go fmt ./...

test:
	go test ./...

lint:
	golangci-lint run ./...

fix:
	golangci-lint  run ./... --fix

deps:
	go mod tidy