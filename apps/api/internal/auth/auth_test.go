package auth_test

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/mjengofiti/api/internal/auth"
)

const testSecret = "0123456789abcdef0123456789abcdef"

func TestAuthMeRequiresAuthenticatedUser(t *testing.T) {
	handler := auth.NewHandler()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	response := httptest.NewRecorder()

	handler.HandleMe(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for anonymous profile request, got %d", response.Code)
	}
	if strings.Contains(response.Body.String(), "nairobi.go.ke") || strings.Contains(response.Body.String(), "Joseph Kamau") {
		t.Fatal("anonymous profile response exposed a fabricated personal identity")
	}
}

func TestAuthMeReturnsContextUser(t *testing.T) {
	handler := auth.NewHandler()
	user := &auth.User{ID: "subject-123", Role: auth.RoleViewer, Permissions: auth.DefaultPermissionsForRole(auth.RoleViewer)}
	request := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	request = request.WithContext(auth.WithUser(request.Context(), user))
	response := httptest.NewRecorder()

	handler.HandleMe(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200 for authenticated profile request, got %d", response.Code)
	}
	var actual auth.User
	if err := json.NewDecoder(response.Body).Decode(&actual); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	if actual.ID != user.ID || actual.Role != user.Role {
		t.Fatalf("unexpected current user: %#v", actual)
	}
}

func TestAuthRoles(t *testing.T) {
	handler := auth.NewHandler()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/auth/roles", nil)
	response := httptest.NewRecorder()

	handler.HandleRoles(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
	var result struct {
		Roles []auth.RoleDefinition `json:"roles"`
		Total int                   `json:"total"`
	}
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		t.Fatalf("decode roles: %v", err)
	}
	if result.Total != 6 {
		t.Fatalf("expected 6 role definitions, got %d", result.Total)
	}
}

func TestPermissionEnforcement(t *testing.T) {
	protected := auth.RequirePermission("datasets:promote")(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for _, testCase := range []struct {
		name       string
		user       *auth.User
		wantStatus int
	}{
		{name: "anonymous", wantStatus: http.StatusUnauthorized},
		{name: "viewer", user: userWithRole(auth.RoleViewer), wantStatus: http.StatusForbidden},
		{name: "steward", user: userWithRole(auth.RoleDataSteward), wantStatus: http.StatusOK},
		{name: "admin", user: userWithRole(auth.RoleAdmin), wantStatus: http.StatusOK},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/api/v1/ingestion/promote", nil)
			if testCase.user != nil {
				request = request.WithContext(auth.WithUser(request.Context(), testCase.user))
			}
			response := httptest.NewRecorder()
			protected.ServeHTTP(response, request)
			if response.Code != testCase.wantStatus {
				t.Fatalf("expected %d, got %d", testCase.wantStatus, response.Code)
			}
		})
	}
}

func TestMiddlewareRejectsCallerControlledIdentities(t *testing.T) {
	middleware := auth.Middleware(auth.MiddlewareConfig{
		Environment: "production",
		Enabled:     true,
		Secret:      testSecret,
		Issuer:      "https://identity.example.test",
		Audience:    "mjengofiti-api",
	})
	protected := middleware(auth.RequireAuthentication(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))

	for _, testCase := range []struct {
		name   string
		header string
	}{
		{name: "bare role bearer", header: "Bearer admin"},
		{name: "role header", header: "X-User-Role: admin"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/protected", nil)
			if strings.HasPrefix(testCase.header, "X-User-Role:") {
				request.Header.Set("X-User-Role", "admin")
			} else {
				request.Header.Set("Authorization", testCase.header)
			}
			response := httptest.NewRecorder()
			protected.ServeHTTP(response, request)
			if response.Code != http.StatusUnauthorized {
				t.Fatalf("expected spoofed identity to receive 401, got %d", response.Code)
			}
		})
	}
}

func TestMiddlewareAcceptsValidJWTAndIgnoresClaimPermissions(t *testing.T) {
	middleware := auth.Middleware(auth.MiddlewareConfig{
		Environment: "production",
		Enabled:     true,
		Secret:      testSecret,
		Issuer:      "https://identity.example.test",
		Audience:    "mjengofiti-api",
	})
	protected := middleware(auth.RequirePermission("buildings:read")(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		user := auth.UserFromContext(request.Context())
		if user == nil || user.ID != "subject-123" || user.Role != auth.RoleViewer {
			t.Errorf("unexpected authenticated user: %#v", user)
		}
		w.WriteHeader(http.StatusOK)
	})))

	request := httptest.NewRequest(http.MethodGet, "/protected", nil)
	request.Header.Set("Authorization", "Bearer "+signedToken(t, map[string]any{
		"sub": "subject-123", "role": "viewer", "permissions": []string{"*"},
		"iss": "https://identity.example.test", "aud": []string{"other-api", "mjengofiti-api"},
		"exp": time.Now().Add(time.Hour).Unix(),
	}))
	response := httptest.NewRecorder()
	protected.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected valid viewer token to read buildings, got %d", response.Code)
	}
}

