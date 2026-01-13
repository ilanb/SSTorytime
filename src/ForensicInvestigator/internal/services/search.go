package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"sort"
	"strings"
	"unicode"

	"forensicinvestigator/internal/models"
)

// SearchService gère la recherche hybride (BM25 + Sémantique via Model2vec)
type SearchService struct {
	model2vecURL string
}

// NewSearchService crée une nouvelle instance du service de recherche
func NewSearchService(ollamaURL, embeddingModel string) *SearchService {
	// On ignore les paramètres Ollama, on utilise Model2vec sur le port 8085
	return &SearchService{
		model2vecURL: "http://localhost:8085",
	}
}

// SearchResult représente un résultat de recherche
type SearchResult struct {
	ID            string   `json:"id"`
	Type          string   `json:"type"` // entity, evidence, event
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	Score         float64  `json:"score"`
	BM25Score     float64  `json:"bm25_score"`
	SemanticScore float64  `json:"semantic_score"`
	Highlights    []string `json:"highlights,omitempty"`
}

// SearchRequest représente une requête de recherche hybride
type SearchRequest struct {
	Query      string   `json:"query"`
	CaseID     string   `json:"case_id"`
	Types      []string `json:"types,omitempty"` // entity, evidence, event
	Limit      int      `json:"limit,omitempty"`
	BM25Weight float64  `json:"bm25_weight,omitempty"` // Poids de BM25 (0-1), semantic = 1 - bm25_weight
}

// Model2vecEmbedRequest représente une requête d'embedding à Model2vec
type Model2vecEmbedRequest struct {
	Text string `json:"text"`
}

// Model2vecEmbedResponse représente la réponse d'embedding de Model2vec
type Model2vecEmbedResponse struct {
	Embedding []float64 `json:"embedding"`
	Dimension int       `json:"dimension"`
}

// Model2vecSimilarityRequest représente une requête de similarité à Model2vec
type Model2vecSimilarityRequest struct {
	Query     string   `json:"query"`
	Documents []string `json:"documents"`
	TopK      int      `json:"top_k"`
}

// Model2vecSimilarityResult représente un résultat de similarité
type Model2vecSimilarityResult struct {
	Index int     `json:"index"`
	Text  string  `json:"text"`
	Score float64 `json:"score"`
}

// Model2vecSimilarityResponse représente la réponse de similarité de Model2vec
type Model2vecSimilarityResponse struct {
	Results []Model2vecSimilarityResult `json:"results"`
}

// Document représente un document indexable pour BM25
type Document struct {
	ID          string
	Type        string
	Name        string
	Description string
	Content     string // Contenu combiné pour la recherche
	Tokens      []string
}

// BM25Parameters contient les paramètres de l'algorithme BM25
type BM25Parameters struct {
	K1 float64 // Saturation du terme (typiquement 1.2-2.0)
	B  float64 // Normalisation de la longueur (typiquement 0.75)
}

// DefaultBM25Params retourne les paramètres BM25 par défaut
func DefaultBM25Params() BM25Parameters {
	return BM25Parameters{
		K1: 1.5,
		B:  0.75,
	}
}

// tokenize tokenise un texte en mots
func tokenize(text string) []string {
	text = strings.ToLower(text)
	var tokens []string
	var current strings.Builder

	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			current.WriteRune(r)
		} else if current.Len() > 0 {
			token := current.String()
			if len(token) > 1 { // Ignorer les tokens d'un seul caractère
				tokens = append(tokens, token)
			}
			current.Reset()
		}
	}

	if current.Len() > 1 {
		tokens = append(tokens, current.String())
	}

	return tokens
}

// computeIDF calcule l'IDF (Inverse Document Frequency) pour un terme
func computeIDF(term string, documents []Document) float64 {
	docCount := 0
	for _, doc := range documents {
		for _, token := range doc.Tokens {
			if token == term {
				docCount++
				break
			}
		}
	}

	if docCount == 0 {
		return 0
	}

	n := float64(len(documents))
	return math.Log((n - float64(docCount) + 0.5) / (float64(docCount) + 0.5) + 1)
}

// computeTF calcule le TF (Term Frequency) d'un terme dans un document
func computeTF(term string, tokens []string) float64 {
	count := 0
	for _, token := range tokens {
		if token == term {
			count++
		}
	}
	return float64(count)
}

// computeBM25Score calcule le score BM25 pour un document
func computeBM25Score(queryTokens []string, doc Document, documents []Document, avgDocLen float64, params BM25Parameters) float64 {
	score := 0.0
	docLen := float64(len(doc.Tokens))

	for _, term := range queryTokens {
		idf := computeIDF(term, documents)
		tf := computeTF(term, doc.Tokens)

		numerator := tf * (params.K1 + 1)
		denominator := tf + params.K1*(1-params.B+params.B*(docLen/avgDocLen))

		if denominator > 0 {
			score += idf * (numerator / denominator)
		}
	}

	return score
}

