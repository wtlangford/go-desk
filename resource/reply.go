package resource

import (
	. "github.com/wtlangford/go-desk/types"
)

type Reply struct {
	Direction        *string    `json:"direction,omitempty"`
	Body             *string    `json:"body,omitempty"`
	BodyText         *string    `json:"body_text,omitempty"`
	BodyHtml         *string    `json:"body_html,omitempty"`
	Headers          *string    `json:"headers,omitempty"`
	HeadersRaw       *string    `json:"headers_raw,omitempty"`
	Status           *string    `json:"status,omitempty"`
	Subject          *string    `json:"subject,omitempty"`
	To               *string    `json:"to,omitempty"`
	From             *string    `json:"from,omitempty"`
	Type             *string    `json:"type,omitempty"`
	Cc               *string    `json:"cc,omitempty"`
	Bcc              *string    `json:"bcc,omitempty"`
	ClientType       *string    `json:"client_type,omitempty"`
	FromFacebookName *string    `json:"from_facebook_name,omitempty"`
	PublicUrl        *string    `json:"public_url,omitempty"`
	IsBestAnswer     *bool      `json:"is_best_answer,omitempty"`
	Rating           *float32   `json:"rating,omitempty"`
	RatingCount      *int       `json:"rating_count,omitempty"`
	RatingScore      *int       `json:"rating_score,omitempty"`
	EnteredAt        *Timestamp `json:"entered_at,omitempty"`
	HiddentAt        *Timestamp `json:"hidden_at,omitempty"`
	CreatedAt        *Timestamp `json:"created_at,omitempty"`
	UpdatedAt        *Timestamp `json:"updated_at,omitempty"`
	Resource
}

func NewReply() *Reply {
	reply := &Reply{}
	reply.InitializeResource(reply)
	reply.requireSelfId = true
	return reply
}

func (c Reply) String() string {
	return Stringify(c)
}
