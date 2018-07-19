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

func startCommonProc(s *server) {
	go doApplyLogRunner(s)
}

func doApplyLogRunner(s *server) {
	for {
		if s.commitIndex > s.lastApplied {
			//TODO Apply to State Machine
		}
	}
}


func Random(min, max int64) int64 {
	return rand.Int63n(max - min) + min
}

