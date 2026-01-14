package models

import "time"

// Case représente une affaire d'investigation
type Case struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Type        string    `json:"type"` // homicide, vol, fraude, etc.
	Status      string    `json:"status"` // en_cours, resolu, classe
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Entities    []Entity  `json:"entities"`
	Evidence    []Evidence `json:"evidence"`
	Timeline    []Event   `json:"timeline"`
	Hypotheses  []Hypothesis `json:"hypotheses"`
	N4LContent  string    `json:"n4l_content"`
}

// EntityType définit les types d'entités possibles
type EntityType string

const (
	EntityPerson   EntityType = "personne"
	EntityPlace    EntityType = "lieu"
	EntityObject   EntityType = "objet"
	EntityEvent    EntityType = "evenement"
	EntityOrg      EntityType = "organisation"
	EntityDocument EntityType = "document"
)

// EntityRole définit les rôles possibles d'une entité
type EntityRole string

const (
	RoleVictim    EntityRole = "victime"
	RoleSuspect   EntityRole = "suspect"
	RoleWitness   EntityRole = "temoin"
	RoleInvestigator EntityRole = "enqueteur"
	RoleOther     EntityRole = "autre"
)

// Entity représente une entité dans l'enquête
type Entity struct {
	ID          string            `json:"id"`
	CaseID      string            `json:"case_id"`
	Name        string            `json:"name"`
	Type        EntityType        `json:"type"`
	Role        EntityRole        `json:"role"`
	Description string            `json:"description"`
	Attributes  map[string]string `json:"attributes"`
	Relations   []Relation        `json:"relations"`
	CreatedAt   time.Time         `json:"created_at"`
}

// Relation représente un lien entre deux entités
type Relation struct {
	ID        string    `json:"id"`
	FromID    string    `json:"from_id"`
	ToID      string    `json:"to_id"`
	Type      string    `json:"type"` // connaît, possède, était_à, etc.
	Label     string    `json:"label"`
	Context   string    `json:"context"`
	Timestamp time.Time `json:"timestamp,omitempty"`
	Verified  bool      `json:"verified"`
	Source    string    `json:"source"` // témoignage, preuve, hypothèse
}

// EvidenceType définit les types de preuves
type EvidenceType string

const (
	EvidencePhysical    EvidenceType = "physique"
	EvidenceTestimonial EvidenceType = "testimoniale"
	EvidenceDocumentary EvidenceType = "documentaire"
	EvidenceDigital     EvidenceType = "numerique"
	EvidenceForensic    EvidenceType = "forensique"
)

// Evidence représente une preuve ou un indice
type Evidence struct {
	ID           string       `json:"id"`
	CaseID       string       `json:"case_id"`
	Name         string       `json:"name"`
	Type         EvidenceType `json:"type"`
	Description  string       `json:"description"`
	Location     string       `json:"location"`
	CollectedAt  time.Time    `json:"collected_at"`
	CollectedBy  string       `json:"collected_by"`
	ChainOfCustody []string   `json:"chain_of_custody"`
	Reliability  int          `json:"reliability"` // 1-10
	LinkedEntities []string   `json:"linked_entities"`
	Notes        string       `json:"notes"`
}

// Event représente un événement dans la timeline
type Event struct {
	ID          string    `json:"id"`
	CaseID      string    `json:"case_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Timestamp   time.Time `json:"timestamp"`
	EndTime     *time.Time `json:"end_time,omitempty"`
	Location    string    `json:"location"`
	Entities    []string  `json:"entities"` // IDs des entités impliquées
	Evidence    []string  `json:"evidence"` // IDs des preuves liées
	Verified    bool      `json:"verified"`
	Source      string    `json:"source"`
	Importance  string    `json:"importance"` // high, medium, low
}

// HypothesisStatus définit le statut d'une hypothèse
type HypothesisStatus string

const (
	HypothesisPending   HypothesisStatus = "en_attente"
	HypothesisSupported HypothesisStatus = "corroboree"
	HypothesisRefuted   HypothesisStatus = "refutee"
	HypothesisPartial   HypothesisStatus = "partielle"
)

// Hypothesis représente une hypothèse d'investigation
type Hypothesis struct {
	ID              string           `json:"id"`
	CaseID          string           `json:"case_id"`
	Title           string           `json:"title"`
	Description     string           `json:"description"`
	Status          HypothesisStatus `json:"status"`
	ConfidenceLevel int              `json:"confidence_level"` // 0-100
	SupportingEvidence []string      `json:"supporting_evidence"`
	ContradictingEvidence []string   `json:"contradicting_evidence"`
	Questions       []string         `json:"questions"` // Questions à investiguer
	GeneratedBy     string           `json:"generated_by"` // "user" ou "ai"
	CreatedAt       time.Time        `json:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at"`
}

