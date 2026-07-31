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

fmt-check:
	@unformatted=$$(gofmt -l .); \
	if [ -n "$$unformatted" ]; then echo "These files need gofmt:"; echo "$$unformatted"; exit 1; fi

# What CI runs. Keep the two in step: if this passes locally, CI should pass.
check: race lint fmt-check

deps:
	go mod tidy
