IMAGE_NAME=astronomerio/cs-event-router

GIT_COMMIT=$(shell git rev-parse --short HEAD)
GOTAGS ?= event-router
GOFILES ?= $(shell go list ./... | grep -v /vendor/)
GOOS=$(shell go env GOOS)
GOARCH=$(shell go env GOARCH)

VERSION ?= SNAPSHOT-$(GIT_COMMIT)

all: build

dep:
	dep ensure -v
	dep prune -v

buildit:
	go build -tags static -o event-router main.go

# build: staticcheck gosimple buildit
build: buildit

# install: staticcheck gosimple
# 	go install -tags static

install: build
	mkdir -p $(DESTDIR)
	cp event-router $(DESTDIR)

uninstall:
	rm -rf $(DESTDIR)

run:
	go run cmd/main.go

build-image:
	docker build -t $(IMAGE_NAME):$(VERSION) .

tag-latest:
	docker tag $(IMAGE_NAME):$(VERSION) $(IMAGE_NAME):latest

push-image: image
	docker push $(IMAGE_NAME):$(VERSION)

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

test:
	go test $(shell go list ./... | grep -v /vendor/)

style:
	@echo ">> checking code style"
	@! gofmt -d $(shell find . -path ./vendor -prune -o -name '*.go' -print) | grep '^'

staticcheck:
	@echo ">> running staticcheck"
	@staticcheck $(GOFILES)

gosimple:
	@echo ">> running gosimple"
	@gosimple $(GOFILES)

tools:
	@echo ">> installing some extra tools"
	@go get -u -v honnef.co/go/tools/...

clean:
	-rm event-router
	-docker rmi `docker images | grep "astronomerio/cs-event-router" | awk '{print $3}'`

.PHONY : clean
