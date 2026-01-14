package services

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"forensicinvestigator/internal/models"

	"github.com/google/uuid"
)

// =========================================
// Fonctions statistiques avancées
// =========================================

// calculateMean calcule la moyenne d'une série de valeurs
func calculateMean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	var sum float64
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

// calculateStdDev calcule l'écart-type d'une série de valeurs
func calculateStdDev(values []float64, mean float64) float64 {
	if len(values) < 2 {
		return 0
	}
	var sumSquares float64
	for _, v := range values {
		diff := v - mean
		sumSquares += diff * diff
	}
	return math.Sqrt(sumSquares / float64(len(values)-1))
}

// calculateZScore calcule le Z-score d'une valeur
func calculateZScore(value, mean, stdDev float64) float64 {
	if stdDev == 0 {
		return 0
	}
	return (value - mean) / stdDev
}

// calculateMedian calcule la médiane d'une série de valeurs
func calculateMedian(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)
	n := len(sorted)
	if n%2 == 0 {
		return (sorted[n/2-1] + sorted[n/2]) / 2
	}
	return sorted[n/2]
}

// calculateMAD calcule la Median Absolute Deviation (plus robuste que stdDev)
func calculateMAD(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	median := calculateMedian(values)
	deviations := make([]float64, len(values))
	for i, v := range values {
		deviations[i] = math.Abs(v - median)
	}
	return calculateMedian(deviations)
}

// calculateModifiedZScore calcule le Z-score modifié basé sur la MAD
func calculateModifiedZScore(value, median, mad float64) float64 {
	if mad == 0 {
		return 0
	}
	// Facteur 0.6745 pour normaliser à une distribution normale standard
	return 0.6745 * (value - median) / mad
}

// AnomalyService gère la détection d'anomalies
type AnomalyService struct {
	anomalies map[string]map[string]*models.Anomaly // caseID -> anomalyID -> Anomaly
	alerts    map[string][]*models.AnomalyAlert     // caseID -> alerts
	configs   map[string]*models.AnomalyDetectionConfig // caseID -> config
	mu        sync.RWMutex
	cases     *CaseService
	ollama    *OllamaService
}

// NewAnomalyService crée un nouveau service d'anomalies
func NewAnomalyService(cases *CaseService, ollama *OllamaService) *AnomalyService {
	return &AnomalyService{
		anomalies: make(map[string]map[string]*models.Anomaly),
		alerts:    make(map[string][]*models.AnomalyAlert),
		configs:   make(map[string]*models.AnomalyDetectionConfig),
		cases:     cases,
		ollama:    ollama,
	}
}

// DetectAnomalies lance une détection d'anomalies pour un cas
func (a *AnomalyService) DetectAnomalies(caseID string) (*models.AnomalyDetectionResult, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	caseData, err := a.cases.GetCase(caseID)
	if err != nil {
		return nil, fmt.Errorf("cas non trouvé: %w", err)
	}

	// Obtenir ou créer la configuration
	config := a.getOrCreateConfig(caseID)

	// Initialiser le stockage si nécessaire
	if a.anomalies[caseID] == nil {
		a.anomalies[caseID] = make(map[string]*models.Anomaly)
	}

	result := &models.AnomalyDetectionResult{
		CaseID:     caseID,
		DetectedAt: time.Now(),
		Anomalies:  []models.Anomaly{},
		Alerts:     []models.AnomalyAlert{},
	}

	// Détecter les différents types d'anomalies
	if config.EnableTimeline {
		anomalies := a.detectTimelineAnomalies(caseData)
		result.Anomalies = append(result.Anomalies, anomalies...)
	}

	if config.EnableFinancial {
		anomalies := a.detectFinancialAnomalies(caseData)
		result.Anomalies = append(result.Anomalies, anomalies...)
	}

	if config.EnableCommunication {
		anomalies := a.detectCommunicationAnomalies(caseData)
		result.Anomalies = append(result.Anomalies, anomalies...)
	}

	if config.EnableBehavior {
		anomalies := a.detectBehaviorAnomalies(caseData)
		result.Anomalies = append(result.Anomalies, anomalies...)
	}

	if config.EnableLocation {
		anomalies := a.detectLocationAnomalies(caseData)
		result.Anomalies = append(result.Anomalies, anomalies...)
	}

	if config.EnableRelation {
		anomalies := a.detectRelationAnomalies(caseData)
		result.Anomalies = append(result.Anomalies, anomalies...)
	}

	if config.EnablePattern {
		anomalies := a.detectPatternAnomalies(caseData)
		result.Anomalies = append(result.Anomalies, anomalies...)
	}

	// Détecter les corrélations croisées entre anomalies
	crossCorrelationAnomalies := a.detectCrossCorrelations(caseData, result.Anomalies)
	result.Anomalies = append(result.Anomalies, crossCorrelationAnomalies...)

	// Appliquer le scoring bayésien pour affiner les confiances
	result.Anomalies = a.applyBayesianScoring(result.Anomalies, caseData)

	// Filtrer par confiance minimum et dédupliquer
	filteredAnomalies := []models.Anomaly{}
	for _, anomaly := range result.Anomalies {
		if anomaly.Confidence >= config.MinConfidence {
			// Vérifier si une anomalie similaire existe déjà
			existingID := a.findExistingAnomaly(caseID, anomaly)
			if existingID != "" {
				// L'anomalie existe déjà, on la récupère avec son ID existant
				existingAnomaly := a.anomalies[caseID][existingID]
				existingAnomaly.IsNew = false
				filteredAnomalies = append(filteredAnomalies, *existingAnomaly)
			} else {
				// Nouvelle anomalie
				anomaly.IsNew = true
				a.anomalies[caseID][anomaly.ID] = &anomaly
				filteredAnomalies = append(filteredAnomalies, anomaly)

				// Générer des alertes uniquement pour les nouvelles anomalies
				if config.AutoAlert && a.shouldAlert(anomaly, config) {
					alert := a.createAlert(caseID, &anomaly)
					result.Alerts = append(result.Alerts, alert)
				}
			}
		}
	}

	result.Anomalies = filteredAnomalies
	result.TotalAnomalies = len(filteredAnomalies)
	result.NewAnomalies = a.countNew(filteredAnomalies)
	result.CriticalCount = a.countBySeverity(filteredAnomalies, models.SeverityCritical)
	result.HighCount = a.countBySeverity(filteredAnomalies, models.SeverityHigh)
	result.MediumCount = a.countBySeverity(filteredAnomalies, models.SeverityMedium)
	result.LowCount = a.countBySeverity(filteredAnomalies, models.SeverityLow)
	result.Summary = a.generateSummary(result)

	return result, nil
}

// GetAnomalies retourne toutes les anomalies d'un cas
func (a *AnomalyService) GetAnomalies(caseID string) []models.Anomaly {
	a.mu.RLock()
	defer a.mu.RUnlock()

	anomalies := []models.Anomaly{}
	if caseAnomalies, ok := a.anomalies[caseID]; ok {
		for _, anomaly := range caseAnomalies {
			anomalies = append(anomalies, *anomaly)
		}
	}

	// Trier par sévérité puis par date
	sort.Slice(anomalies, func(i, j int) bool {
		if severityOrder(anomalies[i].Severity) != severityOrder(anomalies[j].Severity) {
			return severityOrder(anomalies[i].Severity) < severityOrder(anomalies[j].Severity)
		}
		return anomalies[i].DetectedAt.After(anomalies[j].DetectedAt)
	})

	return anomalies
}

// GetAnomaly retourne une anomalie spécifique
func (a *AnomalyService) GetAnomaly(caseID, anomalyID string) (*models.Anomaly, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if caseAnomalies, ok := a.anomalies[caseID]; ok {
		if anomaly, ok := caseAnomalies[anomalyID]; ok {
			return anomaly, nil
		}
	}
	return nil, fmt.Errorf("anomalie non trouvée")
}

// AcknowledgeAnomaly marque une anomalie comme acquittée
func (a *AnomalyService) AcknowledgeAnomaly(caseID, anomalyID string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if caseAnomalies, ok := a.anomalies[caseID]; ok {
		if anomaly, ok := caseAnomalies[anomalyID]; ok {
			anomaly.IsAcknowledged = true
			anomaly.IsNew = false
			return nil
		}
	}
	return fmt.Errorf("anomalie non trouvée")
}

// GetStatistics retourne les statistiques d'anomalies pour un cas
func (a *AnomalyService) GetStatistics(caseID string) *models.AnomalyStatistics {
	a.mu.RLock()
	defer a.mu.RUnlock()

	stats := &models.AnomalyStatistics{
		CaseID:     caseID,
		ByType:     make(map[models.AnomalyType]int),
		BySeverity: make(map[models.AnomalySeverity]int),
	}

	anomalies := a.GetAnomalies(caseID)
	stats.TotalDetected = len(anomalies)

	var totalConfidence int
	for _, anomaly := range anomalies {
		if anomaly.IsAcknowledged {
			stats.Acknowledged++
		} else {
			stats.Pending++
		}

		stats.ByType[anomaly.Type]++
		stats.BySeverity[anomaly.Severity]++
		totalConfidence += anomaly.Confidence

		if stats.LastDetection.Before(anomaly.DetectedAt) {
			stats.LastDetection = anomaly.DetectedAt
		}
	}

	if stats.TotalDetected > 0 {
		stats.AvgConfidence = float64(totalConfidence) / float64(stats.TotalDetected)
	}

	stats.TrendDirection = "stable"

	return stats
}

// GetAlerts retourne les alertes d'un cas
func (a *AnomalyService) GetAlerts(caseID string, unreadOnly bool) []models.AnomalyAlert {
	a.mu.RLock()
	defer a.mu.RUnlock()

	alerts := []models.AnomalyAlert{}
	if caseAlerts, ok := a.alerts[caseID]; ok {
		for _, alert := range caseAlerts {
			if !unreadOnly || !alert.IsRead {
				alerts = append(alerts, *alert)
			}
		}
	}

	return alerts
}

// MarkAlertRead marque une alerte comme lue
func (a *AnomalyService) MarkAlertRead(caseID, alertID string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if caseAlerts, ok := a.alerts[caseID]; ok {
		for _, alert := range caseAlerts {
			if alert.ID == alertID {
				alert.IsRead = true
				return nil
			}
		}
	}
	return fmt.Errorf("alerte non trouvée")
}

// UpdateConfig met à jour la configuration de détection
func (a *AnomalyService) UpdateConfig(config *models.AnomalyDetectionConfig) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.configs[config.CaseID] = config
	return nil
}

// GetConfig retourne la configuration de détection
func (a *AnomalyService) GetConfig(caseID string) *models.AnomalyDetectionConfig {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.getOrCreateConfig(caseID)
}

// ExplainAnomaly génère une explication IA pour une anomalie
func (a *AnomalyService) ExplainAnomaly(caseID, anomalyID string) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	anomaly, err := a.GetAnomaly(caseID, anomalyID)
	if err != nil {
		return "", err
	}

	caseData, err := a.cases.GetCase(caseID)
	if err != nil {
		return "", err
	}

	prompt := a.buildExplanationPrompt(anomaly, caseData)
	caseContext := fmt.Sprintf("Affaire: %s\nDescription: %s", caseData.Name, caseData.Description)
	explanation, err := a.ollama.Chat(prompt, caseContext)
	if err != nil {
		return "", fmt.Errorf("erreur analyse IA: %w", err)
	}

	anomaly.AIExplanation = explanation

	return explanation, nil
}

// =========================================
// Méthodes de détection d'anomalies
// =========================================

