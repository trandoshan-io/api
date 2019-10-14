# build image
FROM golang:1.13.1-alpine3.10 as builder

RUN apk update && apk upgrade && \
    apk add --no-cache bash git openssh

COPY . /app/
WORKDIR /app

RUN go build -v api.go handler.go service.go

# runtime image
FROM alpine:latest
COPY --from=builder /app/api /app/

WORKDIR /app/
CMD ["./api"]

EXPOSE 8080
