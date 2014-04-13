package utils

import (
	"fmt"
	"testing"
)

type T struct {
	id   int
	name string
	by   []byte
}

func (this *T) ft() {
	this.id = 100
	this.name = "lijie"
	this.by = []byte("hello world")
}

func (this *T) fp() {
	fmt.Println(this.id, this.name, string(this.by))
}

func f1() {
	fmt.Println("f1()")
}

func f2(i int, s string, arr []int, t *T) {
	fmt.Println("start f2()")
	fmt.Println(i, s)
	fmt.Println(arr)
	t.ft()
	t.fp()
	fmt.Println("end f2()")
}

func f3(in ...interface{}) []int {
	arr := make([]int, 0)
	for _, v := range in {
		arr = append(arr, v.(int))
	}
	return arr
}

func TestFuncMap(t *testing.T) {

	funcMap := NewFuncMap()
	funcMap.Bind(1, f1)
	funcMap.Bind(2, f2)
	funcMap.Bind(3, f3)

	funcMap.Call(1)
	funcMap.Call(2, 99, "sss", []int{1, 2, 3}, &T{})
	arr, err := funcMap.Call(3, 7, 8, 9, 10)

	if err == nil {
		for i := 0; i < arr[0].Len(); i++ {
			fmt.Println(arr[0].Index(i).Int())
		}
	} else {
		t.Error(err)
	}

	_, err = funcMap.Call(4, 9)
	if err == nil {
		t.Error(err)
	}

	if funcMap.Exist(10) {
		t.Error("Exist error")
	}
	if !funcMap.Exist(3) {
		t.Error("Exist error")
	}

}