// detectTimelineAnomalies détecte les anomalies temporelles avec seuils adaptatifs (Z-score)
func (a *AnomalyService) detectTimelineAnomalies(caseData *models.Case) []models.Anomaly {
	anomalies := []models.Anomaly{}

	// Trier les événements par date
	events := make([]models.Event, len(caseData.Timeline))
	copy(events, caseData.Timeline)
	sort.Slice(events, func(i, j int) bool {
		return events[i].Timestamp.Before(events[j].Timestamp)
	})

	if len(events) < 2 {
		return anomalies
	}

	// Calculer tous les gaps entre événements consécutifs
	gaps := make([]float64, len(events)-1)
	for i := 1; i < len(events); i++ {
		gaps[i-1] = events[i].Timestamp.Sub(events[i-1].Timestamp).Hours()
	}

	// Calculer les statistiques pour les seuils adaptatifs
	median := calculateMedian(gaps)
	mad := calculateMAD(gaps)
	mean := calculateMean(gaps)
	stdDev := calculateStdDev(gaps, mean)

	// Seuil adaptatif: utiliser Z-score modifié (plus robuste aux outliers)
	// Un gap est anormal si son Z-score modifié > 2.5 (ou Z-score standard > 2)
	const zScoreThreshold = 2.5
	const minGapHours = 24.0 // Minimum 24h pour éviter les faux positifs

	for i := 1; i < len(events); i++ {
		gap := events[i].Timestamp.Sub(events[i-1].Timestamp)
		gapHours := gap.Hours()

		// Calculer le Z-score modifié pour ce gap
		modifiedZScore := calculateModifiedZScore(gapHours, median, mad)
		standardZScore := calculateZScore(gapHours, mean, stdDev)

		// Utiliser le Z-score le plus significatif
		effectiveZScore := math.Max(math.Abs(modifiedZScore), math.Abs(standardZScore))

		// Détecter si le gap est statistiquement anormal
		if effectiveZScore > zScoreThreshold && gapHours > minGapHours {
			// Calculer la confiance basée sur le Z-score
			confidence := int(math.Min(50+effectiveZScore*15, 95))

			// Calculer la sévérité basée sur le Z-score
			severity := a.calculateAdaptiveGapSeverity(effectiveZScore, gapHours)

			anomalies = append(anomalies, models.Anomaly{
				ID:          uuid.New().String(),
				CaseID:      caseData.ID,
				Type:        models.AnomalyTimeline,
				Severity:    severity,
				Title:       "Gap temporel statistiquement anormal",
				Description: fmt.Sprintf("Période de %.1f jours sans événement (Z-score: %.2f, médiane: %.1f jours) entre '%s' et '%s'",
					gapHours/24, effectiveZScore, median/24, events[i-1].Title, events[i].Title),
				DetectedAt: time.Now(),
				EventIDs:   []string{events[i-1].ID, events[i].ID},
				Confidence: confidence,
				Details: map[string]interface{}{
					"gap_hours":         gapHours,
					"gap_days":          gapHours / 24,
					"z_score":           effectiveZScore,
					"modified_z_score":  modifiedZScore,
					"standard_z_score":  standardZScore,
					"median_gap_hours":  median,
					"mad":               mad,
					"mean_gap_hours":    mean,
					"std_dev":           stdDev,
					"event_before":      events[i-1].Title,
					"event_after":       events[i].Title,
					"detection_method":  "adaptive_zscore",
				},
			})
		}
	}

	// Détecter les contradictions temporelles avec seuils adaptatifs basés sur la distance
	entityEvents := make(map[string][]models.Event)
	for _, event := range events {
		for _, entityID := range event.Entities {
			entityEvents[entityID] = append(entityEvents[entityID], event)
		}
	}

	for entityID, evts := range entityEvents {
		for i := 1; i < len(evts); i++ {
			timeDiff := evts[i].Timestamp.Sub(evts[i-1].Timestamp)

			// Vérifier si les locations sont différentes et non vides
			if evts[i].Location == evts[i-1].Location || evts[i].Location == "" || evts[i-1].Location == "" {
				continue
			}

			// Seuil adaptatif basé sur la "distance" des locations (heuristique simple)
			// Pour une vraie implémentation, utiliser des coordonnées GPS
			maxTravelTime := a.estimateTravelTime(evts[i-1].Location, evts[i].Location)

			if timeDiff.Hours() < maxTravelTime {
				// Calculer la sévérité et confiance basées sur l'écart
				ratio := timeDiff.Hours() / maxTravelTime
				confidence := int(math.Min(95-ratio*30, 95))
				severity := models.SeverityHigh
				if ratio < 0.25 {
					severity = models.SeverityCritical
					confidence = 95
				}

				anomalies = append(anomalies, models.Anomaly{
					ID:          uuid.New().String(),
					CaseID:      caseData.ID,
					Type:        models.AnomalyTimeline,
					Severity:    severity,
					Title:       "Conflit temporel potentiel",
					Description: fmt.Sprintf("Entité présente à '%s' puis '%s' en %.0f minutes (temps minimum estimé: %.0f min)",
						evts[i-1].Location, evts[i].Location, timeDiff.Minutes(), maxTravelTime*60),
					DetectedAt: time.Now(),
					EntityIDs:  []string{entityID},
					EventIDs:   []string{evts[i-1].ID, evts[i].ID},
					Confidence: confidence,
					Details: map[string]interface{}{
						"time_diff_minutes":    timeDiff.Minutes(),
						"estimated_travel_min": maxTravelTime * 60,
						"location1":            evts[i-1].Location,
						"location2":            evts[i].Location,
						"time_ratio":           ratio,
						"detection_method":     "adaptive_travel_time",
					},
				})
			}
		}
	}

	return anomalies
}

// calculateAdaptiveGapSeverity calcule la sévérité basée sur le Z-score
func (a *AnomalyService) calculateAdaptiveGapSeverity(zScore, gapHours float64) models.AnomalySeverity {
	// Combinaison du Z-score et de la durée absolue
	if zScore > 4 || gapHours > 720 { // 30 jours
		return models.SeverityCritical
	} else if zScore > 3 || gapHours > 336 { // 14 jours
		return models.SeverityHigh
	} else if zScore > 2.5 || gapHours > 168 { // 7 jours
		return models.SeverityMedium
	}
	return models.SeverityLow
}

// estimateTravelTime estime le temps de trajet entre deux lieux (en heures)
func (a *AnomalyService) estimateTravelTime(loc1, loc2 string) float64 {
	// Heuristique simple basée sur les mots-clés dans les noms de lieux
	// Pour une vraie implémentation, utiliser une API de géolocalisation
	loc1Lower := strings.ToLower(loc1)
	loc2Lower := strings.ToLower(loc2)

	// Même ville/quartier
	sameCity := false
	cities := []string{"paris", "lyon", "marseille", "bordeaux", "lille", "toulouse", "nice", "nantes"}
	for _, city := range cities {
		if strings.Contains(loc1Lower, city) && strings.Contains(loc2Lower, city) {
			sameCity = true
			break
		}
	}

	if sameCity {
		return 1.0 // 1 heure pour traverser une ville
	}

	// Différentes villes en France
	if strings.Contains(loc1Lower, "france") || strings.Contains(loc2Lower, "france") {
		return 4.0 // 4 heures en moyenne
	}

	// Par défaut, 2 heures (distance moyenne)
	return 2.0
}

// Regex patterns pour l'extraction de montants financiers
var (
	// Pattern pour les montants en euros: 1000€, 1 000€, 1,000.00€, 1.000,00 €, etc.
	amountPatternEuro = regexp.MustCompile(`(\d{1,3}(?:[\s.,]\d{3})*(?:[.,]\d{1,2})?)\s*(?:€|EUR|euros?)\b`)
	// Pattern pour les montants avec symbole devant: €1000, € 1,000.00
	amountPatternEuroPre = regexp.MustCompile(`(?:€|EUR)\s*(\d{1,3}(?:[\s.,]\d{3})*(?:[.,]\d{1,2})?)`)
	// Pattern pour les montants écrits: "mille euros", "million d'euros"
	amountPatternWords = regexp.MustCompile(`(\d+(?:[.,]\d+)?)\s*(mille|million|milliard)s?\s*(?:d')?(?:€|euros?)?`)
	// Pattern générique pour les montants sans devise explicite mais dans un contexte financier
	amountPatternGeneric = regexp.MustCompile(`(?:montant|somme|total|virement|transaction|paiement|versement)\s*(?:de|:)?\s*(\d{1,3}(?:[\s.,]\d{3})*(?:[.,]\d{1,2})?)`)
)

// extractAmounts extrait tous les montants financiers d'un texte
func extractAmounts(text string) []float64 {
	amounts := []float64{}
	textLower := strings.ToLower(text)

	// Extraire les montants avec le pattern euro suffixe
	matches := amountPatternEuro.FindAllStringSubmatch(textLower, -1)
	for _, match := range matches {
		if len(match) > 1 {
			if amount := parseAmount(match[1]); amount > 0 {
				amounts = append(amounts, amount)
			}
		}
	}

	// Extraire les montants avec le pattern euro préfixe
	matches = amountPatternEuroPre.FindAllStringSubmatch(textLower, -1)
	for _, match := range matches {
		if len(match) > 1 {
			if amount := parseAmount(match[1]); amount > 0 {
				amounts = append(amounts, amount)
			}
		}
	}

	// Extraire les montants écrits en mots
	matches = amountPatternWords.FindAllStringSubmatch(textLower, -1)
	for _, match := range matches {
		if len(match) > 2 {
			baseAmount := parseAmount(match[1])
			multiplier := 1.0
			switch match[2] {
			case "mille":
				multiplier = 1000
			case "million":
				multiplier = 1000000
			case "milliard":
				multiplier = 1000000000
			}
			if baseAmount > 0 {
				amounts = append(amounts, baseAmount*multiplier)
			}
		}
	}

	// Extraire les montants génériques dans un contexte financier
	matches = amountPatternGeneric.FindAllStringSubmatch(textLower, -1)
	for _, match := range matches {
		if len(match) > 1 {
			if amount := parseAmount(match[1]); amount > 0 {
				amounts = append(amounts, amount)
			}
		}
	}

	return amounts
}

// parseAmount convertit une chaîne de montant en float64
func parseAmount(s string) float64 {
	// Nettoyer la chaîne
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, " ", "")

	// Détecter le format (européen vs américain)
	// Format européen: 1.234,56 ou 1 234,56
	// Format américain: 1,234.56

	hasComma := strings.Contains(s, ",")
	hasDot := strings.Contains(s, ".")

	if hasComma && hasDot {
		// Déterminer lequel est le séparateur décimal (le dernier)
		lastComma := strings.LastIndex(s, ",")
		lastDot := strings.LastIndex(s, ".")

		if lastComma > lastDot {
			// Format européen: 1.234,56
			s = strings.ReplaceAll(s, ".", "")
			s = strings.ReplaceAll(s, ",", ".")
		} else {
			// Format américain: 1,234.56
			s = strings.ReplaceAll(s, ",", "")
		}
	} else if hasComma {
		// Vérifier si c'est un séparateur de milliers ou décimal
		parts := strings.Split(s, ",")
		if len(parts) == 2 && len(parts[1]) <= 2 {
			// Probablement décimal européen: 1234,56
			s = strings.ReplaceAll(s, ",", ".")
		} else {
			// Séparateur de milliers: 1,234,567
			s = strings.ReplaceAll(s, ",", "")
		}
	}

	// Parser le montant
	var amount float64
	fmt.Sscanf(s, "%f", &amount)
	return amount
}

