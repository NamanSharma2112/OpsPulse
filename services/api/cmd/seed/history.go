package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"
)

// delivery is one webhook request waiting to be signed and sent.
type delivery struct {
	id    string
	event string
	body  []byte
}

// generator produces a repository's history. Everything it emits is derived
// from its own rand, so a given -seed always yields the same history: a
// screenshot or a bug report made from seeded data can be reproduced.
type generator struct {
	rnd    *rand.Rand
	owner  string
	name   string
	branch string
	days   int

	n         int // delivery counter, for unique delivery ids
	prNumber  int
	issueNo   int
	deployID  int64
	workflow  int64
	out       []delivery
	openIssue []int // incident issues still open, closed later in the history
}

var (
	authors  = []string{"ada", "grace", "linus", "margaret", "ken"}
	features = []string{
		"Cache the project summary query",
		"Retry webhook delivery on 5xx",
		"Tighten the deployment status mapping",
		"Add an index on events(project_id, occurred_at)",
		"Split the health score into pillars",
		"Stop counting tag pushes as commits",
		"Paginate the event feed",
		"Seal GitHub tokens with a per-install key",
		"Drop the unused metrics view",
		"Make the dashboard render partial data",
	}
	chores = []string{
		"Bump pgx to 5.7.6",
		"Tidy go.mod",
		"Fix a typo in the API reference",
		"Format with gofmt",
		"Extend the ingest tests",
	}
)

// history returns a full set of deliveries, oldest first.
//
// Ordering matters: the API records each delivery as it arrives, and an
// incident must be opened before the run that clears it, so the generator
// walks day by day rather than emitting each event type in its own pass.
func (g *generator) history() []delivery {
	start := time.Now().UTC().AddDate(0, 0, -g.days).Truncate(24 * time.Hour)

	for day := 0; day < g.days; day++ {
		date := start.AddDate(0, 0, day)
		weekend := date.Weekday() == time.Saturday || date.Weekday() == time.Sunday

		// A quiet weekend is what a real repository looks like, and it is
		// what gives the anomaly detector a baseline with some shape to it.
		commits := g.between(2, 7)
		if weekend {
			commits = g.between(0, 2)
		}
		if commits > 0 {
			g.push(date.Add(10*time.Hour), commits)
		}

		for range g.pullRequestsFor(weekend) {
			g.pullRequest(date)
		}

		// Roughly every other weekday ships something.
		if !weekend && g.rnd.Float64() < 0.55 {
			g.deployment(date.Add(16 * time.Hour))
		}

		if !weekend {
			g.workflowRun(date.Add(11 * time.Hour))
		}

		if g.rnd.Float64() < 0.12 {
			g.issue(date.Add(13 * time.Hour))
		}
		g.closeAnOldIssue(date.Add(15 * time.Hour))
	}

	return g.out
}

func (g *generator) pullRequestsFor(weekend bool) []struct{} {
	n := g.between(0, 3)
	if weekend {
		n = 0
	}
	return make([]struct{}, n)
}

// push emits one push with n commits on the default branch.
func (g *generator) push(at time.Time, n int) {
	type commit struct {
		ID        string    `json:"id"`
		Message   string    `json:"message"`
		Timestamp time.Time `json:"timestamp"`
		Author    struct {
			Name     string `json:"name"`
			Username string `json:"username"`
		} `json:"author"`
	}

	commits := make([]commit, 0, n)
	for i := 0; i < n; i++ {
		var c commit
		c.ID = g.sha()
		c.Message = g.pick(append(append([]string{}, features...), chores...))
		c.Timestamp = at.Add(time.Duration(i) * 37 * time.Minute)
		c.Author.Username = g.pick(authors)
		c.Author.Name = c.Author.Username
		commits = append(commits, c)
	}

	g.emit("push", map[string]any{
		"ref":         "refs/heads/" + g.branch,
		"repository":  g.repository(),
		"sender":      g.sender(commits[0].Author.Username),
		"head_commit": commits[len(commits)-1],
		"commits":     commits,
	})
}

// pullRequest opens a pull request and, usually, merges it — emitting both
// deliveries, because lead time is measured between them.
func (g *generator) pullRequest(date time.Time) {
	g.prNumber++
	author := g.pick(authors)
	opened := date.Add(time.Duration(g.between(9, 17)) * time.Hour)

	base := map[string]any{
		"number":     g.prNumber,
		"title":      g.pick(features),
		"state":      "open",
		"draft":      false,
		"merged":     false,
		"html_url":   fmt.Sprintf("https://github.com/%s/%s/pull/%d", g.owner, g.name, g.prNumber),
		"created_at": opened,
		"updated_at": opened,
		"user":       map[string]string{"login": author},
	}
	g.emit("pull_request", map[string]any{
		"action":       "opened",
		"repository":   g.repository(),
		"sender":       g.sender(author),
		"pull_request": base,
	})

	// Most pull requests merge; the lead time is skewed, with a long tail.
	if g.rnd.Float64() < 0.12 {
		return
	}
	lead := time.Duration(g.between(2, 30)) * time.Hour
	if g.rnd.Float64() < 0.2 {
		lead += time.Duration(g.between(24, 96)) * time.Hour
	}
	merged := opened.Add(lead)
	if merged.After(time.Now().UTC()) {
		return // still open at the end of the window
	}

	closed := map[string]any{}
	for k, v := range base {
		closed[k] = v
	}
	closed["state"] = "closed"
	closed["merged"] = true
	closed["merged_at"] = merged
	closed["closed_at"] = merged
	closed["updated_at"] = merged

	g.emit("pull_request", map[string]any{
		"action":       "closed",
		"repository":   g.repository(),
		"sender":       g.sender(author),
		"pull_request": closed,
	})
}

