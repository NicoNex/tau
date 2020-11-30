package utils

import (
	"fmt"
	"sync"
)

type Stack struct {
	s []interface{}
	sync.RWMutex
}

func NewStack() Stack {
	return Stack{
		make([]interface{}, 0),
		sync.RWMutex{},
	}
}

func (s *Stack) Push(n interface{}) {
	s.Lock()
	defer s.Unlock()
	s.s = append(s.s, n)
}

func (s *Stack) Pop() interface{} {
	s.Lock()
	defer s.Unlock()
	var l = len(s.s)

	if l == 0 {
		return nil
	}
	ret := s.s[l-1]
	s.s = s.s[:l-1]
	return ret
}

func (s Stack) Peek() interface{} {
	s.Lock()
	defer s.Unlock()
	var l = len(s.s)

	if l == 0 {
		return nil
	}
	return s.s[l-1]
}

func (s Stack) String() string {
	return fmt.Sprintf("stack: %v", s.s)
}

func (s Stack) Slice() []interface{} {
	return s.s
}