// detectFinancialAnomalies détecte les anomalies financières avec extraction de montants réels
func (a *AnomalyService) detectFinancialAnomalies(caseData *models.Case) []models.Anomaly {
	anomalies := []models.Anomaly{}

	// Collecter tous les montants de toutes les preuves
	allAmounts := []float64{}
	evidenceAmounts := make(map[string][]float64) // evidenceID -> amounts

	for _, evidence := range caseData.Evidence {
		amounts := extractAmounts(evidence.Description)

		// Aussi chercher dans le nom de la preuve
		nameAmounts := extractAmounts(evidence.Name)
		amounts = append(amounts, nameAmounts...)

		if len(amounts) > 0 {
			evidenceAmounts[evidence.ID] = amounts
			allAmounts = append(allAmounts, amounts...)
		}
	}

	// Si pas assez de montants pour une analyse statistique, utiliser des seuils fixes
	if len(allAmounts) < 3 {
		// Revenir à l'ancienne méthode basée sur les mots-clés
		return a.detectFinancialAnomaliesKeywords(caseData)
	}

	// Calculer les statistiques pour les seuils adaptatifs
	median := calculateMedian(allAmounts)
	mad := calculateMAD(allAmounts)
	mean := calculateMean(allAmounts)
	stdDev := calculateStdDev(allAmounts, mean)

	// Définir des seuils absolus pour les montants "élevés"
	const absoluteHighThreshold = 10000.0   // 10k€
	const absoluteCriticalThreshold = 100000.0 // 100k€

	for _, evidence := range caseData.Evidence {
		amounts, hasAmounts := evidenceAmounts[evidence.ID]
		if !hasAmounts || len(amounts) == 0 {
			continue
		}

		for _, amount := range amounts {
			// Calculer le Z-score modifié
			modifiedZScore := calculateModifiedZScore(amount, median, mad)
			standardZScore := calculateZScore(amount, mean, stdDev)
			effectiveZScore := math.Max(math.Abs(modifiedZScore), math.Abs(standardZScore))

			// Détecter les anomalies basées sur le Z-score ou les seuils absolus
			isStatisticallyAnormal := effectiveZScore > 2.5
			isAbsolutelyHigh := amount >= absoluteHighThreshold

			if isStatisticallyAnormal || isAbsolutelyHigh {
				severity := a.calculateFinancialSeverity(amount, effectiveZScore)
				confidence := a.calculateFinancialConfidence(amount, effectiveZScore, isStatisticallyAnormal, isAbsolutelyHigh)

				title := "Montant financier anormal détecté"
				if amount >= absoluteCriticalThreshold {
					title = "Transaction financière majeure détectée"
				}

				anomalies = append(anomalies, models.Anomaly{
					ID:          uuid.New().String(),
					CaseID:      caseData.ID,
					Type:        models.AnomalyFinancial,
					Severity:    severity,
					Title:       title,
					Description: fmt.Sprintf("Montant de %.2f€ détecté dans '%s' (Z-score: %.2f, médiane: %.2f€)",
						amount, evidence.Name, effectiveZScore, median),
					DetectedAt:  time.Now(),
					EvidenceIDs: []string{evidence.ID},
					Confidence:  confidence,
					Details: map[string]interface{}{
						"amount":              amount,
						"evidence_name":       evidence.Name,
						"evidence_type":       string(evidence.Type),
						"z_score":             effectiveZScore,
						"modified_z_score":    modifiedZScore,
						"standard_z_score":    standardZScore,
						"median_amount":       median,
						"mean_amount":         mean,
						"std_dev":             stdDev,
						"mad":                 mad,
						"is_statistical":      isStatisticallyAnormal,
						"is_absolute_high":    isAbsolutelyHigh,
						"detection_method":    "amount_extraction_zscore",
					},
				})
			}
		}
	}

	// Détecter les patterns de transactions fractionnées (structuring)
	structuringAnomalies := a.detectStructuring(caseData, evidenceAmounts)
	anomalies = append(anomalies, structuringAnomalies...)

	return anomalies
}

// calculateFinancialSeverity calcule la sévérité basée sur le montant et le Z-score
func (a *AnomalyService) calculateFinancialSeverity(amount, zScore float64) models.AnomalySeverity {
	if amount >= 100000 || zScore > 4 {
		return models.SeverityCritical
	} else if amount >= 50000 || zScore > 3 {
		return models.SeverityHigh
	} else if amount >= 10000 || zScore > 2.5 {
		return models.SeverityMedium
	}
	return models.SeverityLow
}

// calculateFinancialConfidence calcule la confiance pour une anomalie financière
func (a *AnomalyService) calculateFinancialConfidence(amount, zScore float64, isStatistical, isAbsolute bool) int {
	baseConfidence := 50

	if isStatistical && isAbsolute {
		baseConfidence = 85
	} else if isStatistical {
		baseConfidence = 70
	} else if isAbsolute {
		baseConfidence = 65
	}

	// Bonus basé sur le Z-score
	zBonus := int(math.Min(zScore*5, 15))

	return int(math.Min(float64(baseConfidence+zBonus), 95))
}

// detectStructuring détecte les patterns de fractionnement de transactions
func (a *AnomalyService) detectStructuring(caseData *models.Case, evidenceAmounts map[string][]float64) []models.Anomaly {
	anomalies := []models.Anomaly{}

	// Collecter tous les montants avec leurs timestamps (si disponibles)
	type amountRecord struct {
		amount     float64
		evidenceID string
	}

	allRecords := []amountRecord{}
	for evidID, amounts := range evidenceAmounts {
		for _, amt := range amounts {
			allRecords = append(allRecords, amountRecord{amount: amt, evidenceID: evidID})
		}
	}

	// Détecter les montants suspects (juste en dessous de seuils réglementaires)
	const threshold1 = 10000.0 // Seuil de déclaration bancaire
	const threshold2 = 15000.0 // Autre seuil courant
	const marginPercent = 0.15 // 15% en dessous du seuil

	suspiciousAmounts := []amountRecord{}
	for _, record := range allRecords {
		// Vérifier si le montant est juste en dessous d'un seuil
		if record.amount > threshold1*(1-marginPercent) && record.amount < threshold1 {
			suspiciousAmounts = append(suspiciousAmounts, record)
		} else if record.amount > threshold2*(1-marginPercent) && record.amount < threshold2 {
			suspiciousAmounts = append(suspiciousAmounts, record)
		}
	}

	// Si plusieurs transactions sont juste en dessous du seuil, c'est suspect
	if len(suspiciousAmounts) >= 2 {
		evidenceIDs := []string{}
		totalAmount := 0.0
		for _, record := range suspiciousAmounts {
			evidenceIDs = append(evidenceIDs, record.evidenceID)
			totalAmount += record.amount
		}

		anomalies = append(anomalies, models.Anomaly{
			ID:          uuid.New().String(),
			CaseID:      caseData.ID,
			Type:        models.AnomalyFinancial,
			Severity:    models.SeverityHigh,
			Title:       "Pattern de fractionnement suspect (structuring)",
			Description: fmt.Sprintf("%d transactions juste en dessous des seuils de déclaration (total: %.2f€)",
				len(suspiciousAmounts), totalAmount),
			DetectedAt:  time.Now(),
			EvidenceIDs: evidenceIDs,
			Confidence:  75 + len(suspiciousAmounts)*5,
			Details: map[string]interface{}{
				"suspicious_count":    len(suspiciousAmounts),
				"total_amount":        totalAmount,
				"regulatory_threshold": threshold1,
				"pattern_type":        "structuring",
				"detection_method":    "threshold_avoidance",
			},
		})
	}

	return anomalies
}

// detectFinancialAnomaliesKeywords est la méthode de fallback basée sur les mots-clés
func (a *AnomalyService) detectFinancialAnomaliesKeywords(caseData *models.Case) []models.Anomaly {
	anomalies := []models.Anomaly{}

	for _, evidence := range caseData.Evidence {
		desc := strings.ToLower(evidence.Description)

		if strings.Contains(desc, "€") || strings.Contains(desc, "euro") ||
			strings.Contains(desc, "transaction") || strings.Contains(desc, "virement") {

			if strings.Contains(desc, "million") || strings.Contains(desc, "important") ||
				strings.Contains(desc, "suspect") || strings.Contains(desc, "inhabituel") {
				anomalies = append(anomalies, models.Anomaly{
					ID:          uuid.New().String(),
					CaseID:      caseData.ID,
					Type:        models.AnomalyFinancial,
					Severity:    models.SeverityMedium,
					Title:       "Transaction financière à examiner",
					Description: fmt.Sprintf("Preuve '%s' contient des références financières potentiellement suspectes", evidence.Name),
					DetectedAt:  time.Now(),
					EvidenceIDs: []string{evidence.ID},
					Confidence:  55,
					Details: map[string]interface{}{
						"evidence_name":    evidence.Name,
						"evidence_type":    string(evidence.Type),
						"detection_method": "keyword_fallback",
					},
				})
			}
		}
	}

	return anomalies
}

// NetworkGraph représente un graphe de communication pour l'analyse de centralité
type NetworkGraph struct {
	nodes     map[string]bool              // entityID -> exists
	edges     map[string]map[string]int    // from -> to -> weight (nombre de communications)
	degree    map[string]int               // entityID -> degré total
	inDegree  map[string]int               // entityID -> degré entrant
	outDegree map[string]int               // entityID -> degré sortant
}

// newNetworkGraph crée un nouveau graphe de réseau
func newNetworkGraph() *NetworkGraph {
	return &NetworkGraph{
		nodes:     make(map[string]bool),
		edges:     make(map[string]map[string]int),
		degree:    make(map[string]int),
		inDegree:  make(map[string]int),
		outDegree: make(map[string]int),
	}
}

// addEdge ajoute une arête au graphe
func (g *NetworkGraph) addEdge(from, to string, weight int) {
	g.nodes[from] = true
	g.nodes[to] = true

	if g.edges[from] == nil {
		g.edges[from] = make(map[string]int)
	}
	g.edges[from][to] += weight

	g.outDegree[from] += weight
	g.inDegree[to] += weight
	g.degree[from] += weight
	g.degree[to] += weight
}

// calculateBetweennessCentrality calcule la centralité d'intermédiarité simplifiée
func (g *NetworkGraph) calculateBetweennessCentrality() map[string]float64 {
	betweenness := make(map[string]float64)

	// Initialiser tous les nœuds à 0
	for node := range g.nodes {
		betweenness[node] = 0
	}

	// Pour chaque paire de nœuds, compter combien de chemins passent par chaque nœud
	nodeList := make([]string, 0, len(g.nodes))
	for node := range g.nodes {
		nodeList = append(nodeList, node)
	}

	// Algorithme simplifié: pour chaque source, faire un BFS et compter les chemins
	for _, source := range nodeList {
		// BFS depuis la source
		distances := make(map[string]int)
		paths := make(map[string]int)
		predecessors := make(map[string][]string)

		distances[source] = 0
		paths[source] = 1
		queue := []string{source}

		for len(queue) > 0 {
			current := queue[0]
			queue = queue[1:]

			if neighbors, ok := g.edges[current]; ok {
				for neighbor := range neighbors {
					if _, visited := distances[neighbor]; !visited {
						distances[neighbor] = distances[current] + 1
						queue = append(queue, neighbor)
					}
					if distances[neighbor] == distances[current]+1 {
						paths[neighbor] += paths[current]
						predecessors[neighbor] = append(predecessors[neighbor], current)
					}
				}
			}
		}

		// Calculer la contribution à la centralité (backtrack)
		delta := make(map[string]float64)
		// Trier les nœuds par distance décroissante
		nodesByDist := make([][]string, 0)
		for node, dist := range distances {
			for len(nodesByDist) <= dist {
				nodesByDist = append(nodesByDist, []string{})
			}
			nodesByDist[dist] = append(nodesByDist[dist], node)
		}

		for i := len(nodesByDist) - 1; i >= 0; i-- {
			for _, w := range nodesByDist[i] {
				for _, v := range predecessors[w] {
					if paths[w] > 0 {
						delta[v] += float64(paths[v]) / float64(paths[w]) * (1 + delta[w])
					}
				}
				if w != source {
					betweenness[w] += delta[w]
				}
			}
		}
	}

	// Normaliser
	n := float64(len(g.nodes))
	if n > 2 {
		normFactor := 2.0 / ((n - 1) * (n - 2))
		for node := range betweenness {
			betweenness[node] *= normFactor
		}
	}

	return betweenness
}

