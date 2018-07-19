package server

import (
	"context"
	"fmt"
	pb "github.com/xuchenCN/raft-demo/protocol"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
	"time"
)

type server struct {
	id string
	grpcSrv *grpc.Server
	//Persistent state on all servers
	currentTerm string
	voteFor string
	//TODO log[]

	//Volatile state on all servers
	commitIndex int
	lastApplied int

	//TODO nextIndex matchIndex

	//Heartbeat relevant
	timeout time.Duration
	lastHeartbeat time.Time

	//Context
	ctx context.Context

	peers []string
}

func (s *server) AppendEntries(context.Context, *pb.AppendEntriesParam) (*pb.AppendEntriesResult, error) {
	fmt.Println("implement me")
	return nil,nil
}

func (s *server) RequestVote(context.Context, *pb.RequestVoteParam) (*pb.RequestVoteResult, error) {
	fmt.Println("implement me")
	return nil,nil
}


func NewServer(host string, port int) *server {
	s := server{}
	s.id = fmt.Sprintf("%s:%d",host,port)
	return &s
}

func (s *server) WithServers(peers... string) {
	s.peers = peers
}

func (s *server) Start() {

	if(len(s.peers) < 2) {
		log.Fatal("At least 2 peers needed")
	}

	lis , err := net.Listen("tcp",s.id)

	if err != nil {
		log.Fatal(err)
	}

	log.Info("Start server using ID " + s.id)

	s.grpcSrv = grpc.NewServer()

	pb.RegisterServerServiceServer(s.grpcSrv, s)

	// Register reflection service on gRPC server.
	reflection.Register(s.grpcSrv)
	if err := s.grpcSrv.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
