package middleware

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockIdempotencyStore is a mock implementation of IdempotencyStore.
type MockIdempotencyStore struct {
	mock.Mock
}

func (m *MockIdempotencyStore) Get(ctx context.Context, key string) (*CachedResponse, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		err := args.Error(1)
		if err != nil {
			return nil, fmt.Errorf("mock get: %w", err)
		}
		return nil, nil
	}
	return args.Get(0).(*CachedResponse), nil
}

func (m *MockIdempotencyStore) SetLocked(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	args := m.Called(ctx, key, ttl)
	err := args.Error(1)
	if err != nil {
		return false, fmt.Errorf("mock setlocked: %w", err)
	}
	return args.Bool(0), nil
}

func (m *MockIdempotencyStore) SetResponse(ctx context.Context, key string, resp CachedResponse, ttl time.Duration) error {
	args := m.Called(ctx, key, resp, ttl)
	err := args.Error(0)
	if err != nil {
		return fmt.Errorf("mock setresponse: %w", err)
	}
	return nil
}

func TestIdempotencyMiddleware(t *testing.T) {
	t.Parallel()

	userID := "user_123"
	key := "test-key"
	redisKey := "idempotency:user_123:test-key"

	t.Run("missing header returns 400", func(t *testing.T) {
		t.Parallel()
		store := new(MockIdempotencyStore)
		mw := Idempotency(store)
		handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

		req := httptest.NewRequest(http.MethodPost, "/", nil)
		ctx := context.WithValue(req.Context(), UserIDKey, userID)
		req = req.WithContext(ctx)

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "missing_idempotency_key")
	})

	t.Run("key too long returns 400", func(t *testing.T) {
		t.Parallel()
		store := new(MockIdempotencyStore)
		mw := Idempotency(store)
		handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.Header.Set(idempotencyHeader, string(make([]byte, 256)))
		ctx := context.WithValue(req.Context(), UserIDKey, userID)
		req = req.WithContext(ctx)

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "invalid_idempotency_key")
	})

	t.Run("first request (cache miss) executes handler and caches response", func(t *testing.T) {
		t.Parallel()
		store := new(MockIdempotencyStore)
		mw := Idempotency(store)
		handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":"tx_1"}`))
		}))

		store.On("Get", mock.Anything, redisKey).Return(nil, nil)
		store.On("SetLocked", mock.Anything, redisKey, idempotencyTTL).Return(true, nil)
		store.On("SetResponse", mock.Anything, redisKey, mock.MatchedBy(func(resp CachedResponse) bool {
			return resp.StatusCode == http.StatusCreated && string(resp.Body) == `{"id":"tx_1"}`
		}), idempotencyTTL).Return(nil)

		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.Header.Set(idempotencyHeader, key)
		ctx := context.WithValue(req.Context(), UserIDKey, userID)
		req = req.WithContext(ctx)

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.JSONEq(t, `{"id":"tx_1"}`, rec.Body.String())
		store.AssertExpectations(t)
	})

	t.Run("duplicate request (cache hit) returns cached response", func(t *testing.T) {
		t.Parallel()
		store := new(MockIdempotencyStore)
		mw := Idempotency(store)
		handlerCalled := false
		handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
		}))

		cached := &CachedResponse{
			StatusCode: http.StatusCreated,
			Body:       []byte(`{"id":"tx_1"}`),
		}
		store.On("Get", mock.Anything, redisKey).Return(cached, nil)

		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.Header.Set(idempotencyHeader, key)
		ctx := context.WithValue(req.Context(), UserIDKey, userID)
		req = req.WithContext(ctx)

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.JSONEq(t, `{"id":"tx_1"}`, rec.Body.String())
		assert.False(t, handlerCalled)
		store.AssertExpectations(t)
	})

	t.Run("in-flight request (locked) returns 409", func(t *testing.T) {
		t.Parallel()
		store := new(MockIdempotencyStore)
		mw := Idempotency(store)
		handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

		store.On("Get", mock.Anything, redisKey).Return(nil, nil)
		store.On("SetLocked", mock.Anything, redisKey, idempotencyTTL).Return(false, nil)

		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.Header.Set(idempotencyHeader, key)
		ctx := context.WithValue(req.Context(), UserIDKey, userID)
		req = req.WithContext(ctx)

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusConflict, rec.Code)
		assert.Contains(t, rec.Body.String(), "idempotency_key_in_flight")
		store.AssertExpectations(t)
	})

	t.Run("5xx response is not cached", func(t *testing.T) {
		t.Parallel()
		store := new(MockIdempotencyStore)
		mw := Idempotency(store)
		handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))

		store.On("Get", mock.Anything, redisKey).Return(nil, nil)
		store.On("SetLocked", mock.Anything, redisKey, idempotencyTTL).Return(true, nil)

		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.Header.Set(idempotencyHeader, key)
		ctx := context.WithValue(req.Context(), UserIDKey, userID)
		req = req.WithContext(ctx)

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		store.AssertExpectations(t)
	})
}
