FROM golang:1.15.6 AS builder

WORKDIR /go/src/github.com/alibaba/open-object
COPY . .
RUN make build && chmod +x bin/open-object

FROM alpine:3.9

RUN apk update && apk upgrade && apk add util-linux coreutils file

COPY --from=builder /go/src/github.com/alibaba/open-object/bin/open-object /bin/open-object
COPY ./build/run-connector.sh /run-connector.sh
COPY ./build/fuse-connector.service /fuse-connector.service
COPY ./build/fuse-connector.conf /fuse-connector.conf 

ENTRYPOINT ["/bin/open-object"]
