package pipeline

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type ScheduledRuleConfig struct {
	Type      ScheduleType `json:"type"`
	CronExpr  string       `json:"cron_expr"`
	StartTime time.Time    `json:"start_time"`
	EndTime   *time.Time   `json:"end_time,omitempty"`

	DayOfWeek  *int `json:"day_of_week,omitempty"`
	DayOfMonth *int `json:"day_of_month,omitempty"`
	Hour       int  `json:"hour"`
	Minute     int  `json:"minute"`
}

type ScheduleType string

const (
	ScheduleOnce    ScheduleType = "once"
	ScheduleDaily   ScheduleType = "daily"
	ScheduleWeekly  ScheduleType = "weekly"
	ScheduleMonthly ScheduleType = "monthly"
	ScheduleCron    ScheduleType = "cron"
)

func (s *ScheduledRuleConfig) Validate() error {
	if s.Type == "" {
		return errors.New("schedule type cannot be empty")
	}

	if s.Hour < 0 || s.Hour > 23 {
		return errors.New("hour must be between 0 and 23")
	}
	if s.Minute < 0 || s.Minute > 59 {
		return errors.New("minute must be between 0 and 59")
	}

	switch s.Type {
	case ScheduleOnce:
		if s.StartTime.IsZero() {
			return errors.New("start_time is required for 'once' schedule")
		}

	case ScheduleDaily:

	case ScheduleWeekly:
		if s.DayOfWeek == nil {
			return errors.New("day_of_week is required for 'weekly' schedule")
		}
		if *s.DayOfWeek < 0 || *s.DayOfWeek > 6 {
			return errors.New("day_of_week must be between 0 (Sunday) and 6 (Saturday)")
		}

	case ScheduleMonthly:
		if s.DayOfMonth == nil {
			return errors.New("day_of_month is required for 'monthly' schedule")
		}
		if *s.DayOfMonth < 1 || *s.DayOfMonth > 31 {
			return errors.New("day_of_month must be between 1 and 31")
		}

	case ScheduleCron:
		if s.CronExpr == "" {
			return errors.New("cron_expr is required for 'cron' schedule")
		}

	default:
		return errors.New("invalid schedule type")
	}

	return nil
}

func (s *ScheduledRuleConfig) ShouldRunNow(now time.Time) bool {

	if s.EndTime != nil && now.After(*s.EndTime) {
		return false
	}

	switch s.Type {
	case ScheduleOnce:
		diff := now.Sub(s.StartTime)
		return diff >= 0 && diff < time.Minute

	case ScheduleDaily:
		return now.Hour() == s.Hour && now.Minute() == s.Minute

	case ScheduleWeekly:
		if s.DayOfWeek == nil {
			return false
		}
		return int(now.Weekday()) == *s.DayOfWeek &&
			now.Hour() == s.Hour &&
			now.Minute() == s.Minute

	case ScheduleMonthly:
		if s.DayOfMonth == nil {
			return false
		}
		return now.Day() == *s.DayOfMonth &&
			now.Hour() == s.Hour &&
			now.Minute() == s.Minute

	case ScheduleCron:

		return false

	default:
		return false
	}
}

func (s *ScheduledRuleConfig) NextExecution(after time.Time) time.Time {
	switch s.Type {
	case ScheduleOnce:
		if after.Before(s.StartTime) {
			return s.StartTime
		}

		return time.Time{}

	case ScheduleDaily:
		next := time.Date(after.Year(), after.Month(), after.Day(), s.Hour, s.Minute, 0, 0, after.Location())
		if next.Before(after) || next.Equal(after) {

			next = next.Add(24 * time.Hour)
		}
		return next

	case ScheduleWeekly:
		if s.DayOfWeek == nil {
			return time.Time{}
		}

		current := after
		for i := 0; i < 7; i++ {
			if int(current.Weekday()) == *s.DayOfWeek {
				next := time.Date(current.Year(), current.Month(), current.Day(), s.Hour, s.Minute, 0, 0, current.Location())
				if next.After(after) {
					return next
				}
			}
			current = current.Add(24 * time.Hour)
		}

		return s.NextExecution(current)

	case ScheduleMonthly:
		if s.DayOfMonth == nil {
			return time.Time{}
		}

		next := time.Date(after.Year(), after.Month(), *s.DayOfMonth, s.Hour, s.Minute, 0, 0, after.Location())
		if next.After(after) {
			return next
		}

		nextMonth := after.AddDate(0, 1, 0)
		return time.Date(nextMonth.Year(), nextMonth.Month(), *s.DayOfMonth, s.Hour, s.Minute, 0, 0, after.Location())

	case ScheduleCron:

		return time.Time{}

	default:
		return time.Time{}
	}
}

type ScheduledAutomationRule struct {
	*Automation
	Schedule      ScheduledRuleConfig `json:"schedule"`
	LastExecuted  *time.Time          `json:"last_executed,omitempty"`
	NextExecution *time.Time          `json:"next_execution,omitempty"`
}

func NewScheduledAutomationRule(
	pipelineID uuid.UUID,
	tenantID string,
	name string,
	schedule ScheduledRuleConfig,
) (*ScheduledAutomationRule, error) {
	if err := schedule.Validate(); err != nil {
		return nil, err
	}

	rule, err := NewAutomation(AutomationTypePipelineBased, tenantID, name, TriggerScheduled, &pipelineID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	nextExec := schedule.NextExecution(now)

	return &ScheduledAutomationRule{
		Automation:    rule,
		Schedule:      schedule,
		LastExecuted:  nil,
		NextExecution: &nextExec,
	}, nil
}

func ReconstructScheduledAutomationRule(
	rule *Automation,
	schedule ScheduledRuleConfig,
	lastExecuted *time.Time,
	nextExecution *time.Time,
) *ScheduledAutomationRule {
	return &ScheduledAutomationRule{
		Automation:    rule,
		Schedule:      schedule,
		LastExecuted:  lastExecuted,
		NextExecution: nextExecution,
	}
}

func (r *ScheduledAutomationRule) MarkExecuted(executedAt time.Time) {
	r.LastExecuted = &executedAt
	next := r.Schedule.NextExecution(executedAt)
	if !next.IsZero() {
		r.NextExecution = &next
	} else {
		r.NextExecution = nil
	}
}

func (r *ScheduledAutomationRule) IsReadyToExecute(now time.Time) bool {
	if !r.IsEnabled() {
		return false
	}

	if r.NextExecution != nil {
		return now.After(*r.NextExecution) || now.Equal(*r.NextExecution)
	}

	return r.Schedule.ShouldRunNow(now)
}
