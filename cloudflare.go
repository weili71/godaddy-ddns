package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

var client = &http.Client{}

type DNSRecord struct {
	ID      string `json:"id,omitempty"`
	Type    string `json:"type"`
	Name    string `json:"name"`
	Content string `json:"content"`
	Proxied bool   `json:"proxied"`
	ZoneID  string `json:"zone_id"`
	TTL     int    `json:"ttl"`
}

func init() {
	//log.SetFlags()
	log.SetFlags(log.Lshortfile | log.Lmicroseconds | log.Ldate)
	//proxyUrl, _ := url.Parse("http://127.0.0.1:8888")
	//client = &http.Client{
	//	Transport: &http.Transport{
	//		Proxy: http.ProxyURL(proxyUrl),
	//	},
	//}
}
func getRecordID(zoneID, name, email, key, secret string) (string, error) {
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records?name=%s", zoneID, name)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("X-Auth-Email", email)
	req.Header.Set("X-Auth-Key", key)
	req.Header.Set("Authorization", "Bearer "+secret)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("Failed to retrieve record ID")
	}

	var data map[string]interface{}
	json.Unmarshal(body, &data)
	records := data["result"].([]interface{})
	if len(records) == 0 {
		return "", fmt.Errorf("Record not found")
	}
	return records[0].(map[string]interface{})["id"].(string), nil
}

func updateDNSRecord(record *DNSRecord, recordID, email, apiKey, secret string) (string, error) {
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records/%s", record.ZoneID, recordID)
	jsonRecord, _ := json.Marshal(record)
	req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(jsonRecord))
	req.Header.Set("X-Auth-Email", email)
	req.Header.Set("X-Auth-Key", apiKey)
	req.Header.Set("Authorization", "Bearer "+secret)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("Failed to create DNS record")
	}

	var data map[string]interface{}
	json.Unmarshal(body, &data)
	return string(body), nil
}

func createDNSRecord(record *DNSRecord, email, apiKey, secret string) (string, error) {
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records", record.ZoneID)
	jsonRecord, _ := json.Marshal(record)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonRecord))
	req.Header.Set("X-Auth-Email", email)
	req.Header.Set("X-Auth-Key", apiKey)
	req.Header.Set("Authorization", "Bearer "+secret)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("Failed to create DNS record")
	}

	var data map[string]interface{}
	json.Unmarshal(body, &data)
	recordID := data["result"].(map[string]interface{})["id"].(string)
	return recordID, nil
}

func getZoneID(domain, email, apiKey, secret string) (string, error) {
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones?name=%s", domain)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("X-Auth-Email", email)
	req.Header.Set("X-Auth-Key", apiKey)
	req.Header.Set("Authorization", "Bearer "+secret)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("Failed to retrieve zone ID")
	}

	var data map[string]interface{}
	json.Unmarshal(body, &data)
	zones := data["result"].([]interface{})
	if len(zones) == 0 {
		return "", fmt.Errorf("Zone not found")
	}
	return zones[0].(map[string]interface{})["id"].(string), nil
}

func runCloudflare(config Config, ip string) error {
	domain := config.Domain
	name := config.Name
	email := config.Email
	apiKey := config.Key
	secret := config.Secret
	recordType := config.RecordType
	z, ok := zoneIDCache[domain]
	var zoneID string
	if ok && len(z) > 0 {
		zoneID = z
	} else {
		zoneID1, err := getZoneID(domain, email, apiKey, secret)
		if err != nil {
			log.Println(err)
			return err
		}
		zoneID = zoneID1
	}
	zoneIDCache[domain] = zoneID

	r, ok1 := recordIDCache[domain+name]
	var recordID string
	if ok1 && len(r) > 0 {
		recordID = r
	} else {
		recordID1, err := getRecordID(zoneID, name+"."+domain, email, apiKey, secret)
		recordID = recordID1
		if err != nil {
			if err.Error() == "Record not found" {
				log.Println("创建新纪录")
				newRecord := &DNSRecord{
					Type:    recordType,
					Name:    name,
					Content: ip,
					Proxied: config.Proxy,
					ZoneID:  zoneID,
					TTL:     300,
				}
				recordID, err = createDNSRecord(newRecord, email, apiKey, secret)
				if err != nil {
					log.Println(err)
					return err
				}
				return nil
			} else {
				log.Println(err)
				return err
			}
		}
	}
	log.Println("更新纪录")
	recordIDCache[domain+name] = recordID

	newRecord := &DNSRecord{
		Type:    recordType,
		Name:    name,
		Content: ip,
		Proxied: config.Proxy,
		ZoneID:  zoneID,
		TTL:     300,
	}
	_, err := updateDNSRecord(newRecord, recordID, email, apiKey, secret)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

var (
	zoneIDCache   = make(map[string]string)
	recordIDCache = make(map[string]string)
)
