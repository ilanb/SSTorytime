package services

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

// fakeEmbeddingServer sert des vecteurs déterministes et compte les textes encodés,
// ce qui permet de vérifier que le cache évite les appels redondants.
type fakeEmbeddingServer struct {
	mu sync.Mutex
	// embeddedTexts liste tous les textes réellement transmis au serveur.
	embeddedTexts []string
	requestCount  int
	status        int
	// vectorFor produit le vecteur d'un texte; nil pour le comportement par défaut.
	vectorFor func(text string) []float64
}

func (f *fakeEmbeddingServer) start(t *testing.T) *httptest.Server {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("/models", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"data":[{"id":"multilingual-e5-base"}]}`))
	})
	mux.HandleFunc("/embeddings", func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		f.requestCount++
		status := f.status
		f.mu.Unlock()

		if status != 0 && status != http.StatusOK {
			w.WriteHeader(status)
			w.Write([]byte(`{"error":"indisponible"}`))
			return
		}

		var req embeddingRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		f.mu.Lock()
		f.embeddedTexts = append(f.embeddedTexts, req.Input...)
		f.mu.Unlock()

		type item struct {
			Index     int       `json:"index"`
			Embedding []float64 `json:"embedding"`
		}
		out := struct {
			Data []item `json:"data"`
		}{}

		for i, text := range req.Input {
			vec := defaultTestVector(text)
			if f.vectorFor != nil {
				vec = f.vectorFor(text)
			}
			out.Data = append(out.Data, item{Index: i, Embedding: vec})
		}

		json.NewEncoder(w).Encode(out)
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func (f *fakeEmbeddingServer) texts() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.embeddedTexts...)
}

// defaultTestVector produit un vecteur non normalisé dépendant du texte.
func defaultTestVector(text string) []float64 {
	return []float64{float64(len(text)), 2, 3}
}

func newTestClient(t *testing.T, f *fakeEmbeddingServer) *EmbeddingClient {
	t.Helper()
	srv := f.start(t)
	return NewEmbeddingClient(srv.URL, "multilingual-e5-base")
}

func TestNormalizeProducesUnitVector(t *testing.T) {
	got := normalize([]float64{3, 4})

	var sumSquares float64
	for _, v := range got {
		sumSquares += v * v
	}
	if math.Abs(math.Sqrt(sumSquares)-1) > 1e-9 {
		t.Fatalf("norme = %v, attendu 1", math.Sqrt(sumSquares))
	}
	if math.Abs(got[0]-0.6) > 1e-9 || math.Abs(got[1]-0.8) > 1e-9 {
		t.Fatalf("normalize([3,4]) = %v, attendu [0.6 0.8]", got)
	}
}

func TestNormalizeHandlesZeroVector(t *testing.T) {
	got := normalize([]float64{0, 0, 0})
	if len(got) != 3 {
		t.Fatalf("longueur = %d, attendu 3", len(got))
	}
	for _, v := range got {
		if v != 0 {
			t.Fatalf("normalize d'un vecteur nul a produit %v", got)
		}
	}
}

func TestNormalizeDoesNotMutateInput(t *testing.T) {
	input := []float64{3, 4}
	normalize(input)
	if input[0] != 3 || input[1] != 4 {
		t.Fatalf("l'entrée a été modifiée: %v", input)
	}
}

func TestDotProduct(t *testing.T) {
	tests := []struct {
		name string
		a, b []float64
		want float64
	}{
		{"vecteurs identiques", []float64{1, 0}, []float64{1, 0}, 1},
		{"vecteurs orthogonaux", []float64{1, 0}, []float64{0, 1}, 0},
		{"vecteurs opposés", []float64{1, 0}, []float64{-1, 0}, -1},
		{"tailles différentes", []float64{1, 0}, []float64{1, 0, 0}, 0},
		{"vecteurs vides", []float64{}, []float64{}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := dotProduct(tt.a, tt.b); math.Abs(got-tt.want) > 1e-9 {
				t.Fatalf("dotProduct = %v, attendu %v", got, tt.want)
			}
		})
	}
}

func TestEmbedAppliesE5Prefixes(t *testing.T) {
	fake := &fakeEmbeddingServer{}
	client := newTestClient(t, fake)

	if _, err := client.EmbedQuery("vol de voiture"); err != nil {
		t.Fatalf("EmbedQuery: %v", err)
	}
	if _, err := client.EmbedPassages([]string{"le suspect a fui"}); err != nil {
		t.Fatalf("EmbedPassages: %v", err)
	}

	texts := fake.texts()
	if len(texts) != 2 {
		t.Fatalf("%d textes encodés, attendu 2: %v", len(texts), texts)
	}
	if !strings.HasPrefix(texts[0], "query: ") {
		t.Errorf("requête sans préfixe e5: %q", texts[0])
	}
	if !strings.HasPrefix(texts[1], "passage: ") {
		t.Errorf("document sans préfixe e5: %q", texts[1])
	}
}

func TestEmbedCachesVectorsAcrossCalls(t *testing.T) {
	fake := &fakeEmbeddingServer{}
	client := newTestClient(t, fake)

	docs := []string{"alpha", "beta"}
	if _, err := client.EmbedPassages(docs); err != nil {
		t.Fatalf("premier appel: %v", err)
	}
	if got := len(fake.texts()); got != 2 {
		t.Fatalf("%d textes encodés au premier appel, attendu 2", got)
	}

	// Deuxième appel identique: rien ne doit repartir vers le serveur.
	if _, err := client.EmbedPassages(docs); err != nil {
		t.Fatalf("second appel: %v", err)
	}
	if got := len(fake.texts()); got != 2 {
		t.Fatalf("%d textes encodés après le second appel, attendu 2 (cache non utilisé)", got)
	}

	// Troisième appel: seul le document inédit doit être encodé.
	if _, err := client.EmbedPassages([]string{"alpha", "gamma"}); err != nil {
		t.Fatalf("troisième appel: %v", err)
	}
	texts := fake.texts()
	if len(texts) != 3 {
		t.Fatalf("%d textes encodés au total, attendu 3: %v", len(texts), texts)
	}
	if texts[2] != "passage: gamma" {
		t.Errorf("texte encodé = %q, attendu \"passage: gamma\"", texts[2])
	}
}

func TestEmbedDeduplicatesWithinBatch(t *testing.T) {
	fake := &fakeEmbeddingServer{}
	client := newTestClient(t, fake)

	vectors, err := client.EmbedPassages([]string{"doublon", "unique", "doublon"})
	if err != nil {
		t.Fatalf("EmbedPassages: %v", err)
	}
	if len(vectors) != 3 {
		t.Fatalf("%d vecteurs retournés, attendu 3", len(vectors))
	}

	if got := len(fake.texts()); got != 2 {
		t.Fatalf("%d textes encodés, attendu 2 (doublon non dédupliqué)", got)
	}
	if dotProduct(vectors[0], vectors[2]) < 0.999 {
		t.Error("les deux occurrences du doublon ont des vecteurs différents")
	}
}

func TestEmbedChunksLargeBatches(t *testing.T) {
	fake := &fakeEmbeddingServer{}
	client := newTestClient(t, fake)

	total := embeddingMaxBatchSize*2 + 5
	docs := make([]string, total)
	for i := range docs {
		docs[i] = fmt.Sprintf("document numéro %d", i)
	}

	vectors, err := client.EmbedPassages(docs)
	if err != nil {
		t.Fatalf("EmbedPassages: %v", err)
	}
	if len(vectors) != total {
		t.Fatalf("%d vecteurs, attendu %d", len(vectors), total)
	}

	fake.mu.Lock()
	requests := fake.requestCount
	fake.mu.Unlock()
	if requests != 3 {
		t.Errorf("%d requêtes HTTP, attendu 3 lots pour %d documents", requests, total)
	}
}

func TestCacheReturnsCopiesNotReferences(t *testing.T) {
	fake := &fakeEmbeddingServer{}
	client := newTestClient(t, fake)

	first, err := client.EmbedPassages([]string{"immuable"})
	if err != nil {
		t.Fatalf("EmbedPassages: %v", err)
	}

	// Altérer le vecteur retourné ne doit pas corrompre le cache.
	first[0][0] = 99999

	second, err := client.EmbedPassages([]string{"immuable"})
	if err != nil {
		t.Fatalf("second EmbedPassages: %v", err)
	}
	if second[0][0] == 99999 {
		t.Fatal("le cache expose ses vecteurs par référence: une mutation externe l'a corrompu")
	}
}

func TestSimilarityScoresRanksRelevantDocumentFirst(t *testing.T) {
	// Vecteurs orthogonaux: le document "pertinent" partage l'axe de la requête.
	fake := &fakeEmbeddingServer{
		vectorFor: func(text string) []float64 {
			if strings.Contains(text, "pertinent") || strings.HasPrefix(text, "query: ") {
				return []float64{1, 0}
			}
			return []float64{0, 1}
		},
	}
	client := newTestClient(t, fake)

	scores, err := client.SimilarityScores("recherche", []string{"hors sujet", "document pertinent"})
	if err != nil {
		t.Fatalf("SimilarityScores: %v", err)
	}
	if len(scores) != 2 {
		t.Fatalf("%d scores, attendu 2", len(scores))
	}
	if scores[1] <= scores[0] {
		t.Fatalf("le document pertinent (%v) n'est pas mieux classé que le hors-sujet (%v)", scores[1], scores[0])
	}
}

func TestSimilarityScoresPreservesDocumentOrder(t *testing.T) {
	fake := &fakeEmbeddingServer{}
	client := newTestClient(t, fake)

	docs := []string{"a", "bb", "ccc", "dddd"}
	scores, err := client.SimilarityScores("requête", docs)
	if err != nil {
		t.Fatalf("SimilarityScores: %v", err)
	}
	if len(scores) != len(docs) {
		t.Fatalf("%d scores pour %d documents", len(scores), len(docs))
	}
}

func TestSimilarityScoresEmptyDocuments(t *testing.T) {
	fake := &fakeEmbeddingServer{}
	client := newTestClient(t, fake)

	scores, err := client.SimilarityScores("requête", nil)
	if err != nil {
		t.Fatalf("SimilarityScores: %v", err)
	}
	if len(scores) != 0 {
		t.Fatalf("%d scores pour aucun document", len(scores))
	}
}

func TestEmbedReturnsErrorOnServerFailure(t *testing.T) {
	fake := &fakeEmbeddingServer{status: http.StatusServiceUnavailable}
	client := newTestClient(t, fake)

	if _, err := client.EmbedPassages([]string{"texte"}); err == nil {
		t.Fatal("une erreur était attendue quand le service répond 503")
	}
}

func TestEmbedReturnsErrorOnCountMismatch(t *testing.T) {
	fake := &fakeEmbeddingServer{}
	srv := fake.start(t)

	// Serveur qui renvoie moins de vecteurs que demandé.
	mux := http.NewServeMux()
	mux.HandleFunc("/embeddings", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"data":[{"index":0,"embedding":[1,0]}]}`))
	})
	broken := httptest.NewServer(mux)
	t.Cleanup(broken.Close)

	client := NewEmbeddingClient(broken.URL, "m")
	if _, err := client.EmbedPassages([]string{"un", "deux"}); err == nil {
		t.Fatal("une erreur était attendue quand le serveur renvoie trop peu de vecteurs")
	}
	_ = srv
}

