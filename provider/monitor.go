package provider

import (
	"context"
	"net/http"
	"strconv"

	"github.com/pulumi/pulumi-go-provider/infer"
)

// Monitor is an Oh Dear monitor (https://ohdear.app/docs/api/monitors).
//
// The per-check tuning lives in the `*CheckSettings` maps. They are open
// passthroughs: put whatever the matching `*_check_settings` object at
// https://ohdear.app/docs/api/monitors accepts. On refresh only the keys the
// program set are reconciled against the API response (see mergeSettings), so a
// change made in the Oh Dear UI shows as drift while the read-only keys the API
// adds to its response are ignored.
type Monitor struct{}

type MonitorArgs struct {
	URL           string   `pulumi:"url"`
	TeamID        int      `pulumi:"teamId" provider:"replaceOnChanges"`
	Type          string   `pulumi:"type,optional" provider:"replaceOnChanges"`
	Checks        []string `pulumi:"checks,optional"`
	FriendlyName  string   `pulumi:"friendlyName,optional"`
	GroupName     string   `pulumi:"groupName,optional"`
	Tags          []string `pulumi:"tags,optional"`
	Notes         string   `pulumi:"notes,optional"`
	Description   string   `pulumi:"description,optional"`
	RealIPAddress string   `pulumi:"realIpAddress,optional"`

	SendReportToEmails        []string `pulumi:"sendReportToEmails,optional"`
	IncludeCheckTypesInReport []string `pulumi:"includeCheckTypesInReport,optional"`

	UptimeCheckSettings            map[string]interface{} `pulumi:"uptimeCheckSettings,optional"`
	PerformanceCheckSettings       map[string]interface{} `pulumi:"performanceCheckSettings,optional"`
	BrokenLinksCheckSettings       map[string]interface{} `pulumi:"brokenLinksCheckSettings,optional"`
	CertificateHealthCheckSettings map[string]interface{} `pulumi:"certificateHealthCheckSettings,optional"`
	DNSCheckSettings               map[string]interface{} `pulumi:"dnsCheckSettings,optional"`
	DomainCheckSettings            map[string]interface{} `pulumi:"domainCheckSettings,optional"`
	LighthouseCheckSettings        map[string]interface{} `pulumi:"lighthouseCheckSettings,optional"`
	ApplicationHealthCheckSettings map[string]interface{} `pulumi:"applicationHealthCheckSettings,optional"`
	SitemapCheckSettings           map[string]interface{} `pulumi:"sitemapCheckSettings,optional"`
	PortsCheckSettings             map[string]interface{} `pulumi:"portsCheckSettings,optional"`
	DNSBlocklistCheckSettings      map[string]interface{} `pulumi:"dnsBlocklistCheckSettings,optional"`
	AICheckSettings                map[string]interface{} `pulumi:"aiCheckSettings,optional"`
	CrawlerSettings                map[string]interface{} `pulumi:"crawlerSettings,optional"`
}

type MonitorState struct {
	MonitorArgs

	MonitorID             int    `pulumi:"monitorId"`
	Label                 string `pulumi:"label"`
	UsesHTTPS             bool   `pulumi:"usesHttps"`
	SortURL               string `pulumi:"sortUrl"`
	SummarizedCheckResult string `pulumi:"summarizedCheckResult"`
	CreatedAt             string `pulumi:"createdAt"`
	UpdatedAt             string `pulumi:"updatedAt"`
}

func (m *Monitor) Annotate(a infer.Annotator) {
	a.Describe(&m, "An Oh Dear monitor (http, ping, tcp or ai).")
}

func (a *MonitorArgs) Annotate(an infer.Annotator) {
	an.Describe(&a.Type, `Monitor type: "http", "ping", "tcp" or "ai".`)
	an.SetDefault(&a.Type, "http")
}

