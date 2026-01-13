package services

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"forensicinvestigator/internal/models"
)

// GraphAnalyzerService fournit des analyses avancées de graphes
type GraphAnalyzerService struct{}

// NewGraphAnalyzerService crée une nouvelle instance du service
func NewGraphAnalyzerService() *GraphAnalyzerService {
	return &GraphAnalyzerService{}
}

// ClusterResult représente un cluster détecté dans le graphe
type ClusterResult struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Nodes    []string `json:"nodes"`
	Size     int      `json:"size"`
	Density  float64  `json:"density"`
	CentralNode string `json:"central_node"`
}

// PathResult représente un chemin entre deux nœuds
type PathResult struct {
	From   string     `json:"from"`
	To     string     `json:"to"`
	Path   []string   `json:"path"`
	Edges  []PathEdge `json:"edges"`
	Length int        `json:"length"`
}

// PathEdge représente une arête dans un chemin
type PathEdge struct {
	From  string `json:"from"`
	To    string `json:"to"`
	Label string `json:"label"`
}

// LayeredGraphResult représente le graphe organisé en couches
type LayeredGraphResult struct {
	Layers []GraphLayer       `json:"layers"`
	Nodes  []models.GraphNode `json:"nodes"`
	Edges  []models.GraphEdge `json:"edges"`
}

// GraphLayer représente une couche du graphe
type GraphLayer struct {
	Level int      `json:"level"`
	Nodes []string `json:"nodes"`
	Name  string   `json:"name"`
}

// ExpansionConeResult représente le cône d'expansion d'un nœud
type ExpansionConeResult struct {
	CenterNode  string               `json:"center_node"`
	Depth       int                  `json:"depth"`
	Levels      []ExpansionLevel     `json:"levels"`
	TotalNodes  int                  `json:"total_nodes"`
	TotalEdges  int                  `json:"total_edges"`
}

// ExpansionLevel représente un niveau du cône d'expansion
type ExpansionLevel struct {
	Level int                  `json:"level"`
	Nodes []models.GraphNode   `json:"nodes"`
	Edges []models.GraphEdge   `json:"edges"`
}

// DensityMapResult représente la carte de densité du graphe
type DensityMapResult struct {
	Zones          []DensityZone `json:"zones"`
	OverallDensity float64       `json:"overall_density"`
	Suggestions    []string      `json:"suggestions"`
}

// DensityZone représente une zone de densité
type DensityZone struct {
	Name       string   `json:"name"`
	Nodes      []string `json:"nodes"`
	Density    float64  `json:"density"`
	Status     string   `json:"status"` // "explored", "partial", "unexplored"
	EdgeCount  int      `json:"edge_count"`
}

// TemporalPattern représente un pattern temporel détecté
type TemporalPattern struct {
	ID          string   `json:"id"`
	Type        string   `json:"type"` // "sequence", "cycle", "gap", "cluster"
	Description string   `json:"description"`
	Nodes       []string `json:"nodes"`
	Confidence  float64  `json:"confidence"`
}

// ConsistencyResult représente le résultat de vérification de cohérence
type ConsistencyResult struct {
	IsConsistent    bool                   `json:"is_consistent"`
	Contradictions  []GraphContradiction   `json:"contradictions"`
	Warnings        []string               `json:"warnings"`
	OrphanNodes     []string               `json:"orphan_nodes"`
	CyclicRelations [][]string             `json:"cyclic_relations"`
}

// GraphContradiction représente une contradiction détectée dans le graphe
type GraphContradiction struct {
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Nodes       []string `json:"nodes"`
	Severity    string   `json:"severity"` // "high", "medium", "low"
}

// FindClusters détecte les clusters dans le graphe
func (s *GraphAnalyzerService) FindClusters(graph models.GraphData) []ClusterResult {
	if len(graph.Nodes) == 0 {
		return []ClusterResult{}
	}

	// Construire la matrice d'adjacence
	adjacency := make(map[string]map[string]bool)
	for _, node := range graph.Nodes {
		adjacency[node.ID] = make(map[string]bool)
	}
	for _, edge := range graph.Edges {
		if adjacency[edge.From] != nil {
			adjacency[edge.From][edge.To] = true
		}
		if adjacency[edge.To] != nil {
			adjacency[edge.To][edge.From] = true
		}
	}

	// Trouver les composantes connexes (clusters)
	visited := make(map[string]bool)
	var clusters []ClusterResult
	clusterID := 0

	for _, node := range graph.Nodes {
		if visited[node.ID] {
			continue
		}

		// BFS pour trouver tous les nœuds connectés
		cluster := []string{}
		queue := []string{node.ID}

		for len(queue) > 0 {
			current := queue[0]
			queue = queue[1:]

			if visited[current] {
				continue
			}
			visited[current] = true
			cluster = append(cluster, current)

			for neighbor := range adjacency[current] {
				if !visited[neighbor] {
					queue = append(queue, neighbor)
				}
			}
		}

		if len(cluster) > 0 {
			// Calculer la densité du cluster
			edgeCount := 0
			for _, n1 := range cluster {
				for _, n2 := range cluster {
					if adjacency[n1][n2] {
						edgeCount++
					}
				}
			}
			maxEdges := len(cluster) * (len(cluster) - 1)
			density := 0.0
			if maxEdges > 0 {
				density = float64(edgeCount) / float64(maxEdges)
			}

			// Trouver le nœud central (plus de connexions)
			centralNode := cluster[0]
			maxConnections := 0
			for _, n := range cluster {
				connections := len(adjacency[n])
				if connections > maxConnections {
					maxConnections = connections
					centralNode = n
				}
			}

			clusters = append(clusters, ClusterResult{
				ID:          fmt.Sprintf("cluster_%d", clusterID),
				Name:        fmt.Sprintf("Groupe %d", clusterID+1),
				Nodes:       cluster,
				Size:        len(cluster),
				Density:     density,
				CentralNode: centralNode,
			})
			clusterID++
		}
	}

	// Trier par taille décroissante
	sort.Slice(clusters, func(i, j int) bool {
		return clusters[i].Size > clusters[j].Size
	})

	return clusters
}

