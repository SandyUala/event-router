IMAGE_NAME=astronomerio/event-router:latest

GOTAGS ?= event-router
GOFILES ?= $(shell go list ./... | grep -v /vendor/)
GOOS=$(shell go env GOOS)
GOARCH=$(shell go env GOARCH)

all: build

build:
	go build -o event-router main.go

static:
	go build -tags static -o event-router main.go

install:
	go install

run:
	go run cmd/main.go

image:
	docker build -t $(IMAGE_NAME) .

push: image
	docker push $(IMAGE_NAME)

format:
	@echo "--> Running go fmt"
	@go fmt $(GOFILES)

vet:
	@echo "--> Running go vet"
	@go vet $(GOFILES); if [ $$? -eq 1 ]; then \
		echo ""; \
		echo "Vet found suspicious constructs. Please check the reported constructs"; \
		echo "and fix them if necessary before submitting the code for review."; \
		exit 1; \
	fi

clean:
	rm event-router
