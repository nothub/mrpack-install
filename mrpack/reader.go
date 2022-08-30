package mrpack

import (
	"encoding/json"
	"os"
)

func ReadFile(file *os.File) (*Index, error) {
	jsonParser := json.NewDecoder(file)
	index := Index{}
	err := jsonParser.Decode(&index)
	if err != nil {
		return nil, err
	}
	return &index, nil
}
