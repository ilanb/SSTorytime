package data

import (
	"time"

	"forensicinvestigator/internal/models"
)

// GetDemoCases retourne un ensemble d'affaires de démonstration
func GetDemoCases() []*models.Case {
	return []*models.Case{
		createAffaireMoreau(),
		createAffaireDisparition(),
		createAffaireFraude(),
		createAffaireCambriolage(),
		createAffaireIncendie(),
		createAffaireTraficArt(),
	}
}

// Affaire 1: Homicide - Affaire Victor Moreau (basé sur enquete.n4l)
// Dataset enrichi pour démontrer les capacités N4L complètes de SSTorytime
func createAffaireMoreau() *models.Case {
	return &models.Case{
		ID:          "case-moreau-001",
		Name:        "Affaire Victor Moreau",
		Description: "Homicide par empoisonnement d'un antiquaire renommé dans son manoir parisien. Suspicions multiples sur le neveu héritier et une ex-associée en conflit judiciaire. Réseau complexe de relations financières et personnelles.",
		Type:        "homicide",
		Status:      "en_cours",
		CreatedAt:   parseDate("2025-08-30"),
		UpdatedAt:   parseDate("2025-09-01"),
		Entities: []models.Entity{
			// Victime
			{
				ID:          "ent-moreau-001",
				CaseID:      "case-moreau-001",
				Name:        "Victor Moreau",
				Type:        models.EntityPerson,
				Role:        models.RoleVictim,
				Description: "Antiquaire renommé, 67 ans, veuf depuis 2019. Fortune estimée à 12 millions d'euros. Décédé le 29/08/2025 vers 20h30 par empoisonnement.",
				Attributes: map[string]string{
					"age":             "67 ans",
					"profession":      "Antiquaire",
					"domicile":        "Manoir rue de Varenne, Paris 7e",
					"fortune":         "12 millions d'euros",
					"cause_deces":     "Empoisonnement (alcaloïde)",
					"date_deces":      "29/08/2025",
					"heure_deces":     "20h30",
					"membre_de":       "Cercle des Bibliophiles Parisiens",
					"collectionneur":  "Livres rares XVIIIe siècle",
					"etat_civil":      "Veuf depuis 2019",
					"latitude":        "48.8566",
					"longitude":       "2.3175",
				},
				Relations: []models.Relation{
					{ID: "rel-001", FromID: "ent-moreau-001", ToID: "ent-moreau-002", Type: "famille", Label: "oncle de", Context: "famille", Verified: true},
					{ID: "rel-001b", FromID: "ent-moreau-001", ToID: "ent-moreau-002", Type: "argent", Label: "transfère de l'argent à", Context: "finances", Verified: true},
					{ID: "rel-002", FromID: "ent-moreau-001", ToID: "ent-moreau-003", Type: "conflit", Label: "en conflit avec", Context: "affaires", Verified: true},
					{ID: "rel-002b", FromID: "ent-moreau-003", ToID: "ent-moreau-001", Type: "menace", Label: "a menacé", Context: "intimidation", Verified: true},
					{ID: "rel-003", FromID: "ent-moreau-001", ToID: "ent-moreau-004", Type: "emploi", Label: "emploie et paie", Context: "travail", Verified: true},
					{ID: "rel-003b", FromID: "ent-moreau-001", ToID: "ent-moreau-005", Type: "emploi", Label: "emploie et paie", Context: "travail", Verified: true},
					{ID: "rel-003c", FromID: "ent-moreau-001", ToID: "ent-moreau-007", Type: "propriete", Label: "propriétaire de", Context: "immobilier", Verified: true},
					{ID: "rel-003d", FromID: "ent-moreau-001", ToID: "ent-moreau-008", Type: "propriete", Label: "contrôle et dirige", Context: "affaires", Verified: true},
					{ID: "rel-003e", FromID: "ent-moreau-001", ToID: "ent-moreau-009", Type: "membre", Label: "membre de", Context: "social", Verified: true},
					{ID: "rel-003f", FromID: "ent-moreau-001", ToID: "ent-moreau-010", Type: "relation", Label: "communique régulièrement avec", Context: "social", Verified: true},
				},
			},
			// Suspect 1 - Neveu
			{
				ID:          "ent-moreau-002",
				CaseID:      "case-moreau-001",
				Name:        "Jean Moreau",
				Type:        models.EntityPerson,
				Role:        models.RoleSuspect,
				Description: "Neveu de la victime, 35 ans, sans emploi fixe. Dettes de jeu importantes (150 000€). Héritier principal (8 millions). Alibi: cinéma UGC Bercy 19h-22h.",
				Attributes: map[string]string{
					"age":             "35 ans",
					"profession":      "Sans emploi",
					"mobile":          "Héritage 8 millions",
					"dettes":          "150 000€ au Casino de Deauville",
					"alibi":           "Cinéma UGC Bercy 19h-22h",
					"alibi_verifie":   "Partiellement",
					"vehicule":        "BMW série 3 noire AB-123-CD",
					"dernier_appel":   "29/08 à 18h45 vers Victor",
					"pointure":        "42",
					"comportement":    "Nerveux lors interrogatoire",
					"latitude":        "48.8396",
					"longitude":       "2.3876",
				},
				Relations: []models.Relation{
					{ID: "rel-004", FromID: "ent-moreau-002", ToID: "ent-moreau-001", Type: "famille", Label: "neveu de", Context: "famille", Verified: true},
					{ID: "rel-004b", FromID: "ent-moreau-002", ToID: "ent-moreau-001", Type: "communication", Label: "appel téléphonique vers", Context: "information", Verified: true},
					{ID: "rel-005", FromID: "ent-moreau-002", ToID: "ent-moreau-006", Type: "dette", Label: "doit de l'argent à", Context: "finances", Verified: true},
					{ID: "rel-005a", FromID: "ent-moreau-006", ToID: "ent-moreau-002", Type: "pression", Label: "menace pour recouvrement", Context: "intimidation", Verified: true},
					{ID: "rel-005b", FromID: "ent-moreau-002", ToID: "ent-moreau-011", Type: "frequentation", Label: "rencontre fréquemment à", Context: "loisirs", Verified: true},
					{ID: "rel-005c", FromID: "ent-moreau-002", ToID: "ent-moreau-007", Type: "connaissance", Label: "connaît le code d'accès de", Context: "accès", Verified: true},
					{ID: "rel-005d", FromID: "ent-moreau-002", ToID: "ent-moreau-012", Type: "contact", Label: "a envoyé un message à", Context: "suspect", Verified: false},
					{ID: "rel-005e", FromID: "ent-moreau-002", ToID: "ent-moreau-003", Type: "communication", Label: "a eu un appel avec", Context: "suspect", Verified: false},
				},
			},
			// Suspect 2 - Ex-associée
			{
				ID:          "ent-moreau-003",
				CaseID:      "case-moreau-001",
				Name:        "Élodie Dubois",
				Type:        models.EntityPerson,
				Role:        models.RoleSuspect,
				Description: "Ex-associée de Victor, 42 ans. Perte financière de 2 millions en 2024. Procédure judiciaire en cours. Menaces proférées: 'Il paiera pour ce qu'il m'a fait'.",
				Attributes: map[string]string{
					"age":                "42 ans",
					"profession":         "Femme d'affaires",
					"mobile":             "Vengeance - perte 2 millions",
					"alibi":              "Dîner charité Hôtel Crillon",
					"procedure":          "Procès contre Victor en cours",
					"menaces":            "Il paiera pour ce qu'il m'a fait",
					"avocat":             "Maître Lefebvre",
					"latitude":           "48.8686",
					"longitude":          "2.3215",
				},
				Relations: []models.Relation{
					{ID: "rel-006", FromID: "ent-moreau-003", ToID: "ent-moreau-001", Type: "conflit", Label: "accuse de fraude", Context: "judiciaire", Verified: true},
					{ID: "rel-006b", FromID: "ent-moreau-003", ToID: "ent-moreau-001", Type: "menace", Label: "a menacé publiquement", Context: "intimidation", Verified: true},
					{ID: "rel-006c", FromID: "ent-moreau-001", ToID: "ent-moreau-003", Type: "argent", Label: "aurait détourné de l'argent de", Context: "finances", Verified: false},
					{ID: "rel-006d", FromID: "ent-moreau-003", ToID: "ent-moreau-014", Type: "emploi", Label: "emploie comme avocat", Context: "judiciaire", Verified: true},
				},
			},
			// Témoin - Gouvernante
			{
				ID:          "ent-moreau-004",
				CaseID:      "case-moreau-001",
				Name:        "Madame Chen",
				Type:        models.EntityPerson,
				Role:        models.RoleWitness,
				Description: "Gouvernante du manoir depuis 15 ans. Présente le soir du crime jusqu'à 19h. A servi le thé à 19h15. A découvert le corps à 21h45.",
				Attributes: map[string]string{
					"age":         "45 ans",
					"profession":  "Gouvernante",
					"anciennete":  "15 ans",
					"dernier_the": "19h15 le 29/08",
					"observation": "Victor semblait nerveux",
					"latitude":    "48.8566",
					"longitude":   "2.3175",
				},
				Relations: []models.Relation{
					{ID: "rel-007", FromID: "ent-moreau-004", ToID: "ent-moreau-001", Type: "emploi", Label: "employée de", Context: "travail", Verified: true},
				},
			},
			// Témoin - Jardinier
			{
				ID:          "ent-moreau-005",
				CaseID:      "case-moreau-001",
				Name:        "Robert Duval",
				Type:        models.EntityPerson,
				Role:        models.RoleWitness,
				Description: "Jardinier. Présent jusqu'à 18h. A observé la fenêtre bibliothèque ouverte à 17h30 et des traces de pas inhabituelles près des rosiers.",
				Attributes: map[string]string{
					"age":         "58 ans",
					"profession":  "Jardinier",
					"observation": "Traces pas inhabituelles près rosiers",
					"latitude":    "48.8570",
					"longitude":   "2.3180",
				},
			},
			// Lieu - Casino
			{
				ID:          "ent-moreau-006",
				CaseID:      "case-moreau-001",
				Name:        "Casino de Deauville",
				Type:        models.EntityPlace,
				Role:        models.RoleOther,
				Description: "Casino où Jean Moreau a contracté des dettes importantes.",
				Attributes: map[string]string{
					"adresse":   "2 Rue Edmond Blanc, 14800 Deauville",
					"latitude":  "49.3565",
					"longitude": "-0.0742",
					"type":      "Casino",
				},
			},
			// Lieu - Scène de crime
			{
				ID:          "ent-moreau-007",
				CaseID:      "case-moreau-001",
				Name:        "Bibliothèque du Manoir",
				Type:        models.EntityPlace,
				Role:        models.RoleOther,
				Description: "Scène de crime principale. RDC, 8x6m. Fenêtre ouest ouverte. Corps trouvé dans fauteuil près de la cheminée.",
				Attributes: map[string]string{
					"etage":        "Rez-de-chaussée",
					"dimensions":   "8m x 6m",
					"acces":        "Porte principale + porte-fenêtre jardin",
					"etat_fenetre": "Ouverte, empreintes essuyées",
					"adresse":      "Manoir Moreau, Rue de Varenne, Paris 7e",
					"latitude":     "48.8556",
					"longitude":    "2.3177",
				},
				Relations: []models.Relation{
					{ID: "rel-007a", FromID: "ent-moreau-007", ToID: "ent-moreau-013", Type: "contient", Label: "connectée à", Context: "architecture", Verified: true},
				},
			},
			// Nouvelles entités pour enrichir le réseau N4L
			// Lieu - Galerie
			{
				ID:          "ent-moreau-008",
				CaseID:      "case-moreau-001",
				Name:        "Galerie Moreau Antiquités",
				Type:        models.EntityPlace,
				Role:        models.RoleOther,
				Description: "Galerie d'antiquités appartenant à Victor Moreau, située rue du Faubourg Saint-Honoré.",
				Attributes: map[string]string{
					"adresse":      "12 rue du Faubourg Saint-Honoré, Paris 8e",
					"specialite":   "Livres rares et manuscrits",
					"valeur_stock": "3.5 millions d'euros",
					"latitude":     "48.8699",
					"longitude":    "2.3189",
				},
				Relations: []models.Relation{
					{ID: "rel-008a", FromID: "ent-moreau-008", ToID: "ent-moreau-010", Type: "concurrent", Label: "en concurrence avec", Context: "affaires", Verified: true},
				},
			},
			// Organisation - Club
			{
				ID:          "ent-moreau-009",
				CaseID:      "case-moreau-001",
				Name:        "Cercle des Bibliophiles Parisiens",
				Type:        models.EntityOrg,
				Role:        models.RoleOther,
				Description: "Club exclusif de collectionneurs de livres anciens. Victor en était membre depuis 20 ans.",
				Relations: []models.Relation{
					{ID: "rel-009a", FromID: "ent-moreau-009", ToID: "ent-moreau-010", Type: "membre", Label: "compte comme membre", Context: "social", Verified: true},
				},
			},
			// Personne - Concurrent
			{
				ID:          "ent-moreau-010",
				CaseID:      "case-moreau-001",
				Name:        "Antoine Mercier",
				Type:        models.EntityPerson,
				Role:        models.RoleSuspect,
				Description: "Expert en livres anciens et concurrent de Victor. Conflit sur vente aux enchères en juillet 2025.",
				Attributes: map[string]string{
					"age":        "55 ans",
					"profession": "Expert en livres anciens",
					"galerie":    "Mercier & Fils",
					"mobile":     "Rivalité commerciale",
					"alibi":      "Non vérifié",
				},
				Relations: []models.Relation{
					{ID: "rel-010a", FromID: "ent-moreau-010", ToID: "ent-moreau-001", Type: "conflit", Label: "en rivalité avec", Context: "affaires", Verified: true},
					{ID: "rel-010b", FromID: "ent-moreau-010", ToID: "ent-moreau-009", Type: "membre", Label: "membre de", Context: "social", Verified: true},
				},
			},
			// Lieu - Bar
			{
				ID:          "ent-moreau-011",
				CaseID:      "case-moreau-001",
				Name:        "Bar Le Diplomate",
				Type:        models.EntityPlace,
				Role:        models.RoleOther,
				Description: "Bar fréquenté par Jean Moreau. Lieu de rencontres suspects.",
				Attributes: map[string]string{
					"adresse":   "45 rue de Bercy, Paris 12e",
					"type":      "Bar de nuit",
					"latitude":  "48.8387",
					"longitude": "2.3826",
				},
			},
			// Personne - Contact mystérieux
			{
				ID:          "ent-moreau-012",
				CaseID:      "case-moreau-001",
				Name:        "Homme au téléphone inconnu",
				Type:        models.EntityPerson,
				Role:        models.RoleSuspect,
				Description: "Personne non identifiée ayant appelé Victor à 19h05 le soir du crime. Durée: 2 minutes.",
				Attributes: map[string]string{
					"numero":       "Numéro masqué",
					"heure_appel":  "19h05",
					"duree":        "2 minutes",
					"identification": "En cours",
				},
				Relations: []models.Relation{
					{ID: "rel-012a", FromID: "ent-moreau-012", ToID: "ent-moreau-001", Type: "contact", Label: "a appelé", Context: "suspect", Verified: true},
				},
			},
			// Lieu - Jardin
			{
				ID:          "ent-moreau-013",
				CaseID:      "case-moreau-001",
				Name:        "Jardin du Manoir",
				Type:        models.EntityPlace,
				Role:        models.RoleOther,
				Description: "Jardin de 500m² avec haie haute côté rue. Traces de pas près des rosiers.",
				Attributes: map[string]string{
					"superficie":    "500m²",
					"particularite": "Haie haute côté rue",
					"acces":         "Portillon avec clé",
					"indices":       "Traces de pas, terre argileuse",
					"adresse":       "Manoir Moreau, Rue de Varenne, Paris 7e",
					"latitude":      "48.8557",
					"longitude":     "2.3175",
				},
				Relations: []models.Relation{
					{ID: "rel-013a", FromID: "ent-moreau-013", ToID: "ent-moreau-007", Type: "acces", Label: "donne accès à", Context: "architecture", Verified: true},
				},
			},
			// Lieu - Cinéma UGC Bercy (alibi Jean)
			{
				ID:          "ent-moreau-020",
				CaseID:      "case-moreau-001",
				Name:        "UGC Bercy",
				Type:        models.EntityPlace,
				Role:        models.RoleOther,
				Description: "Cinéma où Jean Moreau prétend avoir passé la soirée du crime (19h-22h).",
				Attributes: map[string]string{
					"adresse":   "2 Cour Saint-Émilion, Paris 12e",
					"latitude":  "48.8335",
					"longitude": "2.3867",
					"type":      "Cinéma",
				},
				Relations: []models.Relation{
					{ID: "rel-020a", FromID: "ent-moreau-020", ToID: "ent-moreau-002", Type: "alibi", Label: "lieu d'alibi de", Context: "investigation", Verified: true},
				},
			},
			// Lieu - Hôtel Crillon (alibi Élodie)
			{
				ID:          "ent-moreau-021",
				CaseID:      "case-moreau-001",
				Name:        "Hôtel de Crillon",
				Type:        models.EntityPlace,
				Role:        models.RoleOther,
				Description: "Palace parisien où Élodie Dubois assistait à un dîner de charité le soir du crime.",
				Attributes: map[string]string{
					"adresse":   "10 Place de la Concorde, Paris 8e",
					"latitude":  "48.8677",
					"longitude": "2.3216",
					"type":      "Hôtel de luxe",
				},
				Relations: []models.Relation{
					{ID: "rel-021a", FromID: "ent-moreau-021", ToID: "ent-moreau-003", Type: "alibi", Label: "lieu d'alibi de", Context: "investigation", Verified: true},
				},
			},
			// Objet - Testament
			{
				ID:          "ent-moreau-014",
				CaseID:      "case-moreau-001",
				Name:        "Testament de Victor Moreau",
				Type:        models.EntityDocument,
				Role:        models.RoleOther,
				Description: "Testament original daté du 15/03/2025. Jean héritier à 80%, œuvres caritatives 20%.",
				Attributes: map[string]string{
					"date_redaction": "15/03/2025",
					"notaire":        "Maître Durand",
					"beneficiaire":   "Jean Moreau 80%",
					"statut":         "Valide",
				},
				Relations: []models.Relation{
					{ID: "rel-014a", FromID: "ent-moreau-014", ToID: "ent-moreau-002", Type: "benefice", Label: "désigne comme héritier", Context: "succession", Verified: true},
				},
			},
			// Objet - Brouillon testament
			{
				ID:          "ent-moreau-015",
				CaseID:      "case-moreau-001",
				Name:        "Brouillon nouveau testament",
				Type:        models.EntityDocument,
				Role:        models.RoleOther,
				Description: "Brouillon non signé daté du 28/08. Jean réduit à 10%, Fondation Moreau 90%.",
				Attributes: map[string]string{
					"date_redaction":     "28/08/2025",
					"statut":             "Non signé",
					"nouveau_beneficiaire": "Fondation Moreau 90%",
					"lieu_decouverte":     "Corbeille bureau",
				},
				Relations: []models.Relation{
					{ID: "rel-015a", FromID: "ent-moreau-015", ToID: "ent-moreau-014", Type: "remplace", Label: "devait remplacer", Context: "succession", Verified: true},
					{ID: "rel-015b", FromID: "ent-moreau-015", ToID: "ent-moreau-002", Type: "prejudice", Label: "déshérite partiellement", Context: "mobile", Verified: true},
				},
			},
			// Personne - Avocat
			{
				ID:          "ent-moreau-016",
				CaseID:      "case-moreau-001",
				Name:        "Maître Lefebvre",
				Type:        models.EntityPerson,
				Role:        models.RoleWitness,
				Description: "Avocat d'Élodie Dubois. Agressif mais efficace. A demandé une saisie conservatoire.",
				Attributes: map[string]string{
					"profession":  "Avocat",
					"specialite":  "Droit des affaires",
					"reputation":  "Agressif mais efficace",
				},
				Relations: []models.Relation{
					{ID: "rel-016a", FromID: "ent-moreau-016", ToID: "ent-moreau-003", Type: "representation", Label: "représente", Context: "juridique", Verified: true},
					{ID: "rel-016b", FromID: "ent-moreau-016", ToID: "ent-moreau-001", Type: "opposition", Label: "s'opposait à", Context: "juridique", Verified: true},
				},
			},
			// Personne - Médecin légiste
			{
				ID:          "ent-moreau-017",
				CaseID:      "case-moreau-001",
				Name:        "Dr. Sarah Martin",
				Type:        models.EntityPerson,
				Role:        models.RoleWitness,
				Description: "Médecin légiste. Rapport préliminaire: décès entre 20h et 21h par alcaloïde non identifié.",
				Attributes: map[string]string{
					"profession":      "Médecin légiste",
					"rapport":         "Décès 20h-21h",
					"cause":           "Alcaloïde végétal",
					"remarque":        "Aucun signe de lutte",
				},
				Relations: []models.Relation{
					{ID: "rel-017a", FromID: "ent-moreau-017", ToID: "ent-moreau-001", Type: "expertise", Label: "a examiné", Context: "médico-légal", Verified: true},
				},
			},
		},
		Evidence: []models.Evidence{
			{
				ID:          "ev-moreau-001",
				CaseID:      "case-moreau-001",
				Name:        "Tasse de thé",
				Type:        models.EvidencePhysical,
				Description: "Résidus d'Earl Grey et substance inconnue. Empreintes de Victor uniquement.",
				Location:    "Table basse bibliothèque",
				Reliability: 9,
				LinkedEntities: []string{"ent-moreau-001", "ent-moreau-007"},
			},
			{
				ID:          "ev-moreau-002",
				CaseID:      "case-moreau-001",
				Name:        "Livre 'Traité des Poisons Exotiques'",
				Type:        models.EvidencePhysical,
				Description: "Édition 1923. Ouvert page 247 (Alcaloïdes végétaux). Annotations récentes. Empreintes partielles non identifiées.",
				Location:    "Bureau de Victor",
				Reliability: 7,
				LinkedEntities: []string{"ent-moreau-001"},
			},
			{
				ID:          "ev-moreau-003",
				CaseID:      "case-moreau-001",
				Name:        "Traces de boue sur tapis",
				Type:        models.EvidenceForensic,
				Description: "Terre argileuse des rosiers. Empreintes pointure 42. Direction: fenêtre vers fauteuil. Fraîcheur < 24h.",
				Location:    "Tapis persan bibliothèque",
				Reliability: 8,
				LinkedEntities: []string{"ent-moreau-007"},
			},
			{
				ID:          "ev-moreau-004",
				CaseID:      "case-moreau-001",
				Name:        "Téléphone de Victor",
				Type:        models.EvidenceDigital,
				Description: "Derniers appels: Jean 18h45, numéro inconnu 19h05. SMS menaçants d'Élodie. Recherches sur poisons le 28/08.",
				Location:    "Sur la victime",
				Reliability: 9,
				LinkedEntities: []string{"ent-moreau-001", "ent-moreau-002", "ent-moreau-003"},
			},
			{
				ID:          "ev-moreau-005",
				CaseID:      "case-moreau-001",
				Name:        "Brouillon nouveau testament",
				Type:        models.EvidenceDocumentary,
				Description: "Non signé, daté 28/08. Jean réduit à 10%, Fondation Moreau 90%. Trouvé dans corbeille.",
				Location:    "Bureau premier étage",
				Reliability: 8,
				LinkedEntities: []string{"ent-moreau-001", "ent-moreau-002"},
			},
			{
				ID:          "ev-moreau-006",
				CaseID:      "case-moreau-001",
				Name:        "Carnet de notes Victor",
				Type:        models.EvidenceDocumentary,
				Description: "Dernière entrée 28/08: 'E.D. devient dangereuse'",
				Location:    "Tiroir bureau",
				Reliability: 7,
				LinkedEntities: []string{"ent-moreau-001", "ent-moreau-003"},
			},
			{
				ID:          "ev-moreau-007",
				CaseID:      "case-moreau-001",
				Name:        "Câble caméra sectionné",
				Type:        models.EvidenceDigital,
				Description: "Caméra surveillance dysfonctionnelle depuis 28/08 18h. Dernière image: silhouette non identifiable.",
				Location:    "Système surveillance manoir",
				Reliability: 6,
			},
		},
		Timeline: []models.Event{
			// Événements antérieurs (contexte)
			{ID: "evt-m-00a", CaseID: "case-moreau-001", Title: "Vente aux enchères contestée", Description: "Victor remporte un manuscrit convoité par Mercier pour 450 000€", Timestamp: parseDateTime("2025-07-15T14:00:00"), Location: "Drouot", Entities: []string{"ent-moreau-001", "ent-moreau-010"}, Importance: "medium", Verified: true},
			{ID: "evt-m-00b", CaseID: "case-moreau-001", Title: "Confrontation tribunal Victor/Élodie", Description: "Élodie menace Victor: 'Il paiera pour ce qu'il m'a fait'", Timestamp: parseDateTime("2025-08-25T10:00:00"), Location: "Tribunal de Commerce", Entities: []string{"ent-moreau-001", "ent-moreau-003", "ent-moreau-016"}, Evidence: []string{"ev-moreau-006"}, Importance: "high", Verified: true},
			// Semaine du crime
			{ID: "evt-m-01", CaseID: "case-moreau-001", Title: "Visite houleuse de Jean", Description: "Discussion avec Victor sur l'argent. Claque la porte en partant.", Timestamp: parseDateTime("2025-08-27T14:00:00"), Location: "Manoir", Entities: []string{"ent-moreau-001", "ent-moreau-002"}, Importance: "high", Verified: true},
			{ID: "evt-m-01b", CaseID: "case-moreau-001", Title: "Jean vu au Bar Le Diplomate", Description: "Rencontre avec individu non identifié", Timestamp: parseDateTime("2025-08-27T22:00:00"), Location: "Bar Le Diplomate", Entities: []string{"ent-moreau-002", "ent-moreau-011"}, Importance: "medium", Verified: false},
			{ID: "evt-m-02", CaseID: "case-moreau-001", Title: "Victor chez le notaire", Description: "Évoque une modification du testament - veut déshériter Jean", Timestamp: parseDateTime("2025-08-28T10:00:00"), Location: "Étude notariale", Entities: []string{"ent-moreau-001", "ent-moreau-014"}, Evidence: []string{"ev-moreau-005"}, Importance: "high", Verified: true},
			{ID: "evt-m-02b", CaseID: "case-moreau-001", Title: "Câble caméra sectionné", Description: "Système de surveillance neutralisé", Timestamp: parseDateTime("2025-08-28T18:00:00"), Location: "Manoir", Entities: []string{}, Evidence: []string{"ev-moreau-007"}, Importance: "high", Verified: true},
			// Jour du crime
			{ID: "evt-m-03", CaseID: "case-moreau-001", Title: "Élodie vue près du manoir", Description: "Aperçue par voisin M. Bertrand à 15h", Timestamp: parseDateTime("2025-08-29T15:00:00"), Location: "Aux abords du manoir", Entities: []string{"ent-moreau-003"}, Importance: "high", Verified: false},
			{ID: "evt-m-04", CaseID: "case-moreau-001", Title: "Fenêtre bibliothèque ouverte", Description: "Observée par le jardinier Robert Duval - traces de boue découvertes plus tard", Timestamp: parseDateTime("2025-08-29T17:30:00"), Location: "Bibliothèque", Entities: []string{"ent-moreau-005", "ent-moreau-007"}, Evidence: []string{"ev-moreau-003"}, Importance: "medium", Verified: true},
			{ID: "evt-m-05", CaseID: "case-moreau-001", Title: "Départ du jardinier", Description: "Ferme le portillon à clé", Timestamp: parseDateTime("2025-08-29T18:00:00"), Location: "Jardin", Entities: []string{"ent-moreau-005", "ent-moreau-013"}, Importance: "medium", Verified: true},
			{ID: "evt-m-06", CaseID: "case-moreau-001", Title: "Appel Jean vers Victor", Description: "Durée 3 minutes - contenu inconnu", Timestamp: parseDateTime("2025-08-29T18:45:00"), Location: "Téléphone", Entities: []string{"ent-moreau-001", "ent-moreau-002"}, Evidence: []string{"ev-moreau-004"}, Importance: "high", Verified: true},
			{ID: "evt-m-07", CaseID: "case-moreau-001", Title: "Départ Madame Chen", Description: "Laisse Victor seul après avoir servi le thé", Timestamp: parseDateTime("2025-08-29T19:00:00"), Location: "Manoir", Entities: []string{"ent-moreau-004", "ent-moreau-001"}, Evidence: []string{"ev-moreau-001"}, Importance: "high", Verified: true},
			{ID: "evt-m-08", CaseID: "case-moreau-001", Title: "Appel numéro inconnu", Description: "Appel entrant sur téléphone Victor - durée 2min", Timestamp: parseDateTime("2025-08-29T19:05:00"), Location: "Téléphone", Entities: []string{"ent-moreau-001", "ent-moreau-012"}, Evidence: []string{"ev-moreau-004"}, Importance: "high", Verified: true},
			{ID: "evt-m-09", CaseID: "case-moreau-001", Title: "Victor boit son thé", Description: "Dernière activité connue - thé possiblement empoisonné", Timestamp: parseDateTime("2025-08-29T19:15:00"), Location: "Bibliothèque", Entities: []string{"ent-moreau-001", "ent-moreau-007"}, Evidence: []string{"ev-moreau-001"}, Importance: "high", Verified: false},
			{ID: "evt-m-09b", CaseID: "case-moreau-001", Title: "Entrée cinéma Jean (alibi)", Description: "Ticket acheté pour séance 19h30", Timestamp: parseDateTime("2025-08-29T19:25:00"), Location: "UGC Bercy", Entities: []string{"ent-moreau-002"}, Importance: "high", Verified: true},
			{ID: "evt-m-10", CaseID: "case-moreau-001", Title: "Heure estimée du décès", Description: "Selon rapport médecin légiste Dr. Martin", Timestamp: parseDateTime("2025-08-29T20:30:00"), Location: "Bibliothèque", Entities: []string{"ent-moreau-001", "ent-moreau-017"}, Evidence: []string{"ev-moreau-001", "ev-moreau-002"}, Importance: "high", Verified: true},
			{ID: "evt-m-11", CaseID: "case-moreau-001", Title: "Découverte du corps", Description: "Par Madame Chen revenue au manoir", Timestamp: parseDateTime("2025-08-29T21:45:00"), Location: "Bibliothèque", Entities: []string{"ent-moreau-001", "ent-moreau-004", "ent-moreau-007"}, Evidence: []string{"ev-moreau-001", "ev-moreau-003"}, Importance: "high", Verified: true},
			{ID: "evt-m-12", CaseID: "case-moreau-001", Title: "Arrivée police", Description: "Début de l'enquête officielle", Timestamp: parseDateTime("2025-08-29T22:00:00"), Location: "Manoir", Entities: []string{}, Evidence: []string{"ev-moreau-001", "ev-moreau-002", "ev-moreau-003", "ev-moreau-004", "ev-moreau-005", "ev-moreau-006", "ev-moreau-007"}, Importance: "medium", Verified: true},
		},
		Hypotheses: []models.Hypothesis{
			{
				ID:              "hyp-m-01",
				CaseID:          "case-moreau-001",
				Title:           "Crime passionnel - Vengeance Élodie",
				Description:     "Élodie Dubois aurait empoisonné Victor pour se venger de la perte de 2 millions d'euros. Elle connaissait les habitudes de la victime et avait accès à la cuisine lors de visites antérieures. Sa présence près du manoir le jour du crime est troublante.",
				Status:          models.HypothesisPending,
				ConfidenceLevel: 65,
				SupportingEvidence: []string{"ev-moreau-004", "ev-moreau-006"},
				Questions:       []string{"Comment expliquer son alibi au dîner de charité?", "A-t-elle pu mandater quelqu'un?", "Source du poison?", "Qui est l'homme au téléphone inconnu?"},
				GeneratedBy:     "user",
			},
			{
				ID:              "hyp-m-02",
				CaseID:          "case-moreau-001",
				Title:           "Crime d'intérêt - Héritage Jean",
				Description:     "Jean Moreau aurait agi pour sécuriser son héritage avant la modification du testament. Ses dettes de jeu (150 000€) créent une urgence financière. Il connaît le code d'accès du manoir et sa pointure (42) correspond aux traces de boue.",
				Status:          models.HypothesisPending,
				ConfidenceLevel: 75,
				SupportingEvidence: []string{"ev-moreau-005", "ev-moreau-004", "ev-moreau-003"},
				Questions:       []string{"Alibi cinéma: a-t-il pu sortir pendant le film?", "A-t-il eu connaissance du changement de testament?", "Qui a-t-il rencontré au Bar Le Diplomate?", "Complice possible?"},
				GeneratedBy:     "user",
			},
			{
				ID:              "hyp-m-03",
				CaseID:          "case-moreau-001",
				Title:           "Suicide maquillé",
				Description:     "Victor aurait orchestré sa propre mort pour incriminer ses héritiers. Les recherches sur les poisons (historique navigateur) et les indices trop évidents (livre ouvert page des alcaloïdes) suggèrent une mise en scène.",
				Status:          models.HypothesisPending,
				ConfidenceLevel: 25,
				SupportingEvidence: []string{"ev-moreau-002"},
				Questions:       []string{"Problèmes de santé cachés?", "Personnalité compatible avec ce scénario?", "Motivation: punir Jean et Élodie?"},
				GeneratedBy:     "user",
			},
			{
				ID:              "hyp-m-04",
				CaseID:          "case-moreau-001",
				Title:           "Complot commercial - Piste Mercier",
				Description:     "Antoine Mercier, concurrent de Victor et rival au Cercle des Bibliophiles, aurait pu commanditer le crime pour éliminer un concurrent et récupérer sa clientèle. La rivalité lors de la vente aux enchères de juillet était intense.",
				Status:          models.HypothesisPending,
				ConfidenceLevel: 35,
				SupportingEvidence: []string{},
				Questions:       []string{"Mercier a-t-il un alibi pour le soir du crime?", "Connexion avec l'homme au téléphone inconnu?", "Intérêt pour les livres de Victor?", "Capacité à se procurer du poison?"},
				GeneratedBy:     "ai",
			},
			{
				ID:              "hyp-m-05",
				CaseID:          "case-moreau-001",
				Title:           "Complice interne - Madame Chen",
				Description:     "La gouvernante avait accès total à la cuisine et connaissait les habitudes de Victor. Son comportement 'étrangement calme' après la découverte du corps et le dîner privé avec Victor le 26/08 soulèvent des questions.",
				Status:          models.HypothesisPending,
				ConfidenceLevel: 20,
				SupportingEvidence: []string{"ev-moreau-001"},
				Questions:       []string{"Nature exacte de sa relation avec Victor?", "A-t-elle été approchée par Jean ou Élodie?", "Héritage prévu pour elle dans le nouveau testament?"},
				GeneratedBy:     "ai",
			},
		},
		// Contenu N4L natif avec toutes les fonctionnalités SSTorytime
		N4LContent: GetMoreauN4LContent(),
	}
}

