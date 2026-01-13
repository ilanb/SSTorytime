package services

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"forensicinvestigator/internal/models"

	"github.com/google/uuid"
)

// NotebookService gère les notebooks et les notes avec persistance JSON
type NotebookService struct {
	notebooks   map[string]*models.Notebook // clé = caseID
	mu          sync.RWMutex
	dataDir     string // Répertoire de stockage des notebooks
}

// NewNotebookService crée une nouvelle instance du service avec persistance
func NewNotebookService() *NotebookService {
	// Déterminer le répertoire de données
	dataDir := "data/notebooks"

	// Créer le répertoire s'il n'existe pas
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Printf("Erreur création répertoire notebooks: %v", err)
	}

	service := &NotebookService{
		notebooks: make(map[string]*models.Notebook),
		dataDir:   dataDir,
	}

	// Charger les notebooks existants
	service.loadAllNotebooks()

	return service
}

// loadAllNotebooks charge tous les notebooks depuis le disque
func (s *NotebookService) loadAllNotebooks() {
	files, err := os.ReadDir(s.dataDir)
	if err != nil {
		log.Printf("Erreur lecture répertoire notebooks: %v", err)
		return
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		filePath := filepath.Join(s.dataDir, file.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			log.Printf("Erreur lecture notebook %s: %v", file.Name(), err)
			continue
		}

		var notebook models.Notebook
		if err := json.Unmarshal(data, &notebook); err != nil {
			log.Printf("Erreur parsing notebook %s: %v", file.Name(), err)
			continue
		}

		s.notebooks[notebook.CaseID] = &notebook
		log.Printf("Notebook chargé: %s (%d notes)", notebook.CaseName, len(notebook.Notes))
	}

	log.Printf("Total: %d notebooks chargés", len(s.notebooks))
}

// saveNotebook sauvegarde un notebook sur le disque
func (s *NotebookService) saveNotebook(notebook *models.Notebook) error {
	// Générer le nom de fichier à partir du caseID (en remplaçant les caractères spéciaux)
	fileName := strings.ReplaceAll(notebook.CaseID, "/", "_")
	fileName = strings.ReplaceAll(fileName, "\\", "_")
	fileName = strings.ReplaceAll(fileName, ":", "_")
	filePath := filepath.Join(s.dataDir, fileName+".json")

	data, err := json.MarshalIndent(notebook, "", "  ")
	if err != nil {
		return fmt.Errorf("erreur serialisation notebook: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("erreur écriture notebook: %w", err)
	}

	return nil
}

// GetNotebook récupère le notebook d'une affaire (le crée s'il n'existe pas)
func (s *NotebookService) GetNotebook(caseID, caseName string) *models.Notebook {
	s.mu.Lock()
	defer s.mu.Unlock()

	notebook, exists := s.notebooks[caseID]
	if !exists {
		notebook = &models.Notebook{
			CaseID:    caseID,
			CaseName:  caseName,
			Notes:     []models.Note{},
			UpdatedAt: time.Now(),
		}
		s.notebooks[caseID] = notebook
		// Sauvegarder le nouveau notebook
		if err := s.saveNotebook(notebook); err != nil {
			log.Printf("Erreur sauvegarde nouveau notebook: %v", err)
		}
	}
	return notebook
}

// AddNote ajoute une note au notebook d'une affaire
func (s *NotebookService) AddNote(caseID, caseName string, note models.Note) (*models.Note, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	notebook, exists := s.notebooks[caseID]
	if !exists {
		notebook = &models.Notebook{
			CaseID:    caseID,
			CaseName:  caseName,
			Notes:     []models.Note{},
			UpdatedAt: time.Now(),
		}
		s.notebooks[caseID] = notebook
	}

	note.ID = uuid.New().String()
	note.CaseID = caseID
	note.CreatedAt = time.Now()
	note.UpdatedAt = time.Now()

	if note.Tags == nil {
		note.Tags = []string{}
	}

	notebook.Notes = append(notebook.Notes, note)
	notebook.UpdatedAt = time.Now()

	// Sauvegarder le notebook
	if err := s.saveNotebook(notebook); err != nil {
		log.Printf("Erreur sauvegarde notebook après ajout note: %v", err)
	}

	return &note, nil
}

// GetNote récupère une note par ID
func (s *NotebookService) GetNote(caseID, noteID string) (*models.Note, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	notebook, exists := s.notebooks[caseID]
	if !exists {
		return nil, fmt.Errorf("notebook non trouvé pour l'affaire: %s", caseID)
	}

	for _, note := range notebook.Notes {
		if note.ID == noteID {
			return &note, nil
		}
	}

	return nil, fmt.Errorf("note non trouvée: %s", noteID)
}

// UpdateNote met à jour une note existante
func (s *NotebookService) UpdateNote(caseID string, note models.Note) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	notebook, exists := s.notebooks[caseID]
	if !exists {
		return fmt.Errorf("notebook non trouvé pour l'affaire: %s", caseID)
	}

	for i, n := range notebook.Notes {
		if n.ID == note.ID {
			note.CaseID = caseID
			note.CreatedAt = n.CreatedAt
			note.UpdatedAt = time.Now()
			if note.Tags == nil {
				note.Tags = n.Tags
			}
			notebook.Notes[i] = note
			notebook.UpdatedAt = time.Now()
			// Sauvegarder le notebook
			if err := s.saveNotebook(notebook); err != nil {
				log.Printf("Erreur sauvegarde notebook après mise à jour note: %v", err)
			}
			return nil
		}
	}

	return fmt.Errorf("note non trouvée: %s", note.ID)
}

