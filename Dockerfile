FROM golang:1.23.4-alpine3.21 AS builder
ENV APP_NAME=ponder
ENV WORKDIR=/app
WORKDIR $WORKDIR
COPY . .
RUN go mod download
RUN go build -o /$APP_NAME

## Deploy
FROM alpine:3.21.0
ENV APP_NAME=ponder
WORKDIR /
COPY --from=builder /$APP_NAME /$APP_NAME
ENTRYPOINT ["/ponder"]
