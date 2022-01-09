package internal

import (
	"log"

	"google.golang.org/api/gmail/v1"
)

const bulkLimit int = 1000

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
