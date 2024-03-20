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

type JobView struct {
	name string
	*tview.TextView
	pages *tview.Pages
}

func NewJobView(name string, pages *tview.Pages) *JobView {
	//	tv := tview.NewTextView()
	jv := JobView{name, tview.NewTextView(), pages}
	jv.SetBorder(true)
	jv.SetDynamicColors(true)
	return &jv
}

func (jv *JobView) Show(job EntityModelEncoreJob) {
		//		jobJson,_ := json.MarshalIndent(job, "", "  ")
		jobJson,_ := yaml.Marshal(job)
		jv.SetTitle(fmt.Sprintf("Job %s", job.Id))
		x,y,w,h := jv.pages.GetRect()
		jv.Clear()
		writer := tview.ANSIWriter(jv)
		quick.Highlight(writer, fmt.Sprint(string(jobJson)), "yaml", "terminal256", "dracula")
		//		jobView.SetText(string(jobJson))
		jv.SetRect(x+2,y+2,w-4,h-4)
		jv.pages.ShowPage(jv.name)
}

type JobCreate struct {
	name string
	*tview.Form
	pages *tview.Pages
}

func NewJobCreate(name string, pages *tview.Pages) *JobCreate {
	profiles := []string {"program", "x264-1080p-medium"}
	jc := JobCreate{name, tview.NewForm(), pages}
	jc.SetTitle("Create job")
	jc.SetBorder(true)
	jc.AddInputField("Input file", "", 90, nil, nil)
	jc.AddDropDown("Profile", profiles, 0, nil)

	jc.AddButton("Create job", func() {
		jc.PostJob()
		pages.HidePage(name)
	})
	jc.AddButton("Cancel",  func() {
		pages.HidePage(name)
	})
	return &jc
}

func (jc *JobCreate) Show() {
	x,y,w,h := jc.pages.GetRect()
	jc.SetRect(x+2,y+2,w-4,h-4)
	jc.pages.ShowPage(jc.name)
}

func (jc *JobCreate) PostJob() {
	input := jc.GetFormItemByLabel("Input file").(*tview.InputField).GetText()
	_, profile := jc.GetFormItemByLabel("Profile").(*tview.DropDown).GetCurrentOption()
	job := CreateJob(input, profile)
	_, err := encoreClient.postJob(job)
	if err != nil {
		panic(err)
	}
}


func main() {
	var jobsTable JobsTable

	app := tview.NewApplication()
	pages := tview.NewPages()
	jobView := NewJobView("job", pages)

	newJob := NewJobCreate("newJob", pages)
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 'N' {
			newJob.Show()
		}
		if event.Key() == tcell.KeyEscape {
			pages.SwitchToPage("main")
			return nil
		}
		return event
	})

	
	table := tview.NewTable().SetContent(&jobsTable).SetSelectable(true, false)
	table.SetSelectedFunc(func(row int, column int) {
		job := jobsTable.jobs[row]
		//		jobJson,_ := json.MarshalIndent(job, "", "  ")
		jobView.Show(job)
	})
	updated := tview.NewTextView().SetSize(1,0).SetLabel("Last updated:  ")
	flex := tview.NewFlex()
	flex.SetTitle("Encore TUI")
	flex.SetBorder(true)
	flex.SetDirection(tview.FlexRow)
	flex.AddItem(table, 0, 1, true)
	flex.AddItem(updated, 1, 0, false)


	pages.AddPage("main", flex, true, true)
	pages.AddPage(jobView.name, jobView, false, false)
	pages.AddPage("newJob", newJob, false, false)

	

	
	go getJobsRoutine(app, &jobsTable, updated)
	if err := app.SetRoot(pages, true).SetFocus(pages).Run(); err != nil {
		panic(err)
	}
}
