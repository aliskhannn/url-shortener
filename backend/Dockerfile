FROM golang:alpine

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o notifier ./cmd/url-shortener/main.go

EXPOSE 8080