# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOVENDOR=$(GOCMD) mod vendor
BINARY_FOLDER=dist
BINARY_NAME=do-dyndns
ENTRYPOINT=main

all: test build

install:
	$(GOVENDOR)

test:
	$(GOTEST) -v ./...

clean:
	$(GOCLEAN)
	rm -f $(BINARY_FOLDER)/$(BINARY_NAME)*

# Cross compilation

build: build-linux-amd64 build-linux-arm build-linux-arm64

build-linux-amd64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_FOLDER)/$(BINARY_NAME)-linux-x64

build-linux-arm7:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 $(GOBUILD) -o $(BINARY_FOLDER)/$(BINARY_NAME)-linux-arm

build-linux-arm64:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GOBUILD) -o $(BINARY_FOLDER)/$(BINARY_NAME)-linux-arm64

.PHONY: build
