package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/internal/testutil/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestIdempotencyMiddleware(t *testing.T) {
	t.Parallel()

	userID := "user_123"
	key := "test-key"
	redisKey := "idempotency:user_123:test-key"

	t.Run("missing header returns 400", func(t *testing.T) {
		t.Parallel()
		store := new(mocks.IdempotencyStore)
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
		store := new(mocks.IdempotencyStore)
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
		store := new(mocks.IdempotencyStore)
		mw := Idempotency(store)
		handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
			if _, err := w.Write([]byte(`{"id":"tx_1"}`)); err != nil {
				t.Errorf("failed to write response: %v", err)
			}
		}))

		store.On("Get", mock.Anything, redisKey).Return(nil, nil)
		store.On("SetLocked", mock.Anything, redisKey, idempotencyTTL).Return(true, nil)
		store.On("SetResponse", mock.Anything, redisKey, mock.MatchedBy(func(resp domain.CachedResponse) bool {
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
		store := new(mocks.IdempotencyStore)
		mw := Idempotency(store)
		handlerCalled := false
		handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
		}))

		cached := &domain.CachedResponse{
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
		store := new(mocks.IdempotencyStore)
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
		store := new(mocks.IdempotencyStore)
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