// deployment emits a deployment and its terminal status.
func (g *generator) deployment(at time.Time) {
	g.deployID++
	actor := g.pick(authors)
	sha := g.sha()

	deployment := map[string]any{
		"id":          g.deployID,
		"ref":         g.branch,
		"sha":         sha,
		"environment": "production",
		"created_at":  at,
		"creator":     map[string]string{"login": actor},
	}
	g.emit("deployment", map[string]any{
		"action":     "created",
		"repository": g.repository(),
		"sender":     g.sender(actor),
		"deployment": deployment,
	})

	state := "success"
	if g.rnd.Float64() < 0.15 {
		state = "failure"
	}
	finished := at.Add(time.Duration(g.between(3, 14)) * time.Minute)
	g.emit("deployment_status", map[string]any{
		"action":     "created",
		"repository": g.repository(),
		"sender":     g.sender(actor),
		"deployment": deployment,
		"deployment_status": map[string]any{
			"state":       state,
			"environment": "production",
			"target_url":  fmt.Sprintf("https://github.com/%s/%s/deployments", g.owner, g.name),
			"created_at":  finished,
		},
	})
}

// workflowRun emits a completed CI run on the default branch. A failure opens
// an incident; the next run's success resolves it, which is what gives the
// dashboard a mean-time-to-resolve worth showing.
func (g *generator) workflowRun(at time.Time) {
	g.workflow++
	conclusion := "success"
	if g.rnd.Float64() < 0.1 {
		conclusion = "failure"
	}
	g.emit("workflow_run", map[string]any{
		"action":     "completed",
		"repository": g.repository(),
		"sender":     g.sender(g.pick(authors)),
		"workflow_run": map[string]any{
			"id":          g.workflow,
			"workflow_id": 1,
			"name":        "CI",
			"head_branch": g.branch,
			"head_sha":    g.sha(),
			"status":      "completed",
			"conclusion":  conclusion,
			"html_url":    fmt.Sprintf("https://github.com/%s/%s/actions", g.owner, g.name),
			"created_at":  at,
			"updated_at":  at.Add(time.Duration(g.between(4, 20)) * time.Minute),
		},
	})
}

// issue opens an issue, sometimes labelled as an incident.
func (g *generator) issue(at time.Time) {
	g.issueNo++
	author := g.pick(authors)

	labels := []map[string]string{{"name": "bug"}}
	if g.rnd.Float64() < 0.5 {
		labels = []map[string]string{
			{"name": "incident"},
			{"name": g.pick([]string{"severity:minor", "severity:major", "severity:critical"})},
		}
	}

	g.emit("issues", map[string]any{
		"action":     "opened",
		"repository": g.repository(),
		"sender":     g.sender(author),
		"issue": map[string]any{
			"number":     g.issueNo,
			"title":      g.pick([]string{"Checkout times out under load", "Duplicate webhook deliveries", "Stale health score after a deploy", "Session cookie dropped on Safari"}),
			"state":      "open",
			"html_url":   fmt.Sprintf("https://github.com/%s/%s/issues/%d", g.owner, g.name, g.issueNo),
			"created_at": at,
			"labels":     labels,
			"user":       map[string]string{"login": author},
		},
	})
	g.openIssue = append(g.openIssue, g.issueNo)
}

// closeAnOldIssue closes the oldest open issue once in a while, so the
// dashboard shows both open and resolved work rather than only a backlog.
func (g *generator) closeAnOldIssue(at time.Time) {
	if len(g.openIssue) == 0 || g.rnd.Float64() > 0.4 {
		return
	}
	number := g.openIssue[0]
	g.openIssue = g.openIssue[1:]
	author := g.pick(authors)

	g.emit("issues", map[string]any{
		"action":     "closed",
		"repository": g.repository(),
		"sender":     g.sender(author),
		"issue": map[string]any{
			"number":     number,
			"title":      "Resolved",
			"state":      "closed",
			"html_url":   fmt.Sprintf("https://github.com/%s/%s/issues/%d", g.owner, g.name, number),
			"created_at": at.AddDate(0, 0, -2),
			"closed_at":  at,
			"labels":     []map[string]string{{"name": "incident"}},
			"user":       map[string]string{"login": author},
		},
	})
}

func (g *generator) emit(event string, payload map[string]any) {
	body, err := json.Marshal(payload)
	if err != nil {
		// Every payload here is built from plain maps and marshalable types,
		// so this cannot fail without a programming error.
		panic(fmt.Sprintf("seed: encode %s: %v", event, err))
	}
	g.n++
	g.out = append(g.out, delivery{
		id:    fmt.Sprintf("seed-%s-%04d", event, g.n),
		event: event,
		body:  body,
	})
}

func (g *generator) repository() map[string]any {
	return map[string]any{
		"name":           g.name,
		"full_name":      g.owner + "/" + g.name,
		"default_branch": g.branch,
		"owner":          map[string]string{"login": g.owner},
	}
}

func (g *generator) sender(login string) map[string]string {
	return map[string]string{"login": login}
}

// between returns an inclusive random integer in [low, high].
func (g *generator) between(low, high int) int {
	if high <= low {
		return low
	}
	return low + g.rnd.Intn(high-low+1)
}

func (g *generator) pick(options []string) string {
	return options[g.rnd.Intn(len(options))]
}

func (g *generator) sha() string {
	const hexdigits = "0123456789abcdef"
	out := make([]byte, 40)
	for i := range out {
		out[i] = hexdigits[g.rnd.Intn(len(hexdigits))]
	}
	return string(out)
}
