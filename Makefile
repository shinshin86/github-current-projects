.PHONY: build test test-race lint fmt vet clean cover

BINARY := github-current-projects

build:
	go build -o $(BINARY) ./cmd/github-current-projects

test:
	go test ./... -v

test-race:
	go test -race ./...

cover:
	go test ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out
	go tool cover -html=coverage.out -o coverage.html

fmt:
	gofmt -w .

vet:
	go vet ./...

lint: fmt vet

clean:
	rm -f $(BINARY) coverage.out coverage.html