// getSemanticScores obtient les scores sémantiques via Model2vec
func (s *SearchService) getSemanticScores(query string, documents []Document) (map[string]float64, error) {
	scores := make(map[string]float64)

	// Préparer les contenus des documents
	var docContents []string
	for _, doc := range documents {
		docContents = append(docContents, doc.Content)
	}

	// Appeler le service Model2vec
	reqBody := Model2vecSimilarityRequest{
		Query:     query,
		Documents: docContents,
		TopK:      len(documents),
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("erreur marshalling similarity request: %w", err)
	}

	resp, err := http.Post(s.model2vecURL+"/similarity", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("erreur appel Model2vec: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Model2vec erreur %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("erreur lecture réponse: %w", err)
	}

	var simResp Model2vecSimilarityResponse
	if err := json.Unmarshal(body, &simResp); err != nil {
		return nil, fmt.Errorf("erreur parsing similarity response: %w", err)
	}

	// Mapper les scores aux IDs des documents
	for _, result := range simResp.Results {
		if result.Index >= 0 && result.Index < len(documents) {
			scores[documents[result.Index].ID] = result.Score
		}
	}

	return scores, nil
}

// isModel2vecAvailable vérifie si le service Model2vec est disponible
func (s *SearchService) isModel2vecAvailable() bool {
	resp, err := http.Get(s.model2vecURL + "/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// HybridSearch effectue une recherche hybride sur les données d'une affaire
func (s *SearchService) HybridSearch(caseData *models.Case, req SearchRequest) ([]SearchResult, error) {
	if req.Limit == 0 {
		req.Limit = 20
	}
	if req.BM25Weight == 0 {
		req.BM25Weight = 0.5 // 50% BM25, 50% Sémantique par défaut
	}

	// Construire les documents à partir des données de l'affaire
	var documents []Document

	// Types à inclure
	includeTypes := make(map[string]bool)
	if len(req.Types) == 0 {
		includeTypes["entity"] = true
		includeTypes["evidence"] = true
		includeTypes["event"] = true
	} else {
		for _, t := range req.Types {
			includeTypes[t] = true
		}
	}

	// Indexer les entités
	if includeTypes["entity"] {
		for _, entity := range caseData.Entities {
			content := entity.Name + " " + entity.Description + " " + string(entity.Type) + " " + string(entity.Role)
			documents = append(documents, Document{
				ID:          entity.ID,
				Type:        "entity",
				Name:        entity.Name,
				Description: entity.Description,
				Content:     content,
				Tokens:      tokenize(content),
			})
		}
	}

	// Indexer les preuves
	if includeTypes["evidence"] {
		for _, evidence := range caseData.Evidence {
			content := evidence.Name + " " + evidence.Description + " " + string(evidence.Type) + " " + evidence.Location
			documents = append(documents, Document{
				ID:          evidence.ID,
				Type:        "evidence",
				Name:        evidence.Name,
				Description: evidence.Description,
				Content:     content,
				Tokens:      tokenize(content),
			})
		}
	}

	// Indexer les événements de la timeline
	if includeTypes["event"] {
		for _, event := range caseData.Timeline {
			content := event.Title + " " + event.Description + " " + event.Location
			documents = append(documents, Document{
				ID:          event.ID,
				Type:        "event",
				Name:        event.Title,
				Description: event.Description,
				Content:     content,
				Tokens:      tokenize(content),
			})
		}
	}

	if len(documents) == 0 {
		return []SearchResult{}, nil
	}

	// Calculer la longueur moyenne des documents pour BM25
	totalTokens := 0
	for _, doc := range documents {
		totalTokens += len(doc.Tokens)
	}
	avgDocLen := float64(totalTokens) / float64(len(documents))

	// Tokeniser la requête
	queryTokens := tokenize(req.Query)

	// Calculer les scores BM25
	bm25Params := DefaultBM25Params()
	bm25Scores := make(map[string]float64)
	maxBM25 := 0.0

	for _, doc := range documents {
		score := computeBM25Score(queryTokens, doc, documents, avgDocLen, bm25Params)
		bm25Scores[doc.ID] = score
		if score > maxBM25 {
			maxBM25 = score
		}
	}

	// Normaliser les scores BM25
	if maxBM25 > 0 {
		for id := range bm25Scores {
			bm25Scores[id] = bm25Scores[id] / maxBM25
		}
	}

	// Calculer les scores sémantiques via Model2vec
	semanticScores := make(map[string]float64)
	semanticWeight := 1 - req.BM25Weight

	// Essayer d'obtenir les scores sémantiques
	if s.isModel2vecAvailable() {
		scores, err := s.getSemanticScores(req.Query, documents)
		if err == nil {
			semanticScores = scores
		} else {
			// Si Model2vec échoue, utiliser uniquement BM25
			semanticWeight = 0
			req.BM25Weight = 1
		}
	} else {
		// Model2vec non disponible, utiliser uniquement BM25
		semanticWeight = 0
		req.BM25Weight = 1
	}

	// Combiner les scores
	var results []SearchResult
	for _, doc := range documents {
		bm25Score := bm25Scores[doc.ID]
		semScore := semanticScores[doc.ID]

		// Score hybride pondéré
		hybridScore := req.BM25Weight*bm25Score + semanticWeight*semScore

		// Générer les highlights (mots correspondants)
		var highlights []string
		queryLower := strings.ToLower(req.Query)
		for _, token := range queryTokens {
			if strings.Contains(strings.ToLower(doc.Content), token) {
				highlights = append(highlights, token)
			}
		}

		// Ne garder que les résultats avec un score minimum
		if hybridScore > 0.01 || strings.Contains(strings.ToLower(doc.Content), queryLower) {
			results = append(results, SearchResult{
				ID:            doc.ID,
				Type:          doc.Type,
				Name:          doc.Name,
				Description:   doc.Description,
				Score:         hybridScore,
				BM25Score:     bm25Score,
				SemanticScore: semScore,
				Highlights:    highlights,
			})
		}
	}

	// Trier par score décroissant
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// Limiter les résultats
	if len(results) > req.Limit {
		results = results[:req.Limit]
	}

	return results, nil
}

// QuickBM25Search effectue une recherche BM25 rapide (sans embeddings)
func (s *SearchService) QuickBM25Search(caseData *models.Case, query string, limit int) []SearchResult {
	req := SearchRequest{
		Query:      query,
		Limit:      limit,
		BM25Weight: 1.0, // 100% BM25
	}

	results, _ := s.HybridSearch(caseData, req)
	return results
}

// ============================================
// Expansion Cones - Inspired by SSTorytime
// ============================================

// ConeDirection représente la direction d'expansion du cône
type ConeDirection string

const (
	ConeForward     ConeDirection = "forward"     // Cône vers l'avant (suivre les arêtes sortantes)
	ConeBackward    ConeDirection = "backward"    // Cône vers l'arrière (suivre les arêtes entrantes)
	ConeBidirectional ConeDirection = "bidirectional" // Cône dans les deux directions
)

// ConeSearchRequest représente une requête de recherche par cône d'expansion
type ConeSearchRequest struct {
	CaseID     string        `json:"case_id"`
	StartNode  string        `json:"start_node"`   // ID du nœud de départ
	Direction  ConeDirection `json:"direction"`    // forward, backward, bidirectional
	Depth      int           `json:"depth"`        // Profondeur maximale d'expansion
	Context    []string      `json:"context,omitempty"`   // Filtres contextuels optionnels
	ArrowTypes []string      `json:"arrow_types,omitempty"` // Types d'arêtes à suivre
	Limit      int           `json:"limit,omitempty"`      // Limite de nœuds
}

// ConeLevel représente un niveau du cône d'expansion
type ConeLevel struct {
	Level int                  `json:"level"`
	Nodes []ConeNode           `json:"nodes"`
	Edges []ConeEdge           `json:"edges"`
}

// ConeNode représente un nœud dans le cône avec ses métadonnées
type ConeNode struct {
	ID       string  `json:"id"`
	Label    string  `json:"label"`
	Type     string  `json:"type"`
	Distance int     `json:"distance"`    // Distance depuis le nœud source
	Weight   float64 `json:"weight"`      // Poids cumulé des chemins
}

// ConeEdge représente une arête dans le cône
type ConeEdge struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Label    string `json:"label"`
	Type     string `json:"type"`
	Context  string `json:"context,omitempty"`
	Level    int    `json:"level"`
}

// ConeSearchResult représente le résultat d'une recherche par cône
type ConeSearchResult struct {
	StartNode   string      `json:"start_node"`
	StartLabel  string      `json:"start_label"`
	Direction   string      `json:"direction"`
	Depth       int         `json:"depth"`
	Levels      []ConeLevel `json:"levels"`
	TotalNodes  int         `json:"total_nodes"`
	TotalEdges  int         `json:"total_edges"`
	Paths       []ConePath  `json:"paths,omitempty"`      // Chemins découverts
	Suggestions []string    `json:"suggestions,omitempty"` // Suggestions d'exploration
}

// ConePath représente un chemin découvert dans le cône
type ConePath struct {
	Nodes  []string `json:"nodes"`
	Labels []string `json:"labels"`
	Edges  []string `json:"edges"`  // Labels des arêtes
	Length int      `json:"length"`
}

// ConeSearch effectue une recherche par cône d'expansion sur le graphe
func (s *SearchService) ConeSearch(graph *models.GraphData, req ConeSearchRequest) (*ConeSearchResult, error) {
	if req.Depth <= 0 {
		req.Depth = 3
	}
	if req.Limit <= 0 {
		req.Limit = 100
	}
	if req.Direction == "" {
		req.Direction = ConeBidirectional
	}

	// Construire les maps d'adjacence selon la direction
	forwardAdj := make(map[string][]models.GraphEdge)
	backwardAdj := make(map[string][]models.GraphEdge)
	nodeMap := make(map[string]models.GraphNode)

	for _, node := range graph.Nodes {
		nodeMap[node.ID] = node
	}

	for _, edge := range graph.Edges {
		// Vérifier le filtre de type d'arête
		if len(req.ArrowTypes) > 0 {
			match := false
			for _, at := range req.ArrowTypes {
				if strings.EqualFold(edge.Label, at) || strings.EqualFold(edge.Type, at) {
					match = true
					break
				}
			}
			if !match {
				continue
			}
		}

		// Vérifier le filtre de contexte
		if len(req.Context) > 0 && edge.Context != "" {
			match := false
			for _, ctx := range req.Context {
				if strings.Contains(strings.ToLower(edge.Context), strings.ToLower(ctx)) {
					match = true
					break
				}
			}
			if !match {
				continue
			}
		}

		forwardAdj[edge.From] = append(forwardAdj[edge.From], edge)
		backwardAdj[edge.To] = append(backwardAdj[edge.To], edge)
	}

	// Trouver le nœud de départ
	startNode, exists := nodeMap[req.StartNode]
	if !exists {
		return nil, fmt.Errorf("nœud de départ '%s' non trouvé", req.StartNode)
	}

	// Structures pour le BFS
	visited := make(map[string]bool)
	nodeDistance := make(map[string]int)
	nodeWeight := make(map[string]float64)
	levels := make([]ConeLevel, 0)
	allPaths := make([]ConePath, 0)

	// File d'attente: (nodeID, distance, path)
	type queueItem struct {
		nodeID   string
		distance int
		path     []string
		edges    []string
	}

	queue := []queueItem{{req.StartNode, 0, []string{req.StartNode}, []string{}}}
	visited[req.StartNode] = true
	nodeDistance[req.StartNode] = 0
	nodeWeight[req.StartNode] = 1.0

	levelNodes := make(map[int][]ConeNode)
	levelEdges := make(map[int][]ConeEdge)

	// Ajouter le nœud de départ au niveau 0
	levelNodes[0] = append(levelNodes[0], ConeNode{
		ID:       startNode.ID,
		Label:    startNode.Label,
		Type:     startNode.Type,
		Distance: 0,
		Weight:   1.0,
	})

	totalNodes := 1
	totalEdges := 0

	for len(queue) > 0 && totalNodes < req.Limit {
		current := queue[0]
		queue = queue[1:]

		if current.distance >= req.Depth {
			continue
		}

		// Collecter les voisins selon la direction
		var neighbors []models.GraphEdge

		switch req.Direction {
		case ConeForward:
			neighbors = forwardAdj[current.nodeID]
		case ConeBackward:
			neighbors = backwardAdj[current.nodeID]
		case ConeBidirectional:
			neighbors = append(forwardAdj[current.nodeID], backwardAdj[current.nodeID]...)
		}

		for _, edge := range neighbors {
			// Déterminer le nœud cible
			targetID := edge.To
			if targetID == current.nodeID {
				targetID = edge.From
			}

			nextLevel := current.distance + 1

			// Ajouter l'arête au niveau actuel
			coneEdge := ConeEdge{
				From:    edge.From,
				To:      edge.To,
				Label:   edge.Label,
				Type:    edge.Type,
				Context: edge.Context,
				Level:   nextLevel,
			}
			levelEdges[nextLevel] = append(levelEdges[nextLevel], coneEdge)
			totalEdges++

			if !visited[targetID] {
				visited[targetID] = true
				nodeDistance[targetID] = nextLevel

				// Calculer le poids cumulé
				weight := nodeWeight[current.nodeID] * 0.8 // Décroissance
				nodeWeight[targetID] = weight

				// Ajouter le nœud
				targetNode := nodeMap[targetID]
				levelNodes[nextLevel] = append(levelNodes[nextLevel], ConeNode{
					ID:       targetID,
					Label:    targetNode.Label,
					Type:     targetNode.Type,
					Distance: nextLevel,
					Weight:   weight,
				})
				totalNodes++

				// Construire le chemin
				newPath := make([]string, len(current.path)+1)
				copy(newPath, current.path)
				newPath[len(current.path)] = targetID

				newEdges := make([]string, len(current.edges)+1)
				copy(newEdges, current.edges)
				newEdges[len(current.edges)] = edge.Label

				queue = append(queue, queueItem{targetID, nextLevel, newPath, newEdges})

				// Enregistrer le chemin
				if nextLevel >= 2 { // Chemins d'au moins 2 nœuds
					pathLabels := make([]string, len(newPath))
					for i, nodeID := range newPath {
						if n, ok := nodeMap[nodeID]; ok {
							pathLabels[i] = n.Label
						} else {
							pathLabels[i] = nodeID
						}
					}
					allPaths = append(allPaths, ConePath{
						Nodes:  newPath,
						Labels: pathLabels,
						Edges:  newEdges,
						Length: len(newPath),
					})
				}
			}
		}
	}

	// Construire les niveaux
	for level := 0; level <= req.Depth; level++ {
		if nodes, ok := levelNodes[level]; ok && len(nodes) > 0 {
			levels = append(levels, ConeLevel{
				Level: level,
				Nodes: nodes,
				Edges: levelEdges[level],
			})
		}
	}

	// Générer des suggestions
	suggestions := s.generateConeSuggestions(levels, totalNodes, totalEdges, req.Direction)

	// Limiter les chemins retournés
	if len(allPaths) > 20 {
		allPaths = allPaths[:20]
	}

	return &ConeSearchResult{
		StartNode:   req.StartNode,
		StartLabel:  startNode.Label,
		Direction:   string(req.Direction),
		Depth:       req.Depth,
		Levels:      levels,
		TotalNodes:  totalNodes,
		TotalEdges:  totalEdges,
		Paths:       allPaths,
		Suggestions: suggestions,
	}, nil
}

// generateConeSuggestions génère des suggestions basées sur l'analyse du cône
func (s *SearchService) generateConeSuggestions(levels []ConeLevel, totalNodes, totalEdges int, direction ConeDirection) []string {
	var suggestions []string

	if totalNodes == 1 {
		if direction == ConeForward {
			suggestions = append(suggestions, "Ce nœud n'a pas de connexions sortantes. Essayez la direction 'backward' ou 'bidirectional'.")
		} else if direction == ConeBackward {
			suggestions = append(suggestions, "Ce nœud n'a pas de connexions entrantes. Essayez la direction 'forward' ou 'bidirectional'.")
		} else {
			suggestions = append(suggestions, "Ce nœud est isolé dans le graphe. Considérez ajouter des relations.")
		}
	}

	if len(levels) > 0 && len(levels) < 3 {
		suggestions = append(suggestions, "Le cône est peu profond. Augmentez la profondeur pour explorer plus loin.")
	}

	if totalEdges > 0 {
		density := float64(totalEdges) / float64(totalNodes)
		if density > 2 {
			suggestions = append(suggestions, fmt.Sprintf("Zone dense détectée (%.1f relations/nœud). Nœud potentiellement central.", density))
		}
	}

	// Analyser la distribution des nœuds par niveau
	if len(levels) >= 2 {
		growth := float64(len(levels[len(levels)-1].Nodes)) / float64(len(levels[0].Nodes))
		if growth > 3 {
			suggestions = append(suggestions, "Expansion rapide du réseau. Ce nœud est un point de départ important.")
		}
	}

	return suggestions
}

// ForwardConeSearch effectue une recherche de cône avant (raccourci)
func (s *SearchService) ForwardConeSearch(graph *models.GraphData, startNode string, depth int) (*ConeSearchResult, error) {
	return s.ConeSearch(graph, ConeSearchRequest{
		StartNode: startNode,
		Direction: ConeForward,
		Depth:     depth,
	})
}

// BackwardConeSearch effectue une recherche de cône arrière (raccourci)
func (s *SearchService) BackwardConeSearch(graph *models.GraphData, startNode string, depth int) (*ConeSearchResult, error) {
	return s.ConeSearch(graph, ConeSearchRequest{
		StartNode: startNode,
		Direction: ConeBackward,
		Depth:     depth,
	})
}

// DiracPathSearch recherche des chemins entre deux ensembles de nœuds (notation Dirac <end|start>)
func (s *SearchService) DiracPathSearch(graph *models.GraphData, startNodes, endNodes []string, maxDepth int) ([]ConePath, error) {
	if maxDepth <= 0 {
		maxDepth = 5
	}

	// Construire l'adjacence bidirectionnelle
	adj := make(map[string][]struct {
		Target string
		Label  string
	})
	nodeMap := make(map[string]models.GraphNode)

	for _, node := range graph.Nodes {
		nodeMap[node.ID] = node
	}

	for _, edge := range graph.Edges {
		adj[edge.From] = append(adj[edge.From], struct {
			Target string
			Label  string
		}{edge.To, edge.Label})
		adj[edge.To] = append(adj[edge.To], struct {
			Target string
			Label  string
		}{edge.From, "← " + edge.Label})
	}

	// Créer un set des nœuds cibles
	endSet := make(map[string]bool)
	for _, end := range endNodes {
		endSet[end] = true
	}

	var allPaths []ConePath

	// DFS depuis chaque nœud de départ
	for _, start := range startNodes {
		var dfs func(current string, path []string, edges []string, visited map[string]bool, depth int)
		dfs = func(current string, path []string, edges []string, visited map[string]bool, depth int) {
			if depth > maxDepth {
				return
			}

			if endSet[current] && len(path) > 1 {
				// Chemin trouvé
				pathLabels := make([]string, len(path))
				for i, nodeID := range path {
					if n, ok := nodeMap[nodeID]; ok {
						pathLabels[i] = n.Label
					} else {
						pathLabels[i] = nodeID
					}
				}
				allPaths = append(allPaths, ConePath{
					Nodes:  append([]string{}, path...),
					Labels: pathLabels,
					Edges:  append([]string{}, edges...),
					Length: len(path),
				})
				return
			}

			for _, neighbor := range adj[current] {
				if !visited[neighbor.Target] {
					visited[neighbor.Target] = true
					dfs(neighbor.Target, append(path, neighbor.Target), append(edges, neighbor.Label), visited, depth+1)
					visited[neighbor.Target] = false
				}
			}
		}

		visited := make(map[string]bool)
		visited[start] = true
		dfs(start, []string{start}, []string{}, visited, 1)
	}

	// Trier par longueur et limiter
	sort.Slice(allPaths, func(i, j int) bool {
		return allPaths[i].Length < allPaths[j].Length
	})

	if len(allPaths) > 20 {
		allPaths = allPaths[:20]
	}

	return allPaths, nil
}

// ============================================
// Contrawave Search - Bidirectional Wavefront Collision
// Inspired by SSTorytime's contrawave algorithm
// ============================================

// ContrawaveRequest représente une requête de recherche contrawave
type ContrawaveRequest struct {
	CaseID     string   `json:"case_id"`
	StartNodes []string `json:"start_nodes"`  // Nœuds de départ (front avant)
	EndNodes   []string `json:"end_nodes"`    // Nœuds cibles (front arrière)
	MaxDepth   int      `json:"max_depth"`    // Profondeur maximale par front
	ArrowTypes []string `json:"arrow_types,omitempty"` // Filtres optionnels
}

// CollisionNode représente un nœud où deux fronts d'onde se rencontrent
type CollisionNode struct {
	NodeID        string   `json:"node_id"`
	NodeLabel     string   `json:"node_label"`
	NodeType      string   `json:"node_type"`
	ForwardDepth  int      `json:"forward_depth"`   // Distance depuis les nœuds de départ
	BackwardDepth int      `json:"backward_depth"`  // Distance depuis les nœuds cibles
	TotalDistance int      `json:"total_distance"`  // forward + backward
	Criticality   float64  `json:"criticality"`     // Importance du nœud (plus c'est central, plus c'est critique)
	PathsThrough  int      `json:"paths_through"`   // Nombre de chemins passant par ce nœud
	FromSources   []string `json:"from_sources"`    // Sources qui atteignent ce nœud
	ToTargets     []string `json:"to_targets"`      // Cibles atteignables depuis ce nœud
}

// ContrawaveResult représente le résultat d'une recherche contrawave
type ContrawaveResult struct {
	StartNodes     []string         `json:"start_nodes"`
	EndNodes       []string         `json:"end_nodes"`
	CollisionNodes []CollisionNode  `json:"collision_nodes"`
	Paths          []ConePath       `json:"paths"`
	ForwardWave    []WaveLevel      `json:"forward_wave"`
	BackwardWave   []WaveLevel      `json:"backward_wave"`
	WaveDepths     [2]int           `json:"wave_depths"`     // [forward_max, backward_max]
	TotalExpanded  int              `json:"total_expanded"`  // Nombre total de nœuds explorés
	Insights       []string         `json:"insights"`
}

// WaveLevel représente un niveau d'expansion dans une vague
type WaveLevel struct {
	Depth int      `json:"depth"`
	Nodes []string `json:"nodes"`
	Count int      `json:"count"`
}

// ContrawaveSearch effectue une recherche par collision de fronts d'onde
// Cette méthode est plus efficace que la recherche unidirectionnelle pour trouver
// des chemins entre deux ensembles de nœuds, et identifie les nœuds critiques
func (s *SearchService) ContrawaveSearch(graph *models.GraphData, req ContrawaveRequest) (*ContrawaveResult, error) {
	if req.MaxDepth <= 0 {
		req.MaxDepth = 5
	}
	if len(req.StartNodes) == 0 || len(req.EndNodes) == 0 {
		return nil, fmt.Errorf("start_nodes et end_nodes sont requis")
	}

	// Créer les maps de nœuds et d'adjacence
	nodeMap := make(map[string]models.GraphNode)
	forwardAdj := make(map[string][]models.GraphEdge)
	backwardAdj := make(map[string][]models.GraphEdge)

	for _, node := range graph.Nodes {
		nodeMap[node.ID] = node
	}

	for _, edge := range graph.Edges {
		// Appliquer le filtre de type d'arête si spécifié
		if len(req.ArrowTypes) > 0 {
			match := false
			for _, at := range req.ArrowTypes {
				if strings.EqualFold(edge.Label, at) || strings.EqualFold(edge.Type, at) {
					match = true
					break
				}
			}
			if !match {
				continue
			}
		}
		forwardAdj[edge.From] = append(forwardAdj[edge.From], edge)
		backwardAdj[edge.To] = append(backwardAdj[edge.To], edge)
	}

	// BFS depuis les nœuds de départ (front avant)
	forwardVisited := make(map[string]int)    // nodeID -> depth
	forwardSources := make(map[string][]string) // nodeID -> sources qui l'atteignent
	forwardWave := make([]WaveLevel, 0)

	forwardQueue := make([]string, len(req.StartNodes))
	copy(forwardQueue, req.StartNodes)
	for _, start := range req.StartNodes {
		forwardVisited[start] = 0
		forwardSources[start] = []string{start}
	}

	// Expansion du front avant
	for depth := 0; depth <= req.MaxDepth && len(forwardQueue) > 0; depth++ {
		levelNodes := make([]string, 0)
		nextQueue := make([]string, 0)

		for _, nodeID := range forwardQueue {
			if forwardVisited[nodeID] != depth {
				continue
			}
			levelNodes = append(levelNodes, nodeID)

			for _, edge := range forwardAdj[nodeID] {
				if _, visited := forwardVisited[edge.To]; !visited {
					forwardVisited[edge.To] = depth + 1
					forwardSources[edge.To] = append([]string{}, forwardSources[nodeID]...)
					nextQueue = append(nextQueue, edge.To)
				} else if forwardVisited[edge.To] == depth+1 {
					// Ajouter les sources supplémentaires
					for _, src := range forwardSources[nodeID] {
						found := false
						for _, existing := range forwardSources[edge.To] {
							if existing == src {
								found = true
								break
							}
						}
						if !found {
							forwardSources[edge.To] = append(forwardSources[edge.To], src)
						}
					}
				}
			}
		}

		if len(levelNodes) > 0 {
			forwardWave = append(forwardWave, WaveLevel{
				Depth: depth,
				Nodes: levelNodes,
				Count: len(levelNodes),
			})
		}
		forwardQueue = nextQueue
	}

	// BFS depuis les nœuds cibles (front arrière)
	backwardVisited := make(map[string]int)     // nodeID -> depth
	backwardTargets := make(map[string][]string) // nodeID -> targets atteignables
	backwardWave := make([]WaveLevel, 0)

	backwardQueue := make([]string, len(req.EndNodes))
	copy(backwardQueue, req.EndNodes)
	for _, end := range req.EndNodes {
		backwardVisited[end] = 0
		backwardTargets[end] = []string{end}
	}

	// Expansion du front arrière
	for depth := 0; depth <= req.MaxDepth && len(backwardQueue) > 0; depth++ {
		levelNodes := make([]string, 0)
		nextQueue := make([]string, 0)

		for _, nodeID := range backwardQueue {
			if backwardVisited[nodeID] != depth {
				continue
			}
			levelNodes = append(levelNodes, nodeID)

			for _, edge := range backwardAdj[nodeID] {
				if _, visited := backwardVisited[edge.From]; !visited {
					backwardVisited[edge.From] = depth + 1
					backwardTargets[edge.From] = append([]string{}, backwardTargets[nodeID]...)
					nextQueue = append(nextQueue, edge.From)
				} else if backwardVisited[edge.From] == depth+1 {
					// Ajouter les targets supplémentaires
					for _, tgt := range backwardTargets[nodeID] {
						found := false
						for _, existing := range backwardTargets[edge.From] {
							if existing == tgt {
								found = true
								break
							}
						}
						if !found {
							backwardTargets[edge.From] = append(backwardTargets[edge.From], tgt)
						}
					}
				}
			}
		}

		if len(levelNodes) > 0 {
			backwardWave = append(backwardWave, WaveLevel{
				Depth: depth,
				Nodes: levelNodes,
				Count: len(levelNodes),
			})
		}
		backwardQueue = nextQueue
	}

	// Trouver les nœuds de collision (atteints par les deux fronts)
	collisionNodes := make([]CollisionNode, 0)
	maxCriticality := 0.0

	for nodeID, fwdDepth := range forwardVisited {
		if bwdDepth, found := backwardVisited[nodeID]; found {
			// Ce nœud est atteint par les deux fronts
			totalDist := fwdDepth + bwdDepth
			pathsThrough := len(forwardSources[nodeID]) * len(backwardTargets[nodeID])

			// Calculer la criticité: nœuds plus centraux avec plus de chemins = plus critiques
			criticality := float64(pathsThrough) / float64(totalDist+1)
			if criticality > maxCriticality {
				maxCriticality = criticality
			}

			node := nodeMap[nodeID]
			collisionNodes = append(collisionNodes, CollisionNode{
				NodeID:        nodeID,
				NodeLabel:     node.Label,
				NodeType:      node.Type,
				ForwardDepth:  fwdDepth,
				BackwardDepth: bwdDepth,
				TotalDistance: totalDist,
				Criticality:   criticality,
				PathsThrough:  pathsThrough,
				FromSources:   forwardSources[nodeID],
				ToTargets:     backwardTargets[nodeID],
			})
		}
	}

	// Normaliser la criticité
	if maxCriticality > 0 {
		for i := range collisionNodes {
			collisionNodes[i].Criticality = collisionNodes[i].Criticality / maxCriticality
		}
	}

	// Trier par criticité décroissante
	sort.Slice(collisionNodes, func(i, j int) bool {
		return collisionNodes[i].Criticality > collisionNodes[j].Criticality
	})

	// Calculer les profondeurs maximales atteintes
	maxForward := 0
	maxBackward := 0
	for _, depth := range forwardVisited {
		if depth > maxForward {
			maxForward = depth
		}
	}
	for _, depth := range backwardVisited {
		if depth > maxBackward {
			maxBackward = depth
		}
	}

	// Reconstruire les chemins à travers les nœuds de collision les plus critiques
	paths := s.reconstructContrawavePaths(graph, nodeMap, forwardAdj, collisionNodes, req.StartNodes, req.EndNodes, 10)

	// Générer des insights
	insights := s.generateContrawaveInsights(collisionNodes, len(forwardVisited), len(backwardVisited), req)

	return &ContrawaveResult{
		StartNodes:     req.StartNodes,
		EndNodes:       req.EndNodes,
		CollisionNodes: collisionNodes,
		Paths:          paths,
		ForwardWave:    forwardWave,
		BackwardWave:   backwardWave,
		WaveDepths:     [2]int{maxForward, maxBackward},
		TotalExpanded:  len(forwardVisited) + len(backwardVisited),
		Insights:       insights,
	}, nil
}

// reconstructContrawavePaths reconstruit les chemins à travers les nœuds de collision
func (s *SearchService) reconstructContrawavePaths(graph *models.GraphData, nodeMap map[string]models.GraphNode, adj map[string][]models.GraphEdge, collisions []CollisionNode, starts, ends []string, maxPaths int) []ConePath {
	var paths []ConePath

	// Limiter aux top nœuds de collision
	topCollisions := collisions
	if len(topCollisions) > 5 {
		topCollisions = collisions[:5]
	}

	// Pour chaque nœud de collision critique, reconstruire un chemin
	for _, collision := range topCollisions {
		if len(paths) >= maxPaths {
			break
		}

		// Trouver un chemin start -> collision -> end
		for _, start := range collision.FromSources {
			for _, end := range collision.ToTargets {
				if len(paths) >= maxPaths {
					break
				}

				// BFS pour trouver le chemin
				path := s.findPathBFS(start, end, adj, nodeMap)
				if len(path.Nodes) > 0 {
					paths = append(paths, path)
				}
			}
		}
	}

	return paths
}

// findPathBFS trouve un chemin entre deux nœuds via BFS
func (s *SearchService) findPathBFS(from, to string, adj map[string][]models.GraphEdge, nodeMap map[string]models.GraphNode) ConePath {
	if from == to {
		node := nodeMap[from]
		return ConePath{
			Nodes:  []string{from},
			Labels: []string{node.Label},
			Edges:  []string{},
			Length: 1,
		}
	}

	visited := make(map[string]bool)
	parent := make(map[string]string)
	edgeLabel := make(map[string]string)
	queue := []string{from}
	visited[from] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		for _, edge := range adj[current] {
			if !visited[edge.To] {
				visited[edge.To] = true
				parent[edge.To] = current
				edgeLabel[edge.To] = edge.Label
				queue = append(queue, edge.To)

				if edge.To == to {
					// Reconstruire le chemin
					var nodes, labels, edges []string
					for node := to; node != ""; node = parent[node] {
						nodes = append([]string{node}, nodes...)
						labels = append([]string{nodeMap[node].Label}, labels...)
						if label, ok := edgeLabel[node]; ok {
							edges = append([]string{label}, edges...)
						}
						if node == from {
							break
						}
					}
					return ConePath{
						Nodes:  nodes,
						Labels: labels,
						Edges:  edges,
						Length: len(nodes),
					}
				}
			}
		}
	}

	return ConePath{}
}

// generateContrawaveInsights génère des insights sur la recherche contrawave
func (s *SearchService) generateContrawaveInsights(collisions []CollisionNode, fwdExpanded, bwdExpanded int, req ContrawaveRequest) []string {
	var insights []string

	if len(collisions) == 0 {
		insights = append(insights, "Aucune collision détectée. Les ensembles de nœuds ne sont pas connectés dans la limite de profondeur spécifiée.")
		return insights
	}

	// Analyser les collisions
	insights = append(insights, fmt.Sprintf("%d nœud(s) de collision détecté(s) entre les fronts d'onde.", len(collisions)))

	// Nœud le plus critique
	if len(collisions) > 0 {
		top := collisions[0]
		insights = append(insights, fmt.Sprintf("'%s' est le point de passage le plus critique (criticité: %.2f, chemins: %d).",
			top.NodeLabel, top.Criticality, top.PathsThrough))
	}

	// Analyser l'efficacité
	expandedNodes := fwdExpanded + bwdExpanded
	efficiency := float64(len(collisions)) / float64(expandedNodes)

	if efficiency > 0.1 {
		insights = append(insights, "Recherche efficace: de nombreux points de connexion trouvés.")
	} else if efficiency < 0.01 {
		insights = append(insights, "Les groupes sont faiblement connectés. Peu de chemins directs existent.")
	}

	// Analyser la distance
	if len(collisions) > 0 {
		minDist := collisions[0].TotalDistance
		maxDist := collisions[0].TotalDistance
		for _, c := range collisions {
			if c.TotalDistance < minDist {
				minDist = c.TotalDistance
			}
			if c.TotalDistance > maxDist {
				maxDist = c.TotalDistance
			}
		}
		if maxDist > minDist {
			insights = append(insights, fmt.Sprintf("Les chemins varient de %d à %d sauts.", minDist, maxDist))
		}
	}

	// Bottlenecks
	bottlenecks := 0
	for _, c := range collisions {
		if c.Criticality > 0.8 {
			bottlenecks++
		}
	}
	if bottlenecks > 0 {
		insights = append(insights, fmt.Sprintf("%d goulot(s) d'étranglement identifié(s) - nœuds clés pour la connectivité.", bottlenecks))
	}

	return insights
}

// ============================================
// Super-Nodes Detection - Functional Equivalence
// Inspired by SSTorytime's supernode algorithm
// ============================================

// SuperNodeGroup représente un groupe de nœuds fonctionnellement équivalents
type SuperNodeGroup struct {
	GroupID      string      `json:"group_id"`
	Nodes        []SuperNode `json:"nodes"`
	Size         int         `json:"size"`
	Equivalence  string      `json:"equivalence"`  // Type d'équivalence: "structural", "role", "flow"
	CommonInLinks  []string  `json:"common_in_links"`   // Arêtes entrantes communes
	CommonOutLinks []string  `json:"common_out_links"`  // Arêtes sortantes communes
	Replaceable  bool        `json:"replaceable"`  // Si les nœuds peuvent se substituer
	Description  string      `json:"description"`
}

// SuperNode représente un nœud dans un groupe de super-nœuds
type SuperNode struct {
	NodeID     string  `json:"node_id"`
	NodeLabel  string  `json:"node_label"`
	NodeType   string  `json:"node_type"`
	InDegree   int     `json:"in_degree"`
	OutDegree  int     `json:"out_degree"`
	Similarity float64 `json:"similarity"`  // Similarité avec le groupe
}

// SuperNodesResult représente le résultat de la détection de super-nœuds
type SuperNodesResult struct {
	Groups          []SuperNodeGroup `json:"groups"`
	TotalGroups     int              `json:"total_groups"`
	TotalNodes      int              `json:"total_nodes"`  // Nœuds dans des groupes
	SimilarityThreshold float64      `json:"similarity_threshold"`
	Insights        []string         `json:"insights"`
}

// SuperNodesRequest représente une requête de détection de super-nœuds
type SuperNodesRequest struct {
	CaseID              string  `json:"case_id"`
	SimilarityThreshold float64 `json:"similarity_threshold"`  // 0.0 à 1.0
	MinGroupSize        int     `json:"min_group_size"`
	IncludeTypes        []string `json:"include_types,omitempty"` // Types de nœuds à considérer
}

// DetectSuperNodes détecte les groupes de nœuds fonctionnellement équivalents
// Ces nœuds occupent des positions similaires dans le graphe et peuvent être interchangeables
func (s *SearchService) DetectSuperNodes(graph *models.GraphData, req SuperNodesRequest) (*SuperNodesResult, error) {
	if req.SimilarityThreshold <= 0 {
		req.SimilarityThreshold = 0.7
	}
	if req.MinGroupSize <= 0 {
		req.MinGroupSize = 2
	}

	// Créer les structures de données
	nodeMap := make(map[string]models.GraphNode)
	inLinks := make(map[string]map[string]bool)  // nodeID -> set of (source+label)
	outLinks := make(map[string]map[string]bool) // nodeID -> set of (target+label)
	inDegree := make(map[string]int)
	outDegree := make(map[string]int)

	// Filtrer les nœuds par type si spécifié
	includeTypes := make(map[string]bool)
	if len(req.IncludeTypes) > 0 {
		for _, t := range req.IncludeTypes {
			includeTypes[strings.ToLower(t)] = true
		}
	}

	for _, node := range graph.Nodes {
		if len(includeTypes) > 0 && !includeTypes[strings.ToLower(node.Type)] {
			continue
		}
		nodeMap[node.ID] = node
		inLinks[node.ID] = make(map[string]bool)
		outLinks[node.ID] = make(map[string]bool)
	}

	// Construire les profils de connexion
	for _, edge := range graph.Edges {
		if _, ok := nodeMap[edge.From]; ok {
			outLinks[edge.From][edge.To+"|"+edge.Label] = true
			outDegree[edge.From]++
		}
		if _, ok := nodeMap[edge.To]; ok {
			inLinks[edge.To][edge.From+"|"+edge.Label] = true
			inDegree[edge.To]++
		}
	}

	// Calculer la similarité Jaccard entre chaque paire de nœuds
	nodeIDs := make([]string, 0, len(nodeMap))
	for id := range nodeMap {
		nodeIDs = append(nodeIDs, id)
	}

	// Union-Find pour regrouper les nœuds similaires
	parent := make(map[string]string)
	for _, id := range nodeIDs {
		parent[id] = id
	}

	var find func(x string) string
	find = func(x string) string {
		if parent[x] != x {
			parent[x] = find(parent[x])
		}
		return parent[x]
	}

	union := func(x, y string) {
		px, py := find(x), find(y)
		if px != py {
			parent[px] = py
		}
	}

	// Calculer les similarités et regrouper
	similarityMatrix := make(map[string]map[string]float64)
	for i := 0; i < len(nodeIDs); i++ {
		nodeI := nodeIDs[i]
		similarityMatrix[nodeI] = make(map[string]float64)

		for j := i + 1; j < len(nodeIDs); j++ {
			nodeJ := nodeIDs[j]

			// Calculer la similarité des liens entrants
			inSim := jaccardSimilarity(inLinks[nodeI], inLinks[nodeJ])
			// Calculer la similarité des liens sortants
			outSim := jaccardSimilarity(outLinks[nodeI], outLinks[nodeJ])

			// Similarité combinée
			similarity := (inSim + outSim) / 2.0

			// Bonus si même type
			if nodeMap[nodeI].Type == nodeMap[nodeJ].Type {
				similarity += 0.1
				if similarity > 1.0 {
					similarity = 1.0
				}
			}

			similarityMatrix[nodeI][nodeJ] = similarity

			if similarity >= req.SimilarityThreshold {
				union(nodeI, nodeJ)
			}
		}
	}

	// Regrouper les nœuds par leur racine
	groupMap := make(map[string][]string)
	for _, id := range nodeIDs {
		root := find(id)
		groupMap[root] = append(groupMap[root], id)
	}

	// Construire les résultats
	var groups []SuperNodeGroup
	totalNodes := 0
	groupID := 0

	for _, members := range groupMap {
		if len(members) < req.MinGroupSize {
			continue
		}

		// Créer le groupe
		nodes := make([]SuperNode, len(members))
		commonIn := findCommonLinks(members, inLinks)
		commonOut := findCommonLinks(members, outLinks)

		// Calculer la similarité moyenne au groupe
		for i, nodeID := range members {
			node := nodeMap[nodeID]

			// Similarité moyenne avec les autres membres
			avgSim := 0.0
			count := 0
			for _, other := range members {
				if other != nodeID {
					if sim, ok := similarityMatrix[nodeID][other]; ok {
						avgSim += sim
						count++
					} else if sim, ok := similarityMatrix[other][nodeID]; ok {
						avgSim += sim
						count++
					}
				}
			}
			if count > 0 {
				avgSim /= float64(count)
			}

			nodes[i] = SuperNode{
				NodeID:     nodeID,
				NodeLabel:  node.Label,
				NodeType:   node.Type,
				InDegree:   inDegree[nodeID],
				OutDegree:  outDegree[nodeID],
				Similarity: avgSim,
			}
		}

		// Déterminer le type d'équivalence
		equivalence := "structural"
		if len(commonIn) > 0 && len(commonOut) > 0 {
			equivalence = "flow"
		} else if len(commonIn) > 0 || len(commonOut) > 0 {
			equivalence = "role"
		}

		// Vérifier si les nœuds sont remplaçables
		replaceable := len(commonIn) > 0 || len(commonOut) > 0

		// Générer une description
		description := fmt.Sprintf("Groupe de %d nœuds avec équivalence %s", len(members), equivalence)
		if len(commonIn) > 0 {
			description += fmt.Sprintf(", %d liens entrants communs", len(commonIn))
		}
		if len(commonOut) > 0 {
			description += fmt.Sprintf(", %d liens sortants communs", len(commonOut))
		}

		groups = append(groups, SuperNodeGroup{
			GroupID:        fmt.Sprintf("supernode_%d", groupID),
			Nodes:          nodes,
			Size:           len(members),
			Equivalence:    equivalence,
			CommonInLinks:  commonIn,
			CommonOutLinks: commonOut,
			Replaceable:    replaceable,
			Description:    description,
		})

		totalNodes += len(members)
		groupID++
	}

	// Trier par taille décroissante
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].Size > groups[j].Size
	})

	// Générer des insights
	insights := s.generateSuperNodesInsights(groups, len(nodeIDs), req.SimilarityThreshold)

	return &SuperNodesResult{
		Groups:              groups,
		TotalGroups:         len(groups),
		TotalNodes:          totalNodes,
		SimilarityThreshold: req.SimilarityThreshold,
		Insights:            insights,
	}, nil
}