// FindAllPaths trouve tous les chemins entre deux nœuds
func (s *GraphAnalyzerService) FindAllPaths(graph models.GraphData, from, to string, maxDepth int) []PathResult {
	if maxDepth <= 0 {
		maxDepth = 5
	}

	// Construire l'adjacence avec les labels
	adjacency := make(map[string][]struct {
		To    string
		Label string
	})
	for _, edge := range graph.Edges {
		adjacency[edge.From] = append(adjacency[edge.From], struct {
			To    string
			Label string
		}{edge.To, edge.Label})
		// Ajouter aussi dans l'autre sens pour les relations non-directionnelles
		adjacency[edge.To] = append(adjacency[edge.To], struct {
			To    string
			Label string
		}{edge.From, edge.Label})
	}

	var paths []PathResult
	var dfs func(current string, target string, path []string, edges []PathEdge, visited map[string]bool, depth int)

	dfs = func(current string, target string, path []string, edges []PathEdge, visited map[string]bool, depth int) {
		if depth > maxDepth {
			return
		}
		if current == target && len(path) > 1 {
			pathCopy := make([]string, len(path))
			copy(pathCopy, path)
			edgesCopy := make([]PathEdge, len(edges))
			copy(edgesCopy, edges)
			paths = append(paths, PathResult{
				From:   from,
				To:     to,
				Path:   pathCopy,
				Edges:  edgesCopy,
				Length: len(path) - 1,
			})
			return
		}

		for _, neighbor := range adjacency[current] {
			if !visited[neighbor.To] {
				visited[neighbor.To] = true
				newPath := append(path, neighbor.To)
				newEdges := append(edges, PathEdge{
					From:  current,
					To:    neighbor.To,
					Label: neighbor.Label,
				})
				dfs(neighbor.To, target, newPath, newEdges, visited, depth+1)
				visited[neighbor.To] = false
			}
		}
	}

	visited := make(map[string]bool)
	visited[from] = true
	dfs(from, to, []string{from}, []PathEdge{}, visited, 1)

	// Trier par longueur
	sort.Slice(paths, func(i, j int) bool {
		return paths[i].Length < paths[j].Length
	})

	// Limiter le nombre de résultats
	if len(paths) > 10 {
		paths = paths[:10]
	}

	return paths
}

// GetLayeredGraph organise le graphe en couches hiérarchiques
func (s *GraphAnalyzerService) GetLayeredGraph(graph models.GraphData) LayeredGraphResult {
	if len(graph.Nodes) == 0 {
		return LayeredGraphResult{Layers: []GraphLayer{}}
	}

	// Calculer le degré entrant de chaque nœud
	inDegree := make(map[string]int)
	outDegree := make(map[string]int)
	for _, node := range graph.Nodes {
		inDegree[node.ID] = 0
		outDegree[node.ID] = 0
	}
	for _, edge := range graph.Edges {
		inDegree[edge.To]++
		outDegree[edge.From]++
	}

	// Assigner les niveaux par ordre topologique
	levels := make(map[string]int)
	assigned := make(map[string]bool)

	// Niveau 0: nœuds sans arêtes entrantes (sources)
	var currentLevel []string
	for _, node := range graph.Nodes {
		if inDegree[node.ID] == 0 {
			levels[node.ID] = 0
			assigned[node.ID] = true
			currentLevel = append(currentLevel, node.ID)
		}
	}

	// Propager les niveaux
	level := 1
	for len(assigned) < len(graph.Nodes) && level < 20 {
		var nextLevel []string
		for _, node := range graph.Nodes {
			if assigned[node.ID] {
				continue
			}
			// Vérifier si tous les prédécesseurs sont assignés
			allPredAssigned := true
			maxPredLevel := -1
			for _, edge := range graph.Edges {
				if edge.To == node.ID {
					if !assigned[edge.From] {
						allPredAssigned = false
						break
					}
					if levels[edge.From] > maxPredLevel {
						maxPredLevel = levels[edge.From]
					}
				}
			}
			if allPredAssigned {
				levels[node.ID] = maxPredLevel + 1
				assigned[node.ID] = true
				nextLevel = append(nextLevel, node.ID)
			}
		}
		// Si aucun progrès, assigner les restants au niveau courant
		if len(nextLevel) == 0 {
			for _, node := range graph.Nodes {
				if !assigned[node.ID] {
					levels[node.ID] = level
					assigned[node.ID] = true
				}
			}
		}
		currentLevel = nextLevel
		level++
	}

	// Grouper par niveau
	layerMap := make(map[int][]string)
	maxLevel := 0
	for nodeID, lvl := range levels {
		layerMap[lvl] = append(layerMap[lvl], nodeID)
		if lvl > maxLevel {
			maxLevel = lvl
		}
	}

	// Construire le résultat
	var layers []GraphLayer
	for l := 0; l <= maxLevel; l++ {
		if nodes, ok := layerMap[l]; ok {
			layerName := "Niveau " + fmt.Sprintf("%d", l)
			if l == 0 {
				layerName = "Sources"
			} else if l == maxLevel {
				layerName = "Destinations"
			}
			layers = append(layers, GraphLayer{
				Level: l,
				Nodes: nodes,
				Name:  layerName,
			})
		}
	}

	return LayeredGraphResult{
		Layers: layers,
		Nodes:  graph.Nodes,
		Edges:  graph.Edges,
	}
}

// GetExpansionCone retourne le cône d'expansion d'un nœud
func (s *GraphAnalyzerService) GetExpansionCone(graph models.GraphData, nodeID string, depth int) ExpansionConeResult {
	if depth <= 0 {
		depth = 3
	}

	// Construire l'adjacence bidirectionnelle
	adjacency := make(map[string][]models.GraphEdge)
	for _, edge := range graph.Edges {
		adjacency[edge.From] = append(adjacency[edge.From], edge)
		// Ajouter aussi l'arête inverse
		reverseEdge := edge
		reverseEdge.From = edge.To
		reverseEdge.To = edge.From
		adjacency[edge.To] = append(adjacency[edge.To], reverseEdge)
	}

	// Construire un map des nœuds
	nodeMap := make(map[string]models.GraphNode)
	for _, node := range graph.Nodes {
		nodeMap[node.ID] = node
	}

	visited := make(map[string]bool)
	visitedEdges := make(map[string]bool)
	var levels []ExpansionLevel

	currentLevel := []string{nodeID}
	visited[nodeID] = true

	for d := 0; d <= depth && len(currentLevel) > 0; d++ {
		var levelNodes []models.GraphNode
		var levelEdges []models.GraphEdge
		var nextLevel []string

		for _, current := range currentLevel {
			if node, ok := nodeMap[current]; ok {
				levelNodes = append(levelNodes, node)
			}

			for _, edge := range adjacency[current] {
				edgeKey := edge.From + "->" + edge.To
				reverseKey := edge.To + "->" + edge.From
				if !visitedEdges[edgeKey] && !visitedEdges[reverseKey] {
					visitedEdges[edgeKey] = true
					levelEdges = append(levelEdges, edge)
				}

				if !visited[edge.To] {
					visited[edge.To] = true
					nextLevel = append(nextLevel, edge.To)
				}
			}
		}

		if len(levelNodes) > 0 {
			levels = append(levels, ExpansionLevel{
				Level: d,
				Nodes: levelNodes,
				Edges: levelEdges,
			})
		}

		currentLevel = nextLevel
	}

	totalNodes := 0
	totalEdges := 0
	for _, level := range levels {
		totalNodes += len(level.Nodes)
		totalEdges += len(level.Edges)
	}

	return ExpansionConeResult{
		CenterNode: nodeID,
		Depth:      depth,
		Levels:     levels,
		TotalNodes: totalNodes,
		TotalEdges: totalEdges,
	}
}

