package common

import "context"

type Response struct {
	Code   StatusCode  `json:"code"`
	Msg    string      `json:"msg"`
	Result interface{} `json:"result"`
}

func NewResponse(c context.Context, code StatusCode, result interface{}) Response {
	return Response{code, code.GetMsg(), result}
}

func SuccessResponse(c context.Context, result interface{}) Response {
	return NewResponse(c, Success, result)
}
