package commands

import "encoding/json"

func mergeRawMessages(items []json.RawMessage) json.RawMessage {
	if len(items) == 0 {
		return json.RawMessage("[]")
	}
	// Check if items are individual array elements or already arrays
	// If GetAll returned array items, wrap them
	result := []byte("[")
	for i, item := range items {
		if i > 0 {
			result = append(result, ',')
		}
		result = append(result, item...)
	}
	result = append(result, ']')
	return json.RawMessage(result)
}