// GetDensityMap génère une carte de densité du graphe
func (s *GraphAnalyzerService) GetDensityMap(graph models.GraphData) DensityMapResult {
	clusters := s.FindClusters(graph)

	var zones []DensityZone
	var suggestions []string

	for _, cluster := range clusters {
		status := "explored"
		if cluster.Density < 0.3 {
			status = "unexplored"
			suggestions = append(suggestions, fmt.Sprintf("Le groupe '%s' a une faible densité (%.1f%%). Envisagez d'explorer plus de relations.", cluster.Name, cluster.Density*100))
		} else if cluster.Density < 0.6 {
			status = "partial"
		}

		zones = append(zones, DensityZone{
			Name:      cluster.Name,
			Nodes:     cluster.Nodes,
			Density:   cluster.Density,
			Status:    status,
			EdgeCount: int(cluster.Density * float64(len(cluster.Nodes)*(len(cluster.Nodes)-1))),
		})
	}

	// Calculer la densité globale
	maxEdges := len(graph.Nodes) * (len(graph.Nodes) - 1)
	overallDensity := 0.0
	if maxEdges > 0 {
		overallDensity = float64(len(graph.Edges)) / float64(maxEdges)
	}

	if overallDensity < 0.1 {
		suggestions = append(suggestions, "Le graphe global est très peu dense. Beaucoup de relations restent à explorer.")
	}

	return DensityMapResult{
		Zones:          zones,
		OverallDensity: overallDensity,
		Suggestions:    suggestions,
	}
}

// DetectTemporalPatterns détecte les patterns temporels dans le graphe
func (s *GraphAnalyzerService) DetectTemporalPatterns(graph models.GraphData) []TemporalPattern {
	var patterns []TemporalPattern

	// Détecter les séquences (chaînes de nœuds)
	sequences := s.detectSequences(graph)
	for i, seq := range sequences {
		if len(seq) >= 3 {
			patterns = append(patterns, TemporalPattern{
				ID:          fmt.Sprintf("seq_%d", i),
				Type:        "sequence",
				Description: fmt.Sprintf("Séquence de %d événements: %s", len(seq), strings.Join(seq, " → ")),
				Nodes:       seq,
				Confidence:  0.8,
			})
		}
	}

	// Détecter les cycles
	cycles := s.detectCycles(graph)
	for i, cycle := range cycles {
		patterns = append(patterns, TemporalPattern{
			ID:          fmt.Sprintf("cycle_%d", i),
			Type:        "cycle",
			Description: fmt.Sprintf("Cycle détecté: %s", strings.Join(cycle, " → ")),
			Nodes:       cycle,
			Confidence:  0.9,
		})
	}

	return patterns
}

// detectSequences trouve les chaînes linéaires dans le graphe
func (s *GraphAnalyzerService) detectSequences(graph models.GraphData) [][]string {
	// Calculer le degré entrant et sortant
	inDegree := make(map[string]int)
	outDegree := make(map[string]int)
	next := make(map[string]string)

	for _, node := range graph.Nodes {
		inDegree[node.ID] = 0
		outDegree[node.ID] = 0
	}

	for _, edge := range graph.Edges {
		inDegree[edge.To]++
		outDegree[edge.From]++
		if outDegree[edge.From] == 1 {
			next[edge.From] = edge.To
		}
	}

	var sequences [][]string
	visited := make(map[string]bool)

	for _, node := range graph.Nodes {
		if visited[node.ID] || inDegree[node.ID] != 0 {
			continue
		}

		// Suivre la chaîne
		var seq []string
		current := node.ID
		for current != "" && !visited[current] {
			visited[current] = true
			seq = append(seq, current)
			if outDegree[current] == 1 {
				current = next[current]
			} else {
				break
			}
		}

		if len(seq) >= 2 {
			sequences = append(sequences, seq)
		}
	}

	return sequences
}

// detectCycles trouve les cycles dans le graphe
func (s *GraphAnalyzerService) detectCycles(graph models.GraphData) [][]string {
	adjacency := make(map[string][]string)
	for _, edge := range graph.Edges {
		adjacency[edge.From] = append(adjacency[edge.From], edge.To)
	}

	var cycles [][]string
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	path := []string{}

	var dfs func(node string) bool
	dfs = func(node string) bool {
		visited[node] = true
		recStack[node] = true
		path = append(path, node)

		for _, neighbor := range adjacency[node] {
			if !visited[neighbor] {
				if dfs(neighbor) {
					return true
				}
			} else if recStack[neighbor] {
				// Cycle trouvé
				cycleStart := -1
				for i, n := range path {
					if n == neighbor {
						cycleStart = i
						break
					}
				}
				if cycleStart >= 0 {
					cycle := make([]string, len(path)-cycleStart)
					copy(cycle, path[cycleStart:])
					cycles = append(cycles, cycle)
				}
			}
		}

		path = path[:len(path)-1]
		recStack[node] = false
		return false
	}

	for _, node := range graph.Nodes {
		if !visited[node.ID] {
			dfs(node.ID)
		}
	}

	return cycles
}

// CheckConsistency vérifie la cohérence du graphe
func (s *GraphAnalyzerService) CheckConsistency(graph models.GraphData) ConsistencyResult {
	result := ConsistencyResult{
		IsConsistent:    true,
		Contradictions:  []GraphContradiction{},
		Warnings:        []string{},
		OrphanNodes:     []string{},
		CyclicRelations: [][]string{},
	}

	// Trouver les nœuds orphelins
	connected := make(map[string]bool)
	for _, edge := range graph.Edges {
		connected[edge.From] = true
		connected[edge.To] = true
	}
	for _, node := range graph.Nodes {
		if !connected[node.ID] {
			result.OrphanNodes = append(result.OrphanNodes, node.ID)
			result.Warnings = append(result.Warnings, fmt.Sprintf("Nœud isolé: %s", node.Label))
		}
	}

	// Détecter les cycles
	cycles := s.detectCycles(graph)
	result.CyclicRelations = cycles
	for _, cycle := range cycles {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Relation cyclique détectée: %s", strings.Join(cycle, " → ")))
	}

	// Détecter les contradictions potentielles (arêtes opposées)
	edgeSet := make(map[string]string) // from-to -> label
	for _, edge := range graph.Edges {
		key := edge.From + "|" + edge.To
		reverseKey := edge.To + "|" + edge.From
		if existingLabel, exists := edgeSet[reverseKey]; exists {
			// Arête inverse existe
			if isContradictoryRelation(edge.Label, existingLabel) {
				result.Contradictions = append(result.Contradictions, GraphContradiction{
					Type:        "opposing_relations",
					Description: fmt.Sprintf("Relations potentiellement contradictoires entre %s et %s: '%s' vs '%s'", edge.From, edge.To, edge.Label, existingLabel),
					Nodes:       []string{edge.From, edge.To},
					Severity:    "medium",
				})
				result.IsConsistent = false
			}
		}
		edgeSet[key] = edge.Label
	}

	return result
}

// isContradictoryRelation vérifie si deux labels de relation sont contradictoires
func isContradictoryRelation(label1, label2 string) bool {
	contradictions := map[string][]string{
		"ami":      {"ennemi", "adversaire"},
		"ennemi":   {"ami", "allié"},
		"innocent": {"coupable"},
		"coupable": {"innocent"},
		"vrai":     {"faux"},
		"faux":     {"vrai"},
		"confirme": {"infirme", "contredit"},
		"infirme":  {"confirme"},
	}

	l1 := strings.ToLower(label1)
	l2 := strings.ToLower(label2)

	if opposites, ok := contradictions[l1]; ok {
		for _, opp := range opposites {
			if strings.Contains(l2, opp) {
				return true
			}
		}
	}
	return false
}

