package services

import (
	"fmt"
	"regexp"
	"strings"

	"forensicinvestigator/internal/models"
)

// N4LService gère le parsing et l'export N4L - Support complet du langage SSTorytime N4L
type N4LService struct {
	// Contextes
	contextRegex       *regexp.Regexp // :: contexte ::
	extendContextRegex *regexp.Regexp // +:: ajouter ::
	removeContextRegex *regexp.Regexp // -:: supprimer ::

	// Relations
	relationArrowRegex  *regexp.Regexp // A -> relation -> B
	relationParenRegex  *regexp.Regexp // A (relation) B
	chainedRelRegex     *regexp.Regexp // A (rel) B (rel) C
	equivalenceRegex    *regexp.Regexp // A <-> B
	groupRegex          *regexp.Regexp // A => {B, C, D}

	// Références et alias
	aliasDefRegex       *regexp.Regexp // @monalias texte
	aliasRefRegex       *regexp.Regexp // $alias.1
	continuationRegex   *regexp.Regexp // " (relation) B
	varRefRegex         *regexp.Regexp // $1, $2, etc.
	entityRefRegex      *regexp.Regexp // >entite

	// Modificateurs temporels
	neverRegex *regexp.Regexp // \never
	newRegex   *regexp.Regexp // \new

	// Séquences et sections
	sectionRegex  *regexp.Regexp // -section/chapter
	sequenceMode  bool
	currentAlias  string
	aliases       map[string][]string
	previousItems []string
}

// NewN4LService crée une nouvelle instance du service N4L avec support complet
func NewN4LService() *N4LService {
	return &N4LService{
		// Contextes
		contextRegex:       regexp.MustCompile(`^:{2,}\s*(.*?)\s*:{2,}$`),
		extendContextRegex: regexp.MustCompile(`^\+:{2,}\s*(.*?)\s*:{2,}$`),
		removeContextRegex: regexp.MustCompile(`^-:{2,}\s*(.*?)\s*:{2,}$`),

		// Relations - formats multiples
		relationArrowRegex:  regexp.MustCompile(`^(.+?)\s+->\s+(.+?)\s+->\s+(.+)$`),
		relationParenRegex:  regexp.MustCompile(`^(.+?)\s+\(([^)]+)\)\s+(.+)$`),
		chainedRelRegex:     regexp.MustCompile(`^(.+?)\s+\(([^)]+)\)\s+(.+?)\s+\(([^)]+)\)\s+(.+)$`),
		equivalenceRegex:    regexp.MustCompile(`^(.+?)\s+<->\s+(.+)$`),
		groupRegex:          regexp.MustCompile(`^(.+?)\s+=>\s+\{(.+)\}$`),

		// Références et alias
		aliasDefRegex:     regexp.MustCompile(`^@(\w+)\s+(.+)$`),
		aliasRefRegex:     regexp.MustCompile(`\$(\w+)\.(\d+)`),
		continuationRegex: regexp.MustCompile(`^"\s+\(([^)]+)\)\s+(.+)$`),
		varRefRegex:       regexp.MustCompile(`\$(\d+)`),
		entityRefRegex:    regexp.MustCompile(`>(\w+)`),

		// Modificateurs
		neverRegex: regexp.MustCompile(`^\\never\s+(.+)$`),
		newRegex:   regexp.MustCompile(`^\\new\s+(.+)$`),

		// Sections
		sectionRegex: regexp.MustCompile(`^-(\w+(?:/\w+)*)$`),

		// État
		sequenceMode:  false,
		aliases:       make(map[string][]string),
		previousItems: []string{},
	}
}

// ParsedN4L représente les données parsées d'un fichier N4L
type ParsedN4L struct {
	Notes      map[string][]string `json:"notes"`
	Subjects   []string            `json:"subjects"`
	Graph      models.GraphData    `json:"graph"`
	Sections   []string            `json:"sections"`
	Aliases    map[string][]string `json:"aliases"`
	Contexts   []string            `json:"contexts"`
	Sequences  [][]string          `json:"sequences"`
	TodoItems  []string            `json:"todo_items"`
}

