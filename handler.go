package yun

type (
	//Handlers handler数组
	Handlers []HandlerFunc

	//HandlerFunc ...
	HandlerFunc func(*Context)
)