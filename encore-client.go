package main


import (
	"io/ioutil"
	"net/http"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"github.com/google/uuid"
	"strings"
	"bytes"
)

type EncoreClient struct {
	url string
}


func NewEncoreClient(url string) *EncoreClient {
	ec := EncoreClient{url: url}
	return &ec
}

func (ec *EncoreClient) getJobs() (*PagedModelEntityModelEncoreJob, error) {
	resp, err := http.Get(ec.url + "/encoreJobs?sort=createdDate,desc")
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("Failed to get jobs code=%d body=%s", resp.StatusCode, string(body)))
	}
	var jobPage PagedModelEntityModelEncoreJob
	//	jobPage.Links = &[]Link{}
	err = json.Unmarshal(body, &jobPage)
	if err != nil {
		return nil, err;
	}
	return &jobPage, nil
}

func (ec *EncoreClient) postJob(job EncoreJobRequestBody) error {
	jsonVal, _ := json.Marshal(job)
	resp, err := http.Post(ec.url + "/encoreJobs", "application/json", bytes.NewBuffer(jsonVal))
	if err != nil {
		return err
	}
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != 201 {
		return errors.New(fmt.Sprintf("Failed to post job code=%d body=%s", resp.StatusCode, string(body)))
	}
	return nil
}

func (ec *EncoreClient) CancelJob(jobId *uuid.UUID) error {
	resp, err := http.Post(fmt.Sprintf("%s/encoreJobs/%s/cancel", ec.url, *jobId), "", nil)
	if err != nil {
		return err
	}
	resp.Body.Close()
	if resp.StatusCode != 204 {
		return errors.New(fmt.Sprintf("Failed to cancel job code=%d", resp.StatusCode))
	}
	return nil
}

func CreateJob(inputUri string, profile string) EncoreJobRequestBody {
	baseName := filepath.Base(inputUri)
	id := uuid.New()
	outputFolder := fmt.Sprintf("/tmp/%s", id)
	input := Input{
		Type: "AudioVideo",
		Uri: inputUri,
		Params: make(map[string]string),
	}
	inputs := []Input{
		input,
	}
	return EncoreJobRequestBody{
		BaseName: strings.TrimSuffix(baseName, filepath.Ext(baseName)),
		Id: &id,
		OutputFolder: outputFolder,
		Profile: profile,
		Inputs: inputs,
		ProfileParams: make(map[string]map[string]interface{}),
		LogContext: make(map[string]string),
		Priority: 50,
	}
}