// InvestigationStep représente une étape du mode investigation
type InvestigationStep struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Icon        string   `json:"icon"`
	Status      string   `json:"status"` // "pending", "in_progress", "completed"
	Questions   []string `json:"questions"`
	Findings    []string `json:"findings"`
}

// InvestigationSession représente une session d'investigation guidée
type InvestigationSession struct {
	ID             string              `json:"id"`
	CaseID         string              `json:"case_id"`
	CurrentStep    int                 `json:"current_step"`
	Steps          []InvestigationStep `json:"steps"`
	StartedAt      string              `json:"started_at"`
	Insights       []string            `json:"insights"`
	Recommendations []string           `json:"recommendations"`
}

// CreateInvestigationSession crée une nouvelle session d'investigation
func (s *GraphAnalyzerService) CreateInvestigationSession(caseID string, graph models.GraphData) InvestigationSession {
	steps := []InvestigationStep{
		{
			ID:          "actors",
			Name:        "Identification des Acteurs",
			Description: "Identifier toutes les personnes impliquées et leurs rôles",
			Icon:        "people",
			Status:      "pending",
			Questions: []string{
				"Qui sont les victimes ?",
				"Qui sont les suspects potentiels ?",
				"Qui sont les témoins clés ?",
				"Y a-t-il des liens familiaux ou professionnels ?",
			},
			Findings: []string{},
		},
		{
			ID:          "locations",
			Name:        "Analyse des Lieux",
			Description: "Cartographier les lieux pertinents et leurs connexions",
			Icon:        "place",
			Status:      "pending",
			Questions: []string{
				"Où les faits se sont-ils déroulés ?",
				"Quels lieux sont fréquentés par les suspects ?",
				"Y a-t-il des correspondances géographiques ?",
				"Quelle est la proximité entre les différents lieux ?",
			},
			Findings: []string{},
		},
		{
			ID:          "timeline",
			Name:        "Reconstitution Chronologique",
			Description: "Établir la séquence des événements",
			Icon:        "schedule",
			Status:      "pending",
			Questions: []string{
				"Quelle est la chronologie des événements ?",
				"Y a-t-il des alibis à vérifier ?",
				"Existe-t-il des incohérences temporelles ?",
				"Quels sont les moments clés ?",
			},
			Findings: []string{},
		},
		{
			ID:          "motives",
			Name:        "Analyse des Mobiles",
			Description: "Explorer les motivations potentielles",
			Icon:        "psychology",
			Status:      "pending",
			Questions: []string{
				"Quels sont les mobiles possibles ?",
				"Qui avait intérêt à commettre les faits ?",
				"Y a-t-il des conflits préexistants ?",
				"Quels sont les gains potentiels ?",
			},
			Findings: []string{},
		},
		{
			ID:          "evidence",
			Name:        "Évaluation des Preuves",
			Description: "Analyser la solidité des éléments de preuve",
			Icon:        "fact_check",
			Status:      "pending",
			Questions: []string{
				"Quelles preuves sont disponibles ?",
				"Quel est le niveau de fiabilité de chaque preuve ?",
				"Y a-t-il des preuves manquantes ?",
				"Les preuves corroborent-elles les témoignages ?",
			},
			Findings: []string{},
		},
		{
			ID:          "synthesis",
			Name:        "Synthèse et Hypothèses",
			Description: "Formuler les conclusions et hypothèses principales",
			Icon:        "lightbulb",
			Status:      "pending",
			Questions: []string{
				"Quelle hypothèse principale se dégage ?",
				"Quelles pistes restent à explorer ?",
				"Quelles sont les zones d'ombre ?",
				"Quelles actions recommandez-vous ?",
			},
			Findings: []string{},
		},
	}

	// Pré-remplir certaines informations à partir du graphe
	session := InvestigationSession{
		ID:             fmt.Sprintf("inv_%s", caseID),
		CaseID:         caseID,
		CurrentStep:    0,
		Steps:          steps,
		Insights:       []string{},
		Recommendations: []string{},
	}

	// Analyser le graphe pour générer des insights initiaux
	clusters := s.FindClusters(graph)
	if len(clusters) > 1 {
		session.Insights = append(session.Insights, fmt.Sprintf("%d groupes distincts détectés dans le graphe", len(clusters)))
	}

	density := s.GetDensityMap(graph)
	if density.OverallDensity < 0.2 {
		session.Recommendations = append(session.Recommendations, "Le graphe est peu dense - envisagez d'explorer plus de relations")
	}

	consistency := s.CheckConsistency(graph)
	if len(consistency.OrphanNodes) > 0 {
		session.Insights = append(session.Insights, fmt.Sprintf("%d entités isolées détectées", len(consistency.OrphanNodes)))
	}
	if len(consistency.Contradictions) > 0 {
		session.Recommendations = append(session.Recommendations, fmt.Sprintf("%d contradictions potentielles à examiner", len(consistency.Contradictions)))
	}

	return session
}

// GetStepSuggestions génère des suggestions pour une étape d'investigation
func (s *GraphAnalyzerService) GetStepSuggestions(graph models.GraphData, stepID string) []string {
	var suggestions []string

	// Créer une map ID -> Label pour afficher les noms
	nodeLabels := make(map[string]string)
	for _, node := range graph.Nodes {
		nodeLabels[node.ID] = node.Label
	}

	// Helper pour obtenir le nom d'un nœud
	getNodeName := func(id string) string {
		if label, ok := nodeLabels[id]; ok && label != "" {
			return label
		}
		return id
	}

	switch stepID {
	case "actors":
		// Suggérer les nœuds les plus connectés
		connections := make(map[string]int)
		for _, edge := range graph.Edges {
			connections[edge.From]++
			connections[edge.To]++
		}
		for nodeID, count := range connections {
			if count >= 3 {
				suggestions = append(suggestions, fmt.Sprintf("'%s' est fortement connecté (%d relations) - acteur clé potentiel", getNodeName(nodeID), count))
			}
		}

	case "locations":
		// Identifier les nœuds de type lieu
		for _, node := range graph.Nodes {
			if strings.Contains(strings.ToLower(node.Type), "lieu") || strings.Contains(strings.ToLower(node.Label), "rue") || strings.Contains(strings.ToLower(node.Label), "maison") {
				suggestions = append(suggestions, fmt.Sprintf("Lieu identifié: %s", node.Label))
			}
		}

	case "timeline":
		// Détecter les patterns temporels
		patterns := s.DetectTemporalPatterns(graph)
		for _, p := range patterns {
			if p.Type == "sequence" {
				// Remplacer les IDs par les noms dans la description
				desc := p.Description
				for id, label := range nodeLabels {
					desc = strings.ReplaceAll(desc, id, label)
				}
				suggestions = append(suggestions, desc)
			}
		}

	case "motives":
		// Identifier les relations de conflit
		for _, edge := range graph.Edges {
			label := strings.ToLower(edge.Label)
			if strings.Contains(label, "conflit") || strings.Contains(label, "dette") || strings.Contains(label, "héritage") || strings.Contains(label, "rivalité") {
				suggestions = append(suggestions, fmt.Sprintf("Mobile potentiel: %s entre %s et %s", edge.Label, getNodeName(edge.From), getNodeName(edge.To)))
			}
		}

	case "evidence":
		// Analyser la cohérence
		consistency := s.CheckConsistency(graph)
		for _, c := range consistency.Contradictions {
			// Remplacer les IDs par les noms dans la description
			desc := c.Description
			for id, label := range nodeLabels {
				desc = strings.ReplaceAll(desc, id, label)
			}
			suggestions = append(suggestions, fmt.Sprintf("Contradiction à vérifier: %s", desc))
		}
	}

	return suggestions
}

