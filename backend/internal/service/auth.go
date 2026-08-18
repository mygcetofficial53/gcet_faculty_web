package service

import (
	"fmt"
	"sync"
	"time"

	"gcet-web-backend/internal/config"
	"gcet-web-backend/internal/logger"
	"gcet-web-backend/internal/models"

	"github.com/golang-jwt/jwt/v5"
)

// AuthService handles JWT tokens and faculty session management
type AuthService struct {
	cfg      *config.Config
	sessions map[string]*FacultySession // sessionID → session
	mu       sync.RWMutex
}

// FacultySession holds a live scraper session for a faculty member
type FacultySession struct {
	Scraper    *GMSScraper
	Faculty    *models.Faculty
	LoginID    string
	LastAccess time.Time
	SessionID  string
}

// NewAuthService creates a new auth service
func NewAuthService(cfg *config.Config) *AuthService {
	svc := &AuthService{
		cfg:      cfg,
		sessions: make(map[string]*FacultySession),
	}
	// Start session cleanup goroutine
	go svc.cleanupLoop()
	return svc
}

// Login authenticates a faculty member and returns JWT tokens
func (a *AuthService) Login(loginID, password string) (*models.LoginResponse, error) {
	scraper := NewGMSScraper(a.cfg.GMSPortalURL)

	faculty, err := scraper.Login(loginID, password)
	if err != nil {
		return &models.LoginResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	// Create session
	sessionID := fmt.Sprintf("%s_%d", loginID, time.Now().UnixNano())

	a.mu.Lock()
	// Remove old sessions for the same user
	for id, sess := range a.sessions {
		if sess.LoginID == loginID {
			sess.Scraper.Logout()
			delete(a.sessions, id)
		}
	}
	a.sessions[sessionID] = &FacultySession{
		Scraper:    scraper,
		Faculty:    faculty,
		LoginID:    loginID,
		LastAccess: time.Now(),
		SessionID:  sessionID,
	}
	a.mu.Unlock()

	// Generate JWT tokens
	token, err := a.generateToken(loginID, faculty.EmployeeID, faculty.Name, sessionID, a.cfg.JWTExpiry)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	refreshToken, err := a.generateToken(loginID, faculty.EmployeeID, faculty.Name, sessionID, a.cfg.JWTRefreshExp)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	logger.Log.Infof("AuthService: Login successful for %s (session: %s)", loginID, sessionID)

	return &models.LoginResponse{
		Success:      true,
		Token:        token,
		RefreshToken: refreshToken,
		Faculty:      faculty,
	}, nil
}

// Logout ends a faculty session
func (a *AuthService) Logout(sessionID string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if sess, ok := a.sessions[sessionID]; ok {
		sess.Scraper.Logout()
		delete(a.sessions, sessionID)
		logger.Log.Infof("AuthService: Logout for session %s", sessionID)
	}
}

// GetSession retrieves a live session by session ID
func (a *AuthService) GetSession(sessionID string) (*FacultySession, error) {
	a.mu.RLock()
	sess, ok := a.sessions[sessionID]
	a.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("session not found or expired")
	}

	// Update last access
	a.mu.Lock()
	sess.LastAccess = time.Now()
	a.mu.Unlock()

	return sess, nil
}

// RefreshToken generates a new token pair from a valid refresh token
func (a *AuthService) RefreshToken(refreshTokenStr string) (*models.LoginResponse, error) {
	claims, err := a.ValidateToken(refreshTokenStr)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	sess, err := a.GetSession(claims.SessionID)
	if err != nil {
		return nil, fmt.Errorf("session expired — please login again")
	}

	token, err := a.generateToken(claims.LoginID, claims.EmployeeID, claims.Name, claims.SessionID, a.cfg.JWTExpiry)
	if err != nil {
		return nil, err
	}

	refreshToken, err := a.generateToken(claims.LoginID, claims.EmployeeID, claims.Name, claims.SessionID, a.cfg.JWTRefreshExp)
	if err != nil {
		return nil, err
	}

	return &models.LoginResponse{
		Success:      true,
		Token:        token,
		RefreshToken: refreshToken,
		Faculty:      sess.Faculty,
	}, nil
}

// ValidateToken validates a JWT token and returns claims
func (a *AuthService) ValidateToken(tokenStr string) (*models.TokenClaims, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(a.cfg.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return &models.TokenClaims{
			LoginID:    claimStr(claims, "login_id"),
			EmployeeID: claimStr(claims, "employee_id"),
			Name:       claimStr(claims, "name"),
			SessionID:  claimStr(claims, "session_id"),
		}, nil
	}

	return nil, fmt.Errorf("invalid token claims")
}

func (a *AuthService) generateToken(loginID, empID, name, sessionID string, expiry time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"login_id":    loginID,
		"employee_id": empID,
		"name":        name,
		"session_id":  sessionID,
		"exp":         time.Now().Add(expiry).Unix(),
		"iat":         time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(a.cfg.JWTSecret))
}

func claimStr(claims jwt.MapClaims, key string) string {
	if v, ok := claims[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// cleanupLoop removes stale sessions every 10 minutes
func (a *AuthService) cleanupLoop() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		a.mu.Lock()
		now := time.Now()
		for id, sess := range a.sessions {
			// Remove sessions inactive for more than 2 hours
			if now.Sub(sess.LastAccess) > 2*time.Hour {
				sess.Scraper.Logout()
				delete(a.sessions, id)
				logger.Log.Infof("AuthService: Cleaned up stale session %s", id)
			}
		}
		a.mu.Unlock()
	}
}