// GraphData représente les données du graphe pour la visualisation
type GraphData struct {
	Nodes []GraphNode `json:"nodes"`
	Edges []GraphEdge `json:"edges"`
}

// GraphNode représente un nœud du graphe
type GraphNode struct {
	ID      string            `json:"id"`
	Label   string            `json:"label"`
	Type    string            `json:"type"`
	Role    string            `json:"role,omitempty"`
	Context string            `json:"context,omitempty"`
	Data    map[string]string `json:"data,omitempty"`
}

// GraphEdge représente une arête du graphe
type GraphEdge struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Label   string `json:"label"`
	Type    string `json:"type"`
	Context string `json:"context,omitempty"`
}

// AnalysisRequest représente une demande d'analyse
type AnalysisRequest struct {
	CaseID    string    `json:"case_id"`
	Query     string    `json:"query"`
	GraphData GraphData `json:"graph_data,omitempty"`
	Context   string    `json:"context,omitempty"`
}

// AnalysisResponse représente une réponse d'analyse
type AnalysisResponse struct {
	Analysis    string   `json:"analysis"`
	Suggestions []string `json:"suggestions,omitempty"`
	Questions   []string `json:"questions,omitempty"`
	Confidence  int      `json:"confidence,omitempty"`
}

// CrossCaseMatchType définit les types de correspondances inter-affaires
type CrossCaseMatchType string

const (
	MatchEntity   CrossCaseMatchType = "entity"
	MatchLocation CrossCaseMatchType = "location"
	MatchModus    CrossCaseMatchType = "modus"
	MatchTemporal CrossCaseMatchType = "temporal"
)

// CrossCaseMatch représente une correspondance entre deux affaires
type CrossCaseMatch struct {
	ID              string             `json:"id"`
	CurrentCaseID   string             `json:"current_case_id"`
	CurrentCaseName string             `json:"current_case_name"`
	OtherCaseID     string             `json:"other_case_id"`
	OtherCaseName   string             `json:"other_case_name"`
	MatchType       CrossCaseMatchType `json:"match_type"`
	Confidence      int                `json:"confidence"` // 0-100
	Description     string             `json:"description"`
	CurrentElement  string             `json:"current_element"`
	OtherElement    string             `json:"other_element"`
	Details         map[string]string  `json:"details,omitempty"`
}

// CrossCaseResult représente le résultat d'une analyse inter-affaires
type CrossCaseResult struct {
	Matches   []CrossCaseMatch `json:"matches"`
	Summary   string           `json:"summary"`
	Alerts    []string         `json:"alerts"`
	GraphData *GraphData       `json:"graph_data,omitempty"`
}

// NoteType définit les types de notes (source de l'analyse IA)
type NoteType string

const (
	NoteTypeGraphAnalysis       NoteType = "graph_analysis"
	NoteTypeHypothesis          NoteType = "hypothesis"
	NoteTypeContradiction       NoteType = "contradiction"
	NoteTypeQuestion            NoteType = "question"
	NoteTypeEntityAnalysis      NoteType = "entity_analysis"
	NoteTypeEvidenceAnalysis    NoteType = "evidence_analysis"
	NoteTypeComparisonAnalysis  NoteType = "comparison_analysis"
	NoteTypePathAnalysis        NoteType = "path_analysis"
	NoteTypeHRMReasoning        NoteType = "hrm_reasoning"
	NoteTypeInvestigation       NoteType = "investigation"
	NoteTypeCrossCaseAnalysis   NoteType = "cross_case_analysis"
	NoteTypeChat                NoteType = "chat"
	NoteTypeManual              NoteType = "manual"
)

