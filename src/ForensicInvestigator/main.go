package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"forensicinvestigator/data"
	"forensicinvestigator/internal/handlers"
	"forensicinvestigator/internal/services"
)

// Configuration des services distants (SPARK GB10).
//
// Le port 8001 sert le LLM via vLLM (Qwen3.8-27B-FP8, contexte 128k, MTP
// self-speculation, prefix caching), le port 8002 les embeddings
// (multilingual-e5-base, 768 dimensions).
//
// Le SPARK est un serveur partagé: d'autres applications consomment les mêmes
// endpoints. Sa configuration est donc prise telle quelle, sans rien exiger de lui.
//
// Concrètement, vLLM n'accepte que la valeur de son --served-model-name. Ici c'est
// "Qwen3.5-9B", un nom hérité qui désigne en réalité Qwen/Qwen3.8-27B-FP8; d'autres
// clients s'appuient dessus, il ne doit pas être renommé. Plutôt que de coder cet
// identifiant en dur, on le découvre au démarrage via /v1/models: l'application
// suit le SPARK quelle que soit sa configuration. LLM_MODEL permet de le forcer,
// LLM_BASE_URL de changer de serveur.
const (
	defaultLLMBaseURL = "http://86.204.69.30:8001"

	// Utilisé uniquement si la découverte échoue ET que LLM_MODEL n'est pas défini:
	// l'application démarre alors en mode dégradé plutôt que sans modèle du tout.
	fallbackLLMModel = "Qwen3.5-9B"
)

// getenv retourne la variable d'environnement key ou fallback si elle est vide.
func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// resolveLLMModel détermine l'identifiant de modèle à envoyer au serveur.
//
// LLM_MODEL est prioritaire; sinon le modèle servi est découvert via /v1/models.
func resolveLLMModel(baseURL string) string {
	if forced := os.Getenv("LLM_MODEL"); forced != "" {
		log.Printf("LLM: modèle forcé par LLM_MODEL: %s", forced)
		return forced
	}

	served, err := services.DiscoverServedModel(baseURL)
	if err != nil {
		log.Printf("LLM: découverte impossible (%v), repli sur %q", err, fallbackLLMModel)
		return fallbackLLMModel
	}

	log.Printf("LLM: modèle découvert: %s", served)
	return served.ID
}

