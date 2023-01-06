package contentful

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
)

// AssetsService service
type AssetsService service

// File model
type File struct {
	Name        string      `json:"fileName,omitempty"`
	ContentType string      `json:"contentType,omitempty"`
	URL         string      `json:"url,omitempty"`
	UploadURL   string      `json:"upload,omitempty"`
	Detail      *FileDetail `json:"details,omitempty"`
	UploadFrom  *UploadFrom `json:"uploadFrom,omitempty"`
}

// FileDetail model
type FileDetail struct {
	Size  int        `json:"size,omitempty"`
	Image *FileImage `json:"image,omitempty"`
}

// FileImage model
type FileImage struct {
	Width  int `json:"width,omitempty"`
	Height int `json:"height,omitempty"`
}

// FileFields model
type FileFields struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	File        *File  `json:"file,omitempty"`
}

// Asset model
type Asset struct {
	Locale string
	Sys    *Sys        `json:"sys"`
	Fields *FileFields `json:"fields"`
}

// UploadFrom model
type UploadFrom struct {
	Sys *Sys `json:"sys,omitempty"`
}

// MarshalJSON for custom json marshaling
func (asset *Asset) MarshalJSON() ([]byte, error) {
	payload := map[string]interface{}{
		"sys": "",
		"fields": map[string]interface{}{
			"title":       map[string]string{},
			"description": map[string]string{},
			"file":        map[string]interface{}{},
		},
	}

	if asset.Sys != nil {
		payload["sys"] = asset.Sys
	} else {
		delete(payload, "sys")
	}

	//payload["sys"] = asset.Sys
	fields := payload["fields"].(map[string]interface{})

	// title
	title := fields["title"].(map[string]string)
	title[asset.Locale] = asset.Fields.Title

	// description
	description := fields["description"].(map[string]string)
	description[asset.Locale] = asset.Fields.Description

	// file
	file := fields["file"].(map[string]interface{})
	file[asset.Locale] = asset.Fields.File

	return json.Marshal(payload)
}

// UnmarshalJSON for custom json unmarshaling
func (asset *Asset) UnmarshalJSON(data []byte) error {
	type Alias *Asset

	var payload map[string]interface{}
	if err := json.Unmarshal(data, &payload); err != nil {
		return err
	}

	if asset.Locale == "" {
		asset.Locale = "en"
	}

	asset.Sys = &Sys{}
	b, _ := json.Marshal(payload["sys"])
	if err := json.Unmarshal(b, asset.Sys); err != nil {
		return err
	}

	title := payload["fields"].(map[string]interface{})["title"]
	if title != nil {
		switch t := title.(type) {
		case string:
			title = t
		default:
			title = t.(map[string]interface{})[asset.Locale]
		}
	} else {
		title = ""
	}

	description := payload["fields"].(map[string]interface{})["description"]
	if description != nil {
		description = description.(map[string]interface{})[asset.Locale]
	} else {
		description = ""
	}

	asset.Fields = &FileFields{
		Title:       title.(string),
		Description: description.(string),
		File:        &File{},
	}

	file := payload["fields"].(map[string]interface{})["file"].(map[string]interface{})[asset.Locale]

	bFile, _ := json.Marshal(file)

	if err := json.Unmarshal(bFile, asset.Fields.File); err != nil {
		return err
	}

	return nil
}

// GetVersion returns entity version
func (asset *Asset) GetVersion() int {
	version := 1
	if asset.Sys != nil {
		version = asset.Sys.Version
	}

	return version
}

// List returns asset collection
func (service *AssetsService) List(spaceID string) *Collection {
	path := fmt.Sprintf("/spaces/%s/assets", spaceID)
	method := "GET"

	req, err := service.c.newRequest(method, path, nil, nil)
	if err != nil {
		return &Collection{}
	}

	col := NewCollection(&CollectionOptions{})
	col.c = service.c
	col.req = req

	return col
}

// Get returns a single asset entity
func (service *AssetsService) Get(spaceID, assetID string) (*Asset, error) {
	path := fmt.Sprintf("/spaces/%s/assets/%s", spaceID, assetID)
	method := "GET"

	req, err := service.c.newRequest(method, path, nil, nil)
	if err != nil {
		return nil, err
	}

	var asset Asset
	if err := service.c.do(req, &asset); err != nil {
		return nil, err
	}

	return &asset, nil
}

// Upsert updates or creates a new asset entity
func (service *AssetsService) Upsert(spaceID string, asset *Asset) error {
	bytesArray, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	var path string
	var method string

	if asset.Sys != nil && asset.Sys.CreatedAt != "" {
		path = fmt.Sprintf("/spaces/%s/assets/%s", spaceID, asset.Sys.ID)
		method = "PUT"
	} else {
		path = fmt.Sprintf("/spaces/%s/assets", spaceID)
		method = "POST"
	}

	req, err := service.c.newRequest(method, path, nil, bytes.NewReader(bytesArray))
	if err != nil {
		return err
	}

	req.Header.Set("X-Contentful-Version", strconv.Itoa(asset.GetVersion()))

	return service.c.do(req, asset)
}

// Delete sends delete request
func (service *AssetsService) Delete(spaceID string, asset *Asset) error {
	path := fmt.Sprintf("/spaces/%s/assets/%s", spaceID, asset.Sys.ID)
	method := "DELETE"

	req, err := service.c.newRequest(method, path, nil, nil)
	if err != nil {
		return err
	}

	version := strconv.Itoa(asset.Sys.Version)
	req.Header.Set("X-Contentful-Version", version)

	return service.c.do(req, nil)
}

// Process the asset
func (service *AssetsService) Process(spaceID string, asset *Asset) error {
	path := fmt.Sprintf("/spaces/%s/assets/%s/files/%s/process", spaceID, asset.Sys.ID, asset.Locale)
	method := "PUT"

	req, err := service.c.newRequest(method, path, nil, nil)
	if err != nil {
		return err
	}

	version := strconv.Itoa(asset.Sys.Version)
	req.Header.Set("X-Contentful-Version", version)

	return service.c.do(req, nil)
}

// Publish published the asset
func (service *AssetsService) Publish(spaceID string, asset *Asset) error {
	path := fmt.Sprintf("/spaces/%s/assets/%s/published", spaceID, asset.Sys.ID)
	method := "PUT"

	req, err := service.c.newRequest(method, path, nil, nil)
	if err != nil {
		return err
	}

	version := strconv.Itoa(asset.Sys.Version)
	req.Header.Set("X-Contentful-Version", version)

	return service.c.do(req, asset)
}
