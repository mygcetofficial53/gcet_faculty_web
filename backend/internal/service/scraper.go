package service

import (
	"crypto/md5"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"gcet-web-backend/internal/logger"
	"gcet-web-backend/internal/models"

	"github.com/PuerkitoBio/goquery"
)

// GMSScraper handles all communication with the GMS Faculty Portal
type GMSScraper struct {
	baseURL     string
	client      *http.Client
	cookieJar   *cookiejar.Jar
	username    string
	password    string
	mu          sync.Mutex
	isRelogging bool
	staffName   string
	empID       string
}

// NewGMSScraper creates a new scraper instance for a faculty session
func NewGMSScraper(baseURL string) *GMSScraper {
	jar, _ := cookiejar.New(nil)
	
	transport := http.DefaultTransport.(*http.Transport).Clone()
	// Force disable HTTP/2 to prevent "http2: unsupported scheme" errors with proxies
	transport.ForceAttemptHTTP2 = false
	transport.TLSNextProto = make(map[string]func(authority string, c *tls.Conn) http.RoundTripper)
	
	var proxyURLStr string
	
	// Check if a static proxy is provided
	if staticProxy := os.Getenv("PROXY_URL"); staticProxy != "" {
		proxyURLStr = staticProxy
	} else if GlobalProxyPool != nil {
		// Otherwise, get a random proxy from the pool
		proxyURLStr = GlobalProxyPool.GetRandomProxy()
	}

	if proxyURLStr != "" {
		// Convert https proxy scheme to http to avoid http2 issues
		if strings.HasPrefix(strings.ToLower(proxyURLStr), "https://") {
			proxyURLStr = "http://" + proxyURLStr[8:]
		}
		if proxyURL, err := url.Parse(proxyURLStr); err == nil {
			transport.Proxy = http.ProxyURL(proxyURL)
			logger.Log.Infof("GMS Scraper configured to use proxy: %s", proxyURLStr)
		}
	}
	
	client := &http.Client{
		Transport: transport,
		Jar:       jar,
		Timeout:   60 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}

	return &GMSScraper{
		baseURL:   baseURL,
		client:    client,
		cookieJar: jar,
	}
}

// URL helpers
func (s *GMSScraper) loginPageURL() string   { return s.baseURL + "/index.jsp" }
func (s *GMSScraper) loginActionURL() string { return s.baseURL + "/LoginCheck.do" }
func (s *GMSScraper) profileURL() string     { return s.baseURL + "/Faculty/Profile/PersonalDetails.jsp" }
func (s *GMSScraper) qualificationURL() string {
	return s.baseURL + "/Faculty/Profile/QualificationDetails.jsp"
}
func (s *GMSScraper) experienceURL() string {
	return s.baseURL + "/Faculty/Profile/ExperienceDetails.jsp"
}
func (s *GMSScraper) uploadMaterialURL() string {
	return s.baseURL + "/Faculty/Teaching/UploadMaterial.jsp"
}
func (s *GMSScraper) timetableURL() string {
	return s.baseURL + "/Faculty/Teaching/FacultyTimetable.jsp"
}
func (s *GMSScraper) attendanceSheetURL() string {
	return s.baseURL + "/Faculty/Teaching/Attendance/ViewAttendanceSheet.jsp"
}
func (s *GMSScraper) avgAttendanceURL() string {
	return s.baseURL + "/Faculty/Teaching/Attendance/AvgAttendance.jsp"
}
func (s *GMSScraper) viewTopicURL() string {
	return s.baseURL + "/Faculty/Teaching/Attendance/ViewTopic.jsp"
}
func (s *GMSScraper) dateWiseAttendanceURL() string {
	return s.baseURL + "/Faculty/Teaching/Attendance/DateWiseAttendance.jsp"
}
func (s *GMSScraper) enterAttendanceURL() string {
	return s.baseURL + "/Faculty/Teaching/Attendance/EnterAttendance_OLD1.jsp"
}
func (s *GMSScraper) enterAttendanceLibURL() string {
	return s.baseURL + "/Faculty/Teaching/Attendance/EnterAttendance_LibID.jsp"
}
func (s *GMSScraper) editAttendanceURL() string {
	return s.baseURL + "/Faculty/Teaching/Attendance/EditAttendance.jsp"
}
func (s *GMSScraper) editAttendanceActionURL() string {
	return s.baseURL + "/Faculty/Teaching/Attendance/EditAttendanceAction.do"
}
func (s *GMSScraper) deleteAttendanceURL() string {
	return s.baseURL + "/Faculty/Teaching/Attendance/DeleteAttendance.jsp"
}
func (s *GMSScraper) deleteAttendanceActionURL() string {
	return s.baseURL + "/Faculty/Teaching/Attendance/DeleteAttendanceAction.do"
}
func (s *GMSScraper) printAttendanceSheetURL() string {
	return s.baseURL + "/Faculty/Teaching/Attendance/PrintAttendanceSheet.jsp"
}
func (s *GMSScraper) attendanceSheetDetailsURL() string {
	return s.baseURL + "/Faculty/Teaching/Attendance/ViewAttendanceSheet_Details.jsp"
}
func (s *GMSScraper) saveProfileURL() string {
	return s.baseURL + "/Faculty/Profile/SavePersonalDetails.do"
}
func (s *GMSScraper) saveQualificationURL() string {
	return s.baseURL + "/Faculty/Profile/SaveQualificationDetails.do"
}
func (s *GMSScraper) saveExperienceURL() string {
	return s.baseURL + "/Faculty/Profile/SaveExperienceDetails.do"
}
func (s *GMSScraper) logoutURL() string { return s.baseURL + "/Logout.jsp" }

// Login authenticates with the GMS portal
func (s *GMSScraper) Login(username, password string) (*models.Faculty, error) {
	s.username = username
	s.password = password

	logger.Log.Infof("GMS Scraper: Starting login for %s", username)

	// MD5 hash the password (GMS portal JavaScript does calcMD5(pass))
	passwordMD5 := fmt.Sprintf("%x", md5.Sum([]byte(password)))

	// Step 1: GET login page to establish session cookies
	_, err := s.doGet(s.loginPageURL())
	if err != nil {
		return nil, fmt.Errorf("failed to load login page: %w", err)
	}

	// Step 2: POST login form
	form := url.Values{
		"login_id": {username},
		"pass":     {passwordMD5},
	}

	req, err := http.NewRequest("POST", s.loginActionURL(), strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create login request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", s.loginPageURL())
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("login request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read login response: %w", err)
	}
	body := string(bodyBytes)
	bodyLower := strings.ToLower(body)

	// Check if login form is still showing (login failed)
	if strings.Contains(body, `name="f1"`) && strings.Contains(body, "Faculty Login") {
		return nil, fmt.Errorf("invalid Login ID or Password")
	}

	// Verify login by checking for faculty-specific content
	isLoggedIn := strings.Contains(bodyLower, "logout") ||
		strings.Contains(bodyLower, "material") ||
		strings.Contains(bodyLower, "academic") ||
		strings.Contains(bodyLower, "faculty")

	if isLoggedIn {
		faculty, err := s.scrapeProfile(username)
		if err != nil {
			return nil, fmt.Errorf("failed to scrape profile: %w", err)
		}
		return faculty, nil
	}

	// Fallback: try fetching upload material page
	testBody, err := s.doGet(s.uploadMaterialURL())
	if err == nil {
		testBodyLower := strings.ToLower(testBody)
		if strings.Contains(testBodyLower, "upload") && strings.Contains(testBodyLower, "material") && !strings.Contains(testBodyLower, "faculty login") {
			faculty, err := s.scrapeProfile(username)
			if err != nil {
				return nil, fmt.Errorf("failed to scrape profile: %w", err)
			}
			return faculty, nil
		}
	}

	return nil, fmt.Errorf("login failed — please check your credentials")
}