// Affaire 2: Disparition - Enrichie pour N4L
func createAffaireDisparition() *models.Case {
	return &models.Case{
		ID:          "case-disparition-002",
		Name:        "Disparition Sophie Laurent",
		Description: "Disparition inquiétante d'une journaliste d'investigation travaillant sur une affaire de corruption impliquant des élus locaux et des entrepreneurs du BTP. Réseau complexe de relations politico-économiques.",
		Type:        "disparition",
		Status:      "en_cours",
		CreatedAt:   parseDate("2025-09-16"),
		UpdatedAt:   parseDate("2025-09-20"),
		Entities: []models.Entity{
			{
				ID:          "ent-disp-001",
				CaseID:      "case-disparition-002",
				Name:        "Sophie Laurent",
				Type:        models.EntityPerson,
				Role:        models.RoleVictim,
				Description: "Journaliste d'investigation, 34 ans. Travaillait sur un dossier sensible impliquant des élus locaux et marchés publics truqués. Disparue depuis le 15/09/2025.",
				Attributes: map[string]string{
					"age":            "34 ans",
					"profession":     "Journaliste d'investigation",
					"employeur":      "Le Courrier du Midi",
					"enquete":        "Corruption mairie Toulouse",
					"derniere_vue":   "15/09/2025 19h30",
					"vehicule":       "Renault Clio grise EF-456-GH",
					"telephone":      "Dernier signal 19h45",
					"articles_pub":   "3 articles sur l'affaire",
					"statut":         "Disparue",
				},
				Relations: []models.Relation{
					{ID: "rel-d-001", FromID: "ent-disp-001", ToID: "ent-disp-002", Type: "enquete", Label: "enquêtait sur", Context: "journalisme", Verified: true},
					{ID: "rel-d-002", FromID: "ent-disp-001", ToID: "ent-disp-003", Type: "collegue", Label: "collègue de", Context: "travail", Verified: true},
					{ID: "rel-d-003", FromID: "ent-disp-001", ToID: "ent-disp-005", Type: "contact", Label: "en contact avec", Context: "enquête", Verified: true},
					{ID: "rel-d-004", FromID: "ent-disp-001", ToID: "ent-disp-004", Type: "lieu", Label: "vue pour dernière fois à", Context: "disparition", Verified: true},
					{ID: "rel-d-004b", FromID: "ent-disp-001", ToID: "ent-disp-008", Type: "enquete", Label: "enquêtait sur", Context: "journalisme", Verified: true},
					{ID: "rel-d-004c", FromID: "ent-disp-001", ToID: "ent-disp-009", Type: "emploi", Label: "employée par", Context: "travail", Verified: true},
				},
			},
			{
				ID:          "ent-disp-002",
				CaseID:      "case-disparition-002",
				Name:        "Marc Delmas",
				Type:        models.EntityPerson,
				Role:        models.RoleSuspect,
				Description: "Adjoint au maire de Toulouse, délégué aux marchés publics. Principal cité dans l'enquête de Sophie. A proféré des menaces voilées au rédacteur en chef.",
				Attributes: map[string]string{
					"age":          "52 ans",
					"fonction":     "Adjoint au maire - Marchés publics",
					"parti":        "Coalition locale",
					"mobile":       "Éviter révélations corruption",
					"menaces":      "Elle ferait mieux d'arrêter si elle tient à sa carrière",
					"alibi":        "Réunion conseil municipal 19h-21h",
					"patrimoine":   "Suspect - enrichissement récent",
				},
				Relations: []models.Relation{
					{ID: "rel-d-005", FromID: "ent-disp-002", ToID: "ent-disp-001", Type: "menace", Label: "a menacé", Context: "intimidation", Verified: true},
					{ID: "rel-d-005b", FromID: "ent-disp-002", ToID: "ent-disp-008", Type: "affaires", Label: "favorise dans les marchés", Context: "corruption", Verified: false},
					{ID: "rel-d-005c", FromID: "ent-disp-002", ToID: "ent-disp-010", Type: "politique", Label: "collabore avec", Context: "mairie", Verified: true},
				},
			},
			{
				ID:          "ent-disp-003",
				CaseID:      "case-disparition-002",
				Name:        "Thomas Blanc",
				Type:        models.EntityPerson,
				Role:        models.RoleWitness,
				Description: "Collègue photographe et ami proche de Sophie. Dernier à l'avoir vue au journal. Travaillait avec elle sur l'enquête.",
				Attributes: map[string]string{
					"age":          "31 ans",
					"profession":   "Photographe de presse",
					"derniere_vue": "15/09 vers 19h30",
					"observation":  "Sophie semblait stressée, parlait d'un RDV important",
					"relation":     "Ami proche, possible relation amoureuse",
				},
				Relations: []models.Relation{
					{ID: "rel-d-006", FromID: "ent-disp-003", ToID: "ent-disp-001", Type: "temoin", Label: "dernier à avoir vu", Context: "disparition", Verified: true},
					{ID: "rel-d-006b", FromID: "ent-disp-003", ToID: "ent-disp-009", Type: "emploi", Label: "employé par", Context: "travail", Verified: true},
				},
			},
			{
				ID:          "ent-disp-004",
				CaseID:      "case-disparition-002",
				Name:        "Parking du Journal",
				Type:        models.EntityPlace,
				Role:        models.RoleOther,
				Description: "Dernier lieu où Sophie a été vue. Sa voiture y a été retrouvée portes non verrouillées.",
				Attributes: map[string]string{
					"adresse":          "Rue des Médias, Toulouse",
					"surveillance":     "Caméra - image SUV noir",
					"indices":          "Sac à main dans voiture, téléphone absent",
				},
			},
			{
				ID:          "ent-disp-005",
				CaseID:      "case-disparition-002",
				Name:        "Source Anonyme X",
				Type:        models.EntityPerson,
				Role:        models.RoleWitness,
				Description: "Informateur de Sophie sur l'affaire de corruption. Identité inconnue. Communiquait par messagerie cryptée.",
				Attributes: map[string]string{
					"identite":     "Inconnue",
					"communication": "Signal - messages cryptés",
					"informations": "Documents sur marchés truqués",
					"statut":       "Introuvable depuis disparition",
				},
				Relations: []models.Relation{
					{ID: "rel-d-007", FromID: "ent-disp-005", ToID: "ent-disp-001", Type: "information", Label: "informateur de", Context: "enquête", Verified: true},
					{ID: "rel-d-007b", FromID: "ent-disp-005", ToID: "ent-disp-010", Type: "connaissance", Label: "pourrait être proche de", Context: "suspect", Verified: false},
				},
			},
			// Nouvelles entités pour enrichir le réseau
			{
				ID:          "ent-disp-006",
				CaseID:      "case-disparition-002",
				Name:        "SUV Noir immatriculé partiellement",
				Type:        models.EntityObject,
				Role:        models.RoleOther,
				Description: "Véhicule dans lequel Sophie est montée à 19h48. Immatriculation partielle: ...BD-31",
				Attributes: map[string]string{
					"type":           "SUV noir",
					"immatriculation": "Partielle: ...BD-31",
					"proprietaire":   "Recherche en cours",
				},
				Relations: []models.Relation{
					{ID: "rel-d-008", FromID: "ent-disp-006", ToID: "ent-disp-001", Type: "transport", Label: "a transporté", Context: "disparition", Verified: true},
				},
			},
			{
				ID:          "ent-disp-007",
				CaseID:      "case-disparition-002",
				Name:        "Dossier Marchés Publics",
				Type:        models.EntityDocument,
				Role:        models.RoleOther,
				Description: "Notes manuscrites de Sophie sur les marchés publics truqués. Noms de plusieurs élus et entrepreneurs.",
				Attributes: map[string]string{
					"contenu":     "Surfacturations, rétro-commissions",
					"montants":    "Estimés à 2.3 millions d'euros",
					"personnes":   "Delmas, Roux Constructions, Maire",
				},
				Relations: []models.Relation{
					{ID: "rel-d-009", FromID: "ent-disp-007", ToID: "ent-disp-002", Type: "incrimine", Label: "incrimine", Context: "enquête", Verified: true},
					{ID: "rel-d-009b", FromID: "ent-disp-007", ToID: "ent-disp-008", Type: "incrimine", Label: "incrimine", Context: "enquête", Verified: true},
				},
			},
			{
				ID:          "ent-disp-008",
				CaseID:      "case-disparition-002",
				Name:        "Roux Constructions SARL",
				Type:        models.EntityOrg,
				Role:        models.RoleSuspect,
				Description: "Entreprise de BTP bénéficiaire de marchés publics suspects. Dirigée par Philippe Roux.",
				Attributes: map[string]string{
					"dirigeant":       "Philippe Roux",
					"secteur":         "BTP - Travaux publics",
					"marches_obtenus": "12 depuis 2023",
					"surfacturation":  "Estimée à 30%",
				},
				Relations: []models.Relation{
					{ID: "rel-d-010", FromID: "ent-disp-008", ToID: "ent-disp-002", Type: "corruption", Label: "verse des commissions à", Context: "corruption", Verified: false},
				},
			},
			{
				ID:          "ent-disp-009",
				CaseID:      "case-disparition-002",
				Name:        "Le Courrier du Midi",
				Type:        models.EntityOrg,
				Role:        models.RoleOther,
				Description: "Journal régional employeur de Sophie. Rédacteur en chef: Jean-Pierre Faure.",
				Attributes: map[string]string{
					"type":            "Quotidien régional",
					"redacteur_chef":  "Jean-Pierre Faure",
					"tirage":          "45 000 exemplaires",
				},
				Relations: []models.Relation{
					{ID: "rel-d-011", FromID: "ent-disp-009", ToID: "ent-disp-002", Type: "pression", Label: "a reçu des pressions de", Context: "intimidation", Verified: true},
				},
			},
			{
				ID:          "ent-disp-010",
				CaseID:      "case-disparition-002",
				Name:        "Maire Bernard Castex",
				Type:        models.EntityPerson,
				Role:        models.RoleSuspect,
				Description: "Maire de Toulouse. Supérieur hiérarchique de Delmas. Mentionné dans le dossier de Sophie.",
				Attributes: map[string]string{
					"age":          "61 ans",
					"fonction":     "Maire de Toulouse",
					"mandat":       "Depuis 2020",
					"implication":  "Indirecte - supervision marchés",
				},
				Relations: []models.Relation{
					{ID: "rel-d-012", FromID: "ent-disp-010", ToID: "ent-disp-002", Type: "hierarchie", Label: "supérieur de", Context: "politique", Verified: true},
				},
			},
		},
		Evidence: []models.Evidence{
			{
				ID:          "ev-disp-001",
				CaseID:      "case-disparition-002",
				Name:        "Véhicule abandonné",
				Type:        models.EvidencePhysical,
				Description: "Renault Clio retrouvée au parking du journal. Portes non verrouillées, sac à main à l'intérieur avec portefeuille et cartes.",
				Location:    "Parking Le Courrier du Midi",
				Reliability: 9,
				LinkedEntities: []string{"ent-disp-001", "ent-disp-004"},
			},
			{
				ID:          "ev-disp-002",
				CaseID:      "case-disparition-002",
				Name:        "Téléphone portable",
				Type:        models.EvidenceDigital,
				Description: "Dernier signal à 19h45 près du parking. Messages supprimés récupérés: RDV avec source X à 19h30. Absent de la voiture.",
				Reliability: 8,
				LinkedEntities: []string{"ent-disp-001", "ent-disp-005"},
			},
			{
				ID:          "ev-disp-003",
				CaseID:      "case-disparition-002",
				Name:        "Dossier d'enquête manuscrit",
				Type:        models.EvidenceDocumentary,
				Description: "Notes manuscrites sur marchés publics truqués. Noms: Delmas, Roux, Castex. Montants estimés: 2.3M€.",
				Location:    "Bureau de Sophie",
				Reliability: 9,
				LinkedEntities: []string{"ent-disp-001", "ent-disp-002", "ent-disp-008", "ent-disp-010"},
			},
			{
				ID:          "ev-disp-004",
				CaseID:      "case-disparition-002",
				Name:        "Vidéosurveillance parking",
				Type:        models.EvidenceDigital,
				Description: "Sophie monte volontairement dans un SUV noir à 19h48. Immatriculation partielle: ...BD-31. Conducteur non identifiable.",
				Reliability: 7,
				LinkedEntities: []string{"ent-disp-001", "ent-disp-006"},
			},
			{
				ID:          "ev-disp-005",
				CaseID:      "case-disparition-002",
				Name:        "Messages Signal cryptés",
				Type:        models.EvidenceDigital,
				Description: "Historique partiellement récupéré. Source X: 'J'ai les preuves définitives. RDV 19h30 parking habituel.'",
				Reliability: 8,
				LinkedEntities: []string{"ent-disp-001", "ent-disp-005"},
			},
			{
				ID:          "ev-disp-006",
				CaseID:      "case-disparition-002",
				Name:        "Enregistrement appel Delmas",
				Type:        models.EvidenceDigital,
				Description: "Appel au rédacteur en chef le 12/09: 'Vos journalistes feraient mieux de se calmer si le journal veut garder ses annonceurs publics.'",
				Reliability: 9,
				LinkedEntities: []string{"ent-disp-002", "ent-disp-009"},
			},
			{
				ID:          "ev-disp-007",
				CaseID:      "case-disparition-002",
				Name:        "Relevé bancaire Roux Constructions",
				Type:        models.EvidenceDocumentary,
				Description: "Virements réguliers vers compte offshore. Correspondances avec dates d'attribution de marchés.",
				Reliability: 7,
				LinkedEntities: []string{"ent-disp-008", "ent-disp-002"},
			},
		},
		Timeline: []models.Event{
			// Contexte
			{ID: "evt-d-00", CaseID: "case-disparition-002", Title: "Début enquête corruption", Description: "Sophie commence ses investigations sur les marchés publics", Timestamp: parseDateTime("2025-07-01T09:00:00"), Location: "Toulouse", Entities: []string{"ent-disp-001"}, Importance: "medium", Verified: true},
			{ID: "evt-d-01", CaseID: "case-disparition-002", Title: "Premier article publié", Description: "Révélations sur surfacturation marché école Jean-Jaurès", Timestamp: parseDateTime("2025-09-10T08:00:00"), Location: "Le Courrier du Midi", Entities: []string{"ent-disp-001", "ent-disp-009"}, Importance: "high", Verified: true},
			{ID: "evt-d-01b", CaseID: "case-disparition-002", Title: "Réaction de la mairie", Description: "Communiqué contestant les accusations", Timestamp: parseDateTime("2025-09-10T14:00:00"), Location: "Mairie Toulouse", Entities: []string{"ent-disp-010"}, Importance: "medium", Verified: true},
			{ID: "evt-d-02", CaseID: "case-disparition-002", Title: "Menaces de Delmas", Description: "Appel au rédacteur en chef - menaces voilées sur annonceurs", Timestamp: parseDateTime("2025-09-12T14:00:00"), Location: "Téléphone", Entities: []string{"ent-disp-002", "ent-disp-009"}, Importance: "high", Verified: true},
			{ID: "evt-d-02b", CaseID: "case-disparition-002", Title: "Second article publié", Description: "Nouvelles révélations sur Roux Constructions", Timestamp: parseDateTime("2025-09-13T08:00:00"), Location: "Le Courrier du Midi", Entities: []string{"ent-disp-001", "ent-disp-008"}, Importance: "high", Verified: true},
			// Jour de la disparition
			{ID: "evt-d-03", CaseID: "case-disparition-002", Title: "Message de Source X", Description: "'J'ai les preuves définitives. RDV 19h30 parking habituel.'", Timestamp: parseDateTime("2025-09-15T15:00:00"), Location: "Signal", Entities: []string{"ent-disp-001", "ent-disp-005"}, Importance: "high", Verified: true},
			{ID: "evt-d-03b", CaseID: "case-disparition-002", Title: "Sophie informe Thomas", Description: "Parle d'un RDV important, semble stressée", Timestamp: parseDateTime("2025-09-15T18:00:00"), Location: "Rédaction", Entities: []string{"ent-disp-001", "ent-disp-003"}, Importance: "medium", Verified: true},
			{ID: "evt-d-04", CaseID: "case-disparition-002", Title: "Départ du journal", Description: "Sophie quitte la rédaction - vue par Thomas Blanc", Timestamp: parseDateTime("2025-09-15T19:30:00"), Location: "Le Courrier du Midi", Entities: []string{"ent-disp-001", "ent-disp-003"}, Importance: "high", Verified: true},
			{ID: "evt-d-06", CaseID: "case-disparition-002", Title: "Dernier signal téléphone", Description: "Localisation près du parking puis perdue", Timestamp: parseDateTime("2025-09-15T19:45:00"), Location: "Parking", Entities: []string{"ent-disp-001"}, Importance: "high", Verified: true},
			{ID: "evt-d-05", CaseID: "case-disparition-002", Title: "Monte dans SUV noir", Description: "Sophie monte VOLONTAIREMENT - capté par vidéosurveillance", Timestamp: parseDateTime("2025-09-15T19:48:00"), Location: "Parking", Entities: []string{"ent-disp-001", "ent-disp-006"}, Importance: "high", Verified: true},
			// Après disparition
			{ID: "evt-d-07", CaseID: "case-disparition-002", Title: "Voiture retrouvée", Description: "Renault Clio découverte par gardien - portes non verrouillées", Timestamp: parseDateTime("2025-09-16T07:00:00"), Location: "Parking", Entities: []string{"ent-disp-004"}, Importance: "high", Verified: true},
			{ID: "evt-d-08", CaseID: "case-disparition-002", Title: "Signalement disparition", Description: "Thomas Blanc alerte la police", Timestamp: parseDateTime("2025-09-16T10:00:00"), Location: "Commissariat", Entities: []string{"ent-disp-003"}, Importance: "high", Verified: true},
		},
		Hypotheses: []models.Hypothesis{
			{
				ID:              "hyp-d-01",
				CaseID:          "case-disparition-002",
				Title:           "Enlèvement commandité par Delmas",
				Description:     "Sophie aurait été enlevée sur ordre de Marc Delmas pour l'empêcher de publier de nouvelles révélations. Le SUV pourrait appartenir à un sbire ou à Roux Constructions.",
				Status:          models.HypothesisPending,
				ConfidenceLevel: 75,
				SupportingEvidence: []string{"ev-disp-004", "ev-disp-006"},
				Questions:       []string{"Qui conduit le SUV noir?", "Lien entre le SUV et Roux Constructions?", "Où a-t-elle été emmenée?"},
				GeneratedBy:     "user",
			},
			{
				ID:              "hyp-d-02",
				CaseID:          "case-disparition-002",
				Title:           "Piège de la Source X",
				Description:     "La Source X pourrait être un agent double travaillant pour les corrompus. Le RDV était un piège pour attirer Sophie.",
				Status:          models.HypothesisPending,
				ConfidenceLevel: 60,
				SupportingEvidence: []string{"ev-disp-005", "ev-disp-002"},
				Questions:       []string{"Source X est-elle complice ou victime?", "Qui connaissait l'existence de Source X?", "Source X a-t-elle aussi disparu?"},
				GeneratedBy:     "user",
			},
			{
				ID:              "hyp-d-03",
				CaseID:          "case-disparition-002",
				Title:           "Implication du Maire",
				Description:     "Le maire Castex pourrait avoir ordonné l'enlèvement pour protéger sa réélection. Delmas n'est qu'un exécutant.",
				Status:          models.HypothesisPending,
				ConfidenceLevel: 45,
				SupportingEvidence: []string{"ev-disp-003"},
				Questions:       []string{"Castex était-il au courant des menaces de Delmas?", "Quel est le niveau d'implication du maire?", "Liens avec le crime organisé?"},
				GeneratedBy:     "ai",
			},
			{
				ID:              "hyp-d-04",
				CaseID:          "case-disparition-002",
				Title:           "Disparition volontaire",
				Description:     "Sophie aurait pu organiser sa propre disparition pour se protéger ou pour mener une enquête sous couverture plus risquée.",
				Status:          models.HypothesisPending,
				ConfidenceLevel: 15,
				SupportingEvidence: []string{},
				Questions:       []string{"Sophie avait-elle des raisons de se cacher?", "A-t-elle préparé une fuite?", "Contact avec sa famille depuis?"},
				GeneratedBy:     "ai",
			},
		},
		// Contenu N4L natif avec toutes les fonctionnalités SSTorytime
		N4LContent: GetDisparitionSophieN4LContent(),
	}
}

