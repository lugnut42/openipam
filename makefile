# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
BINARY_NAME=ipam

all: build

build:
	$(GOBUILD) -o $(BINARY_NAME) -v

clean:
	rm -f $(BINARY_NAME)

.PHONY: all build clean