package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type SegmentBatchRequest struct {
	MembersToAdd    []string `json:"members_to_add"`
	MembersToRemove []string `json:"members_to_remove"`
}

type SegmentBatchResponse struct {
	MembersAdded   []Member `json:"members_added"`
	MembersRemoved []Member `json:"members_removed"`

	TotalAdded   int `json:"total_added"`
	TotalRemoved int `json:"total_removed"`
	ErrorCount   int `json:"error_count"`
}

func SetTags(segmentID string, members []Member, method string) (*SegmentBatchResponse, error) {
	actionURL := fmt.Sprintf("/lists/%s/segments/%s", listID, segmentID)
	requestURL := baseURL + actionURL

	requetPayload, err := GetSegmentPayload(method, segmentID, members)
	if err != nil {
		log.Println(err)
	}

	data, _ := json.Marshal(requetPayload)
	client := &http.Client{}
	request := getHTTPRequest("POST", requestURL, data)

	resp, err := client.Do(request)
	if err != nil {
		log.Println("GET Request Error", err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Request Error: ", err)
	}
	defer resp.Body.Close()

	response := new(SegmentBatchResponse)
	json.Unmarshal(body, response)
	return response, err
}

func GetSegmentPayload(method string, segmentID string, data []Member) (SegmentBatchRequest, error) {
	var payload []string
	for _, member := range data {
		payload = append(payload, member.EmailAddress)
	}
	switch method {
	case "add":
		return SegmentBatchRequest{
			MembersToAdd:    payload,
			MembersToRemove: []string{},
		}, nil
	case "remove":
		return SegmentBatchRequest{
			MembersToAdd:    []string{},
			MembersToRemove: payload,
		}, nil
	default:
		return SegmentBatchRequest{}, errors.New("Failed to create request payload")
	}
}
