package resource

import (
	. "github.com/wtlangford/go-desk/types"
)

type Case struct {
	ExternalID      *string                `json:"external_id,omitempty"`
	Type            *string                `json:"type,omitempty"`
	Status          *string                `json:"status,omitempty"`
	Description     *string                `json:"description,omitempty"`
	Subject         *string                `json:"subject,omitempty"`
	Blurb           *string                `json:"blurb,omitempty"`
	Language        *string                `json:"language,omitempty"`
	Priority        *int                   `json:"priority,omitempty"`
	Labels          []string               `json:"labels,omitempty"`
	LabelIDs        []int                  `json:"label_ids,omitempty"`
	SuppressRules   *bool                  `json:"suppress_rules,omitempty"`
	CustomFields    map[string]interface{} `json:"custom_fields,omitempty"`
	LockedUntil     *Timestamp             `json:"locked_until",omitempty`
	CreatedAt       *Timestamp             `json:"created_at,omitempty"`
	UpdatedAt       *Timestamp             `json:"updated_at,omitempty"`
	ChangedAt       *Timestamp             `json:"changed_at,omitempty"`
	ReceivedAt      *Timestamp             `json:"received_at,omitempty"`
	ActiveAt        *Timestamp             `json:"active_at,omitempty"`
	OpenedAt        *Timestamp             `json:"opened_at,omitempty"`
	FirstOpenedAt   *Timestamp             `json:"first_opened_at,omitempty"`
	ResolvedAt      *Timestamp             `json:"resolved_at,omitempty"`
	FirstResolvedAt *Timestamp             `json:"first_resolved_at,omitempty"`
	Message         *Message               `json:"message,omitempty"`
	Resource
}

func NewCase() *Case {
	caze := &Case{}
	caze.InitializeResource(caze)
	return caze
}

func (c Case) String() string {
	return Stringify(c)
}
