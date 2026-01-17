package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"forensicinvestigator/internal/models"
	"forensicinvestigator/internal/services"
)

// Handler gère les requêtes HTTP
type Handler struct {
	ollama        *services.OllamaService
	cases         *services.CaseService
	n4l           *services.N4LService
	n4lGenerator  *services.N4LGeneratorService  // Service de génération N4L (UI -> N4L)
	hrm           *services.HRMService           // Service HRM externe (sapientinc/HRM + Ollama)
	search        *services.SearchService
	graphAnalyzer *services.GraphAnalyzerService // Service d'analyse de graphe
	notebook      *services.NotebookService      // Service de gestion des notebooks
	scenario      *services.ScenarioService      // Service de simulation What-If
	anomaly       *services.AnomalyService       // Service de détection d'anomalies
}

// NewHandler crée un nouveau handler
func NewHandler(ollama *services.OllamaService, cases *services.CaseService, n4l *services.N4LService) *Handler {
	h := &Handler{
		ollama:        ollama,
		cases:         cases,
		n4l:           n4l,
		n4lGenerator:  services.NewN4LGeneratorService(n4l),
		hrm:           services.NewHRMService("http://localhost:8081"),
		search:        services.NewSearchService("http://localhost:11434", "nomic-embed-text"),
		graphAnalyzer: services.NewGraphAnalyzerService(),
		notebook:      services.NewNotebookService(),
	}
	// Initialiser les nouveaux services
	h.scenario = services.NewScenarioService(cases, ollama)
	h.anomaly = services.NewAnomalyService(cases, ollama)
	return h
}

// HandleCases gère les opérations sur les affaires
func (h *Handler) HandleCases(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		cases := h.cases.GetAllCases()
		json.NewEncoder(w).Encode(cases)

	case http.MethodPost:
		var req struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			Type        string `json:"type"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		c, err := h.cases.CreateCase(req.Name, req.Description, req.Type)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(c)

	default:
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
	}
}

// HandleCase gère une affaire spécifique
func (h *Handler) HandleCase(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extraire l'ID de l'URL
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		http.Error(w, "ID manquant", http.StatusBadRequest)
		return
	}
	caseID := parts[len(parts)-1]

	switch r.Method {
	case http.MethodGet:
		c, err := h.cases.GetCase(caseID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(c)

	case http.MethodPut:
		var c models.Case
		if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		c.ID = caseID
		if err := h.cases.UpdateCase(&c); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(c)

	case http.MethodDelete:
		if err := h.cases.DeleteCase(caseID); err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
	}
}

// HandleEntities gère les entités
func (h *Handler) HandleEntities(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	caseID := r.URL.Query().Get("case_id")
	if caseID == "" {
		http.Error(w, "case_id requis", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		entities, err := h.cases.GetEntities(caseID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(entities)

	case http.MethodPost:
		var entity models.Entity
		if err := json.NewDecoder(r.Body).Decode(&entity); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		result, err := h.cases.AddEntity(caseID, entity)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(result)

	default:
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
	}
}

// HandleRelations gère les relations entre entités
func (h *Handler) HandleRelations(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	caseID := r.URL.Query().Get("case_id")
	if caseID == "" {
		http.Error(w, "case_id requis", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodPost:
		var relation models.Relation
		if err := json.NewDecoder(r.Body).Decode(&relation); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := h.cases.AddRelation(caseID, relation); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})

	default:
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
	}
}

// HandleEvidence gère les preuves
func (h *Handler) HandleEvidence(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	caseID := r.URL.Query().Get("case_id")
	if caseID == "" {
		http.Error(w, "case_id requis", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		evidence, err := h.cases.GetEvidence(caseID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(evidence)

	case http.MethodPost:
		var ev models.Evidence
		if err := json.NewDecoder(r.Body).Decode(&ev); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		result, err := h.cases.AddEvidence(caseID, ev)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(result)

	default:
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
	}
}

// HandleTimeline gère la chronologie
func (h *Handler) HandleTimeline(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	caseID := r.URL.Query().Get("case_id")
	if caseID == "" {
		http.Error(w, "case_id requis", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		timeline, err := h.cases.GetTimeline(caseID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(timeline)

	case http.MethodPost:
		var event models.Event
		if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		result, err := h.cases.AddEvent(caseID, event)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(result)

	default:
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
	}
}

// HandleHypotheses gère les hypothèses
func (h *Handler) HandleHypotheses(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	caseID := r.URL.Query().Get("case_id")
	if caseID == "" {
		http.Error(w, "case_id requis", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		hypotheses, err := h.cases.GetHypotheses(caseID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(hypotheses)

	case http.MethodPost:
		var hyp models.Hypothesis
		if err := json.NewDecoder(r.Body).Decode(&hyp); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		result, err := h.cases.AddHypothesis(caseID, hyp)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(result)

	default:
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
	}
}

// HandleAnalyze gère les analyses IA
func (h *Handler) HandleAnalyze(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req models.AnalysisRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Si un case_id est fourni, récupérer les données
	if req.CaseID != "" {
		c, err := h.cases.GetCase(req.CaseID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		analysis, err := h.ollama.AnalyzeCase(*c)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(models.AnalysisResponse{
			Analysis: analysis,
		})
		return
	}

	// Sinon, analyser le graphe fourni
	hypotheses, err := h.ollama.GenerateHypotheses(req.GraphData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"hypotheses": hypotheses,
	})
}

// HandleContradictions détecte les contradictions
func (h *Handler) HandleContradictions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req struct {
		GraphData models.GraphData `json:"graph_data"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	analysis, err := h.ollama.DetectContradictions(req.GraphData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"analysis": analysis,
	})
}

// HandleQuestions génère des questions d'investigation
func (h *Handler) HandleQuestions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req struct {
		GraphData models.GraphData `json:"graph_data"`
		Context   string           `json:"context"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	questions, err := h.ollama.GenerateQuestions(req.GraphData, req.Context)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"questions": questions,
	})
}

// HandleN4LParse parse un fichier N4L et retourne les données forensiques complètes
func (h *Handler) HandleN4LParse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Support ancien format (body direct) et nouveau format (JSON)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var content string
	var caseID string

	// Essayer de parser comme JSON d'abord
	var req struct {
		Content string `json:"content"`
		CaseID  string `json:"case_id"`
	}
	if err := json.Unmarshal(body, &req); err == nil {
		content = req.Content
		caseID = req.CaseID
	} else {
		// Fallback: contenu brut
		content = string(body)
		caseID = r.URL.Query().Get("case_id")
	}

	// Si content est vide mais case_id fourni, charger le N4L du cas
	if content == "" && caseID != "" {
		c, err := h.cases.GetCase(caseID)
		if err == nil {
			if c.N4LContent != "" {
				content = c.N4LContent
			} else {
				content = h.n4lGenerator.ExportCaseToN4L(c)
			}
		}
	}

	// Parser avec extraction forensique complète
	result := h.n4l.ParseForensicN4L(content, caseID)
	json.NewEncoder(w).Encode(result)
}

// HandleN4LExport exporte une affaire en N4L
func (h *Handler) HandleN4LExport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	caseID := r.URL.Query().Get("case_id")
	if caseID == "" {
		http.Error(w, "case_id requis", http.StatusBadRequest)
		return
	}

	format := r.URL.Query().Get("format") // "text" ou "json"

	c, err := h.cases.GetCase(caseID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Si le cas a déjà du contenu N4L, le retourner
	var n4lContent string
	if c.N4LContent != "" {
		n4lContent = c.N4LContent
	} else {
		// Générer le contenu N4L avec le nouveau générateur
		n4lContent = h.n4lGenerator.ExportCaseToN4L(c)
	}

	if format == "json" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"n4l_content": n4lContent,
		})
	} else {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(n4lContent))
	}
}

// HandleGraph construit et retourne le graphe d'une affaire
// Utilise toujours le N4L comme source unique (généré à la volée si nécessaire)
func (h *Handler) HandleGraph(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	caseID := r.URL.Query().Get("case_id")
	if caseID == "" {
		http.Error(w, "case_id requis", http.StatusBadRequest)
		return
	}

	caseData, err := h.cases.GetCase(caseID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Obtenir le contenu N4L (existant ou généré à la volée)
	var n4lContent string
	if caseData.N4LContent != "" {
		n4lContent = caseData.N4LContent
	} else {
		// Générer le N4L à partir des données du cas
		n4lContent = h.n4lGenerator.ExportCaseToN4L(caseData)
	}

	// Parser le N4L et retourner le graphe
	parsed := h.n4l.ParseForensicN4L(n4lContent, caseID)
	json.NewEncoder(w).Encode(parsed.Graph)
}

// HandleLoadDemo charge les données de démonstration
func (h *Handler) HandleLoadDemo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Cette fonction sera appelée depuis main.go avec les données de démo
	json.NewEncoder(w).Encode(map[string]string{
		"status": "Demo data must be loaded at startup",
	})
}

// GetCaseService retourne le service des affaires (pour le chargement démo)
func (h *Handler) GetCaseService() *services.CaseService {
	return h.cases
}

// HandleChat gère les requêtes de chat conversationnel avec recherche hybride
func (h *Handler) HandleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req struct {
		CaseID  string `json:"case_id"`
		Message string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Récupérer les données de l'affaire si disponible
	var caseContext string
	if req.CaseID != "" {
		c, err := h.cases.GetCase(req.CaseID)
		if err == nil {
			// Construire le contexte enrichi avec recherche hybride
			caseContext = h.buildEnrichedContext(c, req.Message)
		}
	}

	// Générer la réponse via Ollama
	response, err := h.ollama.Chat(req.Message, caseContext)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"response": response,
	})
}

// HandleChatStream gère les requêtes de chat en streaming (SSE)
func (h *Handler) HandleChatStream(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	// Configurer les headers SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// S'assurer que le ResponseWriter supporte le flush
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming non supporté", http.StatusInternalServerError)
		return
	}

	var req struct {
		CaseID  string `json:"case_id"`
		Message string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		fmt.Fprintf(w, "data: {\"error\": \"%s\"}\n\n", err.Error())
		flusher.Flush()
		return
	}

	// Récupérer les données de l'affaire si disponible
	var caseContext string
	if req.CaseID != "" {
		c, err := h.cases.GetCase(req.CaseID)
		if err == nil {
			caseContext = h.buildEnrichedContext(c, req.Message)
		}
	}

	// Streamer la réponse via Ollama
	err := h.ollama.ChatStream(req.Message, caseContext, func(chunk string, done bool) error {
		// Encoder le chunk en JSON pour éviter les problèmes d'échappement
		chunkJSON, _ := json.Marshal(map[string]interface{}{
			"chunk": chunk,
			"done":  done,
		})
		fmt.Fprintf(w, "data: %s\n\n", chunkJSON)
		flusher.Flush()
		return nil
	})

	if err != nil {
		errorJSON, _ := json.Marshal(map[string]string{"error": err.Error()})
		fmt.Fprintf(w, "data: %s\n\n", errorJSON)
		flusher.Flush()
	}
}

// HandleAnalyzeStream gère les analyses IA en streaming
func (h *Handler) HandleAnalyzeStream(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	// Configurer les headers SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming non supporté", http.StatusInternalServerError)
		return
	}

	var req models.AnalysisRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		fmt.Fprintf(w, "data: {\"error\": \"%s\"}\n\n", err.Error())
		flusher.Flush()
		return
	}

	// Si un case_id est fourni, analyser l'affaire
	if req.CaseID != "" {
		c, err := h.cases.GetCase(req.CaseID)
		if err != nil {
			fmt.Fprintf(w, "data: {\"error\": \"Affaire non trouvée\"}\n\n", )
			flusher.Flush()
			return
		}

		err = h.ollama.AnalyzeCaseStream(*c, func(chunk string, done bool) error {
			chunkJSON, _ := json.Marshal(map[string]interface{}{
				"chunk": chunk,
				"done":  done,
			})
			fmt.Fprintf(w, "data: %s\n\n", chunkJSON)
			flusher.Flush()
			return nil
		})

		if err != nil {
			errorJSON, _ := json.Marshal(map[string]string{"error": err.Error()})
			fmt.Fprintf(w, "data: %s\n\n", errorJSON)
			flusher.Flush()
		}
		return
	}

	// Sinon, erreur
	fmt.Fprintf(w, "data: {\"error\": \"case_id requis\"}\n\n")
	flusher.Flush()
}

// HandleContradictionsStream détecte les contradictions en streaming
func (h *Handler) HandleContradictionsStream(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming non supporté", http.StatusInternalServerError)
		return
	}

	var req struct {
		GraphData models.GraphData `json:"graph_data"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		fmt.Fprintf(w, "data: {\"error\": \"%s\"}\n\n", err.Error())
		flusher.Flush()
		return
	}

	err := h.ollama.DetectContradictionsStream(req.GraphData, func(chunk string, done bool) error {
		chunkJSON, _ := json.Marshal(map[string]interface{}{
			"chunk": chunk,
			"done":  done,
		})
		fmt.Fprintf(w, "data: %s\n\n", chunkJSON)
		flusher.Flush()
		return nil
	})

	if err != nil {
		errorJSON, _ := json.Marshal(map[string]string{"error": err.Error()})
		fmt.Fprintf(w, "data: %s\n\n", errorJSON)
		flusher.Flush()
	}
}

