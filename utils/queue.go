package utils

import (
	"fmt"
	"sync"
)

type Queue struct {
	q []interface{}
	sync.RWMutex
}

func NewQueue() Queue {
	return Queue{
		make([]interface{}, 0),
		sync.RWMutex{},
	}
}

func (q *Queue) Push(n interface{}) {
	q.Lock()
	defer q.Unlock()
	q.q = append(q.q, n)
}

func (q *Queue) Pop() interface{} {
	q.Lock()
	defer q.Unlock()

	if len(q.q) == 0 {
		return nil
	}
	ret := q.q[0]
	q.q = q.q[1:]
	return ret
}

func (q Queue) Peek() interface{} {
	q.Lock()
	defer q.Unlock()
	if len(q.q) == 0 {
		return nil
	}
	return q.q[0]
}

func (q Queue) String() string {
	return fmt.Sprintf("queue: %v", q.q)
}

func (q Queue) Slice() []interface{} {
	return q.q
}
