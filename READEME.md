# 并发有序链表

## 项目背景

Golang训练营大作业：实现一个并发安全的有序链表（数据严格有序且没有重复元素）

## 功能特性

1. 基本功能

- 实现并发安全的插入、删除、遍历等功能
- 数据严格有序
- 不包含重复的元素

```golang
    // 检查一个元素是否存在，如果存在则返回 true，否则返回 false
    Contains(value int) bool
    
    // 插入一个元素，如果此操作成功插入一个元素，则返回 true，否则返回 false
    Insert(value int) bool
    
    // 删除一个元素，如果此操作成功删除一个元素，则返回 true，否则返回 false
    Delete(value int) bool
    
    // 遍历此有序链表的所有元素，如果 f 返回 false，则停止遍历
    Range(f func(value int) bool)
    
    // 返回有序链表的元素个数
    Len() int
```

2. 特性

- 插入时锁的粒度为要插入位置的前一个节点
- 删除时锁的粒度为要删除节点与期前一个节点
- 遍历等功能均为lock-free