// jaccardSimilarity calcule la similarité de Jaccard entre deux ensembles
func jaccardSimilarity(set1, set2 map[string]bool) float64 {
	if len(set1) == 0 && len(set2) == 0 {
		return 1.0 // Deux ensembles vides sont identiques
	}
	if len(set1) == 0 || len(set2) == 0 {
		return 0.0
	}

	intersection := 0
	union := len(set1)

	for k := range set2 {
		if set1[k] {
			intersection++
		} else {
			union++
		}
	}

	return float64(intersection) / float64(union)
}

// findCommonLinks trouve les liens communs à tous les nœuds d'un groupe
func findCommonLinks(nodes []string, links map[string]map[string]bool) []string {
	if len(nodes) == 0 {
		return []string{}
	}

	// Commencer avec les liens du premier nœud
	common := make(map[string]bool)
	for link := range links[nodes[0]] {
		common[link] = true
	}

	// Intersection avec les liens des autres nœuds
	for i := 1; i < len(nodes); i++ {
		newCommon := make(map[string]bool)
		for link := range common {
			if links[nodes[i]][link] {
				newCommon[link] = true
			}
		}
		common = newCommon
	}

	result := make([]string, 0, len(common))
	for link := range common {
		result = append(result, link)
	}
	return result
}

// generateSuperNodesInsights génère des insights sur les super-nœuds détectés
func (s *SearchService) generateSuperNodesInsights(groups []SuperNodeGroup, totalNodes int, threshold float64) []string {
	var insights []string

	if len(groups) == 0 {
		insights = append(insights, "Aucun groupe de nœuds équivalents détecté avec le seuil de similarité actuel.")
		insights = append(insights, "Essayez de réduire le seuil de similarité pour trouver plus de groupes.")
		return insights
	}

	// Statistiques globales
	nodesInGroups := 0
	for _, g := range groups {
		nodesInGroups += g.Size
	}
	coverage := float64(nodesInGroups) / float64(totalNodes) * 100

	insights = append(insights, fmt.Sprintf("%d groupe(s) de nœuds équivalents détecté(s) (%.1f%% des nœuds).", len(groups), coverage))

	// Analyser le plus grand groupe
	if len(groups) > 0 {
		largest := groups[0]
		insights = append(insights, fmt.Sprintf("Plus grand groupe: %d nœuds avec équivalence '%s'.", largest.Size, largest.Equivalence))

		if largest.Replaceable {
			insights = append(insights, "Ces nœuds peuvent potentiellement se substituer les uns aux autres dans l'enquête.")
		}
	}

	// Compter par type d'équivalence
	equivCount := make(map[string]int)
	for _, g := range groups {
		equivCount[g.Equivalence]++
	}

	if equivCount["flow"] > 0 {
		insights = append(insights, fmt.Sprintf("%d groupe(s) avec équivalence de flux - ces nœuds ont les mêmes entrées ET sorties.", equivCount["flow"]))
	}
	if equivCount["role"] > 0 {
		insights = append(insights, fmt.Sprintf("%d groupe(s) avec équivalence de rôle - ces nœuds jouent des rôles similaires.", equivCount["role"]))
	}

	return insights
}