// Logout ends the GMS session
func (s *GMSScraper) Logout() {
	_, _ = s.doGet(s.logoutURL())
	s.username = ""
	s.password = ""
	// Clear cookies
	jar, _ := cookiejar.New(nil)
	s.client.Jar = jar
	s.cookieJar = jar
}

// doGet performs a GET request and returns the response body
func (s *GMSScraper) doGet(url string) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(bodyBytes), nil
}

// doPost performs a POST request with form data and returns the response body
func (s *GMSScraper) doPost(targetURL string, form url.Values) (string, error) {
	req, err := http.NewRequest("POST", targetURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", targetURL)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(bodyBytes), nil
}

// doPostRaw posts a raw body string (not url.Values)
func (s *GMSScraper) doPostRaw(targetURL string, body string) (string, error) {
	req, err := http.NewRequest("POST", targetURL, strings.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", targetURL)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(bodyBytes), nil
}

// getWithRetry performs GET with auto-relogin on session expiry
func (s *GMSScraper) getWithRetry(url string) (string, error) {
	body, err := s.doGet(url)
	if err != nil {
		return "", err
	}

	bodyLower := strings.ToLower(body)
	if strings.Contains(bodyLower, "faculty login") || (strings.Contains(bodyLower, "login_id") && strings.Contains(body, `name="f1"`)) {
		logger.Log.Infof("GMS Scraper: Session expired for %s, re-authenticating...", url)

		if s.username != "" && s.password != "" {
			s.mu.Lock()
			if s.isRelogging {
				s.mu.Unlock()
				time.Sleep(500 * time.Millisecond)
				return s.doGet(url)
			}
			s.isRelogging = true
			s.mu.Unlock()

			defer func() {
				s.mu.Lock()
				s.isRelogging = false
				s.mu.Unlock()
			}()

			_, err := s.Login(s.username, s.password)
			if err != nil {
				return "", fmt.Errorf("auto-relogin failed: %w", err)
			}
			return s.doGet(url)
		}
	}
	return body, nil
}

// postWithRetry performs POST with auto-relogin on session expiry
func (s *GMSScraper) postWithRetry(targetURL string, form url.Values) (string, error) {
	body, err := s.doPost(targetURL, form)
	if err != nil {
		return "", err
	}

	bodyLower := strings.ToLower(body)
	if strings.Contains(bodyLower, "faculty login") || (strings.Contains(bodyLower, "login_id") && strings.Contains(body, `name="f1"`)) {
		if s.username != "" && s.password != "" {
			s.mu.Lock()
			if s.isRelogging {
				s.mu.Unlock()
				time.Sleep(500 * time.Millisecond)
				return s.doPost(targetURL, form)
			}
			s.isRelogging = true
			s.mu.Unlock()

			defer func() {
				s.mu.Lock()
				s.isRelogging = false
				s.mu.Unlock()
			}()

			_, err := s.Login(s.username, s.password)
			if err != nil {
				return "", fmt.Errorf("auto-relogin failed: %w", err)
			}
			return s.doPost(targetURL, form)
		}
	}
	return body, nil
}

// scrapeProfile parses the PersonalDetails.jsp page
func (s *GMSScraper) scrapeProfile(loginID string) (*models.Faculty, error) {
	body, err := s.doGet(s.profileURL())
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		return nil, err
	}

	f := &models.Faculty{
		ID:      fmt.Sprintf("%d", time.Now().UnixMilli()),
		LoginID: loginID,
	}

	var firstName, middleName, surname string

	// Parse input fields
	doc.Find("input, select, textarea").Each(func(_ int, sel *goquery.Selection) {
		name := strings.ToLower(sel.AttrOr("name", ""))
		value := sel.AttrOr("value", "")
		if value == "" {
			value = strings.TrimSpace(sel.Text())
		}
		if value == "" || len(value) > 500 {
			return
		}

		switch {
		case contains(name, "emp_id") || contains(name, "staff_code"):
			f.EmployeeID = value
			s.empID = value
		case contains(name, "first_name"):
			firstName = value
		case contains(name, "middle_name"):
			middleName = value
		case contains(name, "surname") || contains(name, "last_name"):
			surname = value
		case contains(name, "email"):
			f.Email = value
		case contains(name, "mobile") || contains(name, "phone"):
			f.Phone = value
		case contains(name, "birth") || contains(name, "dob"):
			f.DateOfBirth = value
		case name == "address" || contains(name, "current_address"):
			f.Address = value
		case contains(name, "perm") && contains(name, "address"):
			f.PermanentAddress = value
		case contains(name, "gender") || name == "sex":
			f.Gender = value
		case contains(name, "aadhar"):
			f.Aadhar = value
		case contains(name, "pan_no"):
			f.PanNo = value
		case contains(name, "blood"):
			f.BloodGroup = value
		case contains(name, "religion"):
			f.Religion = value
		case contains(name, "caste"):
			f.Caste = value
		case contains(name, "bank_acc") || contains(name, "account_no"):
			f.BankAccount = value
		case contains(name, "bank_name"):
			f.BankName = value
		case contains(name, "ifsc"):
			f.IFSCCode = value
		}
	})

	// Construct name
	if firstName != "" && strings.ToLower(firstName) != "first name" {
		f.Name = strings.TrimSpace(firstName + " " + middleName + " " + surname)
		f.FirstName = firstName
		f.MiddleName = middleName
		f.LastName = surname
		s.staffName = f.Name
	}

	// Parse table rows for label:value data
	doc.Find("table tr").Each(func(_ int, row *goquery.Selection) {
		cols := row.Find("td")
		if cols.Length() >= 2 {
			label := strings.ToLower(strings.TrimSpace(cols.Eq(0).Text()))
			value := strings.TrimSpace(cols.Eq(1).Text())
			if value == "" || len(value) > 500 {
				return
			}

			switch {
			case contains(label, "name") && !contains(label, "father") && !contains(label, "mother") && !contains(label, "spouse"):
				if f.Name == "" {
					f.Name = value
				}
			case contains(label, "department") || contains(label, "dept"):
				if f.Department == "" {
					f.Department = value
				}
			case contains(label, "designation"):
				if f.Designation == "" {
					f.Designation = value
				}
			case contains(label, "email"):
				if f.Email == "" {
					f.Email = value
				}
			case contains(label, "mobile") || contains(label, "phone") || contains(label, "contact"):
				if f.Phone == "" {
					f.Phone = value
				} else if f.ContactNo2 == "" && value != f.Phone {
					f.ContactNo2 = value
				}
			case contains(label, "qualification"):
				if f.Qualification == "" {
					f.Qualification = value
				}
			case contains(label, "experience"):
				if f.Experience == "" {
					f.Experience = value
				}
			case contains(label, "dob") || contains(label, "birth"):
				if f.DateOfBirth == "" {
					f.DateOfBirth = value
				}
			case contains(label, "gender") || contains(label, "sex"):
				if f.Gender == "" {
					f.Gender = value
				}
			case contains(label, "blood"):
				if f.BloodGroup == "" {
					f.BloodGroup = value
				}
			case contains(label, "category"):
				if f.Category == "" {
					f.Category = value
				}
			case contains(label, "marital"):
				if f.MaritalStatus == "" {
					f.MaritalStatus = value
				}
			case contains(label, "join"):
				if f.JoiningDate == "" {
					f.JoiningDate = value
				}
			case contains(label, "employee id") || contains(label, "staff code"):
				if f.EmployeeID == "" {
					f.EmployeeID = value
				}
			}
		}
	})

	// Fallback name
	if f.Name == "" || strings.ToLower(f.Name) == "first name" || strings.ToLower(f.Name) == "first name surname" {
		parts := strings.Split(loginID, ".")
		nameParts := make([]string, len(parts))
		for i, p := range parts {
			if len(p) > 0 {
				nameParts[i] = strings.ToUpper(p[:1]) + p[1:]
			}
		}
		f.Name = strings.Join(nameParts, " ")
		s.staffName = f.Name
	}

	// Split name into parts if not already set
	if f.FirstName == "" && f.Name != "" {
		parts := strings.Fields(f.Name)
		if len(parts) > 0 {
			f.FirstName = parts[0]
		}
		if len(parts) > 2 {
			f.MiddleName = strings.Join(parts[1:len(parts)-1], " ")
			f.LastName = parts[len(parts)-1]
		} else if len(parts) == 2 {
			f.LastName = parts[1]
		}
	}

	if f.EmployeeID == "" {
		f.EmployeeID = loginID
	}

	// Get subjects from upload material page
	matBody, err := s.doGet(s.uploadMaterialURL())
	if err == nil {
		matDoc, err := goquery.NewDocumentFromReader(strings.NewReader(matBody))
		if err == nil {
			matDoc.Find(`select[name="course_code"] option`).Each(func(_ int, opt *goquery.Selection) {
				val := opt.AttrOr("value", "")
				text := strings.TrimSpace(opt.Text())
				if val != "" && text != "" && !strings.Contains(text, "--Select") {
					f.Subjects = append(f.Subjects, text)
				}
			})
		}
	}

	logger.Log.Infof("GMS Scraper: Profile scraped for %s (%s)", f.Name, f.EmployeeID)
	return f, nil
}

// ScrapeQualifications parses QualificationDetails.jsp
func (s *GMSScraper) ScrapeQualifications() ([]models.QualificationDetail, error) {
	body, err := s.getWithRetry(s.qualificationURL())
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		return nil, err
	}

	var qualifications []models.QualificationDetail
	doc.Find("table").Each(func(_ int, table *goquery.Selection) {
		if len(qualifications) > 0 {
			return
		}
		tableText := strings.ToLower(table.Text())
		if contains(tableText, "qualification") || contains(tableText, "degree") || contains(tableText, "university") {
			rows := parseTableWithMapping(table)
			for _, row := range rows {
				qualifications = append(qualifications, models.QualificationDetail{
					SrNo:           getMapVal(row, "sr.no.", "sr. no.", "srno"),
					Degree:         getMapVal(row, "degree", "name of degree"),
					Specialization: getMapVal(row, "specialization", "branch"),
					University:     getMapVal(row, "university", "board/university"),
					YearOfPassing:  getMapVal(row, "year of passing", "passing year"),
					Percentage:     getMapVal(row, "percentage", "%"),
					Grade:          getMapVal(row, "grade", "class"),
					Remarks:        getMapVal(row, "remarks"),
					CVMUResult:     getMapVal(row, "cvmu result"),
				})
			}
		}
	})

	return qualifications, nil
}

