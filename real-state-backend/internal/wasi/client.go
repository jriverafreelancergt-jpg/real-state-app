package wasi

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"real-state-backend/config"
	"sync"
	"time"
)

// HTTPError representa un error retornado por la API HTTP de Wasi.
type HTTPError struct {
	StatusCode int
	Message    string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("Wasi API HTTP error: status=%d, message=%s", e.StatusCode, e.Message)
}

// isTransientError determina si el error es de carácter transitorio y justifica reintento.
func isTransientError(err error) bool {
	if err == nil {
		return false
	}
	if httpErr, ok := err.(*HTTPError); ok {
		// Reintentar en límite de tasa (429) o errores del servidor (5xx)
		return httpErr.StatusCode == 429 || httpErr.StatusCode >= 500
	}
	// Otros errores (timeouts de red, fallos DNS, conexión rechazada) son transitorios
	return true
}

// CircuitBreakerState representa los estados posibles del Circuit Breaker.
type CircuitBreakerState int

const (
	StateClosed CircuitBreakerState = iota
	StateOpen
	StateHalfOpen
)

// CircuitBreaker protege al sistema de propagar fallos en cascada si Wasi se cae.
type CircuitBreaker struct {
	mu                  sync.RWMutex
	state               CircuitBreakerState
	failureThreshold    int
	consecutiveFailures int
	cooldownPeriod      time.Duration
	lastStateChange     time.Time
}

// NewCircuitBreaker inicializa un nuevo Circuit Breaker.
func NewCircuitBreaker(threshold int, cooldown time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:           StateClosed,
		failureThreshold: threshold,
		cooldownPeriod:  cooldown,
		lastStateChange: time.Now(),
	}
}

// Execute ejecuta la operación encapsulada en el Circuit Breaker.
func (cb *CircuitBreaker) Execute(operation func() error) error {
	if !cb.allowRequest() {
		return fmt.Errorf("circuit breaker is open (fast-fail): calls to Wasi API blocked temporarily")
	}

	err := operation()
	cb.recordResult(err)
	return err
}

func (cb *CircuitBreaker) allowRequest() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.state == StateOpen {
		if time.Since(cb.lastStateChange) > cb.cooldownPeriod {
			cb.state = StateHalfOpen
			cb.lastStateChange = time.Now()
			return true
		}
		return false
	}
	return true
}

func (cb *CircuitBreaker) recordResult(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	isFailure := err != nil && isTransientError(err)

	if isFailure {
		cb.consecutiveFailures++
		if cb.state == StateHalfOpen || cb.consecutiveFailures >= cb.failureThreshold {
			cb.state = StateOpen
			cb.lastStateChange = time.Now()
		}
	} else {
		// Si es exitoso o es un error no transitorio (ej: 400 Bad Request que no cambiará reintentando)
		if cb.state == StateHalfOpen {
			cb.state = StateClosed
			cb.consecutiveFailures = 0
		} else if cb.state == StateClosed {
			cb.consecutiveFailures = 0
		}
	}
}

// GetState retorna el estado actual del Circuit Breaker (útil para testing y monitoreo).
func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// wasiRoundTripper intercepta las peticiones y añade las credenciales a la URL de forma segura.
type wasiRoundTripper struct {
	next      http.RoundTripper
	companyID string
	token     string
}

func (rt *wasiRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// Clonar la URL para evitar mutar el request original
	newURL := *req.URL
	q := newURL.Query()
	q.Set("id_company", rt.companyID)
	q.Set("wasi_token", rt.token)
	newURL.RawQuery = q.Encode()

	newReq := req.Clone(req.Context())
	newReq.URL = &newURL

	return rt.next.RoundTrip(newReq)
}

// Client define el cliente HTTP resiliente para la API de Wasi.
type Client struct {
	baseURL    string
	httpClient *http.Client
	breaker    *CircuitBreaker
}

// NewClient crea un cliente configurado con credenciales, timeout por defecto y Circuit Breaker.
func NewClient(baseURL string, companyID string, token config.SecretToken, timeout time.Duration) *Client {
	// RoundTripper customizado para inyectar credenciales automáticamente
	transport := &wasiRoundTripper{
		next:      http.DefaultTransport,
		companyID: companyID,
		token:     string(token),
	}

	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   timeout,
		},
		// Circuit Breaker por defecto: se abre tras 5 fallos consecutivos, tiempo de enfriamiento de 30 segundos
		breaker: NewCircuitBreaker(5, 30*time.Second),
	}
}

// Do ejecuta una petición HTTP con Exponential Backoff y Circuit Breaker.
func (c *Client) Do(ctx context.Context, method, path string, body io.Reader) ([]byte, error) {
	var responseBody []byte

	// Definir la operación a ejecutar dentro del Circuit Breaker
	operation := func() error {
		reqURL, err := url.JoinPath(c.baseURL, path)
		if err != nil {
			return err
		}

		// Implementación de Exponential Backoff nativa en Go
		minBackoff := 200 * time.Millisecond
		maxBackoff := 5 * time.Second
		factor := 2.0
		currentBackoff := minBackoff

		for {
			req, err := http.NewRequestWithContext(ctx, method, reqURL, body)
			if err != nil {
				return err
			}

			resp, err := c.httpClient.Do(req)
			if err != nil {
				// Error de red (transitorio)
				if ctx.Err() != nil {
					return ctx.Err()
				}
				if !isTransientError(err) {
					return err
				}
				goto waitAndRetry
			}

			// Validar estatus HTTP
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				defer resp.Body.Close()
				data, err := io.ReadAll(resp.Body)
				if err != nil {
					return err
				}
				responseBody = data
				return nil
			}

			// Manejar error HTTP
			resp.Body.Close()
			err = &HTTPError{
				StatusCode: resp.StatusCode,
				Message:    resp.Status,
			}

			if !isTransientError(err) {
				return err // Error no transitorio (ej: 400 Bad Request, 401 Unauthorized) -> fallar rápido
			}

		waitAndRetry:
			// Aplicar tiempo de espera del Backoff antes de reintentar
			timer := time.NewTimer(currentBackoff)
			select {
			case <-ctx.Done():
				timer.Stop()
				return ctx.Err()
			case <-timer.C:
			}

			currentBackoff = time.Duration(float64(currentBackoff) * factor)
			if currentBackoff > maxBackoff {
				currentBackoff = maxBackoff
			}
		}
	}

	// Ejecutar a través del Circuit Breaker
	if err := c.breaker.Execute(operation); err != nil {
		return nil, err
	}

	return responseBody, nil
}