// ============================================
// Betweenness Centrality - Path Flow Analysis
// Enhanced version inspired by SSTorytime
// ============================================

// BetweennessResult représente le résultat du calcul de betweenness centrality
type BetweennessResult struct {
	Centralities    []BetweennessNode `json:"centralities"`
	TotalPaths      int               `json:"total_paths"`
	AverageBetweenness float64        `json:"average_betweenness"`
	MaxBetweenness  float64           `json:"max_betweenness"`
	Insights        []string          `json:"insights"`
}

// BetweennessNode représente un nœud avec sa betweenness centrality
type BetweennessNode struct {
	NodeID        string  `json:"node_id"`
	NodeLabel     string  `json:"node_label"`
	NodeType      string  `json:"node_type"`
	Betweenness   float64 `json:"betweenness"`    // Betweenness brute
	Normalized    float64 `json:"normalized"`     // Betweenness normalisée [0,1]
	PathsThrough  int     `json:"paths_through"`  // Nombre de chemins passant par ce nœud
	Rank          int     `json:"rank"`
	Role          string  `json:"role"`           // "bridge", "hub", "peripheral"
}

// CalculateBetweennessCentrality calcule la betweenness centrality pour tous les nœuds
// Version améliorée inspirée de SSTorytime qui compte les apparitions dans les chemins
func (s *SearchService) CalculateBetweennessCentrality(graph *models.GraphData) (*BetweennessResult, error) {
	n := len(graph.Nodes)
	if n == 0 {
		return &BetweennessResult{Centralities: []BetweennessNode{}}, nil
	}

	// Créer les structures
	nodeMap := make(map[string]models.GraphNode)
	nodeIndex := make(map[string]int)
	indexNode := make(map[int]string)

	for i, node := range graph.Nodes {
		nodeMap[node.ID] = node
		nodeIndex[node.ID] = i
		indexNode[i] = node.ID
	}

	// Construire la liste d'adjacence bidirectionnelle
	adj := make(map[string][]string)
	for _, edge := range graph.Edges {
		adj[edge.From] = append(adj[edge.From], edge.To)
		adj[edge.To] = append(adj[edge.To], edge.From)
	}

	// Calculer la betweenness pour chaque nœud
	betweenness := make(map[string]float64)
	pathsThrough := make(map[string]int)
	totalPaths := 0

	// Pour chaque paire de nœuds, trouver les plus courts chemins
	for i := 0; i < n; i++ {
		sourceID := indexNode[i]

		// BFS depuis la source pour trouver les plus courts chemins
		dist := make(map[string]int)
		sigma := make(map[string]float64)  // Nombre de plus courts chemins
		pred := make(map[string][]string)  // Prédécesseurs sur les plus courts chemins

		dist[sourceID] = 0
		sigma[sourceID] = 1.0

		queue := []string{sourceID}
		stack := []string{}

		for len(queue) > 0 {
			v := queue[0]
			queue = queue[1:]
			stack = append(stack, v)

			for _, w := range adj[v] {
				// Premier chemin vers w
				if _, seen := dist[w]; !seen {
					dist[w] = dist[v] + 1
					queue = append(queue, w)
				}
				// Chemin le plus court vers w via v?
				if dist[w] == dist[v]+1 {
					sigma[w] += sigma[v]
					pred[w] = append(pred[w], v)
				}
			}
		}

		// Accumulation de la betweenness
		delta := make(map[string]float64)
		for len(stack) > 0 {
			w := stack[len(stack)-1]
			stack = stack[:len(stack)-1]

			for _, v := range pred[w] {
				contribution := (sigma[v] / sigma[w]) * (1.0 + delta[w])
				delta[v] += contribution
			}
			if w != sourceID {
				betweenness[w] += delta[w]
				pathsThrough[w]++
				totalPaths++
			}
		}
	}

	// Normaliser (le graphe non-orienté compte chaque chemin 2 fois)
	maxBetweenness := 0.0
	for _, b := range betweenness {
		if b > maxBetweenness {
			maxBetweenness = b
		}
	}

	// Construire les résultats
	results := make([]BetweennessNode, n)
	totalBetweenness := 0.0

	for i := 0; i < n; i++ {
		nodeID := indexNode[i]
		node := nodeMap[nodeID]
		b := betweenness[nodeID] / 2.0 // Corriger pour le double comptage

		normalized := 0.0
		if maxBetweenness > 0 {
			normalized = b / (maxBetweenness / 2.0)
		}

		// Déterminer le rôle
		role := "peripheral"
		if normalized >= 0.7 {
			role = "bridge"  // Nœud pont critique
		} else if normalized >= 0.3 {
			role = "hub"     // Nœud important
		}

		results[i] = BetweennessNode{
			NodeID:       nodeID,
			NodeLabel:    node.Label,
			NodeType:     node.Type,
			Betweenness:  b,
			Normalized:   normalized,
			PathsThrough: pathsThrough[nodeID],
			Role:         role,
		}
		totalBetweenness += b
	}

	// Trier par betweenness décroissante
	sort.Slice(results, func(i, j int) bool {
		return results[i].Betweenness > results[j].Betweenness
	})

	// Assigner les rangs
	for i := range results {
		results[i].Rank = i + 1
	}

	// Calculer la moyenne
	avgBetweenness := 0.0
	if n > 0 {
		avgBetweenness = totalBetweenness / float64(n)
	}

	// Générer des insights
	insights := s.generateBetweennessInsights(results, totalPaths, avgBetweenness)

	return &BetweennessResult{
		Centralities:       results,
		TotalPaths:         totalPaths / 2, // Corriger pour le double comptage
		AverageBetweenness: avgBetweenness,
		MaxBetweenness:     maxBetweenness / 2.0,
		Insights:           insights,
	}, nil
}