func TestWarmUpCountsOnlyNewVectors(t *testing.T) {
	fake := &fakeEmbeddingServer{}
	client := newTestClient(t, fake)

	added, err := client.WarmUp([]string{"un", "deux", "trois"})
	if err != nil {
		t.Fatalf("WarmUp: %v", err)
	}
	if added != 3 {
		t.Fatalf("%d vecteurs ajoutés, attendu 3", added)
	}

	// Deux déjà connus, un nouveau.
	added, err = client.WarmUp([]string{"un", "deux", "quatre"})
	if err != nil {
		t.Fatalf("second WarmUp: %v", err)
	}
	if added != 1 {
		t.Fatalf("%d vecteurs ajoutés au second passage, attendu 1", added)
	}
	if client.CachedVectors() != 4 {
		t.Fatalf("%d vecteurs en cache, attendu 4", client.CachedVectors())
	}
}

func TestResetCacheClearsVectors(t *testing.T) {
	fake := &fakeEmbeddingServer{}
	client := newTestClient(t, fake)

	if _, err := client.EmbedPassages([]string{"un", "deux"}); err != nil {
		t.Fatalf("EmbedPassages: %v", err)
	}
	client.ResetCache()

	if client.CachedVectors() != 0 {
		t.Fatalf("%d vecteurs après ResetCache, attendu 0", client.CachedVectors())
	}
}

func TestIsAvailable(t *testing.T) {
	fake := &fakeEmbeddingServer{}
	client := newTestClient(t, fake)

	if !client.IsAvailable() {
		t.Error("IsAvailable = false alors que le serveur répond")
	}

	unreachable := NewEmbeddingClient("http://127.0.0.1:1", "m")
	if unreachable.IsAvailable() {
		t.Error("IsAvailable = true sur une adresse injoignable")
	}
}

func TestConcurrentEmbedIsRaceFree(t *testing.T) {
	fake := &fakeEmbeddingServer{}
	client := newTestClient(t, fake)

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			docs := []string{"partagé", fmt.Sprintf("propre-%d", n)}
			if _, err := client.EmbedPassages(docs); err != nil {
				t.Errorf("EmbedPassages concurrent: %v", err)
			}
		}(i)
	}
	wg.Wait()

	// 20 documents propres + 1 partagé.
	if got := client.CachedVectors(); got != 21 {
		t.Fatalf("%d vecteurs en cache, attendu 21", got)
	}
}
