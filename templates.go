package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type ListOfTemplates struct {
	Templates []TemplateResponse `json:"templates"`
}

type TemplateResponse struct {
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

func GetTemplates() *ListOfTemplates {
	var requestURL = baseURL + "/templates"
	client := &http.Client{}
	request := getHTTPRequest("GET", requestURL, nil)

	resp, err := client.Do(request)
	if err != nil {
		log.Println("GET Request Failed, GetTemplates error: ", err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Failed reading responde body from templates request: ", err)
	}
	defer resp.Body.Close()

	var templates = new(ListOfTemplates)
	json.Unmarshal(body, templates)

	return templates
}

// GetTemplateFromYesterday returns the template id that created yesterday with the name ZONE
func GetTemplateFromYesterday() uint {
	var yesterdayDate = time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	const templateName = "testing"

	var templateID uint

	listOfTemplates := GetTemplates()
	for _, template := range listOfTemplates.Templates {
		createdDate := convertDateStringToDate(template.DateCreated)
		if createdDate == yesterdayDate && template.Name == templateName {
			templateID = template.ID
			break
		}
	}

	return templateID
}
