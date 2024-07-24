package queue

import (
	"testing"
)

func TestQueue(t *testing.T) {
	// Test New function
	q := New[int](3)
	if q == nil {
		t.Error("New function returned nil")
	}
	if q.Size() != 0 {
		t.Errorf("Expected size 0, got %d", q.Size())
	}

	// Test Enqueue function
	q.Enqueue(1)
	if q.Size() != 1 {
		t.Errorf("Expected size 1, got %d", q.Size())
	}
	q.Enqueue(2)
	q.Enqueue(3)
	q.Enqueue(4)
	// Test EnqueueWithDequeue function
	head, full := q.EnqueueWithDequeue(5)
	if head != 1 {
		t.Errorf("Expected head 1, got %d", head)
	}
	if full {
		t.Error("Expected full to be false")
	}

	// Test Dequeue function
	value, ok := q.Dequeue()
	if value != 1 {
		t.Errorf("Expected value 1, got %d", value)
	}
	if !ok {
		t.Error("Expected ok to be true")
	}

	// Test Peek function
	q.Enqueue(3)
	value, ok = q.Peek()
	if value != 3 {
		t.Errorf("Expected value 3, got %d", value)
	}
	if !ok {
		t.Error("Expected ok to be true")
	}

	// Test PeekTail function
	value, ok = q.PeekTail()
	if value != 3 {
		t.Errorf("Expected value 3, got %d", value)
	}
	if !ok {
		t.Error("Expected ok to be true")
	}

	// Test Get function
	value, ok = q.Get(0)
	if value != 3 {
		t.Errorf("Expected value 3, got %d", value)
	}
	if !ok {
		t.Error("Expected ok to be true")
	}

	// Test MustGet function
	value = q.MustGet(0)
	if value != 3 {
		t.Errorf("Expected value 3, got %d", value)
	}

	// Test ToSlice function
	slice := q.ToSlice()
	if len(slice) != 1 || slice[0] != 3 {
		t.Errorf("Expected slice [3], got %v", slice)
	}

	// Test Len function
	if q.Len() != 1 {
		t.Errorf("Expected length 1, got %d", q.Len())
	}

	// Test Size function
	if q.Size() != 1 {
		t.Errorf("Expected size 1, got %d", q.Size())
	}

	// Test Full function
	if q.Full() {
		t.Error("Expected full to be false")
	}

	// Test Empty function
	if q.Empty() {
		t.Error("Expected empty to be false")
	}

	// Test Clear function
	q.Clear()
	if q.Size() != 0 {
		t.Errorf("Expected size 0, got %d", q.Size())
	}
}
