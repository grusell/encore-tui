package main

import "github.com/mohae/deepcopy"

func NewEncoreJobRequestBody() *EncoreJobRequestBody {
	input := Input{
		Type:   "AudioVideo",
		Uri:    "",
		Params: make(map[string]string),
	}
	inputs := []Input{
		input,
	}
	return &EncoreJobRequestBody{
		BaseName:      "",
		OutputFolder:  "",
		Profile:       "",
		Inputs:        inputs,
		ProfileParams: make(map[string]map[string]interface{}),
		LogContext:    make(map[string]string),
		Priority:      50,
	}
}

func RequestFromJob(job *EntityModelEncoreJob) *EncoreJobRequestBody {
	jobRequest := NewEncoreJobRequestBody()
	jobRequest.BaseName = job.BaseName
	jobRequest.Duration = new(float64)
	*jobRequest.Duration = *job.Duration
	jobRequest.ExternalId = job.ExternalId
	jobRequest.Inputs = deepcopy.Copy(job.Inputs).([]Input)
	for i := 0; i < len(jobRequest.Inputs); i++ {
		jobRequest.Inputs[i].Analyzed = nil
	}
	jobRequest.LogContext = deepcopy.Copy(job.LogContext).(map[string]string)
	jobRequest.OutputFolder = job.OutputFolder
	jobRequest.Priority = job.Priority
	jobRequest.Profile = job.Profile
	jobRequest.ProfileParams = deepcopy.Copy(job.ProfileParams).(map[string]map[string]interface{})
	jobRequest.ProgressCallbackUri = job.ProgressCallbackUri
	jobRequest.SeekTo = deepcopy.Copy(job.SeekTo).(*float64)
	jobRequest.SegmentLength = deepcopy.Copy(job.SegmentLength).(*float64)
	jobRequest.ThumbnailTime = deepcopy.Copy(job.ThumbnailTime).(*float64)
	return jobRequest
}

func copyMap[V any](m map[string]V) map[string]V {
	c := make(map[string]V)
	for k, v := range m {
		c[k] = v
	}
	return c
}
