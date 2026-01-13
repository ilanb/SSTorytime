package services

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// PromptConfig représente la configuration d'un prompt
type PromptConfig struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	System       string `json:"system"`
	Instruction  string `json:"instruction,omitempty"`
	ContextIntro string `json:"context_intro,omitempty"`
	OutputFormat string `json:"output_format"`
}

// ModelsConfig représente la configuration des modèles
type ModelsConfig struct {
	Default       string `json:"default"`
	N4LConversion string `json:"n4l_conversion"`
}

// PromptsConfig représente la configuration complète des prompts
type PromptsConfig struct {
	Version             string                  `json:"version"`
	LanguageInstruction string                  `json:"language_instruction"`
	Prompts             map[string]PromptConfig `json:"prompts"`
	Models              ModelsConfig            `json:"models"`
}

// ConfigService gère la configuration des prompts
type ConfigService struct {
	configPath string
	config     *PromptsConfig
	mu         sync.RWMutex
}

// NewConfigService crée une nouvelle instance du service de configuration
func NewConfigService() *ConfigService {
	cs := &ConfigService{}
	cs.loadConfig()
	return cs
}

// getConfigPath retourne le chemin du fichier de configuration
func (s *ConfigService) getConfigPath() string {
	if s.configPath != "" {
		return s.configPath
	}

	// Chercher le fichier config dans plusieurs emplacements
	paths := []string{
		"config/prompts.json",
		"../config/prompts.json",
		"/Users/ilan/_INFOSTRATES/_AI/SSTorytime-1/src/ForensicInvestigator/config/prompts.json",
	}

	// Essayer de trouver depuis le répertoire de l'exécutable
	if execPath, err := os.Executable(); err == nil {
		execDir := filepath.Dir(execPath)
		paths = append([]string{filepath.Join(execDir, "config", "prompts.json")}, paths...)
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			s.configPath = p
			return p
		}
	}

	// Fallback: créer dans le répertoire courant
	s.configPath = "config/prompts.json"
	return s.configPath
}

// loadConfig charge la configuration depuis le fichier
func (s *ConfigService) loadConfig() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	configPath := s.getConfigPath()
	data, err := os.ReadFile(configPath)
	if err != nil {
		// Si le fichier n'existe pas, utiliser la config par défaut
		s.config = s.getDefaultConfig()
		return nil
	}

	var config PromptsConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("erreur parsing config: %w", err)
	}

	s.config = &config
	return nil
}

// SaveConfig sauvegarde la configuration dans le fichier
func (s *ConfigService) SaveConfig() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	configPath := s.getConfigPath()

	// Créer le répertoire si nécessaire
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("erreur création répertoire config: %w", err)
	}

	data, err := json.MarshalIndent(s.config, "", "  ")
	if err != nil {
		return fmt.Errorf("erreur marshalling config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("erreur écriture config: %w", err)
	}

	return nil
}

// GetConfig retourne la configuration complète
func (s *ConfigService) GetConfig() *PromptsConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config
}

// GetPrompt retourne un prompt spécifique
func (s *ConfigService) GetPrompt(name string) *PromptConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.config == nil {
		return nil
	}

	if prompt, ok := s.config.Prompts[name]; ok {
		return &prompt
	}
	return nil
}

// GetLanguageInstruction retourne l'instruction de langue
func (s *ConfigService) GetLanguageInstruction() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.config != nil {
		return s.config.LanguageInstruction
	}
	return "IMPORTANT: Tu DOIS répondre UNIQUEMENT en FRANÇAIS.\n\n"
}

// UpdatePrompt met à jour un prompt spécifique
func (s *ConfigService) UpdatePrompt(name string, prompt PromptConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.config == nil {
		s.config = s.getDefaultConfig()
	}

	s.config.Prompts[name] = prompt
	return nil
}

// UpdateConfig met à jour la configuration complète
func (s *ConfigService) UpdateConfig(config *PromptsConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.config = config
	return nil
}

// ReloadConfig recharge la configuration depuis le fichier
func (s *ConfigService) ReloadConfig() error {
	return s.loadConfig()
}

// getDefaultConfig retourne la configuration par défaut
func (s *ConfigService) getDefaultConfig() *PromptsConfig {
	return &PromptsConfig{
		Version:             "1.0",
		LanguageInstruction: "IMPORTANT: Tu DOIS répondre UNIQUEMENT en FRANÇAIS.\n\n",
		Prompts: map[string]PromptConfig{
			"analyze_case": {
				Name:        "Analyse d'Affaire",
				Description: "Prompt utilisé pour analyser une affaire et générer un résumé structuré",
				System:      "Tu es un assistant d'enquête criminalistique expert.",
				Instruction: "Analyse les informations suivantes sur cette affaire:",
				OutputFormat: `Rédige un résumé structuré identifiant:
1. Les points clés de l'affaire
2. Les suspects potentiels et leurs mobiles
3. Les pistes à explorer prioritairement
4. Les incohérences ou contradictions éventuelles
5. Les questions d'investigation restantes

Utilise un format markdown structuré.`,
			},
			"generate_hypotheses": {
				Name:        "Génération d'Hypothèses",
				Description: "Prompt pour générer des hypothèses d'investigation basées sur le graphe de connaissances",
				System:      "Tu es un analyste criminalistique expert.",
				Instruction: "En te basant sur le graphe de connaissances suivant, génère des hypothèses d'investigation.",
				OutputFormat: `Génère 3 à 5 hypothèses d'investigation au format JSON:
[{"title": "...", "description": "...", "confidence_level": 0-100, "questions": ["..."]}]
Réponds UNIQUEMENT avec le JSON, sans autre texte.`,
			},
			"detect_contradictions": {
				Name:        "Détection de Contradictions",
				Description: "Prompt pour identifier les incohérences et contradictions dans les données",
				System:      "Tu es un analyste expert en détection d'incohérences.",
				Instruction: "Analyse les relations suivantes et identifie toutes les contradictions potentielles:",
				OutputFormat: `Identifie:
1. Les contradictions directes (A dit X, B dit non-X)
2. Les incohérences temporelles (A était quelque part quand il prétend être ailleurs)
3. Les alibis impossibles à vérifier ou contradictoires
4. Les relations suspectes ou non expliquées

Format ta réponse en markdown avec des sections claires.`,
			},
			"chat": {
				Name:        "Assistant Chat",
				Description: "Prompt pour la conversation avec l'assistant IA",
				System:      "Tu es un assistant d'enquête criminalistique expert, spécialisé dans les méthodes PEACE et PROGREAI.",
				ContextIntro: `## DONNÉES DE L'AFFAIRE
Le contexte ci-dessous contient:
- Les résultats de recherche sémantique (éléments les plus pertinents à ta question)
- Toutes les entités avec leurs relations détaillées
- Les preuves et leurs liens avec les entités
- La chronologie des événements
- Les hypothèses d'investigation en cours
- Une représentation N4L (Notes for Linking) des relations

UTILISE CES DONNÉES pour répondre de façon précise et contextuelle.
BASE TA RÉPONSE SUR LES FAITS fournis, pas sur des suppositions.`,
				OutputFormat: `## INSTRUCTIONS
- Réponds de manière concise et structurée
- Cite les éléments spécifiques de l'affaire qui justifient ta réponse
- Si des relations ou connexions sont pertinentes, mentionne-les explicitement
- Utilise le format markdown pour structurer ta réponse`,
			},
		},
		Models: ModelsConfig{
			Default:       "Qwen/Qwen2.5-7B-Instruct",
			N4LConversion: "n4l-qwen:latest",
		},
	}
}
