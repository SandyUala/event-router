IMAGE_NAME=astronomerio/asds:latest

GOTAGS ?= go-provisioner
GOFILES ?= $(shell go list ./... | grep -v /vendor/)
GOOS=$(shell go env GOOS)
GOARCH=$(shell go env GOARCH)

# Get the git commit
GIT_COMMIT=$(shell git rev-parse --short HEAD)
GIT_DIRTY=$(shell test -n "`git status --porcelain`" && echo "+CHANGES" || true)
GIT_DESCRIBE=$(shell git describe --tags --always)
GOLDFLAGS=-X $(GIT_IMPORT).GitCommit=$(GIT_COMMIT)$(GIT_DIRTY)


build:
	go build -o ASDS -ldflags '$(GOLDFLAGS)' -tags '$(GOTAGS)' cmd/main.go

run:
	go run cmd/main.go

image:
	docker build -t $(IMAGE_NAME) .

push: image
	docker push $(IMAGE_NAME)