// Affaire 3: Fraude financière
func createAffaireFraude() *models.Case {
	return &models.Case{
		ID:          "case-fraude-003",
		Name:        "Fraude Pyramidale FinanceMax",
		Description: "Escroquerie de type Ponzi ayant fait 847 victimes pour un préjudice estimé à 23 millions d'euros.",
		Type:        "fraude",
		Status:      "en_cours",
		CreatedAt:   parseDate("2025-07-15"),
		UpdatedAt:   parseDate("2025-09-25"),
		Entities: []models.Entity{
			{
				ID:          "ent-fraud-001",
				CaseID:      "case-fraude-003",
				Name:        "Philippe Martin",
				Type:        models.EntityPerson,
				Role:        models.RoleSuspect,
				Description: "Fondateur et PDG de FinanceMax. 48 ans. Promettait des rendements de 15% par mois.",
				Attributes: map[string]string{
					"age":         "48 ans",
					"profession":  "Financier",
					"societe":     "FinanceMax SARL",
					"prejudice":   "23 millions €",
					"victimes":    "847 personnes",
				},
				Relations: []models.Relation{
					{ID: "rel-f-001", FromID: "ent-fraud-001", ToID: "ent-fraud-002", Type: "hierarchie", Label: "supérieur hiérarchique de", Context: "travail", Verified: true},
					{ID: "rel-f-002", FromID: "ent-fraud-001", ToID: "ent-fraud-003", Type: "prejudice", Label: "a escroqué", Context: "fraude", Verified: true},
					{ID: "rel-f-003", FromID: "ent-fraud-001", ToID: "ent-fraud-004", Type: "finance", Label: "a transféré vers", Context: "blanchiment", Verified: true},
				},
			},
			{
				ID:          "ent-fraud-002",
				CaseID:      "case-fraude-003",
				Name:        "Céline Roux",
				Type:        models.EntityPerson,
				Role:        models.RoleSuspect,
				Description: "Directrice commerciale. Recrutait les investisseurs avec des promesses mensongères.",
				Attributes: map[string]string{
					"age":        "39 ans",
					"role":       "Recrutement investisseurs",
				},
				Relations: []models.Relation{
					{ID: "rel-f-004", FromID: "ent-fraud-002", ToID: "ent-fraud-001", Type: "complice", Label: "complice de", Context: "fraude", Verified: true},
					{ID: "rel-f-005", FromID: "ent-fraud-002", ToID: "ent-fraud-003", Type: "recrutement", Label: "a recruté", Context: "fraude", Verified: true},
				},
			},
			{
				ID:          "ent-fraud-003",
				CaseID:      "case-fraude-003",
				Name:        "Association Victimes FinanceMax",
				Type:        models.EntityOrg,
				Role:        models.RoleOther,
				Description: "Regroupement de 612 victimes pour action collective.",
				Relations: []models.Relation{
					{ID: "rel-f-006", FromID: "ent-fraud-003", ToID: "ent-fraud-001", Type: "plainte", Label: "porte plainte contre", Context: "judiciaire", Verified: true},
				},
			},
			{
				ID:          "ent-fraud-004",
				CaseID:      "case-fraude-003",
				Name:        "Compte Suisse HSBC",
				Type:        models.EntityDocument,
				Role:        models.RoleOther,
				Description: "Compte bancaire offshore où transitaient les fonds.",
			},
		},
		Evidence: []models.Evidence{
			{
				ID:          "ev-fraud-001",
				CaseID:      "case-fraude-003",
				Name:        "Contrats d'investissement",
				Type:        models.EvidenceDocumentary,
				Description: "847 contrats promettant 15% de rendement mensuel garanti. Clauses abusives.",
				Reliability: 10,
			},
			{
				ID:          "ev-fraud-002",
				CaseID:      "case-fraude-003",
				Name:        "Relevés bancaires offshore",
				Type:        models.EvidenceDocumentary,
				Description: "Transferts vers compte HSBC Genève. 18 millions tracés.",
				Reliability: 9,
			},
			{
				ID:          "ev-fraud-003",
				CaseID:      "case-fraude-003",
				Name:        "Emails internes",
				Type:        models.EvidenceDigital,
				Description: "Communications entre Martin et Roux confirmant la connaissance du schéma frauduleux.",
				Reliability: 9,
			},
			{
				ID:          "ev-fraud-004",
				CaseID:      "case-fraude-003",
				Name:        "Témoignages victimes",
				Type:        models.EvidenceTestimonial,
				Description: "Dépositions de 127 victimes décrivant les méthodes de recrutement.",
				Reliability: 8,
			},
		},
		Timeline: []models.Event{
			{ID: "evt-f-01", CaseID: "case-fraude-003", Title: "Création FinanceMax", Timestamp: parseDateTime("2023-03-15T09:00:00"), Importance: "high", Verified: true},
			{ID: "evt-f-02", CaseID: "case-fraude-003", Title: "Premiers investisseurs", Description: "50 personnes recrutées", Timestamp: parseDateTime("2023-06-01T09:00:00"), Importance: "medium", Verified: true},
			{ID: "evt-f-03", CaseID: "case-fraude-003", Title: "Premier transfert offshore", Description: "2 millions vers Suisse", Timestamp: parseDateTime("2024-01-15T09:00:00"), Importance: "high", Verified: true},
			{ID: "evt-f-04", CaseID: "case-fraude-003", Title: "Pic d'investissement", Description: "12 millions collectés en 3 mois", Timestamp: parseDateTime("2024-09-01T09:00:00"), Importance: "high", Verified: true},
			{ID: "evt-f-05", CaseID: "case-fraude-003", Title: "Premiers défauts de paiement", Timestamp: parseDateTime("2025-05-01T09:00:00"), Importance: "high", Verified: true},
			{ID: "evt-f-06", CaseID: "case-fraude-003", Title: "Plainte collective déposée", Timestamp: parseDateTime("2025-07-10T09:00:00"), Importance: "high", Verified: true},
		},
		Hypotheses: []models.Hypothesis{
			{
				ID:              "hyp-f-01",
				CaseID:          "case-fraude-003",
				Title:           "Schéma Ponzi classique",
				Description:     "Les rendements des anciens investisseurs étaient payés avec l'argent des nouveaux, sans aucun investissement réel.",
				Status:          models.HypothesisSupported,
				ConfidenceLevel: 95,
				GeneratedBy:     "user",
			},
			{
				ID:              "hyp-f-02",
				CaseID:          "case-fraude-003",
				Title:           "Complices bancaires",
				Description:     "Des employés de banque auraient facilité les transferts offshore en échange de commissions.",
				Status:          models.HypothesisPending,
				ConfidenceLevel: 40,
				GeneratedBy:     "ai",
			},
		},
		// Contenu N4L natif avec toutes les fonctionnalités SSTorytime
		N4LContent: GetFraudeN4LContent(),
	}
}

