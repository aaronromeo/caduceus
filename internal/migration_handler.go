package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"
)

const UpdateMessagesMigration string = "update-messages"
const DeleteLabelMigration string = "delete-label"
const UpdateLabelMigration string = "update-label"
const UpdateLabelsMigration string = "update-labels"
const CreateLabelMigration string = "create-label"
const ReplaceFiltersMigration string = "replace-filters"
const DeleteFilterMigration string = "delete-filter"
const DeleteFiltersMigration string = "delete-filters"
const CreateFilterMigration string = "create-filter"

type CadUpdateMessagesMigration struct {
	QueryLabelIds  *[]string `json:"queryLabelIds"`
	MessageIds     *[]string `json:"messageIds"`
	QueryString    *string   `json:"query"`
	RemoveLabelIds *[]string `json:"removeLabelIds"`
	AddLabelIds    *[]string `json:"addLabelIds"`
}

type CadCreateFilterMigration struct {
	Criteria *CadCriteria `json:"criteria,omitempty"`
	Action   *CadAction   `json:"action,omitempty"`
}

type CadReplaceFiltersMigration struct {
	Ids    *[]string  `json:"ids"`
	Action *CadAction `json:"action,omitempty"`
}

type CadDeleteFilterMigration struct {
	Id *string `json:"id"`
}

type CadDeleteFiltersMigration struct {
	Ids *[]string `json:"ids"`
}

type CadUpdateLabelMigration struct {
	Id                    *string        `json:"id"`
	Name                  *string        `json:"name,omitempty"`
	LabelListVisibility   *string        `json:"labelListVisibility,omitempty"`
	MessageListVisibility *string        `json:"messageListVisibility,omitempty"`
	Color                 *CadLabelColor `json:"color,omitempty"`
}

type CadUpdateLabelsMigration struct {
	Ids                   *string        `json:"ids"`
	LabelListVisibility   *string        `json:"labelListVisibility,omitempty"`
	MessageListVisibility *string        `json:"messageListVisibility,omitempty"`
	Color                 *CadLabelColor `json:"color,omitempty"`
}

type CadCreateLabelMigration struct {
	Name                  *string        `json:"name,omitempty"`
	LabelListVisibility   *string        `json:"labelListVisibility,omitempty"`
	MessageListVisibility *string        `json:"messageListVisibility,omitempty"`
	Color                 *CadLabelColor `json:"color,omitempty"`
}

type CadDeleteLabelMigration struct {
	Id *string `json:"id"`
}

type CadRawMigration struct {
	Operation  *string         `json:"operation"`
	Details    interface{}     `json:"-"`
	RawDetails json.RawMessage `json:"details"`
	Note       *string         `json:"note,omitempty"`
}

const migrationsPath string = "migrations"

var indent string = ""

