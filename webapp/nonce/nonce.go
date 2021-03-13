package nonce

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"sync"
	"time"
)

type Store struct {
	mu     sync.RWMutex
	salt   string
	ttl    time.Duration
	nonces map[string]time.Time
	h      hash.Hash
	newTTL chan struct{}
}

func (s *Store) hash(exp time.Time) string {

	s.h.Reset()
	io.WriteString(s.h, fmt.Sprintf("%s:%s", exp.String(), s.salt))
	theHash := s.h.Sum(nil)
	rval := make([]byte, len(theHash))

	for k, v := range theHash {
		rval[k] = v
	}

	return hex.EncodeToString(rval)
}

func (s *Store) Nonce() string {

	s.mu.Lock()
	defer s.mu.Unlock()

	exp := time.Now().Add(s.ttl)
	n := s.hash(exp)
	s.nonces[n] = exp

	return n
}

func (s *Store) prune() {

	for {

		select {
		case when := <-time.After(5 * s.ttl):
			s.mu.Lock()
			for k, v := range s.nonces {
				if when.After(v) {
					delete(s.nonces, k)
				}
			}
			s.mu.Unlock()
		case <-s.newTTL:
			continue
		}
	}
}

func (s *Store) Salt(salt string) {
	s.salt = salt
}

func (s *Store) TTL(d time.Duration) {
	s.ttl = d
	s.newTTL <- struct{}{}
}

func (s *Store) Validate(n string) bool {

	s.mu.RLock()
	defer s.mu.RUnlock()

	if exp, ok := s.nonces[n]; ok {

		if time.Now().After(exp) {
			return false
		}

		if s.hash(exp) == n {
			delete(s.nonces, n)
			return true
		}
	}

	return false
}

func New() (*Store, error) {

	var randBytes = make([]byte, 20)
	if _, e := rand.Read(randBytes); e != nil {
		return nil, e
	}
	var rval = &Store{
		ttl:    time.Duration(10 * time.Minute),
		nonces: map[string]time.Time{},
		salt:   string(randBytes),
		h:      sha1.New(),
	}

	go rval.prune()

	return rval, nil
}
