FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY ["go.mod", "go.sum", "./"]
RUN go mod download

COPY . .

RUN go build -o /app/server ./cmd/server/main.go

COPY config/config.yaml /config.yaml

EXPOSE 18000 10800
CMD ["/app/server"]
