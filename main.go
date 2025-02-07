package main

type Node[ValueType any] struct {
	value    []ValueType
	children []*Node[ValueType]
}

type PersistentVec[ValueType any] struct {
	cnt   uint
	shift uint
	width uint
	mask  uint
	root  *Node[ValueType]
	tail  *Node[ValueType]
}

func NewPersistentVec[ValueType any](values []ValueType) *PersistentVec[ValueType] {
	vec := &PersistentVec[ValueType]{
		cnt:   0,
		shift: 5,
		width: 32,
		mask:  31,
		root:  nil,
	}
	for _, value := range values {
		vec = vec.Append(value)
	}
	return vec
}

// returns the offset of the tail portion in the vector
//
// example: cnt = 33, shift = 5, width = 32
// tree structure:
// [filled, filled, filled ...] [filled, empty, empty...]
//
// calculation:
// 33 - 1 = 32: 100000
// 100000 >> 5: 000001
// 000001 << 5: 100000 -> 32
// 32 points to where the tail starts
func (vec *PersistentVec[ValueType]) tailOffset() uint {
	if vec.cnt < vec.width {
		return 0
	}
	return ((vec.cnt - 1) >> vec.shift) << vec.shift
}

func (vec *PersistentVec[ValueType]) Append(value ...ValueType) *PersistentVec[ValueType] {
	out := vec
	i := vec.cnt

	for _, v := range value {
		// we have space in the tail
		newTailNode := &Node[ValueType]{
			value: make([]ValueType, vec.width),
		}
		copy(newTailNode.value, vec.tail.value)
		if i-vec.tailOffset() < vec.width {
			newTailNode.value[i&vec.mask] = v
			out = &PersistentVec[ValueType]{
				cnt:   vec.cnt + 1,
				shift: vec.shift,
				width: vec.width,
				mask:  vec.mask,
				root:  vec.root,
				tail:  newTailNode,
			}
			continue
		}
		nodeToInsertInTree := newTailNode
		newTailNode = &Node[ValueType]{
			value: make([]ValueType, vec.width),
		}
		newTailNode.value[0] = v
		out = &PersistentVec[ValueType]{
			cnt:   vec.cnt + 1,
			shift: vec.shift,
			width: vec.width,
			mask:  vec.mask,
			root:  vec.root,
			tail:  newTailNode,
		}
	}
	return out
}

func (vec *PersistentVec[ValueType]) pushTailNodeToTree(level int, parent *Node[ValueType], tailNodeToPush *Node[ValueType]) {

}
