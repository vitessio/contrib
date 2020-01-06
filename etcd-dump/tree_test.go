package main

import (
	"fmt"
	"testing"
)

func TestTree(t *testing.T) {
	tree := NewTree("Test Tree")
	root := tree.setRoot("root")
	level11 := tree.addChild(root, "level1")
	level12 := tree.addChild(root, "level2")
	tree.addChild(level11, "level2_11")
	tree.addChild(level12, "level2_11")
	if len(tree.root.children) != 2 {
		t.Errorf("Want %d, Got %d", 2, len(tree.root.children))
	}
	level := 0
	s := ""
	root.traverse(level, func(name string, level2 int) {
		s = s + fmt.Sprintf("%d:%s", level2, name)
	})
	expected := "0:root1:level12:level2_111:level22:level2_11"
	if s != expected {
		t.Errorf("Want %s, Got %s", expected, s)
	}
}
