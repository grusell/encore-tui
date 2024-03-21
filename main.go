package main


//go:generate go run github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml https://petstore3.swagger.io/api/v3/openapi.json
//go:generate go run github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml encore-api.yaml


import (
	//	"log"
	"fmt"

	"encoding/json"
	//	"net/url"
	//	"path/filepath"
	"time"
	"os"
	"github.com/rivo/tview"
	"github.com/gdamore/tcell/v2"
	"github.com/alecthomas/chroma/quick"
	//	"gopkg.in/yaml.v3"
	"strings"
	"os/exec"
	"errors"
	//	"io/ioutil"
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

type JsonView struct {
	*tview.TextView
}

func NewJsonView() *JsonView {
	jtv := JsonView{tview.NewTextView()}
	jtv.SetDynamicColors(true)
	return &jtv
}

func (jtv *JsonView) SetObj(obj interface{}) {
	json,_ := json.MarshalIndent(obj, "", "  ")
	jtv.Clear()
	writer := tview.ANSIWriter(jtv)
	quick.Highlight(writer, fmt.Sprint(string(json)), "json", "terminal256", "dracula")
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
	err := encoreClient.postJob(job)
	if err != nil {
		panic(err)
	}
}

type JobCreateJson struct {
	name string
	*tview.Flex
	text *JsonView
	pages *tview.Pages
	job *EncoreJobRequestBody
	editFile func(string)
}

func NewJobCreateJson(name string, pages *tview.Pages, editFile func(string)) *JobCreateJson {
	jc := JobCreateJson{name, tview.NewFlex(), NewJsonView(), pages, nil, editFile}
	jc.Box = tview.NewBox()
	jc.SetTitle("Create job")
	jc.SetBorder(true)
	jc.SetDirection(tview.FlexRow)
	jc.AddItem(jc.text, 0, 1, true)

	keys := []string{"e", "p", "c"}
	descs := []string{"Edit job", "Post Job", "Cancel"}
	helpRow := helpRow(keys, descs, 5)
	jc.AddItem(helpRow, 1, 0, false)

	jc.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 'c' {
			pages.HidePage(name)
			return nil
		}
		if event.Rune() == 'p' {
			jc.PostJob()
			pages.HidePage(name)
			return nil
		}
		if event.Rune() == 'e' {
			jc.EditJob()
			return nil
		}
		return event
	})

	return &jc
}

func (jc *JobCreateJson) Show() {
	job := CreateJob("", "")
	job.Id = nil
	job.OutputFolder = ""
	jc.job = &job
	//	text, _ := json.MarshalIndent(job, "", "  ")
	//	jc.text.SetText(string(text))
	jc.text.SetObj(&job)
	x,y,w,h := jc.pages.GetRect()
	jc.SetRect(x+2,y+2,w-4,h-4)
	jc.pages.ShowPage(jc.name)
}

func (jc *JobCreateJson) EditJob() {
	file, err := os.CreateTemp("", "encoreJob.*.json")
	if err != nil {
		panic(err)
	}
	jsonBytes,_ := json.MarshalIndent(*jc.job, "", "  ")
	file.Write(jsonBytes)
	file.Close()
	defer os.Remove(file.Name())
	jc.editFile(file.Name())
	newJson,err := os.ReadFile(file.Name())
	if err != nil {
		panic(errors.New("Failed to read file: " + err.Error()))
	}
	//	jc.text.SetText(string(newJson))
	
	err = json.Unmarshal(newJson, jc.job)
	if err != nil {
		panic(errors.New("Failed to unmarshall: " + err.Error()))
	}
	jc.text.SetObj(jc.job)
	
}

func (jc *JobCreateJson) PostJob() error {
	err := encoreClient.postJob(*jc.job)
	if err != nil {
		panic(err)
	}
	return nil
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


	editFile := func(file string) {
		app.Suspend(func() {
			cmd := exec.Command("/usr/bin/vim", file)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			if err := cmd.Run(); err != nil {
				panic(err)
			}
			return
		})
	}
	
	pages := tview.NewPages()
	jobView := NewJobView("job", pages)

	newJob := NewJobCreate("newJob", pages)
	newJobJson := NewJobCreateJson("newJobJson", pages, editFile)
	
	table := tview.NewTable().SetContent(&jobsTable).SetSelectable(true, false)
	table.SetSelectedFunc(func(row int, column int) {
		job := jobsTable.jobs[row]
		//		jobJson,_ := json.MarshalIndent(job, "", "  ")
		jobView.Show(&job)
	})

	updated := tview.NewTextView().SetSize(1,0).SetLabel("Last updated:  ")
	messages := tview.NewTextView().SetSize(1,0).SetText("some message")
	statusRow := tview.NewFlex()
	statusRow.AddItem(updated, 24, 0, false)
	statusRow.AddItem(nil, 5, 0, false)
	statusRow.AddItem(messages, 0, 1, true)

	help := tview.NewTextView().
		SetSize(1,0).
		SetText(" [black:white]j/k[white:black] Up/Down  [black:white]Enter[white:black] View job  [black:white]C[white:black] Cancel job  [black:white]N[white:black] Create job (simple)  [black:white]^C[white:black] Quit").
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
	pages.AddPage("newJob", newJob, false, false)
	pages.AddPage("newJobJson", newJobJson, false, false)

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			pages.SwitchToPage("main")
			return nil
		}
		return event
	})

	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 'n' {
			newJobJson.Show()
			return nil
		}
		if event.Rune() == 'N' {
			newJob.Show()
			return nil
		}
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
