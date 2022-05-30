package main

import (
	"errors"
	"sync"
	"time"
)

var (
	ErrBeeAlreadyExists error = errors.New("bee already exists")
)

type Bee struct {
	BeeName           string `json:"bee_name"`
	InstanceID        string `json:"instance_id"`
	HeartbeatDuration int    `json:"heartbeat_duration"`

	mu            sync.RWMutex
	lastHeartbeat time.Time
	h             *Hive
}

type Hive struct {
	mu            sync.RWMutex
	bees          map[string][]*Bee
	beesInstances map[string]*Bee
}

func NewHive() *Hive {
	return &Hive{
		bees:          make(map[string][]*Bee),
		beesInstances: make(map[string]*Bee),
	}
}

func NewBee(hive *Hive, name string, instanceID string, heartbeatDuration int) *Bee {
	return &Bee{
		BeeName:           name,
		InstanceID:        instanceID,
		HeartbeatDuration: heartbeatDuration,
		lastHeartbeat:     time.Now(),
		h:                 hive,
	}
}

func (h *Hive) AddBee(bee *Bee) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.beesInstances[bee.InstanceID]; ok {
		return ErrBeeAlreadyExists
	}

	h.bees[bee.BeeName] = append(h.bees[bee.BeeName], bee)
	h.beesInstances[bee.InstanceID] = bee

	go func(b *Bee) {
		for {
			if b.IsTimeout() {
				b.Kill()
				return
			}
			time.Sleep(time.Duration(b.HeartbeatDuration) * time.Second)
		}
	}(bee)
	return nil
}

func (h *Hive) RemoveBee(bee *Bee) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for i, b := range h.bees[bee.BeeName] {
		if b.InstanceID == bee.InstanceID {
			h.bees[bee.BeeName] = append(h.bees[bee.BeeName][:i], h.bees[bee.BeeName][i+1:]...)
			break
		}
	}
	delete(h.beesInstances, bee.InstanceID)
}

func (h *Hive) Bees(beeName string) []*Bee {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return h.bees[beeName]
}

func (h *Hive) AllBees() []*Bee {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var bees []*Bee
	for _, bs := range h.bees {
		bees = append(bees, bs...)
	}
	return bees
}

func (h *Hive) Bee(instanceID string) *Bee {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return h.beesInstances[instanceID]
}

func (h *Hive) BeesByName(name string) []*Bee {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return h.bees[name]
}

func (b *Bee) UpdateHeartbeat() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.lastHeartbeat = time.Now()
}

func (b *Bee) Kill() {
	b.h.RemoveBee(b)
}

func (b *Bee) IsTimeout() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return time.Now().Sub(b.lastHeartbeat) > 2*time.Duration(b.HeartbeatDuration)*time.Second
}