// ScrapeExperience parses ExperienceDetails.jsp
func (s *GMSScraper) ScrapeExperience() ([]models.ExperienceDetail, error) {
	body, err := s.getWithRetry(s.experienceURL())
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		return nil, err
	}

	var experiences []models.ExperienceDetail
	doc.Find("table").Each(func(_ int, table *goquery.Selection) {
		if len(experiences) > 0 {
			return
		}
		tableText := strings.ToLower(table.Text())
		if contains(tableText, "experience") || contains(tableText, "organization") {
			rows := parseTableWithMapping(table)
			for _, row := range rows {
				experiences = append(experiences, models.ExperienceDetail{
					SrNo:         getMapVal(row, "sr.no.", "sr. no.", "srno"),
					Organization: getMapVal(row, "organization", "name of organization"),
					Designation:  getMapVal(row, "designation"),
					FromDate:     getMapVal(row, "from date", "joining date"),
					ToDate:       getMapVal(row, "to date", "leaving date"),
					Years:        getMapVal(row, "years", "total years"),
					Type:         getMapVal(row, "type", "experience type"),
					Remarks:      getMapVal(row, "remarks"),
				})
			}
		}
	})

	return experiences, nil
}

// ScrapeAttendanceSheets parses ViewAttendanceSheet.jsp
func (s *GMSScraper) ScrapeAttendanceSheets() ([]models.AttendanceSheet, error) {
	body, err := s.getWithRetry(s.attendanceSheetURL())
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		return nil, err
	}

	var sheets []models.AttendanceSheet
	doc.Find("table").Each(func(_ int, table *goquery.Selection) {
		if len(sheets) > 0 {
			return
		}
		tableText := strings.ToLower(table.Text())
		if contains(tableText, "course") && (contains(tableText, "present") || contains(tableText, "absent") || contains(tableText, "attendance")) {
			rows := parseTableWithMapping(table)
			for _, row := range rows {
				avgStr := getMapVal(row, "avg percentage", "avg_percentage")
				avgStr = strings.ReplaceAll(avgStr, "%", "")
				avg, _ := strconv.ParseFloat(avgStr, 64)

				sheets = append(sheets, models.AttendanceSheet{
					CourseCode:        getMapVal(row, "course code", "course_code"),
					CourseName:        getMapVal(row, "course name", "course_name"),
					Type:              getMapValDefault(row, "Theory", "type"),
					Batch:             getMapVal(row, "batch"),
					AverageAttendance: avg,
				})
			}
		}
	})

	return sheets, nil
}

