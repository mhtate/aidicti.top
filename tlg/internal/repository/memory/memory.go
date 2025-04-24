package memory

import (
	"fmt"
	"sync"
	"time"

	"aidicti.top/ai/internal/model"
)

type dialog struct {
	model.Dialog
	lastAccessed time.Time
}

// Temporarily memory storage implementation with auto-deletion
type temporarilyDialogStorage struct {
	data       map[model.ID]dialog
	nextID     model.ID
	mu         sync.RWMutex
	expiryTime time.Duration
	stopChan   chan struct{}
}

func NewTemporarilyDialogStorage(expiryTime time.Duration) *temporarilyDialogStorage {
	storage := &temporarilyDialogStorage{
		data:       make(map[model.ID]dialog),
		nextID:     1,
		expiryTime: expiryTime,
		stopChan:   make(chan struct{}),
	}
	go storage.cleanupExpiredDialogs()
	return storage
}

func (s *temporarilyDialogStorage) Get(id model.ID) (model.Dialog, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	dialog, exists := s.data[id]
	if !exists {
		return model.Dialog{}, fmt.Errorf("dialog with id %d not found", id)
	}

	dialog.lastAccessed = time.Now()
	s.data[id] = dialog

	return dialog.Dialog, nil
}

func (s *temporarilyDialogStorage) Create() (model.Dialog, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := s.nextID
	s.nextID++
	dialog := dialog{
		Dialog:       model.Dialog{ID: id},
		lastAccessed: time.Now(),
	}
	s.data[id] = dialog
	return dialog.Dialog, nil
}

func (s *temporarilyDialogStorage) Store(d model.Dialog) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data[d.ID] = dialog{d, time.Now()}
	return nil
}

func (s *temporarilyDialogStorage) Delete(d model.Dialog) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.data, d.ID)

	return nil
}

func (s *temporarilyDialogStorage) cleanupExpiredDialogs() {
	ticker := time.NewTicker(s.expiryTime)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.mu.Lock()
			now := time.Now()
			for id, dialog := range s.data {
				if now.Sub(dialog.lastAccessed) > s.expiryTime {
					delete(s.data, id)
				}
			}
			s.mu.Unlock()
		case <-s.stopChan:
			return
		}
	}
}

func (s *temporarilyDialogStorage) StopCleanup() {
	close(s.stopChan)
}
