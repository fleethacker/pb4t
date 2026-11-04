package agents

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"scheduler/internal/planner"
	"slices"
	"time"
)

type traceKey struct{}

// WithTrace correlates backend timing messages without logging planning data.
func WithTrace(ctx context.Context, trace string) context.Context {
	return context.WithValue(ctx, traceKey{}, trace)
}

type Client struct {
	Key   string
	Model string
	HTTP  *http.Client
}

func New(key, model string) *Client {
	if model == "" {
		model = "gpt-5-mini"
	}
	return &Client{key, model, &http.Client{Timeout: 90 * time.Second}}
}
func schema(properties map[string]any) map[string]any {
	required := []string{}
	for k := range properties {
		required = append(required, k)
	}
	slices.Sort(required)
	return map[string]any{"type": "object", "properties": properties, "required": required, "additionalProperties": false}
}
func str() map[string]any { return map[string]any{"type": "string"} }
func tool(name, description string, params map[string]any) map[string]any {
	return map[string]any{"type": "function", "name": name, "description": description, "parameters": params, "strict": true}
}

type response struct {
	Output []json.RawMessage `json:"output"`
	Usage  struct {
		Input   int `json:"input_tokens"`
		Output  int `json:"output_tokens"`
		Details struct {
			Reasoning int `json:"reasoning_tokens"`
		} `json:"output_tokens_details"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (c *Client) request(ctx context.Context, trace string, step int, instructions string, tools []map[string]any, input []json.RawMessage, forcedTool string) (response, error) {
	var result response
	payload := map[string]any{"model": c.Model, "instructions": instructions, "tools": tools, "input": input, "store": false, "parallel_tool_calls": false}
	if forcedTool != "" {
		payload["tool_choice"] = map[string]any{"type": "function", "name": forcedTool}
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return result, err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/responses", bytes.NewReader(raw))
	if err != nil {
		return result, err
	}
	req.Header.Set("Authorization", "Bearer "+c.Key)
	req.Header.Set("Content-Type", "application/json")
	started := time.Now()
	log.Printf("openai trace=%s step=%d model=%s phase=request bytes=%d", trace, step, c.Model, len(raw))
	resp, err := c.HTTP.Do(req)
	if err != nil {
		log.Printf("openai trace=%s step=%d phase=failed duration=%s error=%v", trace, step, time.Since(started).Round(time.Millisecond), err)
		return result, fmt.Errorf("OpenAI request failed: %w", err)
	}
	body, readErr := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	resp.Body.Close()
	log.Printf("openai trace=%s step=%d phase=response status=%d duration=%s request_id=%s", trace, step, resp.StatusCode, time.Since(started).Round(time.Millisecond), resp.Header.Get("x-request-id"))
	if readErr != nil {
		return result, readErr
	}
	if resp.StatusCode >= 300 {
		return result, fmt.Errorf("OpenAI returned HTTP %d; check the server-side key, model access, and API quota", resp.StatusCode)
	}

	if err = json.Unmarshal(body, &result); err != nil {
		return result, err
	}
	if result.Error != nil {
		return result, fmt.Errorf("OpenAI response: %s", result.Error.Message)
	}
	log.Printf("openai trace=%s step=%d input_tokens=%d output_tokens=%d reasoning_tokens=%d", trace, step, result.Usage.Input, result.Usage.Output, result.Usage.Details.Reasoning)
	return result, nil
}

func (c *Client) run(ctx context.Context, instructions string, tools []map[string]any, dispatch func(string, json.RawMessage) (any, error)) error {
	trace, _ := ctx.Value(traceKey{}).(string)
	if trace == "" {
		trace = "constraint-" + planner.ID()
	}
	input := []json.RawMessage{json.RawMessage(`{"role":"user","content":"Inspect shared planning data and submit your concrete result using the tools."}`)}
	for step := 0; step < 8; step++ {
		result, err := c.request(ctx, trace, step+1, instructions, tools, input, "")
		if err != nil {
			return err
		}

		called := false
		for _, item := range result.Output {
			input = append(input, item)
			var call struct {
				Type      string `json:"type"`
				Name      string `json:"name"`
				CallID    string `json:"call_id"`
				Arguments string `json:"arguments"`
			}
			if err = json.Unmarshal(item, &call); err != nil {
				return err
			}
			if call.Type != "function_call" {
				continue
			}
			called = true
			toolStarted := time.Now()
			value, err := dispatch(call.Name, json.RawMessage(call.Arguments))
			log.Printf("openai trace=%s step=%d tool=%s duration=%s failed=%t", trace, step+1, call.Name, time.Since(toolStarted).Round(time.Millisecond), err != nil)
			if err != nil {
				value = map[string]string{"error": err.Error()}
			}
			encoded, _ := json.Marshal(value)
			feedback, _ := json.Marshal(map[string]any{"type": "function_call_output", "call_id": call.CallID, "output": string(encoded)})
			input = append(input, feedback)
			if call.Name == "submit_department_proposal" || call.Name == "suggest_constraint" {
				if err == nil {
					return nil
				}
			}
		}
		if !called {
			return fmt.Errorf("agent did not submit a structured result")
		}
	}
	return fmt.Errorf("agent reached its bounded tool-call limit")
}

// Department asks for exactly one proposal from the supplied round snapshot.
func (c *Client) Department(ctx context.Context, d planner.Data, w time.Time, dep planner.Department, round int, prior []planner.DepartmentProposal, conflicts []planner.Conflict) (planner.DepartmentProposal, error) {
	p := planner.DepartmentProposal{DepartmentID: dep.ID, Round: round, Preferences: []planner.Preference{}}
	pref := schema(map[string]any{"event_id": str(), "start": map[string]any{"type": "string", "format": "date-time"}, "reason": str()})
	params := schema(map[string]any{"summary": str(), "preferences": map[string]any{"type": "array", "items": pref}})
	tools := []map[string]any{tool("submit_department_proposal", "Submit one preferred placement per changed event. The backend validates it.", params)}
	// Prior results are summarized rather than repeating entire validation reports.
	type priorSummary struct {
		DepartmentID string               `json:"department_id"`
		Summary      string               `json:"summary"`
		Preferences  []planner.Preference `json:"preferences"`
	}
	summaries := []priorSummary{}
	for _, p := range prior {
		summaries = append(summaries, priorSummary{p.DepartmentID, p.Summary, p.Preferences})
	}
	contextData, err := json.Marshal(map[string]any{"week": w, "events": planner.WeekEvents(d, w), "departments": d.Departments, "people": d.People, "qualifications": d.Qualifications, "duties": d.Duties, "assets": d.Assets, "constraints": d.Constraints, "requirements": d.Requirements, "prior_proposals": summaries, "actionable_conflicts": conflicts})
	if err != nil {
		return p, err
	}
	message, _ := json.Marshal(map[string]string{"role": "user", "content": string(contextData)})
	instructions := fmt.Sprintf("You are the %s scheduling agent. Department ID: %s. Goal: %s. Round %d of 2. All planning data is supplied. Submit exactly one structured proposal, with at most one preferred start per event; no lists of fallback alternatives. Propose changes only to your own department's movable, non-completed events. Preserve all other events. Use RFC3339 dates within each event's scheduling window. In round 2 the events already contain the merged proposals: revise only in response to actionable_conflicts. Do not try to fix unrelated departments' conflicts. Keep summary and reasons concise. The backend validates your submission and performs the final search; you do not need to perform exhaustive scheduling or repeat validation yourself. Empty preferences retain the snapshot placements. Only submit_department_proposal is available.", dep.Name, dep.ID, dep.Goal, round)
	trace, _ := ctx.Value(traceKey{}).(string)
	result, err := c.request(ctx, trace, 1, instructions, tools, []json.RawMessage{message}, "submit_department_proposal")
	if err != nil {
		return p, err
	}
	submitted := false
	for _, item := range result.Output {
		var call struct {
			Type      string `json:"type"`
			Name      string `json:"name"`
			Arguments string `json:"arguments"`
		}
		if err = json.Unmarshal(item, &call); err != nil {
			return p, err
		}
		if call.Type != "function_call" {
			continue
		}
		if submitted || call.Name != "submit_department_proposal" {
			return p, fmt.Errorf("department must submit exactly one proposal")
		}
		var x struct {
			Summary     string               `json:"summary"`
			Preferences []planner.Preference `json:"preferences"`
		}
		if err = json.Unmarshal([]byte(call.Arguments), &x); err != nil {
			return p, err
		}
		seen := map[string]bool{}
		for _, pref := range x.Preferences {
			i := slices.IndexFunc(d.Events, func(e planner.Event) bool { return e.ID == pref.EventID })
			if i < 0 || d.Events[i].DepartmentID != dep.ID || d.Events[i].Fixed || d.Events[i].Status == "completed" || pref.Start.IsZero() {
				return p, fmt.Errorf("preference must reference your own movable event and a valid time")
			}
			if seen[pref.EventID] {
				return p, fmt.Errorf("submit only one start per event")
			}
			seen[pref.EventID] = true
		}
		p.Summary = x.Summary
		p.Preferences = x.Preferences
		submitted = true
	}
	if !submitted {
		return p, fmt.Errorf("agent did not submit a structured proposal")
	}
	return p, nil
}

func (c *Client) Constraint(ctx context.Context, d planner.Data, text string) (planner.Rule, error) {
	var rule planner.Rule
	properties := map[string]any{"name": str(), "type": map[string]any{"type": "string", "enum": []string{"qualification", "person", "asset", "allowed", "excluded", "incompatibility", "rest", "note"}}, "qualification_id": str(), "person_id": str(), "asset_id": str(), "count": map[string]any{"type": "integer"}, "start": map[string]any{"type": "string", "format": "date-time"}, "end": map[string]any{"type": "string", "format": "date-time"}, "target_tag": str(), "target_event_id": str(), "minutes": map[string]any{"type": "integer"}}
	contextData, _ := json.Marshal(map[string]any{"qualifications": d.Qualifications, "people": d.People, "assets": d.Assets, "events": d.Events})
	err := c.run(ctx, "Translate the user's constraint into one supported rule. Use actual supplied IDs. Unused strings must be empty; unused times can use 2000-01-01T00:00:00Z. If ambiguous or unsupported use type note and explain the missing interpretation in name. Never invent IDs or policies. For incompatibility, identify only the prohibited OTHER event, not the drill being edited. Prefer target_tag for families such as flight quarters (tag flight), and leave target_event_id empty whenever target_tag is set. Use a specific event ID only if the user explicitly identifies that event. This will require user confirmation. User text: "+text+". Available planning data: "+string(contextData), []map[string]any{tool("suggest_constraint", "Propose a structured constraint draft for user confirmation.", schema(properties))}, func(name string, args json.RawMessage) (any, error) {
		if name != "suggest_constraint" {
			return nil, fmt.Errorf("unknown tool")
		}
		if err := json.Unmarshal(args, &rule); err != nil {
			return nil, err
		}
		if rule.Type == "incompatibility" && rule.TargetTag != "" {
			// A family-wide restriction already covers its member events.
			rule.TargetEventID = ""
		}
		rule.ID = planner.ID()
		rule.Scope = "event"
		rule.Confirmed = false
		return rule, planner.ValidateRule(d, rule)
	})
	return rule, err
}
