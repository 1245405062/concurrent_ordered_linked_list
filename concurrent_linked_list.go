package concurrent_ordered_linked_list

import (
	"sync"
	"sync/atomic"
	"unsafe"
)

const (
	StatusContinue    = 0 // 自旋
	StatusReturnTrue  = 1 // 返回成功插入/删除
	StatusReturnFalse = 2 // 返回失败插入/删除

	StatusNotDeleted = 0
	StatusDeleted    = 1
)

type IntList struct {
	head   *intNode
	length int
}

type intNode struct {
	mtx     sync.Mutex
	value   int
	next    *intNode
	deleted int
}

func newIntNode(value int) *intNode {
	return &intNode{value: value, deleted: StatusNotDeleted}
}

func NewInt() *IntList {
	return &IntList{head: newIntNode(0)}
}

func (l *IntList) Insert(value int) bool {
	for {
		// 原子找到value需要插入的位置
		a := (*intNode)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&l.head))))
		b := (*intNode)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&a.next))))

		for b != nil &&
			int(atomic.LoadInt32((*int32)(unsafe.Pointer(&(b.value))))) < value {
			a = b
			b = (*intNode)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&a.next))))
		}

		status := func() int {
			a.mtx.Lock()
			defer a.mtx.Unlock()
			// 如果a.next不为b或a已经被标记删除了，自旋
			if (atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&a.next))) !=
				atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&b)))) ||
				(atomic.LoadInt32((*int32)(unsafe.Pointer(&a.deleted))) == StatusDeleted) {
				return StatusContinue
			}
			// 如果当前b的值与value相同，直接返回false，保证相同的值不重复插入
			if b != nil && int(atomic.LoadInt32((*int32)(unsafe.Pointer(&(b.value))))) == value {
				return StatusReturnFalse
			}

			// 在a和b之间插入节点c
			c := newIntNode(value)
			atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&c.next)), unsafe.Pointer(b))
			atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&a.next)), unsafe.Pointer(c))

			// 长度原子++
			atomic.AddInt32((*int32)(unsafe.Pointer(&l.length)), 1)
			return StatusReturnTrue
		}()
		switch status {
		case StatusReturnTrue:
			return true
		case StatusReturnFalse:
			return false
		}
	}
}

func (l *IntList) Delete(value int) bool {
	for {
		// 原子找到value需要插入的位置
		a := (*intNode)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&l.head))))
		b := (*intNode)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&a.next))))

		for b != nil &&
			int(atomic.LoadInt32((*int32)(unsafe.Pointer(&(b.value))))) < value {
			a = b
			b = (*intNode)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&a.next))))
		}
		// 遍历到最后仍找不到value直接返回false
		if b == nil {
			return false
		}

		status := func() int {
			// 锁定b节点，如果b被删除了，自旋
			b.mtx.Lock()
			defer b.mtx.Unlock()
			if atomic.LoadInt32((*int32)(unsafe.Pointer(&b.deleted))) == StatusDeleted {
				return StatusContinue
			}
			// 如果a.next不为b或a已经被标记删除了，自旋
			a.mtx.Lock()
			defer a.mtx.Unlock()
			if (*intNode)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&a.next)))) != b ||
				atomic.LoadInt32((*int32)(unsafe.Pointer(&a.deleted))) == StatusDeleted {
				return StatusContinue
			}
			// 如果b节点的值不是value，找不到要删除的节点，返回false
			if int(atomic.LoadInt32((*int32)(unsafe.Pointer(&b.value)))) != value {
				return StatusReturnFalse
			}
			// 标记b为要删除的节点并进行删除
			atomic.StoreInt32((*int32)(unsafe.Pointer(&b.deleted)), StatusDeleted)
			atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&a.next)), unsafe.Pointer(b.next))

			// 长度原子--
			atomic.AddInt32((*int32)(unsafe.Pointer(&l.length)), -1)
			return StatusReturnTrue
		}()
		switch status {
		case StatusReturnTrue:
			return true
		case StatusReturnFalse:
			return false
		}
	}
}

func (l *IntList) Contains(value int) bool {
	// 从第一个节点并依次原子遍历
	head := (*intNode)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&l.head.next))))
	for head != nil && int(atomic.LoadInt32((*int32)(unsafe.Pointer(&(head.value))))) <= value {
		if int(atomic.LoadInt32((*int32)(unsafe.Pointer(&(head.value))))) == value {
			return true
		}
		head = (*intNode)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&head.next))))
	}
	return false
}

func (l *IntList) Range(f func(value int) bool) {
	// 从第一个节点并依次原子遍历，执行f()
	head := (*intNode)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&l.head.next))))
	for head != nil {
		if !f(int(atomic.LoadInt32((*int32)(unsafe.Pointer(&(head.value)))))) {
			break
		}
		head = (*intNode)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&head.next))))
	}
}

func (l *IntList) Len() int {
	// 原子读length
	return (int)(atomic.LoadInt32((*int32)(unsafe.Pointer(&l.length))))
}
