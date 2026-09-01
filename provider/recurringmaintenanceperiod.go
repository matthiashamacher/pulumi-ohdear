package provider

import (
	"context"
	"net/http"
	"strconv"

	"github.com/pulumi/pulumi-go-provider/infer"
)

// RecurringMaintenancePeriod is an Oh Dear recurring maintenance schedule
// (https://ohdear.app/docs/api/maintenance-windows#recurring-maintenance-periods).
//
// It generates concrete maintenance periods on a daily, weekly or monthly
// cadence. One-off maintenance periods are transient events, not desired state,
// so only the recurring schedule is modelled here.
//
// ponytail: the generate_now / regenerate_future_periods / delete_future_periods
// backfill knobs are imperative side effects, not state — left at their API
// defaults.
type RecurringMaintenancePeriod struct{}

type RecurringMaintenancePeriodArgs struct {
	// monitorId is only accepted on create; a change replaces the resource.
	MonitorID      int    `pulumi:"monitorId" provider:"replaceOnChanges"`
	Name           string `pulumi:"name"`
	RecurrenceType string `pulumi:"recurrenceType"`
	StartTime      string `pulumi:"startTime"`
	EndTime        string `pulumi:"endTime"`
	DaysOfWeek     []int  `pulumi:"daysOfWeek,optional"`
	DayOfMonth     int    `pulumi:"dayOfMonth,optional"`
}

type RecurringMaintenancePeriodState struct {
	RecurringMaintenancePeriodArgs

	RecurringMaintenancePeriodID int    `pulumi:"recurringMaintenancePeriodId"`
	HumanReadableSchedule        string `pulumi:"humanReadableSchedule"`
	LastGeneratedUntil           string `pulumi:"lastGeneratedUntil"`
	CreatedAt                    string `pulumi:"createdAt"`
	UpdatedAt                    string `pulumi:"updatedAt"`
}

func (r *RecurringMaintenancePeriod) Annotate(a infer.Annotator) {
	a.Describe(&r, "An Oh Dear recurring maintenance schedule that generates maintenance periods on a cadence.")
}

func (a *RecurringMaintenancePeriodArgs) Annotate(an infer.Annotator) {
	an.Describe(&a.RecurrenceType, `One of "daily", "weekly" or "monthly".`)
	an.Describe(&a.StartTime, "Start of the window, `HH:MM`.")
	an.Describe(&a.EndTime, "End of the window, `HH:MM`.")
	an.Describe(&a.DaysOfWeek, "Days for a weekly schedule, 0 (Sunday) to 6 (Saturday).")
	an.Describe(&a.DayOfMonth, "Day for a monthly schedule, 1 to 31.")
}

type recurringWire struct {
	ID                    int    `json:"id"`
	MonitorID             int    `json:"monitor_id"`
	Name                  string `json:"name"`
	RecurrenceType        string `json:"recurrence_type"`
	DaysOfWeek            []int  `json:"days_of_week"`
	DayOfMonth            int    `json:"day_of_month"`
	StartTime             string `json:"start_time"`
	EndTime               string `json:"end_time"`
	HumanReadableSchedule string `json:"human_readable_schedule"`
	LastGeneratedUntil    string `json:"last_generated_until"`
	CreatedAt             string `json:"created_at"`
	UpdatedAt             string `json:"updated_at"`
}

func (w recurringWire) toState(inputs RecurringMaintenancePeriodArgs) RecurringMaintenancePeriodState {
	inputs.MonitorID = w.MonitorID
	inputs.Name = w.Name
	inputs.RecurrenceType = w.RecurrenceType
	inputs.StartTime = w.StartTime
	inputs.EndTime = w.EndTime
	inputs.DaysOfWeek = w.DaysOfWeek
	inputs.DayOfMonth = w.DayOfMonth
	return RecurringMaintenancePeriodState{
		RecurringMaintenancePeriodArgs: inputs,
		RecurringMaintenancePeriodID:   w.ID,
		HumanReadableSchedule:          w.HumanReadableSchedule,
		LastGeneratedUntil:             w.LastGeneratedUntil,
		CreatedAt:                      w.CreatedAt,
		UpdatedAt:                      w.UpdatedAt,
	}
}