// HandleContradictionsDetectStream détecte les contradictions en streaming (accepte case_id)
func (h *Handler) HandleContradictionsDetectStream(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming non supporté", http.StatusInternalServerError)
		return
	}

	var req struct {
		CaseID string `json:"case_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		fmt.Fprintf(w, "data: {\"error\": \"%s\"}\n\n", err.Error())
		flusher.Flush()
		return
	}

	if req.CaseID == "" {
		fmt.Fprintf(w, "data: {\"error\": \"case_id requis\"}\n\n")
		flusher.Flush()
		return
	}

	graphData, err := h.cases.BuildGraphData(req.CaseID)
	if err != nil {
		fmt.Fprintf(w, "data: {\"error\": \"Affaire non trouvée\"}\n\n")
		flusher.Flush()
		return
	}

	err = h.ollama.DetectContradictionsStream(*graphData, func(chunk string, done bool) error {
		chunkJSON, _ := json.Marshal(map[string]interface{}{
			"chunk": chunk,
			"done":  done,
		})
		fmt.Fprintf(w, "data: %s\n\n", chunkJSON)
		flusher.Flush()
		return nil
	})

	if err != nil {
		errorJSON, _ := json.Marshal(map[string]string{"error": err.Error()})
		fmt.Fprintf(w, "data: %s\n\n", errorJSON)
		flusher.Flush()
	}
}

// HandleHypothesesGenerateStream génère des hypothèses en streaming
func (h *Handler) HandleHypothesesGenerateStream(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming non supporté", http.StatusInternalServerError)
		return
	}

	var req struct {
		CaseID string `json:"case_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		fmt.Fprintf(w, "data: {\"error\": \"%s\"}\n\n", err.Error())
		flusher.Flush()
		return
	}

	if req.CaseID == "" {
		fmt.Fprintf(w, "data: {\"error\": \"case_id requis\"}\n\n")
		flusher.Flush()
		return
	}

	graphData, err := h.cases.BuildGraphData(req.CaseID)
	if err != nil {
		fmt.Fprintf(w, "data: {\"error\": \"Affaire non trouvée\"}\n\n")
		flusher.Flush()
		return
	}

	err = h.ollama.GenerateHypothesesStream(*graphData, func(chunk string, done bool) error {
		chunkJSON, _ := json.Marshal(map[string]interface{}{
			"chunk": chunk,
			"done":  done,
		})
		fmt.Fprintf(w, "data: %s\n\n", chunkJSON)
		flusher.Flush()
		return nil
	})

	if err != nil {
		errorJSON, _ := json.Marshal(map[string]string{"error": err.Error()})
		fmt.Fprintf(w, "data: %s\n\n", errorJSON)
		flusher.Flush()
	}
}

// HandleQuestionsGenerateStream génère des questions d'investigation en streaming
func (h *Handler) HandleQuestionsGenerateStream(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming non supporté", http.StatusInternalServerError)
		return
	}

	var req struct {
		CaseID  string `json:"case_id"`
		Context string `json:"context"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		fmt.Fprintf(w, "data: {\"error\": \"%s\"}\n\n", err.Error())
		flusher.Flush()
		return
	}

	if req.CaseID == "" {
		fmt.Fprintf(w, "data: {\"error\": \"case_id requis\"}\n\n")
		flusher.Flush()
		return
	}

	graphData, err := h.cases.BuildGraphData(req.CaseID)
	if err != nil {
		fmt.Fprintf(w, "data: {\"error\": \"Affaire non trouvée\"}\n\n")
		flusher.Flush()
		return
	}

	err = h.ollama.GenerateQuestionsStream(*graphData, req.Context, func(chunk string, done bool) error {
		chunkJSON, _ := json.Marshal(map[string]interface{}{
			"chunk": chunk,
			"done":  done,
		})
		fmt.Fprintf(w, "data: %s\n\n", chunkJSON)
		flusher.Flush()
		return nil
	})

	if err != nil {
		errorJSON, _ := json.Marshal(map[string]string{"error": err.Error()})
		fmt.Fprintf(w, "data: %s\n\n", errorJSON)
		flusher.Flush()
	}
}

// buildEnrichedContext construit un contexte enrichi avec recherche hybride, relations et N4L
func (h *Handler) buildEnrichedContext(c *models.Case, query string) string {
	var sb strings.Builder

	// === Section 1: Informations générales de l'affaire ===
	sb.WriteString("=== AFFAIRE ===\n")
	sb.WriteString("Nom: " + c.Name + "\n")
	sb.WriteString("Type: " + c.Type + "\n")
	sb.WriteString("Statut: " + c.Status + "\n")
	sb.WriteString("Description: " + c.Description + "\n\n")

	// === Section 2: Recherche hybride - éléments pertinents à la question ===
	searchReq := services.SearchRequest{
		Query:      query,
		CaseID:     c.ID,
		Limit:      10,
		BM25Weight: 0.4, // 40% BM25, 60% sémantique pour privilégier le sens
	}

	results, err := h.search.HybridSearch(c, searchReq)
	if err == nil && len(results) > 0 {
		sb.WriteString("=== ÉLÉMENTS PERTINENTS (recherche sémantique) ===\n")
		sb.WriteString("Les éléments suivants sont les plus pertinents par rapport à la question:\n\n")

		for i, result := range results {
			if result.Score > 0.1 { // Seuil de pertinence
				sb.WriteString(fmt.Sprintf("%d. [%s] %s (score: %.2f)\n", i+1, result.Type, result.Name, result.Score))
				if result.Description != "" {
					sb.WriteString("   Description: " + result.Description + "\n")
				}
			}
		}
		sb.WriteString("\n")
	}

	// === Section 3: Toutes les entités avec leurs relations ===
	if len(c.Entities) > 0 {
		sb.WriteString("=== ENTITÉS ET RELATIONS ===\n")

		// Créer une map pour résoudre les IDs en noms
		entityMap := make(map[string]string)
		for _, e := range c.Entities {
			entityMap[e.ID] = e.Name
		}

		for _, e := range c.Entities {
			sb.WriteString(fmt.Sprintf("\n• %s (%s, %s)\n", e.Name, string(e.Type), string(e.Role)))
			if e.Description != "" {
				sb.WriteString("  Description: " + e.Description + "\n")
			}

			// Ajouter les attributs
			if len(e.Attributes) > 0 {
				sb.WriteString("  Attributs: ")
				attrs := []string{}
				for k, v := range e.Attributes {
					attrs = append(attrs, k+"="+v)
				}
				sb.WriteString(strings.Join(attrs, ", ") + "\n")
			}

			// Ajouter les relations de cette entité
			if len(e.Relations) > 0 {
				sb.WriteString("  Relations:\n")
				for _, rel := range e.Relations {
					targetName := entityMap[rel.ToID]
					if targetName == "" {
						targetName = rel.ToID
					}
					relType := rel.Type
					if rel.Label != "" {
						relType = rel.Label
					}
					sb.WriteString(fmt.Sprintf("    → %s: %s\n", relType, targetName))
					if rel.Context != "" {
						sb.WriteString(fmt.Sprintf("      (contexte: %s)\n", rel.Context))
					}
				}
			}
		}
		sb.WriteString("\n")
	}

	// === Section 4: Preuves avec liens ===
	if len(c.Evidence) > 0 {
		sb.WriteString("=== PREUVES ===\n")
		for _, ev := range c.Evidence {
			sb.WriteString(fmt.Sprintf("• %s (%s) - Fiabilité: %d/10\n", ev.Name, string(ev.Type), ev.Reliability))
			if ev.Description != "" {
				sb.WriteString("  Description: " + ev.Description + "\n")
			}
			if ev.Location != "" {
				sb.WriteString("  Localisation: " + ev.Location + "\n")
			}
			// Liens avec entités
			if len(ev.LinkedEntities) > 0 {
				linkedNames := []string{}
				for _, id := range ev.LinkedEntities {
					for _, e := range c.Entities {
						if e.ID == id {
							linkedNames = append(linkedNames, e.Name)
							break
						}
					}
				}
				if len(linkedNames) > 0 {
					sb.WriteString("  Entités liées: " + strings.Join(linkedNames, ", ") + "\n")
				}
			}
		}
		sb.WriteString("\n")
	}

	// === Section 5: Chronologie ===
	if len(c.Timeline) > 0 {
		sb.WriteString("=== CHRONOLOGIE ===\n")
		for _, t := range c.Timeline {
			sb.WriteString(fmt.Sprintf("• %s: %s\n", t.Timestamp.Format("02/01/2006 15:04"), t.Title))
			if t.Description != "" {
				sb.WriteString("  " + t.Description + "\n")
			}
			if t.Location != "" {
				sb.WriteString("  Lieu: " + t.Location + "\n")
			}
		}
		sb.WriteString("\n")
	}

	// === Section 6: Hypothèses en cours ===
	if len(c.Hypotheses) > 0 {
		sb.WriteString("=== HYPOTHÈSES D'INVESTIGATION ===\n")
		for _, hyp := range c.Hypotheses {
			sb.WriteString(fmt.Sprintf("• %s (confiance: %d%%, statut: %s)\n", hyp.Title, hyp.ConfidenceLevel, string(hyp.Status)))
			if hyp.Description != "" {
				sb.WriteString("  " + hyp.Description + "\n")
			}
		}
		sb.WriteString("\n")
	}

	// === Section 7: Format N4L (représentation sémantique) ===
	n4lContent := h.n4l.ExportToN4L(c)
	if n4lContent != "" {
		sb.WriteString("=== REPRÉSENTATION N4L (Notes for Linking) ===\n")
		sb.WriteString("Format sémantique des relations:\n")
		sb.WriteString(n4lContent)
		sb.WriteString("\n")
	}

	return sb.String()
}

// buildCaseContext construit le contexte de l'affaire pour le LLM (version simple)
func (h *Handler) buildCaseContext(c *models.Case) string {
	var sb strings.Builder
	sb.WriteString("Contexte de l'affaire:\n")
	sb.WriteString("Nom: " + c.Name + "\n")
	sb.WriteString("Type: " + c.Type + "\n")
	sb.WriteString("Description: " + c.Description + "\n\n")

	if len(c.Entities) > 0 {
		sb.WriteString("Entités:\n")
		for _, e := range c.Entities {
			sb.WriteString("- " + e.Name + " (" + string(e.Type) + ", " + string(e.Role) + "): " + e.Description + "\n")
		}
	}

	if len(c.Evidence) > 0 {
		sb.WriteString("\nPreuves:\n")
		for _, ev := range c.Evidence {
			sb.WriteString("- " + ev.Name + " (" + string(ev.Type) + "): " + ev.Description + "\n")
		}
	}

	if len(c.Timeline) > 0 {
		sb.WriteString("\nChronologie:\n")
		for _, t := range c.Timeline {
			sb.WriteString("- " + t.Timestamp.Format("02/01/2006 15:04") + ": " + t.Title + "\n")
		}
	}

	return sb.String()
}

// HandleDeleteEntity supprime une entité
func (h *Handler) HandleDeleteEntity(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	caseID := r.URL.Query().Get("case_id")
	entityID := r.URL.Query().Get("entity_id")

	if caseID == "" || entityID == "" {
		http.Error(w, "case_id et entity_id requis", http.StatusBadRequest)
		return
	}

	if err := h.cases.DeleteEntity(caseID, entityID); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandleUpdateEntity met à jour une entité existante
func (h *Handler) HandleUpdateEntity(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	caseID := r.URL.Query().Get("case_id")
	if caseID == "" {
		http.Error(w, "case_id requis", http.StatusBadRequest)
		return
	}

	var entity models.Entity
	if err := json.NewDecoder(r.Body).Decode(&entity); err != nil {
		http.Error(w, "JSON invalide", http.StatusBadRequest)
		return
	}

	if entity.ID == "" {
		http.Error(w, "entity ID requis", http.StatusBadRequest)
		return
	}

	if err := h.cases.UpdateEntity(caseID, entity); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

// HandleDeleteEvidence supprime une preuve
func (h *Handler) HandleDeleteEvidence(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	caseID := r.URL.Query().Get("case_id")
	evidenceID := r.URL.Query().Get("evidence_id")

	if caseID == "" || evidenceID == "" {
		http.Error(w, "case_id et evidence_id requis", http.StatusBadRequest)
		return
	}

	if err := h.cases.DeleteEvidence(caseID, evidenceID); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandleUpdateEvidence met à jour une preuve existante
func (h *Handler) HandleUpdateEvidence(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	caseID := r.URL.Query().Get("case_id")
	if caseID == "" {
		http.Error(w, "case_id requis", http.StatusBadRequest)
		return
	}

	var evidence models.Evidence
	if err := json.NewDecoder(r.Body).Decode(&evidence); err != nil {
		http.Error(w, "JSON invalide", http.StatusBadRequest)
		return
	}

	if evidence.ID == "" {
		http.Error(w, "evidence ID requis", http.StatusBadRequest)
		return
	}

	if err := h.cases.UpdateEvidence(caseID, evidence); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

// HandleUpdateEvent met à jour un événement existant
func (h *Handler) HandleUpdateEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	caseID := r.URL.Query().Get("case_id")
	eventID := r.URL.Query().Get("event_id")

	if caseID == "" || eventID == "" {
		http.Error(w, "case_id et event_id requis", http.StatusBadRequest)
		return
	}

	var event models.Event
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	event.ID = eventID

	if err := h.cases.UpdateEvent(caseID, event); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

// HandleDeleteEvent supprime un événement
func (h *Handler) HandleDeleteEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	caseID := r.URL.Query().Get("case_id")
	eventID := r.URL.Query().Get("event_id")

	if caseID == "" || eventID == "" {
		http.Error(w, "case_id et event_id requis", http.StatusBadRequest)
		return
	}

	if err := h.cases.DeleteEvent(caseID, eventID); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandleUpdateHypothesis met à jour une hypothèse existante
func (h *Handler) HandleUpdateHypothesis(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	// Support deux formats:
	// 1. case_id en query param + hypothesis directement dans le body
	// 2. { case_id, hypothesis } dans le body (format frontend HRM)
	caseID := r.URL.Query().Get("case_id")

	var hypothesis models.Hypothesis

	// Essayer d'abord le format avec wrapper { case_id, hypothesis }
	var wrapper struct {
		CaseID     string           `json:"case_id"`
		Hypothesis models.Hypothesis `json:"hypothesis"`
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Erreur lecture body", http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(body, &wrapper); err == nil && wrapper.Hypothesis.ID != "" {
		// Format wrapper détecté
		hypothesis = wrapper.Hypothesis
		if caseID == "" {
			caseID = wrapper.CaseID
		}
	} else {
		// Format direct: hypothesis dans le body
		if err := json.Unmarshal(body, &hypothesis); err != nil {
			http.Error(w, "JSON invalide", http.StatusBadRequest)
			return
		}
	}

	if caseID == "" {
		http.Error(w, "case_id requis", http.StatusBadRequest)
		return
	}

	if hypothesis.ID == "" {
		http.Error(w, "hypothesis ID requis", http.StatusBadRequest)
		return
	}

	if err := h.cases.UpdateHypothesis(caseID, hypothesis); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

// HandleDeleteHypothesis supprime une hypothèse
func (h *Handler) HandleDeleteHypothesis(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	caseID := r.URL.Query().Get("case_id")
	hypothesisID := r.URL.Query().Get("hypothesis_id")

	if caseID == "" || hypothesisID == "" {
		http.Error(w, "case_id et hypothesis_id requis", http.StatusBadRequest)
		return
	}

	if err := h.cases.DeleteHypothesis(caseID, hypothesisID); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandleAnalyzeHypothesis analyse une hypothèse spécifique avec l'IA
func (h *Handler) HandleAnalyzeHypothesis(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req struct {
		CaseID       string `json:"case_id"`
		HypothesisID string `json:"hypothesis_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON invalide", http.StatusBadRequest)
		return
	}

	// Récupérer l'affaire et l'hypothèse
	c, err := h.cases.GetCase(req.CaseID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	var hypothesis *models.Hypothesis
	for _, hyp := range c.Hypotheses {
		if hyp.ID == req.HypothesisID {
			hypothesis = &hyp
			break
		}
	}
	if hypothesis == nil {
		http.Error(w, "Hypothèse non trouvée", http.StatusNotFound)
		return
	}

	// Construire le contexte pour l'analyse
	graphData, _ := h.cases.BuildGraphData(req.CaseID)
	analysis, err := h.ollama.AnalyzeHypothesis(*hypothesis, c, graphData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"analysis": analysis,
	})
}

