// SPDX-FileCopyrightText: 2024 Gustav Grusell
//
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"errors"
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"os"
	"strconv"
	"strings"
	"time"
)

var encoreClient *EncoreClient = NewEncoreClient(getEnv("ENCORE_URL", "http://localhost:8080"),
	getEnv("ENCORE_AUTH_HEADER", ""))

func getEnv(key string, defaultValue string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	} else {
		return val
	}
}

var HandleError func(err error)

func NewKeyHelp(keyHelp []string) *tview.TextView {
	var sb strings.Builder
	for i := 0; i < len(keyHelp)-1; i += 2 {
		sb.WriteString(
			fmt.Sprintf("  [black:white]%s[white:black] %s",
				keyHelp[i], keyHelp[i+1]))
	}
	return tview.NewTextView().
		SetSize(1, 0).
		SetText(sb.String()).
		SetDynamicColors(true)
}

func main() {
	jobsPollIntervalStr := getEnv("POLL_INTERVAL", "10")
	jobsPollInterval, err := strconv.Atoi(jobsPollIntervalStr)
	if err != nil {
		jobsPollInterval = 10
	}
	var jobsPoller *JobsPoller

	app := tview.NewApplication()
	externalEditor := NewExternalEditor(app)
	pages := tview.NewPages()

	updated := tview.NewTextView().SetSize(1, 0).SetLabel("Last updated:  ")
	messages := tview.NewTextView().SetSize(1, 0).SetDynamicColors(true)
	statusRow := tview.NewFlex()
	statusRow.AddItem(updated, 24, 0, false)
	statusRow.AddItem(nil, 5, 0, false)
	statusRow.AddItem(messages, 0, 1, true)

	HandleError = func(err error) {
		t := time.Now()
		text := fmt.Sprintf("[green]%s[white]: [red]%s", t.Format(time.TimeOnly), err.Error())
		messages.SetText(text)
	}

	jobView := NewJobView("job", pages)
	createJob := NewCreateJob("createJob", pages, externalEditor,
		func(job *EncoreJobRequestBody) error {
			err := encoreClient.PostJob(*job)
			if err == nil {
				jobsPoller.Poll()
			}
			return err
		})

	jobActions := JobActions{
		func(job *EntityModelEncoreJob) {
			jobView.Show(job)
		},
		func() {
			createJob.Show(nil)
		},
		func(job *EntityModelEncoreJob) {
			if *job.Status == "IN_PROGRESS" || *job.Status == "QUEUED" {
				err := encoreClient.CancelJob(job.Id)
				if err != nil {
					HandleError(errors.New(fmt.Sprintf("Cancel job failed: %s", err)))
				}
			} else {
				messages.SetText(fmt.Sprintf("Cannot cancel job with status %s",
					*job.Status))
			}
		},
		func(job *EntityModelEncoreJob) {
			err := encoreClient.DeleteJob(job.Id)
			if err != nil {
				HandleError(errors.New(fmt.Sprintf("Delete job failed: %s", err)))
			}
			jobsPoller.Poll()
		},
		func(job *EntityModelEncoreJob) {
			jobRequest := RequestFromJob(job)
			createJob.Show(jobRequest)
		},
	}
	table := NewJobsTable(jobActions)

	keyDescriptions := []string{"j/k", "Up/Down", "Enter", "View job", "C", "Cancel job", "n", "New job",
		"d", "Duplicate job", "^D", "Delete job", "^C", "Quit"}
	help := NewKeyHelp(keyDescriptions)

	flex := tview.NewFlex()
	flex.SetTitle("Encore TUI")
	flex.SetBorder(true)
	flex.SetDirection(tview.FlexRow)
	flex.AddItem(table, 0, 1, true)
	flex.AddItem(help, 1, 0, false)
	flex.AddItem(nil, 1, 0, false)
	flex.AddItem(statusRow, 1, 0, false)

	pages.AddPage("main", flex, true, true)
	pages.AddPage(jobView.name, jobView, false, false)
	pages.AddPage("createJob", createJob, false, false)

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			pages.SwitchToPage("main")
			return nil
		}
		return event
	})

	jobsPoller = NewJobsPoller(encoreClient, jobsPollInterval, func(jobs []EntityModelEncoreJob, err error) {
		if err != nil {
			HandleError(errors.New("Failed to fetch jobs: " + err.Error()))
		} else {
			app.QueueUpdateDraw(func() {
				table.SetData(jobs)
				t := time.Now()
				updated.SetText(t.Format(time.TimeOnly))
			})
		}
	})
	jobsPoller.start()
	jobsPoller.Poll()
	if err := app.SetRoot(pages, true).SetFocus(pages).Run(); err != nil {
		panic(err)
	}

}
