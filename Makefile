# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_FOLDER=dist
BINARY_NAME=do-dyndns
ENTRYPOINT=main

all: test build

test:
				$(GOTEST) -v ./...

clean:
				$(GOCLEAN)
				rm -f $(BINARY_FOLDER)/$(BINARY_NAME)

# Cross compilation
build:
		CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_FOLDER)/$(BINARY_NAME)-linux-x64
