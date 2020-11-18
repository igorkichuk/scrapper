// Package fatalfuncs contains functions which
// will call log.Fatal with error description inside if something is wrong.
package fatalfuncs

import (
	"encoding/json"
	"log"
	"os"
)

// SaveJsonToFile receives data, marshals it to json and saves to a file by filename path.
// If data can't be marshaled to json format, it will call log.Fatal.
// Warning! This function recreates file by the filename.
func SaveJsonToFile(fileName string, data interface{}) {
	f, err := os.Create(fileName)
	CheckErr(err)
	defer f.Close()

	jResult, err := json.Marshal(data)
	CheckErr(err)

	_, err = f.Write(jResult)
	CheckErr(err)
}

// CheckErr checks if err is not nil.
func CheckErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
