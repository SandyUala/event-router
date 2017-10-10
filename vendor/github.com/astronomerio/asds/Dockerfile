FROM astronomerio/docker-golang-alpine
MAINTAINER Ken Herner <ken@astronomer.io>

WORKDIR /go/src/github.com/astronomerio/asds
COPY . .

RUN make build

FROM alpine
MAINTAINER Ken Herner <ken@astronomer.io>

COPY --from=0 /go/src/github.com/astronomerio/asds/ASDS /usr/local/bin/ASDS

ENV GIN_MODE=release
EXPOSE 8081

ENTRYPOINT ["ASDS"]