package jsonfmt

import (
	"bytes"
	"encoding/json"
)

func Marshal(v any, indent string) ([]byte, error) {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", indent)

	err := encoder.Encode(v)
	if err != nil {
		return nil, err
	}

	return bytes.TrimSuffix(buf.Bytes(), []byte("\n")), nil
}
