package internal

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"sort"
	"time"

	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/googleapi"
)

const filterdatafile string = "data/filters.json"
const consolidatedfilterdatafile string = "data/consolidated_filters.json"

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
	Id       string         `json:"id,omitempty"`
	Criteria *CadCriteria   `json:"criteria,omitempty"`
	Action   *CadAction     `json:"action,omitempty"`
	Meta     *CadFilterMeta `json:"meta,omitempty"`
}

type CadConsolidatedAction struct {
	AddLabelIds    []string `json:"addLabelIds,omitempty"`
	RemoveLabelIds []string `json:"removeLabelIds,omitempty"`
	Forwards       []string `json:"forwards,omitempty"`
}

type CadConsolidatedFilter struct {
	Ids      *[]string              `json:"ids,omitempty"`
	Criteria *CadCriteria           `json:"criteria,omitempty"`
	Action   *CadConsolidatedAction `json:"action,omitempty"`
	Meta     *CadFilterMeta         `json:"meta,omitempty"`
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
	sort.SliceStable(filters, func(i, j int) bool {
		fi := fmt.Sprintf(
			"%s%s%s",
			filters[i].Criteria.From,
			filters[i].Criteria.Query,
			filters[i].Criteria.To,
		)
		fj := fmt.Sprintf(
			"%s%s%s",
			filters[j].Criteria.From,
			filters[j].Criteria.Query,
			filters[j].Criteria.To,
		)

		return fi < fj
	})

	return filters, nil
}

func SelectArchiveFilters() ([]CadFilter, error) {
	filters, err := ReadLocalFilters()
	if err != nil {
		log.Printf("Unable to read local filters file: %v", err)
		return nil, err
	}

	archiveFilters := []CadFilter{}
	for _, filter := range filters {
		for _, label := range filter.Action.RemoveLabelIds {
			if label == "INBOX" {
				archiveFilters = append(archiveFilters, filter)
				break
			}
		}
	}

	return archiveFilters, nil
}

func GetFilter(cadFilter *CadFilter) (*CadFilter, error) {
	srv, err := GetService()
	if err != nil {
		log.Printf("Unable to retrieve Gmail client: %v", err)
		return nil, err
	}

	user := "me"
	filter, err := srv.Users.Settings.Filters.Get(user, cadFilter.Id).Do()
	if err != nil {
		log.Printf("Unable to retrieve filter: %s\n%v", cadFilter.Id, err)
		return nil, err
	}
	return MarshalCadFilter(filter), nil
}

func CreateFilter(cadFilter *CadFilter) (*CadFilter, error) {
	srv, err := GetService()
	if err != nil {
		log.Printf("Unable to retrieve Gmail client: %v", err)
		return nil, err
	}

	user := "me"
	gmailFilter := cadFilter.MarshalGmail()
	retry := 3
	for retry > 0 {
		filter, err := srv.Users.Settings.Filters.Create(user, gmailFilter).Do()
		if err != nil {
			gErr, ok := err.(*googleapi.Error)

			log.Printf("Unable to create filter: %v\n", err)
			if ok && (gErr.Code == 503 || gErr.Code == 400) {
				retry -= 1
				log.Printf("Retrying after a nap...\n")
				time.Sleep(60 * time.Second * time.Duration(10/retry))
			} else {
				return nil, err
			}
		} else {
			return MarshalCadFilter(filter), nil
		}
	}
	return nil, err
}

func DeleteFilter(cadFilter *CadFilter) error {
	srv, err := GetService()
	if err != nil {
		log.Printf("Unable to retrieve Gmail client: %v", err)
		return err
	}

	user := "me"
	err = srv.Users.Settings.Filters.Delete(user, cadFilter.Id).Do()
	if err != nil {
		log.Printf("Unable to delete filter: %s\n%v", cadFilter.Id, err)
		return err
	}
	return nil
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
	action := &CadAction{}
	if filter.Action != nil {
		action.AddLabelIds = filter.Action.AddLabelIds
		action.RemoveLabelIds = filter.Action.RemoveLabelIds
		action.Forward = filter.Action.Forward
	}
	data := &CadFilter{
		Id:       filter.Id,
		Criteria: critera,
		Action:   action,
	}

	return data
}

