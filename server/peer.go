package server

import (
	"context"
	"google.golang.org/grpc"
	pb "github.com/xuchenCN/raft-demo/protocol"
	log "github.com/sirupsen/logrus"
	"time"
)

/**

Volatile state on leaders:
(Reinitialized after election)
nextIndex[]
matchIndex[]
for each server, index of the next log entry to send to that server (initialized to leader
last log index + 1)
for each server, index of highest log entry known to be replicated on server (initialized to 0, increases monotonically)

 */
type peer struct {
	address string
	nextIndex,matchIndex int32
	ctx context.Context
	client pb.ServerServiceClient
	shouldRun bool
}

func (p *peer) connect() {

	conn,err := grpc.Dial(p.address,grpc.WithInsecure())

	if err != nil {
		log.Error(err)
		log.Warn("Retry to connect client " + p.address)
		for _ = range time.Tick(2 * time.Second) {

			if !p.shouldRun {
				log.Info("Peer stop!")
				return
			}

			conn,err = grpc.Dial(p.address,grpc.WithInsecure())
			if err == nil {
				break
			}
		}
	}

	p.client = pb.NewServerServiceClient(conn)
	log.Info("Client " + p.address + " connected")
	p.ctx = context.Background()
	p.nextIndex = 1
}

func (p *peer) getRpcClient() pb.ServerServiceClient {
	return p.client
}



