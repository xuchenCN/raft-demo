package server

import (
	"context"
	"fmt"
	pb "github.com/xuchenCN/raft-demo/protocol"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type server struct {
	id string
	grpcSrv *grpc.Server
	//Persistent state on all servers
	currentTerm int32
	voteFor string
	logs []pb.LogEntry

	//Volatile state on all servers
	commitIndex int32
	lastApplied int32

	//Heartbeat relevant
	timeout time.Duration
	lastHeartbeat time.Time

	//Current state
	role pb.ServerRole

	peers []*peer

	sync.RWMutex
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

func (s *server) WithServers(peers... string) *server {
	for _, address := range peers {
		p := peer{}
		p.address = address
		s.peers = append(s.peers,&p)
	}
	return s
}

func (s *server) Start() {

	if(len(s.peers) < 2) {
		log.Fatal("At least 2 peers needed")
	}

	stopCh := make(chan int)

	//TODO load logs to State Machine

	lis , err := net.Listen("tcp",s.id)

	if err != nil {
		log.Fatal(err)
	}

	// Connect peers
	for _ , peer := range s.peers {
		peer.connect()
	}

	s.grpcSrv = grpc.NewServer()

	pb.RegisterServerServiceServer(s.grpcSrv, s)

	// Register reflection service on gRPC server.
	reflection.Register(s.grpcSrv)

	go func() {
		if err := s.grpcSrv.Serve(lis); err != nil {
			log.Warn("failed to serve: %v", err)
		}
		close(stopCh)
	}()

	log.Info("Start server using ID " + s.id)



	s.startCommonProc()
	//First start become follower
	s.becomeFollower()

	<- stopCh
}

func (s *server) incrementTerm(i int32) {
	atomic.AddInt32(&s.currentTerm,i)
}

func (s *server) setCurrentTermTo (i int32) {
	atomic.CompareAndSwapInt32(&s.currentTerm,atomic.LoadInt32(&s.currentTerm),i)
}

func (s * server) getCurrentTerm() int32 {
	return atomic.LoadInt32(&s.currentTerm)
}


func (s *server) becomeFollower() {
	s.stopLeaderProc()
	s.stopCandidate()
	s.startFollowerProc()
	s.role = pb.ServerRole_FOLLOWER
	log.Info(s.id + " become a " + s.role.String())
}

func (s *server) GetID() string {
	return s.id
}

func (s *server) becomeCandidate() {
	s.stopLeaderProc()
	s.stopFollowerProc()
	s.startCandidateProc()
	s.role = pb.ServerRole_CANDIDATE
	log.Info(s.id + " become a " + s.role.String())
}

func (s *server) Stop() {
	s.stopLeaderProc()
	s.stopCandidate()
	s.stopFollowerProc()
	s.grpcSrv.Stop()
}
