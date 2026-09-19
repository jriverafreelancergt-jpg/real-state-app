package wasi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"real-state-backend/config"
	"sync/atomic"
	"testing"
	"time"
)

func TestSecretTokenString(t *testing.T) {
	token := config.SecretToken("my-super-secret-token")
	if token.String() != "***[REDACTED]***" {
		t.Errorf("Expected REDACTED string representation, got %s", token.String())
	}
}

func TestCircuitBreakerStates(t *testing.T) {
	// Breaker que se abre tras 2 fallos consecutivos y tiene 50ms de cooldown
	cb := NewCircuitBreaker(2, 50*time.Millisecond)

	if cb.GetState() != StateClosed {
		t.Errorf("Expected Closed state initially, got %v", cb.GetState())
	}

	// Primer fallo (transitorio)
	errTransient := &HTTPError{StatusCode: 503, Message: "Service Unavailable"}
	err := cb.Execute(func() error {
		return errTransient
	})
	if !errors.Is(err, errTransient) {
		t.Errorf("Expected original error, got %v", err)
	}
	if cb.GetState() != StateClosed {
		t.Errorf("Expected Closed state after 1 failure, got %v", cb.GetState())
	}

	// Segundo fallo (debe abrir el breaker)
	_ = cb.Execute(func() error {
		return errTransient
	})
	if cb.GetState() != StateOpen {
		t.Errorf("Expected Open state after 2 failures, got %v", cb.GetState())
	}

	// Petición bloqueada inmediatamente por breaker abierto
	err = cb.Execute(func() error {
		return nil
	})
	if err == nil || err.Error() != "circuit breaker is open (fast-fail): calls to Wasi API blocked temporarily" {
		t.Errorf("Expected fast-fail error, got %v", err)
	}

	// Esperar periodo de enfriamiento (cooldown)
	time.Sleep(60 * time.Millisecond)

	// Debe transicionar a HalfOpen y dejar pasar una petición
	called := false
	err = cb.Execute(func() error {
		called = true
		return nil // Éxito
	})
	if err != nil {
		t.Errorf("Expected success in half-open request, got %v", err)
	}
	if !called {
		t.Errorf("Expected operation to be called in half-open state")
	}

	// El éxito debe cerrar el breaker
	if cb.GetState() != StateClosed {
		t.Errorf("Expected Closed state after success in half-open, got %v", cb.GetState())
	}
}

func TestClientExponentialBackoffAndTimeout(t *testing.T) {
	var requestCount int32

	// Servidor Mock que retorna error 503 las primeras 2 veces y luego éxito
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&requestCount, 1)

		// Verificar que las credenciales de wasiRoundTripper se inyectan correctamente en la query
		if r.URL.Query().Get("id_company") != "123" {
			t.Errorf("Expected id_company = 123, got %s", r.URL.Query().Get("id_company"))
		}
		if r.URL.Query().Get("wasi_token") != "secret" {
			t.Errorf("Expected wasi_token = secret, got %s", r.URL.Query().Get("wasi_token"))
		}

		if count <= 2 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"success","total":1}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "123", config.SecretToken("secret"), 5*time.Second)

	ctx := context.Background()
	resp, err := client.Do(ctx, "GET", "/test", nil)
	if err != nil {
		t.Fatalf("Expected request to succeed eventually, got error: %v", err)
	}

	if atomic.LoadInt32(&requestCount) != 3 {
		t.Errorf("Expected 3 total request attempts due to backoff, got %d", requestCount)
	}

	if string(resp) != `{"status":"success","total":1}` {
		t.Errorf("Expected mock response body, got %s", string(resp))
	}
}

func TestClientRequestTimeoutCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond) // Simular delay
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL, "123", config.SecretToken("secret"), 5*time.Second)

	// Contexto que expira rápido para forzar cancelación
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	_, err := client.Do(ctx, "GET", "/test", nil)
	if err == nil {
		t.Fatal("Expected context timeout error, got nil")
	}

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Expected context.DeadlineExceeded, got %v", err)
	}
}
