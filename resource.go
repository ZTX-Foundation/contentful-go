package contentful

import (
	"bytes"
	"fmt"
	"io"
	"net/url"

	"github.com/spf13/afero"
)

// ResourcesService service
type ResourcesService service

// Resource model
type Resource struct {
	Sys      *Sys `json:"sys,omitempty"`
	FilePath string
}

func (service *ResourcesService) List(spaceID string) *Collection {
	path := fmt.Sprintf("/spaces/%s/uploads", spaceID)
	query := url.Values{}
	method := "GET"

	req, _ := service.c.newRequest(method, path, query, nil)

	col := NewCollection(&CollectionOptions{})
	col.c = service.c
	col.req = req

	return col
}

// Get returns a single resource/upload
func (service *ResourcesService) Get(spaceID, resourceID string) (*Resource, error) {
	path := fmt.Sprintf("/spaces/%s/uploads/%s", spaceID, resourceID)
	query := url.Values{}
	method := "GET"

	req, err := service.c.newRequest(method, path, query, nil)
	if err != nil {
		return &Resource{}, err
	}

	var resource Resource
	if ok := service.c.do(req, &resource); ok != nil {
		return nil, err
	}

	return &resource, err
}

// UpsertWithFs creates an upload resource using the provided filesystem
func (service *ResourcesService) UpsertWithFs(spaceID string, resource *Resource, fs afero.Fs) error {
	file, err := fs.Open(resource.FilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	return service.UpsertByteBuffer(spaceID, resource, file)
}

// Create creates an upload resource
func (service *ResourcesService) Upsert(spaceID string, resource *Resource) error {
	return service.UpsertWithFs(spaceID, resource, afero.NewOsFs())
}

// UpsertByteBuffer is a streaming upload
func (service *ResourcesService) UpsertByteBuffer(spaceID string, resource *Resource, reader io.Reader) error {
	body := &bytes.Buffer{}

	if _, err := io.Copy(body, reader); err != nil {
		return err
	}

	path := fmt.Sprintf("/spaces/%s/uploads", spaceID)
	method := "POST"

	req, err := service.c.newRequest(method, path, nil, body)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/octet-stream")

	return service.c.do(req, resource)
}

// Delete the resource
func (service *ResourcesService) Delete(spaceID string, resource *Resource) error {
	path := fmt.Sprintf("/spaces/%s/uploads/%s", spaceID, resource.Sys.ID)
	method := "DELETE"

	req, err := service.c.newRequest(method, path, nil, nil)
	if err != nil {
		return err
	}

	return service.c.do(req, nil)
}
