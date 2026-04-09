package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"gophermart/internal/app/services/auth"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestRequestLoggerMiddleware(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	})

	middleware := RequestLoggerMiddleware(logger)
	wrappedHandler := middleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %v", w.Code)
	}

	logOutput := buf.String()
	if logOutput == "" {
		t.Error("Expected log output, got empty string")
	}

	expectedFields := []string{"Request starting", "Request finished", "method", "uri", "status", "duration"}
	for _, field := range expectedFields {
		if !bytes.Contains([]byte(logOutput), []byte(field)) {
			t.Errorf("Log output doesn't contain expected field: %s", field)
		}
	}
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	userID := uuid.New()
	token, err := auth.GenerateToken(userID)
	require.NoError(t, err)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctxUserID := r.Context().Value(UserIDKey)
		if ctxUserID == nil {
			t.Error("Expected userID in context, got nil")
			return
		}
		if ctxUserID.(uuid.UUID) != userID {
			t.Errorf("Expected userID %v, got %v", userID, ctxUserID)
		}
		w.WriteHeader(http.StatusOK)
	})

	middleware := AuthMiddleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: token})
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %v", w.Code)
	}
}

func TestAuthMiddleware_MissingToken(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called when token is missing")
		w.WriteHeader(http.StatusOK)
	})

	middleware := AuthMiddleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status Unauthorized, got %v", w.Code)
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called when token is invalid")
		w.WriteHeader(http.StatusOK)
	})

	middleware := AuthMiddleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: "invalid.token.value"})
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status Unauthorized, got %v", w.Code)
	}
}

func TestLoggingResponseWriter_Write(t *testing.T) {
	responseData := &responseData{}
	w := httptest.NewRecorder()
	lw := &loggingResponseWriter{
		ResponseWriter: w,
		responseData:   responseData,
	}

	data := []byte("test data")
	n, err := lw.Write(data)
	require.NoError(t, err)

	if n != len(data) {
		t.Errorf("Write() returned %v bytes, want %v", n, len(data))
	}

	if responseData.size != len(data) {
		t.Errorf("responseData.size = %v, want %v", responseData.size, len(data))
	}
}

func TestLoggingResponseWriter_WriteHeader(t *testing.T) {
	responseData := &responseData{}
	w := httptest.NewRecorder()
	lw := &loggingResponseWriter{
		ResponseWriter: w,
		responseData:   responseData,
	}

	lw.WriteHeader(http.StatusCreated)

	if responseData.status != http.StatusCreated {
		t.Errorf("responseData.status = %v, want %v", responseData.status, http.StatusCreated)
	}
}

func TestGzipMiddleware_NoAcceptEncoding(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	})

	middleware := GzipMiddleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %v", w.Code)
	}

	if w.Header().Get("Content-Encoding") == "gzip" {
		t.Error("Expected no gzip encoding when Accept-Encoding not set")
	}
}

func TestGzipMiddleware_WithAcceptEncoding(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response that should be compressed"))
	})

	middleware := GzipMiddleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %v", w.Code)
	}

	if w.Header().Get("Content-Encoding") != "gzip" {
		t.Error("Expected gzip encoding when Accept-Encoding is gzip")
	}
}

func TestGzipWriter_Write(t *testing.T) {
	w := httptest.NewRecorder()
	var buf bytes.Buffer
	gw := gzipWriter{
		ResponseWriter: w,
		Writer:         &buf,
	}

	testData := []byte("test data")
	n, err := gw.Write(testData)
	require.NoError(t, err)

	if n != len(testData) {
		t.Errorf("Write() returned %v bytes, want %v", n, len(testData))
	}

	if buf.String() != string(testData) {
		t.Errorf("Buffer contains %q, want %q", buf.String(), string(testData))
	}
}

func TestGzipMiddleware_CompressedRequest(t *testing.T) {
	var receivedBody string
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		receivedBody = string(body)
		w.WriteHeader(http.StatusOK)
	})

	middleware := GzipMiddleware(handler)

	originalBody := "test request body"
	var compressedBuf bytes.Buffer
	gzWriter := gzip.NewWriter(&compressedBuf)
	gzWriter.Write([]byte(originalBody))
	gzWriter.Close()

	req := httptest.NewRequest(http.MethodPost, "/test", &compressedBuf)
	req.Header.Set("Content-Encoding", "gzip")
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %v", w.Code)
	}

	if receivedBody != originalBody {
		t.Errorf("Expected decompressed body %q, got %q", originalBody, receivedBody)
	}
}

func TestGzipMiddleware_InvalidGzipRequest(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called for invalid gzip")
	})

	middleware := GzipMiddleware(handler)

	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader([]byte("not gzip data")))
	req.Header.Set("Content-Encoding", "gzip")
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %v", w.Code)
	}
}

func TestGzipMiddleware_GzipWriterError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	})

	middleware := GzipMiddleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %v", w.Code)
	}
}
