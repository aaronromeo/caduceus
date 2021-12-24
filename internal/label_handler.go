package internal

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"sort"

	"google.golang.org/api/gmail/v1"
)

const user string = "user"
const labeldatafile string = "data/labels.json"

type CadLabelColor struct {
	BackgroundColor string
	TextColor       string
}

type CadLabel struct {
	Id                    string
	Name                  string
	LabelListVisibility   string
	MessageListVisibility string
	MessagesTotal         int64
	MessagesUnread        int64
	Type                  string
	Color                 CadLabelColor
}

func GetLabels() ([]*CadLabel, error) {
	srv, err := GetService()
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
	cadlabels := []*CadLabel{}
	for _, label := range labels {
		r, err := srv.Users.Labels.Get(user, label.Id).Do()
		if err != nil {
			log.Fatalf("Unable to retrieve labels: %v", err)
			return nil, err
		}

		cadlabels = append(cadlabels, MarshalCadLabel(r))
	}
	return cadlabels, nil
}

func GetUserLabels() ([]*CadLabel, error) {
	labels, err := GetLabels()
	userLabels := []*CadLabel{}

	for i := range labels {
		if labels[i].Type == user {
			userLabels = append(userLabels, labels[i])
		}
	}

	return userLabels, err
}

func PatchUserLabel(id string, cadlabel *CadLabel) (*gmail.Label, error) {
	srv, err := GetService()
	if err != nil {
		log.Fatalf("Unable to update Gmail client: %v", err)
		return nil, err
	}

	label := MarshalGmailLabel(cadlabel)
	user := "me"
	r, err := srv.Users.Labels.Patch(user, id, label).Do()
	if err != nil {
		log.Fatalf("Unable to update label: %v", err)
		return nil, err
	}

	return r, nil
}

func MarshalCadLabel(label *gmail.Label) *CadLabel {
	labelcolor := &CadLabelColor{}
	if label.Color != nil {
		labelcolor = &CadLabelColor{
			BackgroundColor: label.Color.BackgroundColor,
			TextColor:       label.Color.TextColor,
		}
	}
	data := &CadLabel{
		Id:                    label.Id,
		Name:                  label.Name,
		LabelListVisibility:   label.LabelListVisibility,
		MessageListVisibility: label.MessageListVisibility,
		MessagesTotal:         label.MessagesTotal,
		MessagesUnread:        label.MessagesUnread,
		Type:                  label.Type,
		Color:                 *labelcolor,
	}

	return data
}

func MarshalGmailLabel(label *CadLabel) *gmail.Label {
	data := &gmail.Label{}
	if label.Id != "" {
		data.Id = label.Id
	}
	if label.Name != "" {
		data.Name = label.Name
	}
	if label.LabelListVisibility != "" {
		data.LabelListVisibility = label.LabelListVisibility
	}
	if label.MessageListVisibility != "" {
		data.MessageListVisibility = label.MessageListVisibility
	}
	if label.Color.BackgroundColor != "" || label.Color.TextColor != "" {
		data.Color = &gmail.LabelColor{
			BackgroundColor: label.Color.BackgroundColor,
			TextColor:       label.Color.TextColor,
		}
	}

	return data
}

func SaveLocalLabels(labels []*CadLabel) error {
	b, err := json.MarshalIndent(labels, "", "  ")
	if err != nil {
		log.Fatalf("Unable to marshal labels to JSON: %v", err)
		return err
	}

	err = ioutil.WriteFile(labeldatafile, b, 0664)
	if err != nil {
		log.Fatalf("Unable to persist labels: %v", err)
		return err
	}
	return nil
}

func ReadLocalLabels() ([]CadLabel, error) {
	if !fileExists(labeldatafile) {
		return []CadLabel{}, nil
	}

	b, err := ioutil.ReadFile(labeldatafile)
	if err != nil {
		log.Fatalf("Unable to read local label data file: %v", err)
		return nil, err
	}
	var labels []CadLabel
	if err := json.Unmarshal(b, &labels); err != nil {
		return []CadLabel{}, err
	}

	return labels, nil
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
