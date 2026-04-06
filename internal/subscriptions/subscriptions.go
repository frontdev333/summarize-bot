package subscriptions

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
)

type SubscriptionStore interface {
	AddTopic(userID int64, topic string) error
	RemoveTopic(userID int64, topic string) error
	GetTopics(userID int64) []string
}

type InMemoryStore struct {
	Mtx  *sync.RWMutex
	Data map[int64][]string
}

type FileStore struct {
	mtx  *sync.RWMutex
	path string
	data map[int64][]string `json:"data"`
}

func NewFileStore(pth string) (*FileStore, error) {
	bytes, err := os.ReadFile(pth)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}

		file, err := os.Create(pth)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		return &FileStore{
			mtx:  &sync.RWMutex{},
			path: pth,
			data: make(map[int64][]string),
		}, nil
	}

	store := &FileStore{}

	if err = json.Unmarshal(bytes, store); err != nil {
		return nil, err
	}

	store.path = pth

	return store, nil
}

func (s *InMemoryStore) AddTopic(userID int64, topic string) error {
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

func (s *InMemoryStore) RemoveTopic(userID int64, topic string) error {
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

func (s *InMemoryStore) GetTopics(userID int64) []string {
	s.Mtx.RLock()
	defer s.Mtx.RUnlock()
	topics := s.Data[userID]
	result := make([]string, len(topics))
	copy(result, topics)
	return result
}

func (s *FileStore) Save() error {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	bytes, err := json.MarshalIndent(s, "", "	")
	if err != nil {
		return err
	}

	pth, _ := filepath.Split(s.path)

	tmpPath := filepath.Join(pth, "tmp_res.json")
	file, err := os.Create(tmpPath)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err = file.Write(bytes); err != nil {
		return err
	}

	return os.Rename(tmpPath, s.path)
}

func (s *FileStore) AddTopic(userID int64, topic string) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if topic == "" {
		return fmt.Errorf("topic name can't be empty")
	}

	ok := slices.Contains(s.data[userID], topic)
	if ok {
		return fmt.Errorf("topic is already added")
	}

	s.data[userID] = append(s.data[userID], topic)
	return s.Save()

}

func (s *FileStore) RemoveTopic(userID int64, topic string) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if len(s.data[userID]) <= 1 {
		s.data[userID] = []string{}
		return s.Save()
	}

	existsTopics := strings.Join(s.data[userID], ",")
	before, after, found := strings.Cut(existsTopics, topic+",")
	if !found {
		return fmt.Errorf("user has no such topic")
	}

	s.data[userID] = strings.Split(before+after, ",")

	return s.Save()
}

func (s *FileStore) GetTopics(userID int64) []string {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	topics := s.data[userID]
	result := make([]string, len(topics))
	copy(result, topics)
	return result
}
