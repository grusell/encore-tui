package main

import (
	"time"
)

type JobsPoller struct {
	encoreClient *EncoreClient
	onUpdate     func([]EntityModelEncoreJob, error)
	pollInterval int
}

func NewJobsPoller(encoreClient *EncoreClient, pollInterval int,
	onUpdate func([]EntityModelEncoreJob, error)) *JobsPoller {
	jp := JobsPoller{encoreClient, onUpdate, pollInterval}
	return &jp
}

func (jp *JobsPoller) start() {
	go jp.pollJobs()
}

func (jp *JobsPoller) pollJobs() {
	for {
		jobs, err := encoreClient.getJobs()
		jp.onUpdate(*jobs.Embedded.EncoreJobs, err)
		time.Sleep(time.Duration(jp.pollInterval) * time.Second)
	}
}