// generateBetweennessInsights génère des insights sur la betweenness centrality
func (s *SearchService) generateBetweennessInsights(results []BetweennessNode, totalPaths int, avgBetweenness float64) []string {
	var insights []string

	if len(results) == 0 {
		return insights
	}

	// Compter les rôles
	bridges := 0
	hubs := 0
	for _, r := range results {
		switch r.Role {
		case "bridge":
			bridges++
		case "hub":
			hubs++
		}
	}

	if bridges > 0 {
		insights = append(insights, fmt.Sprintf("%d nœud(s) pont(s) critique(s) identifié(s) - leur suppression fragmenterait le réseau.", bridges))
	}

	if hubs > 0 {
		insights = append(insights, fmt.Sprintf("%d nœud(s) hub(s) important(s) - ils facilitent de nombreuses connexions.", hubs))
	}

	// Analyser le top nœud
	if len(results) > 0 && results[0].Normalized > 0.5 {
		top := results[0]
		insights = append(insights, fmt.Sprintf("'%s' est le nœud le plus central (betweenness normalisée: %.2f).", top.NodeLabel, top.Normalized))
		insights = append(insights, "Ce nœud contrôle le flux d'information dans le graphe.")
	}

	// Analyser la distribution
	if len(results) >= 5 {
		top5Sum := 0.0
		for i := 0; i < 5; i++ {
			top5Sum += results[i].Betweenness
		}
		total := 0.0
		for _, r := range results {
			total += r.Betweenness
		}

		if total > 0 {
			concentration := top5Sum / total * 100
			if concentration > 80 {
				insights = append(insights, fmt.Sprintf("Forte concentration: les 5 premiers nœuds contrôlent %.0f%% du flux.", concentration))
			} else if concentration < 40 {
				insights = append(insights, "Le flux est bien distribué dans le graphe.")
			}
		}
	}

	return insights
}

