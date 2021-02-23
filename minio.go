package main

import (
	"context"
	"errors"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/minio/minio-go"
	"log"
	"strings"
)

func getDockerContainers(namePrefix string) ([]types.ContainerJSON, error) {
	cli, err := client.NewClientWithOpts(client.WithHost("unix:///var/run/docker.sock"), client.WithHTTPClient(nil))
	if err != nil {
		return nil, err
	}

	filterNames := filters.NewArgs()
	filterNames.Add("name", namePrefix)
	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{
		Filters: filterNames,
	})
	if err != nil {
		return nil, err
	}

	var results []types.ContainerJSON
	for _, container := range containers {
		result, err := cli.ContainerInspect(context.Background(), container.ID)
		if err != nil {
			log.Printf("Could not inspect container %s", container.ID)
		}
		results = append(results, result)
	}

	return results, err
}

type MinioConnection struct {
	endpoint        string
	accessKeyID     string
	secretAccessKey string
}

func NewMinioConnection(network string, port string, container types.ContainerJSON) (*MinioConnection, error) {
	var (
		endpoint,
		accessKeyId,
		secretAccessKey string
	)

	networks := container.NetworkSettings.Networks
	if network, ok := networks[network]; ok {
		endpoint += network.IPAddress + ":" + port
	} else {
		return nil, errors.New("Container is not connected to the network")
	}

	envs := container.Config.Env
	for _, env := range envs {
		kv := strings.Split(env, "=")
		if kv[0] == "MINIO_ACCESS_KEY" {
			accessKeyId = kv[1]
		} else if kv[0] == "MINIO_SECRET_KEY" {
			secretAccessKey = kv[1]
		}
	}

	if accessKeyId == "" || secretAccessKey == "" {
		return nil, errors.New("Missing ENV secrets")
	}

	return &MinioConnection{
		endpoint:        endpoint,
		accessKeyID:     accessKeyId,
		secretAccessKey: secretAccessKey,
	}, nil
}

func (c *MinioConnection) GetValue(bucket, id string) (string, error) {
	minioClient, err := minio.New(c.endpoint, c.accessKeyID, c.secretAccessKey, false)
	if err != nil {
		log.Fatalln(err)
		return "", err
	}
	if ok, err := minioClient.BucketExists(bucket); ok {
		if err != nil {
			log.Fatalln(err)
			return "", err
		}
		//obj, err := minioClient.GetObject(bucket, id, nil)
	}
	return "", nil
}
