package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

type ResponseAddress struct {
	IP   string `json:"IP"`
	Port int    `json:"Port"`
}

func RunIpService() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		address := strings.Split(r.RemoteAddr, ":")
		port, _ := strconv.Atoi(address[1])
		resp, _ := json.Marshal(ResponseAddress{
			IP:   address[0],
			Port: port,
		})
		w.Write(resp)
	})
	http.ListenAndServe(":9999", nil)
}
