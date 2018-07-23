package server

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	pb "github.com/xuchenCN/raft-demo/protocol"
	"golang.org/x/net/html/atom"
)

/**

Candidates (§5.2):
• On conversion to candidate, start election:
• Increment currentTerm
• Vote for self
• Reset election timer
• Send RequestVote RPCs to all other servers
• If votes received from majority of servers: become leader
• If AppendEntries RPC received from new leader: convert to
follower
• If election timeout elapses: start new election

*/

var doCandidate = true;
var voteResultCh = make(chan bool,1)


func (s *server) startCandidateProc() {

}

func (s *server) doVoteUntilTimeout() {
	for doCandidate {

		lastLog := s.logs[len(s.logs) - 1]
		req := pb.RequestVoteParam{
			Term:s.currentTerm,
			CandidateId:s.id,
			LastLogIndex:lastLog.Index,
			LastLogTerm:lastLog.Term}

		s.launchVote(&req)

		c , cancel := context.WithTimeout(s.ctx,s.timeout)

		select {
		case <- c.Done():
			//TODO continue
		case true == <- voteResultCh:
			//TODO Become leader
			cancel()
		case false == <- voteResultCh:
			//TODO Rejected
		}
	}
}

func (s *server) stopCandidate() {
	doCandidate = false
}

func (s *server) launchVote(req *pb.RequestVoteParam) {

	answered := 0
	cancelMap := make(map[string]context.CancelFunc)
	voteResponseCh := make(chan *pb.RequestVoteResult,len(s.peers))
	defer close(voteResponseCh)

	doVoteFunc := func(p peer) {
		c,cancel := context.WithTimeout(s.ctx,s.timeout)
		cancelMap[p.address] = cancel

		result ,err := p.getRpcClient().RequestVote(c, req)
		if err != nil {
			result = &pb.RequestVoteResult{Term:-1,VoteGranted:false}
		}

		voteResponseCh <- result
		answered += 1
		delete(cancelMap,p.address)
	}

	for _,peer := range s.peers {
		go doVoteFunc(peer)
	}

	for answered < (len(s.peers) / 2 + 1) {

	}

	for resp := range voteResponseCh {
		fmt.Println(resp)
	}
}