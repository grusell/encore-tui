package main

import (
	"time"
)

type JobsPoller struct {
	encoreClient  *EncoreClient
	onUpdate      func([]EntityModelEncoreJob, error)
	ticker        *time.Ticker
	updateChannel chan bool
}

func NewJobsPoller(encoreClient *EncoreClient, pollInterval int,
	onUpdate func([]EntityModelEncoreJob, error)) *JobsPoller {
	jp := JobsPoller{encoreClient, onUpdate,
		time.NewTicker(time.Duration(pollInterval) * time.Second),
		make(chan bool)}
	return &jp
}

func (jp *JobsPoller) start() {
	go jp.pollJobs()
}

func (jp *JobsPoller) Poll() {
	jp.updateChannel <- true
}

func (jp *JobsPoller) pollJobs() {
	for {
		select {
		case <-jp.updateChannel:
		case <-jp.ticker.C:
		}
		jobs, err := encoreClient.getJobs()
		if err != nil {
			jp.onUpdate(nil, err)
		} else {
			jp.onUpdate(*jobs.Embedded.EncoreJobs, nil)
		}
	}
}
