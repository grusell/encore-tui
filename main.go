package main


//go:generate go run github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml https://petstore3.swagger.io/api/v3/openapi.json
//go:generate go run github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml encore-api.yaml


import (
	//	"log"
	"fmt"

	//	"encoding/json"
	//	"net/url"
	//	"path/filepath"
	"time"
	"os"
	"github.com/rivo/tview"
	"github.com/gdamore/tcell/v2"
	"github.com/alecthomas/chroma/quick"
	"gopkg.in/yaml.v3"
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

/*
func formatDate(date time.Time) string {
	return fmt.Sprintf("%04d-%02d-%02d %02d:%02d",
		date.Year(), date.Month(), date.Day(), date.Hour(), date.Minute())
}
*/


func main() {
	//var jobPfage PagedModelEntityModelEncoreJob
	//	jobChannel := make(chan PagedModelEntityModelEncoreJob)
	//	errorChannel := make(chan error)
	var jobsTable JobsTable
	//	quitChannel := make(chan int)

	jobView := tview.NewTextView()
	jobView.SetBorder(true)
	jobView.SetDynamicColors(true)
	app := tview.NewApplication()
	pages := tview.NewPages()

	newJob := tview.NewForm()
	newJob.SetTitle("Create job")
	newJob.SetBorder(true)
	newJob.AddInputField("Input file", "", 90, nil, nil)
	newJob.AddButton("Create job", func() {
		pages.SwitchToPage("main")
	})
	newJob.AddButton("Cancel",  func() {
		pages.SwitchToPage("main")
	})
	
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 'N' {
			x,y,w,h := pages.GetRect()
			newJob.SetRect(x+2,y+2,w-4,h-4)
			pages.ShowPage("newJob")
		}
		if event.Key() == tcell.KeyEscape {
			pages.SwitchToPage("main")
			return nil
		}
		return event
	})


	profiles := []string {"program", "x264-1080p-medium"}
	newJob.AddDropDown("Profile", profiles, 0, nil)
	
	table := tview.NewTable().SetContent(&jobsTable).SetSelectable(true, false)
	table.SetSelectedFunc(func(row int, column int) {
		job := jobsTable.jobs[row]
		//		jobJson,_ := json.MarshalIndent(job, "", "  ")
		jobJson,_ := yaml.Marshal(job)
		jobView.SetTitle(fmt.Sprintf("Job %s", job.Id))
		x,y,w,h := pages.GetRect()
		jobView.Clear()
		writer := tview.ANSIWriter(jobView)
		quick.Highlight(writer, fmt.Sprint(string(jobJson)), "yaml", "terminal256", "dracula")
		//		jobView.SetText(string(jobJson))
		jobView.SetRect(x+2,y+2,w-4,h-4)
		pages.ShowPage("job")
	})
	updated := tview.NewTextView().SetSize(1,0).SetLabel("Last updated:  ")
	flex := tview.NewFlex()
	flex.SetTitle("Encore TUI")
	flex.SetBorder(true)
	flex.SetDirection(tview.FlexRow)
	flex.AddItem(table, 0, 1, true)
	flex.AddItem(updated, 1, 0, false)


	pages.AddPage("main", flex, true, true)
	pages.AddPage("job", jobView, false, false)
	pages.AddPage("newJob", newJob, false, false)

	

	
	go getJobsRoutine(app, &jobsTable, updated)
	if err := app.SetRoot(pages, true).SetFocus(pages).Run(); err != nil {
		panic(err)
	}
}
