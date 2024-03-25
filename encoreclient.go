package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"io"
	"net/http"
	"path/filepath"
	"strings"
)

type EncoreClient struct {
	url        string
	client     *http.Client
	authHeader string
}

func NewEncoreClient(url string, authHeader string) *EncoreClient {
	ec := EncoreClient{url, &http.Client{}, authHeader}
	return &ec
}

func (ec *EncoreClient) get(url string) (*http.Response, error) {
	req, err := ec.newRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	return ec.client.Do(req)
}

func (ec *EncoreClient) post(url string, contentType string, body io.Reader) (*http.Response, error) {
	req, err := ec.newRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	return ec.client.Do(req)
}

func (ec *EncoreClient) newRequest(method string, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	if len(ec.authHeader) > 0 {
		ah := strings.Split(ec.authHeader, ":")
		req.Header.Set(ah[0], ah[1])
	}
	return req, nil
}

func (ec *EncoreClient) getJobs() (*PagedModelEntityModelEncoreJob, error) {
	resp, err := ec.get(ec.url + "/encoreJobs?sort=createdDate,desc")
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = resp.Body.Close()
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
		return nil, err
	}
	return &jobPage, nil
}

func (ec *EncoreClient) PostJob(job EncoreJobRequestBody) error {
	jsonVal, _ := json.Marshal(job)
	resp, err := ec.post(ec.url+"/encoreJobs", "application/json", bytes.NewBuffer(jsonVal))
	if err != nil {
		return err
	}
	if resp.StatusCode != 201 {
		return errors.New(fmt.Sprintf("Failed to post job code=%d", resp.StatusCode))
	}
	return nil
}

func (ec *EncoreClient) CancelJob(jobId *uuid.UUID) error {
	resp, err := ec.post(fmt.Sprintf("%s/encoreJobs/%s/cancel", ec.url, *jobId), "", nil)
	if err != nil {
		return err
	}
	_ = resp.Body.Close()
	if resp.StatusCode != 204 {
		return errors.New(fmt.Sprintf("Failed to cancel job code=%d", resp.StatusCode))
	}
	return nil
}

func (ec *EncoreClient) DeleteJob(jobId *uuid.UUID) error {
	req, err := ec.newRequest("DELETE", fmt.Sprintf("%s/encoreJobs/%s", ec.url, *jobId), nil)
	if err != nil {
		return err
	}
	resp, err := ec.client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 204 {
		return errors.New(fmt.Sprintf("Failed to delete job code=%d", resp.StatusCode))
	}
	return nil
}

func NewEncoreJobRequestBody(inputUri string, profile string) EncoreJobRequestBody {
	baseName := filepath.Base(inputUri)
	id := uuid.New()
	outputFolder := fmt.Sprintf("/tmp/%s", id)
	input := Input{
		Type:   "AudioVideo",
		Uri:    inputUri,
		Params: make(map[string]string),
	}
	inputs := []Input{
		input,
	}
	return EncoreJobRequestBody{
		BaseName:      strings.TrimSuffix(baseName, filepath.Ext(baseName)),
		Id:            &id,
		OutputFolder:  outputFolder,
		Profile:       profile,
		Inputs:        inputs,
		ProfileParams: make(map[string]map[string]interface{}),
		LogContext:    make(map[string]string),
		Priority:      50,
	}
}
