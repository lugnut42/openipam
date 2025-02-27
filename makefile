# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOLANGCI_LINT=golangci-lint
BINARY_NAME=ipam
VALIDATOR_NAME=validate-blocks

all: lint test build

build: main validator
	
main:
	$(GOBUILD) -o $(BINARY_NAME) -v

validator:
	$(GOBUILD) -o $(VALIDATOR_NAME) scripts/validate_blocks.go

lint:
	$(GOLANGCI_LINT) run ./...

test:
	$(GOTEST) -race -coverprofile=coverage.out -covermode=atomic ./...

clean:
	rm -f $(BINARY_NAME) $(VALIDATOR_NAME) coverage.out

.PHONY: all build main validator lint test clean