// CentralityResult représente les métriques de centralité d'un nœud
type CentralityResult struct {
	NodeID            string  `json:"node_id"`
	NodeLabel         string  `json:"node_label"`
	NodeType          string  `json:"node_type"`
	DegreeCentrality  int     `json:"degree_centrality"`
	BetweennessCentrality float64 `json:"betweenness_centrality"`
	ClosenessCentrality   float64 `json:"closeness_centrality"`
	Score             float64 `json:"score"`
}

// SuspicionResult représente le score de suspicion d'une personne
type SuspicionResult struct {
	NodeID       string   `json:"node_id"`
	NodeLabel    string   `json:"node_label"`
	Score        int      `json:"score"`
	Factors      []SuspicionFactor `json:"factors"`
	Level        string   `json:"level"` // high, medium, low
}

// SuspicionFactor représente un facteur de suspicion
type SuspicionFactor struct {
	Name   string `json:"name"`
	Value  int    `json:"value"`
	Level  string `json:"level"` // high, medium, low
}

// AlibiInfo représente les informations d'alibi d'une personne
type AlibiInfo struct {
	PersonID    string      `json:"person_id"`
	PersonName  string      `json:"person_name"`
	PersonRole  string      `json:"person_role"`
	Alibis      []AlibiBlock `json:"alibis"`
	HasOpportunity bool     `json:"has_opportunity"`
}

// AlibiBlock représente une période d'alibi
type AlibiBlock struct {
	StartTime   string `json:"start_time"`
	EndTime     string `json:"end_time"`
	Location    string `json:"location"`
	Verified    bool   `json:"verified"`
	Description string `json:"description"`
}

// AlibiTimeline représente la timeline complète des alibis
type AlibiTimeline struct {
	Persons     []AlibiInfo `json:"persons"`
	CrimeTime   string      `json:"crime_time"`
	TimeRange   TimeRange   `json:"time_range"`
}

// TimeRange représente une plage horaire
type TimeRange struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

// CalculateCentrality calcule les métriques de centralité pour tous les nœuds
func (s *GraphAnalyzerService) CalculateCentrality(graph models.GraphData) []CentralityResult {
	results := make([]CentralityResult, 0)

	// Créer une map ID -> Label
	nodeLabels := make(map[string]string)
	nodeTypes := make(map[string]string)
	for _, node := range graph.Nodes {
		nodeLabels[node.ID] = node.Label
		nodeTypes[node.ID] = node.Type
	}

	// Calculer le degré de chaque nœud
	degrees := make(map[string]int)
	for _, edge := range graph.Edges {
		degrees[edge.From]++
		degrees[edge.To]++
	}

	// Trouver le degré maximum pour normaliser
	maxDegree := 0
	for _, d := range degrees {
		if d > maxDegree {
			maxDegree = d
		}
	}

	// Calculer la centralité de proximité (simplifiée)
	closeness := s.calculateCloseness(graph)

	// Calculer la centralité d'intermédiarité (simplifiée)
	betweenness := s.calculateBetweenness(graph)

	for _, node := range graph.Nodes {
		degree := degrees[node.ID]
		normalizedDegree := 0.0
		if maxDegree > 0 {
			normalizedDegree = float64(degree) / float64(maxDegree)
		}

		// Score combiné (pondéré)
		score := normalizedDegree*0.4 + closeness[node.ID]*0.3 + betweenness[node.ID]*0.3

		results = append(results, CentralityResult{
			NodeID:              node.ID,
			NodeLabel:           node.Label,
			NodeType:            node.Type,
			DegreeCentrality:    degree,
			BetweennessCentrality: betweenness[node.ID],
			ClosenessCentrality:   closeness[node.ID],
			Score:               score,
		})
	}

	// Trier par score décroissant
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results
}

// calculateCloseness calcule la centralité de proximité
func (s *GraphAnalyzerService) calculateCloseness(graph models.GraphData) map[string]float64 {
	closeness := make(map[string]float64)

	// Construire la liste d'adjacence
	adj := make(map[string][]string)
	for _, edge := range graph.Edges {
		adj[edge.From] = append(adj[edge.From], edge.To)
		adj[edge.To] = append(adj[edge.To], edge.From)
	}

	for _, node := range graph.Nodes {
		// BFS pour calculer les distances
		distances := s.bfsDistances(node.ID, adj)
		totalDist := 0
		reachable := 0
		for _, d := range distances {
			if d > 0 {
				totalDist += d
				reachable++
			}
		}

		if totalDist > 0 && reachable > 0 {
			closeness[node.ID] = float64(reachable) / float64(totalDist)
		} else {
			closeness[node.ID] = 0
		}
	}

	// Normaliser
	maxCloseness := 0.0
	for _, c := range closeness {
		if c > maxCloseness {
			maxCloseness = c
		}
	}
	if maxCloseness > 0 {
		for id := range closeness {
			closeness[id] = closeness[id] / maxCloseness
		}
	}

	return closeness
}

// calculateBetweenness calcule la centralité d'intermédiarité (simplifiée)
func (s *GraphAnalyzerService) calculateBetweenness(graph models.GraphData) map[string]float64 {
	betweenness := make(map[string]float64)

	// Construire la liste d'adjacence
	adj := make(map[string][]string)
	for _, edge := range graph.Edges {
		adj[edge.From] = append(adj[edge.From], edge.To)
		adj[edge.To] = append(adj[edge.To], edge.From)
	}

	// Pour chaque paire de nœuds, compter combien de plus courts chemins passent par chaque nœud
	nodeIDs := make([]string, len(graph.Nodes))
	for i, n := range graph.Nodes {
		nodeIDs[i] = n.ID
		betweenness[n.ID] = 0
	}

	for i := 0; i < len(nodeIDs); i++ {
		for j := i + 1; j < len(nodeIDs); j++ {
			path := s.findShortestPath(nodeIDs[i], nodeIDs[j], adj)
			// Ajouter 1 à tous les nœuds intermédiaires du chemin
			for k := 1; k < len(path)-1; k++ {
				betweenness[path[k]]++
			}
		}
	}

	// Normaliser
	maxBetweenness := 0.0
	for _, b := range betweenness {
		if b > maxBetweenness {
			maxBetweenness = b
		}
	}
	if maxBetweenness > 0 {
		for id := range betweenness {
			betweenness[id] = betweenness[id] / maxBetweenness
		}
	}

	return betweenness
}

