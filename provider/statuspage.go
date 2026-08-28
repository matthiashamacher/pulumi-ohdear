package provider

import (
	"context"
	"net/http"
	"strconv"

	"github.com/pulumi/pulumi-go-provider/infer"
)

// StatusPage is an Oh Dear status page (https://ohdear.app/docs/api/status-pages).
//
// The API has no endpoint to edit the page itself, only its monitor list, so
// changing `title` or `teamId` replaces the resource while `monitors` changes
// are synced in place.
type StatusPage struct{}

type StatusPageMonitor struct {
	ID        int  `pulumi:"id"`
	Clickable bool `pulumi:"clickable,optional"`
}

type StatusPageArgs struct {
	TeamID   int                 `pulumi:"teamId" provider:"replaceOnChanges"`
	Title    string              `pulumi:"title" provider:"replaceOnChanges"`
	Monitors []StatusPageMonitor `pulumi:"monitors"`
}

type StatusPageState struct {
	StatusPageArgs

	StatusPageID     int    `pulumi:"statusPageId"`
	Slug             string `pulumi:"slug"`
	Domain           string `pulumi:"domain"`
	FullURL          string `pulumi:"fullUrl"`
	Timezone         string `pulumi:"timezone"`
	SummarizedStatus string `pulumi:"summarizedStatus"`
	BadgeID          string `pulumi:"badgeId"`
	CreatedAt        string `pulumi:"createdAt"`
	UpdatedAt        string `pulumi:"updatedAt"`
}

func (s *StatusPage) Annotate(a infer.Annotator) {
	a.Describe(&s, "An Oh Dear status page. `monitors` is the full set of monitors shown on the page.")
}

type statusPageWire struct {
	ID               int    `json:"id"`
	Title            string `json:"title"`
	Domain           string `json:"domain"`
	Slug             string `json:"slug"`
	FullURL          string `json:"full_url"`
	Timezone         string `json:"timezone"`
	SummarizedStatus string `json:"summarized_status"`
	BadgeID          string `json:"badge_id"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
	Team             struct {
		ID int `json:"id"`
	} `json:"team"`
}

// toState refreshes server-owned fields. `monitors` is kept from inputs: the API
// echoes full monitor objects, not the {id, clickable} pairs we send.
func (w statusPageWire) toState(inputs StatusPageArgs) StatusPageState {
	inputs.TeamID = w.Team.ID
	inputs.Title = w.Title
	return StatusPageState{
		StatusPageArgs:   inputs,
		StatusPageID:     w.ID,
		Slug:             w.Slug,
		Domain:           w.Domain,
		FullURL:          w.FullURL,
		Timezone:         w.Timezone,
		SummarizedStatus: w.SummarizedStatus,
		BadgeID:          w.BadgeID,
		CreatedAt:        w.CreatedAt,
		UpdatedAt:        w.UpdatedAt,
	}
}

type statusPageMonitorBody struct {
	ID        int  `json:"id"`
	Clickable bool `json:"clickable,omitempty"`
}

func monitorBodies(ms []StatusPageMonitor) []statusPageMonitorBody {
	out := make([]statusPageMonitorBody, len(ms))
	for i, m := range ms {
		out[i] = statusPageMonitorBody{ID: m.ID, Clickable: m.Clickable}
	}
	return out
}

func (StatusPage) Create(ctx context.Context, req infer.CreateRequest[StatusPageArgs]) (infer.CreateResponse[StatusPageState], error) {
	in := req.Inputs
	if req.DryRun {
		return infer.CreateResponse[StatusPageState]{Output: StatusPageState{StatusPageArgs: in}}, nil
	}

	body := struct {
		TeamID   int                     `json:"team_id"`
		Title    string                  `json:"title"`
		Monitors []statusPageMonitorBody `json:"monitors"`
	}{in.TeamID, in.Title, monitorBodies(in.Monitors)}

	var out statusPageWire
	if err := infer.GetConfig[Config](ctx).client.Do(ctx, http.MethodPost, "/status-pages", body, &out); err != nil {
		return infer.CreateResponse[StatusPageState]{}, err
	}
	return infer.CreateResponse[StatusPageState]{ID: strconv.Itoa(out.ID), Output: out.toState(in)}, nil
}

func (StatusPage) Read(ctx context.Context, req infer.ReadRequest[StatusPageArgs, StatusPageState]) (infer.ReadResponse[StatusPageArgs, StatusPageState], error) {
	var out statusPageWire
	err := infer.GetConfig[Config](ctx).client.Do(ctx, http.MethodGet, "/status-pages/"+req.ID, nil, &out)
	if err != nil {
		if apiStatus(err) == http.StatusNotFound {
			return infer.ReadResponse[StatusPageArgs, StatusPageState]{}, nil
		}
		return infer.ReadResponse[StatusPageArgs, StatusPageState]{}, err
	}
	st := out.toState(req.Inputs)
	return infer.ReadResponse[StatusPageArgs, StatusPageState]{ID: req.ID, Inputs: st.StatusPageArgs, State: st}, nil
}

func (StatusPage) Update(ctx context.Context, req infer.UpdateRequest[StatusPageArgs, StatusPageState]) (infer.UpdateResponse[StatusPageState], error) {
	in := req.Inputs
	in.TeamID = req.State.TeamID
	in.Title = req.State.Title
	if req.DryRun {
		out := req.State
		out.StatusPageArgs = in
		return infer.UpdateResponse[StatusPageState]{Output: out}, nil
	}

	c := infer.GetConfig[Config](ctx).client
	sync := struct {
		Monitors []statusPageMonitorBody `json:"monitors"`
		Sync     bool                    `json:"sync"`
	}{monitorBodies(in.Monitors), true}
	if err := c.Do(ctx, http.MethodPost, "/status-pages/"+req.ID+"/monitors", sync, nil); err != nil {
		return infer.UpdateResponse[StatusPageState]{}, err
	}

	var out statusPageWire
	if err := c.Do(ctx, http.MethodGet, "/status-pages/"+req.ID, nil, &out); err != nil {
		return infer.UpdateResponse[StatusPageState]{}, err
	}
	return infer.UpdateResponse[StatusPageState]{Output: out.toState(in)}, nil
}

func (StatusPage) Delete(ctx context.Context, req infer.DeleteRequest[StatusPageState]) (infer.DeleteResponse, error) {
	err := infer.GetConfig[Config](ctx).client.Do(ctx, http.MethodDelete, "/status-pages/"+req.ID, nil, nil)
	if err != nil && apiStatus(err) == http.StatusNotFound {
		return infer.DeleteResponse{}, nil
	}
	return infer.DeleteResponse{}, err
}
