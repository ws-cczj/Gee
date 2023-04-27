package gee

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"net/http"
	"testing"
)

func TestBind(t *testing.T) {
	// 允许嵌套结构体
	type Student struct {
		Id int `form:"id"`
	}
	type User struct {
		*Student
		Name string `form:"name" binding:"required"`
		Age  int    `form:"age"`
	}
	r := Default()
	r.GET("/", func(c *Context) {
		c.String(http.StatusOK, "Hello Geektutu\n")
	})
	r.POST("/go", func(c *Context) {
		u := &User{Student: new(Student)}
		if err := c.ShouldBind(u); err != nil {
			_, ok := err.(validator.ValidationErrors)
			fmt.Println(ok)
			fmt.Println(err)
		}
		fmt.Println(u)
		fmt.Println(u.Student)
	})
	r.PUT("/go", func(c *Context) {
		u := &User{}
		if err := c.ShouldBind(u); err != nil {
			_, ok := err.(validator.ValidationErrors)
			fmt.Println(ok)
			fmt.Println(err)
		}
		fmt.Println(u)
	})
	r.DELETE("/go", func(c *Context) {
		u := &User{}
		if err := c.ShouldBind(u); err != nil {
			_, ok := err.(validator.ValidationErrors)
			fmt.Println(ok)
			fmt.Println(err)
		}
		fmt.Println(u)
	})
	// index out of range for testing Recovery()
	r.GET("/panic", func(c *Context) {
		names := []string{"geektutu"}
		c.String(http.StatusOK, names[100])
	})

	r.Run(":9999")
}
