package services

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"forensicinvestigator/internal/models"

	"github.com/google/uuid"
)

// ScenarioService gère les simulations "What-If"
type ScenarioService struct {
	scenarios map[string]map[string]*models.Scenario // caseID -> scenarioID -> Scenario
	mu        sync.RWMutex
	cases     *CaseService
	ollama    *OllamaService
}

// NewScenarioService crée un nouveau service de scénarios
func NewScenarioService(cases *CaseService, ollama *OllamaService) *ScenarioService {
	return &ScenarioService{
		scenarios: make(map[string]map[string]*models.Scenario),
		cases:     cases,
		ollama:    ollama,
	}
}

// findExistingScenario retourne l'ID d'un scénario similaire existant, ou "" si aucun
func (s *ScenarioService) findExistingScenario(caseID string, assumption string, assumptionType string, targetEntityID string) string {
	if caseScenarios, ok := s.scenarios[caseID]; ok {
		// Normaliser l'assumption pour la comparaison
		normalizedAssumption := strings.TrimSpace(strings.ToLower(assumption))
		for id, existing := range caseScenarios {
			existingNormalized := strings.TrimSpace(strings.ToLower(existing.Assumption))
			// Un scénario est considéré comme dupliqué si:
			// - Même type d'hypothèse ET
			// - Même assumption (insensible à la casse) ET
			// - Même entité cible (ou les deux vides)
			if existing.AssumptionType == assumptionType &&
				existingNormalized == normalizedAssumption &&
				existing.TargetEntityID == targetEntityID {
				return id
			}
		}
	}
	return ""
}

// CreateScenario crée un nouveau scénario "What-If"
func (s *ScenarioService) CreateScenario(caseID string, req models.ScenarioSimulationRequest) (*models.Scenario, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Vérifier que le cas existe
	caseData, err := s.cases.GetCase(caseID)
	if err != nil {
		return nil, fmt.Errorf("cas non trouvé: %w", err)
	}

	// Vérifier si un scénario similaire existe déjà
	existingID := s.findExistingScenario(caseID, req.Assumption, req.AssumptionType, req.TargetEntityID)
	if existingID != "" {
		// Retourner le scénario existant au lieu d'en créer un nouveau
		return s.scenarios[caseID][existingID], nil
	}

	// Créer le scénario
	scenario := &models.Scenario{
		ID:             uuid.New().String(),
		CaseID:         caseID,
		Name:           s.generateScenarioName(req.Assumption, req.AssumptionType),
		Description:    req.Assumption,
		Assumption:     req.Assumption,
		AssumptionType: req.AssumptionType,
		TargetEntityID: req.TargetEntityID,
		Status:         models.ScenarioActive,
		Implications:   []models.ScenarioImplication{},
		SupportingFacts: []string{},
		ContradictingFacts: []string{},
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		CreatedBy:      "user",
	}

	// Calculer les implications et le score de plausibilité
	s.calculateImplications(scenario, caseData)
	s.calculatePlausibility(scenario, caseData)

	// Générer le graphe modifié
	scenario.ModifiedGraph = s.generateModifiedGraph(scenario, caseData)

	// Stocker le scénario
	if s.scenarios[caseID] == nil {
		s.scenarios[caseID] = make(map[string]*models.Scenario)
	}
	s.scenarios[caseID][scenario.ID] = scenario

	return scenario, nil
}

// GetScenarios retourne tous les scénarios d'un cas
func (s *ScenarioService) GetScenarios(caseID string) []*models.Scenario {
	s.mu.RLock()
	defer s.mu.RUnlock()

	scenarios := make([]*models.Scenario, 0)
	if caseScenarios, ok := s.scenarios[caseID]; ok {
		for _, scenario := range caseScenarios {
			scenarios = append(scenarios, scenario)
		}
	}

	// Trier par date de création
	sort.Slice(scenarios, func(i, j int) bool {
		return scenarios[i].CreatedAt.After(scenarios[j].CreatedAt)
	})

	return scenarios
}

// GetScenario retourne un scénario spécifique
func (s *ScenarioService) GetScenario(caseID, scenarioID string) (*models.Scenario, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if caseScenarios, ok := s.scenarios[caseID]; ok {
		if scenario, ok := caseScenarios[scenarioID]; ok {
			return scenario, nil
		}
	}
	return nil, fmt.Errorf("scénario non trouvé")
}

