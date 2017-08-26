package lbbnet

import (
	"math/rand"
	"sync"
)

type (
	Session struct {
		id  uint32
		tsp *Transport
	}

	SessionManager struct {
		sync.RWMutex
		mp map[uint64]*Session
	}
)

func NewSessionManager() *SessionManager {
	s := &SessionManager{mp: make(map[uint64]*Session, 100000)}
	return s
}

func (s *SessionManager) Exist(uid uint64) bool {
	s.RLock()
	_, ok := s.mp[uid]
	s.RUnlock()
	return ok
}

func (s *SessionManager) Get(uid uint64) *Session {
	s.RLock()
	ss := s.mp[uid]
	s.RUnlock()
	return ss
}

func (s *SessionManager) Set(uid uint64, tsp *Transport) bool {
	sid := rand.Uint32()
	s.RLock()
	_, ok := s.mp[uid]
	s.RUnlock()
	if !ok {
		s.Lock()
		s.mp[uid] = &Session{sid, tsp}
		s.Unlock()
		return true
	}
	return false
}
func (s *SessionManager) Del(uid uint64) {
	s.RLock()
	_, ok := s.mp[uid]
	s.RUnlock()
	if ok {
		s.Lock()
		delete(s.mp, uid)
		s.Unlock()
	}
}
