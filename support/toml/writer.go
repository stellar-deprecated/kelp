package toml

import (
	"bytes"
	"fmt"
	"io/ioutil"

	"github.com/BurntSushi/toml"
)

// WriteFile is a helper method to write toml files
func WriteFile(filePath string, v interface{}) error {
	var fileBuf bytes.Buffer
	encoder := toml.NewEncoder(&fileBuf)

	e := encoder.Encode(v)
	if e != nil {
		return fmt.Errorf("error encoding file as toml: %s", e)
	}

	ioutil.WriteFile(filePath, fileBuf.Bytes(), 0644)
	return nil
}
