package server

import (
	pb "github.com/xuchenCN/raft-demo/protocol"
)

func (s *server) receiveClientRequest(date []byte) {
	lastLog := s.GetLastLog()

	logEntry := pb.LogEntry{Term:s.currentTerm,Index:lastLog.Index + 1,Data:date}

	s.Lock()
	s.logs = append(s.logs,&logEntry)
	s.commitIndex += 1
	s.Unlock()
}