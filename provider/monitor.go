package provider

import (
	"context"
	"net/http"
	"strconv"

	"github.com/pulumi/pulumi-go-provider/infer"
)

// Monitor is an Oh Dear monitor (https://ohdear.app/docs/api/monitors).
//
// ponytail: first cut covers the monitor itself plus the simple top-level knobs.
// The per-check *_check_settings trees (uptime, ports, dns, lighthouse, ...) are
// not modelled yet — add them when a stack actually needs to tune a check.
type Monitor struct{}

type MonitorArgs struct {
	URL           string   `pulumi:"url"`
	TeamID        int      `pulumi:"teamId" provider:"replaceOnChanges"`
	Type          string   `pulumi:"type,optional" provider:"replaceOnChanges"`
	Checks        []string `pulumi:"checks,optional"`
	FriendlyName  string   `pulumi:"friendlyName,optional"`
	GroupName     string   `pulumi:"groupName,optional"`
	Tags          []string `pulumi:"tags,optional"`
	Notes         string   `pulumi:"notes,optional"`
	Description   string   `pulumi:"description,optional"`
	RealIPAddress string   `pulumi:"realIpAddress,optional"`
}

type MonitorState struct {
	MonitorArgs

	MonitorID             int    `pulumi:"monitorId"`
	Label                 string `pulumi:"label"`
	UsesHTTPS             bool   `pulumi:"usesHttps"`
	SortURL               string `pulumi:"sortUrl"`
	SummarizedCheckResult string `pulumi:"summarizedCheckResult"`
	CreatedAt             string `pulumi:"createdAt"`
	UpdatedAt             string `pulumi:"updatedAt"`
}

func (m *Monitor) Annotate(a infer.Annotator) {
	a.Describe(&m, "An Oh Dear monitor (http, ping, tcp or ai).")
}

func (a *MonitorArgs) Annotate(an infer.Annotator) {
	an.Describe(&a.Type, `Monitor type: "http", "ping", "tcp" or "ai".`)
	an.SetDefault(&a.Type, "http")
}

type monitorWire struct {
	ID                    int      `json:"id"`
	TeamID                int      `json:"team_id"`
	Type                  string   `json:"type"`
	URL                   string   `json:"url"`
	UsesHTTPS             bool     `json:"uses_https"`
	SortURL               string   `json:"sort_url"`
	Label                 string   `json:"label"`
	GroupName             string   `json:"group_name"`
	Tags                  []string `json:"tags"`
	Notes                 string   `json:"notes"`
	Description           string   `json:"description"`
	RealIPAddress         string   `json:"real_ip_address"`
	SummarizedCheckResult string   `json:"summarized_check_result"`
	CreatedAt             string   `json:"created_at"`
	UpdatedAt             string   `json:"updated_at"`
}

// toState refreshes server-owned fields from a response while preserving the
// inputs that the API never echoes back (friendlyName, checks).
func (w monitorWire) toState(inputs MonitorArgs) MonitorState {
	inputs.URL = w.URL
	inputs.TeamID = w.TeamID
	inputs.Type = w.Type
	inputs.GroupName = w.GroupName
	inputs.Tags = w.Tags
	inputs.Notes = w.Notes
	inputs.Description = w.Description
	inputs.RealIPAddress = w.RealIPAddress
	return MonitorState{
		MonitorArgs:           inputs,
		MonitorID:             w.ID,
		Label:                 w.Label,
		UsesHTTPS:             w.UsesHTTPS,
		SortURL:               w.SortURL,
		SummarizedCheckResult: w.SummarizedCheckResult,
		CreatedAt:             w.CreatedAt,
		UpdatedAt:             w.UpdatedAt,
	}
}

// createBody omits empty optionals so Oh Dear applies its own defaults.
type monitorCreateBody struct {
	URL           string   `json:"url"`
	TeamID        int      `json:"team_id"`
	Type          string   `json:"type"`
	Checks        []string `json:"checks,omitempty"`
	FriendlyName  string   `json:"friendly_name,omitempty"`
	GroupName     string   `json:"group_name,omitempty"`
	Tags          []string `json:"tags,omitempty"`
	Notes         string   `json:"notes,omitempty"`
	Description   string   `json:"description,omitempty"`
	RealIPAddress string   `json:"real_ip_address,omitempty"`
}

