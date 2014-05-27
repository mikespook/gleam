package gleam

import "encoding/json"

type Func struct {
	Name string
	Data interface{}
}

func MarshalFunc(value string) (f *Func, err error) {
	if err = json.Unmarshal([]byte(value), &f); err != nil {
		return
	}
	return
}

func (f *Func) Unmarshal() (string, error) {
	data, err := json.Marshal(f)
	return string(data), err
}
