package services

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"forensicinvestigator/internal/models"
)

// OllamaService gère les interactions avec le LLM (vLLM compatible OpenAI)
type OllamaService struct {
	baseURL       string
	model         string
	configService *ConfigService
}

// NewOllamaService crée une nouvelle instance du service LLM (vLLM)
func NewOllamaService(baseURL, model string) *OllamaService {
	return &OllamaService{
		baseURL:       baseURL,
		model:         model,
		configService: NewConfigService(),
	}
}

// SetConfigService permet d'injecter un service de configuration externe
func (s *OllamaService) SetConfigService(cs *ConfigService) {
	s.configService = cs
}

// GetConfigService retourne le service de configuration
func (s *OllamaService) GetConfigService() *ConfigService {
	return s.configService
}

// getPromptConfig retourne la configuration d'un prompt ou une valeur par défaut
func (s *OllamaService) getPromptConfig(name string) *PromptConfig {
	if s.configService != nil {
		if pc := s.configService.GetPrompt(name); pc != nil {
			return pc
		}
	}
	return nil
}

// getLanguageInstruction retourne l'instruction de langue depuis la config
func (s *OllamaService) getLanguageInstruction() string {
	if s.configService != nil {
		return s.configService.GetLanguageInstruction()
	}
	return "IMPORTANT: Tu DOIS répondre UNIQUEMENT en FRANÇAIS.\n\n"
}

// ChatMessage représente un message dans le format chat
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// VLLMChatRequest représente une requête chat à vLLM (format OpenAI)
type VLLMChatRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Temperature float64       `json:"temperature,omitempty"`
	Stream      bool          `json:"stream"`
}

