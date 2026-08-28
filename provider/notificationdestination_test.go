package provider

import "testing"

func TestNDRoutesFor(t *testing.T) {
	cases := []struct {
		level                  string
		list, create, put, del string
	}{
		{"team",
			"/team-notification-destinations",
			"/team-notification-destinations/7",
			"/team-notification-destinations/7/destination",
			"/team-notification-destinations/7/destination"},
		{"monitor",
			"/monitors/7/notification-destinations",
			"/monitors/7/notification-destinations",
			"/monitors/7/notification-destinations",
			"/monitors/7/notification-destinations"},
		{"tag",
			"/tags/notification-destinations",
			"/tags/7/notification-destinations",
			"/tags/7/notification-destinations",
			"/tags/7/notification-destinations/destination"},
		{"tagGroup",
			"/tag-groups/7/notification-destinations",
			"/tag-groups/7/notification-destinations",
			"/tag-groups/7/notification-destinations",
			"/tag-groups/7/notification-destinations"},
	}
	for _, c := range cases {
		t.Run(c.level, func(t *testing.T) {
			r, err := ndRoutesFor(c.level, 7)
			if err != nil {
				t.Fatal(err)
			}
			if r.list != c.list || r.create != c.create || r.put != c.put || r.del != c.del {
				t.Fatalf("got %+v", r)
			}
		})
	}

	if _, err := ndRoutesFor("nope", 1); err == nil {
		t.Fatal("expected error for unknown level")
	}
}
