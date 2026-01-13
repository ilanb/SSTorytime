package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"unicode"
)

// NOTE: Modifiez cette URL pour pointer vers votre instance Ollama locale
var ollamaApiUrl = "http://localhost:11434/api/generate"

// --- Structures ---
type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
	Format string `json:"format,omitempty"`
}
type OllamaResponse struct {
	Response string `json:"response"`
}
type GraphData struct {
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}
type Node struct {
	ID      string `json:"id"`
	Label   string `json:"label"`
	Context string `json:"context"`
}
type Edge struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Label   string `json:"label"`
	Type    string `json:"type"` // "relation", "equivalence", "group"
	Context string `json:"context"`
}
type ParsedN4L struct {
	Subjects []string            `json:"subjects"`
	Notes    map[string][]string `json:"notes"`
}
type TimelineEvent struct {
	Order       int    `json:"order"`
	TimeHint    string `json:"timeHint"`
	Description string `json:"description"`
}
type AnalyzePathRequest struct {
	Path  []string            `json:"path"`
	Notes map[string][]string `json:"notes"`
}

type InvestigationQuestion struct {
	Question string   `json:"question"`
	Type     string   `json:"type"`
	Priority string   `json:"priority"`
	Context  string   `json:"context"`
	Nodes    []string `json:"nodes"`
	Hint     string   `json:"hint"`
}

// --- Handlers HTTP ---

func extractConceptsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}
	file, _, err := r.FormFile("textFile")
	if err != nil {
		http.Error(w, "Erreur lors de la lecture du fichier", http.StatusBadRequest)
		return
	}
	defer file.Close()
	content, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Erreur lors de la lecture du contenu du fichier", http.StatusInternalServerError)
		return
	}
	re := regexp.MustCompile(`[.!?]\s*`)
	sentences := re.Split(string(content), -1)
	var concepts []string
	for _, s := range sentences {
		trimmed := strings.TrimSpace(s)
		if trimmed != "" {
			concepts = append(concepts, trimmed)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(concepts)
}

func cleanOllamaJSON(rawJson string) string {
	start := strings.Index(rawJson, "{")
	end := strings.LastIndex(rawJson, "}")
	if start == -1 || end == -1 || start > end {
		start = strings.Index(rawJson, "[")
		end = strings.LastIndex(rawJson, "]")
		if start == -1 || end == -1 || start > end {
			return ""
		}
	}
	return rawJson[start : end+1]
}

func autoExtractSubjectsHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Impossible de lire le corps de la requête", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	prompt := `À partir du texte suivant, identifie les entités nommées (personnes, lieux, objets). Réponds uniquement avec un objet JSON contenant des listes pour chaque catégorie (ex: {"personnes": [...], "lieux": [...]}). Le texte : ` + string(body)
	reqPayload := OllamaRequest{
		Model: "zeffmuks/universal-ner:latest", Prompt: prompt, Stream: false, Format: "json",
	}
	reqBody, err := json.Marshal(reqPayload)
	if err != nil {
		http.Error(w, "Erreur de création de la requête Ollama", http.StatusInternalServerError)
		return
	}
	resp, err := http.Post(ollamaApiUrl, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		http.Error(w, "Erreur de connexion à Ollama", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Erreur de lecture de la réponse d'Ollama", http.StatusInternalServerError)
		return
	}
	log.Printf("Réponse brute d'Ollama (NER): %s", string(respBody))

	var ollamaResp OllamaResponse
	if err := json.Unmarshal(respBody, &ollamaResp); err != nil {
		http.Error(w, "Erreur de parsing de la réponse d'Ollama", http.StatusInternalServerError)
		return
	}
	cleanedJson := cleanOllamaJSON(ollamaResp.Response)
	if cleanedJson == "" {
		http.Error(w, "L'IA a renvoyé une réponse non-JSON.", http.StatusInternalServerError)
		return
	}
	var subjects []string
	var structuredResp map[string]interface{}
	if err := json.Unmarshal([]byte(cleanedJson), &structuredResp); err == nil {
		for _, value := range structuredResp {
			if items, ok := value.([]interface{}); ok {
				for _, item := range items {
					if s, ok := item.(string); ok {
						subjects = append(subjects, s)
					}
				}
			}
		}
	} else if err2 := json.Unmarshal([]byte(cleanedJson), &subjects); err2 != nil {
		http.Error(w, "Impossible de parser la réponse nettoyée de l'IA.", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(subjects)
}

func analyzeGraphHandler(w http.ResponseWriter, r *http.Request) {
	var graphData GraphData
	if err := json.NewDecoder(r.Body).Decode(&graphData); err != nil {
		http.Error(w, "Données de graphe invalides", http.StatusBadRequest)
		return
	}

	var sb strings.Builder
	sb.WriteString("Faits connus:\n")
	for _, edge := range graphData.Edges {
		switch edge.Type {
		case "relation":
			sb.WriteString(fmt.Sprintf("- %s %s %s.\n", edge.From, edge.Label, edge.To))
		case "equivalence":
			sb.WriteString(fmt.Sprintf("- %s est équivalent à %s.\n", edge.From, edge.To))
		case "group":
			sb.WriteString(fmt.Sprintf("- Le groupe '%s' contient %s.\n", edge.From, edge.To))
		}
	}

	prompt := fmt.Sprintf("Tu es un assistant d'enquête intelligent. Basé *uniquement* sur les faits suivants, rédige une synthèse de la situation. Quels sont les points clés, les suspects principaux, et les pistes à explorer ? Sois concis et direct.\n\n%s", sb.String())

	reqPayload := OllamaRequest{
		Model:  "gpt-oss:20b",
		Prompt: prompt,
		Stream: false,
	}

	reqBody, err := json.Marshal(reqPayload)
	if err != nil {
		http.Error(w, "Erreur de création de la requête Ollama", http.StatusInternalServerError)
		return
	}

	resp, err := http.Post(ollamaApiUrl, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		http.Error(w, "Erreur de connexion à Ollama", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Erreur de lecture de la réponse d'Ollama", http.StatusInternalServerError)
		return
	}
	log.Printf("Réponse brute d'Ollama (Analyse): %s", string(respBody))

	var ollamaResp OllamaResponse
	if err := json.Unmarshal(respBody, &ollamaResp); err != nil {
		http.Error(w, "Erreur de parsing de la réponse d'Ollama", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(ollamaResp.Response))
}

func analyzePathHandler(w http.ResponseWriter, r *http.Request) {
	var req AnalyzePathRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Données de chemin invalides", http.StatusBadRequest)
		return
	}

	var storyBuilder strings.Builder
	relationRegex := regexp.MustCompile(`^(.*) -> (.*) -> (.*)$`)
	groupRegex := regexp.MustCompile(`^(.*) => {(.*)}$`)

	for i := 0; i < len(req.Path)-1; i++ {
		fromNode := req.Path[i]
		toNode := req.Path[i+1]
		foundRelation := false

		for _, notes := range req.Notes {
			for _, note := range notes {
				if matches := relationRegex.FindStringSubmatch(note); len(matches) == 4 {
					source, _, target := strings.TrimSpace(matches[1]), strings.TrimSpace(matches[2]), strings.TrimSpace(matches[3])
					// La variable label (matches[2]) n'est pas utilisée, on la remplace par _
					if (source == fromNode && target == toNode) || (source == toNode && target == fromNode) {
						storyBuilder.WriteString(fmt.Sprintf("Fait %d: %s.\n", i+1, note))
						foundRelation = true
						break
					}
				} else if matches := groupRegex.FindStringSubmatch(note); len(matches) == 3 {
					parent := strings.TrimSpace(matches[1])
					children := strings.Split(matches[2], ";")
					for _, child := range children {
						childName := strings.TrimSpace(child)
						if (parent == fromNode && childName == toNode) || (parent == toNode && childName == fromNode) {
							storyBuilder.WriteString(fmt.Sprintf("Fait %d: %s.\n", i+1, note))
							foundRelation = true
							break
						}
					}
				}
			}
			if foundRelation {
				break
			}
		}
	}

	prompt := fmt.Sprintf("Tu es un analyste sémantique. La séquence de faits suivante représente un chemin logique découvert dans un graphe de connaissances : \n%s\nAnalyse cette séquence et détermine s'il s'agit principalement d'une chaîne de causalité, d'une simple corrélation, ou si elle révèle une possible contradiction. Justifie ta réponse en une ou deux phrases.", storyBuilder.String())

	reqPayload := OllamaRequest{
		Model:  "gpt-oss:20b",
		Prompt: prompt,
		Stream: false,
	}

	reqBody, err := json.Marshal(reqPayload)
	if err != nil {
		http.Error(w, "Erreur de création de la requête Ollama", http.StatusInternalServerError)
		return
	}

	resp, err := http.Post(ollamaApiUrl, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		http.Error(w, "Erreur de connexion à Ollama", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Erreur de lecture de la réponse d'Ollama", http.StatusInternalServerError)
		return
	}

	var ollamaResp OllamaResponse
	if err := json.Unmarshal(respBody, &ollamaResp); err != nil {
		http.Error(w, "Erreur de parsing de la réponse d'Ollama", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(ollamaResp.Response))
}

// Fonction helper pour extraire le premier sujet d'une note
func extractFirstSubject(note string) string {
	// Essayer de trouver le premier mot/concept significatif
	parts := strings.Fields(note)
	for _, part := range parts {
		part = strings.Trim(part, `"'`)
		if len(part) > 2 && !strings.Contains(part, "->") && !strings.Contains(part, "<->") {
			return part
		}
	}
	return ""
}

func parseN4LHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Erreur de lecture du corps de la requête", http.StatusInternalServerError)
		return
	}

	notes := make(map[string][]string)
	subjectsMap := make(map[string]bool)
	var currentContext string = "general"
	var lastSubject string = ""

	scanner := bufio.NewScanner(strings.NewReader(string(body)))

	// Regex pour différentes syntaxes
	contextRegex := regexp.MustCompile(`^:{2,}\s*(.*)\s*:{2,}$`)
	relationRegex := regexp.MustCompile(`^(.*) -> (.*) -> (.*)$`)
	equivalenceRegex := regexp.MustCompile(`^(.*) <-> (.*)$`)
	groupRegex := regexp.MustCompile(`^(.*) => {(.*)}$`)

	// Nouvelles regex pour syntaxes étendues
	parenthesesRegex := regexp.MustCompile(`^([^()]+)\s*\(([^)]+)\)\s*(.+)$`)
	annotationRegex := regexp.MustCompile(`>"([^"]+)"`)
	referenceRegex := regexp.MustCompile(`\$(\w+)\.(\d+)`)
	altEquivalenceRegex := regexp.MustCompile(`^(.+)\s*\(=\)\s*(.+)$`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Ignorer les lignes vides, commentaires et séparateurs de séquence
		if line == "" || strings.HasPrefix(line, "#") ||
			strings.HasPrefix(line, "+::") || strings.HasPrefix(line, "-::") {
			continue
		}

		// Gérer les contextes (avec multiple :)
		if matches := contextRegex.FindStringSubmatch(line); len(matches) > 1 {
			contextName := strings.TrimSpace(matches[1])
			// Ignorer les marqueurs de séquence
			if contextName == "_sequence_" || contextName == "sequence" {
				continue
			}
			currentContext = contextName
			if _, ok := notes[currentContext]; !ok {
				notes[currentContext] = []string{}
			}
			continue
		}

		// Nettoyer les annotations >"..." pour extraire les concepts
		cleanedLine := line
		if matches := annotationRegex.FindAllStringSubmatch(line, -1); len(matches) > 0 {
			for _, match := range matches {
				concept := match[1]
				cleanedLine = strings.ReplaceAll(cleanedLine, match[0], concept)
				subjectsMap[concept] = true
			}
		}

		// Gérer les références $variable
		if matches := referenceRegex.FindAllStringSubmatch(cleanedLine, -1); len(matches) > 0 {
			for _, match := range matches {
				// Remplacer par le contexte précédent si possible
				if match[1] == "goal" || match[1] == "PREV" {
					if lastSubject != "" {
						cleanedLine = strings.ReplaceAll(cleanedLine, match[0], lastSubject)
					} else {
						cleanedLine = strings.ReplaceAll(cleanedLine, match[0], "[REF:"+match[1]+"]")
					}
				}
			}
		}

		// Vérifier d'abord la syntaxe avec parenthèses : A (relation) B
		if matches := parenthesesRegex.FindStringSubmatch(cleanedLine); len(matches) == 4 {
			source := strings.TrimSpace(matches[1])
			relation := strings.TrimSpace(matches[2])
			target := strings.TrimSpace(matches[3])

			// Gérer le cas où source est "" (référence au sujet précédent)
			if source == `""` || source == `"` || source == "" {
				if lastSubject != "" {
					source = lastSubject
				} else if len(notes[currentContext]) > 0 {
					// Utiliser le dernier sujet mentionné dans le contexte
					lastNote := notes[currentContext][len(notes[currentContext])-1]
					source = extractFirstSubject(lastNote)
				}
			}

			if source != "" && source != `""` && target != "" {
				// Nettoyer les guillemets éventuels
				source = strings.Trim(source, `"`)
				target = strings.Trim(target, `"`)

				subjectsMap[source] = true
				subjectsMap[target] = true
				lastSubject = source

				// Convertir au format standard pour le graphe
				normalizedNote := fmt.Sprintf("%s -> %s -> %s", source, relation, target)
				notes[currentContext] = append(notes[currentContext], normalizedNote)
			}
			continue
		}

		// Parser l'équivalence alternative : A (=) B
		if matches := altEquivalenceRegex.FindStringSubmatch(cleanedLine); len(matches) == 3 {
			source := strings.TrimSpace(matches[1])
			target := strings.TrimSpace(matches[2])

			// Nettoyer les guillemets
			source = strings.Trim(source, `"`)
			target = strings.Trim(target, `"`)

			if source != "" && target != "" {
				subjectsMap[source] = true
				subjectsMap[target] = true
				lastSubject = source

				normalizedNote := fmt.Sprintf("%s <-> %s", source, target)
				notes[currentContext] = append(notes[currentContext], normalizedNote)
			}
			continue
		}

		// Syntaxe relation standard : A -> B -> C
		if matches := relationRegex.FindStringSubmatch(cleanedLine); len(matches) == 4 {
			source, _, target := strings.TrimSpace(matches[1]), strings.TrimSpace(matches[2]), strings.TrimSpace(matches[3])
			subjectsMap[source] = true
			subjectsMap[target] = true
			lastSubject = source
			notes[currentContext] = append(notes[currentContext], cleanedLine)
			continue
		}

		// Syntaxe équivalence standard : A <-> B
		if matches := equivalenceRegex.FindStringSubmatch(cleanedLine); len(matches) == 3 {
			source, target := strings.TrimSpace(matches[1]), strings.TrimSpace(matches[2])
			subjectsMap[source] = true
			subjectsMap[target] = true
			lastSubject = source
			notes[currentContext] = append(notes[currentContext], cleanedLine)
			continue
		}

		// Syntaxe groupe : A => { B; C; D }
		if matches := groupRegex.FindStringSubmatch(cleanedLine); len(matches) == 3 {
			parent := strings.TrimSpace(matches[1])
			subjectsMap[parent] = true
			lastSubject = parent
			children := strings.Split(matches[2], ";")
			for _, child := range children {
				childName := strings.Trim(strings.TrimSpace(child), `"`)
				if childName != "" {
					subjectsMap[childName] = true
				}
			}
			notes[currentContext] = append(notes[currentContext], cleanedLine)
			continue
		}

		// Si ce n'est pas une relation reconnue mais contient des informations utiles
		// on peut essayer d'extraire des sujets des lignes simples
		if cleanedLine != "" && !strings.HasPrefix(cleanedLine, "::") {
			// Extraire les mots qui pourraient être des sujets (mots capitalisés)
			words := strings.Fields(cleanedLine)
			for _, word := range words {
				word = strings.Trim(word, `"'.,;:!?`)
				if len(word) > 2 && unicode.IsUpper(rune(word[0])) {
					subjectsMap[word] = true
				}
			}
			// Ajouter comme note contextuelle
			notes[currentContext] = append(notes[currentContext], cleanedLine)
		}
	}

	var subjects []string
	for s := range subjectsMap {
		if s != "" && s != `""` && s != "[" && s != "]" {
			subjects = append(subjects, s)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ParsedN4L{Subjects: subjects, Notes: notes})
}

func parseN4LToGraph(w http.ResponseWriter, r *http.Request) {
	var n4lNotes map[string][]string
	if err := json.NewDecoder(r.Body).Decode(&n4lNotes); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	nodesMap := make(map[string]string) // Map ID to context
	var edges []Edge

	// Regex pour toutes les syntaxes
	relationRegex := regexp.MustCompile(`^(.*) -> (.*) -> (.*)$`)
	equivalenceRegex := regexp.MustCompile(`^(.*) <-> (.*)$`)
	groupRegex := regexp.MustCompile(`^(.*) => {(.*)}$`)

	// Nouvelles regex pour syntaxes étendues (au cas où elles n'ont pas été normalisées)
	parenthesesRegex := regexp.MustCompile(`^([^()]+)\s*\(([^)]+)\)\s*(.+)$`)
	altEquivalenceRegex := regexp.MustCompile(`^(.+)\s*\(=\)\s*(.+)$`)
	annotationRegex := regexp.MustCompile(`>"([^"]+)"`)

	for context, notes := range n4lNotes {
		for _, note := range notes {
			// Nettoyer les annotations si présentes
			cleanedNote := note
			if matches := annotationRegex.FindAllStringSubmatch(note, -1); len(matches) > 0 {
				for _, match := range matches {
					cleanedNote = strings.ReplaceAll(cleanedNote, match[0], match[1])
				}
			}

			// Essayer d'abord le format standard (déjà normalisé)
			if matches := relationRegex.FindStringSubmatch(cleanedNote); len(matches) == 4 {
				source, label, target := strings.TrimSpace(matches[1]), strings.TrimSpace(matches[2]), strings.TrimSpace(matches[3])

				// Nettoyer les références [REF:...]
				source = strings.TrimPrefix(source, "[REF:")
				source = strings.TrimSuffix(source, "]")
				target = strings.TrimPrefix(target, "[REF:")
				target = strings.TrimSuffix(target, "]")

				if source != "" && target != "" {
					nodesMap[source] = context
					nodesMap[target] = context
					edges = append(edges, Edge{From: source, To: target, Label: label, Type: "relation", Context: context})
				}
			} else if matches := equivalenceRegex.FindStringSubmatch(cleanedNote); len(matches) == 3 {
				source, target := strings.TrimSpace(matches[1]), strings.TrimSpace(matches[2])
				if source != "" && target != "" {
					nodesMap[source] = context
					nodesMap[target] = context
					edges = append(edges, Edge{From: source, To: target, Label: "", Type: "equivalence", Context: context})
				}
			} else if matches := groupRegex.FindStringSubmatch(cleanedNote); len(matches) == 3 {
				parent := strings.TrimSpace(matches[1])
				childrenStr := strings.TrimSpace(matches[2])
				children := strings.Split(childrenStr, ";")

				if parent != "" {
					nodesMap[parent] = context
					for _, child := range children {
						childName := strings.Trim(strings.TrimSpace(child), `"`)
						if childName != "" {
							nodesMap[childName] = context
							edges = append(edges, Edge{From: parent, To: childName, Label: "contient", Type: "group", Context: context})
						}
					}
				}
			} else if matches := parenthesesRegex.FindStringSubmatch(cleanedNote); len(matches) == 4 {
				// Cas où la syntaxe parenthèse n'a pas été normalisée
				source := strings.Trim(strings.TrimSpace(matches[1]), `"`)
				relation := strings.TrimSpace(matches[2])
				target := strings.Trim(strings.TrimSpace(matches[3]), `"`)

				if source != "" && source != `""` && target != "" {
					nodesMap[source] = context
					nodesMap[target] = context
					edges = append(edges, Edge{From: source, To: target, Label: relation, Type: "relation", Context: context})
				}
			} else if matches := altEquivalenceRegex.FindStringSubmatch(cleanedNote); len(matches) == 3 {
				// Cas où l'équivalence alternative n'a pas été normalisée
				source := strings.Trim(strings.TrimSpace(matches[1]), `"`)
				target := strings.Trim(strings.TrimSpace(matches[2]), `"`)

				if source != "" && target != "" {
					nodesMap[source] = context
					nodesMap[target] = context
					edges = append(edges, Edge{From: source, To: target, Label: "", Type: "equivalence", Context: context})
				}
			}
		}
	}

	var nodes []Node
	for nodeID, context := range nodesMap {
		if nodeID != "" && nodeID != `""` && nodeID != "[" && nodeID != "]" {
			nodes = append(nodes, Node{ID: nodeID, Label: nodeID, Context: context})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(GraphData{Nodes: nodes, Edges: edges})
}

func timelineHandler(w http.ResponseWriter, r *http.Request) {
	var n4lNotes map[string][]string
	if err := json.NewDecoder(r.Body).Decode(&n4lNotes); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var events []TimelineEvent
	eventMap := make(map[string]TimelineEvent)

	timeRegexes := map[string]int{
		`soirée`:    1,
		`22h`:       2,
		`lendemain`: 3,
	}

	for _, notes := range n4lNotes {
		for _, note := range notes {
			for pattern, order := range timeRegexes {
				re := regexp.MustCompile(pattern)
				if re.MatchString(note) {
					if _, exists := eventMap[note]; !exists {
						eventMap[note] = TimelineEvent{
							Order:       order,
							TimeHint:    pattern,
							Description: note,
						}
					}
				}
			}
		}
	}

	for _, event := range eventMap {
		events = append(events, event)
	}

	sort.Slice(events, func(i, j int) bool {
		return events[i].Order < events[j].Order
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}

func findAllPathsHandler(w http.ResponseWriter, r *http.Request) {
	var req map[string][]string
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	adj := make(map[string][]string)
	allNodes := make(map[string]bool)
	relationRegex := regexp.MustCompile(`^(.*) -> (.*) -> (.*)$`)
	equivalenceRegex := regexp.MustCompile(`^(.*) <-> (.*)$`)
	groupRegex := regexp.MustCompile(`^(.*) => {(.*)}$`)

	for _, notes := range req {
		for _, note := range notes {
			if matches := relationRegex.FindStringSubmatch(note); len(matches) == 4 {
				source, target := strings.TrimSpace(matches[1]), strings.TrimSpace(matches[3])
				adj[source] = append(adj[source], target)
				adj[target] = append(adj[target], source)
				allNodes[source] = true
				allNodes[target] = true
			} else if matches := equivalenceRegex.FindStringSubmatch(note); len(matches) == 3 {
				source, target := strings.TrimSpace(matches[1]), strings.TrimSpace(matches[2])
				adj[source] = append(adj[source], target)
				adj[target] = append(adj[target], source)
				allNodes[source] = true
				allNodes[target] = true
			} else if matches := groupRegex.FindStringSubmatch(note); len(matches) == 3 {
				parent := strings.TrimSpace(matches[1])
				children := strings.Split(matches[2], ";")
				allNodes[parent] = true
				for _, child := range children {
					childName := strings.TrimSpace(child)
					adj[parent] = append(adj[parent], childName)
					adj[childName] = append(adj[childName], parent)
					allNodes[childName] = true
				}
			}
		}
	}

	var nodesList []string
	for node := range allNodes {
		nodesList = append(nodesList, node)
	}

	var allPaths [][]string
	for i := 0; i < len(nodesList); i++ {
		for j := i + 1; j < len(nodesList); j++ {
			startNode, endNode := nodesList[i], nodesList[j]

			queue := [][]string{{startNode}}
			visited := make(map[string]bool)
			visited[startNode] = true

			for len(queue) > 0 {
				path := queue[0]
				queue = queue[1:]
				node := path[len(path)-1]

				if node == endNode {
					if len(path) > 2 {
						allPaths = append(allPaths, path)
					}
					break
				}

				for _, neighbor := range adj[node] {
					if !visited[neighbor] {
						visited[neighbor] = true
						newPath := make([]string, len(path))
						copy(newPath, path)
						newPath = append(newPath, neighbor)
						queue = append(queue, newPath)
					}
				}
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(allPaths)
}

func detectTemporalPatternsHandler(w http.ResponseWriter, r *http.Request) {
	var n4lNotes map[string][]string
	if err := json.NewDecoder(r.Body).Decode(&n4lNotes); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	type TemporalPattern struct {
		Pattern     string   `json:"pattern"`
		Occurrences []string `json:"occurrences"`
		Suggestions []string `json:"suggestions"`
	}

	// Élargir la liste des marqueurs temporels
	temporalMarkers := map[string]string{
		"avant":     "précède",
		"après":     "suit",
		"puis":      "puis",
		"ensuite":   "ensuite",
		"pendant":   "pendant",
		"durant":    "durant",
		"alors que": "en parallèle de",
		"jusqu'à":   "jusqu'à",
		"depuis":    "depuis",
		"vers":      "vers",
		"à":         "à",
		"lorsque":   "au moment où",
		"quand":     "quand",
		"lendemain": "suit",
		"veille":    "précède",
		"soirée":    "pendant",
		"matin":     "au début de",
		"soir":      "à la fin de",
	}

	var patterns []TemporalPattern
	detectedPhrases := make(map[string]bool)

	// Analyser toutes les notes pour détecter les marqueurs temporels
	for _, notes := range n4lNotes {
		for _, note := range notes {
			// Nettoyer la note
			cleanNote := strings.TrimSpace(note)
			lowerNote := strings.ToLower(cleanNote)

			// Log pour debug
			log.Printf("Analysing note: %s", cleanNote)

			for marker, relation := range temporalMarkers {
				if strings.Contains(lowerNote, marker) && !detectedPhrases[cleanNote] {
					detectedPhrases[cleanNote] = true

					// Extraire les entités autour du marqueur
					suggestions := analyzeTemporalContext(cleanNote, marker, relation)

					pattern := TemporalPattern{
						Pattern:     marker,
						Occurrences: []string{cleanNote},
						Suggestions: suggestions,
					}

					// Vérifier si ce pattern existe déjà
					found := false
					for i, p := range patterns {
						if p.Pattern == marker {
							patterns[i].Occurrences = append(patterns[i].Occurrences, cleanNote)
							// Ajouter les nouvelles suggestions
							for _, sug := range suggestions {
								if !contains(patterns[i].Suggestions, sug) {
									patterns[i].Suggestions = append(patterns[i].Suggestions, sug)
								}
							}
							found = true
							break
						}
					}
					if !found {
						patterns = append(patterns, pattern)
					}
				}
			}
		}
	}

	// Détecter les séquences implicites (heures, dates)
	timeRegex := regexp.MustCompile(`(\d{1,2}h\d{0,2}|\d{1,2}:\d{2})`)
	dateRegex := regexp.MustCompile(`(\d{1,2}[/-]\d{1,2}[/-]\d{2,4}|lendemain|veille|matin|soir|soirée|midi|minuit)`)

	for _, notes := range n4lNotes {
		for _, note := range notes {
			cleanNote := strings.TrimSpace(note)

			if timeMatches := timeRegex.FindAllString(cleanNote, -1); len(timeMatches) > 0 {
				for _, timeMatch := range timeMatches {
					// Extraire les éléments autour de l'heure
					suggestions := []string{
						fmt.Sprintf("Créer un événement temporel à %s", timeMatch),
					}

					// Essayer d'extraire le sujet de la note
					if subject := extractSubjectFromNote(cleanNote); subject != "" {
						suggestions = append(suggestions, fmt.Sprintf("%s -> se passe à -> %s", subject, timeMatch))
					}

					pattern := TemporalPattern{
						Pattern:     "heure",
						Occurrences: []string{cleanNote},
						Suggestions: suggestions,
					}

					// Vérifier si un pattern "heure" existe déjà
					found := false
					for i, p := range patterns {
						if p.Pattern == "heure" {
							if !contains(patterns[i].Occurrences, cleanNote) {
								patterns[i].Occurrences = append(patterns[i].Occurrences, cleanNote)
							}
							for _, sug := range suggestions {
								if !contains(patterns[i].Suggestions, sug) {
									patterns[i].Suggestions = append(patterns[i].Suggestions, sug)
								}
							}
							found = true
							break
						}
					}
					if !found {
						patterns = append(patterns, pattern)
					}
				}
			}

			if dateMatches := dateRegex.FindAllString(cleanNote, -1); len(dateMatches) > 0 {
				for _, dateMatch := range dateMatches {
					suggestions := []string{
						fmt.Sprintf("Marquer '%s' comme repère temporel", dateMatch),
					}

					// Essayer d'extraire le sujet
					if subject := extractSubjectFromNote(cleanNote); subject != "" {
						suggestions = append(suggestions, fmt.Sprintf("%s -> a lieu le -> %s", subject, dateMatch))
					}

					pattern := TemporalPattern{
						Pattern:     "date/moment",
						Occurrences: []string{cleanNote},
						Suggestions: suggestions,
					}

					// Vérifier si un pattern "date/moment" existe déjà
					found := false
					for i, p := range patterns {
						if p.Pattern == "date/moment" {
							if !contains(patterns[i].Occurrences, cleanNote) {
								patterns[i].Occurrences = append(patterns[i].Occurrences, cleanNote)
							}
							for _, sug := range suggestions {
								if !contains(patterns[i].Suggestions, sug) {
									patterns[i].Suggestions = append(patterns[i].Suggestions, sug)
								}
							}
							found = true
							break
						}
					}
					if !found {
						patterns = append(patterns, pattern)
					}
				}
			}
		}
	}

	// Analyser spécifiquement les relations temporelles déjà existantes
	relationRegex := regexp.MustCompile(`^(.*) -> (.*) -> (.*)$`)
	for _, notes := range n4lNotes {
		for _, note := range notes {
			if matches := relationRegex.FindStringSubmatch(note); len(matches) == 4 {
				relation := strings.ToLower(strings.TrimSpace(matches[2]))
				// Vérifier si c'est une relation temporelle
				if strings.Contains(relation, "arrivé") || strings.Contains(relation, "prévu") ||
					strings.Contains(relation, "passé") || strings.Contains(relation, "être lu") {

					source := strings.TrimSpace(matches[1])
					target := strings.TrimSpace(matches[3])

					pattern := TemporalPattern{
						Pattern:     "relation temporelle",
						Occurrences: []string{note},
						Suggestions: []string{
							fmt.Sprintf("Créer une chronologie: %s -> puis -> %s", source, target),
							fmt.Sprintf("Ajouter au contexte 'Chronologie': %s", note),
						},
					}

					// Ajouter ou fusionner le pattern
					found := false
					for i, p := range patterns {
						if p.Pattern == "relation temporelle" {
							if !contains(patterns[i].Occurrences, note) {
								patterns[i].Occurrences = append(patterns[i].Occurrences, note)
							}
							found = true
							break
						}
					}
					if !found {
						patterns = append(patterns, pattern)
					}
				}
			}
		}
	}

	log.Printf("Found %d temporal patterns", len(patterns))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(patterns)
}

func checkSemanticConsistencyHandler(w http.ResponseWriter, r *http.Request) {
	var graphData GraphData
	if err := json.NewDecoder(r.Body).Decode(&graphData); err != nil {
		http.Error(w, "Données invalides", http.StatusBadRequest)
		return
	}

	type Inconsistency struct {
		Type        string   `json:"type"`
		Description string   `json:"description"`
		Nodes       []string `json:"nodes"`
		Severity    string   `json:"severity"` // "error", "warning", "info"
		Suggestion  string   `json:"suggestion"`
	}

	var inconsistencies []Inconsistency

	// 1. Détecter les contradictions temporelles
	temporalEdges := make(map[string][]Edge)
	for _, edge := range graphData.Edges {
		if edge.Type == "relation" {
			lowerLabel := strings.ToLower(edge.Label)
			if strings.Contains(lowerLabel, "précède") || strings.Contains(lowerLabel, "avant") ||
				strings.Contains(lowerLabel, "puis") || strings.Contains(lowerLabel, "ensuite") {
				temporalEdges[edge.From] = append(temporalEdges[edge.From], edge)
			}
		}
	}

	// Vérifier les cycles temporels
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	var detectCycle func(node string, path []string) []string
	detectCycle = func(node string, path []string) []string {
		visited[node] = true
		recStack[node] = true
		path = append(path, node)

		for _, edge := range temporalEdges[node] {
			if !visited[edge.To] {
				if cycle := detectCycle(edge.To, path); cycle != nil {
					return cycle
				}
			} else if recStack[edge.To] {
				// Cycle détecté
				cycleStart := -1
				for i, n := range path {
					if n == edge.To {
						cycleStart = i
						break
					}
				}
				if cycleStart >= 0 {
					return append(path[cycleStart:], edge.To)
				}
			}
		}

		recStack[node] = false
		return nil
	}

	for node := range temporalEdges {
		if !visited[node] {
			if cycle := detectCycle(node, []string{}); cycle != nil {
				inconsistencies = append(inconsistencies, Inconsistency{
					Type:        "temporal_cycle",
					Description: fmt.Sprintf("Boucle temporelle détectée : %s", strings.Join(cycle, " → ")),
					Nodes:       cycle,
					Severity:    "error",
					Suggestion:  "Vérifiez l'ordre chronologique des événements. Un événement ne peut pas précéder et suivre le même élément.",
				})
				break
			}
		}
	}

	// 2. Détecter les relations contradictoires
	relationMap := make(map[string]map[string][]string)
	for _, edge := range graphData.Edges {
		if edge.Type == "relation" {
			if relationMap[edge.From] == nil {
				relationMap[edge.From] = make(map[string][]string)
			}
			relationMap[edge.From][edge.To] = append(relationMap[edge.From][edge.To], edge.Label)
		}
	}

	// Vérifier les relations contradictoires
	contradictoryPairs := map[string]string{
		"cause":     "empêche",
		"contient":  "exclut",
		"précède":   "suit",
		"identique": "différent",
		"ami":       "ennemi",
	}

	for from, targets := range relationMap {
		for to, labels := range targets {
			for _, label1 := range labels {
				for _, label2 := range labels {
					if label1 != label2 {
						l1Lower := strings.ToLower(label1)
						l2Lower := strings.ToLower(label2)

						for word1, word2 := range contradictoryPairs {
							if (strings.Contains(l1Lower, word1) && strings.Contains(l2Lower, word2)) ||
								(strings.Contains(l1Lower, word2) && strings.Contains(l2Lower, word1)) {
								inconsistencies = append(inconsistencies, Inconsistency{
									Type:        "contradictory_relations",
									Description: fmt.Sprintf("%s a des relations contradictoires avec %s : '%s' et '%s'", from, to, label1, label2),
									Nodes:       []string{from, to},
									Severity:    "warning",
									Suggestion:  "Clarifiez la nature de la relation entre ces éléments.",
								})
							}
						}
					}
				}
			}
		}
	}

	// 3. Détecter les équivalences incohérentes
	equivalenceGroups := make(map[string][]string)
	for _, edge := range graphData.Edges {
		if edge.Type == "equivalence" {
			found := false
			for key, group := range equivalenceGroups {
				if contains(group, edge.From) || contains(group, edge.To) {
					if !contains(group, edge.From) {
						equivalenceGroups[key] = append(group, edge.From)
					}
					if !contains(group, edge.To) {
						equivalenceGroups[key] = append(group, edge.To)
					}
					found = true
					break
				}
			}
			if !found {
				equivalenceGroups[edge.From] = []string{edge.From, edge.To}
			}
		}
	}

	// Vérifier que les éléments équivalents ont des relations similaires
	for _, group := range equivalenceGroups {
		if len(group) > 1 {
			relationsPerNode := make(map[string]map[string]bool)
			for _, node := range group {
				relationsPerNode[node] = make(map[string]bool)
				for _, edge := range graphData.Edges {
					if edge.From == node && edge.Type == "relation" {
						relationsPerNode[node][edge.To+":"+edge.Label] = true
					}
				}
			}

			// Comparer les relations
			for i := 0; i < len(group)-1; i++ {
				for j := i + 1; j < len(group); j++ {
					node1, node2 := group[i], group[j]
					diff := 0
					for rel := range relationsPerNode[node1] {
						if !relationsPerNode[node2][rel] {
							diff++
						}
					}
					for rel := range relationsPerNode[node2] {
						if !relationsPerNode[node1][rel] {
							diff++
						}
					}

					if diff > 2 {
						inconsistencies = append(inconsistencies, Inconsistency{
							Type:        "inconsistent_equivalence",
							Description: fmt.Sprintf("%s et %s sont marqués comme équivalents mais ont des relations très différentes", node1, node2),
							Nodes:       []string{node1, node2},
							Severity:    "info",
							Suggestion:  "Vérifiez si ces éléments sont vraiment équivalents ou s'il s'agit d'une relation différente.",
						})
					}
				}
			}
		}
	}

	// 4. Détecter les nœuds orphelins suspects
	connectedNodes := make(map[string]bool)
	for _, edge := range graphData.Edges {
		connectedNodes[edge.From] = true
		connectedNodes[edge.To] = true
	}

	for _, node := range graphData.Nodes {
		if !connectedNodes[node.ID] {
			// Vérifier si c'est un nom propre ou un concept important
			if isLikelyImportant(node.Label) {
				inconsistencies = append(inconsistencies, Inconsistency{
					Type:        "orphan_node",
					Description: fmt.Sprintf("'%s' semble important mais n'a aucune connexion", node.Label),
					Nodes:       []string{node.ID},
					Severity:    "info",
					Suggestion:  "Considérez ajouter des relations pour connecter cet élément au reste du graphe.",
				})
			}
		}
	}

	// 5. Détecter les groupes incohérents
	for _, edge := range graphData.Edges {
		if edge.Type == "group" {
			// Vérifier si les membres du groupe ont des relations entre eux
			groupMembers := []string{}
			for _, e := range graphData.Edges {
				if e.Type == "group" && e.From == edge.From {
					groupMembers = append(groupMembers, e.To)
				}
			}

			if len(groupMembers) > 2 {
				// Vérifier la cohérence sémantique du groupe
				hasInternalRelations := false
				for _, m1 := range groupMembers {
					for _, m2 := range groupMembers {
						if m1 != m2 {
							for _, e := range graphData.Edges {
								if (e.From == m1 && e.To == m2) || (e.From == m2 && e.To == m1) {
									hasInternalRelations = true
									break
								}
							}
						}
					}
				}

				if !hasInternalRelations && len(groupMembers) > 3 {
					inconsistencies = append(inconsistencies, Inconsistency{
						Type:        "disconnected_group",
						Description: fmt.Sprintf("Le groupe '%s' contient des éléments sans relations entre eux", edge.From),
						Nodes:       append([]string{edge.From}, groupMembers...),
						Severity:    "info",
						Suggestion:  "Les membres d'un groupe devraient avoir des relations ou propriétés communes.",
					})
				}
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(inconsistencies)
}

// Fonction helper pour extraire le sujet principal d'une note
func extractSubjectFromNote(note string) string {
	// Essayer d'extraire le premier élément significatif
	relationRegex := regexp.MustCompile(`^(.*) -> .* -> .*$`)
	if matches := relationRegex.FindStringSubmatch(note); len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	// Pour les autres formats, prendre le premier nom propre ou concept
	words := strings.Fields(note)
	for _, word := range words {
		word = strings.Trim(word, `"'.,;:!?`)
		if len(word) > 2 && unicode.IsUpper(rune(word[0])) {
			return word
		}
	}

	return ""
}

func investigationModeHandler(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Step        string              `json:"step"`
		GraphData   GraphData           `json:"graphData"`
		CurrentData map[string][]string `json:"currentData"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Données invalides", http.StatusBadRequest)
		return
	}

	type InvestigationStep struct {
		Question    string   `json:"question"`
		Suggestions []string `json:"suggestions"`
		ActionType  string   `json:"actionType"` // "subjects", "relations", "groups"
		NextStep    string   `json:"nextStep"`
		Tips        string   `json:"tips"`
	}

	steps := map[string]InvestigationStep{
		"actors": {
			Question:    "Qui sont les acteurs principaux de votre enquête ?",
			Suggestions: []string{"Victime", "Suspect", "Témoin", "Enquêteur", "Expert"},
			ActionType:  "subjects",
			NextStep:    "locations",
			Tips:        "Identifiez toutes les personnes impliquées, même indirectement.",
		},
		"locations": {
			Question:    "Quels sont les lieux importants ?",
			Suggestions: []string{"Scène de crime", "Domicile", "Lieu de travail", "Lieu public"},
			ActionType:  "subjects",
			NextStep:    "timeline",
			Tips:        "Notez tous les endroits mentionnés, ils peuvent révéler des connexions.",
		},
		"timeline": {
			Question:    "Quelle est la chronologie des événements ?",
			Suggestions: []string{"avant -> précède -> après", "pendant -> simultané -> pendant", "cause -> entraîne -> conséquence"},
			ActionType:  "relations",
			NextStep:    "motives",
			Tips:        "Établissez l'ordre temporel pour comprendre la séquence causale.",
		},
		"motives": {
			Question:    "Quels sont les mobiles identifiés ?",
			Suggestions: []string{"Argent", "Vengeance", "Jalousie", "Protection", "Secret"},
			ActionType:  "subjects",
			NextStep:    "evidence",
			Tips:        "Un mobile fort peut révéler le coupable.",
		},
		"evidence": {
			Question:    "Quelles preuves sont disponibles ?",
			Suggestions: []string{"Preuve physique", "Témoignage", "Document", "Enregistrement", "Trace numérique"},
			ActionType:  "subjects",
			NextStep:    "connections",
			Tips:        "Cataloguez toutes les preuves, même celles qui semblent insignifiantes.",
		},
		"connections": {
			Question:    "Comment relier les éléments entre eux ?",
			Suggestions: []string{"possède", "a rencontré", "connaît", "travaille avec", "est lié à"},
			ActionType:  "relations",
			NextStep:    "groups",
			Tips:        "Cherchez les patterns et les connexions cachées.",
		},
		"groups": {
			Question:    "Comment regrouper les éléments similaires ?",
			Suggestions: []string{"Suspects => {}", "Preuves => {}", "Lieux visités => {}", "Alibis => {}"},
			ActionType:  "groups",
			NextStep:    "complete",
			Tips:        "Organisez vos découvertes en catégories logiques.",
		},
	}

	// Analyser l'état actuel du graphe pour suggestions contextuelles
	var contextualSuggestions []string

	if request.Step == "actors" {
		// Détecter les noms propres dans les données existantes
		for _, notes := range request.CurrentData {
			for _, note := range notes {
				words := strings.Fields(note)
				for _, word := range words {
					if len(word) > 2 && unicode.IsUpper(rune(word[0])) {
						if !contains(contextualSuggestions, word) {
							contextualSuggestions = append(contextualSuggestions, word)
						}
					}
				}
			}
		}
	} else if request.Step == "connections" {
		// Suggérer des connexions basées sur les nœuds orphelins
		orphans := findOrphanNodes(request.GraphData)
		for _, orphan := range orphans {
			suggestion := fmt.Sprintf("Connecter '%s' au graphe", orphan)
			contextualSuggestions = append(contextualSuggestions, suggestion)
		}
	}

	step, exists := steps[request.Step]
	if !exists {
		step = steps["actors"] // Commencer par le début
	}

	// Ajouter les suggestions contextuelles
	if len(contextualSuggestions) > 0 {
		step.Suggestions = append(contextualSuggestions, step.Suggestions...)
	}

	// Analyser la progression
	progress := analyzeInvestigationProgress(request.GraphData)
	if progress.ActorsCount < 2 && request.Step != "actors" {
		step.Tips = "⚠️ Conseil: Ajoutez plus d'acteurs pour enrichir votre enquête. " + step.Tips
	}
	if progress.IsolatedNodes > 3 {
		step.Tips = "⚠️ Attention: Vous avez plusieurs éléments isolés. Pensez à les connecter. " + step.Tips
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(step)
}

func getLayeredGraphHandler(w http.ResponseWriter, r *http.Request) {
	var graphData GraphData
	if err := json.NewDecoder(r.Body).Decode(&graphData); err != nil {
		http.Error(w, "Données invalides", http.StatusBadRequest)
		return
	}

	type LayeredNode struct {
		ID      string  `json:"id"`
		Label   string  `json:"label"`
		Context string  `json:"context"`
		Layer   string  `json:"layer"`
		X       float64 `json:"x"`
		Y       float64 `json:"y"`
		Color   string  `json:"color"`
		Shape   string  `json:"shape"`
		Size    int     `json:"size"`
	}

	type LayeredGraph struct {
		Nodes  []LayeredNode `json:"nodes"`
		Edges  []Edge        `json:"edges"`
		Layers map[string]struct {
			Y     int    `json:"y"`
			Color string `json:"color"`
			Label string `json:"label"`
		} `json:"layers"`
	}

	// Définir les couches et leurs propriétés
	layers := map[string]struct {
		Y     int    `json:"y"`
		Color string `json:"color"`
		Label string `json:"label"`
	}{
		"actors": {
			Y:     0,
			Color: "#3b82f6",
			Label: "Acteurs",
		},
		"locations": {
			Y:     200,
			Color: "#10b981",
			Label: "Lieux",
		},
		"events": {
			Y:     400,
			Color: "#f59e0b",
			Label: "Événements",
		},
		"evidence": {
			Y:     600,
			Color: "#ef4444",
			Label: "Preuves",
		},
		"concepts": {
			Y:     800,
			Color: "#8b5cf6",
			Label: "Concepts",
		},
	}

	// Classifier les nœuds par couche
	layeredNodes := []LayeredNode{}
	nodeLayerCount := make(map[string]int)

	for _, node := range graphData.Nodes {
		layer := classifyNodeLayer(node.Label, node.Context)
		nodeLayerCount[layer]++

		// Calculer la position X pour répartir les nœuds horizontalement
		xOffset := nodeLayerCount[layer] * 150

		layeredNode := LayeredNode{
			ID:      node.ID,
			Label:   node.Label,
			Context: node.Context,
			Layer:   layer,
			X:       float64(xOffset),
			Y:       float64(layers[layer].Y),
			Color:   layers[layer].Color,
			Shape:   getNodeShape(layer),
			Size:    calculateNodeSize(node, graphData.Edges),
		}

		layeredNodes = append(layeredNodes, layeredNode)
	}

	// Ajuster les positions X pour centrer chaque couche
	for layer := range layers {
		count := nodeLayerCount[layer]
		if count > 0 {
			totalWidth := count * 150
			startX := -totalWidth / 2

			currentCount := 0
			for i := range layeredNodes {
				if layeredNodes[i].Layer == layer {
					layeredNodes[i].X = float64(startX + currentCount*150)
					currentCount++
				}
			}
		}
	}

	result := LayeredGraph{
		Nodes:  layeredNodes,
		Edges:  graphData.Edges,
		Layers: layers,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func classifyNodeLayer(label, context string) string {
	lowerLabel := strings.ToLower(label)
	lowerContext := strings.ToLower(context)

	// Détecter les acteurs (personnes)
	if unicode.IsUpper(rune(label[0])) && !strings.Contains(lowerLabel, " ") {
		return "actors"
	}
	if strings.Contains(lowerContext, "personnage") || strings.Contains(lowerContext, "suspect") ||
		strings.Contains(lowerLabel, "victime") || strings.Contains(lowerLabel, "témoin") ||
		strings.Contains(lowerLabel, "enquêteur") || strings.Contains(lowerLabel, "detective") {
		return "actors"
	}

	// Détecter les lieux
	if strings.Contains(lowerContext, "lieu") || strings.Contains(lowerLabel, "scène") ||
		strings.Contains(lowerLabel, "maison") || strings.Contains(lowerLabel, "bureau") ||
		strings.Contains(lowerLabel, "bibliothèque") || strings.Contains(lowerLabel, "manoir") ||
		strings.Contains(lowerLabel, "jardin") || strings.Contains(lowerLabel, "rue") {
		return "locations"
	}

	// Détecter les événements
	if strings.Contains(lowerContext, "chronologie") || strings.Contains(lowerContext, "timeline") ||
		strings.Contains(lowerLabel, "arrivé") || strings.Contains(lowerLabel, "découvert") ||
		strings.Contains(lowerLabel, "rencontré") || strings.Contains(lowerLabel, "heure") ||
		strings.Contains(lowerLabel, "moment") || strings.Contains(lowerLabel, "avant") ||
		strings.Contains(lowerLabel, "après") {
		return "events"
	}

	// Détecter les preuves
	if strings.Contains(lowerContext, "preuve") || strings.Contains(lowerContext, "indice") ||
		strings.Contains(lowerLabel, "document") || strings.Contains(lowerLabel, "trace") ||
		strings.Contains(lowerLabel, "empreinte") || strings.Contains(lowerLabel, "tasse") ||
		strings.Contains(lowerLabel, "livre") || strings.Contains(lowerLabel, "lettre") {
		return "evidence"
	}

	// Par défaut, concepts
	return "concepts"
}

func getNodeShape(layer string) string {
	switch layer {
	case "actors":
		return "circle"
	case "locations":
		return "square"
	case "events":
		return "diamond"
	case "evidence":
		return "triangle"
	default:
		return "box"
	}
}

func calculateNodeSize(node Node, edges []Edge) int {
	// Calculer la taille basée sur le nombre de connexions
	connections := 0
	for _, edge := range edges {
		if edge.From == node.ID || edge.To == node.ID {
			connections++
		}
	}

	// Taille de base + bonus pour les connexions
	baseSize := 25
	return baseSize + (connections * 3)
}

func findOrphanNodes(graphData GraphData) []string {
	connected := make(map[string]bool)
	for _, edge := range graphData.Edges {
		connected[edge.From] = true
		connected[edge.To] = true
	}

	var orphans []string
	for _, node := range graphData.Nodes {
		if !connected[node.ID] {
			orphans = append(orphans, node.ID)
		}
	}
	return orphans
}

type InvestigationProgress struct {
	ActorsCount    int
	LocationsCount int
	EvidenceCount  int
	RelationsCount int
	IsolatedNodes  int
}

func analyzeInvestigationProgress(graphData GraphData) InvestigationProgress {
	progress := InvestigationProgress{
		RelationsCount: len(graphData.Edges),
	}

	// Compter les types de nœuds
	for _, node := range graphData.Nodes {
		label := strings.ToLower(node.Label)
		if strings.Contains(label, "suspect") || strings.Contains(label, "victime") ||
			strings.Contains(label, "témoin") || unicode.IsUpper(rune(node.Label[0])) {
			progress.ActorsCount++
		}
		if strings.Contains(label, "lieu") || strings.Contains(label, "scène") ||
			strings.Contains(label, "maison") || strings.Contains(label, "bureau") {
			progress.LocationsCount++
		}
		if strings.Contains(label, "preuve") || strings.Contains(label, "indice") ||
			strings.Contains(label, "document") || strings.Contains(label, "trace") {
			progress.EvidenceCount++
		}
	}

	// Compter les nœuds isolés
	connected := make(map[string]bool)
	for _, edge := range graphData.Edges {
		connected[edge.From] = true
		connected[edge.To] = true
	}
	for _, node := range graphData.Nodes {
		if !connected[node.ID] {
			progress.IsolatedNodes++
		}
	}

	return progress
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func isLikelyImportant(label string) bool {
	// Détecte si un nœud est probablement important (nom propre, concept clé)
	if len(label) < 3 {
		return false
	}

	// Vérifie si c'est un nom propre (commence par une majuscule)
	if unicode.IsUpper(rune(label[0])) {
		return true
	}

	// Liste de mots-clés importants
	importantKeywords := []string{"principal", "important", "clé", "central", "critique", "essentiel"}
	lowerLabel := strings.ToLower(label)
	for _, keyword := range importantKeywords {
		if strings.Contains(lowerLabel, keyword) {
			return true
		}
	}

	return false
}

// Amélioration de la fonction analyzeTemporalContext
func analyzeTemporalContext(text, marker, relation string) []string {
	var suggestions []string

	// Nettoyer le texte
	cleanText := strings.TrimSpace(text)

	// Si c'est déjà une relation, proposer une version temporelle
	relationRegex := regexp.MustCompile(`^(.*) -> (.*) -> (.*)$`)
	if matches := relationRegex.FindStringSubmatch(cleanText); len(matches) == 4 {
		source := strings.TrimSpace(matches[1])
		target := strings.TrimSpace(matches[3])

		// Proposer une relation temporelle basée sur le marqueur trouvé
		suggestions = append(suggestions, fmt.Sprintf("%s -> %s -> %s", source, relation, target))
		suggestions = append(suggestions, fmt.Sprintf("Ajouter au contexte 'Chronologie': %s", cleanText))
		return suggestions
	}

	// Diviser le texte autour du marqueur pour les phrases simples
	lowerText := strings.ToLower(text)
	markerIndex := strings.Index(lowerText, marker)

	if markerIndex > 0 {
		beforeMarker := text[:markerIndex]
		afterMarker := text[markerIndex+len(marker):]

		// Extraire les mots significatifs avant et après
		beforeWords := extractSignificantWords(beforeMarker)
		afterWords := extractSignificantWords(afterMarker)

		if len(beforeWords) > 0 && len(afterWords) > 0 {
			// Prendre le dernier mot significatif avant et le premier après
			subject := beforeWords[len(beforeWords)-1]
			object := afterWords[0]

			suggestion := fmt.Sprintf("%s -> %s -> %s", subject, relation, object)
			suggestions = append(suggestions, suggestion)
		}
	}

	// Suggérer une annotation temporelle générale
	suggestions = append(suggestions, fmt.Sprintf("Annoter comme événement temporel avec '%s'", marker))
	suggestions = append(suggestions, fmt.Sprintf("Ajouter au contexte 'Chronologie': %s", text))

	return suggestions
}

// Fonction pour extraire les mots significatifs
func extractSignificantWords(text string) []string {
	var significant []string
	words := strings.Fields(text)

	for _, word := range words {
		// Nettoyer la ponctuation
		word = strings.Trim(word, ".,;:!?()[]{}\"'")

		// Ignorer les mots trop courts ou les articles
		if len(word) < 3 {
			continue
		}

		// Ignorer les articles et prépositions communes
		commonWords := []string{"les", "une", "des", "dans", "sur", "avec", "pour", "par"}
		isCommon := false
		for _, common := range commonWords {
			if strings.ToLower(word) == common {
				isCommon = true
				break
			}
		}

		if !isCommon {
			significant = append(significant, word)
		}
	}

	return significant
}

func generateInvestigationQuestionsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("generateInvestigationQuestionsHandler called")

	var graphData GraphData
	if err := json.NewDecoder(r.Body).Decode(&graphData); err != nil {
		http.Error(w, "Données invalides", http.StatusBadRequest)
		return
	}

	var questions []InvestigationQuestion

	// 1. Analyser les nœuds orphelins
	orphans := findOrphanNodes(graphData)
	for _, orphan := range orphans {
		// Analyser l'importance basée sur les connexions sémantiques, pas sur des noms prédéfinis
		importance := calculateNodeImportance(orphan, graphData)
		if importance > 0.5 {
			questions = append(questions, InvestigationQuestion{
				Question: fmt.Sprintf("Comment '%s' est-il lié aux autres éléments ?", orphan),
				Type:     "orphan",
				Priority: getPriorityFromImportance(importance),
				Context:  "Connexions manquantes",
				Nodes:    []string{orphan},
				Hint:     "Cet élément semble isolé. Cherchez des relations possibles.",
			})
		}
	}

	// 2. Détecter les patterns incomplets basés sur la structure du graphe
	patterns := analyzeGraphPatterns(graphData)
	for _, pattern := range patterns {
		questions = append(questions, pattern)
	}

	// 3. Analyser les clusters déconnectés
	clusters := findDisconnectedClusters(graphData)
	for _, cluster := range clusters {
		if len(cluster) > 1 {
			questions = append(questions, InvestigationQuestion{
				Question: fmt.Sprintf("Quelle connexion existe entre ces groupes : %s ?", strings.Join(cluster[:2], " et ")),
				Type:     "missing_link",
				Priority: "medium",
				Context:  "Groupes isolés",
				Nodes:    cluster,
				Hint:     "Ces éléments forment des groupes séparés qui pourraient être liés.",
			})
		}
	}

	// 4. Détecter les nœuds centraux sans connexions importantes
	centralNodes := findCentralNodes(graphData)
	for _, node := range centralNodes {
		missingConnections := analyzeMissingConnections(node, graphData)
		for _, missing := range missingConnections {
			questions = append(questions, missing)
		}
	}

	// Trier par priorité calculée dynamiquement
	sort.Slice(questions, func(i, j int) bool {
		priorityOrder := map[string]int{"high": 0, "medium": 1, "low": 2}
		return priorityOrder[questions[i].Priority] < priorityOrder[questions[j].Priority]
	})

	if len(questions) > 10 {
		questions = questions[:10]
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(questions)
}

func calculateNodeImportance(nodeID string, graphData GraphData) float64 {
	var node Node
	for _, n := range graphData.Nodes {
		if n.ID == nodeID {
			node = n
			break
		}
	}

	importance := 0.3 // Base

	// Majuscule initiale suggère un nom propre ou concept important
	if len(node.Label) > 0 && unicode.IsUpper(rune(node.Label[0])) {
		importance += 0.3
	}

	// Les labels plus descriptifs sont souvent plus importants
	if len(node.Label) > 10 {
		importance += 0.2
	}

	// Contexte non-général suggère une importance spécifique
	if node.Context != "" && node.Context != "general" {
		importance += 0.2
	}

	return importance
}

func getPriorityFromImportance(importance float64) string {
	if importance > 0.7 {
		return "high"
	} else if importance > 0.5 {
		return "medium"
	}
	return "low"
}

func analyzeGraphPatterns(graphData GraphData) []InvestigationQuestion {
	var questions []InvestigationQuestion

	// Analyser les patterns basés sur la structure actuelle
	nodeTypes := classifyNodesByConnections(graphData)

	// Pour chaque type détecté, vérifier s'il y a des patterns incomplets
	for _, nodes := range nodeTypes { // Supprimé nodeType qui n'était pas utilisé
		if len(nodes) > 0 {
			avgConnections := calculateAverageConnections(nodes, graphData)

			for _, node := range nodes {
				connections := countNodeConnections(node, graphData)
				if float64(connections) < avgConnections*0.5 {
					questions = append(questions, InvestigationQuestion{
						Question: fmt.Sprintf("Pourquoi '%s' a-t-il moins de connexions que les autres éléments similaires ?", node),
						Type:     "pattern",
						Priority: "medium",
						Context:  "Pattern incomplet",
						Nodes:    []string{node},
						Hint:     fmt.Sprintf("Cet élément a %d connexions alors que la moyenne est %.1f", connections, avgConnections),
					})
				}
			}
		}
	}

	return questions
}

func classifyNodesByConnections(graphData GraphData) map[string][]string {
	// Classifier les nœuds par leurs patterns de connexion, pas par des types prédéfinis
	classified := make(map[string][]string)

	for _, node := range graphData.Nodes {
		pattern := getConnectionPattern(node.ID, graphData)
		classified[pattern] = append(classified[pattern], node.ID)
	}

	return classified
}

func getConnectionPattern(nodeID string, graphData GraphData) string {
	inCount := 0
	outCount := 0

	for _, edge := range graphData.Edges {
		if edge.To == nodeID {
			inCount++
		}
		if edge.From == nodeID {
			outCount++
		}
	}

	// Classifier par pattern de connexion
	if inCount > outCount*2 {
		return "receiver" // Reçoit beaucoup plus qu'il n'émet
	} else if outCount > inCount*2 {
		return "emitter" // Émet beaucoup plus qu'il ne reçoit
	} else if inCount+outCount > 5 {
		return "hub" // Nœud très connecté
	} else if inCount+outCount == 0 {
		return "isolated" // Nœud isolé
	}
	return "standard" // Connexions équilibrées
}

func findDisconnectedClusters(graphData GraphData) [][]string {
	// Utiliser un algorithme de parcours pour trouver les composantes connexes
	visited := make(map[string]bool)
	var clusters [][]string

	for _, node := range graphData.Nodes {
		if !visited[node.ID] {
			cluster := []string{}
			dfs(node.ID, graphData, visited, &cluster)
			if len(cluster) > 0 {
				clusters = append(clusters, cluster)
			}
		}
	}

	return clusters
}

func dfs(nodeID string, graphData GraphData, visited map[string]bool, cluster *[]string) {
	visited[nodeID] = true
	*cluster = append(*cluster, nodeID)

	for _, edge := range graphData.Edges {
		if edge.From == nodeID && !visited[edge.To] {
			dfs(edge.To, graphData, visited, cluster)
		}
		if edge.To == nodeID && !visited[edge.From] {
			dfs(edge.From, graphData, visited, cluster)
		}
	}
}

func findCentralNodes(graphData GraphData) []string {
	// Trouver les nœuds avec le plus de connexions (hubs)
	connectionCount := make(map[string]int)

	for _, edge := range graphData.Edges {
		connectionCount[edge.From]++
		connectionCount[edge.To]++
	}

	var central []string
	avgConnections := 0
	if len(connectionCount) > 0 {
		total := 0
		for _, count := range connectionCount {
			total += count
		}
		avgConnections = total / len(connectionCount)
	}

	for node, count := range connectionCount {
		if count > avgConnections*2 {
			central = append(central, node)
		}
	}

	return central
}

func analyzeMissingConnections(nodeID string, graphData GraphData) []InvestigationQuestion {
	var questions []InvestigationQuestion

	// Analyser les connexions transitives potentielles
	// Si A->B et B->C mais pas A->C, suggérer une connexion possible

	connected := make(map[string]bool)
	secondDegree := make(map[string]int)

	// Connexions directes
	for _, edge := range graphData.Edges {
		if edge.From == nodeID {
			connected[edge.To] = true
		}
		if edge.To == nodeID {
			connected[edge.From] = true
		}
	}

	// Connexions de second degré
	for connectedNode := range connected {
		for _, edge := range graphData.Edges {
			if edge.From == connectedNode && !connected[edge.To] && edge.To != nodeID {
				secondDegree[edge.To]++
			}
			if edge.To == connectedNode && !connected[edge.From] && edge.From != nodeID {
				secondDegree[edge.From]++
			}
		}
	}

	// Suggérer des connexions pour les nœuds souvent atteints au second degré
	for node, count := range secondDegree {
		if count >= 2 {
			questions = append(questions, InvestigationQuestion{
				Question: fmt.Sprintf("Y a-t-il une relation directe entre '%s' et '%s' ?", nodeID, node),
				Type:     "missing_link",
				Priority: "low",
				Context:  "Connexion transitive",
				Nodes:    []string{nodeID, node},
				Hint:     fmt.Sprintf("Ces éléments sont connectés via %d intermédiaires", count),
			})
		}
	}

	return questions
}

func calculateAverageConnections(nodes []string, graphData GraphData) float64 {
	if len(nodes) == 0 {
		return 0
	}

	total := 0
	for _, node := range nodes {
		total += countNodeConnections(node, graphData)
	}

	return float64(total) / float64(len(nodes))
}

func countNodeConnections(nodeID string, graphData GraphData) int {
	count := 0
	for _, edge := range graphData.Edges {
		if edge.From == nodeID || edge.To == nodeID {
			count++
		}
	}
	return count
}

func main() {
	http.HandleFunc("/api/extract-concepts", extractConceptsHandler)
	http.HandleFunc("/api/auto-extract-subjects", autoExtractSubjectsHandler)
	http.HandleFunc("/api/parse-n4l", parseN4LHandler)
	http.HandleFunc("/api/graph-data", parseN4LToGraph)
	http.HandleFunc("/api/analyze-graph", analyzeGraphHandler)
	http.HandleFunc("/api/timeline-data", timelineHandler)
	http.HandleFunc("/api/find-all-paths", findAllPathsHandler)
	http.HandleFunc("/api/analyze-path", analyzePathHandler)
	http.HandleFunc("/api/detect-temporal-patterns", detectTemporalPatternsHandler)
	http.HandleFunc("/api/check-consistency", checkSemanticConsistencyHandler)
	http.HandleFunc("/api/investigation-mode", investigationModeHandler)
	http.HandleFunc("/api/layered-graph", getLayeredGraphHandler)
	http.HandleFunc("/api/generate-questions", generateInvestigationQuestionsHandler)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "n4l_editor.html")
	})

	fmt.Println("Serveur démarré. Accédez à l'application sur http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
