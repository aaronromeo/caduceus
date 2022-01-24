/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	internal "aaronromeo/mailboxorg/caduceus/internal"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

// doctorCmd represents the doctor command
var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: runDoctor,
}

const create_filter string = "Create Filter"
const skip string = "Skip"
const ignore string = "Ignore"
const neverImportant string = "Never mark it as important"
const alwaysImportant string = "Always mark it as important"

func runDoctor(cmd *cobra.Command, args []string) {
	FetchLabels()
	// FetchFilters()

	// fmt.Println("Analyzing results...")
	// err := emptyLabelMigrations()
	// if err != nil {
	// 	panic(err)
	// }

	err := unsubscribeMigrations()
	if err != nil {
		panic(err)
	}
}

func emptyLabel(l internal.CadLabel) bool {
	// For some reason MessagesTotal and ThreadsTotal come back empty
	// in spite of there being messages
	return l.MessagesTotal == 0 &&
		l.MessagesUnread == 0 &&
		l.ThreadsTotal == 0 &&
		l.ThreadsUnread == 0
}

func emptyLabelMigrations() error {
	localLabels, err := internal.ReadLocalLabels()
	if err != nil {
		log.Printf("Unable to read local labels: %v", err)
		return err
	}

	emptyLabels := []internal.CadLabel{}
	nestedLabelLookup := map[string][]internal.CadLabel{}

	for _, label := range localLabels {
		labelParts := strings.Split(label.Name, "/")
		for i := range labelParts {
			key := strings.Join(labelParts[:i+1], "/")
			nestedLabelLookup[key] = append(nestedLabelLookup[key], label)
		}
		if emptyLabel(label) {
			emptyLabels = append(emptyLabels, label)
		}
	}

	emptyLabelMigrations := []internal.CadRawMigration{}
	for _, label := range emptyLabels {
		if len(nestedLabelLookup[label.Name]) == 1 {
			operation := internal.DeleteLabelMigration
			note := fmt.Sprintf("%s: Empty Label identified by the doctor", label.Name)
			labelId := label.Id
			labelMigration := internal.CadRawMigration{
				Operation: &operation,
				Details: internal.CadDeleteLabelMigration{
					Id: &labelId,
				},
				Note: &note,
			}

			emptyLabelMigrations = append(emptyLabelMigrations, labelMigration)
		} else {
			fmt.Printf("Skipping parent label %s\n", label.Name)
		}
	}
	return internal.CreateMigrationFile(&emptyLabelMigrations)
}

func unsubscribeMigrations() error {
	criteriaAndSampleMessages, err := internal.GetMessageCriteriaForUnsubscribe(time.Now().Add(-time.Hour * 72).UTC())
	if err != nil {
		return err
	}
	unsubscribeMigrations := []internal.CadRawMigration{}

	for _, cAndSM := range criteriaAndSampleMessages {
		criteria := *cAndSM.Criteria
		selectedFilter := internal.CadCriteria{}
		introMessage := "\nMessage filtered by\n"
		if criteria.Query != "" {
			selectedFilter.Query = criteria.Query
			fmt.Printf(introMessage)
			fmt.Printf("\tQuery: %s\n", selectedFilter.Query)
		} else if criteria.From != "" {
			selectedFilter.From = criteria.From
			fmt.Printf(introMessage)
			fmt.Printf("\tFrom: %s\n", selectedFilter.From)
		} else if criteria.To != "" {
			selectedFilter.To = criteria.To
			fmt.Printf(introMessage)
			fmt.Printf("\tTo: %s\n", selectedFilter.To)
		} else {
			continue
		}

		subject := ""
		for _, header := range (*cAndSM.SampleMessage).Payload.Headers {
			if header.Name == "Subject" && header.Value != "" {
				subject = header.Value
			}

		}
		fmt.Println("\tSample subject: ", subject)

		prompt := promptui.Select{
			Label: "Select action",
			Items: []string{
				create_filter,
				skip,
				// ignore,
			},
		}

		_, result, err := prompt.Run()

		if err != nil {
			return err
		}

		// fmt.Printf("You choose %q\n", result)

		if result == create_filter {
			selectedAction := internal.CadAction{}

			validateApplyLabel := func(input string) error {
				localLabels, err := internal.ReadLocalLabels()
				if err != nil {
					log.Printf("Unable to read local labels: %v", err)
					return err
				}

				found := false
				for _, label := range localLabels {
					if strings.ToLower(label.Name) == strings.ToLower(input) {
						found = true
						break
					}
				}

				if !found {
					return errors.New("Unable to find the label: " + input)
				}

				return nil
			}

			applyTheLabelPrompt := promptui.Prompt{
				Label:    "Apply the label name",
				Validate: validateApplyLabel,
			}

			applyTheLabelResult, err := applyTheLabelPrompt.Run()

			if err != nil {
				fmt.Printf("Prompt failed %v\n", err)
				return err
			}

			if applyTheLabelResult != "" {
				localLabels, _ := internal.ReadLocalLabels()

				labelId := ""
				for _, label := range localLabels {
					if strings.ToLower(label.Name) == strings.ToLower(applyTheLabelResult) {
						labelId = label.Id
						break
					}
				}

				selectedAction.AddLabelIds = append(selectedAction.AddLabelIds, labelId)
			}

			skipInboxPrompt := promptui.Select{
				Label: "Skip the Inbox (Archive it)",
				Items: []string{"Yes", "No"},
			}

			_, skipInboxResult, err := skipInboxPrompt.Run()

			if err != nil {
				fmt.Printf("Prompt failed %v\n", err)
				return err
			}

			if skipInboxResult == "Yes" {
				selectedAction.RemoveLabelIds = append(
					selectedAction.RemoveLabelIds,
					"INBOX",
				)
			}

			importancePrompt := promptui.Select{
				Label: "Should this be important",
				Items: []string{neverImportant, alwaysImportant, skip},
			}

			_, importanceResult, err := importancePrompt.Run()

			if err != nil {
				fmt.Printf("Prompt failed %v\n", err)
				return err
			}

			if importanceResult == alwaysImportant {
				selectedAction.AddLabelIds = append(
					selectedAction.AddLabelIds,
					"IMPORTANT",
				)
			} else if importanceResult == neverImportant {
				selectedAction.RemoveLabelIds = append(
					selectedAction.RemoveLabelIds,
					"IMPORTANT",
				)
			}

			// fmt.Printf("You choose %q\n", applyTheLabelResult)
			// fmt.Printf("You choose %q\n", skipInboxResult)
			// fmt.Printf("You choose %q\n", importanceResult)

			operation := internal.CreateFilterMigration
			note := fmt.Sprint("Unsubscribed by the doctor")
			labelMigration := internal.CadRawMigration{
				Operation: &operation,
				Details: internal.CadCreateFilterMigration{
					Criteria: &selectedFilter,
					Action:   &selectedAction,
				},
				Note: &note,
			}

			unsubscribeMigrations = append(unsubscribeMigrations, labelMigration)
		}
	}

	err = internal.CreateMigrationFile(&unsubscribeMigrations)
	if err != nil {
		fmt.Printf("Unable to create migration file %v\n", err)
		return err
	}

	return nil
}

func init() {
	rootCmd.AddCommand(doctorCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// doctorCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// doctorCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
