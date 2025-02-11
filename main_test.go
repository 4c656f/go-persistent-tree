package main

import (
	"fmt"
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
}