// ParseN4L parse le contenu d'un fichier N4L avec support complet SSTorytime
func (s *N4LService) ParseN4L(content string) ParsedN4L {
	result := ParsedN4L{
		Notes:     make(map[string][]string),
		Subjects:  []string{},
		Sections:  []string{},
		Aliases:   make(map[string][]string),
		Contexts:  []string{},
		Sequences: [][]string{},
		TodoItems: []string{},
		Graph: models.GraphData{
			Nodes: []models.GraphNode{},
			Edges: []models.GraphEdge{},
		},
	}

	// Reset l'état du parser
	s.aliases = make(map[string][]string)
	s.previousItems = []string{}
	s.sequenceMode = false
	s.currentAlias = ""

	subjectsMap := make(map[string]bool)
	currentContext := "general"
	contextSet := make(map[string]bool)
	contextSet["general"] = true
	var currentSequence []string
	var previousItem string

	lines := strings.Split(content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Ignorer lignes vides et commentaires
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}

		// Détecter les TODO (lignes tout en majuscules)
		if isAllCaps(line) && len(line) > 3 {
			result.TodoItems = append(result.TodoItems, line)
			continue
		}

		// Section: -section/chapter
		if matches := s.sectionRegex.FindStringSubmatch(line); len(matches) == 2 {
			section := matches[1]
			result.Sections = append(result.Sections, section)
			continue
		}

		// Extension de contexte: +:: ajouter ::
		if matches := s.extendContextRegex.FindStringSubmatch(line); len(matches) == 2 {
			newContexts := strings.Split(matches[1], ",")
			for _, ctx := range newContexts {
				ctx = strings.TrimSpace(ctx)
				if ctx != "" {
					contextSet[ctx] = true
					// Activer le mode séquence si _sequence_ est présent
					if ctx == "_sequence_" || ctx == "_timeline_" {
						s.sequenceMode = true
						currentSequence = []string{}
					}
				}
			}
			currentContext = buildContextString(contextSet)
			continue
		}

		// Suppression de contexte: -:: supprimer ::
		if matches := s.removeContextRegex.FindStringSubmatch(line); len(matches) == 2 {
			removeContexts := strings.Split(matches[1], ",")
			for _, ctx := range removeContexts {
				ctx = strings.TrimSpace(ctx)
				delete(contextSet, ctx)
				// Désactiver le mode séquence
				if ctx == "_sequence_" || ctx == "_timeline_" {
					s.sequenceMode = false
					if len(currentSequence) > 0 {
						result.Sequences = append(result.Sequences, currentSequence)
					}
					currentSequence = []string{}
				}
			}
			currentContext = buildContextString(contextSet)
			if currentContext == "" {
				currentContext = "general"
				contextSet["general"] = true
			}
			continue
		}

		// Changement de contexte: :: contexte ::
		if matches := s.contextRegex.FindStringSubmatch(line); len(matches) == 2 {
			// Réinitialiser le contexte
			contextSet = make(map[string]bool)
			newContexts := strings.Split(matches[1], ",")
			for _, ctx := range newContexts {
				ctx = strings.TrimSpace(ctx)
				if ctx != "" {
					contextSet[ctx] = true
					if ctx == "_sequence_" || ctx == "_timeline_" {
						s.sequenceMode = true
						currentSequence = []string{}
					}
				}
			}
			currentContext = buildContextString(contextSet)
			if currentContext == "" {
				currentContext = "general"
			}
			if !contains(result.Contexts, currentContext) {
				result.Contexts = append(result.Contexts, currentContext)
			}
			if result.Notes[currentContext] == nil {
				result.Notes[currentContext] = []string{}
			}
			continue
		}

		// Définition d'alias: @alias texte
		if matches := s.aliasDefRegex.FindStringSubmatch(line); len(matches) == 3 {
			aliasName := matches[1]
			aliasContent := matches[2]
			s.currentAlias = aliasName
			s.aliases[aliasName] = []string{aliasContent}
			result.Aliases[aliasName] = []string{aliasContent}
			line = aliasContent // Continuer à parser le contenu
		}

		// Continuation: " (relation) B
		if strings.HasPrefix(line, "\"") {
			if previousItem != "" {
				if matches := s.continuationRegex.FindStringSubmatch(line); len(matches) == 3 {
					label := matches[1]
					target := matches[2]
					edge := &models.GraphEdge{
						From:    previousItem,
						To:      target,
						Label:   label,
						Type:    "relation",
						Context: currentContext,
					}
					result.Graph.Edges = append(result.Graph.Edges, *edge)
					if !subjectsMap[target] {
						subjectsMap[target] = true
						result.Subjects = append(result.Subjects, target)
					}
				}
			}
			continue
		}

		// Ajouter la note au contexte actuel
		if result.Notes[currentContext] == nil {
			result.Notes[currentContext] = []string{}
		}
		result.Notes[currentContext] = append(result.Notes[currentContext], line)

		// Parser la ligne pour extraire les arêtes et sujets
		edges, subjects := s.parseNoteToEdges(line, currentContext)
		for _, edge := range edges {
			result.Graph.Edges = append(result.Graph.Edges, edge)
		}
		for _, subj := range subjects {
			if !subjectsMap[subj] {
				subjectsMap[subj] = true
				result.Subjects = append(result.Subjects, subj)
			}
		}

		// Mode séquence: lier les éléments consécutifs
		if s.sequenceMode && len(subjects) > 0 {
			firstSubject := subjects[0]
			currentSequence = append(currentSequence, firstSubject)
			if previousItem != "" && previousItem != firstSubject {
				seqEdge := models.GraphEdge{
					From:    previousItem,
					To:      firstSubject,
					Label:   "puis",
					Type:    "sequence",
					Context: currentContext,
				}
				result.Graph.Edges = append(result.Graph.Edges, seqEdge)
			}
			previousItem = firstSubject
		} else if len(subjects) > 0 {
			previousItem = subjects[0]
		}
	}

	// Sauvegarder la dernière séquence
	if len(currentSequence) > 0 {
		result.Sequences = append(result.Sequences, currentSequence)
	}

	// Construire les nœuds à partir des sujets
	for _, subj := range result.Subjects {
		node := models.GraphNode{
			ID:    subj,
			Label: subj,
			Type:  "entity",
		}
		result.Graph.Nodes = append(result.Graph.Nodes, node)
	}

	return result
}

