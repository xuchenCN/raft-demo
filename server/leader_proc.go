package server

import (
	"context"
	log "github.com/sirupsen/logrus"
	pb "github.com/xuchenCN/raft-demo/protocol"
	"time"
)


/**

Leaders:
• Upon election: send initial empty AppendEntries RPCs (heartbeat) to each server; repeat during idle periods to prevent election timeouts (§5.2)
• If command received from client: append entry to local log, respond after entry applied to state machine (§5.3)
• If last log index ≥ nextIndex for a follower: send AppendEntries RPC with log entries starting at nextIndex
• If successful: update nextIndex and matchIndex for
follower (§5.3)
• If AppendEntries fails because of log inconsistency:
decrement nextIndex and retry (§5.3)
• If there exists an N such that N > commitIndex, a majority
of matchIndex[i] ≥ N, and log[N].term == currentTerm: set commitIndex = N (§5.3, §5.4).

 */

var doLeaderProc = true
var heartbeatContext context.Context
var cancelHeartbeatFunc context.CancelFunc

type peerHeartbeatResult struct {
	peer *peer
	result *pb.AppendEntriesResult
}

func (s *server) startLeaderProc() {
	doLeaderProc = true
	heartbeatContext,cancelHeartbeatFunc = context.WithCancel(context.Background())
	s.startHeartbeatToMembers()
}

func (s *server) stopLeaderProc() {
	doLeaderProc = false;
	if cancelHeartbeatFunc != nil {
		cancelHeartbeatFunc()
	}
}

func (s *server) startHeartbeatToMembers() {

	for _,p := range s.peers {
		log.Infof("Start heartbart to %s",p.address)
		go s.doHeartbeat(heartbeatContext,p)
	}
}

func (s *server) doHeartbeat(ctx context.Context, peer *peer) {

	for doLeaderProc {

		select {
		case <- time.After(500 * time.Millisecond): {

			request, lastIndex := s.buildAppendEntriesParam(peer)

			peer.client.AppendEntries(ctx,request)
			log.Infof("Sent heartbeat to %s" , peer.address)
			peer.nextIndex, peer.matchIndex = lastIndex,lastIndex
		}
		case <- ctx.Done():
			break;
		}
	}

}

func (s *server) buildAppendEntriesParam(peer *peer) (*pb.AppendEntriesParam,int32 ){
	prevLog := s.GetPrevLog(peer.nextIndex)
	lastLog := s.GetLastLog()
	entries := make([]*pb.LogEntry,0)

	if lastLog.Index >= peer.nextIndex {
		entries = s.logs[peer.nextIndex:]
	}

	request := pb.AppendEntriesParam{
		Term:s.getCurrentTerm(),
		LeaderId:s.id,
		PrevLogIndex:prevLog.Index,
		PrevLogTerm:prevLog.Term,
		Entries:entries,
		LeaderCommit:s.commitIndex,
	}

	return &request,lastLog.Index
}


func (s *server) logRampUpRunner() {


	for range time.Tick(1 * time.Millisecond) {

		minReplicated,minReplicatedMember := s.GetLastLog().Index,0
		majority := len(s.peers) / 2


		for _,peer := range s.peers {

			if peer.matchIndex <= minReplicated &&
				s.logs[peer.matchIndex].Term == s.currentTerm &&
				peer.matchIndex > s.commitIndex {

				if minReplicated == peer.matchIndex {
					minReplicatedMember += 1
				} else {
					minReplicated = peer.matchIndex
					minReplicatedMember = 1
				}
			}

		}

		//Majority members replicated this index
		if majority >= minReplicatedMember {
			s.commitIndex = int32(minReplicatedMember)
		}
	}
}

