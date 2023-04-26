package gee

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
)

const abortLen = 1 << 10

type H map[string]any

type Context struct {
	Writer http.ResponseWriter
	Req    *http.Request

	Params map[string]string // 用于在上下文传递路径参数, 比如 /p/:id -> 可以通过id找到对应的路径值

	Path       string
	Method     string
	StatusCode int

	index    int           // 控制当前处理进度
	handlers []HandlerFunc // 存储当前请求对应的 中间件 和 handler.

	// This mutex protects Keys map.
	mu sync.RWMutex

	// Keys is a key/value pair exclusively for the context of each request.
	Keys map[string]any

	engine *Engine // 存储引擎
}

func newContext(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{
		Writer: w,
		Req:    r,
		Params: map[string]string{},
		Path:   r.URL.Path,
		Method: r.Method,
		index:  -1,
	}
}

// Next 请求处理中枢
func (c *Context) Next() {
	c.index++
	s := len(c.handlers)
	for ; c.index < s; c.index++ {
		c.handlers[c.index](c)
	}
}

// Abort 请求处理终止
func (c *Context) Abort() {
	c.index = abortLen
}

// AbortWithStatus 请求终止并设置相关状态码
func (c *Context) AbortWithStatus(code int) {
	c.Abort()
	c.Status(code)
}

// AbortWithJson 请求终止并写入消息
func (c *Context) AbortWithJson(code int, msg string) {
	c.Abort()
	c.JSON(code, H{"message": msg})
}

// Header Tip: 在调用Header之后会返回一个 Header的请求头。如果先调用WriteHeader，则会使用默认的HTTP请求头和状态码。
// 因此要在WriteHeader设置之前先设置好请求头信息，否则Header不会生效。
// 正确的调用顺序为: Header 响应头-> Status 状态码-> Write 写回
// 如果先写状态码，那么底层会发给客户端 响应头和状态码。
// 如果先写回 Write，那么底层会检查是否有状态码，如果没有就手动修改为200.然后执行写回操作
func (c *Context) Header(key, val string) {
	if val == "" {
		c.Writer.Header().Del(key)
		return
	}
	c.Writer.Header().Set(key, val)
}

// Status 修改状态码
func (c *Context) Status(code int) {
	if c.StatusCode > 0 && c.StatusCode != code {
		_, _ = fmt.Fprintf(os.Stdout, "[WARNING] Headers were already written. Wanted to override status code %d with %d", c.StatusCode, code)
	}
	c.StatusCode = code
	c.Writer.WriteHeader(code)
}

func (c *Context) ClientIP() (addr string) {
	if addr = c.Req.Header.Get("X-Forwarded-For"); addr == "" {
		addr = c.Req.RemoteAddr
	}
	return
}

func (c *Context) String(code int, format string, v ...any) {
	c.Header("Content-Type", "text/plain;charset=utf-8")
	c.Status(code)
	_, _ = c.Writer.Write([]byte(fmt.Sprintf(format, v...)))
}

// JSON 这里无法规避掉错误，因为 Header 和 Status 已经被设置完毕，即使错误
// 也无法修改掉状态码和响应头。因此应该直接 panic
func (c *Context) JSON(code int, obj ...any) {
	c.Header("Content-Type", "application/json;charset=utf-8")
	c.Status(code)
	encoder := json.NewEncoder(c.Writer)
	if err := encoder.Encode(obj); err != nil {
		panic("[GEE] | JSON format err: " + err.Error())
	}
}

// HTML 支持根据模板文件名选择模板进行渲染
func (c *Context) HTML(code int, suffixType string, data any) {
	c.Header("Content-Type", "text/html;charset=utf-8")
	c.Status(code)

	if err := c.engine.htmlTemplates.ExecuteTemplate(c.Writer, suffixType, data); err != nil {
		c.AbortWithJson(http.StatusInternalServerError, err.Error())
	}
}

func (c *Context) Data(code int, data []byte) {
	c.Status(code)
	_, _ = c.Writer.Write(data)
}

func (c *Context) PostForm(key string) string {
	return c.Req.Form.Get(key)
}

func (c *Context) Query(key string) string {
	return c.Req.URL.Query().Get(key)
}

func (c *Context) Param(key string) string {
	return c.Params[key]
}

// Set 通过上下文传递信息
func (c *Context) Set(key string, val interface{}) {
	c.mu.Lock()
	if c.Keys == nil {
		c.Keys = map[string]interface{}{}
	}

	c.Keys[key] = val
	c.mu.Unlock()
}

// Get 通过上下文获取信息
func (c *Context) Get(key string) (val interface{}, exist bool) {
	c.mu.RLock()
	val, exist = c.Keys[key]
	c.mu.RUnlock()
	return
}

// ShouldBind 处理Form表单数据
func (c *Context) ShouldBind(obj interface{}) error {
	if obj == nil {
		return ErrNullData
	}
	return c.Validator().ShouldBindForm(obj, c.Req)
}

func (c *Context) Validator() *Validator {
	return c.engine.validator.lazyInit()
}
