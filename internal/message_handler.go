package internal

import (
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"time"

	"google.golang.org/api/gmail/v1"
)

const bulkLimit int = 1000

type CadCriteraAndSampleMessage struct {
	Criteria      *CadCriteria
	SampleMessage *gmail.Message
}

func GetMessageCriteriaForUnsubscribe(until time.Time) ([]*CadCriteraAndSampleMessage, error) {
	srv, err := GetService()
	if err != nil {
		log.Printf("Unable to retrieve Gmail client: %v", err)
		return nil, err
	}

	user := "me"
	messageSearchCriteria := map[string]*CadCriteraAndSampleMessage{}
	moreResults := true
	pageToken := ""
	headerNamesMap := map[string]bool{}
	for moreResults {
		fmt.Printf("+unsubscribe in:INBOX after:%d\n", until.Unix())
		r, err := srv.Users.Messages.List(user).
			MaxResults(500).
			PageToken(pageToken).
			Q(fmt.Sprintf("+unsubscribe in:INBOX after:%d", until.Unix())).
			Do()
		if err != nil {
			log.Printf("Unable to retrieve messages: %v", err)
			return nil, err
		}
		for _, messageFragment := range r.Messages {
			// Previously tried Regex
			// <a.*?href.*?>[\S]*?[uU]nsubscribe[^<]?<\/a>
			// <a.*?href.*?>[\S]*?[uU]nsubscribe[\S]*?<\/a>
			// <a.*?href.*?>[\S]*?[uU]nsubscribe[\S\W]*?<\/a>
			regUnsubscribe, _ := regexp.Compile(`<a.*?href.*?>[\S\W]*?[uU]nsubscribe[\S\W]*?<\/a>`)
			message, err := srv.Users.Messages.Get(user, messageFragment.Id).Format("full").Do()
			if err != nil {
				log.Printf("Unable to retrieve message: %s %v", messageFragment.Id, err)
				return nil, err
			}

			criteria := &CadCriteria{}
			for _, part := range message.Payload.Parts {
				if part.MimeType == "text/html" {
					data, _ := base64.URLEncoding.DecodeString(part.Body.Data)
					html := string(data)
					if regUnsubscribe.MatchString(html) {
						for _, header := range message.Payload.Headers {
							headerNamesMap[strings.ToLower(header.Name)] = true

							switch strings.ToLower(header.Name) {
							case "from":
								regFrom := regexp.MustCompile(`.*<([^>]*)>\z`)
								var from string
								parts := regFrom.FindStringSubmatch(header.Value)
								if len(parts) == 0 {
									from = header.Value
								} else {
									from = parts[1]
								}
								criteria.From = from
							case "to":
								criteria.To = header.Value
							case "list-id":
								regListId := regexp.MustCompile(`.*list <([^>]*)>\z`)
								var listId string
								parts := regListId.FindStringSubmatch(header.Value)
								if len(parts) == 0 {
									listId = header.Value
								} else {
									listId = parts[1]
								}
								criteria.Query = fmt.Sprintf("list:\\\"%s\\\"", listId)
							}
						}
					} else {
						fmt.Println(html)
					}
				} else if part.MimeType == "text/plain" {
					// fmt.Printf("Skipping plain text message\n")
				} else {
					fmt.Printf("\t-> Message found of type %s\n", part.MimeType)
				}
			}
			if criteria.Query != "" {
				messageSearchCriteria["Query: "+criteria.Query] = &CadCriteraAndSampleMessage{
					Criteria:      criteria,
					SampleMessage: message,
				}
			} else if criteria.From != "" {
				messageSearchCriteria["From: "+criteria.From] = &CadCriteraAndSampleMessage{
					Criteria:      criteria,
					SampleMessage: message,
				}
			} else if criteria.To != "" {
				messageSearchCriteria["To: "+criteria.To] = &CadCriteraAndSampleMessage{
					Criteria:      criteria,
					SampleMessage: message,
				}
			}

		}
		headerNameKeys := []string{}
		for headerName := range headerNamesMap {
			headerNameKeys = append(headerNameKeys, headerName)
		}
		sort.Strings(headerNameKeys)
		fmt.Printf("Found %d messages\n", len(r.Messages))
		fmt.Printf("Filtered %d messages\n", len(messageSearchCriteria))

		if r.NextPageToken != "" {
			pageToken = r.NextPageToken
		} else {
			moreResults = false
		}
	}
	criteria := make([]*CadCriteraAndSampleMessage, 0, len(messageSearchCriteria))
	for _, value := range messageSearchCriteria {
		criteria = append(criteria, value)
	}
	return criteria, nil
}

