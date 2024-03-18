package main

import (
	"strings"
	"path/filepath"
	"fmt"
	"net/url"
	"time"
	"github.com/rivo/tview"
)

type JobsTable struct {
	tview.TableContentReadOnly

	// Nevermind the hard-coded values, this is just an example.
	//	data       [200][5]string
	jobs []EntityModelEncoreJob
}

func (d *JobsTable) GetCell(row, column int) *tview.TableCell {
	if len(d.jobs) == 0 {
		return tview.NewTableCell(fmt.Sprintf("data-%d", column))
	}
	job := d.jobs[row]
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

func (d *JobsTable) GetRowCount() int {
	if len(d.jobs) == 0 {
		return 1
	}
	return len(d.jobs)
}

func (d *JobsTable) GetColumnCount() int {
	// input, createdDate, status, profile, progress percent, progress
	return 6
}

func (d *JobsTable) SetData(jobs []EntityModelEncoreJob) {
	d.jobs = jobs
}

func formatInputs(job EntityModelEncoreJob) string {
	if len(job.Inputs) == 1 {
		return filenameFromUrl(job.Inputs[0].Uri)
	}
	var inputs []string
	for i := 0; i < len(job.Inputs); i++ {
		inputs = append(inputs,filenameFromUrl(job.Inputs[i].Uri))
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
		color="green"
	case "FAILED":
		color="red"
	case "IN_PROGRESS":
		color="blue"
	}
	return fmt.Sprintf("[%s]%s", color, *status)
}



func formatProgressBar(job EntityModelEncoreJob) string {
	ticks := int(*job.Progress) / 10;
	return strings.Repeat("#", ticks) +
		strings.Repeat(" ", 10 - ticks)
}

/*
func (d *JobsTable) AppendRow(row [5]string) {
	d.data[d.startIndex] = row
	d.startIndex = (d.startIndex + 1) % 200
        }
*/