// ============================================
// Constrained Paths - Filtered Path Search
// Paths limited by relation types
// ============================================

// ConstrainedPathRequest représente une requête de chemins contraints
type ConstrainedPathRequest struct {
	CaseID        string   `json:"case_id"`
	FromNode      string   `json:"from_node"`
	ToNode        string   `json:"to_node"`
	AllowedTypes  []string `json:"allowed_types,omitempty"`  // Types de relations autorisés
	ExcludedTypes []string `json:"excluded_types,omitempty"` // Types de relations exclus
	MaxDepth      int      `json:"max_depth"`
	MaxPaths      int      `json:"max_paths"`
}

// ConstrainedPathResult représente le résultat de la recherche
type ConstrainedPathResult struct {
	FromNode      string             `json:"from_node"`
	FromLabel     string             `json:"from_label"`
	ToNode        string             `json:"to_node"`
	ToLabel       string             `json:"to_label"`
	Paths         []ConstrainedPath  `json:"paths"`
	TotalPaths    int                `json:"total_paths"`
	UsedTypes     map[string]int     `json:"used_types"`      // Types de relations utilisés
	FilteredEdges int                `json:"filtered_edges"`  // Arêtes filtrées
	Insights      []string           `json:"insights"`
}

// ConstrainedPath représente un chemin avec détails des arêtes
type ConstrainedPath struct {
	Nodes      []string         `json:"nodes"`
	Labels     []string         `json:"labels"`
	Edges      []PathEdgeDetail `json:"edges"`
	Length     int              `json:"length"`
	TypesUsed  []string         `json:"types_used"`
}

// PathEdgeDetail représente une arête avec ses détails
type PathEdgeDetail struct {
	From      string `json:"from"`
	To        string `json:"to"`
	Label     string `json:"label"`
	EdgeType  string `json:"edge_type"`
}

// FindConstrainedPaths trouve des chemins filtrés par type de relation
func (s *SearchService) FindConstrainedPaths(graph *models.GraphData, req ConstrainedPathRequest) (*ConstrainedPathResult, error) {
	if req.MaxDepth <= 0 {
		req.MaxDepth = 5
	}
	if req.MaxPaths <= 0 {
		req.MaxPaths = 10
	}

	// Créer les structures
	nodeMap := make(map[string]models.GraphNode)
	for _, node := range graph.Nodes {
		nodeMap[node.ID] = node
	}

	// Créer une liste d'adjacence filtrée
	allowedSet := make(map[string]bool)
	for _, t := range req.AllowedTypes {
		allowedSet[strings.ToLower(t)] = true
	}
	excludedSet := make(map[string]bool)
	for _, t := range req.ExcludedTypes {
		excludedSet[strings.ToLower(t)] = true
	}

	filteredEdges := 0
	usedTypes := make(map[string]int)
	adj := make(map[string][]models.GraphEdge)

	for _, edge := range graph.Edges {
		edgeType := strings.ToLower(edge.Label)
		if edge.Type != "" {
			edgeType = strings.ToLower(edge.Type)
		}

		// Appliquer les filtres
		if len(allowedSet) > 0 {
			match := false
			for allowed := range allowedSet {
				if strings.Contains(edgeType, allowed) || strings.Contains(strings.ToLower(edge.Label), allowed) {
					match = true
					break
				}
			}
			if !match {
				filteredEdges++
				continue
			}
		}

		if len(excludedSet) > 0 {
			excluded := false
			for ex := range excludedSet {
				if strings.Contains(edgeType, ex) || strings.Contains(strings.ToLower(edge.Label), ex) {
					excluded = true
					break
				}
			}
			if excluded {
				filteredEdges++
				continue
			}
		}

		adj[edge.From] = append(adj[edge.From], edge)
		adj[edge.To] = append(adj[edge.To], edge) // Bidirectionnel
	}

	// BFS pour trouver les chemins
	var paths []ConstrainedPath
	type pathState struct {
		current string
		path    []string
		edges   []models.GraphEdge
	}

	queue := []pathState{{current: req.FromNode, path: []string{req.FromNode}, edges: nil}}
	visited := make(map[string]bool)

	for len(queue) > 0 && len(paths) < req.MaxPaths {
		state := queue[0]
		queue = queue[1:]

		if len(state.path) > req.MaxDepth+1 {
			continue
		}

		if state.current == req.ToNode && len(state.path) > 1 {
			// Chemin trouvé
			labels := make([]string, len(state.path))
			for i, nid := range state.path {
				labels[i] = nodeMap[nid].Label
			}

			edgeDetails := make([]PathEdgeDetail, len(state.edges))
			typesUsed := make([]string, len(state.edges))
			for i, e := range state.edges {
				edgeDetails[i] = PathEdgeDetail{
					From:     e.From,
					To:       e.To,
					Label:    e.Label,
					EdgeType: e.Type,
				}
				typesUsed[i] = e.Label
				usedTypes[e.Label]++
			}

			paths = append(paths, ConstrainedPath{
				Nodes:     state.path,
				Labels:    labels,
				Edges:     edgeDetails,
				Length:    len(state.path),
				TypesUsed: typesUsed,
			})
			continue
		}

		// Explorer les voisins
		for _, edge := range adj[state.current] {
			var nextNode string
			if edge.From == state.current {
				nextNode = edge.To
			} else {
				nextNode = edge.From
			}

			// Éviter les cycles dans ce chemin
			inPath := false
			for _, n := range state.path {
				if n == nextNode {
					inPath = true
					break
				}
			}
			if inPath {
				continue
			}

			// Éviter de revisiter trop de fois le même nœud globalement
			visitKey := state.current + "->" + nextNode
			if visited[visitKey] && len(paths) > 3 {
				continue
			}
			visited[visitKey] = true

			newPath := append([]string{}, state.path...)
			newPath = append(newPath, nextNode)
			newEdges := append([]models.GraphEdge{}, state.edges...)
			newEdges = append(newEdges, edge)

			queue = append(queue, pathState{
				current: nextNode,
				path:    newPath,
				edges:   newEdges,
			})
		}
	}

	// Trier par longueur
	sort.Slice(paths, func(i, j int) bool {
		return paths[i].Length < paths[j].Length
	})

	// Générer les insights
	insights := s.generateConstrainedPathInsights(paths, filteredEdges, usedTypes, req)

	fromLabel := nodeMap[req.FromNode].Label
	toLabel := nodeMap[req.ToNode].Label

	return &ConstrainedPathResult{
		FromNode:      req.FromNode,
		FromLabel:     fromLabel,
		ToNode:        req.ToNode,
		ToLabel:       toLabel,
		Paths:         paths,
		TotalPaths:    len(paths),
		UsedTypes:     usedTypes,
		FilteredEdges: filteredEdges,
		Insights:      insights,
	}, nil
}

