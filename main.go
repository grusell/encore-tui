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


func getJobsRoutine(app *tview.Application, jobsTable *JobsTableContent, updated *tview.TextView) {
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

func NewKeyHelp(keyHelp []string) *tview.TextView{
	var sb strings.Builder
	for i := 0; i < len(keyHelp) - 1; i += 2 {
		sb.WriteString(
			fmt.Sprintf("  [black:white]%s[white:black] %s",
				keyHelp[i], keyHelp[i+1]))
	}
	return tview.NewTextView().
		SetSize(1,0).
		SetText(sb.String()).
		SetDynamicColors(true)
}

func main() {
	app := tview.NewApplication()
	externalEditor := NewExternalEditor(app)	
	pages := tview.NewPages()

	updated := tview.NewTextView().SetSize(1,0).SetLabel("Last updated:  ")
	messages := tview.NewTextView().SetSize(1,0)
	statusRow := tview.NewFlex()
	statusRow.AddItem(updated, 24, 0, false)
	statusRow.AddItem(nil, 5, 0, false)
	statusRow.AddItem(messages, 0, 1, true)

	
	jobView := NewJobView("job", pages)
	createJob := NewCreateJob("createJob", pages, externalEditor)

	jobActions := JobActions{
		func(job *EntityModelEncoreJob) {
			jobView.Show(job)
		},
		func() {
			createJob.Show()
		},
		func(job *EntityModelEncoreJob) {
			if *job.Status == "IN_PROGRESS" || *job.Status == "QUEUED" {
				err := encoreClient.CancelJob(job.Id)
				if err != nil {
					messages.SetText(fmt.Sprintf("Error: %s", err))
				}
			} else {
				messages.SetText(fmt.Sprintf("Cannot cancel job with status %s",
					*job.Status))
			}
		},
	}
	table := NewJobsTable(jobActions)


	keyDescriptions := []string{"j/k", "Up/Down", "Enter", "View job",  "C", "Cancel job",  "n", "New job",  "^C", "Quit"}
	help := NewKeyHelp(keyDescriptions)


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

	
	go getJobsRoutine(app, table.content, updated)
	if err := app.SetRoot(pages, true).SetFocus(pages).Run(); err != nil {
		panic(err)
	}
}