// DeleteScenario supprime un scénario
func (s *ScenarioService) DeleteScenario(caseID, scenarioID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if caseScenarios, ok := s.scenarios[caseID]; ok {
		if _, ok := caseScenarios[scenarioID]; ok {
			delete(caseScenarios, scenarioID)
			return nil
		}
	}
	return fmt.Errorf("scénario non trouvé")
}

// CompareScenarios compare deux scénarios
func (s *ScenarioService) CompareScenarios(caseID, scenarioID1, scenarioID2 string) (*models.ScenarioComparison, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Accès direct sans appeler GetScenario (évite deadlock potentiel)
	var scenario1, scenario2 *models.Scenario
	if caseScenarios, ok := s.scenarios[caseID]; ok {
		scenario1 = caseScenarios[scenarioID1]
		scenario2 = caseScenarios[scenarioID2]
	}
	if scenario1 == nil {
		return nil, fmt.Errorf("scénario 1 non trouvé")
	}
	if scenario2 == nil {
		return nil, fmt.Errorf("scénario 2 non trouvé")
	}

	comparison := &models.ScenarioComparison{
		Scenario1ID:       scenario1.ID,
		Scenario2ID:       scenario2.ID,
		Scenario1Name:     scenario1.Name,
		Scenario2Name:     scenario2.Name,
		CommonFacts:       s.findCommonFacts(scenario1, scenario2),
		DifferentFacts:    s.findDifferentFacts(scenario1, scenario2),
		PlausibilityDelta: scenario1.PlausibilityScore - scenario2.PlausibilityScore,
	}

	// Calculer les différences de graphe
	if scenario1.ModifiedGraph != nil && scenario2.ModifiedGraph != nil {
		comparison.GraphDifferences = s.calculateGraphDifferences(scenario1.ModifiedGraph, scenario2.ModifiedGraph)
	}

	// Générer la recommandation
	comparison.Recommendation = s.generateRecommendation(comparison, scenario1, scenario2)

	return comparison, nil
}

// SimulateWithAI utilise l'IA pour analyser un scénario
func (s *ScenarioService) SimulateWithAI(caseID, scenarioID string) (string, error) {
	s.mu.Lock()

	// Accès direct sans appeler GetScenario (évite deadlock)
	var scenario *models.Scenario
	if caseScenarios, ok := s.scenarios[caseID]; ok {
		scenario = caseScenarios[scenarioID]
	}
	if scenario == nil {
		s.mu.Unlock()
		return "", fmt.Errorf("scénario non trouvé")
	}

	caseData, err := s.cases.GetCase(caseID)
	if err != nil {
		s.mu.Unlock()
		return "", err
	}
	s.mu.Unlock()

	// Construire le prompt pour l'analyse IA
	prompt := s.buildAIPrompt(scenario, caseData)

	// Appeler le service Ollama
	caseContext := fmt.Sprintf("Affaire: %s\nDescription: %s", caseData.Name, caseData.Description)
	analysis, err := s.ollama.Chat(prompt, caseContext)
	if err != nil {
		return "", fmt.Errorf("erreur analyse IA: %w", err)
	}

	// Mettre à jour le scénario avec l'analyse
	scenario.AIAnalysis = analysis
	scenario.UpdatedAt = time.Now()

	return analysis, nil
}

// SimulateWithAIStream utilise l'IA pour analyser un scénario en streaming
func (s *ScenarioService) SimulateWithAIStream(caseID, scenarioID string, callback StreamCallback) error {
	s.mu.Lock()

	// Accès direct sans appeler GetScenario (évite deadlock)
	var scenario *models.Scenario
	if caseScenarios, ok := s.scenarios[caseID]; ok {
		scenario = caseScenarios[scenarioID]
	}
	if scenario == nil {
		s.mu.Unlock()
		return fmt.Errorf("scénario non trouvé")
	}

	caseData, err := s.cases.GetCase(caseID)
	if err != nil {
		s.mu.Unlock()
		return err
	}
	s.mu.Unlock()

	// Construire le prompt pour l'analyse IA
	prompt := s.buildAIPrompt(scenario, caseData)

	// Appeler le service Ollama en streaming
	caseContext := fmt.Sprintf("Affaire: %s\nDescription: %s", caseData.Name, caseData.Description)

	var fullAnalysis strings.Builder
	err = s.ollama.ChatStream(prompt, caseContext, func(chunk string, done bool) error {
		fullAnalysis.WriteString(chunk)
		return callback(chunk, done)
	})

	if err != nil {
		return fmt.Errorf("erreur analyse IA: %w", err)
	}

	// Mettre à jour le scénario avec l'analyse complète
	s.mu.Lock()
	if caseScenarios, ok := s.scenarios[caseID]; ok {
		if sc := caseScenarios[scenarioID]; sc != nil {
			sc.AIAnalysis = fullAnalysis.String()
			sc.UpdatedAt = time.Now()
		}
	}
	s.mu.Unlock()

	return nil
}

