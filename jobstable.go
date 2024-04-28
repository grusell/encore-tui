// SPDX-FileCopyrightText: 2024 Gustav Grusell
//
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"net/url"
	"path/filepath"
	"strings"
	"time"
)

type JobActions struct {
	onViewJob      func(*EntityModelEncoreJob)
	onCreateJob    func()
	onCancelJob    func(*EntityModelEncoreJob)
	onDeleteJob    func(*EntityModelEncoreJob)
	onDuplicateJob func(*EntityModelEncoreJob)
}

type JobsTable struct {
	*tview.Table
	content    *JobsTableContent
	jobActions JobActions
}

func NewJobsTable(jobActions JobActions) *JobsTable {
	var jtc JobsTableContent
	jt := JobsTable{tview.NewTable().SetSelectable(true, false), &jtc, jobActions}
	jt.SetContent(&jtc)
	jt.SetSelectedFunc(func(row int, column int) {
		job := jtc.jobs[row]
		//		jobJson,_ := json.MarshalIndent(job, "", "  ")
		jobActions.onViewJob(&job)
	})
	jt.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 'n' {
			jt.jobActions.onCreateJob()
			return nil
		}
		if event.Rune() == 'd' {
			jt.jobActions.onDuplicateJob(jt.GetSelectedJob())
			return nil
		}
		if event.Rune() == 'C' {
			job := jt.GetSelectedJob()
			jt.jobActions.onCancelJob(job)
			return nil
		}

		if event.Key() == tcell.KeyCtrlD {
			job := jt.GetSelectedJob()
			jt.jobActions.onDeleteJob(job)
			return nil
		}

		return event
	})
	return &jt
}

func (jt *JobsTable) GetSelectedJob() *EntityModelEncoreJob {
	if len(jt.content.jobs) == 0 {
		return nil
	}
	row, _ := jt.GetSelection()
	return &jt.content.jobs[row]
}

func (jt *JobsTable) SetData(jobs []EntityModelEncoreJob) {
	jt.content.jobs = jobs
}

type JobsTableContent struct {
	tview.TableContentReadOnly

	// Nevermind the hard-coded values, this is just an example.
	//	data       [200][5]string
	jobs []EntityModelEncoreJob
}

func (jtc *JobsTableContent) GetCell(row, column int) *tview.TableCell {
	if len(jtc.jobs) == 0 {
		return nil
	}
	job := jtc.jobs[row]
	var content string
	switch column {
	case 0:
		return tview.NewTableCell(formatInputs(job)).SetExpansion(3)
	case 1:
		return tview.NewTableCell(formatDate(*job.CreatedDate)).SetExpansion(1)
	case 2:
		return tview.NewTableCell(formatStatus(job.Status)).SetExpansion(1)
	case 3:
		return tview.NewTableCell(job.Profile).SetExpansion(1)
	case 4:
		return tview.NewTableCell(fmt.Sprintf("%d%%", *job.Progress))
	case 5:
		return tview.NewTableCell(formatProgressBar(job))
	default:
		content = ""
	}

	return tview.NewTableCell(content)
}

func (jtc *JobsTableContent) GetRowCount() int {
	if len(jtc.jobs) == 0 {
		return 1
	}
	return len(jtc.jobs)
}

func (jtc *JobsTableContent) GetColumnCount() int {
	// input, createdDate, status, profile, progress percent, progress
	return 6
}

func formatInputs(job EntityModelEncoreJob) string {
	if len(job.Inputs) == 1 {
		return filenameFromUrl(job.Inputs[0].Uri)
	}
	var inputs []string
	for i := 0; i < len(job.Inputs); i++ {
		inputs = append(inputs, filenameFromUrl(job.Inputs[i].Uri))
	}
	return fmt.Sprintf("%v", inputs)
}

func filenameFromUrl(urlStr string) string {
	url, _ := url.Parse(urlStr)
	return filepath.Base(url.Path)
}

func formatDate(date time.Time) string {
	return fmt.Sprintf("%04d-%02d-%02d %02d:%02d",
		date.Year(), date.Month(), date.Day(), date.Hour(), date.Minute())
}

func formatStatus(status *EntityModelEncoreJobStatus) string {
	var color string = ""
	switch *status {
	case "SUCCESSFUL":
		color = "green"
	case "FAILED":
		color = "red"
	case "IN_PROGRESS":
		color = "blue"
	}
	return fmt.Sprintf("[%s]%s", color, *status)
}

func formatProgressBar(job EntityModelEncoreJob) string {
	ticks := int(*job.Progress) / 10
	return strings.Repeat("#", ticks) +
		strings.Repeat(" ", 10-ticks)
}

/*
func (d *JobsTableContent) AppendRow(row [5]string) {
	d.data[d.startIndex] = row
	d.startIndex = (d.startIndex + 1) % 200
        }
*/
