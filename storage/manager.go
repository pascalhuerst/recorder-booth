package storage

import (
	"fmt"
	"sync"
)

// Manager can store an audio stream
type Manager struct {
	byteStream chan []byte
	mutex      sync.Mutex
	handlers   []Handler
}

// Handler used to add storage handlers
type Handler interface {
	store([]byte)
}

// NewManager factory for manager
func NewManager() *Manager {
	ret := Manager{
		byteStream: make(chan []byte),
	}
	go ret.run()
	return &ret
}

// InputChannel returns the input channel for Manager
func (m *Manager) InputChannel() chan []byte {
	return m.byteStream
}

// Add adds handlers
func (m *Manager) Add(h Handler) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.handlers = append(m.handlers, h)
}

func (m *Manager) run() {
	fmt.Printf("Starting manager\n")
	for {
		data := <-m.byteStream

		m.mutex.Lock()
		curHandlers := m.handlers
		m.mutex.Unlock()

		for _, ca := range curHandlers {
			go ca.store(data)
		}
	}
}