// bfsDistances calcule les distances depuis un nœud source
func (s *GraphAnalyzerService) bfsDistances(source string, adj map[string][]string) map[string]int {
	distances := make(map[string]int)
	visited := make(map[string]bool)
	queue := []string{source}
	distances[source] = 0
	visited[source] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		for _, neighbor := range adj[current] {
			if !visited[neighbor] {
				visited[neighbor] = true
				distances[neighbor] = distances[current] + 1
				queue = append(queue, neighbor)
			}
		}
	}

	return distances
}

// findShortestPath trouve le plus court chemin entre deux nœuds
func (s *GraphAnalyzerService) findShortestPath(from, to string, adj map[string][]string) []string {
	if from == to {
		return []string{from}
	}

	visited := make(map[string]bool)
	parent := make(map[string]string)
	queue := []string{from}
	visited[from] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if current == to {
			// Reconstruire le chemin
			path := []string{}
			for node := to; node != ""; node = parent[node] {
				path = append([]string{node}, path...)
				if node == from {
					break
				}
			}
			return path
		}

		for _, neighbor := range adj[current] {
			if !visited[neighbor] {
				visited[neighbor] = true
				parent[neighbor] = current
				queue = append(queue, neighbor)
			}
		}
	}

	return []string{}
}

// CalculateSuspicionScores calcule les scores de suspicion pour les personnes
func (s *GraphAnalyzerService) CalculateSuspicionScores(graph models.GraphData, caseData *models.Case) []SuspicionResult {
	results := make([]SuspicionResult, 0)

	// Créer une map ID -> Label
	nodeLabels := make(map[string]string)
	for _, node := range graph.Nodes {
		nodeLabels[node.ID] = node.Label
	}

	// Calculer les connexions
	connections := make(map[string]int)
	for _, edge := range graph.Edges {
		connections[edge.From]++
		connections[edge.To]++
	}

	// Analyser uniquement les personnes (suspects potentiels)
	for _, entity := range caseData.Entities {
		if entity.Type != "personne" {
			continue
		}

		// Ignorer les victimes décédées
		if strings.Contains(strings.ToLower(entity.Description), "décédé") ||
		   strings.Contains(strings.ToLower(entity.Description), "victime") {
			continue
		}

		factors := make([]SuspicionFactor, 0)
		totalScore := 0

		// Facteur 1: Connexions (plus de connexions = plus d'importance)
		connScore := connections[entity.ID]
		if connScore >= 5 {
			factors = append(factors, SuspicionFactor{Name: "Connexions élevées", Value: 20, Level: "medium"})
			totalScore += 20
		} else if connScore >= 3 {
			factors = append(factors, SuspicionFactor{Name: "Connexions modérées", Value: 10, Level: "low"})
			totalScore += 10
		}

		// Facteur 2: Mobile (recherche de mots-clés dans la description)
		desc := strings.ToLower(entity.Description)
		if strings.Contains(desc, "dette") || strings.Contains(desc, "argent") || strings.Contains(desc, "hériti") {
			factors = append(factors, SuspicionFactor{Name: "Mobile financier", Value: 30, Level: "high"})
			totalScore += 30
		}
		if strings.Contains(desc, "conflit") || strings.Contains(desc, "menace") || strings.Contains(desc, "dispute") {
			factors = append(factors, SuspicionFactor{Name: "Conflit connu", Value: 25, Level: "high"})
			totalScore += 25
		}

		// Facteur 3: Accès (présence sur les lieux)
		for _, edge := range graph.Edges {
			if edge.From == entity.ID || edge.To == entity.ID {
				label := strings.ToLower(edge.Label)
				if strings.Contains(label, "accès") || strings.Contains(label, "code") || strings.Contains(label, "clé") {
					factors = append(factors, SuspicionFactor{Name: "Accès aux lieux", Value: 20, Level: "medium"})
					totalScore += 20
					break
				}
			}
		}

		// Facteur 4: Alibi faible ou absent
		hasStrongAlibi := false
		if strings.Contains(desc, "alibi") && strings.Contains(desc, "vérifié") {
			hasStrongAlibi = true
		}
		if !hasStrongAlibi && strings.Contains(desc, "alibi") {
			factors = append(factors, SuspicionFactor{Name: "Alibi non vérifié", Value: 15, Level: "medium"})
			totalScore += 15
		}

		// Facteur 5: Preuves liées
		for _, evidence := range caseData.Evidence {
			for _, linkedID := range evidence.LinkedEntities {
				if linkedID == entity.ID {
					if strings.Contains(strings.ToLower(evidence.Description), "suspect") ||
					   strings.Contains(strings.ToLower(evidence.Description), "empreinte") {
						factors = append(factors, SuspicionFactor{Name: "Preuves liées", Value: 20, Level: "high"})
						totalScore += 20
					}
					break
				}
			}
		}

		// Limiter le score à 100
		if totalScore > 100 {
			totalScore = 100
		}

		// Déterminer le niveau
		level := "low"
		if totalScore >= 60 {
			level = "high"
		} else if totalScore >= 30 {
			level = "medium"
		}

		if len(factors) > 0 || totalScore > 0 {
			results = append(results, SuspicionResult{
				NodeID:    entity.ID,
				NodeLabel: entity.Name,
				Score:     totalScore,
				Factors:   factors,
				Level:     level,
			})
		}
	}

	// Trier par score décroissant
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results
}

// BuildAlibiTimeline construit la timeline des alibis
func (s *GraphAnalyzerService) BuildAlibiTimeline(caseData *models.Case) AlibiTimeline {
	timeline := AlibiTimeline{
		Persons:   make([]AlibiInfo, 0),
		CrimeTime: "20:30", // Heure par défaut, à extraire des données
		TimeRange: TimeRange{Start: "15:00", End: "23:00"},
	}

	// Extraire l'heure du crime depuis les événements
	for _, event := range caseData.Timeline {
		if strings.Contains(strings.ToLower(event.Title), "décès") ||
		   strings.Contains(strings.ToLower(event.Title), "crime") ||
		   strings.Contains(strings.ToLower(event.Title), "meurtre") {
			timeline.CrimeTime = event.Timestamp.Format("15:04")
			break
		}
	}

	// Pour chaque personne, extraire les alibis
	for _, entity := range caseData.Entities {
		if entity.Type != "personne" {
			continue
		}

		// Déterminer le rôle
		role := "Personne d'intérêt"
		desc := strings.ToLower(entity.Description)
		if strings.Contains(desc, "suspect") {
			role = "Suspect"
		} else if strings.Contains(desc, "témoin") {
			role = "Témoin"
		} else if strings.Contains(desc, "victime") {
			role = "Victime"
			continue // Ne pas inclure les victimes
		}

		alibis := make([]AlibiBlock, 0)

		// Chercher les alibis dans les événements
		for _, event := range caseData.Timeline {
			// Vérifier si cet événement concerne cette personne
			for _, entID := range event.Entities {
				if entID == entity.ID {
					// C'est un événement concernant cette personne
					alibi := AlibiBlock{
						StartTime:   event.Timestamp.Format("15:04"),
						Location:    event.Location,
						Verified:    event.Verified,
						Description: event.Title,
					}

					// Estimer l'heure de fin
					if event.EndTime != nil {
						alibi.EndTime = event.EndTime.Format("15:04")
					} else {
						// Ajouter 1 heure par défaut
						alibi.EndTime = event.Timestamp.Add(1 * 60 * 60 * 1000000000).Format("15:04")
					}

					alibis = append(alibis, alibi)
					break
				}
			}
		}

		// Aussi chercher dans la description de l'entité pour les alibis mentionnés
		if strings.Contains(desc, "alibi") {
			// Extraire l'alibi de la description si possible
			if strings.Contains(desc, "cinéma") || strings.Contains(desc, "cinema") {
				alibis = append(alibis, AlibiBlock{
					StartTime:   "19:00",
					EndTime:     "22:00",
					Location:    "Cinéma",
					Verified:    strings.Contains(desc, "ticket") || strings.Contains(desc, "vérifié"),
					Description: "Au cinéma",
				})
			}
		}

		// Vérifier si la personne a une opportunité (pas d'alibi pendant le crime)
		hasOpportunity := true
		for _, alibi := range alibis {
			if alibi.Verified && alibi.StartTime <= timeline.CrimeTime && alibi.EndTime >= timeline.CrimeTime {
				hasOpportunity = false
				break
			}
		}

		if len(alibis) > 0 || hasOpportunity {
			timeline.Persons = append(timeline.Persons, AlibiInfo{
				PersonID:       entity.ID,
				PersonName:     entity.Name,
				PersonRole:     role,
				Alibis:         alibis,
				HasOpportunity: hasOpportunity,
			})
		}
	}

	return timeline
}

