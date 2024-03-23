package main

import (
	"github.com/rivo/tview"
	"errors"
	"github.com/gdamore/tcell/v2"
	"encoding/json"
)

type CreateJob struct {
	name string
	*tview.Flex
	text *JsonView
	pages *tview.Pages
	externalEditor *ExternalEditor
	job *EncoreJobRequestBody
}

func NewCreateJob(name string, pages *tview.Pages, externalEditor *ExternalEditor) *CreateJob {
	jc := CreateJob{name, tview.NewFlex(), NewJsonView(), pages, externalEditor, nil}
	jc.Box = tview.NewBox()
	jc.SetTitle("Create job")
	jc.SetBorder(true)
	jc.SetDirection(tview.FlexRow)
	jc.AddItem(jc.text, 0, 1, true)

	keys := []string{"e", "p", "c"}
	descs := []string{"Edit job", "Post Job", "Cancel"}
	helpRow := helpRow(keys, descs, 5)
	jc.AddItem(helpRow, 1, 0, false)

	jc.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 'c' {
			pages.HidePage(name)
			return nil
		}
		if event.Rune() == 'p' {
			jc.PostJob()
			pages.HidePage(name)
			return nil
		}
		if event.Rune() == 'e' {
			jc.EditJob()
			return nil
		}
		return event
	})

	return &jc
}

func (jc *CreateJob) Show() {
	job := NewEncoreJobRequestBody("", "")
	job.Id = nil
	job.OutputFolder = ""
	jc.job = &job
	//	text, _ := json.MarshalIndent(job, "", "  ")
	//	jc.text.SetText(string(text))
	jc.text.SetObj(&job)
	x,y,w,h := jc.pages.GetRect()
	jc.SetRect(x+2,y+2,w-4,h-4)
	jc.pages.ShowPage(jc.name)
}

func (jc *CreateJob) EditJob() {
	jsonBytes,_ := json.MarshalIndent(*jc.job, "", "  ")
	newJson, _ := jc.externalEditor.EditString(string(jsonBytes), ".json")
	err := json.Unmarshal([]byte(newJson), jc.job)
	if err != nil {
		panic(errors.New("Failed to unmarshall: " + err.Error()))
	}
	jc.text.SetObj(jc.job)
	
}

func (jc *CreateJob) PostJob() error {
	err := encoreClient.postJob(*jc.job)
	if err != nil {
		panic(err)
	}
	return nil
}
