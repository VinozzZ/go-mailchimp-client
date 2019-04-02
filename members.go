package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

// ListOfMembers represents the response payload from GET - /lists/${listID}/members
type ListOfMembers struct {
	TotalItems int      `json:"total_items"`
	Members    []Member `json:"members"`
}

type MemberStats struct {
	AvgOpenRate  float64 `json:"avg_open_rate"`
	AvgClickRate float64 `json:"avg_click_rate"`
}

type Member struct {
	ID              string      `json:"id"`
	ListID          string      `json:"list_id"`
	EmailAddress    string      `json:"email_address"`
	UniqueEMailID   string      `json:"unique_email_id"`
	EmailType       string      `json:"email_type"`
	Stats           MemberStats `json:"stats"`
	IPSignup        string      `json:"ip_signup"`
	TimestampSignup string      `json:"timestamp_signup"`
	TimestampOpt    string      `json:"timestamp_opt"`
	MemberRating    int         `json:"member_rating"`
	LastChanged     string      `json:"last_changed"`
	EmailClient     string      `json:"email_client"`
}

func GetMembers(segmentID string) (*ListOfMembers, error) {
	remainingListURL := fmt.Sprintf("/lists/%s/segments/%s/members?offset=0&count=100", listID, segmentID)
	getRemainingListURL := baseURL + remainingListURL

	client := &http.Client{}
	request := getHTTPRequest("GET", getRemainingListURL, nil)

	resp, err := client.Do(request)
	if err != nil {
		log.Println("GET Request Error", err)
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Request Error: ", err)
	}

	var response = new(ListOfMembers)
	json.Unmarshal(data, response)

	return response, err
}
