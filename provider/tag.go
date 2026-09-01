package provider

import (
	"context"
	"net/http"
	"strconv"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
)

// Tag is an Oh Dear tag (https://ohdear.app/docs/api/tags-and-tag-groups).
//
// ponytail: the API only exposes create + list for tags. There is no update or
// delete endpoint, so any input change replaces the tag and `pulumi destroy`
// only drops it from state — the tag stays in Oh Dear.
type Tag struct{}

type TagArgs struct {
	TeamID   int    `pulumi:"teamId"`
	Name     string `pulumi:"name"`
	Monitors []int  `pulumi:"monitors,optional"`
}

type TagState struct {
	TagArgs

	TagID     int    `pulumi:"tagId"`
	Slug      string `pulumi:"slug"`
	TeamName  string `pulumi:"teamName"`
	Sites     []int  `pulumi:"sites"`
	CreatedAt string `pulumi:"createdAt"`
	UpdatedAt string `pulumi:"updatedAt"`
}

func (t *Tag) Annotate(a infer.Annotator) {
	a.Describe(&t, "An Oh Dear tag. The API cannot update or delete tags: changing an input "+
		"replaces the resource, and destroy only removes it from Pulumi state.")
}

func (a *TagArgs) Annotate(an infer.Annotator) {
	an.Describe(&a.TeamID, "Team that owns the tag.")
	an.Describe(&a.Name, "Tag name.")
	an.Describe(&a.Monitors, "IDs of monitors to attach the tag to.")
}

func (s *TagState) Annotate(an infer.Annotator) {
	an.Describe(&s.TagID, "Numeric Oh Dear tag ID.")
	an.Describe(&s.Slug, "URL-safe form of the tag name.")
	an.Describe(&s.TeamName, "Name of the owning team.")
	an.Describe(&s.Sites, "IDs of the sites/monitors currently carrying the tag.")
	an.Describe(&s.CreatedAt, "When the tag was created, as an ISO 8601 timestamp.")
	an.Describe(&s.UpdatedAt, "When the tag was last updated, as an ISO 8601 timestamp.")
}

type tagWire struct {
	ID        int    `json:"id"`
	TeamID    int    `json:"team_id"`
	TeamName  string `json:"team_name"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	Sites     []int  `json:"sites"`
}

func (w tagWire) toState(inputs TagArgs) TagState {
	return TagState{
		TagArgs:   inputs,
		TagID:     w.ID,
		Slug:      w.Slug,
		TeamName:  w.TeamName,
		Sites:     w.Sites,
		CreatedAt: w.CreatedAt,
		UpdatedAt: w.UpdatedAt,
	}
}

func (Tag) Create(ctx context.Context, req infer.CreateRequest[TagArgs]) (infer.CreateResponse[TagState], error) {
	in := req.Inputs
	if req.DryRun {
		return infer.CreateResponse[TagState]{Output: TagState{TagArgs: in}}, nil
	}

	body := struct {
		TeamID   int    `json:"team_id"`
		Name     string `json:"name"`
		Monitors []int  `json:"monitors,omitempty"`
	}{in.TeamID, in.Name, in.Monitors}

	var out tagWire
	if err := infer.GetConfig[Config](ctx).client.Do(ctx, http.MethodPost, "/tags", body, &out); err != nil {
		return infer.CreateResponse[TagState]{}, err
	}
	return infer.CreateResponse[TagState]{ID: strconv.Itoa(out.ID), Output: out.toState(in)}, nil
}

func (Tag) Read(ctx context.Context, req infer.ReadRequest[TagArgs, TagState]) (infer.ReadResponse[TagArgs, TagState], error) {
	tags, err := List[tagWire](ctx, infer.GetConfig[Config](ctx).client, "/tags")
	if err != nil {
		return infer.ReadResponse[TagArgs, TagState]{}, err
	}
	id, _ := strconv.Atoi(req.ID)
	for _, w := range tags {
		if w.ID != id {
			continue
		}
		st := w.toState(req.Inputs)
		return infer.ReadResponse[TagArgs, TagState]{ID: req.ID, Inputs: st.TagArgs, State: st}, nil
	}
	// Not found: report the resource as deleted.
	return infer.ReadResponse[TagArgs, TagState]{}, nil
}

func (Tag) Delete(ctx context.Context, req infer.DeleteRequest[TagState]) (infer.DeleteResponse, error) {
	p.GetLogger(ctx).Warningf(
		"Oh Dear has no tag-deletion endpoint; tag %q (id %s) was removed from Pulumi state but still exists in Oh Dear.",
		req.State.Name, req.ID)
	return infer.DeleteResponse{}, nil
}
