package service

import (
	"encoding/json"
	. "github.com/wtlangford/go-desk/resource"
	"net/http"
)

type JobService struct {
	client *Client
}

// Get retrieves a job.
// See Desk API: http://dev.desk.com/API/jobs/#show
func (c *JobService) Get(id string) (*Job, *http.Response, error) {
	restful := Restful{}
	job := NewJob()
	path := NewIdentityResourcePath(id, job)
	resp, err := restful.
		Get(path.Path()).
		Json(job).
		Client(c.client).
		Do()
	return job, resp, err
}

// List jobs with pagination.
// See Desk API: http://dev.desk.com/API/jobs/#list
func (c *JobService) List() (*Page, *http.Response, error) {
	restful := Restful{}
	page := new(Page)
	path := NewResourcePath(NewJob())
	resp, err := restful.
		Get(path.Path()).
		Json(page).
		Client(c.client).
		Do()
	if err != nil {
		return nil, resp, err
	}
	err = c.unravelPage(page)
	if err != nil {
		return nil, nil, err
	}
	return page, resp, err
}

func (c *JobService) unravelPage(page *Page) error {
	jobs := new([]Job)
	err := json.Unmarshal(*page.Embedded.RawEntries, &jobs)
	if err != nil {
		return err
	}
	page.Embedded.Entries = make([]interface{}, len(*jobs))
	for i, v := range *jobs {
		v.InitializeResource(v)
		page.Embedded.Entries[i] = interface{}(v)
	}
	page.Embedded.RawEntries = nil
	return err
}