// recurringBody carries the writable fields. omitempty drops the scheduling
// knobs that do not apply to the chosen recurrence_type.
type recurringBody struct {
	MonitorID      int    `json:"monitor_id,omitempty"`
	Name           string `json:"name"`
	RecurrenceType string `json:"recurrence_type"`
	StartTime      string `json:"start_time"`
	EndTime        string `json:"end_time"`
	DaysOfWeek     []int  `json:"days_of_week,omitempty"`
	DayOfMonth     int    `json:"day_of_month,omitempty"`
}

func recurringBodyOf(in RecurringMaintenancePeriodArgs, withMonitor bool) recurringBody {
	b := recurringBody{
		Name: in.Name, RecurrenceType: in.RecurrenceType,
		StartTime: in.StartTime, EndTime: in.EndTime,
		DaysOfWeek: in.DaysOfWeek, DayOfMonth: in.DayOfMonth,
	}
	if withMonitor {
		b.MonitorID = in.MonitorID
	}
	return b
}

func (RecurringMaintenancePeriod) Create(ctx context.Context, req infer.CreateRequest[RecurringMaintenancePeriodArgs]) (infer.CreateResponse[RecurringMaintenancePeriodState], error) {
	in := req.Inputs
	if req.DryRun {
		return infer.CreateResponse[RecurringMaintenancePeriodState]{Output: RecurringMaintenancePeriodState{RecurringMaintenancePeriodArgs: in}}, nil
	}

	var out recurringWire
	if err := infer.GetConfig[Config](ctx).client.Do(ctx, http.MethodPost, "/recurring-maintenance-periods", recurringBodyOf(in, true), &out); err != nil {
		return infer.CreateResponse[RecurringMaintenancePeriodState]{}, err
	}
	return infer.CreateResponse[RecurringMaintenancePeriodState]{ID: strconv.Itoa(out.ID), Output: out.toState(in)}, nil
}

func (RecurringMaintenancePeriod) Read(ctx context.Context, req infer.ReadRequest[RecurringMaintenancePeriodArgs, RecurringMaintenancePeriodState]) (infer.ReadResponse[RecurringMaintenancePeriodArgs, RecurringMaintenancePeriodState], error) {
	var out recurringWire
	err := infer.GetConfig[Config](ctx).client.Do(ctx, http.MethodGet, "/recurring-maintenance-periods/"+req.ID, nil, &out)
	if err != nil {
		if apiStatus(err) == http.StatusNotFound {
			return infer.ReadResponse[RecurringMaintenancePeriodArgs, RecurringMaintenancePeriodState]{}, nil
		}
		return infer.ReadResponse[RecurringMaintenancePeriodArgs, RecurringMaintenancePeriodState]{}, err
	}
	st := out.toState(req.Inputs)
	return infer.ReadResponse[RecurringMaintenancePeriodArgs, RecurringMaintenancePeriodState]{ID: req.ID, Inputs: st.RecurringMaintenancePeriodArgs, State: st}, nil
}

func (RecurringMaintenancePeriod) Update(ctx context.Context, req infer.UpdateRequest[RecurringMaintenancePeriodArgs, RecurringMaintenancePeriodState]) (infer.UpdateResponse[RecurringMaintenancePeriodState], error) {
	in := req.Inputs
	in.MonitorID = req.State.MonitorID
	if req.DryRun {
		out := req.State
		out.RecurringMaintenancePeriodArgs = in
		return infer.UpdateResponse[RecurringMaintenancePeriodState]{Output: out}, nil
	}

	var out recurringWire
	if err := infer.GetConfig[Config](ctx).client.Do(ctx, http.MethodPut, "/recurring-maintenance-periods/"+req.ID, recurringBodyOf(in, false), &out); err != nil {
		return infer.UpdateResponse[RecurringMaintenancePeriodState]{}, err
	}
	return infer.UpdateResponse[RecurringMaintenancePeriodState]{Output: out.toState(in)}, nil
}

func (RecurringMaintenancePeriod) Delete(ctx context.Context, req infer.DeleteRequest[RecurringMaintenancePeriodState]) (infer.DeleteResponse, error) {
	err := infer.GetConfig[Config](ctx).client.Do(ctx, http.MethodDelete, "/recurring-maintenance-periods/"+req.ID, nil, nil)
	if err != nil && apiStatus(err) == http.StatusNotFound {
		return infer.DeleteResponse{}, nil
	}
	return infer.DeleteResponse{}, err
}