type monitorWire struct {
	ID                    int      `json:"id"`
	TeamID                int      `json:"team_id"`
	Type                  string   `json:"type"`
	URL                   string   `json:"url"`
	UsesHTTPS             bool     `json:"uses_https"`
	SortURL               string   `json:"sort_url"`
	Label                 string   `json:"label"`
	GroupName             string   `json:"group_name"`
	Tags                  []string `json:"tags"`
	Notes                 string   `json:"notes"`
	Description           string   `json:"description"`
	RealIPAddress         string   `json:"real_ip_address"`
	SummarizedCheckResult string   `json:"summarized_check_result"`
	CreatedAt             string   `json:"created_at"`
	UpdatedAt             string   `json:"updated_at"`

	SendReportToEmails        []string `json:"send_report_to_emails"`
	IncludeCheckTypesInReport []string `json:"include_check_types_in_report"`

	UptimeCheckSettings            map[string]interface{} `json:"uptime_check_settings"`
	PerformanceCheckSettings       map[string]interface{} `json:"performance_check_settings"`
	BrokenLinksCheckSettings       map[string]interface{} `json:"broken_links_check_settings"`
	CertificateHealthCheckSettings map[string]interface{} `json:"certificate_health_check_settings"`
	DNSCheckSettings               map[string]interface{} `json:"dns_check_settings"`
	DomainCheckSettings            map[string]interface{} `json:"domain_check_settings"`
	LighthouseCheckSettings        map[string]interface{} `json:"lighthouse_check_settings"`
	ApplicationHealthCheckSettings map[string]interface{} `json:"application_health_check_settings"`
	SitemapCheckSettings           map[string]interface{} `json:"sitemap_check_settings"`
	PortsCheckSettings             map[string]interface{} `json:"ports_check_settings"`
	DNSBlocklistCheckSettings      map[string]interface{} `json:"dns_blocklist_check_settings"`
	AICheckSettings                map[string]interface{} `json:"ai_check_settings"`
	CrawlerSettings                map[string]interface{} `json:"crawler_settings"`
}

// mergeSettings reconciles one settings map after a refresh. It keeps only the
// keys the program set (want), taking each value from the API response (got), so
// an edit made in the Oh Dear UI shows as drift while the read-only keys and
// defaults the API adds to its response do not. If the response omits the block
// entirely, the sent values are kept — the API did not report on it.
//
// ponytail: value-level comparison is left to Pulumi's differ; JSON type
// coercion by the API (e.g. a number echoed as a string) can still surface a
// cosmetic diff.
func mergeSettings(want, got map[string]interface{}) map[string]interface{} {
	if want == nil || len(got) == 0 {
		return want
	}
	out := make(map[string]interface{}, len(want))
	for k := range want {
		if v, ok := got[k]; ok {
			out[k] = v
		}
	}
	return out
}

// toState refreshes server-owned fields from a response while preserving the
// inputs that the API never echoes back (friendlyName, checks).
func (w monitorWire) toState(inputs MonitorArgs) MonitorState {
	inputs.URL = w.URL
	inputs.TeamID = w.TeamID
	inputs.Type = w.Type
	inputs.GroupName = w.GroupName
	inputs.Tags = w.Tags
	inputs.Notes = w.Notes
	inputs.Description = w.Description
	inputs.RealIPAddress = w.RealIPAddress
	if w.SendReportToEmails != nil {
		inputs.SendReportToEmails = w.SendReportToEmails
	}
	if w.IncludeCheckTypesInReport != nil {
		inputs.IncludeCheckTypesInReport = w.IncludeCheckTypesInReport
	}
	inputs.UptimeCheckSettings = mergeSettings(inputs.UptimeCheckSettings, w.UptimeCheckSettings)
	inputs.PerformanceCheckSettings = mergeSettings(inputs.PerformanceCheckSettings, w.PerformanceCheckSettings)
	inputs.BrokenLinksCheckSettings = mergeSettings(inputs.BrokenLinksCheckSettings, w.BrokenLinksCheckSettings)
	inputs.CertificateHealthCheckSettings = mergeSettings(inputs.CertificateHealthCheckSettings, w.CertificateHealthCheckSettings)
	inputs.DNSCheckSettings = mergeSettings(inputs.DNSCheckSettings, w.DNSCheckSettings)
	inputs.DomainCheckSettings = mergeSettings(inputs.DomainCheckSettings, w.DomainCheckSettings)
	inputs.LighthouseCheckSettings = mergeSettings(inputs.LighthouseCheckSettings, w.LighthouseCheckSettings)
	inputs.ApplicationHealthCheckSettings = mergeSettings(inputs.ApplicationHealthCheckSettings, w.ApplicationHealthCheckSettings)
	inputs.SitemapCheckSettings = mergeSettings(inputs.SitemapCheckSettings, w.SitemapCheckSettings)
	inputs.PortsCheckSettings = mergeSettings(inputs.PortsCheckSettings, w.PortsCheckSettings)
	inputs.DNSBlocklistCheckSettings = mergeSettings(inputs.DNSBlocklistCheckSettings, w.DNSBlocklistCheckSettings)
	inputs.AICheckSettings = mergeSettings(inputs.AICheckSettings, w.AICheckSettings)
	inputs.CrawlerSettings = mergeSettings(inputs.CrawlerSettings, w.CrawlerSettings)
	return MonitorState{
		MonitorArgs:           inputs,
		MonitorID:             w.ID,
		Label:                 w.Label,
		UsesHTTPS:             w.UsesHTTPS,
		SortURL:               w.SortURL,
		SummarizedCheckResult: w.SummarizedCheckResult,
		CreatedAt:             w.CreatedAt,
		UpdatedAt:             w.UpdatedAt,
	}
}

