package provider

import (
	"context"
	"net/http"
	"strconv"

	"github.com/pulumi/pulumi-go-provider/infer"
)

// TagGroup is an Oh Dear tag group (https://ohdear.app/docs/api/tags-and-tag-groups).
type TagGroup struct{}

type TagGroupArgs struct {
	TeamID int      `pulumi:"teamId"`
	Label  string   `pulumi:"label"`
	Tags   []string `pulumi:"tags,optional"`
}

type TagGroupState struct {
	TagGroupArgs

	GroupID   int    `pulumi:"groupId"`
	TeamName  string `pulumi:"teamName"`
	CreatedAt string `pulumi:"createdAt"`
}

func (g *TagGroup) Annotate(a infer.Annotator) {
	a.Describe(&g, "An Oh Dear tag group. `tags` holds tag names and supports wildcards such as `prod-*`.")
}

func (a *TagGroupArgs) Annotate(an infer.Annotator) {
	an.Describe(&a.TeamID, "Team that owns the tag group.")
	an.Describe(&a.Label, "Display label for the group.")
	an.Describe(&a.Tags, "Tag names in the group; wildcards such as prod-* are allowed.")
}

func (s *TagGroupState) Annotate(an infer.Annotator) {
	an.Describe(&s.GroupID, "Numeric Oh Dear tag group ID.")
	an.Describe(&s.TeamName, "Name of the owning team.")
	an.Describe(&s.CreatedAt, "When the tag group was created, as an ISO 8601 timestamp.")
}

type tagGroupWire struct {
	ID        int    `json:"id"`
	Label     string `json:"label"`
	TeamID    int    `json:"team_id"`
	TeamName  string `json:"team_name"`
	CreatedAt string `json:"created_at"`
}

// tag-group request body, shared by create (POST) and update (PUT).
type tagGroupBody struct {
	TeamID int      `json:"team_id"`
	Label  string   `json:"label"`
	Tags   []string `json:"tags,omitempty"`
}

func (in TagGroupArgs) body() tagGroupBody {
	return tagGroupBody{TeamID: in.TeamID, Label: in.Label, Tags: in.Tags}
}

// merge keeps the inputs (tag names, possibly wildcards) and layers the
// server-computed fields on top.
func (w tagGroupWire) toState(inputs TagGroupArgs) TagGroupState {
	return TagGroupState{
		TagGroupArgs: inputs,
		GroupID:      w.ID,
		TeamName:     w.TeamName,
		CreatedAt:    w.CreatedAt,
	}
}

func (TagGroup) Create(ctx context.Context, req infer.CreateRequest[TagGroupArgs]) (infer.CreateResponse[TagGroupState], error) {
	in := req.Inputs
	if req.DryRun {
		return infer.CreateResponse[TagGroupState]{Output: TagGroupState{TagGroupArgs: in}}, nil
	}

	var out tagGroupWire
	if err := infer.GetConfig[Config](ctx).client.Do(ctx, http.MethodPost, "/tag-groups", in.body(), &out); err != nil {
		return infer.CreateResponse[TagGroupState]{}, err
	}
	return infer.CreateResponse[TagGroupState]{ID: strconv.Itoa(out.ID), Output: out.toState(in)}, nil
}

func (TagGroup) Read(ctx context.Context, req infer.ReadRequest[TagGroupArgs, TagGroupState]) (infer.ReadResponse[TagGroupArgs, TagGroupState], error) {
	groups, err := List[tagGroupWire](ctx, infer.GetConfig[Config](ctx).client, "/tag-groups")
	if err != nil {
		return infer.ReadResponse[TagGroupArgs, TagGroupState]{}, err
	}
	id, _ := strconv.Atoi(req.ID)
	for _, w := range groups {
		if w.ID != id {
			continue
		}
		st := w.toState(req.Inputs)
		return infer.ReadResponse[TagGroupArgs, TagGroupState]{ID: req.ID, Inputs: st.TagGroupArgs, State: st}, nil
	}
	return infer.ReadResponse[TagGroupArgs, TagGroupState]{}, nil
}

func (TagGroup) Update(ctx context.Context, req infer.UpdateRequest[TagGroupArgs, TagGroupState]) (infer.UpdateResponse[TagGroupState], error) {
	in := req.Inputs
	if req.DryRun {
		return infer.UpdateResponse[TagGroupState]{Output: in.merge(req.State)}, nil
	}

	var out tagGroupWire
	if err := infer.GetConfig[Config](ctx).client.Do(ctx, http.MethodPut, "/tag-groups/"+req.ID, in.body(), &out); err != nil {
		return infer.UpdateResponse[TagGroupState]{}, err
	}
	return infer.UpdateResponse[TagGroupState]{Output: out.toState(in)}, nil
}

func (TagGroup) Delete(ctx context.Context, req infer.DeleteRequest[TagGroupState]) (infer.DeleteResponse, error) {
	err := infer.GetConfig[Config](ctx).client.Do(ctx, http.MethodDelete, "/tag-groups/"+req.ID, nil, nil)
	return infer.DeleteResponse{}, err
}

// merge produces the previewed post-update state without contacting the API.
func (in TagGroupArgs) merge(prev TagGroupState) TagGroupState {
	prev.TagGroupArgs = in
	return prev
}