// ============================================
// Appointed Nodes - Inspired by SSTorytime
// ============================================

// AppointedNode représente un nœud pointé par plusieurs autres (corrélation)
type AppointedNode struct {
	NodeID       string            `json:"node_id"`
	NodeLabel    string            `json:"node_label"`
	NodeType     string            `json:"node_type"`
	PointedBy    []AppointedSource `json:"pointed_by"`
	PointerCount int               `json:"pointer_count"`
	ArrowTypes   map[string]int    `json:"arrow_types"`   // Type d'arête -> nombre
	Correlation  float64           `json:"correlation"`   // Score de corrélation
	Significance string            `json:"significance"`  // high, medium, low
}

// AppointedSource représente une source qui pointe vers le nœud appointé
type AppointedSource struct {
	NodeID    string `json:"node_id"`
	NodeLabel string `json:"node_label"`
	ArrowType string `json:"arrow_type"`
	Context   string `json:"context,omitempty"`
}

// AppointedNodesResult représente le résultat de la détection de nœuds appointés
type AppointedNodesResult struct {
	Nodes           []AppointedNode `json:"nodes"`
	TotalAppointed  int             `json:"total_appointed"`
	AveragePointers float64         `json:"average_pointers"`
	MaxPointers     int             `json:"max_pointers"`
	Insights        []string        `json:"insights"`
}

// FindAppointedNodes détecte les nœuds pointés par plusieurs autres nœuds
// Ces nœuds créent des corrélations importantes dans le graphe
func (s *GraphAnalyzerService) FindAppointedNodes(graph models.GraphData, minPointers int) AppointedNodesResult {
	if minPointers <= 0 {
		minPointers = 2
	}

	// Créer un map des nœuds
	nodeMap := make(map[string]models.GraphNode)
	for _, node := range graph.Nodes {
		nodeMap[node.ID] = node
	}

	// Compter les arêtes entrantes par nœud et par type
	incomingEdges := make(map[string][]AppointedSource)
	arrowTypeCount := make(map[string]map[string]int)

	for _, edge := range graph.Edges {
		// Ajouter la source
		sourceNode := nodeMap[edge.From]
		incomingEdges[edge.To] = append(incomingEdges[edge.To], AppointedSource{
			NodeID:    edge.From,
			NodeLabel: sourceNode.Label,
			ArrowType: edge.Label,
			Context:   edge.Context,
		})

		// Compter par type d'arête
		if arrowTypeCount[edge.To] == nil {
			arrowTypeCount[edge.To] = make(map[string]int)
		}
		arrowTypeCount[edge.To][edge.Label]++
	}

	// Filtrer les nœuds avec au moins minPointers sources distinctes
	var appointedNodes []AppointedNode
	maxPointers := 0
	totalPointers := 0

	for nodeID, sources := range incomingEdges {
		// Compter les sources distinctes
		distinctSources := make(map[string]bool)
		for _, src := range sources {
			distinctSources[src.NodeID] = true
		}

		pointerCount := len(distinctSources)
		if pointerCount >= minPointers {
			node := nodeMap[nodeID]

			// Calculer le score de corrélation
			// Plus il y a de sources et de types d'arêtes différents, plus c'est significatif
			arrowDiversity := len(arrowTypeCount[nodeID])
			correlation := float64(pointerCount) * (1.0 + float64(arrowDiversity)*0.1)

			// Déterminer la significativité
			significance := "low"
			if pointerCount >= 5 || correlation >= 5.0 {
				significance = "high"
			} else if pointerCount >= 3 || correlation >= 3.0 {
				significance = "medium"
			}

			appointedNodes = append(appointedNodes, AppointedNode{
				NodeID:       nodeID,
				NodeLabel:    node.Label,
				NodeType:     node.Type,
				PointedBy:    sources,
				PointerCount: pointerCount,
				ArrowTypes:   arrowTypeCount[nodeID],
				Correlation:  correlation,
				Significance: significance,
			})

			if pointerCount > maxPointers {
				maxPointers = pointerCount
			}
			totalPointers += pointerCount
		}
	}

	// Trier par nombre de pointeurs décroissant
	sort.Slice(appointedNodes, func(i, j int) bool {
		return appointedNodes[i].PointerCount > appointedNodes[j].PointerCount
	})

	// Calculer la moyenne
	avgPointers := 0.0
	if len(appointedNodes) > 0 {
		avgPointers = float64(totalPointers) / float64(len(appointedNodes))
	}

	// Générer des insights
	insights := s.generateAppointedInsights(appointedNodes, graph)

	return AppointedNodesResult{
		Nodes:           appointedNodes,
		TotalAppointed:  len(appointedNodes),
		AveragePointers: avgPointers,
		MaxPointers:     maxPointers,
		Insights:        insights,
	}
}

// generateAppointedInsights génère des insights sur les nœuds appointés
func (s *GraphAnalyzerService) generateAppointedInsights(nodes []AppointedNode, graph models.GraphData) []string {
	var insights []string

	if len(nodes) == 0 {
		insights = append(insights, "Aucun nœud de corrélation détecté. Le graphe est peut-être trop fragmenté.")
		return insights
	}

	// Analyser le nœud le plus appointé
	if len(nodes) > 0 {
		top := nodes[0]
		insights = append(insights, fmt.Sprintf("'%s' est le nœud le plus central avec %d sources distinctes.", top.NodeLabel, top.PointerCount))

		// Analyser les types d'arêtes
		if len(top.ArrowTypes) > 1 {
			insights = append(insights, fmt.Sprintf("Ce nœud reçoit %d types de relations différentes, indiquant un rôle multifonctionnel.", len(top.ArrowTypes)))
		}
	}

	// Compter par significativité
	highCount := 0
	for _, n := range nodes {
		if n.Significance == "high" {
			highCount++
		}
	}

	if highCount > 0 {
		insights = append(insights, fmt.Sprintf("%d nœud(s) hautement significatif(s) détecté(s) - points de convergence importants.", highCount))
	}

	// Analyser par type
	typeCount := make(map[string]int)
	for _, n := range nodes {
		typeCount[n.NodeType]++
	}
	for nodeType, count := range typeCount {
		if count >= 2 {
			insights = append(insights, fmt.Sprintf("Les éléments de type '%s' sont fréquemment corrélés (%d occurrences).", nodeType, count))
		}
	}

	return insights
}

