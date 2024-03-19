package main


//go:generate go run github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml https://petstore3.swagger.io/api/v3/openapi.json
//go:generate go run github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml encore-api.yaml


import (
	//	"log"
	"fmt"
	"io/ioutil"
	"net/http"
	"encoding/json"
	//	"net/url"
	//	"path/filepath"
	"time"
	"errors"
	"github.com/rivo/tview"
	"github.com/gdamore/tcell/v2"
	"github.com/alecthomas/chroma/quick"
)

func getJobs() (*PagedModelEntityModelEncoreJob, error) {
	resp, err := http.Get("http://localhost:8080/encoreJobs?sort=createdDate,desc")
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("Failed to get jobs code=%d body=%s", resp.StatusCode, string(body)))
	}
	if err != nil {
		return nil, err
	}
	var jobPage PagedModelEntityModelEncoreJob
	//	jobPage.Links = &[]Link{}
	err = json.Unmarshal(body, &jobPage)
	if err != nil {
		return nil, err;
	}
	return &jobPage, nil
}


func getJobsRoutine(app *tview.Application, jobsTable *JobsTable, updated *tview.TextView) {
	for true {
		//		log.Printf("Getting jobs")
		jobs, err := getJobs()
		if err != nil {
			panic(err)
		} else {
						app.QueueUpdateDraw(func() {
							jobsTable.SetData(*jobs.Embedded.EncoreJobs)
							t := time.Now()
							updated.SetText(t.Format(time.TimeOnly))
						})
			//			jobsOut <- *jobs
		}
		for i:= 0; i < 10; i++ {
			time.Sleep(100 * time.Millisecond)
			/*
			select {
			case q := <-quit:
				return
			default
				time.Sleep(1 * time.Second)
			}
			*/
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
	app := tview.NewApplication()
	pages := tview.NewPages()

	jobView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			pages.HidePage("job")
			return nil
		}
		return event
	})
	jobView.SetDynamicColors(true)
	
	table := tview.NewTable().SetContent(&jobsTable).SetSelectable(true, false)
	table.SetSelectedFunc(func(row int, column int) {
		job := jobsTable.jobs[row]
		jobJson,_ := json.MarshalIndent(job, "", "  ")
		jobView.SetTitle(fmt.Sprintf("Job %s", job.Id))
		x,y,w,h := pages.GetRect()
		jobView.Clear()
		writer := tview.ANSIWriter(jobView)
		quick.Highlight(writer, fmt.Sprint(string(jobJson)), "json", "terminal256", "dracula")
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

	

	
	go getJobsRoutine(app, &jobsTable, updated)
	if err := app.SetRoot(pages, true).SetFocus(pages).Run(); err != nil {
		panic(err)
	}
}
