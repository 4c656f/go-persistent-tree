package main

import "fmt"

type Node[ValueType any] struct {
	value    []ValueType
	children []*Node[ValueType]
}

type PersistentVec[ValueType any] struct {
	cnt   uint
	shift uint
	bits  uint
	width uint
	mask  uint
	root  *Node[ValueType]
	tail  *Node[ValueType]
}

func NewPersistentVec[ValueType any](values []ValueType, power uint) *PersistentVec[ValueType] {
	var width uint = 1 << power
	tailNode := &Node[ValueType]{
		value: make([]ValueType, 0, width),
	}
	rootNode := &Node[ValueType]{
		children: make([]*Node[ValueType], width),
	}
	vec := &PersistentVec[ValueType]{
		cnt:   0,
		shift: power,
		bits:  power,
		width: width,
		mask:  (1 << power) - 1,
		root:  rootNode,
		tail:  tailNode,
	}
	vec = vec.Append(values...)
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
	return ((vec.cnt - 1) >> vec.bits) << vec.bits
}

func (vec *PersistentVec[ValueType]) Append(value ...ValueType) *PersistentVec[ValueType] {
	out := vec

	for _, v := range value {

		// we have space in the tail
		if out.cnt-out.tailOffset() < out.width {

			newTailNode := &Node[ValueType]{
				value: make([]ValueType, len(out.tail.value), out.width),
			}
			// copy previus tail values to new tail node
			copy(newTailNode.value, out.tail.value)
			// add value to the last avalible index
			newTailNode.value = append(newTailNode.value, v)
			// return new instance of persisted vec with new cloned tail node
			out = &PersistentVec[ValueType]{
				cnt:   out.cnt + 1,
				shift: out.shift,
				bits:  out.bits,
				width: out.width,
				mask:  out.mask,
				root:  out.root,
				tail:  newTailNode,
			}
			continue
		}
		// we need create new tail node
		newTailNode := &Node[ValueType]{
			value: make([]ValueType, 0, out.width),
		}
		// we dont have space in tail, then we need to move current tail inside tree
		nodeToInsertInTree := out.tail
		// init variable, so when we hit root overflow we can increment and assign proper shift to new instance
		newShift := out.shift
		// new instance of the root node that we will use in new instance of PerVec
		var newRoot *Node[ValueType]
		// as we move current tail to tree, we need new tail, so assign value to first position in new instance of tail
		newTailNode.value = append(newTailNode.value, v)
		// is root overflow check (when tree needs to grow in height)
		// vec.cnt >> vec.bits will give use number of filled chunks
		// 1 << vec.shift wil give us tree capacity of it's current hegith
		if (out.cnt >> out.bits) > (1 << out.shift) {
			// init new node for root
			newRoot = &Node[ValueType]{
				children: make([]*Node[ValueType], out.width),
			}
			// set first children as previus tree
			newRoot.children[0] = out.root
			// init new brunch with correct height
			newRoot.children[1] = out.newPath(out.shift, nodeToInsertInTree)
			// grow capacity of the tree
			newShift += out.bits
		} else {
			// we don't need to grow height of the tree
			// so just push tail somewhere in the ree and it will be our new root
			newRoot = out.pushTailNodeToTree(out.shift, out.root, nodeToInsertInTree)
		}
		// create new instance of vec with new instance of root and ceated branches
		out = &PersistentVec[ValueType]{
			cnt:   out.cnt + 1,
			bits:  out.bits,
			width: out.width,
			shift: newShift,
			mask:  out.mask,
			root:  newRoot,
			tail:  newTailNode,
		}
	}
	return out
}

func (vec *PersistentVec[ValueType]) pushTailNodeToTree(level uint, parent *Node[ValueType], tailNodeToPush *Node[ValueType]) *Node[ValueType] {
	// create new instance of node to keep peprsitents
	out := &Node[ValueType]{
		children: make([]*Node[ValueType], vec.width),
	}
	// clone previus childrens to new instance of node
	copy(out.children, parent.children)

	// calculate acces idx of next node
	// example: cnt = 67
	// 67 - 1 = 66 : 01000010
	// level: 10   : 00000000 & 011111 = 0
	// level: 5    : 01000010 & 011111 = 2
	// level: 0    : 01000010 & 011111 = 2
	// Level 10:   [Root]
	//              |
	// Level 5:     [Node] (slot 2)
	//              |
	// Level 0:     [Leaf] (slot 2)
	idx := ((vec.cnt - 1) >> level) & vec.mask
	var nodeToInser *Node[ValueType]
	// curent cloned parent node is a leaf node, we can insert leaf node to it
	if level == vec.bits {
		nodeToInser = tailNodeToPush
	} else {
		// curent node is not a leaf, so we need to continue cloning path

		// next children
		child := out.children[idx]
		if child != nil {
			// we have children, clone it
			nodeToInser = vec.pushTailNodeToTree(level-vec.bits, child, tailNodeToPush)
		} else {
			// we don't have needed branch, create it
			nodeToInser = vec.newPath(level-vec.bits, tailNodeToPush)
		}
	}
	out.children[idx] = nodeToInser
	return out
}

// create branch from defined level in tree with correct height
func (vec *PersistentVec[ValueType]) newPath(level uint, node *Node[ValueType]) *Node[ValueType] {
	// we reach neded height, just return insert node
	if level == 0 {
		return node
	}
	// create new node that will be in new branch
	out := &Node[ValueType]{
		children: make([]*Node[ValueType], vec.width),
	}

	// we chose always most-left path couse we in a newly created path
	out.children[0] = vec.newPath(level-vec.bits, node)

	return out
}

func (vec *PersistentVec[ValueType]) sliceFor(index uint) []ValueType {
	if index >= vec.tailOffset() {
		return vec.tail.value
	}

	node := vec.root

	for level := vec.shift; level > 0; level -= vec.bits {
		node = node.children[(index>>level)&vec.mask]
	}

	return node.value
}

func (vec *PersistentVec[ValueType]) ToGenericVec() []ValueType {
	out := make([]ValueType, 0, vec.cnt)

	for i := uint(0); i < vec.cnt; i += vec.width {
		out = append(out, vec.sliceFor(i)...)
	}

	return out
}

func (vec *PersistentVec[ValueType]) PrintTree() {
	fmt.Println("Tail:", vec.tail.value)
	printNode(vec.root, 0)
}

func printNode[ValueType any](node *Node[ValueType], level int) {
	indent := ""
	for i := 0; i < level; i++ {
		indent += "  "
	}

	if node == nil {
		fmt.Println(indent + "<nil>")
		return
	}

	fmt.Printf(indent+"Node (Level %d, Value: %v) {\n", level, node.value)
	for i, child := range node.children {
		fmt.Printf(indent + "  Child %d: ", i)
		printNode(child, level+1)
	}
	fmt.Println(indent + "}")
}
