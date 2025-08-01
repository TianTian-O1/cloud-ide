package serialize

import (
	"net/http"
	"sync"
)

type CodeMessager func(int) string

var cm CodeMessager

// SetCodeMessager 设置一个CodeMessager，该函数可以从code获取message
func SetCodeMessager(messager CodeMessager) {
	cm = messager
}

type resResult struct {
	Data    interface{} `json:"data"`
	Status  int         `json:"status"`
	Message string      `json:"message"`
}

type Response struct {
	HttpStatus int
	R          resResult
}

const (
	SuccessCode = iota
	FailCode
	ErrorCode
)

var pool = sync.Pool{
	New: func() interface{} {
		return &Response{}
	},
}

func PutResponse(resp *Response) {
	pool.Put(resp)
}

func NewResponse(status, code int, data interface{}, message string) *Response {
	res := pool.Get().(*Response)
	res.HttpStatus = status
	res.R.Status = code
	res.R.Data = data
	res.R.Message = message

	return res
}

func Ok() *Response {
	return NewResponse(http.StatusOK, SuccessCode, nil, "success")
}

func OkData(data interface{}) *Response {
	return NewResponse(http.StatusOK, SuccessCode, data, "success")
}

func OkCode(code int) *Response {
	if cm != nil {
		return NewResponse(http.StatusOK, code, nil, cm(code))
	}
	return NewResponse(http.StatusOK, code, nil, "success")
}

func OkCodeData(code int, data interface{}) *Response {
	if cm != nil {
		return NewResponse(http.StatusOK, code, data, cm(code))
	}
	return NewResponse(http.StatusOK, code, data, "success")
}

func Fail(code int) *Response {
	return FailData(code, nil)
}

func FailData(code int, data interface{}) *Response {
	if cm != nil {
		return NewResponse(http.StatusOK, code, data, cm(code))
	}
	return NewResponse(http.StatusOK, FailCode, data, "failed")
}

func Error(status int) *Response {
	return NewResponse(status, ErrorCode, nil, "error")
}