// ScrapeAverageAttendance parses AvgAttendance.jsp
func (s *GMSScraper) ScrapeAverageAttendance() ([]models.SubjectAvgAttendance, error) {
	body, err := s.getWithRetry(s.avgAttendanceURL())
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		return nil, err
	}

	var avgList []models.SubjectAvgAttendance
	doc.Find("table").Each(func(_ int, table *goquery.Selection) {
		if len(avgList) > 0 {
			return
		}
		tableText := strings.ToLower(table.Text())
		if contains(tableText, "course code") && (contains(tableText, "avg") || contains(tableText, "percentage")) {
			rows := parseTableWithMapping(table)
			for _, row := range rows {
				avgStr := getMapVal(row, "avg percentage", "avg_percentage", "avg %")
				avgStr = strings.ReplaceAll(avgStr, "%", "")
				avg, _ := strconv.ParseFloat(strings.TrimSpace(avgStr), 64)

				avgList = append(avgList, models.SubjectAvgAttendance{
					CourseCode:    getMapVal(row, "course code", "course_code"),
					CourseName:    getMapVal(row, "course name", "course_name"),
					Type:          getMapValDefault(row, "Theory", "type"),
					AvgPercentage: avg,
				})
			}
		}
	})

	return avgList, nil
}

// ScrapeViewTopics parses ViewTopic.jsp
func (s *GMSScraper) ScrapeViewTopics() ([]models.TopicRecord, error) {
	body, err := s.getWithRetry(s.viewTopicURL())
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		return nil, err
	}

	var topics []models.TopicRecord
	doc.Find("table").Each(func(_ int, table *goquery.Selection) {
		if len(topics) > 0 {
			return
		}
		tableText := strings.ToLower(table.Text())
		if contains(tableText, "topic") && (contains(tableText, "hours") || contains(tableText, "covered")) {
			rows := parseTableWithMapping(table)
			for _, row := range rows {
				topics = append(topics, models.TopicRecord{
					SrNo:         getMapVal(row, "sr.no.", "sr no", "srno"),
					Date:         getMapVal(row, "date"),
					CourseCode:   getMapVal(row, "course code", "course_code"),
					CourseName:   getMapVal(row, "course name", "course_name"),
					Topic:        getMapVal(row, "topic covered", "topic"),
					HoursPlanned: getMapVal(row, "hours planned", "hours_planned"),
					HoursCovered: getMapVal(row, "hours covered", "hours_covered"),
					Type:         getMapVal(row, "type"),
				})
			}
		}
	})

	return topics, nil
}

// ScrapeTimetable parses FacultyTimetable.jsp
func (s *GMSScraper) ScrapeTimetable() ([]models.TimetableEntry, error) {
	body, err := s.getWithRetry(s.timetableURL())
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		return nil, err
	}

	var entries []models.TimetableEntry
	doc.Find("table").Each(func(_ int, table *goquery.Selection) {
		if len(entries) > 0 {
			return
		}
		tableText := strings.ToLower(table.Text())
		if contains(tableText, "monday") && contains(tableText, "tuesday") {
			rows := table.Find("tr")
			if rows.Length() <= 2 {
				return
			}

			// Get headers
			var headers []string
			rows.First().Find("th, td").Each(func(_ int, cell *goquery.Selection) {
				headers = append(headers, strings.ToLower(strings.TrimSpace(cell.Text())))
			})

			rows.Each(func(i int, row *goquery.Selection) {
				if i == 0 {
					return
				}
				cols := row.Find("td")
				if cols.Length() < 3 {
					return
				}

				typeVal := findInCols(headers, cols, "type")
				if typeVal == "" {
					typeVal = "Theory"
				}

				entries = append(entries, models.TimetableEntry{
					Day:       findInCols(headers, cols, "day"),
					Time:      findInCols(headers, cols, "time"),
					Subject:   findInCols(headers, cols, "subject"),
					Type:      typeVal,
					Room:      findInCols(headers, cols, "room"),
					Batch:     findInCols(headers, cols, "batch"),
					ClassRoom: findInCols(headers, cols, "classroom"),
				})
			})
		}
	})

	return entries, nil
}

// FetchAttendanceCourses parses course dropdown from EnterAttendance page
func (s *GMSScraper) FetchAttendanceCourses() ([]models.AttendanceCourseOption, error) {
	body, err := s.getWithRetry(s.enterAttendanceURL())
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		return nil, err
	}

	// Parse metadata from hidden inputs
	metadataMap := map[string][]string{}
	doc.Find("input").Each(func(_ int, input *goquery.Selection) {
		name := input.AttrOr("name", "")
		value := input.AttrOr("value", "")
		if strings.HasSuffix(name, "_h") {
			metadataMap[name] = append(metadataMap[name], value)
		}
	})

	var courses []models.AttendanceCourseOption
	doc.Find(`select[name="course_code"] option, select[name="courseCode"] option`).Each(func(i int, opt *goquery.Selection) {
		val := opt.AttrOr("value", "")
		text := strings.TrimSpace(opt.Text())

		if val == "" || val == "select" || strings.Contains(text, "--Select") || strings.Contains(text, "Select Course") {
			return
		}

		var metadata *models.AttendanceMetadata
		metaLength := len(metadataMap["course_name_h"])
		if i < metaLength {
			metadata = &models.AttendanceMetadata{
				CourseName:     getMetaVal(metadataMap, "course_name_h", i),
				Semester:       getMetaVal(metadataMap, "semester_h", i),
				Division:       getMetaVal(metadataMap, "div_code_h", i),
				AcademicYear:   getMetaVal(metadataMap, "academic_year_h", i),
				Term:           getMetaVal(metadataMap, "term_h", i),
				DeptName:       getMetaVal(metadataMap, "dept_name_h", i),
				CourseTeacher:  getMetaVal(metadataMap, "course_teacher_h", i),
				ClassType:      getMetaVal(metadataMap, "class_type_h", i),
				ProgramCode:    getMetaVal(metadataMap, "program_code_h", i),
				ClassName:      getMetaVal(metadataMap, "class_name_h", i),
				GTUBranchCode:  getMetaVal(metadataMap, "gtu_branch_code_h", i),
				CVMUBranchCode: getMetaVal(metadataMap, "cvmu_branch_code_h", i),
				CLS:            getMetaVal(metadataMap, "cls_h", i),
			}
		}

		parts := strings.SplitN(val, ",", 2)
		courseCode := strings.TrimSpace(parts[0])
		courseName := text
		if len(parts) > 1 {
			courseName = strings.TrimSpace(parts[1])
		}

		courses = append(courses, models.AttendanceCourseOption{
			CourseCode:  courseCode,
			CourseName:  courseName,
			RawValue:    val,
			DisplayText: text,
			OptionIndex: i,
			Metadata:    metadata,
		})
	})

	return courses, nil
}

