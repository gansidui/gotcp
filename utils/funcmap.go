package utils

import (
	"errors"
	"reflect"
	"strconv"
)

type FuncMap struct {
	funcs map[uint16]reflect.Value
}

func NewFuncMap() *FuncMap {
	return &FuncMap{
		funcs: make(map[uint16]reflect.Value),
	}
}

func (this *FuncMap) Bind(fnid uint16, fn interface{}) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = errors.New(strconv.Itoa(int(fnid)) + " is not callable.")
		}
	}()

	v := reflect.ValueOf(fn)
	v.Type().NumIn() // fn不是函数则panic
	this.funcs[fnid] = v
	return
}

func (this *FuncMap) Call(fnid uint16, params ...interface{}) (result []reflect.Value, err error) {
	if _, ok := this.funcs[fnid]; !ok {
		err = errors.New(strconv.Itoa(int(fnid)) + " does not exist.")
		return
	}

	in := make([]reflect.Value, len(params))
	for k, param := range params {
		in[k] = reflect.ValueOf(param)
	}

	result = this.funcs[fnid].Call(in)
	return
}

func (this *FuncMap) Exist(fnid uint16) bool {
	_, ok := this.funcs[fnid]
	return ok
}
