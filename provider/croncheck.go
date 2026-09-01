package provider

import (
	"context"
	"net/http"
	"strconv"

	"github.com/pulumi/pulumi-go-provider/infer"
)

// CronCheck is an Oh Dear cron job monitor
// (https://ohdear.app/docs/api/cron-job-monitoring).
//
// A check is either `simple` (ping expected every `frequencyInMinutes`) or
// `cron` (ping expected per `cronExpression` in `serverTimezone`). Every check
// gets a `pingUrl` the job curls on success.
//
// ponytail: the snooze/unsnooze and bulk `sync` endpoints are imperative
// actions, not desired state, so they are not modelled.
type CronCheck struct{}

type CronCheckArgs struct {
	// monitorId is the path parent; the API has no endpoint to move a check
	// between monitors, so a change replaces the resource.
	MonitorID          int    `pulumi:"monitorId" provider:"replaceOnChanges"`
	Name               string `pulumi:"name"`
	Type               string `pulumi:"type"`
	Description         string `pulumi:"description,optional"`
	FrequencyInMinutes int    `pulumi:"frequencyInMinutes,optional"`
	GraceTimeInMinutes int    `pulumi:"graceTimeInMinutes,optional"`
	CronExpression     string `pulumi:"cronExpression,optional"`
	ServerTimezone     string `pulumi:"serverTimezone,optional"`
}

type CronCheckState struct {
	CronCheckArgs

	CronCheckID                 int    `pulumi:"cronCheckId"`
	UUID                        string `pulumi:"uuid"`
	PingURL                     string `pulumi:"pingUrl"`
	HumanReadableCronExpression string `pulumi:"humanReadableCronExpression"`
	CreatedAt                   string `pulumi:"createdAt"`
}

func (c *CronCheck) Annotate(a infer.Annotator) {
	a.Describe(&c, "An Oh Dear cron job monitor. Curl `pingUrl` from the job on success.")
}

func (a *CronCheckArgs) Annotate(an infer.Annotator) {
	an.Describe(&a.Type, `One of "simple" (uses frequencyInMinutes) or "cron" (uses cronExpression + serverTimezone).`)
	an.Describe(&a.GraceTimeInMinutes, "How long a ping may be late before the check alerts.")
}

type cronCheckWire struct {
	ID                          int    `json:"id"`
	UUID                        string `json:"uuid"`
	Name                        string `json:"name"`
	Type                        string `json:"type"`
	Description                 string `json:"description"`
	FrequencyInMinutes          int    `json:"frequency_in_minutes"`
	GraceTimeInMinutes          int    `json:"grace_time_in_minutes"`
	CronExpression              string `json:"cron_expression"`
	HumanReadableCronExpression string `json:"human_readable_cron_expression"`
	ServerTimezone              string `json:"server_timezone"`
	PingURL                     string `json:"ping_url"`
	CreatedAt                   string `json:"created_at"`
}

// toState refreshes server-owned fields. monitorId is kept from inputs: the
// per-monitor response does not echo it.
func (w cronCheckWire) toState(inputs CronCheckArgs) CronCheckState {
	inputs.Name = w.Name
	inputs.Type = w.Type
	inputs.Description = w.Description
	inputs.FrequencyInMinutes = w.FrequencyInMinutes
	inputs.GraceTimeInMinutes = w.GraceTimeInMinutes
	inputs.CronExpression = w.CronExpression
	inputs.ServerTimezone = w.ServerTimezone
	return CronCheckState{
		CronCheckArgs:               inputs,
		CronCheckID:                 w.ID,
		UUID:                        w.UUID,
		PingURL:                     w.PingURL,
		HumanReadableCronExpression: w.HumanReadableCronExpression,
		CreatedAt:                   w.CreatedAt,
	}
}

