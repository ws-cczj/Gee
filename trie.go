package gee

import "strings"

type node struct {
	isWild   bool    // 精确匹配，是否含有 '*' || ':'
	pattern  string  // 当前结点正在匹配的总路径, 比如: /p/:lang
	part     string  // 当前结点存储的部分路径, 比如: /:lang
	children []*node // 当前结点的所有子节点
}

// matchChild 用于匹配第一个结点，用于插入
func (n *node) matchChild(part string) *node {
	for _, child := range n.children {
		if child.part == part || child.isWild {
			return child
		}
	}
	return nil
}

// matchChildren 用户查找所有匹配的子节点
func (n *node) matchChildren(part string) []*node {
	children := make([]*node, 0)
	for _, child := range n.children {
		if child.part == part || child.isWild {
			children = append(children, child)
		}
	}
	return children
}

// insert 终止条件为，遍历到了 parts 最高层，将调用该函数的结点的 pattern 设置为总路径。
// 表明该处就是整个匹配路径的终点，查找时也按照这个终点进行结果判断。最后逐层返回递归
func (n *node) insert(pattern string, parts []string, height int) {
	if len(parts) == height {
		n.pattern = pattern
		return
	}

	part := parts[height]
	child := n.matchChild(part)
	if child == nil {
		child = &node{part: part, isWild: part[0] == ':' || part[0] == '*'}
		n.children = append(n.children, child)
	}
	child.insert(pattern, parts, height+1)
}

// search 根据 insert 函数的终止条件来看，这个函数只要遍历到最终结果，去判断 pattern 的值是否为 ""即可。
// 如果遍历不到最终结果，或者匹配不合格则返回 nil
func (n *node) search(parts []string, height int) *node {
	if len(parts) == height || strings.HasPrefix(n.part, "*") {
		if n.pattern == "" {
			return nil
		}
		return n
	}

	part := parts[height]
	children := n.matchChildren(part)
	for _, child := range children {
		if childVal := child.search(parts, height+1); childVal != nil {
			return childVal
		}
	}
	return nil
}
