package handlers

import (
	"encoding/json"
	"net/http"
	"regexp"

	"github.com/go-chi/chi/v5"
)

// Healthz: pure liveness, no dependencies
func Healthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Readyz: would check downstream deps in real life
func Readyz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

// Ethereum address validator — minimal sanity check
var ethAddressRegex = regexp.MustCompile(`^0x[0-9a-fA-F]{40}$`)

func GetBalance(w http.ResponseWriter, r *http.Request) {
	address := chi.URLParam(r, "address")
	if !ethAddressRegex.MatchString(address) {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid ethereum address",
		})
		return
	}

	// Mock balance — real impl would query chain RPC via signed read
	writeJSON(w, http.StatusOK, map[string]any{
		"address":  address,
		"balance":  "0",
		"currency": "ETH",
		"note":     "mock data",
	})
}

func SignTransaction(w http.ResponseWriter, r *http.Request) {
	// Sensitive endpoint placeholder.
	// Real implementation: never holds private keys — would delegate to HSM/KMS.
	// Week 4 pentest report will discuss the threat model around this endpoint.
	writeJSON(w, http.StatusNotImplemented, map[string]string{
		"error": "signing requires KMS integration",
	})
}

func GetMarket(w http.ResponseWriter, r *http.Request) {
	pair := chi.URLParam(r, "pair")
	// Trading pair validation — restrict charset
	matched, _ := regexp.MatchString(`^[A-Z]{2,8}-[A-Z]{2,8}$`, pair)
	if !matched {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid trading pair",
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"pair":  pair,
		"price": "0.00",
		"note":  "mock data",
	})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
