package provider

import (
	"reflect"
	"testing"
)

func TestMergeSettings(t *testing.T) {
	cases := map[string]struct {
		want, got, expect map[string]interface{}
	}{
		"drops server-only read-only keys": {
			want:   map[string]interface{}{"continent": "eu"},
			got:    map[string]interface{}{"continent": "eu", "baseline_confirmed": true},
			expect: map[string]interface{}{"continent": "eu"},
		},
		"surfaces a UI edit as drift": {
			want:   map[string]interface{}{"continent": "eu"},
			got:    map[string]interface{}{"continent": "us"},
			expect: map[string]interface{}{"continent": "us"},
		},
		"keeps sent values when the API omits the block": {
			want:   map[string]interface{}{"continent": "eu"},
			got:    nil,
			expect: map[string]interface{}{"continent": "eu"},
		},
		"unmanaged block stays nil": {
			want:   nil,
			got:    map[string]interface{}{"continent": "eu"},
			expect: nil,
		},
		"managed key removed server-side is dropped": {
			want:   map[string]interface{}{"continent": "eu", "device_emulation": "mobile"},
			got:    map[string]interface{}{"continent": "eu"},
			expect: map[string]interface{}{"continent": "eu"},
		},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			if out := mergeSettings(c.want, c.got); !reflect.DeepEqual(out, c.expect) {
				t.Fatalf("mergeSettings(%v, %v) = %v, want %v", c.want, c.got, out, c.expect)
			}
		})
	}
}
