package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const apiBase = "https://api.simonegentili.com/quadrato"

func doRequest(method, path string, body interface{}, token string) ([]byte, error) {
	var bodyReader io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, apiBase+path, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "curl/8.7.1")
	req.Header.Set("Origin", "https://quadrato.simonegentili.com")
	req.Header.Set("Referer", "https://quadrato.simonegentili.com/")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		var apiErr APIError
		if json.Unmarshal(respBody, &apiErr) == nil && apiErr.Message != "" {
			return nil, fmt.Errorf("%s", apiErr.Message)
		}
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	return respBody, nil
}

func fetchData(token string) (Data, error) {
	body, err := doRequest("GET", "/data", nil, token)
	if err != nil {
		return Data{}, err
	}
	var d Data
	err = json.Unmarshal(body, &d)
	return d, err
}
