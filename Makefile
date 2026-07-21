fmt:
	go fmt ./...

test:
	go test ./...

race:
	go test -race ./...

lint:
	golangci-lint run ./...

fix:
	golangci-lint run ./... --fix

bench:
	go test -run=^$$ -bench=. -benchmem ./...

check: race lint

deps:
	go mod tidy
