package main

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"
)

func TestInsert(t *testing.T) {
	var wg sync.WaitGroup
	intList := NewIntList()
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			randSec := int64(rand.Intn(5))
			sec := time.Duration(randSec) * time.Second
			time.Sleep(sec)
			intList.Insert(i)
			wg.Done()
		}(i)
	}
	wg.Wait()
	node := intList.head
	for node != nil {
		fmt.Println(node.value)
		node = node.next
	}

}
