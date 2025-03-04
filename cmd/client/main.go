package main

import "CBA-Backproxy/internal/client"

func main() {
	cl := client.New()
	cl.Connect()
}