// DeleteNote supprime une note
func (s *NotebookService) DeleteNote(caseID, noteID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	notebook, exists := s.notebooks[caseID]
	if !exists {
		return fmt.Errorf("notebook non trouvé pour l'affaire: %s", caseID)
	}

	for i, note := range notebook.Notes {
		if note.ID == noteID {
			notebook.Notes = append(notebook.Notes[:i], notebook.Notes[i+1:]...)
			notebook.UpdatedAt = time.Now()
			// Sauvegarder le notebook
			if err := s.saveNotebook(notebook); err != nil {
				log.Printf("Erreur sauvegarde notebook après suppression note: %v", err)
			}
			return nil
		}
	}

	return fmt.Errorf("note non trouvée: %s", noteID)
}

// TogglePinNote épingle/désépingle une note
func (s *NotebookService) TogglePinNote(caseID, noteID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	notebook, exists := s.notebooks[caseID]
	if !exists {
		return fmt.Errorf("notebook non trouvé pour l'affaire: %s", caseID)
	}

	for i, note := range notebook.Notes {
		if note.ID == noteID {
			notebook.Notes[i].IsPinned = !notebook.Notes[i].IsPinned
			notebook.Notes[i].UpdatedAt = time.Now()
			notebook.UpdatedAt = time.Now()
			// Sauvegarder le notebook
			if err := s.saveNotebook(notebook); err != nil {
				log.Printf("Erreur sauvegarde notebook après toggle pin: %v", err)
			}
			return nil
		}
	}

	return fmt.Errorf("note non trouvée: %s", noteID)
}

// ToggleFavoriteNote marque/démarque une note comme favorite
func (s *NotebookService) ToggleFavoriteNote(caseID, noteID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	notebook, exists := s.notebooks[caseID]
	if !exists {
		return fmt.Errorf("notebook non trouvé pour l'affaire: %s", caseID)
	}

	for i, note := range notebook.Notes {
		if note.ID == noteID {
			notebook.Notes[i].IsFavorite = !notebook.Notes[i].IsFavorite
			notebook.Notes[i].UpdatedAt = time.Now()
			notebook.UpdatedAt = time.Now()
			// Sauvegarder le notebook
			if err := s.saveNotebook(notebook); err != nil {
				log.Printf("Erreur sauvegarde notebook après toggle favorite: %v", err)
			}
			return nil
		}
	}

	return fmt.Errorf("note non trouvée: %s", noteID)
}

// SortOrder définit l'ordre de tri
type SortOrder string

const (
	SortByDateDesc   SortOrder = "date_desc"
	SortByDateAsc    SortOrder = "date_asc"
	SortByTitleAsc   SortOrder = "title_asc"
	SortByTitleDesc  SortOrder = "title_desc"
	SortByType       SortOrder = "type"
	SortByPinned     SortOrder = "pinned"
)

// SearchNotes recherche des notes dans un notebook
func (s *NotebookService) SearchNotes(caseID, query string, noteType string, sortBy SortOrder) []models.Note {
	s.mu.RLock()
	defer s.mu.RUnlock()

	notebook, exists := s.notebooks[caseID]
	if !exists {
		return []models.Note{}
	}

	query = strings.ToLower(query)
	results := []models.Note{}

	for _, note := range notebook.Notes {
		// Filtrer par type si spécifié
		if noteType != "" && noteType != "all" && string(note.Type) != noteType {
			continue
		}

		// Filtrer par recherche textuelle
		if query != "" {
			titleMatch := strings.Contains(strings.ToLower(note.Title), query)
			contentMatch := strings.Contains(strings.ToLower(note.Content), query)
			contextMatch := strings.Contains(strings.ToLower(note.Context), query)
			tagMatch := false
			for _, tag := range note.Tags {
				if strings.Contains(strings.ToLower(tag), query) {
					tagMatch = true
					break
				}
			}

			if !titleMatch && !contentMatch && !contextMatch && !tagMatch {
				continue
			}
		}

		results = append(results, note)
	}

	// Trier les résultats
	s.sortNotes(results, sortBy)

	return results
}

