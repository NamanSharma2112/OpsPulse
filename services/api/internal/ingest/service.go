// Package ingest turns verified GitHub deliveries into stored events and the
// projections the dashboard reads.
package ingest

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/NamanSharma2112/OpsPulse/services/api/internal/domain"
	"github.com/NamanSharma2112/OpsPulse/services/api/internal/github"
	"github.com/NamanSharma2112/OpsPulse/services/api/internal/store"
)

// Service records deliveries and keeps the projections up to date.
type Service struct {
	events       store.Events
	metrics      store.Metrics
	deployments  store.Deployments
	pullRequests store.PullRequests
	incidents    store.Incidents
	issues       store.Issues
	commits      store.Commits
	log          *slog.Logger
}

// NewService wires the ingest pipeline.
func NewService(
	events store.Events,
	metrics store.Metrics,
	deployments store.Deployments,
	pullRequests store.PullRequests,
	incidents store.Incidents,
	issues store.Issues,
	commits store.Commits,
	log *slog.Logger,
) *Service {
	return &Service{
		events:       events,
		metrics:      metrics,
		deployments:  deployments,
		pullRequests: pullRequests,
		incidents:    incidents,
		issues:       issues,
		commits:      commits,
		log:          log,
	}
}

// Result describes what a delivery did.
type Result struct {
	EventID   string `json:"event_id,omitempty"`
	Duplicate bool   `json:"duplicate"`
}

// Handle stores the delivery and applies it to the projections. Replayed
// deliveries are recognised by their delivery id and skipped.
func (s *Service) Handle(ctx context.Context, repo *domain.Repository, d github.Delivery, p *github.Payload) (Result, error) {
	occurred := occurredAt(p)
	event := &domain.Event{
		ProjectID:    repo.ProjectID,
		RepositoryID: &repo.ID,
		Source:       repo.Provider,
		DeliveryID:   d.ID,
		Type:         d.Event,
		Action:       p.Action,
		Actor:        p.Sender.Login,
		Payload:      json.RawMessage(d.Body),
		OccurredAt:   occurred,
	}

	fresh, err := s.events.Insert(ctx, event)
	if err != nil {
		return Result{}, fmt.Errorf("store event: %w", err)
	}
	if !fresh {
		s.log.Debug("duplicate delivery ignored", "delivery_id", d.ID, "repository_id", repo.ID)
		return Result{Duplicate: true}, nil
	}

	if err := s.project(ctx, repo, d.Event, p, occurred); err != nil {
		// The raw event is already safe in the database, so a projection
		// failure is logged rather than failing the delivery: GitHub would
		// otherwise retry an event we have already stored.
		s.log.Error("projection failed", "error", err, "type", d.Event, "repository_id", repo.ID)
	}
	return Result{EventID: event.ID, Duplicate: false}, nil
}

func (s *Service) project(ctx context.Context, repo *domain.Repository, eventType string, p *github.Payload, occurred time.Time) error {
	switch eventType {
	case "deployment", "deployment_status":
		return s.applyDeployment(ctx, repo, p, occurred)
	case "pull_request":
		return s.applyPullRequest(ctx, repo, p)
	case "workflow_run":
		return s.applyWorkflowRun(ctx, repo, p, occurred)
	case "issues":
		return s.applyIssue(ctx, repo, p, occurred)
	case "push":
		return s.applyPush(ctx, repo, p, occurred)
	default:
		return nil
	}
}

