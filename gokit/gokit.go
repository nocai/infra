package gokit

import (
	"context"
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/nocai/infra/returncoder"
	"net/http"
)

func ServerOptions(l log.Logger) []httptransport.ServerOption {
	return []httptransport.ServerOption{
		httptransport.ServerErrorHandler(ErrorHandler{l}),
		httptransport.ServerErrorEncoder(ErrorEncoder),
	}
}

// EncodeJSONResponse 成功时的统一数据处理
func EncodeJSONResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	switch response.(type) {
	case returncoder.ReturnCoder:
		return httptransport.EncodeJSONResponse(ctx, w, response)
	default:
		return httptransport.EncodeJSONResponse(ctx, w, returncoder.S(response))
	}
}

func DecodeJSONResponse2ReturnCoder(_ context.Context, resp *http.Response) (interface{}, error) {
	return returncoder.Unmarshal(resp.Body)
}

// ErrorEncoder 失败时的统一数据处理
func ErrorEncoder(ctx context.Context, err error, w http.ResponseWriter) {
	switch err.(type) {
	case returncoder.ReturnCoder:
		httptransport.DefaultErrorEncoder(ctx, err, w)
	default:
		httptransport.DefaultErrorEncoder(ctx, returncoder.F(err), w)
	}
}

type ErrorHandler struct {
	log.Logger
}

func (h ErrorHandler) Handle(ctx context.Context, err error) {
	_ = level.Error(h.Logger).Log("msg", fmt.Sprintf("%+v", err))
}