// FetchStudentList gets the student list for a course for attendance marking
func (s *GMSScraper) FetchStudentList(courseCode, date string, byLibID bool, isEdit bool, metadata *models.AttendanceMetadata, optionIndex int) ([]models.AttendanceStudentEntry, error) {
	baseURL := s.enterAttendanceURL()
	if byLibID {
		baseURL = s.enterAttendanceLibURL()
	}
	if isEdit {
		baseURL = s.editAttendanceURL()
	}

	// Step 1: GET page to prime session
	getBody, err := s.getWithRetry(baseURL)
	if err != nil {
		return nil, err
	}

	getDoc, err := goquery.NewDocumentFromReader(strings.NewReader(getBody))
	if err != nil {
		return nil, err
	}

	// Find raw option value
	matchedRawValue := courseCode
	getDoc.Find(`select[name="course_code"] option`).Each(func(_ int, opt *goquery.Selection) {
		val := opt.AttrOr("value", "")
		if val == "" || val == "select" {
			return
		}
		codeOnly := courseCode
		if idx := strings.Index(courseCode, ","); idx >= 0 {
			codeOnly = strings.TrimSpace(courseCode[:idx])
		}
		valCode := val
		if idx := strings.Index(val, ","); idx >= 0 {
			valCode = strings.TrimSpace(val[:idx])
		}
		if valCode == codeOnly {
			matchedRawValue = val
		}
	})

	// Step 2: Build form data
	form := url.Values{
		"course_code": {matchedRawValue},
		"b1":          {"Enter"},
	}

	if metadata != nil {
		form.Set("course_name", metadata.CourseName)
		form.Set("semester", metadata.Semester)
		form.Set("div_code", metadata.Division)
		form.Set("academic_year", metadata.AcademicYear)
		form.Set("term", metadata.Term)
		form.Set("dept_name", metadata.DeptName)
		form.Set("class_type", metadata.ClassType)
		form.Set("program_code", metadata.ProgramCode)
		form.Set("class_name", metadata.ClassName)
		form.Set("gtu_branch_code", metadata.GTUBranchCode)
		form.Set("cvmu_branch_code", metadata.CVMUBranchCode)
		form.Set("cls", metadata.CLS)
	}

	// Step 3: POST to get student list
	respBody, err := s.postWithRetry(baseURL, form)
	if err != nil {
		return nil, err
	}

	respDoc, err := goquery.NewDocumentFromReader(strings.NewReader(respBody))
	if err != nil {
		return nil, err
	}

	var students []models.AttendanceStudentEntry

	// Strategy 1: Portal uses individual tables per student (ctable1, ctable2, ...)
	searchRoot := respDoc.Find(`form[name="f2"]`)
	if searchRoot.Length() == 0 {
		searchRoot = respDoc.Selection
	}

	searchRoot.Find(`input[name="enrollment_no"]`).Each(func(_ int, enrollInput *goquery.Selection) {
		enrollment := strings.TrimSpace(enrollInput.AttrOr("value", ""))
		if enrollment == "" {
			return
		}

		// Find parent table
		parentTable := enrollInput.Closest("table")

		name := ""
		if parentTable.Length() > 0 {
			parentTable.Find("i").Each(func(_ int, italic *goquery.Selection) {
				if name != "" {
					return
				}
				text := strings.TrimSpace(italic.Text())
				if text != "" {
					name = text
				}
			})
		}

		students = append(students, models.AttendanceStudentEntry{
			Enrollment: enrollment,
			Name:       name,
			IsPresent:  true,
		})
	})

	// Strategy 2: Fallback - traditional table format
	if len(students) == 0 {
		respDoc.Find("table").Each(func(_ int, table *goquery.Selection) {
			if len(students) > 0 {
				return
			}
			tableText := strings.ToLower(table.Text())
			if contains(tableText, "enrollment") && (contains(tableText, "student") || contains(tableText, "name")) {
				rows := parseTableWithMapping(table)
				for _, row := range rows {
					students = append(students, models.AttendanceStudentEntry{
						Enrollment: getMapVal(row, "enrollment no", "enrollment_no", "enrollment"),
						Name:       getMapVal(row, "student name", "student_name", "name"),
						IsPresent:  true,
					})
				}
			}
		})
	}

	logger.Log.Infof("GMS Scraper: Found %d students for course %s", len(students), courseCode)
	return students, nil
}

