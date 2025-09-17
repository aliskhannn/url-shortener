FROM golang:alpine

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go mod tidy
RUN go mod download
RUN go build -o url-shortener ./cmd/url-shortener/main.go

EXPOSE 8080