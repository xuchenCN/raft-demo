package server

import (
	"context"
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
	s.timeout = time.Duration(Random(50,200)) * time.Millisecond
}

func (s *server) doVoteUntilTimeout() {
	for doCandidate {

		//TODO launch a vote

		result := make(chan bool)

		c , cancel := context.WithTimeout(s.ctx,s.timeout)

		select {
		case <- c.Done():
			//TODO continue
		case <- result:
			//TODO Become leader
			cancel()
		}
	}
}

func (s *server) stopCandidate() {
	doCandidate = false
}