// GenerateScenariosWithAI génère automatiquement des scénarios basés sur l'affaire
func (s *ScenarioService) GenerateScenariosWithAI(caseID string) ([]*models.Scenario, error) {
	// Récupérer les données du cas
	caseData, err := s.cases.GetCase(caseID)
	if err != nil {
		return nil, fmt.Errorf("cas non trouvé: %w", err)
	}

	// Construire le prompt pour générer des scénarios
	prompt := s.buildScenarioGenerationPrompt(caseData)

	// Appeler le LLM
	caseContext := fmt.Sprintf("Affaire: %s\nDescription: %s", caseData.Name, caseData.Description)
	response, err := s.ollama.Chat(prompt, caseContext)
	if err != nil {
		return nil, fmt.Errorf("erreur IA: %w", err)
	}

	// Parser la réponse pour extraire les scénarios
	scenarios := s.parseGeneratedScenarios(response, caseID, caseData)

	return scenarios, nil
}

// buildScenarioGenerationPrompt construit le prompt pour générer des scénarios
func (s *ScenarioService) buildScenarioGenerationPrompt(caseData *models.Case) string {
	var sb strings.Builder

	sb.WriteString("Tu es un enquêteur expérimenté. Analyse cette affaire et propose 3 à 5 scénarios \"What-If\" plausibles.\n\n")

	sb.WriteString("## AFFAIRE\n")
	sb.WriteString(fmt.Sprintf("Titre: %s\n", caseData.Name))
	sb.WriteString(fmt.Sprintf("Description: %s\n\n", caseData.Description))

	// Ajouter les suspects
	sb.WriteString("## SUSPECTS ET PERSONNES D'INTÉRÊT\n")
	for _, entity := range caseData.Entities {
		if entity.Role == models.RoleSuspect || entity.Role == models.RoleWitness {
			sb.WriteString(fmt.Sprintf("- %s (%s): %s\n", entity.Name, entity.Role, entity.Description))
		}
	}

	// Ajouter les preuves clés
	sb.WriteString("\n## PREUVES PRINCIPALES\n")
	for i, ev := range caseData.Evidence {
		if i >= 10 {
			break
		}
		sb.WriteString(fmt.Sprintf("- %s: %s\n", ev.Name, ev.Description))
	}

	// Ajouter les événements clés
	sb.WriteString("\n## CHRONOLOGIE\n")
	for i, evt := range caseData.Timeline {
		if i >= 8 {
			break
		}
		sb.WriteString(fmt.Sprintf("- %s: %s\n", evt.Title, evt.Description))
	}

	sb.WriteString("\n## FORMAT DE RÉPONSE\n")
	sb.WriteString("Réponds UNIQUEMENT avec une liste de scénarios au format suivant (un par ligne):\n")
	sb.WriteString("TYPE|ENTITY_ID|HYPOTHÈSE\n\n")
	sb.WriteString("Types disponibles: guilt, presence, motive, timeline, relation\n")
	sb.WriteString("ENTITY_ID: l'ID de l'entité cible (ex: ent-moreau-002) ou vide si non applicable\n\n")
	sb.WriteString("Exemple:\n")
	sb.WriteString("guilt|ent-moreau-002|Jean Moreau a empoisonné son oncle pour hériter de sa fortune\n")
	sb.WriteString("motive|ent-moreau-003|Élodie Dubois avait un mobile financier suite à la fraude présumée\n")
	sb.WriteString("timeline||L'empoisonnement a eu lieu entre 19h et 20h, pendant que l'alibi de Jean n'est pas vérifié\n\n")
	sb.WriteString("Génère 3 à 5 scénarios pertinents basés sur les éléments de l'affaire:\n")

	return sb.String()
}

