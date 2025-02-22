package frame

import "strings"

// 实现前缀树
// treeNode 表示路由树中的一个节点。
// 它包含节点的名称、子节点、路由名称以及一个标志，指示该节点是否是路由的终点。
type treeNode struct {
	name       string
	children   []*treeNode
	routerName string
	isEnd      bool
}

// Put 用于将路由添加到路由树中。
// 它接受一个路径字符串作为输入，将其拆分为多个段，并构建相应的树结构。
// 如果路径的某个段已经存在，则遍历到该节点；如果不存在，则创建一个新节点。
func (t *treeNode) Put(path string) {
	// 将路径插入到树中。该函数通过路径字符串导航或创建树节点，以维护树的结构。
	// 参数:
	// - t: 树的根节点引用，用于开始遍历。
	// - path: 待插入的路径字符串，使用"/"分隔各个节点。
	root := t
	strs := strings.Split(path, "/")
	for index, name := range strs {
		// 跳过空字符串，通常出现在路径开头或连续的"/"之间。
		if index == 0 {
			continue
		}
		children := t.children
		isMatch := false
		// 遍历当前节点的子节点，查找匹配的节点。
		for _, node := range children {
			if node.name == name {
				isMatch = true
				t = node
				break
			}
		}
		// 如果没有找到匹配的节点，则创建新的节点并添加到子节点列表中。
		if !isMatch {
			isEnd := false
			// 判断当前节点是否是路径的终点。
			if index == len(strs)-1 {
				isEnd = true
			}
			// 创建并添加新节点。
			node := &treeNode{name: name, children: make([]*treeNode, 0), isEnd: isEnd}
			children = append(children, node)
			t.children = children
			t = node
		}
	}
	// 遍历完成后，将t重置为根节点，以供下一次插入使用。
	t = root
}

// Get 用于根据路径查询路由树。
// 它将路径拆分为多个段，并搜索匹配的节点。
// 它支持精确匹配、参数匹配（用 ":" 表示）和通配符匹配（用 "*" 表示）。
// 如果找到匹配项，则返回相应的 treeNode 指针；否则返回 nil。
func (t *treeNode) Get(path string) *treeNode {
	// 将路径按"/"分割成字符串数组
	strs := strings.Split(path, "/")
	// 初始化路由器名称
	routerName := ""
	// 遍历路径中的每个部分，从第二个部分开始（第一个部分是空字符串，因为路径以"/"开头）
	for index, name := range strs {
		if index == 0 {
			continue
		}
		// 获取当前节点的子节点
		children := t.children
		// 标记是否找到匹配的子节点
		isMatch := false
		// 遍历所有子节点，寻找匹配的节点
		for _, node := range children {
			// 如果子节点名称与当前路径部分匹配，或者子节点名称为"*"（匹配任何单个路径部分），或者子节点名称包含":"（匹配任何参数）
			if node.name == name || node.name == "*" || strings.Contains(node.name, ":") {
				// 设置匹配标志为true
				isMatch = true
				// 将匹配的节点名称添加到路由器名称中
				routerName += "/" + node.name
				// 更新节点的路由器名称
				node.routerName = routerName
				// 将当前节点设置为匹配的子节点，以便在下一次迭代中继续搜索
				t = node
				// 如果当前路径部分是最后一个部分，返回匹配的节点
				if index == len(strs)-1 {
					return node
				}
				// 找到匹配项后，跳出循环
				break
			}
		}
		// 如果没有找到匹配的子节点
		if !isMatch {
			// 再次遍历所有子节点，寻找名称为"**"的节点（匹配任何路径）
			for _, node := range children {
				if node.name == "**" {
					// 将匹配的节点名称添加到路由器名称中
					routerName += "/" + node.name
					// 更新节点的路由器名称
					node.routerName = routerName
					// 返回匹配的节点
					return node
				}
			}
		}
	}
	// 如果没有找到任何匹配的节点，返回nil
	return nil
}
