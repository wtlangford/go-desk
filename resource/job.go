package resource

import (
	. "github.com/wtlangford/go-desk/types"
)

type Job struct {
	Type          *string    `json:"type,omitempty"`
	StatusMessage *string    `json:"status_message,omitempty"`
	Progress      float64    `json:"progress,omitempty"`
	CreatedAt     *Timestamp `json:"created_at,omitempty"`
	CompletedAt   *Timestamp `json:"completed_at,omitempty"`
	LastError     *string    `json:"last_error,omitempty"`
	Resource
}

func NewJob() *Job {
	job := &Job{}
	job.InitializeResource(job)
	return job
}

func (j Job) String() string {
	return Stringify(j)
}
