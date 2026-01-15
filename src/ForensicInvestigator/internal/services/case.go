package services

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"forensicinvestigator/internal/models"

	"github.com/google/uuid"
)

// CaseService gère les affaires et leurs données
type CaseService struct {
	cases    map[string]*models.Case
	mu       sync.RWMutex
}

// NewCaseService crée une nouvelle instance du service
func NewCaseService() *CaseService {
	return &CaseService{
		cases: make(map[string]*models.Case),
	}
}

// CreateCase crée une nouvelle affaire
func (s *CaseService) CreateCase(name, description, caseType string) (*models.Case, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	c := &models.Case{
		ID:          uuid.New().String(),
		Name:        name,
		Description: description,
		Type:        caseType,
		Status:      "en_cours",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Entities:    []models.Entity{},
		Evidence:    []models.Evidence{},
		Timeline:    []models.Event{},
		Hypotheses:  []models.Hypothesis{},
	}

	s.cases[c.ID] = c
	return c, nil
}

// GetCase récupère une affaire par ID
func (s *CaseService) GetCase(id string) (*models.Case, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	c, exists := s.cases[id]
	if !exists {
		return nil, fmt.Errorf("affaire non trouvée: %s", id)
	}
	return c, nil
}

// GetAllCases récupère toutes les affaires
func (s *CaseService) GetAllCases() []*models.Case {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cases := make([]*models.Case, 0, len(s.cases))
	for _, c := range s.cases {
		cases = append(cases, c)
	}
	return cases
}

// UpdateCase met à jour une affaire
func (s *CaseService) UpdateCase(c *models.Case) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.cases[c.ID]; !exists {
		return fmt.Errorf("affaire non trouvée: %s", c.ID)
	}

	c.UpdatedAt = time.Now()
	s.cases[c.ID] = c
	return nil
}

// DeleteCase supprime une affaire
func (s *CaseService) DeleteCase(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.cases[id]; !exists {
		return fmt.Errorf("affaire non trouvée: %s", id)
	}

	delete(s.cases, id)
	return nil
}

// AddEntity ajoute une entité à une affaire
func (s *CaseService) AddEntity(caseID string, entity models.Entity) (*models.Entity, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	c, exists := s.cases[caseID]
	if !exists {
		return nil, fmt.Errorf("affaire non trouvée: %s", caseID)
	}

	entity.ID = uuid.New().String()
	entity.CaseID = caseID
	entity.CreatedAt = time.Now()
	if entity.Attributes == nil {
		entity.Attributes = make(map[string]string)
	}
	if entity.Relations == nil {
		entity.Relations = []models.Relation{}
	}

	c.Entities = append(c.Entities, entity)
	c.UpdatedAt = time.Now()

	return &entity, nil
}

// GetEntities récupère toutes les entités d'une affaire
func (s *CaseService) GetEntities(caseID string) ([]models.Entity, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	c, exists := s.cases[caseID]
	if !exists {
		return nil, fmt.Errorf("affaire non trouvée: %s", caseID)
	}

	return c.Entities, nil
}

// AddRelation ajoute une relation entre deux entités
func (s *CaseService) AddRelation(caseID string, relation models.Relation) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	c, exists := s.cases[caseID]
	if !exists {
		return fmt.Errorf("affaire non trouvée: %s", caseID)
	}

	relation.ID = uuid.New().String()

	// Trouver l'entité source et ajouter la relation
	for i, e := range c.Entities {
		if e.ID == relation.FromID {
			c.Entities[i].Relations = append(c.Entities[i].Relations, relation)
			c.UpdatedAt = time.Now()
			return nil
		}
	}

	return fmt.Errorf("entité source non trouvée: %s", relation.FromID)
}

