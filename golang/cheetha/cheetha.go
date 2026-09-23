package cheetah

import (
	"sync"
)

// Cheetah is a generic pub/sub system keyed by string.
// T is the message payload type.
type Cheetah[K comparable, T any] struct {
	mu          sync.Mutex
	subscribers map[K]map[chan *T]struct{}
	last        map[K]*T // ← latched last value per key
}

// New returns a new Cheetah instance.
func New[K comparable, T any](buffer int) *Cheetah[K, T] {
	return &Cheetah[K, T]{
		subscribers: make(map[K]map[chan *T]struct{}, buffer),
		last:        make(map[K]*T),
	}
}

func (l *Cheetah[K, T]) Subscribe(key K) chan *T {
	ch := make(chan *T, 1)
	l.mu.Lock()
	defer l.mu.Unlock()

	if _, ok := l.subscribers[key]; !ok {
		l.subscribers[key] = make(map[chan *T]struct{}, 1)
	}
	l.subscribers[key][ch] = struct{}{}

	if val, ok := l.last[key]; ok {
		ch <- val
		delete(l.last, key) // ← evict here, only one subscriber gets the replay
	}

	return ch
}

func (l *Cheetah[K, T]) Publish(key K, parcel *T) {
	//	l.mu.Lock()
	//	defer l.mu.Unlock()
	//
	//	l.last[key] = parcel

	//for ch := range l.subscribers[key] {
	//	select {
	//	case ch <- parcel:
	//	default:
	//	}
	//}
	l.mu.Lock()
	l.last[key] = parcel
	chans := make([]chan *T, 0, len(l.subscribers[key]))
	for ch := range l.subscribers[key] {
		chans = append(chans, ch)
	}
	l.mu.Unlock()
	for _, ch := range chans {
		ch <- parcel // blocking send — never silently dropped
	}
}

// Unsubscribe removes and closes the channel for key.
func (l *Cheetah[K, T]) Unsubscribe(key K, body chan *T) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if sub, ok := l.subscribers[key]; ok {
		delete(sub, body)
		close(body)
		if len(sub) == 0 {
			delete(l.subscribers, key)
		}
	}
}

// Evict clears the latched value for key (call after result is consumed).
func (l *Cheetah[K, T]) Evict(key K) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.last, key)
}
