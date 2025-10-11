package pipeline

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create int pointer
func intPtr2(i int) *int {
	return &i
}

func TestScheduledRuleConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  ScheduledRuleConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid once schedule",
			config: ScheduledRuleConfig{
				Type:      ScheduleOnce,
				StartTime: time.Now().Add(time.Hour),
				Hour:      10,
				Minute:    30,
			},
			wantErr: false,
		},
		{
			name: "valid daily schedule",
			config: ScheduledRuleConfig{
				Type:   ScheduleDaily,
				Hour:   14,
				Minute: 0,
			},
			wantErr: false,
		},
		{
			name: "valid weekly schedule",
			config: ScheduledRuleConfig{
				Type:      ScheduleWeekly,
				DayOfWeek: intPtr2(1), // Monday
				Hour:      9,
				Minute:    0,
			},
			wantErr: false,
		},
		{
			name: "valid monthly schedule",
			config: ScheduledRuleConfig{
				Type:       ScheduleMonthly,
				DayOfMonth: intPtr2(15),
				Hour:       12,
				Minute:     0,
			},
			wantErr: false,
		},
		{
			name: "valid cron schedule",
			config: ScheduledRuleConfig{
				Type:     ScheduleCron,
				CronExpr: "0 0 * * *",
				Hour:     0,
				Minute:   0,
			},
			wantErr: false,
		},
		{
			name: "empty type",
			config: ScheduledRuleConfig{
				Hour:   10,
				Minute: 0,
			},
			wantErr: true,
			errMsg:  "schedule type cannot be empty",
		},
		{
			name: "invalid hour - too low",
			config: ScheduledRuleConfig{
				Type:   ScheduleDaily,
				Hour:   -1,
				Minute: 0,
			},
			wantErr: true,
			errMsg:  "hour must be between 0 and 23",
		},
		{
			name: "invalid hour - too high",
			config: ScheduledRuleConfig{
				Type:   ScheduleDaily,
				Hour:   24,
				Minute: 0,
			},
			wantErr: true,
			errMsg:  "hour must be between 0 and 23",
		},
		{
			name: "invalid minute - too low",
			config: ScheduledRuleConfig{
				Type:   ScheduleDaily,
				Hour:   10,
				Minute: -1,
			},
			wantErr: true,
			errMsg:  "minute must be between 0 and 59",
		},
		{
			name: "invalid minute - too high",
			config: ScheduledRuleConfig{
				Type:   ScheduleDaily,
				Hour:   10,
				Minute: 60,
			},
			wantErr: true,
			errMsg:  "minute must be between 0 and 59",
		},
		{
			name: "once schedule without start_time",
			config: ScheduledRuleConfig{
				Type:   ScheduleOnce,
				Hour:   10,
				Minute: 0,
			},
			wantErr: true,
			errMsg:  "start_time is required",
		},
		{
			name: "weekly schedule without day_of_week",
			config: ScheduledRuleConfig{
				Type:   ScheduleWeekly,
				Hour:   10,
				Minute: 0,
			},
			wantErr: true,
			errMsg:  "day_of_week is required",
		},
		{
			name: "weekly schedule with invalid day_of_week",
			config: ScheduledRuleConfig{
				Type:      ScheduleWeekly,
				DayOfWeek: intPtr2(7),
				Hour:      10,
				Minute:    0,
			},
			wantErr: true,
			errMsg:  "day_of_week must be between 0",
		},
		{
			name: "monthly schedule without day_of_month",
			config: ScheduledRuleConfig{
				Type:   ScheduleMonthly,
				Hour:   10,
				Minute: 0,
			},
			wantErr: true,
			errMsg:  "day_of_month is required",
		},
		{
			name: "monthly schedule with invalid day_of_month - too low",
			config: ScheduledRuleConfig{
				Type:       ScheduleMonthly,
				DayOfMonth: intPtr2(0),
				Hour:       10,
				Minute:     0,
			},
			wantErr: true,
			errMsg:  "day_of_month must be between 1 and 31",
		},
		{
			name: "monthly schedule with invalid day_of_month - too high",
			config: ScheduledRuleConfig{
				Type:       ScheduleMonthly,
				DayOfMonth: intPtr2(32),
				Hour:       10,
				Minute:     0,
			},
			wantErr: true,
			errMsg:  "day_of_month must be between 1 and 31",
		},
		{
			name: "cron schedule without cron_expr",
			config: ScheduledRuleConfig{
				Type:   ScheduleCron,
				Hour:   10,
				Minute: 0,
			},
			wantErr: true,
			errMsg:  "cron_expr is required",
		},
		{
			name: "invalid schedule type",
			config: ScheduledRuleConfig{
				Type:   "invalid",
				Hour:   10,
				Minute: 0,
			},
			wantErr: true,
			errMsg:  "invalid schedule type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestScheduledRuleConfig_ShouldRunNow(t *testing.T) {
	t.Run("once schedule - should run", func(t *testing.T) {
		startTime := time.Now().Add(-30 * time.Second)
		config := ScheduledRuleConfig{
			Type:      ScheduleOnce,
			StartTime: startTime,
			Hour:      10,
			Minute:    0,
		}

		assert.True(t, config.ShouldRunNow(time.Now()))
	})

	t.Run("once schedule - should not run (too late)", func(t *testing.T) {
		startTime := time.Now().Add(-2 * time.Minute)
		config := ScheduledRuleConfig{
			Type:      ScheduleOnce,
			StartTime: startTime,
			Hour:      10,
			Minute:    0,
		}

		assert.False(t, config.ShouldRunNow(time.Now()))
	})

	t.Run("once schedule - after end time", func(t *testing.T) {
		startTime := time.Now().Add(-30 * time.Second)
		endTime := time.Now().Add(-10 * time.Second)
		config := ScheduledRuleConfig{
			Type:      ScheduleOnce,
			StartTime: startTime,
			EndTime:   &endTime,
			Hour:      10,
			Minute:    0,
		}

		assert.False(t, config.ShouldRunNow(time.Now()))
	})

	t.Run("daily schedule - matches hour and minute", func(t *testing.T) {
		now := time.Now()
		config := ScheduledRuleConfig{
			Type:   ScheduleDaily,
			Hour:   now.Hour(),
			Minute: now.Minute(),
		}

		assert.True(t, config.ShouldRunNow(now))
	})

	t.Run("daily schedule - different hour", func(t *testing.T) {
		now := time.Now()
		config := ScheduledRuleConfig{
			Type:   ScheduleDaily,
			Hour:   (now.Hour() + 1) % 24,
			Minute: now.Minute(),
		}

		assert.False(t, config.ShouldRunNow(now))
	})

	t.Run("weekly schedule - matches day, hour, minute", func(t *testing.T) {
		now := time.Now()
		dayOfWeek := int(now.Weekday())
		config := ScheduledRuleConfig{
			Type:      ScheduleWeekly,
			DayOfWeek: &dayOfWeek,
			Hour:      now.Hour(),
			Minute:    now.Minute(),
		}

		assert.True(t, config.ShouldRunNow(now))
	})

	t.Run("weekly schedule - wrong day", func(t *testing.T) {
		now := time.Now()
		wrongDay := (int(now.Weekday()) + 1) % 7
		config := ScheduledRuleConfig{
			Type:      ScheduleWeekly,
			DayOfWeek: &wrongDay,
			Hour:      now.Hour(),
			Minute:    now.Minute(),
		}

		assert.False(t, config.ShouldRunNow(now))
	})

	t.Run("monthly schedule - matches day, hour, minute", func(t *testing.T) {
		now := time.Now()
		dayOfMonth := now.Day()
		config := ScheduledRuleConfig{
			Type:       ScheduleMonthly,
			DayOfMonth: &dayOfMonth,
			Hour:       now.Hour(),
			Minute:     now.Minute(),
		}

		assert.True(t, config.ShouldRunNow(now))
	})

	t.Run("monthly schedule - wrong day", func(t *testing.T) {
		now := time.Now()
		wrongDay := (now.Day() % 28) + 1 // Different day
		config := ScheduledRuleConfig{
			Type:       ScheduleMonthly,
			DayOfMonth: &wrongDay,
			Hour:       now.Hour(),
			Minute:     now.Minute(),
		}

		assert.False(t, config.ShouldRunNow(now))
	})

	t.Run("cron schedule - not implemented", func(t *testing.T) {
		config := ScheduledRuleConfig{
			Type:     ScheduleCron,
			CronExpr: "0 0 * * *",
			Hour:     0,
			Minute:   0,
		}

		assert.False(t, config.ShouldRunNow(time.Now()))
	})
}

