// SPDX-FileCopyrightText: 2024 Gustav Grusell
//
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"fmt"
	"github.com/rivo/tview"
)

type JobView struct {
	name string
	*JsonView
	pages *tview.Pages
}

func NewJobView(name string, pages *tview.Pages) *JobView {
	jv := JobView{name, NewJsonView(), pages}
	jv.SetBorder(true)
	return &jv
}

func (jv *JobView) Show(job *EntityModelEncoreJob) {
	jv.SetTitle(fmt.Sprintf("Job %s", job.Id))
	jv.SetObj(*job)
	x, y, w, h := jv.pages.GetRect()
	jv.SetRect(x+2, y+2, w-4, h-4)
	jv.pages.ShowPage(jv.name)
}