// SubmitAttendance enters new attendance on the portal
func (s *GMSScraper) SubmitAttendance(req *models.SubmitAttendanceRequest) (bool, error) {
	// Step 1: GET enter attendance page
	getBody, err := s.getWithRetry(s.enterAttendanceURL())
	if err != nil {
		return false, err
	}

	getDoc, err := goquery.NewDocumentFromReader(strings.NewReader(getBody))
	if err != nil {
		return false, err
	}

	// Find raw value
	matchedRawValue := req.CourseCode
	getDoc.Find(`select[name="course_code"] option`).Each(func(_ int, opt *goquery.Selection) {
		val := opt.AttrOr("value", "")
		if val == "" || val == "select" {
			return
		}
		codeOnly := req.CourseCode
		if idx := strings.Index(req.CourseCode, ","); idx >= 0 {
			codeOnly = strings.TrimSpace(req.CourseCode[:idx])
		}
		valCode := val
		if idx := strings.Index(val, ","); idx >= 0 {
			valCode = strings.TrimSpace(val[:idx])
		}
		if valCode == codeOnly {
			matchedRawValue = val
		}
	})

	// Build course form data
	courseForm := url.Values{
		"course_code": {matchedRawValue},
		"b1":          {"Enter"},
	}

	if req.Metadata != nil {
		courseForm.Set("course_name", req.Metadata.CourseName)
		courseForm.Set("semester", req.Metadata.Semester)
		courseForm.Set("div_code", req.Metadata.Division)
		courseForm.Set("academic_year", req.Metadata.AcademicYear)
		courseForm.Set("term", req.Metadata.Term)
		courseForm.Set("dept_name", req.Metadata.DeptName)
		courseForm.Set("class_type", req.Metadata.ClassType)
		courseForm.Set("program_code", req.Metadata.ProgramCode)
		courseForm.Set("class_name", req.Metadata.ClassName)
		courseForm.Set("gtu_branch_code", req.Metadata.GTUBranchCode)
		courseForm.Set("cvmu_branch_code", req.Metadata.CVMUBranchCode)
		courseForm.Set("cls", req.Metadata.CLS)
	}

	// POST to get student list page (form f2)
	studentPageBody, err := s.postWithRetry(s.enterAttendanceURL(), courseForm)
	if err != nil {
		return false, err
	}

	studentDoc, err := goquery.NewDocumentFromReader(strings.NewReader(studentPageBody))
	if err != nil {
		return false, err
	}

	formF2 := studentDoc.Find(`form[name="f2"]`)
	if formF2.Length() == 0 {
		return false, fmt.Errorf("form f2 not found in response")
	}

	// Get portal date
	attDateInput := formF2.Find(`input[name="att_date"]`)
	portalDate := attDateInput.AttrOr("value", "")
	if portalDate == "" {
		parts := strings.Split(req.Date, "-")
		if len(parts) == 3 {
			day, _ := strconv.Atoi(parts[0])
			month, _ := strconv.Atoi(parts[1])
			portalDate = fmt.Sprintf("%s-%d-%d", parts[2], month, day)
		} else {
			portalDate = req.Date
		}
	}

	topicValue := req.Topic
	if strings.TrimSpace(topicValue) == "" {
		topicValue = "Regular Session"
	}

	extraLecture := "No"
	lecNo := "1"
	if req.ExtraLecture {
		extraLecture = "Yes"
		lecNo = req.LecNo
	}

	// Extract _f hidden fields
	fFields := map[string]string{}
	formF2.Find(`input[type="hidden"]`).Each(func(_ int, input *goquery.Selection) {
		name := input.AttrOr("name", "")
		value := input.AttrOr("value", "")
		if strings.HasSuffix(name, "_f") && name != "" {
			fFields[name] = value
		}
	})

	// Build submission body
	var bodyParts []string
	bodyParts = append(bodyParts, "att_date="+url.QueryEscape(portalDate))
	bodyParts = append(bodyParts, "extra_lecture="+url.QueryEscape(extraLecture))
	bodyParts = append(bodyParts, "lec_no="+url.QueryEscape(lecNo))
	bodyParts = append(bodyParts, "topic="+url.QueryEscape(topicValue))
	bodyParts = append(bodyParts, "s1=true")

	for _, student := range req.Students {
		bodyParts = append(bodyParts, "enrollment_no="+url.QueryEscape(student.Enrollment+",true,false"))
		if student.IsPresent {
			bodyParts = append(bodyParts, "enrollment_status=true")
		}
	}

	for name, value := range fFields {
		bodyParts = append(bodyParts, url.QueryEscape(name)+"="+url.QueryEscape(value))
	}
	bodyParts = append(bodyParts, "action=Save")

	encodedBody := strings.Join(bodyParts, "&")

	// Get submit URL
	formAction := formF2.AttrOr("action", "")
	submitURL := s.enterAttendanceURL()
	if formAction != "" && formAction != "null" {
		if strings.HasPrefix(formAction, "http") {
			submitURL = formAction
		} else {
			submitURL = s.baseURL + formAction
		}
	}

	respBody, err := s.doPostRaw(submitURL, encodedBody)
	if err != nil {
		return false, err
	}

	bodyLower := strings.ToLower(respBody)

	// Check for success
	if strings.Contains(bodyLower, "deleteattendance") || strings.Contains(respBody, "--Select Date--") {
		return true, nil
	}
	if strings.Contains(bodyLower, "success") || strings.Contains(bodyLower, "saved") || strings.Contains(bodyLower, "inserted") {
		return true, nil
	}
	hasF1 := strings.Contains(respBody, `name="f1"`)
	hasF2 := strings.Contains(respBody, `name="f2"`)
	if hasF1 && !hasF2 && !strings.Contains(bodyLower, "please") && !strings.Contains(bodyLower, "invalid") {
		return true, nil
	}

	return false, fmt.Errorf("could not confirm attendance submission")
}

// EditAttendance edits existing attendance
func (s *GMSScraper) EditAttendance(req *models.EditAttendanceRequest) (bool, error) {
	form := url.Values{
		"course_code":    {req.CourseCode},
		"attend_date":    {req.Date},
		"time_slot":      {req.TimeSlot},
		"topic":          {req.Topic},
		"attend_type":    {req.Type},
		"total_students": {strconv.Itoa(len(req.Students))},
	}

	for i, student := range req.Students {
		form.Set(fmt.Sprintf("enroll_%d", i), student.Enrollment)
		status := "A"
		if student.IsPresent {
			status = "P"
		}
		form.Set(fmt.Sprintf("status_%d", i), status)
	}

	body, err := s.postWithRetry(s.editAttendanceActionURL(), form)
	if err != nil {
		return false, err
	}

	bodyLower := strings.ToLower(body)
	return strings.Contains(bodyLower, "success") || strings.Contains(bodyLower, "updated") || strings.Contains(bodyLower, "saved"), nil
}

// DeleteAttendance deletes attendance for a date
func (s *GMSScraper) DeleteAttendance(req *models.DeleteAttendanceRequest) (bool, error) {
	// Prime session
	_, _ = s.getWithRetry(s.deleteAttendanceURL())

	form := url.Values{
		"course_code": {req.CourseCode},
		"attend_date": {req.Date},
		"time_slot":   {req.TimeSlot},
	}

	body, err := s.postWithRetry(s.deleteAttendanceActionURL(), form)
	if err != nil {
		return false, err
	}

	bodyLower := strings.ToLower(body)
	return strings.Contains(bodyLower, "success") || strings.Contains(bodyLower, "deleted") || strings.Contains(bodyLower, "removed"), nil
}

