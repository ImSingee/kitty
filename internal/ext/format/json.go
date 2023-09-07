package format

import (
	"encoding/json"

	"github.com/ImSingee/go-ex/ee"
)

type JsonFormatter struct {
	Indent string
}

func (f *JsonFormatter) Format(filename string, content []byte) ([]byte, error) {
	var v any
	err := json.Unmarshal(content, &v)
	if err != nil {
		return nil, ee.New("invalid json")
	}

	return json.MarshalIndent(v, "", f.Indent)
}

func (o *options) jsonFormatter() *JsonFormatter {
	return &JsonFormatter{
		Indent: "  ",
	}
}

func (o *options) jsonFormat(filename string, content []byte) error {
	return o.doFormat(filename, content, o.jsonFormatter())
}
