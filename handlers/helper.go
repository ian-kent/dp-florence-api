package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

func unmarshal(req *http.Request, i interface{}) error {
	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return err
	}

	defer req.Body.Close()

	return json.Unmarshal(b, &i)
}
