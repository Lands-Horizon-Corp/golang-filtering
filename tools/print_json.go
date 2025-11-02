package tools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
)

// PrintAsJSON prints any value as formatted JSON
func PrintAsJSON(v any) {
	data, err := json.Marshal(v)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling to JSON: %v\n", err)
		return
	}

	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, data, "", "  "); err != nil {
		fmt.Fprintf(os.Stderr, "Error formatting JSON: %v\n", err)
		return
	}

	fmt.Println(prettyJSON.String())
}

// PrintAsJSONCompact prints any value as compact JSON
func PrintAsJSONCompact(v any) {
	data, err := json.Marshal(v)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling to JSON: %v\n", err)
		return
	}
	fmt.Println(string(data))
}

// ToJSONString returns a formatted JSON string
func ToJSONString(v any) (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", err
	}

	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, data, "", "  "); err != nil {
		return "", err
	}

	return prettyJSON.String(), nil
}