func TestMiddlewareRejectsInvalidJWTClaims(t *testing.T) {
	middleware := auth.Middleware(auth.MiddlewareConfig{
		Environment: "production",
		Enabled:     true,
		Secret:      testSecret,
		Issuer:      "https://identity.example.test",
		Audience:    "mjengofiti-api",
	})
	protected := middleware(auth.RequireAuthentication(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))
	now := time.Now().Unix()

	for _, testCase := range []struct {
		name   string
		claims map[string]any
	}{
		{name: "expired", claims: tokenClaims(now - 1, "viewer", "https://identity.example.test", "mjengofiti-api")},
		{name: "wrong issuer", claims: tokenClaims(now + 600, "viewer", "https://untrusted.example.test", "mjengofiti-api")},
		{name: "wrong audience", claims: tokenClaims(now + 600, "viewer", "https://identity.example.test", "other-api")},
		{name: "unknown role", claims: tokenClaims(now + 600, "root", "https://identity.example.test", "mjengofiti-api")},
		{name: "missing subject", claims: map[string]any{"role": "admin", "iss": "https://identity.example.test", "aud": "mjengofiti-api", "exp": now + 600}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/protected", nil)
			request.Header.Set("Authorization", "Bearer "+signedToken(t, testCase.claims))
			response := httptest.NewRecorder()
			protected.ServeHTTP(response, request)
			if response.Code != http.StatusUnauthorized {
				t.Fatalf("expected invalid token to receive 401, got %d", response.Code)
			}
		})
	}
}

func TestDevelopmentIdentityIsLoopbackOnly(t *testing.T) {
	middleware := auth.Middleware(auth.MiddlewareConfig{Environment: "development"})
	protected := middleware(auth.RequireAuthentication(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		user := auth.UserFromContext(request.Context())
		if user == nil || user.ID != "local-development-user" {
			t.Errorf("expected loopback development identity, got %#v", user)
		}
		w.WriteHeader(http.StatusOK)
	})))

	localRequest := httptest.NewRequest(http.MethodGet, "/protected", nil)
	localRequest.RemoteAddr = "127.0.0.1:1234"
	localResponse := httptest.NewRecorder()
	protected.ServeHTTP(localResponse, localRequest)
	if localResponse.Code != http.StatusOK {
		t.Fatalf("expected loopback development access, got %d", localResponse.Code)
	}

	remoteRequest := httptest.NewRequest(http.MethodGet, "/protected", nil)
	remoteRequest.RemoteAddr = "192.0.2.10:1234"
	remoteResponse := httptest.NewRecorder()
	protected.ServeHTTP(remoteResponse, remoteRequest)
	if remoteResponse.Code != http.StatusUnauthorized {
		t.Fatalf("expected non-loopback development access to fail, got %d", remoteResponse.Code)
	}
}

func userWithRole(role auth.Role) *auth.User {
	return &auth.User{ID: "test-user", Role: role, Permissions: auth.DefaultPermissionsForRole(role)}
}

func tokenClaims(expiresAt int64, role, issuer, audience string) map[string]any {
	return map[string]any{"sub": "subject-123", "role": role, "iss": issuer, "aud": audience, "exp": expiresAt}
}

func signedToken(t *testing.T, claims map[string]any) string {
	t.Helper()
	headerJSON, err := json.Marshal(map[string]string{"alg": "HS256", "typ": "JWT"})
	if err != nil {
		t.Fatalf("marshal JWT header: %v", err)
	}
	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		t.Fatalf("marshal JWT claims: %v", err)
	}
	unsigned := base64.RawURLEncoding.EncodeToString(headerJSON) + "." + base64.RawURLEncoding.EncodeToString(claimsJSON)
	mac := hmac.New(sha256.New, []byte(testSecret))
	_, _ = mac.Write([]byte(unsigned))
	return unsigned + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func TestWithUserContext(t *testing.T) {
	user := userWithRole(auth.RoleViewer)
	if actual := auth.UserFromContext(auth.WithUser(context.Background(), user)); actual != user {
		t.Fatal("context did not retain authenticated user")
	}
}