// AddEvidence ajoute une preuve à une affaire
func (s *CaseService) AddEvidence(caseID string, evidence models.Evidence) (*models.Evidence, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	c, exists := s.cases[caseID]
	if !exists {
		return nil, fmt.Errorf("affaire non trouvée: %s", caseID)
	}

	evidence.ID = uuid.New().String()
	evidence.CaseID = caseID
	evidence.CollectedAt = time.Now()
	if evidence.ChainOfCustody == nil {
		evidence.ChainOfCustody = []string{}
	}
	if evidence.LinkedEntities == nil {
		evidence.LinkedEntities = []string{}
	}

	c.Evidence = append(c.Evidence, evidence)
	c.UpdatedAt = time.Now()

	return &evidence, nil
}

// GetEvidence récupère toutes les preuves d'une affaire
func (s *CaseService) GetEvidence(caseID string) ([]models.Evidence, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	c, exists := s.cases[caseID]
	if !exists {
		return nil, fmt.Errorf("affaire non trouvée: %s", caseID)
	}

	return c.Evidence, nil
}

// AddEvent ajoute un événement à la timeline
func (s *CaseService) AddEvent(caseID string, event models.Event) (*models.Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	c, exists := s.cases[caseID]
	if !exists {
		return nil, fmt.Errorf("affaire non trouvée: %s", caseID)
	}

	event.ID = uuid.New().String()
	event.CaseID = caseID
	if event.Entities == nil {
		event.Entities = []string{}
	}
	if event.Evidence == nil {
		event.Evidence = []string{}
	}

	c.Timeline = append(c.Timeline, event)
	c.UpdatedAt = time.Now()

	return &event, nil
}

// GetTimeline récupère la timeline d'une affaire
func (s *CaseService) GetTimeline(caseID string) ([]models.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	c, exists := s.cases[caseID]
	if !exists {
		return nil, fmt.Errorf("affaire non trouvée: %s", caseID)
	}

	return c.Timeline, nil
}

// AddHypothesis ajoute une hypothèse à une affaire
func (s *CaseService) AddHypothesis(caseID string, hypothesis models.Hypothesis) (*models.Hypothesis, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	c, exists := s.cases[caseID]
	if !exists {
		return nil, fmt.Errorf("affaire non trouvée: %s", caseID)
	}

	hypothesis.ID = uuid.New().String()
	hypothesis.CaseID = caseID
	hypothesis.CreatedAt = time.Now()
	hypothesis.UpdatedAt = time.Now()
	if hypothesis.SupportingEvidence == nil {
		hypothesis.SupportingEvidence = []string{}
	}
	if hypothesis.ContradictingEvidence == nil {
		hypothesis.ContradictingEvidence = []string{}
	}
	if hypothesis.Questions == nil {
		hypothesis.Questions = []string{}
	}

	c.Hypotheses = append(c.Hypotheses, hypothesis)
	c.UpdatedAt = time.Now()

	return &hypothesis, nil
}

// GetHypotheses récupère les hypothèses d'une affaire
func (s *CaseService) GetHypotheses(caseID string) ([]models.Hypothesis, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	c, exists := s.cases[caseID]
	if !exists {
		return nil, fmt.Errorf("affaire non trouvée: %s", caseID)
	}

	return c.Hypotheses, nil
}

// UpdateHypothesis met à jour une hypothèse existante
func (s *CaseService) UpdateHypothesis(caseID string, hypothesis models.Hypothesis) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	c, exists := s.cases[caseID]
	if !exists {
		return fmt.Errorf("affaire non trouvée: %s", caseID)
	}

	for i, h := range c.Hypotheses {
		if h.ID == hypothesis.ID {
			// Conserver les données de création si non fournies
			if hypothesis.CreatedAt.IsZero() {
				hypothesis.CreatedAt = h.CreatedAt
			}
			// Conserver les preuves si non fournies
			if hypothesis.SupportingEvidence == nil {
				hypothesis.SupportingEvidence = h.SupportingEvidence
			}
			if hypothesis.ContradictingEvidence == nil {
				hypothesis.ContradictingEvidence = h.ContradictingEvidence
			}
			if hypothesis.Questions == nil {
				hypothesis.Questions = h.Questions
			}
			if hypothesis.GeneratedBy == "" {
				hypothesis.GeneratedBy = h.GeneratedBy
			}
			hypothesis.CaseID = caseID
			hypothesis.UpdatedAt = time.Now()
			c.Hypotheses[i] = hypothesis
			c.UpdatedAt = time.Now()
			return nil
		}
	}

	return fmt.Errorf("hypothèse non trouvée: %s", hypothesis.ID)
}

