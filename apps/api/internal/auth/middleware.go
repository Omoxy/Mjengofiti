package auth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"time"
)

type contextKey string

const userCtxKey contextKey = "auth_user"

// MiddlewareConfig defines the trust settings for signed bearer tokens.
type MiddlewareConfig struct {
	Environment string
	Enabled     bool
	Secret      string
	Issuer      string
	Audience    string
}

// WithUser injects an authenticated User into the request context.
func WithUser(ctx context.Context, user *User) context.Context {
	return context.WithValue(ctx, userCtxKey, user)
}

// UserFromContext extracts the authenticated User from the request context.
func UserFromContext(ctx context.Context) *User {
	user, _ := ctx.Value(userCtxKey).(*User)
	return user
}

// Middleware accepts only configured HS256 JWTs. With authentication disabled,
// a local administrator identity is available only to loopback development clients.
func Middleware(cfg MiddlewareConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var user *User
			authorization := strings.TrimSpace(r.Header.Get("Authorization"))
			if authorization != "" {
				if cfg.Enabled {
					user = authenticateBearer(authorization, cfg)
				}
			} else if !cfg.Enabled && cfg.Environment == "development" && isLoopbackRemote(r.RemoteAddr) {
				user = &User{
					ID:          "local-development-user",
					DisplayName: "Local Development User",
					Role:        RoleAdmin,
					Permissions: DefaultPermissionsForRole(RoleAdmin),
				}
			}

			if user != nil {
				r = r.WithContext(WithUser(r.Context(), user))
			}
			next.ServeHTTP(w, r)
		})
	}
}

type tokenHeader struct {
	Algorithm string `json:"alg"`
	Type      string `json:"typ"`
}

type tokenClaims struct {
	Subject   string          `json:"sub"`
	Email     string          `json:"email"`
	Name      string          `json:"name"`
	Role      string          `json:"role"`
	Issuer    string          `json:"iss"`
	Audience  json.RawMessage `json:"aud"`
	ExpiresAt int64           `json:"exp"`
	NotBefore int64           `json:"nbf,omitempty"`
	IssuedAt  int64           `json:"iat,omitempty"`
}

func authenticateBearer(header string, cfg MiddlewareConfig) *User {
	parts := strings.Fields(header)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || len(cfg.Secret) < 32 || cfg.Issuer == "" || cfg.Audience == "" {
		return nil
	}

	segments := strings.Split(parts[1], ".")
	if len(segments) != 3 {
		return nil
	}

	headerJSON, err := base64.RawURLEncoding.DecodeString(segments[0])
	if err != nil {
		return nil
	}
	var tokenHeaderValue tokenHeader
	if json.Unmarshal(headerJSON, &tokenHeaderValue) != nil || tokenHeaderValue.Algorithm != "HS256" {
		return nil
	}

	providedSignature, err := base64.RawURLEncoding.DecodeString(segments[2])
	if err != nil {
		return nil
	}
	mac := hmac.New(sha256.New, []byte(cfg.Secret))
	_, _ = mac.Write([]byte(segments[0] + "." + segments[1]))
	if !hmac.Equal(providedSignature, mac.Sum(nil)) {
		return nil
	}

	claimsJSON, err := base64.RawURLEncoding.DecodeString(segments[1])
	if err != nil {
		return nil
	}
	var claims tokenClaims
	if json.Unmarshal(claimsJSON, &claims) != nil {
		return nil
	}

	now := time.Now().Unix()
	role, knownRole := ParseRole(claims.Role)
	if !knownRole || claims.Subject == "" || claims.Issuer != cfg.Issuer || !hasAudience(claims.Audience, cfg.Audience) || claims.ExpiresAt <= now || (claims.NotBefore != 0 && claims.NotBefore > now) || (claims.IssuedAt != 0 && claims.IssuedAt > now+30) {
		return nil
	}

	name := claims.Name
	if name == "" {
		name = claims.Subject
	}

	return &User{
		ID:          claims.Subject,
		Email:       claims.Email,
		DisplayName: name,
		Role:        role,
		Permissions: DefaultPermissionsForRole(role),
	}
}

func hasAudience(raw json.RawMessage, expected string) bool {
	var audience string
	if json.Unmarshal(raw, &audience) == nil {
		return audience == expected
	}
	var audiences []string
	if json.Unmarshal(raw, &audiences) != nil {
		return false
	}
	for _, audience := range audiences {
		if audience == expected {
			return true
		}
	}
	return false
}

func isLoopbackRemote(remoteAddr string) bool {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return false
	}
	ip := net.ParseIP(strings.Trim(host, "[]"))
	return ip != nil && ip.IsLoopback()
}

// RequireAuthentication rejects requests without an authenticated user.
func RequireAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if UserFromContext(r.Context()) == nil {
			writeAuthError(w, http.StatusUnauthorized, "unauthorized", "Authentication is required to access this resource")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequirePermission verifies the authenticated user possesses the requested permission.
func RequirePermission(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := UserFromContext(r.Context())
			if user == nil {
				writeAuthError(w, http.StatusUnauthorized, "unauthorized", "Authentication is required to access this resource")
				return
			}
			if !user.HasPermission(permission) {
				writeAuthError(w, http.StatusForbidden, "forbidden", "Insufficient permissions to perform this operation")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireRole verifies the authenticated user holds one of the specified roles.
func RequireRole(roles ...Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := UserFromContext(r.Context())
			if user == nil {
				writeAuthError(w, http.StatusUnauthorized, "unauthorized", "Authentication is required to access this resource")
				return
			}
			for _, allowed := range roles {
				if user.Role == allowed || user.Role == RoleAdmin {
					next.ServeHTTP(w, r)
					return
				}
			}
			writeAuthError(w, http.StatusForbidden, "forbidden", "User role is not authorized for this operation")
		})
	}
}

func writeAuthError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error":   code,
		"message": message,
	})
}
