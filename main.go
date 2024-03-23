package main


import (
	"fmt"
	"time"
	"os"
	"github.com/rivo/tview"
	"github.com/gdamore/tcell/v2"
	"strings"
)

var encoreClient *EncoreClient = NewEncoreClient(getEnv("ENCORE_URL", "http://localhost:8080"))


func getEnv(key string, defaultValue string) string {
	val, ok := os.LookupEnv(key)
		if !ok {
			return defaultValue
		} else {
			return val
		}
}


func getJobsRoutine(app *tview.Application, jobsTable *JobsTable, updated *tview.TextView) {
	for true {
		jobs, err := encoreClient.getJobs()
		if err != nil {
			panic(err)
		} else {
			app.QueueUpdateDraw(func() {
				jobsTable.SetData(*jobs.Embedded.EncoreJobs)
				t := time.Now()
				updated.SetText(t.Format(time.TimeOnly))
			})
		}
		for i:= 0; i < 10; i++ {
			time.Sleep(100 * time.Millisecond)
		}
	}
}


type JobView struct {
	name string
	*JsonView
	pages *tview.Pages
}

func NewJobView(name string, pages *tview.Pages) *JobView {
	//	tv := tview.NewTextView()
	jv := JobView{name, NewJsonView(), pages}
	jv.SetBorder(true)
	return &jv
}

func (jv *JobView) Show(job *EntityModelEncoreJob) {
	//		jobJson,_ := json.MarshalIndent(job, "", "  ")
	//	jv.SetJob(job)
	jv.SetTitle(fmt.Sprintf("Job %s", job.Id))
	jv.SetObj(*job)
	x,y,w,h := jv.pages.GetRect()
	jv.SetRect(x+2,y+2,w-4,h-4)
	jv.pages.ShowPage(jv.name)
}

func helpRow(keys []string, helpTexts []string, cmdsPerRow int) *tview.TextView{
	rows := len(keys) / cmdsPerRow
	if len(keys) % cmdsPerRow != 0 {
		rows++
	}
	var sb strings.Builder
	for idx, key := range keys {
		if idx % cmdsPerRow > 0 {
			sb.WriteString("  ")
		}
		if idx < len(helpTexts) {
			sb.WriteString(fmt.Sprintf("[black:white]%s[white:black] %s",
				key, helpTexts[idx]))
		}
	}
	return tview.NewTextView().
		SetSize(rows,0).
		SetText(sb.String()).
		SetDynamicColors(true)
}




func main() {
	var jobsTable JobsTable
	app := tview.NewApplication()
	externalEditor := NewExternalEditor(app)	
	pages := tview.NewPages()
	jobView := NewJobView("job", pages)
	createJob := NewCreateJob("createJob", pages, externalEditor)
	
	table := tview.NewTable().SetContent(&jobsTable).SetSelectable(true, false)
	table.SetSelectedFunc(func(row int, column int) {
		job := jobsTable.jobs[row]
		//		jobJson,_ := json.MarshalIndent(job, "", "  ")
		jobView.Show(&job)
	})

	updated := tview.NewTextView().SetSize(1,0).SetLabel("Last updated:  ")
	messages := tview.NewTextView().SetSize(1,0)
	statusRow := tview.NewFlex()
	statusRow.AddItem(updated, 24, 0, false)
	statusRow.AddItem(nil, 5, 0, false)
	statusRow.AddItem(messages, 0, 1, true)

	help := tview.NewTextView().
		SetSize(1,0).
		SetText(" [black:white]j/k[white:black] Up/Down  [black:white]Enter[white:black] View job  [black:white]C[white:black] Cancel job  [black:white]n[white:black] create job  [black:white]^C[white:black] Quit").
		SetDynamicColors(true)


	flex := tview.NewFlex()
	flex.SetTitle("Encore TUI")
	flex.SetBorder(true)
	flex.SetDirection(tview.FlexRow)
	flex.AddItem(table, 0, 1, true)
	flex.AddItem(help, 1, 0, false)
	flex.AddItem(nil, 1, 0, false)
	//	flex.AddItem(updated, 1, 0, false)
	flex.AddItem(statusRow, 1, 0, false)


	pages.AddPage("main", flex, true, true)
	pages.AddPage(jobView.name, jobView, false, false)
	//	pages.AddPage("newJob", newJob, false, false)
	pages.AddPage("createJob", createJob, false, false)

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			pages.SwitchToPage("main")
			return nil
		}
		return event
	})

	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 'n' {
			createJob.Show()
			return nil
		}
		/*		if event.Rune() == 'N' {
			newJob.Show()
			return nil
		} */
		if event.Rune() == 'C' {
			row, _ := table.GetSelection()
			job := jobsTable.jobs[row]
			if *job.Status == "IN_PROGRESS" || *job.Status == "QUEUED" {
				err := encoreClient.CancelJob(job.Id)
				if err != nil {
					messages.SetText(fmt.Sprintf("Error: %s", err))
				}
			} else {
				messages.SetText(fmt.Sprintf("Cannot cancel job with status %s", *job.Status))
			}
			return nil
		}
		return event
	})
	
	go getJobsRoutine(app, &jobsTable, updated)
	if err := app.SetRoot(pages, true).SetFocus(pages).Run(); err != nil {
		panic(err)
	}
}