// DeleteHypothesis supprime une hypothèse d'une affaire
func (s *CaseService) DeleteHypothesis(caseID, hypothesisID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	c, exists := s.cases[caseID]
	if !exists {
		return fmt.Errorf("affaire non trouvée: %s", caseID)
	}

	for i, h := range c.Hypotheses {
		if h.ID == hypothesisID {
			c.Hypotheses = append(c.Hypotheses[:i], c.Hypotheses[i+1:]...)
			c.UpdatedAt = time.Now()
			return nil
		}
	}

	return fmt.Errorf("hypothèse non trouvée: %s", hypothesisID)
}

// BuildGraphData construit les données du graphe pour une affaire
func (s *CaseService) BuildGraphData(caseID string) (*models.GraphData, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	c, exists := s.cases[caseID]
	if !exists {
		return nil, fmt.Errorf("affaire non trouvée: %s", caseID)
	}

	graph := &models.GraphData{
		Nodes: []models.GraphNode{},
		Edges: []models.GraphEdge{},
	}

	// Ajouter les entités comme nœuds
	for _, e := range c.Entities {
		node := models.GraphNode{
			ID:    e.ID,
			Label: e.Name,
			Type:  string(e.Type),
			Role:  string(e.Role),
			Data:  e.Attributes,
		}
		graph.Nodes = append(graph.Nodes, node)

		// Ajouter les relations comme arêtes
		for _, r := range e.Relations {
			edge := models.GraphEdge{
				From:    r.FromID,
				To:      r.ToID,
				Label:   r.Label,
				Type:    r.Type,
				Context: r.Context,
			}
			graph.Edges = append(graph.Edges, edge)
		}
	}

	// Ajouter les preuves comme nœuds
	for _, ev := range c.Evidence {
		node := models.GraphNode{
			ID:    ev.ID,
			Label: ev.Name,
			Type:  "preuve",
			Data: map[string]string{
				"type":       string(ev.Type),
				"fiabilite":  fmt.Sprintf("%d", ev.Reliability),
			},
		}
		graph.Nodes = append(graph.Nodes, node)

		// Lier aux entités
		for _, entityID := range ev.LinkedEntities {
			edge := models.GraphEdge{
				From:  ev.ID,
				To:    entityID,
				Label: "concerne",
				Type:  "evidence_link",
			}
			graph.Edges = append(graph.Edges, edge)
		}
	}

	return graph, nil
}

// UpdateN4LContent met à jour le contenu N4L d'une affaire
func (s *CaseService) UpdateN4LContent(caseID, content string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	c, exists := s.cases[caseID]
	if !exists {
		return fmt.Errorf("affaire non trouvée: %s", caseID)
	}

	c.N4LContent = content
	c.UpdatedAt = time.Now()
	return nil
}

// LoadDemoCases charge les affaires de démonstration
func (s *CaseService) LoadDemoCases(demoCases []*models.Case) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	count := 0
	for _, c := range demoCases {
		if existing, exists := s.cases[c.ID]; !exists {
			// Nouveau cas - ajouter
			s.cases[c.ID] = c
			count++
		} else if c.N4LContent != "" && existing.N4LContent != c.N4LContent {
			// Cas existant mais N4LContent différent - mettre à jour le contenu N4L
			existing.N4LContent = c.N4LContent
		}
	}
	return count
}

// ClearAllCases supprime toutes les affaires
func (s *CaseService) ClearAllCases() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cases = make(map[string]*models.Case)
}

