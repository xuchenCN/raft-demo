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
	commitIndex int32
	lastApplied int32

	//Heartbeat relevant
	timeout time.Duration
	lastHeartbeat time.Time

	//Context
	ctx context.Context

	//Current state
	role pb.ServerRole

	peers []peer
}



func (s *server) AppendEntries(ctx context.Context, req *pb.AppendEntriesParam) (*pb.AppendEntriesResult, error) {
	fmt.Println("Receive AppendEntries " + req.String())
	return nil,nil
}

func (s *server) RequestVote(ctx context.Context,  req *pb.RequestVoteParam) (*pb.RequestVoteResult, error) {
	fmt.Println("Receive RequestVote " + req.String())
	return nil,nil
}


func NewServer(host string, port int) *server {
	s := server{}
	s.id = fmt.Sprintf("%s:%d",host,port)
	return &s
}

func (s *server) WithServers(peers... string) {
	for _, address := range peers {
		p := peer{}
		p.address = address
		s.peers = append(s.peers,p)
	}
}

func (s *server) Start() {

	if(len(s.peers) < 2) {
		log.Fatal("At least 2 peers needed")
	}

	//TODO load logs to State Machine

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

	s.startCommonProc()

	//First start become follower
	s.becomeFollower()
}


func (s *server) becomeFollower() {
	s.stopLeaderProc()
	s.stopCandidate()
	s.startFollowerProc()
	s.role = pb.ServerRole_FOLLOWER
	log.Info(s.id + " become a " + s.role.String())
}

