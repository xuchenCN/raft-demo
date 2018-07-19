package server

import "time"

/**
	Followers (§5.2):
	• Respond to RPCs from candidates and leaders
	• If election timeout elapses without receiving AppendEntries
		RPC from current leader or granting vote to candidate: convert to candidate
 */

var doFollower = true

func (s *server) startFollowerProc() {

}

func (s *server) pingChecker() {
	for doFollower {
		//Check ping duration
		if time.Now().Sub(s.lastHeartbeat) > time.Second * 5 {
			//TODO become a candidate
		}
	}
}

func (s *server) stopFollowerProc() {
	doFollower = false;
}