func (filter *CadFilter) MarshalGmail() *gmail.Filter {
	data := &gmail.Filter{Id: filter.Id}

	if filter.Action != nil {
		data.Action = &gmail.FilterAction{
			AddLabelIds:    filter.Action.AddLabelIds,
			RemoveLabelIds: filter.Action.RemoveLabelIds,
			Forward:        filter.Action.Forward,
		}
	}

	if filter.Criteria != nil {
		data.Criteria = &gmail.FilterCriteria{
			To:             filter.Criteria.To,
			From:           filter.Criteria.From,
			Subject:        filter.Criteria.Subject,
			Query:          filter.Criteria.Query,
			ExcludeChats:   filter.Criteria.ExcludeChats,
			HasAttachment:  filter.Criteria.HasAttachment,
			NegatedQuery:   filter.Criteria.NegatedQuery,
			Size:           filter.Criteria.Size,
			SizeComparison: filter.Criteria.SizeComparison,
		}
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
		meta := &CadFilterMeta{}
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

	consolidatedFilters, _ := consolidateFiltersByCriteria(filters)
	for i, filter := range consolidatedFilters {
		meta := &CadFilterMeta{}
		for _, labelId := range filter.Action.AddLabelIds {
			meta.Labels = append(meta.Labels, labelmap[labelId])
		}
		for _, labelId := range filter.Action.RemoveLabelIds {
			meta.Labels = append(meta.Labels, labelmap[labelId])
		}
		consolidatedFilters[i].Meta = meta
	}

	b, err = json.MarshalIndent(consolidatedFilters, "", "  ")
	if err != nil {
		log.Printf("Unable to marshal filters: %v", err)
		return err
	}

	err = ioutil.WriteFile(consolidatedfilterdatafile, b, 0664)
	if err != nil {
		log.Printf("Unable to persist filters: %v", err)
		return err
	}

	return nil
}

func ReadLocalFilters() ([]CadFilter, error) {
	if !fileExists(filterdatafile) {
		return []CadFilter{}, nil
	}

	b, err := ioutil.ReadFile(filterdatafile)
	if err != nil {
		log.Printf("Unable to read local filter data file: %v", err)
		return nil, err
	}
	var filters []CadFilter
	if err := json.Unmarshal(b, &filters); err != nil {
		return []CadFilter{}, err
	}

	return filters, nil
}

func consolidateFiltersByCriteria(filters []*CadFilter) ([]*CadConsolidatedFilter, error) {
	filterIdMap := map[string]*CadConsolidatedFilter{}

	for _, filter := range filters {
		criteriaKey := criteriaKey(*filter.Criteria)

		if filterIdMap[criteriaKey] != nil {
			consolidatedFilter := *filterIdMap[criteriaKey]
			ids := *consolidatedFilter.Ids
			action := *consolidatedFilter.Action
			ids = append(ids, filter.Id)
			action.AddLabelIds = append(action.AddLabelIds, filter.Action.AddLabelIds...)
			action.RemoveLabelIds = append(action.RemoveLabelIds, filter.Action.RemoveLabelIds...)
			if filter.Action.Forward != "" {
				action.Forwards = append(action.Forwards, filter.Action.Forward)
			}

			filterIdMap[criteriaKey] = &CadConsolidatedFilter{
				Ids:      &ids,
				Action:   &action,
				Criteria: filterIdMap[criteriaKey].Criteria,
			}
		} else {
			ids := []string{filter.Id}
			forwards := []string{}
			if filter.Action.Forward != "" {
				forwards = append(forwards, filter.Action.Forward)
			}
			action := CadConsolidatedAction{
				AddLabelIds:    filter.Action.AddLabelIds,
				RemoveLabelIds: filter.Action.RemoveLabelIds,
				Forwards:       forwards,
			}
			filterIdMap[criteriaKey] = &CadConsolidatedFilter{
				Ids:      &ids,
				Action:   &action,
				Criteria: filter.Criteria,
			}
		}
	}

	consolidatedFilters := []*CadConsolidatedFilter{}
	for _, ccf := range filterIdMap {
		consolidatedFilters = append(consolidatedFilters, ccf)
	}
	sort.SliceStable(consolidatedFilters, func(i, j int) bool {
		fi := criteriaKey(*consolidatedFilters[i].Criteria)
		fj := criteriaKey(*consolidatedFilters[j].Criteria)

		return fi < fj
	})

	return consolidatedFilters, nil
}

func criteriaKey(criteria CadCriteria) string {
	return fmt.Sprintf(
		"%s|%s|%s|%s|%d|%s|%t|%t",
		criteria.From,
		criteria.To,
		criteria.Query,
		criteria.NegatedQuery,
		criteria.Size,
		criteria.SizeComparison,
		criteria.HasAttachment,
		criteria.ExcludeChats,
	)
}