func TestScheduledRuleConfig_NextExecution(t *testing.T) {
	t.Run("once schedule - before start time", func(t *testing.T) {
		startTime := time.Now().Add(2 * time.Hour)
		config := ScheduledRuleConfig{
			Type:      ScheduleOnce,
			StartTime: startTime,
			Hour:      10,
			Minute:    0,
		}

		next := config.NextExecution(time.Now())
		assert.Equal(t, startTime, next)
	})

	t.Run("once schedule - after start time", func(t *testing.T) {
		startTime := time.Now().Add(-2 * time.Hour)
		config := ScheduledRuleConfig{
			Type:      ScheduleOnce,
			StartTime: startTime,
			Hour:      10,
			Minute:    0,
		}

		next := config.NextExecution(time.Now())
		assert.True(t, next.IsZero())
	})

	t.Run("daily schedule - next day", func(t *testing.T) {
		now := time.Now()
		config := ScheduledRuleConfig{
			Type:   ScheduleDaily,
			Hour:   10,
			Minute: 30,
		}

		next := config.NextExecution(now)
		assert.False(t, next.IsZero())
		assert.Equal(t, 10, next.Hour())
		assert.Equal(t, 30, next.Minute())
	})

	t.Run("weekly schedule", func(t *testing.T) {
		now := time.Now()
		dayOfWeek := (int(now.Weekday()) + 1) % 7
		config := ScheduledRuleConfig{
			Type:      ScheduleWeekly,
			DayOfWeek: &dayOfWeek,
			Hour:      10,
			Minute:    30,
		}

		next := config.NextExecution(now)
		assert.False(t, next.IsZero())
		assert.Equal(t, dayOfWeek, int(next.Weekday()))
		assert.Equal(t, 10, next.Hour())
		assert.Equal(t, 30, next.Minute())
	})

	t.Run("monthly schedule - current month", func(t *testing.T) {
		now := time.Now()
		futureDay := now.Day() + 5
		if futureDay > 28 {
			futureDay = 1
		}
		config := ScheduledRuleConfig{
			Type:       ScheduleMonthly,
			DayOfMonth: &futureDay,
			Hour:       10,
			Minute:     30,
		}

		next := config.NextExecution(now)
		assert.False(t, next.IsZero())
		assert.Equal(t, 10, next.Hour())
		assert.Equal(t, 30, next.Minute())
	})

	t.Run("cron schedule - not implemented", func(t *testing.T) {
		config := ScheduledRuleConfig{
			Type:     ScheduleCron,
			CronExpr: "0 0 * * *",
			Hour:     0,
			Minute:   0,
		}

		next := config.NextExecution(time.Now())
		assert.True(t, next.IsZero())
	})
}

