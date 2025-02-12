package main

import (
	"reflect"
	"testing"
)

func TestMain(t *testing.T) {
	// Test cases
	t.Run("Should correctly convert vector to slice", func(t *testing.T) {
		vec := NewPersistentVec([]int{1, 2, 3}, 5)
		sliceFromTree := vec.ToGenericVec()
		if !reflect.DeepEqual(sliceFromTree, []int{1, 2, 3}) {
			t.Errorf("got %v, want %v", sliceFromTree, []int{1, 2, 3})
		}
	})

	t.Run("Should correctly append to tree", func(t *testing.T) {
		vec := NewPersistentVec([]int{1, 2, 3}, 5)
		vec = vec.Append(4, 5, 6)
		vecFromTree := vec.ToGenericVec()
		if !reflect.DeepEqual(vecFromTree, []int{1, 2, 3, 4, 5, 6}) {
			t.Errorf("got %v, want %v", vecFromTree, []int{1, 2, 3, 4, 5, 6})
		}
	})

	t.Run("Should correctly append and grow tree", func(t *testing.T) {
		vec := NewPersistentVec([]int{1, 2, 3}, 1)
		vec = vec.Append(4, 5, 6)
		vecFromTree := vec.ToGenericVec()
		if !reflect.DeepEqual(vecFromTree, []int{1, 2, 3, 4, 5, 6}) {
			t.Errorf("got %v, want %v", vecFromTree, []int{1, 2, 3, 4, 5, 6})
		}
	})

	t.Run("Should correctly set index in tail", func(t *testing.T) {
		vec := NewPersistentVec([]int{1, 2, 3}, 1)
		vec = vec.Set(2, 4)
		vecFromTree := vec.ToGenericVec()
		if !reflect.DeepEqual(vecFromTree, []int{1, 2, 4}) {
			t.Errorf("got %v, want %v", vecFromTree, []int{1, 2, 4})
		}
	})

	t.Run("Should correctly set index left node inside tree", func(t *testing.T) {
		vec := NewPersistentVec([]int{1, 2, 3}, 1)
		vec = vec.Set(0, 4)
		vecFromTree := vec.ToGenericVec()
		if !reflect.DeepEqual(vecFromTree, []int{4, 2, 3}) {
			t.Errorf("got %v, want %v", vecFromTree, []int{4, 2, 3})
		}
	})

	t.Run("Should correctly set index right node inside tree", func(t *testing.T) {
		vec := NewPersistentVec([]int{1, 2, 3, 4, 5, 6}, 1)
		vec = vec.Set(2, 4)
		vecFromTree := vec.ToGenericVec()
		if !reflect.DeepEqual(vecFromTree, []int{1, 2, 4, 4, 5, 6}) {
			t.Errorf("got %v, want %v", vecFromTree, []int{1, 2, 4, 4, 5, 6})
		}
	})

	t.Run("Should correctly pop from tail", func(t *testing.T) {
		vec := NewPersistentVec([]int{1, 2}, 1)
		value, newVec := vec.Pop()
		vecFromTree := newVec.ToGenericVec()
		if value != 2 {
			t.Errorf("got %v, want %v", value, 2)
		}
		if !reflect.DeepEqual(vecFromTree, []int{1}) {
			t.Errorf("got %v, want %v", vecFromTree, []int{1})
		}
	})

	t.Run("Should correctly pop from tree", func(t *testing.T) {
		vec := NewPersistentVec([]int{1, 2, 3}, 1)
		value, newVec := vec.Pop()
		vecFromTree := newVec.ToGenericVec()
		if value != 3 {
			t.Errorf("got %v, want %v", value, 3)
		}
		if !reflect.DeepEqual(vecFromTree, []int{1, 2}) {
			t.Errorf("got %v, want %v", vecFromTree, []int{1, 2})
		}
	})

	t.Run("Should correctly pop from tree and shrink tree height", func(t *testing.T) {
		vec := NewPersistentVec([]int{1, 2, 3, 4, 5}, 1)
		_, vec = vec.Pop()
		_, vec = vec.Pop()
		value, vec := vec.Pop()
		vecFromTree := vec.ToGenericVec()
		if value != 3 {
			t.Errorf("got %v, want %v", value, 3)
		}
		if !reflect.DeepEqual(vecFromTree, []int{1, 2}) {
			t.Errorf("got %v, want %v", vecFromTree, []int{1, 2})
		}
	})

	t.Run("Should restore tree after popping an elems", func(t *testing.T) {
		vec := NewPersistentVec([]int{1, 2, 3, 4, 5}, 1)
		_, vec = vec.Pop()
		_, vec = vec.Pop()
		_, vec = vec.Pop()
		vec = vec.Append(1, 2, 3)
		vecFromTree := vec.ToGenericVec()
		if !reflect.DeepEqual(vecFromTree, []int{1, 2, 1, 2, 3}) {
			t.Errorf("got %v, want %v", vecFromTree, []int{1, 2, 1, 2, 3})
		}
	})

	t.Run("Should correctly impliment iterators", func(t *testing.T) {
		values := []int{1, 2, 3, 4, 5}
		vec := NewPersistentVec(values, 1)
		fromIterIndeses := []int{}
		fromIter := []int{}
		for idx, value := range vec.All() {
			fromIterIndeses = append(fromIterIndeses, idx)
			fromIter = append(fromIter, value)
		}
		if !reflect.DeepEqual(fromIterIndeses, []int{0, 1, 2, 3, 4}) {
			t.Errorf("got %v, want %v", fromIterIndeses, []int{0, 1, 2, 3, 4})
		}
		if !reflect.DeepEqual(fromIter, values) {
			t.Errorf("got %v, want %v", fromIter, values)
		}
	})
}
