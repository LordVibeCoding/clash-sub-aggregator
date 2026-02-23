package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"clash-sub-aggregator/internal/model"
)

type Store struct {
	path string
	mu   sync.RWMutex
	data *model.StoreData
}

func New(dataDir string) (*Store, error) {
	path := filepath.Join(dataDir, "subscriptions.json")
	s := &Store{
		path: path,
		data: &model.StoreData{},
	}
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}
	if err := s.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	return s, nil
}

func (s *Store) load() error {
	raw, err := os.ReadFile(s.path)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, s.data)
}

func (s *Store) save() error {
	raw, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, raw, 0644)
}

func (s *Store) List() []model.Subscription {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]model.Subscription, len(s.data.Subscriptions))
	copy(result, s.data.Subscriptions)
	return result
}

func (s *Store) Get(id string) (model.Subscription, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, sub := range s.data.Subscriptions {
		if sub.ID == id {
			return sub, true
		}
	}
	return model.Subscription{}, false
}

func (s *Store) Add(sub model.Subscription) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data.Subscriptions = append(s.data.Subscriptions, sub)
	return s.save()
}

func (s *Store) Update(sub model.Subscription) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, existing := range s.data.Subscriptions {
		if existing.ID == sub.ID {
			s.data.Subscriptions[i] = sub
			return s.save()
		}
	}
	return nil
}

func (s *Store) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, sub := range s.data.Subscriptions {
		if sub.ID == id {
			s.data.Subscriptions = append(s.data.Subscriptions[:i], s.data.Subscriptions[i+1:]...)
			return s.save()
		}
	}
	return nil
}
