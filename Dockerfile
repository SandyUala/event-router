FROM astronomerio/alpine-librdkafka-golang
MAINTAINER Ken Herner <ken@astronomer.io>


WORKDIR /go/src/github.com/astronomerio/event-router
COPY . .

RUN make build

FROM alpine
MAINTAINER Ken Herner <ken@astronomer.io>

COPY --from=0 /go/src/github.com/astronomerio/event-router/event-router /usr/local/bin/event-router

ENV GIN_MODE=release
EXPOSE 8080

ENTRYPOINT ["event-router"]