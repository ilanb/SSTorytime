package services

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// serveModels démarre un serveur qui répond à /v1/models avec le corps donné.
func serveModels(t *testing.T, status int, body string) string {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/models", func(w http.ResponseWriter, r *http.Request) {
		if status != 0 && status != http.StatusOK {
			w.WriteHeader(status)
		}
		w.Write([]byte(body))
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv.URL
}

func TestDiscoverServedModelDetectsAlias(t *testing.T) {
	// Cas réel du SPARK: vLLM sert Qwen3.8-27B-FP8 sous un nom hérité.
	url := serveModels(t, http.StatusOK,
		`{"data":[{"id":"Qwen3.5-9B","root":"Qwen/Qwen3.8-27B-FP8"}]}`)

	served, err := DiscoverServedModel(url)
	if err != nil {
		t.Fatalf("DiscoverServedModel: %v", err)
	}

	if served.ID != "Qwen3.5-9B" {
		t.Errorf("ID = %q, attendu \"Qwen3.5-9B\"", served.ID)
	}
	if !served.IsAliased() {
		t.Error("IsAliased = false alors que root diffère de id")
	}
	if got, want := served.String(), "Qwen3.5-9B (alias de Qwen/Qwen3.8-27B-FP8)"; got != want {
		t.Errorf("String() = %q, attendu %q", got, want)
	}
}

func TestDiscoverServedModelWithoutAlias(t *testing.T) {
	url := serveModels(t, http.StatusOK,
		`{"data":[{"id":"Qwen3.8-27B-FP8","root":"Qwen3.8-27B-FP8"}]}`)

	served, err := DiscoverServedModel(url)
	if err != nil {
		t.Fatalf("DiscoverServedModel: %v", err)
	}

	if served.IsAliased() {
		t.Error("IsAliased = true alors que id et root sont identiques")
	}
	if got := served.String(); got != "Qwen3.8-27B-FP8" {
		t.Errorf("String() = %q, attendu le nom seul", got)
	}
}

func TestDiscoverServedModelWithEmptyRoot(t *testing.T) {
	// Certains serveurs OpenAI-compatibles n'exposent pas "root".
	url := serveModels(t, http.StatusOK, `{"data":[{"id":"un-modele"}]}`)

	served, err := DiscoverServedModel(url)
	if err != nil {
		t.Fatalf("DiscoverServedModel: %v", err)
	}
	if served.IsAliased() {
		t.Error("IsAliased = true alors que root est absent")
	}
}

func TestDiscoverServedModelErrors(t *testing.T) {
	tests := []struct {
		name   string
		status int
		body   string
	}{
		{"aucun modèle servi", http.StatusOK, `{"data":[]}`},
		{"réponse illisible", http.StatusOK, `pas du json`},
		{"erreur serveur", http.StatusInternalServerError, `{"error":"boom"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := serveModels(t, tt.status, tt.body)
			if _, err := DiscoverServedModel(url); err == nil {
				t.Fatal("une erreur était attendue")
			}
		})
	}
}

func TestDiscoverServedModelUnreachableServer(t *testing.T) {
	if _, err := DiscoverServedModel("http://127.0.0.1:1"); err == nil {
		t.Fatal("une erreur était attendue pour un serveur injoignable")
	}
}

func TestEmbeddingClientAccessors(t *testing.T) {
	client := NewEmbeddingClient("http://exemple/v1", "un-modele")

	if got := client.BaseURL(); got != "http://exemple/v1" {
		t.Errorf("BaseURL = %q", got)
	}
	if got := client.Model(); got != "un-modele" {
		t.Errorf("Model = %q", got)
	}
}

func TestNewEmbeddingClientFallsBackToEnvThenDefaults(t *testing.T) {
	t.Setenv("EMBEDDING_BASE_URL", "http://depuis-env/v1")
	t.Setenv("EMBEDDING_MODEL", "modele-env")

	fromEnv := NewEmbeddingClient("", "")
	if fromEnv.BaseURL() != "http://depuis-env/v1" || fromEnv.Model() != "modele-env" {
		t.Errorf("les variables d'environnement ne sont pas prises en compte: %s / %s",
			fromEnv.BaseURL(), fromEnv.Model())
	}

	// Les arguments explicites restent prioritaires sur l'environnement.
	explicit := NewEmbeddingClient("http://explicite/v1", "modele-explicite")
	if explicit.BaseURL() != "http://explicite/v1" || explicit.Model() != "modele-explicite" {
		t.Errorf("les arguments explicites ont été ignorés: %s / %s",
			explicit.BaseURL(), explicit.Model())
	}
}