// FetchEditOptions gets available dates for editing a course's attendance
func (s *GMSScraper) FetchEditOptions(course *models.AttendanceCourseOption) ([]models.AttendanceEditOption, error) {
	getBody, err := s.getWithRetry(s.editAttendanceURL())
	if err != nil {
		return nil, err
	}

	getDoc, err := goquery.NewDocumentFromReader(strings.NewReader(getBody))
	if err != nil {
		return nil, err
	}

	matchedRawValue := course.RawValue
	if matchedRawValue == "" {
		matchedRawValue = course.CourseCode
	}
	getDoc.Find(`select[name="course_code"] option`).Each(func(_ int, opt *goquery.Selection) {
		val := opt.AttrOr("value", "")
		if val == "" || val == "select" {
			return
		}
		codeOnly := course.CourseCode
		if idx := strings.Index(course.CourseCode, ","); idx >= 0 {
			codeOnly = strings.TrimSpace(course.CourseCode[:idx])
		}
		valCode := val
		if idx := strings.Index(val, ","); idx >= 0 {
			valCode = strings.TrimSpace(val[:idx])
		}
		if valCode == codeOnly {
			matchedRawValue = val
		}
	})

	form := url.Values{
		"course_code": {matchedRawValue},
		"b1":          {"Enter"},
	}

	if course.Metadata != nil {
		for _, key := range []string{"course_name", "semester", "div_code", "academic_year", "term", "dept_name", "class_type", "program_code", "class_name", "gtu_branch_code", "cvmu_branch_code", "cls"} {
			var val string
			switch key {
			case "course_name":
				val = course.Metadata.CourseName
			case "semester":
				val = course.Metadata.Semester
			case "div_code":
				val = course.Metadata.Division
			case "academic_year":
				val = course.Metadata.AcademicYear
			case "term":
				val = course.Metadata.Term
			case "dept_name":
				val = course.Metadata.DeptName
			case "class_type":
				val = course.Metadata.ClassType
			case "program_code":
				val = course.Metadata.ProgramCode
			case "class_name":
				val = course.Metadata.ClassName
			case "gtu_branch_code":
				val = course.Metadata.GTUBranchCode
			case "cvmu_branch_code":
				val = course.Metadata.CVMUBranchCode
			case "cls":
				val = course.Metadata.CLS
			}
			form.Set(key+"_h", val)
		}
	}

	respBody, err := s.postWithRetry(s.editAttendanceURL(), form)
	if err != nil {
		return nil, err
	}

	respDoc, err := goquery.NewDocumentFromReader(strings.NewReader(respBody))
	if err != nil {
		return nil, err
	}

	var options []models.AttendanceEditOption

	// Try attend_date or att_date select
	sel := respDoc.Find(`select[name="attend_date"]`)
	if sel.Length() == 0 {
		sel = respDoc.Find(`select[name="att_date"]`)
	}

	sel.Find("option").Each(func(_ int, opt *goquery.Selection) {
		value := opt.AttrOr("value", "")
		text := strings.TrimSpace(opt.Text())

		if value == "" || strings.Contains(strings.ToLower(text), "select date") || strings.HasPrefix(text, "--") {
			return
		}

		parts := strings.SplitN(text, " No: ", 2)
		lectureNo := "1"
		if len(parts) > 1 {
			lectureNo = parts[1]
		}

		options = append(options, models.AttendanceEditOption{
			Date:      parts[0],
			LectureNo: lectureNo,
			RawValue:  value,
		})
	})

	return options, nil
}

// ScrapeStudentWiseAttendance gets student-wise attendance for analytics
func (s *GMSScraper) ScrapeStudentWiseAttendance(course *models.AttendanceCourseOption) ([]models.StudentAttendanceSummary, error) {
	getBody, err := s.getWithRetry(s.attendanceSheetURL())
	if err != nil {
		return nil, err
	}

	getDoc, err := goquery.NewDocumentFromReader(strings.NewReader(getBody))
	if err != nil {
		return nil, err
	}

	var targetIdx int = -1
	var targetVal string
	getDoc.Find(`select[name="cname"] option`).Each(func(i int, opt *goquery.Selection) {
		if targetIdx >= 0 {
			return
		}
		text := strings.TrimSpace(opt.Text())
		val := opt.AttrOr("value", "")
		if text == strings.TrimSpace(course.DisplayText) || (strings.Contains(val, course.CourseCode) && strings.Contains(text, course.Batch)) {
			targetIdx = i
			targetVal = val
		}
	})

	if targetIdx <= 0 {
		getDoc.Find(`select[name="cname"] option`).Each(func(i int, opt *goquery.Selection) {
			if targetIdx >= 0 {
				return
			}
			val := opt.AttrOr("value", "")
			if strings.Contains(val, course.CourseCode) {
				targetIdx = i
				targetVal = val
			}
		})
	}

	if targetIdx <= 0 || targetVal == "" {
		return nil, nil
	}

	p := strings.Split(targetVal, ",")
	if len(p) < 9 {
		return nil, nil
	}

	form := url.Values{
		"course_code":     {p[0]},
		"course_name":     {p[1]},
		"dept_name":       {p[2]},
		"div_code":        {p[3]},
		"class_type":      {p[4]},
		"class_name":      {p[5]},
		"cls":             {p[6]},
		"academic_year":   {p[7]},
		"term":            {p[8]},
		"gtu_branch_code": {""},
		"b1":              {"View"},
	}

	respBody, err := s.postWithRetry(s.attendanceSheetDetailsURL(), form)
	if err != nil {
		return nil, err
	}

	respDoc, err := goquery.NewDocumentFromReader(strings.NewReader(respBody))
	if err != nil {
		return nil, err
	}

	var summaries []models.StudentAttendanceSummary
	respDoc.Find("table").Each(func(_ int, table *goquery.Selection) {
		if len(summaries) > 0 {
			return
		}
		tableText := strings.ToLower(table.Text())
		if contains(tableText, "enrollment no") && contains(tableText, "student name") {
			table.Find("tr").Each(func(i int, row *goquery.Selection) {
				if i == 0 {
					return
				}
				cols := row.Find("td")
				if cols.Length() >= 6 {
					attended, _ := strconv.Atoi(strings.TrimSpace(cols.Eq(3).Text()))
					total, _ := strconv.Atoi(strings.TrimSpace(cols.Eq(4).Text()))
					pctStr := strings.ReplaceAll(strings.TrimSpace(cols.Eq(5).Text()), "%", "")
					pct, _ := strconv.ParseFloat(pctStr, 64)

					summaries = append(summaries, models.StudentAttendanceSummary{
						Enrollment: strings.TrimSpace(cols.Eq(1).Text()),
						Name:       strings.TrimSpace(cols.Eq(2).Text()),
						CourseCode: course.CourseCode,
						CourseName: course.CourseName,
						Attended:   attended,
						Total:      total,
						Percentage: pct,
					})
				}
			})
		}
	})

	return summaries, nil
}