// calculatePageRank calcule le PageRank simplifié
func (g *NetworkGraph) calculatePageRank(iterations int, dampingFactor float64) map[string]float64 {
	n := float64(len(g.nodes))
	if n == 0 {
		return make(map[string]float64)
	}

	// Initialiser tous les scores à 1/n
	pageRank := make(map[string]float64)
	for node := range g.nodes {
		pageRank[node] = 1.0 / n
	}

	// Itérer
	for iter := 0; iter < iterations; iter++ {
		newRank := make(map[string]float64)

		// Contribution de base (téléportation)
		for node := range g.nodes {
			newRank[node] = (1 - dampingFactor) / n
		}

		// Contribution des liens
		for from := range g.nodes {
			outLinks := g.edges[from]
			if len(outLinks) > 0 {
				totalWeight := 0
				for _, w := range outLinks {
					totalWeight += w
				}
				for to, weight := range outLinks {
					contribution := dampingFactor * pageRank[from] * float64(weight) / float64(totalWeight)
					newRank[to] += contribution
				}
			} else {
				// Nœud sans lien sortant: distribuer uniformément
				for node := range g.nodes {
					newRank[node] += dampingFactor * pageRank[from] / n
				}
			}
		}

		pageRank = newRank
	}

	return pageRank
}

// detectCommunicationAnomalies détecte les anomalies de communication avec analyse de centralité
func (a *AnomalyService) detectCommunicationAnomalies(caseData *models.Case) []models.Anomaly {
	anomalies := []models.Anomaly{}

	// Construire le graphe de communication
	graph := newNetworkGraph()
	communicationCount := make(map[string]int)

	for _, entity := range caseData.Entities {
		graph.nodes[entity.ID] = true
		for _, rel := range entity.Relations {
			relType := strings.ToLower(rel.Type)
			if strings.Contains(relType, "appel") || strings.Contains(relType, "message") ||
				strings.Contains(relType, "contact") || strings.Contains(relType, "communiqu") ||
				strings.Contains(relType, "email") || strings.Contains(relType, "sms") {
				graph.addEdge(entity.ID, rel.ToID, 1)
				communicationCount[entity.ID]++
				communicationCount[rel.ToID]++
			}
		}
	}

	// Si pas assez de nœuds pour une analyse de réseau
	if len(graph.nodes) < 3 {
		return a.detectCommunicationAnomaliesSimple(caseData, communicationCount)
	}

	// Calculer les métriques de centralité
	betweenness := graph.calculateBetweennessCentrality()
	pageRank := graph.calculatePageRank(20, 0.85)

	// Collecter les valeurs pour calculer les statistiques
	betweennessValues := make([]float64, 0, len(betweenness))
	pageRankValues := make([]float64, 0, len(pageRank))
	degreeValues := make([]float64, 0, len(graph.degree))

	for _, v := range betweenness {
		betweennessValues = append(betweennessValues, v)
	}
	for _, v := range pageRank {
		pageRankValues = append(pageRankValues, v)
	}
	for _, v := range graph.degree {
		degreeValues = append(degreeValues, float64(v))
	}

	// Calculer les statistiques
	betweennessMean := calculateMean(betweennessValues)
	betweennessStd := calculateStdDev(betweennessValues, betweennessMean)
	pageRankMean := calculateMean(pageRankValues)
	pageRankStd := calculateStdDev(pageRankValues, pageRankMean)
	degreeMean := calculateMean(degreeValues)
	degreeStd := calculateStdDev(degreeValues, degreeMean)

	// Détecter les anomalies basées sur la centralité
	for _, entity := range caseData.Entities {
		entityID := entity.ID

		// Calculer les Z-scores
		betweennessZ := calculateZScore(betweenness[entityID], betweennessMean, betweennessStd)
		pageRankZ := calculateZScore(pageRank[entityID], pageRankMean, pageRankStd)
		degreeZ := calculateZScore(float64(graph.degree[entityID]), degreeMean, degreeStd)

		// Score composite de centralité
		centralityScore := (math.Abs(betweennessZ) + math.Abs(pageRankZ) + math.Abs(degreeZ)) / 3

		// Détecter si l'entité a une centralité anormalement élevée
		if centralityScore > 2.0 || betweennessZ > 2.5 || pageRankZ > 2.5 {
			severity := a.calculateCentralitySeverity(centralityScore, betweennessZ, pageRankZ)
			confidence := int(math.Min(50+centralityScore*15, 95))

			// Déterminer le type d'anomalie
			anomalyTitle := "Position centrale inhabituelle dans le réseau"
			if betweennessZ > pageRankZ && betweennessZ > degreeZ {
				anomalyTitle = "Rôle d'intermédiaire clé (broker) dans les communications"
			} else if pageRankZ > betweennessZ && pageRankZ > degreeZ {
				anomalyTitle = "Hub de communication influent"
			}

			anomalies = append(anomalies, models.Anomaly{
				ID:          uuid.New().String(),
				CaseID:      caseData.ID,
				Type:        models.AnomalyCommunication,
				Severity:    severity,
				Title:       anomalyTitle,
				Description: fmt.Sprintf("'%s' occupe une position centrale anormale (centralité: %.2f, betweenness Z: %.2f, PageRank Z: %.2f)",
					entity.Name, centralityScore, betweennessZ, pageRankZ),
				DetectedAt: time.Now(),
				EntityIDs:  []string{entityID},
				Confidence: confidence,
				Details: map[string]interface{}{
					"betweenness":           betweenness[entityID],
					"betweenness_z_score":   betweennessZ,
					"pagerank":              pageRank[entityID],
					"pagerank_z_score":      pageRankZ,
					"degree":                graph.degree[entityID],
					"degree_z_score":        degreeZ,
					"in_degree":             graph.inDegree[entityID],
					"out_degree":            graph.outDegree[entityID],
					"centrality_score":      centralityScore,
					"network_size":          len(graph.nodes),
					"detection_method":      "network_centrality",
				},
			})
		}

		// Détecter les entités avec un déséquilibre in/out (communication asymétrique)
		if graph.inDegree[entityID] > 0 || graph.outDegree[entityID] > 0 {
			inOut := float64(graph.inDegree[entityID])
			outIn := float64(graph.outDegree[entityID])
			total := inOut + outIn

			if total > 0 {
				ratio := math.Abs(inOut-outIn) / total
				if ratio > 0.7 && total >= 3 { // Très asymétrique avec au moins 3 communications
					direction := "entrant"
					if outIn > inOut {
						direction = "sortant"
					}

					anomalies = append(anomalies, models.Anomaly{
						ID:          uuid.New().String(),
						CaseID:      caseData.ID,
						Type:        models.AnomalyCommunication,
						Severity:    models.SeverityLow,
						Title:       "Pattern de communication asymétrique",
						Description: fmt.Sprintf("'%s' a un flux de communication majoritairement %s (ratio: %.0f%%)",
							entity.Name, direction, ratio*100),
						DetectedAt: time.Now(),
						EntityIDs:  []string{entityID},
						Confidence: 55,
						Details: map[string]interface{}{
							"in_degree":        graph.inDegree[entityID],
							"out_degree":       graph.outDegree[entityID],
							"asymmetry_ratio":  ratio,
							"dominant_flow":    direction,
							"detection_method": "flow_asymmetry",
						},
					})
				}
			}
		}
	}

	// Détecter les cliques ou sous-groupes isolés
	cliqueAnomalies := a.detectCommunicationCliques(caseData, graph)
	anomalies = append(anomalies, cliqueAnomalies...)

	return anomalies
}

// calculateCentralitySeverity calcule la sévérité basée sur les scores de centralité
func (a *AnomalyService) calculateCentralitySeverity(composite, betweenness, pagerank float64) models.AnomalySeverity {
	maxZ := math.Max(composite, math.Max(betweenness, pagerank))
	if maxZ > 4 {
		return models.SeverityCritical
	} else if maxZ > 3 {
		return models.SeverityHigh
	} else if maxZ > 2 {
		return models.SeverityMedium
	}
	return models.SeverityLow
}

// detectCommunicationCliques détecte les sous-groupes de communication isolés
func (a *AnomalyService) detectCommunicationCliques(caseData *models.Case, graph *NetworkGraph) []models.Anomaly {
	anomalies := []models.Anomaly{}

	// Trouver les composantes connexes
	visited := make(map[string]bool)
	components := [][]string{}

	var dfs func(node string, component *[]string)
	dfs = func(node string, component *[]string) {
		if visited[node] {
			return
		}
		visited[node] = true
		*component = append(*component, node)

		// Voisins sortants
		if neighbors, ok := graph.edges[node]; ok {
			for neighbor := range neighbors {
				dfs(neighbor, component)
			}
		}
		// Voisins entrants (graphe non-dirigé pour cette analyse)
		for from, edges := range graph.edges {
			if _, ok := edges[node]; ok {
				dfs(from, component)
			}
		}
	}

	for node := range graph.nodes {
		if !visited[node] {
			component := []string{}
			dfs(node, &component)
			if len(component) > 1 {
				components = append(components, component)
			}
		}
	}

	// Si plusieurs composantes, signaler les sous-groupes isolés
	if len(components) > 1 {
		for _, component := range components {
			if len(component) >= 2 && len(component) <= len(graph.nodes)/2 {
				// Trouver les noms des entités
				names := []string{}
				for _, id := range component {
					for _, e := range caseData.Entities {
						if e.ID == id {
							names = append(names, e.Name)
							break
						}
					}
				}

				anomalies = append(anomalies, models.Anomaly{
					ID:          uuid.New().String(),
					CaseID:      caseData.ID,
					Type:        models.AnomalyCommunication,
					Severity:    models.SeverityMedium,
					Title:       "Sous-groupe de communication isolé",
					Description: fmt.Sprintf("Groupe de %d entités communiquant entre elles mais isolées du reste: %s",
						len(component), strings.Join(names, ", ")),
					DetectedAt: time.Now(),
					EntityIDs:  component,
					Confidence: 65,
					Details: map[string]interface{}{
						"component_size":    len(component),
						"total_components":  len(components),
						"entity_names":      names,
						"detection_method":  "connected_components",
					},
				})
			}
		}
	}

	return anomalies
}

// detectCommunicationAnomaliesSimple est la méthode de fallback pour les petits réseaux
func (a *AnomalyService) detectCommunicationAnomaliesSimple(caseData *models.Case, communicationCount map[string]int) []models.Anomaly {
	anomalies := []models.Anomaly{}

	// Collecter les valeurs de communication
	commValues := make([]float64, 0, len(communicationCount))
	for _, count := range communicationCount {
		commValues = append(commValues, float64(count))
	}

	if len(commValues) == 0 {
		return anomalies
	}

	mean := calculateMean(commValues)
	stdDev := calculateStdDev(commValues, mean)

	for entityID, count := range communicationCount {
		zScore := calculateZScore(float64(count), mean, stdDev)

		if zScore > 2.0 {
			var entityName string
			for _, e := range caseData.Entities {
				if e.ID == entityID {
					entityName = e.Name
					break
				}
			}

			severity := models.SeverityMedium
			if zScore > 3 {
				severity = models.SeverityHigh
			}

			anomalies = append(anomalies, models.Anomaly{
				ID:          uuid.New().String(),
				CaseID:      caseData.ID,
				Type:        models.AnomalyCommunication,
				Severity:    severity,
				Title:       "Activité de communication élevée",
				Description: fmt.Sprintf("'%s' a un nombre de communications anormalement élevé (%d, Z-score: %.2f)",
					entityName, count, zScore),
				DetectedAt: time.Now(),
				EntityIDs:  []string{entityID},
				Confidence: int(math.Min(50+zScore*15, 90)),
				Details: map[string]interface{}{
					"communication_count": count,
					"mean_count":          mean,
					"std_dev":             stdDev,
					"z_score":             zScore,
					"detection_method":    "zscore_simple",
				},
			})
		}
	}

	return anomalies
}