// parseGeneratedScenarios parse la réponse du LLM et crée les scénarios
func (s *ScenarioService) parseGeneratedScenarios(response string, caseID string, caseData *models.Case) []*models.Scenario {
	var scenarios []*models.Scenario

	lines := strings.Split(response, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "Exemple") {
			continue
		}

		parts := strings.SplitN(line, "|", 3)
		if len(parts) < 3 {
			continue
		}

		assumptionType := strings.TrimSpace(strings.ToLower(parts[0]))
		targetEntityID := strings.TrimSpace(parts[1])
		assumption := strings.TrimSpace(parts[2])

		// Valider le type
		validTypes := map[string]bool{"guilt": true, "presence": true, "motive": true, "timeline": true, "relation": true}
		if !validTypes[assumptionType] {
			continue
		}

		if assumption == "" {
			continue
		}

		// Créer le scénario via la méthode standard
		req := models.ScenarioSimulationRequest{
			CaseID:         caseID,
			Assumption:     assumption,
			AssumptionType: assumptionType,
			TargetEntityID: targetEntityID,
		}

		scenario, err := s.CreateScenario(caseID, req)
		if err == nil && scenario != nil {
			scenarios = append(scenarios, scenario)
		}
	}

	return scenarios
}

// PropagateImplications propage les implications d'un scénario sur le graphe
func (s *ScenarioService) PropagateImplications(caseID, scenarioID string) (*models.GraphData, error) {
	s.mu.Lock()

	// Accès direct sans appeler GetScenario (évite deadlock)
	var scenario *models.Scenario
	if caseScenarios, ok := s.scenarios[caseID]; ok {
		scenario = caseScenarios[scenarioID]
	}
	if scenario == nil {
		s.mu.Unlock()
		return nil, fmt.Errorf("scénario non trouvé")
	}

	caseData, err := s.cases.GetCase(caseID)
	if err != nil {
		s.mu.Unlock()
		return nil, err
	}

	// Recalculer les implications
	s.calculateImplications(scenario, caseData)
	s.calculatePlausibility(scenario, caseData)

	// Régénérer le graphe modifié
	scenario.ModifiedGraph = s.generateModifiedGraph(scenario, caseData)
	scenario.UpdatedAt = time.Now()

	s.mu.Unlock()
	return scenario.ModifiedGraph, nil
}

// generateScenarioName génère un nom pour le scénario
func (s *ScenarioService) generateScenarioName(assumption, assumptionType string) string {
	typeLabels := map[string]string{
		"guilt":    "Culpabilité",
		"presence": "Présence",
		"motive":   "Mobile",
		"timeline": "Chronologie",
		"relation": "Relation",
	}

	typeLabel := typeLabels[assumptionType]
	if typeLabel == "" {
		typeLabel = "Hypothèse"
	}

	// Extraire un nom court de l'assumption
	shortName := assumption
	if len(shortName) > 50 {
		shortName = shortName[:47] + "..."
	}

	return fmt.Sprintf("%s: %s", typeLabel, shortName)
}

// calculateImplications calcule les implications d'un scénario
func (s *ScenarioService) calculateImplications(scenario *models.Scenario, caseData *models.Case) {
	implications := []models.ScenarioImplication{}

	switch scenario.AssumptionType {
	case "guilt":
		// Si X est coupable, quelles sont les implications ?
		implications = append(implications, s.calculateGuiltImplications(scenario, caseData)...)
	case "presence":
		// Si X était présent à Y, quelles sont les implications ?
		implications = append(implications, s.calculatePresenceImplications(scenario, caseData)...)
	case "motive":
		// Si X avait le mobile Y, quelles sont les implications ?
		implications = append(implications, s.calculateMotiveImplications(scenario, caseData)...)
	case "timeline":
		// Si l'événement X s'est produit à Y, quelles sont les implications ?
		implications = append(implications, s.calculateTimelineImplications(scenario, caseData)...)
	case "relation":
		// Si X et Y sont liés par Z, quelles sont les implications ?
		implications = append(implications, s.calculateRelationImplications(scenario, caseData)...)
	}

	scenario.Implications = implications
}

