package services

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"sync"
	"time"
)

// Configuration du service d'embeddings distant (SPARK GB10, port 8002).
//
// Le serveur expose une API compatible OpenAI (/v1/embeddings) servant
// multilingual-e5-base (768 dimensions), ainsi qu'un reranker bge-reranker-v2-m3.
const (
	defaultEmbeddingBaseURL = "http://86.204.69.30:8002/v1"
	defaultEmbeddingModel   = "multilingual-e5-base"
	embeddingDimension      = 768

	// Les modèles e5 sont entraînés avec des préfixes asymétriques: la requête et
	// les documents ne sont pas encodés de la même façon. Les omettre dégrade le
	// classement (MRR mesuré 0.538 sans préfixe contre 0.547 avec, sur corpus FR).
	embeddingQueryPrefix   = "query: "
	embeddingPassagePrefix = "passage: "

	// Le serveur accepte de gros lots (512 textes en ~1.7s) mais on découpe pour
	// borner la taille des requêtes HTTP et la latence unitaire.
	embeddingMaxBatchSize = 128

	embeddingTimeout      = 60 * time.Second
	embeddingProbeTimeout = 5 * time.Second
)

// embeddingRequest est le corps d'une requête /v1/embeddings.
type embeddingRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

// embeddingResponse est la réponse de /v1/embeddings.
type embeddingResponse struct {
	Data []struct {
		Index     int       `json:"index"`
		Embedding []float64 `json:"embedding"`
	} `json:"data"`
}

// EmbeddingClient interroge le service d'embeddings distant et mémorise les
// vecteurs déjà calculés.
//
// Le cache tient lieu d'index: la recherche hybride ré-encode sinon l'intégralité
// du corpus à chaque requête. Il est indexé par empreinte du texte préfixé, donc
// une modification d'un document invalide naturellement son entrée.
type EmbeddingClient struct {
	baseURL string
	model   string
	client  *http.Client

	mu    sync.RWMutex
	cache map[string][]float64
}

// NewEmbeddingClient crée un client d'embeddings.
//
// baseURL et model peuvent être vides: les valeurs sont alors lues dans
// EMBEDDING_BASE_URL / EMBEDDING_MODEL, puis dans les défauts SPARK.
func NewEmbeddingClient(baseURL, model string) *EmbeddingClient {
	if baseURL == "" {
		baseURL = getenvOr("EMBEDDING_BASE_URL", defaultEmbeddingBaseURL)
	}
	if model == "" {
		model = getenvOr("EMBEDDING_MODEL", defaultEmbeddingModel)
	}

	return &EmbeddingClient{
		baseURL: baseURL,
		model:   model,
		client:  &http.Client{Timeout: embeddingTimeout},
		cache:   make(map[string][]float64),
	}
}

// getenvOr retourne la variable d'environnement key, ou fallback si elle est vide.
func getenvOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// Model retourne le nom du modèle d'embeddings utilisé.
func (c *EmbeddingClient) Model() string { return c.model }

// BaseURL retourne l'URL du service d'embeddings.
func (c *EmbeddingClient) BaseURL() string { return c.baseURL }

// CachedVectors retourne le nombre de vecteurs actuellement mémorisés.
func (c *EmbeddingClient) CachedVectors() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.cache)
}

// ResetCache vide le cache de vecteurs.
//
// À utiliser lors d'un changement de modèle: des vecteurs produits par deux
// modèles différents ne sont pas comparables entre eux.
func (c *EmbeddingClient) ResetCache() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache = make(map[string][]float64)
}

// IsAvailable indique si le service d'embeddings répond.
func (c *EmbeddingClient) IsAvailable() bool {
	client := &http.Client{Timeout: embeddingProbeTimeout}
	resp, err := client.Get(c.baseURL + "/models")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)
	return resp.StatusCode == http.StatusOK
}

// cacheKey calcule l'empreinte d'un texte déjà préfixé.
func cacheKey(prefixedText string) string {
	sum := sha256.Sum256([]byte(prefixedText))
	return hex.EncodeToString(sum[:])
}

