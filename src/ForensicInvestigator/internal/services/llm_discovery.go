package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// modelsResponse est la réponse de /v1/models (format OpenAI).
type modelsResponse struct {
	Data []struct {
		ID   string `json:"id"`
		Root string `json:"root"`
	} `json:"data"`
}

// ServedModel décrit un modèle tel qu'exposé par le serveur d'inférence.
type ServedModel struct {
	// ID est l'identifiant à envoyer dans le champ "model" des requêtes.
	ID string
	// Root est le modèle réellement chargé. Il diffère de ID lorsque le serveur
	// a été démarré avec --served-model-name.
	Root string
}

// IsAliased indique que l'identifiant d'API masque un autre modèle.
func (m ServedModel) IsAliased() bool {
	return m.Root != "" && m.Root != m.ID
}

// String décrit le modèle de façon lisible dans les journaux.
func (m ServedModel) String() string {
	if m.IsAliased() {
		return fmt.Sprintf("%s (alias de %s)", m.ID, m.Root)
	}
	return m.ID
}

// DiscoverServedModel interroge /v1/models et retourne le premier modèle servi.
//
// Cela évite de coder en dur l'identifiant attendu par le serveur: vLLM n'accepte
// que la valeur de --served-model-name, qui peut différer du modèle réellement
// chargé. L'application s'adapte ainsi à la configuration du serveur au lieu
// d'imposer la sienne, ce qui importe pour un serveur partagé entre applications.
func DiscoverServedModel(baseURL string) (ServedModel, error) {
	client := &http.Client{Timeout: 5 * time.Second}

	resp, err := client.Get(baseURL + "/v1/models")
	if err != nil {
		return ServedModel{}, fmt.Errorf("découverte du modèle: appel à %s: %w", baseURL, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ServedModel{}, fmt.Errorf("découverte du modèle: lecture de la réponse: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return ServedModel{}, fmt.Errorf("découverte du modèle: le serveur a répondu %d", resp.StatusCode)
	}

	var parsed modelsResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return ServedModel{}, fmt.Errorf("découverte du modèle: réponse illisible: %w", err)
	}

	if len(parsed.Data) == 0 {
		return ServedModel{}, fmt.Errorf("découverte du modèle: aucun modèle servi par %s", baseURL)
	}

	return ServedModel{ID: parsed.Data[0].ID, Root: parsed.Data[0].Root}, nil
}