func TestNewScheduledAutomationRule(t *testing.T) {
	pipelineID := uuid.New()
	tenantID := "tenant-123"

	t.Run("valid scheduled rule", func(t *testing.T) {
		schedule := ScheduledRuleConfig{
			Type:   ScheduleDaily,
			Hour:   14,
			Minute: 30,
		}

		rule, err := NewScheduledAutomationRule(pipelineID, tenantID, "Daily Report", schedule)
		require.NoError(t, err)
		require.NotNil(t, rule)

		assert.NotNil(t, rule.Automation)
		assert.Equal(t, schedule.Type, rule.Schedule.Type)
		assert.Nil(t, rule.LastExecuted)
		assert.NotNil(t, rule.NextExecution)
	})

	t.Run("invalid schedule", func(t *testing.T) {
		schedule := ScheduledRuleConfig{
			Type:   ScheduleWeekly,
			Hour:   10,
			Minute: 0,
			// Missing DayOfWeek
		}

		rule, err := NewScheduledAutomationRule(pipelineID, tenantID, "Weekly Report", schedule)
		require.Error(t, err)
		assert.Nil(t, rule)
		assert.Contains(t, err.Error(), "day_of_week is required")
	})
}

func TestReconstructScheduledAutomationRule(t *testing.T) {
	pipelineID := uuid.New()
	baseRule, err := NewAutomation(AutomationTypePipelineBased, "tenant-123", "Test Rule", TriggerScheduled, &pipelineID)
	require.NoError(t, err)

	schedule := ScheduledRuleConfig{
		Type:   ScheduleDaily,
		Hour:   10,
		Minute: 0,
	}

	lastExecuted := time.Now().Add(-24 * time.Hour)
	nextExecution := time.Now().Add(1 * time.Hour)

	rule := ReconstructScheduledAutomationRule(baseRule, schedule, &lastExecuted, &nextExecution)

	assert.NotNil(t, rule)
	assert.Equal(t, baseRule, rule.Automation)
	assert.Equal(t, schedule.Type, rule.Schedule.Type)
	assert.Equal(t, lastExecuted, *rule.LastExecuted)
	assert.Equal(t, nextExecution, *rule.NextExecution)
}