// VLLMChatResponse représente une réponse chat de vLLM (format OpenAI)
type VLLMChatResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		Index        int    `json:"index"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
}

// VLLMChatStreamResponse représente une réponse streaming chat de vLLM
type VLLMChatStreamResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Choices []struct {
		Delta struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"delta"`
		Index        int    `json:"index"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
}

// Generate génère une réponse à partir d'un prompt (vLLM Chat API)
func (s *OllamaService) Generate(prompt string) (string, error) {
	reqBody := VLLMChatRequest{
		Model: s.model,
		Messages: []ChatMessage{
			{Role: "user", Content: prompt},
		},
		MaxTokens:   8192,
		Temperature: 0.7,
		Stream:      false,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("erreur marshalling request: %w", err)
	}

	resp, err := http.Post(s.baseURL+"/v1/chat/completions", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("erreur appel vLLM: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("erreur lecture réponse: %w", err)
	}

	var vllmResp VLLMChatResponse
	if err := json.Unmarshal(body, &vllmResp); err != nil {
		return "", fmt.Errorf("erreur parsing réponse: %w", err)
	}

	if len(vllmResp.Choices) > 0 {
		return vllmResp.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("pas de réponse de vLLM")
}

// BuildAnalyzeCasePrompt construit le prompt pour l'analyse d'affaire
func (s *OllamaService) BuildAnalyzeCasePrompt(caseData models.Case) string {
	var sb strings.Builder

	// Utiliser la configuration si disponible
	pc := s.getPromptConfig("analyze_case")
	if pc != nil {
		sb.WriteString(s.getLanguageInstruction())
		sb.WriteString(pc.System + "\n")
		sb.WriteString(pc.Instruction + "\n\n")
	} else {
		sb.WriteString("IMPORTANT: Tu DOIS répondre UNIQUEMENT en FRANÇAIS.\n\n")
		sb.WriteString("Tu es un assistant d'enquête criminalistique expert.\n")
		sb.WriteString("Analyse les informations suivantes sur cette affaire:\n\n")
	}

	sb.WriteString(fmt.Sprintf("**Affaire**: %s\n", caseData.Name))
	sb.WriteString(fmt.Sprintf("**Type**: %s\n", caseData.Type))
	sb.WriteString(fmt.Sprintf("**Description**: %s\n\n", caseData.Description))

	if len(caseData.Entities) > 0 {
		sb.WriteString("**Entités impliquées**:\n")
		for _, e := range caseData.Entities {
			sb.WriteString(fmt.Sprintf("- %s (%s, %s): %s\n", e.Name, e.Type, e.Role, e.Description))
		}
		sb.WriteString("\n")
	}

	if len(caseData.Evidence) > 0 {
		sb.WriteString("**Preuves collectées**:\n")
		for _, ev := range caseData.Evidence {
			sb.WriteString(fmt.Sprintf("- %s (%s): %s [Fiabilité: %d/10]\n", ev.Name, ev.Type, ev.Description, ev.Reliability))
		}
		sb.WriteString("\n")
	}

	if len(caseData.Timeline) > 0 {
		sb.WriteString("**Chronologie des événements**:\n")
		for _, t := range caseData.Timeline {
			sb.WriteString(fmt.Sprintf("- %s: %s - %s\n", t.Timestamp.Format("02/01/2006 15:04"), t.Title, t.Description))
		}
		sb.WriteString("\n")
	}

	// Utiliser le format de sortie de la config ou le défaut
	if pc != nil && pc.OutputFormat != "" {
		sb.WriteString("\n" + pc.OutputFormat)
	} else {
		sb.WriteString("\nRédige un résumé structuré identifiant:\n")
		sb.WriteString("1. Les points clés de l'affaire\n")
		sb.WriteString("2. Les suspects potentiels et leurs mobiles\n")
		sb.WriteString("3. Les pistes à explorer prioritairement\n")
		sb.WriteString("4. Les incohérences ou contradictions éventuelles\n")
		sb.WriteString("5. Les questions d'investigation restantes\n")
		sb.WriteString("\nUtilise un format markdown structuré.")
	}

	return sb.String()
}

// AnalyzeCase analyse une affaire et génère un résumé
func (s *OllamaService) AnalyzeCase(caseData models.Case) (string, error) {
	return s.Generate(s.BuildAnalyzeCasePrompt(caseData))
}

// AnalyzeCaseStream analyse une affaire en streaming
func (s *OllamaService) AnalyzeCaseStream(caseData models.Case, callback StreamCallback) error {
	return s.GenerateStream(s.BuildAnalyzeCasePrompt(caseData), callback)
}

// GenerateHypotheses génère des hypothèses basées sur les données
func (s *OllamaService) GenerateHypotheses(graphData models.GraphData) ([]models.Hypothesis, error) {
	// Créer un map pour résoudre les IDs en noms
	nodeNames := make(map[string]string)
	for _, node := range graphData.Nodes {
		nodeNames[node.ID] = node.Label
	}

	// Fonction pour résoudre un ID en nom
	resolveName := func(id string) string {
		if name, ok := nodeNames[id]; ok {
			return name
		}
		return id
	}

	var sb strings.Builder

	// Utiliser la configuration si disponible
	pc := s.getPromptConfig("generate_hypotheses")
	if pc != nil {
		sb.WriteString(s.getLanguageInstruction())
		sb.WriteString(pc.System + "\n")
		sb.WriteString(pc.Instruction + "\n\n")
	} else {
		sb.WriteString("IMPORTANT: Tu DOIS répondre UNIQUEMENT en FRANÇAIS.\n\n")
		sb.WriteString("Tu es un analyste criminalistique expert.\n")
		sb.WriteString("En te basant sur le graphe de connaissances suivant, génère des hypothèses d'investigation.\n\n")
	}

	sb.WriteString("**Relations connues**:\n")
	for _, edge := range graphData.Edges {
		fromName := resolveName(edge.From)
		toName := resolveName(edge.To)
		sb.WriteString(fmt.Sprintf("- %s %s %s\n", fromName, edge.Label, toName))
	}

	sb.WriteString("\n**Entités**:\n")
	for _, node := range graphData.Nodes {
		role := ""
		if node.Role != "" {
			role = fmt.Sprintf(" [%s]", node.Role)
		}
		sb.WriteString(fmt.Sprintf("- %s (%s)%s\n", node.Label, node.Type, role))
	}

	// Utiliser le format de sortie de la config ou le défaut
	if pc != nil && pc.OutputFormat != "" {
		sb.WriteString("\n" + pc.OutputFormat)
	} else {
		sb.WriteString("\nGénère 3 à 5 hypothèses d'investigation au format JSON:\n")
		sb.WriteString(`[{"title": "...", "description": "...", "confidence_level": 0-100, "questions": ["..."]}]`)
		sb.WriteString("\nRéponds UNIQUEMENT avec le JSON, sans autre texte.")
	}

	response, err := s.Generate(sb.String())
	if err != nil {
		return nil, err
	}

	// Nettoyer la réponse et extraire le JSON
	response = cleanJSON(response)

	var hypotheses []models.Hypothesis
	if err := json.Unmarshal([]byte(response), &hypotheses); err != nil {
		// Si le parsing échoue, créer une hypothèse par défaut
		return []models.Hypothesis{{
			Title:       "Analyse en cours",
			Description: response,
			Status:      models.HypothesisPending,
			GeneratedBy: "ai",
		}}, nil
	}

	for i := range hypotheses {
		hypotheses[i].Status = models.HypothesisPending
		hypotheses[i].GeneratedBy = "ai"
	}

	return hypotheses, nil
}

// BuildHypothesesPrompt construit le prompt pour la génération d'hypothèses en streaming
func (s *OllamaService) BuildHypothesesPrompt(graphData models.GraphData) string {
	nodeNames := make(map[string]string)
	for _, node := range graphData.Nodes {
		nodeNames[node.ID] = node.Label
	}

	resolveName := func(id string) string {
		if name, ok := nodeNames[id]; ok {
			return name
		}
		return id
	}

	var sb strings.Builder
	sb.WriteString("IMPORTANT: Tu DOIS répondre UNIQUEMENT en FRANÇAIS.\n\n")
	sb.WriteString("Tu es un analyste criminalistique expert spécialisé dans la formulation d'hypothèses d'investigation.\n\n")

	sb.WriteString("## Données de l'affaire\n\n")
	sb.WriteString("### Relations connues:\n")
	for _, edge := range graphData.Edges {
		fromName := resolveName(edge.From)
		toName := resolveName(edge.To)
		sb.WriteString(fmt.Sprintf("- %s **%s** %s\n", fromName, edge.Label, toName))
	}

	sb.WriteString("\n### Entités impliquées:\n")
	for _, node := range graphData.Nodes {
		role := ""
		if node.Role != "" {
			role = fmt.Sprintf(" [Rôle: %s]", node.Role)
		}
		sb.WriteString(fmt.Sprintf("- **%s** (%s)%s\n", node.Label, node.Type, role))
	}

	sb.WriteString("\n## Ta mission\n\n")
	sb.WriteString("Génère 3 à 5 hypothèses d'investigation détaillées. Pour chaque hypothèse:\n\n")
	sb.WriteString("1. **Titre**: Un nom court et descriptif\n")
	sb.WriteString("2. **Description**: Explique l'hypothèse en détail\n")
	sb.WriteString("3. **Niveau de confiance**: Évalue la probabilité (0-100%)\n")
	sb.WriteString("4. **Éléments à l'appui**: Quelles données soutiennent cette hypothèse?\n")
	sb.WriteString("5. **Points faibles**: Quels éléments pourraient la contredire?\n")
	sb.WriteString("6. **Questions à investiguer**: Que faut-il vérifier?\n\n")
	sb.WriteString("Structure ta réponse en markdown avec des sections claires pour chaque hypothèse.")

	return sb.String()
}

// GenerateHypothesesStream génère des hypothèses en streaming
func (s *OllamaService) GenerateHypothesesStream(graphData models.GraphData, callback StreamCallback) error {
	return s.GenerateStream(s.BuildHypothesesPrompt(graphData), callback)
}

// BuildContradictionsPrompt construit le prompt pour la détection de contradictions
func (s *OllamaService) BuildContradictionsPrompt(graphData models.GraphData) string {
	nodeNames := make(map[string]string)
	for _, node := range graphData.Nodes {
		nodeNames[node.ID] = node.Label
	}

	resolveName := func(id string) string {
		if name, ok := nodeNames[id]; ok {
			return name
		}
		return id
	}

	var sb strings.Builder

	// Utiliser la configuration si disponible
	pc := s.getPromptConfig("detect_contradictions")
	if pc != nil {
		sb.WriteString(s.getLanguageInstruction())
		sb.WriteString(pc.System + "\n")
		sb.WriteString(pc.Instruction + "\n\n")
	} else {
		sb.WriteString("IMPORTANT: Tu DOIS répondre UNIQUEMENT en FRANÇAIS.\n\n")
		sb.WriteString("Tu es un analyste expert en détection d'incohérences.\n")
		sb.WriteString("Analyse les relations suivantes et identifie toutes les contradictions potentielles:\n\n")
	}

	for _, edge := range graphData.Edges {
		fromName := resolveName(edge.From)
		toName := resolveName(edge.To)
		sb.WriteString(fmt.Sprintf("- %s %s %s\n", fromName, edge.Label, toName))
	}

	// Utiliser le format de sortie de la config ou le défaut
	if pc != nil && pc.OutputFormat != "" {
		sb.WriteString("\n" + pc.OutputFormat)
	} else {
		sb.WriteString("\nIdentifie:\n")
		sb.WriteString("1. Les contradictions directes (A dit X, B dit non-X)\n")
		sb.WriteString("2. Les incohérences temporelles (A était quelque part quand il prétend être ailleurs)\n")
		sb.WriteString("3. Les alibis impossibles à vérifier ou contradictoires\n")
		sb.WriteString("4. Les relations suspectes ou non expliquées\n")
		sb.WriteString("\nFormat ta réponse en markdown avec des sections claires.")
	}

	return sb.String()
}

// DetectContradictions détecte les contradictions dans les données
func (s *OllamaService) DetectContradictions(graphData models.GraphData) (string, error) {
	return s.Generate(s.BuildContradictionsPrompt(graphData))
}

// DetectContradictionsStream détecte les contradictions en streaming
func (s *OllamaService) DetectContradictionsStream(graphData models.GraphData, callback StreamCallback) error {
	return s.GenerateStream(s.BuildContradictionsPrompt(graphData), callback)
}

// QuestionWithExplanation représente une question avec son explication contextuelle
type QuestionWithExplanation struct {
	Question    string `json:"question"`
	Explanation string `json:"explanation"`
}

// BuildQuestionsPrompt construit le prompt pour la génération de questions (version markdown)
func (s *OllamaService) BuildQuestionsPrompt(graphData models.GraphData, context string) string {
	nodeNames := make(map[string]string)
	for _, node := range graphData.Nodes {
		nodeNames[node.ID] = node.Label
	}

	resolveName := func(id string) string {
		if name, ok := nodeNames[id]; ok {
			return name
		}
		return id
	}

	var sb strings.Builder
	sb.WriteString("IMPORTANT: Tu DOIS répondre UNIQUEMENT en FRANÇAIS.\n\n")
	sb.WriteString("Tu es un enquêteur criminaliste expérimenté utilisant la méthode PEACE.\n")
	sb.WriteString("Génère des questions d'investigation pertinentes basées sur ces informations:\n\n")

	for _, edge := range graphData.Edges {
		fromName := resolveName(edge.From)
		toName := resolveName(edge.To)
		sb.WriteString(fmt.Sprintf("- %s %s %s\n", fromName, edge.Label, toName))
	}

	if context != "" {
		sb.WriteString(fmt.Sprintf("\nContexte supplémentaire: %s\n", context))
	}

	sb.WriteString("\nGénère 5 à 10 questions d'investigation pertinentes.\n")
	sb.WriteString("Pour chaque question:\n")
	sb.WriteString("1. Pose la question clairement\n")
	sb.WriteString("2. Explique pourquoi elle est importante pour l'enquête\n")
	sb.WriteString("3. Indique quels éléments de l'affaire la motivent\n\n")
	sb.WriteString("Format ta réponse en markdown avec des sections numérotées.")

	return sb.String()
}

// GenerateQuestionsStream génère des questions d'investigation en streaming
func (s *OllamaService) GenerateQuestionsStream(graphData models.GraphData, context string, callback StreamCallback) error {
	return s.GenerateStream(s.BuildQuestionsPrompt(graphData, context), callback)
}

// GenerateQuestions génère des questions d'investigation avec explications contextuelles
func (s *OllamaService) GenerateQuestions(graphData models.GraphData, context string) ([]QuestionWithExplanation, error) {
	// Créer un map pour résoudre les IDs en noms
	nodeNames := make(map[string]string)
	for _, node := range graphData.Nodes {
		nodeNames[node.ID] = node.Label
	}

	// Fonction pour résoudre un ID en nom
	resolveName := func(id string) string {
		if name, ok := nodeNames[id]; ok {
			return name
		}
		return id
	}

	var sb strings.Builder
	sb.WriteString("IMPORTANT: Tu DOIS répondre UNIQUEMENT en FRANÇAIS.\n\n")
	sb.WriteString("Tu es un enquêteur criminaliste expérimenté utilisant la méthode PEACE.\n")
	sb.WriteString("Génère des questions d'investigation pertinentes basées sur ces informations:\n\n")

	for _, edge := range graphData.Edges {
		fromName := resolveName(edge.From)
		toName := resolveName(edge.To)
		sb.WriteString(fmt.Sprintf("- %s %s %s\n", fromName, edge.Label, toName))
	}

	if context != "" {
		sb.WriteString(fmt.Sprintf("\nContexte supplémentaire: %s\n", context))
	}

	sb.WriteString("\nGénère 5 à 10 questions d'investigation. Pour chaque question, explique pourquoi elle est pertinente en référençant les données de l'enquête.\n")
	sb.WriteString("Format JSON attendu:\n")
	sb.WriteString(`[{"question": "La question?", "explanation": "Explication basée sur les données de l'enquête..."}]`)
	sb.WriteString("\nRéponds UNIQUEMENT avec le JSON.")

	response, err := s.Generate(sb.String())
	if err != nil {
		return nil, err
	}

	response = cleanJSON(response)

	var questions []QuestionWithExplanation
	if err := json.Unmarshal([]byte(response), &questions); err != nil {
		// Essayer l'ancien format (liste simple de strings) pour rétrocompatibilité
		var simpleQuestions []string
		if err2 := json.Unmarshal([]byte(response), &simpleQuestions); err2 == nil {
			for _, q := range simpleQuestions {
				questions = append(questions, QuestionWithExplanation{Question: q, Explanation: ""})
			}
			return questions, nil
		}
		// Fallback: retourner la réponse comme une seule question
		return []QuestionWithExplanation{{Question: response, Explanation: ""}}, nil
	}

	return questions, nil
}

// AnalyzeHypothesis analyse une hypothèse spécifique en profondeur
func (s *OllamaService) AnalyzeHypothesis(hypothesis models.Hypothesis, caseData *models.Case, graphData *models.GraphData) (string, error) {
	// Créer un map pour résoudre les IDs en noms
	entityNames := make(map[string]string)
	for _, e := range caseData.Entities {
		entityNames[e.ID] = e.Name
	}
	evidenceNames := make(map[string]string)
	for _, ev := range caseData.Evidence {
		evidenceNames[ev.ID] = ev.Name
	}

	var sb strings.Builder
	sb.WriteString("IMPORTANT: Tu DOIS répondre UNIQUEMENT en FRANÇAIS.\n\n")
	sb.WriteString("Tu es un analyste criminalistique expert. Analyse en profondeur cette hypothèse d'investigation.\n\n")

	// Hypothèse
	sb.WriteString(fmt.Sprintf("## Hypothèse: %s\n", hypothesis.Title))
	sb.WriteString(fmt.Sprintf("**Description**: %s\n", hypothesis.Description))
	sb.WriteString(fmt.Sprintf("**Niveau de confiance actuel**: %d%%\n", hypothesis.ConfidenceLevel))
	sb.WriteString(fmt.Sprintf("**Statut**: %s\n\n", hypothesis.Status))

	// Preuves à l'appui
	if len(hypothesis.SupportingEvidence) > 0 {
		sb.WriteString("**Preuves à l'appui**:\n")
		for _, evID := range hypothesis.SupportingEvidence {
			if name, ok := evidenceNames[evID]; ok {
				sb.WriteString(fmt.Sprintf("- %s\n", name))
			}
		}
		sb.WriteString("\n")
	}

	// Preuves contradictoires
	if len(hypothesis.ContradictingEvidence) > 0 {
		sb.WriteString("**Preuves contradictoires**:\n")
		for _, evID := range hypothesis.ContradictingEvidence {
			if name, ok := evidenceNames[evID]; ok {
				sb.WriteString(fmt.Sprintf("- %s\n", name))
			}
		}
		sb.WriteString("\n")
	}

	// Questions existantes
	if len(hypothesis.Questions) > 0 {
		sb.WriteString("**Questions en suspens**:\n")
		for _, q := range hypothesis.Questions {
			sb.WriteString(fmt.Sprintf("- %s\n", q))
		}
		sb.WriteString("\n")
	}

	// Contexte de l'affaire
	sb.WriteString("## Contexte de l'affaire\n")
	sb.WriteString(fmt.Sprintf("**Affaire**: %s\n", caseData.Name))
	sb.WriteString(fmt.Sprintf("**Type**: %s\n\n", caseData.Type))

	// Relations pertinentes
	if graphData != nil && len(graphData.Edges) > 0 {
		sb.WriteString("**Relations clés**:\n")
		nodeNames := make(map[string]string)
		for _, node := range graphData.Nodes {
			nodeNames[node.ID] = node.Label
		}
		for _, edge := range graphData.Edges {
			fromName := nodeNames[edge.From]
			if fromName == "" {
				fromName = edge.From
			}
			toName := nodeNames[edge.To]
			if toName == "" {
				toName = edge.To
			}
			sb.WriteString(fmt.Sprintf("- %s (%s) %s\n", fromName, edge.Label, toName))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("## Analyse demandée\n")
	sb.WriteString("Fournis une analyse structurée avec:\n")
	sb.WriteString("1. **Forces de l'hypothèse**: Quels éléments la soutiennent?\n")
	sb.WriteString("2. **Faiblesses**: Quels sont les points faibles ou manquants?\n")
	sb.WriteString("3. **Preuves à rechercher**: Quelles preuves supplémentaires confirmeraient ou infirmeraient cette hypothèse?\n")
	sb.WriteString("4. **Scénarios alternatifs**: Y a-t-il d'autres explications possibles?\n")
	sb.WriteString("5. **Recommandation**: L'hypothèse devrait-elle être maintenue, renforcée ou abandonnée?\n")
	sb.WriteString("6. **Niveau de confiance suggéré**: Quel pourcentage de confiance recommandes-tu?\n")

	return s.Generate(sb.String())
}

// AnalyzePath analyse un chemin sémantique
func (s *OllamaService) AnalyzePath(path []string, context string) (string, error) {
	var sb strings.Builder
	sb.WriteString("IMPORTANT: Tu DOIS répondre UNIQUEMENT en FRANÇAIS.\n\n")
	sb.WriteString("Tu es un analyste de liens criminalistique.\n")
	sb.WriteString("Analyse cette chaîne de relations et détermine sa signification:\n\n")

	sb.WriteString("**Chemin**: ")
	sb.WriteString(strings.Join(path, " → "))
	sb.WriteString("\n\n")

	if context != "" {
		sb.WriteString(fmt.Sprintf("**Contexte**: %s\n\n", context))
	}

	sb.WriteString("Analyse:\n")
	sb.WriteString("1. S'agit-il d'une chaîne causale, d'une corrélation ou d'une coïncidence?\n")
	sb.WriteString("2. Quelles implications ce chemin suggère-t-il?\n")
	sb.WriteString("3. Quelles vérifications sont nécessaires?\n")

	return s.Generate(sb.String())
}

// Chat permet une conversation libre avec le LLM
func (s *OllamaService) Chat(message string, caseContext string) (string, error) {
	prompt := s.BuildChatPrompt(message, caseContext)
	return s.Generate(prompt)
}

// BuildChatPrompt construit le prompt pour le chat
func (s *OllamaService) BuildChatPrompt(message string, caseContext string) string {
	var sb strings.Builder

	// Utiliser la configuration si disponible
	pc := s.getPromptConfig("chat")
	if pc != nil {
		sb.WriteString(s.getLanguageInstruction())
		sb.WriteString(pc.System + "\n\n")

		if caseContext != "" {
			if pc.ContextIntro != "" {
				sb.WriteString(pc.ContextIntro + "\n\n")
			}
			sb.WriteString(caseContext + "\n\n")
		}

		sb.WriteString("## QUESTION DE L'ENQUÊTEUR\n")
		sb.WriteString(message + "\n\n")

		if pc.OutputFormat != "" {
			sb.WriteString(pc.OutputFormat + "\n")
		}
	} else {
		sb.WriteString("IMPORTANT: Tu DOIS répondre UNIQUEMENT en FRANÇAIS.\n\n")
		sb.WriteString("Tu es un assistant d'enquête criminalistique expert, spécialisé dans les méthodes PEACE et PROGREAI.\n\n")

		if caseContext != "" {
			sb.WriteString("## DONNÉES DE L'AFFAIRE\n")
			sb.WriteString("Le contexte ci-dessous contient:\n")
			sb.WriteString("- Les résultats de recherche sémantique (éléments les plus pertinents à ta question)\n")
			sb.WriteString("- Toutes les entités avec leurs relations détaillées\n")
			sb.WriteString("- Les preuves et leurs liens avec les entités\n")
			sb.WriteString("- La chronologie des événements\n")
			sb.WriteString("- Les hypothèses d'investigation en cours\n")
			sb.WriteString("- Une représentation N4L (Notes for Linking) des relations\n\n")
			sb.WriteString("UTILISE CES DONNÉES pour répondre de façon précise et contextuelle.\n")
			sb.WriteString("BASE TA RÉPONSE SUR LES FAITS fournis, pas sur des suppositions.\n\n")
			sb.WriteString(caseContext + "\n\n")
		}

		sb.WriteString("## QUESTION DE L'ENQUÊTEUR\n")
		sb.WriteString(message + "\n\n")
		sb.WriteString("## INSTRUCTIONS\n")
		sb.WriteString("- Réponds de manière concise et structurée\n")
		sb.WriteString("- Cite les éléments spécifiques de l'affaire qui justifient ta réponse\n")
		sb.WriteString("- Si des relations ou connexions sont pertinentes, mentionne-les explicitement\n")
		sb.WriteString("- Utilise le format markdown pour structurer ta réponse\n")
	}

	return sb.String()
}

// StreamCallback est appelée pour chaque chunk de réponse
type StreamCallback func(chunk string, done bool) error

// GenerateStream génère une réponse en streaming (vLLM Chat API)
func (s *OllamaService) GenerateStream(prompt string, callback StreamCallback) error {
	reqBody := VLLMChatRequest{
		Model: s.model,
		Messages: []ChatMessage{
			{Role: "user", Content: prompt},
		},
		MaxTokens:   8192,
		Temperature: 0.7,
		Stream:      true,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("erreur marshalling request: %w", err)
	}

	log.Printf("[STREAM] Début streaming vers %s/v1/chat/completions", s.baseURL)

	resp, err := http.Post(s.baseURL+"/v1/chat/completions", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("erreur appel vLLM: %w", err)
	}
	defer resp.Body.Close()

	log.Printf("[STREAM] Réponse HTTP status: %d", resp.StatusCode)

	scanner := bufio.NewScanner(resp.Body)
	totalChars := 0
	chunkCount := 0
	streamEnded := false

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		chunkCount++

		// vLLM SSE format: "data: {...}" ou "data: [DONE]"
		if !strings.HasPrefix(line, "data:") {
			continue // Ignorer les lignes qui ne sont pas des données SSE
		}

		// Extraire le contenu après "data:" (avec ou sans espace)
		data := strings.TrimPrefix(line, "data:")
		data = strings.TrimSpace(data)

		// Vérifier signal de fin SSE standard "[DONE]"
		if data == "[DONE]" {
			log.Printf("[STREAM] Signal [DONE] reçu, totalChars=%d, chunks=%d", totalChars, chunkCount)
			streamEnded = true
			callback("", true)
			break
		}

		var vllmResp VLLMChatStreamResponse
		if err := json.Unmarshal([]byte(data), &vllmResp); err != nil {
			continue // Ignorer les lignes JSON mal formées
		}

		if len(vllmResp.Choices) > 0 {
			choice := vllmResp.Choices[0]
			content := choice.Delta.Content
			totalChars += len(content)

			// Log tous les 100 chunks ou si finish_reason présent
			if chunkCount%100 == 0 || choice.FinishReason != "" {
				log.Printf("[STREAM] Chunk #%d: totalChars=%d, finish_reason='%s'",
					chunkCount, totalChars, choice.FinishReason)
			}

			// Détecter la fin via finish_reason ("stop", "length", etc.)
			done := choice.FinishReason != ""

			if err := callback(content, done); err != nil {
				log.Printf("[STREAM] Erreur callback: %v", err)
				return err
			}

			if done {
				streamEnded = true
				break
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("[STREAM] Erreur scanner: %v", err)
		return fmt.Errorf("erreur lecture stream: %w", err)
	}

	log.Printf("[STREAM] Fin streaming: totalChars=%d, chunkCount=%d, streamEnded=%v", totalChars, chunkCount, streamEnded)

	// Envoyer un signal de fin explicite seulement si pas déjà envoyé
	if !streamEnded {
		log.Printf("[STREAM] Envoi signal de fin forcé (streamEnded=false)")
		callback("", true)
	}

	return nil
}

// ChatStream permet une conversation en streaming
func (s *OllamaService) ChatStream(message string, caseContext string, callback StreamCallback) error {
	prompt := s.BuildChatPrompt(message, caseContext)
	return s.GenerateStream(prompt, callback)
}

// AnalyzeCrossCase analyse les connexions inter-affaires avec l'IA
func (s *OllamaService) AnalyzeCrossCase(currentCase *models.Case, relatedCases map[string]*models.Case, matches []models.CrossCaseMatch) (string, error) {
	var sb strings.Builder
	sb.WriteString("IMPORTANT: Tu DOIS répondre UNIQUEMENT en FRANÇAIS.\n\n")

	// Instructions de rigueur analytique
	sb.WriteString("## RÈGLES D'ANALYSE STRICTES\n\n")
	sb.WriteString("Tu es un analyste criminalistique RIGOUREUX. Tu dois:\n")
	sb.WriteString("- **NE JAMAIS EXTRAPOLER** au-delà des données fournies\n")
	sb.WriteString("- **DISTINGUER CLAIREMENT** les faits vérifiés des hypothèses\n")
	sb.WriteString("- **PRIVILÉGIER les relations DIRECTES** et documentées\n")
	sb.WriteString("- **IGNORER les correspondances faibles** (confiance < 70%) sauf mention explicite\n")
	sb.WriteString("- **NE PAS INVENTER** de liens qui ne sont pas dans les données\n")
	sb.WriteString("- **ÊTRE SCEPTIQUE**: une similarité de nom n'implique PAS une connexion réelle\n")
	sb.WriteString("- **QUALIFIER chaque affirmation** avec son niveau de certitude (certain/probable/possible/spéculatif)\n\n")

	// Affaire courante
	sb.WriteString(fmt.Sprintf("## Affaire principale: %s\n", currentCase.Name))
	sb.WriteString(fmt.Sprintf("**Type**: %s\n", currentCase.Type))
	sb.WriteString(fmt.Sprintf("**Description**: %s\n\n", currentCase.Description))

	// Entités de l'affaire courante
	if len(currentCase.Entities) > 0 {
		sb.WriteString("**Entités principales**:\n")
		for _, e := range currentCase.Entities {
			sb.WriteString(fmt.Sprintf("- %s (%s, %s)\n", e.Name, e.Type, e.Role))
		}
		sb.WriteString("\n")
	}

	// Correspondances trouvées
	sb.WriteString("## Correspondances détectées par le système\n\n")
	sb.WriteString("**ATTENTION**: Ces correspondances sont générées automatiquement par comparaison de chaînes.\n")
	sb.WriteString("Une confiance < 80% indique souvent un FAUX POSITIF (noms similaires mais personnes différentes).\n\n")

	// Grouper par affaire liée
	matchesByCase := make(map[string][]models.CrossCaseMatch)
	for _, m := range matches {
		matchesByCase[m.OtherCaseID] = append(matchesByCase[m.OtherCaseID], m)
	}

	for caseID, caseMatches := range matchesByCase {
		if relatedCase, ok := relatedCases[caseID]; ok {
			sb.WriteString(fmt.Sprintf("### Affaire liée: %s\n", relatedCase.Name))
			sb.WriteString(fmt.Sprintf("**Type**: %s\n", relatedCase.Type))
			sb.WriteString("**Correspondances (à vérifier)**:\n")
			for _, m := range caseMatches {
				reliability := "FAIBLE"
				if m.Confidence >= 90 {
					reliability = "FORTE"
				} else if m.Confidence >= 70 {
					reliability = "MOYENNE"
				}
				sb.WriteString(fmt.Sprintf("- [%s] %s (confiance: %d%% - fiabilité: %s)\n", m.MatchType, m.Description, m.Confidence, reliability))
			}
			sb.WriteString("\n")
		}
	}

	sb.WriteString("## Analyse demandée\n\n")
	sb.WriteString("Fournis une analyse PRUDENTE et FACTUELLE:\n\n")

	sb.WriteString("1. **Correspondances fiables** (confiance ≥ 80% uniquement):\n")
	sb.WriteString("   - Liste UNIQUEMENT les correspondances à haute confiance\n")
	sb.WriteString("   - Pour chacune, indique si c'est la MÊME personne/lieu ou juste un nom similaire\n\n")

	sb.WriteString("2. **Évaluation critique des liens**:\n")
	sb.WriteString("   - Y a-t-il des preuves DIRECTES d'un lien entre ces affaires?\n")
	sb.WriteString("   - Une coïncidence temporelle n'est PAS une preuve de lien\n")
	sb.WriteString("   - Un nom similaire n'est PAS une preuve d'identité commune\n\n")

	sb.WriteString("3. **Faux positifs probables**:\n")
	sb.WriteString("   - Identifie les correspondances qui semblent être des erreurs du système\n")
	sb.WriteString("   - Explique pourquoi certaines correspondances sont probablement non pertinentes\n\n")

	sb.WriteString("4. **Pistes à vérifier** (si liens plausibles):\n")
	sb.WriteString("   - UNIQUEMENT si des correspondances fiables existent\n")
	sb.WriteString("   - Quelles vérifications concrètes permettraient de confirmer/infirmer le lien?\n\n")

	sb.WriteString("5. **Conclusion**:\n")
	sb.WriteString("   - LIEN CONFIRMÉ: preuves directes d'une connexion réelle\n")
	sb.WriteString("   - LIEN PROBABLE: indices forts mais vérification nécessaire\n")
	sb.WriteString("   - LIEN POSSIBLE: quelques indices, enquête approfondie requise\n")
	sb.WriteString("   - PAS DE LIEN: correspondances probablement fortuites\n")
	sb.WriteString("   - Niveau de priorité (1-5) JUSTIFIÉ par les faits\n")

	return s.Generate(sb.String())
}

// cleanJSON extrait et nettoie le JSON d'une réponse
func cleanJSON(response string) string {
	// Trouver le début du JSON
	start := strings.Index(response, "[")
	if start == -1 {
		start = strings.Index(response, "{")
	}
	if start == -1 {
		return response
	}

	// Trouver la fin du JSON
	end := strings.LastIndex(response, "]")
	if end == -1 {
		end = strings.LastIndex(response, "}")
	}
	if end == -1 || end < start {
		return response
	}

	return response[start : end+1]
}
