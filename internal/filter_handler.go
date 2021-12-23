package internal

import (
	"encoding/json"
	"io/ioutil"
	"log"

	"google.golang.org/api/gmail/v1"
)

const filterdatafile string = "data/filters.json"

type CadCriteria struct {
	From           string
	To             string
	Subject        string
	Query          string
	NegatedQuery   string
	HasAttachment  bool
	ExcludeChats   bool
	Size           int64
	SizeComparison string
}

type CadAction struct {
	AddLabelIds    []string
	RemoveLabelIds []string
	Forward        string
}

type CadFilterMeta struct {
	Labels []CadLabel
}

type CadFilter struct {
	Id       string
	Criteria CadCriteria
	Action   CadAction
	Meta     CadFilterMeta
}

func GetFilters() ([]*CadFilter, error) {
	srv, err := GetService()
	if err != nil {
		log.Fatalf("Unable to retrieve Gmail client: %v", err)
		return nil, err
	}

	user := "me"
	r, err := srv.Users.Settings.Filters.List(user).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve filters: %v", err)
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
	b, err := json.MarshalIndent(filters, "", "  ")
	if err != nil {
		log.Fatalf("Unable to marshal filters: %v", err)
		return err
	}

	err = ioutil.WriteFile(filterdatafile, b, 0664)
	if err != nil {
		log.Fatalf("Unable to persist filters: %v", err)
		return err
	}
	return nil
}
