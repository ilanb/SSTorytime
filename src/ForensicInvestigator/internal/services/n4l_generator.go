package services

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"forensicinvestigator/internal/models"
)

// N4LGeneratorService génère des fragments N4L à partir des structures de données
// Cela permet la synchronisation bidirectionnelle : UI -> N4L
type N4LGeneratorService struct {
	n4lService *N4LService
}

// NewN4LGeneratorService crée une nouvelle instance du générateur N4L
func NewN4LGeneratorService(n4lService *N4LService) *N4LGeneratorService {
	return &N4LGeneratorService{
		n4lService: n4lService,
	}
}

// N4LPatch représente une modification incrémentale du contenu N4L
type N4LPatch struct {
	Operation   string `json:"operation"`    // "add", "update", "delete"
	EntityType  string `json:"entity_type"`  // "entity", "evidence", "timeline", "hypothesis", "relation"
	EntityID    string `json:"entity_id"`    // ID de l'élément à modifier
	N4LFragment string `json:"n4l_fragment"` // Fragment N4L à insérer/remplacer
	Context     string `json:"context"`      // Contexte cible (ex: "victimes", "preuves")
}

// N4LPatchResult représente le résultat d'une opération de patch
type N4LPatchResult struct {
	Success     bool   `json:"success"`
	N4LContent  string `json:"n4l_content"`  // Contenu N4L mis à jour
	Message     string `json:"message"`
	ParsedData  *ForensicParsedN4L `json:"parsed_data,omitempty"`
}

// GenerateEntityN4L génère un fragment N4L pour une entité
func (g *N4LGeneratorService) GenerateEntityN4L(entity models.Entity) string {
	var sb strings.Builder

	// Déterminer le contexte approprié
	context := g.getContextForRole(entity.Role)
	sb.WriteString(fmt.Sprintf(":: %s ::\n\n", context))

	// Générer l'alias et la définition principale
	alias := sanitizeN4LName(entity.ID)
	sb.WriteString(fmt.Sprintf("@%s %s (type) %s\n", alias, entity.Name, entity.Type))
	sb.WriteString(fmt.Sprintf("    \" (role) %s\n", entity.Role))

	// Ajouter la description si présente
	if entity.Description != "" {
		sb.WriteString(fmt.Sprintf("    \" (description) %s\n", entity.Description))
	}

	// Ajouter les attributs
	for key, value := range entity.Attributes {
		sb.WriteString(fmt.Sprintf("    \" (%s) %s\n", key, value))
	}

	// Ajouter les relations
	for _, rel := range entity.Relations {
		if rel.Verified {
			sb.WriteString(fmt.Sprintf("    \" (%s) %s\n", rel.Label, rel.ToID))
		} else {
			sb.WriteString(fmt.Sprintf("    \" (\\new %s) %s\n", rel.Label, rel.ToID))
		}
	}

	sb.WriteString("\n")
	return sb.String()
}

// GenerateEvidenceN4L génère un fragment N4L pour une preuve
func (g *N4LGeneratorService) GenerateEvidenceN4L(evidence models.Evidence) string {
	var sb strings.Builder

	sb.WriteString(":: preuves ::\n\n")

	alias := sanitizeN4LName(evidence.ID)
	sb.WriteString(fmt.Sprintf("@%s %s (type) preuve\n", alias, evidence.Name))
	sb.WriteString(fmt.Sprintf("    \" (categorie) %s\n", evidence.Type))
	sb.WriteString(fmt.Sprintf("    \" (fiabilite) %d\n", evidence.Reliability))

	if evidence.Location != "" {
		sb.WriteString(fmt.Sprintf("    \" (localisation) %s\n", evidence.Location))
	}

	if evidence.Description != "" {
		sb.WriteString(fmt.Sprintf("    \" (description) %s\n", evidence.Description))
	}

	if evidence.CollectedBy != "" {
		sb.WriteString(fmt.Sprintf("    \" (collecte_par) %s\n", evidence.CollectedBy))
	}

	// Entités liées
	if len(evidence.LinkedEntities) > 0 {
		sb.WriteString(fmt.Sprintf("    \" (concerne) %s\n", strings.Join(evidence.LinkedEntities, ", ")))
	}

	if evidence.Notes != "" {
		sb.WriteString(fmt.Sprintf("    \" (notes) %s\n", evidence.Notes))
	}

	sb.WriteString("\n")
	return sb.String()
}

