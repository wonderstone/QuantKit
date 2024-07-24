package queue

type Queue[T any] struct {
	data       []T
	head, tail int
	size, cap  int
}

func New[T any](n int) *Queue[T] {
	return &Queue[T]{
		data: make([]T, n),
		cap:  n,
	}
}

// Enqueue 现在会在队列满时自动出队头部元素。
func (q *Queue[T]) Enqueue(value T) bool {
	if q.size == q.cap {
		return false
	}
	q.data[q.tail] = value
	q.tail = (q.tail + 1) % q.cap
	q.size++

	return true
}

// EnqueueWithDequeue 现在会在队列满时自动出队头部元素。
func (q *Queue[T]) EnqueueWithDequeue(value T) (head T, full bool) {
	if q.size == q.cap {
		head, _ = q.Dequeue() // 自动出队头部元素
	}

	q.data[q.tail] = value
	q.tail = (q.tail + 1) % q.cap
	if q.size < q.cap {
		q.size++
	}

	full = q.size == q.cap

	return
}

func (q *Queue[T]) Dequeue() (value T, ok bool) {
	if q.size == 0 {
		return value, false
	}
	value = q.data[q.head]
	q.head = (q.head + 1) % q.cap
	q.size--
	return value, true
}

// Peek 返回队列头部的元素但不移除它，如果队列为空则返回false和零值。
func (q *Queue[T]) Peek() (value T, ok bool) {
	if q.size == 0 {
		return value, false // 队列为空
	}
	value = q.data[q.head]
	return value, true
}

// PeekTail 返回尾部的元素但不移除它，如果队列为空则返回false和零值。
func (q *Queue[T]) PeekTail() (value T, ok bool) {
	if q.size == 0 {
		return value, false // 队列为空
	}
	value = q.data[(q.tail-1+q.cap)%q.cap]
	return value, true
}

// Get 返回从头部开始的第N个元素，如果不存在则返回false和零值。
func (q *Queue[T]) Get(n int) (value T, ok bool) {
	if n >= q.size {
		return value, false
	}
	index := (q.head + n) % q.cap
	return q.data[index], true
}

// Get 返回从头部开始的第N个元素，如果不存在则返回false和零值。
func (q *Queue[T]) MustGet(n int) (value T) {
	if n >= q.size {
		panic("队列元素下标越界")
	}
	index := (q.head + n) % q.cap
	return q.data[index]
}

// ToSlice 返回队列中的元素，按照从头部到尾部的顺序。
func (q *Queue[T]) ToSlice() []T {
	if len(q.data) <= 5 {
		return q.toSlice2()
	} else {
		return q.toSlice()
	}
}

// ToSlice 返回队列中的元素，按照从头部到尾部的顺序。
func (q *Queue[T]) toSlice() []T {
	result := make([]T, q.size)
	if q.head < q.tail { // 如果没有环绕
		copy(result, q.data[q.head:q.tail])
	} else { // 如果环绕了
		n := copy(result, q.data[q.head:])
		copy(result[n:], q.data[:q.tail])
	}
	return result
}

// ToSlice2 返回队列中的元素，按照从头部到尾部的顺序。
func (q *Queue[T]) toSlice2() []T {
	result := make([]T, q.size)
	for i := 0; i < q.size; i++ {
		index := (q.head + i) % q.cap
		result[i] = q.data[index]
	}
	return result
}

func (q *Queue[T]) Len() int {
	return q.size
}

func (q *Queue[T]) Size() int {
	return q.size
}

func (q *Queue[T]) Full() bool {
		return q.size == q.cap
}

func (q *Queue[T]) Empty() bool {
	return q.size == 0
	}
	
func (q *Queue[T]) Clear() {
	q.head = 0
	q.tail = 0
	q.size = 0
}
