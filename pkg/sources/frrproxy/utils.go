package frrproxy

import (
	"crypto/sha1"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"sync"

	"github.com/alice-lg/alice-lg/pkg/api"
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

// PeerHash creates a peer hash by its config.ID and neighbor address
func PeerHash(configID string, address string) string {
	h := sha1.New()
	io.WriteString(h, configID)
	io.WriteString(h, address)
	sum := h.Sum(nil)
	return fmt.Sprintf("%x", sum[0:5])
}

// Parse a list of FRR returned community strings
func parseBgpCommunityList(list []string) api.Communities {
	var result api.Communities

	for _, item := range list {
		parts := strings.Split(item, ":")
		if len(parts) > 2 {
			var community api.Community

			first, err1 := strconv.Atoi(parts[0])
			second, err2 := strconv.Atoi(parts[1])

			if err1 != nil || err2 != nil {
				continue
			}

			community = append(community, first)
			community = append(community, second)

			if len(parts) == 3 {
				third, err3 := strconv.Atoi(parts[2])

				if err3 != nil {
					continue
				}

				community = append(community, third)
			}

			result = append(result, community)
		} else {
			continue
		}
	}

	return result
}

// Parse the extendedCommunity string
func parseExtBgpCommunities(str string) api.ExtCommunities {
	var result api.ExtCommunities

	extComm := strings.Split(str, " ")
	for _, item := range extComm {
		if item != "" {
			log.Printf("ExtCommunity: %s", item)
			result = append(result, api.ExtCommunity{item})
		}
	}

	return result
}
