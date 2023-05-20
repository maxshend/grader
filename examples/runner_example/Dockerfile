FROM golang:1.19-alpine

RUN apk update && apk upgrade && \
    apk add --no-cache bash git openssh

WORKDIR /app
COPY ./go.mod ./go.sum ./
RUN go mod download && chmod 0777 /app && mkdir -p -m 0777 ./cache
COPY . .
ENV GOCACHE=/app/cache
