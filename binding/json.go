package binding

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

var (
	EnableDecoderUseNumber             bool // 是否开启高精度判断,避免转化为float64造成精度损失
	EnableDecoderDisallowUnknownFields bool // 是否开启未知类型判断，一旦开启此判断，那么Json中只要存在go语言结构体中没有的类型就会出现错误
)

type jsonBinding struct{}

func (jsonBinding) Name() string {
	return "json"
}

func (jsonBinding) Bind(req *http.Request, obj any) error {
	if req == nil || req.Body == nil {
		return errors.New("invalid request")
	}
	return decodeJSON(req.Body, obj)
}

func decodeJSON(r io.Reader, obj any) error {
	decoder := json.NewDecoder(r)
	if EnableDecoderUseNumber {
		decoder.UseNumber()
	}
	if EnableDecoderDisallowUnknownFields {
		decoder.DisallowUnknownFields()
	}
	if err := decoder.Decode(obj); err != nil {
		return err
	}
	return validate(obj)
}
