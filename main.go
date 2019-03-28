/*
	A golang snippet that helps the team to send out mailchimp email updates based on tags
*/
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type listOfMembers struct {
	TotalItems int      `json:"total_items"`
	Members    []member `json:"members"`
}

type memberStats struct {
	AvgOpenRate  float64 `json:"avg_open_rate"`
	AvgClickRate float64 `json:"avg_click_rate"`
}

type member struct {
	ID              string      `json:"id"`
	ListID          string      `json:"list_id"`
	EmailAddress    string      `json:"email_address"`
	UniqueEMailID   string      `json:"unique_email_id"`
	EmailType       string      `json:"email_type"`
	Stats           memberStats `json:"stats"`
	IPSignup        string      `json:"ip_signup"`
	TimestampSignup string      `json:"timestamp_signup"`
	TimestampOpt    string      `json:"timestamp_opt"`
	MemberRating    int         `json:"member_rating"`
	LastChanged     string      `json:"last_changed"`
	EmailClient     string      `json:"email_client"`
}

type segmentBatchRequest struct {
	MembersToAdd    []string `json:"members_to_add"`
	MembersToRemove []string `json:"members_to_remove"`
}

type segmentBatchResponse struct {
	MembersAdded   []member `json:"members_added"`
	MembersRemoved []member `json:"members_removed"`

	TotalAdded   int `json:"total_added"`
	TotalRemoved int `json:"total_removed"`
	ErrorCount   int `json:"error_count"`
}

type listOfTemplates struct {
	Templates []templateResponse `json:"templates"`
}

type templateResponse struct {
	ID          uint   `json:"id"`
	Type        string `json:"type"`
	Name        string `json:"name"`
	DragAndDrop bool   `json:"drag_and_drop"`
	Responsive  bool   `json:"responsive"`
	Category    string `json:"category"`
	DateCreated string `json:"date_created"`
	CreatedBy   string `json:"created_by"`
	Active      bool   `json:"activer"`
	FolderId    string `json:"folder_id"`
	Thumbnail   string `json:"thumbnail"`
	ShareUrl    string `json:"share_url"`
}

type campaignCreationRequest struct {
	Type       string                     `json:"type"` // must be one of the CAMPAIGN_TYPE_* consts
	Recipients campaignCreationRecipients `json:"recipients"`
	Settings   campaignCreationSettings   `json:"settings"`
}

type campaignCreationSettings struct {
	SubjectLine string `json:"subject_line"`
	Title       string `json:"title"`
	FromName    string `json:"from_name"`
	ReplyTo     string `json:"reply_to"`
	TemplateID  uint   `json:"template_id"`
}

type campaignCreationRecipients struct {
	ListID         string                         `json:"list_id"`
	SegmentOptions campaignCreationSegmentOptions `json:"segment_opts"`
}

type campaignCreationSegmentOptions struct {
	SavedSegmentID int    `json:"saved_segment_id"`
	Match          string `json:"match"` // one of CONDITION_MATCH_*

	Conditions []campaignCreationConditions `json:"conditions"`
}

type campaignCreationConditions struct {
	ConditionType string `json:"condition_type"`
	Field         string `json:"field"`
	Op            string `json:"op"`
	Value         int    `json:"value"`
}

type campaignResponse struct {
	ID         string `json:"id"`
	WebID      uint   `json:"web_id"`
	Type       string `json:"type"`
	CreateTime string `json:"create_time"`
	Status     string `json:"status"`
	EmailsSent uint   `json:"emails_sent"`
	SendTime   string `json:"send_time"`
}

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

	remainingMembers, err := getMembers(remainingSegmentID)
	if err != nil || len(remainingMembers.Members) < 1 {
		log.Panicln("Failed to get remaining members")
	}

	if _, err := setTags(queuedID, remainingMembers.Members, "add"); err != nil {
		log.Panicln("queue tag failed to be set")
	}

	templateID := getTemplateFromYesterday()
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

	if _, err := setTags(sentID, remainingMembers.Members, "add"); err != nil {
		log.Panicln("Sent tag failed to be set")
	}

	if _, err := setTags(queuedID, remainingMembers.Members, "remove"); err != nil {
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

func getMembers(segmentID string) (*listOfMembers, error) {
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

	var response = new(listOfMembers)
	json.Unmarshal(data, response)

	return response, err
}

func setTags(segmentID string, members []member, method string) (*segmentBatchResponse, error) {
	actionURL := fmt.Sprintf("/lists/%s/segments/%s", listID, segmentID)
	requestURL := baseURL + actionURL

	requetPayload, err := getSegmentPayload(method, segmentID, members)
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

	response := new(segmentBatchResponse)
	json.Unmarshal(body, response)
	return response, err
}

func getSegmentPayload(method string, segmentID string, data []member) (segmentBatchRequest, error) {
	var payload []string
	for _, member := range data {
		payload = append(payload, member.EmailAddress)
	}
	switch method {
	case "add":
		return segmentBatchRequest{
			MembersToAdd:    payload,
			MembersToRemove: []string{},
		}, nil
	case "remove":
		return segmentBatchRequest{
			MembersToAdd:    []string{},
			MembersToRemove: payload,
		}, nil
	default:
		return segmentBatchRequest{}, errors.New("Failed to create request payload")
	}
}

func getTemplates() *listOfTemplates {
	var requestURL = baseURL + "/templates"
	client := &http.Client{}
	request := getHTTPRequest("GET", requestURL, nil)

	resp, err := client.Do(request)
	if err != nil {
		log.Println("GET Request Failed, getTemplates error: ", err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Failed reading responde body from templates request: ", err)
	}
	defer resp.Body.Close()

	var templates = new(listOfTemplates)
	json.Unmarshal(body, templates)

	return templates
}

// getTemplateFromYesterday returns the template id that created yesterday with the name ZONE
func getTemplateFromYesterday() uint {
	var yesterdayDate = time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	const templateName = "testing"

	var templateID uint

	listOfTemplates := getTemplates()
	for _, template := range listOfTemplates.Templates {
		createdDate := convertDateStringToDate(template.DateCreated)
		if createdDate == yesterdayDate && template.Name == templateName {
			templateID = template.ID
			break
		}
	}

	return templateID
}

func createNewCampaign(templateID uint) (*campaignResponse, error) {
	var requestURL = baseURL + "/campaigns"
	segmentID, _ := strconv.Atoi(queuedID)

	//TODO: Create command-line input function to allow dynamically setting campaign settings
	requestPayload := campaignCreationRequest{
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
		Settings: campaignCreationSettings{
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
