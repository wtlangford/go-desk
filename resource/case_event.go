package resource

import (
	. "github.com/wtlangford/go-desk/types"
)

type CaseEvent struct {
	Type      *string                  `json:"type,omitempty"`
	Context   *string                  `json:"context,omitempty"`
	CreatedAt *Timestamp               `json:"created_at,omitempty"`
	Changes   []map[string]interface{} `json:"changes,omitempty"`
	Resource
}

func NewCaseEvent() *CaseEvent {
	case_event := &CaseEvent{}
	return case_event
}

func (c CaseEvent) String() string {
	return Stringify(c)
}
