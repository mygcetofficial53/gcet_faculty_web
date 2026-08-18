package models

import "time"

// Faculty represents a faculty member's profile
type Faculty struct {
	ID               string   `json:"id"`
	LoginID          string   `json:"login_id"`
	Name             string   `json:"name"`
	FirstName        string   `json:"first_name"`
	MiddleName       string   `json:"middle_name"`
	LastName         string   `json:"last_name"`
	EmployeeID       string   `json:"employee_id"`
	Department       string   `json:"department"`
	Designation      string   `json:"designation"`
	Email            string   `json:"email"`
	Phone            string   `json:"phone"`
	Qualification    string   `json:"qualification"`
	Experience       string   `json:"experience"`
	Address          string   `json:"address"`
	DateOfBirth      string   `json:"dob"`
	Gender           string   `json:"gender"`
	BloodGroup       string   `json:"blood_group"`
	Aadhar           string   `json:"aadhar"`
	PanNo            string   `json:"pan_no"`
	Category         string   `json:"category"`
	MaritalStatus    string   `json:"marital_status"`
	JoiningDate      string   `json:"joining_date"`
	ContactNo2       string   `json:"contact_no2"`
	PermanentAddress string   `json:"permanent_address"`
	Religion         string   `json:"religion"`
	Caste            string   `json:"caste"`
	BankAccount      string   `json:"bank_account"`
	BankName         string   `json:"bank_name"`
	IFSCCode         string   `json:"ifsc_code"`
	Subjects         []string `json:"subjects"`
}

// QualificationDetail from QualificationDetails.jsp
type QualificationDetail struct {
	SrNo           string `json:"sr_no"`
	Degree         string `json:"degree"`
	Specialization string `json:"specialization"`
	University     string `json:"university"`
	YearOfPassing  string `json:"year_of_passing"`
	Percentage     string `json:"percentage"`
	Grade          string `json:"grade"`
	Remarks        string `json:"remarks"`
	CVMUResult     string `json:"cvmu_result"`
}

// ExperienceDetail from ExperienceDetails.jsp
type ExperienceDetail struct {
	SrNo         string `json:"sr_no"`
	Organization string `json:"organization"`
	Designation  string `json:"designation"`
	FromDate     string `json:"from_date"`
	ToDate       string `json:"to_date"`
	Years        string `json:"years"`
	Type         string `json:"type"`
	Remarks      string `json:"remarks"`
}

