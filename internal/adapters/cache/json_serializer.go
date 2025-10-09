package cache

import (
	"bytes"
	"encoding/json"
	"fmt"

	"zpwoot/internal/core/ports/output"
)

type JSONSerializer struct{}

func NewJSONSerializer() output.CacheSerializer {
	return &JSONSerializer{}
}

func (s *JSONSerializer) Serialize(data interface{}) ([]byte, error) {
	if data == nil {
		return nil, fmt.Errorf("cannot serialize nil data")
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize data to JSON: %w", err)
	}

	return bytes, nil
}

func (s *JSONSerializer) Deserialize(data []byte, target interface{}) error {
	if len(data) == 0 {
		return fmt.Errorf("cannot deserialize empty data")
	}

	if target == nil {
		return fmt.Errorf("target cannot be nil")
	}

	err := json.Unmarshal(data, target)
	if err != nil {
		return fmt.Errorf("failed to deserialize JSON data: %w", err)
	}

	return nil
}

func (s *JSONSerializer) SerializeString(str string) ([]byte, error) {
	return []byte(str), nil
}

func (s *JSONSerializer) DeserializeString(data []byte) (string, error) {
	return string(data), nil
}

func (s *JSONSerializer) SerializeMap(m map[string]interface{}) ([]byte, error) {
	return s.Serialize(m)
}

func (s *JSONSerializer) DeserializeMap(data []byte) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := s.Deserialize(data, &result)
	return result, err
}

func (s *JSONSerializer) SerializeSlice(slice []interface{}) ([]byte, error) {
	return s.Serialize(slice)
}

func (s *JSONSerializer) DeserializeSlice(data []byte) ([]interface{}, error) {
	var result []interface{}
	err := s.Deserialize(data, &result)
	return result, err
}

func (s *JSONSerializer) IsValidJSON(data []byte) bool {
	var js json.RawMessage
	return json.Unmarshal(data, &js) == nil
}

func (s *JSONSerializer) PrettyPrint(data interface{}) ([]byte, error) {
	return json.MarshalIndent(data, "", "  ")
}

func (s *JSONSerializer) Compact(data []byte) ([]byte, error) {
	var compact bytes.Buffer
	err := json.Compact(&compact, data)
	return compact.Bytes(), err
}
