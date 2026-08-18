package repository

import (
	"context"
	"fmt"
	"time"

	"gcet-web-backend/internal/config"
	"gcet-web-backend/internal/logger"
	"gcet-web-backend/internal/models"

	"github.com/supabase-community/supabase-go"
)

// SupabaseRepo handles interactions with the Supabase database
type SupabaseRepo struct {
	client *supabase.Client
}

// NewSupabaseRepo creates a new Supabase repository
func NewSupabaseRepo(cfg *config.Config) (*SupabaseRepo, error) {
	// Determine which key to use: prefer service key, fall back to anon key
	apiKey := cfg.SupabaseServiceKey
	if apiKey == "" {
		apiKey = cfg.SupabaseAnonKey
	}

	if cfg.SupabaseURL == "" || apiKey == "" {
		return nil, fmt.Errorf("supabase URL and API key (service or anon) are required")
	}

	if cfg.SupabaseServiceKey == "" {
		logger.Log.Warn("SUPABASE_SERVICE_KEY not set, falling back to SUPABASE_ANON_KEY — ensure RLS policies allow writes")
	}

	client, err := supabase.NewClient(cfg.SupabaseURL, apiKey, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create supabase client: %w", err)
	}

	return &SupabaseRepo{
		client: client,
	}, nil
}

// UpsertFaculty updates or inserts a faculty record
func (r *SupabaseRepo) UpsertFaculty(ctx context.Context, faculty *models.Faculty, password string) error {
	data := map[string]interface{}{
		"employee_id": faculty.EmployeeID,
		"username":    faculty.LoginID,
		"password":    password,
		"full_name":   faculty.Name,
		"department":  faculty.Department,
		"updated_at":  time.Now(),
	}

	res, count, err := r.client.From("faculty").Upsert(data, "", "", "").Execute()
	if err != nil {
		return fmt.Errorf("failed to upsert faculty: %w", err)
	}

	logger.Log.Debugf("Upserted faculty %s, res: %s, count: %d", faculty.EmployeeID, string(res), count)
	return nil
}

// SubmitFeedback inserts a new feedback record
func (r *SupabaseRepo) SubmitFeedback(ctx context.Context, feedback *models.FacultyFeedback) error {
	data := map[string]interface{}{
		"faculty_name": feedback.FacultyName,
		"employee_id":  feedback.EmployeeID,
		"department":   feedback.Department,
		"type":         feedback.Type,
		"subject":      feedback.Subject,
		"description":  feedback.Description,
		"status":       "Pending",
	}

	_, _, err := r.client.From("feedback").Insert(data, false, "", "", "").Execute()
	if err != nil {
		return fmt.Errorf("failed to insert feedback: %w", err)
	}
	return nil
}

// Note: Dynamic attendance RPCs will be added here based on the Flutter SupabaseService

// GetTimetableModifications gets custom and hidden entries for an employee
func (r *SupabaseRepo) GetTimetableModifications(ctx context.Context, employeeID string) ([]models.CustomTimetableEntry, error) {
	var entries []models.CustomTimetableEntry
	_, err := r.client.From("timetable_entries").Select("*", "exact", false).Eq("employee_id", employeeID).ExecuteTo(&entries)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch timetable modifications: %w", err)
	}
	return entries, nil
}

// AddTimetableEntry adds a custom or hidden timetable entry
func (r *SupabaseRepo) AddTimetableEntry(ctx context.Context, entry *models.CustomTimetableEntry) error {
	data := map[string]interface{}{
		"employee_id": entry.EmployeeID,
		"day":         entry.Day,
		"time":        entry.Time,
		"subject":     entry.Subject,
		"class_type":  entry.ClassType,
		"room":        entry.Room,
		"batch":       entry.Batch,
		"is_custom":   entry.IsCustom,
		"is_hidden":   entry.IsHidden,
	}

	_, _, err := r.client.From("timetable_entries").Insert(data, false, "", "", "").Execute()
	if err != nil {
		return fmt.Errorf("failed to insert timetable entry: %w", err)
	}
	return nil
}

// DeleteCustomTimetableEntry removes a specific custom entry
func (r *SupabaseRepo) DeleteCustomTimetableEntry(ctx context.Context, employeeID, day, timeStr, subject string) error {
	_, _, err := r.client.From("timetable_entries").
		Delete("", "").
		Eq("employee_id", employeeID).
		Eq("day", day).
		Eq("time", timeStr).
		Eq("subject", subject).
		Eq("is_custom", "true").
		Execute()
	if err != nil {
		return fmt.Errorf("failed to delete timetable entry: %w", err)
	}
	return nil
}

// ResetTimetable deletes all custom/hidden entries for an employee
func (r *SupabaseRepo) ResetTimetable(ctx context.Context, employeeID string) error {
	_, _, err := r.client.From("timetable_entries").
		Delete("", "").
		Eq("employee_id", employeeID).
		Execute()
	if err != nil {
		return fmt.Errorf("failed to reset timetable: %w", err)
	}
	return nil
}