// DeleteEntity supprime une entité d'une affaire
func (s *CaseService) DeleteEntity(caseID, entityID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	c, exists := s.cases[caseID]
	if !exists {
		return fmt.Errorf("affaire non trouvée: %s", caseID)
	}

	for i, e := range c.Entities {
		if e.ID == entityID {
			c.Entities = append(c.Entities[:i], c.Entities[i+1:]...)
			c.UpdatedAt = time.Now()
			return nil
		}
	}

	return fmt.Errorf("entité non trouvée: %s", entityID)
}

// UpdateEntity met à jour une entité existante
func (s *CaseService) UpdateEntity(caseID string, entity models.Entity) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	c, exists := s.cases[caseID]
	if !exists {
		return fmt.Errorf("affaire non trouvée: %s", caseID)
	}

	for i, e := range c.Entities {
		if e.ID == entity.ID {
			// Conserver les relations existantes si non fournies
			if entity.Relations == nil {
				entity.Relations = e.Relations
			}
			// Conserver les attributs existants si non fournis
			if entity.Attributes == nil {
				entity.Attributes = e.Attributes
			}
			c.Entities[i] = entity
			c.UpdatedAt = time.Now()
			return nil
		}
	}

	return fmt.Errorf("entité non trouvée: %s", entity.ID)
}

// DeleteEvidence supprime une preuve d'une affaire
func (s *CaseService) DeleteEvidence(caseID, evidenceID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	c, exists := s.cases[caseID]
	if !exists {
		return fmt.Errorf("affaire non trouvée: %s", caseID)
	}

	for i, ev := range c.Evidence {
		if ev.ID == evidenceID {
			c.Evidence = append(c.Evidence[:i], c.Evidence[i+1:]...)
			c.UpdatedAt = time.Now()
			return nil
		}
	}

	return fmt.Errorf("preuve non trouvée: %s", evidenceID)
}

// DeleteEvent supprime un événement d'une affaire
func (s *CaseService) DeleteEvent(caseID, eventID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	c, exists := s.cases[caseID]
	if !exists {
		return fmt.Errorf("affaire non trouvée: %s", caseID)
	}

	for i, evt := range c.Timeline {
		if evt.ID == eventID {
			c.Timeline = append(c.Timeline[:i], c.Timeline[i+1:]...)
			c.UpdatedAt = time.Now()
			return nil
		}
	}

	return fmt.Errorf("événement non trouvé: %s", eventID)
}

// UpdateEvent met à jour un événement existant
func (s *CaseService) UpdateEvent(caseID string, event models.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	c, exists := s.cases[caseID]
	if !exists {
		return fmt.Errorf("affaire non trouvée: %s", caseID)
	}

	for i, evt := range c.Timeline {
		if evt.ID == event.ID {
			// Conserver le CaseID original
			event.CaseID = caseID
			// Conserver les entités liées si non fournies
			if event.Entities == nil {
				event.Entities = evt.Entities
			}
			// Conserver les preuves liées si non fournies
			if event.Evidence == nil {
				event.Evidence = evt.Evidence
			}
			c.Timeline[i] = event
			c.UpdatedAt = time.Now()
			return nil
		}
	}

	return fmt.Errorf("événement non trouvé: %s", event.ID)
}

// UpdateEvidence met à jour une preuve existante
func (s *CaseService) UpdateEvidence(caseID string, evidence models.Evidence) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	c, exists := s.cases[caseID]
	if !exists {
		return fmt.Errorf("affaire non trouvée: %s", caseID)
	}

	for i, ev := range c.Evidence {
		if ev.ID == evidence.ID {
			// Conserver les données de collection si non fournies
			if evidence.CollectedAt.IsZero() {
				evidence.CollectedAt = ev.CollectedAt
			}
			// Conserver la chaîne de possession si non fournie
			if evidence.ChainOfCustody == nil {
				evidence.ChainOfCustody = ev.ChainOfCustody
			}
			// Conserver les entités liées si non fournies
			if evidence.LinkedEntities == nil {
				evidence.LinkedEntities = ev.LinkedEntities
			}
			evidence.CaseID = caseID
			c.Evidence[i] = evidence
			c.UpdatedAt = time.Now()
			return nil
		}
	}

	return fmt.Errorf("preuve non trouvée: %s", evidence.ID)
}

