package gee

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"
)

// Bind 通过反射来绑定参数
func Bind(req *http.Request, obj interface{}) error {
	// 获取结构体类型和值
	objType := reflect.TypeOf(obj).Elem()
	objValue := reflect.ValueOf(obj).Elem()

	// 遍历结构体的字段
	for i := 0; i < objType.NumField(); i++ {
		field := objType.Field(i)
		tag := field.Tag.Get("form")

		// 如果tag为 '-' 或者 该字段为私有字段，则跳过该字段
		if tag == "-" || !field.IsExported() {
			continue
		}

		// 将参数值转换为字段类型
		fieldType := field.Type
		fieldValue := objValue.Field(i)
		// 从req.Form中获取参数值 如果获取不到就跳过
		values, ok := req.Form[tag]
		if !ok {
			// 处理嵌套结构体 使用一个新的结构体来进入递归。
			// 将新结构体中的字段 赋值 给旧结构体。
			if fieldType.Kind() == reflect.Struct {
				nestedObj := reflect.New(fieldType).Interface()
				if err := Bind(req, nestedObj); err != nil {
					return err
				}
				objValue.Field(i).Set(reflect.ValueOf(nestedObj).Elem())
			} else if fieldType.Kind() == reflect.Ptr && fieldType.Elem().Kind() == reflect.Struct {
				nestedObj := reflect.New(fieldType.Elem()).Interface()
				if err := Bind(req, nestedObj); err != nil {
					return err
				}
				objValue.Field(i).Set(reflect.ValueOf(nestedObj).Elem().Addr())
			}
			continue
		}

		if fieldType.Kind() == reflect.Slice {
			// 如果字段是切片类型，则需要特殊处理
			sliceType := fieldType.Elem()
			slice := reflect.MakeSlice(fieldType, len(values), len(values))
			for j, value := range values {
				sliceValue, err := convertValue(value, sliceType)
				if err != nil {
					return fmt.Errorf("invalid form parameter: %s", tag)
				}
				slice.Index(j).Set(sliceValue)
			}
			fieldValue.Set(slice)
		} else {
			// 否则直接转换为字段类型
			rftVal, err := convertValue(values[0], fieldType)
			if err != nil {
				return fmt.Errorf("invalid form parameter: %s", tag)
			}
			fieldValue.Set(rftVal)
		}
	}
	return nil
}

// convertValue 将字符串转换为指定类型的值
func convertValue(value string, typ reflect.Type) (reflect.Value, error) {
	switch typ.Kind() {
	case reflect.String:
		return reflect.ValueOf(value), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intValue, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(intValue).Convert(typ), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintValue, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(uintValue).Convert(typ), nil
	case reflect.Float32, reflect.Float64:
		floatValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(floatValue).Convert(typ), nil
	case reflect.Bool:
		boolValue, err := strconv.ParseBool(value)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(boolValue), nil
	default:
		return reflect.Value{}, fmt.Errorf("unsupported type: %s", typ.String())
	}
}
