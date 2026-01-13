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
