package services

import (
	"fmt"
	"os"
	"testing"
	"time"

	"forensicinvestigator/internal/models"
)

// Ces tests interrogent réellement les services du SPARK. Ils sont ignorés par
// défaut pour que `go test ./...` reste hors-ligne et déterministe.
//
//	SPARK_INTEGRATION=1 go test ./internal/services/ -run Integration -v
func requireSparkIntegration(t *testing.T) {
	t.Helper()
	if os.Getenv("SPARK_INTEGRATION") == "" {
		t.Skip("test d'intégration ignoré (définir SPARK_INTEGRATION=1 pour l'activer)")
	}
}

func TestIntegrationDiscoverServedModel(t *testing.T) {
	requireSparkIntegration(t)

	baseURL := getenvOr("LLM_BASE_URL", "http://86.204.69.30:8001")
	served, err := DiscoverServedModel(baseURL)
	if err != nil {
		t.Fatalf("DiscoverServedModel(%s): %v", baseURL, err)
	}

	if served.ID == "" {
		t.Fatal("identifiant de modèle vide")
	}
	t.Logf("modèle servi: %s", served)

	if served.IsAliased() {
		t.Logf("l'identifiant d'API %q masque le modèle réel %q", served.ID, served.Root)
	}
}

func TestIntegrationEmbeddingsDimensionAndNormalization(t *testing.T) {
	requireSparkIntegration(t)

	client := NewEmbeddingClient("", "")
	if !client.IsAvailable() {
		t.Fatalf("service d'embeddings injoignable: %s", client.BaseURL())
	}

	vectors, err := client.EmbedPassages([]string{"empreinte digitale sur la poignée"})
	if err != nil {
		t.Fatalf("EmbedPassages: %v", err)
	}

	if len(vectors[0]) != embeddingDimension {
		t.Fatalf("dimension = %d, attendu %d", len(vectors[0]), embeddingDimension)
	}

	// Après normalisation, le produit scalaire d'un vecteur avec lui-même vaut 1.
	if self := dotProduct(vectors[0], vectors[0]); self < 0.999 || self > 1.001 {
		t.Fatalf("vecteur non unitaire: produit scalaire avec lui-même = %v", self)
	}
}

// TestIntegrationSemanticSearchBeatsLexical vérifie l'apport réel du volet
// sémantique: retrouver un document pertinent qui ne partage aucun mot avec la
// requête, ce que BM25 seul ne peut pas faire.
func TestIntegrationSemanticSearchBeatsLexical(t *testing.T) {
	requireSparkIntegration(t)

	service := NewSearchService("", "")
	if !service.Embeddings().IsAvailable() {
		t.Fatalf("service d'embeddings injoignable: %s", service.Embeddings().BaseURL())
	}

	caseData := &models.Case{
		ID: "affaire-integration",
		Evidence: []models.Evidence{
			{ID: "p1", Name: "Relevé bancaire", Description: "virements répétés vers un compte à l'étranger", Type: "document"},
			{ID: "p2", Name: "Recette de cuisine", Description: "tarte aux pommes et pâte brisée", Type: "document"},
		},
	}

	// Aucun mot de la requête n'apparaît dans le document attendu.
	results, err := service.HybridSearch(caseData, SearchRequest{
		Query:      "blanchiment d'argent",
		Limit:      10,
		BM25Weight: 0.3,
	})
	if err != nil {
		t.Fatalf("HybridSearch: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("aucun résultat")
	}

	for _, r := range results {
		t.Logf("%-20s score=%.4f bm25=%.4f sem=%.4f", r.Name, r.Score, r.BM25Score, r.SemanticScore)
	}

	if results[0].ID != "p1" {
		t.Errorf("meilleur résultat = %q, attendu le relevé bancaire (p1)", results[0].ID)
	}
}

// TestIntegrationReindexAvoidsReencoding mesure le bénéfice de la réindexation:
// une fois le cache chaud, une recherche ne doit plus encoder que la requête.
func TestIntegrationReindexAvoidsReencoding(t *testing.T) {
	requireSparkIntegration(t)

	service := NewSearchService("", "")
	if !service.Embeddings().IsAvailable() {
		t.Fatalf("service d'embeddings injoignable: %s", service.Embeddings().BaseURL())
	}

	// Chaque description doit être unique, sinon la déduplication du cache produit
	// moins de vecteurs que de documents. Un compteur garantit l'unicité; un
	// horodatage ne le ferait pas (collisions possibles dans une boucle serrée).
	// Le suffixe aléatoire évite de réutiliser le cache d'une exécution précédente.
	runID := time.Now().UnixNano()
	caseData := &models.Case{ID: "affaire-perf"}
	for i := 0; i < 60; i++ {
		caseData.Evidence = append(caseData.Evidence, models.Evidence{
			ID:          fmt.Sprintf("piece-%d-%d", runID, i),
			Name:        "Pièce à conviction",
			Description: fmt.Sprintf("description détaillée numéro %d de l'exécution %d", i, runID),
			Type:        "physique",
		})
	}

	start := time.Now()
	added, err := service.ReindexCase(caseData)
	if err != nil {
		t.Fatalf("ReindexCase: %v", err)
	}
	indexDuration := time.Since(start)

	if added != len(caseData.Evidence) {
		t.Fatalf("%d vecteurs indexés pour %d pièces", added, len(caseData.Evidence))
	}

	req := SearchRequest{Query: "empreinte", Limit: 10}
	start = time.Now()
	if _, err := service.HybridSearch(caseData, req); err != nil {
		t.Fatalf("HybridSearch: %v", err)
	}
	warmDuration := time.Since(start)

	t.Logf("indexation de %d documents: %s | recherche à cache chaud: %s",
		added, indexDuration.Round(time.Millisecond), warmDuration.Round(time.Millisecond))

	if service.Embeddings().CachedVectors() < added {
		t.Errorf("%d vecteurs en cache, au moins %d attendus", service.Embeddings().CachedVectors(), added)
	}
}
