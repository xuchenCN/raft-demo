package server

import "time"
import log "github.com/sirupsen/logrus"

/**
	Followers (§5.2):
	• Respond to RPCs from candidates and leaders
	• If election timeout elapses without receiving AppendEntries
		RPC from current leader or granting vote to candidate: convert to candidate
 */

var doFollower = true

func (s *server) startFollowerProc() {
	s.timeout = time.Duration(Random(50,200)) * time.Millisecond
	log.Info(s.id + " start with timeout " + s.timeout.String())
	doFollower = true
	go s.pingChecker()
}

func (s *server) pingChecker() {
	for doFollower {
		//Check ping duration
		if time.Now().Sub(s.lastHeartbeat) > s.timeout {
			log.Info(s.id + " want become a candidate")
			s.becomeCandidate()
		}
	}
}

func (s *server) stopFollowerProc() {
	doFollower = false;
}


