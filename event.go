package events

import (
	"sync"
)

type event[T any] struct {
	handlers []*eventHandler[T]
	lock     sync.Mutex
	wg       sync.WaitGroup
}

type eventHandler[T any] struct {
	subscriber    func(T)
	once          bool
	async         bool
	transactional bool
}

// Returns new event
func New[T any]() *event[T] {
	return &event[T]{}
}

// Subscribes to event and returns unsubscribe function
func (e *event[T]) Subscribe(subscriber func(T)) func() {
	return e.addSubscriber(&eventHandler[T]{
		subscriber: subscriber,
	})
}

// Subscribe to event once and returns unsubscribe function.
// Subscriber will be called only once
func (e *event[T]) SubscribeOnce(subscriber func(T)) func() {
	return e.addSubscriber(&eventHandler[T]{
		subscriber: subscriber,
		once:       true,
	})
}

// Subscribe to event and call subscriber in separate goroutine.
// Can be passed transactional bool to make subscriber transactional
func (e *event[T]) SubscribeAsync(subscriber func(T), opts ...bool) func() {
	transactional := false
	if len(opts) > 0 {
		transactional = opts[0]
	}

	return e.addSubscriber(&eventHandler[T]{
		subscriber:    subscriber,
		async:         true,
		transactional: transactional,
	})
}

// Subscribe to event and call subscriber in separate goroutine.
// Will be called only once.
// Can be passed transactional bool to make subscriber transactional
func (e *event[T]) SubscribeOnceAsync(subscriber func(T), opts ...bool) func() {
	transactional := false
	if len(opts) > 0 {
		transactional = opts[0]
	}
	return e.addSubscriber(&eventHandler[T]{
		subscriber:    subscriber,
		once:          true,
		async:         true,
		transactional: transactional,
	})
}

// Dispatch event to all subscribers
func (e *event[T]) Dispatch(data T) {
	e.lock.Lock()
	defer e.lock.Unlock()

	handlersCopy := make([]*eventHandler[T], len(e.handlers))
	copy(handlersCopy, e.handlers)

	for i, handler := range handlersCopy {
		if handler.once {
			e.removeHandler(i)
		}
		if !handler.async {
			handler.subscriber(data)
		} else {
			e.callHandlerAsync(handler, data)
		}
	}
}

// Wait for all async subscribers to finish
func (e *event[T]) Wait() {
	e.wg.Wait()
}

func (e *event[T]) callHandlerAsync(h *eventHandler[T], data T) {
	e.wg.Add(1)

	var wg sync.WaitGroup
	if h.transactional {
		wg.Add(1)
	}
	go func() {
		defer e.wg.Done()
		h.subscriber(data)
		if h.transactional {
			wg.Done()
		}
	}()
	if h.transactional {
		wg.Wait()
	}
}

func (e *event[T]) removeHandler(idx int) {
	l := len(e.handlers)
	if !(0 <= idx && idx < l) {
		return
	}
	e.handlers = append(e.handlers[:idx], e.handlers[idx+1:]...)
}

func (e *event[T]) findHandler(h *eventHandler[T]) int {
	for i := range e.handlers {
		if e.handlers[i] == h {
			return i
		}
	}
	return -1
}

func (e *event[T]) addSubscriber(h *eventHandler[T]) func() {
	e.lock.Lock()
	defer e.lock.Unlock()
	e.handlers = append(e.handlers, h)

	return func() {
		e.lock.Lock()
		defer e.lock.Unlock()
		e.removeHandler(e.findHandler(h))
	}
}
