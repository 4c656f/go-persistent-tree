package main

import (
	"fmt"
	"strings"
)

type Node[ValueType any] struct {
	value    []ValueType
	children []*Node[ValueType]
}

type PersistentVec[ValueType any] struct {
	// count of elements held by the tree
	cnt uint
	// height of the tree
	shift uint
	// power of 2, used to calculate width of the tree, height and mask
	power uint
	// width of every node
	width uint
	// used to calculate index for current node instead of module operation
	mask uint
	// root of the tree
	root *Node[ValueType]
	// tail node used for optimization of average runtime, average 0(1) operations
	tail *Node[ValueType]
}

// used to fill nullable root after popping
func newNullNode[ValueType any](width uint) *Node[ValueType] {
	return &Node[ValueType]{
		value:    []ValueType{},
		children: make([]*Node[ValueType], width),
	}
}

func NewPersistentVec[ValueType any](values []ValueType, power uint) *PersistentVec[ValueType] {
	var width uint = 1 << power
	tailNode := &Node[ValueType]{
		value:    make([]ValueType, 0, width),
		children: make([]*Node[ValueType], width),
	}
	rootNode := &Node[ValueType]{
		children: make([]*Node[ValueType], width),
	}
	vec := &PersistentVec[ValueType]{
		cnt:   0,
		shift: power,
		power: power,
		width: width,
		mask:  width - 1,
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
	return ((vec.cnt - 1) >> vec.power) << vec.power
}

func (vec *PersistentVec[ValueType]) Append(value ...ValueType) *PersistentVec[ValueType] {
	out := vec

	for _, v := range value {
		out = out.Clone()
		// we have space in the tail
		if out.cnt-out.tailOffset() < out.width {

			newTailNode := &Node[ValueType]{
				value: make([]ValueType, len(out.tail.value), out.width),
			}
			// copy previous tail values to new tail node
			copy(newTailNode.value, out.tail.value)
			// add value to the last available index
			newTailNode.value = append(newTailNode.value, v)
			// return new instance of persisted vec with new cloned tail node
			out.cnt++
			out.tail = newTailNode
			continue
		}
		// we need to create new tail node
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
		if (out.cnt >> out.power) > (1 << out.shift) {
			// init new node for root
			newRoot = &Node[ValueType]{
				children: make([]*Node[ValueType], out.width),
			}
			// set first children as previous tree
			newRoot.children[0] = out.root
			// init new branch with correct height
			newRoot.children[1] = out.newPath(out.shift, nodeToInsertInTree)
			// grow capacity of the tree
			newShift += out.power
		} else {
			// we don't need to grow height of the tree
			// so just push tail somewhere in the tree and it will be our new root
			newRoot = out.pushTailNodeToTree(out.shift, out.root, nodeToInsertInTree)
		}
		// create new instance of vec with new instance of root and ceated branches
		out.cnt++
		out.shift = newShift
		out.root = newRoot
		out.tail = newTailNode
	}
	return out
}

func (vec *PersistentVec[ValueType]) Pop() (ValueType, *PersistentVec[ValueType]) {
	if vec.cnt == 0 {
		panic("Pop from empty tree")
	}
	newVec := vec.Clone()
	// tail contains some values, simply pop from it and create new instance
	if vec.cnt-vec.tailOffset() > 1 {
		newTail := vec.tail.Clone(vec)
		popped := newTail.value[len(newTail.value)-1]
		newTail.value = newTail.value[:len(newTail.value)-1]
		newVec.tail = newTail
		newVec.cnt--
		// return popped element from the tail and new instance of vec with new tail node
		return popped, newVec
	}
	// tail does contains one value, we need to pop it and set most right node as new tail node
	popped := vec.tail.value[len(vec.tail.value)-1]
	newRoot, newTail := vec.popNodeFromTreeToTail(vec.shift, vec.root)
	newVec.cnt--
	newVec.root = newRoot
	if newVec.root == nil {
		newVec.root = newNullNode[ValueType](vec.width)
	}
	newVec.tail = newTail

	if newVec.shift > vec.power && newRoot.children[1] == nil {
		newVec.shift -= vec.power
		newVec.root = newVec.root.children[0]
	}

	return popped, newVec
}

// will recursively goes to most right node and return leaf and new root
func (vec *PersistentVec[ValueType]) popNodeFromTreeToTail(level uint, current *Node[ValueType]) (newRoot *Node[ValueType], newTail *Node[ValueType]) {
	// we reach leaf node that will be our new tail node, clone it and return to previous iter of recursion
	if level == 0 {
		return nil, current.Clone(vec)
	}
	// calculate most right idx that we have in child array
	idx := ((vec.cnt - 2) >> level) & vec.mask
	// clone the node for persistants
	out := current.Clone(vec)
	// go to next level
	nxt, tail := vec.popNodeFromTreeToTail(level-vec.power, current.children[idx])
	// that means that child is no longer contains any nodes, so it's seted to nil, and in current elemets we to won't have any elements
	if nxt == nil && idx == 0 {
		// in that case we need to remove current node, because it no longer have any elements
		return nil, tail
	}
	// else set in cloned node need to pop node as nil, or as new instance of it
	out.children[idx] = nxt
	return out, tail
}

func (vec *PersistentVec[ValueType]) pushTailNodeToTree(level uint, parent *Node[ValueType], tailNodeToPush *Node[ValueType]) *Node[ValueType] {
	// create new instance of node to keep persitents
	out := parent.Clone(vec)

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
	// current cloned parent node is a leaf node, we can insert leaf node to it
	if level == vec.power {
		nodeToInser = tailNodeToPush
	} else {
		// current node is not a leaf, so we need to continue cloning path

		// next children
		child := out.children[idx]
		if child != nil {
			// we have children, clone it
			nodeToInser = vec.pushTailNodeToTree(level-vec.power, child, tailNodeToPush)
		} else {
			// we don't have needed branch, create it
			nodeToInser = vec.newPath(level-vec.power, tailNodeToPush)
		}
	}
	out.children[idx] = nodeToInser
	return out
}

// create branch from defined level in tree with correct height
func (vec *PersistentVec[ValueType]) newPath(level uint, node *Node[ValueType]) *Node[ValueType] {
	// we reach needed height, just return insert node
	if level == 0 {
		return node
	}
	// create new node that will be in new branch
	out := &Node[ValueType]{
		children: make([]*Node[ValueType], vec.width),
	}

	// we chose always most-left path couse we in a newly created path
	out.children[0] = vec.newPath(level-vec.power, node)

	return out
}

func (vec *PersistentVec[ValueType]) Set(index uint, value ValueType) *PersistentVec[ValueType] {
	if index < 0 || index >= vec.cnt {
		return vec
	}
	tailOffset := vec.tailOffset()
	newVec := vec.Clone()
	// index that we need to set is in tail
	if index >= tailOffset {
		// create new instance of the tail node
		newTail := vec.tail.Clone(vec)
		// set index to new value in new instance of the tail
		//
		newTail.value[index&vec.mask] = value
		newVec.tail = newTail
		return newVec
	}
	newVec.root = vec.cloneIdxPath(index, value)
	return newVec
}

// clones all fields of the node
func (node *Node[ValueType]) Clone(vec *PersistentVec[ValueType]) *Node[ValueType] {
	values := make([]ValueType, len(node.value), vec.width)
	children := make([]*Node[ValueType], vec.width)
	copy(values, node.value)
	copy(children, node.children)
	return &Node[ValueType]{
		value:    values,
		children: children,
	}
}

// clones all fields of the vectree
func (vec *PersistentVec[ValueType]) Clone() *PersistentVec[ValueType] {
	return &PersistentVec[ValueType]{
		cnt:   vec.cnt,
		shift: vec.shift,
		power: vec.power,
		width: vec.width,
		mask:  vec.mask,
		root:  vec.root,
		tail:  vec.tail,
	}
}

// used to perform set operation
func (vec *PersistentVec[ValueType]) cloneIdxPath(index uint, value ValueType) *Node[ValueType] {
	node := vec.root.Clone(vec)
	// save root to return it
	newRoot := node
	for level := vec.shift; level > 0; level -= vec.power {
		nxtCloned := node.children[(index>>level)&vec.mask].Clone(vec)
		node.children[(index>>level)&vec.mask] = nxtCloned
		node = nxtCloned
	}
	node.value[index&vec.mask] = value
	// new root will contain newly created branch with seted value
	return newRoot
}

func (vec *PersistentVec[ValueType]) sliceFor(index uint) []ValueType {
	if index >= vec.tailOffset() {
		return vec.tail.value
	}

	node := vec.root

	for level := vec.shift; level > 0; level -= vec.power {
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

func (vec *PersistentVec[ValueType]) String() string {
	var result strings.Builder
	result.WriteString(fmt.Sprintf("Cnt: %v\n", vec.cnt))
	result.WriteString(fmt.Sprintf("Shift: %v\n", vec.shift))
	result.WriteString(fmt.Sprintf("Tail: %v\n", vec.tail.value))
	result.WriteString(nodeToString(vec.root, 0))
	return result.String()
}

func nodeToString[ValueType any](node *Node[ValueType], level int) string {
	var result strings.Builder
	indent := strings.Repeat("  ", level)

	if node == nil {
		return indent + "<nil>\n"
	}

	result.WriteString(fmt.Sprintf("%sNode (Level %d, Value: %v) {\n", indent, level, node.value))
	for i, child := range node.children {
		result.WriteString(fmt.Sprintf("%s  Child %d: ", indent, i))
		result.WriteString(nodeToString(child, level+1))
	}
	result.WriteString(indent + "}\n")

	return result.String()
}
