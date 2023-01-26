package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

type UpdateDnsBody struct {
	Data string `json:"data"` //ip
	Name string `json:"name"`
	TTL  int    `json:"ttl"`
	Type string `json:"type"`
}

func runGodaddy(config Config, ip string) error {
	url := fmt.Sprintf("https://api.godaddy.com/v1/domains/%s/records/%s/%s", config.Domain, config.RecordType, config.Name)
	updateDnsBody := []UpdateDnsBody{{
		Data: ip,
		Name: "home",
		TTL:  600,
		Type: "A",
	}}

	updateDnsBodyBytes, _ := json.Marshal(updateDnsBody)
	req, _ := http.NewRequest("PUT", url, bytes.NewReader(updateDnsBodyBytes))
	req.Header.Set("Authorization", fmt.Sprintf("sso-key %s:%s", config.Key, config.Secret))
	req.Header.Set("Content-Type", "application/json")
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode == 200 {
		return nil
	}

	respBytes, err1 := ioutil.ReadAll(resp.Body)
	if err1 != nil {
		return err1
	}
	return errors.New(string(respBytes))
}
