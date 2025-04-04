include config/config.yaml

SERVER_BINARY_NAME ?= server
CLIENT_BINARY_NAME ?= client

# Сборка бинарника
server_build:
	go build -o $(SERVER_BINARY_NAME) ./cmd/server/main.go

# Запуск сервера
server_run:
	./$(SERVER_BINARY_NAME)


# Сборка бинарника
client_build:
	go build -o $(CLIENT_BINARY_NAME) ./cmd/client/main.go

# Запуск клиента
client_run:
	./$(CLIENT_BINARY_NAME)

# Очистка
clean:
	rm -rf $(SERVER_BINARY_NAME) $(CLIENT_BINARY_NAME)

.PHONY: server_build server_run client_build client_run clean help

help:
	@echo Makefile commands:
	@echo server_build - build server
	@echo server_run - run server
	@echo client_build - build client
	@echo client_run - run client
	@echo clean - clean project