func RunMigrations(daily bool) error {
	migrationFiles, err := getMigrationFiles(daily)
	if err != nil {
		log.Printf("Unable to fetch migration files: %v", err)
		return err
	}

	for _, migrationFile := range migrationFiles {
		var migrations []CadRawMigration

		fmt.Printf("Processing migration %s\n", migrationFile)
		b, err := ioutil.ReadFile(migrationFile)
		if err != nil {
			log.Printf("Unable to read the migration file: %v", err)
			return err
		}
		if err := json.Unmarshal(b, &migrations); err != nil {
			return err
		}

		for _, migration := range migrations {
			switch *migration.Operation {
			case UpdateMessagesMigration:
				messageMigration := CadUpdateMessagesMigration{}
				b, _ := migration.RawDetails.MarshalJSON()
				json.Unmarshal(b, &messageMigration)
				err := updateMessages(messageMigration)
				if err != nil {
					return err
				}
			case CreateFilterMigration:
				filterMigration := CadCreateFilterMigration{}
				b, _ := migration.RawDetails.MarshalJSON()
				json.Unmarshal(b, &filterMigration)
				err := createFilter(filterMigration)
				if err != nil {
					return err
				}
			case DeleteFilterMigration:
				filterMigration := CadDeleteFilterMigration{}
				b, _ := migration.RawDetails.MarshalJSON()
				json.Unmarshal(b, &filterMigration)
				err := deleteFilter(filterMigration)
				if err != nil {
					return err
				}
			case DeleteFiltersMigration:
				filtersMigration := CadDeleteFiltersMigration{}
				b, _ := migration.RawDetails.MarshalJSON()
				json.Unmarshal(b, &filtersMigration)
				err := deleteFilters(filtersMigration)
				if err != nil {
					return err
				}
			case UpdateLabelMigration:
				labelMigration := CadUpdateLabelMigration{}
				b, _ := migration.RawDetails.MarshalJSON()
				json.Unmarshal(b, &labelMigration)
				err := updateLabel(labelMigration)
				if err != nil {
					return err
				}
			case UpdateLabelsMigration:
				// labelMigration := CadUpdateLabelMigration{}
				// b, _ := migration.RawDetails.MarshalJSON()
				// json.Unmarshal(b, &labelMigration)
				// err := updateLabel(labelMigration)
				// if err != nil {
				// 	return err
				// }
				return errors.New("not implemented operation " + *migration.Operation)
			case CreateLabelMigration:
				labelMigration := CadCreateLabelMigration{}
				b, _ := migration.RawDetails.MarshalJSON()
				json.Unmarshal(b, &labelMigration)
				err := createLabel(labelMigration)
				if err != nil {
					return err
				}
			case DeleteLabelMigration:
				labelMigration := CadDeleteLabelMigration{}
				b, _ := migration.RawDetails.MarshalJSON()
				json.Unmarshal(b, &labelMigration)
				err := deleteLabel(labelMigration)
				if err != nil {
					return err
				}
			case ReplaceFiltersMigration:
				filterMigration := CadReplaceFiltersMigration{}
				b, _ := migration.RawDetails.MarshalJSON()
				json.Unmarshal(b, &filterMigration)
				err := replaceFilters(filterMigration)
				if err != nil {
					return err
				}
			default:
				return errors.New("unknown operation " + *migration.Operation)
			}
		}

		if !daily {
			err = os.Rename(
				migrationFile,
				strings.ReplaceAll(
					strings.ToLower(migrationFile), ".json", "-complete.json",
				),
			)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (m *CadRawMigration) UnmarshalJSON(b []byte) error {
	type cadRawMigration CadRawMigration

	err := json.Unmarshal(b, (*cadRawMigration)(m))
	if err != nil {
		log.Printf("Unable to marshal data: %v", err)
		return err
	}

	return nil
}

func (m *CadRawMigration) MarshalJSON() ([]byte, error) {
	type cadRawMigration CadRawMigration

	if m.Details != nil {
		b, err := json.Marshal(m.Details)

		if err != nil {
			return nil, err
		}
		m.RawDetails = b
	}
	return json.Marshal((*cadRawMigration)(m))
}

func getMigrationFiles(daily bool) ([]string, error) {
	files, err := os.ReadDir(migrationsPath)
	if err != nil {
		log.Printf("Unable to read the migrations directory: %v", err)
		return nil, err
	}

	migrationFiles := []string{}
	for _, file := range files {
		r, _ := regexp.Compile("^[0-9]+.json$")
		if daily {
			r, _ = regexp.Compile("^daily-[0-9]+.json$")
		}
		if r.MatchString(file.Name()) {
			migrationFiles = append(migrationFiles, strings.Join([]string{migrationsPath, file.Name()}, "/"))
		}
	}
	if len(migrationFiles) == 0 {
		log.Printf("No files to migrate")
		return nil, errors.New("no migrations files")
	}
	sort.SliceStable(migrationFiles, func(i, j int) bool {
		return migrationFiles[i] < migrationFiles[j]
	})

	return migrationFiles, nil
}

func CreateMigrationFile(migrations *[]CadRawMigration) error {
	data, err := json.MarshalIndent(migrations, "", " ")
	if err != nil {
		log.Printf("Unable to marshal migration JSON: %v", err)
		return err
	}

	t := time.Now()
	err = ioutil.WriteFile(fmt.Sprintf("migrations/%s.json", t.Format("20060201-0304")), data, 0644)
	if err != nil {
		log.Printf("Unable write migrations: %v", err)
		return err
	}

	return nil
}

func updateMessages(migration CadUpdateMessagesMigration) error {
	printMessages := []string{}
	if migration.MessageIds != nil {
		printMessages = append(printMessages, *migration.MessageIds...)
	}
	if migration.QueryLabelIds != nil {
		printMessages = append(printMessages, *migration.QueryLabelIds...)
	}
	if migration.QueryString != nil {
		printMessages = append(printMessages, *migration.QueryString)
	}
	fmt.Println("Migrating messages...", printMessages)

	labels := []*CadLabel{}
	if migration.QueryLabelIds != nil {
		for _, labelId := range *migration.QueryLabelIds {
			labels = append(labels, &CadLabel{Id: labelId})
		}
	}
	messageIds := []string{}
	var err error
	if len(labels) > 0 {
		messageIds, err = GetMessagesIDsByLabelIDs(labels, migration.QueryString)
		if err != nil {
			log.Printf("Unable to retrieve message Ids: %v", err)
			return err
		}

	}
	messageIds = append(messageIds, *migration.MessageIds...)

	err = BulkUpdateMessageLabels(
		messageIds,
		*migration.AddLabelIds,
		*migration.RemoveLabelIds,
	)
	if err != nil {
		log.Printf("Unable to modify messages: %v", err)
		return err
	}

	return nil
}

func createFilter(migration CadCreateFilterMigration) error {
	fmt.Printf("%sCreating filter...%s %s %s\n", indent, migration.Criteria.From, migration.Criteria.To, migration.Criteria.Subject)

	labels, err := GetLabels()
	if err != nil {
		log.Printf("Unable to retrieve labels\n")
		return err
	}

	labelIdToType := map[string]string{}
	for _, label := range labels {
		labelIdToType[label.Id] = label.Type
	}

	sort.SliceStable(migration.Action.AddLabelIds, func(i, j int) bool {
		return labelIdToType[migration.Action.AddLabelIds[i]] < labelIdToType[migration.Action.AddLabelIds[j]]
	})

	for _, labelId := range migration.Action.RemoveLabelIds {
		if labelIdToType[labelId] == "user" {
			log.Printf("Unable to create filter with user label removeLabelId %s \n", labelId)
			return fmt.Errorf("unable to create filter with user label removeLabelId %s", labelId)
		}
	}

	newFilters := []*CadFilter{}
	action := &CadAction{RemoveLabelIds: migration.Action.RemoveLabelIds}
	currentNewCadFilter := &CadFilter{Action: action, Criteria: migration.Criteria}
	userLabelCount := 0
	for _, labelId := range migration.Action.AddLabelIds {
		if labelIdToType[labelId] == "user" && userLabelCount >= 1 {
			newFilters = append(newFilters, currentNewCadFilter)
			action = &CadAction{AddLabelIds: []string{labelId}}
			userLabelCount = 1
		} else {
			action.AddLabelIds = append(currentNewCadFilter.Action.AddLabelIds, labelId)
			if labelIdToType[labelId] == "user" {
				userLabelCount += 1
			}
		}
		currentNewCadFilter = &CadFilter{Action: action, Criteria: migration.Criteria}
	}
	newFilters = append(newFilters, currentNewCadFilter)

	indent = fmt.Sprintf("%s\t", indent)
	for _, filter := range newFilters {
		fmt.Printf("%sCreating subfilter...\n", indent)
		_, err := CreateFilter(filter)
		if err != nil {
			log.Printf("Unable to create new filter")
			return err
		}
	}
	indent = indent[:len(indent)-1]

	return nil
}

func deleteFilter(migration CadDeleteFilterMigration) error {
	fmt.Printf("%sDeleting filter... %s\n", indent, *migration.Id)

	oldCadFilter := &CadFilter{Id: *migration.Id}
	err := DeleteFilter(oldCadFilter)
	if err != nil {
		log.Printf("Unable to delete filter %v", *migration.Id)
		return err
	}

	return nil
}

func deleteFilters(bulkMigration CadDeleteFiltersMigration) error {
	fmt.Println("Deleting multiple filters...")

	indent = fmt.Sprintf("%s\t", indent)
	for _, id := range *bulkMigration.Ids {
		migration := &CadDeleteFilterMigration{Id: &id}
		err := deleteFilter(*migration)
		if err != nil {
			log.Printf("Unable to delete filter %v", *migration.Id)
			return err
		}
	}
	indent = indent[:len(indent)-1]

	return nil
}

func createLabel(migration CadCreateLabelMigration) error {
	fmt.Println("Creating label...", *migration.Name)

	newCadLabel := &CadLabel{}
	if migration.Name != nil {
		newCadLabel.Name = *migration.Name
	}
	if migration.LabelListVisibility != nil {
		newCadLabel.LabelListVisibility = *migration.LabelListVisibility
	}
	if migration.MessageListVisibility != nil {
		newCadLabel.MessageListVisibility = *migration.MessageListVisibility
	}
	if migration.Color != nil {
		newCadLabel.Color = *migration.Color
	}

	_, err := CreateUserLabel(newCadLabel)
	if err != nil {
		log.Printf("Unable to create new filter")
		return err
	}

	return nil
}

func deleteLabel(migration CadDeleteLabelMigration) error {
	fmt.Println("Deleting label...", *migration.Id)

	if migration.Id == nil {
		log.Printf("Label Id cannot be nil")
		return errors.New("delete Label called with missing label id")
	}

	oldCadLabel := &CadLabel{Id: *migration.Id}
	err := DeleteUserLabel(oldCadLabel)
	if err != nil {
		log.Printf("Unable to delete label %v", *migration.Id)
		return err
	}

	return nil
}

func updateLabel(migration CadUpdateLabelMigration) error {
	fmt.Println("Update label...", *migration.Id)

	if migration.Id == nil {
		log.Printf("Label Id cannot be nil")
		return errors.New("update Label called with missing label id")
	}

	updatedCadLabel := &CadLabel{Id: *migration.Id}
	if migration.Name != nil {
		updatedCadLabel.Name = *migration.Name
	}
	if migration.LabelListVisibility != nil {
		updatedCadLabel.LabelListVisibility = *migration.LabelListVisibility
	}
	if migration.MessageListVisibility != nil {
		updatedCadLabel.MessageListVisibility = *migration.MessageListVisibility
	}
	if migration.Color != nil {
		updatedCadLabel.Color = *migration.Color
	}

	_, err := PatchUserLabel(updatedCadLabel.Id, updatedCadLabel)
	if err != nil {
		log.Printf("Unable to update new filter")
		return err
	}

	return nil
}

func replaceFilters(migration CadReplaceFiltersMigration) error {
	fmt.Println("Replacing filters...")

	if migration.Ids == nil || len(*migration.Ids) == 0 {
		log.Printf("Label Id cannot be nil")
		return errors.New("replace Filters called without ids")
	}

	filterIdCriteriaMap := map[string]CadCriteria{}
	filtersToDelete := []*CadFilter{}

	// Verify the filters with the associated IDs exist and get criteria
	for _, id := range *migration.Ids {
		cadFilter := &CadFilter{Id: id}
		cadFilter, err := GetFilter(cadFilter)
		if err != nil {
			log.Printf("Unable to retrieve filter %s", id)
			return err
		}
		filterIdCriteriaMap[id] = *cadFilter.Criteria
		filtersToDelete = append(filtersToDelete, cadFilter)
	}

	indent = fmt.Sprintf("%s\t", indent)
	for _, cadfilter := range filtersToDelete {
		criteria := filterIdCriteriaMap[cadfilter.Id]

		fmt.Printf("%sDeleting filter... %s\n", indent, cadfilter.Id)
		deleteFilterMigration := CadDeleteFilterMigration{
			Id: &cadfilter.Id,
		}
		err := deleteFilter(deleteFilterMigration)
		if err != nil {
			log.Printf("Unable to delete filter %s", cadfilter.Id)
			return err
		}

		fmt.Printf("%sCreate filter... %s %s %s\n", indent, criteria.From, criteria.To, criteria.Subject)
		createFilterMigration := CadCreateFilterMigration{
			Criteria: &criteria,
			Action:   migration.Action,
		}
		err = createFilter(createFilterMigration)
		if err != nil {
			log.Printf("Unable to create filter")
			return err
		}
	}
	indent = indent[:len(indent)-1]

	return nil
}
