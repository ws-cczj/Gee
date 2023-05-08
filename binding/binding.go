package binding

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

// Binding describes the interface which needs to be implemented for binding the
// data present in the request such as JSON request body, query parameters or
// the form POST.
type Binding interface {
	Name() string
	Bind(*http.Request, any) error
}

// Default returns the appropriate Binding instance based on the HTTP method
// and the content type.
func Default(method, contentType string) Binding {
	if method == http.MethodGet {
		return formBinding{}
	}

	switch contentType {
	case "json":
		return jsonBinding{}
	default: // case MIMEPOSTForm:
		return formBinding{}
	}
}

type Validator struct {
	*validator.Validate
}

func validate(obj any) error {
	if validatorTol == nil {
		return nil
	}
	return validatorTol.lazyInit().validate(obj)
}

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

// ValidateStruct receives any kind of type,
// but only performed struct or pointer to struct type.
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
var validateOnce sync.Once
var validatorTol *Validator

// ValidatorTol 只允许调用，不允许修改
func ValidatorTol() *Validator {
	if validatorTol == nil {
		validatorTol = new(Validator)
	}
	return validatorTol.lazyInit()
}

// LazyInit 懒加载
func (v *Validator) lazyInit() *Validator {
	validateOnce.Do(func() {
		v.Validate = validator.New()
		v.setTagName()
	})
	return v
}

// setTagName 允许去设置标签名, 将 Validate 改为 binding，为了后边识别.
func (v *Validator) setTagName() {
	v.SetTagName("binding")
}
