package internal

import (
	"log"

	"google.golang.org/api/gmail/v1"
)

func GetFilters() ([]*gmail.Filter, error) {
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
	filters := r.Filter
	// sort.SliceStable(labels, func(i, j int) bool {
	// 	return labels[i].Name < labels[j].Name
	// })
	return filters, nil
}
