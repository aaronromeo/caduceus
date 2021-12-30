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
	BackgroundColor string `json:"backgroundColor,omitempty"`
	TextColor       string `json:"textColor,omitempty"`
}

type CadLabel struct {
	Id                    string        `json:"id,omitempty"`
	Name                  string        `json:"name,omitempty"`
	LabelListVisibility   string        `json:"labelListVisibility,omitempty"`
	MessageListVisibility string        `json:"messageListVisibility,omitempty"`
	MessagesTotal         int64         `json:"messagesTotal,omitempty"`
	MessagesUnread        int64         `json:"messagesUnread,omitempty"`
	ThreadsTotal          int64         `json:"threadsTotal,omitempty"`
	ThreadsUnread         int64         `json:"threadsUnread,omitempty"`
	Type                  string        `json:"type,omitempty"`
	Color                 CadLabelColor `json:"color,omitempty"`
}

func GetLabels() ([]*CadLabel, error) {
	srv, err := GetService()
	if err != nil {
		log.Printf("Unable to retrieve Gmail client: %v", err)
		return nil, err
	}

	user := "me"
	r, err := srv.Users.Labels.List(user).Do()
	if err != nil {
		log.Printf("Unable to retrieve labels: %v", err)
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
			log.Printf("Unable to retrieve labels: %v", err)
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

func CreateUserLabel(cadLabel *CadLabel) (*CadLabel, error) {
	srv, err := GetService()
	if err != nil {
		log.Printf("Unable to retrieve Gmail client: %v", err)
		return nil, err
	}

	user := "me"
	gmailLabel := cadLabel.MarshalGmail()
	label, err := srv.Users.Labels.Create(user, gmailLabel).Do()
	if err != nil {
		log.Printf("Unable to create label: %v", err)
		return nil, err
	}
	return MarshalCadLabel(label), nil
}

func DeleteUserLabel(cadLabel *CadLabel) error {
	srv, err := GetService()
	if err != nil {
		log.Printf("Unable to retrieve Gmail client: %v", err)
		return err
	}

	user := "me"
	err = srv.Users.Labels.Delete(user, cadLabel.Id).Do()
	if err != nil {
		log.Printf("Unable to delete label: %s\n%v", cadLabel.Id, err)
		return err
	}
	return nil
}

func PatchUserLabel(id string, updatedCadlabel *CadLabel) (*gmail.Label, error) {
	srv, err := GetService()
	if err != nil {
		log.Printf("Unable to update Gmail client: %v", err)
		return nil, err
	}

	label := updatedCadlabel.MarshalGmail()
	user := "me"
	r, err := srv.Users.Labels.Patch(user, id, label).Do()
	if err != nil {
		log.Printf("Unable to update label: %v", err)
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
		ThreadsTotal:          label.ThreadsTotal,
		ThreadsUnread:         label.ThreadsUnread,
		Type:                  label.Type,
		Color:                 *labelcolor,
	}

	return data
}

func (label *CadLabel) MarshalGmail() *gmail.Label {
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
		log.Printf("Unable to marshal labels to JSON: %v", err)
		return err
	}

	err = ioutil.WriteFile(labeldatafile, b, 0664)
	if err != nil {
		log.Printf("Unable to persist labels: %v", err)
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
		log.Printf("Unable to read local label data file: %v", err)
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
