package subscriptions

import (
	"fmt"
	"slices"
	"strings"
	"sync"
)

type SubscriptionStore interface {
	AddTopic(userID int64, topic string) error
	RemoveTopic(userID int64, topic string) error
	GetTopics(userID int64) []string
}

type Store struct {
	Mtx  *sync.RWMutex
	Data map[int64][]string
}

func (s *Store) AddTopic(userID int64, topic string) error {
	s.Mtx.RLock()
	ok := slices.Contains(s.Data[userID], topic)
	s.Mtx.RUnlock()
	if ok {
		return fmt.Errorf("topic is already added")
	}

	if topic == "" {
		return fmt.Errorf("topic name can't be empty")
	}

	s.Mtx.Lock()
	defer s.Mtx.Unlock()
	s.Data[userID] = append(s.Data[userID], topic)

	return nil
}

func (s *Store) RemoveTopic(userID int64, topic string) error {
	s.Mtx.Lock()
	defer s.Mtx.Unlock()

	if len(s.Data[userID]) <= 1 {
		s.Data[userID] = []string{}
		return nil
	}

	existsTopics := strings.Join(s.Data[userID], ",")
	fmt.Println(existsTopics)
	before, after, found := strings.Cut(existsTopics, topic+",")
	if !found {
		return fmt.Errorf("user has no such topic")
	}

	s.Data[userID] = strings.Split(before+after, ",")

	return nil
}

func (s *Store) GetTopics(userID int64) []string {
	s.Mtx.RLock()
	defer s.Mtx.RUnlock()
	topics := s.Data[userID]
	result := make([]string, len(topics))
	copy(result, topics)
	return result
}
