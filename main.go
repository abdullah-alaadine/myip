package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

func main() {
	var IPs struct {
		PublicIP  string
		PrivateIP string
	}

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	wg := &sync.WaitGroup{}

	wg.Add(1)
	go func(httpClient *http.Client) {
		publicIP, err := getPublicIP(httpClient)
		if err != nil {
			log.Println("Failed to retrieve public IP:", err)
			return
		}
		IPs.PublicIP = publicIP
		wg.Done()
	}(httpClient)

	wg.Add(1)
	go func() {
		privateIP, err := getPrivateIP()
		if err != nil {
			log.Println("Failed to retrieve private IP:", err)
			return
		}
		IPs.PrivateIP = privateIP
		wg.Done()
	}()

	wg.Wait()
}

func getPrivateIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String(), nil
}

func getPublicIP(client *http.Client) (string, error) {
	resp, err := client.Get("https://ifconfig.me/all.json")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP request failed with status code: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result struct {
		IpAddress string `json:"ip_addr"`
	}

	err = json.Unmarshal(data, &result)
	if err != nil {
		return "", err
	}

	return result.IpAddress, nil
}
