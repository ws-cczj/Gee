package main

import (
	"fmt"
	"net/http"
	"path"
)

type RouterGroup struct {
	prefix      string        // 路由组的前缀
	middlewares []HandlerFunc // 路由组的中间件, 因为Engine作为根路由组，因此可以在启动时自带一些中间件
	engine      *Engine       // 公用引擎
}

// Use 通过当前调用的路由分组，添加中间件到其分组中间件切片中
// 然后通过 ctx 统一进行处理
func (rg *RouterGroup) Use(middlewares ...HandlerFunc) {
	rg.middlewares = append(rg.middlewares, middlewares...)
}

func (rg *RouterGroup) addRoute(method, comp string, handlers ...HandlerFunc) {
	pattern := rg.prefix + comp
	if !rg.engine.releaseMode {
		_, _ = fmt.Printf("[GEE] %v |Route %4s-%s\n", getCurrentTime(), method, pattern)
	}
	rg.engine.router.addRoute(method, pattern, handlers...)
}

func (rg *RouterGroup) GET(pattern string, handlers ...HandlerFunc) {
	rg.addRoute(http.MethodGet, pattern, handlers...)
}

func (rg *RouterGroup) POST(pattern string, handlers ...HandlerFunc) {
	rg.addRoute(http.MethodPost, pattern, handlers...)
}

func (rg *RouterGroup) PUT(pattern string, handlers ...HandlerFunc) {
	rg.addRoute(http.MethodPut, pattern, handlers...)
}

func (rg *RouterGroup) DELETE(pattern string, handlers ...HandlerFunc) {
	rg.addRoute(http.MethodDelete, pattern, handlers...)
}

// createStaticHandler 根据相对路径和当前文件系统生成对应的处理句柄 handler
func (rg *RouterGroup) createStaticHandler(relativePath string, fs http.FileSystem) HandlerFunc {
	absolutePath := path.Join(relativePath, rg.prefix)
	fileServer := http.StripPrefix(absolutePath, http.FileServer(fs))

	return func(c *Context) {
		file := c.Param("filepath")
		// 检查, 如果该文件无法打开就直接返回, 否则继续
		if _, err := fs.Open(file); err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		fileServer.ServeHTTP(c.Writer, c.Req)
	}
}

// Static 根据相对路径和根路径，建立起来一个监听服务，用于监听相对路径下静态文件，
// 并将这些文件进行返回的句柄 handler
func (rg *RouterGroup) Static(relativePath, root string) {
	handler := rg.createStaticHandler(relativePath, http.Dir(root))

	urlPattern := path.Join(relativePath, "/*filepath")
	rg.GET(urlPattern, handler)
}

func (rg *RouterGroup) Group(prefix string) *RouterGroup {
	engine := rg.engine
	newGroup := &RouterGroup{
		prefix: rg.prefix + prefix,
		engine: engine,
	}
	engine.groups = append(engine.groups, newGroup)
	return newGroup
}
