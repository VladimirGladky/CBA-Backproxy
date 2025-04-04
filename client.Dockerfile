FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY ["go.mod", "go.sum", "./"]
RUN go mod download

COPY . .

RUN go build -o /app/client ./cmd/client/main.go

CMD ["/app/client"]
