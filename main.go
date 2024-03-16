package main


//go:generate go run github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml https://petstore3.swagger.io/api/v3/openapi.json
//go:generate go run github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml encore-api.yaml


import (
	"log"
	"fmt"
	"io/ioutil"
	"net/http"
	"encoding/json"
	"net/url"
	"path/filepath"
	"time"
	"errors"
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

func formatDate(date time.Time) string {
	return fmt.Sprintf("%04d-%02d-%02d %02d:%02d",
		date.Year(), date.Month(), date.Day(), date.Hour(), date.Minute())
}

func printJob(job EntityModelEncoreJob) {
	var progressStr = ""
	for i :=1; i <= 10; i++ {
		if int(*job.Progress) >= i*10 {
			progressStr = progressStr + "#"
		} else {
			progressStr = progressStr + " "
		}
	}
	var inputsStr = ""
	for i := 0; i < len(job.Inputs); i++ {
		url, _ := url.Parse(job.Inputs[i].Uri)
		if i > 0 {
			inputsStr = inputsStr + ","
		}
		inputsStr = inputsStr + filepath.Base(url.Path)
	}
	color := 0
	switch {
	case *job.Status == "FAILED":
		color = 31;
	case *job.Status == "SUCCESSFUL":
		color = 32;
	case *job.Status == "IN_PROGRESS":
		color = 33;
	}
	fmt.Printf("[%d]%-30s %12s \x1b[%dm%11s\x1b[0m %20s %3d%% %10s\n", len(job.Inputs),
		inputsStr,
		formatDate(*job.CreatedDate),
		color,
		*job.Status,
		job.Profile,
		*job.Progress,
		progressStr)
}

func main() {
	//var jobPfage PagedModelEntityModelEncoreJob
	fmt.Printf("\x1b[s")
	var lines = 0
	for true {
		fmt.Printf("\x1b[%dA", lines)
		fmt.Printf("\x1b[0J")
		jobPage, err := getJobs();
		if err != nil {
			log.Printf("Error: %s\n", err)
			lines=1
			//		return;
		} else {
			//		fmt.Printf("\x1b[u")
			//		fmt.Printf("\x1b[0J")
			lines = 0
			for _, job := range *jobPage.Embedded.EncoreJobs {
				printJob(job)
				lines++
			}
		}
		time.Sleep(1 * time.Second)
	}
	//	log.Printf("PageNo: %d", *jobPage.Page.Number)
}
