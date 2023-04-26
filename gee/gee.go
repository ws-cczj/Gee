package gee

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

type HandlerFunc func(*Context)

type Engine struct {
	*RouterGroup // 引擎本身作为根路由组
	router       *router
	groups       []*RouterGroup

	releaseMode bool // 是否为发行版本
	exitOp      bool // 是否开启优雅关机

	htmlTemplates *template.Template // 静态模板
	funcMap       template.FuncMap

	validator *Validator // 绑定校验
}

// ServeHTTP 实现Handler接口，底层进行HTTP服务解析。
func (engine *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := newContext(w, r)
	c.engine = engine
	for _, group := range engine.groups {
		if strings.HasPrefix(r.URL.Path, group.prefix) {
			c.handlers = append(c.handlers, group.middlewares...)
		}
	}
	engine.router.handle(c)
}

// Run 在addr开启监听
func (engine *Engine) Run(addr string) (err error) {
	srv := &http.Server{
		Addr:    addr,
		Handler: engine,
	}
	// 是否开启优雅关机
	if engine.exitOp {
		go func() {
			if err = srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				if !engine.releaseMode {
					_, _ = fmt.Fprintf(os.Stdout, "[GEE] listen is fail! err: %v\n", err)
				}
			}
		}()

		exit := make(chan os.Signal, 1)
		signal.Notify(exit, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
		<-exit

		ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
		err = srv.Shutdown(ctx)
		cancelFunc()
	} else {
		err = srv.ListenAndServe()
	}
	return
}

// SetFuncMap 将所有的模板加载进内存
func (engine *Engine) SetFuncMap(funcMap template.FuncMap) {
	engine.funcMap = funcMap
}

// LoadHTMLGlob 所有的自定义模板渲染函数
func (engine *Engine) LoadHTMLGlob(pattern string) {
	engine.htmlTemplates = template.Must(template.New("").Funcs(engine.funcMap).ParseGlob(pattern))
}

// New 默认配置
func New() *Engine {
	engine := &Engine{router: newRouter()}
	engine.RouterGroup = &RouterGroup{engine: engine}
	engine.groups = []*RouterGroup{engine.RouterGroup}
	engine.validator = new(Validator)
	return engine
}

type IEngine interface {
	Apply(*Engine)
}

type setupEngine struct {
	f func(*Engine)
}

var _ IEngine = &setupEngine{}

func (set *setupEngine) Apply(e *Engine) {
	set.f(e)
}

// WithExitOp 优雅关机枢纽
func WithExitOp(exitop bool) IEngine {
	return newSetupEngine(func(engine *Engine) {
		engine.exitOp = exitop
	})
}

// WithReleaseMode 开启发行版本枢纽
func WithReleaseMode(release bool) IEngine {
	return newSetupEngine(func(engine *Engine) {
		engine.releaseMode = release
	})
}

// WithMiddlewares 自定义全局中间件枢纽
func WithMiddlewares(middlewares ...HandlerFunc) IEngine {
	return newSetupEngine(func(engine *Engine) {
		engine.middlewares = append(engine.middlewares, middlewares...)
	})
}

func newSetupEngine(f func(*Engine)) *setupEngine {
	return &setupEngine{f: f}
}

// Default 自动化配置
func Default(ies ...IEngine) *Engine {
	engine := New()

	for _, ie := range ies {
		ie.Apply(engine)
	}
	// 如果没有指定中间件，就采用默认中间件
	if len(engine.middlewares) == 0 {
		engine.Use(Cors(), Logger(), Recover())
	}
	return engine
}
