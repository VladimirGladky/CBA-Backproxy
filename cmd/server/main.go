package main

import (
	"CBA-Backproxy/internal/server"
)

func main() {
	server := server.NewServer()
	server.Run()
	server.Process()
	server.RunSocks5()
}
