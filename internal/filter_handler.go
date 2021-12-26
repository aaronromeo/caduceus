package internal

import (
	"encoding/json"
	"io/ioutil"
	"log"

	"google.golang.org/api/gmail/v1"
)

const filterdatafile string = "data/filters.json"

type CadCriteria struct {
	From           string `json:"from,omitempty"`
	To             string `json:"to,omitempty"`
	Subject        string `json:"subject,omitempty"`
	Query          string `json:"query,omitempty"`
	NegatedQuery   string `json:"negativeQuery,omitempty"`
	HasAttachment  bool   `json:"hasAttachment,omitempty"`
	ExcludeChats   bool   `json:"excludeChats,omitempty"`
	Size           int64  `json:"size,omitempty"`
	SizeComparison string `json:"sizeComparison,omitempty"`
}

type CadAction struct {
	AddLabelIds    []string `json:"addLabelIds,omitempty"`
	RemoveLabelIds []string `json:"removeLabelIds,omitempty"`
	Forward        string   `json:"forward,omitempty"`
}

type CadFilterMeta struct {
	Labels []CadLabel `json:"labels,omitempty"`
}

type CadFilter struct {
	Id       string        `json:"id,omitempty"`
	Criteria CadCriteria   `json:"criteria,omitempty"`
	Action   CadAction     `json:"action,omitempty"`
	Meta     CadFilterMeta `json:"meta,omitempty"`
}

func GetFilters() ([]*CadFilter, error) {
	srv, err := GetService()
	if err != nil {
		log.Printf("Unable to retrieve Gmail client: %v", err)
		return nil, err
	}

	user := "me"
	r, err := srv.Users.Settings.Filters.List(user).Do()
	if err != nil {
		log.Printf("Unable to retrieve filters: %v", err)
		return nil, err
	}
	filters := []*CadFilter{}
	for _, filter := range r.Filter {
		filters = append(filters, MarshalCadFilter(filter))
	}
	return filters, nil
}

func MarshalCadFilter(filter *gmail.Filter) *CadFilter {
	critera := &CadCriteria{
		From:           filter.Criteria.From,
		To:             filter.Criteria.To,
		Subject:        filter.Criteria.Subject,
		Query:          filter.Criteria.Query,
		NegatedQuery:   filter.Criteria.NegatedQuery,
		HasAttachment:  filter.Criteria.HasAttachment,
		ExcludeChats:   filter.Criteria.ExcludeChats,
		Size:           filter.Criteria.Size,
		SizeComparison: filter.Criteria.SizeComparison,
	}
	action := &CadAction{
		AddLabelIds:    filter.Action.AddLabelIds,
		RemoveLabelIds: filter.Action.RemoveLabelIds,
		Forward:        filter.Action.Forward,
	}
	data := &CadFilter{
		Id:       filter.Id,
		Criteria: *critera,
		Action:   *action,
	}

	return data
}

func SaveLocalFilters(filters []*CadFilter) error {
	localLabels, err := ReadLocalLabels()
	if err != nil {
		log.Printf("Unable to read local labels: %v", err)
		return err
	}

	labelmap := make(map[string]CadLabel)
	for _, label := range localLabels {
		labelmap[label.Id] = label
	}

	for i, filter := range filters {
		meta := CadFilterMeta{}
		for _, labelId := range filter.Action.AddLabelIds {
			meta.Labels = append(meta.Labels, labelmap[labelId])
		}
		for _, labelId := range filter.Action.RemoveLabelIds {
			meta.Labels = append(meta.Labels, labelmap[labelId])
		}
		filters[i].Meta = meta
	}

	b, err := json.MarshalIndent(filters, "", "  ")
	if err != nil {
		log.Printf("Unable to marshal filters: %v", err)
		return err
	}

	err = ioutil.WriteFile(filterdatafile, b, 0664)
	if err != nil {
		log.Printf("Unable to persist filters: %v", err)
		return err
	}
	return nil
}