// parseNoteToEdges retourne plusieurs arêtes (pour les relations chaînées)
func (s *N4LService) parseNoteToEdges(note, context string) ([]models.GraphEdge, []string) {
	edge, subjects := s.parseNoteToEdge(note, context)
	if edge != nil {
		return []models.GraphEdge{*edge}, subjects
	}
	return []models.GraphEdge{}, subjects
}

// Fonctions utilitaires
func buildContextString(contextSet map[string]bool) string {
	var contexts []string
	for ctx := range contextSet {
		if ctx != "_sequence_" && ctx != "_timeline_" {
			contexts = append(contexts, ctx)
		}
	}
	return strings.Join(contexts, ", ")
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func isAllCaps(s string) bool {
	hasLetter := false
	for _, r := range s {
		if r >= 'a' && r <= 'z' {
			return false
		}
		if r >= 'A' && r <= 'Z' {
			hasLetter = true
		}
	}
	return hasLetter
}

// sanitizeN4LName nettoie un nom pour l'utiliser comme alias N4L
func sanitizeN4LName(name string) string {
	// Remplacer les caractères non alphanumériques par des underscores
	result := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			return r
		}
		return '_'
	}, name)
	// Supprimer les underscores consécutifs
	for strings.Contains(result, "__") {
		result = strings.ReplaceAll(result, "__", "_")
	}
	// Supprimer les underscores en début et fin
	result = strings.Trim(result, "_")
	return result
}

// resolveEntityName retourne le nom d'une entité à partir de son ID
func resolveEntityName(entityNames map[string]string, id string) string {
	if name, ok := entityNames[id]; ok {
		return name
	}
	return id
}

// parseNoteToEdge convertit une note en arête - Support complet N4L
func (s *N4LService) parseNoteToEdge(note, context string) (*models.GraphEdge, []string) {
	edgeType := "relation"
	processedNote := note

	// Vérifier les modificateurs temporels
	if matches := s.neverRegex.FindStringSubmatch(note); len(matches) == 2 {
		edgeType = "never"
		processedNote = strings.TrimSpace(matches[1])
	} else if matches := s.newRegex.FindStringSubmatch(note); len(matches) == 2 {
		edgeType = "new"
		processedNote = strings.TrimSpace(matches[1])
	}

	// Nettoyer les références d'entité >nom
	processedNote = s.entityRefRegex.ReplaceAllString(processedNote, "$1")

	// Relation format flèche: A -> relation -> B
	if matches := s.relationArrowRegex.FindStringSubmatch(processedNote); len(matches) == 4 {
		source := strings.TrimSpace(matches[1])
		label := strings.TrimSpace(matches[2])
		target := strings.TrimSpace(matches[3])

		if source != "" && target != "" {
			return &models.GraphEdge{
				From:    source,
				To:      target,
				Label:   label,
				Type:    edgeType,
				Context: context,
			}, []string{source, target}
		}
	}

	// Relation format parenthèses: A (relation) B (style N4L original)
	if matches := s.relationParenRegex.FindStringSubmatch(processedNote); len(matches) == 4 {
		source := strings.TrimSpace(matches[1])
		label := strings.TrimSpace(matches[2])
		target := strings.TrimSpace(matches[3])

		if source != "" && target != "" {
			return &models.GraphEdge{
				From:    source,
				To:      target,
				Label:   label,
				Type:    edgeType,
				Context: context,
			}, []string{source, target}
		}
	}

	// Équivalence: A <-> B
	if matches := s.equivalenceRegex.FindStringSubmatch(processedNote); len(matches) == 3 {
		source := strings.TrimSpace(matches[1])
		target := strings.TrimSpace(matches[2])

		if source != "" && target != "" {
			return &models.GraphEdge{
				From:    source,
				To:      target,
				Label:   "équivalent à",
				Type:    "equivalence",
				Context: context,
			}, []string{source, target}
		}
	}

	// Groupe: A => {B, C, D}
	if matches := s.groupRegex.FindStringSubmatch(processedNote); len(matches) == 3 {
		source := strings.TrimSpace(matches[1])
		members := strings.Split(matches[2], ",")
		subjects := []string{source}

		for _, m := range members {
			member := strings.TrimSpace(m)
			if member != "" {
				subjects = append(subjects, member)
			}
		}

		// Retourner la première relation du groupe
		if len(subjects) > 1 {
			return &models.GraphEdge{
				From:    source,
				To:      subjects[1],
				Label:   "contient",
				Type:    "group",
				Context: context,
			}, subjects
		}
	}

	return nil, nil
}

