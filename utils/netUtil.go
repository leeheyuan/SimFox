package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
)

func HttpGet(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, errors.New("访问失败：" + url)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	return body, err
}

func HttpGetJSON(url string) (map[string]interface{}, error) {
	body, err := HttpGet(url)
	if err != nil {
		return nil, err
	}
	var f map[string]interface{}
	err = json.Unmarshal(body, &f)
	if err != nil {
		return nil, errors.New("JSON解析失败：" + err.Error())
	}
	return f, nil
}

func HttpPost(url string, xml string) string {
	client := &http.Client{}

	req, err := http.NewRequest("POST", url, strings.NewReader(xml))
	if err != nil {
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	return string(body)

}

func HttpPostJson(url string, v interface{}) (*http.Response, error) {
	postBody, _ := json.Marshal(v)
	req_new := bytes.NewBuffer([]byte(postBody))
	request, _ := http.NewRequest("POST", url, req_new)
	request.Header.Set("Content-type", "application/json")
	client := &http.Client{}
	response, _ := client.Do(request)
	if response.StatusCode == 200 {
		return response, nil
	}
	return nil, errors.New("errorStatusCode not 200")
}
