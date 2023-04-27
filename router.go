package main

import (
	"net/http"
	"strings"
)

type router struct {
	roots    map[string]*node
	handlers map[string][]HandlerFunc
}

func newRouter() *router {
	return &router{
		roots:    make(map[string]*node),
		handlers: make(map[string][]HandlerFunc),
	}
}

// parseRoute 自定义分割函数 结尾使用 / 进行分割
// 避免提前分割为字符串的操作，使用 pre 存储上一个 分隔符的位置
// 如果遇到下一个分隔符，先将字符串追加然后判断字符串开头是否为 *
func parseRoute(pattern string) []string {
	res := make([]string, 0)
	pattern += "/"
	n, pre := len(pattern), -1
	for i := 0; i < n; i++ {
		if pattern[i] == '/' {
			if pre != -1 && pre != i {
				res = append(res, pattern[pre:i])
			}
			if pre != -1 && pattern[pre] == '*' {
				break
			}
			pre = i + 1
		}
	}
	return res
}

// addRoute 添加路由
// 1. 解析路径为多个路线
// 2. 判断结点是否开启
// 3. 插入路由树，进行构建。存储 handler 函数
func (r *router) addRoute(method, pattern string, handlers ...HandlerFunc) {
	parts := parseRoute(pattern)
	key := method + "-" + pattern
	if _, ok := r.roots[method]; !ok {
		r.roots[method] = new(node)
	}
	r.roots[method].insert(pattern, parts, 0)
	r.handlers[key] = append(r.handlers[key], handlers...)
}

// findRouter 查找路由路线
func (r *router) findRouter(method, pattern string) (n *node, mapper map[string]string) {
	root, ok := r.roots[method]
	if !ok {
		return
	}

	mapper = make(map[string]string)
	real_parts := parseRoute(pattern)

	if n = root.search(real_parts, 0); n != nil {
		virtual_parts := parseRoute(n.pattern)
		for i, part := range virtual_parts {
			if part[0] == ':' {
				mapper[part[1:]] = real_parts[i]
			}
			if part[0] == '*' && len(part) > 1 {
				mapper[part[1:]] = strings.Join(real_parts[i:], "/")
				break
			}
		}
	}
	return
}

// handle 请求处理
func (r *router) handle(ctx *Context) {
	n, mapper := r.findRouter(ctx.Method, ctx.Path)
	if n != nil {
		ctx.Params = mapper
		key := ctx.Method + "-" + n.pattern
		ctx.handlers = append(ctx.handlers, r.handlers[key]...)
	} else {
		ctx.handlers = append(ctx.handlers, func(context *Context) {
			http.NotFound(ctx.Writer, ctx.Req)
		})
	}
	ctx.Next()
}
