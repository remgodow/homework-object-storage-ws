package main

import (
	"bytes"
	"context"
	"errors"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/minio/minio-go"
	"hash/fnv"
	"io"
	"log"
	"strings"
)

const (
	containerNamePrefix    = "amazin-object-storage-node"
	containerNetworkPrefix = "homework-object-storage-ws_amazin-object-storage"
	minioPort              = "9000"
	bucketName             = "homework"
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

func (c *MinioConnection) PutValue(bucket, id, data string) error {
	minioClient, err := minio.New(c.endpoint, c.accessKeyID, c.secretAccessKey, false)
	if err != nil {
		log.Println(err)
		return err
	}
	if ok, err := minioClient.BucketExists(bucket); !ok {
		if err != nil {
			log.Println(err)
			return err
		}
		err := minioClient.MakeBucket(bucket, "")
		if err != nil {
			log.Println(err)
			return err
		}
	}

	b := bytes.NewReader([]byte(data))
	size := int64(b.Len())
	written, err := minioClient.PutObject(bucket, id, b, size, minio.PutObjectOptions{})
	if err != nil {
		log.Println(err)
		return err
	}
	if written != size {
		log.Println("data uploaded partially")
		return errors.New("data uploaded patrially")
	}

	return nil
}

func HashIdToRange(id string, n uint16) (int, error) {
	strlen := len(id)
	h := fnv.New64a()
	written, err := h.Write([]byte(id))
	if err != nil {
		return 0, err
	}
	if written != strlen {
		return 0, errors.New("Could not hash id")
	}
	val := h.Sum64()
	modulo := uint64(n)
	ret := val % modulo
	return int(ret), nil

}

type ErrBucketDoesNotExist struct{}

func (ErrBucketDoesNotExist) Error() string {
	return "bucket does not exist"
}

func (c *MinioConnection) GetValue(bucket, id string) (string, error) {
	minioClient, err := minio.New(c.endpoint, c.accessKeyID, c.secretAccessKey, false)
	if err != nil {
		log.Println(err)
		return "", err
	}
	if ok, err := minioClient.BucketExists(bucket); ok {
		if err != nil {
			log.Println(err)
			return "", err
		}
		obj, err := minioClient.GetObject(bucket, id, minio.GetObjectOptions{})
		if err != nil {
			log.Println(err)
			return "nil", err
		}
		stat, err := obj.Stat()
		if err != nil {
			log.Println(err)
			return "", err
		}

		var writeBuf bytes.Buffer

		written, err := io.CopyN(&writeBuf, obj, stat.Size)
		if err != nil {
			log.Println(err)
			return "", err
		}
		if written != stat.Size {
			log.Println("Object partially read")
			return "", errors.New("Object patrially read")
		}

		return writeBuf.String(), nil
	} else {
		return "", ErrBucketDoesNotExist{}
	}

}

// MinioGetRoute Both MinioGetRoute and MinioPutRoute assume that the minio cluster does not chage while the app is working
//they use Hash function + modulo to select index of container from the array
type MinioGetRoute struct {
}

func (MinioGetRoute) GetType() string {
	return "GET"
}

func (MinioGetRoute) Handle(request map[string]interface{}) interface{} {
	var id string
	if idval, ok := request["id"]; !ok {
		return ErrorResponse{Code: 400, Message: "id field is missing"}
	} else {
		if val, ok := idval.(string); !ok {
			return ErrorResponse{Code: 400, Message: "id field must be a string"}
		} else {
			id = val
		}
	}

	containers, err := getDockerContainers(containerNamePrefix)
	if err != nil {
		return ErrorResponse{Code: 500, Message: "Internal Server Error"}
	}
	idx, err := HashIdToRange(id, uint16(len(containers)))
	if err != nil {
		return ErrorResponse{Code: 500, Message: "Internal Server Error"}
	}
	container := containers[idx]
	connection, err := NewMinioConnection(containerNetworkPrefix, minioPort, container)
	if err != nil {
		return ErrorResponse{Code: 500, Message: "Internal Server Error"}
	}

	value, err := connection.GetValue(bucketName, id)
	if err != nil {
		if merr, ok := err.(minio.ErrorResponse); ok {
			if merr.Code == "NoSuchKey" {
				return struct {
					Data string `json:"data"`
				}{"null"}
			}
		} else if errors.As(err, &ErrBucketDoesNotExist{}) {
			return struct {
				Data string `json:"data"`
			}{"null"}
		}

		return ErrorResponse{Code: 500, Message: "Internal Server Error"}
	}

	return struct {
		Data string `json:"data"`
	}{value}
}

type MinioPutRoute struct {
}

func (MinioPutRoute) GetType() string {
	return "PUT"
}

func (MinioPutRoute) Handle(request map[string]interface{}) interface{} {
	var id, data string
	if idval, ok := request["id"]; !ok {
		return ErrorResponse{Code: 400, Message: "id field is missing"}
	} else {
		if val, ok := idval.(string); !ok {
			return ErrorResponse{Code: 400, Message: "id field must be a string"}
		} else {
			id = val
		}
	}
	if dataval, ok := request["data"]; !ok {
		return ErrorResponse{Code: 400, Message: "data field is missing"}
	} else {
		if val, ok := dataval.(string); !ok {
			return ErrorResponse{Code: 400, Message: "data field must be a string"}
		} else {
			data = val
		}
	}

	containers, err := getDockerContainers(containerNamePrefix)
	if err != nil {
		return ErrorResponse{Code: 500, Message: "Internal Server Error"}
	}
	idx, err := HashIdToRange(id, uint16(len(containers)))
	if err != nil {
		return ErrorResponse{Code: 500, Message: "Internal Server Error"}
	}
	container := containers[idx]
	connection, err := NewMinioConnection(containerNetworkPrefix, minioPort, container)
	if err != nil {
		return ErrorResponse{Code: 500, Message: "Internal Server Error"}
	}

	err = connection.PutValue(bucketName, id, data)
	if err != nil {
		return ErrorResponse{Code: 500, Message: "Internal Server Error"}
	}

	return struct {
		Code    int16  `json:"code"`
		Message string `json:"message"`
	}{
		200,
		"OK",
	}
}