// Affaire 4: Cambriolage
func createAffaireCambriolage() *models.Case {
	return &models.Case{
		ID:          "case-cambriolage-004",
		Name:        "Cambriolage Musée des Arts Premiers",
		Description: "Vol de 5 statuettes africaines d'une valeur estimée à 2.8 millions d'euros. Mode opératoire sophistiqué.",
		Type:        "vol",
		Status:      "en_cours",
		CreatedAt:   parseDate("2025-10-02"),
		UpdatedAt:   parseDate("2025-10-05"),
		Entities: []models.Entity{
			{
				ID:          "ent-camb-001",
				CaseID:      "case-cambriolage-004",
				Name:        "Musée des Arts Premiers",
				Type:        models.EntityPlace,
				Role:        models.RoleOther,
				Description: "Musée municipal, sécurité niveau 3. Système alarme neutralisé.",
				Relations: []models.Relation{
					{ID: "rel-c-001", FromID: "ent-camb-001", ToID: "ent-camb-004", Type: "possession", Label: "exposait", Context: "musée", Verified: true},
				},
			},
			{
				ID:          "ent-camb-002",
				CaseID:      "case-cambriolage-004",
				Name:        "Pierre Lafont",
				Type:        models.EntityPerson,
				Role:        models.RoleSuspect,
				Description: "Ancien agent de sécurité licencié il y a 6 mois. Connaît le système d'alarme.",
				Attributes: map[string]string{
					"age":           "41 ans",
					"ancien_emploi": "Agent sécurité musée",
					"licenciement":  "Mars 2025 - faute grave",
				},
				Relations: []models.Relation{
					{ID: "rel-c-002", FromID: "ent-camb-002", ToID: "ent-camb-001", Type: "emploi", Label: "ancien employé de", Context: "travail", Verified: true},
					{ID: "rel-c-003", FromID: "ent-camb-002", ToID: "ent-camb-003", Type: "complicite", Label: "aurait collaboré avec", Context: "vol", Verified: false},
				},
			},
			{
				ID:          "ent-camb-003",
				CaseID:      "case-cambriolage-004",
				Name:        "Collectionneurs suspects",
				Type:        models.EntityOrg,
				Role:        models.RoleSuspect,
				Description: "Réseau de collectionneurs d'art africain identifié par Interpol.",
				Relations: []models.Relation{
					{ID: "rel-c-004", FromID: "ent-camb-003", ToID: "ent-camb-004", Type: "interet", Label: "intéressé par", Context: "art", Verified: true},
					{ID: "rel-c-005", FromID: "ent-camb-003", ToID: "ent-camb-001", Type: "cible", Label: "a ciblé", Context: "vol", Verified: false},
				},
			},
			{
				ID:          "ent-camb-004",
				CaseID:      "case-cambriolage-004",
				Name:        "Statuettes Dogon",
				Type:        models.EntityObject,
				Role:        models.RoleOther,
				Description: "5 statuettes rituelles du Mali, XIVe siècle. Valeur: 2.8 millions €.",
			},
		},
		Evidence: []models.Evidence{
			{
				ID:          "ev-camb-001",
				CaseID:      "case-cambriolage-004",
				Name:        "Vidéosurveillance neutralisée",
				Type:        models.EvidenceDigital,
				Description: "Signal coupé de 2h15 à 3h45. Image de boucle détectée.",
				Reliability: 7,
			},
			{
				ID:          "ev-camb-002",
				CaseID:      "case-cambriolage-004",
				Name:        "Gants en latex",
				Type:        models.EvidencePhysical,
				Description: "Paire retrouvée près sortie secours. ADN en cours d'analyse.",
				Reliability: 8,
			},
			{
				ID:          "ev-camb-003",
				CaseID:      "case-cambriolage-004",
				Name:        "Traces de fourgon",
				Type:        models.EvidenceForensic,
				Description: "Empreintes pneus dans allée service. Modèle Renault Master.",
				Reliability: 6,
			},
		},
		Timeline: []models.Event{
			{ID: "evt-c-01", CaseID: "case-cambriolage-004", Title: "Fermeture musée", Timestamp: parseDateTime("2025-10-01T18:00:00"), Importance: "medium", Verified: true},
			{ID: "evt-c-02", CaseID: "case-cambriolage-004", Title: "Coupure vidéosurveillance", Timestamp: parseDateTime("2025-10-02T02:15:00"), Importance: "high", Verified: true},
			{ID: "evt-c-03", CaseID: "case-cambriolage-004", Title: "Intrusion estimée", Timestamp: parseDateTime("2025-10-02T02:30:00"), Importance: "high", Verified: false},
			{ID: "evt-c-04", CaseID: "case-cambriolage-004", Title: "Retour vidéosurveillance", Timestamp: parseDateTime("2025-10-02T03:45:00"), Importance: "high", Verified: true},
			{ID: "evt-c-05", CaseID: "case-cambriolage-004", Title: "Découverte vol", Description: "Par gardien de nuit", Timestamp: parseDateTime("2025-10-02T06:00:00"), Importance: "high", Verified: true},
		},
		Hypotheses: []models.Hypothesis{
			{
				ID:              "hyp-c-01",
				CaseID:          "case-cambriolage-004",
				Title:           "Complicité interne",
				Description:     "L'ancien agent de sécurité Lafont aurait fourni les codes et les plans à une équipe professionnelle.",
				Status:          models.HypothesisPending,
				ConfidenceLevel: 70,
				GeneratedBy:     "user",
			},
		},
		// Contenu N4L natif avec toutes les fonctionnalités SSTorytime
		N4LContent: GetCambriolageN4LContent(),
	}
}

