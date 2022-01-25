package events

import (
	"sync"
)

type Event[T any] interface {
	Subscribe(subscriber func(data T)) (unsubscribe func())
	SubscribeOnce(subscriber func(data T)) (unsubscribe func())
	SubscribeAsync(subscriber func(data T), opts ...bool) (unsubscribe func())
	SubscribeOnceAsync(subscriber func(data T), opts ...bool) (unsubscribe func())
	Dispatch(data T)
	Wait()
}

type event[T any] struct {
	handlers []*eventHandler[T]
	lock 			sync.Mutex
	wg 				sync.WaitGroup
}

type eventHandler[T any] struct {
	subscriber func(T)
	once bool
	async bool
	transactional bool
}

func New[T any]() Event[T] {
	return &event[T]{}
}

func (e *event[T]) Subscribe(subscriber func(T)) func() {
	return e.addSubscriber(&eventHandler[T]{
		subscriber: subscriber,
	})
}

func (e *event[T]) SubscribeOnce(subscriber func(T)) func() {
	return e.addSubscriber(&eventHandler[T]{
		subscriber: subscriber,
		once: true,
	})
}

func (e *event[T]) SubscribeAsync(subscriber func(T), opts ...bool) func() { 
	transactional := false
	if len(opts) > 0 {
		transactional = opts[0]
	}

	return e.addSubscriber(&eventHandler[T]{
		subscriber: subscriber,
		async: true,
		transactional: transactional,
	})
}

func (e *event[T]) SubscribeOnceAsync(subscriber func(T), opts ...bool) func() {
	transactional := false
	if len(opts) > 0 {
		transactional = opts[0]
	}
	return e.addSubscriber(&eventHandler[T]{
		subscriber: subscriber,
		once: true,
		async: true,
		transactional: transactional,
	})
}

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
			e.wg.Add(1)
			go e.callHandlerAsync(handler, data)
		}
	}
}

func (e *event[T]) Wait() {
	e.wg.Wait()
}

func (e *event[T]) callHandlerAsync(h *eventHandler[T], data T) {
	defer e.wg.Done()
	if h.transactional {
		e.lock.Lock()
		defer e.lock.Unlock()
	}
	h.subscriber(data)
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
