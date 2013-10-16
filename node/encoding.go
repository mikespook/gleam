// Copyright 2012 Xing Xing <mikespook@gmail.com>.
// All rights reserved.
// Use of this source code is governed by a commercial
// license that can be found in the LICENSE file.

package node

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
)

type Encoding interface {
	Encode(fn *ZFunc) ([]byte, error)
	Decode(data []byte, fn *ZFunc) error
}

type JSON struct{}

func (j JSON) Decode(data []byte, fn *ZFunc) error {
	return json.Unmarshal(data, fn)
}

func (j JSON) Encode(fn *ZFunc) ([]byte, error) {
	return json.Marshal(fn)
}

type Gob struct{}

func (g Gob) Decode(data []byte, fn *ZFunc) error {
	d := gob.NewDecoder(bytes.NewReader(data))
	return d.Decode(fn)
}

func (g Gob) Encode(fn *ZFunc) ([]byte, error) {
	var b bytes.Buffer
	e := gob.NewEncoder(&b)
	err := e.Encode(fn)
	return b.Bytes(), err
}
