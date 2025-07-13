package frrproxy

import (
	"crypto/sha1"
	"fmt"
	"io"
	"sync"
)

/*
LockMap uses the sync.Map to manage locks, accessed by a key.
TODO: Maybe this would be a nice generic helper
*/
type LockMap struct {
	locks *sync.Map
}

// NewLockMap creates a new LockMap
func NewLockMap() *LockMap {
	return &LockMap{
		locks: &sync.Map{},
	}
}

// Lock locks the lock.
func (m *LockMap) Lock(key string) {
	mutex, _ := m.locks.LoadOrStore(key, &sync.Mutex{})
	mutex.(*sync.Mutex).Lock()
}

// Unlock unlocks the locked LockMap-lock.
func (m *LockMap) Unlock(key string) {
	mutex, ok := m.locks.Load(key)
	if !ok {
		return // no lock
	}
	mutex.(*sync.Mutex).Unlock()
}

// PeerHashWithASAndAddress creates a peer hash (sha1) from
// the ASN and the address.
func PeerHashWithASAndAddress(asn uint32, address string) string {
	h := sha1.New()
	io.WriteString(h, fmt.Sprintf("%v", asn))
	io.WriteString(h, address)
	sum := h.Sum(nil)
	return fmt.Sprintf("%x", sum[0:5])
}

// PeerHash creates a peer hash by its neighbor address
func PeerHash(address string) string {
	h := sha1.New()
	io.WriteString(h, address)
	sum := h.Sum(nil)
	return fmt.Sprintf("%x", sum[0:5])
}
