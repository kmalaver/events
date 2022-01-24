package events

import (
	"sync"
	"reflect"
)

type Event[T any] interface {
	Subscribe(subscriber func(data T))
	SubscribeOnce(subscriber func(data T))
	Unsubscribe(subscriber func(data T))
	SubscribeAsync(subscriber func(data T), transactional bool)
	SubscribeOnceAsync(subscriber func(data T), transactional bool)
	Dispatch(data T)
	Wait()
}


type event[T any] struct {
	handlers []*handler[T]
	lock 			sync.Mutex
	wg 				sync.WaitGroup
}

type handler[T any] struct {
	subscriber func(T)
	once bool
	async bool
	transactional bool
}

func New[T any]() Event[T] {
	return &event[T]{}
}

func (e *event[T]) Subscribe(subscriber func(T)) {
	e.addSubscriber(&handler[T]{
		subscriber: subscriber,
	})
}

func (e *event[T]) SubscribeOnce(subscriber func(T)) {
	e.addSubscriber(&handler[T]{
		subscriber: subscriber,
		once: true,
	})
}

func (e *event[T]) SubscribeAsync(subscriber func(T), transactional bool) {
	e.addSubscriber(&handler[T]{
		subscriber: subscriber,
		async: true,
		transactional: transactional,
	})
}

func (e *event[T]) SubscribeOnceAsync(subscriber func(T), transactional bool) {
	e.addSubscriber(&handler[T]{
		subscriber: subscriber,
		once: true,
		async: true,
		transactional: transactional,
	})
}

func (e *event[T]) Unsubscribe(subscriber func(T)) {
	e.lock.Lock()
	defer e.lock.Unlock()
	idx := e.findHandler(subscriber)
	e.removeHandler(idx)
}


func (e *event[T]) Dispatch(data T) {
	e.lock.Lock()
	defer e.lock.Unlock()

	handlersCopy := make([]*handler[T], len(e.handlers))
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

func (e *event[T]) callHandlerAsync(h *handler[T], data T) {
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

func (e *event[T]) findHandler(subscriber func(T)) int {
	for idx, s := range e.handlers {
		fn1 := reflect.ValueOf(s.subscriber).Pointer()
		fn2 := reflect.ValueOf(subscriber).Pointer()
		if fn1 == fn2 {
			return idx
		}
	}
	return -1
}


func (e *event[T]) addSubscriber(h *handler[T]) {
	e.lock.Lock()
	defer e.lock.Unlock()
	e.handlers = append(e.handlers, h)
}