// Note représente une note sauvegardée dans le notebook
type Note struct {
	ID          string    `json:"id"`
	CaseID      string    `json:"case_id"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	Type        NoteType  `json:"type"`
	Tags        []string  `json:"tags,omitempty"`
	Context     string    `json:"context,omitempty"`   // Contexte additionnel (ex: entités analysées)
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	IsPinned    bool      `json:"is_pinned"`
	IsFavorite  bool      `json:"is_favorite"`
}

// Notebook représente le carnet de notes d'une affaire
type Notebook struct {
	CaseID    string    `json:"case_id"`
	CaseName  string    `json:"case_name"`
	Notes     []Note    `json:"notes"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ============================================
// Simulation de Scénarios "What-If"
// ============================================

// ScenarioStatus définit le statut d'un scénario
type ScenarioStatus string

const (
	ScenarioActive    ScenarioStatus = "active"
	ScenarioArchived  ScenarioStatus = "archived"
	ScenarioComparing ScenarioStatus = "comparing"
)

// Scenario représente un scénario "What-If"
type Scenario struct {
	ID                string                 `json:"id"`
	CaseID            string                 `json:"case_id"`
	Name              string                 `json:"name"`
	Description       string                 `json:"description"`
	Assumption        string                 `json:"assumption"`        // "X est coupable", "Y était présent", etc.
	AssumptionType    string                 `json:"assumption_type"`   // guilt, presence, motive, timeline, relation
	TargetEntityID    string                 `json:"target_entity_id"`  // Entité principale du scénario
	PlausibilityScore int                    `json:"plausibility_score"` // 0-100
	Status            ScenarioStatus         `json:"status"`
	Implications      []ScenarioImplication  `json:"implications"`      // Conséquences sur le graphe
	SupportingFacts   []string               `json:"supporting_facts"`  // Faits qui supportent
	ContradictingFacts []string              `json:"contradicting_facts"` // Faits qui contredisent
	ModifiedGraph     *GraphData             `json:"modified_graph,omitempty"` // Graphe modifié pour ce scénario
	AIAnalysis        string                 `json:"ai_analysis,omitempty"`    // Analyse IA du scénario
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
	CreatedBy         string                 `json:"created_by"` // "user" ou "ai"
}

// ScenarioImplication représente une implication d'un scénario
type ScenarioImplication struct {
	ID          string `json:"id"`
	Type        string `json:"type"`        // add_relation, remove_relation, change_role, add_motive, timeline_conflict
	Description string `json:"description"`
	EntityID    string `json:"entity_id,omitempty"`
	RelationID  string `json:"relation_id,omitempty"`
	Impact      string `json:"impact"`      // high, medium, low
	Confidence  int    `json:"confidence"`  // 0-100 - certitude de l'implication
	AutoDetected bool  `json:"auto_detected"` // true si détecté automatiquement
}

// ScenarioComparison représente une comparaison entre scénarios
type ScenarioComparison struct {
	Scenario1ID       string                `json:"scenario1_id"`
	Scenario2ID       string                `json:"scenario2_id"`
	Scenario1Name     string                `json:"scenario1_name"`
	Scenario2Name     string                `json:"scenario2_name"`
	CommonFacts       []string              `json:"common_facts"`       // Faits partagés
	DifferentFacts    []ScenarioDifference  `json:"different_facts"`    // Divergences
	PlausibilityDelta int                   `json:"plausibility_delta"` // Différence de score
	Recommendation    string                `json:"recommendation"`     // Recommandation IA
	GraphDifferences  *GraphDifference      `json:"graph_differences,omitempty"`
}

// ScenarioDifference représente une différence entre deux scénarios
type ScenarioDifference struct {
	Aspect         string `json:"aspect"`          // Ce qui diffère
	Scenario1Value string `json:"scenario1_value"` // Valeur dans scénario 1
	Scenario2Value string `json:"scenario2_value"` // Valeur dans scénario 2
	Significance   string `json:"significance"`    // high, medium, low
}

// GraphDifference représente les différences de graphe entre scénarios
type GraphDifference struct {
	AddedNodes   []GraphNode `json:"added_nodes"`
	RemovedNodes []GraphNode `json:"removed_nodes"`
	AddedEdges   []GraphEdge `json:"added_edges"`
	RemovedEdges []GraphEdge `json:"removed_edges"`
	ModifiedNodes []GraphNode `json:"modified_nodes"`
}

// ScenarioSimulationRequest représente une demande de simulation
type ScenarioSimulationRequest struct {
	CaseID         string `json:"case_id"`
	Assumption     string `json:"assumption"`
	AssumptionType string `json:"assumption_type"`
	TargetEntityID string `json:"target_entity_id,omitempty"`
}

// ============================================
// Détection d'Anomalies
// ============================================

// AnomalyType définit les types d'anomalies
type AnomalyType string

const (
	AnomalyTimeline      AnomalyType = "timeline"      // Comportement inhabituel dans la timeline
	AnomalyFinancial     AnomalyType = "financial"     // Transaction financière suspecte
	AnomalyCommunication AnomalyType = "communication" // Pattern de communication anormal
	AnomalyBehavior      AnomalyType = "behavior"      // Comportement inhabituel
	AnomalyLocation      AnomalyType = "location"      // Présence anormale à un lieu
	AnomalyRelation      AnomalyType = "relation"      // Relation suspecte
	AnomalyPattern       AnomalyType = "pattern"       // Pattern récurrent suspect
)

// AnomalySeverity définit la sévérité d'une anomalie
type AnomalySeverity string

const (
	SeverityCritical AnomalySeverity = "critical"
	SeverityHigh     AnomalySeverity = "high"
	SeverityMedium   AnomalySeverity = "medium"
	SeverityLow      AnomalySeverity = "low"
	SeverityInfo     AnomalySeverity = "info"
)

// Anomaly représente une anomalie détectée
type Anomaly struct {
	ID              string          `json:"id"`
	CaseID          string          `json:"case_id"`
	Type            AnomalyType     `json:"type"`
	Severity        AnomalySeverity `json:"severity"`
	Title           string          `json:"title"`
	Description     string          `json:"description"`
	DetectedAt      time.Time       `json:"detected_at"`
	EntityIDs       []string        `json:"entity_ids,omitempty"`   // Entités impliquées
	EvidenceIDs     []string        `json:"evidence_ids,omitempty"` // Preuves liées
	EventIDs        []string        `json:"event_ids,omitempty"`    // Événements liés
	Confidence      int             `json:"confidence"`             // 0-100
	IsNew           bool            `json:"is_new"`                 // Nouvelle anomalie (non vue)
	IsAcknowledged  bool            `json:"is_acknowledged"`        // Acquittée par l'utilisateur
	RelatedAnomalies []string       `json:"related_anomalies,omitempty"` // IDs d'anomalies liées
	Details         map[string]interface{} `json:"details,omitempty"` // Détails spécifiques au type
	AIExplanation   string          `json:"ai_explanation,omitempty"`
	Recommendations []string        `json:"recommendations,omitempty"`
}

// TimelineAnomaly détails spécifiques pour anomalie timeline
type TimelineAnomaly struct {
	EventID          string    `json:"event_id"`
	ExpectedBehavior string    `json:"expected_behavior"`
	ActualBehavior   string    `json:"actual_behavior"`
	TimeGap          string    `json:"time_gap,omitempty"`     // Écart temporel anormal
	Contradiction    string    `json:"contradiction,omitempty"` // Contradiction avec autre événement
}

// FinancialAnomaly détails spécifiques pour anomalie financière
type FinancialAnomaly struct {
	Amount           float64   `json:"amount"`
	Currency         string    `json:"currency"`
	TransactionType  string    `json:"transaction_type"` // transfer, cash, crypto, etc.
	FromEntity       string    `json:"from_entity"`
	ToEntity         string    `json:"to_entity"`
	Timestamp        time.Time `json:"timestamp"`
	DeviationPercent float64   `json:"deviation_percent"` // Écart par rapport à la normale
	PatternBreak     string    `json:"pattern_break,omitempty"` // Description du pattern rompu
}

// CommunicationAnomaly détails spécifiques pour anomalie de communication
type CommunicationAnomaly struct {
	FromEntity       string    `json:"from_entity"`
	ToEntity         string    `json:"to_entity"`
	CommunicationType string   `json:"communication_type"` // call, message, email, meeting
	Frequency        int       `json:"frequency"`          // Nombre de communications
	TimeRange        string    `json:"time_range"`         // Période concernée
	UnusualPattern   string    `json:"unusual_pattern"`    // Description du pattern anormal
	NormalBaseline   string    `json:"normal_baseline"`    // Comportement normal de référence
}

// AnomalyAlert représente une alerte automatique
type AnomalyAlert struct {
	ID          string          `json:"id"`
	CaseID      string          `json:"case_id"`
	AnomalyID   string          `json:"anomaly_id"`
	AlertType   string          `json:"alert_type"`   // immediate, daily_digest, threshold
	Message     string          `json:"message"`
	Priority    AnomalySeverity `json:"priority"`
	CreatedAt   time.Time       `json:"created_at"`
	IsRead      bool            `json:"is_read"`
	ActionTaken string          `json:"action_taken,omitempty"`
}

// AnomalyDetectionConfig configuration pour la détection d'anomalies
type AnomalyDetectionConfig struct {
	CaseID                string   `json:"case_id"`
	EnableTimeline        bool     `json:"enable_timeline"`
	EnableFinancial       bool     `json:"enable_financial"`
	EnableCommunication   bool     `json:"enable_communication"`
	EnableBehavior        bool     `json:"enable_behavior"`
	EnableLocation        bool     `json:"enable_location"`
	EnableRelation        bool     `json:"enable_relation"`
	EnablePattern         bool     `json:"enable_pattern"`
	MinConfidence         int      `json:"min_confidence"`         // Seuil minimum de confiance
	AutoAlert             bool     `json:"auto_alert"`             // Alertes automatiques
	AlertSeverityThreshold AnomalySeverity `json:"alert_severity_threshold"`
	WatchedEntities       []string `json:"watched_entities,omitempty"` // Entités à surveiller
}

// AnomalyDetectionResult résultat d'une détection d'anomalies
type AnomalyDetectionResult struct {
	CaseID            string    `json:"case_id"`
	DetectedAt        time.Time `json:"detected_at"`
	TotalAnomalies    int       `json:"total_anomalies"`
	NewAnomalies      int       `json:"new_anomalies"`
	CriticalCount     int       `json:"critical_count"`
	HighCount         int       `json:"high_count"`
	MediumCount       int       `json:"medium_count"`
	LowCount          int       `json:"low_count"`
	Anomalies         []Anomaly `json:"anomalies"`
	Summary           string    `json:"summary"`
	Alerts            []AnomalyAlert `json:"alerts,omitempty"`
}

// AnomalyStatistics statistiques d'anomalies pour un cas
type AnomalyStatistics struct {
	CaseID           string                 `json:"case_id"`
	TotalDetected    int                    `json:"total_detected"`
	Acknowledged     int                    `json:"acknowledged"`
	Pending          int                    `json:"pending"`
	ByType           map[AnomalyType]int    `json:"by_type"`
	BySeverity       map[AnomalySeverity]int `json:"by_severity"`
	TrendDirection   string                 `json:"trend_direction"` // increasing, decreasing, stable
	AvgConfidence    float64                `json:"avg_confidence"`
	LastDetection    time.Time              `json:"last_detection"`
}