// monitorSettingsBody holds the fields shared by create and update: the report
// lists and the per-check settings maps. Every entry is omitempty — an unset
// block is left untouched upstream.
type monitorSettingsBody struct {
	SendReportToEmails        []string `json:"send_report_to_emails,omitempty"`
	IncludeCheckTypesInReport []string `json:"include_check_types_in_report,omitempty"`

	UptimeCheckSettings            map[string]interface{} `json:"uptime_check_settings,omitempty"`
	PerformanceCheckSettings       map[string]interface{} `json:"performance_check_settings,omitempty"`
	BrokenLinksCheckSettings       map[string]interface{} `json:"broken_links_check_settings,omitempty"`
	CertificateHealthCheckSettings map[string]interface{} `json:"certificate_health_check_settings,omitempty"`
	DNSCheckSettings               map[string]interface{} `json:"dns_check_settings,omitempty"`
	DomainCheckSettings            map[string]interface{} `json:"domain_check_settings,omitempty"`
	LighthouseCheckSettings        map[string]interface{} `json:"lighthouse_check_settings,omitempty"`
	ApplicationHealthCheckSettings map[string]interface{} `json:"application_health_check_settings,omitempty"`
	SitemapCheckSettings           map[string]interface{} `json:"sitemap_check_settings,omitempty"`
	PortsCheckSettings             map[string]interface{} `json:"ports_check_settings,omitempty"`
	DNSBlocklistCheckSettings      map[string]interface{} `json:"dns_blocklist_check_settings,omitempty"`
	AICheckSettings                map[string]interface{} `json:"ai_check_settings,omitempty"`
	CrawlerSettings                map[string]interface{} `json:"crawler_settings,omitempty"`
}

func settingsBodyOf(in MonitorArgs) monitorSettingsBody {
	return monitorSettingsBody{
		SendReportToEmails:             in.SendReportToEmails,
		IncludeCheckTypesInReport:      in.IncludeCheckTypesInReport,
		UptimeCheckSettings:            in.UptimeCheckSettings,
		PerformanceCheckSettings:       in.PerformanceCheckSettings,
		BrokenLinksCheckSettings:       in.BrokenLinksCheckSettings,
		CertificateHealthCheckSettings: in.CertificateHealthCheckSettings,
		DNSCheckSettings:               in.DNSCheckSettings,
		DomainCheckSettings:            in.DomainCheckSettings,
		LighthouseCheckSettings:        in.LighthouseCheckSettings,
		ApplicationHealthCheckSettings: in.ApplicationHealthCheckSettings,
		SitemapCheckSettings:           in.SitemapCheckSettings,
		PortsCheckSettings:             in.PortsCheckSettings,
		DNSBlocklistCheckSettings:      in.DNSBlocklistCheckSettings,
		AICheckSettings:                in.AICheckSettings,
		CrawlerSettings:                in.CrawlerSettings,
	}
}

// createBody omits empty optionals so Oh Dear applies its own defaults.
type monitorCreateBody struct {
	URL           string   `json:"url"`
	TeamID        int      `json:"team_id"`
	Type          string   `json:"type"`
	Checks        []string `json:"checks,omitempty"`
	FriendlyName  string   `json:"friendly_name,omitempty"`
	GroupName     string   `json:"group_name,omitempty"`
	Tags          []string `json:"tags,omitempty"`
	Notes         string   `json:"notes,omitempty"`
	Description   string   `json:"description,omitempty"`
	RealIPAddress string   `json:"real_ip_address,omitempty"`
	monitorSettingsBody
}

