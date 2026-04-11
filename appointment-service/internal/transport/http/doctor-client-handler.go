package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type HTTP_doctor_client struct {
	baseURL    string
	httpClient *http.Client
}

func New_HTTP_doctor_client(baseURL string) *HTTP_doctor_client {
	return &HTTP_doctor_client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 3 * time.Second,
		},
	}
}

func (client *HTTP_doctor_client) Doctor_exists(doctorID string) (bool, error) {
	url := fmt.Sprintf("%s/doctors/%s", client.baseURL, doctorID)
	res, err := client.httpClient.Get(url)
	if err != nil {
		return false, fmt.Errorf("doctor service unreachable: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return false, nil
	}
	if res.StatusCode != http.StatusOK {
		return false, fmt.Errorf("unexpected status from doctor service: %d", res.StatusCode)
	}

	var body map[string]interface{}
	err = json.NewDecoder(res.Body).Decode(&body)
	if err != nil {
		return false, err
	}
	return true, nil
}