// calculateGuiltImplications calcule les implications si une entité est coupable
func (s *ScenarioService) calculateGuiltImplications(scenario *models.Scenario, caseData *models.Case) []models.ScenarioImplication {
	implications := []models.ScenarioImplication{}

	// Trouver l'entité cible
	var targetEntity *models.Entity
	for i := range caseData.Entities {
		if caseData.Entities[i].ID == scenario.TargetEntityID {
			targetEntity = &caseData.Entities[i]
			break
		}
	}

	if targetEntity == nil {
		return implications
	}

	// Implication 1: Changement de rôle vers suspect
	if targetEntity.Role != models.RoleSuspect {
		implications = append(implications, models.ScenarioImplication{
			ID:           uuid.New().String(),
			Type:         "change_role",
			Description:  fmt.Sprintf("%s devient suspect principal", targetEntity.Name),
			EntityID:     targetEntity.ID,
			Impact:       "high",
			Confidence:   100,
			AutoDetected: true,
		})
	}

	// Implication 2: Vérifier les alibis
	for _, event := range caseData.Timeline {
		for _, entityID := range event.Entities {
			if entityID == targetEntity.ID && event.Verified {
				implications = append(implications, models.ScenarioImplication{
					ID:           uuid.New().String(),
					Type:         "timeline_conflict",
					Description:  fmt.Sprintf("Conflit potentiel avec l'alibi: %s était présent à '%s'", targetEntity.Name, event.Title),
					EntityID:     targetEntity.ID,
					Impact:       "medium",
					Confidence:   70,
					AutoDetected: true,
				})
			}
		}
	}

	// Implication 3: Analyser les relations
	for _, entity := range caseData.Entities {
		for _, rel := range entity.Relations {
			if rel.FromID == targetEntity.ID || rel.ToID == targetEntity.ID {
				implications = append(implications, models.ScenarioImplication{
					ID:           uuid.New().String(),
					Type:         "relation_review",
					Description:  fmt.Sprintf("Relation à examiner: %s - %s", rel.Type, rel.Label),
					EntityID:     targetEntity.ID,
					RelationID:   rel.ID,
					Impact:       "medium",
					Confidence:   60,
					AutoDetected: true,
				})
			}
		}
	}

	return implications
}

// calculatePresenceImplications calcule les implications de présence
func (s *ScenarioService) calculatePresenceImplications(scenario *models.Scenario, caseData *models.Case) []models.ScenarioImplication {
	implications := []models.ScenarioImplication{}

	// Analyser si la présence contredit d'autres événements
	for _, event := range caseData.Timeline {
		for _, entityID := range event.Entities {
			if entityID == scenario.TargetEntityID {
				implications = append(implications, models.ScenarioImplication{
					ID:           uuid.New().String(),
					Type:         "presence_verification",
					Description:  fmt.Sprintf("Vérifier la cohérence avec l'événement: %s", event.Title),
					EntityID:     scenario.TargetEntityID,
					Impact:       "medium",
					Confidence:   50,
					AutoDetected: true,
				})
			}
		}
	}

	return implications
}

// calculateMotiveImplications calcule les implications de mobile
func (s *ScenarioService) calculateMotiveImplications(scenario *models.Scenario, caseData *models.Case) []models.ScenarioImplication {
	implications := []models.ScenarioImplication{}

	// Identifier les preuves liées au mobile
	for _, evidence := range caseData.Evidence {
		for _, linkedID := range evidence.LinkedEntities {
			if linkedID == scenario.TargetEntityID {
				implications = append(implications, models.ScenarioImplication{
					ID:           uuid.New().String(),
					Type:         "evidence_review",
					Description:  fmt.Sprintf("Preuve à réévaluer avec ce mobile: %s", evidence.Name),
					EntityID:     scenario.TargetEntityID,
					Impact:       "high",
					Confidence:   65,
					AutoDetected: true,
				})
			}
		}
	}

	return implications
}

// calculateTimelineImplications calcule les implications temporelles
func (s *ScenarioService) calculateTimelineImplications(scenario *models.Scenario, caseData *models.Case) []models.ScenarioImplication {
	implications := []models.ScenarioImplication{}

	// Détecter les conflits temporels
	for i, event1 := range caseData.Timeline {
		for j, event2 := range caseData.Timeline {
			if i >= j {
				continue
			}

			// Vérifier si les événements sont proches temporellement
			timeDiff := event2.Timestamp.Sub(event1.Timestamp).Hours()
			if timeDiff > 0 && timeDiff < 24 {
				// Vérifier si les mêmes entités sont impliquées
				for _, e1 := range event1.Entities {
					for _, e2 := range event2.Entities {
						if e1 == e2 {
							implications = append(implications, models.ScenarioImplication{
								ID:           uuid.New().String(),
								Type:         "temporal_proximity",
								Description:  fmt.Sprintf("Proximité temporelle entre '%s' et '%s'", event1.Title, event2.Title),
								Impact:       "medium",
								Confidence:   55,
								AutoDetected: true,
							})
						}
					}
				}
			}
		}
	}

	return implications
}