// HandleAnalyzePath analyse un chemin entre entités
func (h *Handler) HandleAnalyzePath(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req struct {
		CaseID   string `json:"case_id"`
		FromID   string `json:"from_id"`
		ToID     string `json:"to_id"`
		MaxDepth int    `json:"max_depth"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.CaseID == "" || req.FromID == "" || req.ToID == "" {
		http.Error(w, "case_id, from_id et to_id requis", http.StatusBadRequest)
		return
	}

	if req.MaxDepth == 0 {
		req.MaxDepth = 5
	}

	// Récupérer l'affaire
	caseData, err := h.cases.GetCase(req.CaseID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Utiliser le graphe N4L parsé (comme HandleGraph) pour avoir les vraies relations
	var n4lContent string
	if caseData.N4LContent != "" {
		n4lContent = caseData.N4LContent
	} else {
		n4lContent = h.n4lGenerator.ExportCaseToN4L(caseData)
	}
	parsed := h.n4l.ParseForensicN4L(n4lContent, req.CaseID)
	graphData := &parsed.Graph

	// Résoudre les IDs des nœuds (supporter les labels/noms pour compatibilité N4L)
	resolveNodeID := func(nodeID string) string {
		// D'abord chercher par ID exact
		for _, node := range graphData.Nodes {
			if node.ID == nodeID {
				return nodeID
			}
		}
		// Ensuite chercher par label (nom)
		for _, node := range graphData.Nodes {
			if node.Label == nodeID {
				return node.ID
			}
		}
		return nodeID
	}

	// Fonction pour normaliser les IDs (remplacer underscores par tirets)
	normalizeID := func(id string) string {
		return strings.ReplaceAll(id, "_", "-")
	}

	// DEBUG: Afficher les IDs reçus
	fmt.Printf("[HandleAnalyzePath] IDs reçus: from=%s, to=%s, case=%s\n", req.FromID, req.ToID, req.CaseID)
	fmt.Printf("[HandleAnalyzePath] Nombre d'entités dans l'affaire: %d\n", len(caseData.Entities))

	// Chercher les noms des entités par leurs IDs
	if caseData != nil {
		// Chercher le nom de l'entité par son ID (avec normalisation pour supporter tirets et underscores)
		normalizedFromID := normalizeID(req.FromID)
		normalizedToID := normalizeID(req.ToID)

		fromFound := false
		toFound := false

		for _, entity := range caseData.Entities {
			if entity.ID == req.FromID || entity.ID == normalizedFromID {
				fmt.Printf("[HandleAnalyzePath] FromID trouvé: %s -> %s\n", entity.ID, entity.Name)
				// L'entité a cet ID, chercher le nœud par son nom
				resolved := resolveNodeID(entity.Name)
				if resolved != entity.Name {
					req.FromID = resolved
				} else {
					req.FromID = entity.Name // Utiliser le nom directement car N4L utilise les noms comme IDs
				}
				fromFound = true
				break
			}
		}
		for _, entity := range caseData.Entities {
			if entity.ID == req.ToID || entity.ID == normalizedToID {
				fmt.Printf("[HandleAnalyzePath] ToID trouvé: %s -> %s\n", entity.ID, entity.Name)
				resolved := resolveNodeID(entity.Name)
				if resolved != entity.Name {
					req.ToID = resolved
				} else {
					req.ToID = entity.Name
				}
				toFound = true
				break
			}
		}

		if !fromFound {
			fmt.Printf("[HandleAnalyzePath] ATTENTION: FromID '%s' non trouvé dans les entités!\n", req.FromID)
			// Essayer de trouver directement dans le graphe N4L (peut être un alias N4L)
			for _, node := range graphData.Nodes {
				if node.ID == req.FromID || node.Label == req.FromID {
					fmt.Printf("[HandleAnalyzePath] FromID trouvé dans graphe N4L: %s\n", node.ID)
					req.FromID = node.ID
					fromFound = true
					break
				}
			}
			// Essayer de résoudre via les aliases N4L parsés
			if !fromFound {
				for alias, names := range parsed.Aliases {
					if alias == req.FromID && len(names) > 0 {
						// Extraire le nom de l'entité depuis l'alias
						name := names[0]
						if parenIdx := strings.Index(name, "("); parenIdx > 0 {
							name = strings.TrimSpace(name[:parenIdx])
						}
						fmt.Printf("[HandleAnalyzePath] FromID résolu via alias N4L: %s -> %s\n", alias, name)
						req.FromID = name
						fromFound = true
						break
					}
				}
			}
		}
		if !toFound {
			fmt.Printf("[HandleAnalyzePath] ATTENTION: ToID '%s' non trouvé dans les entités!\n", req.ToID)
			// Essayer de trouver directement dans le graphe N4L (peut être un alias N4L)
			for _, node := range graphData.Nodes {
				if node.ID == req.ToID || node.Label == req.ToID {
					fmt.Printf("[HandleAnalyzePath] ToID trouvé dans graphe N4L: %s\n", node.ID)
					req.ToID = node.ID
					toFound = true
					break
				}
			}
			// Essayer de résoudre via les aliases N4L parsés
			if !toFound {
				for alias, names := range parsed.Aliases {
					if alias == req.ToID && len(names) > 0 {
						// Extraire le nom de l'entité depuis l'alias
						name := names[0]
						if parenIdx := strings.Index(name, "("); parenIdx > 0 {
							name = strings.TrimSpace(name[:parenIdx])
						}
						fmt.Printf("[HandleAnalyzePath] ToID résolu via alias N4L: %s -> %s\n", alias, name)
						req.ToID = name
						toFound = true
						break
					}
				}
			}
		}
	}

	// DEBUG: Afficher les paramètres résolus
	fmt.Printf("[HandleAnalyzePath] FromID résolu: %s, ToID résolu: %s\n", req.FromID, req.ToID)
	fmt.Printf("[HandleAnalyzePath] Graphe: %d noeuds, %d edges\n", len(graphData.Nodes), len(graphData.Edges))

	// Trouver les chemins entre les deux entités
	paths := h.findPaths(graphData, req.FromID, req.ToID, req.MaxDepth)

	fmt.Printf("[HandleAnalyzePath] Chemins trouvés: %d\n", len(paths))

	if len(paths) == 0 {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"paths":    [][]string{},
			"analysis": "Aucun chemin trouvé entre ces deux entités.",
		})
		return
	}

	// Préparer le contexte pour l'analyse
	context := ""
	if caseData != nil {
		context = caseData.Description
	}

	// Analyser le premier chemin avec l'IA
	analysis, err := h.ollama.AnalyzePath(paths[0], context)
	if err != nil {
		analysis = "Erreur lors de l'analyse IA: " + err.Error()
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"paths":    paths,
		"analysis": analysis,
	})
}

// findPaths trouve tous les chemins entre deux nœuds du graphe
func (h *Handler) findPaths(graph *models.GraphData, fromID, toID string, maxDepth int) [][]string {
	// Construire une map d'adjacence
	adjacency := make(map[string][]struct {
		TargetID string
		Label    string
	})

	// Créer un map pour retrouver les labels des nœuds
	nodeLabels := make(map[string]string)
	for _, node := range graph.Nodes {
		nodeLabels[node.ID] = node.Label
	}

	// Construire le graphe d'adjacence (bidirectionnel pour la découverte)
	for _, edge := range graph.Edges {
		adjacency[edge.From] = append(adjacency[edge.From], struct {
			TargetID string
			Label    string
		}{edge.To, edge.Label})
		// Ajouter aussi la relation inverse pour explorer dans les deux sens
		adjacency[edge.To] = append(adjacency[edge.To], struct {
			TargetID string
			Label    string
		}{edge.From, "← " + edge.Label})
	}

	var allPaths [][]string
	visited := make(map[string]bool)

	var dfs func(current string, path []string, depth int)
	dfs = func(current string, path []string, depth int) {
		if depth > maxDepth {
			return
		}

		if current == toID {
			// Chemin trouvé - convertir les IDs en labels
			labelPath := make([]string, len(path))
			for i, id := range path {
				if label, ok := nodeLabels[id]; ok {
					labelPath[i] = label
				} else {
					labelPath[i] = id
				}
			}
			pathCopy := make([]string, len(labelPath))
			copy(pathCopy, labelPath)
			allPaths = append(allPaths, pathCopy)
			return
		}

		visited[current] = true
		defer func() { visited[current] = false }()

		for _, neighbor := range adjacency[current] {
			if !visited[neighbor.TargetID] {
				newPath := append(path, neighbor.TargetID)
				dfs(neighbor.TargetID, newPath, depth+1)
			}
		}
	}

	// Démarrer la recherche
	startPath := []string{fromID}
	dfs(fromID, startPath, 0)

	return allPaths
}

// HandleCrossCase gère la recherche de connexions inter-affaires
func (h *Handler) HandleCrossCase(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req struct {
		CaseID string `json:"case_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.CaseID == "" {
		http.Error(w, "case_id requis", http.StatusBadRequest)
		return
	}

	matches, err := h.cases.FindCrossReferences(req.CaseID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(matches)
}

// HandleCrossCaseAnalyze effectue une analyse IA des patterns inter-affaires
func (h *Handler) HandleCrossCaseAnalyze(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req struct {
		CaseID  string                    `json:"case_id"`
		Matches []models.CrossCaseMatch   `json:"matches"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.CaseID == "" {
		http.Error(w, "case_id requis", http.StatusBadRequest)
		return
	}

	// Récupérer l'affaire courante
	currentCase, err := h.cases.GetCase(req.CaseID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Récupérer toutes les affaires liées
	relatedCases := make(map[string]*models.Case)
	for _, match := range req.Matches {
		if _, exists := relatedCases[match.OtherCaseID]; !exists {
			if c, err := h.cases.GetCase(match.OtherCaseID); err == nil {
				relatedCases[match.OtherCaseID] = c
			}
		}
	}

	// Générer l'analyse IA
	analysis, err := h.ollama.AnalyzeCrossCase(currentCase, relatedCases, req.Matches)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"analysis": analysis,
	})
}

// HandleCrossCaseAnalyzeStream effectue une analyse IA des patterns inter-affaires en streaming
func (h *Handler) HandleCrossCaseAnalyzeStream(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	// Configurer les headers SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming non supporté", http.StatusInternalServerError)
		return
	}

	var req struct {
		CaseID  string                   `json:"case_id"`
		Matches []models.CrossCaseMatch  `json:"matches"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		fmt.Fprintf(w, "data: {\"error\": \"%s\"}\n\n", err.Error())
		flusher.Flush()
		return
	}

	if req.CaseID == "" {
		fmt.Fprintf(w, "data: {\"error\": \"case_id requis\"}\n\n")
		flusher.Flush()
		return
	}

	// Récupérer l'affaire courante
	currentCase, err := h.cases.GetCase(req.CaseID)
	if err != nil {
		fmt.Fprintf(w, "data: {\"error\": \"Affaire non trouvée\"}\n\n")
		flusher.Flush()
		return
	}

	// Récupérer toutes les affaires liées
	relatedCases := make(map[string]*models.Case)
	for _, match := range req.Matches {
		if _, exists := relatedCases[match.OtherCaseID]; !exists {
			if c, err := h.cases.GetCase(match.OtherCaseID); err == nil {
				relatedCases[match.OtherCaseID] = c
			}
		}
	}

	// Streamer l'analyse IA
	err = h.ollama.AnalyzeCrossCaseStream(currentCase, relatedCases, req.Matches, func(chunk string, done bool) error {
		chunkJSON, _ := json.Marshal(map[string]interface{}{
			"chunk": chunk,
			"done":  done,
		})
		fmt.Fprintf(w, "data: %s\n\n", chunkJSON)
		flusher.Flush()
		return nil
	})

	if err != nil {
		errorJSON, _ := json.Marshal(map[string]string{"error": err.Error()})
		fmt.Fprintf(w, "data: %s\n\n", errorJSON)
		flusher.Flush()
	}
}

// HandleCrossCaseGraph construit le graphe multi-affaires
func (h *Handler) HandleCrossCaseGraph(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req struct {
		CaseID  string                    `json:"case_id"`
		Matches []models.CrossCaseMatch   `json:"matches"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	graph, err := h.cases.BuildCrossCaseGraph(req.CaseID, req.Matches)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(graph)
}

// ============================================
// HRM (Hypothetical Reasoning Model) Handlers
// ============================================

// HandleHRMStatus vérifie le statut du service HRM
func (h *Handler) HandleHRMStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Service HRM externe (sapientinc/HRM + Ollama)
	available := h.hrm.IsAvailable()
	json.NewEncoder(w).Encode(map[string]interface{}{
		"available": available,
		"service":   "HRM - Hierarchical Reasoning Model (sapientinc)",
		"url":       "http://localhost:8081",
		"engine":    "sapientinc/HRM + Ollama (raisonnement hiérarchique)",
		"features": []string{
			"Raisonnement hiérarchique en deux niveaux (planification + exécution)",
			"Vérification d'hypothèses avec analyse détaillée",
			"Détection de contradictions multi-niveaux",
			"Analyse inter-affaires avec patterns",
			"Intégration Ollama pour raisonnement textuel",
		},
	})
}

// HandleHRMReason effectue un raisonnement HRM sur une affaire
func (h *Handler) HandleHRMReason(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req struct {
		CaseID        string `json:"case_id"`
		Question      string `json:"question"`
		ReasoningType string `json:"reasoning_type"`
		MaxDepth      int    `json:"max_depth"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.CaseID == "" || req.Question == "" {
		http.Error(w, "case_id et question requis", http.StatusBadRequest)
		return
	}

	// Récupérer l'affaire
	c, err := h.cases.GetCase(req.CaseID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Construire les preuves au format HRM
	var evidence []services.Evidence
	for _, ev := range c.Evidence {
		evidence = append(evidence, services.Evidence{
			ID:          ev.ID,
			Type:        string(ev.Type),
			Description: ev.Description,
			Confidence:  0.7,
		})
	}

	// Configurer les valeurs par défaut
	if req.ReasoningType == "" {
		req.ReasoningType = "deductive"
	}
	if req.MaxDepth == 0 {
		req.MaxDepth = 3
	}

	// Appeler le service HRM basé sur Ollama (raisonnement hiérarchique)
	result, err := h.hrm.Reason(services.ReasoningRequest{
		Context:       h.buildCaseContext(c),
		Question:      req.Question,
		Evidence:      evidence,
		ReasoningType: req.ReasoningType,
		MaxDepth:      req.MaxDepth,
	})
	if err != nil {
		http.Error(w, "Erreur HRM: "+err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(result)
}

// HandleHRMVerifyHypothesis vérifie une hypothèse avec HRM
func (h *Handler) HandleHRMVerifyHypothesis(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req struct {
		CaseID       string `json:"case_id"`
		HypothesisID string `json:"hypothesis_id"`
		StrictMode   bool   `json:"strict_mode"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.CaseID == "" || req.HypothesisID == "" {
		http.Error(w, "case_id et hypothesis_id requis", http.StatusBadRequest)
		return
	}

	// Récupérer l'affaire
	c, err := h.cases.GetCase(req.CaseID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Trouver l'hypothèse
	var hypothesis *models.Hypothesis
	for _, hyp := range c.Hypotheses {
		if hyp.ID == req.HypothesisID {
			hypothesis = &hyp
			break
		}
	}
	if hypothesis == nil {
		http.Error(w, "Hypothèse non trouvée", http.StatusNotFound)
		return
	}

	// Construire les preuves au format HRM
	var evidence []services.Evidence
	for _, ev := range c.Evidence {
		evidence = append(evidence, services.Evidence{
			ID:          ev.ID,
			Type:        string(ev.Type),
			Description: ev.Description,
			Confidence:  0.7,
		})
	}

	// Construire l'hypothèse HRM
	hrmHypothesis := services.Hypothesis{
		ID:                    hypothesis.ID,
		Statement:             hypothesis.Description,
		SupportingEvidence:    []string{},
		ContradictingEvidence: []string{},
		Confidence:            float64(hypothesis.ConfidenceLevel) / 100.0,
	}

	// Appeler le service HRM basé sur Ollama
	result, err := h.hrm.VerifyHypothesis(services.HypothesisVerificationRequest{
		Hypothesis:  hrmHypothesis,
		Evidence:    evidence,
		CaseContext: h.buildCaseContext(c),
		StrictMode:  req.StrictMode,
	})
	if err != nil {
		http.Error(w, "Erreur HRM: "+err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(result)
}

// HandleHRMContradictions détecte les contradictions avec HRM
func (h *Handler) HandleHRMContradictions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req struct {
		CaseID string `json:"case_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.CaseID == "" {
		http.Error(w, "case_id requis", http.StatusBadRequest)
		return
	}

	// Récupérer l'affaire
	c, err := h.cases.GetCase(req.CaseID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Construire les statements à partir des témoignages et hypothèses
	var statements []map[string]string
	for _, hyp := range c.Hypotheses {
		statements = append(statements, map[string]string{
			"id":      hyp.ID,
			"content": hyp.Description,
			"source":  "hypothesis",
		})
	}
	// Ajouter les descriptions des preuves
	for _, ev := range c.Evidence {
		statements = append(statements, map[string]string{
			"id":      ev.ID,
			"content": ev.Description,
			"source":  "evidence",
		})
	}

	// Construire les preuves au format HRM
	var evidence []services.Evidence
	for _, ev := range c.Evidence {
		evidence = append(evidence, services.Evidence{
			ID:          ev.ID,
			Type:        string(ev.Type),
			Description: ev.Description,
			Confidence:  0.7,
		})
	}

	// Appeler le service HRM basé sur Ollama
	result, err := h.hrm.FindContradictions(services.ContradictionRequest{
		Statements:  statements,
		Evidence:    evidence,
		CaseContext: h.buildCaseContext(c),
	})
	if err != nil {
		http.Error(w, "Erreur HRM: "+err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(result)
}

// HandleHRMCrossCase analyse les connexions inter-affaires avec HRM
func (h *Handler) HandleHRMCrossCase(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req struct {
		CaseID       string   `json:"case_id"`
		OtherCaseIDs []string `json:"other_case_ids"`
		FocusAreas   []string `json:"focus_areas"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.CaseID == "" {
		http.Error(w, "case_id requis", http.StatusBadRequest)
		return
	}

	// Récupérer l'affaire principale
	primaryCase, err := h.cases.GetCase(req.CaseID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Convertir en format HRM
	primaryHRM := h.caseToHRMFormat(primaryCase)

	// Récupérer les affaires de comparaison
	var comparisonCases []map[string]interface{}
	if len(req.OtherCaseIDs) > 0 {
		for _, caseID := range req.OtherCaseIDs {
			if c, err := h.cases.GetCase(caseID); err == nil {
				comparisonCases = append(comparisonCases, h.caseToHRMFormat(c))
			}
		}
	} else {
		// Si aucune affaire spécifiée, comparer avec toutes les autres
		allCases := h.cases.GetAllCases()
		for _, c := range allCases {
			if c.ID != req.CaseID {
				comparisonCases = append(comparisonCases, h.caseToHRMFormat(c))
			}
		}
	}

	// Appeler le service HRM basé sur Ollama
	result, err := h.hrm.CrossCaseReasoning(services.CrossCaseRequest{
		PrimaryCase:     primaryHRM,
		ComparisonCases: comparisonCases,
		FocusAreas:      req.FocusAreas,
	})
	if err != nil {
		http.Error(w, "Erreur HRM: "+err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(result)
}

// caseToHRMFormat convertit une affaire au format HRM
func (h *Handler) caseToHRMFormat(c *models.Case) map[string]interface{} {
	// Convertir la timeline
	var timeline []map[string]interface{}
	for _, ev := range c.Timeline {
		timeline = append(timeline, map[string]interface{}{
			"id":          ev.ID,
			"timestamp":   ev.Timestamp,
			"description": ev.Description,
		})
	}

	// Convertir les preuves
	var evidence []map[string]interface{}
	for _, ev := range c.Evidence {
		evidence = append(evidence, map[string]interface{}{
			"id":          ev.ID,
			"type":        string(ev.Type),
			"description": ev.Description,
		})
	}

	// Convertir les hypothèses
	var hypotheses []map[string]interface{}
	for _, hyp := range c.Hypotheses {
		hypotheses = append(hypotheses, map[string]interface{}{
			"id":          hyp.ID,
			"description": hyp.Description,
			"confidence":  float64(hyp.ConfidenceLevel) / 100.0,
		})
	}

	return map[string]interface{}{
		"id":          c.ID,
		"name":        c.Name,
		"type":        c.Type,
		"description": c.Description,
		"timeline":    timeline,
		"evidence":    evidence,
		"hypotheses":  hypotheses,
	}
}

// ============================================
// Recherche Hybride (BM25 + Sémantique)
// ============================================

// HandleHybridSearch effectue une recherche hybride sur une affaire
func (h *Handler) HandleHybridSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req services.SearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.CaseID == "" || req.Query == "" {
		http.Error(w, "case_id et query requis", http.StatusBadRequest)
		return
	}

	// Récupérer l'affaire
	c, err := h.cases.GetCase(req.CaseID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Effectuer la recherche hybride
	results, err := h.search.HybridSearch(c, req)
	if err != nil {
		http.Error(w, "Erreur de recherche: "+err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"query":       req.Query,
		"results":     results,
		"count":       len(results),
		"bm25_weight": req.BM25Weight,
	})
}

// HandleConvertToN4L convertit du texte en format N4L via Ollama avec le modèle n4l-qwen
func (h *Handler) HandleConvertToN4L(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Text == "" {
		http.Error(w, "text requis", http.StatusBadRequest)
		return
	}

	// Utiliser le modèle n4l-qwen:latest pour la conversion
	n4lOllama := services.NewOllamaService("http://localhost:11434", "n4l-qwen:latest")

	prompt := fmt.Sprintf(`Transforme ce texte en format N4L (Notes for Learning).

Texte à transformer:
%s

Génère un graphe sémantique N4L structuré avec:
- Sections: -NomSection ou ---
- Contextes: :: mot1, mot2 ::
- Relations: Sujet (relation) Objet
- Ditto: " pour répéter le sujet précédent
- Groupes: Nom => { element1; element2 }
- Timeline: +:: _timeline_ ::
- Références: @alias, $alias.1

Réponds UNIQUEMENT avec le format N4L, sans explication.`, req.Text)

	n4lContent, err := n4lOllama.Generate(prompt)
	if err != nil {
		http.Error(w, "Erreur conversion N4L: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Parser le contenu N4L généré
	parsedData := h.n4l.ParseN4L(n4lContent)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"n4l_content": n4lContent,
		"parsed":      parsedData,
	})
}

// HandleQuickSearch effectue une recherche rapide BM25 uniquement
func (h *Handler) HandleQuickSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	caseID := r.URL.Query().Get("case_id")
	query := r.URL.Query().Get("q")

	if caseID == "" || query == "" {
		http.Error(w, "case_id et q requis", http.StatusBadRequest)
		return
	}

	// Récupérer l'affaire
	c, err := h.cases.GetCase(caseID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Effectuer la recherche BM25 rapide
	results := h.search.QuickBM25Search(c, query, 10)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"query":   query,
		"results": results,
		"count":   len(results),
	})
}

// HandleConfigPrompts gère la configuration des prompts
func (h *Handler) HandleConfigPrompts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	configService := h.ollama.GetConfigService()
	if configService == nil {
		http.Error(w, "Service de configuration non disponible", http.StatusInternalServerError)
		return
	}

	switch r.Method {
	case http.MethodGet:
		// Retourner la configuration complète
		config := configService.GetConfig()
		json.NewEncoder(w).Encode(config)

	case http.MethodPut:
		// Mettre à jour la configuration complète
		var config services.PromptsConfig
		if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
			http.Error(w, "Erreur de parsing: "+err.Error(), http.StatusBadRequest)
			return
		}

		if err := configService.UpdateConfig(&config); err != nil {
			http.Error(w, "Erreur de mise à jour: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if err := configService.SaveConfig(); err != nil {
			http.Error(w, "Erreur de sauvegarde: "+err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Configuration sauvegardée",
		})

	default:
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
	}
}

// HandleConfigPrompt gère un prompt spécifique
func (h *Handler) HandleConfigPrompt(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	configService := h.ollama.GetConfigService()
	if configService == nil {
		http.Error(w, "Service de configuration non disponible", http.StatusInternalServerError)
		return
	}

	// Extraire le nom du prompt de l'URL
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "Nom du prompt manquant", http.StatusBadRequest)
		return
	}
	promptName := parts[len(parts)-1]

	switch r.Method {
	case http.MethodGet:
		prompt := configService.GetPrompt(promptName)
		if prompt == nil {
			http.Error(w, "Prompt non trouvé", http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(prompt)

	case http.MethodPut:
		var prompt services.PromptConfig
		if err := json.NewDecoder(r.Body).Decode(&prompt); err != nil {
			http.Error(w, "Erreur de parsing: "+err.Error(), http.StatusBadRequest)
			return
		}

		if err := configService.UpdatePrompt(promptName, prompt); err != nil {
			http.Error(w, "Erreur de mise à jour: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if err := configService.SaveConfig(); err != nil {
			http.Error(w, "Erreur de sauvegarde: "+err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Prompt sauvegardé",
		})

	default:
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
	}
}

// HandleConfigReload recharge la configuration depuis le fichier
func (h *Handler) HandleConfigReload(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	configService := h.ollama.GetConfigService()
	if configService == nil {
		http.Error(w, "Service de configuration non disponible", http.StatusInternalServerError)
		return
	}

	if err := configService.ReloadConfig(); err != nil {
		http.Error(w, "Erreur de rechargement: "+err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Configuration rechargée",
	})
}

// ==================== GRAPH ANALYSIS HANDLERS ====================

// HandleFindClusters détecte les clusters dans le graphe
func (h *Handler) HandleFindClusters(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req struct {
		CaseID string `json:"case_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	graph, err := h.cases.BuildGraphData(req.CaseID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	clusters := h.graphAnalyzer.FindClusters(*graph)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"clusters": clusters,
		"count":    len(clusters),
	})
}

// HandleFindPaths trouve tous les chemins entre deux nœuds
func (h *Handler) HandleFindPaths(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req struct {
		CaseID   string `json:"case_id"`
		From     string `json:"from"`
		To       string `json:"to"`
		MaxDepth int    `json:"max_depth"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	graph, err := h.cases.BuildGraphData(req.CaseID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if req.MaxDepth <= 0 {
		req.MaxDepth = 5
	}

	// Résoudre les IDs des nœuds (supporter les labels/noms pour compatibilité N4L)
	resolveNodeID := func(nodeID string) string {
		// D'abord chercher par ID exact
		for _, node := range graph.Nodes {
			if node.ID == nodeID {
				return nodeID
			}
		}
		// Ensuite chercher par label (nom)
		for _, node := range graph.Nodes {
			if node.Label == nodeID {
				return node.ID
			}
		}
		return nodeID
	}

	resolvedFrom := resolveNodeID(req.From)
	resolvedTo := resolveNodeID(req.To)

	paths := h.graphAnalyzer.FindAllPaths(*graph, resolvedFrom, resolvedTo, req.MaxDepth)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"paths": paths,
		"count": len(paths),
	})
}

// HandleLayeredGraph retourne le graphe organisé en couches
func (h *Handler) HandleLayeredGraph(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	caseID := r.URL.Query().Get("case_id")
	if caseID == "" {
		http.Error(w, "case_id requis", http.StatusBadRequest)
		return
	}

	graph, err := h.cases.BuildGraphData(caseID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	layered := h.graphAnalyzer.GetLayeredGraph(*graph)
	json.NewEncoder(w).Encode(layered)
}

// HandleExpansionCone retourne le cône d'expansion d'un nœud
func (h *Handler) HandleExpansionCone(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req struct {
		CaseID string `json:"case_id"`
		NodeID string `json:"node_id"`
		Depth  int    `json:"depth"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	graph, err := h.cases.BuildGraphData(req.CaseID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if req.Depth <= 0 {
		req.Depth = 3
	}

	cone := h.graphAnalyzer.GetExpansionCone(*graph, req.NodeID, req.Depth)
	json.NewEncoder(w).Encode(cone)
}

// HandleDensityMap retourne la carte de densité du graphe
func (h *Handler) HandleDensityMap(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	caseID := r.URL.Query().Get("case_id")
	if caseID == "" {
		http.Error(w, "case_id requis", http.StatusBadRequest)
		return
	}

	graph, err := h.cases.BuildGraphData(caseID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	density := h.graphAnalyzer.GetDensityMap(*graph)
	json.NewEncoder(w).Encode(density)
}

// HandleTemporalPatterns détecte les patterns temporels
func (h *Handler) HandleTemporalPatterns(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	caseID := r.URL.Query().Get("case_id")
	if caseID == "" {
		http.Error(w, "case_id requis", http.StatusBadRequest)
		return
	}

	graph, err := h.cases.BuildGraphData(caseID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	patterns := h.graphAnalyzer.DetectTemporalPatterns(*graph)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"patterns": patterns,
		"count":    len(patterns),
	})
}

// HandleCheckConsistency vérifie la cohérence du graphe
func (h *Handler) HandleCheckConsistency(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	caseID := r.URL.Query().Get("case_id")
	if caseID == "" {
		http.Error(w, "case_id requis", http.StatusBadRequest)
		return
	}

	graph, err := h.cases.BuildGraphData(caseID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	result := h.graphAnalyzer.CheckConsistency(*graph)
	json.NewEncoder(w).Encode(result)
}

// ==================== INVESTIGATION MODE HANDLERS ====================

// HandleStartInvestigation démarre une nouvelle session d'investigation
func (h *Handler) HandleStartInvestigation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req struct {
		CaseID string `json:"case_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	graph, err := h.cases.BuildGraphData(req.CaseID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	session := h.graphAnalyzer.CreateInvestigationSession(req.CaseID, *graph)
	json.NewEncoder(w).Encode(session)
}

// HandleInvestigationSuggestions génère des suggestions pour une étape
func (h *Handler) HandleInvestigationSuggestions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req struct {
		CaseID string `json:"case_id"`
		StepID string `json:"step_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	graph, err := h.cases.BuildGraphData(req.CaseID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	suggestions := h.graphAnalyzer.GetStepSuggestions(*graph, req.StepID)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"suggestions": suggestions,
		"step_id":     req.StepID,
	})
}

// HandleInvestigationAnalyze analyse une étape d'investigation avec l'IA
func (h *Handler) HandleInvestigationAnalyze(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req struct {
		CaseID   string   `json:"case_id"`
		StepID   string   `json:"step_id"`
		Question string   `json:"question"`
		Context  []string `json:"context"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Récupérer l'affaire pour le contexte
	c, err := h.cases.GetCase(req.CaseID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Récupérer le graphe pour avoir les relations
	graph, _ := h.cases.BuildGraphData(req.CaseID)

	// Construire le prompt d'investigation
	stepNames := map[string]string{
		"actors":    "Identification des Acteurs",
		"locations": "Analyse des Lieux",
		"timeline":  "Reconstitution Chronologique",
		"motives":   "Analyse des Mobiles",
		"evidence":  "Évaluation des Preuves",
		"synthesis": "Synthèse et Hypothèses",
	}

	stepName := stepNames[req.StepID]
	if stepName == "" {
		stepName = req.StepID
	}

	// Construire le contexte enrichi avec les vraies données
	contextStr := fmt.Sprintf("# Affaire: %s\nType: %s\n\n", c.Name, c.Type)

	// Ajouter les entités avec leurs types
	if len(c.Entities) > 0 {
		contextStr += "## Entités impliquées:\n"
		for _, e := range c.Entities {
			contextStr += fmt.Sprintf("- %s (%s): %s\n", e.Name, e.Type, e.Description)
		}
		contextStr += "\n"
	}

	// Ajouter les relations depuis le graphe
	if graph != nil && len(graph.Edges) > 0 {
		contextStr += "## Relations entre entités:\n"
		// Créer une map ID -> Label
		nodeLabels := make(map[string]string)
		for _, n := range graph.Nodes {
			nodeLabels[n.ID] = n.Label
		}
		for _, edge := range graph.Edges {
			fromName := nodeLabels[edge.From]
			if fromName == "" {
				fromName = edge.From
			}
			toName := nodeLabels[edge.To]
			if toName == "" {
				toName = edge.To
			}
			contextStr += fmt.Sprintf("- %s -[%s]-> %s\n", fromName, edge.Label, toName)
		}
		contextStr += "\n"
	}

	// Ajouter la timeline
	if len(c.Timeline) > 0 {
		contextStr += "## Chronologie des événements:\n"
		for _, t := range c.Timeline {
			contextStr += fmt.Sprintf("- %s: %s - %s\n", t.Timestamp.Format("2006-01-02 15:04"), t.Title, t.Description)
		}
		contextStr += "\n"
	}

	// Ajouter les preuves
	if len(c.Evidence) > 0 {
		contextStr += "## Preuves collectées:\n"
		for _, e := range c.Evidence {
			contextStr += fmt.Sprintf("- [%s] %s: %s\n", e.Type, e.Name, e.Description)
		}
		contextStr += "\n"
	}

	if len(req.Context) > 0 {
		contextStr += "## Notes de l'enquêteur:\n"
		for _, ctx := range req.Context {
			contextStr += "- " + ctx + "\n"
		}
	}

	// Construire le prompt spécifique à l'étape
	var question string
	switch req.StepID {
	case "actors":
		question = `Analyse les acteurs de cette affaire. Pour chaque personne impliquée:
1. Identifie son rôle (victime, suspect, témoin, complice)
2. Liste ses connexions avec les autres acteurs
3. Évalue son importance dans l'enquête (centrale, périphérique)
4. Note tout comportement suspect ou alibi mentionné

Réponds de manière concise et structurée.`
	case "locations":
		question = `Analyse les lieux liés à cette affaire:
1. Identifie tous les lieux mentionnés
2. Détermine leur importance (scène de crime, lieu de résidence, point de rencontre)
3. Note les connexions entre lieux et personnes
4. Identifie les déplacements significatifs

Réponds de manière concise.`
	case "timeline":
		question = `Reconstitue la chronologie de l'affaire:
1. Liste les événements dans l'ordre chronologique
2. Identifie les trous ou incohérences temporelles
3. Note les alibis et leur validité
4. Suggère les moments clés à investiguer

Réponds de manière structurée.`
	case "motives":
		question = `Analyse les mobiles potentiels:
1. Pour chaque suspect, identifie les mobiles possibles
2. Cherche les conflits, dettes, héritages, rivalités
3. Évalue la force de chaque mobile
4. Note les indices supportant chaque théorie

Réponds de manière analytique.`
	case "evidence":
		question = `Évalue les preuves de l'affaire:
1. Classe les preuves par type (matérielle, testimoniale, documentaire)
2. Évalue leur fiabilité et leur force probante
3. Identifie les preuves manquantes
4. Note les contradictions entre preuves

Réponds de manière objective.`
	case "synthesis":
		question = `Synthétise l'ensemble de l'affaire:
1. Résume les faits établis
2. Liste les hypothèses principales
3. Identifie les zones d'ombre
4. Propose les prochaines étapes d'investigation

Réponds de manière concise et actionnable.`
	default:
		question = fmt.Sprintf("Analyse l'étape '%s' pour cette affaire. Identifie les éléments clés et propose des pistes concrètes.", stepName)
	}

	if req.Question != "" {
		question = req.Question
	}

	// Log pour debug
	log.Printf("=== INVESTIGATION ANALYZE ===")
	log.Printf("Step: %s", req.StepID)
	log.Printf("Context length: %d chars", len(contextStr))
	log.Printf("--- CONTEXT ---\n%s", contextStr)
	log.Printf("--- QUESTION ---\n%s", question)

	// Appeler l'IA
	response, err := h.ollama.Chat(question, contextStr)
	if err != nil {
		log.Printf("ERROR: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("--- RESPONSE ---\n%s", response)
	log.Printf("=== END INVESTIGATION ANALYZE ===")

	json.NewEncoder(w).Encode(map[string]interface{}{
		"analysis": response,
		"step_id":  req.StepID,
	})
}

// HandleInvestigationAnalyzeStream analyse une étape d'investigation avec l'IA en streaming
func (h *Handler) HandleInvestigationAnalyzeStream(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	// Configurer les headers SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming non supporté", http.StatusInternalServerError)
		return
	}

	var req struct {
		CaseID   string   `json:"case_id"`
		StepID   string   `json:"step_id"`
		Question string   `json:"question"`
		Context  []string `json:"context"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		fmt.Fprintf(w, "data: {\"error\": \"%s\"}\n\n", err.Error())
		flusher.Flush()
		return
	}

	// Récupérer l'affaire pour le contexte
	c, err := h.cases.GetCase(req.CaseID)
	if err != nil {
		fmt.Fprintf(w, "data: {\"error\": \"Affaire non trouvée\"}\n\n")
		flusher.Flush()
		return
	}

	// Récupérer le graphe pour avoir les relations
	graph, _ := h.cases.BuildGraphData(req.CaseID)

	// Construire le prompt d'investigation
	stepNames := map[string]string{
		"actors":    "Identification des Acteurs",
		"locations": "Analyse des Lieux",
		"timeline":  "Reconstitution Chronologique",
		"motives":   "Analyse des Mobiles",
		"evidence":  "Évaluation des Preuves",
		"synthesis": "Synthèse et Hypothèses",
	}

	stepName := stepNames[req.StepID]
	if stepName == "" {
		stepName = req.StepID
	}

	// Construire le contexte enrichi avec les vraies données
	contextStr := fmt.Sprintf("# Affaire: %s\nType: %s\n\n", c.Name, c.Type)

	// Ajouter les entités avec leurs types
	if len(c.Entities) > 0 {
		contextStr += "## Entités impliquées:\n"
		for _, e := range c.Entities {
			contextStr += fmt.Sprintf("- %s (%s): %s\n", e.Name, e.Type, e.Description)
		}
		contextStr += "\n"
	}

	// Ajouter les relations depuis le graphe
	if graph != nil && len(graph.Edges) > 0 {
		contextStr += "## Relations entre entités:\n"
		nodeLabels := make(map[string]string)
		for _, n := range graph.Nodes {
			nodeLabels[n.ID] = n.Label
		}
		for _, edge := range graph.Edges {
			fromName := nodeLabels[edge.From]
			if fromName == "" {
				fromName = edge.From
			}
			toName := nodeLabels[edge.To]
			if toName == "" {
				toName = edge.To
			}
			contextStr += fmt.Sprintf("- %s -[%s]-> %s\n", fromName, edge.Label, toName)
		}
		contextStr += "\n"
	}

	// Ajouter la timeline
	if len(c.Timeline) > 0 {
		contextStr += "## Chronologie des événements:\n"
		for _, t := range c.Timeline {
			contextStr += fmt.Sprintf("- %s: %s - %s\n", t.Timestamp.Format("2006-01-02 15:04"), t.Title, t.Description)
		}
		contextStr += "\n"
	}

	// Ajouter les preuves
	if len(c.Evidence) > 0 {
		contextStr += "## Preuves collectées:\n"
		for _, e := range c.Evidence {
			contextStr += fmt.Sprintf("- [%s] %s: %s\n", e.Type, e.Name, e.Description)
		}
		contextStr += "\n"
	}

	if len(req.Context) > 0 {
		contextStr += "## Notes de l'enquêteur:\n"
		for _, ctx := range req.Context {
			contextStr += "- " + ctx + "\n"
		}
	}

	// Construire le prompt spécifique à l'étape
	var question string
	switch req.StepID {
	case "actors":
		question = `Analyse les acteurs de cette affaire. Pour chaque personne impliquée:
1. Identifie son rôle (victime, suspect, témoin, complice)
2. Liste ses connexions avec les autres acteurs
3. Évalue son importance dans l'enquête (centrale, périphérique)
4. Note tout comportement suspect ou alibi mentionné

Réponds de manière concise et structurée.`
	case "locations":
		question = `Analyse les lieux liés à cette affaire:
1. Identifie tous les lieux mentionnés
2. Détermine leur importance (scène de crime, lieu de résidence, point de rencontre)
3. Note les connexions entre lieux et personnes
4. Identifie les déplacements significatifs

Réponds de manière concise.`
	case "timeline":
		question = `Reconstitue la chronologie de l'affaire:
1. Liste les événements dans l'ordre chronologique
2. Identifie les trous ou incohérences temporelles
3. Note les alibis et leur validité
4. Suggère les moments clés à investiguer

Réponds de manière structurée.`
	case "motives":
		question = `Analyse les mobiles potentiels:
1. Pour chaque suspect, identifie les mobiles possibles
2. Cherche les conflits, dettes, héritages, rivalités
3. Évalue la force de chaque mobile
4. Note les indices supportant chaque théorie

Réponds de manière analytique.`
	case "evidence":
		question = `Évalue les preuves de l'affaire:
1. Classe les preuves par type (matérielle, testimoniale, documentaire)
2. Évalue leur fiabilité et leur force probante
3. Identifie les preuves manquantes
4. Note les contradictions entre preuves

Réponds de manière objective.`
	case "synthesis":
		question = `Synthétise l'ensemble de l'affaire:
1. Résume les faits établis
2. Liste les hypothèses principales
3. Identifie les zones d'ombre
4. Propose les prochaines étapes d'investigation

Réponds de manière concise et actionnable.`
	default:
		question = fmt.Sprintf("Analyse l'étape '%s' pour cette affaire. Identifie les éléments clés et propose des pistes concrètes.", stepName)
	}

	if req.Question != "" {
		question = req.Question
	}

	// Appeler l'IA en streaming
	err = h.ollama.ChatStream(question, contextStr, func(chunk string, done bool) error {
		chunkJSON, _ := json.Marshal(map[string]interface{}{
			"chunk": chunk,
			"done":  done,
		})
		fmt.Fprintf(w, "data: %s\n\n", chunkJSON)
		flusher.Flush()
		return nil
	})

	if err != nil {
		errorJSON, _ := json.Marshal(map[string]string{"error": err.Error()})
		fmt.Fprintf(w, "data: %s\n\n", errorJSON)
		flusher.Flush()
	}
}

// HandleGraphAnalyzeComplete effectue une analyse complète du graphe
func (h *Handler) HandleGraphAnalyzeComplete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	caseID := r.URL.Query().Get("case_id")
	if caseID == "" {
		http.Error(w, "case_id requis", http.StatusBadRequest)
		return
	}

	// Récupérer l'affaire pour les analyses avancées
	caseData, err := h.cases.GetCase(caseID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Utiliser le graphe N4L comme source de données (comme HandleGraph)
	var n4lContent string
	if caseData.N4LContent != "" {
		n4lContent = caseData.N4LContent
	} else {
		// Générer le N4L à partir des données du cas
		n4lContent = h.n4lGenerator.ExportCaseToN4L(caseData)
	}

	// Parser le N4L pour obtenir le graphe
	parsed := h.n4l.ParseForensicN4L(n4lContent, caseID)
	graph := &parsed.Graph

	// Effectuer toutes les analyses
	clusters := h.graphAnalyzer.FindClusters(*graph)
	layered := h.graphAnalyzer.GetLayeredGraph(*graph)
	density := h.graphAnalyzer.GetDensityMap(*graph)
	patterns := h.graphAnalyzer.DetectTemporalPatterns(*graph)
	consistency := h.graphAnalyzer.CheckConsistency(*graph)

	// Nouvelles analyses
	centrality := h.graphAnalyzer.CalculateCentrality(*graph)
	var suspicion []services.SuspicionResult
	var alibis services.AlibiTimeline
	if caseData != nil {
		suspicion = h.graphAnalyzer.CalculateSuspicionScores(*graph, caseData)
		alibis = h.graphAnalyzer.BuildAlibiTimeline(caseData)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"clusters":    clusters,
		"layered":     layered,
		"density":     density,
		"patterns":    patterns,
		"consistency": consistency,
		"centrality":  centrality,
		"suspicion":   suspicion,
		"alibis":      alibis,
		"graph":       graph, // Ajout du graphe pour le mini-graph des clusters
		"summary": map[string]interface{}{
			"total_nodes":         len(graph.Nodes),
			"total_edges":         len(graph.Edges),
			"cluster_count":       len(clusters),
			"layer_count":         len(layered.Layers),
			"pattern_count":       len(patterns),
			"is_consistent":       consistency.IsConsistent,
			"orphan_count":        len(consistency.OrphanNodes),
			"contradiction_count": len(consistency.Contradictions),
		},
	})
}

// ============================================
// NOTEBOOK HANDLERS
// ============================================

// HandleNotebook gère les opérations sur le notebook d'une affaire
func (h *Handler) HandleNotebook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	caseID := r.URL.Query().Get("case_id")
	if caseID == "" {
		http.Error(w, "case_id requis", http.StatusBadRequest)
		return
	}

	// Récupérer le nom de l'affaire
	caseData, err := h.cases.GetCase(caseID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	switch r.Method {
	case http.MethodGet:
		// Récupérer le notebook avec les paramètres de recherche/tri
		query := r.URL.Query().Get("q")
		noteType := r.URL.Query().Get("type")
		sortBy := r.URL.Query().Get("sort")

		if query != "" || noteType != "" {
			// Recherche filtrée
			notes := h.notebook.SearchNotes(caseID, query, noteType, services.SortOrder(sortBy))
			json.NewEncoder(w).Encode(map[string]interface{}{
				"case_id":   caseID,
				"case_name": caseData.Name,
				"notes":     notes,
				"count":     len(notes),
			})
		} else {
			// Récupérer tout le notebook
			notebook := h.notebook.GetNotebook(caseID, caseData.Name)
			json.NewEncoder(w).Encode(notebook)
		}

	default:
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
	}
}

// HandleNotebookStats retourne les statistiques du notebook
func (h *Handler) HandleNotebookStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	caseID := r.URL.Query().Get("case_id")
	if caseID == "" {
		http.Error(w, "case_id requis", http.StatusBadRequest)
		return
	}

	stats := h.notebook.GetNotebookStats(caseID)
	json.NewEncoder(w).Encode(stats)
}

// HandleNotes gère les opérations CRUD sur les notes
func (h *Handler) HandleNotes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	caseID := r.URL.Query().Get("case_id")
	if caseID == "" {
		http.Error(w, "case_id requis", http.StatusBadRequest)
		return
	}

	// Récupérer le nom de l'affaire
	caseData, err := h.cases.GetCase(caseID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	switch r.Method {
	case http.MethodPost:
		// Ajouter une nouvelle note
		var note models.Note
		if err := json.NewDecoder(r.Body).Decode(&note); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		result, err := h.notebook.AddNote(caseID, caseData.Name, note)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(result)

	default:
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
	}
}

// HandleNote gère une note spécifique (GET, PUT, DELETE)
func (h *Handler) HandleNote(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	caseID := r.URL.Query().Get("case_id")
	noteID := r.URL.Query().Get("note_id")

	if caseID == "" || noteID == "" {
		http.Error(w, "case_id et note_id requis", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		note, err := h.notebook.GetNote(caseID, noteID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(note)

	case http.MethodPut:
		var note models.Note
		if err := json.NewDecoder(r.Body).Decode(&note); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		note.ID = noteID

		if err := h.notebook.UpdateNote(caseID, note); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})

	case http.MethodDelete:
		if err := h.notebook.DeleteNote(caseID, noteID); err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
	}
}

// HandleNotePin épingle/désépingle une note
func (h *Handler) HandleNotePin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	caseID := r.URL.Query().Get("case_id")
	noteID := r.URL.Query().Get("note_id")

	if caseID == "" || noteID == "" {
		http.Error(w, "case_id et note_id requis", http.StatusBadRequest)
		return
	}

	if err := h.notebook.TogglePinNote(caseID, noteID); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// HandleNoteFavorite marque/démarque une note comme favorite
func (h *Handler) HandleNoteFavorite(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	caseID := r.URL.Query().Get("case_id")
	noteID := r.URL.Query().Get("note_id")

	if caseID == "" || noteID == "" {
		http.Error(w, "case_id et note_id requis", http.StatusBadRequest)
		return
	}

	if err := h.notebook.ToggleFavoriteNote(caseID, noteID); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// HandleNoteTag gère les tags d'une note
func (h *Handler) HandleNoteTag(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	caseID := r.URL.Query().Get("case_id")
	noteID := r.URL.Query().Get("note_id")
	tag := r.URL.Query().Get("tag")

	if caseID == "" || noteID == "" || tag == "" {
		http.Error(w, "case_id, note_id et tag requis", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodPost:
		// Ajouter un tag
		if err := h.notebook.AddTagToNote(caseID, noteID, tag); err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})

	case http.MethodDelete:
		// Supprimer un tag
		if err := h.notebook.RemoveTagFromNote(caseID, noteID, tag); err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})

	default:
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
	}
}

// ============================================
// ADVANCED GRAPH ANALYSIS HANDLERS
// Inspired by SSTorytime features
// ============================================

// HandleConeSearch effectue une recherche par cône d'expansion
func (h *Handler) HandleConeSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req services.ConeSearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.CaseID == "" {
		http.Error(w, "case_id requis", http.StatusBadRequest)
		return
	}

	graph, err := h.cases.BuildGraphData(req.CaseID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Debug logging
	log.Printf("[ConeSearch] CaseID=%s, StartNode=%s, Direction=%s, Depth=%d", req.CaseID, req.StartNode, req.Direction, req.Depth)
	log.Printf("[ConeSearch] Graph has %d nodes, %d edges", len(graph.Nodes), len(graph.Edges))

	result, err := h.search.ConeSearch(graph, req)
	if err != nil {
		log.Printf("[ConeSearch] Error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("[ConeSearch] Result: %d total nodes, %d levels", result.TotalNodes, len(result.Levels))
	json.NewEncoder(w).Encode(result)
}

// HandleDiracPathSearch recherche des chemins style Dirac <end|start>
func (h *Handler) HandleDiracPathSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req struct {
		CaseID     string   `json:"case_id"`
		StartNodes []string `json:"start_nodes"`
		EndNodes   []string `json:"end_nodes"`
		MaxDepth   int      `json:"max_depth"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.CaseID == "" {
		http.Error(w, "case_id requis", http.StatusBadRequest)
		return
	}

	graph, err := h.cases.BuildGraphData(req.CaseID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	paths, err := h.search.DiracPathSearch(graph, req.StartNodes, req.EndNodes, req.MaxDepth)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"paths": paths,
		"count": len(paths),
	})
}

// HandleAppointedNodes détecte les nœuds appointés (corrélations)
func (h *Handler) HandleAppointedNodes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	caseID := r.URL.Query().Get("case_id")
	if caseID == "" {
		http.Error(w, "case_id requis", http.StatusBadRequest)
		return
	}

	minPointers := 2
	if mp := r.URL.Query().Get("min_pointers"); mp != "" {
		fmt.Sscanf(mp, "%d", &minPointers)
	}

	graph, err := h.cases.BuildGraphData(caseID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	result := h.graphAnalyzer.FindAppointedNodes(*graph, minPointers)
	json.NewEncoder(w).Encode(result)
}

// HandleEigenvectorCentrality calcule la centralité eigenvector
func (h *Handler) HandleEigenvectorCentrality(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	caseID := r.URL.Query().Get("case_id")
	if caseID == "" {
		http.Error(w, "case_id requis", http.StatusBadRequest)
		return
	}

	maxIterations := 100
	if mi := r.URL.Query().Get("max_iterations"); mi != "" {
		fmt.Sscanf(mi, "%d", &maxIterations)
	}

	graph, err := h.cases.BuildGraphData(caseID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	result := h.graphAnalyzer.CalculateEigenvectorCentrality(*graph, maxIterations)
	json.NewEncoder(w).Encode(result)
}

// HandleSTTypeAnalysis analyse la structure sémantique par STTypes
func (h *Handler) HandleSTTypeAnalysis(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	caseID := r.URL.Query().Get("case_id")
	if caseID == "" {
		http.Error(w, "case_id requis", http.StatusBadRequest)
		return
	}

	graph, err := h.cases.BuildGraphData(caseID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	result := h.n4l.AnalyzeGraphBySTTypes(*graph)
	json.NewEncoder(w).Encode(result)
}

// HandleAdvancedGraphAnalysis effectue toutes les analyses avancées en une requête
func (h *Handler) HandleAdvancedGraphAnalysis(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	caseID := r.URL.Query().Get("case_id")
	if caseID == "" {
		http.Error(w, "case_id requis", http.StatusBadRequest)
		return
	}

	graph, err := h.cases.BuildGraphData(caseID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Effectuer toutes les analyses
	appointed := h.graphAnalyzer.FindAppointedNodes(*graph, 2)
	eigenvector := h.graphAnalyzer.CalculateEigenvectorCentrality(*graph, 100)
	stTypes := h.n4l.AnalyzeGraphBySTTypes(*graph)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"appointed_nodes":        appointed,
		"eigenvector_centrality": eigenvector,
		"st_type_analysis":       stTypes,
		"summary": map[string]interface{}{
			"total_appointed":  appointed.TotalAppointed,
			"max_pointers":     appointed.MaxPointers,
			"convergence":      eigenvector.Convergence,
			"top_influencer":   getTopInfluencer(eigenvector),
			"causal_chains":    len(stTypes.CausalChains),
			"containers":       len(stTypes.Containers),
		},
	})
}

// getTopInfluencer retourne le nœud le plus influent
func getTopInfluencer(result services.EigenvectorResult) string {
	if len(result.Centralities) > 0 {
		return result.Centralities[0].NodeLabel
	}
	return ""
}

// ============================================
// Contrawave Search - Collision de fronts d'onde
// ============================================

// HandleContrawaveSearch effectue une recherche par collision de fronts d'onde bidirectionnels
func (h *Handler) HandleContrawaveSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req services.ContrawaveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Erreur de parsing JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.CaseID == "" {
		http.Error(w, "case_id requis", http.StatusBadRequest)
		return
	}

	graph, err := h.cases.BuildGraphData(req.CaseID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	result, err := h.search.ContrawaveSearch(graph, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(result)
}

// ============================================
// Super-Nodes Detection - Équivalence fonctionnelle
// ============================================

// HandleSuperNodesDetection détecte les groupes de nœuds fonctionnellement équivalents
func (h *Handler) HandleSuperNodesDetection(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req services.SuperNodesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Erreur de parsing JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.CaseID == "" {
		http.Error(w, "case_id requis", http.StatusBadRequest)
		return
	}

	graph, err := h.cases.BuildGraphData(req.CaseID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	result, err := h.search.DetectSuperNodes(graph, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(result)
}

// ============================================
// Betweenness Centrality - Intermédiarité améliorée
// ============================================

// HandleBetweennessCentrality calcule la centralité d'intermédiarité pour tous les nœuds
func (h *Handler) HandleBetweennessCentrality(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	caseID := r.URL.Query().Get("case_id")
	if caseID == "" {
		http.Error(w, "case_id requis", http.StatusBadRequest)
		return
	}

	graph, err := h.cases.BuildGraphData(caseID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	result, err := h.search.CalculateBetweennessCentrality(graph)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(result)
}

// ============================================
// Analyse Avancée Combinée SSTorytime
// ============================================

// HandleSSTorytimeAnalysis effectue toutes les analyses SSTorytime en une requête
func (h *Handler) HandleSSTorytimeAnalysis(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	caseID := r.URL.Query().Get("case_id")
	if caseID == "" {
		http.Error(w, "case_id requis", http.StatusBadRequest)
		return
	}

	// Utiliser le graphe N4L comme source de données
	caseData, err := h.cases.GetCase(caseID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	var n4lContent string
	if caseData.N4LContent != "" {
		n4lContent = caseData.N4LContent
	} else {
		n4lContent = h.n4lGenerator.ExportCaseToN4L(caseData)
	}

	parsed := h.n4l.ParseForensicN4L(n4lContent, caseID)
	graph := &parsed.Graph

	// Effectuer toutes les analyses SSTorytime
	appointed := h.graphAnalyzer.FindAppointedNodes(*graph, 2)
	eigenvector := h.graphAnalyzer.CalculateEigenvectorCentrality(*graph, 100)
	stTypes := h.n4l.AnalyzeGraphBySTTypes(*graph)
	betweenness, _ := h.search.CalculateBetweennessCentrality(graph)

	// Détecter les super-nœuds avec paramètres par défaut
	superNodes, _ := h.search.DetectSuperNodes(graph, services.SuperNodesRequest{
		CaseID:              caseID,
		SimilarityThreshold: 0.7,
		MinGroupSize:        2,
	})

	// Construire la réponse combinée
	response := map[string]interface{}{
		"appointed_nodes":        appointed,
		"eigenvector_centrality": eigenvector,
		"st_type_analysis":       stTypes,
		"betweenness_centrality": betweenness,
		"super_nodes":            superNodes,
		"summary": map[string]interface{}{
			"total_nodes":           len(graph.Nodes),
			"total_edges":           len(graph.Edges),
			"appointed_count":       appointed.TotalAppointed,
			"max_pointers":          appointed.MaxPointers,
			"convergence":           eigenvector.Convergence,
			"top_influencer":        getTopInfluencer(eigenvector),
			"causal_chains":         len(stTypes.CausalChains),
			"containers":            len(stTypes.Containers),
			"super_node_groups":     0,
			"bridge_nodes":          0,
		},
	}

	// Ajouter les stats des super-nœuds si disponibles
	if superNodes != nil {
		response["summary"].(map[string]interface{})["super_node_groups"] = superNodes.TotalGroups
	}

	// Compter les nœuds ponts
	if betweenness != nil {
		bridges := 0
		for _, node := range betweenness.Centralities {
			if node.Role == "bridge" {
				bridges++
			}
		}
		response["summary"].(map[string]interface{})["bridge_nodes"] = bridges
	}

	json.NewEncoder(w).Encode(response)
}

// HandleConstrainedPaths recherche des chemins avec filtrage par type de relation
func (h *Handler) HandleConstrainedPaths(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req services.ConstrainedPathRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Requête invalide: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.CaseID == "" {
		http.Error(w, "case_id requis", http.StatusBadRequest)
		return
	}
	if req.FromNode == "" || req.ToNode == "" {
		http.Error(w, "from_node et to_node requis", http.StatusBadRequest)
		return
	}

	graph, err := h.cases.BuildGraphData(req.CaseID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	result, err := h.search.FindConstrainedPaths(graph, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(result)
}

// HandleDiracSearch effectue une recherche avec notation Dirac <cible|source>
func (h *Handler) HandleDiracSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req services.DiracRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Requête invalide: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.CaseID == "" {
		http.Error(w, "case_id requis", http.StatusBadRequest)
		return
	}
	if req.Query == "" {
		http.Error(w, "query requis (format: <cible|source>)", http.StatusBadRequest)
		return
	}

	graph, err := h.cases.BuildGraphData(req.CaseID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	result, err := h.search.SearchDirac(graph, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(result)
}

// HandleOrbits analyse les orbites (voisinage structuré) autour d'un nœud
func (h *Handler) HandleOrbits(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req services.OrbitRequest

	if r.Method == http.MethodGet {
		req.CaseID = r.URL.Query().Get("case_id")
		req.NodeID = r.URL.Query().Get("node_id")
		if maxLevel := r.URL.Query().Get("max_level"); maxLevel != "" {
			fmt.Sscanf(maxLevel, "%d", &req.MaxLevel)
		}
	} else {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Requête invalide: "+err.Error(), http.StatusBadRequest)
			return
		}
	}

	if req.CaseID == "" {
		http.Error(w, "case_id requis", http.StatusBadRequest)
		return
	}
	if req.NodeID == "" {
		http.Error(w, "node_id requis", http.StatusBadRequest)
		return
	}

	graph, err := h.cases.BuildGraphData(req.CaseID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	result, err := h.search.AnalyzeOrbits(graph, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(result)
}

// ============================================
// N4L Source Unique - Endpoints de migration et génération
// ============================================

// HandleN4LGenerate génère un fragment N4L à partir d'une structure
func (h *Handler) HandleN4LGenerate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	// Déterminer le type d'élément à générer depuis l'URL
	entityType := r.URL.Query().Get("type")
	if entityType == "" {
		// Essayer d'extraire du path
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) > 0 {
			entityType = parts[len(parts)-1]
		}
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Erreur lecture body: "+err.Error(), http.StatusBadRequest)
		return
	}

	var n4lFragment string

	switch entityType {
	case "entity":
		var entity models.Entity
		if err := json.Unmarshal(body, &entity); err != nil {
			http.Error(w, "Entité invalide: "+err.Error(), http.StatusBadRequest)
			return
		}
		n4lFragment = h.n4lGenerator.GenerateEntityN4L(entity)

	case "evidence":
		var evidence models.Evidence
		if err := json.Unmarshal(body, &evidence); err != nil {
			http.Error(w, "Preuve invalide: "+err.Error(), http.StatusBadRequest)
			return
		}
		n4lFragment = h.n4lGenerator.GenerateEvidenceN4L(evidence)

	case "timeline":
		var event models.Event
		if err := json.Unmarshal(body, &event); err != nil {
			http.Error(w, "Événement invalide: "+err.Error(), http.StatusBadRequest)
			return
		}
		n4lFragment = h.n4lGenerator.GenerateTimelineEventN4L(event)

	case "hypothesis":
		var hypothesis models.Hypothesis
		if err := json.Unmarshal(body, &hypothesis); err != nil {
			http.Error(w, "Hypothèse invalide: "+err.Error(), http.StatusBadRequest)
			return
		}
		n4lFragment = h.n4lGenerator.GenerateHypothesisN4L(hypothesis)

	case "relation":
		var req struct {
			Relation    models.Relation   `json:"relation"`
			EntityNames map[string]string `json:"entity_names"`
		}
		if err := json.Unmarshal(body, &req); err != nil {
			http.Error(w, "Relation invalide: "+err.Error(), http.StatusBadRequest)
			return
		}
		n4lFragment = h.n4lGenerator.GenerateRelationN4L(req.Relation, req.EntityNames)

	default:
		http.Error(w, "Type non supporté: "+entityType, http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"n4l_fragment": n4lFragment,
	})
}

// HandleN4LPatch applique un patch au contenu N4L d'un cas
func (h *Handler) HandleN4LPatch(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPatch && r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	caseID := r.URL.Query().Get("case_id")
	if caseID == "" {
		http.Error(w, "case_id requis", http.StatusBadRequest)
		return
	}

	var patch services.N4LPatch
	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
		http.Error(w, "Patch invalide: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Récupérer le cas et son contenu N4L actuel
	c, err := h.cases.GetCase(caseID)
	if err != nil {
		http.Error(w, "Cas non trouvé: "+err.Error(), http.StatusNotFound)
		return
	}

	// Appliquer le patch
	result := h.n4lGenerator.ApplyPatch(c.N4LContent, patch, caseID)

	if result.Success {
		// Sauvegarder le nouveau contenu N4L
		c.N4LContent = result.N4LContent
		if err := h.cases.UpdateCase(c); err != nil {
			http.Error(w, "Erreur sauvegarde: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	json.NewEncoder(w).Encode(result)
}

// HandleN4LMigrate migre un cas de la base de données vers le format N4L
func (h *Handler) HandleN4LMigrate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	caseID := r.URL.Query().Get("case_id")
	if caseID == "" {
		http.Error(w, "case_id requis", http.StatusBadRequest)
		return
	}

	// Récupérer le cas complet
	c, err := h.cases.GetCase(caseID)
	if err != nil {
		http.Error(w, "Cas non trouvé: "+err.Error(), http.StatusNotFound)
		return
	}

	// Générer le contenu N4L à partir des données existantes
	n4lContent := h.n4lGenerator.ExportCaseToN4L(c)

	// Valider le N4L généré
	errors := h.n4lGenerator.ValidateN4L(n4lContent)

	// Sauvegarder le contenu N4L dans le cas
	c.N4LContent = n4lContent
	if err := h.cases.UpdateCase(c); err != nil {
		http.Error(w, "Erreur sauvegarde: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Parser le N4L pour vérifier le round-trip
	parsed := h.n4l.ParseForensicN4L(n4lContent, caseID)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":           "migrated",
		"case_id":          caseID,
		"n4l_content":      n4lContent,
		"validation_errors": errors,
		"entity_count":     len(parsed.Entities),
		"evidence_count":   len(parsed.Evidence),
		"timeline_count":   len(parsed.Timeline),
		"hypothesis_count": len(parsed.Hypotheses),
		"relation_count":   len(parsed.Relations),
	})
}

// HandleN4LMigrateAll migre tous les cas vers le format N4L
func (h *Handler) HandleN4LMigrateAll(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	cases := h.cases.GetAllCases()
	results := []map[string]interface{}{}

	for _, c := range cases {
		// Générer le contenu N4L
		n4lContent := h.n4lGenerator.ExportCaseToN4L(c)

		// Sauvegarder
		c.N4LContent = n4lContent
		if err := h.cases.UpdateCase(c); err != nil {
			results = append(results, map[string]interface{}{
				"case_id": c.ID,
				"status":  "error",
				"error":   err.Error(),
			})
			continue
		}

		// Parser pour vérification
		parsed := h.n4l.ParseForensicN4L(n4lContent, c.ID)

		results = append(results, map[string]interface{}{
			"case_id":        c.ID,
			"case_name":      c.Name,
			"status":         "migrated",
			"entity_count":   len(parsed.Entities),
			"evidence_count": len(parsed.Evidence),
		})
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"total_cases":    len(cases),
		"migrated_count": len(results),
		"results":        results,
	})
}

// HandleN4LValidate valide le contenu N4L
func (h *Handler) HandleN4LValidate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Content string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Requête invalide: "+err.Error(), http.StatusBadRequest)
		return
	}

	errors := h.n4lGenerator.ValidateN4L(req.Content)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"valid":  len(errors) == 0,
		"errors": errors,
	})
}

// HandleN4LSync synchronise les données depuis le contenu N4L vers les structures du cas
func (h *Handler) HandleN4LSync(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	caseID := r.URL.Query().Get("case_id")
	if caseID == "" {
		http.Error(w, "case_id requis", http.StatusBadRequest)
		return
	}

	var req struct {
		N4LContent string `json:"n4l_content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Requête invalide: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Parser le contenu N4L
	parsed := h.n4l.ParseForensicN4L(req.N4LContent, caseID)

	// Récupérer le cas existant
	c, err := h.cases.GetCase(caseID)
	if err != nil {
		http.Error(w, "Cas non trouvé: "+err.Error(), http.StatusNotFound)
		return
	}

	// Mettre à jour les données du cas depuis le N4L parsé
	c.N4LContent = req.N4LContent
	c.Entities = parsed.Entities
	c.Evidence = parsed.Evidence
	c.Timeline = parsed.Timeline
	c.Hypotheses = parsed.Hypotheses

	// Sauvegarder
	if err := h.cases.UpdateCase(c); err != nil {
		http.Error(w, "Erreur sauvegarde: "+err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":           "synced",
		"case_id":          caseID,
		"entity_count":     len(parsed.Entities),
		"evidence_count":   len(parsed.Evidence),
		"timeline_count":   len(parsed.Timeline),
		"hypothesis_count": len(parsed.Hypotheses),
		"parsed_data":      parsed,
	})
}

// HandleGetCaseWithN4L retourne le cas avec ses données parsées depuis N4L
func (h *Handler) HandleGetCaseWithN4L(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	caseID := r.URL.Query().Get("case_id")
	if caseID == "" {
		http.Error(w, "case_id requis", http.StatusBadRequest)
		return
	}

	c, err := h.cases.GetCase(caseID)
	if err != nil {
		http.Error(w, "Cas non trouvé: "+err.Error(), http.StatusNotFound)
		return
	}

	// Si le cas a du contenu N4L, parser et retourner les données
	if c.N4LContent != "" {
		parsed := h.n4l.ParseForensicN4L(c.N4LContent, caseID)

		// Remplacer les données du cas par celles du N4L
		c.Entities = parsed.Entities
		c.Evidence = parsed.Evidence
		c.Timeline = parsed.Timeline
		c.Hypotheses = parsed.Hypotheses
	}

	json.NewEncoder(w).Encode(c)
}

// ============================================
// Simulation de Scénarios "What-If"
// ============================================

// HandleScenarios gère les opérations CRUD sur les scénarios
func (h *Handler) HandleScenarios(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	caseID := r.URL.Query().Get("case_id")
	if caseID == "" {
		http.Error(w, "case_id requis", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		// Lister tous les scénarios d'un cas
		scenarios := h.scenario.GetScenarios(caseID)
		json.NewEncoder(w).Encode(scenarios)

	case http.MethodPost:
		// Créer un nouveau scénario
		var req models.ScenarioSimulationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Erreur de parsing: "+err.Error(), http.StatusBadRequest)
			return
		}
		req.CaseID = caseID

		scenario, err := h.scenario.CreateScenario(caseID, req)
		if err != nil {
			http.Error(w, "Erreur création scénario: "+err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(scenario)

	default:
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
	}
}

// HandleScenario gère les opérations sur un scénario spécifique
func (h *Handler) HandleScenario(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	caseID := r.URL.Query().Get("case_id")
	scenarioID := r.URL.Query().Get("scenario_id")

	if caseID == "" || scenarioID == "" {
		http.Error(w, "case_id et scenario_id requis", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		scenario, err := h.scenario.GetScenario(caseID, scenarioID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(scenario)

	case http.MethodDelete:
		if err := h.scenario.DeleteScenario(caseID, scenarioID); err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Scénario supprimé",
		})

	default:
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
	}
}

// HandleScenarioSimulate lance une simulation IA pour un scénario
func (h *Handler) HandleScenarioSimulate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req struct {
		CaseID     string `json:"case_id"`
		ScenarioID string `json:"scenario_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.CaseID == "" || req.ScenarioID == "" {
		http.Error(w, "case_id et scenario_id requis", http.StatusBadRequest)
		return
	}

	analysis, err := h.scenario.SimulateWithAI(req.CaseID, req.ScenarioID)
	if err != nil {
		http.Error(w, "Erreur simulation: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Récupérer le scénario mis à jour
	scenario, _ := h.scenario.GetScenario(req.CaseID, req.ScenarioID)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"analysis": analysis,
		"scenario": scenario,
	})
}

// HandleScenarioSimulateStream lance une simulation IA pour un scénario en streaming
func (h *Handler) HandleScenarioSimulateStream(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	// Configurer les headers SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming non supporté", http.StatusInternalServerError)
		return
	}

	var req struct {
		CaseID     string `json:"case_id"`
		ScenarioID string `json:"scenario_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		fmt.Fprintf(w, "data: {\"error\": \"%s\"}\n\n", err.Error())
		flusher.Flush()
		return
	}

	if req.CaseID == "" || req.ScenarioID == "" {
		fmt.Fprintf(w, "data: {\"error\": \"case_id et scenario_id requis\"}\n\n")
		flusher.Flush()
		return
	}

	err := h.scenario.SimulateWithAIStream(req.CaseID, req.ScenarioID, func(chunk string, done bool) error {
		chunkJSON, _ := json.Marshal(map[string]interface{}{
			"chunk": chunk,
			"done":  done,
		})
		fmt.Fprintf(w, "data: %s\n\n", chunkJSON)
		flusher.Flush()
		return nil
	})

	if err != nil {
		errorJSON, _ := json.Marshal(map[string]string{"error": err.Error()})
		fmt.Fprintf(w, "data: %s\n\n", errorJSON)
		flusher.Flush()
	}
}

// HandleScenarioCompare compare deux scénarios
func (h *Handler) HandleScenarioCompare(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req struct {
		CaseID      string `json:"case_id"`
		ScenarioID1 string `json:"scenario1_id"`
		ScenarioID2 string `json:"scenario2_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.CaseID == "" || req.ScenarioID1 == "" || req.ScenarioID2 == "" {
		http.Error(w, "case_id, scenario1_id et scenario2_id requis", http.StatusBadRequest)
		return
	}

	comparison, err := h.scenario.CompareScenarios(req.CaseID, req.ScenarioID1, req.ScenarioID2)
	if err != nil {
		http.Error(w, "Erreur comparaison: "+err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(comparison)
}

// HandleScenarioPropagate propage les implications d'un scénario sur le graphe
func (h *Handler) HandleScenarioPropagate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req struct {
		CaseID     string `json:"case_id"`
		ScenarioID string `json:"scenario_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.CaseID == "" || req.ScenarioID == "" {
		http.Error(w, "case_id et scenario_id requis", http.StatusBadRequest)
		return
	}

	graph, err := h.scenario.PropagateImplications(req.CaseID, req.ScenarioID)
	if err != nil {
		http.Error(w, "Erreur propagation: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Récupérer le scénario mis à jour
	scenario, _ := h.scenario.GetScenario(req.CaseID, req.ScenarioID)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"graph":    graph,
		"scenario": scenario,
	})
}

// HandleScenarioGenerate génère automatiquement des scénarios avec l'IA
func (h *Handler) HandleScenarioGenerate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	caseID := r.URL.Query().Get("case_id")
	if caseID == "" {
		http.Error(w, "case_id requis", http.StatusBadRequest)
		return
	}

	scenarios, err := h.scenario.GenerateScenariosWithAI(caseID)
	if err != nil {
		http.Error(w, "Erreur génération: "+err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"scenarios": scenarios,
		"count":     len(scenarios),
	})
}

// ============================================
// Détection d'Anomalies
// ============================================

// HandleAnomalies gère les opérations sur les anomalies
func (h *Handler) HandleAnomalies(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	caseID := r.URL.Query().Get("case_id")
	if caseID == "" {
		http.Error(w, "case_id requis", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		// Lister toutes les anomalies d'un cas
		anomalies := h.anomaly.GetAnomalies(caseID)
		json.NewEncoder(w).Encode(anomalies)

	default:
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
	}
}

// HandleAnomalyDetect lance une détection d'anomalies
func (h *Handler) HandleAnomalyDetect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req struct {
		CaseID string `json:"case_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.CaseID == "" {
		http.Error(w, "case_id requis", http.StatusBadRequest)
		return
	}

	result, err := h.anomaly.DetectAnomalies(req.CaseID)
	if err != nil {
		http.Error(w, "Erreur détection: "+err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(result)
}

// HandleAnomaly gère les opérations sur une anomalie spécifique
func (h *Handler) HandleAnomaly(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	caseID := r.URL.Query().Get("case_id")
	anomalyID := r.URL.Query().Get("anomaly_id")

	if caseID == "" || anomalyID == "" {
		http.Error(w, "case_id et anomaly_id requis", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		anomaly, err := h.anomaly.GetAnomaly(caseID, anomalyID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(anomaly)

	default:
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
	}
}

// HandleAnomalyAcknowledge marque une anomalie comme acquittée
func (h *Handler) HandleAnomalyAcknowledge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req struct {
		CaseID    string `json:"case_id"`
		AnomalyID string `json:"anomaly_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.CaseID == "" || req.AnomalyID == "" {
		http.Error(w, "case_id et anomaly_id requis", http.StatusBadRequest)
		return
	}

	if err := h.anomaly.AcknowledgeAnomaly(req.CaseID, req.AnomalyID); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Anomalie acquittée",
	})
}

// HandleAnomalyExplain génère une explication IA pour une anomalie
func (h *Handler) HandleAnomalyExplain(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req struct {
		CaseID    string `json:"case_id"`
		AnomalyID string `json:"anomaly_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.CaseID == "" || req.AnomalyID == "" {
		http.Error(w, "case_id et anomaly_id requis", http.StatusBadRequest)
		return
	}

	explanation, err := h.anomaly.ExplainAnomaly(req.CaseID, req.AnomalyID)
	if err != nil {
		http.Error(w, "Erreur explication: "+err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"explanation": explanation,
	})
}

// HandleAnomalyStatistics retourne les statistiques d'anomalies
func (h *Handler) HandleAnomalyStatistics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	caseID := r.URL.Query().Get("case_id")
	if caseID == "" {
		http.Error(w, "case_id requis", http.StatusBadRequest)
		return
	}

	stats := h.anomaly.GetStatistics(caseID)
	json.NewEncoder(w).Encode(stats)
}

// HandleAnomalyAlerts gère les alertes d'anomalies
func (h *Handler) HandleAnomalyAlerts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	caseID := r.URL.Query().Get("case_id")
	if caseID == "" {
		http.Error(w, "case_id requis", http.StatusBadRequest)
		return
	}

	unreadOnly := r.URL.Query().Get("unread_only") == "true"

	switch r.Method {
	case http.MethodGet:
		alerts := h.anomaly.GetAlerts(caseID, unreadOnly)
		json.NewEncoder(w).Encode(alerts)

	default:
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
	}
}

// HandleAnomalyAlertRead marque une alerte comme lue
func (h *Handler) HandleAnomalyAlertRead(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req struct {
		CaseID  string `json:"case_id"`
		AlertID string `json:"alert_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.CaseID == "" || req.AlertID == "" {
		http.Error(w, "case_id et alert_id requis", http.StatusBadRequest)
		return
	}

	if err := h.anomaly.MarkAlertRead(req.CaseID, req.AlertID); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Alerte marquée comme lue",
	})
}

// HandleAnomalyConfig gère la configuration de détection d'anomalies
func (h *Handler) HandleAnomalyConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	caseID := r.URL.Query().Get("case_id")
	if caseID == "" {
		http.Error(w, "case_id requis", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		config := h.anomaly.GetConfig(caseID)
		json.NewEncoder(w).Encode(config)

	case http.MethodPut:
		var config models.AnomalyDetectionConfig
		if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		config.CaseID = caseID

		if err := h.anomaly.UpdateConfig(&config); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Configuration mise à jour",
		})

	default:
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
	}
}
