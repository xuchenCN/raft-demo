package server

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	pb "github.com/xuchenCN/raft-demo/protocol"
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

var (
	doCandidate = true
	voteContext context.Context
	voteCancel context.CancelFunc
)

func (s *server) startCandidateProc() {
	s.incrementTerm(1)
	voteContext, voteCancel = context.WithCancel(context.Background())
	doCandidate = true
	s.doVoteUntilTimeout()
}

func (s *server) doVoteUntilTimeout() {

	i := 0
	for doCandidate {
		select {
		case <-time.Tick(s.timeout):
			{
				i++
				log.Infof("Start vote %d", i)
				lastLog := pb.LogEntry{Term: 0, Index: 0}
				if len(s.logs) > 0 {
					lastLog = s.logs[len(s.logs)-1]
				}
				majority := (int32)(len(s.peers) / 2)
				req := pb.RequestVoteParam{
					Term:         s.currentTerm,
					CandidateId:  s.id,
					LastLogIndex: lastLog.Index,
					LastLogTerm:  lastLog.Term}

				voteResultCh := make(chan bool, 1)
				voteResponseCh := make(chan *pb.RequestVoteResult, len(s.peers))

				c, cancel := context.WithCancel(context.Background())

				s.launchVote(c, majority, &req, voteResponseCh, voteResultCh)

				select {
				case <-time.After(s.timeout):
					//TODO continue
					log.Warn("timeout")
					cancel()
				case <-voteResultCh:
					log.Infof("Vote completed")
				}
				close(voteResponseCh)

				accepted, timeout, lastTerm := checkVoteResult(voteResponseCh)
				log.Infof("accepted %d timeout %d lastTerm %d", accepted, timeout, lastTerm)

				//Become a leader
				if accepted >= majority {
					s.becomeLeader()
				}

				log.Infof("======= BLOCK %d =========", i)
			}
		case <- voteContext.Done():
			log.Infof("Vote canceled")
			break;
		}
	}
}

func (s *server) stopCandidate() {
	log.Infof("Stop candidate %s" , s.id)
	doCandidate = false
	if voteCancel != nil {
		voteCancel()
	}
}

func checkVoteResult(voteResponseCh chan *pb.RequestVoteResult) (accepted, timeout, lastTerm int32) {
	for resp := range voteResponseCh {
		if resp.VoteGranted {
			accepted += 1
		} else if resp.Term < 0 {
			timeout += 1
		} else if lastTerm < resp.Term {
			lastTerm = resp.Term
		}
	}
	return accepted, timeout, lastTerm
}

func (s *server) launchVote(ctx context.Context, majority int32, req *pb.RequestVoteParam, voteResponseCh chan<- *pb.RequestVoteResult, voteResultCh chan<- bool) {

	var answered int32 = 0

	doVoteFunc := func(p *peer) {

		defer func() {
			if err := recover(); err != nil {
				fmt.Println(err)
			}
		}()

		result, err := p.getRpcClient().RequestVote(ctx, req)
		if err != nil {
			result = &pb.RequestVoteResult{Term: -1, VoteGranted: false}
		}

		voteResponseCh <- result
		atomic.AddInt32(&answered, 1)
	}

	for _, peer := range s.peers {
		go doVoteFunc(peer)
	}

	for atomic.LoadInt32(&answered) < majority {

	}
	log.Infof("Vote to %d member wait for %d response", len(s.peers), majority)
	close(voteResultCh)

}
