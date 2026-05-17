package provider

import "encoding/json"

// prRecord is one entry from `gh pr list --json headRefName,state`.
type prRecord struct {
	HeadRefName string `json:"headRefName"`
	State       string `json:"state"`
}

// parsePRStates decodes `gh pr list` JSON into a branch -> PR-state map.
func parsePRStates(data []byte) map[string]string {
	var records []prRecord
	if json.Unmarshal(data, &records) != nil {
		return map[string]string{}
	}
	states := make(map[string]string, len(records))
	for _, r := range records {
		states[r.HeadRefName] = r.State
	}
	return states
}

// loadPRStates runs one `gh pr list` for the whole repo (PLAN.md §2.3 — never
// N round-trips) and returns a branch -> PR-state map (OPEN, CLOSED, MERGED).
// gh being unavailable yields an empty map, not an error, so stack resolution
// degrades gracefully.
func loadPRStates(dir string) map[string]string {
	if !available("gh") {
		return map[string]string{}
	}
	res, err := run("gh", dir, nil,
		"pr", "list", "--state", "all", "--json", "headRefName,state", "--limit", "300")
	if err != nil {
		return map[string]string{}
	}
	return parsePRStates([]byte(res.stdout))
}
