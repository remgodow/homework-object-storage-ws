package main

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestGetDockerContainers(t *testing.T) {
	c, e := getDockerContainers("amazin-object-storage-node")
	if e != nil {
		t.Fatal(e)
	}
	for _, cont := range c {
		fmt.Printf("%s %s\n", cont.ID[:10], cont.Image)
	}
}

func TestNewMinioConnection(t *testing.T) {
	c, e := getDockerContainers("amazin-object-storage-node")
	if e != nil {
		t.Fatal(e)
	}

	for _, cont := range c {
		_, err := NewMinioConnection("homework-object-storage-ws_amazin-object-storage", "9000", cont)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestMinioPut(t *testing.T) {
	req := struct {
		Type string `json:"type"`
		Id   string `json:"id"`
		Data string `json:"data"`
	}{
		"PUT",
		"ABCD123",
		"Test minio data",
	}

	requestBytes, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	request := make(map[string]interface{})
	err = json.Unmarshal(requestBytes, &request)
	if err != nil {
		t.Fatal(err)
	}

	handler := MinioPutMethod{}
	resp := handler.Handle(request)
	respBytes, err := json.Marshal(resp)
	if err != nil {
		t.Fatal(err)
	}

	response := make(map[string]interface{})
	err = json.Unmarshal(respBytes, &response)
	if err != nil {
		t.Fatal(err)
	}

	if code, ok := response["code"]; !ok {
		t.Fatal("missing code")
	} else if val, ok := code.(float64); !ok {
		t.Fatal("response code is not a number")
	} else {
		if val != 200 {
			t.Fatal("response code is not a 200")
		}
	}
}

func TestMinioGetRoute(t *testing.T) {
	req := struct {
		Type string `json:"type"`
		Id   string `json:"id"`
	}{
		"GET",
		"ABCD123",
	}

	requestBytes, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	request := make(map[string]interface{})
	err = json.Unmarshal(requestBytes, &request)
	if err != nil {
		t.Fatal(err)
	}

	handler := MinioGetMethod{}
	resp := handler.Handle(request)
	respBytes, err := json.Marshal(resp)
	if err != nil {
		t.Fatal(err)
	}

	response := make(map[string]interface{})
	err = json.Unmarshal(respBytes, &response)
	if err != nil {
		t.Fatal(err)
	}

	if _, ok := response["data"]; !ok {
		t.Fatal("missing data")
	}

}

func TestMinioGetNoKey(t *testing.T) {
	req := struct {
		Type string `json:"type"`
		Id   string `json:"id"`
	}{
		"GET",
		"ABCD1231",
	}

	requestBytes, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	request := make(map[string]interface{})
	err = json.Unmarshal(requestBytes, &request)
	if err != nil {
		t.Fatal(err)
	}

	handler := MinioGetMethod{}
	resp := handler.Handle(request)
	respBytes, err := json.Marshal(resp)
	if err != nil {
		t.Fatal(err)
	}

	response := make(map[string]interface{})
	err = json.Unmarshal(respBytes, &response)
	if err != nil {
		t.Fatal(err)
	}

	if data, ok := response["data"]; !ok {
		t.Fatal("missing data")
	} else if val, ok := data.(string); ok {
		if val != "null" {
			t.Fatal("null data expected")
		}
	} else {
		t.Fatal("data is not a string")
	}

}