// FindCrossReferences recherche les correspondances entre une affaire et toutes les autres
func (s *CaseService) FindCrossReferences(caseID string) (*models.CrossCaseResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	currentCase, exists := s.cases[caseID]
	if !exists {
		return nil, fmt.Errorf("affaire non trouvée: %s", caseID)
	}

	result := &models.CrossCaseResult{
		Matches: []models.CrossCaseMatch{},
		Alerts:  []string{},
	}

	matchID := 0

	// Comparer avec toutes les autres affaires
	for otherID, otherCase := range s.cases {
		if otherID == caseID {
			continue
		}

		// 1. Rechercher les entités similaires (même nom ou attributs similaires)
		for _, currentEntity := range currentCase.Entities {
			for _, otherEntity := range otherCase.Entities {
				if similarity := s.compareEntities(currentEntity, otherEntity); similarity > 50 {
					matchID++
					match := models.CrossCaseMatch{
						ID:              fmt.Sprintf("match-%d", matchID),
						CurrentCaseID:   caseID,
						CurrentCaseName: currentCase.Name,
						OtherCaseID:     otherID,
						OtherCaseName:   otherCase.Name,
						MatchType:       models.MatchEntity,
						Confidence:      similarity,
						Description:     fmt.Sprintf("Entité similaire: %s ↔ %s", currentEntity.Name, otherEntity.Name),
						CurrentElement:  currentEntity.Name,
						OtherElement:    otherEntity.Name,
						Details: map[string]string{
							"current_type": string(currentEntity.Type),
							"other_type":   string(otherEntity.Type),
							"current_role": string(currentEntity.Role),
							"other_role":   string(otherEntity.Role),
						},
					}
					result.Matches = append(result.Matches, match)

					if similarity > 80 {
						result.Alerts = append(result.Alerts,
							fmt.Sprintf("Forte correspondance: '%s' (%s) apparaît aussi dans l'affaire '%s'",
								currentEntity.Name, currentEntity.Role, otherCase.Name))
					}
				}
			}
		}

		// 2. Rechercher les lieux communs
		currentLocations := s.extractLocations(currentCase)
		otherLocations := s.extractLocations(otherCase)

		for _, currentLoc := range currentLocations {
			for _, otherLoc := range otherLocations {
				if similarity := s.compareStrings(currentLoc, otherLoc); similarity > 60 {
					matchID++
					match := models.CrossCaseMatch{
						ID:              fmt.Sprintf("match-%d", matchID),
						CurrentCaseID:   caseID,
						CurrentCaseName: currentCase.Name,
						OtherCaseID:     otherID,
						OtherCaseName:   otherCase.Name,
						MatchType:       models.MatchLocation,
						Confidence:      similarity,
						Description:     fmt.Sprintf("Lieu commun: %s ↔ %s", currentLoc, otherLoc),
						CurrentElement:  currentLoc,
						OtherElement:    otherLoc,
					}
					result.Matches = append(result.Matches, match)
				}
			}
		}

		// 3. Rechercher les modus operandi similaires (même type d'affaire)
		if currentCase.Type == otherCase.Type {
			matchID++
			match := models.CrossCaseMatch{
				ID:              fmt.Sprintf("match-%d", matchID),
				CurrentCaseID:   caseID,
				CurrentCaseName: currentCase.Name,
				OtherCaseID:     otherID,
				OtherCaseName:   otherCase.Name,
				MatchType:       models.MatchModus,
				Confidence:      70,
				Description:     fmt.Sprintf("Même type d'affaire: %s", currentCase.Type),
				CurrentElement:  currentCase.Type,
				OtherElement:    otherCase.Type,
			}
			result.Matches = append(result.Matches, match)
		}

		// 4. Rechercher les chevauchements temporels
		if overlap := s.checkTemporalOverlap(currentCase, otherCase); overlap {
			matchID++
			match := models.CrossCaseMatch{
				ID:              fmt.Sprintf("match-%d", matchID),
				CurrentCaseID:   caseID,
				CurrentCaseName: currentCase.Name,
				OtherCaseID:     otherID,
				OtherCaseName:   otherCase.Name,
				MatchType:       models.MatchTemporal,
				Confidence:      60,
				Description:     "Les périodes d'événements se chevauchent",
				CurrentElement:  "Timeline",
				OtherElement:    "Timeline",
			}
			result.Matches = append(result.Matches, match)
		}
	}

	// Générer un résumé
	if len(result.Matches) > 0 {
		result.Summary = fmt.Sprintf("%d correspondances trouvées avec %d autres affaires",
			len(result.Matches), s.countUniqueCases(result.Matches))
	} else {
		result.Summary = "Aucune correspondance significative trouvée"
	}

	return result, nil
}

