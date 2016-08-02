package yun

type (
	Handler interface {
		GetType() string
	}

	Handlers []HandlerFunc

	HandlerFunc func(*Context)

	handler struct {
		typename string
		h interface{}
	}
)

func NewHandler(typename string, h interface{}) *handler {
	hh := new(handler)
	hh.typename = typename
	hh.h = h

	return hh
}

func (h handler) GetType() string {
	return h.typename
}