// LoginRequest represents the login form data
type LoginRequest struct {
	LoginID  string `json:"login_id" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse returned after successful login
type LoginResponse struct {
	Success      bool     `json:"success"`
	Token        string   `json:"token,omitempty"`
	RefreshToken string   `json:"refresh_token,omitempty"`
	Faculty      *Faculty `json:"faculty,omitempty"`
	Error        string   `json:"error,omitempty"`
}

// TokenClaims for JWT
type TokenClaims struct {
	LoginID    string `json:"login_id"`
	EmployeeID string `json:"employee_id"`
	Name       string `json:"name"`
	SessionID  string `json:"session_id"`
}

// AttendanceRecord for a single date
type AttendanceRecord struct {
	Date          string `json:"date"`
	Day           string `json:"day"`
	TimeSlot      string `json:"time_slot"`
	PresentCount  int    `json:"present_count"`
	AbsentCount   int    `json:"absent_count"`
	TotalStudents int    `json:"total_students"`
	Topic         string `json:"topic"`
}

// AttendanceMetadata holds detailed context for attendance operations
type AttendanceMetadata struct {
	CourseName     string `json:"course_name"`
	Semester       string `json:"semester"`
	Division       string `json:"division"`
	AcademicYear   string `json:"academic_year"`
	Term           string `json:"term"`
	DeptName       string `json:"dept_name"`
	CourseTeacher  string `json:"course_teacher"`
	ClassType      string `json:"class_type"`
	ProgramCode    string `json:"program_code"`
	ClassName      string `json:"class_name"`
	GTUBranchCode  string `json:"gtu_branch_code"`
	CVMUBranchCode string `json:"cvmu_branch_code"`
	CLS            string `json:"cls"`
}

// AttendanceSheet for a course
type AttendanceSheet struct {
	CourseCode        string              `json:"course_code"`
	CourseName        string              `json:"course_name"`
	Type              string              `json:"type"`
	Batch             string              `json:"batch"`
	FromDate          string              `json:"from_date"`
	ToDate            string              `json:"to_date"`
	Records           []AttendanceRecord  `json:"records"`
	AverageAttendance float64             `json:"average_attendance"`
	Metadata          *AttendanceMetadata `json:"metadata,omitempty"`
}

// SubjectAvgAttendance holds average attendance per subject
type SubjectAvgAttendance struct {
	CourseCode    string  `json:"course_code"`
	CourseName    string  `json:"course_name"`
	Type          string  `json:"type"`
	Batch         string  `json:"batch"`
	AvgPercentage float64 `json:"avg_percentage"`
	TotalStudents int     `json:"total_students"`
}

// DateWiseAttendance from DateWiseAttendance.jsp
type DateWiseAttendance struct {
	Date       string `json:"date"`
	CourseCode string `json:"course_code"`
	CourseName string `json:"course_name"`
	TimeSlot   string `json:"time_slot"`
	Type       string `json:"type"`
	Present    int    `json:"present"`
	Absent     int    `json:"absent"`
	Total      int    `json:"total"`
	Topic      string `json:"topic"`
}

// TopicRecord from ViewTopic.jsp
type TopicRecord struct {
	SrNo         string `json:"sr_no"`
	Date         string `json:"date"`
	CourseCode   string `json:"course_code"`
	CourseName   string `json:"course_name"`
	Topic        string `json:"topic"`
	HoursPlanned string `json:"hours_planned"`
	HoursCovered string `json:"hours_covered"`
	Type         string `json:"type"`
}

// StudentAttendanceSummary for student-wise analytics
type StudentAttendanceSummary struct {
	Enrollment string  `json:"enrollment"`
	Name       string  `json:"name"`
	CourseCode string  `json:"course_code"`
	CourseName string  `json:"course_name"`
	Attended   int     `json:"attended"`
	Total      int     `json:"total"`
	Percentage float64 `json:"percentage"`
}

// AttendanceCourseOption for course selection dropdown
type AttendanceCourseOption struct {
	CourseCode       string              `json:"course_code"`
	CourseName       string              `json:"course_name"`
	RawValue         string              `json:"raw_value"`
	DisplayText      string              `json:"display_text"`
	Type             string              `json:"type"`
	Batch            string              `json:"batch"`
	OptionIndex      int                 `json:"option_index"`
	LastEnteredDates []string            `json:"last_entered_dates"`
	Metadata         *AttendanceMetadata `json:"metadata,omitempty"`
}

// AttendanceStudentEntry for student list in attendance marking
type AttendanceStudentEntry struct {
	Enrollment string `json:"enrollment"`
	Name       string `json:"name"`
	LibraryID  string `json:"library_id"`
	IsPresent  bool   `json:"is_present"`
}

// AttendanceEditOption for editing past attendance
type AttendanceEditOption struct {
	Date      string `json:"date"`
	LectureNo string `json:"lecture_no"`
	RawValue  string `json:"raw_value"`
}

// PrintAttendanceRow for blank attendance sheets
type PrintAttendanceRow struct {
	SrNo       string `json:"sr_no"`
	Enrollment string `json:"enrollment"`
	Name       string `json:"name"`
}

// TimetableEntry for a faculty member's schedule
type TimetableEntry struct {
	Day       string `json:"day"`
	Time      string `json:"time"`
	Subject   string `json:"subject"`
	Type      string `json:"type"`
	Room      string `json:"room"`
	Batch     string `json:"batch"`
	ClassRoom string `json:"classroom"`
	IsCustom  bool   `json:"is_custom"`
	IsHidden  bool   `json:"is_hidden"`
}

// CustomTimetableEntry for Supabase mapping
type CustomTimetableEntry struct {
	ID         string    `json:"id,omitempty"`
	EmployeeID string    `json:"employee_id"`
	Day        string    `json:"day"`
	Time       string    `json:"time"`
	Subject    string    `json:"subject"`
	ClassType  string    `json:"class_type"`
	Room       string    `json:"room"`
	Batch      string    `json:"batch"`
	IsCustom   bool      `json:"is_custom"`
	IsHidden   bool      `json:"is_hidden"`
	CreatedAt  time.Time `json:"created_at,omitempty"`
}

// FacultyFeedback for the support/feedback form
type FacultyFeedback struct {
	ID          string    `json:"id,omitempty"`
	FacultyName string    `json:"faculty_name"`
	EmployeeID  string    `json:"employee_id"`
	Department  string    `json:"department"`
	Type        string    `json:"type"`
	Subject     string    `json:"subject"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

// SubmitAttendanceRequest for entering new attendance
type SubmitAttendanceRequest struct {
	CourseCode   string                   `json:"course_code"`
	Date         string                   `json:"date"`
	TimeSlot     string                   `json:"time_slot"`
	Topic        string                   `json:"topic"`
	Type         string                   `json:"type"`
	Students     []AttendanceStudentEntry `json:"students"`
	ExtraLecture bool                     `json:"extra_lecture"`
	LecNo        string                   `json:"lec_no"`
	Metadata     *AttendanceMetadata      `json:"metadata,omitempty"`
	OptionIndex  int                      `json:"option_index"`
}

// EditAttendanceRequest for editing existing attendance
type EditAttendanceRequest struct {
	CourseCode string                   `json:"course_code"`
	Date       string                   `json:"date"`
	TimeSlot   string                   `json:"time_slot"`
	Topic      string                   `json:"topic"`
	Type       string                   `json:"type"`
	Students   []AttendanceStudentEntry `json:"students"`
}

// DeleteAttendanceRequest for deleting attendance
type DeleteAttendanceRequest struct {
	CourseCode string `json:"course_code"`
	Date       string `json:"date"`
	TimeSlot   string `json:"time_slot"`
}

// DynamicSessionRequest for creating a dynamic attendance session
type DynamicSessionRequest struct {
	CourseID        string `json:"course_id"`
	CourseName      string `json:"course_name"`
	Batch           string `json:"batch"`
	Semester        string `json:"semester"`
	CodeLength      int    `json:"code_length"`
	DurationSeconds int    `json:"duration_seconds"`
	SessionType     string `json:"session_type"`
}

// APIResponse is a standard API response wrapper
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

// PaginatedResponse wraps paginated data
type PaginatedResponse struct {
	Success    bool        `json:"success"`
	Data       interface{} `json:"data"`
	Total      int         `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}
