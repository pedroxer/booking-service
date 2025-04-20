package storage

import (
	"fmt"
	"strings"
)

type Field struct {
	Name  string
	Value interface{}
}
type SearchField struct {
	NameWhere string
	NameOrder string
}

func GenerateSearch(columns map[string]SearchField, filters []Field) (string, error) {
	var conditions strings.Builder
	andPrefix := ""
	for _, column := range filters {
		var ok bool
		var field SearchField
		if field, ok = columns[column.Name]; !ok {
			return "", fmt.Errorf("bad search by column %s", column.Name)
		}
		if column.Value == nil {
			conditions.WriteString(fmt.Sprintf("%s %s = IS NULL\n", andPrefix, field.NameWhere))
		} else {
			conditions.WriteString(fmt.Sprintf("%s %s = '%s'\n", andPrefix, field.NameWhere, fmt.Sprint(column.Value)))
		}
		andPrefix = "AND"
	}
	return conditions.String(), nil

}

func GenerateLimits(page, pageSize int64) string {
	return fmt.Sprintf(" OFFSET %d LIMIT %d", (page-1)*pageSize, pageSize)
}

func GenerateUpdates(columns map[string]SearchField, updateFields []Field) (string, error) {
	var updates strings.Builder
	andPrefix := ""
	for _, column := range updateFields {
		var ok bool
		var field SearchField
		if field, ok = columns[column.Name]; !ok {
			return "", fmt.Errorf("bad update by column %s", column.Name)
		}
		updates.WriteString(fmt.Sprintf("%s %s = '%s'\n", andPrefix, field.NameWhere, fmt.Sprint(column.Value)))
		andPrefix = ", "
	}
	return updates.String(), nil
}
