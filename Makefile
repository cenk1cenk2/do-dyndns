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

update:
	$(GOGET) -u all
	$(GOVENDOR)
	$(GOCMD) mod tidy -compat=1.17

test:
	$(GOTEST) -v ./...

clean:
	$(GOCLEAN)
	rm -f $(BINARY_FOLDER)/$(BINARY_NAME)*

# Cross compilation

build: build-linux-amd64 build-linux-armv7 build-linux-arm64

build-linux-amd64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -mod=readonly -o $(BINARY_FOLDER)/$(BINARY_NAME)-linux-amd64

build-linux-armv7:
	GOOS=linux GOARCH=arm GOARM=7 $(GOBUILD) -mod=readonly -a -installsuffix cgo -ldflags "$$LD_FLAGS" -o $(BINARY_FOLDER)/$(BINARY_NAME)-linux-armv7

build-linux-arm64:
	GOOS=linux GOARCH=arm64 $(GOBUILD) -mod=readonly -a -installsuffix cgo -ldflags "$$LD_FLAGS" -o $(BINARY_FOLDER)/$(BINARY_NAME)-linux-arm64

.PHONY: build