// GenerateTimelineEventN4L génère un fragment N4L pour un événement timeline
func (g *N4LGeneratorService) GenerateTimelineEventN4L(event models.Event) string {
	var sb strings.Builder

	sb.WriteString(":: chronologie ::\n\n")
	sb.WriteString("+:: _timeline_ ::\n\n")

	alias := sanitizeN4LName(event.ID)
	timestamp := event.Timestamp.Format("2006-01-02T15:04:05")
	sb.WriteString(fmt.Sprintf("@%s %s %s (type) evenement\n", alias, timestamp, event.Title))

	sb.WriteString(fmt.Sprintf("    \" (importance) %s\n", event.Importance))

	if event.Verified {
		sb.WriteString("    \" (verifie) true\n")
	}

	if event.Location != "" {
		sb.WriteString(fmt.Sprintf("    \" (lieu) %s\n", event.Location))
	}

	if event.Description != "" {
		sb.WriteString(fmt.Sprintf("    \" (description) %s\n", event.Description))
	}

	// Entités impliquées
	if len(event.Entities) > 0 {
		sb.WriteString(fmt.Sprintf("    \" (implique) %s\n", strings.Join(event.Entities, ", ")))
	}

	// Preuves liées
	if len(event.Evidence) > 0 {
		sb.WriteString(fmt.Sprintf("    \" (preuves) %s\n", strings.Join(event.Evidence, ", ")))
	}

	sb.WriteString("\n-:: _timeline_ ::\n\n")
	return sb.String()
}

// GenerateHypothesisN4L génère un fragment N4L pour une hypothèse
func (g *N4LGeneratorService) GenerateHypothesisN4L(hypothesis models.Hypothesis) string {
	var sb strings.Builder

	sb.WriteString(":: hypotheses ::\n\n")

	alias := sanitizeN4LName(hypothesis.ID)
	sb.WriteString(fmt.Sprintf("@%s %s (type) hypothese\n", alias, hypothesis.Title))

	// Statut
	status := g.formatHypothesisStatus(hypothesis.Status)
	sb.WriteString(fmt.Sprintf("    \" (statut) %s\n", status))
	sb.WriteString(fmt.Sprintf("    \" (confiance) %d%%\n", hypothesis.ConfidenceLevel))

	if hypothesis.Description != "" {
		sb.WriteString(fmt.Sprintf("    \" (description) %s\n", hypothesis.Description))
	}

	// Preuves supportant
	if len(hypothesis.SupportingEvidence) > 0 {
		sb.WriteString(fmt.Sprintf("    \" (supporte) %s\n", strings.Join(hypothesis.SupportingEvidence, ", ")))
	}

	// Preuves contradictoires
	if len(hypothesis.ContradictingEvidence) > 0 {
		sb.WriteString(fmt.Sprintf("    \" (contredit) %s\n", strings.Join(hypothesis.ContradictingEvidence, ", ")))
	}

	sb.WriteString(fmt.Sprintf("    \" (genere_par) %s\n", hypothesis.GeneratedBy))

	// Questions
	if len(hypothesis.Questions) > 0 {
		sb.WriteString(fmt.Sprintf("    \" (questions) %s\n", strings.Join(hypothesis.Questions, "; ")))
	}

	sb.WriteString("\n")
	return sb.String()
}

// GenerateRelationN4L génère un fragment N4L pour une relation
func (g *N4LGeneratorService) GenerateRelationN4L(relation models.Relation, entityNames map[string]string) string {
	var sb strings.Builder

	fromName := resolveEntityName(entityNames, relation.FromID)
	toName := resolveEntityName(entityNames, relation.ToID)

	// Ajouter le contexte si spécifié
	if relation.Context != "" {
		sb.WriteString(fmt.Sprintf(":: %s ::\n\n", relation.Context))
	}

	// Générer la relation
	if relation.Verified {
		sb.WriteString(fmt.Sprintf("%s (%s) %s\n", fromName, relation.Label, toName))
	} else {
		sb.WriteString(fmt.Sprintf("\\new %s (%s) %s\n", fromName, relation.Label, toName))
	}

	return sb.String()
}

