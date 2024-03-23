package main

import (

	"fmt"
	"encoding/json"
	"github.com/rivo/tview"
	"github.com/alecthomas/chroma/quick"
)

type JsonView struct {
	*tview.TextView
}

func NewJsonView() *JsonView {
	jtv := JsonView{tview.NewTextView()}
	jtv.SetDynamicColors(true)
	return &jtv
}

func (jtv *JsonView) SetObj(obj interface{}) {
	json,_ := json.MarshalIndent(obj, "", "  ")
	jtv.Clear()
	writer := tview.ANSIWriter(jtv)
	quick.Highlight(writer, fmt.Sprint(string(json)), "json", "terminal256", "dracula")
}
