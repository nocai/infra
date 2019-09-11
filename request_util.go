package infra

import (
	"github.com/gorilla/mux"
	"github.com/nocai/infra/returncode"
	"net/http"
	"strconv"
)

func RestfulID(req *http.Request) (uint64, error) {
	ID, exist := mux.Vars(req)["ID"]
	if !exist {
		return 0, returncode.ErrBadRequest
	}

	return strconv.ParseUint(ID, 10, 64)
}

func FormPagination(req *http.Request) (page, pageSize uint64, err error) {
	if err = req.ParseForm(); err != nil {
		return 0, 0, err
	}

	page, err = strconv.ParseUint(req.Form.Get("Page"), 10, 64)
	if err != nil {
		return 0, 0, err
	}
	pageSize, err = strconv.ParseUint(req.Form.Get("PageSize"), 10, 64)
	if err != nil {
		return 0, 0, err
	}
	return page, pageSize, nil
}

func FormBool(req *http.Request, arg string) (exist bool, b bool, err error) {
	if err = req.ParseForm(); err != nil {
		return false, false, err
	}
	if arg = req.Form.Get(arg); arg != "" {
		b, err = strconv.ParseBool(arg)
		return true, b, err
	}
	return false, false, nil
}

func FormInt(req *http.Request, arg string) (exist bool, i int, err error) {
	if err = req.ParseForm(); err != nil {
		return false, 0, err
	}
	if arg = req.Form.Get(arg); arg != "" {
		i64, err := strconv.ParseInt(arg, 10, 10)
		if err != nil {
			return true, 0, err
		}
		return true, int(i64), nil
	}
	return false, 0, nil
}
