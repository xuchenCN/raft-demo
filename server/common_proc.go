package server

import (
	"math/rand"
	"time"
)

/**
	All Servers:
	• If commitIndex > lastApplied: increment lastApplied, apply log[lastApplied] to state machine (§5.3)
	• If RPC request or response contains term T > currentTerm: set currentTerm = T, convert to follower (§5.1)
 */

func init() {
	rand.Seed(time.Now().Unix())
}

func (s *server) startCommonProc() {
	go s.doApplyLogRunner()
	s.lastHeartbeat = time.Now()
}

func (s *server) doApplyLogRunner() {
	for range time.Tick(1 * time.Millisecond) {
		if s.commitIndex > s.lastApplied {
			s.applyLog(&s.logs[s.lastApplied:s.commitIndex])
			s.lastApplied = s.commitIndex
		}
	}
}


func Random(min, max int64) int64 {
	return rand.Int63n(max - min) + min
}