func TestScheduledAutomationRule_MarkExecuted(t *testing.T) {
	pipelineID := uuid.New()
	schedule := ScheduledRuleConfig{
		Type:   ScheduleDaily,
		Hour:   14,
		Minute: 30,
	}

	rule, err := NewScheduledAutomationRule(pipelineID, "tenant-123", "Daily Rule", schedule)
	require.NoError(t, err)

	executedAt := time.Now()
	rule.MarkExecuted(executedAt)

	assert.NotNil(t, rule.LastExecuted)
	assert.Equal(t, executedAt, *rule.LastExecuted)
	assert.NotNil(t, rule.NextExecution)
	assert.True(t, rule.NextExecution.After(executedAt))
}

func TestScheduledAutomationRule_MarkExecuted_OnceSchedule(t *testing.T) {
	pipelineID := uuid.New()
	startTime := time.Now().Add(-1 * time.Hour)
	schedule := ScheduledRuleConfig{
		Type:      ScheduleOnce,
		StartTime: startTime,
		Hour:      10,
		Minute:    0,
	}

	baseRule, err := NewAutomation(AutomationTypePipelineBased, "tenant-123", "Once Rule", TriggerScheduled, &pipelineID)
	require.NoError(t, err)

	rule := ReconstructScheduledAutomationRule(baseRule, schedule, nil, nil)

	executedAt := time.Now()
	rule.MarkExecuted(executedAt)

	assert.NotNil(t, rule.LastExecuted)
	assert.Nil(t, rule.NextExecution) // Once schedule has no next execution
}

func TestScheduledAutomationRule_IsReadyToExecute(t *testing.T) {
	pipelineID := uuid.New()

	t.Run("ready - next execution time reached", func(t *testing.T) {
		schedule := ScheduledRuleConfig{
			Type:   ScheduleDaily,
			Hour:   10,
			Minute: 0,
		}

		rule, err := NewScheduledAutomationRule(pipelineID, "tenant-123", "Test", schedule)
		require.NoError(t, err)

		pastTime := time.Now().Add(-1 * time.Hour)
		rule.NextExecution = &pastTime

		assert.True(t, rule.IsReadyToExecute(time.Now()))
	})

	t.Run("not ready - next execution in future", func(t *testing.T) {
		schedule := ScheduledRuleConfig{
			Type:   ScheduleDaily,
			Hour:   10,
			Minute: 0,
		}

		rule, err := NewScheduledAutomationRule(pipelineID, "tenant-123", "Test", schedule)
		require.NoError(t, err)

		futureTime := time.Now().Add(2 * time.Hour)
		rule.NextExecution = &futureTime

		assert.False(t, rule.IsReadyToExecute(time.Now()))
	})

	t.Run("not ready - rule disabled", func(t *testing.T) {
		schedule := ScheduledRuleConfig{
			Type:   ScheduleDaily,
			Hour:   10,
			Minute: 0,
		}

		rule, err := NewScheduledAutomationRule(pipelineID, "tenant-123", "Test", schedule)
		require.NoError(t, err)

		rule.Disable()

		assert.False(t, rule.IsReadyToExecute(time.Now()))
	})

	t.Run("fallback to schedule logic", func(t *testing.T) {
		now := time.Now()
		schedule := ScheduledRuleConfig{
			Type:   ScheduleDaily,
			Hour:   now.Hour(),
			Minute: now.Minute(),
		}

		rule, err := NewScheduledAutomationRule(pipelineID, "tenant-123", "Test", schedule)
		require.NoError(t, err)

		rule.NextExecution = nil

		assert.True(t, rule.IsReadyToExecute(now))
	})
}

func TestScheduleType_Constants(t *testing.T) {
	assert.Equal(t, ScheduleType("once"), ScheduleOnce)
	assert.Equal(t, ScheduleType("daily"), ScheduleDaily)
	assert.Equal(t, ScheduleType("weekly"), ScheduleWeekly)
	assert.Equal(t, ScheduleType("monthly"), ScheduleMonthly)
	assert.Equal(t, ScheduleType("cron"), ScheduleCron)
}
