package internal

import (
	"context"
	"io/ioutil"
	"log"
	"sort"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

const user string = "user"

func GetLabels() ([]*gmail.Label, error) {
	srv, err := getService()
	if err != nil {
		log.Fatalf("Unable to retrieve Gmail client: %v", err)
		return nil, err
	}

	user := "me"
	r, err := srv.Users.Labels.List(user).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve labels: %v", err)
		return nil, err
	}
	labels := r.Labels
	sort.SliceStable(labels, func(i, j int) bool {
		return labels[i].Name < labels[j].Name
	})
	return labels, nil
}

func GetUserLabels() ([]*gmail.Label, error) {
	labels, err := GetLabels()
	userLabels := []*gmail.Label{}

	for i := range labels {
		if labels[i].Type == user {
			userLabels = append(userLabels, labels[i])
		}
	}

	return userLabels, err
}

func PatchUserLabel(id string, label *gmail.Label) (*gmail.Label, error) {
	srv, err := getService()
	if err != nil {
		log.Fatalf("Unable to update Gmail client: %v", err)
		return nil, err
	}

	user := "me"
	r, err := srv.Users.Labels.Patch(user, id, label).Do()
	if err != nil {
		log.Fatalf("Unable to update label: %v", err)
		return nil, err
	}

	return r, nil
}

func getService() (*gmail.Service, error) {
	ctx := context.Background()
	b, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
		return nil, err
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, gmail.GmailModifyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
		return nil, err
	}
	client := GetClient(config)

	return gmail.NewService(ctx, option.WithHTTPClient(client))
}
