package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"time"
)

type MyInfo struct {
	domain     string
	key        string
	secret     string
	recordType string
	name       string
	ipserver   string
}

type Address struct {
	IP   string `json:"IP"`
	Port int    `json:"Port"`
}

type UpdateDnsBody struct {
	Data string `json:"data"` //ip
	Name string `json:"name"`
	TTL  int    `json:"ttl"`
	Type string `json:"type"`
}

const (
	ipserver = "http://120.78.173.214:9999/"
)

func main() {
	run(MyInfo{
		domain:     "xxx.com",
		key:        "",
		secret:     "",
		recordType: "A",
		name:       "www",
		ipserver: ipserver,
	})
}

func run(info MyInfo) {
	for {
		oldIP, err := net.ResolveIPAddr("ip", info.name+"."+info.domain)
		newIP, err1 := getNewAddress(info)
		if err1 != nil {
			time.Sleep(time.Minute)
			continue
		}
		if err == nil || newIP == oldIP.String() {
			log.Println("ip不变")
		} else {
			fmt.Println("ip变化，正在更新dns记录")
			err := updateDnsRecord(info, newIP)
			if err != nil {
				fmt.Printf("dns更新成功，旧地址：%s,新地址：%s\n", oldIP.String(), newIP)
			} else {
				fmt.Println(err)
			}
		}
		time.Sleep(time.Minute)
	}
}

func getNewAddress(info MyInfo) (string, error) {
	resp, err := http.Get(info.ipserver)
	if err != nil {
		return "", err
	}
	respBytes, err1 := ioutil.ReadAll(resp.Body)
	if err1 != nil {
		return "", err1
	}
	newAddress := Address{}
	err2 := json.Unmarshal(respBytes, &newAddress)
	if err2 != nil {
		return "", err2
	}
	return newAddress.IP, nil
}

func updateDnsRecord(info MyInfo, ip string) error {
	url := fmt.Sprintf("https://api.godaddy.com/v1/domains/%s/records/%s/%s", info.domain, info.recordType, info.name)
	updateDnsBody := []UpdateDnsBody{{
		Data: ip,
		Name: "home",
		TTL:  600,
		Type: "A",
	}}

	updateDnsBodyBytes, _ := json.Marshal(updateDnsBody)
	req, _ := http.NewRequest("PUT", url, bytes.NewReader(updateDnsBodyBytes))
	req.Header.Set("Authorization", fmt.Sprintf("sso-key %s:%s", info.key, info.secret))
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
