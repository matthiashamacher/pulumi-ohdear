package provider

import (
	"context"
	"net/http"
	"strconv"

	"github.com/pulumi/pulumi-go-provider/infer"
)

// StatusPageUpdateTemplate is a reusable status-page update message
// (https://ohdear.app/docs/api/status-pages).
type StatusPageUpdateTemplate struct{}

type StatusPageUpdateTemplateArgs struct {
	TeamID   int    `pulumi:"teamId" provider:"replaceOnChanges"`
	Name     string `pulumi:"name,optional"`
	Title    string `pulumi:"title,optional"`
	Text     string `pulumi:"text,optional"`
	Severity string `pulumi:"severity,optional"`
}

type StatusPageUpdateTemplateState struct {
	StatusPageUpdateTemplateArgs

	TemplateID int `pulumi:"templateId"`
}

func (t *StatusPageUpdateTemplate) Annotate(a infer.Annotator) {
	a.Describe(&t, "A reusable status page update template.")
}

func (a *StatusPageUpdateTemplateArgs) Annotate(an infer.Annotator) {
	an.Describe(&a.Severity, `One of "info", "warning", "high", "resolved", "scheduled".`)
}

type templateWire struct {
	ID       int    `json:"id"`
	TeamID   int    `json:"team_id"`
	Name     string `json:"name"`
	Title    string `json:"title"`
	Text     string `json:"text"`
	Severity string `json:"severity"`
}

func (w templateWire) toState(inputs StatusPageUpdateTemplateArgs) StatusPageUpdateTemplateState {
	inputs.TeamID = w.TeamID
	inputs.Name = w.Name
	inputs.Title = w.Title
	inputs.Text = w.Text
	inputs.Severity = w.Severity
	return StatusPageUpdateTemplateState{StatusPageUpdateTemplateArgs: inputs, TemplateID: w.ID}
}

func (StatusPageUpdateTemplate) Create(ctx context.Context, req infer.CreateRequest[StatusPageUpdateTemplateArgs]) (infer.CreateResponse[StatusPageUpdateTemplateState], error) {
	in := req.Inputs
	if req.DryRun {
		return infer.CreateResponse[StatusPageUpdateTemplateState]{Output: StatusPageUpdateTemplateState{StatusPageUpdateTemplateArgs: in}}, nil
	}

	body := struct {
		TeamID   int    `json:"team_id"`
		Name     string `json:"name,omitempty"`
		Title    string `json:"title,omitempty"`
		Text     string `json:"text,omitempty"`
		Severity string `json:"severity,omitempty"`
	}{in.TeamID, in.Name, in.Title, in.Text, in.Severity}

	var out templateWire
	if err := infer.GetConfig[Config](ctx).client.Do(ctx, http.MethodPost, "/status-page-update-templates", body, &out); err != nil {
		return infer.CreateResponse[StatusPageUpdateTemplateState]{}, err
	}
	return infer.CreateResponse[StatusPageUpdateTemplateState]{ID: strconv.Itoa(out.ID), Output: out.toState(in)}, nil
}

func (StatusPageUpdateTemplate) Read(ctx context.Context, req infer.ReadRequest[StatusPageUpdateTemplateArgs, StatusPageUpdateTemplateState]) (infer.ReadResponse[StatusPageUpdateTemplateArgs, StatusPageUpdateTemplateState], error) {
	templates, err := List[templateWire](ctx, infer.GetConfig[Config](ctx).client, "/status-page-update-templates")
	if err != nil {
		return infer.ReadResponse[StatusPageUpdateTemplateArgs, StatusPageUpdateTemplateState]{}, err
	}
	id, _ := strconv.Atoi(req.ID)
	for _, w := range templates {
		if w.ID != id {
			continue
		}
		st := w.toState(req.Inputs)
		return infer.ReadResponse[StatusPageUpdateTemplateArgs, StatusPageUpdateTemplateState]{ID: req.ID, Inputs: st.StatusPageUpdateTemplateArgs, State: st}, nil
	}
	return infer.ReadResponse[StatusPageUpdateTemplateArgs, StatusPageUpdateTemplateState]{}, nil
}

func (StatusPageUpdateTemplate) Update(ctx context.Context, req infer.UpdateRequest[StatusPageUpdateTemplateArgs, StatusPageUpdateTemplateState]) (infer.UpdateResponse[StatusPageUpdateTemplateState], error) {
	in := req.Inputs
	in.TeamID = req.State.TeamID
	if req.DryRun {
		return infer.UpdateResponse[StatusPageUpdateTemplateState]{
			Output: StatusPageUpdateTemplateState{StatusPageUpdateTemplateArgs: in, TemplateID: req.State.TemplateID},
		}, nil
	}

	// No omitempty: clearing an input clears it upstream.
	body := struct {
		Name     string `json:"name"`
		Title    string `json:"title"`
		Text     string `json:"text"`
		Severity string `json:"severity"`
	}{in.Name, in.Title, in.Text, in.Severity}

	var out templateWire
	if err := infer.GetConfig[Config](ctx).client.Do(ctx, http.MethodPut, "/status-page-update-templates/"+req.ID, body, &out); err != nil {
		return infer.UpdateResponse[StatusPageUpdateTemplateState]{}, err
	}
	return infer.UpdateResponse[StatusPageUpdateTemplateState]{Output: out.toState(in)}, nil
}

func (StatusPageUpdateTemplate) Delete(ctx context.Context, req infer.DeleteRequest[StatusPageUpdateTemplateState]) (infer.DeleteResponse, error) {
	err := infer.GetConfig[Config](ctx).client.Do(ctx, http.MethodDelete, "/status-page-update-templates/"+req.ID, nil, nil)
	if err != nil && apiStatus(err) == http.StatusNotFound {
		return infer.DeleteResponse{}, nil
	}
	return infer.DeleteResponse{}, err
}
