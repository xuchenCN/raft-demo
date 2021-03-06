package server

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	pb "github.com/xuchenCN/raft-demo/protocol"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"math"
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
	logs []*pb.LogEntry

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


/**
Receiver implementation:
1. Reply false if term < currentTerm (§5.1)
2. Reply false if log doesn’t contain an entry at prevLogIndex
whose term matches prevLogTerm (§5.3)
3. If an existing entry conflicts with a new one (same index
but different terms), delete the existing entry and all that
follow it (§5.3)
4. Append any new entries not already in the log
5. If leaderCommit > commitIndex, set commitIndex =
min(leaderCommit, index of last new entry)
 */

func (s *server) AppendEntries(ctx context.Context, req *pb.AppendEntriesParam) (*pb.AppendEntriesResult, error) {
	fmt.Println("Reveive AppendEntries " + req.String())


	resp := pb.AppendEntriesResult{Term:s.getCurrentTerm(),Success:true}

	if s.role == pb.ServerRole_CANDIDATE {
		s.becomeFollower()
	}

	if s.getCurrentTerm() < req.Term {
		s.setCurrentTerm(req.Term)
	}

	if (s.currentTerm > req.Term) {
		resp.Success = false
	}

	if req.Entries != nil && len(req.Entries) > 0 {
		//TODO receive entries
		hasIndexLog := s.GetLogEntryByIndex(req.Entries[0].Index)
		if hasIndexLog.Term != req.Entries[0].Term {
			s.DeleteLogAndFollows(req.Entries[0].Index)
		}
		s.logs = append(s.logs,req.Entries...)

		if req.LeaderCommit > s.commitIndex {
			//TODO Fix min func 
			//https://mrekucci.blogspot.com/2015/07/dont-abuse-mathmax-mathmin.html
			s.commitIndex = int32(math.Min(float64(s.GetLastLog().Index),float64(req.LeaderCommit)))
		}

	}

	s.lastHeartbeat = time.Now()

	return &resp,nil
}

func (s *server) RequestVote(ctx context.Context,  req *pb.RequestVoteParam) (*pb.RequestVoteResult, error) {
	fmt.Println("Receive RequestVote " + req.String())

	resp := pb.RequestVoteResult{}

	lastLog := s.GetLastLog()

	if s.role == pb.ServerRole_LEADER {
		resp.Term = s.getCurrentTerm()
		resp.VoteGranted = false;
		return &resp, nil
	}

	if s.voteFor != s.id && s.voteFor != req.CandidateId {
		resp.Term = s.getCurrentTerm()
		resp.VoteGranted = false;
		return &resp, nil
	}

	if lastLog.Term <= req.Term && lastLog.Index <= req.LastLogIndex {
		resp.Term = s.getCurrentTerm()
		resp.VoteGranted = true;
		//Stop vote request
		//FIXME If concurrent from other candidate?
		if s.role == pb.ServerRole_CANDIDATE {
			s.stopCandidate()
			s.voteFor = req.CandidateId
		}

	} else {
		resp.Term = s.getCurrentTerm()
		resp.VoteGranted = false;
	}

	log.Infof("Vote response to %s %v",req.CandidateId,resp.VoteGranted)

	return &resp,nil
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

func (s *server) setCurrentTerm (i int32) {
	atomic.StoreInt32(&s.currentTerm,i)
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


func (s *server) becomeCandidate() {
	s.stopLeaderProc()
	s.stopFollowerProc()
	s.role = pb.ServerRole_CANDIDATE
	s.voteFor = s.id
	log.Info(s.id + " become a " + s.role.String())
	s.startCandidateProc()
}

func (s *server) becomeLeader() {
	s.stopLeaderProc()
	s.stopFollowerProc()
	s.stopCandidate()
	s.role = pb.ServerRole_LEADER
	log.Info(s.id + " become a " + s.role.String())
	s.startLeaderProc()
}

func (s *server) Stop() {
	s.stopLeaderProc()
	s.stopCandidate()
	s.stopFollowerProc()
	s.grpcSrv.Stop()
}

func (s *server) GetID() string {
	return s.id
}

func (s *server) GetLastLog() pb.LogEntry {
	lastLog := pb.LogEntry{Term: 0, Index: 0}

	s.RLock()
	defer s.RUnlock()

	if len(s.logs) > 1 {
		lastLog = *s.logs[len(s.logs)-1]
	}
	return lastLog
}

func (s *server) GetPrevLog(index int32) pb.LogEntry {
	//Fix index
	index = index - 1

	s.RLock()
	defer s.RUnlock()

	prevLog := pb.LogEntry{Term: 0, Index: 0}
	if index >= 0 && len(s.logs) > int(index) {
		prevLog = *s.logs[index]
	}
	return prevLog
}

func (s *server) GetLogEntryByIndex(index int32) *pb.LogEntry {
	h := len(s.logs)
	for i := h ; i >= 0 ; i -- {
		logEntry := s.logs[i]
		if logEntry.Index == index {
			return logEntry
		}
	}
	return nil
}

func (s *server) DeleteLogAndFollows(index int32) {
	h := len(s.logs)
	for i := h ; i >= 0 ; i -- {
		logEntry := s.logs[i]
		if logEntry.Index == index {
			s.logs = s.logs[:i]
		}
	}
}