// ApplyPatch applique un patch au contenu N4L
func (g *N4LGeneratorService) ApplyPatch(originalN4L string, patch N4LPatch, caseID string) N4LPatchResult {
	result := N4LPatchResult{
		Success: false,
	}

	switch patch.Operation {
	case "add":
		result = g.applyAddPatch(originalN4L, patch, caseID)
	case "update":
		result = g.applyUpdatePatch(originalN4L, patch, caseID)
	case "delete":
		result = g.applyDeletePatch(originalN4L, patch, caseID)
	default:
		result.Message = fmt.Sprintf("Opération inconnue: %s", patch.Operation)
	}

	return result
}

// applyAddPatch ajoute un nouveau fragment au contenu N4L
func (g *N4LGeneratorService) applyAddPatch(originalN4L string, patch N4LPatch, caseID string) N4LPatchResult {
	result := N4LPatchResult{
		Success: true,
	}

	// Trouver la section appropriée et ajouter le fragment
	contextMarker := fmt.Sprintf(":: %s ::", patch.Context)

	if strings.Contains(originalN4L, contextMarker) {
		// Ajouter après le marqueur de contexte existant
		idx := strings.Index(originalN4L, contextMarker)
		endIdx := idx + len(contextMarker)

		// Trouver la fin de la ligne
		if newlineIdx := strings.Index(originalN4L[endIdx:], "\n"); newlineIdx >= 0 {
			endIdx += newlineIdx + 1
		}

		// Insérer le fragment (sans le header de contexte car il existe déjà)
		fragmentWithoutContext := g.removeContextHeader(patch.N4LFragment)
		result.N4LContent = originalN4L[:endIdx] + "\n" + fragmentWithoutContext + originalN4L[endIdx:]
	} else {
		// Ajouter à la fin avec le contexte complet
		result.N4LContent = originalN4L + "\n" + patch.N4LFragment
	}

	// Parser le résultat pour validation
	parsed := g.n4lService.ParseForensicN4L(result.N4LContent, caseID)
	result.ParsedData = &parsed
	result.Message = "Fragment ajouté avec succès"

	return result
}

// applyUpdatePatch met à jour un fragment existant
func (g *N4LGeneratorService) applyUpdatePatch(originalN4L string, patch N4LPatch, caseID string) N4LPatchResult {
	result := N4LPatchResult{
		Success: true,
	}

	// Chercher l'alias existant
	aliasPattern := regexp.MustCompile(fmt.Sprintf(`(?m)^@%s\s+.+$`, regexp.QuoteMeta(patch.EntityID)))

	if aliasPattern.MatchString(originalN4L) {
		// Trouver le bloc complet (alias + continuations)
		blockPattern := regexp.MustCompile(fmt.Sprintf(`(?m)^@%s\s+.+\n(?:\s+"\s+\(.+\)\s+.+\n)*`, regexp.QuoteMeta(patch.EntityID)))

		// Extraire le fragment sans le header de contexte
		fragmentWithoutContext := g.removeContextHeader(patch.N4LFragment)

		result.N4LContent = blockPattern.ReplaceAllString(originalN4L, fragmentWithoutContext)
	} else {
		// L'élément n'existe pas, l'ajouter
		return g.applyAddPatch(originalN4L, patch, caseID)
	}

	// Parser le résultat pour validation
	parsed := g.n4lService.ParseForensicN4L(result.N4LContent, caseID)
	result.ParsedData = &parsed
	result.Message = "Fragment mis à jour avec succès"

	return result
}

// applyDeletePatch supprime un fragment existant
func (g *N4LGeneratorService) applyDeletePatch(originalN4L string, patch N4LPatch, caseID string) N4LPatchResult {
	result := N4LPatchResult{
		Success: true,
	}

	// Chercher et supprimer le bloc complet (alias + continuations)
	blockPattern := regexp.MustCompile(fmt.Sprintf(`(?m)^@%s\s+.+\n(?:\s+"\s+\(.+\)\s+.+\n)*\n?`, regexp.QuoteMeta(patch.EntityID)))

	if blockPattern.MatchString(originalN4L) {
		result.N4LContent = blockPattern.ReplaceAllString(originalN4L, "")
		result.Message = "Élément supprimé avec succès"
	} else {
		// Essayer de supprimer une ligne simple (entité sans alias)
		simplePattern := regexp.MustCompile(fmt.Sprintf(`(?m)^%s\s+\(.+\)\s+.+\n?`, regexp.QuoteMeta(patch.EntityID)))
		if simplePattern.MatchString(originalN4L) {
			result.N4LContent = simplePattern.ReplaceAllString(originalN4L, "")
			result.Message = "Élément supprimé avec succès"
		} else {
			result.Success = false
			result.Message = fmt.Sprintf("Élément non trouvé: %s", patch.EntityID)
			result.N4LContent = originalN4L
		}
	}

	// Parser le résultat pour validation
	if result.Success {
		parsed := g.n4lService.ParseForensicN4L(result.N4LContent, caseID)
		result.ParsedData = &parsed
	}

	return result
}