// calculateRelationImplications calcule les implications relationnelles
func (s *ScenarioService) calculateRelationImplications(scenario *models.Scenario, caseData *models.Case) []models.ScenarioImplication {
	implications := []models.ScenarioImplication{}

	// Analyser les relations existantes
	for _, entity := range caseData.Entities {
		if entity.ID == scenario.TargetEntityID {
			for _, rel := range entity.Relations {
				implications = append(implications, models.ScenarioImplication{
					ID:           uuid.New().String(),
					Type:         "relation_impact",
					Description:  fmt.Sprintf("Impact sur la relation '%s' avec '%s'", rel.Type, rel.ToID),
					EntityID:     entity.ID,
					RelationID:   rel.ID,
					Impact:       "medium",
					Confidence:   60,
					AutoDetected: true,
				})
			}
		}
	}

	return implications
}

// calculatePlausibility calcule le score de plausibilité
func (s *ScenarioService) calculatePlausibility(scenario *models.Scenario, caseData *models.Case) {
	score := 50 // Score de base

	// Réinitialiser les listes pour éviter les doublons lors des recalculs
	scenario.SupportingFacts = []string{}
	scenario.ContradictingFacts = []string{}

	// Facteurs positifs
	supportingCount := 0
	contradictingCount := 0

	// Analyser les preuves
	for _, evidence := range caseData.Evidence {
		for _, linkedID := range evidence.LinkedEntities {
			if linkedID == scenario.TargetEntityID {
				if evidence.Reliability >= 7 {
					supportingCount++
					scenario.SupportingFacts = append(scenario.SupportingFacts, evidence.Name)
				} else if evidence.Reliability <= 3 {
					contradictingCount++
					scenario.ContradictingFacts = append(scenario.ContradictingFacts, evidence.Name)
				}
			}
		}
	}

	// Analyser la cohérence avec les hypothèses existantes
	for _, hypothesis := range caseData.Hypotheses {
		if strings.Contains(strings.ToLower(hypothesis.Description), strings.ToLower(scenario.Assumption)) {
			if hypothesis.Status == models.HypothesisSupported {
				supportingCount += 2
				score += 10
			} else if hypothesis.Status == models.HypothesisRefuted {
				contradictingCount += 2
				score -= 15
			}
		}
	}

	// Ajuster le score
	score += supportingCount * 5
	score -= contradictingCount * 8

	// Borner le score entre 0 et 100
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	scenario.PlausibilityScore = score
}

// normalizeID normalise un ID pour permettre la comparaison entre formats
// (ent-moreau-001 vs ent_moreau_001)
func normalizeID(id string) string {
	return strings.ReplaceAll(id, "_", "-")
}