// ScrapeDateWiseForCourse gets date-wise attendance for a specific course
func (s *GMSScraper) ScrapeDateWiseForCourse(course *models.AttendanceCourseOption) ([]models.DateWiseAttendance, error) {
	getBody, err := s.getWithRetry(s.dateWiseAttendanceURL())
	if err != nil {
		return nil, err
	}

	getDoc, err := goquery.NewDocumentFromReader(strings.NewReader(getBody))
	if err != nil {
		return nil, err
	}

	sel := getDoc.Find(`select[name="course_code"]`)
	if sel.Length() == 0 {
		return nil, nil
	}

	var targetIdx int = -1
	var options []*goquery.Selection
	sel.Find("option").Each(func(i int, opt *goquery.Selection) {
		options = append(options, opt)
		if targetIdx >= 0 {
			return
		}
		text := strings.TrimSpace(opt.Text())
		if text == strings.TrimSpace(course.DisplayText) {
			targetIdx = i
		}
	})

	if targetIdx <= 0 {
		for i, opt := range options {
			val := opt.AttrOr("value", "")
			if val == course.CourseCode {
				targetIdx = i
				break
			}
		}
	}

	if targetIdx <= 0 || targetIdx >= len(options) {
		return nil, nil
	}

	// Build payload using hidden _h inputs
	hiddenVal := func(name string, idx int) string {
		elements := getDoc.Find(fmt.Sprintf(`input[name="%s_h"]`, name))
		if idx < elements.Length() {
			return elements.Eq(idx).AttrOr("value", "")
		}
		return ""
	}

	form := url.Values{
		"course_code":     {options[targetIdx].AttrOr("value", "")},
		"course_name":     {hiddenVal("course_name", targetIdx)},
		"dept_name":       {hiddenVal("dept_name", targetIdx)},
		"class_type":      {hiddenVal("class_type", targetIdx)},
		"class_name":      {hiddenVal("class_name", targetIdx)},
		"div_code":        {hiddenVal("div_code", targetIdx)},
		"gtu_branch_code": {hiddenVal("gtu_branch_code", targetIdx)},
		"semester":        {hiddenVal("semester", targetIdx)},
		"cls":             {hiddenVal("cls", targetIdx)},
		"academic_year":   {hiddenVal("academic_year", targetIdx)},
		"term":            {hiddenVal("term", targetIdx)},
		"program_code":    {hiddenVal("program_code", targetIdx)},
		"course_teacher":  {hiddenVal("course_teacher", targetIdx)},
		"b1":              {"Enter"},
	}

	respBody, err := s.doPost(s.dateWiseAttendanceURL(), form)
	if err != nil {
		return nil, err
	}

	respDoc, err := goquery.NewDocumentFromReader(strings.NewReader(respBody))
	if err != nil {
		return nil, err
	}

	var records []models.DateWiseAttendance
	respDoc.Find("table").Each(func(_ int, table *goquery.Selection) {
		if len(records) > 0 {
			return
		}
		tableText := strings.ToLower(table.Text())
		if contains(tableText, "date") && (contains(tableText, "present") || contains(tableText, "attendance")) && !contains(tableText, "select course") {
			rows := parseTableWithMapping(table)
			for _, row := range rows {
				present, _ := strconv.Atoi(getMapVal(row, "present"))
				absent, _ := strconv.Atoi(getMapVal(row, "absent"))
				total, _ := strconv.Atoi(getMapVal(row, "total"))

				records = append(records, models.DateWiseAttendance{
					Date:       getMapVal(row, "date"),
					CourseCode: getMapVal(row, "course code", "course_code"),
					CourseName: getMapVal(row, "course name", "course_name"),
					TimeSlot:   getMapVal(row, "time slot", "time_slot"),
					Type:       getMapValDefault(row, "L", "type"),
					Present:    present,
					Absent:     absent,
					Total:      total,
					Topic:      getMapVal(row, "topic"),
				})
			}
		}
	})

	return records, nil
}

// UpdateProfile updates profile data on the portal
func (s *GMSScraper) UpdateProfile(data map[string]string) (bool, error) {
	form := url.Values{}
	for k, v := range data {
		form.Set(k, v)
	}
	body, err := s.postWithRetry(s.saveProfileURL(), form)
	if err != nil {
		return false, err
	}
	bodyLower := strings.ToLower(body)
	return strings.Contains(bodyLower, "success") || strings.Contains(bodyLower, "updated") || strings.Contains(bodyLower, "saved"), nil
}

// UpdateQualification updates qualification on the portal
func (s *GMSScraper) UpdateQualification(data map[string]string) (bool, error) {
	form := url.Values{}
	for k, v := range data {
		form.Set(k, v)
	}
	body, err := s.postWithRetry(s.saveQualificationURL(), form)
	if err != nil {
		return false, err
	}
	bodyLower := strings.ToLower(body)
	return strings.Contains(bodyLower, "success") || strings.Contains(bodyLower, "updated") || strings.Contains(bodyLower, "saved"), nil
}

// UpdateExperience updates experience on the portal
func (s *GMSScraper) UpdateExperience(data map[string]string) (bool, error) {
	form := url.Values{}
	for k, v := range data {
		form.Set(k, v)
	}
	body, err := s.postWithRetry(s.saveExperienceURL(), form)
	if err != nil {
		return false, err
	}
	bodyLower := strings.ToLower(body)
	return strings.Contains(bodyLower, "success") || strings.Contains(bodyLower, "updated") || strings.Contains(bodyLower, "saved"), nil
}

// ===== HTML Parsing Helpers =====

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// parseTableWithMapping parses an HTML table into a list of maps based on header text
func parseTableWithMapping(table *goquery.Selection) []map[string]string {
	var results []map[string]string
	rows := table.Find("tr")
	if rows.Length() == 0 {
		return results
	}

	// Get headers from first row
	var headers []string
	rows.First().Find("th, td").Each(func(_ int, cell *goquery.Selection) {
		headers = append(headers, strings.ToLower(strings.TrimSpace(cell.Text())))
	})

	if len(headers) == 0 {
		return results
	}

	rows.Each(func(i int, row *goquery.Selection) {
		if i == 0 {
			return
		}
		cols := row.Find("td")
		if cols.Length() == 0 {
			return
		}

		rowMap := map[string]string{}
		for j := 0; j < len(headers); j++ {
			if j < cols.Length() && headers[j] != "" {
				rowMap[headers[j]] = strings.TrimSpace(cols.Eq(j).Text())
			}
		}
		if len(rowMap) > 0 {
			results = append(results, rowMap)
		}
	})

	return results
}

func findInCols(headers []string, cols *goquery.Selection, search string) string {
	for i, h := range headers {
		if contains(h, search) && i < cols.Length() {
			return strings.TrimSpace(cols.Eq(i).Text())
		}
	}
	return ""
}

func getMapVal(m map[string]string, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k]; ok && v != "" {
			return v
		}
	}
	return ""
}

func getMapValDefault(m map[string]string, defaultVal string, keys ...string) string {
	v := getMapVal(m, keys...)
	if v == "" {
		return defaultVal
	}
	return v
}

func getMetaVal(m map[string][]string, key string, index int) string {
	if vals, ok := m[key]; ok && index < len(vals) {
		return strings.TrimSpace(vals[index])
	}
	return ""
}

// Compile regex patterns once
var emailRegex = regexp.MustCompile(`[\w.\-]+@[\w.\-]+\.\w+`)
var welcomeRegex = regexp.MustCompile(`(?i)welcome\s+(?:prof\.?\s+)?(\w[\w\s]+)`)