// detectBehaviorAnomalies détecte les anomalies comportementales
func (a *AnomalyService) detectBehaviorAnomalies(caseData *models.Case) []models.Anomaly {
	anomalies := []models.Anomaly{}

	// Chercher des changements de comportement dans la description des entités
	behaviorKeywords := []string{"inhabituel", "soudain", "changement", "bizarre", "suspect",
		"nerveux", "agité", "différent", "anormal", "étrange"}

	for _, entity := range caseData.Entities {
		desc := strings.ToLower(entity.Description)
		for _, keyword := range behaviorKeywords {
			if strings.Contains(desc, keyword) {
				anomalies = append(anomalies, models.Anomaly{
					ID:          uuid.New().String(),
					CaseID:      caseData.ID,
					Type:        models.AnomalyBehavior,
					Severity:    models.SeverityLow,
					Title:       "Comportement inhabituel signalé",
					Description: fmt.Sprintf("La description de '%s' mentionne un comportement '%s'", entity.Name, keyword),
					DetectedAt:  time.Now(),
					EntityIDs:   []string{entity.ID},
					Confidence:  45,
					Details: map[string]interface{}{
						"keyword_found": keyword,
						"entity_name":   entity.Name,
						"entity_role":   string(entity.Role),
					},
				})
				break // Un seul match par entité
			}
		}
	}

	return anomalies
}

// detectLocationAnomalies détecte les anomalies de localisation
func (a *AnomalyService) detectLocationAnomalies(caseData *models.Case) []models.Anomaly {
	anomalies := []models.Anomaly{}

	// Compter les occurrences par lieu
	locationCount := make(map[string]int)
	for _, event := range caseData.Timeline {
		if event.Location != "" {
			locationCount[event.Location]++
		}
	}

	// Identifier les lieux qui apparaissent très fréquemment
	var totalLocs int
	for _, count := range locationCount {
		totalLocs += count
	}

	avgLocs := float64(totalLocs) / math.Max(float64(len(locationCount)), 1)

	for location, count := range locationCount {
		if float64(count) > avgLocs*3 { // Plus de 3x la moyenne
			anomalies = append(anomalies, models.Anomaly{
				ID:          uuid.New().String(),
				CaseID:      caseData.ID,
				Type:        models.AnomalyLocation,
				Severity:    models.SeverityLow,
				Title:       "Point chaud géographique",
				Description: fmt.Sprintf("Le lieu '%s' apparaît de manière récurrente (%d occurrences)", location, count),
				DetectedAt:  time.Now(),
				Confidence:  50,
				Details: map[string]interface{}{
					"location":         location,
					"occurrence_count": count,
					"average_count":    avgLocs,
				},
			})
		}
	}

	return anomalies
}

// detectRelationAnomalies détecte les anomalies relationnelles
func (a *AnomalyService) detectRelationAnomalies(caseData *models.Case) []models.Anomaly {
	anomalies := []models.Anomaly{}

	// Chercher des relations non vérifiées entre suspects et victimes
	suspectIDs := make(map[string]string) // ID -> name
	victimIDs := make(map[string]string)

	for _, entity := range caseData.Entities {
		if entity.Role == models.RoleSuspect {
			suspectIDs[entity.ID] = entity.Name
		} else if entity.Role == models.RoleVictim {
			victimIDs[entity.ID] = entity.Name
		}
	}

	// Chercher les relations directes suspect-victime
	for _, entity := range caseData.Entities {
		if _, isSuspect := suspectIDs[entity.ID]; isSuspect {
			for _, rel := range entity.Relations {
				if victimName, isVictim := victimIDs[rel.ToID]; isVictim {
					if !rel.Verified {
						anomalies = append(anomalies, models.Anomaly{
							ID:          uuid.New().String(),
							CaseID:      caseData.ID,
							Type:        models.AnomalyRelation,
							Severity:    models.SeverityHigh,
							Title:       "Relation suspect-victime non vérifiée",
							Description: fmt.Sprintf("Relation '%s' entre suspect '%s' et victime '%s' non vérifiée",
								rel.Type, entity.Name, victimName),
							DetectedAt: time.Now(),
							EntityIDs:  []string{entity.ID, rel.ToID},
							Confidence: 75,
							Details: map[string]interface{}{
								"relation_type": rel.Type,
								"suspect_name":  entity.Name,
								"victim_name":   victimName,
								"verified":      rel.Verified,
							},
						})
					}
				}
			}
		}
	}

	return anomalies
}

// detectPatternAnomalies détecte les patterns suspects avec détection de ruptures
func (a *AnomalyService) detectPatternAnomalies(caseData *models.Case) []models.Anomaly {
	anomalies := []models.Anomaly{}

	if len(caseData.Timeline) < 3 {
		return anomalies
	}

	// Trier les événements par date
	events := make([]models.Event, len(caseData.Timeline))
	copy(events, caseData.Timeline)
	sort.Slice(events, func(i, j int) bool {
		return events[i].Timestamp.Before(events[j].Timestamp)
	})

	// 1. Détecter les patterns temporels (concentrations horaires et journalières)
	hourCount := make(map[int]int)
	dayCount := make(map[int]int) // 0=Sunday, 6=Saturday
	weekCount := make(map[int]int) // Numéro de semaine

	for _, event := range events {
		hourCount[event.Timestamp.Hour()]++
		dayCount[int(event.Timestamp.Weekday())]++
		_, week := event.Timestamp.ISOWeek()
		weekCount[week]++
	}

	// Analyser les concentrations horaires avec Z-score
	hourValues := make([]float64, 0, len(hourCount))
	for _, count := range hourCount {
		hourValues = append(hourValues, float64(count))
	}

	if len(hourValues) > 0 {
		hourMean := calculateMean(hourValues)
		hourStd := calculateStdDev(hourValues, hourMean)

		for hour, count := range hourCount {
			zScore := calculateZScore(float64(count), hourMean, hourStd)
			if zScore > 2.5 {
				severity := models.SeverityLow
				if zScore > 3.5 {
					severity = models.SeverityMedium
				}

				anomalies = append(anomalies, models.Anomaly{
					ID:          uuid.New().String(),
					CaseID:      caseData.ID,
					Type:        models.AnomalyPattern,
					Severity:    severity,
					Title:       "Pattern horaire significatif",
					Description: fmt.Sprintf("Concentration d'événements autour de %dh (%d occurrences, Z-score: %.2f)", hour, count, zScore),
					DetectedAt:  time.Now(),
					Confidence:  int(math.Min(45+zScore*10, 85)),
					Details: map[string]interface{}{
						"hour":            hour,
						"count":           count,
						"z_score":         zScore,
						"mean":            hourMean,
						"std_dev":         hourStd,
						"pattern_type":    "hourly_concentration",
						"detection_method": "zscore",
					},
				})
			}
		}
	}

	// 2. Détecter les ruptures de pattern (changements brusques dans l'activité)
	ruptureAnomalies := a.detectPatternRuptures(caseData, events)
	anomalies = append(anomalies, ruptureAnomalies...)

	// 3. Détecter les patterns périodiques
	periodicAnomalies := a.detectPeriodicPatterns(caseData, events)
	anomalies = append(anomalies, periodicAnomalies...)

	// 4. Détecter les patterns comportementaux par entité
	entityPatternAnomalies := a.detectEntityPatternChanges(caseData, events)
	anomalies = append(anomalies, entityPatternAnomalies...)

	return anomalies
}

// detectPatternRuptures détecte les changements brusques dans les patterns d'activité
func (a *AnomalyService) detectPatternRuptures(caseData *models.Case, events []models.Event) []models.Anomaly {
	anomalies := []models.Anomaly{}

	if len(events) < 5 {
		return anomalies
	}

	// Calculer l'activité par période (fenêtre glissante)
	// Utiliser des fenêtres de 7 jours pour détecter les ruptures

	// Grouper les événements par jour
	eventsByDay := make(map[string][]models.Event)
	for _, event := range events {
		dayKey := event.Timestamp.Format("2006-01-02")
		eventsByDay[dayKey] = append(eventsByDay[dayKey], event)
	}

	// Créer une série temporelle d'activité quotidienne
	if len(eventsByDay) < 7 {
		return anomalies
	}

	// Obtenir les dates triées
	dates := make([]string, 0, len(eventsByDay))
	for date := range eventsByDay {
		dates = append(dates, date)
	}
	sort.Strings(dates)

	// Calculer l'activité par jour
	dailyActivity := make([]float64, len(dates))
	for i, date := range dates {
		dailyActivity[i] = float64(len(eventsByDay[date]))
	}

	// Détecter les ruptures avec CUSUM simplifié
	ruptures := a.detectCUSUM(dailyActivity, dates)
	for _, rupture := range ruptures {
		anomalies = append(anomalies, models.Anomaly{
			ID:          uuid.New().String(),
			CaseID:      caseData.ID,
			Type:        models.AnomalyPattern,
			Severity:    rupture.severity,
			Title:       "Rupture de pattern détectée",
			Description: rupture.description,
			DetectedAt:  time.Now(),
			Confidence:  rupture.confidence,
			Details:     rupture.details,
		})
	}

	// Détecter les changements de variance (stabilité -> instabilité)
	varianceAnomalies := a.detectVarianceChanges(dailyActivity, dates, caseData.ID)
	anomalies = append(anomalies, varianceAnomalies...)

	return anomalies
}

// RuptureInfo contient les informations sur une rupture détectée
type RuptureInfo struct {
	date        string
	severity    models.AnomalySeverity
	confidence  int
	description string
	details     map[string]interface{}
}

// detectCUSUM détecte les ruptures avec l'algorithme CUSUM simplifié
func (a *AnomalyService) detectCUSUM(values []float64, dates []string) []RuptureInfo {
	ruptures := []RuptureInfo{}

	if len(values) < 10 {
		return ruptures
	}

	// Calculer la moyenne et l'écart-type de référence (première moitié)
	halfLen := len(values) / 2
	refValues := values[:halfLen]
	refMean := calculateMean(refValues)
	refStd := calculateStdDev(refValues, refMean)

	if refStd == 0 {
		refStd = 1 // Éviter la division par zéro
	}

	// Seuil de détection (typiquement 4-5 pour CUSUM)
	threshold := 4.0 * refStd

	// Calculer les sommes cumulées
	var cusumPos, cusumNeg float64
	slackValue := 0.5 * refStd // Slack pour réduire les faux positifs

	for i := halfLen; i < len(values); i++ {
		deviation := values[i] - refMean

		// CUSUM positif (détecte les augmentations)
		cusumPos = math.Max(0, cusumPos+deviation-slackValue)

		// CUSUM négatif (détecte les diminutions)
		cusumNeg = math.Max(0, cusumNeg-deviation-slackValue)

		// Vérifier si un seuil est dépassé
		if cusumPos > threshold {
			severity := models.SeverityMedium
			if cusumPos > threshold*1.5 {
				severity = models.SeverityHigh
			}

			ruptures = append(ruptures, RuptureInfo{
				date:     dates[i],
				severity: severity,
				confidence: int(math.Min(60+cusumPos/refStd*5, 90)),
				description: fmt.Sprintf("Augmentation significative de l'activité détectée le %s (CUSUM: %.2f, seuil: %.2f)",
					dates[i], cusumPos, threshold),
				details: map[string]interface{}{
					"rupture_date":     dates[i],
					"cusum_value":      cusumPos,
					"threshold":        threshold,
					"reference_mean":   refMean,
					"reference_std":    refStd,
					"direction":        "increase",
					"detection_method": "cusum",
				},
			})

			// Réinitialiser après détection
			cusumPos = 0
		}

		if cusumNeg > threshold {
			severity := models.SeverityMedium
			if cusumNeg > threshold*1.5 {
				severity = models.SeverityHigh
			}

			ruptures = append(ruptures, RuptureInfo{
				date:     dates[i],
				severity: severity,
				confidence: int(math.Min(60+cusumNeg/refStd*5, 90)),
				description: fmt.Sprintf("Diminution significative de l'activité détectée le %s (CUSUM: %.2f, seuil: %.2f)",
					dates[i], cusumNeg, threshold),
				details: map[string]interface{}{
					"rupture_date":     dates[i],
					"cusum_value":      cusumNeg,
					"threshold":        threshold,
					"reference_mean":   refMean,
					"reference_std":    refStd,
					"direction":        "decrease",
					"detection_method": "cusum",
				},
			})

			// Réinitialiser après détection
			cusumNeg = 0
		}
	}

	return ruptures
}