func (s *Service) applyDeployment(ctx context.Context, repo *domain.Repository, p *github.Payload, occurred time.Time) error {
	if p.Deployment == nil {
		return nil
	}
	d := domain.Deployment{
		ProjectID:    repo.ProjectID,
		RepositoryID: &repo.ID,
		ExternalID:   strconv.FormatInt(p.Deployment.ID, 10),
		Environment:  firstNonEmpty(p.Deployment.Environment, "production"),
		Ref:          p.Deployment.Ref,
		CommitSHA:    p.Deployment.SHA,
		Status:       domain.DeploymentPending,
		Actor:        firstNonEmpty(p.Deployment.Creator.Login, p.Sender.Login),
		StartedAt:    nonZeroTime(p.Deployment.CreatedAt, occurred),
	}

	metric := domain.MetricDeploymentStarted
	if st := p.DeploymentStatus; st != nil {
		d.Status = mapDeploymentState(st.State)
		d.URL = st.TargetURL
		if st.Environment != "" {
			d.Environment = st.Environment
		}
		if terminal(d.Status) {
			finished := nonZeroTime(st.CreatedAt, occurred)
			d.FinishedAt = &finished
		}
		switch d.Status {
		case domain.DeploymentSuccess:
			metric = domain.MetricDeploymentSucceeded
		case domain.DeploymentFailure:
			metric = domain.MetricDeploymentFailed
		default:
			metric = ""
		}
	}

	if err := s.deployments.Upsert(ctx, &d); err != nil {
		return err
	}
	if metric == "" {
		return nil
	}
	return s.record(ctx, repo.ProjectID, metric, 1, occurred, map[string]string{
		"environment": d.Environment,
		"ref":         d.Ref,
	})
}

func (s *Service) applyPullRequest(ctx context.Context, repo *domain.Repository, p *github.Payload) error {
	if p.PullRequest == nil {
		return nil
	}
	pr := p.PullRequest
	state := pr.State
	if pr.Merged {
		state = "merged"
	}
	record := domain.PullRequest{
		ProjectID:    repo.ProjectID,
		RepositoryID: &repo.ID,
		Number:       pr.Number,
		Title:        pr.Title,
		Author:       pr.User.Login,
		State:        state,
		Draft:        pr.Draft,
		URL:          pr.HTMLURL,
		OpenedAt:     pr.CreatedAt,
		MergedAt:     pr.MergedAt,
		ClosedAt:     pr.ClosedAt,
	}
	if err := s.pullRequests.Upsert(ctx, &record); err != nil {
		return err
	}

	labels := map[string]string{"author": pr.User.Login}
	switch {
	case p.Action == "opened":
		return s.record(ctx, repo.ProjectID, domain.MetricPullRequestOpened, 1, pr.CreatedAt, labels)
	case p.Action == "closed" && pr.Merged:
		at := time.Now().UTC()
		if pr.MergedAt != nil {
			at = *pr.MergedAt
		}
		// Lead time in hours is the headline delivery metric.
		labels["lead_time_hours"] = strconv.FormatFloat(at.Sub(pr.CreatedAt).Hours(), 'f', 2, 64)
		return s.record(ctx, repo.ProjectID, domain.MetricPullRequestMerged, 1, at, labels)
	}
	return nil
}

func (s *Service) applyWorkflowRun(ctx context.Context, repo *domain.Repository, p *github.Payload, occurred time.Time) error {
	run := p.WorkflowRun
	if run == nil || run.Status != "completed" {
		return nil
	}

	// Only runs on the default branch page the team: a failing feature branch
	// is the author's problem, a failing default branch is an incident.
	if run.HeadBranch != repo.DefaultBranch {
		return nil
	}
	externalID := workflowIncidentKey(run.WorkflowID, run.Name, run.HeadBranch)

	switch run.Conclusion {
	case "failure", "timed_out":
		incident := domain.Incident{
			ProjectID:  repo.ProjectID,
			ExternalID: externalID,
			Title:      fmt.Sprintf("%s failed on %s", firstNonEmpty(run.Name, "Workflow"), run.HeadBranch),
			Severity:   domain.SeverityMajor,
			Status:     domain.IncidentOpen,
			Source:     "workflow_run",
			URL:        run.HTMLURL,
			OpenedAt:   nonZeroTime(run.UpdatedAt, occurred),
		}
		if err := s.incidents.Upsert(ctx, &incident); err != nil {
			return err
		}
		return s.record(ctx, repo.ProjectID, domain.MetricIncidentOpened, 1, incident.OpenedAt, map[string]string{
			"source": "workflow_run",
			"branch": run.HeadBranch,
		})
	case "success":
		// A green run on the default branch clears the incident its earlier
		// failure opened.
		at := nonZeroTime(run.UpdatedAt, occurred)
		if err := s.incidents.Resolve(ctx, repo.ProjectID, externalID, at); err != nil {
			return err
		}
		return s.record(ctx, repo.ProjectID, domain.MetricIncidentResolved, 1, at, map[string]string{
			"source": "workflow_run",
			"branch": run.HeadBranch,
		})
	}
	return nil
}

