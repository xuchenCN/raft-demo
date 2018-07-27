package server


import (
	log "github.com/sirupsen/logrus"
	pb "github.com/xuchenCN/raft-demo/protocol"
)

func (s *server) applyLog(logEntry []*pb.LogEntry) {
	log.Infof("Apply to State Machine %v", logEntry)
}

