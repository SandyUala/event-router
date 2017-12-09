FROM astronomerio/alpine-librdkafka-golang
MAINTAINER Ken Herner <ken@astronomer.io>

WORKDIR /go/src/github.com/astronomerio/event-router
COPY . .
RUN make buildit

FROM astronomerio/alpine-librdkafka:3.6-0.11.0-r0
COPY --from=0 /go/src/github.com/astronomerio/event-router/event-router /usr/local/bin/event-router
COPY ./wait-for-it.sh /usr/local/bin/wait-for-it
ENV GIN_MODE=release
EXPOSE 9091
ENTRYPOINT ["event-router"]
