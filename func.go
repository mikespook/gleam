package gleam

import "encoding/json"

type Func struct {
	Name string
	Data interface{}
}

func marshal(value string) (f *Func, err error) {
	if err = json.Unmarshal([]byte(value), &f); err != nil {
		return
	}
	return
}

func unmarshal(f *Func) (string, error) {
	data, err := json.Marshal(f)
	return string(data), err
}