// normalize retourne une copie L2-normalisée du vecteur.
//
// Le serveur renvoie déjà des vecteurs unitaires, mais on ne s'en remet pas à
// cette garantie: la normalisation rend le produit scalaire équivalent au cosinus
// quel que soit le backend configuré.
func normalize(vec []float64) []float64 {
	var sumSquares float64
	for _, v := range vec {
		sumSquares += v * v
	}
	if sumSquares == 0 {
		return append([]float64(nil), vec...)
	}

	norm := math.Sqrt(sumSquares)
	out := make([]float64, len(vec))
	for i, v := range vec {
		out[i] = v / norm
	}
	return out
}

// dotProduct calcule le produit scalaire de deux vecteurs normalisés,
// c'est-à-dire leur similarité cosinus. Retourne 0 si les tailles diffèrent.
func dotProduct(a, b []float64) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}
	var sum float64
	for i := range a {
		sum += a[i] * b[i]
	}
	return sum
}

// fetch appelle le service d'embeddings pour une liste de textes déjà préfixés,
// en découpant en lots. Les vecteurs sont retournés dans l'ordre des entrées.
func (c *EmbeddingClient) fetch(prefixedTexts []string) ([][]float64, error) {
	vectors := make([][]float64, 0, len(prefixedTexts))

	for start := 0; start < len(prefixedTexts); start += embeddingMaxBatchSize {
		end := start + embeddingMaxBatchSize
		if end > len(prefixedTexts) {
			end = len(prefixedTexts)
		}

		batch, err := c.fetchBatch(prefixedTexts[start:end])
		if err != nil {
			return nil, err
		}
		vectors = append(vectors, batch...)
	}

	return vectors, nil
}

// fetchBatch effectue un unique appel /v1/embeddings.
func (c *EmbeddingClient) fetchBatch(prefixedTexts []string) ([][]float64, error) {
	body, err := json.Marshal(embeddingRequest{Model: c.model, Input: prefixedTexts})
	if err != nil {
		return nil, fmt.Errorf("embeddings: sérialisation de la requête: %w", err)
	}

	resp, err := c.client.Post(c.baseURL+"/embeddings", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("embeddings: appel à %s: %w", c.baseURL, err)
	}
	defer resp.Body.Close()

	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("embeddings: lecture de la réponse: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("embeddings: le service a répondu %d: %s", resp.StatusCode, truncateForError(payload))
	}

	var parsed embeddingResponse
	if err := json.Unmarshal(payload, &parsed); err != nil {
		return nil, fmt.Errorf("embeddings: réponse illisible: %w", err)
	}

	if len(parsed.Data) != len(prefixedTexts) {
		return nil, fmt.Errorf("embeddings: %d vecteurs reçus pour %d textes envoyés", len(parsed.Data), len(prefixedTexts))
	}

	// Le champ "index" fait foi pour l'ordre, l'ordre du tableau n'est pas garanti.
	vectors := make([][]float64, len(prefixedTexts))
	for _, item := range parsed.Data {
		if item.Index < 0 || item.Index >= len(vectors) {
			return nil, fmt.Errorf("embeddings: index %d hors bornes", item.Index)
		}
		if len(item.Embedding) == 0 {
			return nil, fmt.Errorf("embeddings: vecteur vide à l'index %d", item.Index)
		}
		vectors[item.Index] = normalize(item.Embedding)
	}

	for i, v := range vectors {
		if v == nil {
			return nil, fmt.Errorf("embeddings: aucun vecteur reçu pour l'index %d", i)
		}
	}

	return vectors, nil
}

// truncateForError borne la taille d'un corps d'erreur remonté à l'appelant.
func truncateForError(payload []byte) string {
	const maxLen = 200
	if len(payload) > maxLen {
		return string(payload[:maxLen]) + "..."
	}
	return string(payload)
}

