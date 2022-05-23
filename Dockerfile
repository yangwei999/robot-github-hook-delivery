FROM golang:latest as BUILDER

MAINTAINER zengchen1024<chenzeng765@gmail.com>

# build binary
WORKDIR /go/src/github.com/opensourceways/robot-github-hook-delivery
COPY . .
RUN GO111MODULE=on CGO_ENABLED=0 go build -a -o robot-github-hook-delivery .

# copy binary config and utils
FROM alpine:3.14
COPY  --from=BUILDER /go/src/github.com/opensourceways/robot-github-hook-delivery/robot-github-hook-delivery /opt/app/robot-github-hook-delivery

ENTRYPOINT ["/opt/app/robot-github-hook-delivery"]
