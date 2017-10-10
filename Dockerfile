FROM astronomerio/docker-golang-alpine:gcc
MAINTAINER Ken Herner <ken@astronomer.io>


WORKDIR /go/src/github.com/astronomerio/event-router
COPY . .
RUN apk --no-cache add ca-certificates openssl && \
  wget https://github.com/edenhill/librdkafka/archive/v0.11.0.tar.gz && \
  tar -xf v0.11.0.tar.gz && \
  cd librdkafka-0.11.0 && \
  ./configure && \
  make && make install

RUN make build

FROM alpine
MAINTAINER Ken Herner <ken@astronomer.io>

COPY --from=0 /go/src/github.com/astronomerio/event-router/event-router /usr/local/bin/event-router

ENV GIN_MODE=release
EXPOSE 8081

ENTRYPOINT ["event-router"]