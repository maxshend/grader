package submission_tasks

type SubmissionTask struct {
	WebhookURL   string            `json:"webhook_url"`
	Container    string            `json:"container"`
	PartID       string            `json:"part_id"`
	Files        []*SubmissionFile `json:"files"`
	SubmissionID int64             `json:"submission_id"`
	AccessToken  string            `json:"access_token"`
}

type SubmissionFile struct {
	URL  string `json:"url"`
	Name string `json:"name"`
}
