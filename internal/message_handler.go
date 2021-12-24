package internal

import (
	"log"
)

func GetMessagesIDsByLableIDs(labels []*CadLabel) ([]string, error) {
	srv, err := GetService()
	if err != nil {
		log.Fatalf("Unable to retrieve Gmail client: %v", err)
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
		r, err := srv.Users.Messages.List(user).
			MaxResults(500).
			PageToken(pageToken).
			LabelIds(labelIds...).
			Do()
		if err != nil {
			log.Fatalf("Unable to retrieve messages: %v", err)
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
