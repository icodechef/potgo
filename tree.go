package potgo

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	staticNode uint8 = iota // default
	paramNode
	catchAllNode
)

type node struct {
	nType     uint8
	key       string
	path      string
	pKeys     []string
	children  []*node
	pChildren []*node
	wChildren *node // 只有一个 '*'
	regex     *regexp.Regexp
	route     *Route
}

// hasChild 是否存在匹配的子节点
func (n *node) hasChild(key string) bool {
	switch key[0] {
	case ':':
		for _, child := range n.pChildren {
			if child.key == key {
				return true
			}
		}
	case '*':
		return n.wChildren != nil
	default:
		for _, child := range n.children {
			if child.key == key {
				return true
			}
		}
	}

	return false
}

// getChild 返回匹配的子节点
func (n *node) getChild(key string) *node {
	switch key[0] {
	case ':':
		for _, child := range n.pChildren {
			if child.key == key {
				return child
			}
		}
	case '*':
		return n.wChildren
	default:
		for _, child := range n.children {
			if child.key == key {
				return child
			}
		}
	}

	return nil
}

// addRoute 添加路由
func (n *node) addRoute(path string, route *Route) {
	var parts []string
	if len(path) == 0 || path == "/" {
		parts = []string{}
	} else {
		parts = strings.Split(strings.Trim(path, "/"), "/")
	}

	pn := n
	var pKeys []string
	for _, s := range parts {
		nType := staticNode
		var pattern string
		key := s
		if s[0] == '{' && s[len(s)-1] == '}' {
			nType = paramNode
			m := strings.IndexByte(s, ':')
			if m < 0 {
				key = ":" + s[1:len(s)-1]
				pKeys = append(pKeys, s[1:len(s)-1])
			} else {
				key = ":" + s[1:m]
				pKeys = append(pKeys, s[1:m])
				if s[m+1:len(s)-1] == "*" {
					nType = catchAllNode
					key = "*" + s[1:m]
				} else {
					pattern = s[m+1 : len(s)-1]
				}
			}
		}

		if !pn.hasChild(key) {
			child := new(node)
			child.nType = nType
			if pattern != "" {
				child.regex = regexp.MustCompile("^" + pattern + "$")
			}
			child.key = key

			switch child.nType {
			case staticNode:
				pn.children = append(pn.children, child)
			case paramNode:
				pn.pChildren = append(pn.pChildren, child)
			case catchAllNode:
				pn.wChildren = child
			}
		}

		pn = pn.getChild(key)
	}

	pn.pKeys = pKeys
	pn.path = path
	pn.route = route
}

type nodeValue struct {
	handlers []HandlerFunc
	pKeys    []string
}

func (n *node) getRoute(path string, pValues []string) (value nodeValue) {
	r := n.get(path, pValues, 0)
	if r != nil && r.route != nil {
		value.handlers = r.route.handlers
		value.pKeys = r.pKeys
	}
	return
}

func (n *node) get(path string, pValues []string, pIndex uint8) *node {
	if path == "/" {
		if n.route != nil {
			return n
		}
	} else {
		l := len(path)
		end := 1
		for end < l && path[end] != '/' {
			end++
		}

		value := path[1:end]

		if n.children != nil {
			for _, child := range n.children {
				if child.key == value {
					if end < l {
						if found := child.get(path[end:], pValues, pIndex); found != nil {
							return found
						}
					} else {
						return child
					}
				}
			}
		}

		if n.pChildren != nil {
			for _, child := range n.pChildren {
				if child.regex != nil {
					if !child.regex.MatchString(value) {
						continue
					}
				}
				if end < l {
					if found := child.get(path[end:], pValues, pIndex+1); found != nil {
						pValues[pIndex] = value
						return found
					}
				} else if child.route != nil {
					pValues[pIndex] = value
					return child
				}
			}
		}
	}

	if n.wChildren != nil {
		pValues[pIndex] = path[1:]
		return n.wChildren
	}

	return nil
}

// print 打印树
func (n *node) print(level int) string {
	s := fmt.Sprintf("%v{key: %v, child: %v, path: %v, regex: %v, route: %v, nType: %v}\n",
		strings.Repeat(" ", level<<2), n.key, len(n.children), n.path, n.regex, n.route, n.nType)
	for _, child := range n.children {
		if child != nil {
			s += child.print(level + 1)
		}
	}
	for _, child := range n.pChildren {
		if child != nil {
			s += child.print(level + 1)
		}
	}
	if n.wChildren != nil {
		s += n.wChildren.print(level + 1)
	}
	return s
}
