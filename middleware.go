package Gee

import (
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"
)

const (
	green   = "\033[97;42m"
	white   = "\033[90;47m"
	yellow  = "\033[90;43m"
	red     = "\033[97;41m"
	blue    = "\033[97;44m"
	magenta = "\033[97;45m"
	cyan    = "\033[97;46m"
	reset   = "\033[0m"
)

func Logger() HandlerFunc {
	return func(c *Context) {
		start := time.Now()
		c.Next()

		cost := time.Since(start)
		if !c.engine.releaseMode {
			_, _ = fmt.Printf("[GEE] %v |%s %s %s| %s |%13v |%s %3d %s| %s  msg:{ip: %s, user-agent: %s}\n",
				getCurrentTime(),
				getMethodColor(c.Req.Method), c.Req.Method, reset,
				c.Req.URL.Path,
				cost,
				getStatusColor(c.StatusCode), c.StatusCode, reset,
				c.Req.URL.RawQuery,
				c.ClientIP(),
				c.Req.UserAgent())
		}
	}
}

func getCurrentTime() string {
	return time.Now().Format("2006/01/02 - 15:04:05")
}

func getStatusColor(code int) string {
	color := cyan
	// 记录异常信息
	switch {
	case code < 300:
		color = green
	case code < 400:
		color = yellow
	case code < 500:
		color = red
	case code < 600:
		color = red
	default:
	}
	return color
}

func getMethodColor(method string) string {
	color := cyan
	switch method {
	case http.MethodGet:
		color = green
	case http.MethodPost:
		color = blue
	case http.MethodDelete:
		color = red
	case http.MethodPut:
		color = magenta
	default:
	}
	return color
}

// Recover 异常恢复中间件
func Recover() HandlerFunc {
	return func(c *Context) {
		defer func() {
			if err := recover(); err != nil {
				_, _ = fmt.Printf("%s\n\n", trace(fmt.Sprintf("%s", err)))
				c.AbortWithJson(http.StatusInternalServerError, "Internal Server Error")
			}
		}()

		c.Next()
	}
}

// print stack trace for debug
func trace(message string) string {
	var pcs [32]uintptr
	n := runtime.Callers(3, pcs[:]) // skip first 3 caller

	var build strings.Builder
	build.Grow(2 + n)
	build.WriteString(message)
	build.WriteString("\nTraceback:")
	for _, pc := range pcs[:n] {
		fn := runtime.FuncForPC(pc)
		file, line := fn.FileLine(pc)
		build.WriteString(fmt.Sprintf("\n\t%s:%d", file, line))
	}
	return build.String()
}

// Cors 跨域中间件
func Cors() HandlerFunc {
	return func(c *Context) {
		method := c.Req.Method
		header := c.Writer.Header()
		origin := c.Req.Header.Get("Origin")
		// 如果origin为 "" 说明不是跨域，null 不等于 ""
		if origin != "" {
			header.Add("Vary", "Origin")
			// 说明这个请求是一个预请求
			if method == http.MethodOptions {
				reqMethod := header.Get("Access-Control-Request-Method")
				if checkMethod(reqMethod) {
					// 如果方法不对，不提供Access Origin字段。
					c.Abort()
					return
				}
				// 预请求最大缓存时间
				c.Header("Access-Control-Max-Age", "86400")
				commonHeader(c)
				c.AbortWithStatus(http.StatusNoContent)
				return
			}
			commonHeader(c)
		}
		c.Next()
	}
}

func commonHeader(c *Context) {
	// 必须，接受指定域的请求，可以使用*不加以限制，但不安全
	//header.Set("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Origin", c.Req.Header.Get("Origin"))
	// 必须，设置服务器支持的所有跨域请求的方法
	c.Header("Access-Control-Allow-Methods", "POST, GET, PUT, DELETE, OPTIONS")
	// 服务器支持的所有头信息字段，不限于浏览器在"预检"中请求的字段
	c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Token")
	// 可选，设置XMLHttpRequest的响应对象能拿到的额外字段
	c.Header("Access-Control-Expose-Headers", "Access-Control-Allow-Headers, Token")
	// 可选，是否允许后续请求携带认证信息Cookie，该值只能是true，不需要则不设置
	c.Header("Access-Control-Allow-Credentials", "true")
}

// checkMethod 用于校验预请求中方法是否合格的函数
func checkMethod(reqMethod string) bool {
	return reqMethod != http.MethodPut && reqMethod != http.MethodDelete
}
