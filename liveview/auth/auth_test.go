package auth

import (
	"context"
	"testing"
	"time"
)

func TestNewAuthManager(t *testing.T) {
	am := NewAuthManager("test-secret", 1*time.Hour)
	
	if am == nil {
		t.Fatal("Expected auth manager to be created")
	}
	
	if len(am.users) != 0 {
		t.Errorf("Expected empty users map, got %d users", len(am.users))
	}
	
	if am.sessionTimeout != 1*time.Hour {
		t.Errorf("Expected session timeout of 1 hour, got %v", am.sessionTimeout)
	}
}

func TestCreateUser(t *testing.T) {
	am := NewAuthManager("test-secret", 1*time.Hour)
	
	tests := []struct {
		name     string
		username string
		email    string
		password string
		roles    []Role
		wantErr  bool
	}{
		{
			name:     "Valid user creation",
			username: "testuser",
			email:    "test@example.com",
			password: "password123",
			roles:    []Role{RoleUser},
			wantErr:  false,
		},
		{
			name:     "Admin user creation",
			username: "admin",
			email:    "admin@example.com",
			password: "admin123",
			roles:    []Role{RoleAdmin},
			wantErr:  false,
		},
		{
			name:     "Multiple roles",
			username: "moderator",
			email:    "mod@example.com",
			password: "mod123",
			roles:    []Role{RoleModerator, RoleUser},
			wantErr:  false,
		},
		{
			name:     "Duplicate username",
			username: "testuser",
			email:    "another@example.com",
			password: "pass123",
			roles:    []Role{RoleUser},
			wantErr:  true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := am.CreateUser(tt.username, tt.email, tt.password, tt.roles)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			
			if user.Username != tt.username {
				t.Errorf("Expected username %s, got %s", tt.username, user.Username)
			}
			
			if user.Email != tt.email {
				t.Errorf("Expected email %s, got %s", tt.email, user.Email)
			}
			
			if !user.IsActive {
				t.Error("Expected user to be active")
			}
			
			for _, role := range tt.roles {
				if !am.HasRole(user, role) {
					t.Errorf("Expected user to have role %s", role)
				}
			}
		})
	}
}

func TestAuthenticate(t *testing.T) {
	am := NewAuthManager("test-secret", 1*time.Hour)
	
	testUser, _ := am.CreateUser("testuser", "test@example.com", "password123", []Role{RoleUser})
	
	tests := []struct {
		name     string
		username string
		password string
		wantErr  bool
	}{
		{
			name:     "Valid credentials",
			username: "testuser",
			password: "password123",
			wantErr:  false,
		},
		{
			name:     "Invalid password",
			username: "testuser",
			password: "wrongpassword",
			wantErr:  true,
		},
		{
			name:     "Non-existent user",
			username: "nonexistent",
			password: "password123",
			wantErr:  true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := am.Authenticate(tt.username, tt.password)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			
			if user.Username != testUser.Username {
				t.Errorf("Expected username %s, got %s", testUser.Username, user.Username)
			}
		})
	}
}

func TestSession(t *testing.T) {
	am := NewAuthManager("test-secret", 1*time.Hour)
	
	user, _ := am.CreateUser("testuser", "test@example.com", "password123", []Role{RoleUser})
	
	t.Run("Create session", func(t *testing.T) {
		session, err := am.CreateSession(user, "127.0.0.1", "TestAgent")
		
		if err != nil {
			t.Fatalf("Failed to create session: %v", err)
		}
		
		if session.UserID != user.ID {
			t.Errorf("Expected user ID %s, got %s", user.ID, session.UserID)
		}
		
		if session.IP != "127.0.0.1" {
			t.Errorf("Expected IP 127.0.0.1, got %s", session.IP)
		}
		
		if !session.IsValid {
			t.Error("Expected session to be valid")
		}
		
		t.Run("Validate session", func(t *testing.T) {
			validSession, validUser, err := am.ValidateSession(session.ID)
			
			if err != nil {
				t.Fatalf("Failed to validate session: %v", err)
			}
			
			if validSession.ID != session.ID {
				t.Errorf("Expected session ID %s, got %s", session.ID, validSession.ID)
			}
			
			if validUser.ID != user.ID {
				t.Errorf("Expected user ID %s, got %s", user.ID, validUser.ID)
			}
		})
		
		t.Run("Invalidate session", func(t *testing.T) {
			err := am.InvalidateSession(session.ID)
			
			if err != nil {
				t.Fatalf("Failed to invalidate session: %v", err)
			}
			
			_, _, err = am.ValidateSession(session.ID)
			
			if err != ErrTokenExpired {
				t.Errorf("Expected ErrTokenExpired, got %v", err)
			}
		})
	})
	
	t.Run("Session not found", func(t *testing.T) {
		_, _, err := am.ValidateSession("non-existent-session")
		
		if err != ErrSessionNotFound {
			t.Errorf("Expected ErrSessionNotFound, got %v", err)
		}
	})
}

