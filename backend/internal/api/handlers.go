package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"gcet-web-backend/internal/logger"
	"gcet-web-backend/internal/middleware"
	"gcet-web-backend/internal/models"
	"gcet-web-backend/internal/repository"
	"gcet-web-backend/internal/service"
)

// Handler holds dependencies for HTTP handlers
type Handler struct {
	authSvc *service.AuthService
	repo    *repository.SupabaseRepo
}

// NewHandler creates a new handler
func NewHandler(authSvc *service.AuthService, repo *repository.SupabaseRepo) *Handler {
	return &Handler{
		authSvc: authSvc,
		repo:    repo,
	}
}

// Helper to write JSON responses
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// Helper to write JSON error
func writeError(w http.ResponseWriter, status int, err string) {
	writeJSON(w, status, models.APIResponse{
		Success: false,
		Error:   err,
	})
}

// --- Auth Endpoints ---

// Login handles faculty login
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if req.LoginID == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "Login ID and Password are required")
		return
	}

	resp, err := h.authSvc.Login(req.LoginID, req.Password)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}

	// Sync to Supabase in background
	if resp.Faculty != nil {
		go func(faculty *models.Faculty, password string) {
			if h.repo != nil {
				if err := h.repo.UpsertFaculty(context.Background(), faculty, password); err != nil {
					logger.Log.Errorf("Failed to upsert faculty %s to Supabase: %v", faculty.EmployeeID, err)
				} else {
					logger.Log.Infof("Successfully synced faculty %s to Supabase", faculty.EmployeeID)
				}
			} else {
				logger.Log.Warn("Supabase repo is nil — faculty not synced to database")
			}
		}(resp.Faculty, req.Password)
	}

	// Set auth cookies for frontend (HTTP only)
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    resp.Token,
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Secure:   true, // Set to false if local dev without HTTPS
		SameSite: http.SameSiteStrictMode,
	})

	writeJSON(w, http.StatusOK, resp)
}

// Logout handles faculty logout
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims != nil {
		h.authSvc.Logout(claims.SessionID)
	}

	// Clear cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Message: "Logged out successfully"})
}

// RefreshToken handles JWT refresh
func (h *Handler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	resp, err := h.authSvc.RefreshToken(req.RefreshToken)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    resp.Token,
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})

	writeJSON(w, http.StatusOK, resp)
}

// --- Profile Endpoints ---

// GetProfile returns the faculty profile
func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	session := middleware.GetSession(r)
	if session == nil || session.Faculty == nil {
		writeError(w, http.StatusUnauthorized, "Session invalid")
		return
	}
	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: session.Faculty})
}

// --- Scraper Proxy Endpoints (Attendance) ---
// Examples for wrapping the scraper service

// GetAttendanceSheets wraps ScrapeAttendanceSheets
func (h *Handler) GetAttendanceSheets(w http.ResponseWriter, r *http.Request) {
	session := middleware.GetSession(r)
	sheets, err := session.Scraper.ScrapeAttendanceSheets()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: sheets})
}

// GetAttendanceCourses wraps FetchAttendanceCourses
func (h *Handler) GetAttendanceCourses(w http.ResponseWriter, r *http.Request) {
	session := middleware.GetSession(r)
	courses, err := session.Scraper.FetchAttendanceCourses()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: courses})
}