// compareEntities compare deux entités et retourne un score de similarité (0-100)
func (s *CaseService) compareEntities(e1, e2 models.Entity) int {
	score := 0

	// Comparaison des noms
	nameScore := s.compareStrings(e1.Name, e2.Name)
	score += nameScore / 2 // 50% du score max

	// Comparaison des types
	if e1.Type == e2.Type {
		score += 20
	}

	// Comparaison des attributs
	if e1.Attributes != nil && e2.Attributes != nil {
		matchingAttrs := 0
		totalAttrs := 0
		for k, v1 := range e1.Attributes {
			totalAttrs++
			if v2, ok := e2.Attributes[k]; ok {
				if s.compareStrings(v1, v2) > 80 {
					matchingAttrs++
				}
			}
		}
		if totalAttrs > 0 {
			score += (matchingAttrs * 30) / totalAttrs
		}
	}

	return score
}

// compareStrings compare deux chaînes et retourne un score de similarité (0-100)
func (s *CaseService) compareStrings(s1, s2 string) int {
	s1 = strings.ToLower(strings.TrimSpace(s1))
	s2 = strings.ToLower(strings.TrimSpace(s2))

	if s1 == s2 {
		return 100
	}

	// Vérifier si l'une contient l'autre
	if strings.Contains(s1, s2) || strings.Contains(s2, s1) {
		shorter := len(s1)
		if len(s2) < shorter {
			shorter = len(s2)
		}
		longer := len(s1)
		if len(s2) > longer {
			longer = len(s2)
		}
		return (shorter * 90) / longer
	}

	// Calculer la distance de Levenshtein simplifiée
	maxLen := len(s1)
	if len(s2) > maxLen {
		maxLen = len(s2)
	}
	if maxLen == 0 {
		return 100
	}

	// Compter les caractères communs
	common := 0
	used := make([]bool, len(s2))
	for _, c1 := range s1 {
		for j, c2 := range s2 {
			if !used[j] && c1 == c2 {
				common++
				used[j] = true
				break
			}
		}
	}

	return (common * 100) / maxLen
}

// extractLocations extrait tous les lieux d'une affaire
func (s *CaseService) extractLocations(c *models.Case) []string {
	locations := []string{}
	seen := make(map[string]bool)

	// Extraire des entités de type lieu
	for _, e := range c.Entities {
		if e.Type == models.EntityPlace && !seen[e.Name] {
			locations = append(locations, e.Name)
			seen[e.Name] = true
		}
		// Extraire l'attribut adresse si présent
		if addr, ok := e.Attributes["adresse"]; ok && !seen[addr] {
			locations = append(locations, addr)
			seen[addr] = true
		}
	}

	// Extraire des preuves (location)
	for _, ev := range c.Evidence {
		if ev.Location != "" && !seen[ev.Location] {
			locations = append(locations, ev.Location)
			seen[ev.Location] = true
		}
	}

	// Extraire des événements
	for _, evt := range c.Timeline {
		if evt.Location != "" && !seen[evt.Location] {
			locations = append(locations, evt.Location)
			seen[evt.Location] = true
		}
	}

	return locations
}