// generateConstrainedPathInsights génère des insights sur les chemins contraints
func (s *SearchService) generateConstrainedPathInsights(paths []ConstrainedPath, filtered int, usedTypes map[string]int, req ConstrainedPathRequest) []string {
	var insights []string

	if len(paths) == 0 {
		insights = append(insights, "Aucun chemin trouvé avec les contraintes spécifiées.")
		if filtered > 0 {
			insights = append(insights, fmt.Sprintf("%d arêtes filtrées par les contraintes de type.", filtered))
		}
		if len(req.AllowedTypes) > 0 {
			insights = append(insights, "Essayez d'élargir les types de relations autorisés.")
		}
		return insights
	}

	insights = append(insights, fmt.Sprintf("%d chemin(s) trouvé(s) entre les nœuds.", len(paths)))

	if len(paths) > 0 {
		shortest := paths[0]
		insights = append(insights, fmt.Sprintf("Chemin le plus court: %d nœuds.", shortest.Length))
	}

	// Analyser les types utilisés
	if len(usedTypes) > 0 {
		maxType := ""
		maxCount := 0
		for t, c := range usedTypes {
			if c > maxCount {
				maxCount = c
				maxType = t
			}
		}
		insights = append(insights, fmt.Sprintf("Type de relation le plus utilisé: '%s' (%d fois).", maxType, maxCount))
	}

	if filtered > 0 {
		insights = append(insights, fmt.Sprintf("%d arêtes exclues par les filtres.", filtered))
	}

	return insights
}

// ============================================
// Dirac Notation - Quantum-inspired Path Search
// <target|source> notation for path queries
// ============================================

// DiracRequest représente une requête en notation Dirac
type DiracRequest struct {
	CaseID        string `json:"case_id"`
	Query         string `json:"query"`         // Format: <target|source> ou <A|B>
	MaxDepth      int    `json:"max_depth"`
	MaxPaths      int    `json:"max_paths"`
	Bidirectional bool   `json:"bidirectional"` // Chercher aussi <source|target>
}

// DiracResult représente le résultat d'une recherche Dirac
type DiracResult struct {
	Query         string     `json:"query"`
	Source        string     `json:"source"`
	SourceLabel   string     `json:"source_label"`
	Target        string     `json:"target"`
	TargetLabel   string     `json:"target_label"`
	ForwardPaths  []ConePath `json:"forward_paths"`  // source -> target
	BackwardPaths []ConePath `json:"backward_paths"` // target -> source (si bidirectionnel)
	TotalPaths    int        `json:"total_paths"`
	Symmetric     bool       `json:"symmetric"` // Les chemins sont-ils symétriques?
	Insights      []string   `json:"insights"`
}

// ParseDiracNotation parse une notation Dirac <target|source>
func ParseDiracNotation(query string) (target, source string, err error) {
	query = strings.TrimSpace(query)

	// Vérifier le format <...|...>
	if !strings.HasPrefix(query, "<") || !strings.HasSuffix(query, ">") {
		return "", "", fmt.Errorf("format invalide: utilisez <cible|source>")
	}

	// Extraire le contenu
	content := query[1 : len(query)-1]
	parts := strings.Split(content, "|")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("format invalide: utilisez <cible|source>")
	}

	target = strings.TrimSpace(parts[0])
	source = strings.TrimSpace(parts[1])

	if target == "" || source == "" {
		return "", "", fmt.Errorf("cible et source doivent être non-vides")
	}

	return target, source, nil
}

// SearchDirac effectue une recherche en notation Dirac
func (s *SearchService) SearchDirac(graph *models.GraphData, req DiracRequest) (*DiracResult, error) {
	if req.MaxDepth <= 0 {
		req.MaxDepth = 5
	}
	if req.MaxPaths <= 0 {
		req.MaxPaths = 10
	}

	// Parser la notation
	target, source, err := ParseDiracNotation(req.Query)
	if err != nil {
		return nil, err
	}

	// Résoudre les noms en IDs
	nodeMap := make(map[string]models.GraphNode)
	namesToIDs := make(map[string]string)
	for _, node := range graph.Nodes {
		nodeMap[node.ID] = node
		// Enregistrer par label exact et ID exact
		namesToIDs[strings.ToLower(node.Label)] = node.ID
		namesToIDs[strings.ToLower(node.ID)] = node.ID
	}

	// Fonction pour résoudre un nom (exact, partiel, ou par rôle)
	resolveNodeID := func(name string) string {
		nameLower := strings.ToLower(name)
		// D'abord essayer exact
		if id, ok := namesToIDs[nameLower]; ok {
			return id
		}
		// Chercher par correspondance partielle sur le label
		for _, node := range graph.Nodes {
			labelLower := strings.ToLower(node.Label)
			if strings.Contains(labelLower, nameLower) || strings.Contains(nameLower, labelLower) {
				return node.ID
			}
		}
		// Chercher par type de nœud (personne, lieu, organisation, etc.)
		for _, node := range graph.Nodes {
			if strings.ToLower(node.Type) == nameLower {
				return node.ID
			}
		}
		// Chercher par rôle (victime, suspect, témoin)
		for _, node := range graph.Nodes {
			if node.Role != "" && strings.ToLower(node.Role) == nameLower {
				return node.ID
			}
			// Chercher aussi dans Data["role"]
			if node.Data != nil {
				if role, ok := node.Data["role"]; ok {
					if strings.ToLower(role) == nameLower {
						return node.ID
					}
				}
			}
		}
		return ""
	}

	sourceID := resolveNodeID(source)
	targetID := resolveNodeID(target)

	if sourceID == "" {
		// Lister les nœuds disponibles pour aide
		availableNodes := make([]string, 0, len(graph.Nodes))
		for _, node := range graph.Nodes {
			availableNodes = append(availableNodes, node.Label)
		}
		return nil, fmt.Errorf("source '%s' non trouvée. Nœuds disponibles: %v", source, availableNodes)
	}
	if targetID == "" {
		availableNodes := make([]string, 0, len(graph.Nodes))
		for _, node := range graph.Nodes {
			availableNodes = append(availableNodes, node.Label)
		}
		return nil, fmt.Errorf("cible '%s' non trouvée. Nœuds disponibles: %v", target, availableNodes)
	}

	// Construire l'adjacence
	adj := make(map[string][]models.GraphEdge)
	for _, edge := range graph.Edges {
		adj[edge.From] = append(adj[edge.From], edge)
	}

	// Recherche forward: source -> target
	forwardPaths := s.findAllPathsBFS(sourceID, targetID, adj, nodeMap, req.MaxDepth, req.MaxPaths)

	// Recherche backward: target -> source (si bidirectionnel)
	var backwardPaths []ConePath
	if req.Bidirectional {
		backwardPaths = s.findAllPathsBFS(targetID, sourceID, adj, nodeMap, req.MaxDepth, req.MaxPaths)
	}

	// Vérifier la symétrie
	symmetric := len(forwardPaths) > 0 && len(backwardPaths) > 0 && len(forwardPaths) == len(backwardPaths)

	// Générer les insights
	insights := s.generateDiracInsights(forwardPaths, backwardPaths, req.Bidirectional, nodeMap[sourceID].Label, nodeMap[targetID].Label)

	return &DiracResult{
		Query:         req.Query,
		Source:        sourceID,
		SourceLabel:   nodeMap[sourceID].Label,
		Target:        targetID,
		TargetLabel:   nodeMap[targetID].Label,
		ForwardPaths:  forwardPaths,
		BackwardPaths: backwardPaths,
		TotalPaths:    len(forwardPaths) + len(backwardPaths),
		Symmetric:     symmetric,
		Insights:      insights,
	}, nil
}