// removeContextHeader supprime le header de contexte d'un fragment N4L
func (g *N4LGeneratorService) removeContextHeader(fragment string) string {
	lines := strings.Split(fragment, "\n")
	var result []string
	skipContext := true

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Sauter les lignes de contexte au début
		if skipContext {
			if strings.HasPrefix(trimmed, "::") || trimmed == "" {
				continue
			}
			skipContext = false
		}

		// Sauter les markers de timeline +:: et -::
		if strings.HasPrefix(trimmed, "+::") || strings.HasPrefix(trimmed, "-::") {
			continue
		}

		result = append(result, line)
	}

	return strings.Join(result, "\n")
}

// MergeN4LFragments fusionne plusieurs fragments N4L
func (g *N4LGeneratorService) MergeN4LFragments(base string, fragments []string) string {
	result := base

	for _, fragment := range fragments {
		// Extraire le contexte du fragment
		contextMatch := regexp.MustCompile(`^::\s*(.+?)\s*::`).FindStringSubmatch(fragment)
		context := "general"
		if len(contextMatch) > 1 {
			context = contextMatch[1]
		}

		patch := N4LPatch{
			Operation:   "add",
			N4LFragment: fragment,
			Context:     context,
		}

		patchResult := g.applyAddPatch(result, patch, "")
		if patchResult.Success {
			result = patchResult.N4LContent
		}
	}

	return result
}

// ExportCaseToN4L génère le contenu N4L complet pour un cas
func (g *N4LGeneratorService) ExportCaseToN4L(caseData *models.Case) string {
	var sb strings.Builder

	// En-tête
	sb.WriteString(fmt.Sprintf("-affaire/%s\n\n", sanitizeN4LName(caseData.ID)))
	sb.WriteString(fmt.Sprintf("# Affaire: %s\n", caseData.Name))
	sb.WriteString(fmt.Sprintf("# Type: %s\n", caseData.Type))
	sb.WriteString(fmt.Sprintf("# Statut: %s\n", caseData.Status))
	sb.WriteString(fmt.Sprintf("# Généré le: %s\n", time.Now().Format("02/01/2006 15:04")))
	sb.WriteString("# Source: N4L Generator\n\n")

	// Créer un map pour résoudre les IDs en noms
	entityNames := make(map[string]string)
	for _, e := range caseData.Entities {
		entityNames[e.ID] = e.Name
	}

	// Grouper les entités par rôle
	entitiesByRole := make(map[models.EntityRole][]models.Entity)
	for _, e := range caseData.Entities {
		entitiesByRole[e.Role] = append(entitiesByRole[e.Role], e)
	}

	// Victimes
	if victims := entitiesByRole[models.RoleVictim]; len(victims) > 0 {
		sb.WriteString(":: victimes ::\n\n")
		for _, v := range victims {
			sb.WriteString(g.generateEntityBlock(v, entityNames))
		}
	}

	// Suspects
	if suspects := entitiesByRole[models.RoleSuspect]; len(suspects) > 0 {
		sb.WriteString(":: suspects ::\n\n")
		for _, s := range suspects {
			sb.WriteString(g.generateEntityBlock(s, entityNames))
		}
	}

	// Témoins
	if witnesses := entitiesByRole[models.RoleWitness]; len(witnesses) > 0 {
		sb.WriteString(":: témoins ::\n\n")
		for _, w := range witnesses {
			sb.WriteString(g.generateEntityBlock(w, entityNames))
		}
	}

	// Autres entités
	if others := entitiesByRole[models.RoleOther]; len(others) > 0 {
		sb.WriteString(":: lieux, objets ::\n\n")
		for _, o := range others {
			sb.WriteString(g.generateEntityBlock(o, entityNames))
		}
	}

	// Preuves
	if len(caseData.Evidence) > 0 {
		sb.WriteString(":: preuves ::\n\n")
		for _, ev := range caseData.Evidence {
			sb.WriteString(g.generateEvidenceBlock(ev))
		}
	}

	// Timeline
	if len(caseData.Timeline) > 0 {
		sb.WriteString(":: chronologie ::\n\n")
		sb.WriteString("+:: _timeline_ ::\n\n")
		for _, evt := range caseData.Timeline {
			sb.WriteString(g.generateTimelineBlock(evt))
		}
		sb.WriteString("-:: _timeline_ ::\n\n")
	}

	// Hypothèses
	if len(caseData.Hypotheses) > 0 {
		sb.WriteString(":: hypotheses ::\n\n")
		for _, h := range caseData.Hypotheses {
			sb.WriteString(g.generateHypothesisBlock(h))
		}
	}

	return sb.String()
}