// detectVarianceChanges détecte les changements de variance (stabilité du pattern)
func (a *AnomalyService) detectVarianceChanges(values []float64, dates []string, caseID string) []models.Anomaly {
	anomalies := []models.Anomaly{}

	if len(values) < 14 {
		return anomalies
	}

	// Diviser en deux périodes
	halfLen := len(values) / 2
	period1 := values[:halfLen]
	period2 := values[halfLen:]

	mean1 := calculateMean(period1)
	mean2 := calculateMean(period2)
	var1 := calculateStdDev(period1, mean1)
	var2 := calculateStdDev(period2, mean2)

	if var1 == 0 {
		var1 = 0.1
	}

	// Ratio de variance (F-test simplifié)
	varianceRatio := var2 / var1
	if varianceRatio < 1 {
		varianceRatio = var1 / var2
	}

	// Un ratio > 2 est significatif
	if varianceRatio > 2.0 {
		direction := "plus instable"
		if var2 < var1 {
			direction = "plus stable"
		}

		severity := models.SeverityLow
		if varianceRatio > 3.0 {
			severity = models.SeverityMedium
		}

		anomalies = append(anomalies, models.Anomaly{
			ID:          uuid.New().String(),
			CaseID:      caseID,
			Type:        models.AnomalyPattern,
			Severity:    severity,
			Title:       "Changement de stabilité du pattern",
			Description: fmt.Sprintf("L'activité est devenue %s (ratio de variance: %.2f)", direction, varianceRatio),
			DetectedAt:  time.Now(),
			Confidence:  int(math.Min(50+varianceRatio*10, 85)),
			Details: map[string]interface{}{
				"variance_period1":   var1 * var1,
				"variance_period2":   var2 * var2,
				"std_dev_period1":    var1,
				"std_dev_period2":    var2,
				"variance_ratio":     varianceRatio,
				"direction":          direction,
				"period1_start":      dates[0],
				"period1_end":        dates[halfLen-1],
				"period2_start":      dates[halfLen],
				"period2_end":        dates[len(dates)-1],
				"detection_method":   "variance_change",
			},
		})
	}

	return anomalies
}

// detectPeriodicPatterns détecte les patterns périodiques (cycles)
func (a *AnomalyService) detectPeriodicPatterns(caseData *models.Case, events []models.Event) []models.Anomaly {
	anomalies := []models.Anomaly{}

	if len(events) < 7 {
		return anomalies
	}

	// Analyser les patterns par jour de la semaine
	dayOfWeekCount := make([]int, 7)
	for _, event := range events {
		dayOfWeekCount[int(event.Timestamp.Weekday())]++
	}

	// Calculer les statistiques
	dowValues := make([]float64, 7)
	for i, count := range dayOfWeekCount {
		dowValues[i] = float64(count)
	}

	mean := calculateMean(dowValues)
	stdDev := calculateStdDev(dowValues, mean)

	// Détecter les jours avec une activité anormale
	dayNames := []string{"Dimanche", "Lundi", "Mardi", "Mercredi", "Jeudi", "Vendredi", "Samedi"}

	for day, count := range dayOfWeekCount {
		zScore := calculateZScore(float64(count), mean, stdDev)
		if math.Abs(zScore) > 2.0 && count > 2 {
			direction := "élevée"
			if zScore < 0 {
				direction = "faible"
			}

			anomalies = append(anomalies, models.Anomaly{
				ID:          uuid.New().String(),
				CaseID:      caseData.ID,
				Type:        models.AnomalyPattern,
				Severity:    models.SeverityLow,
				Title:       fmt.Sprintf("Pattern hebdomadaire: %s", dayNames[day]),
				Description: fmt.Sprintf("Activité %s le %s (%d événements, Z-score: %.2f)",
					direction, dayNames[day], count, zScore),
				DetectedAt: time.Now(),
				Confidence: int(math.Min(45+math.Abs(zScore)*10, 80)),
				Details: map[string]interface{}{
					"day_of_week":      day,
					"day_name":         dayNames[day],
					"event_count":      count,
					"z_score":          zScore,
					"mean":             mean,
					"std_dev":          stdDev,
					"pattern_type":     "weekly_cycle",
					"detection_method": "periodic_analysis",
				},
			})
		}
	}

	return anomalies
}

// detectEntityPatternChanges détecte les changements de pattern par entité
func (a *AnomalyService) detectEntityPatternChanges(caseData *models.Case, events []models.Event) []models.Anomaly {
	anomalies := []models.Anomaly{}

	if len(events) < 10 {
		return anomalies
	}

	// Grouper les événements par entité
	entityEvents := make(map[string][]models.Event)
	for _, event := range events {
		for _, entityID := range event.Entities {
			entityEvents[entityID] = append(entityEvents[entityID], event)
		}
	}

	// Pour chaque entité avec suffisamment d'événements
	for entityID, evts := range entityEvents {
		if len(evts) < 5 {
			continue
		}

		// Trier par date
		sort.Slice(evts, func(i, j int) bool {
			return evts[i].Timestamp.Before(evts[j].Timestamp)
		})

		// Calculer les intervalles entre événements
		intervals := make([]float64, len(evts)-1)
		for i := 1; i < len(evts); i++ {
			intervals[i-1] = evts[i].Timestamp.Sub(evts[i-1].Timestamp).Hours()
		}

		if len(intervals) < 4 {
			continue
		}

		// Détecter les changements de rythme
		halfLen := len(intervals) / 2
		firstHalf := intervals[:halfLen]
		secondHalf := intervals[halfLen:]

		mean1 := calculateMean(firstHalf)
		mean2 := calculateMean(secondHalf)

		if mean1 == 0 {
			mean1 = 1
		}

		changeRatio := mean2 / mean1

		// Changement significatif si ratio > 2 ou < 0.5
		if changeRatio > 2.0 || changeRatio < 0.5 {
			var entityName string
			for _, e := range caseData.Entities {
				if e.ID == entityID {
					entityName = e.Name
					break
				}
			}

			direction := "ralenti"
			if changeRatio < 1 {
				direction = "accéléré"
			}

			anomalies = append(anomalies, models.Anomaly{
				ID:          uuid.New().String(),
				CaseID:      caseData.ID,
				Type:        models.AnomalyPattern,
				Severity:    models.SeverityMedium,
				Title:       "Changement de rythme d'activité",
				Description: fmt.Sprintf("Le rythme d'activité de '%s' a %s (ratio: %.2f)",
					entityName, direction, changeRatio),
				DetectedAt: time.Now(),
				EntityIDs:  []string{entityID},
				Confidence: int(math.Min(55+math.Abs(changeRatio-1)*15, 85)),
				Details: map[string]interface{}{
					"entity_name":        entityName,
					"mean_interval_before": mean1,
					"mean_interval_after":  mean2,
					"change_ratio":         changeRatio,
					"direction":            direction,
					"event_count":          len(evts),
					"detection_method":     "entity_rhythm_change",
				},
			})
		}
	}

	return anomalies
}

// =========================================
// Corrélation croisée entre anomalies
// =========================================

// detectCrossCorrelations détecte les corrélations entre différents types d'anomalies
func (a *AnomalyService) detectCrossCorrelations(caseData *models.Case, existingAnomalies []models.Anomaly) []models.Anomaly {
	anomalies := []models.Anomaly{}

	if len(existingAnomalies) < 2 {
		return anomalies
	}

	// Grouper les anomalies par type
	byType := make(map[models.AnomalyType][]models.Anomaly)
	for _, anomaly := range existingAnomalies {
		byType[anomaly.Type] = append(byType[anomaly.Type], anomaly)
	}

	// Grouper les anomalies par entité
	byEntity := make(map[string][]models.Anomaly)
	for _, anomaly := range existingAnomalies {
		for _, entityID := range anomaly.EntityIDs {
			byEntity[entityID] = append(byEntity[entityID], anomaly)
		}
	}

	// 1. Détecter les entités avec plusieurs types d'anomalies (multi-facteur)
	for entityID, entityAnomalies := range byEntity {
		if len(entityAnomalies) < 2 {
			continue
		}

		// Compter les types distincts
		typeSet := make(map[models.AnomalyType]bool)
		for _, a := range entityAnomalies {
			typeSet[a.Type] = true
		}

		if len(typeSet) >= 2 {
			// Trouver le nom de l'entité
			var entityName string
			for _, e := range caseData.Entities {
				if e.ID == entityID {
					entityName = e.Name
					break
				}
			}

			// Calculer un score de risque composite
			riskScore := a.calculateCompositeRiskScore(entityAnomalies)
			severity := a.riskScoreToSeverity(riskScore)

			// Lister les types d'anomalies
			typeNames := []string{}
			for t := range typeSet {
				typeNames = append(typeNames, string(t))
			}

			anomalies = append(anomalies, models.Anomaly{
				ID:          uuid.New().String(),
				CaseID:      caseData.ID,
				Type:        models.AnomalyPattern, // Type composite
				Severity:    severity,
				Title:       "Convergence d'anomalies multiples",
				Description: fmt.Sprintf("'%s' présente %d types d'anomalies différents (%s) - Score de risque: %.0f%%",
					entityName, len(typeSet), strings.Join(typeNames, ", "), riskScore*100),
				DetectedAt: time.Now(),
				EntityIDs:  []string{entityID},
				Confidence: int(math.Min(60+float64(len(typeSet))*10+riskScore*20, 95)),
				Details: map[string]interface{}{
					"entity_name":       entityName,
					"anomaly_count":     len(entityAnomalies),
					"type_count":        len(typeSet),
					"anomaly_types":     typeNames,
					"risk_score":        riskScore,
					"detection_method":  "cross_correlation_entity",
				},
			})
		}
	}

	// 2. Détecter les corrélations temporelles entre types d'anomalies
	temporalCorrelations := a.detectTemporalCorrelations(caseData, existingAnomalies)
	anomalies = append(anomalies, temporalCorrelations...)

	// 3. Détecter les patterns de causalité potentiels
	causalPatterns := a.detectCausalPatterns(caseData, existingAnomalies, byType)
	anomalies = append(anomalies, causalPatterns...)

	return anomalies
}

