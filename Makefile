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

fuzz:
	go test -run=^$$ -fuzz=FuzzSignalCoW -fuzztime=30s ./signal/
	go test -run=^$$ -fuzz=FuzzGroupOps -fuzztime=30s ./signal/

check: race lint

deps:
	go mod tidy
