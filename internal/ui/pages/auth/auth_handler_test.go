package auth

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/internal/testutil/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAuthHandler_Login(t *testing.T) {
	t.Parallel()

	h := NewAuthHandler(nil, false)
	req := httptest.NewRequest(http.MethodGet, "/web/login", nil)
	w := httptest.NewRecorder()

	h.Login(w, req)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Contains(t, res.Header.Get("Content-Type"), "text/html")
}

func TestAuthHandler_RequestOTP(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		mockAuthSvc := new(mocks.AuthService)
		h := NewAuthHandler(mockAuthSvc, false)

		form := url.Values{}
		form.Add("email", "test@example.com")
		req := httptest.NewRequest(http.MethodPost, "/web/auth/otp/request", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		mockAuthSvc.On("RequestOTP", mock.Anything, "test@example.com").Return(nil)

		h.RequestOTP(w, req)

		res := w.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Contains(t, w.Body.String(), `name="email" value="test@example.com"`)
		mockAuthSvc.AssertExpectations(t)
	})

	t.Run("invalid_email", func(t *testing.T) {
		t.Parallel()
		mockAuthSvc := new(mocks.AuthService)
		h := NewAuthHandler(mockAuthSvc, false)

		form := url.Values{}
		form.Add("email", "invalid")
		req := httptest.NewRequest(http.MethodPost, "/web/auth/otp/request", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		mockAuthSvc.On("RequestOTP", mock.Anything, "invalid").Return(errors.New("invalid email"))

		h.RequestOTP(w, req)

		res := w.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Contains(t, w.Body.String(), `name="email" value="invalid"`)
		mockAuthSvc.AssertExpectations(t)
	})

	t.Run("service_error", func(t *testing.T) {
		t.Parallel()
		mockAuthSvc := new(mocks.AuthService)
		h := NewAuthHandler(mockAuthSvc, false)

		form := url.Values{}
		form.Add("email", "error@example.com")
		req := httptest.NewRequest(http.MethodPost, "/web/auth/otp/request", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		mockAuthSvc.On("RequestOTP", mock.Anything, "error@example.com").Return(errors.New("service failure"))

		h.RequestOTP(w, req)

		res := w.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Contains(t, w.Body.String(), `name="email" value="error@example.com"`)
		mockAuthSvc.AssertExpectations(t)
	})
}

func TestAuthHandler_VerifyOTP(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		mockAuthSvc := new(mocks.AuthService)
		h := NewAuthHandler(mockAuthSvc, false)

		form := url.Values{}
		form.Add("email", "test@example.com")
		form.Add("code", "123456")
		req := httptest.NewRequest(http.MethodPost, "/web/auth/otp/verify", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		mockAuthSvc.On("VerifyOTP", mock.Anything, "test@example.com", "123456").
			Return(&domain.TokenPair{AccessToken: "fake_token"}, nil)

		// Set HX-Request header to get HX-Redirect
		req.Header.Set("HX-Request", "true")

		h.VerifyOTP(w, req)

		res := w.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Equal(t, "/dashboard", res.Header.Get("HX-Redirect"))

		cookies := res.Cookies()
		require.NotEmpty(t, cookies)
		found := false
		for _, c := range cookies {
			if c.Name == "moolah_token" {
				assert.Equal(t, "fake_token", c.Value)
				assert.True(t, c.HttpOnly)
				found = true
			}
		}
		assert.True(t, found, "moolah_token cookie not found")
		mockAuthSvc.AssertExpectations(t)
	})

	t.Run("invalid_code", func(t *testing.T) {
		t.Parallel()
		mockAuthSvc := new(mocks.AuthService)
		h := NewAuthHandler(mockAuthSvc, false)

		form := url.Values{}
		form.Add("email", "test@example.com")
		form.Add("code", "000000")
		req := httptest.NewRequest(http.MethodPost, "/web/auth/otp/verify", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		mockAuthSvc.On("VerifyOTP", mock.Anything, "test@example.com", "000000").
			Return(nil, domain.ErrInvalidOTP)

		h.VerifyOTP(w, req)

		res := w.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Contains(t, w.Body.String(), "Invalid or expired code")
		mockAuthSvc.AssertExpectations(t)
	})
}

func TestAuthHandler_Logout(t *testing.T) {
	t.Parallel()

	h := NewAuthHandler(nil, false)
	req := httptest.NewRequest(http.MethodPost, "/web/auth/logout", nil)
	req.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()

	h.Logout(w, req)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, "/web/login", res.Header.Get("HX-Redirect"))

	cookies := res.Cookies()
	require.NotEmpty(t, cookies)
	found := false
	for _, c := range cookies {
		if c.Name == "moolah_token" {
			assert.Equal(t, -1, c.MaxAge)
			found = true
		}
	}
	assert.True(t, found, "moolah_token cookie deletion not found")
}

func TestMaskEmail(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		email    string
		expected string
	}{
		{
			name:     "standard email",
			email:    "john.doe@example.com",
			expected: "jo*****@example.com",
		},
		{
			name:     "short username",
			email:    "a@domain.com",
			expected: "a*****@domain.com",
		},
		{
			name:     "two char username",
			email:    "ab@domain.com",
			expected: "ab*****@domain.com",
		},
		{
			name:     "invalid email format",
			email:    "invalid-email",
			expected: "****",
		},
		{
			name:     "empty string",
			email:    "",
			expected: "****",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, maskEmail(tt.email))
		})
	}
}