// generateEntityBlock génère le bloc N4L pour une entité
func (g *N4LGeneratorService) generateEntityBlock(entity models.Entity, entityNames map[string]string) string {
	var sb strings.Builder

	alias := sanitizeN4LName(entity.ID)
	sb.WriteString(fmt.Sprintf("@%s %s (type) %s\n", alias, entity.Name, entity.Type))
	sb.WriteString(fmt.Sprintf("    \" (role) %s\n", entity.Role))

	if entity.Description != "" {
		sb.WriteString(fmt.Sprintf("    \" (description) %s\n", entity.Description))
	}

	for key, value := range entity.Attributes {
		sb.WriteString(fmt.Sprintf("    \" (%s) %s\n", key, value))
	}

	for _, rel := range entity.Relations {
		targetName := resolveEntityName(entityNames, rel.ToID)
		if rel.Verified {
			sb.WriteString(fmt.Sprintf("    \" (%s) %s\n", rel.Label, targetName))
		} else {
			sb.WriteString(fmt.Sprintf("    \" (\\new %s) %s\n", rel.Label, targetName))
		}
	}

	sb.WriteString("\n")
	return sb.String()
}

// generateEvidenceBlock génère le bloc N4L pour une preuve
func (g *N4LGeneratorService) generateEvidenceBlock(evidence models.Evidence) string {
	var sb strings.Builder

	alias := sanitizeN4LName(evidence.ID)
	sb.WriteString(fmt.Sprintf("@%s %s (type) preuve\n", alias, evidence.Name))
	sb.WriteString(fmt.Sprintf("    \" (categorie) %s\n", evidence.Type))
	sb.WriteString(fmt.Sprintf("    \" (fiabilite) %d\n", evidence.Reliability))

	if evidence.Location != "" {
		sb.WriteString(fmt.Sprintf("    \" (localisation) %s\n", evidence.Location))
	}

	if evidence.Description != "" {
		sb.WriteString(fmt.Sprintf("    \" (description) %s\n", evidence.Description))
	}

	if len(evidence.LinkedEntities) > 0 {
		sb.WriteString(fmt.Sprintf("    \" (concerne) %s\n", strings.Join(evidence.LinkedEntities, ", ")))
	}

	sb.WriteString("\n")
	return sb.String()
}

// generateTimelineBlock génère le bloc N4L pour un événement timeline
func (g *N4LGeneratorService) generateTimelineBlock(event models.Event) string {
	var sb strings.Builder

	alias := sanitizeN4LName(event.ID)
	timestamp := event.Timestamp.Format("2006-01-02T15:04:05")
	sb.WriteString(fmt.Sprintf("@%s %s %s (type) evenement\n", alias, timestamp, event.Title))

	sb.WriteString(fmt.Sprintf("    \" (importance) %s\n", event.Importance))

	if event.Verified {
		sb.WriteString("    \" (verifie) true\n")
	}

	if event.Location != "" {
		sb.WriteString(fmt.Sprintf("    \" (lieu) %s\n", event.Location))
	}

	if len(event.Entities) > 0 {
		sb.WriteString(fmt.Sprintf("    \" (implique) %s\n", strings.Join(event.Entities, ", ")))
	}

	// Preuves liées à l'événement
	if len(event.Evidence) > 0 {
		sb.WriteString(fmt.Sprintf("    \" (preuves) %s\n", strings.Join(event.Evidence, ", ")))
	}

	sb.WriteString("\n")
	return sb.String()
}

