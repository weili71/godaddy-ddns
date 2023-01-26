package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/wilinz/go-filex"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

type Config struct {
	Server     string `json:"server"`
	Domain     string `json:"domain"`
	Email      string `json:"email"`
	Key        string `json:"key"`
	Secret     string `json:"secret"`
	RecordType string `json:"record_type"`
	Name       string `json:"name"`
	IpServer   string `json:"ip_server"`
	Proxy      bool   `json:"proxy"`
}

const (
	cloudflare = "cloudflare"
	godaddy    = "godaddy"
)

func main() {
	go func() {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			w.Write([]byte("DDNS服务已运行"))
		})
		err := http.ListenAndServe(":20231", nil)
		if err != nil {
			log.Panic(err)
			return
		}
	}()
	programDir := filex.NewFile(os.Args[0]).Parent()
	configFile := filex.NewFile1(programDir, "config.json")

	fmt.Println("File path: " + configFile.Pathname)

	if !configFile.IsExist() {
		configTemplate, _ := json.Marshal(Config{
			Server:     "godaddy or cloudflare",
			Domain:     "xxx.com",
			Email:      "xxx@xxx,com",
			Key:        "key",
			Secret:     "secret",
			RecordType: "A",
			Name:       "www",
			IpServer:   "http://192.168.1.1:20230/wanip or if://pppoe-wan",
			Proxy:      false,
		})

		var out bytes.Buffer
		err := json.Indent(&out, configTemplate, "", "    ")
		if err != nil {
			log.Panicln(err)
			return
		}

		err = configFile.Write(out.Bytes(), 0777)
		if err != nil {
			log.Panicln(err)
			return
		}
		fmt.Println("Please edit config.json first!")
		return
	}

	configBytes, err := configFile.ReadAll()
	if err != nil {
		log.Panicln(err)
		return
	}

	var config Config
	err = json.Unmarshal(configBytes, &config)
	if err != nil {
		log.Panicln(err)
		return
	}
	fmt.Printf("%#v\n", config)
	run(config)
}

func run(config Config) {
	for {
		oldIP := "0.0.0.0"
		oldIPAddr, err := net.ResolveIPAddr("ip", config.Name+"."+config.Domain)
		if err == nil {
			oldIP = oldIPAddr.String()
		}
		newIP, err1 := getNewAddress(config)
		if err1 != nil {
			time.Sleep(time.Minute)
			continue
		}
		if newIP == oldIP {
			log.Println("ip不变")
		} else {
			fmt.Printf("ip变化，正在更新dns记录，旧地址：%s,新地址：%s\n", oldIP, newIP)
			if config.Server == cloudflare {
				err = runCloudflare(config, newIP)
			} else if config.Server == godaddy {
				err = runGodaddy(config, newIP)
			} else {
				log.Panic("配置服务商错误")
			}
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Printf("dns更新成功，旧地址：%s,新地址：%s\n", oldIP, newIP)
			}
		}
		time.Sleep(time.Minute)
	}
}

func getNewAddress(config Config) (string, error) {
	if isInterface, _ := regexp.MatchString("^if://.*$", config.IpServer); isInterface {
		ips, err := Ips()
		if err != nil {
			return "", err
		}
		name := strings.ReplaceAll(config.IpServer, "if://", "")
		ip := ips[name]
		return ip, nil
	}
	resp, err := http.Get(config.IpServer)
	if err != nil {
		return "", err
	}
	respBytes, err1 := ioutil.ReadAll(resp.Body)
	if err1 != nil {
		return "", err1
	}
	newAddress := string(respBytes)
	return newAddress, nil
}

func Ips() (map[string]string, error) {

	ips := make(map[string]string)

	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, i := range interfaces {
		byName, err := net.InterfaceByName(i.Name)
		if err != nil {
			return nil, err
		}
		addresses, err := byName.Addrs()
		for _, address := range addresses {
			ipNet, isVailIpNet := address.(*net.IPNet)
			// 检查ip地址判断是否回环地址
			if isVailIpNet && !ipNet.IP.IsLoopback() {
				if ipNet.IP.To4() != nil {
					ips[byName.Name] = ipNet.IP.String()
				}
			}

		}
	}
	return ips, nil
}