// embedWithPrefix encode des textes, en ne sollicitant le service que pour ceux
// absents du cache. Les vecteurs retournés sont des copies: l'appelant peut les
// manipuler sans altérer le cache.
func (c *EmbeddingClient) embedWithPrefix(texts []string, prefix string) ([][]float64, error) {
	prefixed := make([]string, len(texts))
	keys := make([]string, len(texts))
	for i, t := range texts {
		prefixed[i] = prefix + t
		keys[i] = cacheKey(prefixed[i])
	}

	// Relever les manquants sans dupliquer les textes identiques du même lot.
	missingIdx := make([]int, 0)
	seen := make(map[string]bool, len(keys))

	c.mu.RLock()
	for i, key := range keys {
		if _, ok := c.cache[key]; ok || seen[key] {
			continue
		}
		seen[key] = true
		missingIdx = append(missingIdx, i)
	}
	c.mu.RUnlock()

	if len(missingIdx) > 0 {
		toFetch := make([]string, len(missingIdx))
		for i, idx := range missingIdx {
			toFetch[i] = prefixed[idx]
		}

		fetched, err := c.fetch(toFetch)
		if err != nil {
			return nil, err
		}

		c.mu.Lock()
		for i, idx := range missingIdx {
			c.cache[keys[idx]] = fetched[i]
		}
		c.mu.Unlock()
	}

	out := make([][]float64, len(texts))
	c.mu.RLock()
	defer c.mu.RUnlock()
	for i, key := range keys {
		vec, ok := c.cache[key]
		if !ok {
			return nil, fmt.Errorf("embeddings: vecteur manquant après récupération (texte %d)", i)
		}
		out[i] = append([]float64(nil), vec...)
	}

	return out, nil
}

// EmbedPassages encode des documents (préfixe "passage: ").
func (c *EmbeddingClient) EmbedPassages(texts []string) ([][]float64, error) {
	if len(texts) == 0 {
		return nil, nil
	}
	return c.embedWithPrefix(texts, embeddingPassagePrefix)
}

// EmbedQuery encode une requête de recherche (préfixe "query: ").
func (c *EmbeddingClient) EmbedQuery(text string) ([]float64, error) {
	vectors, err := c.embedWithPrefix([]string{text}, embeddingQueryPrefix)
	if err != nil {
		return nil, err
	}
	return vectors[0], nil
}

// SimilarityScores retourne la similarité cosinus entre une requête et chaque
// document, dans l'ordre des documents.
//
// Les scores sont bruts: les modèles e5 les concentrent dans une plage étroite
// (mesuré 0.74-0.81 sur corpus forensique FR). Ils doivent être normalisés avant
// d'être combinés à un score BM25, sous peine de n'apporter qu'un décalage
// constant. Voir normalizeScores dans search.go.
func (c *EmbeddingClient) SimilarityScores(query string, documents []string) ([]float64, error) {
	if len(documents) == 0 {
		return nil, nil
	}

	queryVec, err := c.EmbedQuery(query)
	if err != nil {
		return nil, fmt.Errorf("embeddings: encodage de la requête: %w", err)
	}

	docVecs, err := c.EmbedPassages(documents)
	if err != nil {
		return nil, fmt.Errorf("embeddings: encodage des documents: %w", err)
	}

	scores := make([]float64, len(documents))
	for i, vec := range docVecs {
		scores[i] = dotProduct(queryVec, vec)
	}

	return scores, nil
}

// WarmUp pré-calcule et met en cache les vecteurs d'une liste de documents.
//
// C'est l'opération de (ré)indexation: sans elle, la première recherche d'une
// affaire paie l'encodage de l'intégralité de son corpus. Retourne le nombre de
// vecteurs effectivement ajoutés au cache.
func (c *EmbeddingClient) WarmUp(documents []string) (int, error) {
	if len(documents) == 0 {
		return 0, nil
	}

	before := c.CachedVectors()
	if _, err := c.EmbedPassages(documents); err != nil {
		return 0, fmt.Errorf("embeddings: préchauffage du cache: %w", err)
	}
	return c.CachedVectors() - before, nil
}
