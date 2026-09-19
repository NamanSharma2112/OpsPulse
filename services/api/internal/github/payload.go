package github

import (
	"encoding/json"
	"time"
)

// Delivery is the subset of a webhook request the API needs to route and
// record an event.
type Delivery struct {
	ID    string // X-GitHub-Delivery
	Event string // X-GitHub-Event
	Body  []byte
}

// Payload holds the fields OpsPulse reads across every event type. Anything
// not modelled here stays available in the stored raw JSON.
type Payload struct {
	Action     string `json:"action"`
	Repository struct {
		Name          string `json:"name"`
		FullName      string `json:"full_name"`
		DefaultBranch string `json:"default_branch"`
		Owner         struct {
			Login string `json:"login"`
		} `json:"owner"`
	} `json:"repository"`
	Sender struct {
		Login string `json:"login"`
	} `json:"sender"`

	Deployment *struct {
		ID          int64     `json:"id"`
		Ref         string    `json:"ref"`
		SHA         string    `json:"sha"`
		Environment string    `json:"environment"`
		CreatedAt   time.Time `json:"created_at"`
		Creator     struct {
			Login string `json:"login"`
		} `json:"creator"`
	} `json:"deployment"`

	DeploymentStatus *struct {
		State       string    `json:"state"`
		Environment string    `json:"environment"`
		TargetURL   string    `json:"target_url"`
		CreatedAt   time.Time `json:"created_at"`
	} `json:"deployment_status"`

	PullRequest *struct {
		Number    int        `json:"number"`
		Title     string     `json:"title"`
		State     string     `json:"state"`
		Draft     bool       `json:"draft"`
		Merged    bool       `json:"merged"`
		HTMLURL   string     `json:"html_url"`
		CreatedAt time.Time  `json:"created_at"`
		UpdatedAt time.Time  `json:"updated_at"`
		MergedAt  *time.Time `json:"merged_at"`
		ClosedAt  *time.Time `json:"closed_at"`
		User      struct {
			Login string `json:"login"`
		} `json:"user"`
	} `json:"pull_request"`

	Issue *struct {
		Number    int        `json:"number"`
		Title     string     `json:"title"`
		State     string     `json:"state"`
		HTMLURL   string     `json:"html_url"`
		CreatedAt time.Time  `json:"created_at"`
		ClosedAt  *time.Time `json:"closed_at"`
		Labels    []struct {
			Name string `json:"name"`
		} `json:"labels"`
		User struct {
			Login string `json:"login"`
		} `json:"user"`
	} `json:"issue"`

	WorkflowRun *struct {
		ID         int64     `json:"id"`
		Name       string    `json:"name"`
		HeadBranch string    `json:"head_branch"`
		HeadSHA    string    `json:"head_sha"`
		Status     string    `json:"status"`
		Conclusion string    `json:"conclusion"`
		HTMLURL    string    `json:"html_url"`
		CreatedAt  time.Time `json:"created_at"`
		UpdatedAt  time.Time `json:"updated_at"`
	} `json:"workflow_run"`

	Ref        string `json:"ref"`
	HeadCommit *struct {
		ID        string    `json:"id"`
		Message   string    `json:"message"`
		Timestamp time.Time `json:"timestamp"`
	} `json:"head_commit"`
}

// ParsePayload decodes the fields OpsPulse understands.
func ParsePayload(body []byte) (*Payload, error) {
	var p Payload
	if err := json.Unmarshal(body, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// Labels flattens issue label names.
func (p *Payload) Labels() []string {
	if p.Issue == nil {
		return nil
	}
	out := make([]string, 0, len(p.Issue.Labels))
	for _, l := range p.Issue.Labels {
		out = append(out, l.Name)
	}
	return out
}
