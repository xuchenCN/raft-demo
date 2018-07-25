package server

import (
	"context"
	"fmt"
	pb "github.com/xuchenCN/raft-demo/protocol"
	log "github.com/sirupsen/logrus"
	"sync/atomic"
	"time"
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

func (s *server) startCandidateProc() {
	doCandidate = true;
	s.doVoteUntilTimeout()
}

func (s *server) doVoteUntilTimeout() {
	i := 0
	for range time.Tick(5 * time.Millisecond) {

		if !doCandidate {
			return
		}
		i ++
		log.Infof("Start vote %d" , i)
		lastLog := pb.LogEntry{Term:s.getCurrentTerm(),Index:0}
		if len(s.logs) > 0 {
			lastLog = s.logs[len(s.logs) - 1]
		//lastLog := s.logs[len(s.logs) - 1]

		req := pb.RequestVoteParam{
			Term:s.currentTerm,
			CandidateId:s.id,
			LastLogIndex:lastLog.Index,
			LastLogTerm:lastLog.Term}

		voteResultCh := make(chan bool,1)
		voteResponseCh := make(chan *pb.RequestVoteResult,len(s.peers))

		c , cancel := context.WithCancel(context.Background())

		s.launchVote(c,&req,voteResponseCh,voteResultCh)

		select {
		case <- time.After(s.timeout):
			//TODO continue
			log.Warn("timeout")
			cancel()
		case <- voteResultCh:
			log.Infof("Vote completed")
		}
		close(voteResponseCh)

		for resp := range voteResponseCh {
			fmt.Println(resp)
		}

		log.Infof("======= BLOCK %d =========" , i)
	}
}

func (s *server) stopCandidate() {
	doCandidate = false
}

func (s *server) launchVote(ctx context.Context , req *pb.RequestVoteParam , voteResponseCh chan<-*pb.RequestVoteResult, voteResultCh chan<-bool) {

	var answered int32 = 1

	doVoteFunc := func(p *peer) {

		defer func() {
			if err := recover(); err != nil {
				fmt.Println(err)
			}
		}()

		result ,err := p.getRpcClient().RequestVote(ctx, req)
		if err != nil {
			result = &pb.RequestVoteResult{Term:-1,VoteGranted:false}
		}

		voteResponseCh <- result
		atomic.AddInt32(&answered,1)
	}

	for _,peer := range s.peers {
		go doVoteFunc(peer)
	}

	majority := (int32)(len(s.peers) / 2)
	for atomic.LoadInt32(&answered) > majority {

	}

	close(voteResultCh)

}