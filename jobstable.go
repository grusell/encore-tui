package main

import (
	"strings"
	"path/filepath"
	"fmt"
	"net/url"
)

/*

type JobsTable struct {
	tview.TableContentReadOnly

	// Nevermind the hard-coded values, this is just an example.
	//	data       [200][5]string
	data []EntityModelEncoreJob
}

func (d *JobsTable) GetCell(row, column int) *tview.TableCell {
	job := *jobPage.Embedded.EncoreJobs[row]
	switch column {
	case 0:
	}
	//	return tview.NewTableCell(d.data[(row+d.startIndex)%200][column])
}

func (d *JobsTable) GetRowCount() int {
	return len(data)
}

func (d *JobsTable) GetColumnCount() int {
	// input, createdDate, status, profile, progress percent, progress
	return 6
}
*/

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