// Affaire 5: Incendie criminel
func createAffaireIncendie() *models.Case {
	return &models.Case{
		ID:          "case-incendie-005",
		Name:        "Incendie Entrepôt Logistique Nord",
		Description: "Incendie criminel ayant détruit un entrepôt de 5000m². Suspicion de fraude à l'assurance.",
		Type:        "incendie",
		Status:      "en_cours",
		CreatedAt:   parseDate("2025-09-28"),
		UpdatedAt:   parseDate("2025-10-01"),
		Entities: []models.Entity{
			{
				ID:          "ent-inc-001",
				CaseID:      "case-incendie-005",
				Name:        "Entrepôt Logistique Nord",
				Type:        models.EntityPlace,
				Role:        models.RoleOther,
				Description: "Entrepôt de 5000m² détruit à 80%. Valeur assurée: 4.5 millions €.",
			},
			{
				ID:          "ent-inc-002",
				CaseID:      "case-incendie-005",
				Name:        "André Petit",
				Type:        models.EntityPerson,
				Role:        models.RoleSuspect,
				Description: "Propriétaire de l'entrepôt. Difficultés financières connues. Assurance augmentée 3 mois avant.",
				Attributes: map[string]string{
					"age":         "56 ans",
					"profession":  "Chef d'entreprise",
					"situation":   "Dettes importantes",
					"assurance":   "Augmentée en juillet 2025",
				},
				Relations: []models.Relation{
					{ID: "rel-i-001", FromID: "ent-inc-002", ToID: "ent-inc-001", Type: "propriete", Label: "propriétaire de", Context: "immobilier", Verified: true},
					{ID: "rel-i-002", FromID: "ent-inc-002", ToID: "ent-inc-004", Type: "contrat", Label: "assuré par", Context: "assurance", Verified: true},
				},
			},
			{
				ID:          "ent-inc-003",
				CaseID:      "case-incendie-005",
				Name:        "Expert Assurance Durand",
				Type:        models.EntityPerson,
				Role:        models.RoleWitness,
				Description: "Expert mandaté par l'assureur. Conclut à une origine criminelle.",
				Relations: []models.Relation{
					{ID: "rel-i-003", FromID: "ent-inc-003", ToID: "ent-inc-001", Type: "expertise", Label: "a expertisé", Context: "enquête", Verified: true},
					{ID: "rel-i-004", FromID: "ent-inc-003", ToID: "ent-inc-004", Type: "mandat", Label: "mandaté par", Context: "assurance", Verified: true},
				},
			},
			{
				ID:          "ent-inc-004",
				CaseID:      "case-incendie-005",
				Name:        "Assurance MutualPro",
				Type:        models.EntityOrg,
				Role:        models.RoleOther,
				Description: "Compagnie d'assurance. Police augmentée à 4.5 millions 3 mois avant l'incendie.",
				Relations: []models.Relation{
					{ID: "rel-i-005", FromID: "ent-inc-004", ToID: "ent-inc-002", Type: "suspicion", Label: "suspecte fraude de", Context: "assurance", Verified: true},
				},
			},
		},
		Evidence: []models.Evidence{
			{
				ID:          "ev-inc-001",
				CaseID:      "case-incendie-005",
				Name:        "Traces d'accélérant",
				Type:        models.EvidenceForensic,
				Description: "Résidus d'essence retrouvés en 3 points distincts. Origine criminelle confirmée.",
				Reliability: 10,
			},
			{
				ID:          "ev-inc-002",
				CaseID:      "case-incendie-005",
				Name:        "Augmentation assurance",
				Type:        models.EvidenceDocumentary,
				Description: "Police d'assurance modifiée le 01/07/2025. Couverture passée de 2 à 4.5 millions.",
				Reliability: 9,
			},
			{
				ID:          "ev-inc-003",
				CaseID:      "case-incendie-005",
				Name:        "Relevés bancaires Petit",
				Type:        models.EvidenceDocumentary,
				Description: "Compte à découvert de 380 000€. Plusieurs rejets de prélèvement.",
				Reliability: 9,
			},
		},
		Timeline: []models.Event{
			{ID: "evt-i-01", CaseID: "case-incendie-005", Title: "Augmentation assurance", Timestamp: parseDateTime("2025-07-01T09:00:00"), Importance: "high", Verified: true},
			{ID: "evt-i-02", CaseID: "case-incendie-005", Title: "Dernier inventaire", Description: "Stock déclaré à 1.2 million", Timestamp: parseDateTime("2025-09-15T09:00:00"), Importance: "medium", Verified: true},
			{ID: "evt-i-03", CaseID: "case-incendie-005", Title: "Départ feu", Timestamp: parseDateTime("2025-09-28T03:15:00"), Importance: "high", Verified: true},
			{ID: "evt-i-04", CaseID: "case-incendie-005", Title: "Arrivée pompiers", Timestamp: parseDateTime("2025-09-28T03:35:00"), Importance: "medium", Verified: true},
			{ID: "evt-i-05", CaseID: "case-incendie-005", Title: "Feu maîtrisé", Timestamp: parseDateTime("2025-09-28T07:00:00"), Importance: "medium", Verified: true},
		},
		Hypotheses: []models.Hypothesis{
			{
				ID:              "hyp-i-01",
				CaseID:          "case-incendie-005",
				Title:           "Fraude à l'assurance",
				Description:     "Le propriétaire aurait commandité l'incendie pour toucher l'indemnisation et rembourser ses dettes.",
				Status:          models.HypothesisPending,
				ConfidenceLevel: 80,
				GeneratedBy:     "user",
			},
		},
		// Contenu N4L natif avec toutes les fonctionnalités SSTorytime
		N4LContent: GetIncendieN4LContent(),
	}
}

