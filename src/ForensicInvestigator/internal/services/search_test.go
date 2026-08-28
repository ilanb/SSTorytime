package services

import (
	"math"
	"testing"

	"forensicinvestigator/internal/models"
)

func TestNormalizeScoresSpreadsCompressedRange(t *testing.T) {
	// Plage typique des similarités e5: étroite et élevée, y compris hors sujet.
	scores := map[string]float64{
		"pertinent":  0.810,
		"moyen":      0.780,
		"hors-sujet": 0.742,
	}

	got := normalizeScores(scores)

	if math.Abs(got["pertinent"]-1) > 1e-9 {
		t.Errorf("meilleur score = %v, attendu 1", got["pertinent"])
	}
	if math.Abs(got["hors-sujet"]) > 1e-9 {
		t.Errorf("pire score = %v, attendu 0", got["hors-sujet"])
	}
	if got["moyen"] <= 0 || got["moyen"] >= 1 {
		t.Errorf("score intermédiaire = %v, attendu dans ]0,1[", got["moyen"])
	}
}

func TestNormalizeScoresPreservesRanking(t *testing.T) {
	scores := map[string]float64{"a": 0.79, "b": 0.81, "c": 0.75}
	got := normalizeScores(scores)

	if !(got["b"] > got["a"] && got["a"] > got["c"]) {
		t.Fatalf("l'ordre n'est pas préservé: %v", got)
	}
}

func TestNormalizeScoresAllEqual(t *testing.T) {
	// Aucune information discriminante: tout doit retomber à 0 plutôt que de
	// produire une division par zéro ou un score arbitraire.
	scores := map[string]float64{"a": 0.8, "b": 0.8, "c": 0.8}
	got := normalizeScores(scores)

	for id, v := range got {
		if v != 0 {
			t.Errorf("score[%s] = %v, attendu 0 quand tous les scores sont égaux", id, v)
		}
	}
}

func TestNormalizeScoresEmpty(t *testing.T) {
	got := normalizeScores(map[string]float64{})
	if len(got) != 0 {
		t.Fatalf("%d scores retournés pour une entrée vide", len(got))
	}
}

func TestNormalizeScoresSingleDocument(t *testing.T) {
	got := normalizeScores(map[string]float64{"seul": 0.8})
	if len(got) != 1 {
		t.Fatalf("%d scores, attendu 1", len(got))
	}
	if got["seul"] != 0 {
		t.Errorf("score = %v; un document unique n'a pas de pertinence relative", got["seul"])
	}
}

func TestNormalizeScoresHandlesNegatives(t *testing.T) {
	// La similarité cosinus peut être négative.
	scores := map[string]float64{"oppose": -0.4, "neutre": 0, "proche": 0.9}
	got := normalizeScores(scores)

	if math.Abs(got["oppose"]) > 1e-9 {
		t.Errorf("score le plus bas = %v, attendu 0", got["oppose"])
	}
	if math.Abs(got["proche"]-1) > 1e-9 {
		t.Errorf("score le plus haut = %v, attendu 1", got["proche"])
	}
}

func testCase() *models.Case {
	return &models.Case{
		ID: "affaire-1",
		Entities: []models.Entity{
			{ID: "e1", Name: "Marc Dubois", Type: "personne", Role: "temoin", Description: "voisin"},
		},
		Evidence: []models.Evidence{
			{ID: "p1", Name: "Empreinte", Description: "sur la poignée", Type: "physique", Location: "porte"},
		},
		Timeline: []models.Event{
			{ID: "t1", Title: "Effraction", Description: "fenêtre brisée", Location: "cuisine"},
		},
	}
}

func TestBuildDocumentsIncludesAllTypesByDefault(t *testing.T) {
	docs := buildDocuments(testCase(), nil)

	if len(docs) != 3 {
		t.Fatalf("%d documents, attendu 3", len(docs))
	}

	seen := make(map[string]bool)
	for _, d := range docs {
		seen[d.Type] = true
		if d.Content == "" {
			t.Errorf("document %s sans contenu", d.ID)
		}
		if len(d.Tokens) == 0 {
			t.Errorf("document %s sans tokens BM25", d.ID)
		}
	}
	for _, want := range []string{"entity", "evidence", "event"} {
		if !seen[want] {
			t.Errorf("type %q absent du corpus", want)
		}
	}
}

func TestBuildDocumentsFiltersByType(t *testing.T) {
	docs := buildDocuments(testCase(), []string{"evidence"})

	if len(docs) != 1 {
		t.Fatalf("%d documents, attendu 1", len(docs))
	}
	if docs[0].Type != "evidence" {
		t.Fatalf("type = %q, attendu \"evidence\"", docs[0].Type)
	}
}

func TestBuildDocumentsNilCase(t *testing.T) {
	if docs := buildDocuments(nil, nil); docs != nil {
		t.Fatalf("%d documents pour une affaire nulle", len(docs))
	}
}

// TestBuildDocumentsIsStable protège l'invariant dont dépend le cache: la
// réindexation et la recherche doivent produire des textes rigoureusement
// identiques, sinon aucun vecteur mémorisé n'est jamais retrouvé.
func TestBuildDocumentsIsStable(t *testing.T) {
	first := buildDocuments(testCase(), nil)
	second := buildDocuments(testCase(), nil)

	if len(first) != len(second) {
		t.Fatalf("tailles différentes: %d et %d", len(first), len(second))
	}
	for i := range first {
		if first[i].Content != second[i].Content {
			t.Fatalf("contenu instable à l'index %d:\n%q\n%q", i, first[i].Content, second[i].Content)
		}
	}
}

func TestReindexCaseWarmsCacheUsedBySearch(t *testing.T) {
	fake := &fakeEmbeddingServer{}
	srv := fake.start(t)

	service := NewSearchService(srv.URL, "multilingual-e5-base")

	added, err := service.ReindexCase(testCase())
	if err != nil {
		t.Fatalf("ReindexCase: %v", err)
	}
	if added != 3 {
		t.Fatalf("%d vecteurs indexés, attendu 3", added)
	}

	encodedAfterIndex := len(fake.texts())

	// Une recherche doit réutiliser l'index: seule la requête est encodée.
	_, err = service.HybridSearch(testCase(), SearchRequest{Query: "empreinte", Limit: 10})
	if err != nil {
		t.Fatalf("HybridSearch: %v", err)
	}

	newlyEncoded := len(fake.texts()) - encodedAfterIndex
	if newlyEncoded != 1 {
		t.Fatalf("%d textes encodés pendant la recherche, attendu 1 (la requête seule); "+
			"la réindexation ne remplit pas le cache utilisé par la recherche", newlyEncoded)
	}
}

func TestHybridSearchFallsBackToBM25WhenEmbeddingsUnavailable(t *testing.T) {
	service := NewSearchService("http://127.0.0.1:1", "m")

	results, err := service.HybridSearch(testCase(), SearchRequest{Query: "empreinte", Limit: 10})
	if err != nil {
		t.Fatalf("HybridSearch doit se replier sur BM25, pas échouer: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("aucun résultat: le repli BM25 n'a pas fonctionné")
	}
	for _, r := range results {
		if r.SemanticScore != 0 {
			t.Errorf("score sémantique non nul (%v) alors que le service est injoignable", r.SemanticScore)
		}
	}
}
