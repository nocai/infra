package returncoder

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"
)

var (
	codes []int

	// 200 成功
	Ok      = New(http.StatusOK, http.StatusText(http.StatusOK))
	Created = New(http.StatusCreated, http.StatusText(http.StatusCreated))
	//NoContent = New(http.StatusNoContent, http.StatusText(http.StatusNoContent)) // 与kit的EncodeJSONResponse不太兼容

	// 400 客户端错误
	ErrBadRequest     = New(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
	ErrRequestTimeout = New(http.StatusRequestTimeout, http.StatusText(http.StatusRequestTimeout))

	// 500 服务端错误
	ErrInternalServer = New(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))

	// 通用600 <= code < 700
	ErrArguments   = New(600, "err: arguments")
	ErrUKDuplicate = New(601, "err: unique key duplicated")

	// ...
)

type ReturnCoder interface {
	error
	Code() int
	Message() string
	WithMessagef(format string, messages ...interface{}) ReturnCoder
	Data() interface{}
	SetData(interface{}) ReturnCoder
	Success() bool
}

var _ ReturnCoder = &returnCode{}

type returnCode struct {
	C int         `json:"Code"`
	M string      `json:"Message,omitempty"`
	D interface{} `json:"Data,omitempty"`
}

func (rc *returnCode) Success() bool {
	return http.StatusOK <= rc.C && rc.C <= http.StatusIMUsed
}

func (rc *returnCode) Code() int {
	return rc.C
}

func (rc *returnCode) Message() string {
	return rc.M
}

func (rc *returnCode) WithMessagef(format string, messages ...interface{}) ReturnCoder {
	return Of(rc.C, rc.M+":"+fmt.Sprintf(format, messages...))
}

func (rc *returnCode) Data() interface{} {
	return rc.D
}
func (rc *returnCode) SetData(d interface{}) ReturnCoder {
	rc.D = d
	return rc
}

func (rc *returnCode) Error() string {
	return rc.Message()
}

// 业务异常返回码
const StatusCode_ErrBusiness = 999

func (rc returnCode) StatusCode() int {
	if rc.C >= StatusCode_ErrBusiness {
		return StatusCode_ErrBusiness
	}
	return rc.C
}

func (rc returnCode) MarshalJSON() ([]byte, error) {
	type alias returnCode
	return json.Marshal(struct{ alias }{alias(rc)})
}

func (rc *returnCode) UnmarshalJSON(data []byte) error {
	type alias returnCode
	return json.Unmarshal(data, &struct{ *alias }{(*alias)(rc)})
}

func New(code int, message string) ReturnCoder {
	checkCode(code)
	return Of(code, message)
}

var lock sync.Mutex

// 检查code是否已经重复
func checkCode(code int) {
	lock.Lock()
	defer lock.Unlock()

	for _, c := range codes {
		if c == code {
			panic("Duplicate code = " + strconv.Itoa(code))
		}
	}

	codes = append(codes, code)
}

// ==================================================================================================================
func Of(code int, message string) ReturnCoder {
	return &returnCode{C: code, M: message}
}

func F(i interface{}) ReturnCoder {
	switch i.(type) {
	case ReturnCoder:
		return i.(ReturnCoder)
	case error:
		return Of(ErrInternalServer.Code(), i.(error).Error())
	default:
		return Of(ErrInternalServer.Code(), fmt.Sprint(i))
	}
}

func S(data interface{}) ReturnCoder {
	return &returnCode{C: Ok.Code(), D: data}
}

func Unmarshal(r io.Reader) (ReturnCoder, error) {
	bytes, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	var rc returnCode
	return &rc, rc.UnmarshalJSON(bytes)
}
