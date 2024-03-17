package main


//go:generate go run github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml https://petstore3.swagger.io/api/v3/openapi.json
//go:generate go run github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml encore-api.yaml


import (
	"log"
	"fmt"
	"io/ioutil"
	"net/http"
	"encoding/json"
	//	"net/url"
	//	"path/filepath"
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


func getJobsRoutine(jobsOut chan PagedModelEntityModelEncoreJob, errors chan error) {
	for true {
		//		log.Printf("Getting jobs")
		jobs, err := getJobs()
		if err != nil {
			errors <- err
		} else {
			//			log.Printf("Sending jobs")
			jobsOut <- *jobs
		}
		time.Sleep(1 * time.Second)
	}
}


func formatDate(date time.Time) string {
	return fmt.Sprintf("%04d-%02d-%02d %02d:%02d",
		date.Year(), date.Month(), date.Day(), date.Hour(), date.Minute())
}

func printJob(job EntityModelEncoreJob) {
	progressStr := formatProgressBar(job)
	inputsStr := formatInputs(job)
	color := 0
	switch {
	case *job.Status == "FAILED":
		color = 31;
	case *job.Status == "SUCCESSFUL":
		color = 32;
	case *job.Status == "IN_PROGRESS":
		color = 33;
	}
	fmt.Printf("%-30s %12s \x1b[%dm%11s\x1b[0m %20s %3d%% %10s\n",
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
	jobChannel := make(chan PagedModelEntityModelEncoreJob)
	errorChannel := make(chan error)
	
	fmt.Printf("\x1b[s")
	var lines = 0
	go getJobsRoutine(jobChannel, errorChannel)
	for true {
		//		fmt.Printf("lines=%d", lines)
				fmt.Printf("\x1b[%dA", lines) // Move cursor up
		//				fmt.Printf("\x1b[0J") // Clear to end of screan
		//		var err error
		//		var jobPage PagedModelEntityModelEncoreJob
		select {
		case err := <- errorChannel:
			log.Printf("Error: %s\n", err)
			lines=1
		case jobPage := <- jobChannel:
			lines = 0
			for _, job := range *jobPage.Embedded.EncoreJobs {
				printJob(job)
				lines++
			}
		}
		/*
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
				*/
	}
	//	log.Printf("PageNo: %d", *jobPage.Page.Number)
}