// cronCheckBody carries the writable fields. omitempty lets Oh Dear ignore the
// knobs that do not apply to the chosen type.
type cronCheckBody struct {
	Name               string `json:"name"`
	Type               string `json:"type"`
	Description         string `json:"description,omitempty"`
	FrequencyInMinutes int    `json:"frequency_in_minutes,omitempty"`
	GraceTimeInMinutes int    `json:"grace_time_in_minutes,omitempty"`
	CronExpression     string `json:"cron_expression,omitempty"`
	ServerTimezone     string `json:"server_timezone,omitempty"`
}

func bodyOf(in CronCheckArgs) cronCheckBody {
	return cronCheckBody{
		Name: in.Name, Type: in.Type, Description: in.Description,
		FrequencyInMinutes: in.FrequencyInMinutes, GraceTimeInMinutes: in.GraceTimeInMinutes,
		CronExpression: in.CronExpression, ServerTimezone: in.ServerTimezone,
	}
}

func (CronCheck) Create(ctx context.Context, req infer.CreateRequest[CronCheckArgs]) (infer.CreateResponse[CronCheckState], error) {
	in := req.Inputs
	if req.DryRun {
		return infer.CreateResponse[CronCheckState]{Output: CronCheckState{CronCheckArgs: in}}, nil
	}

	var out cronCheckWire
	path := "/monitors/" + strconv.Itoa(in.MonitorID) + "/cron-checks"
	if err := infer.GetConfig[Config](ctx).client.Do(ctx, http.MethodPost, path, bodyOf(in), &out); err != nil {
		return infer.CreateResponse[CronCheckState]{}, err
	}
	return infer.CreateResponse[CronCheckState]{ID: strconv.Itoa(out.ID), Output: out.toState(in)}, nil
}

func (CronCheck) Read(ctx context.Context, req infer.ReadRequest[CronCheckArgs, CronCheckState]) (infer.ReadResponse[CronCheckArgs, CronCheckState], error) {
	// No GET-by-id endpoint; list the parent monitor's checks and pick ours.
	path := "/monitors/" + strconv.Itoa(req.Inputs.MonitorID) + "/cron-checks"
	checks, err := List[cronCheckWire](ctx, infer.GetConfig[Config](ctx).client, path)
	if err != nil {
		return infer.ReadResponse[CronCheckArgs, CronCheckState]{}, err
	}
	id, _ := strconv.Atoi(req.ID)
	for _, w := range checks {
		if w.ID != id {
			continue
		}
		st := w.toState(req.Inputs)
		return infer.ReadResponse[CronCheckArgs, CronCheckState]{ID: req.ID, Inputs: st.CronCheckArgs, State: st}, nil
	}
	return infer.ReadResponse[CronCheckArgs, CronCheckState]{}, nil
}

func (CronCheck) Update(ctx context.Context, req infer.UpdateRequest[CronCheckArgs, CronCheckState]) (infer.UpdateResponse[CronCheckState], error) {
	in := req.Inputs
	in.MonitorID = req.State.MonitorID
	if req.DryRun {
		out := req.State
		out.CronCheckArgs = in
		return infer.UpdateResponse[CronCheckState]{Output: out}, nil
	}

	var out cronCheckWire
	if err := infer.GetConfig[Config](ctx).client.Do(ctx, http.MethodPut, "/cron-checks/"+req.ID, bodyOf(in), &out); err != nil {
		return infer.UpdateResponse[CronCheckState]{}, err
	}
	return infer.UpdateResponse[CronCheckState]{Output: out.toState(in)}, nil
}

func (CronCheck) Delete(ctx context.Context, req infer.DeleteRequest[CronCheckState]) (infer.DeleteResponse, error) {
	err := infer.GetConfig[Config](ctx).client.Do(ctx, http.MethodDelete, "/cron-checks/"+req.ID, nil, nil)
	if err != nil && apiStatus(err) == http.StatusNotFound {
		return infer.DeleteResponse{}, nil
	}
	return infer.DeleteResponse{}, err
}
