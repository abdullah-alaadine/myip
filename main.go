package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"reflect"
	"sync"
	"time"
)

var endpoint string = "https://ipinfo.io/"

type result struct {
	IpAddress string `json:"ip"`
	Country   string `json:"country,omitempty"`
	City      string `json:"city,omitempty"`
	Region    string `json:"region,omitempty"`
	Location  string `json:"loc,omitempty"`
	Origin    string `json:"org,omitempty"`
	HostName  string `json:"hostname,omitempty"`
}

func main() {
	inJSON := flag.Bool("json", false, "display results in JSON format")
	rich := flag.Bool("rich", false, "display results in rich JSON format | more information")
	flag.Parse()

	var IPs struct {
		PublicIP  string `json:"publicIP"`
		PrivateIP string `json:"privateIP"`
	}

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	wg := &sync.WaitGroup{}

	var res *result

	wg.Add(1)
	go func(httpClient *http.Client, rich bool) {
		defer wg.Done()
		if !rich {
			publicIP, err := getPublicIP(httpClient)
			if err != nil {
				log.Println("Failed to retrieve public IP:", err)
				return
			}
			IPs.PublicIP = publicIP
			return
		}
		var err error
		res, err = getPublicIPRich(httpClient)
		if err != nil {
			log.Println("Failed to retrieve public IP:", err)
			return
		}
	}(httpClient, *rich)

	wg.Add(1)
	go func() {
		defer wg.Done()
		privateIP, err := getPrivateIP()
		if err != nil {
			log.Println("Failed to retrieve private IP:", err)
			return
		}
		IPs.PrivateIP = privateIP
	}()

	wg.Wait()

	fmt.Println()
	if *inJSON {
		if !*rich {
			err := json.NewEncoder(os.Stdout).Encode(IPs)
			if err != nil {
				log.Fatal("failed to encode IPs to JSON:", err)
			}
			return
		}
		data, err := json.MarshalIndent(struct {
			PublicIP  string `json:"public IP"`
			PrivateIP string `json:"private IP"`
			Info      result `json:"info"`
		}{
			PrivateIP: IPs.PrivateIP,
			PublicIP:  res.IpAddress,
			Info:      *res,
		}, "", "    ")
		if err != nil {
			log.Fatal("failed to encode data to JSON:", err)
		}
		fmt.Println(string(data))
		return
	}

	if *rich {
		v := reflect.ValueOf(*res)
		for i := 0; i < v.NumField(); i++ {
			fmt.Printf("%-11s: %s\n", v.Type().Field(i).Name, v.Field(i).Interface())
		}
		fmt.Printf("%-11s: %s\n", "private IP", IPs.PrivateIP)
		return
	}

	fmt.Printf("%-11s: %s\n", "public IP", IPs.PublicIP)
	fmt.Printf("%-11s: %s\n", "private IP", IPs.PrivateIP)
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
	resp, err := client.Get(endpoint + "ip")
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

	return string(data), nil
}

func getPublicIPRich(client *http.Client) (*result, error) {
	resp, err := client.Get(endpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP request failed with status code: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	result := &result{}

	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