func TestJWT(t *testing.T) {
	am := NewAuthManager("test-secret", 1*time.Hour)
	
	user, _ := am.CreateUser("testuser", "test@example.com", "password123", []Role{RoleUser})
	
	t.Run("Generate and validate JWT", func(t *testing.T) {
		token, err := am.generateJWT(user)
		
		if err != nil {
			t.Fatalf("Failed to generate JWT: %v", err)
		}
		
		if token == "" {
			t.Error("Expected non-empty token")
		}
		
		validatedUser, err := am.ValidateJWT(token)
		
		if err != nil {
			t.Fatalf("Failed to validate JWT: %v", err)
		}
		
		if validatedUser.ID != user.ID {
			t.Errorf("Expected user ID %s, got %s", user.ID, validatedUser.ID)
		}
	})
	
	t.Run("Invalid JWT", func(t *testing.T) {
		_, err := am.ValidateJWT("invalid-token")
		
		if err != ErrInvalidToken {
			t.Errorf("Expected ErrInvalidToken, got %v", err)
		}
	})
}

func TestPermissions(t *testing.T) {
	am := NewAuthManager("test-secret", 1*time.Hour)
	
	adminUser, _ := am.CreateUser("admin", "admin@example.com", "admin123", []Role{RoleAdmin})
	userUser, _ := am.CreateUser("user", "user@example.com", "user123", []Role{RoleUser})
	guestUser, _ := am.CreateUser("guest", "guest@example.com", "guest123", []Role{RoleGuest})
	
	tests := []struct {
		name       string
		user       *User
		permission Permission
		expected   bool
	}{
		{"Admin has admin permission", adminUser, PermissionAdmin, true},
		{"Admin has write permission", adminUser, PermissionWrite, true},
		{"Admin has read permission", adminUser, PermissionRead, true},
		{"User has write permission", userUser, PermissionWrite, true},
		{"User has read permission", userUser, PermissionRead, true},
		{"User lacks admin permission", userUser, PermissionAdmin, false},
		{"Guest has read permission", guestUser, PermissionRead, true},
		{"Guest lacks write permission", guestUser, PermissionWrite, false},
		{"Guest lacks admin permission", guestUser, PermissionAdmin, false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := am.HasPermission(tt.user, tt.permission)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestRoles(t *testing.T) {
	am := NewAuthManager("test-secret", 1*time.Hour)
	
	user, _ := am.CreateUser("testuser", "test@example.com", "password123", []Role{RoleUser, RoleModerator})
	
	tests := []struct {
		name     string
		role     Role
		expected bool
	}{
		{"Has user role", RoleUser, true},
		{"Has moderator role", RoleModerator, true},
		{"Lacks admin role", RoleAdmin, false},
		{"Lacks guest role", RoleGuest, false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := am.HasRole(user, tt.role)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestUpdateUserRoles(t *testing.T) {
	am := NewAuthManager("test-secret", 1*time.Hour)
	
	user, _ := am.CreateUser("testuser", "test@example.com", "password123", []Role{RoleUser})
	
	t.Run("Update roles successfully", func(t *testing.T) {
		err := am.UpdateUserRoles(user.ID, []Role{RoleAdmin, RoleModerator})
		
		if err != nil {
			t.Fatalf("Failed to update roles: %v", err)
		}
		
		if !am.HasRole(user, RoleAdmin) {
			t.Error("Expected user to have admin role")
		}
		
		if !am.HasRole(user, RoleModerator) {
			t.Error("Expected user to have moderator role")
		}
		
		if am.HasRole(user, RoleUser) {
			t.Error("Expected user to not have user role anymore")
		}
		
		if !am.HasPermission(user, PermissionAdmin) {
			t.Error("Expected user to have admin permission")
		}
	})
	
	t.Run("Update non-existent user", func(t *testing.T) {
		err := am.UpdateUserRoles("non-existent-user", []Role{RoleAdmin})
		
		if err != ErrUserNotFound {
			t.Errorf("Expected ErrUserNotFound, got %v", err)
		}
	})
}

func TestContextWithUser(t *testing.T) {
	am := NewAuthManager("test-secret", 1*time.Hour)
	
	user, _ := am.CreateUser("testuser", "test@example.com", "password123", []Role{RoleUser})
	
	ctx := context.Background()
	ctxWithUser := ContextWithUser(ctx, user)
	
	retrievedUser, ok := UserFromContext(ctxWithUser)
	
	if !ok {
		t.Error("Expected to retrieve user from context")
	}
	
	if retrievedUser.ID != user.ID {
		t.Errorf("Expected user ID %s, got %s", user.ID, retrievedUser.ID)
	}
	
	_, ok = UserFromContext(ctx)
	if ok {
		t.Error("Expected no user in empty context")
	}
}

func TestConstantTimeCompare(t *testing.T) {
	tests := []struct {
		name     string
		a        string
		b        string
		expected bool
	}{
		{"Equal strings", "password123", "password123", true},
		{"Different strings", "password123", "password456", false},
		{"Empty strings", "", "", true},
		{"One empty string", "password", "", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConstantTimeCompare(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestSessionCleanup(t *testing.T) {
	am := NewAuthManager("test-secret", 50*time.Millisecond)
	
	user, _ := am.CreateUser("testuser", "test@example.com", "password123", []Role{RoleUser})
	session, _ := am.CreateSession(user, "127.0.0.1", "TestAgent")
	
	time.Sleep(100 * time.Millisecond)
	
	_, _, err := am.ValidateSession(session.ID)
	if err != ErrTokenExpired {
		t.Errorf("Expected session to be expired, got error: %v", err)
	}
}