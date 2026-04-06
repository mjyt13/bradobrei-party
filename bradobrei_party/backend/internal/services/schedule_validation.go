package services

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

var (
	validScheduleDays = map[string]struct{}{
		"mon": {},
		"tue": {},
		"wed": {},
		"thu": {},
		"fri": {},
		"sat": {},
		"sun": {},
	}
	scheduleRangePattern = regexp.MustCompile(`^(?:[01]\d|2[0-3]):[0-5]\d-(?:[01]\d|2[0-3]):[0-5]\d$`)
)

func normalizeScheduleJSON(raw string, fieldName string) (*string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, nil
	}

	var schedule map[string]string
	if err := json.Unmarshal([]byte(trimmed), &schedule); err != nil {
		return nil, fmt.Errorf("%s должен быть корректным JSON-объектом вида {\"mon\":\"10:00-19:00\"}", fieldName)
	}

	if len(schedule) == 0 {
		return nil, fmt.Errorf("%s не должен быть пустым JSON-объектом", fieldName)
	}

	normalized := make(map[string]string, len(schedule))
	for day, hours := range schedule {
		if _, ok := validScheduleDays[day]; !ok {
			return nil, fmt.Errorf("%s содержит неподдерживаемый день: %s", fieldName, day)
		}

		value := strings.TrimSpace(hours)
		if value == "" {
			return nil, fmt.Errorf("%s содержит пустой интервал для дня %s", fieldName, day)
		}
		if !scheduleRangePattern.MatchString(value) {
			return nil, fmt.Errorf("%s для дня %s должен быть в формате HH:MM-HH:MM", fieldName, day)
		}

		parts := strings.SplitN(value, "-", 2)
		if len(parts) != 2 || parts[0] >= parts[1] {
			return nil, fmt.Errorf("%s для дня %s должен содержать корректный интервал времени", fieldName, day)
		}

		normalized[day] = value
	}

	ordered := make(map[string]string, len(normalized))
	keys := make([]string, 0, len(normalized))
	for key := range normalized {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		ordered[key] = normalized[key]
	}

	bytes, err := json.Marshal(ordered)
	if err != nil {
		return nil, fmt.Errorf("не удалось нормализовать %s", fieldName)
	}

	value := string(bytes)
	return &value, nil
}