// generateHypothesisBlock génère le bloc N4L pour une hypothèse
func (g *N4LGeneratorService) generateHypothesisBlock(hypothesis models.Hypothesis) string {
	var sb strings.Builder

	alias := sanitizeN4LName(hypothesis.ID)
	sb.WriteString(fmt.Sprintf("@%s %s (type) hypothese\n", alias, hypothesis.Title))

	status := g.formatHypothesisStatus(hypothesis.Status)
	sb.WriteString(fmt.Sprintf("    \" (statut) %s\n", status))
	sb.WriteString(fmt.Sprintf("    \" (confiance) %d%%\n", hypothesis.ConfidenceLevel))

	if hypothesis.Description != "" {
		sb.WriteString(fmt.Sprintf("    \" (description) %s\n", hypothesis.Description))
	}

	if len(hypothesis.SupportingEvidence) > 0 {
		sb.WriteString(fmt.Sprintf("    \" (supporte) %s\n", strings.Join(hypothesis.SupportingEvidence, ", ")))
	}

	if len(hypothesis.ContradictingEvidence) > 0 {
		sb.WriteString(fmt.Sprintf("    \" (contredit) %s\n", strings.Join(hypothesis.ContradictingEvidence, ", ")))
	}

	sb.WriteString(fmt.Sprintf("    \" (genere_par) %s\n", hypothesis.GeneratedBy))

	sb.WriteString("\n")
	return sb.String()
}

// getContextForRole retourne le contexte N4L approprié pour un rôle d'entité
func (g *N4LGeneratorService) getContextForRole(role models.EntityRole) string {
	switch role {
	case models.RoleVictim:
		return "victimes"
	case models.RoleSuspect:
		return "suspects"
	case models.RoleWitness:
		return "témoins"
	case models.RoleInvestigator:
		return "enquêteurs"
	default:
		return "entités"
	}
}

// formatHypothesisStatus formate le statut d'hypothèse pour N4L
func (g *N4LGeneratorService) formatHypothesisStatus(status models.HypothesisStatus) string {
	switch status {
	case models.HypothesisPending:
		return "en_attente"
	case models.HypothesisSupported:
		return "corroboree"
	case models.HypothesisRefuted:
		return "refutee"
	case models.HypothesisPartial:
		return "partielle"
	default:
		return "en_attente"
	}
}

// ValidateN4L valide le contenu N4L et retourne les erreurs
func (g *N4LGeneratorService) ValidateN4L(content string) []string {
	var errors []string

	lines := strings.Split(content, "\n")
	contextStack := []string{}
	lineNum := 0

	for _, line := range lines {
		lineNum++
		trimmed := strings.TrimSpace(line)

		if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "//") {
			continue
		}

		// Vérifier les contextes non fermés
		if strings.HasPrefix(trimmed, "+::") {
			ctx := strings.TrimSpace(strings.TrimPrefix(strings.TrimSuffix(trimmed, "::"), "+::"))
			contextStack = append(contextStack, ctx)
		}

		if strings.HasPrefix(trimmed, "-::") {
			ctx := strings.TrimSpace(strings.TrimPrefix(strings.TrimSuffix(trimmed, "::"), "-::"))
			if len(contextStack) == 0 {
				errors = append(errors, fmt.Sprintf("Ligne %d: Fermeture de contexte '%s' sans ouverture correspondante", lineNum, ctx))
			} else {
				lastCtx := contextStack[len(contextStack)-1]
				if lastCtx != ctx {
					errors = append(errors, fmt.Sprintf("Ligne %d: Fermeture de contexte '%s' ne correspond pas à l'ouverture '%s'", lineNum, ctx, lastCtx))
				}
				contextStack = contextStack[:len(contextStack)-1]
			}
		}

		// Vérifier les continuations orphelines
		if strings.HasPrefix(trimmed, "\"") && !strings.Contains(trimmed, "(") {
			errors = append(errors, fmt.Sprintf("Ligne %d: Continuation sans attribut valide", lineNum))
		}

		// Vérifier les alias avec caractères invalides
		if strings.HasPrefix(trimmed, "@") {
			parts := strings.SplitN(trimmed[1:], " ", 2)
			if len(parts) > 0 {
				alias := parts[0]
				if !regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString(alias) {
					errors = append(errors, fmt.Sprintf("Ligne %d: Alias '%s' contient des caractères invalides", lineNum, alias))
				}
			}
		}
	}

	// Vérifier les contextes non fermés à la fin
	for _, ctx := range contextStack {
		errors = append(errors, fmt.Sprintf("Contexte '%s' ouvert mais jamais fermé", ctx))
	}

	return errors
}