// workflowIncidentKey identifies the incident a broken workflow owns.
//
// It deliberately keys on the workflow and the branch rather than on the run.
// A run id is unique to one execution, so an incident keyed by it could only
// ever be resolved by a re-run of that same execution: the usual recovery —
// CI fails on one commit and passes on the next — left the incident open
// forever, and every later break added another. Keyed this way, "CI is broken
// on main" is one incident that opens on the first failure and closes on the
// next success.
//
// workflow_id survives a workflow being renamed and is always present on a
// GitHub delivery; the name is a fallback for payloads that omit it.
func workflowIncidentKey(workflowID int64, name, branch string) string {
	workflow := strconv.FormatInt(workflowID, 10)
	if workflowID == 0 {
		workflow = firstNonEmpty(name, "workflow")
	}
	return "workflow_run:" + workflow + ":" + branch
}

func (s *Service) applyIssue(ctx context.Context, repo *domain.Repository, p *github.Payload, occurred time.Time) error {
	if p.Issue == nil {
		return nil
	}

	// Every issue is projected, whatever its labels: the open-issue count is
	// a delivery signal on its own, separate from whether the issue is also
	// an incident.
	if err := s.projectIssue(ctx, repo, p, occurred); err != nil {
		return err
	}

	severity, isIncident := incidentSeverity(p.Labels())
	if !isIncident {
		return nil
	}
	externalID := "issue:" + strconv.Itoa(p.Issue.Number)

	if p.Issue.State == "closed" {
		at := occurred
		if p.Issue.ClosedAt != nil {
			at = *p.Issue.ClosedAt
		}
		if err := s.incidents.Resolve(ctx, repo.ProjectID, externalID, at); err != nil {
			return err
		}
		return s.record(ctx, repo.ProjectID, domain.MetricIncidentResolved, 1, at, map[string]string{"source": "issue"})
	}

	incident := domain.Incident{
		ProjectID:  repo.ProjectID,
		ExternalID: externalID,
		Title:      p.Issue.Title,
		Severity:   severity,
		Status:     domain.IncidentOpen,
		Source:     "issue",
		URL:        p.Issue.HTMLURL,
		OpenedAt:   nonZeroTime(p.Issue.CreatedAt, occurred),
	}
	if err := s.incidents.Upsert(ctx, &incident); err != nil {
		return err
	}
	if p.Action != "opened" && p.Action != "labeled" {
		return nil
	}
	return s.record(ctx, repo.ProjectID, domain.MetricIncidentOpened, 1, incident.OpenedAt, map[string]string{
		"source":   "issue",
		"severity": severity,
	})
}

// projectIssue keeps the issue projection in step with GitHub.
func (s *Service) projectIssue(ctx context.Context, repo *domain.Repository, p *github.Payload, occurred time.Time) error {
	state := domain.IssueOpen
	var closedAt *time.Time
	if p.Issue.State == "closed" {
		state = domain.IssueClosed
		at := occurred
		if p.Issue.ClosedAt != nil {
			at = *p.Issue.ClosedAt
		}
		closedAt = &at
	}

	issue := domain.Issue{
		ProjectID:    repo.ProjectID,
		RepositoryID: &repo.ID,
		Number:       p.Issue.Number,
		Title:        p.Issue.Title,
		Author:       p.Issue.User.Login,
		State:        state,
		URL:          p.Issue.HTMLURL,
		OpenedAt:     nonZeroTime(p.Issue.CreatedAt, occurred),
		ClosedAt:     closedAt,
	}
	if err := s.issues.Upsert(ctx, &issue); err != nil {
		return err
	}

	// Only the transitions are counted. "edited" and "labeled" repeat the
	// issue's current state and would otherwise inflate both series.
	switch p.Action {
	case "opened", "reopened":
		return s.record(ctx, repo.ProjectID, domain.MetricIssueOpened, 1, issue.OpenedAt,
			map[string]string{"author": issue.Author})
	case "closed":
		at := occurred
		if closedAt != nil {
			at = *closedAt
		}
		return s.record(ctx, repo.ProjectID, domain.MetricIssueClosed, 1, at,
			map[string]string{"author": issue.Author})
	}
	return nil
}

