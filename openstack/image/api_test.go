// Copyright (c) 2016 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package image

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

// TBD - can some of this stuff be pulled out into a common test area?
type test struct {
	method           string
	pattern          string
	handler          func(*Context, http.ResponseWriter, *http.Request) (APIResponse, error)
	request          string
	expectedStatus   int
	expectedResponse string
}

var tests = []test{
	{
		"GET",
		"/",
		listAPIVersions,
		"",
		http.StatusOK,
		`{"versions":[{"status":"CURRENT","id":"v2.3","links":[{"href":"` + fmt.Sprintf("https://%s:9292/v2/", myHostname()) + `","rel":"self"}]}]}`,
	},
	{
		"POST",
		"/v2/images",
		createImage,
		`{"container_format":"bare","disk_format":"raw","name":"Ubuntu","id":"b2173dd3-7ad6-4362-baa6-a68bce3565cb"}`,
		http.StatusCreated,
		`{"status":"queued","container_format":"bare","min_ram":0,"updated_at":"2015-11-29T22:21:42Z","owner":"bab7d5c60cd041a0a36f7c4b6e1dd978","min_disk":0,"tags":[],"locations":[],"visibility":"private","id":"b2173dd3-7ad6-4362-baa6-a68bce3565cb","size":null,"virtual_size":null,"name":"Ubuntu","checksum":null,"created_at":"2015-11-29T22:21:42Z","disk_format":"raw","properties":null,"protected":false,"self":"/v2/images/b2173dd3-7ad6-4362-baa6-a68bce3565cb","file":"/v2/images/b2173dd3-7ad6-4362-baa6-a68bce3565cb/file","schema":"/v2/schemas/image"}`,
	},
	{
		"GET",
		"/v2/images",
		listImages,
		"",
		http.StatusOK,
		`{"images":[{"status":"queued","container_format":"bare","min_ram":0,"updated_at":"2015-11-29T22:21:42Z","owner":"bab7d5c60cd041a0a36f7c4b6e1dd978","min_disk":0,"tags":[],"locations":[],"visibility":"private","id":"b2173dd3-7ad6-4362-baa6-a68bce3565cb","size":null,"virtual_size":null,"name":"Ubuntu","checksum":null,"created_at":"2015-11-29T22:21:42Z","disk_format":"raw","properties":null,"protected":false,"self":"/v2/images/b2173dd3-7ad6-4362-baa6-a68bce3565cb","file":"/v2/images/b2173dd3-7ad6-4362-baa6-a68bce3565cb/file","schema":"/v2/schemas/image"}],"schema":"/v2/schemas/images","first":"/v2/images"}`,
	},
}

func myHostname() string {
	host, _ := os.Hostname()
	return host
}

type testImageService struct{}

func (is testImageService) CreateImage(req CreateImageRequest) (DefaultResponse, error) {
	format := Bare
	name := "Ubuntu"
	createdAt, _ := time.Parse(time.RFC3339, "2015-11-29T22:21:42Z")
	updatedAt, _ := time.Parse(time.RFC3339, "2015-11-29T22:21:42Z")
	minDisk := 0
	minRAM := 0
	owner := "bab7d5c60cd041a0a36f7c4b6e1dd978"

	return DefaultResponse{
		Status:          Queued,
		ContainerFormat: &format,
		CreatedAt:       createdAt,
		Tags:            make([]string, 0),
		DiskFormat:      Raw,
		Visibility:      Private,
		UpdatedAt:       &updatedAt,
		Locations:       make([]string, 0),
		Self:            "/v2/images/b2173dd3-7ad6-4362-baa6-a68bce3565cb",
		MinDisk:         &minDisk,
		Protected:       false,
		ID:              "b2173dd3-7ad6-4362-baa6-a68bce3565cb",
		File:            "/v2/images/b2173dd3-7ad6-4362-baa6-a68bce3565cb/file",
		Owner:           &owner,
		MinRAM:          &minRAM,
		Schema:          "/v2/schemas/image",
		Name:            &name,
	}, nil
}

func (is testImageService) ListImages() ([]DefaultResponse, error) {
	format := Bare
	name := "Ubuntu"
	createdAt, _ := time.Parse(time.RFC3339, "2015-11-29T22:21:42Z")
	updatedAt, _ := time.Parse(time.RFC3339, "2015-11-29T22:21:42Z")
	minDisk := 0
	minRAM := 0
	owner := "bab7d5c60cd041a0a36f7c4b6e1dd978"

	image := DefaultResponse{
		Status:          Queued,
		ContainerFormat: &format,
		CreatedAt:       createdAt,
		Tags:            make([]string, 0),
		DiskFormat:      Raw,
		Visibility:      Private,
		UpdatedAt:       &updatedAt,
		Locations:       make([]string, 0),
		Self:            "/v2/images/b2173dd3-7ad6-4362-baa6-a68bce3565cb",
		MinDisk:         &minDisk,
		Protected:       false,
		ID:              "b2173dd3-7ad6-4362-baa6-a68bce3565cb",
		File:            "/v2/images/b2173dd3-7ad6-4362-baa6-a68bce3565cb/file",
		Owner:           &owner,
		MinRAM:          &minRAM,
		Schema:          "/v2/schemas/image",
		Name:            &name,
	}

	var images []DefaultResponse
	images = append(images, image)

	return images, nil
}

func (is testImageService) GetImage(string) (DefaultResponse, error) {
	return DefaultResponse{}, nil
}

func (is testImageService) UploadImage(string, io.Reader) (UploadImageResponse, error) {
	return UploadImageResponse{}, nil
}

func TestRoutes(t *testing.T) {
	var is testImageService
	config := APIConfig{9292, is}

	r := Routes(config)
	if r == nil {
		t.Fatalf("No routes returned")
	}
}

func TestAPIResponse(t *testing.T) {
	var is testImageService

	// TBD: add context to test definition so it can be created per
	// endpoint with either a pass testVolumeService or a failure
	// one.
	context := &Context{9292, is}

	for _, tt := range tests {
		req, err := http.NewRequest(tt.method, tt.pattern, bytes.NewBuffer([]byte(tt.request)))
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := APIHandler{context, tt.handler}

		handler.ServeHTTP(rr, req)

		status := rr.Code
		if status != tt.expectedStatus {
			t.Errorf("got %v, expected %v", status, tt.expectedStatus)
		}

		if rr.Body.String() != tt.expectedResponse {
			t.Errorf("%s: failed\ngot: %v\nexp: %v", tt.pattern, rr.Body.String(), tt.expectedResponse)
		}
	}
}
