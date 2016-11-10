package yun

import (
	"strconv"
	"errors"
)

//Param 参数路由的参数
type Param struct {
	Key   string
	Value string
}

//Params 参数数组
type Params []Param

//Get 获取参数值
//name 参数名称
//return 字符串类型的参数值或错误
func (ps Params) Get(name string) (string, error) {
	return ps.get(name)
}

//GetInt 获取整数型参数值
//name 参数名称
//return 整型值或错误
func (ps Params) GetInt(name string) (int, error) {
	v, err := ps.get(name)
	if err != nil {
		return -1, err
	}

	i, err := strconv.Atoi(v)
	if err != nil {
		return -1, err
	}

	return i, nil
}

func (ps Params) get(name string) (string, error) {
	for i := range ps {
		if ps[i].Key == name {
			return ps[i].Value, nil
		}
	}
	return "", errors.New("not exist")
}