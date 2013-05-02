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

type ZDecodeHandler func([]byte, *ZFunc) error

func JSONDecoder(data []byte, fn *ZFunc) error {
    return json.Unmarshal(data, fn)
}

func JSONEncoder(fn *ZFunc) ([]byte, error) {
    return json.Marshal(fn)
}

func GobDecoder(data []byte, fn *ZFunc) error {
    d := gob.NewDecoder(bytes.NewReader(data))
    return d.Decode(fn)
}

func GobEncoder(fn *ZFunc) ([]byte, error) {
    var b bytes.Buffer
    e := gob.NewEncoder(&b)
    err := e.Encode(fn)
    return b.Bytes(), err
}