// ExportToN4L exporte les données d'une affaire en format N4L authentique SSTorytime
func (s *N4LService) ExportToN4L(caseData *models.Case) string {
	var sb strings.Builder

	// En-tête avec section
	sb.WriteString(fmt.Sprintf("-affaire/%s\n\n", sanitizeN4LName(caseData.ID)))
	sb.WriteString(fmt.Sprintf("# Affaire: %s\n", caseData.Name))
	sb.WriteString(fmt.Sprintf("# Type: %s\n", caseData.Type))
	sb.WriteString(fmt.Sprintf("# Statut: %s\n", caseData.Status))
	sb.WriteString(fmt.Sprintf("# Généré le: %s\n\n", caseData.UpdatedAt.Format("02/01/2006")))

	// Créer un map pour résoudre les IDs en noms
	entityNames := make(map[string]string)
	for _, e := range caseData.Entities {
		entityNames[e.ID] = e.Name
	}

	// Section victimes
	victims := filterEntitiesByRole(caseData.Entities, models.RoleVictim)
	if len(victims) > 0 {
		sb.WriteString(":: victimes ::\n\n")
		for _, v := range victims {
			sb.WriteString(fmt.Sprintf("@%s %s (rôle) victime\n", sanitizeN4LName(v.ID), v.Name))
			for key, val := range v.Attributes {
				sb.WriteString(fmt.Sprintf("    \" (%s) %s\n", key, val))
			}
			for _, r := range v.Relations {
				targetName := resolveEntityName(entityNames, r.ToID)
				sb.WriteString(fmt.Sprintf("    \" (%s) %s\n", r.Label, targetName))
			}
			sb.WriteString("\n")
		}
	}

	// Section suspects
	suspects := filterEntitiesByRole(caseData.Entities, models.RoleSuspect)
	if len(suspects) > 0 {
		sb.WriteString(":: suspects ::\n\n")
		for _, susp := range suspects {
			sb.WriteString(fmt.Sprintf("@%s %s (rôle) suspect\n", sanitizeN4LName(susp.ID), susp.Name))
			for key, val := range susp.Attributes {
				sb.WriteString(fmt.Sprintf("    \" (%s) %s\n", key, val))
			}
			for _, r := range susp.Relations {
				targetName := resolveEntityName(entityNames, r.ToID)
				sb.WriteString(fmt.Sprintf("    \" (%s) %s\n", r.Label, targetName))
			}
			sb.WriteString("\n")
		}
	}

	// Section témoins
	witnesses := filterEntitiesByRole(caseData.Entities, models.RoleWitness)
	if len(witnesses) > 0 {
		sb.WriteString(":: témoins ::\n\n")
		for _, w := range witnesses {
			sb.WriteString(fmt.Sprintf("@%s %s (rôle) témoin\n", sanitizeN4LName(w.ID), w.Name))
			for key, val := range w.Attributes {
				sb.WriteString(fmt.Sprintf("    \" (%s) %s\n", key, val))
			}
			for _, r := range w.Relations {
				targetName := resolveEntityName(entityNames, r.ToID)
				sb.WriteString(fmt.Sprintf("    \" (%s) %s\n", r.Label, targetName))
			}
			sb.WriteString("\n")
		}
	}

	// Section lieux et objets
	others := filterEntitiesByRole(caseData.Entities, models.RoleOther)
	if len(others) > 0 {
		sb.WriteString(":: lieux, objets ::\n\n")
		for _, o := range others {
			typeStr := string(o.Type)
			sb.WriteString(fmt.Sprintf("@%s %s (type) %s\n", sanitizeN4LName(o.ID), o.Name, typeStr))
			for _, r := range o.Relations {
				targetName := resolveEntityName(entityNames, r.ToID)
				sb.WriteString(fmt.Sprintf("    \" (%s) %s\n", r.Label, targetName))
			}
			sb.WriteString("\n")
		}
	}

	// Section preuves avec groupe
	if len(caseData.Evidence) > 0 {
		sb.WriteString(":: preuves, indices ::\n\n")
		// Grouper les preuves par type
		evidenceByType := make(map[string][]models.Evidence)
		for _, ev := range caseData.Evidence {
			typeStr := string(ev.Type)
			evidenceByType[typeStr] = append(evidenceByType[typeStr], ev)
		}
		for evType, evList := range evidenceByType {
			var names []string
			for _, ev := range evList {
				names = append(names, ev.Name)
			}
			sb.WriteString(fmt.Sprintf("Preuves %s => {%s}\n\n", evType, strings.Join(names, ", ")))
		}
		for _, ev := range caseData.Evidence {
			sb.WriteString(fmt.Sprintf("%s (type) %s\n", ev.Name, ev.Type))
			if ev.Location != "" {
				sb.WriteString(fmt.Sprintf("    \" (localisation) %s\n", ev.Location))
			}
			sb.WriteString(fmt.Sprintf("    \" (fiabilité) %d/10\n", ev.Reliability))
			for _, entityID := range ev.LinkedEntities {
				targetName := resolveEntityName(entityNames, entityID)
				sb.WriteString(fmt.Sprintf("    \" (concerne) %s\n", targetName))
			}
			sb.WriteString("\n")
		}
	}

	// Section chronologie avec mode séquence
	if len(caseData.Timeline) > 0 {
		sb.WriteString(":: chronologie ::\n\n")
		sb.WriteString("+:: _timeline_ ::\n\n")
		for _, evt := range caseData.Timeline {
			dateStr := evt.Timestamp.Format("02/01/2006 15:04")
			verifiedStr := ""
			if evt.Verified {
				verifiedStr = " [vérifié]"
			}
			sb.WriteString(fmt.Sprintf("%s %s (quand) %s%s\n", dateStr, evt.Title, evt.Description, verifiedStr))
			if evt.Location != "" {
				sb.WriteString(fmt.Sprintf("    \" (lieu) %s\n", evt.Location))
			}
			for _, entityID := range evt.Entities {
				targetName := resolveEntityName(entityNames, entityID)
				sb.WriteString(fmt.Sprintf("    \" (implique) %s\n", targetName))
			}
		}
		sb.WriteString("\n-:: _timeline_ ::\n\n")
	}

	// Section hypothèses
	if len(caseData.Hypotheses) > 0 {
		sb.WriteString(":: hypothèses, pistes ::\n\n")
		for _, h := range caseData.Hypotheses {
			status := "en attente"
			switch h.Status {
			case models.HypothesisSupported:
				status = "corroborée"
			case models.HypothesisRefuted:
				status = "réfutée"
			case models.HypothesisPartial:
				status = "partielle"
			}
			sb.WriteString(fmt.Sprintf("%s (statut) %s\n", h.Title, status))
			sb.WriteString(fmt.Sprintf("    \" (confiance) %d%%\n", h.ConfidenceLevel))
			if h.Description != "" {
				sb.WriteString(fmt.Sprintf("    \" (description) %s\n", h.Description))
			}
			if len(h.Questions) > 0 {
				sb.WriteString(fmt.Sprintf("    Questions => {%s}\n", strings.Join(h.Questions, "; ")))
			}
			sb.WriteString("\n")
		}
	}

	// Section relations (réseau global)
	sb.WriteString(":: réseau de relations ::\n\n")
	for _, e := range caseData.Entities {
		for _, r := range e.Relations {
			targetName := resolveEntityName(entityNames, r.ToID)
			if r.Verified {
				sb.WriteString(fmt.Sprintf("%s (%s) %s\n", e.Name, r.Label, targetName))
			} else {
				sb.WriteString(fmt.Sprintf("\\new %s (%s) %s\n", e.Name, r.Label, targetName))
			}
		}
	}

	return sb.String()
}

