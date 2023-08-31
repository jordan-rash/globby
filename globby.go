package main

import (
	"net/http"
	"strings"
	"sync"

	"github.com/wasmcloud/actor-tinygo"
	"github.com/wasmcloud/interfaces/blobstore/tinygo"
	"github.com/wasmcloud/interfaces/httpserver/tinygo"
	"github.com/wasmcloud/interfaces/logging/tinygo"
)

const (
	DEFAULT_CONTAINER_NAME string = "default"
	CONTAINER_HEADER_NAME  string = "blobby-container"
	CONTAINER_PARAM_NAME   string = "container"
)

func main() {
	me := Globby{
		logger:                logging.NewProviderLogging(),
		initializedContainers: make(map[string]struct{}),
		mu:                    sync.Mutex{},
	}
	actor.RegisterHandlers(httpserver.HttpServerHandler(&me))
}

type Globby struct {
	logger                *logging.LoggingSender
	initializedContainers map[string]struct{}
	mu                    sync.Mutex
}

func (e *Globby) HandleRequest(ctx *actor.Context, req httpserver.HttpRequest) (*httpserver.HttpResponse, error) {

	cleanID := strings.Trim(req.Path, "/")
	if len(strings.Split(cleanID, "/")) > 1 {
		return &httpserver.HttpResponse{
			StatusCode: http.StatusBadRequest,
			Header:     make(httpserver.HeaderMap, 0),
			Body:       []byte("Cannot use a subpathed file name (e.g. foo/bar.txt)"),
		}, nil
	}

	containerName, err := e.ensureContainer(ctx, getContainerName(&req))
	if err != nil {
		return nil, err
	}

	_ = e.logger.WriteLog(ctx, logging.LogEntry{Level: "info", Text: "clean id: " + cleanID})
	_ = e.logger.WriteLog(ctx, logging.LogEntry{Level: "info", Text: "container name: " + containerName})

	switch req.Method {
	case http.MethodGet:
		return e.getObject(ctx, containerName, cleanID)
	case http.MethodPost:
		_ = e.logger.WriteLog(ctx, logging.LogEntry{Level: "info", Text: "content-type: " + strings.Join(req.Header["content-type"], ",")})
		contentType := req.Header["content-type"]
		return e.putObject(ctx, containerName, cleanID, req.Body, contentType[1])
	case http.MethodDelete:
		return e.deleteObject(ctx, containerName, cleanID)
	default:
		return &httpserver.HttpResponse{
			StatusCode: http.StatusMethodNotAllowed,
			Header:     httpserver.HeaderMap{"Allow": []string{http.MethodGet, http.MethodPut, http.MethodDelete}},
			Body:       []byte("bad request"),
		}, nil
	}
}

func (e *Globby) putObject(ctx *actor.Context, containerName, objectName string, data []byte, contentType string) (*httpserver.HttpResponse, error) {
	bs := blobstore.NewProviderBlobstore()

	_, err := bs.PutObject(ctx,
		blobstore.PutObjectRequest{
			Chunk: blobstore.Chunk{
				ContainerId: blobstore.ContainerId(containerName),
				ObjectId:    blobstore.ObjectId(objectName),
				Bytes:       data,
				Offset:      0,
				IsLast:      true,
			},
			ContentType: contentType,
		})

	if err != nil {
		return nil, err
	}

	return &httpserver.HttpResponse{
		StatusCode: http.StatusOK,
		Header:     make(httpserver.HeaderMap, 0),
		Body:       []byte("upload complete"),
	}, nil
}

func (e *Globby) deleteObject(ctx *actor.Context, containerName, objectName string) (*httpserver.HttpResponse, error) {
	bs := blobstore.NewProviderBlobstore()
	_, err := bs.RemoveObjects(ctx, blobstore.RemoveObjectsRequest{
		ContainerId: blobstore.ContainerId(containerName),
		Objects:     blobstore.ObjectIds{blobstore.ObjectId(objectName)},
	})
	if err != nil {
		return nil, err
	}

	return &httpserver.HttpResponse{
		StatusCode: http.StatusOK,
		Header:     make(httpserver.HeaderMap, 0),
		Body:       []byte("delete successful"),
	}, nil
}

func (e *Globby) getObject(ctx *actor.Context, containerName, objectName string) (*httpserver.HttpResponse, error) {
	bs := blobstore.NewProviderBlobstore()
	exists, err := bs.ObjectExists(ctx,
		blobstore.ContainerObject{
			ContainerId: blobstore.ContainerId(containerName),
			ObjectId:    blobstore.ObjectId(objectName),
		})
	if err != nil {
		return nil, err
	}

	if !exists {
		return &httpserver.HttpResponse{
			StatusCode: http.StatusNotFound,
			Header:     make(httpserver.HeaderMap, 0),
			Body:       []byte("not found"),
		}, nil
	}

	object, err := bs.GetObject(ctx, blobstore.GetObjectRequest{
		ObjectId:    blobstore.ObjectId(objectName),
		ContainerId: blobstore.ContainerId(containerName),
		RangeStart:  0,
		RangeEnd:    18446744073709551614,
	})
	if err != nil {
		return nil, err
	}

	_ = e.logger.WriteLog(ctx, logging.LogEntry{Level: "info", Text: "successfully got an object!"})

	body := object.InitialChunk.Bytes
	if len(body) == 0 {
		return &httpserver.HttpResponse{
			StatusCode: http.StatusBadGateway,
			Header:     make(httpserver.HeaderMap, 0),
			Body:       []byte("Blobstore sent empty data chunk when full file was requested"),
		}, nil
	}

	return &httpserver.HttpResponse{
		StatusCode: http.StatusOK,
		Header:     httpserver.HeaderMap{"Content-Type": []string{object.ContentType}},
		Body:       body,
	}, nil
}

func getContainerName(req *httpserver.HttpRequest) string {
	return DEFAULT_CONTAINER_NAME
}

func (e *Globby) ensureContainer(ctx *actor.Context, name string) (string, error) {
	_, ok := e.initializedContainers[name]
	if !ok {
		bs := blobstore.NewProviderBlobstore()
		exists, err := bs.ContainerExists(ctx, blobstore.ContainerId(name))
		if err != nil {
			return "", err
		}
		if !exists {
			err := bs.CreateContainer(ctx, blobstore.ContainerId(name))
			if err != nil {
				return "", err
			}
		}

		e.mu.Lock()
		defer e.mu.Unlock()

		e.initializedContainers[name] = struct{}{}
	}

	return name, nil
}
