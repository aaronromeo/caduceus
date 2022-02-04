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
const yes string = "Yes"
const no string = "No"
const neverImportant string = "Never mark it as important"
const alwaysImportant string = "Always mark it as important"

var FlagSuggestions bool
var FlagDirect bool
var FlagFilterMaintenance bool
var FlagFetch bool

func runDoctor(cmd *cobra.Command, args []string) {
	updateLabelsResult := yes
	updateFiltersResult := yes

	if !FlagDirect {
		var err error

		updateLabelsPrompt := promptui.Select{
			Label: "Update labels?",
			Items: []string{yes, no},
		}

		_, updateLabelsResult, err = updateLabelsPrompt.Run()

		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			panic(err)
		}

		updateFiltersPrompt := promptui.Select{
			Label: "Update filters?",
			Items: []string{yes, no},
		}

		_, updateFiltersResult, err = updateFiltersPrompt.Run()

		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			panic(err)
		}
	} else if FlagFetch {
		updateLabelsResult = yes
		updateFiltersResult = yes
	} else {
		updateLabelsResult = no
		updateFiltersResult = no
	}

	if updateLabelsResult == yes {
		FetchLabels()
	}

	if updateFiltersResult == yes {
		FetchFilters()
	}

	if FlagSuggestions || FlagFilterMaintenance {
		fmt.Println("Analyzing results...")
	}

	if FlagSuggestions && !FlagDirect {
		interactiveSuggestions()
	}

	if FlagFilterMaintenance {
		totalmigs := []internal.CadRawMigration{}

		filters, err := internal.SelectArchiveFilters()
		if err != nil {
			fmt.Printf("Maintenance failed %v\n", err)
			panic(err)
		}

		for _, archiveFilter := range filters {
			fmt.Printf("\tSearching for filter ID %s", archiveFilter.Id)
			ids, _ := internal.GetMessageIDsInInboxByFilterCriteria(&archiveFilter)

			if len(ids) != 0 {
				fmt.Print("\t\tFound results\n")

				operation := internal.UpdateMessagesMigration
				note := "Archived message identified by the doctor"
				archiveMigration := internal.CadRawMigration{
					Operation: &operation,
					Details: internal.CadUpdateMessagesMigration{
						MessageIds:     &ids,
						RemoveLabelIds: &archiveFilter.Action.RemoveLabelIds,
						AddLabelIds:    &archiveFilter.Action.AddLabelIds,
					},
					Note: &note,
				}

				totalmigs = append(totalmigs, archiveMigration)
			}
		}
		if len(totalmigs) > 0 {
			internal.CreateMigrationFile(&totalmigs)
		}
	}
}

func interactiveSuggestions() {
	emptyLabelCadMigrations, err := emptyLabelMigrations()
	if err != nil {
		panic(err)
	}

	unsubscribeCadMigrations, err := unsubscribeMigrations()
	if err != nil {
		panic(err)
	}

	totalmigs := []internal.CadRawMigration{}
	totalmigs = append(emptyLabelCadMigrations, unsubscribeCadMigrations...)
	internal.CreateMigrationFile(&totalmigs)
}

func emptyLabel(l internal.CadLabel) bool {
	// For some reason MessagesTotal and ThreadsTotal come back empty
	// in spite of there being messages
	return l.MessagesTotal == 0 &&
		l.MessagesUnread == 0 &&
		l.ThreadsTotal == 0 &&
		l.ThreadsUnread == 0
}

func emptyLabelMigrations() ([]internal.CadRawMigration, error) {
	localLabels, err := internal.ReadLocalLabels()
	if err != nil {
		log.Printf("Unable to read local labels: %v", err)
		return nil, err
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

			prompt := promptui.Select{
				Label: fmt.Sprintf("Delete label %s", label.Name),
				Items: []string{
					yes,
					no,
				},
			}

			_, result, err := prompt.Run()

			if err != nil {
				return nil, err
			}

			if result == yes {
				emptyLabelMigrations = append(emptyLabelMigrations, labelMigration)
			}
		} else {
			fmt.Printf("Skipping parent label %s\n", label.Name)
		}
	}
	return emptyLabelMigrations, nil
}

func unsubscribeMigrations() ([]internal.CadRawMigration, error) {
	returnMigrations := []internal.CadRawMigration{}
	criteriaAndSampleMessages, err := internal.GetMessageCriteriaForUnsubscribe(time.Now().Add(-time.Hour * 72).UTC())
	if err != nil {
		return nil, err
	}
	unsubscribeMigrations := []internal.CadRawMigration{}

	for _, cAndSM := range criteriaAndSampleMessages {
		criteria := *cAndSM.Criteria
		selectedFilter := internal.CadCriteria{}
		introMessage := "\nMessage filtered by\n"
		if criteria.Query != "" {
			selectedFilter.Query = criteria.Query
			fmt.Print(introMessage)
			fmt.Printf("\tQuery: %s\n", selectedFilter.Query)
		} else if criteria.From != "" {
			selectedFilter.From = criteria.From
			fmt.Print(introMessage)
			fmt.Printf("\tFrom: %s\n", selectedFilter.From)
		} else if criteria.To != "" {
			selectedFilter.To = criteria.To
			fmt.Print(introMessage)
			fmt.Printf("\tTo: %s\n", selectedFilter.To)
		} else {
			continue
		}

		from := ""
		to := ""
		subject := ""
		for _, header := range (*cAndSM.SampleMessage).Payload.Headers {
			if header.Name == "Subject" && header.Value != "" {
				subject = header.Value
			} else if header.Name == "From" && header.Value != "" {
				from = header.Value
			} else if header.Name == "To" && header.Value != "" {
				to = header.Value
			}

		}
		fmt.Printf("\tFrom: %s\n\tTo: %s\n\tSample subject: %s\n", from, to, subject)

		prompt := promptui.Select{
			Label: "Select action",
			Items: []string{
				create_filter,
				skip,
			},
		}

		_, result, err := prompt.Run()

		if err != nil {
			return returnMigrations, err
		}

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
					if strings.EqualFold(label.Name, input) {
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
				return returnMigrations, err
			}

			if applyTheLabelResult != "" {
				localLabels, _ := internal.ReadLocalLabels()

				labelId := ""
				for _, label := range localLabels {
					if strings.EqualFold(label.Name, applyTheLabelResult) {
						labelId = label.Id
						break
					}
				}

				selectedAction.AddLabelIds = append(selectedAction.AddLabelIds, labelId)
			}

			skipInboxPrompt := promptui.Select{
				Label: "Skip the Inbox (Archive it)",
				Items: []string{yes, no},
			}

			_, skipInboxResult, err := skipInboxPrompt.Run()

			if err != nil {
				fmt.Printf("Prompt failed %v\n", err)
				return returnMigrations, err
			}

			if skipInboxResult == yes {
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
				return returnMigrations, err
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

			operation := internal.CreateFilterMigration
			note := "Unsubscribed by the doctor"
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

	return unsubscribeMigrations, nil
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
	doctorCmd.Flags().BoolVarP(&FlagDirect, "direct", "d", false, "The opposite of interactive mode")
	doctorCmd.Flags().BoolVarP(&FlagFetch, "fetch", "f", true, "Fetch the labels and filters (only used in direct mode)")
	doctorCmd.Flags().BoolVarP(&FlagSuggestions, "suggestions", "s", true, "Generate filter and label suggestions if in interactive mode")
	doctorCmd.Flags().BoolVarP(&FlagFilterMaintenance, "maintenance", "m", false, "Generate message cleanup based on existing filters")
}