// Affaire 6: Trafic d'Art et Blanchiment - Connexions fortes avec Moreau et Disparition
func createAffaireTraficArt() *models.Case {
	return &models.Case{
		ID:          "case-trafic-006",
		Name:        "Trafic d'Art et Blanchiment",
		Description: "Réseau de trafic d'œuvres d'art volées et blanchiment d'argent. Connexions avec le milieu des antiquaires parisiens et des entreprises de BTP.",
		Type:        "trafic",
		Status:      "en_cours",
		CreatedAt:   parseDate("2025-10-10"),
		UpdatedAt:   parseDate("2025-10-14"),
		Entities: []models.Entity{
			// Suspect principal - MÊME NOM que dans Affaire Moreau
			{
				ID:          "ent-traf-001",
				CaseID:      "case-trafic-006",
				Name:        "Jean Moreau",
				Type:        models.EntityPerson,
				Role:        models.RoleSuspect,
				Description: "Intermédiaire présumé dans le trafic d'œuvres d'art. Neveu d'un antiquaire décédé. Dettes de jeu importantes.",
				Attributes: map[string]string{
					"age":            "35 ans",
					"profession":     "Sans emploi",
					"dettes":         "150 000€",
					"role_reseau":    "Intermédiaire - contact acheteurs",
					"vehicule":       "BMW série 3 noire AB-123-CD",
					"lien_moreau":    "Neveu de Victor Moreau (décédé)",
				},
				Relations: []models.Relation{
					{ID: "rel-t-001", FromID: "ent-traf-001", ToID: "ent-traf-002", Type: "complice", Label: "complice de", Context: "trafic", Verified: true},
					{ID: "rel-t-002", FromID: "ent-traf-001", ToID: "ent-traf-005", Type: "frequentation", Label: "fréquente", Context: "rencontres", Verified: true},
					{ID: "rel-t-003", FromID: "ent-traf-001", ToID: "ent-traf-006", Type: "acces", Label: "a accès à", Context: "trafic", Verified: true},
				},
			},
			// Chef du réseau
			{
				ID:          "ent-traf-002",
				CaseID:      "case-trafic-006",
				Name:        "Viktor Sokolov",
				Type:        models.EntityPerson,
				Role:        models.RoleSuspect,
				Description: "Chef présumé du réseau de trafic. Nationalité russe. Recherché par Interpol.",
				Attributes: map[string]string{
					"age":            "52 ans",
					"nationalite":    "Russe",
					"alias":          "Le Collectionneur",
					"interpol":       "Notice rouge",
					"fortune":        "Estimée à 30 millions €",
				},
				Relations: []models.Relation{
					{ID: "rel-t-004", FromID: "ent-traf-002", ToID: "ent-traf-003", Type: "blanchiment", Label: "blanchit via", Context: "finances", Verified: true},
					{ID: "rel-t-005", FromID: "ent-traf-002", ToID: "ent-traf-004", Type: "commanditaire", Label: "commandite", Context: "trafic", Verified: false},
				},
			},
			// Entreprise blanchiment - MÊME NOM que dans Affaire Disparition
			{
				ID:          "ent-traf-003",
				CaseID:      "case-trafic-006",
				Name:        "Roux Constructions SARL",
				Type:        models.EntityOrg,
				Role:        models.RoleSuspect,
				Description: "Entreprise de BTP utilisée pour le blanchiment. Surfacturation de chantiers fictifs.",
				Attributes: map[string]string{
					"dirigeant":       "Philippe Roux",
					"secteur":         "BTP - Travaux publics",
					"blanchiment":     "Via fausses factures",
					"montant_blanchi": "Estimé à 5 millions €",
				},
				Relations: []models.Relation{
					{ID: "rel-t-006", FromID: "ent-traf-003", ToID: "ent-traf-007", Type: "corruption", Label: "verse des commissions à", Context: "corruption", Verified: true},
				},
			},
			// Receleur
			{
				ID:          "ent-traf-004",
				CaseID:      "case-trafic-006",
				Name:        "Galerie L'Éclipse",
				Type:        models.EntityPlace,
				Role:        models.RoleSuspect,
				Description: "Galerie d'art servant de façade pour écouler les œuvres volées.",
				Attributes: map[string]string{
					"adresse":    "8 rue de Seine, Paris 6e",
					"proprietaire": "Société écran luxembourgeoise",
					"activite":   "Vente d'art contemporain (façade)",
				},
				Relations: []models.Relation{
					{ID: "rel-t-007", FromID: "ent-traf-004", ToID: "ent-traf-006", Type: "concurrent", Label: "en relation avec", Context: "art", Verified: true},
				},
			},
			// Lieu de rencontre - MÊME NOM que dans Affaire Moreau
			{
				ID:          "ent-traf-005",
				CaseID:      "case-trafic-006",
				Name:        "Bar Le Diplomate",
				Type:        models.EntityPlace,
				Role:        models.RoleOther,
				Description: "Lieu de rencontre entre les membres du réseau. Transactions en liquide.",
				Attributes: map[string]string{
					"adresse": "45 rue de Bercy, Paris 12e",
					"type":    "Bar de nuit",
					"activite": "Lieu de rencontres clandestines",
				},
			},
			// Galerie liée - MÊME NOM que dans Affaire Moreau
			{
				ID:          "ent-traf-006",
				CaseID:      "case-trafic-006",
				Name:        "Galerie Moreau Antiquités",
				Type:        models.EntityPlace,
				Role:        models.RoleOther,
				Description: "Ancienne galerie de Victor Moreau. Suspectée d'avoir servi au recel avant son décès.",
				Attributes: map[string]string{
					"adresse":      "12 rue du Faubourg Saint-Honoré, Paris 8e",
					"statut":       "Succession en cours",
					"suspicion":    "Recel d'œuvres volées",
				},
			},
			// Politicien corrompu - MÊME NOM que dans Affaire Disparition
			{
				ID:          "ent-traf-007",
				CaseID:      "case-trafic-006",
				Name:        "Marc Delmas",
				Type:        models.EntityPerson,
				Role:        models.RoleSuspect,
				Description: "Élu local facilitant les permis et marchés pour Roux Constructions. Reçoit des œuvres d'art en paiement.",
				Attributes: map[string]string{
					"age":          "52 ans",
					"fonction":     "Adjoint au maire - Marchés publics",
					"corruption":   "Pots-de-vin en œuvres d'art",
					"collection":   "Art africain et antiquités",
				},
				Relations: []models.Relation{
					{ID: "rel-t-008", FromID: "ent-traf-007", ToID: "ent-traf-003", Type: "corruption", Label: "reçoit des pots-de-vin de", Context: "corruption", Verified: true},
				},
			},
			// Témoin clé
			{
				ID:          "ent-traf-008",
				CaseID:      "case-trafic-006",
				Name:        "Claire Fontaine",
				Type:        models.EntityPerson,
				Role:        models.RoleWitness,
				Description: "Ancienne employée de la Galerie L'Éclipse. A dénoncé le réseau anonymement.",
				Attributes: map[string]string{
					"age":         "28 ans",
					"profession":  "Historienne de l'art",
					"statut":      "Protection témoin",
				},
				Relations: []models.Relation{
					{ID: "rel-t-009", FromID: "ent-traf-008", ToID: "ent-traf-004", Type: "emploi", Label: "ancienne employée de", Context: "travail", Verified: true},
					{ID: "rel-t-010", FromID: "ent-traf-008", ToID: "ent-traf-002", Type: "denonciation", Label: "a dénoncé", Context: "enquête", Verified: true},
				},
			},
			// Expert complice - Lien avec Cercle des Bibliophiles (Affaire Moreau)
			{
				ID:          "ent-traf-009",
				CaseID:      "case-trafic-006",
				Name:        "Antoine Mercier",
				Type:        models.EntityPerson,
				Role:        models.RoleSuspect,
				Description: "Expert en art ancien. Authentifie de fausses provenances pour les œuvres volées.",
				Attributes: map[string]string{
					"age":        "55 ans",
					"profession": "Expert en livres anciens et objets d'art",
					"galerie":    "Mercier & Fils",
					"role":       "Faux certificats d'authenticité",
				},
				Relations: []models.Relation{
					{ID: "rel-t-011", FromID: "ent-traf-009", ToID: "ent-traf-002", Type: "complice", Label: "travaille pour", Context: "trafic", Verified: true},
					{ID: "rel-t-012", FromID: "ent-traf-009", ToID: "ent-traf-006", Type: "expertise", Label: "expertisait pour", Context: "art", Verified: true},
				},
			},
		},
		Evidence: []models.Evidence{
			{
				ID:          "ev-traf-001",
				CaseID:      "case-trafic-006",
				Name:        "Factures falsifiées Roux Constructions",
				Type:        models.EvidenceDocumentary,
				Description: "Factures de travaux fictifs pour un total de 2.3 millions €. Correspondent aux dates de ventes d'œuvres.",
				Location:    "Siège Roux Constructions",
				Reliability: 9,
				LinkedEntities: []string{"ent-traf-003", "ent-traf-007"},
			},
			{
				ID:          "ev-traf-002",
				CaseID:      "case-trafic-006",
				Name:        "Écoutes téléphoniques",
				Type:        models.EvidenceDigital,
				Description: "Conversations entre Sokolov et Jean Moreau sur des 'livraisons'. Mentions de 'l'oncle' et 'la galerie'.",
				Reliability: 8,
				LinkedEntities: []string{"ent-traf-001", "ent-traf-002"},
			},
			{
				ID:          "ev-traf-003",
				CaseID:      "case-trafic-006",
				Name:        "Liste d'œuvres volées",
				Type:        models.EvidenceDocumentary,
				Description: "Inventaire Interpol: 12 œuvres transitées par le réseau. 3 retrouvées chez Delmas.",
				Reliability: 10,
				LinkedEntities: []string{"ent-traf-002", "ent-traf-007"},
			},
			{
				ID:          "ev-traf-004",
				CaseID:      "case-trafic-006",
				Name:        "Vidéosurveillance Bar Le Diplomate",
				Type:        models.EvidenceDigital,
				Description: "Images de Jean Moreau remettant une enveloppe à un homme non identifié le 05/10/2025.",
				Location:    "Bar Le Diplomate",
				Reliability: 7,
				LinkedEntities: []string{"ent-traf-001", "ent-traf-005"},
			},
			{
				ID:          "ev-traf-005",
				CaseID:      "case-trafic-006",
				Name:        "Témoignage Claire Fontaine",
				Type:        models.EvidenceTestimonial,
				Description: "Décrit le processus de blanchiment et nomme Sokolov, Moreau et Mercier.",
				Reliability: 8,
				LinkedEntities: []string{"ent-traf-008", "ent-traf-002", "ent-traf-001", "ent-traf-009"},
			},
			{
				ID:          "ev-traf-006",
				CaseID:      "case-trafic-006",
				Name:        "Certificats d'authenticité falsifiés",
				Type:        models.EvidenceDocumentary,
				Description: "5 certificats signés par Antoine Mercier pour des œuvres volées.",
				Reliability: 9,
				LinkedEntities: []string{"ent-traf-009"},
			},
		},
		Timeline: []models.Event{
			{ID: "evt-t-01", CaseID: "case-trafic-006", Title: "Décès de Victor Moreau", Description: "Jean Moreau hérite de la galerie et des contacts", Timestamp: parseDateTime("2025-08-29T20:30:00"), Location: "Paris", Entities: []string{"ent-traf-001", "ent-traf-006"}, Importance: "high", Verified: true},
			{ID: "evt-t-02", CaseID: "case-trafic-006", Title: "Premier contact Sokolov-Moreau", Description: "Rencontre au Bar Le Diplomate", Timestamp: parseDateTime("2025-09-05T22:00:00"), Location: "Bar Le Diplomate", Entities: []string{"ent-traf-001", "ent-traf-002", "ent-traf-005"}, Importance: "high", Verified: true},
			{ID: "evt-t-03", CaseID: "case-trafic-006", Title: "Première transaction via Roux", Description: "Facture fictive de 500 000€", Timestamp: parseDateTime("2025-09-15T10:00:00"), Location: "Toulouse", Entities: []string{"ent-traf-003", "ent-traf-007"}, Importance: "high", Verified: true},
			{ID: "evt-t-04", CaseID: "case-trafic-006", Title: "Vol au Musée des Arts Premiers", Description: "Statuettes Dogon volées - lien suspecté", Timestamp: parseDateTime("2025-10-02T02:30:00"), Location: "Musée", Entities: []string{"ent-traf-002"}, Importance: "medium", Verified: false},
			{ID: "evt-t-05", CaseID: "case-trafic-006", Title: "Dénonciation anonyme", Description: "Claire Fontaine contacte la police", Timestamp: parseDateTime("2025-10-08T14:00:00"), Location: "Commissariat", Entities: []string{"ent-traf-008"}, Importance: "high", Verified: true},
			{ID: "evt-t-06", CaseID: "case-trafic-006", Title: "Perquisition Galerie L'Éclipse", Description: "Saisie de 3 œuvres volées", Timestamp: parseDateTime("2025-10-12T06:00:00"), Location: "Galerie L'Éclipse", Entities: []string{"ent-traf-004"}, Importance: "high", Verified: true},
			{ID: "evt-t-07", CaseID: "case-trafic-006", Title: "Perquisition domicile Delmas", Description: "3 œuvres volées retrouvées", Timestamp: parseDateTime("2025-10-12T06:30:00"), Location: "Toulouse", Entities: []string{"ent-traf-007"}, Importance: "high", Verified: true},
		},
		Hypotheses: []models.Hypothesis{
			{
				ID:              "hyp-t-01",
				CaseID:          "case-trafic-006",
				Title:           "Jean Moreau impliqué avant décès de Victor",
				Description:     "Jean Moreau aurait utilisé les contacts de son oncle Victor pour monter le réseau. Le décès de Victor pourrait être lié à sa découverte du trafic.",
				Status:          models.HypothesisPending,
				ConfidenceLevel: 70,
				SupportingEvidence: []string{"ev-traf-002"},
				Questions:       []string{"Victor était-il au courant?", "Son décès est-il lié au trafic?", "Qui d'autre dans la galerie était impliqué?"},
				GeneratedBy:     "ai",
			},
			{
				ID:              "hyp-t-02",
				CaseID:          "case-trafic-006",
				Title:           "Réseau international coordonné",
				Description:     "Sokolov dirige un réseau européen. La branche française utilise Roux Constructions pour blanchir et Delmas pour la protection politique.",
				Status:          models.HypothesisSupported,
				ConfidenceLevel: 85,
				SupportingEvidence: []string{"ev-traf-001", "ev-traf-003", "ev-traf-005"},
				Questions:       []string{"Autres branches du réseau?", "Lien avec la disparition de Sophie Laurent qui enquêtait sur Roux?"},
				GeneratedBy:     "user",
			},
			{
				ID:              "hyp-t-03",
				CaseID:          "case-trafic-006",
				Title:           "Disparition Sophie Laurent liée au trafic",
				Description:     "Sophie Laurent enquêtait sur Roux Constructions et Delmas. Elle aurait pu découvrir le lien avec le trafic d'art.",
				Status:          models.HypothesisPending,
				ConfidenceLevel: 60,
				SupportingEvidence: []string{"ev-traf-001"},
				Questions:       []string{"Sophie avait-elle des preuves du trafic?", "Source X connaissait-elle le réseau?"},
				GeneratedBy:     "ai",
			},
		},
		// Contenu N4L natif avec toutes les fonctionnalités SSTorytime
		N4LContent: GetTraficArtN4LContent(),
	}
}

func parseDate(dateStr string) time.Time {
	t, _ := time.Parse("2006-01-02", dateStr)
	return t
}

func parseDateTime(dateTimeStr string) time.Time {
	// Format sans timezone: 2025-07-15T14:00:00
	t, err := time.Parse("2006-01-02T15:04:05", dateTimeStr)
	if err != nil {
		// Essayer RFC3339 si le premier format échoue
		t, _ = time.Parse(time.RFC3339, dateTimeStr)
	}
	return t
}
