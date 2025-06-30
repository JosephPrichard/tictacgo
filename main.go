package main

import (
	"TicTacGo/server"
	"flag"
)

func main() {
	flag.Parse()
	server.Exec()
}
