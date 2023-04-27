package Gee

import (
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"net/http"
	"reflect"
	"strings"
	"sync"
)

var ErrNullData = errors.New("obj data cant nil")

const defaultMemory = 32 << 20

type SliceValidationError []error

// Error concatenates all error elements in SliceValidationError into a single string separated by \n.
func (err SliceValidationError) Error() string {
	n := len(err)
	switch n {
	case 0:
		return ""
	default:
		var b strings.Builder
		if err[0] != nil {
			fmt.Fprintf(&b, "[%d]: %s", 0, err[0].Error())
		}
		if n > 1 {
			for i := 1; i < n; i++ {
				if err[i] != nil {
					b.WriteString("\n")
					fmt.Fprintf(&b, "[%d]: %s", i, err[i].Error())
				}
			}
		}
		return b.String()
	}
}

type Validator struct {
	*validator.Validate
}

func (v *Validator) ShouldBindForm(obj any, req *http.Request) error {
	if err := req.ParseForm(); err != nil {
		return err
	}
	if err := req.ParseMultipartForm(defaultMemory); err != nil && !errors.Is(err, http.ErrNotMultipart) {
		return err
	}
	if err := Bind(req, obj); err != nil {
		return err
	}
	return v.validate(obj)
}

// ValidateStruct receives any kind of type, but only performed struct or pointer to struct type.
func (v *Validator) validate(obj any) error {
	value := reflect.ValueOf(obj)
	switch value.Kind() {
	case reflect.Ptr:
		return v.validate(value.Elem().Interface())
	case reflect.Struct:
		return v.Struct(obj)
	case reflect.Slice, reflect.Array:
		count := value.Len()
		validateRet := make(SliceValidationError, 0)
		for i := 0; i < count; i++ {
			if err := v.validate(value.Index(i).Interface()); err != nil {
				validateRet = append(validateRet, err)
			}
		}
		if len(validateRet) == 0 {
			return nil
		}
		return validateRet
	default:
		return nil
	}
}

// 懒汉式单例
var validate sync.Once

func (v *Validator) lazyInit() *Validator {
	validate.Do(func() {
		v.Validate = validator.New()
		v.setTagName()
	})
	return v
}

// setTagName 允许去设置标签名, 将 Validate 改为 binding，为了后边识别.
func (v *Validator) setTagName() {
	v.SetTagName("binding")
}
