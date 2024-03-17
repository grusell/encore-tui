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
	jobPage PagedModelEntityModelEncoreJob
}

func (d *JobsTable) GetCell(row, column int) *tview.TableCell {
	job := (*d.jobPage.Embedded.EncoreJobs)[row]
	var content string
	switch column {
	case 0:
		content = formatInputs(job)
	case 1:
		content = formatDate(*job.CreatedDate)
	case 2:
		content = fmt.Sprintf("%s", *job.Status)
	case 3:
		content = job.Profile
	case 4:
		content = fmt.Sprintf("%d%%", *job.Progress)
	case 5:
		content = formatProgressBar(job)
	default:
		content = ""
	}
	
		
	return tview.NewTableCell(content)
}

func (d *JobsTable) GetRowCount() int {
	return len(*d.jobPage.Embedded.EncoreJobs)
}

func (d *JobsTable) GetColumnCount() int {
	// input, createdDate, status, profile, progress percent, progress
	return 6
}

func (d *JobsTable) SetData(jobPage PagedModelEntityModelEncoreJob) {
	d.jobPage = jobPage
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
