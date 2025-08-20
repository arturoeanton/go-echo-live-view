package auth

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrTokenExpired       = errors.New("token expired")
	ErrInvalidToken       = errors.New("invalid token")
	ErrSessionNotFound    = errors.New("session not found")
	ErrUserNotFound       = errors.New("user not found")
)

type Role string

const (
	RoleAdmin     Role = "admin"
	RoleModerator Role = "moderator"
	RoleUser      Role = "user"
	RoleGuest     Role = "guest"
)

type Permission string

const (
	PermissionRead   Permission = "read"
	PermissionWrite  Permission = "write"
	PermissionDelete Permission = "delete"
	PermissionAdmin  Permission = "admin"
)

type User struct {
	ID          string
	Username    string
	Email       string
	Password    string
	Roles       []Role
	Permissions map[Permission]bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	LastLogin   time.Time
	IsActive    bool
	Metadata    map[string]interface{}
}

type Session struct {
	ID        string
	UserID    string
	Token     string
	ExpiresAt time.Time
	CreatedAt time.Time
	IP        string
	UserAgent string
	IsValid   bool
}

type AuthManager struct {
	users           map[string]*User
	sessions        map[string]*Session
	usersByUsername map[string]*User
	jwtSecret       []byte
	sessionTimeout  time.Duration
	mu              sync.RWMutex
	rolePermissions map[Role][]Permission
}

func NewAuthManager(jwtSecret string, sessionTimeout time.Duration) *AuthManager {
	if jwtSecret == "" {
		secret := make([]byte, 32)
		rand.Read(secret)
		jwtSecret = base64.StdEncoding.EncodeToString(secret)
	}

	am := &AuthManager{
		users:           make(map[string]*User),
		sessions:        make(map[string]*Session),
		usersByUsername: make(map[string]*User),
		jwtSecret:       []byte(jwtSecret),
		sessionTimeout:  sessionTimeout,
		rolePermissions: map[Role][]Permission{
			RoleAdmin:     {PermissionRead, PermissionWrite, PermissionDelete, PermissionAdmin},
			RoleModerator: {PermissionRead, PermissionWrite, PermissionDelete},
			RoleUser:      {PermissionRead, PermissionWrite},
			RoleGuest:     {PermissionRead},
		},
	}

	go am.cleanupExpiredSessions()

	return am
}

func (am *AuthManager) CreateUser(username, email, password string, roles []Role) (*User, error) {
	am.mu.Lock()
	defer am.mu.Unlock()

	if _, exists := am.usersByUsername[username]; exists {
		return nil, errors.New("username already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	userID := generateID()
	permissions := make(map[Permission]bool)

	for _, role := range roles {
		if perms, ok := am.rolePermissions[role]; ok {
			for _, perm := range perms {
				permissions[perm] = true
			}
		}
	}

	user := &User{
		ID:          userID,
		Username:    username,
		Email:       email,
		Password:    string(hashedPassword),
		Roles:       roles,
		Permissions: permissions,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		IsActive:    true,
		Metadata:    make(map[string]interface{}),
	}

	am.users[userID] = user
	am.usersByUsername[username] = user

	return user, nil
}

func (am *AuthManager) Authenticate(username, password string) (*User, error) {
	am.mu.RLock()
	defer am.mu.RUnlock()

	user, exists := am.usersByUsername[username]
	if !exists {
		return nil, ErrInvalidCredentials
	}

	if !user.IsActive {
		return nil, ErrUnauthorized
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	user.LastLogin = time.Now()
	return user, nil
}

func (am *AuthManager) CreateSession(user *User, ip, userAgent string) (*Session, error) {
	am.mu.Lock()
	defer am.mu.Unlock()

	sessionID := generateID()
	token, err := am.generateJWT(user)
	if err != nil {
		return nil, err
	}

	session := &Session{
		ID:        sessionID,
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: time.Now().Add(am.sessionTimeout),
		CreatedAt: time.Now(),
		IP:        ip,
		UserAgent: userAgent,
		IsValid:   true,
	}

	am.sessions[sessionID] = session
	return session, nil
}

func (am *AuthManager) ValidateSession(sessionID string) (*Session, *User, error) {
	am.mu.RLock()
	defer am.mu.RUnlock()

	session, exists := am.sessions[sessionID]
	if !exists {
		return nil, nil, ErrSessionNotFound
	}

	if !session.IsValid || time.Now().After(session.ExpiresAt) {
		return nil, nil, ErrTokenExpired
	}

	user, exists := am.users[session.UserID]
	if !exists {
		return nil, nil, ErrUserNotFound
	}

	return session, user, nil
}

func (am *AuthManager) InvalidateSession(sessionID string) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	session, exists := am.sessions[sessionID]
	if !exists {
		return ErrSessionNotFound
	}

	session.IsValid = false
	return nil
}

func (am *AuthManager) generateJWT(user *User) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"email":    user.Email,
		"roles":    user.Roles,
		"exp":      time.Now().Add(am.sessionTimeout).Unix(),
		"iat":      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(am.jwtSecret)
}

func (am *AuthManager) ValidateJWT(tokenString string) (*User, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return am.jwtSecret, nil
	})

	if err != nil {
		return nil, ErrInvalidToken
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID, ok := claims["user_id"].(string)
		if !ok {
			return nil, ErrInvalidToken
		}

		am.mu.RLock()
		defer am.mu.RUnlock()

		user, exists := am.users[userID]
		if !exists {
			return nil, ErrUserNotFound
		}

		return user, nil
	}

	return nil, ErrInvalidToken
}

func (am *AuthManager) HasPermission(user *User, permission Permission) bool {
	if user == nil {
		return false
	}

	return user.Permissions[permission]
}

func (am *AuthManager) HasRole(user *User, role Role) bool {
	if user == nil {
		return false
	}

	for _, r := range user.Roles {
		if r == role {
			return true
		}
	}

	return false
}

func (am *AuthManager) UpdateUserRoles(userID string, roles []Role) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	user, exists := am.users[userID]
	if !exists {
		return ErrUserNotFound
	}

	permissions := make(map[Permission]bool)
	for _, role := range roles {
		if perms, ok := am.rolePermissions[role]; ok {
			for _, perm := range perms {
				permissions[perm] = true
			}
		}
	}

	user.Roles = roles
	user.Permissions = permissions
	user.UpdatedAt = time.Now()

	return nil
}

func (am *AuthManager) cleanupExpiredSessions() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		am.mu.Lock()
		now := time.Now()
		for id, session := range am.sessions {
			if now.After(session.ExpiresAt) || !session.IsValid {
				delete(am.sessions, id)
			}
		}
		am.mu.Unlock()
	}
}

func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

type contextKey string

const userContextKey contextKey = "user"

func ContextWithUser(ctx context.Context, user *User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

func UserFromContext(ctx context.Context) (*User, bool) {
	user, ok := ctx.Value(userContextKey).(*User)
	return user, ok
}

func ConstantTimeCompare(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}