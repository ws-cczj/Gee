package binding

import (
	"fmt"
	"gee"
	"github.com/go-playground/validator/v10"
	"testing"
)

func Test(t *testing.T) {
	t.Run("TestFormBinding_Bind", func(t *testing.T) {
		TestFormBinding_Bind(t)
	})
	t.Run("TestJsonBinding_Bind", func(t *testing.T) {
		TestJsonBinding_Bind(t)
	})
}

func TestFormBinding_Bind(t *testing.T) {
	// 允许嵌套结构体
	type Student struct {
		Id int `form:"id"`
	}
	type User struct {
		*Student
		Name string `form:"name" binding:"required"`
		Age  int    `form:"age"`
	}
	r := gee.Default()
	r.POST("/go", func(c *gee.Context) {
		u := &User{Student: new(Student)}
		if err := c.ShouldBind(u); err != nil {
			_, ok := err.(validator.ValidationErrors)
			fmt.Println(ok)
			fmt.Println(err)
		}
		fmt.Println(u)
		fmt.Println(u.Student)
	})

	r.Run(":9999")
}

func TestJsonBinding_Bind(t *testing.T) {
	// 允许嵌套结构体
	type Student struct {
		Id int `json:"id"`
	}
	type User struct {
		*Student
		Name string `json:"name" binding:"required"`
		Age  int    `json:"age"`
	}
	r := gee.Default()
	r.POST("/go", func(c *gee.Context) {
		u := &User{Student: new(Student)}
		if err := c.ShouldBind(u); err != nil {
			_, ok := err.(validator.ValidationErrors)
			fmt.Println(ok)
			fmt.Println(err)
		}
		fmt.Println(u)
		fmt.Println(u.Student)
	})

	r.Run(":4399")
}
