/*
	A golang snippet that helps the team to send out mailchimp email updates based on tags
*/
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

var (
	baseURL            string
	listID             string
	remainingSegmentID string
	queuedID           string
	sentID             string
)

var cfg map[string]string

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("File .env not found, loading config from ENV")
	}

	cfg, _ = godotenv.Read()

	baseURL = cfg["BASEURL"]
	listID = cfg["LISTID"]
	remainingSegmentID = cfg["REMAININGSEGMENTID"]
	queuedID = cfg["QUEUEDID"]
	sentID = cfg["SENDID"]

	remainingMembers, err := GetMembers(remainingSegmentID)
	if err != nil || len(remainingMembers.Members) < 1 {
		log.Panicln("Failed to get remaining members")
	}

	if _, err := SetTags(queuedID, remainingMembers.Members, "add"); err != nil {
		log.Panicln("queue tag failed to be set")
	}

	templateID := GetTemplateFromYesterday()
	if templateID == 0 {
		log.Panicln("No template is found")
	}

	newCampaign, err := createNewCampaign(templateID)
	if err != nil {
		log.Panicln("New campaign creation failed")
	}

	isNewCampaignCreated := false
	for !isNewCampaignCreated {
		isNewCampaignCreated = checkCampaignExist(newCampaign.ID)
	}

	if err := sendCampaign(newCampaign.ID); err != nil {
		log.Panicln("Failed to send campaign")
	}

	if _, err := SetTags(sentID, remainingMembers.Members, "add"); err != nil {
		log.Panicln("Sent tag failed to be set")
	}

	if _, err := SetTags(queuedID, remainingMembers.Members, "remove"); err != nil {
		log.Panicln("queue tag failed to be removed")
	}

}

func getHTTPRequest(method string, requestURL string, postData []byte) *http.Request {
	payload := bytes.NewReader(postData)
	req, err := http.NewRequest(method, requestURL, payload)
	if err != nil {
		log.Println("HTTP New Request not creates", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth("yingrong", cfg["APIKEY"])
	return req
}

func createNewCampaign(templateID uint) (*campaignResponse, error) {
	var requestURL = baseURL + "/campaigns"
	segmentID, _ := strconv.Atoi(queuedID)

	//TODO: Create command-line input function to allow dynamically setting campaign settings
	requestPayload := CampaignCreationRequest{
		Type: "regular",
		Recipients: campaignCreationRecipients{
			ListID: listID,
			SegmentOptions: campaignCreationSegmentOptions{
				SavedSegmentID: segmentID,
				Match:          "any",
				Conditions: []campaignCreationConditions{
					campaignCreationConditions{
						ConditionType: "StaticSegment",
						Field:         "static_segment",
						Op:            "static_is",
						Value:         segmentID,
					},
				},
			},
		},
		Settings: CampaignCreationSettings{
			SubjectLine: "Zone",
			Title:       "Zone",
			FromName:    "Storj",
			ReplyTo:     "lionzhao0820@gmail.com",
			TemplateID:  templateID,
		},
	}

	data, err := json.Marshal(requestPayload)
	if err != nil {
		log.Println("Failed to marshal request payload for create new campaign")
	}
	client := &http.Client{}
	request := getHTTPRequest("POST", requestURL, data)

	resp, err := client.Do(request)
	if err != nil {
		log.Println("POST Request Failed, create new campaign error: ", err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Failed reading responde body from new campaign request: ", err)
	}
	defer resp.Body.Close()

	var newCampaign = new(campaignResponse)
	json.Unmarshal(body, newCampaign)

	return newCampaign, err
}

func sendCampaign(campaignID string) error {
	sendURL := fmt.Sprintf("/campaigns/%s/actions/send", campaignID)
	requestURL := baseURL + sendURL

	client := &http.Client{}
	request := getHTTPRequest("POST", requestURL, nil)

	_, err := client.Do(request)
	if err != nil {
		log.Println("POST Request Failed, create new campaign error: ", err)
	}

	return err
}

func getCampaign(campaignID string) (*campaignResponse, error) {
	campaignURL := fmt.Sprintf("/campaigns/%s", campaignID)
	requestURL := baseURL + campaignURL

	client := &http.Client{}
	request := getHTTPRequest("GET", requestURL, nil)
	campaignData := new(campaignResponse)

	resp, err := client.Do(request)
	if err != nil {
		log.Printf("GET Request Failed: campaign id %s does not exist\n", campaignID)
	} else {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println("Failed reading responde body from new campaign request: ", err)
		} else {
			json.Unmarshal(body, campaignData)
		}
		defer resp.Body.Close()
	}

	return campaignData, err

}

func checkCampaignExist(campaignID string) bool {
	if _, err := getCampaign(campaignID); err != nil {
		return false
	}

	return true
}

func convertDateStringToDate(dateString string) string {
	t, err := time.Parse(time.RFC3339, dateString)
	if err != nil {
		log.Printf("Failed parsing date string %s with error: %v", dateString, err)
	}
	return t.Format("2006-01-02")
}
