IMAGE_NAME ?= astronomerinc/ap-event-router

GIT_COMMIT=$(shell git rev-parse HEAD)
GIT_COMMIT_SHORT=$(shell git rev-parse --short HEAD)
VERSION ?= SNAPSHOT-${GIT_COMMIT_SHORT}

LDFLAGS_VERSION=-X github.com/astronomerio/event-router/cmd.version=${VERSION}
LDFLAGS_GIT_COMMIT=-X github.com/astronomerio/event-router/cmd.gitCommit=${GIT_COMMIT}

# Set default for make.
.DEFAULT_GOAL := build-image

.PHONY: build
build:
	go build -ldflags "${LDFLAGS_VERSION} ${LDFLAGS_GIT_COMMIT}" -tags static -o event-router main.go

.PHONY: install
install: build
	mkdir -p $(DESTDIR)
	cp event-router $(DESTDIR)

.PHONY: uninstall
uninstall:
	rm -rf $(DESTDIR)

.PHONY: build-image
build-image:
	docker build -t $(IMAGE_NAME):latest .
	docker tag $(IMAGE_NAME):latest $(IMAGE_NAME):$(VERSION)

.PHONY: test-image
test-image: build-image
	docker run ${IMAGE_NAME}:${VERSION} go test ./...
