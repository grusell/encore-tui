package main

import (
	"fmt"
	"os"
	"os/exec"
	"github.com/rivo/tview"
	"errors"
)

var editorEnvVars = []string{"ETUI_EDITOR", "EDITOR"}
var editors = []string{"vim", "nano", "emacs"}

func findEditorCmd() string {
	var (
		bin string
		err error
	)
	for _, e := range editorEnvVars {
		env := os.Getenv(e)
		if env == "" {
			continue
		}
		if bin, err = exec.LookPath(env); err == nil {
			break
		}
	}
	if bin == "" {
		for _, editor := range editors {
			if bin, err = exec.LookPath(editor); err == nil {
				break
			}
		}

	}
	return bin
}

type ExternalEditor struct {
	app *tview.Application
	editorPath string
}

func NewExternalEditor(app *tview.Application) *ExternalEditor {
	ee := ExternalEditor{app, findEditorCmd()}
	return &ee
}


func (ee *ExternalEditor) EditString(value string, fileSuffix string) (string,error) {
	file, err := os.CreateTemp("", fmt.Sprintf("etui.*%s", fileSuffix))
	if err != nil {
		panic(err)
	}
	filename := file.Name()
	file.Close()
	os.WriteFile(filename, []byte(value), 644)
	defer os.Remove(filename)

	ee.app.Suspend(func() {
		cmd := exec.Command(ee.editorPath, filename)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		if err := cmd.Run(); err != nil {
			
		}
		return
	})
	
	newValue,err := os.ReadFile(filename)
	if err != nil {
		return "", errors.New("Failed to read file: " + err.Error())
	}
	return string(newValue), nil;
}
