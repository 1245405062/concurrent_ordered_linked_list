package main

import (
	"sync"
	"sync/atomic"
)

type IntList struct {
	head   *intNode
	length int32
	mtx    sync.RWMutex
}

type intNode struct {
	mtx   sync.RWMutex
	value int
	next  *intNode
}

func newIntNode(value int) *intNode {
	return &intNode{value: value}
}

func NewIntList() *IntList {
	return &IntList{head: newIntNode(0)}
}

func (l *IntList) Insert(value int) bool {
	for {
		a := l.head
		b := a.next
		for b != nil && b.value < value {
			a = b
			b = b.next
		}

		needReturn := func() bool {
			a.mtx.Lock()
			defer a.mtx.Unlock()
			// 如果a.next改变了，重新找到a和b
			if a.next != b {
				return false
			}
			c := newIntNode(value)
			c.next = b
			a.next = b
			atomic.AddInt32(&l.length, 1)
			return true
		}()
		if needReturn {
			return true
		}
	}
}
