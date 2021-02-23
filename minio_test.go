package main

import (
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
