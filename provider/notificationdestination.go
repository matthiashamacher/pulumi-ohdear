package provider

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/pulumi/pulumi-go-provider/infer"
)

// NotificationDestination is an Oh Dear notification destination
// (https://ohdear.app/docs/api/notification-destinations).
//
// Oh Dear scopes destinations to one of four owners — a team, a monitor, a tag
// or a tag group — each with its own (slightly irregular) set of URLs. `level`
// and `ownerId` pick the owner and are immutable; everything else is editable.
//
// ponytail: `destination` is kept from inputs on refresh (it holds channel
// secrets the API masks) and its values are strings — every channel in the docs
// fits except Opsgenie's boolean `euEndpoint`; send "true"/"false" there.
type NotificationDestination struct{}

type NotificationDestinationArgs struct {
	Level             string            `pulumi:"level" provider:"replaceOnChanges"`
	OwnerID           int               `pulumi:"ownerId" provider:"replaceOnChanges"`
	Channel           string            `pulumi:"channel"`
	Destination       map[string]string `pulumi:"destination" provider:"secret"`
	NotificationTypes []string          `pulumi:"notificationTypes,optional"`
	Label             string            `pulumi:"label,optional"`
	Enabled           bool              `pulumi:"enabled,optional"`
}

type NotificationDestinationState struct {
	NotificationDestinationArgs

	DestinationID  int    `pulumi:"destinationId"`
	DisabledReason string `pulumi:"disabledReason"`
}

func (d *NotificationDestination) Annotate(a infer.Annotator) {
	a.Describe(&d, "An Oh Dear notification destination for a team, monitor, tag or tag group.")
}

func (a *NotificationDestinationArgs) Annotate(an infer.Annotator) {
	an.Describe(&a.Level, `Owner type: "team", "monitor", "tag" or "tagGroup".`)
	an.Describe(&a.OwnerID, "ID of the owning team, monitor, tag or tag group.")
	an.Describe(&a.Channel, `Notification channel, e.g. "mail", "slack", "discord", "webhook".`)
	an.SetDefault(&a.Enabled, true)
}

type ndWire struct {
	ID                int      `json:"id"`
	Label             string   `json:"label"`
	Channel           string   `json:"channel"`
	NotificationTypes []string `json:"notification_types"`
	Enabled           bool     `json:"enabled"`
	DisabledReason    string   `json:"disabled_reason"`
}

// ndRoutes holds the per-level URL prefixes. Item operations append "/{id}".
type ndRoutes struct {
	list, create, put, del string
}

func ndRoutesFor(level string, ownerID int) (ndRoutes, error) {
	o := strconv.Itoa(ownerID)
	switch level {
	case "team":
		base := "/team-notification-destinations"
		return ndRoutes{base, base + "/" + o, base + "/" + o + "/destination", base + "/" + o + "/destination"}, nil
	case "monitor":
		base := "/monitors/" + o + "/notification-destinations"
		return ndRoutes{base, base, base, base}, nil
	case "tag":
		owned := "/tags/" + o + "/notification-destinations"
		return ndRoutes{"/tags/notification-destinations", owned, owned, owned + "/destination"}, nil
	case "tagGroup":
		base := "/tag-groups/" + o + "/notification-destinations"
		return ndRoutes{base, base, base, base}, nil
	}
	return ndRoutes{}, fmt.Errorf("unknown level %q (want team, monitor, tag or tagGroup)", level)
}

type ndBody struct {
	Channel           string            `json:"channel"`
	Destination       map[string]string `json:"destination"`
	NotificationTypes []string          `json:"notification_types,omitempty"`
	Label             string            `json:"label,omitempty"`
}

func (in NotificationDestinationArgs) body() ndBody {
	return ndBody{in.Channel, in.Destination, in.NotificationTypes, in.Label}
}

func (in NotificationDestinationArgs) toState(w ndWire) NotificationDestinationState {
	in.Enabled = w.Enabled
	return NotificationDestinationState{
		NotificationDestinationArgs: in,
		DestinationID:               w.ID,
		DisabledReason:              w.DisabledReason,
	}
}

// reconcileEnabled flips the destination to want via the enable/disable
// endpoints (which take no body and never notify) when it isn't already there.
func reconcileEnabled(ctx context.Context, c *Client, r ndRoutes, id int, want bool, w ndWire) (ndWire, error) {
	if w.Enabled == want {
		return w, nil
	}
	verb := "disable"
	if want {
		verb = "enable"
	}
	var out ndWire
	err := c.Do(ctx, http.MethodPost, r.put+"/"+strconv.Itoa(id)+"/"+verb, nil, &out)
	return out, err
}