// calculateCompositeRiskScore calcule un score de risque composite basé sur plusieurs anomalies
func (a *AnomalyService) calculateCompositeRiskScore(anomalies []models.Anomaly) float64 {
	if len(anomalies) == 0 {
		return 0
	}

	// Pondération par type d'anomalie
	typeWeights := map[models.AnomalyType]float64{
		models.AnomalyTimeline:      0.8,
		models.AnomalyFinancial:     1.0,
		models.AnomalyCommunication: 0.7,
		models.AnomalyBehavior:      0.6,
		models.AnomalyLocation:      0.5,
		models.AnomalyRelation:      0.9,
		models.AnomalyPattern:       0.6,
	}

	// Pondération par sévérité
	severityWeights := map[models.AnomalySeverity]float64{
		models.SeverityCritical: 1.0,
		models.SeverityHigh:     0.8,
		models.SeverityMedium:   0.5,
		models.SeverityLow:      0.3,
		models.SeverityInfo:     0.1,
	}

	var totalScore float64
	var maxPossibleScore float64

	for _, anomaly := range anomalies {
		typeWeight := typeWeights[anomaly.Type]
		if typeWeight == 0 {
			typeWeight = 0.5
		}
		severityWeight := severityWeights[anomaly.Severity]
		if severityWeight == 0 {
			severityWeight = 0.3
		}

		// Score = confiance * poids_type * poids_sévérité
		score := float64(anomaly.Confidence) / 100.0 * typeWeight * severityWeight
		totalScore += score
		maxPossibleScore += typeWeight // Maximum possible pour ce type
	}

	if maxPossibleScore == 0 {
		return 0
	}

	// Normaliser et ajouter un bonus pour la diversité des types
	typeCount := 0
	typeSet := make(map[models.AnomalyType]bool)
	for _, a := range anomalies {
		if !typeSet[a.Type] {
			typeSet[a.Type] = true
			typeCount++
		}
	}

	diversityBonus := float64(typeCount-1) * 0.1 // +10% par type supplémentaire
	normalizedScore := (totalScore / maxPossibleScore) + diversityBonus

	return math.Min(normalizedScore, 1.0)
}

// riskScoreToSeverity convertit un score de risque en niveau de sévérité
func (a *AnomalyService) riskScoreToSeverity(score float64) models.AnomalySeverity {
	if score >= 0.8 {
		return models.SeverityCritical
	} else if score >= 0.6 {
		return models.SeverityHigh
	} else if score >= 0.4 {
		return models.SeverityMedium
	}
	return models.SeverityLow
}

// detectTemporalCorrelations détecte les anomalies qui se produisent dans des fenêtres temporelles proches
func (a *AnomalyService) detectTemporalCorrelations(caseData *models.Case, anomalies []models.Anomaly) []models.Anomaly {
	result := []models.Anomaly{}

	if len(anomalies) < 3 {
		return result
	}

	// Trier les anomalies par date de détection (utiliser les événements associés si possible)
	type timedAnomaly struct {
		anomaly models.Anomaly
		time    time.Time
	}

	timedAnomalies := []timedAnomaly{}
	for _, anomaly := range anomalies {
		// Utiliser la date de détection par défaut
		t := anomaly.DetectedAt

		// Si des événements sont associés, utiliser leur timestamp
		if len(anomaly.EventIDs) > 0 {
			for _, event := range caseData.Timeline {
				if event.ID == anomaly.EventIDs[0] {
					t = event.Timestamp
					break
				}
			}
		}

		timedAnomalies = append(timedAnomalies, timedAnomaly{anomaly: anomaly, time: t})
	}

	sort.Slice(timedAnomalies, func(i, j int) bool {
		return timedAnomalies[i].time.Before(timedAnomalies[j].time)
	})

	// Fenêtre de corrélation: 48 heures
	windowDuration := 48 * time.Hour

	// Détecter les clusters d'anomalies dans la fenêtre
	i := 0
	for i < len(timedAnomalies) {
		cluster := []timedAnomaly{timedAnomalies[i]}
		j := i + 1

		// Trouver toutes les anomalies dans la fenêtre
		for j < len(timedAnomalies) {
			if timedAnomalies[j].time.Sub(timedAnomalies[i].time) <= windowDuration {
				cluster = append(cluster, timedAnomalies[j])
				j++
			} else {
				break
			}
		}

		// Si le cluster contient au moins 3 anomalies de types différents
		if len(cluster) >= 3 {
			typeSet := make(map[models.AnomalyType]bool)
			for _, ta := range cluster {
				typeSet[ta.anomaly.Type] = true
			}

			if len(typeSet) >= 2 {
				typeNames := []string{}
				for t := range typeSet {
					typeNames = append(typeNames, string(t))
				}

				startTime := cluster[0].time.Format("02/01/2006 15:04")
				endTime := cluster[len(cluster)-1].time.Format("02/01/2006 15:04")

				result = append(result, models.Anomaly{
					ID:          uuid.New().String(),
					CaseID:      caseData.ID,
					Type:        models.AnomalyPattern,
					Severity:    models.SeverityHigh,
					Title:       "Cluster temporel d'anomalies",
					Description: fmt.Sprintf("%d anomalies de %d types différents détectées entre %s et %s",
						len(cluster), len(typeSet), startTime, endTime),
					DetectedAt: time.Now(),
					Confidence: int(math.Min(65+float64(len(cluster))*5+float64(len(typeSet))*5, 95)),
					Details: map[string]interface{}{
						"cluster_size":     len(cluster),
						"type_count":       len(typeSet),
						"anomaly_types":    typeNames,
						"window_start":     startTime,
						"window_end":       endTime,
						"window_hours":     windowDuration.Hours(),
						"detection_method": "temporal_clustering",
					},
				})
			}
		}

		i = j
		if i == j && j < len(timedAnomalies) {
			i++
		}
	}

	return result
}

// detectCausalPatterns détecte les patterns de causalité potentiels entre types d'anomalies
func (a *AnomalyService) detectCausalPatterns(caseData *models.Case, anomalies []models.Anomaly, byType map[models.AnomalyType][]models.Anomaly) []models.Anomaly {
	result := []models.Anomaly{}

	// Patterns de causalité connus (type1 -> type2 = suspect)
	causalPatterns := []struct {
		cause      models.AnomalyType
		effect     models.AnomalyType
		name       string
		riskFactor float64
	}{
		{models.AnomalyFinancial, models.AnomalyBehavior, "Transaction suspecte suivie de changement comportemental", 0.8},
		{models.AnomalyCommunication, models.AnomalyTimeline, "Communication intense suivie d'incident temporel", 0.7},
		{models.AnomalyRelation, models.AnomalyFinancial, "Relation suspecte liée à des transactions", 0.9},
		{models.AnomalyBehavior, models.AnomalyLocation, "Changement comportemental associé à des mouvements suspects", 0.6},
		{models.AnomalyPattern, models.AnomalyFinancial, "Rupture de pattern suivie de transaction inhabituelle", 0.75},
	}

	for _, pattern := range causalPatterns {
		causeAnomalies, hasCause := byType[pattern.cause]
		effectAnomalies, hasEffect := byType[pattern.effect]

		if !hasCause || !hasEffect {
			continue
		}

		// Vérifier si des entités sont communes
		causeEntities := make(map[string]bool)
		for _, a := range causeAnomalies {
			for _, eid := range a.EntityIDs {
				causeEntities[eid] = true
			}
		}

		commonEntities := []string{}
		for _, a := range effectAnomalies {
			for _, eid := range a.EntityIDs {
				if causeEntities[eid] {
					commonEntities = append(commonEntities, eid)
				}
			}
		}

		if len(commonEntities) > 0 {
			// Trouver les noms des entités
			entityNames := []string{}
			for _, eid := range commonEntities {
				for _, e := range caseData.Entities {
					if e.ID == eid {
						entityNames = append(entityNames, e.Name)
						break
					}
				}
			}

			severity := models.SeverityMedium
			if pattern.riskFactor >= 0.8 {
				severity = models.SeverityHigh
			}
			if pattern.riskFactor >= 0.9 {
				severity = models.SeverityCritical
			}

			result = append(result, models.Anomaly{
				ID:          uuid.New().String(),
				CaseID:      caseData.ID,
				Type:        models.AnomalyPattern,
				Severity:    severity,
				Title:       "Pattern causal détecté",
				Description: fmt.Sprintf("%s - Entités concernées: %s",
					pattern.name, strings.Join(entityNames, ", ")),
				DetectedAt: time.Now(),
				EntityIDs:  commonEntities,
				Confidence: int(60 + pattern.riskFactor*30),
				Details: map[string]interface{}{
					"cause_type":       string(pattern.cause),
					"effect_type":      string(pattern.effect),
					"pattern_name":     pattern.name,
					"risk_factor":      pattern.riskFactor,
					"common_entities":  entityNames,
					"cause_count":      len(causeAnomalies),
					"effect_count":     len(effectAnomalies),
					"detection_method": "causal_pattern",
				},
			})
		}
	}

	return result
}

// =========================================
// Scoring Bayésien
// =========================================

// BayesianPriors contient les probabilités a priori pour le scoring bayésien
type BayesianPriors struct {
	// P(anomalie réelle | type) - probabilité qu'une anomalie soit réelle selon son type
	TypeReliability map[models.AnomalyType]float64
	// P(anomalie réelle | sévérité) - probabilité selon la sévérité
	SeverityReliability map[models.AnomalySeverity]float64
	// P(faux positif | détection_method)
	MethodFalsePositiveRate map[string]float64
}

// getDefaultPriors retourne les probabilités a priori par défaut
func getDefaultPriors() *BayesianPriors {
	return &BayesianPriors{
		TypeReliability: map[models.AnomalyType]float64{
			models.AnomalyTimeline:      0.85, // Les conflits temporels sont généralement fiables
			models.AnomalyFinancial:     0.75, // Les détections financières ont plus de faux positifs
			models.AnomalyCommunication: 0.70, // Patterns de communication peuvent être normaux
			models.AnomalyBehavior:      0.60, // Comportement est subjectif
			models.AnomalyLocation:      0.65, // Points chauds peuvent être normaux
			models.AnomalyRelation:      0.80, // Relations suspectes sont assez fiables
			models.AnomalyPattern:       0.70, // Patterns peuvent être coïncidences
		},
		SeverityReliability: map[models.AnomalySeverity]float64{
			models.SeverityCritical: 0.90, // Les critiques sont rarement des faux positifs
			models.SeverityHigh:     0.80,
			models.SeverityMedium:   0.65,
			models.SeverityLow:      0.50,
			models.SeverityInfo:     0.40,
		},
		MethodFalsePositiveRate: map[string]float64{
			"adaptive_zscore":          0.15,
			"adaptive_travel_time":     0.10,
			"amount_extraction_zscore": 0.20,
			"keyword_fallback":         0.40,
			"network_centrality":       0.25,
			"flow_asymmetry":           0.35,
			"connected_components":     0.20,
			"zscore_simple":            0.25,
			"cusum":                    0.20,
			"variance_change":          0.30,
			"periodic_analysis":        0.35,
			"entity_rhythm_change":     0.25,
			"cross_correlation_entity": 0.15,
			"temporal_clustering":      0.20,
			"causal_pattern":           0.25,
		},
	}
}

// applyBayesianScoring applique le scoring bayésien à une liste d'anomalies
func (a *AnomalyService) applyBayesianScoring(anomalies []models.Anomaly, caseData *models.Case) []models.Anomaly {
	priors := getDefaultPriors()

	// Calculer le contexte du cas pour ajuster les priors
	caseContext := a.analyzeCaseContext(caseData)

	result := make([]models.Anomaly, len(anomalies))
	for i, anomaly := range anomalies {
		// Calculer le score bayésien
		bayesianScore := a.calculateBayesianConfidence(anomaly, priors, caseContext)

		// Créer une copie avec le score mis à jour
		result[i] = anomaly
		result[i].Confidence = bayesianScore

		// Ajouter les détails bayésiens
		if result[i].Details == nil {
			result[i].Details = make(map[string]interface{})
		}
		result[i].Details["original_confidence"] = anomaly.Confidence
		result[i].Details["bayesian_confidence"] = bayesianScore
		result[i].Details["bayesian_adjustment"] = bayesianScore - anomaly.Confidence
	}

	return result
}

