package main

import (
	"github.com/rivo/tview"
	"errors"
	"github.com/gdamore/tcell/v2"
	"encoding/json"
	"fmt"
	"strings"
)

type CreateJob struct {
	name string
	*tview.Flex
	text *JsonView
	pages *tview.Pages
	messages *tview.TextView
	valid bool
	externalEditor *ExternalEditor
	job *EncoreJobRequestBody
}

func NewCreateJob(name string, pages *tview.Pages, externalEditor *ExternalEditor) *CreateJob {
	cj := CreateJob{name, tview.NewFlex(), NewJsonView(), pages,
		tview.NewTextView(), false, externalEditor, nil}
	cj.Box = tview.NewBox()
	cj.SetTitle("Create job")
	cj.SetBorder(true)
	cj.SetDirection(tview.FlexRow)
	cj.AddItem(cj.text, 0, 1, true)

	cj.messages.SetDynamicColors(true)
	cj.AddItem(cj.messages, 2, 0, false)

	keyDescriptions := []string{"e", "Edit job", "p",  "Post Job", "c,Esc", "Cancel",}

	helpRow := NewKeyHelp(keyDescriptions)
	cj.AddItem(helpRow, 1, 0, false)

	cj.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 'c' {
			pages.HidePage(name)
			return nil
		}
		if event.Rune() == 'p' {
			if cj.valid {
				cj.PostJob()
				pages.HidePage(name)
				return nil
			}
		}
		if event.Rune() == 'e' {
			cj.EditJob()
			return nil
		}
		return event
	})

	return &cj
}

func (cj *CreateJob) Show() {
	job := NewEncoreJobRequestBody("", "")
	job.Id = nil
	job.OutputFolder = ""
	cj.job = &job
	cj.ValidateJob()
	cj.text.SetObj(&job)
	x,y,w,h := cj.pages.GetRect()
	cj.SetRect(x+2,y+2,w-4,h-4)
	cj.pages.ShowPage(cj.name)
}

func (cj *CreateJob) EditJob() {
	jsonBytes,_ := json.MarshalIndent(*cj.job, "", "  ")
	newJson, _ := cj.externalEditor.EditString(string(jsonBytes), ".json")
	err := json.Unmarshal([]byte(newJson), cj.job)
	if err != nil {
		panic(errors.New("Failed to unmarshall: " + err.Error()))
	}
	cj.text.SetObj(cj.job)
	cj.ValidateJob()
	
}

func (cj *CreateJob) PostJob() error {
	err := encoreClient.postJob(*cj.job)
	if err != nil {
		panic(err)
	}
	return nil
}

func (cj *CreateJob) ValidateJob() {
	cj.valid = true
	var messages []string
	if cj.job.OutputFolder == "" {
		messages = append(messages, "outputFolder not set")
		cj.valid = false
	}
	if cj.job.BaseName == "" {
		messages = append(messages, "baseName not set")
		cj.valid = false
	}
	if cj.job.Profile == "" {
		messages = append(messages, "profile not set")
		cj.valid = false
	}
	for idx, input := range cj.job.Inputs {
		if input.Uri == "" {
			messages = append(messages, fmt.Sprintf("inputs[%d].uri not set", idx))
			cj.valid = false
		}
	}
	cj.messages.SetText(fmt.Sprintf("[red]%s", strings.Join(messages, ", ")))

}
