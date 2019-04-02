package main

type CampaignCreationRequest struct {
	Type       string                     `json:"type"` // must be one of the CAMPAIGN_TYPE_* consts
	Recipients campaignCreationRecipients `json:"recipients"`
	Settings   CampaignCreationSettings   `json:"settings"`
}

type CampaignCreationSettings struct {
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
