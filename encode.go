package remote

import (
	"bytes"
	"encoding/gob"
	"io"
)

// GobEncode for encode using gob
func GobEncode(e interface{}) (string, error) {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(e)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// GobDecode for decode using gob
func GobDecode(r io.Reader, e interface{}) error {
	dec := gob.NewDecoder(r)
	err := dec.Decode(e)
	return err
}
