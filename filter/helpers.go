package filter

import (
	"fmt"
	"strings"
	"time"
)

func parseNumber(value any) (float64, error) {
	var num float64

	switch v := value.(type) {
	case int:
		num = float64(v)
	case int32:
		num = float64(v)
	case int64:
		num = float64(v)
	case float32:
		num = float64(v)
	case float64:
		num = v
	default:
		return 0, fmt.Errorf("invalid number type for field %s", value)
	}
	return num, nil
}

func parseText(value any) (string, error) {
	str, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("invalid text type for field %s", value)
	}
	return strings.ToLower(strings.TrimSpace(str)), nil
}

func parseDate(value any) (time.Time, error) {
	t, ok := value.(time.Time)
	if !ok {
		return time.Time{}, fmt.Errorf("invalid date type for field %s", value)
	}
	dateOnly := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	return dateOnly, nil
}

func parseTime(value any) (time.Time, error) {
	t, ok := value.(time.Time)
	if !ok {
		return time.Time{}, fmt.Errorf("invalid time type for field %s", value)
	}
	timeOnly := time.Date(0, 1, 1, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), time.UTC)
	return timeOnly, nil
}

func parseDateTime(value any) (time.Time, error) {
	t, ok := value.(time.Time)
	if !ok {
		return time.Time{}, fmt.Errorf("invalid datetime type for field %s", value)
	}
	return t, nil
}

func parseRangeNumber(value any) (FilterRangeNumber, error) {
	rng, ok := value.(FilterRange)
	if !ok {
		return FilterRangeNumber{}, fmt.Errorf("invalid range type for field %s", value)
	}
	from, err := parseNumber(rng.From)
	if err != nil {
		return FilterRangeNumber{}, err
	}
	to, err := parseNumber(rng.To)
	if err != nil {
		return FilterRangeNumber{}, err
	}
	return FilterRangeNumber{
		From: from,
		To:   to,
	}, nil
}

func parseRangeDate(value any) (FilterRangeDate, error) {
	rng, ok := value.(FilterRange)
	if !ok {
		return FilterRangeDate{}, fmt.Errorf("invalid range type for field %s", value)
	}
	from, err := parseDate(rng.From)
	if err != nil {
		return FilterRangeDate{}, err
	}
	to, err := parseDate(rng.To)
	if err != nil {
		return FilterRangeDate{}, err
	}

	// Validate that from <= to
	if from.After(to) {
		return FilterRangeDate{}, fmt.Errorf("range from date cannot be after to date")
	}

	return FilterRangeDate{
		From: from,
		To:   to,
	}, nil
}

func parseRangeTime(value any) (FilterRangeTime, error) {
	rng, ok := value.(FilterRange)
	if !ok {
		return FilterRangeTime{}, fmt.Errorf("invalid range type for field %s", value)
	}
	from, err := parseTime(rng.From)
	if err != nil {
		return FilterRangeTime{}, err
	}
	to, err := parseTime(rng.To)
	if err != nil {
		return FilterRangeTime{}, err
	}

	// Validate that from <= to
	if from.After(to) {
		return FilterRangeTime{}, fmt.Errorf("range from time cannot be after to time")
	}

	return FilterRangeTime{
		From: from,
		To:   to,
	}, nil
}

func parseBool(value any) (bool, error) {
	b, ok := value.(bool)
	if !ok {
		return false, fmt.Errorf("invalid boolean type for field %s", value)
	}
	return b, nil
}
