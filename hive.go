package main

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/c4pt0r/log"
	"github.com/c4pt0r/tipubsub"
)

var (
	ErrBeeAlreadyExists error = errors.New("bee already exists")
)

type Bee struct {
	BeeName           string `json:"bee_name"`
	InstanceID        string `json:"instance_id"`
	HeartbeatDuration int    `json:"heartbeat_duration"`

	mu            sync.RWMutex
	h             *Hive
	lastHeartbeat time.Time
}

type Hive struct {
	mu            sync.RWMutex
	bees          map[string][]*Bee
	beesInstances map[string]*Bee

	hub *tipubsub.Hub
}

func NewHive() *Hive {
	h := &Hive{
		bees:          make(map[string][]*Bee),
		beesInstances: make(map[string]*Bee),
	}

	c := tipubsub.DefaultConfig()
	c.DSN = *mysqlDSN

	var err error
	h.hub, err = tipubsub.NewHub(c)
	if err != nil {
		log.Fatal(err)
	}
	return h
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
	h.hub.Subscribe(bee.InstanceID, bee, tipubsub.LatestId)

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

func (h *Hive) SendToBee(instanceID string, data string) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	bee := h.beesInstances[instanceID]
	if bee == nil {
		return errors.New(fmt.Sprintf("bee with instance id %s not found", instanceID))
	}
	return bee.h.hub.Publish(bee.InstanceID, &tipubsub.Message{
		Data: data,
	})
}

func (b *Bee) UpdateHeartbeat() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.lastHeartbeat = time.Now()
}

func (b *Bee) GetLastHeartbeat() time.Time {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return b.lastHeartbeat
}

func (b *Bee) Kill() {
	b.h.RemoveBee(b)
}

func (b *Bee) IsTimeout() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return time.Now().Sub(b.lastHeartbeat) > 2*time.Duration(b.HeartbeatDuration)*time.Second
}

var _ tipubsub.Subscriber = (*Bee)(nil)

func (b *Bee) ID() string {
	return b.InstanceID
}

func (b *Bee) OnMessages(topic string, msgs []tipubsub.Message) {
	for _, msg := range msgs {
		// TODO
	}
}