func GetMessageIDsInInboxByFilterCriteria(filter *CadFilter) ([]string, error) {
	labelInbox := []*CadLabel{}
	labelInbox = append(labelInbox, &CadLabel{
		Id: "INBOX", Name: "INBOX",
	})

	extraCriteria := false
	q := "has:nouserlabels"
	criteriaValue := reflect.ValueOf(filter.Criteria).Elem()
	for i := 0; i < criteriaValue.NumField(); i++ {
		if criteriaValue.Field(i).String() != "" {
			switch criteriaValue.Type().Field(i).Name {
			case "Query":
				extraCriteria = true
				q = fmt.Sprintf("%s \"%s\"", q, criteriaValue.Field(i))
			case "HasAttachment":
			case "ExcludeChats":
			case "Size":
			default:
				extraCriteria = true
				q = fmt.Sprintf("%s %s=\"%s\"", q, strings.ToLower(criteriaValue.Type().Field(i).Name), criteriaValue.Field(i))
			}
		}
	}

	if !extraCriteria {
		return nil, errors.New("no filter criteria")
	}

	fmt.Println(q) // TODO: Remove

	returnIDs, err := GetMessagesIDsByLabelIDs(labelInbox, &q)
	return returnIDs, err
}

func GetMessagesIDsByLabelIDs(labels []*CadLabel, query *string) ([]string, error) {
	srv, err := GetService()
	if err != nil {
		log.Printf("Unable to retrieve Gmail client: %v", err)
		return nil, err
	}

	user := "me"
	labelIds := []string{}
	for _, label := range labels {
		labelIds = append(labelIds, label.Id)
	}

	returnIDs := []string{}
	moreResults := true
	pageToken := ""
	for moreResults {
		req := srv.Users.Messages.List(user).
			MaxResults(500).
			PageToken(pageToken).
			LabelIds(labelIds...)

		if query != nil {
			req.Q(*query)
		}

		r, err := req.Do()
		if err != nil {
			log.Printf("Unable to retrieve messages: %v", err)
			return nil, err
		}
		for _, message := range r.Messages {
			returnIDs = append(returnIDs, message.Id)
		}

		if r.NextPageToken != "" {
			pageToken = r.NextPageToken
		} else {
			moreResults = false
		}
	}
	return returnIDs, nil
}

func BulkUpdateMessageLabels(messageIds []string, addLabelIds []string, removeLabelIds []string) error {
	srv, err := GetService()
	if err != nil {
		log.Printf("Unable to retrieve Gmail client: %v", err)
		return err
	}

	for i := 0; i < len(messageIds); i += bulkLimit {
		batchIds := messageIds[i:min(i+bulkLimit, len(messageIds))]
		user := "me"

		req := &gmail.BatchModifyMessagesRequest{Ids: batchIds, AddLabelIds: addLabelIds, RemoveLabelIds: removeLabelIds}

		if err = srv.Users.Messages.BatchModify(user, req).Do(); err != nil {
			log.Printf("Unable to modify messages: %v", err)
			return err
		}
	}

	return nil
}

func min(a int, b int) int {
	if a <= b {
		return a
	}
	return b
}