// generateModifiedGraph génère le graphe modifié pour un scénario
func (s *ScenarioService) generateModifiedGraph(scenario *models.Scenario, caseData *models.Case) *models.GraphData {
	nodes := []models.GraphNode{}
	edges := []models.GraphEdge{}

	// Collecter les IDs des entités impactées par les implications (normalisés)
	impactedEntities := make(map[string]bool)
	if scenario.TargetEntityID != "" {
		// Ajouter à la fois l'ID original et l'ID normalisé
		impactedEntities[scenario.TargetEntityID] = true
		impactedEntities[normalizeID(scenario.TargetEntityID)] = true
		fmt.Printf("[generateModifiedGraph] Added TargetEntityID to impactedEntities: %s (normalized: %s)\n", scenario.TargetEntityID, normalizeID(scenario.TargetEntityID))
	} else {
		fmt.Println("[generateModifiedGraph] WARNING: TargetEntityID is empty!")
	}
	for _, impl := range scenario.Implications {
		if impl.EntityID != "" {
			impactedEntities[impl.EntityID] = true
			impactedEntities[normalizeID(impl.EntityID)] = true
			fmt.Printf("[generateModifiedGraph] Added EntityID from implication: %s (normalized: %s)\n", impl.EntityID, normalizeID(impl.EntityID))
		}
	}
	fmt.Printf("[generateModifiedGraph] Total impactedEntities: %d\n", len(impactedEntities))

	// Copier les nœuds existants
	for _, entity := range caseData.Entities {
		node := models.GraphNode{
			ID:    entity.ID,
			Label: entity.Name,
			Type:  string(entity.Type),
			Role:  string(entity.Role),
		}

		// Vérifier si l'entité est impactée (avec normalisation)
		isImpacted := impactedEntities[entity.ID] || impactedEntities[normalizeID(entity.ID)]
		fmt.Printf("[generateModifiedGraph] Checking entity %s (%s), isImpacted: %v\n", entity.ID, entity.Name, isImpacted)

		// Marquer les nœuds impactés par le scénario
		if isImpacted {
			// Modifier le rôle si c'est l'entité cible d'un scénario de culpabilité
			normalizedTargetID := normalizeID(scenario.TargetEntityID)
			normalizedEntityID := normalizeID(entity.ID)
			if scenario.AssumptionType == "guilt" && (entity.ID == scenario.TargetEntityID || normalizedEntityID == normalizedTargetID) {
				node.Role = "suspect"
			}
			node.Data = map[string]string{
				"scenario_modified": "true",
				"original_role":     string(entity.Role),
			}
			fmt.Printf("[generateModifiedGraph] MARKED node %s as scenario_modified!\n", entity.ID)
		}

		nodes = append(nodes, node)
	}

	// Copier les arêtes existantes
	for _, entity := range caseData.Entities {
		for _, rel := range entity.Relations {
			edges = append(edges, models.GraphEdge{
				From:  rel.FromID,
				To:    rel.ToID,
				Label: rel.Label,
				Type:  rel.Type,
			})
		}
	}

	// Ajouter des arêtes d'implication
	for _, impl := range scenario.Implications {
		if impl.Type == "add_relation" && impl.EntityID != "" {
			edges = append(edges, models.GraphEdge{
				From:    impl.EntityID,
				To:      scenario.TargetEntityID,
				Label:   impl.Description,
				Type:    "implication",
				Context: "scenario:" + scenario.ID,
			})
		}
	}

	return &models.GraphData{
		Nodes: nodes,
		Edges: edges,
	}
}

// findCommonFacts trouve les faits communs entre deux scénarios
func (s *ScenarioService) findCommonFacts(scenario1, scenario2 *models.Scenario) []string {
	common := []string{}

	factMap := make(map[string]bool)
	for _, fact := range scenario1.SupportingFacts {
		factMap[fact] = true
	}

	for _, fact := range scenario2.SupportingFacts {
		if factMap[fact] {
			common = append(common, fact)
		}
	}

	return common
}

// findDifferentFacts trouve les différences entre deux scénarios
func (s *ScenarioService) findDifferentFacts(scenario1, scenario2 *models.Scenario) []models.ScenarioDifference {
	differences := []models.ScenarioDifference{}

	// Comparer les scores de plausibilité
	if scenario1.PlausibilityScore != scenario2.PlausibilityScore {
		differences = append(differences, models.ScenarioDifference{
			Aspect:         "Score de plausibilité",
			Scenario1Value: fmt.Sprintf("%d%%", scenario1.PlausibilityScore),
			Scenario2Value: fmt.Sprintf("%d%%", scenario2.PlausibilityScore),
			Significance:   s.getSignificance(abs(scenario1.PlausibilityScore - scenario2.PlausibilityScore)),
		})
	}

	// Comparer le nombre d'implications
	if len(scenario1.Implications) != len(scenario2.Implications) {
		differences = append(differences, models.ScenarioDifference{
			Aspect:         "Nombre d'implications",
			Scenario1Value: fmt.Sprintf("%d", len(scenario1.Implications)),
			Scenario2Value: fmt.Sprintf("%d", len(scenario2.Implications)),
			Significance:   "medium",
		})
	}

	// Comparer les types d'assomption
	if scenario1.AssumptionType != scenario2.AssumptionType {
		differences = append(differences, models.ScenarioDifference{
			Aspect:         "Type d'hypothèse",
			Scenario1Value: scenario1.AssumptionType,
			Scenario2Value: scenario2.AssumptionType,
			Significance:   "high",
		})
	}

	return differences
}