// findAllPathsBFS trouve tous les chemins entre deux nœuds via BFS
func (s *SearchService) findAllPathsBFS(from, to string, adj map[string][]models.GraphEdge, nodeMap map[string]models.GraphNode, maxDepth, maxPaths int) []ConePath {
	var paths []ConePath

	type pathState struct {
		current string
		path    []string
		edges   []string
	}

	queue := []pathState{{current: from, path: []string{from}, edges: nil}}

	for len(queue) > 0 && len(paths) < maxPaths {
		state := queue[0]
		queue = queue[1:]

		if len(state.path) > maxDepth+1 {
			continue
		}

		if state.current == to && len(state.path) > 1 {
			labels := make([]string, len(state.path))
			for i, nid := range state.path {
				labels[i] = nodeMap[nid].Label
			}
			paths = append(paths, ConePath{
				Nodes:  state.path,
				Labels: labels,
				Edges:  state.edges,
				Length: len(state.path),
			})
			continue
		}

		for _, edge := range adj[state.current] {
			// Éviter les cycles
			inPath := false
			for _, n := range state.path {
				if n == edge.To {
					inPath = true
					break
				}
			}
			if inPath {
				continue
			}

			newPath := append([]string{}, state.path...)
			newPath = append(newPath, edge.To)
			newEdges := append([]string{}, state.edges...)
			newEdges = append(newEdges, edge.Label)

			queue = append(queue, pathState{
				current: edge.To,
				path:    newPath,
				edges:   newEdges,
			})
		}
	}

	// Trier par longueur
	sort.Slice(paths, func(i, j int) bool {
		return paths[i].Length < paths[j].Length
	})

	return paths
}

// generateDiracInsights génère des insights pour la recherche Dirac
func (s *SearchService) generateDiracInsights(forward, backward []ConePath, bidirectional bool, sourceLabel, targetLabel string) []string {
	var insights []string

	if len(forward) == 0 && len(backward) == 0 {
		insights = append(insights, fmt.Sprintf("Aucun chemin trouvé entre '%s' et '%s'.", sourceLabel, targetLabel))
		return insights
	}

	if len(forward) > 0 {
		insights = append(insights, fmt.Sprintf("%d chemin(s) de '%s' vers '%s'.", len(forward), sourceLabel, targetLabel))
		if len(forward) > 0 {
			insights = append(insights, fmt.Sprintf("Chemin le plus court (avant): %d nœuds.", forward[0].Length))
		}
	}

	if bidirectional && len(backward) > 0 {
		insights = append(insights, fmt.Sprintf("%d chemin(s) de '%s' vers '%s' (inverse).", len(backward), targetLabel, sourceLabel))
		if len(backward) > 0 {
			insights = append(insights, fmt.Sprintf("Chemin le plus court (arrière): %d nœuds.", backward[0].Length))
		}
	}

	if len(forward) > 0 && len(backward) > 0 {
		if forward[0].Length == backward[0].Length {
			insights = append(insights, "Les chemins avant et arrière ont la même longueur minimale - relation potentiellement symétrique.")
		} else if forward[0].Length < backward[0].Length {
			insights = append(insights, fmt.Sprintf("Le chemin avant est plus court (%d vs %d) - direction causale possible.", forward[0].Length, backward[0].Length))
		} else {
			insights = append(insights, fmt.Sprintf("Le chemin arrière est plus court (%d vs %d) - direction causale inverse possible.", backward[0].Length, forward[0].Length))
		}
	}

	return insights
}

// ============================================
// Orbits - Structured Neighborhood Analysis
// Concentric analysis of node neighborhoods
// ============================================

// OrbitRequest représente une requête d'analyse d'orbites
type OrbitRequest struct {
	CaseID   string `json:"case_id"`
	NodeID   string `json:"node_id"`
	MaxLevel int    `json:"max_level"` // Nombre de niveaux d'orbite (1-5)
}

// OrbitLevel représente un niveau d'orbite
type OrbitLevel struct {
	Level         int            `json:"level"`
	Nodes         []OrbitNode    `json:"nodes"`
	Count         int            `json:"count"`
	Density       float64        `json:"density"`        // Connexions internes / possibles
	TypeBreakdown map[string]int `json:"type_breakdown"` // Répartition par type de nœud
	EdgeTypes     map[string]int `json:"edge_types"`     // Types d'arêtes vers ce niveau
}

// OrbitNode représente un nœud dans une orbite
type OrbitNode struct {
	NodeID      string   `json:"node_id"`
	NodeLabel   string   `json:"node_label"`
	NodeType    string   `json:"node_type"`
	Connections int      `json:"connections"` // Connexions vers le niveau précédent
	EdgeLabels  []string `json:"edge_labels"` // Labels des arêtes de connexion
}

// OrbitResult représente le résultat de l'analyse d'orbites
type OrbitResult struct {
	CenterNode  string       `json:"center_node"`
	CenterLabel string       `json:"center_label"`
	CenterType  string       `json:"center_type"`
	Orbits      []OrbitLevel `json:"orbits"`
	TotalNodes  int          `json:"total_nodes"`
	MaxReached  int          `json:"max_reached"` // Niveau max atteint
	Insights    []string     `json:"insights"`
}

// AnalyzeOrbits analyse les orbites concentriques autour d'un nœud
func (s *SearchService) AnalyzeOrbits(graph *models.GraphData, req OrbitRequest) (*OrbitResult, error) {
	if req.MaxLevel <= 0 || req.MaxLevel > 5 {
		req.MaxLevel = 3
	}

	// Créer les structures
	nodeMap := make(map[string]models.GraphNode)
	for _, node := range graph.Nodes {
		nodeMap[node.ID] = node
	}

	if _, exists := nodeMap[req.NodeID]; !exists {
		return nil, fmt.Errorf("nœud '%s' non trouvé", req.NodeID)
	}

	// Construire l'adjacence bidirectionnelle avec les labels d'arêtes
	type neighborInfo struct {
		nodeID    string
		edgeLabel string
	}
	adj := make(map[string][]neighborInfo)
	for _, edge := range graph.Edges {
		adj[edge.From] = append(adj[edge.From], neighborInfo{edge.To, edge.Label})
		adj[edge.To] = append(adj[edge.To], neighborInfo{edge.From, edge.Label})
	}

	// BFS pour construire les orbites
	visited := make(map[string]int) // nodeID -> level
	visited[req.NodeID] = 0

	orbits := make([]OrbitLevel, 0)
	currentLevel := []string{req.NodeID}
	totalNodes := 0

	for level := 1; level <= req.MaxLevel && len(currentLevel) > 0; level++ {
		nextLevel := make(map[string]*OrbitNode)
		edgeTypes := make(map[string]int)

		for _, nodeID := range currentLevel {
			for _, neighbor := range adj[nodeID] {
				if _, seen := visited[neighbor.nodeID]; !seen {
					visited[neighbor.nodeID] = level

					if existing, ok := nextLevel[neighbor.nodeID]; ok {
						existing.Connections++
						existing.EdgeLabels = append(existing.EdgeLabels, neighbor.edgeLabel)
					} else {
						node := nodeMap[neighbor.nodeID]
						nextLevel[neighbor.nodeID] = &OrbitNode{
							NodeID:      neighbor.nodeID,
							NodeLabel:   node.Label,
							NodeType:    node.Type,
							Connections: 1,
							EdgeLabels:  []string{neighbor.edgeLabel},
						}
					}
					edgeTypes[neighbor.edgeLabel]++
				}
			}
		}

		if len(nextLevel) == 0 {
			break
		}

		// Convertir en slice et calculer les statistiques
		nodes := make([]OrbitNode, 0, len(nextLevel))
		typeBreakdown := make(map[string]int)
		for _, node := range nextLevel {
			nodes = append(nodes, *node)
			typeBreakdown[node.NodeType]++
		}

		// Calculer la densité interne
		internalEdges := 0
		nodeIDs := make(map[string]bool)
		for id := range nextLevel {
			nodeIDs[id] = true
		}
		for id := range nextLevel {
			for _, neighbor := range adj[id] {
				if nodeIDs[neighbor.nodeID] {
					internalEdges++
				}
			}
		}
		possibleEdges := len(nodes) * (len(nodes) - 1)
		density := 0.0
		if possibleEdges > 0 {
			density = float64(internalEdges) / float64(possibleEdges)
		}

		// Trier par nombre de connexions
		sort.Slice(nodes, func(i, j int) bool {
			return nodes[i].Connections > nodes[j].Connections
		})

		orbits = append(orbits, OrbitLevel{
			Level:         level,
			Nodes:         nodes,
			Count:         len(nodes),
			Density:       density,
			TypeBreakdown: typeBreakdown,
			EdgeTypes:     edgeTypes,
		})

		totalNodes += len(nodes)

		// Préparer le niveau suivant
		currentLevel = make([]string, 0, len(nextLevel))
		for id := range nextLevel {
			currentLevel = append(currentLevel, id)
		}
	}

	// Générer les insights
	centerNode := nodeMap[req.NodeID]
	insights := s.generateOrbitInsights(orbits, centerNode.Label, totalNodes)

	return &OrbitResult{
		CenterNode:  req.NodeID,
		CenterLabel: centerNode.Label,
		CenterType:  centerNode.Type,
		Orbits:      orbits,
		TotalNodes:  totalNodes,
		MaxReached:  len(orbits),
		Insights:    insights,
	}, nil
}

// generateOrbitInsights génère des insights sur l'analyse d'orbites
func (s *SearchService) generateOrbitInsights(orbits []OrbitLevel, centerLabel string, totalNodes int) []string {
	var insights []string

	if len(orbits) == 0 {
		insights = append(insights, fmt.Sprintf("'%s' est un nœud isolé sans connexions.", centerLabel))
		return insights
	}

	insights = append(insights, fmt.Sprintf("Analyse de %d niveau(x) d'orbite autour de '%s'.", len(orbits), centerLabel))
	insights = append(insights, fmt.Sprintf("%d nœuds au total dans l'orbite.", totalNodes))

	// Analyser le premier niveau (connexions directes)
	if len(orbits) > 0 {
		first := orbits[0]
		insights = append(insights, fmt.Sprintf("Niveau 1 (direct): %d connexion(s).", first.Count))

		// Type dominant
		if len(first.TypeBreakdown) > 0 {
			maxType := ""
			maxCount := 0
			for t, c := range first.TypeBreakdown {
				if c > maxCount {
					maxCount = c
					maxType = t
				}
			}
			if maxType != "" {
				insights = append(insights, fmt.Sprintf("Type dominant au niveau 1: '%s' (%d nœuds).", maxType, maxCount))
			}
		}
	}

	// Analyser la décroissance
	if len(orbits) >= 2 {
		growth := float64(orbits[1].Count) / float64(orbits[0].Count)
		if growth > 2 {
			insights = append(insights, "Expansion rapide: le réseau s'élargit fortement au niveau 2.")
		} else if growth < 0.5 {
			insights = append(insights, "Contraction: le réseau se rétrécit au niveau 2 (structure locale dense).")
		}
	}

	// Analyser la densité
	for _, orbit := range orbits {
		if orbit.Density > 0.5 {
			insights = append(insights, fmt.Sprintf("Niveau %d très dense (%.0f%%) - cluster potentiel.", orbit.Level, orbit.Density*100))
		}
	}

	return insights
}
