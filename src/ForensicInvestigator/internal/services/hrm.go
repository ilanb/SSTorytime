package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HRMService handles communication with the HRM (Hypothetical Reasoning Model) server
type HRMService struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewHRMService creates a new HRM service client
func NewHRMService(baseURL string) *HRMService {
	return &HRMService{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 600 * time.Second, // 10 minutes pour le raisonnement hiérarchique avec LLM
		},
	}
}

// Evidence represents a piece of evidence for HRM analysis
type Evidence struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Confidence  float64                `json:"confidence"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Hypothesis represents a hypothesis to verify
type Hypothesis struct {
	ID                    string   `json:"id"`
	Statement             string   `json:"statement"`
	SupportingEvidence    []string `json:"supporting_evidence"`
	ContradictingEvidence []string `json:"contradicting_evidence"`
	Confidence            float64  `json:"confidence"`
}

// ReasoningRequest is the request for general reasoning
type ReasoningRequest struct {
	Context       string     `json:"context"`
	Question      string     `json:"question"`
	Evidence      []Evidence `json:"evidence"`
	ReasoningType string     `json:"reasoning_type"`
	MaxDepth      int        `json:"max_depth"`
}

// ReasoningStep represents a step in the reasoning chain
type ReasoningStep struct {
	StepNumber   int      `json:"step_number"`
	Premise      string   `json:"premise"`
	Inference    string   `json:"inference"`
	Confidence   float64  `json:"confidence"`
	EvidenceUsed []string `json:"evidence_used"`
}

// ReasoningResponse is the response from reasoning endpoint
type ReasoningResponse struct {
	Conclusion             string                   `json:"conclusion"`
	Confidence             float64                  `json:"confidence"`
	ReasoningChain         []ReasoningStep          `json:"reasoning_chain"`
	AlternativeConclusions []map[string]interface{} `json:"alternative_conclusions"`
	Warnings               []string                 `json:"warnings"`
}

// HypothesisVerificationRequest is the request for hypothesis verification
type HypothesisVerificationRequest struct {
	Hypothesis  Hypothesis `json:"hypothesis"`
	Evidence    []Evidence `json:"evidence"`
	CaseContext string     `json:"case_context"`
	StrictMode  bool       `json:"strict_mode"`
}

// HypothesisVerificationResponse is the response from hypothesis verification
type HypothesisVerificationResponse struct {
	HypothesisID         string   `json:"hypothesis_id"`
	IsSupported          bool     `json:"is_supported"`
	Confidence           float64  `json:"confidence"`
	SupportingReasons    []string `json:"supporting_reasons"`
	ContradictingReasons []string `json:"contradicting_reasons"`
	MissingEvidence      []string `json:"missing_evidence"`
	Recommendation       string   `json:"recommendation"`
}

// ContradictionRequest is the request for contradiction detection
type ContradictionRequest struct {
	Statements  []map[string]string `json:"statements"`
	Evidence    []Evidence          `json:"evidence"`
	CaseContext string              `json:"case_context"`
}

// Contradiction represents a detected contradiction
type Contradiction struct {
	StatementIDs          []string `json:"statement_ids"`
	Description           string   `json:"description"`
	Severity              string   `json:"severity"`
	ResolutionSuggestions []string `json:"resolution_suggestions"`
}

// ContradictionResponse is the response from contradiction detection
type ContradictionResponse struct {
	Contradictions   []Contradiction `json:"contradictions"`
	ConsistencyScore float64         `json:"consistency_score"`
	AnalysisSummary  string          `json:"analysis_summary"`
}

// CasePattern represents a pattern found across cases
type CasePattern struct {
	PatternType   string   `json:"pattern_type"`
	Description   string   `json:"description"`
	CasesInvolved []string `json:"cases_involved"`
	Confidence    float64  `json:"confidence"`
	Significance  string   `json:"significance"`
}

// CrossCaseRequest is the request for cross-case reasoning
type CrossCaseRequest struct {
	PrimaryCase     map[string]interface{}   `json:"primary_case"`
	ComparisonCases []map[string]interface{} `json:"comparison_cases"`
	FocusAreas      []string                 `json:"focus_areas"`
}

// CrossCaseResponse is the response from cross-case reasoning
type CrossCaseResponse struct {
	Patterns           []CasePattern            `json:"patterns"`
	Connections        []map[string]interface{} `json:"connections"`
	InvestigativeLeads []string                 `json:"investigative_leads"`
	RiskAssessment     map[string]interface{}   `json:"risk_assessment"`
	Summary            string                   `json:"summary"`
}

// IsAvailable checks if the HRM service is available
func (s *HRMService) IsAvailable() bool {
	resp, err := s.HTTPClient.Get(s.BaseURL + "/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// Reason performs general reasoning analysis
func (s *HRMService) Reason(req ReasoningRequest) (*ReasoningResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := s.HTTPClient.Post(
		s.BaseURL+"/reason",
		"application/json",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to call HRM API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HRM API error: %s - %s", resp.Status, string(bodyBytes))
	}

	var result ReasoningResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// VerifyHypothesis verifies a hypothesis against evidence
func (s *HRMService) VerifyHypothesis(req HypothesisVerificationRequest) (*HypothesisVerificationResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := s.HTTPClient.Post(
		s.BaseURL+"/verify-hypothesis",
		"application/json",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to call HRM API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HRM API error: %s - %s", resp.Status, string(bodyBytes))
	}

	var result HypothesisVerificationResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// FindContradictions detects contradictions in statements and evidence
func (s *HRMService) FindContradictions(req ContradictionRequest) (*ContradictionResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := s.HTTPClient.Post(
		s.BaseURL+"/find-contradictions",
		"application/json",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to call HRM API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HRM API error: %s - %s", resp.Status, string(bodyBytes))
	}

	var result ContradictionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// CrossCaseReasoning analyzes patterns across multiple cases
func (s *HRMService) CrossCaseReasoning(req CrossCaseRequest) (*CrossCaseResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := s.HTTPClient.Post(
		s.BaseURL+"/cross-case-reasoning",
		"application/json",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to call HRM API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HRM API error: %s - %s", resp.Status, string(bodyBytes))
	}

	var result CrossCaseResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// ConvertCaseToHRMFormat converts a ForensicInvestigator case to HRM format
func ConvertCaseToHRMFormat(caseData map[string]interface{}) map[string]interface{} {
	hrmCase := map[string]interface{}{
		"id":          caseData["id"],
		"type":        caseData["type"],
		"description": caseData["description"],
	}

	// Convert timeline
	if timeline, ok := caseData["timeline"].([]interface{}); ok {
		hrmCase["timeline"] = timeline
	}

	// Convert evidence
	if evidence, ok := caseData["evidence"].([]interface{}); ok {
		hrmCase["evidence"] = evidence
	}

	// Convert hypotheses
	if hypotheses, ok := caseData["hypotheses"].([]interface{}); ok {
		hrmCase["hypotheses"] = hypotheses
	}

	return hrmCase
}

// ConvertEvidenceToHRMFormat converts evidence items to HRM format
func ConvertEvidenceToHRMFormat(evidenceList []interface{}) []Evidence {
	var hrmEvidence []Evidence

	for _, ev := range evidenceList {
		if evMap, ok := ev.(map[string]interface{}); ok {
			evidence := Evidence{
				ID:          getString(evMap, "id"),
				Type:        getString(evMap, "type"),
				Description: getString(evMap, "description"),
				Confidence:  getFloat(evMap, "confidence", 0.5),
			}
			if metadata, ok := evMap["metadata"].(map[string]interface{}); ok {
				evidence.Metadata = metadata
			}
			hrmEvidence = append(hrmEvidence, evidence)
		}
	}

	return hrmEvidence
}

// Helper functions
func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getFloat(m map[string]interface{}, key string, defaultVal float64) float64 {
	if v, ok := m[key].(float64); ok {
		return v
	}
	return defaultVal
}