// checkTemporalOverlap vérifie si les timelines de deux affaires se chevauchent
func (s *CaseService) checkTemporalOverlap(c1, c2 *models.Case) bool {
	if len(c1.Timeline) == 0 || len(c2.Timeline) == 0 {
		return false
	}

	// Trouver les bornes temporelles de chaque affaire
	var c1Start, c1End, c2Start, c2End time.Time

	for i, evt := range c1.Timeline {
		if i == 0 || evt.Timestamp.Before(c1Start) {
			c1Start = evt.Timestamp
		}
		if i == 0 || evt.Timestamp.After(c1End) {
			c1End = evt.Timestamp
		}
	}

	for i, evt := range c2.Timeline {
		if i == 0 || evt.Timestamp.Before(c2Start) {
			c2Start = evt.Timestamp
		}
		if i == 0 || evt.Timestamp.After(c2End) {
			c2End = evt.Timestamp
		}
	}

	// Vérifier le chevauchement
	return !(c1End.Before(c2Start) || c2End.Before(c1Start))
}

// countUniqueCases compte le nombre d'affaires uniques dans les correspondances
func (s *CaseService) countUniqueCases(matches []models.CrossCaseMatch) int {
	seen := make(map[string]bool)
	for _, m := range matches {
		seen[m.OtherCaseID] = true
	}
	return len(seen)
}

// BuildCrossCaseGraph construit un graphe des connexions inter-affaires
func (s *CaseService) BuildCrossCaseGraph(caseID string, matches []models.CrossCaseMatch) (*models.GraphData, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	currentCase, exists := s.cases[caseID]
	if !exists {
		return nil, fmt.Errorf("affaire non trouvée: %s", caseID)
	}

	graph := &models.GraphData{
		Nodes: []models.GraphNode{},
		Edges: []models.GraphEdge{},
	}

	// Ajouter l'affaire courante comme nœud central
	graph.Nodes = append(graph.Nodes, models.GraphNode{
		ID:    caseID,
		Label: currentCase.Name,
		Type:  "case_current",
		Data: map[string]string{
			"type":   currentCase.Type,
			"status": currentCase.Status,
		},
	})

	// Map pour éviter les doublons
	addedCases := make(map[string]bool)
	addedCases[caseID] = true

	// Ajouter les affaires liées et les connexions
	for _, match := range matches {
		// Ajouter l'autre affaire si pas déjà ajoutée
		if !addedCases[match.OtherCaseID] {
			if otherCase, exists := s.cases[match.OtherCaseID]; exists {
				graph.Nodes = append(graph.Nodes, models.GraphNode{
					ID:    match.OtherCaseID,
					Label: otherCase.Name,
					Type:  "case_other",
					Data: map[string]string{
						"type":   otherCase.Type,
						"status": otherCase.Status,
					},
				})
				addedCases[match.OtherCaseID] = true
			}
		}

		// Ajouter l'arête de connexion
		edgeLabel := string(match.MatchType)
		switch match.MatchType {
		case models.MatchEntity:
			edgeLabel = "Entité: " + match.CurrentElement
		case models.MatchLocation:
			edgeLabel = "Lieu: " + match.CurrentElement
		case models.MatchModus:
			edgeLabel = "Modus: " + match.CurrentElement
		case models.MatchTemporal:
			edgeLabel = "Période commune"
		}

		graph.Edges = append(graph.Edges, models.GraphEdge{
			From:  caseID,
			To:    match.OtherCaseID,
			Label: edgeLabel,
			Type:  string(match.MatchType),
		})
	}

	return graph, nil
}