// CaseContext contient le contexte analysé du cas pour le scoring bayésien
type CaseContext struct {
	EntityCount           int
	EventCount            int
	EvidenceCount         int
	HasSuspects           bool
	HasVictims            bool
	TimelineSpanDays      float64
	AverageRelationsCount float64
	CaseComplexity        float64 // Score de complexité normalisé 0-1
}

// analyzeCaseContext analyse le contexte d'un cas pour ajuster les priors
func (a *AnomalyService) analyzeCaseContext(caseData *models.Case) *CaseContext {
	ctx := &CaseContext{
		EntityCount:   len(caseData.Entities),
		EventCount:    len(caseData.Timeline),
		EvidenceCount: len(caseData.Evidence),
	}

	// Vérifier la présence de suspects et victimes
	var totalRelations int
	for _, entity := range caseData.Entities {
		if entity.Role == models.RoleSuspect {
			ctx.HasSuspects = true
		} else if entity.Role == models.RoleVictim {
			ctx.HasVictims = true
		}
		totalRelations += len(entity.Relations)
	}

	if ctx.EntityCount > 0 {
		ctx.AverageRelationsCount = float64(totalRelations) / float64(ctx.EntityCount)
	}

	// Calculer l'étendue temporelle
	if len(caseData.Timeline) >= 2 {
		var minTime, maxTime time.Time
		for i, event := range caseData.Timeline {
			if i == 0 || event.Timestamp.Before(minTime) {
				minTime = event.Timestamp
			}
			if i == 0 || event.Timestamp.After(maxTime) {
				maxTime = event.Timestamp
			}
		}
		ctx.TimelineSpanDays = maxTime.Sub(minTime).Hours() / 24
	}

	// Calculer la complexité du cas (normalisée 0-1)
	// Plus il y a d'entités, d'événements et de relations, plus le cas est complexe
	entityScore := math.Min(float64(ctx.EntityCount)/20.0, 1.0)
	eventScore := math.Min(float64(ctx.EventCount)/50.0, 1.0)
	relationScore := math.Min(ctx.AverageRelationsCount/5.0, 1.0)
	ctx.CaseComplexity = (entityScore + eventScore + relationScore) / 3.0

	return ctx
}

// calculateBayesianConfidence calcule la confiance bayésienne pour une anomalie
func (a *AnomalyService) calculateBayesianConfidence(anomaly models.Anomaly, priors *BayesianPriors, ctx *CaseContext) int {
	// P(A) = confiance originale (prior)
	originalConfidence := float64(anomaly.Confidence) / 100.0

	// P(T|A) = fiabilité du type d'anomalie
	typeReliability := priors.TypeReliability[anomaly.Type]
	if typeReliability == 0 {
		typeReliability = 0.6 // Valeur par défaut
	}

	// P(S|A) = fiabilité de la sévérité
	severityReliability := priors.SeverityReliability[anomaly.Severity]
	if severityReliability == 0 {
		severityReliability = 0.5
	}

	// P(FP|M) = taux de faux positifs de la méthode de détection
	detectionMethod := ""
	if anomaly.Details != nil {
		if method, ok := anomaly.Details["detection_method"].(string); ok {
			detectionMethod = method
		}
	}
	falsePositiveRate := priors.MethodFalsePositiveRate[detectionMethod]
	if falsePositiveRate == 0 {
		falsePositiveRate = 0.25 // Valeur par défaut
	}

	// Ajustements contextuels
	contextAdjustment := 1.0

	// Si le cas a des suspects ET des victimes, les anomalies relationnelles sont plus fiables
	if ctx.HasSuspects && ctx.HasVictims && anomaly.Type == models.AnomalyRelation {
		contextAdjustment *= 1.15
	}

	// Pour les cas complexes, réduire légèrement la confiance (plus de bruit)
	if ctx.CaseComplexity > 0.7 {
		contextAdjustment *= 0.95
	}

	// Si l'anomalie concerne plusieurs entités, c'est plus fiable
	if len(anomaly.EntityIDs) > 1 {
		contextAdjustment *= 1.1
	}

	// Si l'anomalie a plusieurs événements associés, c'est plus fiable
	if len(anomaly.EventIDs) > 1 {
		contextAdjustment *= 1.1
	}

	// Formule bayésienne simplifiée:
	// P(A|Evidence) = P(A) * P(T|A) * P(S|A) * (1 - P(FP|M)) * contextAdjustment
	// Puis normalisation

	posterior := originalConfidence * typeReliability * severityReliability * (1 - falsePositiveRate) * contextAdjustment

	// Normaliser pour obtenir une probabilité entre 0 et 1
	// Utiliser une fonction sigmoïde pour lisser les valeurs extrêmes
	normalizedPosterior := 1.0 / (1.0 + math.Exp(-6*(posterior-0.3)))

	// Combiner avec la confiance originale (ne pas trop s'éloigner)
	// 60% bayésien, 40% original pour éviter les changements trop drastiques
	finalConfidence := 0.6*normalizedPosterior + 0.4*originalConfidence

	// Convertir en pourcentage et borner
	confidencePercent := int(math.Round(finalConfidence * 100))
	if confidencePercent < 10 {
		confidencePercent = 10
	}
	if confidencePercent > 99 {
		confidencePercent = 99
	}

	return confidencePercent
}

// RefineConfidenceWithFeedback affine la confiance basée sur le feedback utilisateur
// Cette méthode permet d'apprendre des corrections de l'utilisateur
func (a *AnomalyService) RefineConfidenceWithFeedback(caseID string, anomalyID string, wasCorrect bool) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	anomaly, err := a.GetAnomaly(caseID, anomalyID)
	if err != nil {
		return err
	}

	// Ajuster la confiance basée sur le feedback
	if wasCorrect {
		// L'anomalie était correcte, augmenter légèrement la confiance pour ce type
		anomaly.Confidence = int(math.Min(float64(anomaly.Confidence)*1.05, 99))
	} else {
		// L'anomalie était un faux positif, réduire la confiance
		anomaly.Confidence = int(math.Max(float64(anomaly.Confidence)*0.85, 10))
	}

	// Marquer comme révisé
	if anomaly.Details == nil {
		anomaly.Details = make(map[string]interface{})
	}
	anomaly.Details["user_feedback"] = wasCorrect
	anomaly.Details["feedback_applied"] = true

	return nil
}

// =========================================
// Méthodes utilitaires
// =========================================

func (a *AnomalyService) getOrCreateConfig(caseID string) *models.AnomalyDetectionConfig {
	if config, ok := a.configs[caseID]; ok {
		return config
	}

	// Configuration par défaut
	config := &models.AnomalyDetectionConfig{
		CaseID:                caseID,
		EnableTimeline:        true,
		EnableFinancial:       true,
		EnableCommunication:   true,
		EnableBehavior:        true,
		EnableLocation:        true,
		EnableRelation:        true,
		EnablePattern:         true,
		MinConfidence:         40,
		AutoAlert:             true,
		AlertSeverityThreshold: models.SeverityMedium,
	}
	a.configs[caseID] = config

	return config
}

func (a *AnomalyService) anomalyExists(caseID string, anomaly models.Anomaly) bool {
	return a.findExistingAnomaly(caseID, anomaly) != ""
}

// findExistingAnomaly retourne l'ID d'une anomalie similaire existante, ou "" si aucune
func (a *AnomalyService) findExistingAnomaly(caseID string, anomaly models.Anomaly) string {
	if caseAnomalies, ok := a.anomalies[caseID]; ok {
		for id, existing := range caseAnomalies {
			// Comparer par titre, type et description pour éviter les doublons
			if existing.Title == anomaly.Title &&
			   existing.Type == anomaly.Type &&
			   existing.Description == anomaly.Description {
				return id
			}
		}
	}
	return ""
}

func (a *AnomalyService) shouldAlert(anomaly models.Anomaly, config *models.AnomalyDetectionConfig) bool {
	return severityOrder(anomaly.Severity) <= severityOrder(config.AlertSeverityThreshold)
}

func (a *AnomalyService) createAlert(caseID string, anomaly *models.Anomaly) models.AnomalyAlert {
	alert := models.AnomalyAlert{
		ID:        uuid.New().String(),
		CaseID:    caseID,
		AnomalyID: anomaly.ID,
		AlertType: "immediate",
		Message:   fmt.Sprintf("[%s] %s: %s", anomaly.Severity, anomaly.Title, anomaly.Description),
		Priority:  anomaly.Severity,
		CreatedAt: time.Now(),
		IsRead:    false,
	}

	if a.alerts[caseID] == nil {
		a.alerts[caseID] = []*models.AnomalyAlert{}
	}
	a.alerts[caseID] = append(a.alerts[caseID], &alert)

	return alert
}

func (a *AnomalyService) countNew(anomalies []models.Anomaly) int {
	count := 0
	for _, anomaly := range anomalies {
		if anomaly.IsNew {
			count++
		}
	}
	return count
}

func (a *AnomalyService) countBySeverity(anomalies []models.Anomaly, severity models.AnomalySeverity) int {
	count := 0
	for _, anomaly := range anomalies {
		if anomaly.Severity == severity {
			count++
		}
	}
	return count
}

func (a *AnomalyService) generateSummary(result *models.AnomalyDetectionResult) string {
	if result.TotalAnomalies == 0 {
		return "Aucune anomalie détectée dans cette affaire."
	}

	summary := fmt.Sprintf("Détection terminée: %d anomalies identifiées", result.TotalAnomalies)

	if result.CriticalCount > 0 {
		summary += fmt.Sprintf(" dont %d critiques", result.CriticalCount)
	}
	if result.HighCount > 0 {
		summary += fmt.Sprintf(", %d élevées", result.HighCount)
	}
	if result.NewAnomalies > 0 {
		summary += fmt.Sprintf(". %d nouvelles anomalies", result.NewAnomalies)
	}

	return summary + "."
}

func (a *AnomalyService) calculateGapSeverity(gap time.Duration) models.AnomalySeverity {
	days := gap.Hours() / 24
	if days > 30 {
		return models.SeverityHigh
	} else if days > 14 {
		return models.SeverityMedium
	}
	return models.SeverityLow
}

func (a *AnomalyService) buildExplanationPrompt(anomaly *models.Anomaly, caseData *models.Case) string {
	var sb strings.Builder

	sb.WriteString("Analysez cette anomalie détectée dans une enquête forensique:\n\n")
	sb.WriteString(fmt.Sprintf("**Affaire**: %s\n", caseData.Name))
	sb.WriteString(fmt.Sprintf("**Type d'anomalie**: %s\n", anomaly.Type))
	sb.WriteString(fmt.Sprintf("**Sévérité**: %s\n", anomaly.Severity))
	sb.WriteString(fmt.Sprintf("**Titre**: %s\n", anomaly.Title))
	sb.WriteString(fmt.Sprintf("**Description**: %s\n", anomaly.Description))
	sb.WriteString(fmt.Sprintf("**Confiance**: %d%%\n\n", anomaly.Confidence))

	sb.WriteString("## Contexte de l'affaire\n")
	sb.WriteString(fmt.Sprintf("%s\n\n", caseData.Description))

	sb.WriteString("## Demande\n")
	sb.WriteString("1. Expliquez pourquoi cette anomalie est significative\n")
	sb.WriteString("2. Quelles pourraient être les causes de cette anomalie ?\n")
	sb.WriteString("3. Quelles actions d'investigation recommandez-vous ?\n")
	sb.WriteString("4. Y a-t-il des risques à ignorer cette anomalie ?\n")

	return sb.String()
}

func severityOrder(s models.AnomalySeverity) int {
	switch s {
	case models.SeverityCritical:
		return 1
	case models.SeverityHigh:
		return 2
	case models.SeverityMedium:
		return 3
	case models.SeverityLow:
		return 4
	case models.SeverityInfo:
		return 5
	default:
		return 6
	}
}