// updateBody sends every managed scalar field (no omitempty) so clearing an
// input in code clears it in Oh Dear. team_id and type are immutable and never
// sent. The settings maps stay omitempty (see the Monitor doc comment).
type monitorUpdateBody struct {
	URL           string   `json:"url"`
	Checks        []string `json:"checks,omitempty"`
	FriendlyName  string   `json:"friendly_name"`
	GroupName     string   `json:"group_name"`
	Tags          []string `json:"tags"`
	Notes         string   `json:"notes"`
	Description   string   `json:"description"`
	RealIPAddress string   `json:"real_ip_address"`
	monitorSettingsBody
}

func (Monitor) Create(ctx context.Context, req infer.CreateRequest[MonitorArgs]) (infer.CreateResponse[MonitorState], error) {
	in := req.Inputs
	if in.Type == "" {
		in.Type = "http"
	}
	if req.DryRun {
		return infer.CreateResponse[MonitorState]{Output: MonitorState{MonitorArgs: in}}, nil
	}

	body := monitorCreateBody{
		URL: in.URL, TeamID: in.TeamID, Type: in.Type, Checks: in.Checks,
		FriendlyName: in.FriendlyName, GroupName: in.GroupName, Tags: in.Tags,
		Notes: in.Notes, Description: in.Description, RealIPAddress: in.RealIPAddress,
		monitorSettingsBody: settingsBodyOf(in),
	}
	var out monitorWire
	if err := infer.GetConfig[Config](ctx).client.Do(ctx, http.MethodPost, "/monitors", body, &out); err != nil {
		return infer.CreateResponse[MonitorState]{}, err
	}
	return infer.CreateResponse[MonitorState]{ID: strconv.Itoa(out.ID), Output: out.toState(in)}, nil
}

func (Monitor) Read(ctx context.Context, req infer.ReadRequest[MonitorArgs, MonitorState]) (infer.ReadResponse[MonitorArgs, MonitorState], error) {
	var out monitorWire
	err := infer.GetConfig[Config](ctx).client.Do(ctx, http.MethodGet, "/monitors/"+req.ID, nil, &out)
	if err != nil {
		if apiStatus(err) == http.StatusNotFound {
			return infer.ReadResponse[MonitorArgs, MonitorState]{}, nil
		}
		return infer.ReadResponse[MonitorArgs, MonitorState]{}, err
	}
	st := out.toState(req.Inputs)
	return infer.ReadResponse[MonitorArgs, MonitorState]{ID: req.ID, Inputs: st.MonitorArgs, State: st}, nil
}

func (Monitor) Update(ctx context.Context, req infer.UpdateRequest[MonitorArgs, MonitorState]) (infer.UpdateResponse[MonitorState], error) {
	in := req.Inputs
	in.TeamID = req.State.TeamID
	in.Type = req.State.Type
	if req.DryRun {
		out := req.State
		out.MonitorArgs = in
		return infer.UpdateResponse[MonitorState]{Output: out}, nil
	}

	body := monitorUpdateBody{
		URL: in.URL, Checks: in.Checks, FriendlyName: in.FriendlyName,
		GroupName: in.GroupName, Tags: in.Tags, Notes: in.Notes,
		Description: in.Description, RealIPAddress: in.RealIPAddress,
		monitorSettingsBody: settingsBodyOf(in),
	}
	var out monitorWire
	if err := infer.GetConfig[Config](ctx).client.Do(ctx, http.MethodPut, "/monitors/"+req.ID, body, &out); err != nil {
		return infer.UpdateResponse[MonitorState]{}, err
	}
	return infer.UpdateResponse[MonitorState]{Output: out.toState(in)}, nil
}

func (Monitor) Delete(ctx context.Context, req infer.DeleteRequest[MonitorState]) (infer.DeleteResponse, error) {
	err := infer.GetConfig[Config](ctx).client.Do(ctx, http.MethodDelete, "/monitors/"+req.ID, nil, nil)
	if err != nil && apiStatus(err) == http.StatusNotFound {
		return infer.DeleteResponse{}, nil
	}
	return infer.DeleteResponse{}, err
}