// sortNotes trie les notes selon l'ordre spécifié
func (s *NotebookService) sortNotes(notes []models.Note, sortBy SortOrder) {
	switch sortBy {
	case SortByDateDesc:
		sort.Slice(notes, func(i, j int) bool {
			// Épinglées en premier
			if notes[i].IsPinned != notes[j].IsPinned {
				return notes[i].IsPinned
			}
			return notes[i].CreatedAt.After(notes[j].CreatedAt)
		})
	case SortByDateAsc:
		sort.Slice(notes, func(i, j int) bool {
			if notes[i].IsPinned != notes[j].IsPinned {
				return notes[i].IsPinned
			}
			return notes[i].CreatedAt.Before(notes[j].CreatedAt)
		})
	case SortByTitleAsc:
		sort.Slice(notes, func(i, j int) bool {
			if notes[i].IsPinned != notes[j].IsPinned {
				return notes[i].IsPinned
			}
			return strings.ToLower(notes[i].Title) < strings.ToLower(notes[j].Title)
		})
	case SortByTitleDesc:
		sort.Slice(notes, func(i, j int) bool {
			if notes[i].IsPinned != notes[j].IsPinned {
				return notes[i].IsPinned
			}
			return strings.ToLower(notes[i].Title) > strings.ToLower(notes[j].Title)
		})
	case SortByType:
		sort.Slice(notes, func(i, j int) bool {
			if notes[i].IsPinned != notes[j].IsPinned {
				return notes[i].IsPinned
			}
			return notes[i].Type < notes[j].Type
		})
	case SortByPinned:
		sort.Slice(notes, func(i, j int) bool {
			if notes[i].IsPinned != notes[j].IsPinned {
				return notes[i].IsPinned
			}
			if notes[i].IsFavorite != notes[j].IsFavorite {
				return notes[i].IsFavorite
			}
			return notes[i].CreatedAt.After(notes[j].CreatedAt)
		})
	default:
		// Par défaut: épinglées puis par date décroissante
		sort.Slice(notes, func(i, j int) bool {
			if notes[i].IsPinned != notes[j].IsPinned {
				return notes[i].IsPinned
			}
			return notes[i].CreatedAt.After(notes[j].CreatedAt)
		})
	}
}

// GetAllNotebooks récupère tous les notebooks (pour export ou vue globale)
func (s *NotebookService) GetAllNotebooks() []*models.Notebook {
	s.mu.RLock()
	defer s.mu.RUnlock()

	notebooks := make([]*models.Notebook, 0, len(s.notebooks))
	for _, nb := range s.notebooks {
		notebooks = append(notebooks, nb)
	}
	return notebooks
}

// GetNotebookStats retourne des statistiques sur le notebook
func (s *NotebookService) GetNotebookStats(caseID string) map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := map[string]interface{}{
		"total_notes": 0,
		"pinned":      0,
		"favorites":   0,
		"by_type":     map[string]int{},
	}

	notebook, exists := s.notebooks[caseID]
	if !exists {
		return stats
	}

	stats["total_notes"] = len(notebook.Notes)

	byType := make(map[string]int)
	pinned := 0
	favorites := 0

	for _, note := range notebook.Notes {
		byType[string(note.Type)]++
		if note.IsPinned {
			pinned++
		}
		if note.IsFavorite {
			favorites++
		}
	}

	stats["pinned"] = pinned
	stats["favorites"] = favorites
	stats["by_type"] = byType

	return stats
}

// AddTagToNote ajoute un tag à une note
func (s *NotebookService) AddTagToNote(caseID, noteID, tag string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	notebook, exists := s.notebooks[caseID]
	if !exists {
		return fmt.Errorf("notebook non trouvé pour l'affaire: %s", caseID)
	}

	for i, note := range notebook.Notes {
		if note.ID == noteID {
			// Vérifier si le tag existe déjà
			for _, t := range note.Tags {
				if t == tag {
					return nil // Tag déjà présent
				}
			}
			notebook.Notes[i].Tags = append(notebook.Notes[i].Tags, tag)
			notebook.Notes[i].UpdatedAt = time.Now()
			notebook.UpdatedAt = time.Now()
			// Sauvegarder le notebook
			if err := s.saveNotebook(notebook); err != nil {
				log.Printf("Erreur sauvegarde notebook après ajout tag: %v", err)
			}
			return nil
		}
	}

	return fmt.Errorf("note non trouvée: %s", noteID)
}

// RemoveTagFromNote supprime un tag d'une note
func (s *NotebookService) RemoveTagFromNote(caseID, noteID, tag string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	notebook, exists := s.notebooks[caseID]
	if !exists {
		return fmt.Errorf("notebook non trouvé pour l'affaire: %s", caseID)
	}

	for i, note := range notebook.Notes {
		if note.ID == noteID {
			for j, t := range note.Tags {
				if t == tag {
					notebook.Notes[i].Tags = append(notebook.Notes[i].Tags[:j], notebook.Notes[i].Tags[j+1:]...)
					notebook.Notes[i].UpdatedAt = time.Now()
					notebook.UpdatedAt = time.Now()
					// Sauvegarder le notebook
					if err := s.saveNotebook(notebook); err != nil {
						log.Printf("Erreur sauvegarde notebook après suppression tag: %v", err)
					}
					return nil
				}
			}
			return nil // Tag pas trouvé, pas d'erreur
		}
	}

	return fmt.Errorf("note non trouvée: %s", noteID)
}