// calculateGraphDifferences calcule les différences entre deux graphes
func (s *ScenarioService) calculateGraphDifferences(graph1, graph2 *models.GraphData) *models.GraphDifference {
	diff := &models.GraphDifference{
		AddedNodes:    []models.GraphNode{},
		RemovedNodes:  []models.GraphNode{},
		AddedEdges:    []models.GraphEdge{},
		RemovedEdges:  []models.GraphEdge{},
		ModifiedNodes: []models.GraphNode{},
	}

	// Créer des maps pour comparaison
	nodes1 := make(map[string]models.GraphNode)
	for _, node := range graph1.Nodes {
		nodes1[node.ID] = node
	}

	nodes2 := make(map[string]models.GraphNode)
	for _, node := range graph2.Nodes {
		nodes2[node.ID] = node
	}

	// Trouver les nœuds ajoutés/supprimés/modifiés
	for id, node := range nodes2 {
		if _, exists := nodes1[id]; !exists {
			diff.AddedNodes = append(diff.AddedNodes, node)
		} else if nodes1[id].Role != node.Role {
			diff.ModifiedNodes = append(diff.ModifiedNodes, node)
		}
	}

	for id, node := range nodes1 {
		if _, exists := nodes2[id]; !exists {
			diff.RemovedNodes = append(diff.RemovedNodes, node)
		}
	}

	return diff
}

// generateRecommendation génère une recommandation basée sur la comparaison
func (s *ScenarioService) generateRecommendation(comparison *models.ScenarioComparison, scenario1, scenario2 *models.Scenario) string {
	if comparison.PlausibilityDelta > 20 {
		return fmt.Sprintf("Le scénario '%s' semble plus plausible avec un écart de %d points. Il est recommandé de concentrer l'investigation sur cette piste.",
			scenario1.Name, comparison.PlausibilityDelta)
	} else if comparison.PlausibilityDelta < -20 {
		return fmt.Sprintf("Le scénario '%s' semble plus plausible avec un écart de %d points. Il est recommandé de concentrer l'investigation sur cette piste.",
			scenario2.Name, -comparison.PlausibilityDelta)
	}

	return "Les deux scénarios ont des scores de plausibilité similaires. Il est recommandé d'approfondir l'investigation pour les départager."
}

// buildAIPrompt construit le prompt pour l'analyse IA
func (s *ScenarioService) buildAIPrompt(scenario *models.Scenario, caseData *models.Case) string {
	var sb strings.Builder

	sb.WriteString("Analysez le scénario suivant pour l'affaire d'investigation:\n\n")
	sb.WriteString(fmt.Sprintf("**Affaire**: %s\n", caseData.Name))
	sb.WriteString(fmt.Sprintf("**Description de l'affaire**: %s\n\n", caseData.Description))

	sb.WriteString("## Scénario What-If\n")
	sb.WriteString(fmt.Sprintf("**Hypothèse**: %s\n", scenario.Assumption))
	sb.WriteString(fmt.Sprintf("**Type**: %s\n", scenario.AssumptionType))
	sb.WriteString(fmt.Sprintf("**Score de plausibilité actuel**: %d%%\n\n", scenario.PlausibilityScore))

	sb.WriteString("## Implications détectées\n")
	for _, impl := range scenario.Implications {
		sb.WriteString(fmt.Sprintf("- [%s] %s (Impact: %s, Confiance: %d%%)\n",
			impl.Type, impl.Description, impl.Impact, impl.Confidence))
	}

	sb.WriteString("\n## Faits à l'appui\n")
	for _, fact := range scenario.SupportingFacts {
		sb.WriteString(fmt.Sprintf("- %s\n", fact))
	}

	sb.WriteString("\n## Faits contradictoires\n")
	for _, fact := range scenario.ContradictingFacts {
		sb.WriteString(fmt.Sprintf("- %s\n", fact))
	}

	sb.WriteString("\n## Demande d'analyse\n")
	sb.WriteString("1. Évaluez la plausibilité de ce scénario\n")
	sb.WriteString("2. Identifiez les points forts et les faiblesses\n")
	sb.WriteString("3. Suggérez des pistes d'investigation pour confirmer ou infirmer ce scénario\n")
	sb.WriteString("4. Identifiez les questions clés à résoudre\n")

	return sb.String()
}

// getSignificance retourne le niveau de signification basé sur la différence
func (s *ScenarioService) getSignificance(diff int) string {
	if diff >= 30 {
		return "high"
	} else if diff >= 15 {
		return "medium"
	}
	return "low"
}

// abs retourne la valeur absolue
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
