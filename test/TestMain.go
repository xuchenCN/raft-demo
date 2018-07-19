package main

import "github.com/xuchenCN/raft-demo/server"

func main() {

	s := server.NewServer("0.0.0.0",7788)
	s.Start()
}
