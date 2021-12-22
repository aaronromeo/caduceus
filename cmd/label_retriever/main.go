package main

import (
	"fmt"

	internal "aaronromeo/mailboxorg/caduceus/internal"
)

func main() {
	labels, err := internal.GetUserLabels()

	if err != nil {
		fmt.Println("Unable to retrieve labels.")
		return
	}

	if len(labels) == 0 {
		fmt.Println("No labels found.")
		return
	}
	fmt.Println("Labels:")
	for _, l := range labels {
		fmt.Printf("%s - %s (%s - %s) \n", l.Name, l.Type, l.LabelListVisibility, l.MessageListVisibility)
	}
}
