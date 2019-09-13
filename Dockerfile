# build image
FROM golang:1.12.7-alpine3.10 as builder

RUN apk update && apk upgrade && \
    apk add --no-cache bash git openssh

RUN go get -v github.com/gorilla/mux && \
    go get -v github.com/joho/godotenv

COPY . /app/
WORKDIR /app

RUN go build -v api.go

# runtime image
FROM alpine:latest
COPY --from=builder /app/api /app/
COPY .env /app/
WORKDIR /app/
CMD ["./api"]
