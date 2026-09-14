package provider

import (
	"context"
	"testing"

	"github.com/pulumi/pulumi-go-provider/infer"
)

func TestTagDiff(t *testing.T) {
	cases := map[string]struct {
		state, inputs TagArgs
		wantChanges   bool
	}{
		"identical inputs, no changes": {
			state:       TagArgs{TeamID: 1, Name: "foo", Monitors: []int{1, 2}},
			inputs:      TagArgs{TeamID: 1, Name: "foo", Monitors: []int{1, 2}},
			wantChanges: false,
		},
		"monitors reordered, no changes": {
			state:       TagArgs{TeamID: 1, Name: "foo", Monitors: []int{1, 2}},
			inputs:      TagArgs{TeamID: 1, Name: "foo", Monitors: []int{2, 1}},
			wantChanges: false,
		},
		"name changed, replace": {
			state:       TagArgs{TeamID: 1, Name: "foo"},
			inputs:      TagArgs{TeamID: 1, Name: "bar"},
			wantChanges: true,
		},
		"teamId changed, replace": {
			state:       TagArgs{TeamID: 1, Name: "foo"},
			inputs:      TagArgs{TeamID: 2, Name: "foo"},
			wantChanges: true,
		},
		"monitors content changed, replace": {
			state:       TagArgs{TeamID: 1, Name: "foo", Monitors: []int{1, 2}},
			inputs:      TagArgs{TeamID: 1, Name: "foo", Monitors: []int{1, 3}},
			wantChanges: true,
		},
		"output-only state noise does not affect diff": {
			state:       TagArgs{TeamID: 1, Name: "foo"},
			inputs:      TagArgs{TeamID: 1, Name: "foo"},
			wantChanges: false,
		},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			resp, err := Tag{}.Diff(context.Background(), infer.DiffRequest[TagArgs, TagState]{
				State:  TagState{TagArgs: c.state, TagID: 42, Slug: "stale", UpdatedAt: "stale"},
				Inputs: c.inputs,
			})
			if err != nil {
				t.Fatalf("Diff returned error: %v", err)
			}
			if resp.HasChanges != c.wantChanges {
				t.Fatalf("HasChanges = %v, want %v (diff: %v)", resp.HasChanges, c.wantChanges, resp.DetailedDiff)
			}
		})
	}
}
