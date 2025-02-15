package serial

import (
	"sync"
)

type SyncMap struct {
	mp   map[string]chan struct{}
	lock sync.Mutex
}

func NewSyncMap() *SyncMap {
	return &SyncMap{
		mp:   make(map[string]chan struct{}),
		lock: sync.Mutex{},
	}
}

func (s *SyncMap) Put(key string) chan struct{} {
	s.lock.Lock()
	defer s.lock.Unlock()
	c := make(chan struct{}, 1)
	s.mp[key] = c
	return c
}

func (s *SyncMap) Trick(key string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	_, ok := s.mp[key]
	if ok {
		s.mp[key] <- struct{}{}
	}
}

func (s *SyncMap) Delete(key string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	_, ok := s.mp[key]
	if ok {
		delete(s.mp, key)
	}
}
