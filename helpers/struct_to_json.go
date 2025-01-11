package helpers

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func ConvertStructToJson(w http.ResponseWriter, status int, data interface{}) error {
	res, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal struct to json %w", err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(res)
	return nil

}