// ============================================
// Eigenvector Centrality - Inspired by SSTorytime
// ============================================

// EigenvectorCentralityResult représente le résultat du calcul de centralité eigenvector
type EigenvectorCentralityResult struct {
	NodeID     string  `json:"node_id"`
	NodeLabel  string  `json:"node_label"`
	NodeType   string  `json:"node_type"`
	Score      float64 `json:"score"`
	Rank       int     `json:"rank"`
	Influence  string  `json:"influence"` // high, medium, low
}

// EigenvectorResult représente le résultat global
type EigenvectorResult struct {
	Centralities []EigenvectorCentralityResult `json:"centralities"`
	Convergence  bool                          `json:"convergence"`
	Iterations   int                           `json:"iterations"`
	Insights     []string                      `json:"insights"`
}

// CalculateEigenvectorCentrality calcule la centralité eigenvector pour tous les nœuds
// Cette méthode mesure l'influence d'un nœud basée sur l'influence de ses voisins
func (s *GraphAnalyzerService) CalculateEigenvectorCentrality(graph models.GraphData, maxIterations int) EigenvectorResult {
	if maxIterations <= 0 {
		maxIterations = 100
	}

	n := len(graph.Nodes)
	if n == 0 {
		return EigenvectorResult{Centralities: []EigenvectorCentralityResult{}}
	}

	// Créer un index des nœuds
	nodeIndex := make(map[string]int)
	indexNode := make(map[int]string)
	nodeMap := make(map[string]models.GraphNode)

	for i, node := range graph.Nodes {
		nodeIndex[node.ID] = i
		indexNode[i] = node.ID
		nodeMap[node.ID] = node
	}

	// Construire la matrice d'adjacence (symétrique pour eigenvector)
	adj := make([][]float64, n)
	for i := range adj {
		adj[i] = make([]float64, n)
	}

	for _, edge := range graph.Edges {
		i, okFrom := nodeIndex[edge.From]
		j, okTo := nodeIndex[edge.To]
		if okFrom && okTo {
			adj[i][j] = 1.0
			adj[j][i] = 1.0 // Symétrique
		}
	}

	// Initialiser le vecteur propre avec des valeurs uniformes
	eigenvector := make([]float64, n)
	for i := range eigenvector {
		eigenvector[i] = 1.0 / float64(n)
	}

	// Itération de puissance pour trouver le vecteur propre dominant
	const tolerance = 1e-6
	converged := false
	iterations := 0

	for iter := 0; iter < maxIterations; iter++ {
		iterations = iter + 1
		newVector := make([]float64, n)

		// Multiplier par la matrice d'adjacence
		for i := 0; i < n; i++ {
			sum := 0.0
			for j := 0; j < n; j++ {
				sum += adj[i][j] * eigenvector[j]
			}
			newVector[i] = sum
		}

		// Normaliser
		norm := 0.0
		for _, v := range newVector {
			norm += v * v
		}
		norm = math.Sqrt(norm)

		if norm > 0 {
			for i := range newVector {
				newVector[i] /= norm
			}
		}

		// Vérifier la convergence
		maxDiff := 0.0
		for i := range eigenvector {
			diff := math.Abs(newVector[i] - eigenvector[i])
			if diff > maxDiff {
				maxDiff = diff
			}
		}

		eigenvector = newVector

		if maxDiff < tolerance {
			converged = true
			break
		}
	}

	// Normaliser entre 0 et 1
	maxScore := 0.0
	for _, v := range eigenvector {
		if v > maxScore {
			maxScore = v
		}
	}

	results := make([]EigenvectorCentralityResult, n)
	for i := 0; i < n; i++ {
		nodeID := indexNode[i]
		node := nodeMap[nodeID]

		score := 0.0
		if maxScore > 0 {
			score = eigenvector[i] / maxScore
		}

		influence := "low"
		if score >= 0.7 {
			influence = "high"
		} else if score >= 0.4 {
			influence = "medium"
		}

		results[i] = EigenvectorCentralityResult{
			NodeID:    nodeID,
			NodeLabel: node.Label,
			NodeType:  node.Type,
			Score:     score,
			Influence: influence,
		}
	}

	// Trier par score décroissant
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// Assigner les rangs
	for i := range results {
		results[i].Rank = i + 1
	}

	// Générer des insights
	insights := s.generateEigenvectorInsights(results, converged, iterations)

	return EigenvectorResult{
		Centralities: results,
		Convergence:  converged,
		Iterations:   iterations,
		Insights:     insights,
	}
}

// generateEigenvectorInsights génère des insights sur la centralité eigenvector
func (s *GraphAnalyzerService) generateEigenvectorInsights(results []EigenvectorCentralityResult, converged bool, iterations int) []string {
	var insights []string

	if !converged {
		insights = append(insights, fmt.Sprintf("L'algorithme n'a pas convergé après %d itérations. Les résultats sont approximatifs.", iterations))
	}

	if len(results) == 0 {
		return insights
	}

	// Analyser le top 3
	highInfluence := 0
	for _, r := range results {
		if r.Influence == "high" {
			highInfluence++
		}
	}

	if highInfluence > 0 {
		insights = append(insights, fmt.Sprintf("%d nœud(s) avec une influence élevée détecté(s).", highInfluence))
	}

	// Top node
	if len(results) > 0 && results[0].Score > 0.5 {
		insights = append(insights, fmt.Sprintf("'%s' est le nœud le plus influent (score: %.2f). Il est connecté à d'autres nœuds influents.", results[0].NodeLabel, results[0].Score))
	}

	// Analyser la distribution
	if len(results) >= 3 {
		top3Avg := (results[0].Score + results[1].Score + results[2].Score) / 3
		if top3Avg > 0.7 {
			insights = append(insights, "Le graphe a une structure centralisée avec quelques nœuds dominants.")
		} else if top3Avg < 0.4 {
			insights = append(insights, "Le graphe a une structure distribuée sans nœud dominant clair.")
		}
	}

	// Comparer degree et eigenvector
	// (Les nœuds avec haute eigenvector mais faible degree sont des "brokers")
	if len(results) >= 5 {
		// Vérifier si le top eigenvector inclut des nœuds de type différent
		types := make(map[string]bool)
		for i := 0; i < 5 && i < len(results); i++ {
			types[results[i].NodeType] = true
		}
		if len(types) >= 3 {
			insights = append(insights, "Les nœuds influents sont de types variés, suggérant un graphe bien intégré.")
		}
	}

	return insights
}
