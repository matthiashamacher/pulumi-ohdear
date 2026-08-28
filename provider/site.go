package provider

import (
	"context"

	"github.com/pulumi/pulumi-go-provider/infer"
)

// Site is an Oh Dear monitored site.
//
// ponytail: stub CRUD, echoes inputs and mints a fake ID. Wire the Oh Dear API
// client (https://ohdear.app/docs/integrations/the-oh-dear-api) before real use.
type Site struct{}

type SiteArgs struct {
	URL   string   `pulumi:"url"`
	Teams []int    `pulumi:"teams,optional"`
	Tags  []string `pulumi:"tags,optional"`
}

type SiteState struct {
	SiteArgs
	SiteID int `pulumi:"siteId"`
}

func (s *Site) Annotate(a infer.Annotator) {
	a.Describe(&s, "A website monitored by Oh Dear.")
}

func (Site) Create(ctx context.Context, req infer.CreateRequest[SiteArgs]) (infer.CreateResponse[SiteState], error) {
	state := SiteState{SiteArgs: req.Inputs}
	if req.DryRun {
		return infer.CreateResponse[SiteState]{ID: "", Output: state}, nil
	}
	// ponytail: replace with POST /api/sites
	state.SiteID = 0
	return infer.CreateResponse[SiteState]{ID: req.Inputs.URL, Output: state}, nil
}
