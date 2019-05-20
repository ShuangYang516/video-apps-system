package util

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	log "qiniupkg.com/x/log.v7"
)

func Post(url string, data []byte) (body []byte, err error) {
	if !strings.HasPrefix(url, "http") {
		url = "http://" + url
	}

	client := http.Client{
		Timeout: 30 * time.Second,
	}
	req, err := http.NewRequest("POST", url, bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")

	response, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer response.Body.Close()

	body, err = ioutil.ReadAll(response.Body)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	if response.StatusCode >= 300 || response.StatusCode < 200 {
		log.Println("response status code error :", response.StatusCode, "; body:", string(body))
		return nil, errors.New(string(body))
	}
	// log.Println(string(body), response.StatusCode)
	return body, nil
}

func Get(url string) (body []byte, err error) {
	if !strings.HasPrefix(url, "http") {
		url = "http://" + url
	}

	client := http.Client{
		Timeout: 30 * time.Second,
	}
	req, err := http.NewRequest("GET", url, nil)

	response, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer response.Body.Close()

	body, err = ioutil.ReadAll(response.Body)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	if response.StatusCode >= 300 || response.StatusCode < 200 {
		log.Println("response status code error :", response.StatusCode, "; body:", string(body))
		return nil, errors.New(string(body))
	}
	// log.Println(string(body), response.StatusCode)
	return body, nil
}

func PostRaw(url string, data []byte, header map[string]string) (body []byte, err error) {
	if !strings.HasPrefix(url, "http") {
		url = "http://" + url
	}

	client := http.Client{
		Timeout: 30 * time.Second,
	}
	req, err := http.NewRequest("POST", url, bytes.NewReader(data))
	for k, v := range header {
		req.Header.Set(k, v)
	}
	response, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer response.Body.Close()

	body, err = ioutil.ReadAll(response.Body)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	if response.StatusCode >= 300 || response.StatusCode < 200 {
		log.Println("response status code error :", response.StatusCode, "; body:", string(body))
		return nil, errors.New(string(body))
	}
	// log.Println(string(body), response.StatusCode)
	return body, nil
}
