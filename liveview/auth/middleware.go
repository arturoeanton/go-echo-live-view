package auth

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type AuthConfig struct {
	AuthManager     *AuthManager
	RequireAuth     bool
	RequiredRoles   []Role
	RequiredPerms   []Permission
	TokenLookup     string
	SessionCookieName string
	SkipPaths       []string
}

func DefaultAuthConfig(am *AuthManager) *AuthConfig {
	return &AuthConfig{
		AuthManager:       am,
		RequireAuth:       true,
		TokenLookup:       "header:Authorization,cookie:token",
		SessionCookieName: "session_id",
		SkipPaths:         []string{"/login", "/register", "/public"},
	}
}

func AuthMiddleware(config *AuthConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			path := c.Request().URL.Path
			
			for _, skipPath := range config.SkipPaths {
				if strings.HasPrefix(path, skipPath) {
					return next(c)
				}
			}

			var user *User
			var err error

			sessionID := c.Request().Header.Get("X-Session-ID")
			if sessionID == "" {
				cookie, _ := c.Cookie(config.SessionCookieName)
				if cookie != nil {
					sessionID = cookie.Value
				}
			}

			if sessionID != "" {
				_, user, err = config.AuthManager.ValidateSession(sessionID)
				if err == nil {
					c.Set("session_id", sessionID)
					c.Set("user", user)
					return next(c)
				}
			}

			token := extractToken(c, config.TokenLookup)
			if token != "" {
				user, err = config.AuthManager.ValidateJWT(token)
				if err == nil {
					c.Set("user", user)
					return next(c)
				}
			}

			if config.RequireAuth {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "authentication required",
				})
			}

			return next(c)
		}
	}
}

func RequireRoles(roles ...Role) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			user, ok := c.Get("user").(*User)
			if !ok || user == nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "authentication required",
				})
			}

			hasRole := false
			for _, requiredRole := range roles {
				for _, userRole := range user.Roles {
					if userRole == requiredRole {
						hasRole = true
						break
					}
				}
				if hasRole {
					break
				}
			}

			if !hasRole {
				return c.JSON(http.StatusForbidden, map[string]string{
					"error": "insufficient privileges",
				})
			}

			return next(c)
		}
	}
}

func RequirePermissions(permissions ...Permission) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			user, ok := c.Get("user").(*User)
			if !ok || user == nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "authentication required",
				})
			}

			for _, perm := range permissions {
				if !user.Permissions[perm] {
					return c.JSON(http.StatusForbidden, map[string]string{
						"error": "insufficient permissions",
					})
				}
			}

			return next(c)
		}
	}
}

func CORSConfig() middleware.CORSConfig {
	return middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{
			echo.GET,
			echo.HEAD,
			echo.PUT,
			echo.PATCH,
			echo.POST,
			echo.DELETE,
			echo.OPTIONS,
		},
		AllowHeaders: []string{
			echo.HeaderOrigin,
			echo.HeaderContentType,
			echo.HeaderAccept,
			echo.HeaderAuthorization,
			"X-Session-ID",
			"X-CSRF-Token",
		},
		ExposeHeaders: []string{
			"X-Session-ID",
			"X-Request-ID",
		},
		AllowCredentials: true,
		MaxAge:          3600,
	}
}

func extractToken(c echo.Context, lookup string) string {
	parts := strings.Split(lookup, ",")
	
	for _, part := range parts {
		source := strings.Split(part, ":")
		if len(source) != 2 {
			continue
		}

		switch source[0] {
		case "header":
			token := c.Request().Header.Get(source[1])
			if token != "" {
				if strings.HasPrefix(token, "Bearer ") {
					return strings.TrimPrefix(token, "Bearer ")
				}
				return token
			}
		case "cookie":
			cookie, err := c.Cookie(source[1])
			if err == nil && cookie.Value != "" {
				return cookie.Value
			}
		case "query":
			token := c.QueryParam(source[1])
			if token != "" {
				return token
			}
		}
	}

	return ""
}

func CSRFConfig() middleware.CSRFConfig {
	return middleware.CSRFConfig{
		TokenLookup:    "header:X-CSRF-Token",
		CookieName:     "csrf_token",
		CookiePath:     "/",
		CookieMaxAge:   86400,
		CookieSecure:   true,
		CookieHTTPOnly: true,
		CookieSameSite: http.SameSiteStrictMode,
	}
}

func SessionMiddleware(am *AuthManager) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			sessionID := c.Request().Header.Get("X-Session-ID")
			if sessionID == "" {
				cookie, err := c.Cookie("session_id")
				if err == nil {
					sessionID = cookie.Value
				}
			}

			if sessionID != "" {
				session, user, err := am.ValidateSession(sessionID)
				if err == nil {
					c.Set("session", session)
					c.Set("user", user)
				}
			}

			return next(c)
		}
	}
}

func RateLimitConfig() middleware.RateLimiterConfig {
	store := middleware.NewRateLimiterMemoryStore(100)
	return middleware.RateLimiterConfig{
		Store: store,
		IdentifierExtractor: func(c echo.Context) (string, error) {
			user, ok := c.Get("user").(*User)
			if ok && user != nil {
				return user.ID, nil
			}
			return c.RealIP(), nil
		},
		ErrorHandler: func(c echo.Context, err error) error {
			return c.JSON(http.StatusTooManyRequests, map[string]string{
				"error": "rate limit exceeded",
			})
		},
		DenyHandler: func(c echo.Context, identifier string, err error) error {
			return c.JSON(http.StatusTooManyRequests, map[string]string{
				"error": "too many requests",
			})
		},
	}
}