func (NotificationDestination) Create(ctx context.Context, req infer.CreateRequest[NotificationDestinationArgs]) (infer.CreateResponse[NotificationDestinationState], error) {
	in := req.Inputs
	if req.DryRun {
		return infer.CreateResponse[NotificationDestinationState]{Output: NotificationDestinationState{NotificationDestinationArgs: in}}, nil
	}
	routes, err := ndRoutesFor(in.Level, in.OwnerID)
	if err != nil {
		return infer.CreateResponse[NotificationDestinationState]{}, err
	}

	c := infer.GetConfig[Config](ctx).client
	var out ndWire
	if err := c.Do(ctx, http.MethodPost, routes.create, in.body(), &out); err != nil {
		return infer.CreateResponse[NotificationDestinationState]{}, err
	}
	if out, err = reconcileEnabled(ctx, c, routes, out.ID, in.Enabled, out); err != nil {
		return infer.CreateResponse[NotificationDestinationState]{}, err
	}
	return infer.CreateResponse[NotificationDestinationState]{ID: strconv.Itoa(out.ID), Output: in.toState(out)}, nil
}

func (NotificationDestination) Read(ctx context.Context, req infer.ReadRequest[NotificationDestinationArgs, NotificationDestinationState]) (infer.ReadResponse[NotificationDestinationArgs, NotificationDestinationState], error) {
	routes, err := ndRoutesFor(req.Inputs.Level, req.Inputs.OwnerID)
	if err != nil {
		return infer.ReadResponse[NotificationDestinationArgs, NotificationDestinationState]{}, err
	}
	list, err := List[ndWire](ctx, infer.GetConfig[Config](ctx).client, routes.list)
	if err != nil {
		return infer.ReadResponse[NotificationDestinationArgs, NotificationDestinationState]{}, err
	}
	id, _ := strconv.Atoi(req.ID)
	for _, w := range list {
		if w.ID != id {
			continue
		}
		st := req.Inputs.toState(w)
		return infer.ReadResponse[NotificationDestinationArgs, NotificationDestinationState]{ID: req.ID, Inputs: st.NotificationDestinationArgs, State: st}, nil
	}
	return infer.ReadResponse[NotificationDestinationArgs, NotificationDestinationState]{}, nil
}

func (NotificationDestination) Update(ctx context.Context, req infer.UpdateRequest[NotificationDestinationArgs, NotificationDestinationState]) (infer.UpdateResponse[NotificationDestinationState], error) {
	in := req.Inputs
	in.Level = req.State.Level
	in.OwnerID = req.State.OwnerID
	if req.DryRun {
		out := req.State
		out.NotificationDestinationArgs = in
		return infer.UpdateResponse[NotificationDestinationState]{Output: out}, nil
	}
	routes, err := ndRoutesFor(in.Level, in.OwnerID)
	if err != nil {
		return infer.UpdateResponse[NotificationDestinationState]{}, err
	}

	c := infer.GetConfig[Config](ctx).client
	var out ndWire
	if err := c.Do(ctx, http.MethodPut, routes.put+"/"+req.ID, in.body(), &out); err != nil {
		return infer.UpdateResponse[NotificationDestinationState]{}, err
	}
	if out, err = reconcileEnabled(ctx, c, routes, out.ID, in.Enabled, out); err != nil {
		return infer.UpdateResponse[NotificationDestinationState]{}, err
	}
	return infer.UpdateResponse[NotificationDestinationState]{Output: in.toState(out)}, nil
}

func (NotificationDestination) Delete(ctx context.Context, req infer.DeleteRequest[NotificationDestinationState]) (infer.DeleteResponse, error) {
	routes, err := ndRoutesFor(req.State.Level, req.State.OwnerID)
	if err != nil {
		return infer.DeleteResponse{}, err
	}
	err = infer.GetConfig[Config](ctx).client.Do(ctx, http.MethodDelete, routes.del+"/"+req.ID, nil, nil)
	if err != nil && apiStatus(err) == http.StatusNotFound {
		return infer.DeleteResponse{}, nil
	}
	return infer.DeleteResponse{}, err
}
