package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"enginenosql/internal/engine"
)

// AuthService handles user authentication
type AuthService struct {
	engine   *engine.Engine
	sessions map[string]*Session // sessionID -> Session
}

// User represents a user in the system
type User struct {
	ID           string `json:"id"`
	Username     string `json:"username"`
	Email        string `json:"email"`
	PasswordHash string `json:"password_hash"`
	Salt         string `json:"salt"`
	CreatedAt    string `json:"created_at"`
	LastLogin    string `json:"last_login"`
	IsActive     bool   `json:"is_active"`
}

// Session represents a user session
type Session struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	Username  string `json:"username"`
	CreatedAt string `json:"created_at"`
	ExpiresAt string `json:"expires_at"`
	IsActive  bool   `json:"is_active"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// RegisterRequest represents a registration request
type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	SessionID string `json:"session_id,omitempty"`
	User      *User  `json:"user,omitempty"`
}

// Helper functions for time conversion
func timeToString(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}

func stringToTime(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}
	return time.Parse(time.RFC3339, s)
}

// NewAuthService creates a new authentication service
func NewAuthService() (*AuthService, error) {
	// Get user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %v", err)
	}

	// Create auth data directory
	authDir := filepath.Join(homeDir, ".enginenosql", "auth")
	authEngine := engine.NewEngine(authDir)

	// Try to get existing database first
	systemDB, err := authEngine.GetDatabase("system")
	if err != nil {
		// Database doesn't exist, create it
		err = authEngine.CreateDatabase("system")
		if err != nil {
			return nil, fmt.Errorf("failed to create system database: %v", err)
		}

		// Get the newly created database
		systemDB, err = authEngine.GetDatabase("system")
		if err != nil {
			return nil, fmt.Errorf("failed to get system database: %v", err)
		}

		// Create users collection
		err = systemDB.CreateCollection("users")
		if err != nil {
			return nil, fmt.Errorf("failed to create users collection: %v", err)
		}

		// Create sessions collection
		err = systemDB.CreateCollection("sessions")
		if err != nil {
			return nil, fmt.Errorf("failed to create sessions collection: %v", err)
		}

		// Create indexes
		usersCollection, _ := systemDB.GetCollection("users")
		usersCollection.CreateIndex("username")
		usersCollection.CreateIndex("email")

		sessionsCollection, _ := systemDB.GetCollection("sessions")
		sessionsCollection.CreateIndex("user_id")

		// Save new database structure
		err = authEngine.SaveDatabase("system")
		if err != nil {
			return nil, fmt.Errorf("failed to save system database: %v", err)
		}
	}

	return &AuthService{
		engine:   authEngine,
		sessions: make(map[string]*Session),
	}, nil
}

// Register registers a new user
func (a *AuthService) Register(req RegisterRequest) (*LoginResponse, error) {
	// Validate input
	if req.Username == "" || req.Email == "" || req.Password == "" {
		return &LoginResponse{
			Success: false,
			Message: "Username, email, and password are required",
		}, nil
	}

	if len(req.Password) < 6 {
		return &LoginResponse{
			Success: false,
			Message: "Password must be at least 6 characters long",
		}, nil
	}

	// Get system database
	systemDB, err := a.engine.GetDatabase("system")
	if err != nil {
		return nil, fmt.Errorf("failed to get system database: %v", err)
	}

	usersCollection, err := systemDB.GetCollection("users")
	if err != nil {
		return nil, fmt.Errorf("failed to get users collection: %v", err)
	}

	// Check if username already exists
	existingUsers, err := usersCollection.Find("username", req.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing username: %v", err)
	}

	if len(existingUsers) > 0 {
		return &LoginResponse{
			Success: false,
			Message: "Username already exists",
		}, nil
	}

	// Check if email already exists
	existingEmails, err := usersCollection.Find("email", req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing email: %v", err)
	}

	if len(existingEmails) > 0 {
		return &LoginResponse{
			Success: false,
			Message: "Email already exists",
		}, nil
	}

	// Generate salt and hash password
	salt, err := generateSalt()
	if err != nil {
		return nil, fmt.Errorf("failed to generate salt: %v", err)
	}

	passwordHash := hashPassword(req.Password, salt)

	// Create user
	userID := generateID()
	user := User{
		ID:           userID,
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: passwordHash,
		Salt:         salt,
		CreatedAt:    timeToString(time.Now()),
		LastLogin:    timeToString(time.Time{}),
		IsActive:     true,
	}

	// Convert user to map for storage
	userData := map[string]interface{}{
		"username":      user.Username,
		"email":         user.Email,
		"password_hash": user.PasswordHash,
		"salt":          user.Salt,
		"created_at":    user.CreatedAt,
		"last_login":    user.LastLogin,
		"is_active":     user.IsActive,
	}

	// Save user
	err = usersCollection.Insert(userID, userData)
	if err != nil {
		return nil, fmt.Errorf("failed to save user: %v", err)
	}

	// Save database
	err = a.engine.SaveDatabase("system")
	if err != nil {
		return nil, fmt.Errorf("failed to save database: %v", err)
	}

	return &LoginResponse{
		Success: true,
		Message: "User registered successfully",
	}, nil
}

// Login authenticates a user and creates a session
func (a *AuthService) Login(req LoginRequest) (*LoginResponse, error) {
	// Validate input
	if req.Username == "" || req.Password == "" {
		return &LoginResponse{
			Success: false,
			Message: "Username and password are required",
		}, nil
	}

	// Get system database
	systemDB, err := a.engine.GetDatabase("system")
	if err != nil {
		return nil, fmt.Errorf("failed to get system database: %v", err)
	}

	usersCollection, err := systemDB.GetCollection("users")
	if err != nil {
		return nil, fmt.Errorf("failed to get users collection: %v", err)
	}

	// Find user by username
	userDocs, err := usersCollection.Find("username", req.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %v", err)
	}

	if len(userDocs) == 0 {
		return &LoginResponse{
			Success: false,
			Message: "Invalid username or password",
		}, nil
	}

	userDoc := userDocs[0]

	// Verify password
	storedHash, ok := userDoc.Data["password_hash"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid password hash format")
	}

	salt, ok := userDoc.Data["salt"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid salt format")
	}

	if !verifyPassword(req.Password, salt, storedHash) {
		return &LoginResponse{
			Success: false,
			Message: "Invalid username or password",
		}, nil
	}

	// Check if user is active
	isActive, ok := userDoc.Data["is_active"].(bool)
	if !ok || !isActive {
		return &LoginResponse{
			Success: false,
			Message: "Account is inactive",
		}, nil
	}

	// Create session
	sessionID := generateSessionID()
	session := &Session{
		ID:        sessionID,
		UserID:    userDoc.ID,
		Username:  req.Username,
		CreatedAt: timeToString(time.Now()),
		ExpiresAt: timeToString(time.Now().Add(24 * time.Hour)), // 24 hour session
		IsActive:  true,
	}

	// Store session in memory
	a.sessions[sessionID] = session

	// Store session in database
	sessionsCollection, err := systemDB.GetCollection("sessions")
	if err != nil {
		return nil, fmt.Errorf("failed to get sessions collection: %v", err)
	}

	sessionData := map[string]interface{}{
		"user_id":    session.UserID,
		"username":   session.Username,
		"created_at": session.CreatedAt,
		"expires_at": session.ExpiresAt,
		"is_active":  session.IsActive,
	}

	err = sessionsCollection.Insert(sessionID, sessionData)
	if err != nil {
		return nil, fmt.Errorf("failed to save session: %v", err)
	}

	// Update user's last login
	userDoc.Data["last_login"] = timeToString(time.Now())
	err = usersCollection.Update(userDoc.ID, userDoc.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to update last login: %v", err)
	}

	// Save database
	err = a.engine.SaveDatabase("system")
	if err != nil {
		return nil, fmt.Errorf("failed to save database: %v", err)
	}

	// Create user object for response
	createdAtStr, _ := userDoc.Data["created_at"].(string)
	user := &User{
		ID:        userDoc.ID,
		Username:  req.Username,
		Email:     userDoc.Data["email"].(string),
		CreatedAt: createdAtStr,
		LastLogin: timeToString(time.Now()),
		IsActive:  true,
	}

	return &LoginResponse{
		Success:   true,
		Message:   "Login successful",
		SessionID: sessionID,
		User:      user,
	}, nil
}

// ValidateSession validates a session ID
func (a *AuthService) ValidateSession(sessionID string) (*Session, error) {
	// Check memory first
	if session, exists := a.sessions[sessionID]; exists {
		if session.IsActive {
			expiresAt, err := stringToTime(session.ExpiresAt)
			if err == nil && time.Now().Before(expiresAt) {
				return session, nil
			}
		}
		// Remove expired session
		delete(a.sessions, sessionID)
	}

	// Check database
	systemDB, err := a.engine.GetDatabase("system")
	if err != nil {
		return nil, fmt.Errorf("failed to get system database: %v", err)
	}

	sessionsCollection, err := systemDB.GetCollection("sessions")
	if err != nil {
		return nil, fmt.Errorf("failed to get sessions collection: %v", err)
	}

	// Find session in database
	sessionDocs := sessionsCollection.GetAll()
	for _, sessionDoc := range sessionDocs {
		if sessionDoc.ID == sessionID {
			expiresAtStr, ok := sessionDoc.Data["expires_at"].(string)
			if !ok {
				continue
			}

			expiresAt, err := stringToTime(expiresAtStr)
			if err != nil {
				continue
			}

			isActive, ok := sessionDoc.Data["is_active"].(bool)
			if !ok || !isActive {
				continue
			}

			if time.Now().Before(expiresAt) {
				createdAtStr, _ := sessionDoc.Data["created_at"].(string)
				session := &Session{
					ID:        sessionDoc.ID,
					UserID:    sessionDoc.Data["user_id"].(string),
					Username:  sessionDoc.Data["username"].(string),
					CreatedAt: createdAtStr,
					ExpiresAt: expiresAtStr,
					IsActive:  isActive,
				}

				// Store in memory for faster access
				a.sessions[sessionID] = session
				return session, nil
			}
		}
	}

	return nil, fmt.Errorf("session not found or expired")
}

// Logout logs out a user by invalidating their session
func (a *AuthService) Logout(sessionID string) error {
	// Remove from memory
	delete(a.sessions, sessionID)

	// Mark as inactive in database
	systemDB, err := a.engine.GetDatabase("system")
	if err != nil {
		return fmt.Errorf("failed to get system database: %v", err)
	}

	sessionsCollection, err := systemDB.GetCollection("sessions")
	if err != nil {
		return fmt.Errorf("failed to get sessions collection: %v", err)
	}

	// Find and update session
	sessionDocs := sessionsCollection.GetAll()
	for _, sessionDoc := range sessionDocs {
		if sessionDoc.ID == sessionID {
			sessionDoc.Data["is_active"] = false
			err = sessionsCollection.Update(sessionDoc.ID, sessionDoc.Data)
			if err != nil {
				return fmt.Errorf("failed to update session: %v", err)
			}
			break
		}
	}

	// Save database
	return a.engine.SaveDatabase("system")
}

// Helper functions

func generateSalt() (string, error) {
	bytes := make([]byte, 16)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func generateID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func generateSessionID() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func hashPassword(password, salt string) string {
	hash := sha256.Sum256([]byte(password + salt))
	return hex.EncodeToString(hash[:])
}

func verifyPassword(password, salt, storedHash string) bool {
	return hashPassword(password, salt) == storedHash
}
