// package main

// import (
// 	"fmt"
// 	"log"
// 	"os"

// 	internal "aaronromeo/mailboxorg/caduceus/internal"

// 	"google.golang.org/api/gmail/v1"
// )

// func main() {
// 	filters, err := internal.GetFilters()

// 	if err != nil {
// 		fmt.Println("Unable to retrieve filters.")
// 		return
// 	}

// 	if len(filters) == 0 {
// 		fmt.Println("No filters found.")
// 		return
// 	}
// 	fmt.Println("filters:")
// 	for _, item := range filters {
// 		criteria, _ := item.Criteria.MarshalJSON()
// 		action, _ := item.Action.MarshalJSON()
// 		fmt.Printf("%s,%s,%s\n", item.Id, criteria, action)
// 	}
// 	// dumpLabels(labels)

// 	// labels, err := internal.GetUserLabels()

// 	// if err != nil {
// 	// 	fmt.Println("Unable to retrieve labels.")
// 	// 	return
// 	// }

// 	// if len(labels) == 0 {
// 	// 	fmt.Println("No labels found.")
// 	// 	return
// 	// }
// 	// fmt.Println("Labels:")
// 	// for _, l := range labels {
// 	// 	fmt.Printf("\"%s\",%s,%s,%s\n", l.Name, l.Id, l.LabelListVisibility, l.MessageListVisibility)
// 	// }
// 	// dumpLabels(labels)

// 	// ids := []string{}

// 	// ids = []string{
// 	// 	"Label_209",
// 	// 	"Label_232",
// 	// 	"Label_39",
// 	// 	"Label_36",
// 	// 	"Label_78",
// 	// 	"Label_37",
// 	// 	"Label_32",
// 	// 	"Label_200",
// 	// 	"Label_174",
// 	// 	"Label_34",
// 	// 	"Label_30",
// 	// 	"Label_33",
// 	// 	"Label_29",
// 	// 	"Label_98",
// 	// 	"Label_184",
// 	// 	"Label_22",
// 	// 	"Label_175",
// 	// 	"Label_173",
// 	// 	"Label_26",
// 	// }
// 	// labelHideIds(ids, "labelHide")

// 	// ids = []string{
// 	// 	"Label_1147896115220560280",
// 	// 	"Label_3152417086712246767",
// 	// 	"Label_231",
// 	// 	"Label_4406302739829786055",
// 	// 	"Label_4821462302103786038",
// 	// 	"Label_8867306721709845546",
// 	// 	"Label_4326046855468876637",
// 	// }
// 	// labelHideIds(ids, "labelShowIfUnread")
// }

// func dumpLabels(labels []*gmail.Label) {
// 	f, err := os.OpenFile("labels.csv", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
// 	if err != nil {
// 		log.Fatalf("Unable to cache oauth token: %v", err)
// 	}
// 	defer f.Close()
// 	f.WriteString("Name,Id,LabelListVisibility,MessageListVisibility\n")
// 	for _, l := range labels {
// 		_, err := f.WriteString(
// 			fmt.Sprintf("\"%s\",%s,%s,%s\n", l.Name, l.Id, l.LabelListVisibility, l.MessageListVisibility),
// 		)
// 		if err != nil {
// 			fmt.Println("Unable to write labels.")
// 			panic(err)
// 		}
// 	}
// 	f.Sync()
// }

// func labelHideIds(label_ids []string, visibility string) {
// 	fmt.Printf("Updating the labels\n\n")
// 	for _, id := range label_ids {
// 		tempLabel := gmail.Label{
// 			Color: &gmail.LabelColor{BackgroundColor: "#000000", TextColor: "#ffffff"},
// 			// Id:                    "",
// 			LabelListVisibility: visibility,
// 			// MessageListVisibility: "hide",
// 			// MessagesTotal:         0,
// 			// MessagesUnread:        0,
// 			// Name:                  "",
// 			// ThreadsTotal:          0,
// 			// ThreadsUnread:         0,
// 			// Type:                  "",
// 			// ServerResponse:        googleapi.ServerResponse{},
// 			// ForceSendFields:       []string{},
// 			// NullFields:            []string{},
// 		}

// 		label, err := internal.PatchUserLabel(id, &tempLabel)
// 		if err != nil {
// 			fmt.Printf("Unable to update labels")
// 			panic(err)
// 		}

// 		fmt.Printf("%s - %s (%s - %s) \n", label.Name, label.Id, label.LabelListVisibility, label.MessageListVisibility)
// 	}
// }