// filterEntitiesByRole filtre les entités par rôle
func filterEntitiesByRole(entities []models.Entity, role models.EntityRole) []models.Entity {
	var result []models.Entity
	for _, e := range entities {
		if e.Role == role {
			result = append(result, e)
		}
	}
	return result
}

// ConvertGraphToN4L convertit un graphe en format N4L
func (s *N4LService) ConvertGraphToN4L(graph models.GraphData) string {
	var sb strings.Builder

	// Grouper les arêtes par contexte
	edgesByContext := make(map[string][]models.GraphEdge)
	for _, edge := range graph.Edges {
		ctx := edge.Context
		if ctx == "" {
			ctx = "general"
		}
		edgesByContext[ctx] = append(edgesByContext[ctx], edge)
	}

	// Écrire chaque contexte
	for ctx, edges := range edgesByContext {
		sb.WriteString(fmt.Sprintf(":: %s ::\n\n", ctx))
		for _, edge := range edges {
			prefix := ""
			if edge.Type == "never" {
				prefix = "\\never "
			} else if edge.Type == "new" {
				prefix = "\\new "
			}

			if edge.Type == "equivalence" {
				sb.WriteString(fmt.Sprintf("    %s%s <-> %s\n", prefix, edge.From, edge.To))
			} else if edge.Type == "group" {
				sb.WriteString(fmt.Sprintf("    %s%s => {%s}\n", prefix, edge.From, edge.To))
			} else {
				sb.WriteString(fmt.Sprintf("    %s%s -> %s -> %s\n", prefix, edge.From, edge.Label, edge.To))
			}
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// ============================================
// STTypes - Spacetime Types from SSTorytime
// ============================================

// STType représente un type sémantique spacetime (inspiré de SSTorytime)
// Les STTypes définissent la nature causale/sémantique des relations
type STType int

const (
	STNear     STType = 0  // NEAR: Adjacence, proximité sans direction causale
	STLeadsTo  STType = 1  // LEADS_TO (+1): Causalité positive, mène à
	STLeadsFr  STType = -1 // LEADS_FROM (-1): Causalité inverse
	STContains STType = 2  // CONTAINS (+2): Containment, englobement
	STContainedBy STType = -2 // CONTAINED_BY (-2): Est contenu par
	STExpresses STType = 3  // EXPRESSES (+3): Expression sémantique, intention
	STExpressedBy STType = -3 // EXPRESSED_BY (-3): Est exprimé par
)

// STTypeInfo contient les informations sur un STType
type STTypeInfo struct {
	Type        STType `json:"type"`
	Code        string `json:"code"`        // +L, -L, +C, -C, +E, -E, N
	Name        string `json:"name"`
	Description string `json:"description"`
	Symbol      string `json:"symbol"`      // →, ←, ⊃, ⊂, ⇒, ⇐, ~
}

// STTypeMap contient tous les types spacetime disponibles
var STTypeMap = map[STType]STTypeInfo{
	STNear:     {STNear, "N", "Near", "Adjacence, proximité sans causalité", "~"},
	STLeadsTo:  {STLeadsTo, "+L", "Leads To", "Causalité: A mène à B", "→"},
	STLeadsFr:  {STLeadsFr, "-L", "Leads From", "Causalité inverse: B est source de A", "←"},
	STContains: {STContains, "+C", "Contains", "Containment: A contient B", "⊃"},
	STContainedBy: {STContainedBy, "-C", "Contained By", "Containment inverse: A est dans B", "⊂"},
	STExpresses: {STExpresses, "+E", "Expresses", "Expression: A exprime/définit B", "⇒"},
	STExpressedBy: {STExpressedBy, "-E", "Expressed By", "Expression inverse: A est exprimé par B", "⇐"},
}

// N4LEdgeWithSTType représente une arête avec son STType
type N4LEdgeWithSTType struct {
	models.GraphEdge
	STType  STType  `json:"st_type"`
	Weight  float64 `json:"weight,omitempty"`
}

// ParseSTTypeFromCode convertit un code STType en valeur
func ParseSTTypeFromCode(code string) STType {
	code = strings.ToUpper(strings.TrimSpace(code))
	switch code {
	case "+L", "L", "LEADSTO", "LEADS_TO":
		return STLeadsTo
	case "-L", "LEADSFROM", "LEADS_FROM":
		return STLeadsFr
	case "+C", "C", "CONTAINS":
		return STContains
	case "-C", "CONTAINEDBY", "CONTAINED_BY":
		return STContainedBy
	case "+E", "E", "+P", "P", "EXPRESSES":
		return STExpresses
	case "-E", "-P", "EXPRESSEDBY", "EXPRESSED_BY":
		return STExpressedBy
	default:
		return STNear
	}
}

// GetSTTypeInfo retourne les informations d'un STType
func GetSTTypeInfo(st STType) STTypeInfo {
	if info, ok := STTypeMap[st]; ok {
		return info
	}
	return STTypeMap[STNear]
}

// InferSTTypeFromRelation infère le STType à partir du label de relation
func InferSTTypeFromRelation(label string) STType {
	label = strings.ToLower(label)

	// Causalité positive (+L)
	causalPositive := []string{
		"mène à", "conduit à", "cause", "entraîne", "provoque",
		"leads to", "causes", "results in", "triggers",
		"puis", "ensuite", "alors", "donc",
	}
	for _, kw := range causalPositive {
		if strings.Contains(label, kw) {
			return STLeadsTo
		}
	}

	// Causalité inverse (-L)
	causalNegative := []string{
		"vient de", "résulte de", "est causé par",
		"comes from", "results from", "caused by",
	}
	for _, kw := range causalNegative {
		if strings.Contains(label, kw) {
			return STLeadsFr
		}
	}

	// Containment (+C)
	containsKeywords := []string{
		"contient", "inclut", "comprend", "englobe",
		"contains", "includes", "comprises", "encompasses",
		"possède", "has", "owns",
	}
	for _, kw := range containsKeywords {
		if strings.Contains(label, kw) {
			return STContains
		}
	}

	// Containment inverse (-C)
	containedKeywords := []string{
		"appartient à", "est dans", "fait partie de", "membre de",
		"belongs to", "is in", "part of", "member of",
		"situé à", "located in", "found in",
	}
	for _, kw := range containedKeywords {
		if strings.Contains(label, kw) {
			return STContainedBy
		}
	}

	// Expression (+E)
	expressKeywords := []string{
		"exprime", "signifie", "représente", "définit", "décrit",
		"expresses", "means", "represents", "defines", "describes",
		"est un", "is a", "type de", "type of",
	}
	for _, kw := range expressKeywords {
		if strings.Contains(label, kw) {
			return STExpresses
		}
	}

	// Expression inverse (-E)
	expressedKeywords := []string{
		"est défini par", "est décrit par", "exprimé par",
		"defined by", "described by", "expressed by",
	}
	for _, kw := range expressedKeywords {
		if strings.Contains(label, kw) {
			return STExpressedBy
		}
	}

	// Par défaut: NEAR (adjacence)
	return STNear
}

// ParseN4LWithSTTypes parse le contenu N4L avec support des STTypes
func (s *N4LService) ParseN4LWithSTTypes(content string) (ParsedN4L, []N4LEdgeWithSTType) {
	// Parser normalement
	parsed := s.ParseN4L(content)

	// Enrichir les arêtes avec STTypes
	var enrichedEdges []N4LEdgeWithSTType
	for _, edge := range parsed.Graph.Edges {
		stType := InferSTTypeFromRelation(edge.Label)

		// Vérifier si le label contient un code STType explicite
		// Format: "relation:+L" ou "relation:C"
		if idx := strings.LastIndex(edge.Label, ":"); idx > 0 {
			possibleCode := edge.Label[idx+1:]
			if parsedST := ParseSTTypeFromCode(possibleCode); parsedST != STNear || possibleCode == "N" {
				stType = parsedST
				edge.Label = strings.TrimSpace(edge.Label[:idx])
			}
		}

		// Vérifier si le label contient un poids
		// Format: "relation:0.8" ou "relation (0.8)"
		weight := 1.0
		if matches := regexp.MustCompile(`\((\d+\.?\d*)\)$`).FindStringSubmatch(edge.Label); len(matches) == 2 {
			if w, err := parseFloat(matches[1]); err == nil {
				weight = w
				edge.Label = strings.TrimSpace(edge.Label[:len(edge.Label)-len(matches[0])])
			}
		}

		enrichedEdges = append(enrichedEdges, N4LEdgeWithSTType{
			GraphEdge: edge,
			STType:    stType,
			Weight:    weight,
		})
	}

	return parsed, enrichedEdges
}

// parseFloat helper pour parser un float
func parseFloat(s string) (float64, error) {
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	return f, err
}

// ExportToN4LWithSTTypes exporte avec les annotations STType
func (s *N4LService) ExportToN4LWithSTTypes(caseData *models.Case, includeSTTypes bool) string {
	if !includeSTTypes {
		return s.ExportToN4L(caseData)
	}

	var sb strings.Builder

	// En-tête
	sb.WriteString(fmt.Sprintf("-affaire/%s\n\n", sanitizeN4LName(caseData.ID)))
	sb.WriteString(fmt.Sprintf("# Affaire: %s\n", caseData.Name))
	sb.WriteString(fmt.Sprintf("# Export avec STTypes (Semantic Spacetime)\n\n"))

	// Créer un map pour résoudre les IDs en noms
	entityNames := make(map[string]string)
	for _, e := range caseData.Entities {
		entityNames[e.ID] = e.Name
	}

	// Section réseau avec STTypes
	sb.WriteString(":: réseau sémantique ::\n\n")
	sb.WriteString("# Légende STTypes:\n")
	sb.WriteString("#   N  = Near (adjacence)\n")
	sb.WriteString("#   +L = Leads To (causalité)\n")
	sb.WriteString("#   -L = Leads From (source)\n")
	sb.WriteString("#   +C = Contains (containment)\n")
	sb.WriteString("#   -C = Contained By\n")
	sb.WriteString("#   +E = Expresses (définition)\n")
	sb.WriteString("#   -E = Expressed By\n\n")

	for _, e := range caseData.Entities {
		for _, r := range e.Relations {
			targetName := resolveEntityName(entityNames, r.ToID)
			stType := InferSTTypeFromRelation(r.Label)
			stInfo := GetSTTypeInfo(stType)

			// Format: Source (relation:STType) Target
			sb.WriteString(fmt.Sprintf("%s (%s:%s) %s\n", e.Name, r.Label, stInfo.Code, targetName))
		}
	}

	return sb.String()
}

// AnalyzeGraphSTTypes analyse les STTypes dans un graphe
type STTypeAnalysis struct {
	TypeDistribution map[string]int       `json:"type_distribution"`
	CausalChains     [][]string           `json:"causal_chains"`     // Chaînes de causalité
	Containers       map[string][]string  `json:"containers"`        // Nœuds contenant d'autres
	SemanticClusters [][]string           `json:"semantic_clusters"` // Clusters par expression
	Insights         []string             `json:"insights"`
}

// AnalyzeGraphBySTTypes analyse la structure sémantique du graphe par STTypes
func (s *N4LService) AnalyzeGraphBySTTypes(graph models.GraphData) STTypeAnalysis {
	analysis := STTypeAnalysis{
		TypeDistribution: make(map[string]int),
		Containers:       make(map[string][]string),
	}

	// Enrichir les arêtes avec STTypes
	stEdges := make(map[string][]N4LEdgeWithSTType)
	forwardAdj := make(map[string][]string) // Pour les chaînes causales

	for _, edge := range graph.Edges {
		stType := InferSTTypeFromRelation(edge.Label)
		stInfo := GetSTTypeInfo(stType)
		analysis.TypeDistribution[stInfo.Code]++

		enriched := N4LEdgeWithSTType{
			GraphEdge: edge,
			STType:    stType,
			Weight:    1.0,
		}
		stEdges[stInfo.Code] = append(stEdges[stInfo.Code], enriched)

		// Construire l'adjacence pour les chaînes causales
		if stType == STLeadsTo {
			forwardAdj[edge.From] = append(forwardAdj[edge.From], edge.To)
		}

		// Collecter les containments
		if stType == STContains {
			analysis.Containers[edge.From] = append(analysis.Containers[edge.From], edge.To)
		}
	}

	// Trouver les chaînes causales (DFS depuis les nœuds sources)
	visited := make(map[string]bool)
	var findChains func(node string, path []string)
	findChains = func(node string, path []string) {
		if len(forwardAdj[node]) == 0 {
			// Fin de chaîne
			if len(path) >= 3 {
				chainCopy := make([]string, len(path))
				copy(chainCopy, path)
				analysis.CausalChains = append(analysis.CausalChains, chainCopy)
			}
			return
		}

		for _, next := range forwardAdj[node] {
			if !visited[next] {
				visited[next] = true
				findChains(next, append(path, next))
				visited[next] = false
			}
		}
	}

	// Trouver les sources (nœuds sans arêtes entrantes de type LEADS_TO)
	hasIncoming := make(map[string]bool)
	for _, edges := range stEdges["+L"] {
		hasIncoming[edges.To] = true
	}

	for _, edges := range stEdges["+L"] {
		if !hasIncoming[edges.From] && !visited[edges.From] {
			visited[edges.From] = true
			findChains(edges.From, []string{edges.From})
			visited[edges.From] = false
		}
	}

	// Générer des insights
	analysis.Insights = s.generateSTTypeInsights(analysis)

	return analysis
}

// generateSTTypeInsights génère des insights basés sur l'analyse STType
func (s *N4LService) generateSTTypeInsights(analysis STTypeAnalysis) []string {
	var insights []string

	total := 0
	for _, count := range analysis.TypeDistribution {
		total += count
	}

	if total == 0 {
		insights = append(insights, "Aucune relation analysable détectée.")
		return insights
	}

	// Analyser la distribution
	if analysis.TypeDistribution["+L"] > total/3 {
		insights = append(insights, fmt.Sprintf("Le graphe est fortement causal (%d%% de relations LEADS_TO).", analysis.TypeDistribution["+L"]*100/total))
	}

	if analysis.TypeDistribution["+C"] > total/4 {
		insights = append(insights, fmt.Sprintf("Structure hiérarchique détectée (%d relations de containment).", analysis.TypeDistribution["+C"]))
	}

	if analysis.TypeDistribution["N"] > total/2 {
		insights = append(insights, "Le graphe contient majoritairement des relations de proximité (NEAR).")
	}

	// Chaînes causales
	if len(analysis.CausalChains) > 0 {
		maxLen := 0
		for _, chain := range analysis.CausalChains {
			if len(chain) > maxLen {
				maxLen = len(chain)
			}
		}
		insights = append(insights, fmt.Sprintf("%d chaîne(s) causale(s) détectée(s), longueur max: %d.", len(analysis.CausalChains), maxLen))
	}

	// Containers
	if len(analysis.Containers) > 0 {
		maxContained := 0
		biggestContainer := ""
		for container, contained := range analysis.Containers {
			if len(contained) > maxContained {
				maxContained = len(contained)
				biggestContainer = container
			}
		}
		if maxContained >= 3 {
			insights = append(insights, fmt.Sprintf("'%s' est un conteneur majeur avec %d éléments.", biggestContainer, maxContained))
		}
	}

	return insights
}