// GetStudentList wraps FetchStudentList
func (h *Handler) GetStudentList(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CourseCode  string                     `json:"course_code"`
		Date        string                     `json:"date"`
		ByLibID     bool                       `json:"by_lib_id"`
		IsEdit      bool                       `json:"is_edit"`
		Metadata    *models.AttendanceMetadata `json:"metadata"`
		OptionIndex int                        `json:"option_index"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	session := middleware.GetSession(r)
	students, err := session.Scraper.FetchStudentList(req.CourseCode, req.Date, req.ByLibID, req.IsEdit, req.Metadata, req.OptionIndex)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: students})
}

// SubmitAttendance wraps SubmitAttendance
func (h *Handler) SubmitAttendance(w http.ResponseWriter, r *http.Request) {
	var req models.SubmitAttendanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	session := middleware.GetSession(r)
	success, err := session.Scraper.SubmitAttendance(&req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, models.APIResponse{Success: success})
}

// GetAverageAttendance wraps ScrapeAverageAttendance
func (h *Handler) GetAverageAttendance(w http.ResponseWriter, r *http.Request) {
	session := middleware.GetSession(r)
	avgList, err := session.Scraper.ScrapeAverageAttendance()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: avgList})
}

// GetStudentWiseAttendance wraps ScrapeStudentWiseAttendance
func (h *Handler) GetStudentWiseAttendance(w http.ResponseWriter, r *http.Request) {
	var req models.AttendanceCourseOption
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	session := middleware.GetSession(r)
	summaries, err := session.Scraper.ScrapeStudentWiseAttendance(&req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: summaries})
}

// --- Feedback Endpoints ---

// SubmitFeedback handles saving feedback to Supabase
func (h *Handler) SubmitFeedback(w http.ResponseWriter, r *http.Request) {
	var req models.FacultyFeedback
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	session := middleware.GetSession(r)
	if session != nil && session.Faculty != nil {
		req.FacultyName = session.Faculty.Name
		req.EmployeeID = session.Faculty.EmployeeID
		req.Department = session.Faculty.Department
	}

	if h.repo != nil {
		if err := h.repo.SubmitFeedback(r.Context(), &req); err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to submit feedback")
			return
		}
	} else {
		// Mock success if Supabase isn't configured
		logger.Log.Warn("Supabase not configured, simulating feedback submission")
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Message: "Feedback submitted successfully"})
}

// --- Timetable Endpoints ---

// GetTimetable wraps ScrapeTimetable and combines with custom DB entries
func (h *Handler) GetTimetable(w http.ResponseWriter, r *http.Request) {
	session := middleware.GetSession(r)
	entries, err := session.Scraper.ScrapeTimetable()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if h.repo != nil && session.Faculty != nil {
		mods, err := h.repo.GetTimetableModifications(r.Context(), session.Faculty.EmployeeID)
		if err == nil {
			var finalEntries []models.TimetableEntry
			
			// filter scraped entries based on hidden mods
			for _, entry := range entries {
				isHidden := false
				for _, mod := range mods {
					if mod.IsHidden && mod.Day == entry.Day && mod.Time == entry.Time && mod.Subject == entry.Subject {
						isHidden = true
						break
					}
				}
				if !isHidden {
					finalEntries = append(finalEntries, entry)
				}
			}

			// add custom entries
			for _, mod := range mods {
				if mod.IsCustom {
					finalEntries = append(finalEntries, models.TimetableEntry{
						Day:       mod.Day,
						Time:      mod.Time,
						Subject:   mod.Subject,
						Type:      mod.ClassType,
						Room:      mod.Room,
						Batch:     mod.Batch,
						IsCustom:  true,
						IsHidden:  false,
					})
				}
			}
			entries = finalEntries
		}
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: entries})
}

// AddCustomTimetableEntry adds a custom entry
func (h *Handler) AddCustomTimetableEntry(w http.ResponseWriter, r *http.Request) {
	var req models.CustomTimetableEntry
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	session := middleware.GetSession(r)
	if session == nil || session.Faculty == nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	req.EmployeeID = session.Faculty.EmployeeID
	req.IsCustom = true
	req.IsHidden = false

	if h.repo != nil {
		if err := h.repo.AddTimetableEntry(r.Context(), &req); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Message: "Entry added successfully"})
}

// HideScrapedTimetableEntry hides a scraped entry
func (h *Handler) HideScrapedTimetableEntry(w http.ResponseWriter, r *http.Request) {
	var req models.CustomTimetableEntry
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	session := middleware.GetSession(r)
	if session == nil || session.Faculty == nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	req.EmployeeID = session.Faculty.EmployeeID
	req.IsCustom = false
	req.IsHidden = true

	if h.repo != nil {
		if err := h.repo.AddTimetableEntry(r.Context(), &req); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Message: "Entry hidden successfully"})
}

// DeleteCustomTimetableEntry deletes a custom entry
func (h *Handler) DeleteCustomTimetableEntry(w http.ResponseWriter, r *http.Request) {
	var req models.CustomTimetableEntry
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	session := middleware.GetSession(r)
	if session == nil || session.Faculty == nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if h.repo != nil {
		if err := h.repo.DeleteCustomTimetableEntry(r.Context(), session.Faculty.EmployeeID, req.Day, req.Time, req.Subject); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Message: "Entry deleted successfully"})
}

// ResetTimetable deletes all custom/hidden entries
func (h *Handler) ResetTimetable(w http.ResponseWriter, r *http.Request) {
	session := middleware.GetSession(r)
	if session == nil || session.Faculty == nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if h.repo != nil {
		if err := h.repo.ResetTimetable(r.Context(), session.Faculty.EmployeeID); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Message: "Timetable reset successfully"})
}

// HealthCheck endpoint
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	})
}
