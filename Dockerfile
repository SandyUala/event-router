FROM golang:alpine

ARG LIBRDKAFKA_VERSION="0.11.1-r1"

ENV REPO="github.com/astronomerio/event-router"
WORKDIR /go/src/${REPO}

RUN apk add --no-cache \
		build-base \
		cyrus-sasl-dev \
		git \
		librdkafka-dev=${LIBRDKAFKA_VERSION} \
		libressl \
		openssl-dev \
		yajl-dev \
		zlib-dev

COPY . .
RUN make build

ENV GIN_MODE=release
EXPOSE 9091

# Use ENTRYPOINT in production images
CMD ["./event-router"]
