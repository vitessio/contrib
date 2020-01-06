package main

type Tree struct {
	name string
	root *Node
}

type Node struct {
	name string
	children []*Node
}

func NewTree(name string) *Tree {
	return &Tree{name: name, root: nil}
}

func (tree *Tree) setRoot(name string) *Node {
	tree.root = NewNode(name)
	return tree.root
}

func (tree *Tree) addChild(parentNode *Node, child string) *Node {
	for _, childNode := range parentNode.children {
		if childNode.name == child {
			return childNode
		}
	}
	childNode := NewNode(child)
	parentNode.children = append(parentNode.children, childNode)
	return childNode
}

func NewNode(name string) *Node {
	return &Node{name:name,children:make([]*Node,0)}
}

func (node *Node) traverse(level int, callback func(name string, level int)) {
	callback(node.name, level)
	for _, childNode := range node.children {
		childNode.traverse(level+1, callback)
	}
}

