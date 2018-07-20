package main

import (
	"github.com/xuchenCN/raft-demo/server"
	"github.com/xuchenCN/raft-demo/util"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	s1 := server.NewServer("127.0.0.1",7002)

	stopCh := make(chan int)
	sigCh := make(chan os.Signal)

	signal.Notify(
		sigCh,
		syscall.SIGUSR1,
	)

	go func() {

		for range sigCh {
			util.DumpStacks()
		}

	}()

	go func() {
		s1.WithServers("127.0.0.1:7000","127.0.0.1:7001").Start()
	}()

	<-stopCh
}