func main() {
	log.Println("ForensicInvestigator - Démarrage du serveur...")

	// Initialiser les services (LLM distant - vLLM sur SPARK)
	llmBaseURL := getenv("LLM_BASE_URL", defaultLLMBaseURL)
	llmModel := resolveLLMModel(llmBaseURL)
	log.Printf("LLM: %s (modèle: %s)", llmBaseURL, llmModel)

	ollamaService := services.NewOllamaService(llmBaseURL, llmModel)
	caseService := services.NewCaseService()
	n4lService := services.NewN4LService()

	// Charger les données de démonstration
	demoCases := data.GetDemoCases()
	count := caseService.LoadDemoCases(demoCases)
	log.Printf("Chargement de %d affaires de démonstration", count)

	// Créer le handler principal
	handler := handlers.NewHandler(ollamaService, caseService, n4lService)

	// Indexer les affaires en arrière-plan: le serveur accepte les requêtes
	// immédiatement, la recherche se replie sur BM25 tant que l'index n'est pas prêt.
	go handler.ReindexAllCases()

	// Routes API
	http.HandleFunc("/api/cases", handler.HandleCases)
	http.HandleFunc("/api/cases/", handler.HandleCase)
	http.HandleFunc("/api/entities", handler.HandleEntities)
	http.HandleFunc("/api/entities/update", handler.HandleUpdateEntity)
	http.HandleFunc("/api/entities/delete", handler.HandleDeleteEntity)
	http.HandleFunc("/api/relations", handler.HandleRelations)
	http.HandleFunc("/api/evidence", handler.HandleEvidence)
	http.HandleFunc("/api/evidence/update", handler.HandleUpdateEvidence)
	http.HandleFunc("/api/evidence/delete", handler.HandleDeleteEvidence)
	http.HandleFunc("/api/timeline", handler.HandleTimeline)
	http.HandleFunc("/api/timeline/update", handler.HandleUpdateEvent)
	http.HandleFunc("/api/timeline/delete", handler.HandleDeleteEvent)
	http.HandleFunc("/api/hypotheses", handler.HandleHypotheses)
	http.HandleFunc("/api/hypotheses/update", handler.HandleUpdateHypothesis)
	http.HandleFunc("/api/hypotheses/delete", handler.HandleDeleteHypothesis)
	http.HandleFunc("/api/hypotheses/analyze", handler.HandleAnalyzeHypothesis)
	http.HandleFunc("/api/analyze", handler.HandleAnalyze)
	http.HandleFunc("/api/analyze/contradictions", handler.HandleContradictions)
	http.HandleFunc("/api/analyze/questions", handler.HandleQuestions)
	http.HandleFunc("/api/analyze/path", handler.HandleAnalyzePath)
	http.HandleFunc("/api/chat", handler.HandleChat)
	http.HandleFunc("/api/chat/stream", handler.HandleChatStream)
	http.HandleFunc("/api/analyze/stream", handler.HandleAnalyzeStream)
	http.HandleFunc("/api/analyze/contradictions/stream", handler.HandleContradictionsStream)
	http.HandleFunc("/api/contradictions/detect/stream", handler.HandleContradictionsDetectStream)
	http.HandleFunc("/api/questions/generate/stream", handler.HandleQuestionsGenerateStream)
	http.HandleFunc("/api/hypotheses/generate/stream", handler.HandleHypothesesGenerateStream)
	http.HandleFunc("/api/n4l/parse", handler.HandleN4LParse)
	http.HandleFunc("/api/n4l/export", handler.HandleN4LExport)
	http.HandleFunc("/api/n4l/convert", handler.HandleConvertToN4L)
	http.HandleFunc("/api/n4l/generate", handler.HandleN4LGenerate)
	http.HandleFunc("/api/n4l/patch", handler.HandleN4LPatch)
	http.HandleFunc("/api/n4l/migrate", handler.HandleN4LMigrate)
	http.HandleFunc("/api/n4l/migrate/all", handler.HandleN4LMigrateAll)
	http.HandleFunc("/api/n4l/validate", handler.HandleN4LValidate)
	http.HandleFunc("/api/n4l/sync", handler.HandleN4LSync)
	http.HandleFunc("/api/case/n4l", handler.HandleGetCaseWithN4L)
	http.HandleFunc("/api/graph", handler.HandleGraph)
	http.HandleFunc("/api/cross-case/scan", handler.HandleCrossCase)
	http.HandleFunc("/api/cross-case/analyze", handler.HandleCrossCaseAnalyze)
	http.HandleFunc("/api/cross-case/analyze/stream", handler.HandleCrossCaseAnalyzeStream)
	http.HandleFunc("/api/cross-case/graph", handler.HandleCrossCaseGraph)

	// Routes HRM (Hypothetical Reasoning Model)
	http.HandleFunc("/api/hrm/status", handler.HandleHRMStatus)
	http.HandleFunc("/api/hrm/reason", handler.HandleHRMReason)
	http.HandleFunc("/api/hrm/reason/stream", handler.HandleHRMReasonStream)
	http.HandleFunc("/api/hrm/verify-hypothesis", handler.HandleHRMVerifyHypothesis)
	http.HandleFunc("/api/hrm/contradictions", handler.HandleHRMContradictions)
	http.HandleFunc("/api/hrm/cross-case", handler.HandleHRMCrossCase)

	// Routes Recherche Hybride (BM25 + sémantique via embeddings SPARK)
	http.HandleFunc("/api/search/hybrid", handler.HandleHybridSearch)
	http.HandleFunc("/api/search/quick", handler.HandleQuickSearch)

	// Routes Configuration des Prompts
	http.HandleFunc("/api/config/prompts", handler.HandleConfigPrompts)
	http.HandleFunc("/api/config/prompts/", handler.HandleConfigPrompt)
	http.HandleFunc("/api/config/reload", handler.HandleConfigReload)

	// Routes Analyse de Graphe
	http.HandleFunc("/api/graph/clusters", handler.HandleFindClusters)
	http.HandleFunc("/api/graph/paths", handler.HandleFindPaths)
	http.HandleFunc("/api/graph/layered", handler.HandleLayeredGraph)
	http.HandleFunc("/api/graph/expansion-cone", handler.HandleExpansionCone)
	http.HandleFunc("/api/graph/density", handler.HandleDensityMap)
	http.HandleFunc("/api/graph/temporal-patterns", handler.HandleTemporalPatterns)
	http.HandleFunc("/api/graph/consistency", handler.HandleCheckConsistency)
	http.HandleFunc("/api/graph/analyze-complete", handler.HandleGraphAnalyzeComplete)

	// Routes Mode Investigation
	http.HandleFunc("/api/investigation/start", handler.HandleStartInvestigation)
	http.HandleFunc("/api/investigation/suggestions", handler.HandleInvestigationSuggestions)
	http.HandleFunc("/api/investigation/analyze", handler.HandleInvestigationAnalyze)
	http.HandleFunc("/api/investigation/analyze/stream", handler.HandleInvestigationAnalyzeStream)

	// Routes Notebook (centralisateur d'analyses IA)
	http.HandleFunc("/api/notebook", handler.HandleNotebook)
	http.HandleFunc("/api/notebook/stats", handler.HandleNotebookStats)
	http.HandleFunc("/api/notes", handler.HandleNotes)
	http.HandleFunc("/api/note", handler.HandleNote)
	http.HandleFunc("/api/note/pin", handler.HandleNotePin)
	http.HandleFunc("/api/note/favorite", handler.HandleNoteFavorite)
	http.HandleFunc("/api/note/tag", handler.HandleNoteTag)

	// Routes Advanced Graph Analysis (Inspired by SSTorytime)
	http.HandleFunc("/api/graph/cone-search", handler.HandleConeSearch)
	http.HandleFunc("/api/graph/dirac-paths", handler.HandleDiracPathSearch)
	http.HandleFunc("/api/graph/appointed-nodes", handler.HandleAppointedNodes)
	http.HandleFunc("/api/graph/eigenvector-centrality", handler.HandleEigenvectorCentrality)
	http.HandleFunc("/api/graph/st-type-analysis", handler.HandleSTTypeAnalysis)
	http.HandleFunc("/api/graph/advanced-analysis", handler.HandleAdvancedGraphAnalysis)

	// Routes SSTorytime Avancées (Nouvelles fonctionnalités)
	http.HandleFunc("/api/graph/contrawave", handler.HandleContrawaveSearch)
	http.HandleFunc("/api/graph/super-nodes", handler.HandleSuperNodesDetection)
	http.HandleFunc("/api/graph/betweenness-centrality", handler.HandleBetweennessCentrality)
	http.HandleFunc("/api/graph/sstorytime-analysis", handler.HandleSSTorytimeAnalysis)
	http.HandleFunc("/api/graph/constrained-paths", handler.HandleConstrainedPaths)
	http.HandleFunc("/api/graph/dirac-search", handler.HandleDiracSearch)
	http.HandleFunc("/api/graph/orbits", handler.HandleOrbits)

	// Routes Simulation de Scénarios "What-If"
	http.HandleFunc("/api/scenarios", handler.HandleScenarios)
	http.HandleFunc("/api/scenarios/generate", handler.HandleScenarioGenerate)
	http.HandleFunc("/api/scenario", handler.HandleScenario)
	http.HandleFunc("/api/scenario/simulate", handler.HandleScenarioSimulate)
	http.HandleFunc("/api/scenario/simulate/stream", handler.HandleScenarioSimulateStream)
	http.HandleFunc("/api/scenario/compare", handler.HandleScenarioCompare)
	http.HandleFunc("/api/scenario/propagate", handler.HandleScenarioPropagate)

	// Routes Détection d'Anomalies
	http.HandleFunc("/api/anomalies", handler.HandleAnomalies)
	http.HandleFunc("/api/anomaly", handler.HandleAnomaly)
	http.HandleFunc("/api/anomaly/detect", handler.HandleAnomalyDetect)
	http.HandleFunc("/api/anomaly/acknowledge", handler.HandleAnomalyAcknowledge)
	http.HandleFunc("/api/anomaly/explain", handler.HandleAnomalyExplain)
	http.HandleFunc("/api/anomaly/statistics", handler.HandleAnomalyStatistics)
	http.HandleFunc("/api/anomaly/alerts", handler.HandleAnomalyAlerts)
	http.HandleFunc("/api/anomaly/alert/read", handler.HandleAnomalyAlertRead)
	http.HandleFunc("/api/anomaly/config", handler.HandleAnomalyConfig)

	// Déterminer le répertoire de travail
	execPath, err := os.Executable()
	if err != nil {
		execPath = "."
	}
	workDir := filepath.Dir(execPath)

	// Vérifier si static existe dans le répertoire courant (dev mode)
	if _, err := os.Stat("static"); err == nil {
		workDir = "."
	}

	// Fichiers statiques
	staticPath := filepath.Join(workDir, "static")
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticPath))))

	// Page principale
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, filepath.Join(staticPath, "index.html"))
	})

	port := ":8082"
	log.Printf("Serveur démarré sur http://localhost%s", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
