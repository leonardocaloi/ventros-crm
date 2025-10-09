package pipeline

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// ScheduledRuleConfig configura regras com agendamento recorrente
type ScheduledRuleConfig struct {
	Type      ScheduleType `json:"type"`      // once, daily, weekly, monthly, cron
	CronExpr  string       `json:"cron_expr"` // expressão cron se Type = cron
	StartTime time.Time    `json:"start_time"`
	EndTime   *time.Time   `json:"end_time,omitempty"` // opcional, null = sem fim

	// Para tipos específicos
	DayOfWeek  *int `json:"day_of_week,omitempty"`  // 0-6 (domingo-sábado) para weekly
	DayOfMonth *int `json:"day_of_month,omitempty"` // 1-31 para monthly
	Hour       int  `json:"hour"`                   // hora do dia (0-23)
	Minute     int  `json:"minute"`                 // minuto (0-59)
}

// ScheduleType define o tipo de agendamento
type ScheduleType string

const (
	ScheduleOnce    ScheduleType = "once"    // executa uma vez em start_time
	ScheduleDaily   ScheduleType = "daily"   // executa diariamente em hour:minute
	ScheduleWeekly  ScheduleType = "weekly"  // executa semanalmente em day_of_week, hour:minute
	ScheduleMonthly ScheduleType = "monthly" // executa mensalmente em day_of_month, hour:minute
	ScheduleCron    ScheduleType = "cron"    // usa expressão cron customizada
)

// Validate valida a configuração do agendamento
func (s *ScheduledRuleConfig) Validate() error {
	if s.Type == "" {
		return errors.New("schedule type cannot be empty")
	}

	// Valida hour e minute
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
		// hour e minute já validados acima

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
		// TODO: Validar expressão cron com lib
		// ex: github.com/robfig/cron/v3

	default:
		return errors.New("invalid schedule type")
	}

	return nil
}

// ShouldRunNow verifica se regra deve executar agora
func (s *ScheduledRuleConfig) ShouldRunNow(now time.Time) bool {
	// Se tem end_time e já passou, não executa
	if s.EndTime != nil && now.After(*s.EndTime) {
		return false
	}

	switch s.Type {
	case ScheduleOnce:
		// Executa se estamos próximos do start_time (janela de 1 minuto)
		diff := now.Sub(s.StartTime)
		return diff >= 0 && diff < time.Minute

	case ScheduleDaily:
		// Verifica se hora/minuto batem (janela de 1 minuto)
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
		// TODO: Implementar parsing de cron expression
		// Por enquanto, retorna false
		return false

	default:
		return false
	}
}

// NextExecution calcula próxima execução
func (s *ScheduledRuleConfig) NextExecution(after time.Time) time.Time {
	switch s.Type {
	case ScheduleOnce:
		if after.Before(s.StartTime) {
			return s.StartTime
		}
		// Se já passou, retorna zero (não executa mais)
		return time.Time{}

	case ScheduleDaily:
		next := time.Date(after.Year(), after.Month(), after.Day(), s.Hour, s.Minute, 0, 0, after.Location())
		if next.Before(after) || next.Equal(after) {
			// Próximo dia
			next = next.Add(24 * time.Hour)
		}
		return next

	case ScheduleWeekly:
		if s.DayOfWeek == nil {
			return time.Time{}
		}

		// Encontra próximo dia da semana
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

		// Próxima semana
		return s.NextExecution(current)

	case ScheduleMonthly:
		if s.DayOfMonth == nil {
			return time.Time{}
		}

		// Tenta no mês atual
		next := time.Date(after.Year(), after.Month(), *s.DayOfMonth, s.Hour, s.Minute, 0, 0, after.Location())
		if next.After(after) {
			return next
		}

		// Próximo mês
		nextMonth := after.AddDate(0, 1, 0)
		return time.Date(nextMonth.Year(), nextMonth.Month(), *s.DayOfMonth, s.Hour, s.Minute, 0, 0, after.Location())

	case ScheduleCron:
		// TODO: Implementar parsing de cron
		return time.Time{}

	default:
		return time.Time{}
	}
}

// ScheduledAutomationRule regra com configuração de agendamento
type ScheduledAutomationRule struct {
	*Automation
	Schedule      ScheduledRuleConfig `json:"schedule"`
	LastExecuted  *time.Time          `json:"last_executed,omitempty"`
	NextExecution *time.Time          `json:"next_execution,omitempty"`
}

// NewScheduledAutomationRule cria regra agendada
func NewScheduledAutomationRule(
	pipelineID uuid.UUID,
	tenantID string,
	name string,
	schedule ScheduledRuleConfig,
) (*ScheduledAutomationRule, error) {
	// Valida schedule
	if err := schedule.Validate(); err != nil {
		return nil, err
	}

	// Cria regra base com trigger "scheduled"
	rule, err := NewAutomation(AutomationTypePipeline, tenantID, name, TriggerScheduled, &pipelineID)
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

// ReconstructScheduledAutomationRule reconstrói regra agendada
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

// MarkExecuted marca regra como executada e calcula próxima execução
func (r *ScheduledAutomationRule) MarkExecuted(executedAt time.Time) {
	r.LastExecuted = &executedAt
	next := r.Schedule.NextExecution(executedAt)
	if !next.IsZero() {
		r.NextExecution = &next
	} else {
		r.NextExecution = nil // não há próxima execução
	}
}

// IsReadyToExecute verifica se está pronto para executar agora
func (r *ScheduledAutomationRule) IsReadyToExecute(now time.Time) bool {
	if !r.IsEnabled() {
		return false
	}

	// Se tem próxima execução agendada, verifica se chegou a hora
	if r.NextExecution != nil {
		return now.After(*r.NextExecution) || now.Equal(*r.NextExecution)
	}

	// Fallback: usa lógica do schedule
	return r.Schedule.ShouldRunNow(now)
}
