package yun

import (
	"bytes"
	"log"
)

func init() {
	log.SetFlags(0)
}

//IsDebugging 是否调试模式
func (eng *Engine) IsDebugging() bool {
	return eng.mode == DEBUG
}

func (eng *Engine) printDebugInfo(format string, values ...interface{}) {
	if eng.IsDebugging() {
		var buffer bytes.Buffer
		buffer.WriteString("[yun-Debug] ")
		buffer.WriteString(format)
		log.Printf(buffer.String(), values...)
	}
}

func (eng *Engine) printError(err error) {
	if err != nil {
		eng.printDebugInfo("[ERROR] %v\n", err)
	}
}