// applyPush projects the commits in a push.
//
// Tag pushes carry no branch and are skipped: a tag is not development
// activity, and counting one would make a release look like a day's work.
func (s *Service) applyPush(ctx context.Context, repo *domain.Repository, p *github.Payload, occurred time.Time) error {
	branch := p.Branch()
	if branch == "" || len(p.Commits) == 0 {
		return nil
	}

	commits := make([]domain.Commit, 0, len(p.Commits))
	for _, c := range p.Commits {
		if c.ID == "" {
			continue
		}
		commits = append(commits, domain.Commit{
			ProjectID:    repo.ProjectID,
			RepositoryID: &repo.ID,
			SHA:          c.ID,
			Message:      firstLine(c.Message),
			Author:       firstNonEmpty(c.Author.Username, c.Author.Name),
			Branch:       branch,
			CommittedAt:  nonZeroTime(c.Timestamp, occurred),
		})
	}

	inserted, err := s.commits.InsertBatch(ctx, commits)
	if err != nil {
		return err
	}
	// Commits already seen on another branch are not new work, so the metric
	// counts what was actually inserted rather than what was delivered.
	if inserted == 0 {
		return nil
	}
	return s.record(ctx, repo.ProjectID, domain.MetricCommitPushed, float64(inserted), occurred,
		map[string]string{"branch": branch, "actor": p.Sender.Login})
}

// firstLine is a commit's subject: everything before the first newline,
// bounded so a pathological message cannot bloat a row.
func firstLine(message string) string {
	if i := strings.IndexByte(message, '\n'); i >= 0 {
		message = message[:i]
	}
	message = strings.TrimSpace(message)
	const max = 200
	if len(message) > max {
		return message[:max]
	}
	return message
}

func (s *Service) record(ctx context.Context, projectID, name string, value float64, at time.Time, labels map[string]string) error {
	m := domain.Metric{ProjectID: projectID, Name: name, Value: value, Labels: labels, RecordedAt: at}
	return s.metrics.Insert(ctx, &m)
}

// incidentSeverity reads the incident labels a repository applies to issues.
// Anything labelled "incident" counts; "severity:critical" style labels and
// the common sev1/sev2 shorthand set the severity.
func incidentSeverity(labels []string) (string, bool) {
	severity, found := domain.SeverityMinor, false
	for _, raw := range labels {
		l := strings.ToLower(strings.TrimSpace(raw))
		switch {
		case l == "incident", l == "outage":
			found = true
		case strings.HasSuffix(l, "critical"), l == "sev1":
			severity, found = domain.SeverityCritical, true
		case strings.HasSuffix(l, "major"), l == "sev2":
			severity, found = domain.SeverityMajor, true
		case strings.HasSuffix(l, "minor"), l == "sev3":
			severity, found = domain.SeverityMinor, true
		}
	}
	return severity, found
}

func mapDeploymentState(state string) string {
	switch strings.ToLower(state) {
	case "success":
		return domain.DeploymentSuccess
	case "failure", "error":
		return domain.DeploymentFailure
	case "in_progress", "queued":
		return domain.DeploymentRunning
	case "inactive":
		return domain.DeploymentInactive
	default:
		return domain.DeploymentPending
	}
}

func terminal(status string) bool {
	return status == domain.DeploymentSuccess || status == domain.DeploymentFailure
}

// occurredAt picks the most specific timestamp in the payload, falling back to
// receipt time for events that carry none.
func occurredAt(p *github.Payload) time.Time {
	switch {
	case p.DeploymentStatus != nil && !p.DeploymentStatus.CreatedAt.IsZero():
		return p.DeploymentStatus.CreatedAt
	case p.WorkflowRun != nil && !p.WorkflowRun.UpdatedAt.IsZero():
		return p.WorkflowRun.UpdatedAt
	case p.PullRequest != nil && !p.PullRequest.UpdatedAt.IsZero():
		return p.PullRequest.UpdatedAt
	case p.Deployment != nil && !p.Deployment.CreatedAt.IsZero():
		return p.Deployment.CreatedAt
	case p.HeadCommit != nil && !p.HeadCommit.Timestamp.IsZero():
		return p.HeadCommit.Timestamp
	default:
		return time.Now().UTC()
	}
}

func nonZeroTime(t, fallback time.Time) time.Time {
	if t.IsZero() {
		return fallback
	}
	return t
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
