# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOLANGCI_LINT=golangci-lint
BINARY_NAME=ipam

all: lint test build

build:
	$(GOBUILD) -o $(BINARY_NAME) -v

lint:
	$(GOLANGCI_LINT) run ./...

test:
	$(GOTEST) -race -coverprofile=coverage.out -covermode=atomic ./...

clean:
	rm -f $(BINARY_NAME) coverage.out

.PHONY: all build lint test clean