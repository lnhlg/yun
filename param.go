package yun

import (
	"strconv"
	"errors"
)

type Param struct {
	Key   string
	Value string
}

type Params []Param

func (ps Params) Get(name string) (string, error) {
	return ps.get(name)
}

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
	for i, _ := range ps {
		if ps[i].Key == name {
			return ps[i].Value, nil
		}
	}
	return "", errors.New("not exist")
}