// updateBody sends every managed field (no omitempty) so clearing an input in
// code clears it in Oh Dear. team_id and type are immutable and never sent.
type monitorUpdateBody struct {
	URL           string   `json:"url"`
	Checks        []string `json:"checks,omitempty"`
	FriendlyName  string   `json:"friendly_name"`
	GroupName     string   `json:"group_name"`
	Tags          []string `json:"tags"`
	Notes         string   `json:"notes"`
	Description   string   `json:"description"`
	RealIPAddress string   `json:"real_ip_address"`
}

func (Monitor) Create(ctx context.Context, req infer.CreateRequest[MonitorArgs]) (infer.CreateResponse[MonitorState], error) {
	in := req.Inputs
	if in.Type == "" {
		in.Type = "http"
	}
	if req.DryRun {
		return infer.CreateResponse[MonitorState]{Output: MonitorState{MonitorArgs: in}}, nil
	}

	body := monitorCreateBody{
		URL: in.URL, TeamID: in.TeamID, Type: in.Type, Checks: in.Checks,
		FriendlyName: in.FriendlyName, GroupName: in.GroupName, Tags: in.Tags,
		Notes: in.Notes, Description: in.Description, RealIPAddress: in.RealIPAddress,
	}
	var out monitorWire
	if err := infer.GetConfig[Config](ctx).client.Do(ctx, http.MethodPost, "/monitors", body, &out); err != nil {
		return infer.CreateResponse[MonitorState]{}, err
	}
	return infer.CreateResponse[MonitorState]{ID: strconv.Itoa(out.ID), Output: out.toState(in)}, nil
}

func (Monitor) Read(ctx context.Context, req infer.ReadRequest[MonitorArgs, MonitorState]) (infer.ReadResponse[MonitorArgs, MonitorState], error) {
	var out monitorWire
	err := infer.GetConfig[Config](ctx).client.Do(ctx, http.MethodGet, "/monitors/"+req.ID, nil, &out)
	if err != nil {
		if apiStatus(err) == http.StatusNotFound {
			return infer.ReadResponse[MonitorArgs, MonitorState]{}, nil
		}
		return infer.ReadResponse[MonitorArgs, MonitorState]{}, err
	}
	st := out.toState(req.Inputs)
	return infer.ReadResponse[MonitorArgs, MonitorState]{ID: req.ID, Inputs: st.MonitorArgs, State: st}, nil
}

func (Monitor) Update(ctx context.Context, req infer.UpdateRequest[MonitorArgs, MonitorState]) (infer.UpdateResponse[MonitorState], error) {
	in := req.Inputs
	in.TeamID = req.State.TeamID
	in.Type = req.State.Type
	if req.DryRun {
		out := req.State
		out.MonitorArgs = in
		return infer.UpdateResponse[MonitorState]{Output: out}, nil
	}

	body := monitorUpdateBody{
		URL: in.URL, Checks: in.Checks, FriendlyName: in.FriendlyName,
		GroupName: in.GroupName, Tags: in.Tags, Notes: in.Notes,
		Description: in.Description, RealIPAddress: in.RealIPAddress,
	}
	var out monitorWire
	if err := infer.GetConfig[Config](ctx).client.Do(ctx, http.MethodPut, "/monitors/"+req.ID, body, &out); err != nil {
		return infer.UpdateResponse[MonitorState]{}, err
	}
	return infer.UpdateResponse[MonitorState]{Output: out.toState(in)}, nil
}

func (Monitor) Delete(ctx context.Context, req infer.DeleteRequest[MonitorState]) (infer.DeleteResponse, error) {
	err := infer.GetConfig[Config](ctx).client.Do(ctx, http.MethodDelete, "/monitors/"+req.ID, nil, nil)
	if err != nil && apiStatus(err) == http.StatusNotFound {
		return infer.DeleteResponse{}, nil
	}
	return infer.DeleteResponse{}, err
}
