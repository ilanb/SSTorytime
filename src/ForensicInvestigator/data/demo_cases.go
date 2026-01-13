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
		// Nouvelles affaires - Série Marseille
		createAffairePortMarseille(),
		createAffaireTraficDrogue(),
		createAffaireBlanchimentRestaurants(),
		// Série Lyon - Criminalité en col blanc
		createAffaireBiotech(),
		createAffaireDeliDInitie(),
		createAffaireFalsificationDiplomes(),
		// Série Bordeaux - Vin et patrimoine
		createAffaireContrefaconVins(),
		createAffaireSuccessionDomaine(),
		createAffaireVolChateaux(),
		// Série Lille - Industrie et environnement
		createAffairePollutionIndustrielle(),
		createAffaireTraficDechets(),
		createAffaireAccidentMortelUsine(),
		// Série Nice - Côte d'Azur
		createAffaireEscroquerieImmobilier(),
		createAffaireDisparitionYacht(),
		createAffaireCambriolagesVillas(),
		// Série Strasbourg - Frontières
		createAffaireTraficArmes(),
		createAffairePasseursHumains(),
		createAffaireContrebandeCigarettes(),
		// Série Nantes - Maritime
		createAffairePiratageInformatique(),
		createAffaireNaufrageSuspect(),
		createAffaireTraficAnimaux(),
		// Série Rennes - Agriculture
		createAffaireAbattageIllegal(),
		createAffaireContaminationAlimentaire(),
		createAffaireSubventionsFrauduleuses(),
		// Série Toulouse - Aérospatiale (connexion avec disparition)
		createAffaireEspionnageIndustriel(),
		createAffaireSabotageAeronautique(),
		createAffaireCorruptionMarchesPublics(),
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
			{ID: "evt-m-00b", CaseID: "case-moreau-001", Title: "Confrontation tribunal Victor/Élodie", Description: "Élodie menace Victor: 'Il paiera pour ce qu'il m'a fait'", Timestamp: parseDateTime("2025-08-25T10:00:00"), Location: "Tribunal de Commerce", Entities: []string{"ent-moreau-001", "ent-moreau-003", "ent-moreau-016"}, Importance: "high", Verified: true},
			// Semaine du crime
			{ID: "evt-m-01", CaseID: "case-moreau-001", Title: "Visite houleuse de Jean", Description: "Discussion avec Victor sur l'argent. Claque la porte en partant.", Timestamp: parseDateTime("2025-08-27T14:00:00"), Location: "Manoir", Entities: []string{"ent-moreau-001", "ent-moreau-002"}, Importance: "high", Verified: true},
			{ID: "evt-m-01b", CaseID: "case-moreau-001", Title: "Jean vu au Bar Le Diplomate", Description: "Rencontre avec individu non identifié", Timestamp: parseDateTime("2025-08-27T22:00:00"), Location: "Bar Le Diplomate", Entities: []string{"ent-moreau-002", "ent-moreau-011"}, Importance: "medium", Verified: false},
			{ID: "evt-m-02", CaseID: "case-moreau-001", Title: "Victor chez le notaire", Description: "Évoque une modification du testament - veut déshériter Jean", Timestamp: parseDateTime("2025-08-28T10:00:00"), Location: "Étude notariale", Entities: []string{"ent-moreau-001", "ent-moreau-014"}, Importance: "high", Verified: true},
			{ID: "evt-m-02b", CaseID: "case-moreau-001", Title: "Câble caméra sectionné", Description: "Système de surveillance neutralisé", Timestamp: parseDateTime("2025-08-28T18:00:00"), Location: "Manoir", Entities: []string{}, Importance: "high", Verified: true},
			// Jour du crime
			{ID: "evt-m-03", CaseID: "case-moreau-001", Title: "Élodie vue près du manoir", Description: "Aperçue par voisin M. Bertrand à 15h", Timestamp: parseDateTime("2025-08-29T15:00:00"), Location: "Aux abords du manoir", Entities: []string{"ent-moreau-003"}, Importance: "high", Verified: false},
			{ID: "evt-m-04", CaseID: "case-moreau-001", Title: "Fenêtre bibliothèque ouverte", Description: "Observée par le jardinier Robert Duval", Timestamp: parseDateTime("2025-08-29T17:30:00"), Location: "Bibliothèque", Entities: []string{"ent-moreau-005", "ent-moreau-007"}, Importance: "medium", Verified: true},
			{ID: "evt-m-05", CaseID: "case-moreau-001", Title: "Départ du jardinier", Description: "Ferme le portillon à clé", Timestamp: parseDateTime("2025-08-29T18:00:00"), Location: "Jardin", Entities: []string{"ent-moreau-005", "ent-moreau-013"}, Importance: "medium", Verified: true},
			{ID: "evt-m-06", CaseID: "case-moreau-001", Title: "Appel Jean vers Victor", Description: "Durée 3 minutes - contenu inconnu", Timestamp: parseDateTime("2025-08-29T18:45:00"), Location: "Téléphone", Entities: []string{"ent-moreau-001", "ent-moreau-002"}, Importance: "high", Verified: true},
			{ID: "evt-m-07", CaseID: "case-moreau-001", Title: "Départ Madame Chen", Description: "Laisse Victor seul après avoir servi le thé", Timestamp: parseDateTime("2025-08-29T19:00:00"), Location: "Manoir", Entities: []string{"ent-moreau-004", "ent-moreau-001"}, Importance: "high", Verified: true},
			{ID: "evt-m-08", CaseID: "case-moreau-001", Title: "Appel numéro inconnu", Description: "Appel entrant sur téléphone Victor - durée 2min", Timestamp: parseDateTime("2025-08-29T19:05:00"), Location: "Téléphone", Entities: []string{"ent-moreau-001", "ent-moreau-012"}, Importance: "high", Verified: true},
			{ID: "evt-m-09", CaseID: "case-moreau-001", Title: "Victor boit son thé", Description: "Dernière activité connue - thé possiblement empoisonné", Timestamp: parseDateTime("2025-08-29T19:15:00"), Location: "Bibliothèque", Entities: []string{"ent-moreau-001", "ent-moreau-007"}, Importance: "high", Verified: false},
			{ID: "evt-m-09b", CaseID: "case-moreau-001", Title: "Entrée cinéma Jean (alibi)", Description: "Ticket acheté pour séance 19h30", Timestamp: parseDateTime("2025-08-29T19:25:00"), Location: "UGC Bercy", Entities: []string{"ent-moreau-002"}, Importance: "high", Verified: true},
			{ID: "evt-m-10", CaseID: "case-moreau-001", Title: "Heure estimée du décès", Description: "Selon rapport médecin légiste Dr. Martin", Timestamp: parseDateTime("2025-08-29T20:30:00"), Location: "Bibliothèque", Entities: []string{"ent-moreau-001", "ent-moreau-017"}, Importance: "high", Verified: true},
			{ID: "evt-m-11", CaseID: "case-moreau-001", Title: "Découverte du corps", Description: "Par Madame Chen revenue au manoir", Timestamp: parseDateTime("2025-08-29T21:45:00"), Location: "Bibliothèque", Entities: []string{"ent-moreau-001", "ent-moreau-004", "ent-moreau-007"}, Importance: "high", Verified: true},
			{ID: "evt-m-12", CaseID: "case-moreau-001", Title: "Arrivée police", Description: "Début de l'enquête officielle", Timestamp: parseDateTime("2025-08-29T22:00:00"), Location: "Manoir", Entities: []string{}, Importance: "medium", Verified: true},
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
	}
}

// ============================================================================
// SÉRIE MARSEILLE - Crime organisé et port
// ============================================================================

// Affaire 7: Meurtre au Port de Marseille
func createAffairePortMarseille() *models.Case {
	return &models.Case{
		ID:          "case-port-007",
		Name:        "Meurtre au Port de Marseille",
		Description: "Corps d'un docker retrouvé dans un conteneur frigorifique. Liens suspectés avec le trafic de drogue transitant par le port.",
		Type:        "homicide",
		Status:      "en_cours",
		CreatedAt:   parseDate("2025-11-02"),
		UpdatedAt:   parseDate("2025-11-08"),
		Entities: []models.Entity{
			{
				ID:          "ent-port-001",
				CaseID:      "case-port-007",
				Name:        "Marco Ferretti",
				Type:        models.EntityPerson,
				Role:        models.RoleVictim,
				Description: "Docker, 45 ans. Travaillait au terminal 3. Retrouvé mort par hypothermie dans conteneur frigorifique.",
				Attributes: map[string]string{
					"age":           "45 ans",
					"profession":    "Docker - Chef d'équipe",
					"employeur":     "Port Autonome de Marseille",
					"anciennete":    "18 ans",
					"cause_deces":   "Hypothermie - enfermement",
					"antecedents":   "Aucun",
				},
				Relations: []models.Relation{
					{ID: "rel-port-010", FromID: "ent-port-001", ToID: "ent-port-002", Type: "conflit", Label: "en conflit avec", Context: "travail", Verified: true},
					{ID: "rel-port-010b", FromID: "ent-port-001", ToID: "ent-port-002", Type: "communication", Label: "a eu une altercation verbale avec", Context: "information", Verified: true},
					{ID: "rel-port-011", FromID: "ent-port-001", ToID: "ent-port-003", Type: "lieu_travail", Label: "travaillait à", Context: "emploi", Verified: true},
					{ID: "rel-port-012", FromID: "ent-port-001", ToID: "ent-port-005", Type: "menace", Label: "aurait menacé de dénoncer", Context: "trafic", Verified: false},
					{ID: "rel-port-012b", FromID: "ent-port-001", ToID: "ent-port-004", Type: "communication", Label: "a envoyé des messages à", Context: "information", Verified: true},
				},
			},
			{
				ID:          "ent-port-002",
				CaseID:      "case-port-007",
				Name:        "Youssef Benali",
				Type:        models.EntityPerson,
				Role:        models.RoleSuspect,
				Description: "Collègue de la victime. Altercation violente 3 jours avant. Problèmes de dettes.",
				Attributes: map[string]string{
					"age":        "38 ans",
					"profession": "Docker",
					"mobile":     "Conflit personnel - dettes",
					"alibi":      "Dit être chez lui",
				},
				Relations: []models.Relation{
					{ID: "rel-port-020", FromID: "ent-port-002", ToID: "ent-port-001", Type: "conflit", Label: "altercation avec", Context: "travail", Verified: true},
					{ID: "rel-port-021", FromID: "ent-port-002", ToID: "ent-port-003", Type: "acces", Label: "avait accès à", Context: "travail", Verified: true},
				},
			},
			{
				ID:          "ent-port-003",
				CaseID:      "case-port-007",
				Name:        "Terminal 3 - Zone Frigorifique",
				Type:        models.EntityPlace,
				Role:        models.RoleOther,
				Description: "Zone de stockage des conteneurs réfrigérés. Accès restreint par badge.",
				Attributes: map[string]string{
					"adresse":   "Port de Marseille Fos, Terminal 3, 13002 Marseille",
					"latitude":  "43.3350",
					"longitude": "5.3425",
					"type":      "Zone portuaire",
				},
				Relations: []models.Relation{
					{ID: "rel-port-030", FromID: "ent-port-003", ToID: "ent-port-001", Type: "scene", Label: "lieu de découverte de", Context: "crime", Verified: true},
				},
			},
			{
				ID:          "ent-port-004",
				CaseID:      "case-port-007",
				Name:        "Karim Messaoudi",
				Type:        models.EntityPerson,
				Role:        models.RoleSuspect,
				Description: "Chef de la sécurité du terminal. Soupçonné de faciliter des trafics.",
				Attributes: map[string]string{
					"age":        "52 ans",
					"profession": "Chef sécurité",
					"suspicion":  "Corruption - ferme les yeux sur conteneurs",
				},
				Relations: []models.Relation{
					{ID: "rel-port-001", FromID: "ent-port-004", ToID: "ent-port-005", Type: "complice", Label: "facilite les opérations de", Context: "trafic", Verified: false},
					{ID: "rel-port-001b", FromID: "ent-port-005", ToID: "ent-port-004", Type: "argent", Label: "paie des pots-de-vin à", Context: "corruption", Verified: false},
					{ID: "rel-port-001c", FromID: "ent-port-004", ToID: "ent-port-003", Type: "controle", Label: "contrôle l'accès à", Context: "securite", Verified: true},
				},
			},
			{
				ID:          "ent-port-005",
				CaseID:      "case-port-007",
				Name:        "Clan Ferrara",
				Type:        models.EntityOrg,
				Role:        models.RoleSuspect,
				Description: "Organisation criminelle marseillaise. Contrôle une partie du trafic de stupéfiants.",
				Relations: []models.Relation{
					{ID: "rel-port-050", FromID: "ent-port-005", ToID: "ent-port-003", Type: "exploitation", Label: "utilise pour trafic", Context: "trafic", Verified: false},
					{ID: "rel-port-051", FromID: "ent-port-005", ToID: "ent-port-001", Type: "elimination", Label: "aurait commandité l'élimination de", Context: "crime", Verified: false},
					{ID: "rel-port-052", FromID: "ent-port-005", ToID: "ent-port-002", Type: "influence", Label: "ordonne et contrôle", Context: "organisation", Verified: false},
					{ID: "rel-port-053", FromID: "ent-port-005", ToID: "ent-port-004", Type: "argent", Label: "effectue des virements à", Context: "corruption", Verified: false},
				},
			},
		},
		Evidence: []models.Evidence{
			{
				ID:          "ev-port-001",
				CaseID:      "case-port-007",
				Name:        "Badge d'accès victime",
				Type:        models.EvidencePhysical,
				Description: "Badge retrouvé à 50m du conteneur. Dernier accès: 01/11 à 23h47.",
				Reliability: 9,
			},
			{
				ID:          "ev-port-002",
				CaseID:      "case-port-007",
				Name:        "Vidéosurveillance",
				Type:        models.EvidenceDigital,
				Description: "Images montrant la victime suivie par une silhouette. Qualité médiocre.",
				Reliability: 6,
			},
			{
				ID:          "ev-port-003",
				CaseID:      "case-port-007",
				Name:        "Téléphone de la victime",
				Type:        models.EvidenceDigital,
				Description: "SMS cryptés vers numéro prépayé. Mentions de 'livraison' et 'problème'.",
				Reliability: 8,
			},
		},
		Timeline: []models.Event{
			{ID: "evt-port-01", CaseID: "case-port-007", Title: "Altercation Ferretti-Benali", Timestamp: parseDateTime("2025-10-29T14:30:00"), Location: "Vestiaires Terminal 3", Importance: "high", Verified: true},
			{ID: "evt-port-02", CaseID: "case-port-007", Title: "Dernier badgeage Ferretti", Timestamp: parseDateTime("2025-11-01T23:47:00"), Location: "Terminal 3", Importance: "high", Verified: true},
			{ID: "evt-port-03", CaseID: "case-port-007", Title: "Découverte du corps", Timestamp: parseDateTime("2025-11-02T06:15:00"), Location: "Conteneur C-4521", Importance: "high", Verified: true},
		},
		Hypotheses: []models.Hypothesis{
			{
				ID:              "hyp-port-01",
				CaseID:          "case-port-007",
				Title:           "Règlement de comptes personnel",
				Description:     "Benali aurait enfermé Ferretti suite à leur altercation.",
				Status:          models.HypothesisPending,
				ConfidenceLevel: 45,
				GeneratedBy:     "user",
			},
			{
				ID:              "hyp-port-02",
				CaseID:          "case-port-007",
				Title:           "Élimination par le clan Ferrara",
				Description:     "Ferretti aurait découvert ou menacé de révéler le trafic.",
				Status:          models.HypothesisPending,
				ConfidenceLevel: 65,
				GeneratedBy:     "ai",
			},
		},
	}
}

// Affaire 8: Trafic de Drogue - Réseau Méditerranéen
func createAffaireTraficDrogue() *models.Case {
	return &models.Case{
		ID:          "case-drogue-008",
		Name:        "Trafic Cocaïne - Réseau Méditerranéen",
		Description: "Démantèlement d'un réseau d'importation de cocaïne depuis l'Amérique du Sud. Connexions avec le port de Marseille et le milieu de la restauration.",
		Type:        "trafic",
		Status:      "en_cours",
		CreatedAt:   parseDate("2025-10-15"),
		UpdatedAt:   parseDate("2025-11-10"),
		Entities: []models.Entity{
			{
				ID:          "ent-drogue-001",
				CaseID:      "case-drogue-008",
				Name:        "Clan Ferrara",
				Type:        models.EntityOrg,
				Role:        models.RoleSuspect,
				Description: "Organisation criminelle marseillaise. Réseau d'importation depuis Colombie.",
				Attributes: map[string]string{
					"effectif":   "Estimé à 40 membres",
					"territoire": "Marseille Nord et Port",
					"activite":   "Cocaïne, cannabis",
				},
				Relations: []models.Relation{
					{ID: "rel-dr-010", FromID: "ent-drogue-001", ToID: "ent-drogue-003", Type: "corruption", Label: "paie pour corruption", Context: "trafic", Verified: true},
					{ID: "rel-dr-010b", FromID: "ent-drogue-001", ToID: "ent-drogue-003", Type: "argent", Label: "effectue des virements vers", Context: "corruption", Verified: true},
					{ID: "rel-dr-011", FromID: "ent-drogue-001", ToID: "ent-drogue-004", Type: "propriete", Label: "contrôle et dirige", Context: "blanchiment", Verified: true},
					{ID: "rel-dr-011b", FromID: "ent-drogue-001", ToID: "ent-drogue-004", Type: "argent", Label: "transfère des fonds vers", Context: "blanchiment", Verified: true},
				},
			},
			{
				ID:          "ent-drogue-002",
				CaseID:      "case-drogue-008",
				Name:        "Antonio Ferrara",
				Type:        models.EntityPerson,
				Role:        models.RoleSuspect,
				Description: "Chef présumé du clan. 58 ans. Façade: import-export alimentaire.",
				Attributes: map[string]string{
					"age":        "58 ans",
					"profession": "Gérant SARL Import-Export",
					"casier":     "2 condamnations - relaxé 2019",
				},
				Relations: []models.Relation{
					{ID: "rel-dr-001", FromID: "ent-drogue-002", ToID: "ent-drogue-001", Type: "direction", Label: "dirige et contrôle", Context: "crime", Verified: true},
					{ID: "rel-dr-001b", FromID: "ent-drogue-002", ToID: "ent-drogue-001", Type: "communication", Label: "ordonne par téléphone chiffré", Context: "organisation", Verified: true},
					{ID: "rel-dr-002", FromID: "ent-drogue-002", ToID: "ent-drogue-004", Type: "blanchiment", Label: "blanchit de l'argent via", Context: "finances", Verified: true},
					{ID: "rel-dr-002b", FromID: "ent-drogue-002", ToID: "ent-drogue-003", Type: "argent", Label: "paie en espèces", Context: "corruption", Verified: true},
				},
			},
			{
				ID:          "ent-drogue-003",
				CaseID:      "case-drogue-008",
				Name:        "Karim Messaoudi",
				Type:        models.EntityPerson,
				Role:        models.RoleSuspect,
				Description: "Chef sécurité port. Facilite l'entrée des conteneurs sans contrôle.",
				Attributes: map[string]string{
					"age":           "52 ans",
					"profession":    "Chef sécurité port",
					"remuneration":  "50 000€ par livraison",
				},
				Relations: []models.Relation{
					{ID: "rel-dr-030", FromID: "ent-drogue-003", ToID: "ent-drogue-001", Type: "complice", Label: "facilite les opérations de", Context: "trafic", Verified: true},
					{ID: "rel-dr-031", FromID: "ent-drogue-003", ToID: "ent-drogue-002", Type: "contact", Label: "appelle par téléphone", Context: "trafic", Verified: true},
					{ID: "rel-dr-031b", FromID: "ent-drogue-003", ToID: "ent-drogue-002", Type: "communication", Label: "rencontre secrètement", Context: "organisation", Verified: true},
					{ID: "rel-dr-032", FromID: "ent-drogue-002", ToID: "ent-drogue-003", Type: "argent", Label: "effectue des transactions vers", Context: "corruption", Verified: true},
				},
			},
			{
				ID:          "ent-drogue-004",
				CaseID:      "case-drogue-008",
				Name:        "Restaurant La Bonne Table",
				Type:        models.EntityPlace,
				Role:        models.RoleSuspect,
				Description: "Restaurant servant de façade pour blanchiment. Chiffre d'affaires suspect.",
				Attributes: map[string]string{
					"adresse":   "45 Quai du Port, 13002 Marseille",
					"latitude":  "43.2965",
					"longitude": "5.3698",
					"type":      "Restaurant",
				},
				Relations: []models.Relation{
					{ID: "rel-dr-040", FromID: "ent-drogue-004", ToID: "ent-drogue-002", Type: "gerance", Label: "géré par", Context: "blanchiment", Verified: true},
				},
			},
		},
		Evidence: []models.Evidence{
			{
				ID:          "ev-drogue-001",
				CaseID:      "case-drogue-008",
				Name:        "Saisie de 200kg cocaïne",
				Type:        models.EvidencePhysical,
				Description: "Cachée dans conteneur de fruits exotiques. Valeur: 8 millions €.",
				Reliability: 10,
			},
			{
				ID:          "ev-drogue-002",
				CaseID:      "case-drogue-008",
				Name:        "Écoutes téléphoniques",
				Type:        models.EvidenceDigital,
				Description: "6 mois de surveillance. Conversations codées entre Ferrara et fournisseurs.",
				Reliability: 9,
			},
		},
		Timeline: []models.Event{
			{ID: "evt-dr-01", CaseID: "case-drogue-008", Title: "Début surveillance", Timestamp: parseDateTime("2025-04-01T00:00:00"), Importance: "medium", Verified: true},
			{ID: "evt-dr-02", CaseID: "case-drogue-008", Title: "Saisie conteneur", Timestamp: parseDateTime("2025-10-14T04:30:00"), Location: "Terminal 3", Importance: "high", Verified: true},
			{ID: "evt-dr-03", CaseID: "case-drogue-008", Title: "Interpellations", Timestamp: parseDateTime("2025-10-15T06:00:00"), Location: "Marseille", Importance: "high", Verified: true},
		},
		Hypotheses: []models.Hypothesis{
			{
				ID:              "hyp-dr-01",
				CaseID:          "case-drogue-008",
				Title:           "Complicité au port",
				Description:     "Messaoudi n'est pas le seul complice. D'autres agents sont impliqués.",
				Status:          models.HypothesisPending,
				ConfidenceLevel: 75,
				GeneratedBy:     "user",
			},
		},
	}
}

// Affaire 9: Blanchiment via Restaurants
func createAffaireBlanchimentRestaurants() *models.Case {
	return &models.Case{
		ID:          "case-blanch-009",
		Name:        "Blanchiment - Chaîne de Restaurants",
		Description: "Réseau de 8 restaurants utilisés pour blanchir l'argent du trafic de drogue. Connexion avec le Clan Ferrara.",
		Type:        "fraude",
		Status:      "en_cours",
		CreatedAt:   parseDate("2025-09-20"),
		UpdatedAt:   parseDate("2025-11-05"),
		Entities: []models.Entity{
			{
				ID:          "ent-blanch-001",
				CaseID:      "case-blanch-009",
				Name:        "Groupe Saveurs du Sud",
				Type:        models.EntityOrg,
				Role:        models.RoleSuspect,
				Description: "Holding détenant 8 restaurants. Chiffre d'affaires gonflé de 200%.",
				Attributes: map[string]string{
					"restaurants":  "8 établissements",
					"ca_declare":   "12 millions €/an",
					"ca_reel":      "Estimé 4 millions €",
				},
				Relations: []models.Relation{
					{ID: "rel-bl-010", FromID: "ent-blanch-001", ToID: "ent-blanch-003", Type: "comptabilite", Label: "comptabilité gérée par", Context: "affaires", Verified: true},
				},
			},
			{
				ID:          "ent-blanch-002",
				CaseID:      "case-blanch-009",
				Name:        "Lucia Ferrara",
				Type:        models.EntityPerson,
				Role:        models.RoleSuspect,
				Description: "Gérante du groupe. Sœur d'Antonio Ferrara. Façade légale de la famille.",
				Attributes: map[string]string{
					"age":         "52 ans",
					"profession":  "Restauratrice",
					"lien_familial": "Sœur d'Antonio Ferrara",
				},
				Relations: []models.Relation{
					{ID: "rel-bl-001", FromID: "ent-blanch-002", ToID: "ent-blanch-001", Type: "direction", Label: "dirige", Context: "affaires", Verified: true},
				},
			},
			{
				ID:          "ent-blanch-003",
				CaseID:      "case-blanch-009",
				Name:        "Cabinet Comptable Méditerranée",
				Type:        models.EntityOrg,
				Role:        models.RoleSuspect,
				Description: "Cabinet gérant la comptabilité. Suspicion de complicité.",
				Relations: []models.Relation{
					{ID: "rel-bl-030", FromID: "ent-blanch-003", ToID: "ent-blanch-001", Type: "complice", Label: "falsifie les comptes de", Context: "blanchiment", Verified: false},
					{ID: "rel-bl-031", FromID: "ent-blanch-003", ToID: "ent-blanch-002", Type: "connivence", Label: "en connivence avec", Context: "blanchiment", Verified: false},
				},
			},
		},
		Evidence: []models.Evidence{
			{
				ID:          "ev-blanch-001",
				CaseID:      "case-blanch-009",
				Name:        "Analyse comptable",
				Type:        models.EvidenceDocumentary,
				Description: "Écarts impossibles entre achats et ventes. Tickets de caisse fictifs.",
				Reliability: 9,
			},
			{
				ID:          "ev-blanch-002",
				CaseID:      "case-blanch-009",
				Name:        "Filature équipe",
				Type:        models.EvidenceDigital,
				Description: "Livraisons nocturnes d'argent liquide depuis entrepôt Ferrara.",
				Reliability: 8,
			},
		},
		Timeline: []models.Event{
			{ID: "evt-bl-01", CaseID: "case-blanch-009", Title: "Signalement TRACFIN", Timestamp: parseDateTime("2025-06-15T00:00:00"), Importance: "high", Verified: true},
			{ID: "evt-bl-02", CaseID: "case-blanch-009", Title: "Début enquête", Timestamp: parseDateTime("2025-09-20T00:00:00"), Importance: "medium", Verified: true},
		},
		Hypotheses: []models.Hypothesis{
			{
				ID:              "hyp-bl-01",
				CaseID:          "case-blanch-009",
				Title:           "Expert-comptable complice",
				Description:     "Le cabinet comptable falsifie sciemment les comptes.",
				Status:          models.HypothesisPending,
				ConfidenceLevel: 70,
				GeneratedBy:     "user",
			},
		},
	}
}

// ============================================================================
// SÉRIE LYON - Criminalité en col blanc
// ============================================================================

// Affaire 10: Scandale Biotech
func createAffaireBiotech() *models.Case {
	return &models.Case{
		ID:          "case-biotech-010",
		Name:        "Scandale BioGenix - Essais Cliniques Falsifiés",
		Description: "Falsification de résultats d'essais cliniques pour un médicament anticancéreux. 3 décès suspects.",
		Type:        "fraude",
		Status:      "en_cours",
		CreatedAt:   parseDate("2025-08-10"),
		UpdatedAt:   parseDate("2025-11-01"),
		Entities: []models.Entity{
			{
				ID:          "ent-bio-001",
				CaseID:      "case-biotech-010",
				Name:        "BioGenix Pharmaceuticals",
				Type:        models.EntityOrg,
				Role:        models.RoleSuspect,
				Description: "Laboratoire pharmaceutique. Capitalisation: 2 milliards €. Cotée au CAC Small.",
				Attributes: map[string]string{
					"secteur":        "Pharmaceutique - Oncologie",
					"effectif":       "1200 employés",
					"capitalisation": "2 milliards €",
				},
				Relations: []models.Relation{
					{ID: "rel-bio-010", FromID: "ent-bio-001", ToID: "ent-bio-004", Type: "fabricant", Label: "fabrique", Context: "pharmaceutique", Verified: true},
				},
			},
			{
				ID:          "ent-bio-002",
				CaseID:      "case-biotech-010",
				Name:        "Dr. François Lemaire",
				Type:        models.EntityPerson,
				Role:        models.RoleSuspect,
				Description: "PDG de BioGenix. Aurait ordonné la falsification pour sauver l'entreprise.",
				Attributes: map[string]string{
					"age":        "62 ans",
					"profession": "PDG - Médecin",
					"mobile":     "Sauver l'entreprise et ses stock-options",
				},
				Relations: []models.Relation{
					{ID: "rel-bio-020", FromID: "ent-bio-002", ToID: "ent-bio-001", Type: "direction", Label: "dirige", Context: "affaires", Verified: true},
					{ID: "rel-bio-021", FromID: "ent-bio-002", ToID: "ent-bio-004", Type: "falsification", Label: "a ordonné la falsification des tests de", Context: "fraude", Verified: false},
					{ID: "rel-bio-022", FromID: "ent-bio-002", ToID: "ent-bio-003", Type: "pression", Label: "a fait pression sur", Context: "intimidation", Verified: true},
				},
			},
			{
				ID:          "ent-bio-003",
				CaseID:      "case-biotech-010",
				Name:        "Dr. Marie Dumont",
				Type:        models.EntityPerson,
				Role:        models.RoleWitness,
				Description: "Chercheuse lanceuse d'alerte. A dénoncé les falsifications.",
				Attributes: map[string]string{
					"age":        "41 ans",
					"profession": "Chercheuse en oncologie",
					"statut":     "Protection lanceur d'alerte",
				},
				Relations: []models.Relation{
					{ID: "rel-bio-030", FromID: "ent-bio-003", ToID: "ent-bio-001", Type: "emploi", Label: "employée de", Context: "travail", Verified: true},
					{ID: "rel-bio-031", FromID: "ent-bio-003", ToID: "ent-bio-004", Type: "recherche", Label: "travaillait sur", Context: "recherche", Verified: true},
					{ID: "rel-bio-032", FromID: "ent-bio-003", ToID: "ent-bio-002", Type: "denonciation", Label: "a dénoncé", Context: "alerte", Verified: true},
				},
			},
			{
				ID:          "ent-bio-004",
				CaseID:      "case-biotech-010",
				Name:        "GenoCure-X",
				Type:        models.EntityObject,
				Role:        models.RoleOther,
				Description: "Médicament anticancéreux. Phase 3 falsifiée. 3 décès potentiellement liés.",
				Relations: []models.Relation{
					{ID: "rel-bio-040", FromID: "ent-bio-004", ToID: "ent-bio-001", Type: "produit", Label: "produit par", Context: "pharmaceutique", Verified: true},
				},
			},
		},
		Evidence: []models.Evidence{
			{
				ID:          "ev-bio-001",
				CaseID:      "case-biotech-010",
				Name:        "Emails internes",
				Type:        models.EvidenceDigital,
				Description: "Ordres de Lemaire pour modifier les résultats. 'Ajustez les chiffres'.",
				Reliability: 9,
			},
			{
				ID:          "ev-bio-002",
				CaseID:      "case-biotech-010",
				Name:        "Dossiers patients originaux",
				Type:        models.EvidenceDocumentary,
				Description: "Comparaison avant/après falsification. Effets secondaires masqués.",
				Reliability: 10,
			},
			{
				ID:          "ev-bio-003",
				CaseID:      "case-biotech-010",
				Name:        "Certificats de décès",
				Type:        models.EvidenceDocumentary,
				Description: "3 patients décédés pendant les essais. Causes réelles masquées.",
				Reliability: 8,
			},
		},
		Timeline: []models.Event{
			{ID: "evt-bio-01", CaseID: "case-biotech-010", Title: "Début essais Phase 3", Timestamp: parseDateTime("2024-09-01T00:00:00"), Importance: "medium", Verified: true},
			{ID: "evt-bio-02", CaseID: "case-biotech-010", Title: "Premier décès suspect", Timestamp: parseDateTime("2025-02-15T00:00:00"), Importance: "high", Verified: true},
			{ID: "evt-bio-03", CaseID: "case-biotech-010", Title: "Alerte Dr. Dumont", Timestamp: parseDateTime("2025-08-01T00:00:00"), Importance: "high", Verified: true},
			{ID: "evt-bio-04", CaseID: "case-biotech-010", Title: "Suspension autorisation", Timestamp: parseDateTime("2025-08-10T00:00:00"), Importance: "high", Verified: true},
		},
		Hypotheses: []models.Hypothesis{
			{
				ID:              "hyp-bio-01",
				CaseID:          "case-biotech-010",
				Title:           "Homicide involontaire aggravé",
				Description:     "Les 3 décès sont directement imputables aux falsifications.",
				Status:          models.HypothesisPending,
				ConfidenceLevel: 75,
				GeneratedBy:     "user",
			},
			{
				ID:              "hyp-bio-02",
				CaseID:          "case-biotech-010",
				Title:           "Délit d'initié associé",
				Description:     "Lemaire aurait vendu ses actions avant l'alerte.",
				Status:          models.HypothesisPending,
				ConfidenceLevel: 60,
				GeneratedBy:     "ai",
			},
		},
	}
}

// Affaire 11: Délit d'Initié
func createAffaireDeliDInitie() *models.Case {
	return &models.Case{
		ID:          "case-initie-011",
		Name:        "Délit d'Initié - BioGenix",
		Description: "Ventes massives d'actions BioGenix avant l'annonce du scandale. Gains illicites estimés à 15 millions €.",
		Type:        "fraude",
		Status:      "en_cours",
		CreatedAt:   parseDate("2025-08-20"),
		UpdatedAt:   parseDate("2025-11-05"),
		Entities: []models.Entity{
			{
				ID:          "ent-init-001",
				CaseID:      "case-initie-011",
				Name:        "Dr. François Lemaire",
				Type:        models.EntityPerson,
				Role:        models.RoleSuspect,
				Description: "PDG BioGenix. A vendu 80% de ses actions 2 semaines avant le scandale.",
				Attributes: map[string]string{
					"actions_vendues": "450 000 actions",
					"gain":            "8.5 millions €",
					"date_vente":      "25/07/2025",
				},
				Relations: []models.Relation{
					{ID: "rel-init-010", FromID: "ent-init-001", ToID: "ent-init-002", Type: "coordination", Label: "s'est coordonné avec", Context: "delit", Verified: false},
					{ID: "rel-init-011", FromID: "ent-init-001", ToID: "ent-init-003", Type: "information", Label: "a transmis des informations privilégiées à", Context: "delit", Verified: false},
				},
			},
			{
				ID:          "ent-init-002",
				CaseID:      "case-initie-011",
				Name:        "Hedge Fund Alpina Capital",
				Type:        models.EntityOrg,
				Role:        models.RoleSuspect,
				Description: "Fonds d'investissement. Ventes massives coordonnées avec Lemaire.",
				Attributes: map[string]string{
					"siege":  "Luxembourg",
					"actifs": "3 milliards €",
				},
				Relations: []models.Relation{
					{ID: "rel-init-020", FromID: "ent-init-002", ToID: "ent-init-001", Type: "beneficiaire", Label: "a bénéficié d'informations de", Context: "delit", Verified: false},
				},
			},
			{
				ID:          "ent-init-003",
				CaseID:      "case-initie-011",
				Name:        "Julien Mercier",
				Type:        models.EntityPerson,
				Role:        models.RoleSuspect,
				Description: "Gérant Alpina Capital. Ami de longue date de Lemaire. Golf ensemble.",
				Attributes: map[string]string{
					"age":      "55 ans",
					"relation": "Ami de Lemaire depuis 20 ans",
				},
				Relations: []models.Relation{
					{ID: "rel-init-001", FromID: "ent-init-003", ToID: "ent-init-001", Type: "ami", Label: "ami de", Context: "personnel", Verified: true},
					{ID: "rel-init-002", FromID: "ent-init-003", ToID: "ent-init-002", Type: "direction", Label: "dirige", Context: "affaires", Verified: true},
				},
			},
		},
		Evidence: []models.Evidence{
			{
				ID:          "ev-init-001",
				CaseID:      "case-initie-011",
				Name:        "Ordres de bourse",
				Type:        models.EvidenceDocumentary,
				Description: "Ventes massives entre le 20 et 28 juillet 2025. Timing suspect.",
				Reliability: 10,
			},
			{
				ID:          "ev-init-002",
				CaseID:      "case-initie-011",
				Name:        "Appels téléphoniques",
				Type:        models.EvidenceDigital,
				Description: "12 appels entre Lemaire et Mercier en juillet. Durée totale: 4h.",
				Reliability: 8,
			},
		},
		Timeline: []models.Event{
			{ID: "evt-init-01", CaseID: "case-initie-011", Title: "Premiers appels Lemaire-Mercier", Timestamp: parseDateTime("2025-07-15T00:00:00"), Importance: "high", Verified: true},
			{ID: "evt-init-02", CaseID: "case-initie-011", Title: "Début ventes Alpina", Timestamp: parseDateTime("2025-07-20T09:00:00"), Importance: "high", Verified: true},
			{ID: "evt-init-03", CaseID: "case-initie-011", Title: "Vente actions Lemaire", Timestamp: parseDateTime("2025-07-25T10:30:00"), Importance: "high", Verified: true},
			{ID: "evt-init-04", CaseID: "case-initie-011", Title: "Annonce scandale", Timestamp: parseDateTime("2025-08-10T08:00:00"), Importance: "high", Verified: true},
		},
		Hypotheses: []models.Hypothesis{
			{
				ID:              "hyp-init-01",
				CaseID:          "case-initie-011",
				Title:           "Complicité Lemaire-Mercier",
				Description:     "Lemaire a informé Mercier du scandale à venir pour profits mutuels.",
				Status:          models.HypothesisSupported,
				ConfidenceLevel: 85,
				GeneratedBy:     "user",
			},
		},
	}
}

// Affaire 12: Falsification de Diplômes
func createAffaireFalsificationDiplomes() *models.Case {
	return &models.Case{
		ID:          "case-diplome-012",
		Name:        "Réseau de Faux Diplômes Médicaux",
		Description: "Réseau vendant de faux diplômes de médecine. 47 faux médecins identifiés exerçant en France.",
		Type:        "fraude",
		Status:      "en_cours",
		CreatedAt:   parseDate("2025-07-01"),
		UpdatedAt:   parseDate("2025-10-20"),
		Entities: []models.Entity{
			{
				ID:          "ent-dipl-001",
				CaseID:      "case-diplome-012",
				Name:        "Réseau Hippocrate",
				Type:        models.EntityOrg,
				Role:        models.RoleSuspect,
				Description: "Réseau international de falsification de diplômes médicaux.",
				Attributes: map[string]string{
					"pays":        "France, Roumanie, Maroc",
					"diplomes":    "Médecine, Pharmacie, Dentaire",
					"tarif":       "15 000 à 50 000€",
				},
				Relations: []models.Relation{
					{ID: "rel-dipl-010", FromID: "ent-dipl-001", ToID: "ent-dipl-002", Type: "direction", Label: "dirigé par", Context: "organisation", Verified: true},
					{ID: "rel-dipl-011", FromID: "ent-dipl-001", ToID: "ent-dipl-003", Type: "infiltration", Label: "a infiltré", Context: "fraude", Verified: true},
				},
			},
			{
				ID:          "ent-dipl-002",
				CaseID:      "case-diplome-012",
				Name:        "Dr. Bogdan Ionescu",
				Type:        models.EntityPerson,
				Role:        models.RoleSuspect,
				Description: "Chef présumé du réseau. Ancien universitaire roumain radié.",
				Attributes: map[string]string{
					"age":         "58 ans",
					"nationalite": "Roumaine",
					"statut":      "Radié de l'université de Bucarest",
				},
				Relations: []models.Relation{
					{ID: "rel-dipl-020", FromID: "ent-dipl-002", ToID: "ent-dipl-001", Type: "direction", Label: "dirige", Context: "organisation", Verified: true},
				},
			},
			{
				ID:          "ent-dipl-003",
				CaseID:      "case-diplome-012",
				Name:        "Clinique du Parc - Lyon",
				Type:        models.EntityPlace,
				Role:        models.RoleOther,
				Description: "Clinique employant 3 faux médecins. 2 décès en cours d'investigation.",
				Attributes: map[string]string{
					"adresse":   "155 Boulevard de Stalingrad, 69006 Lyon",
					"latitude":  "45.7676",
					"longitude": "4.8567",
					"type":      "Clinique privée",
				},
				Relations: []models.Relation{
					{ID: "rel-dipl-030", FromID: "ent-dipl-003", ToID: "ent-dipl-001", Type: "victime", Label: "victime de", Context: "fraude", Verified: true},
				},
			},
		},
		Evidence: []models.Evidence{
			{
				ID:          "ev-dipl-001",
				CaseID:      "case-diplome-012",
				Name:        "Faux diplômes saisis",
				Type:        models.EvidencePhysical,
				Description: "127 diplômes vierges et 43 tampons officiels falsifiés.",
				Reliability: 10,
			},
			{
				ID:          "ev-dipl-002",
				CaseID:      "case-diplome-012",
				Name:        "Base de données clients",
				Type:        models.EvidenceDigital,
				Description: "Liste de 312 acheteurs dont 47 exerçant en France.",
				Reliability: 9,
			},
		},
		Timeline: []models.Event{
			{ID: "evt-dipl-01", CaseID: "case-diplome-012", Title: "Signalement Ordre des Médecins", Timestamp: parseDateTime("2025-03-15T00:00:00"), Importance: "high", Verified: true},
			{ID: "evt-dipl-02", CaseID: "case-diplome-012", Title: "Interpellation Ionescu", Timestamp: parseDateTime("2025-07-01T06:00:00"), Location: "Lyon", Importance: "high", Verified: true},
		},
		Hypotheses: []models.Hypothesis{
			{
				ID:              "hyp-dipl-01",
				CaseID:          "case-diplome-012",
				Title:           "Homicides involontaires multiples",
				Description:     "Certains faux médecins auraient causé des décès par incompétence.",
				Status:          models.HypothesisPending,
				ConfidenceLevel: 65,
				GeneratedBy:     "user",
			},
		},
	}
}

// ============================================================================
// SÉRIE BORDEAUX - Vin et patrimoine
// ============================================================================

// Affaire 13: Contrefaçon de Grands Crus
func createAffaireContrefaconVins() *models.Case {
	return &models.Case{
		ID:          "case-vin-013",
		Name:        "Contrefaçon Grands Crus Bordelais",
		Description: "Réseau de contrefaçon de vins de prestige. Faux Pétrus et Margaux vendus pour 12 millions €.",
		Type:        "fraude",
		Status:      "en_cours",
		CreatedAt:   parseDate("2025-09-01"),
		UpdatedAt:   parseDate("2025-11-10"),
		Entities: []models.Entity{
			{
				ID:          "ent-vin-001",
				CaseID:      "case-vin-013",
				Name:        "Cave Saint-Émilion Distribution",
				Type:        models.EntityOrg,
				Role:        models.RoleSuspect,
				Description: "Négociant en vin. Façade pour écouler les contrefaçons.",
				Attributes: map[string]string{
					"adresse":   "12 Rue Guadet, 33330 Saint-Émilion",
					"latitude":  "44.8945",
					"longitude": "-0.1549",
					"type":      "Négociant en vin",
				},
				Relations: []models.Relation{
					{ID: "rel-vin-010", FromID: "ent-vin-001", ToID: "ent-vin-002", Type: "emploi", Label: "emploie", Context: "fraude", Verified: true},
					{ID: "rel-vin-011", FromID: "ent-vin-001", ToID: "ent-vin-003", Type: "collaboration", Label: "collabore avec", Context: "fraude", Verified: true},
				},
			},
			{
				ID:          "ent-vin-002",
				CaseID:      "case-vin-013",
				Name:        "Pierre Delacroix",
				Type:        models.EntityPerson,
				Role:        models.RoleSuspect,
				Description: "Ancien sommelier reconverti. Maître faussaire des étiquettes.",
				Attributes: map[string]string{
					"age":        "48 ans",
					"profession": "Ex-sommelier",
					"expertise":  "Reproduction d'étiquettes anciennes",
				},
				Relations: []models.Relation{
					{ID: "rel-vin-020", FromID: "ent-vin-002", ToID: "ent-vin-001", Type: "emploi", Label: "travaille pour", Context: "fraude", Verified: true},
					{ID: "rel-vin-021", FromID: "ent-vin-002", ToID: "ent-vin-003", Type: "collaboration", Label: "collabore avec", Context: "fraude", Verified: true},
				},
			},
			{
				ID:          "ent-vin-003",
				CaseID:      "case-vin-013",
				Name:        "Antoine Mercier",
				Type:        models.EntityPerson,
				Role:        models.RoleSuspect,
				Description: "Expert en vins anciens. Authentifie les faux. Lié au milieu de l'art.",
				Attributes: map[string]string{
					"age":        "55 ans",
					"profession": "Expert - Commissaire priseur",
					"double_jeu": "Aussi impliqué dans trafic art Paris",
				},
				Relations: []models.Relation{
					{ID: "rel-vin-001", FromID: "ent-vin-003", ToID: "ent-vin-001", Type: "complice", Label: "authentifie pour", Context: "fraude", Verified: true},
				},
			},
		},
		Evidence: []models.Evidence{
			{
				ID:          "ev-vin-001",
				CaseID:      "case-vin-013",
				Name:        "Bouteilles saisies",
				Type:        models.EvidencePhysical,
				Description: "458 bouteilles de faux grands crus. Valeur affichée: 3.2 millions €.",
				Reliability: 10,
			},
			{
				ID:          "ev-vin-002",
				CaseID:      "case-vin-013",
				Name:        "Atelier de falsification",
				Type:        models.EvidencePhysical,
				Description: "Imprimerie clandestine avec étiquettes, capsules, bouchons d'époque.",
				Reliability: 10,
			},
		},
		Timeline: []models.Event{
			{ID: "evt-vin-01", CaseID: "case-vin-013", Title: "Plainte Château Pétrus", Timestamp: parseDateTime("2025-06-20T00:00:00"), Importance: "high", Verified: true},
			{ID: "evt-vin-02", CaseID: "case-vin-013", Title: "Perquisition cave", Timestamp: parseDateTime("2025-09-01T06:00:00"), Location: "Saint-Émilion", Importance: "high", Verified: true},
		},
		Hypotheses: []models.Hypothesis{
			{
				ID:              "hyp-vin-01",
				CaseID:          "case-vin-013",
				Title:           "Réseau international",
				Description:     "Les faux sont exportés vers Chine, USA et Russie.",
				Status:          models.HypothesisPending,
				ConfidenceLevel: 70,
				GeneratedBy:     "ai",
			},
		},
	}
}

// Affaire 14: Succession Contestée d'un Domaine
func createAffaireSuccessionDomaine() *models.Case {
	return &models.Case{
		ID:          "case-success-014",
		Name:        "Succession Domaine de Lamarque",
		Description: "Mort suspecte du propriétaire d'un domaine viticole de 80 hectares. Héritage de 45 millions € contesté.",
		Type:        "homicide",
		Status:      "en_cours",
		CreatedAt:   parseDate("2025-10-01"),
		UpdatedAt:   parseDate("2025-11-08"),
		Entities: []models.Entity{
			{
				ID:          "ent-succ-001",
				CaseID:      "case-success-014",
				Name:        "Henri de Lamarque",
				Type:        models.EntityPerson,
				Role:        models.RoleVictim,
				Description: "Propriétaire du domaine. 78 ans. Décédé d'une 'chute' dans ses caves.",
				Attributes: map[string]string{
					"age":         "78 ans",
					"fortune":     "45 millions €",
					"cause_deces": "Traumatisme crânien - chute",
				},
				Relations: []models.Relation{
					{ID: "rel-succ-010", FromID: "ent-succ-001", ToID: "ent-succ-002", Type: "famille", Label: "père de", Context: "famille", Verified: true},
					{ID: "rel-succ-011", FromID: "ent-succ-001", ToID: "ent-succ-003", Type: "emploi", Label: "employeur de", Context: "travail", Verified: true},
				},
			},
			{
				ID:          "ent-succ-002",
				CaseID:      "case-success-014",
				Name:        "Mathilde de Lamarque",
				Type:        models.EntityPerson,
				Role:        models.RoleSuspect,
				Description: "Fille unique. En conflit avec son père sur la gestion du domaine.",
				Attributes: map[string]string{
					"age":      "52 ans",
					"mobile":   "Héritage et contrôle domaine",
					"conflit":  "Voulait vendre à groupe chinois",
				},
				Relations: []models.Relation{
					{ID: "rel-succ-020", FromID: "ent-succ-002", ToID: "ent-succ-001", Type: "famille", Label: "fille de", Context: "famille", Verified: true},
					{ID: "rel-succ-021", FromID: "ent-succ-002", ToID: "ent-succ-001", Type: "conflit", Label: "en conflit avec", Context: "succession", Verified: true},
				},
			},
			{
				ID:          "ent-succ-003",
				CaseID:      "case-success-014",
				Name:        "Régisseur du domaine",
				Type:        models.EntityPerson,
				Role:        models.RoleWitness,
				Description: "Dernier à avoir vu Henri vivant. 30 ans au service de la famille.",
				Relations: []models.Relation{
					{ID: "rel-succ-030", FromID: "ent-succ-003", ToID: "ent-succ-001", Type: "emploi", Label: "employé de", Context: "travail", Verified: true},
					{ID: "rel-succ-031", FromID: "ent-succ-003", ToID: "ent-succ-001", Type: "temoin", Label: "dernier à avoir vu", Context: "enquête", Verified: true},
				},
			},
		},
		Evidence: []models.Evidence{
			{
				ID:          "ev-succ-001",
				CaseID:      "case-success-014",
				Name:        "Rapport autopsie",
				Type:        models.EvidenceForensic,
				Description: "Traces de lutte sur les avant-bras. Chute pas accidentelle.",
				Reliability: 9,
			},
			{
				ID:          "ev-succ-002",
				CaseID:      "case-success-014",
				Name:        "Testament modifié",
				Type:        models.EvidenceDocumentary,
				Description: "Nouveau testament 2 semaines avant décès. Mathilde déshéritée au profit d'une fondation.",
				Reliability: 10,
			},
		},
		Timeline: []models.Event{
			{ID: "evt-succ-01", CaseID: "case-success-014", Title: "Modification testament", Timestamp: parseDateTime("2025-09-15T00:00:00"), Importance: "high", Verified: true},
			{ID: "evt-succ-02", CaseID: "case-success-014", Title: "Dispute père-fille", Timestamp: parseDateTime("2025-09-28T00:00:00"), Importance: "high", Verified: true},
			{ID: "evt-succ-03", CaseID: "case-success-014", Title: "Décès Henri", Timestamp: parseDateTime("2025-10-01T22:30:00"), Location: "Caves du domaine", Importance: "high", Verified: true},
		},
		Hypotheses: []models.Hypothesis{
			{
				ID:              "hyp-succ-01",
				CaseID:          "case-success-014",
				Title:           "Meurtre par héritière",
				Description:     "Mathilde aurait poussé son père pour hériter avant changement testament.",
				Status:          models.HypothesisPending,
				ConfidenceLevel: 70,
				GeneratedBy:     "user",
			},
		},
	}
}

// Affaire 15: Vol dans les Châteaux
func createAffaireVolChateaux() *models.Case {
	return &models.Case{
		ID:          "case-chateaux-015",
		Name:        "Série de Cambriolages - Châteaux Bordelais",
		Description: "12 cambriolages dans des propriétés viticoles. Œuvres d'art et bouteilles de collection. Butin: 8 millions €.",
		Type:        "vol",
		Status:      "en_cours",
		CreatedAt:   parseDate("2025-06-01"),
		UpdatedAt:   parseDate("2025-10-30"),
		Entities: []models.Entity{
			{
				ID:          "ent-chat-001",
				CaseID:      "case-chateaux-015",
				Name:        "Gang des Châteaux",
				Type:        models.EntityOrg,
				Role:        models.RoleSuspect,
				Description: "Équipe professionnelle. Mode opératoire identique sur 12 sites.",
				Relations: []models.Relation{
					{ID: "rel-chat-010", FromID: "ent-chat-001", ToID: "ent-chat-002", Type: "commanditaire", Label: "financé par", Context: "crime", Verified: false},
				},
			},
			{
				ID:          "ent-chat-002",
				CaseID:      "case-chateaux-015",
				Name:        "Viktor Sokolov",
				Type:        models.EntityPerson,
				Role:        models.RoleSuspect,
				Description: "Commanditaire présumé. Réseau russe de recel d'œuvres d'art.",
				Attributes: map[string]string{
					"nationalite": "Russe",
					"alias":       "Le Collectionneur",
					"interpol":    "Notice rouge",
				},
				Relations: []models.Relation{
					{ID: "rel-chat-020", FromID: "ent-chat-002", ToID: "ent-chat-001", Type: "commanditaire", Label: "commandite", Context: "crime", Verified: false},
				},
			},
		},
		Evidence: []models.Evidence{
			{
				ID:          "ev-chat-001",
				CaseID:      "case-chateaux-015",
				Name:        "ADN sur gant",
				Type:        models.EvidenceForensic,
				Description: "Correspondance avec fichier: ancien détenu spécialisé vols.",
				Reliability: 9,
			},
		},
		Timeline: []models.Event{
			{ID: "evt-chat-01", CaseID: "case-chateaux-015", Title: "Premier cambriolage", Timestamp: parseDateTime("2025-06-01T03:00:00"), Location: "Château Margaux", Importance: "high", Verified: true},
			{ID: "evt-chat-02", CaseID: "case-chateaux-015", Title: "Dernier cambriolage", Timestamp: parseDateTime("2025-10-15T02:30:00"), Location: "Château Palmer", Importance: "high", Verified: true},
		},
		Hypotheses: []models.Hypothesis{
			{
				ID:              "hyp-chat-01",
				CaseID:          "case-chateaux-015",
				Title:           "Connexion réseau Sokolov",
				Description:     "Les œuvres sont revendues via le réseau international de Sokolov.",
				Status:          models.HypothesisPending,
				ConfidenceLevel: 75,
				GeneratedBy:     "ai",
			},
		},
	}
}

// ============================================================================
// SÉRIE LILLE - Industrie et environnement
// ============================================================================

// Affaire 16: Pollution Industrielle
func createAffairePollutionIndustrielle() *models.Case {
	return &models.Case{
		ID:          "case-pollution-016",
		Name:        "Pollution Chimique - Usine ChemNord",
		Description: "Déversements illégaux de produits toxiques. 3 cas de cancers dans village voisin. Nappe phréatique contaminée.",
		Type:        "environnement",
		Status:      "en_cours",
		CreatedAt:   parseDate("2025-05-15"),
		UpdatedAt:   parseDate("2025-10-25"),
		Entities: []models.Entity{
			{
				ID:          "ent-poll-001",
				CaseID:      "case-pollution-016",
				Name:        "ChemNord Industries",
				Type:        models.EntityOrg,
				Role:        models.RoleSuspect,
				Description: "Usine chimique. Production de solvants et peintures industrielles.",
				Attributes: map[string]string{
					"effectif":  "340 employés",
					"ca":        "85 millions €",
					"adresse":   "Zone Industrielle des Deux-Synthe, 59640 Dunkerque",
					"latitude":  "51.0156",
					"longitude": "2.3245",
					"type":      "Usine chimique",
				},
				Relations: []models.Relation{
					{ID: "rel-poll-010", FromID: "ent-poll-001", ToID: "ent-poll-002", Type: "direction", Label: "dirigé par", Context: "organisation", Verified: true},
					{ID: "rel-poll-011", FromID: "ent-poll-001", ToID: "ent-poll-003", Type: "prejudice", Label: "a causé préjudice à", Context: "pollution", Verified: true},
				},
			},
			{
				ID:          "ent-poll-002",
				CaseID:      "case-pollution-016",
				Name:        "Gérard Dupont",
				Type:        models.EntityPerson,
				Role:        models.RoleSuspect,
				Description: "Directeur de site. A ordonné les déversements nocturnes.",
				Relations: []models.Relation{
					{ID: "rel-poll-020", FromID: "ent-poll-002", ToID: "ent-poll-001", Type: "direction", Label: "dirige", Context: "organisation", Verified: true},
					{ID: "rel-poll-021", FromID: "ent-poll-002", ToID: "ent-poll-003", Type: "responsable", Label: "responsable des dommages à", Context: "pollution", Verified: false},
				},
			},
			{
				ID:          "ent-poll-003",
				CaseID:      "case-pollution-016",
				Name:        "Association Victimes ChemNord",
				Type:        models.EntityOrg,
				Role:        models.RoleOther,
				Description: "127 habitants riverains. Plainte collective déposée.",
				Relations: []models.Relation{
					{ID: "rel-poll-030", FromID: "ent-poll-003", ToID: "ent-poll-001", Type: "plainte", Label: "a porté plainte contre", Context: "juridique", Verified: true},
				},
			},
		},
		Evidence: []models.Evidence{
			{
				ID:          "ev-poll-001",
				CaseID:      "case-pollution-016",
				Name:        "Analyses eau",
				Type:        models.EvidenceForensic,
				Description: "Taux de solvants 200x supérieurs aux normes dans puits voisins.",
				Reliability: 10,
			},
			{
				ID:          "ev-poll-002",
				CaseID:      "case-pollution-016",
				Name:        "Registres falsifiés",
				Type:        models.EvidenceDocumentary,
				Description: "Bordereaux de destruction de déchets falsifiés depuis 2020.",
				Reliability: 9,
			},
		},
		Timeline: []models.Event{
			{ID: "evt-poll-01", CaseID: "case-pollution-016", Title: "Premier signalement", Timestamp: parseDateTime("2025-02-10T00:00:00"), Importance: "medium", Verified: true},
			{ID: "evt-poll-02", CaseID: "case-pollution-016", Title: "Analyses DREAL", Timestamp: parseDateTime("2025-05-15T00:00:00"), Importance: "high", Verified: true},
		},
		Hypotheses: []models.Hypothesis{
			{
				ID:              "hyp-poll-01",
				CaseID:          "case-pollution-016",
				Title:           "Pollution depuis 2018",
				Description:     "Les déversements auraient commencé bien avant 2020.",
				Status:          models.HypothesisPending,
				ConfidenceLevel: 60,
				GeneratedBy:     "ai",
			},
		},
	}
}

// Affaire 17: Trafic de Déchets
func createAffaireTraficDechets() *models.Case {
	return &models.Case{
		ID:          "case-dechets-017",
		Name:        "Trafic International de Déchets Toxiques",
		Description: "Export illégal de déchets industriels vers l'Afrique. 45 conteneurs identifiés.",
		Type:        "trafic",
		Status:      "en_cours",
		CreatedAt:   parseDate("2025-08-01"),
		UpdatedAt:   parseDate("2025-11-01"),
		Entities: []models.Entity{
			{
				ID:          "ent-dech-001",
				CaseID:      "case-dechets-017",
				Name:        "EcoRecycle SARL",
				Type:        models.EntityOrg,
				Role:        models.RoleSuspect,
				Description: "Société de traitement de déchets. Façade pour export illégal.",
				Relations: []models.Relation{
					{ID: "rel-dech-010", FromID: "ent-dech-001", ToID: "ent-dech-002", Type: "utilisation", Label: "utilise", Context: "trafic", Verified: true},
					{ID: "rel-dech-011", FromID: "ent-dech-001", ToID: "ent-dech-003", Type: "client", Label: "fournisseur de", Context: "déchets", Verified: true},
				},
			},
			{
				ID:          "ent-dech-002",
				CaseID:      "case-dechets-017",
				Name:        "Port de Dunkerque",
				Type:        models.EntityPlace,
				Role:        models.RoleOther,
				Description: "Point de départ des conteneurs. Complicités portuaires suspectées.",
				Attributes: map[string]string{
					"adresse":   "Port de Dunkerque, 59140 Dunkerque",
					"latitude":  "51.0343",
					"longitude": "2.3767",
					"type":      "Port maritime",
				},
				Relations: []models.Relation{
					{ID: "rel-dech-020", FromID: "ent-dech-002", ToID: "ent-dech-001", Type: "complicite", Label: "facilite les opérations de", Context: "trafic", Verified: false},
				},
			},
			{
				ID:          "ent-dech-003",
				CaseID:      "case-dechets-017",
				Name:        "ChemNord Industries",
				Type:        models.EntityOrg,
				Role:        models.RoleSuspect,
				Description: "Client principal. Économise 2 millions/an sur traitement légal.",
				Relations: []models.Relation{
					{ID: "rel-dech-001", FromID: "ent-dech-003", ToID: "ent-dech-001", Type: "client", Label: "client de", Context: "déchets", Verified: true},
				},
			},
		},
		Evidence: []models.Evidence{
			{
				ID:          "ev-dech-001",
				CaseID:      "case-dechets-017",
				Name:        "Conteneurs saisis",
				Type:        models.EvidencePhysical,
				Description: "15 conteneurs interceptés au port de Dakar. Déchets toxiques.",
				Reliability: 10,
			},
		},
		Timeline: []models.Event{
			{ID: "evt-dech-01", CaseID: "case-dechets-017", Title: "Alerte Interpol Dakar", Timestamp: parseDateTime("2025-07-20T00:00:00"), Importance: "high", Verified: true},
		},
		Hypotheses: []models.Hypothesis{
			{
				ID:              "hyp-dech-01",
				CaseID:          "case-dechets-017",
				Title:           "Réseau organisé multi-entreprises",
				Description:     "ChemNord n'est pas le seul client. D'autres industriels impliqués.",
				Status:          models.HypothesisPending,
				ConfidenceLevel: 70,
				GeneratedBy:     "user",
			},
		},
	}
}

// Affaire 18: Accident Mortel en Usine
func createAffaireAccidentMortelUsine() *models.Case {
	return &models.Case{
		ID:          "case-accident-018",
		Name:        "Accident Mortel - Usine Métallurgique",
		Description: "Décès d'un ouvrier dans un accident 'de travail'. Suspicion de négligence criminelle et falsification sécurité.",
		Type:        "homicide",
		Status:      "en_cours",
		CreatedAt:   parseDate("2025-10-20"),
		UpdatedAt:   parseDate("2025-11-05"),
		Entities: []models.Entity{
			{
				ID:          "ent-acc-001",
				CaseID:      "case-accident-018",
				Name:        "Karim Bouazza",
				Type:        models.EntityPerson,
				Role:        models.RoleVictim,
				Description: "Ouvrier intérimaire, 28 ans. Écrasé par presse hydraulique.",
				Attributes: map[string]string{
					"age":         "28 ans",
					"statut":      "Intérimaire",
					"anciennete":  "3 semaines",
					"formation":   "Minimale",
				},
				Relations: []models.Relation{
					{ID: "rel-acc-010", FromID: "ent-acc-001", ToID: "ent-acc-002", Type: "emploi", Label: "employé par", Context: "travail", Verified: true},
				},
			},
			{
				ID:          "ent-acc-002",
				CaseID:      "case-accident-018",
				Name:        "AcierNord SA",
				Type:        models.EntityOrg,
				Role:        models.RoleSuspect,
				Description: "Usine métallurgique. Nombreuses violations sécurité non sanctionnées.",
				Attributes: map[string]string{
					"adresse":   "27 Avenue de la Métallurgie, 59760 Grande-Synthe",
					"latitude":  "51.0089",
					"longitude": "2.3012",
					"type":      "Usine métallurgique",
				},
				Relations: []models.Relation{
					{ID: "rel-acc-020", FromID: "ent-acc-002", ToID: "ent-acc-001", Type: "negligence", Label: "responsable de la mort de", Context: "accident", Verified: false},
					{ID: "rel-acc-021", FromID: "ent-acc-002", ToID: "ent-acc-003", Type: "infraction", Label: "a ignoré les avertissements de", Context: "sécurité", Verified: true},
				},
			},
			{
				ID:          "ent-acc-003",
				CaseID:      "case-accident-018",
				Name:        "Inspection du Travail",
				Type:        models.EntityOrg,
				Role:        models.RoleWitness,
				Description: "3 rapports d'infraction en 2 ans. Mise en demeure ignorée.",
				Relations: []models.Relation{
					{ID: "rel-acc-030", FromID: "ent-acc-003", ToID: "ent-acc-002", Type: "controle", Label: "a contrôlé", Context: "sécurité", Verified: true},
				},
			},
		},
		Evidence: []models.Evidence{
			{
				ID:          "ev-acc-001",
				CaseID:      "case-accident-018",
				Name:        "Rapports inspection",
				Type:        models.EvidenceDocumentary,
				Description: "3 mises en demeure pour sécurité. Dispositif de la presse défaillant signalé.",
				Reliability: 10,
			},
			{
				ID:          "ev-acc-002",
				CaseID:      "case-accident-018",
				Name:        "Registre formation",
				Type:        models.EvidenceDocumentary,
				Description: "Formation sécurité de Bouazza: 30 minutes au lieu de 8h réglementaires.",
				Reliability: 9,
			},
		},
		Timeline: []models.Event{
			{ID: "evt-acc-01", CaseID: "case-accident-018", Title: "Dernière inspection", Timestamp: parseDateTime("2025-09-10T00:00:00"), Importance: "high", Verified: true},
			{ID: "evt-acc-02", CaseID: "case-accident-018", Title: "Accident mortel", Timestamp: parseDateTime("2025-10-20T14:35:00"), Location: "AcierNord - Atelier 3", Importance: "high", Verified: true},
		},
		Hypotheses: []models.Hypothesis{
			{
				ID:              "hyp-acc-01",
				CaseID:          "case-accident-018",
				Title:           "Homicide involontaire par négligence",
				Description:     "La direction savait que les équipements étaient dangereux.",
				Status:          models.HypothesisPending,
				ConfidenceLevel: 85,
				GeneratedBy:     "user",
			},
		},
	}
}

// ============================================================================
// SÉRIE NICE - Côte d'Azur
// ============================================================================

// Affaire 19: Escroquerie Immobilière
func createAffaireEscroquerieImmobilier() *models.Case {
	return &models.Case{
		ID:          "case-immo-019",
		Name:        "Escroquerie Immobilière Côte d'Azur",
		Description: "Ventes frauduleuses de biens immobiliers à des étrangers. 23 victimes. Préjudice: 18 millions €.",
		Type:        "fraude",
		Status:      "en_cours",
		CreatedAt:   parseDate("2025-07-20"),
		UpdatedAt:   parseDate("2025-10-30"),
		Entities: []models.Entity{
			{
				ID:          "ent-immo-001",
				CaseID:      "case-immo-019",
				Name:        "Riviera Luxury Properties",
				Type:        models.EntityOrg,
				Role:        models.RoleSuspect,
				Description: "Agence immobilière de luxe. Vente de biens inexistants ou sans titre.",
				Attributes: map[string]string{
					"adresse":   "15 Promenade des Anglais, 06000 Nice",
					"latitude":  "43.6947",
					"longitude": "7.2653",
					"type":      "Agence immobilière",
				},
				Relations: []models.Relation{
					{ID: "rel-immo-010", FromID: "ent-immo-001", ToID: "ent-immo-002", Type: "direction", Label: "dirigé par", Context: "organisation", Verified: true},
					{ID: "rel-immo-011", FromID: "ent-immo-001", ToID: "ent-immo-003", Type: "collaboration", Label: "collabore avec", Context: "fraude", Verified: true},
				},
			},
			{
				ID:          "ent-immo-002",
				CaseID:      "case-immo-019",
				Name:        "Alexandre Fontaine",
				Type:        models.EntityPerson,
				Role:        models.RoleSuspect,
				Description: "Gérant de l'agence. 45 ans. Fausse identité. Vrai nom inconnu.",
				Attributes: map[string]string{
					"age":       "45 ans",
					"identite":  "Fausse - en cours de vérification",
					"nationalite": "Inconnue",
				},
				Relations: []models.Relation{
					{ID: "rel-immo-020", FromID: "ent-immo-002", ToID: "ent-immo-001", Type: "direction", Label: "dirige", Context: "organisation", Verified: true},
					{ID: "rel-immo-021", FromID: "ent-immo-002", ToID: "ent-immo-003", Type: "collaboration", Label: "complice de", Context: "fraude", Verified: true},
				},
			},
			{
				ID:          "ent-immo-003",
				CaseID:      "case-immo-019",
				Name:        "Maître Bernard",
				Type:        models.EntityPerson,
				Role:        models.RoleSuspect,
				Description: "Notaire complice. Validait les faux actes de propriété.",
				Relations: []models.Relation{
					{ID: "rel-immo-001", FromID: "ent-immo-003", ToID: "ent-immo-001", Type: "complice", Label: "complice de", Context: "fraude", Verified: true},
				},
			},
		},
		Evidence: []models.Evidence{
			{
				ID:          "ev-immo-001",
				CaseID:      "case-immo-019",
				Name:        "Faux actes notariés",
				Type:        models.EvidenceDocumentary,
				Description: "23 actes falsifiés portant le sceau du notaire Bernard.",
				Reliability: 10,
			},
			{
				ID:          "ev-immo-002",
				CaseID:      "case-immo-019",
				Name:        "Comptes offshore",
				Type:        models.EvidenceDigital,
				Description: "Virements vers Monaco et Suisse. 12 millions tracés.",
				Reliability: 8,
			},
		},
		Timeline: []models.Event{
			{ID: "evt-immo-01", CaseID: "case-immo-019", Title: "Première plainte", Timestamp: parseDateTime("2025-07-15T00:00:00"), Importance: "high", Verified: true},
			{ID: "evt-immo-02", CaseID: "case-immo-019", Title: "Fuite Fontaine", Timestamp: parseDateTime("2025-07-20T00:00:00"), Importance: "high", Verified: true},
		},
		Hypotheses: []models.Hypothesis{
			{
				ID:              "hyp-immo-01",
				CaseID:          "case-immo-019",
				Title:           "Fontaine toujours en France",
				Description:     "Des témoignages le placent à Monaco récemment.",
				Status:          models.HypothesisPending,
				ConfidenceLevel: 55,
				GeneratedBy:     "ai",
			},
		},
	}
}

// Affaire 20: Disparition sur Yacht
func createAffaireDisparitionYacht() *models.Case {
	return &models.Case{
		ID:          "case-yacht-020",
		Name:        "Disparition en Mer - Yacht Athena",
		Description: "Disparition d'un milliardaire russe lors d'une croisière. Yacht retrouvé vide au large de Nice.",
		Type:        "disparition",
		Status:      "en_cours",
		CreatedAt:   parseDate("2025-09-10"),
		UpdatedAt:   parseDate("2025-10-20"),
		Entities: []models.Entity{
			{
				ID:          "ent-yacht-001",
				CaseID:      "case-yacht-020",
				Name:        "Dimitri Volkov",
				Type:        models.EntityPerson,
				Role:        models.RoleVictim,
				Description: "Oligarque russe, 62 ans. Fortune: 2 milliards €. Disparu le 08/09.",
				Attributes: map[string]string{
					"age":      "62 ans",
					"fortune":  "2 milliards €",
					"secteur":  "Mines et métaux",
					"ennemis":  "Nombreux en Russie",
				},
				Relations: []models.Relation{
					{ID: "rel-yacht-010", FromID: "ent-yacht-001", ToID: "ent-yacht-002", Type: "propriete", Label: "propriétaire de", Context: "possession", Verified: true},
					{ID: "rel-yacht-011", FromID: "ent-yacht-001", ToID: "ent-yacht-003", Type: "conflit", Label: "en conflit avec", Context: "politique", Verified: true},
				},
			},
			{
				ID:          "ent-yacht-002",
				CaseID:      "case-yacht-020",
				Name:        "Yacht Athena",
				Type:        models.EntityObject,
				Role:        models.RoleOther,
				Description: "Yacht de 65m. Retrouvé dérivant sans équipage. Traces de lutte.",
				Attributes: map[string]string{
					"longueur":        "65 mètres",
					"valeur":          "45 millions €",
					"equipage":        "12 personnes - tous disparus",
					"lieu_decouverte": "15 miles au large de Nice",
					"latitude":        "43.8200",
					"longitude":       "7.3500",
				},
				Relations: []models.Relation{
					{ID: "rel-yacht-020", FromID: "ent-yacht-002", ToID: "ent-yacht-001", Type: "propriete", Label: "appartient à", Context: "possession", Verified: true},
				},
			},
			{
				ID:          "ent-yacht-003",
				CaseID:      "case-yacht-020",
				Name:        "Services de Renseignement Russes",
				Type:        models.EntityOrg,
				Role:        models.RoleSuspect,
				Description: "Volkov était en conflit avec le Kremlin. Menaces récentes.",
				Relations: []models.Relation{
					{ID: "rel-yacht-030", FromID: "ent-yacht-003", ToID: "ent-yacht-001", Type: "menace", Label: "a menacé", Context: "politique", Verified: false},
				},
			},
		},
		Evidence: []models.Evidence{
			{
				ID:          "ev-yacht-001",
				CaseID:      "case-yacht-020",
				Name:        "Traces de sang",
				Type:        models.EvidenceForensic,
				Description: "Sang de Volkov retrouvé sur pont. Traces de lutte dans cabine.",
				Reliability: 9,
			},
			{
				ID:          "ev-yacht-002",
				CaseID:      "case-yacht-020",
				Name:        "Vidéosurveillance bord",
				Type:        models.EvidenceDigital,
				Description: "Effacée. Récupération partielle: approche d'un zodiac à 3h du matin.",
				Reliability: 7,
			},
		},
		Timeline: []models.Event{
			{ID: "evt-yacht-01", CaseID: "case-yacht-020", Title: "Départ Antibes", Timestamp: parseDateTime("2025-09-07T18:00:00"), Location: "Port Vauban", Importance: "medium", Verified: true},
			{ID: "evt-yacht-02", CaseID: "case-yacht-020", Title: "Dernier contact radio", Timestamp: parseDateTime("2025-09-08T02:30:00"), Importance: "high", Verified: true},
			{ID: "evt-yacht-03", CaseID: "case-yacht-020", Title: "Yacht retrouvé", Timestamp: parseDateTime("2025-09-10T06:00:00"), Location: "15 miles au large Nice", Importance: "high", Verified: true},
		},
		Hypotheses: []models.Hypothesis{
			{
				ID:              "hyp-yacht-01",
				CaseID:          "case-yacht-020",
				Title:           "Assassinat politique",
				Description:     "Opération des services russes pour éliminer un opposant.",
				Status:          models.HypothesisPending,
				ConfidenceLevel: 65,
				GeneratedBy:     "user",
			},
			{
				ID:              "hyp-yacht-02",
				CaseID:          "case-yacht-020",
				Title:           "Mise en scène fuite",
				Description:     "Volkov aurait organisé sa disparition pour échapper à des poursuites.",
				Status:          models.HypothesisPending,
				ConfidenceLevel: 30,
				GeneratedBy:     "ai",
			},
		},
	}
}

// Affaire 21: Cambriolages de Villas
func createAffaireCambriolagesVillas() *models.Case {
	return &models.Case{
		ID:          "case-villas-021",
		Name:        "Série Cambriolages - Villas Côte d'Azur",
		Description: "18 cambriolages ciblant des propriétés de luxe. Joaillerie et œuvres d'art. Butin: 25 millions €.",
		Type:        "vol",
		Status:      "en_cours",
		CreatedAt:   parseDate("2025-05-01"),
		UpdatedAt:   parseDate("2025-11-01"),
		Entities: []models.Entity{
			{
				ID:          "ent-villa-001",
				CaseID:      "case-villas-021",
				Name:        "Gang de la Riviera",
				Type:        models.EntityOrg,
				Role:        models.RoleSuspect,
				Description: "Équipe internationale. Mode opératoire sophistiqué.",
				Relations: []models.Relation{
					{ID: "rel-villa-010", FromID: "ent-villa-001", ToID: "ent-villa-002", Type: "complicite", Label: "complice de", Context: "crime", Verified: false},
					{ID: "rel-villa-011", FromID: "ent-villa-001", ToID: "ent-villa-003", Type: "recel", Label: "revend à", Context: "crime", Verified: false},
				},
			},
			{
				ID:          "ent-villa-002",
				CaseID:      "case-villas-021",
				Name:        "Société de Sécurité SecurAzur",
				Type:        models.EntityOrg,
				Role:        models.RoleSuspect,
				Description: "Toutes les villas cambriolées étaient clientes. Complicité interne ?",
				Attributes: map[string]string{
					"adresse":   "42 Boulevard Gambetta, 06000 Nice",
					"latitude":  "43.7034",
					"longitude": "7.2663",
					"type":      "Société de sécurité",
				},
				Relations: []models.Relation{
					{ID: "rel-villa-020", FromID: "ent-villa-002", ToID: "ent-villa-001", Type: "complicite", Label: "a fourni informations à", Context: "crime", Verified: false},
				},
			},
			{
				ID:          "ent-villa-003",
				CaseID:      "case-villas-021",
				Name:        "Viktor Sokolov",
				Type:        models.EntityPerson,
				Role:        models.RoleSuspect,
				Description: "Commanditaire présumé. Achète les œuvres d'art volées.",
				Attributes: map[string]string{
					"nationalite": "Russe",
					"alias":       "Le Collectionneur",
				},
				Relations: []models.Relation{
					{ID: "rel-villa-030", FromID: "ent-villa-003", ToID: "ent-villa-001", Type: "commanditaire", Label: "commandite", Context: "crime", Verified: false},
				},
			},
		},
		Evidence: []models.Evidence{
			{
				ID:          "ev-villa-001",
				CaseID:      "case-villas-021",
				Name:        "Désactivation alarmes",
				Type:        models.EvidenceDigital,
				Description: "Toutes les alarmes désactivées depuis système central SecurAzur.",
				Reliability: 9,
			},
		},
		Timeline: []models.Event{
			{ID: "evt-villa-01", CaseID: "case-villas-021", Title: "Premier cambriolage", Timestamp: parseDateTime("2025-05-01T03:00:00"), Location: "Cap-Ferrat", Importance: "high", Verified: true},
		},
		Hypotheses: []models.Hypothesis{
			{
				ID:              "hyp-villa-01",
				CaseID:          "case-villas-021",
				Title:           "Taupe chez SecurAzur",
				Description:     "Un employé fournit les codes et plannings aux cambrioleurs.",
				Status:          models.HypothesisPending,
				ConfidenceLevel: 80,
				GeneratedBy:     "user",
			},
		},
	}
}

// ============================================================================
// SÉRIE STRASBOURG - Frontières
// ============================================================================

// Affaire 22: Trafic d'Armes
func createAffaireTraficArmes() *models.Case {
	return &models.Case{
		ID:          "case-armes-022",
		Name:        "Trafic d'Armes - Filière Balkanique",
		Description: "Réseau d'importation d'armes de guerre depuis l'ex-Yougoslavie. Kalachnikovs et explosifs.",
		Type:        "trafic",
		Status:      "en_cours",
		CreatedAt:   parseDate("2025-06-15"),
		UpdatedAt:   parseDate("2025-11-05"),
		Entities: []models.Entity{
			{
				ID:          "ent-armes-001",
				CaseID:      "case-armes-022",
				Name:        "Réseau Balkan Connection",
				Type:        models.EntityOrg,
				Role:        models.RoleSuspect,
				Description: "Filière d'importation depuis Serbie et Bosnie.",
				Relations: []models.Relation{
					{ID: "rel-armes-010", FromID: "ent-armes-001", ToID: "ent-armes-002", Type: "direction", Label: "dirigé par", Context: "organisation", Verified: true},
					{ID: "rel-armes-011", FromID: "ent-armes-001", ToID: "ent-armes-003", Type: "logistique", Label: "utilise pour transport", Context: "trafic", Verified: true},
				},
			},
			{
				ID:          "ent-armes-002",
				CaseID:      "case-armes-022",
				Name:        "Dragan Petrovic",
				Type:        models.EntityPerson,
				Role:        models.RoleSuspect,
				Description: "Chef présumé branche française. Ancien militaire serbe.",
				Attributes: map[string]string{
					"age":         "48 ans",
					"nationalite": "Serbe",
					"passé":       "Ex-militaire - Guerre Bosnie",
				},
				Relations: []models.Relation{
					{ID: "rel-armes-020", FromID: "ent-armes-002", ToID: "ent-armes-001", Type: "direction", Label: "dirige", Context: "organisation", Verified: true},
					{ID: "rel-armes-021", FromID: "ent-armes-002", ToID: "ent-armes-003", Type: "controle", Label: "contrôle", Context: "trafic", Verified: true},
				},
			},
			{
				ID:          "ent-armes-003",
				CaseID:      "case-armes-022",
				Name:        "Transport Européen SARL",
				Type:        models.EntityOrg,
				Role:        models.RoleSuspect,
				Description: "Société de transport. Camions à double-fond pour cacher les armes.",
				Attributes: map[string]string{
					"adresse":   "Zone Industrielle du Port du Rhin, 67100 Strasbourg",
					"latitude":  "48.5734",
					"longitude": "7.7912",
					"type":      "Transport routier",
				},
				Relations: []models.Relation{
					{ID: "rel-armes-030", FromID: "ent-armes-003", ToID: "ent-armes-001", Type: "logistique", Label: "transporte pour", Context: "trafic", Verified: true},
				},
			},
		},
		Evidence: []models.Evidence{
			{
				ID:          "ev-armes-001",
				CaseID:      "case-armes-022",
				Name:        "Saisie arsenal",
				Type:        models.EvidencePhysical,
				Description: "45 Kalachnikovs, 12 000 munitions, 5kg de Semtex.",
				Reliability: 10,
			},
			{
				ID:          "ev-armes-002",
				CaseID:      "case-armes-022",
				Name:        "Camion modifié",
				Type:        models.EvidencePhysical,
				Description: "Double-fond découvert. Capable de cacher 100 armes.",
				Reliability: 10,
			},
		},
		Timeline: []models.Event{
			{ID: "evt-armes-01", CaseID: "case-armes-022", Title: "Renseignement DGSI", Timestamp: parseDateTime("2025-04-01T00:00:00"), Importance: "high", Verified: true},
			{ID: "evt-armes-02", CaseID: "case-armes-022", Title: "Interception frontière", Timestamp: parseDateTime("2025-06-15T04:30:00"), Location: "Poste frontière Kehl", Importance: "high", Verified: true},
		},
		Hypotheses: []models.Hypothesis{
			{
				ID:              "hyp-armes-01",
				CaseID:          "case-armes-022",
				Title:           "Clients: gangs français",
				Description:     "Les armes alimentent les règlements de compte à Marseille.",
				Status:          models.HypothesisPending,
				ConfidenceLevel: 75,
				GeneratedBy:     "user",
			},
		},
	}
}

// Affaire 23: Passeurs de Migrants
func createAffairePasseursHumains() *models.Case {
	return &models.Case{
		ID:          "case-passeurs-023",
		Name:        "Réseau de Passeurs - Filière Iranienne",
		Description: "Réseau faisant passer des migrants depuis l'Europe de l'Est. 5 à 15 000€ par personne.",
		Type:        "trafic",
		Status:      "en_cours",
		CreatedAt:   parseDate("2025-08-20"),
		UpdatedAt:   parseDate("2025-11-01"),
		Entities: []models.Entity{
			{
				ID:          "ent-pass-001",
				CaseID:      "case-passeurs-023",
				Name:        "Filière Perse",
				Type:        models.EntityOrg,
				Role:        models.RoleSuspect,
				Description: "Réseau de passeurs. Transit via Turquie, Grèce, Balkans.",
				Attributes: map[string]string{
					"tarif":     "5 000 à 15 000€",
					"victimes":  "Estimation 2000/an",
					"parcours":  "Iran-Turquie-Grèce-Balkans-France",
				},
				Relations: []models.Relation{
					{ID: "rel-pass-010", FromID: "ent-pass-001", ToID: "ent-pass-002", Type: "coordination", Label: "coordonné par", Context: "trafic", Verified: true},
				},
			},
			{
				ID:          "ent-pass-002",
				CaseID:      "case-passeurs-023",
				Name:        "Mehdi Tehrani",
				Type:        models.EntityPerson,
				Role:        models.RoleSuspect,
				Description: "Coordinateur français du réseau. Tient un restaurant à Strasbourg.",
				Attributes: map[string]string{
					"age":         "42 ans",
					"facade":      "Restaurant Le Persépolis",
					"nationalite": "Franco-iranienne",
					"adresse":     "28 Rue des Juifs, 67000 Strasbourg",
					"latitude":    "48.5812",
					"longitude":   "7.7491",
				},
				Relations: []models.Relation{
					{ID: "rel-pass-020", FromID: "ent-pass-002", ToID: "ent-pass-001", Type: "coordination", Label: "coordonne en France", Context: "trafic", Verified: true},
				},
			},
		},
		Evidence: []models.Evidence{
			{
				ID:          "ev-pass-001",
				CaseID:      "case-passeurs-023",
				Name:        "Écoutes téléphoniques",
				Type:        models.EvidenceDigital,
				Description: "Coordination des 'livraisons'. Mentions de prix et itinéraires.",
				Reliability: 9,
			},
			{
				ID:          "ev-pass-002",
				CaseID:      "case-passeurs-023",
				Name:        "Témoignages migrants",
				Type:        models.EvidenceTestimonial,
				Description: "23 migrants identifient Tehrani comme point de contact en France.",
				Reliability: 8,
			},
		},
		Timeline: []models.Event{
			{ID: "evt-pass-01", CaseID: "case-passeurs-023", Title: "Signalement Frontex", Timestamp: parseDateTime("2025-05-10T00:00:00"), Importance: "medium", Verified: true},
			{ID: "evt-pass-02", CaseID: "case-passeurs-023", Title: "Interpellation groupe", Timestamp: parseDateTime("2025-08-20T05:00:00"), Location: "Strasbourg", Importance: "high", Verified: true},
		},
		Hypotheses: []models.Hypothesis{
			{
				ID:              "hyp-pass-01",
				CaseID:          "case-passeurs-023",
				Title:           "Complices aux frontières",
				Description:     "Des douaniers seraient corrompus pour faciliter les passages.",
				Status:          models.HypothesisPending,
				ConfidenceLevel: 55,
				GeneratedBy:     "ai",
			},
		},
	}
}

// Affaire 24: Contrebande de Cigarettes
func createAffaireContrebandeCigarettes() *models.Case {
	return &models.Case{
		ID:          "case-tabac-024",
		Name:        "Contrebande de Cigarettes - Réseau Ukrainien",
		Description: "Importation massive de cigarettes de contrefaçon. 2 millions de cartouches saisies.",
		Type:        "trafic",
		Status:      "en_cours",
		CreatedAt:   parseDate("2025-09-05"),
		UpdatedAt:   parseDate("2025-10-25"),
		Entities: []models.Entity{
			{
				ID:          "ent-tabac-001",
				CaseID:      "case-tabac-024",
				Name:        "Réseau Odessa",
				Type:        models.EntityOrg,
				Role:        models.RoleSuspect,
				Description: "Réseau ukrainien. Production en usines clandestines en Moldavie.",
				Relations: []models.Relation{
					{ID: "rel-tabac-010", FromID: "ent-tabac-001", ToID: "ent-tabac-002", Type: "direction", Label: "dirigé par", Context: "organisation", Verified: true},
					{ID: "rel-tabac-011", FromID: "ent-tabac-001", ToID: "ent-tabac-003", Type: "logistique", Label: "utilise pour transport", Context: "trafic", Verified: true},
				},
			},
			{
				ID:          "ent-tabac-002",
				CaseID:      "case-tabac-024",
				Name:        "Oleksandr Shevchenko",
				Type:        models.EntityPerson,
				Role:        models.RoleSuspect,
				Description: "Chef du réseau. Ancien colonel de police ukrainien.",
				Attributes: map[string]string{
					"age":         "55 ans",
					"nationalite": "Ukrainienne",
					"passé":       "Ex-colonel de police",
				},
				Relations: []models.Relation{
					{ID: "rel-tabac-020", FromID: "ent-tabac-002", ToID: "ent-tabac-001", Type: "direction", Label: "dirige", Context: "organisation", Verified: true},
				},
			},
			{
				ID:          "ent-tabac-003",
				CaseID:      "case-tabac-024",
				Name:        "Transport Européen SARL",
				Type:        models.EntityOrg,
				Role:        models.RoleSuspect,
				Description: "Même société que pour le trafic d'armes. Multi-trafics.",
				Relations: []models.Relation{
					{ID: "rel-tabac-001", FromID: "ent-tabac-003", ToID: "ent-tabac-001", Type: "transport", Label: "transporte pour", Context: "trafic", Verified: true},
				},
			},
		},
		Evidence: []models.Evidence{
			{
				ID:          "ev-tabac-001",
				CaseID:      "case-tabac-024",
				Name:        "Saisie cartouches",
				Type:        models.EvidencePhysical,
				Description: "2 millions de cartouches. Valeur marché: 40 millions €.",
				Reliability: 10,
			},
		},
		Timeline: []models.Event{
			{ID: "evt-tabac-01", CaseID: "case-tabac-024", Title: "Interception camion", Timestamp: parseDateTime("2025-09-05T03:00:00"), Location: "A35 - Colmar", Importance: "high", Verified: true},
		},
		Hypotheses: []models.Hypothesis{
			{
				ID:              "hyp-tabac-01",
				CaseID:          "case-tabac-024",
				Title:           "Lien avec trafic d'armes",
				Description:     "La même société de transport est utilisée pour les deux trafics.",
				Status:          models.HypothesisSupported,
				ConfidenceLevel: 90,
				GeneratedBy:     "user",
			},
		},
	}
}

// ============================================================================
// SÉRIE NANTES - Maritime
// ============================================================================

// Affaire 25: Piratage Informatique
func createAffairePiratageInformatique() *models.Case {
	return &models.Case{
		ID:          "case-cyber-025",
		Name:        "Cyberattaque - Chantiers de l'Atlantique",
		Description: "Piratage des systèmes informatiques du chantier naval. Vol de plans de sous-marins nucléaires.",
		Type:        "cyber",
		Status:      "en_cours",
		CreatedAt:   parseDate("2025-10-01"),
		UpdatedAt:   parseDate("2025-11-10"),
		Entities: []models.Entity{
			{
				ID:          "ent-cyber-001",
				CaseID:      "case-cyber-025",
				Name:        "Chantiers de l'Atlantique",
				Type:        models.EntityOrg,
				Role:        models.RoleVictim,
				Description: "Chantier naval stratégique. Construit navires militaires et civils.",
				Attributes: map[string]string{
					"adresse":   "Avenue de la Forme Écluse, 44600 Saint-Nazaire",
					"latitude":  "47.2866",
					"longitude": "-2.2006",
					"type":      "Chantier naval",
				},
				Relations: []models.Relation{
					{ID: "rel-cyber-010", FromID: "ent-cyber-001", ToID: "ent-cyber-003", Type: "detenteur", Label: "détenteur de", Context: "secret", Verified: true},
				},
			},
			{
				ID:          "ent-cyber-002",
				CaseID:      "case-cyber-025",
				Name:        "Groupe APT Lazarus",
				Type:        models.EntityOrg,
				Role:        models.RoleSuspect,
				Description: "Groupe de hackers nord-coréens. Signature technique identifiée.",
				Attributes: map[string]string{
					"origine":  "Corée du Nord",
					"cibles":   "Défense, finance, crypto",
				},
				Relations: []models.Relation{
					{ID: "rel-cyber-020", FromID: "ent-cyber-002", ToID: "ent-cyber-001", Type: "attaque", Label: "a attaqué", Context: "cyber", Verified: true},
					{ID: "rel-cyber-021", FromID: "ent-cyber-002", ToID: "ent-cyber-003", Type: "vol", Label: "a volé", Context: "espionnage", Verified: true},
				},
			},
			{
				ID:          "ent-cyber-003",
				CaseID:      "case-cyber-025",
				Name:        "Plans sous-marin Barracuda",
				Type:        models.EntityDocument,
				Role:        models.RoleOther,
				Description: "Plans du nouveau sous-marin nucléaire. Secret Défense.",
				Relations: []models.Relation{
					{ID: "rel-cyber-030", FromID: "ent-cyber-003", ToID: "ent-cyber-001", Type: "appartenance", Label: "appartient à", Context: "secret", Verified: true},
				},
			},
		},
		Evidence: []models.Evidence{
			{
				ID:          "ev-cyber-001",
				CaseID:      "case-cyber-025",
				Name:        "Logs d'intrusion",
				Type:        models.EvidenceDigital,
				Description: "Connexions depuis VPN nord-coréen. Malware signature Lazarus.",
				Reliability: 9,
			},
			{
				ID:          "ev-cyber-002",
				CaseID:      "case-cyber-025",
				Name:        "Exfiltration données",
				Type:        models.EvidenceDigital,
				Description: "2 To de données exfiltrées sur 3 semaines.",
				Reliability: 10,
			},
		},
		Timeline: []models.Event{
			{ID: "evt-cyber-01", CaseID: "case-cyber-025", Title: "Première intrusion", Timestamp: parseDateTime("2025-09-10T00:00:00"), Importance: "high", Verified: true},
			{ID: "evt-cyber-02", CaseID: "case-cyber-025", Title: "Détection attaque", Timestamp: parseDateTime("2025-10-01T14:00:00"), Importance: "high", Verified: true},
		},
		Hypotheses: []models.Hypothesis{
			{
				ID:              "hyp-cyber-01",
				CaseID:          "case-cyber-025",
				Title:           "Complice interne",
				Description:     "L'accès initial nécessitait des identifiants internes.",
				Status:          models.HypothesisPending,
				ConfidenceLevel: 60,
				GeneratedBy:     "ai",
			},
		},
	}
}

// Affaire 26: Naufrage Suspect
func createAffaireNaufrageSuspect() *models.Case {
	return &models.Case{
		ID:          "case-naufrage-026",
		Name:        "Naufrage Suspect - Cargo Atlantis",
		Description: "Naufrage d'un cargo au large de Saint-Nazaire. Suspicion de sabotage pour fraude à l'assurance.",
		Type:        "fraude",
		Status:      "en_cours",
		CreatedAt:   parseDate("2025-08-15"),
		UpdatedAt:   parseDate("2025-10-30"),
		Entities: []models.Entity{
			{
				ID:          "ent-nauf-001",
				CaseID:      "case-naufrage-026",
				Name:        "Cargo Atlantis",
				Type:        models.EntityObject,
				Role:        models.RoleOther,
				Description: "Cargo de 120m. Coulé le 12/08. Cargaison déclarée: 8 millions €.",
				Attributes: map[string]string{
					"longueur":   "120 mètres",
					"cargaison":  "Déclarée: équipements industriels",
					"assurance":  "15 millions €",
				},
				Relations: []models.Relation{
					{ID: "rel-nauf-010", FromID: "ent-nauf-001", ToID: "ent-nauf-002", Type: "propriete", Label: "appartient à", Context: "possession", Verified: true},
					{ID: "rel-nauf-011", FromID: "ent-nauf-001", ToID: "ent-nauf-003", Type: "commandement", Label: "commandé par", Context: "navigation", Verified: true},
				},
			},
			{
				ID:          "ent-nauf-002",
				CaseID:      "case-naufrage-026",
				Name:        "Maritime Global Shipping",
				Type:        models.EntityOrg,
				Role:        models.RoleSuspect,
				Description: "Propriétaire du cargo. Difficultés financières connues.",
				Relations: []models.Relation{
					{ID: "rel-nauf-020", FromID: "ent-nauf-002", ToID: "ent-nauf-001", Type: "propriete", Label: "propriétaire de", Context: "possession", Verified: true},
					{ID: "rel-nauf-021", FromID: "ent-nauf-002", ToID: "ent-nauf-003", Type: "emploi", Label: "employeur de", Context: "travail", Verified: true},
				},
			},
			{
				ID:          "ent-nauf-003",
				CaseID:      "case-naufrage-026",
				Name:        "Capitaine Andréas Papadopoulos",
				Type:        models.EntityPerson,
				Role:        models.RoleSuspect,
				Description: "Capitaine du navire. A quitté le navire en premier. Incohérences dans témoignage.",
				Attributes: map[string]string{
					"age":         "52 ans",
					"nationalite": "Grecque",
					"suspicion":   "A touché bonus de 100 000€ récemment",
				},
				Relations: []models.Relation{
					{ID: "rel-nauf-030", FromID: "ent-nauf-003", ToID: "ent-nauf-001", Type: "commandement", Label: "commandait", Context: "navigation", Verified: true},
					{ID: "rel-nauf-031", FromID: "ent-nauf-003", ToID: "ent-nauf-002", Type: "complicite", Label: "complice présumé de", Context: "fraude", Verified: false},
				},
			},
		},
		Evidence: []models.Evidence{
			{
				ID:          "ev-nauf-001",
				CaseID:      "case-naufrage-026",
				Name:        "Boîte noire récupérée",
				Type:        models.EvidenceDigital,
				Description: "Ouverture vannes de coque 30min avant appel de détresse.",
				Reliability: 10,
			},
			{
				ID:          "ev-nauf-002",
				CaseID:      "case-naufrage-026",
				Name:        "Inventaire réel cargaison",
				Type:        models.EvidenceDocumentary,
				Description: "Conteneurs vides ou remplis de ferraille sans valeur.",
				Reliability: 9,
			},
		},
		Timeline: []models.Event{
			{ID: "evt-nauf-01", CaseID: "case-naufrage-026", Title: "Appel de détresse", Timestamp: parseDateTime("2025-08-12T03:45:00"), Importance: "high", Verified: true},
			{ID: "evt-nauf-02", CaseID: "case-naufrage-026", Title: "Naufrage", Timestamp: parseDateTime("2025-08-12T04:30:00"), Location: "40 miles SW Saint-Nazaire", Importance: "high", Verified: true},
		},
		Hypotheses: []models.Hypothesis{
			{
				ID:              "hyp-nauf-01",
				CaseID:          "case-naufrage-026",
				Title:           "Sabotage pour assurance",
				Description:     "Le cargo a été sabordé volontairement pour toucher l'assurance.",
				Status:          models.HypothesisPending,
				ConfidenceLevel: 85,
				GeneratedBy:     "user",
			},
		},
	}
}

// Affaire 27: Trafic d'Animaux Protégés
func createAffaireTraficAnimaux() *models.Case {
	return &models.Case{
		ID:          "case-animaux-027",
		Name:        "Trafic d'Espèces Protégées",
		Description: "Réseau d'importation illégale d'animaux exotiques. Reptiles, oiseaux, primates. Transit par le port de Nantes.",
		Type:        "trafic",
		Status:      "en_cours",
		CreatedAt:   parseDate("2025-07-10"),
		UpdatedAt:   parseDate("2025-10-20"),
		Entities: []models.Entity{
			{
				ID:          "ent-anim-001",
				CaseID:      "case-animaux-027",
				Name:        "Réseau Jungle Import",
				Type:        models.EntityOrg,
				Role:        models.RoleSuspect,
				Description: "Réseau international. Capture en Amérique du Sud et Afrique.",
				Relations: []models.Relation{
					{ID: "rel-anim-010", FromID: "ent-anim-001", ToID: "ent-anim-002", Type: "distribution", Label: "distribue via", Context: "trafic", Verified: true},
					{ID: "rel-anim-011", FromID: "ent-anim-001", ToID: "ent-anim-003", Type: "collaboration", Label: "collabore avec", Context: "trafic", Verified: true},
				},
			},
			{
				ID:          "ent-anim-002",
				CaseID:      "case-animaux-027",
				Name:        "Animalerie Exotica",
				Type:        models.EntityPlace,
				Role:        models.RoleSuspect,
				Description: "Point de vente clandestin. Façade légale: reptiles d'élevage.",
				Attributes: map[string]string{
					"adresse":   "18 Rue de la Fosse, 44000 Nantes",
					"latitude":  "47.2133",
					"longitude": "-1.5534",
					"type":      "Animalerie",
				},
				Relations: []models.Relation{
					{ID: "rel-anim-020", FromID: "ent-anim-002", ToID: "ent-anim-003", Type: "gerance", Label: "géré par", Context: "affaires", Verified: true},
				},
			},
			{
				ID:          "ent-anim-003",
				CaseID:      "case-animaux-027",
				Name:        "Jean-Marc Leblanc",
				Type:        models.EntityPerson,
				Role:        models.RoleSuspect,
				Description: "Gérant Exotica. Ancien vétérinaire radié.",
				Attributes: map[string]string{
					"age":        "55 ans",
					"profession": "Ex-vétérinaire - radié 2018",
				},
				Relations: []models.Relation{
					{ID: "rel-anim-030", FromID: "ent-anim-003", ToID: "ent-anim-002", Type: "gerance", Label: "gère", Context: "affaires", Verified: true},
					{ID: "rel-anim-031", FromID: "ent-anim-003", ToID: "ent-anim-001", Type: "collaboration", Label: "collabore avec", Context: "trafic", Verified: true},
				},
			},
		},
		Evidence: []models.Evidence{
			{
				ID:          "ev-anim-001",
				CaseID:      "case-animaux-027",
				Name:        "Animaux saisis",
				Type:        models.EvidencePhysical,
				Description: "127 reptiles, 45 oiseaux, 3 primates. Valeur: 800 000€.",
				Reliability: 10,
			},
		},
		Timeline: []models.Event{
			{ID: "evt-anim-01", CaseID: "case-animaux-027", Title: "Signalement douanes", Timestamp: parseDateTime("2025-06-15T00:00:00"), Importance: "medium", Verified: true},
			{ID: "evt-anim-02", CaseID: "case-animaux-027", Title: "Perquisition Exotica", Timestamp: parseDateTime("2025-07-10T06:00:00"), Location: "Nantes", Importance: "high", Verified: true},
		},
		Hypotheses: []models.Hypothesis{
			{
				ID:              "hyp-anim-01",
				CaseID:          "case-animaux-027",
				Title:           "Réseau européen",
				Description:     "Les animaux sont redistribués vers Allemagne et Benelux.",
				Status:          models.HypothesisPending,
				ConfidenceLevel: 65,
				GeneratedBy:     "ai",
			},
		},
	}
}

// ============================================================================
// SÉRIE RENNES - Agriculture
// ============================================================================

// Affaire 28: Abattage Clandestin
func createAffaireAbattageIllegal() *models.Case {
	return &models.Case{
		ID:          "case-abattage-028",
		Name:        "Abattage Clandestin - Filière Ovine",
		Description: "Réseau d'abattage clandestin de moutons. 15 000 bêtes abattues hors circuit légal. Risques sanitaires majeurs.",
		Type:        "fraude",
		Status:      "en_cours",
		CreatedAt:   parseDate("2025-09-01"),
		UpdatedAt:   parseDate("2025-10-25"),
		Entities: []models.Entity{
			{
				ID:          "ent-abat-001",
				CaseID:      "case-abattage-028",
				Name:        "Réseau Filière Directe",
				Type:        models.EntityOrg,
				Role:        models.RoleSuspect,
				Description: "Réseau organisant abattages à la ferme. Vente directe communautaire.",
				Relations: []models.Relation{
					{ID: "rel-abat-010", FromID: "ent-abat-001", ToID: "ent-abat-002", Type: "utilisation", Label: "utilise", Context: "abattage", Verified: true},
					{ID: "rel-abat-011", FromID: "ent-abat-001", ToID: "ent-abat-003", Type: "organisation", Label: "organisé par", Context: "abattage", Verified: true},
				},
			},
			{
				ID:          "ent-abat-002",
				CaseID:      "case-abattage-028",
				Name:        "Ferme des Quatre Vents",
				Type:        models.EntityPlace,
				Role:        models.RoleSuspect,
				Description: "Site principal d'abattage clandestin. 500m² de bâtiments adaptés.",
				Attributes: map[string]string{
					"adresse":   "Lieu-dit Les Quatre Vents, 35520 Melesse",
					"latitude":  "48.2345",
					"longitude": "-1.6978",
					"type":      "Ferme agricole",
				},
				Relations: []models.Relation{
					{ID: "rel-abat-020", FromID: "ent-abat-002", ToID: "ent-abat-001", Type: "site", Label: "site pour", Context: "abattage", Verified: true},
				},
			},
			{
				ID:          "ent-abat-003",
				CaseID:      "case-abattage-028",
				Name:        "Mohamed Benslimane",
				Type:        models.EntityPerson,
				Role:        models.RoleSuspect,
				Description: "Organisateur principal. Ancien employé d'abattoir agréé.",
				Attributes: map[string]string{
					"age":        "48 ans",
					"profession": "Ex-employé abattoir",
				},
				Relations: []models.Relation{
					{ID: "rel-abat-030", FromID: "ent-abat-003", ToID: "ent-abat-001", Type: "organisation", Label: "organise", Context: "abattage", Verified: true},
					{ID: "rel-abat-031", FromID: "ent-abat-003", ToID: "ent-abat-002", Type: "utilisation", Label: "utilise", Context: "abattage", Verified: true},
				},
			},
		},
		Evidence: []models.Evidence{
			{
				ID:          "ev-abat-001",
				CaseID:      "case-abattage-028",
				Name:        "Installation découverte",
				Type:        models.EvidencePhysical,
				Description: "Chaîne d'abattage complète. Chambres froides. Camion réfrigéré.",
				Reliability: 10,
			},
			{
				ID:          "ev-abat-002",
				CaseID:      "case-abattage-028",
				Name:        "Registres comptables",
				Type:        models.EvidenceDocumentary,
				Description: "15 000 bêtes abattues en 18 mois. CA non déclaré: 2.5 millions €.",
				Reliability: 9,
			},
		},
		Timeline: []models.Event{
			{ID: "evt-abat-01", CaseID: "case-abattage-028", Title: "Signalement vétérinaire", Timestamp: parseDateTime("2025-08-10T00:00:00"), Importance: "medium", Verified: true},
			{ID: "evt-abat-02", CaseID: "case-abattage-028", Title: "Descente de police", Timestamp: parseDateTime("2025-09-01T05:00:00"), Location: "Ferme des Quatre Vents", Importance: "high", Verified: true},
		},
		Hypotheses: []models.Hypothesis{
			{
				ID:              "hyp-abat-01",
				CaseID:          "case-abattage-028",
				Title:           "Risques sanitaires avérés",
				Description:     "Absence de contrôle vétérinaire. Animaux potentiellement malades abattus.",
				Status:          models.HypothesisPending,
				ConfidenceLevel: 70,
				GeneratedBy:     "user",
			},
		},
	}
}

// Affaire 29: Contamination Alimentaire
func createAffaireContaminationAlimentaire() *models.Case {
	return &models.Case{
		ID:          "case-contam-029",
		Name:        "Contamination Alimentaire - Usine Agroalimentaire",
		Description: "Épidémie de listériose liée à une usine de charcuterie. 47 cas, 3 décès. Négligences graves suspectées.",
		Type:        "homicide",
		Status:      "en_cours",
		CreatedAt:   parseDate("2025-10-10"),
		UpdatedAt:   parseDate("2025-11-05"),
		Entities: []models.Entity{
			{
				ID:          "ent-cont-001",
				CaseID:      "case-contam-029",
				Name:        "Charcuteries de Bretagne SA",
				Type:        models.EntityOrg,
				Role:        models.RoleSuspect,
				Description: "Usine de production. 450 employés. Fournit grande distribution.",
				Attributes: map[string]string{
					"adresse":   "Zone Industrielle de Kergaradec, 29850 Gouesnou",
					"latitude":  "48.4512",
					"longitude": "-4.4623",
					"type":      "Usine agroalimentaire",
				},
				Relations: []models.Relation{
					{ID: "rel-cont-010", FromID: "ent-cont-001", ToID: "ent-cont-002", Type: "responsabilite", Label: "responsabilité de", Context: "hygiène", Verified: true},
					{ID: "rel-cont-011", FromID: "ent-cont-001", ToID: "ent-cont-003", Type: "prejudice", Label: "a causé préjudice à", Context: "contamination", Verified: true},
				},
			},
			{
				ID:          "ent-cont-002",
				CaseID:      "case-contam-029",
				Name:        "Direction Qualité",
				Type:        models.EntityOrg,
				Role:        models.RoleSuspect,
				Description: "Aurait ignoré les alertes internes sur l'hygiène.",
				Relations: []models.Relation{
					{ID: "rel-cont-020", FromID: "ent-cont-002", ToID: "ent-cont-001", Type: "negligence", Label: "a négligé la sécurité de", Context: "hygiène", Verified: false},
				},
			},
			{
				ID:          "ent-cont-003",
				CaseID:      "case-contam-029",
				Name:        "Victimes épidémie",
				Type:        models.EntityOrg,
				Role:        models.RoleVictim,
				Description: "47 personnes contaminées. 3 décès dont 2 personnes âgées et 1 nourrisson.",
				Relations: []models.Relation{
					{ID: "rel-cont-030", FromID: "ent-cont-003", ToID: "ent-cont-001", Type: "victime", Label: "victime de", Context: "contamination", Verified: true},
				},
			},
		},
		Evidence: []models.Evidence{
			{
				ID:          "ev-cont-001",
				CaseID:      "case-contam-029",
				Name:        "Rapport DGAL",
				Type:        models.EvidenceDocumentary,
				Description: "Inspection révèle: nettoyage insuffisant, température non conforme.",
				Reliability: 10,
			},
			{
				ID:          "ev-cont-002",
				CaseID:      "case-contam-029",
				Name:        "Emails internes",
				Type:        models.EvidenceDigital,
				Description: "Alertes qualité de juin ignorées. 'On ne peut pas arrêter la production'.",
				Reliability: 9,
			},
		},
		Timeline: []models.Event{
			{ID: "evt-cont-01", CaseID: "case-contam-029", Title: "Alertes internes ignorées", Timestamp: parseDateTime("2025-06-15T00:00:00"), Importance: "high", Verified: true},
			{ID: "evt-cont-02", CaseID: "case-contam-029", Title: "Premiers cas déclarés", Timestamp: parseDateTime("2025-10-05T00:00:00"), Importance: "high", Verified: true},
			{ID: "evt-cont-03", CaseID: "case-contam-029", Title: "Rappel produits", Timestamp: parseDateTime("2025-10-10T00:00:00"), Importance: "high", Verified: true},
		},
		Hypotheses: []models.Hypothesis{
			{
				ID:              "hyp-cont-01",
				CaseID:          "case-contam-029",
				Title:           "Homicide involontaire par négligence",
				Description:     "La direction savait et n'a rien fait pour économiser les coûts de nettoyage.",
				Status:          models.HypothesisPending,
				ConfidenceLevel: 80,
				GeneratedBy:     "user",
			},
		},
	}
}

// Affaire 30: Subventions Frauduleuses
func createAffaireSubventionsFrauduleuses() *models.Case {
	return &models.Case{
		ID:          "case-subv-030",
		Name:        "Fraude aux Subventions PAC",
		Description: "Réseau de fraude aux subventions agricoles européennes. Fausses déclarations de surfaces. 8 millions € détournés.",
		Type:        "fraude",
		Status:      "en_cours",
		CreatedAt:   parseDate("2025-07-20"),
		UpdatedAt:   parseDate("2025-10-30"),
		Entities: []models.Entity{
			{
				ID:          "ent-subv-001",
				CaseID:      "case-subv-030",
				Name:        "Réseau Prairies Fantômes",
				Type:        models.EntityOrg,
				Role:        models.RoleSuspect,
				Description: "35 exploitations impliquées. Surfaces déclarées: +200% réelles.",
				Relations: []models.Relation{
					{ID: "rel-subv-010", FromID: "ent-subv-001", ToID: "ent-subv-002", Type: "collaboration", Label: "assisté par", Context: "fraude", Verified: true},
				},
			},
			{
				ID:          "ent-subv-002",
				CaseID:      "case-subv-030",
				Name:        "Cabinet Agriconseil",
				Type:        models.EntityOrg,
				Role:        models.RoleSuspect,
				Description: "Cabinet de conseil agricole. Montait les dossiers frauduleux.",
				Relations: []models.Relation{
					{ID: "rel-subv-020", FromID: "ent-subv-002", ToID: "ent-subv-001", Type: "assistance", Label: "a aidé à frauder", Context: "fraude", Verified: true},
					{ID: "rel-subv-021", FromID: "ent-subv-002", ToID: "ent-subv-003", Type: "direction", Label: "dirigé par", Context: "affaires", Verified: true},
				},
			},
			{
				ID:          "ent-subv-003",
				CaseID:      "case-subv-030",
				Name:        "Jacques Leroy",
				Type:        models.EntityPerson,
				Role:        models.RoleSuspect,
				Description: "Gérant Agriconseil. Ancien fonctionnaire DDT.",
				Attributes: map[string]string{
					"age":        "58 ans",
					"profession": "Consultant agricole",
					"passé":      "Ex-fonctionnaire DDT - démissionnaire",
				},
				Relations: []models.Relation{
					{ID: "rel-subv-001", FromID: "ent-subv-003", ToID: "ent-subv-002", Type: "direction", Label: "dirige", Context: "affaires", Verified: true},
				},
			},
		},
		Evidence: []models.Evidence{
			{
				ID:          "ev-subv-001",
				CaseID:      "case-subv-030",
				Name:        "Images satellites",
				Type:        models.EvidenceDigital,
				Description: "Comparaison déclarations PAC et surfaces réelles. Écarts massifs.",
				Reliability: 10,
			},
			{
				ID:          "ev-subv-002",
				CaseID:      "case-subv-030",
				Name:        "Dossiers falsifiés",
				Type:        models.EvidenceDocumentary,
				Description: "35 dossiers PAC avec surfaces gonflées et faux documents.",
				Reliability: 10,
			},
		},
		Timeline: []models.Event{
			{ID: "evt-subv-01", CaseID: "case-subv-030", Title: "Alerte OLAF", Timestamp: parseDateTime("2025-04-10T00:00:00"), Importance: "high", Verified: true},
			{ID: "evt-subv-02", CaseID: "case-subv-030", Title: "Perquisitions", Timestamp: parseDateTime("2025-07-20T06:00:00"), Location: "35 exploitations", Importance: "high", Verified: true},
		},
		Hypotheses: []models.Hypothesis{
			{
				ID:              "hyp-subv-01",
				CaseID:          "case-subv-030",
				Title:           "Complice à la DDT",
				Description:     "Un fonctionnaire actuel faciliterait les contrôles favorables.",
				Status:          models.HypothesisPending,
				ConfidenceLevel: 55,
				GeneratedBy:     "ai",
			},
		},
	}
}

// ============================================================================
// SÉRIE TOULOUSE - Aérospatiale (connexion avec disparition Sophie Laurent)
// ============================================================================

// Affaire 31: Espionnage Industriel
func createAffaireEspionnageIndustriel() *models.Case {
	return &models.Case{
		ID:          "case-espion-031",
		Name:        "Espionnage Industriel - Airbus",
		Description: "Vol de secrets industriels sur l'A350. Suspicion d'espionnage chinois. Connexion avec corruption locale.",
		Type:        "espionnage",
		Status:      "en_cours",
		CreatedAt:   parseDate("2025-09-25"),
		UpdatedAt:   parseDate("2025-11-08"),
		Entities: []models.Entity{
			{
				ID:          "ent-esp-001",
				CaseID:      "case-espion-031",
				Name:        "Airbus Defence & Space",
				Type:        models.EntityOrg,
				Role:        models.RoleVictim,
				Description: "Division défense d'Airbus. Programmes militaires sensibles.",
				Attributes: map[string]string{
					"adresse":   "31 Rue des Cosmonautes, 31402 Toulouse",
					"latitude":  "43.5833",
					"longitude": "1.3667",
					"type":      "Industrie aérospatiale",
				},
				Relations: []models.Relation{
					{ID: "rel-esp-010", FromID: "ent-esp-001", ToID: "ent-esp-002", Type: "emploi", Label: "employait", Context: "travail", Verified: true},
				},
			},
			{
				ID:          "ent-esp-002",
				CaseID:      "case-espion-031",
				Name:        "Chen Wei",
				Type:        models.EntityPerson,
				Role:        models.RoleSuspect,
				Description: "Ingénieur chinois. 8 ans chez Airbus. Accès aux plans A350.",
				Attributes: map[string]string{
					"age":         "38 ans",
					"nationalite": "Chinoise",
					"poste":       "Ingénieur structures",
					"habilitation": "Secret Défense - retirée",
				},
				Relations: []models.Relation{
					{ID: "rel-esp-020", FromID: "ent-esp-002", ToID: "ent-esp-001", Type: "emploi", Label: "employé par", Context: "travail", Verified: true},
					{ID: "rel-esp-021", FromID: "ent-esp-002", ToID: "ent-esp-003", Type: "agent", Label: "agent de", Context: "espionnage", Verified: false},
					{ID: "rel-esp-022", FromID: "ent-esp-002", ToID: "ent-esp-003", Type: "communication", Label: "communique via email chiffré avec", Context: "espionnage", Verified: true},
					{ID: "rel-esp-023", FromID: "ent-esp-003", ToID: "ent-esp-002", Type: "argent", Label: "effectue des virements à", Context: "corruption", Verified: true},
					{ID: "rel-esp-024", FromID: "ent-esp-002", ToID: "ent-esp-004", Type: "communication", Label: "a rencontré", Context: "facilitation", Verified: true},
				},
			},
			{
				ID:          "ent-esp-003",
				CaseID:      "case-espion-031",
				Name:        "MSS - Services Chinois",
				Type:        models.EntityOrg,
				Role:        models.RoleSuspect,
				Description: "Ministère de la Sécurité d'État chinois. Commanditaire présumé.",
				Relations: []models.Relation{
					{ID: "rel-esp-030", FromID: "ent-esp-003", ToID: "ent-esp-002", Type: "commanditaire", Label: "ordonne et contrôle", Context: "espionnage", Verified: false},
					{ID: "rel-esp-030b", FromID: "ent-esp-003", ToID: "ent-esp-002", Type: "argent", Label: "paie via société-écran", Context: "financement", Verified: true},
					{ID: "rel-esp-031", FromID: "ent-esp-003", ToID: "ent-esp-001", Type: "cible", Label: "ciblait pour vol informations", Context: "espionnage", Verified: false},
				},
			},
			{
				ID:          "ent-esp-004",
				CaseID:      "case-espion-031",
				Name:        "Marc Delmas",
				Type:        models.EntityPerson,
				Role:        models.RoleSuspect,
				Description: "Adjoint au maire Toulouse. Aurait facilité visas et contacts.",
				Attributes: map[string]string{
					"fonction": "Adjoint au maire - Relations internationales",
					"lien":     "Nombreux voyages en Chine",
				},
				Relations: []models.Relation{
					{ID: "rel-esp-001", FromID: "ent-esp-004", ToID: "ent-esp-002", Type: "facilitation", Label: "a facilité l'installation de", Context: "corruption", Verified: false},
					{ID: "rel-esp-001b", FromID: "ent-esp-004", ToID: "ent-esp-003", Type: "communication", Label: "rencontre des représentants de", Context: "diplomatie", Verified: true},
					{ID: "rel-esp-001c", FromID: "ent-esp-003", ToID: "ent-esp-004", Type: "argent", Label: "offre des cadeaux et voyages à", Context: "corruption", Verified: false},
				},
			},
		},
		Evidence: []models.Evidence{
			{
				ID:          "ev-esp-001",
				CaseID:      "case-espion-031",
				Name:        "Clé USB chiffrée",
				Type:        models.EvidenceDigital,
				Description: "Retrouvée chez Chen. 15 Go de plans et spécifications A350.",
				Reliability: 10,
			},
			{
				ID:          "ev-esp-002",
				CaseID:      "case-espion-031",
				Name:        "Transferts bancaires",
				Type:        models.EvidenceDocumentary,
				Description: "150 000€ reçus de société-écran à Hong Kong.",
				Reliability: 9,
			},
		},
		Timeline: []models.Event{
			{ID: "evt-esp-01", CaseID: "case-espion-031", Title: "Alerte DGSI", Timestamp: parseDateTime("2025-09-01T00:00:00"), Importance: "high", Verified: true},
			{ID: "evt-esp-02", CaseID: "case-espion-031", Title: "Arrestation Chen", Timestamp: parseDateTime("2025-09-25T06:00:00"), Location: "Toulouse", Importance: "high", Verified: true},
		},
		Hypotheses: []models.Hypothesis{
			{
				ID:              "hyp-esp-01",
				CaseID:          "case-espion-031",
				Title:           "Réseau plus large",
				Description:     "Chen n'est pas le seul. D'autres ingénieurs pourraient être impliqués.",
				Status:          models.HypothesisPending,
				ConfidenceLevel: 60,
				GeneratedBy:     "ai",
			},
			{
				ID:              "hyp-esp-02",
				CaseID:          "case-espion-031",
				Title:           "Lien avec disparition Laurent",
				Description:     "Sophie Laurent enquêtait sur Delmas. A-t-elle découvert le lien chinois?",
				Status:          models.HypothesisPending,
				ConfidenceLevel: 45,
				GeneratedBy:     "ai",
			},
		},
	}
}

// Affaire 32: Sabotage Aéronautique
func createAffaireSabotageAeronautique() *models.Case {
	return &models.Case{
		ID:          "case-sabotage-032",
		Name:        "Sabotage Ligne d'Assemblage A321",
		Description: "Actes de sabotage sur la chaîne d'assemblage. 3 avions touchés. Suspicion de vengeance syndicale ou espionnage.",
		Type:        "sabotage",
		Status:      "en_cours",
		CreatedAt:   parseDate("2025-10-15"),
		UpdatedAt:   parseDate("2025-11-05"),
		Entities: []models.Entity{
			{
				ID:          "ent-sab-001",
				CaseID:      "case-sabotage-032",
				Name:        "Usine Airbus Blagnac",
				Type:        models.EntityPlace,
				Role:        models.RoleVictim,
				Description: "Site d'assemblage final A320/A321. 3000 employés.",
				Attributes: map[string]string{
					"adresse":   "1 Rond-Point Maurice Bellonte, 31707 Blagnac",
					"latitude":  "43.6283",
					"longitude": "1.3650",
					"type":      "Usine aéronautique",
				},
				Relations: []models.Relation{
					{ID: "rel-sab-010", FromID: "ent-sab-001", ToID: "ent-sab-003", Type: "emploi", Label: "emploie", Context: "travail", Verified: true},
				},
			},
			{
				ID:          "ent-sab-002",
				CaseID:      "case-sabotage-032",
				Name:        "Syndicat CGT Airbus",
				Type:        models.EntityOrg,
				Role:        models.RoleSuspect,
				Description: "En conflit dur avec direction. Menaces de 'faire payer' la direction.",
				Relations: []models.Relation{
					{ID: "rel-sab-020", FromID: "ent-sab-002", ToID: "ent-sab-001", Type: "conflit", Label: "en conflit avec", Context: "social", Verified: true},
					{ID: "rel-sab-021", FromID: "ent-sab-002", ToID: "ent-sab-003", Type: "membre", Label: "compte comme membre", Context: "syndical", Verified: true},
				},
			},
			{
				ID:          "ent-sab-003",
				CaseID:      "case-sabotage-032",
				Name:        "Pierre Gonzales",
				Type:        models.EntityPerson,
				Role:        models.RoleSuspect,
				Description: "Délégué syndical. Accès zone assemblage. Expertise technique.",
				Attributes: map[string]string{
					"age":        "45 ans",
					"profession": "Technicien assemblage",
					"anciennete": "22 ans",
					"role_synd":  "Délégué CGT",
				},
				Relations: []models.Relation{
					{ID: "rel-sab-030", FromID: "ent-sab-003", ToID: "ent-sab-001", Type: "emploi", Label: "employé de", Context: "travail", Verified: true},
					{ID: "rel-sab-031", FromID: "ent-sab-003", ToID: "ent-sab-002", Type: "membre", Label: "délégué de", Context: "syndical", Verified: true},
				},
			},
		},
		Evidence: []models.Evidence{
			{
				ID:          "ev-sab-001",
				CaseID:      "case-sabotage-032",
				Name:        "Câbles sectionnés",
				Type:        models.EvidencePhysical,
				Description: "Câbles de commande sectionnés de façon à ne pas être détectés au sol.",
				Reliability: 10,
			},
			{
				ID:          "ev-sab-002",
				CaseID:      "case-sabotage-032",
				Name:        "Badge Gonzales",
				Type:        models.EvidenceDigital,
				Description: "Présent sur site lors des 3 incidents. Seul point commun.",
				Reliability: 8,
			},
		},
		Timeline: []models.Event{
			{ID: "evt-sab-01", CaseID: "case-sabotage-032", Title: "Premier incident détecté", Timestamp: parseDateTime("2025-10-01T00:00:00"), Importance: "high", Verified: true},
			{ID: "evt-sab-02", CaseID: "case-sabotage-032", Title: "Troisième incident", Timestamp: parseDateTime("2025-10-15T00:00:00"), Importance: "high", Verified: true},
		},
		Hypotheses: []models.Hypothesis{
			{
				ID:              "hyp-sab-01",
				CaseID:          "case-sabotage-032",
				Title:           "Vengeance syndicale",
				Description:     "Gonzales agit seul pour faire pression sur la direction.",
				Status:          models.HypothesisPending,
				ConfidenceLevel: 50,
				GeneratedBy:     "user",
			},
			{
				ID:              "hyp-sab-02",
				CaseID:          "case-sabotage-032",
				Title:           "Lien avec espionnage",
				Description:     "Le sabotage cache une opération d'espionnage plus large.",
				Status:          models.HypothesisPending,
				ConfidenceLevel: 40,
				GeneratedBy:     "ai",
			},
		},
	}
}

// Affaire 33: Corruption Marchés Publics Toulouse
func createAffaireCorruptionMarchesPublics() *models.Case {
	return &models.Case{
		ID:          "case-corrup-033",
		Name:        "Corruption Marchés Publics - Mairie Toulouse",
		Description: "Réseau de corruption autour des marchés publics municipaux. Lien direct avec l'affaire disparition Sophie Laurent.",
		Type:        "corruption",
		Status:      "en_cours",
		CreatedAt:   parseDate("2025-09-20"),
		UpdatedAt:   parseDate("2025-11-10"),
		Entities: []models.Entity{
			{
				ID:          "ent-corr-001",
				CaseID:      "case-corrup-033",
				Name:        "Marc Delmas",
				Type:        models.EntityPerson,
				Role:        models.RoleSuspect,
				Description: "Adjoint au maire. Central dans l'attribution frauduleuse des marchés.",
				Attributes: map[string]string{
					"fonction":      "Adjoint - Marchés publics",
					"enrichissement": "Patrimoine x3 en 5 ans",
					"lien_laurent":  "Cible de son enquête",
				},
				Relations: []models.Relation{
					{ID: "rel-corr-010", FromID: "ent-corr-001", ToID: "ent-corr-002", Type: "corruption", Label: "reçoit des commissions de", Context: "corruption", Verified: true},
					{ID: "rel-corr-011", FromID: "ent-corr-001", ToID: "ent-corr-003", Type: "hierarchie", Label: "sous l'autorité de", Context: "politique", Verified: true},
					{ID: "rel-corr-012", FromID: "ent-corr-001", ToID: "ent-corr-004", Type: "menace", Label: "cible de l'enquête de", Context: "enquête", Verified: true},
				},
			},
			{
				ID:          "ent-corr-002",
				CaseID:      "case-corrup-033",
				Name:        "Roux Constructions SARL",
				Type:        models.EntityOrg,
				Role:        models.RoleSuspect,
				Description: "Principal bénéficiaire des marchés truqués. Verse des commissions.",
				Attributes: map[string]string{
					"marches":     "45 millions € sur 3 ans",
					"surfacturation": "Estimée à 30%",
				},
				Relations: []models.Relation{
					{ID: "rel-corr-001", FromID: "ent-corr-002", ToID: "ent-corr-001", Type: "corruption", Label: "verse des commissions à", Context: "corruption", Verified: true},
				},
			},
			{
				ID:          "ent-corr-003",
				CaseID:      "case-corrup-033",
				Name:        "Maire Bernard Castex",
				Type:        models.EntityPerson,
				Role:        models.RoleSuspect,
				Description: "Maire de Toulouse. Niveau d'implication en cours d'investigation.",
				Attributes: map[string]string{
					"fonction": "Maire de Toulouse",
					"mandat":   "Depuis 2020",
				},
				Relations: []models.Relation{
					{ID: "rel-corr-030", FromID: "ent-corr-003", ToID: "ent-corr-001", Type: "hierarchie", Label: "supérieur de", Context: "politique", Verified: true},
					{ID: "rel-corr-031", FromID: "ent-corr-003", ToID: "ent-corr-002", Type: "beneficiaire", Label: "bénéficiaire présumé de", Context: "corruption", Verified: false},
				},
			},
			{
				ID:          "ent-corr-004",
				CaseID:      "case-corrup-033",
				Name:        "Sophie Laurent",
				Type:        models.EntityPerson,
				Role:        models.RoleVictim,
				Description: "Journaliste disparue. Enquêtait sur cette affaire de corruption.",
				Attributes: map[string]string{
					"profession": "Journaliste d'investigation",
					"statut":     "Disparue depuis 15/09/2025",
					"articles":   "3 articles publiés avant disparition",
				},
				Relations: []models.Relation{
					{ID: "rel-corr-002", FromID: "ent-corr-004", ToID: "ent-corr-001", Type: "enquete", Label: "enquêtait sur", Context: "journalisme", Verified: true},
					{ID: "rel-corr-003", FromID: "ent-corr-004", ToID: "ent-corr-002", Type: "enquete", Label: "enquêtait sur", Context: "journalisme", Verified: true},
				},
			},
		},
		Evidence: []models.Evidence{
			{
				ID:          "ev-corr-001",
				CaseID:      "case-corrup-033",
				Name:        "Dossier Sophie Laurent",
				Type:        models.EvidenceDocumentary,
				Description: "Notes de la journaliste. Preuves de surfacturation et commissions.",
				Reliability: 9,
				LinkedEntities: []string{"ent-corr-004", "ent-corr-001", "ent-corr-002"},
			},
			{
				ID:          "ev-corr-002",
				CaseID:      "case-corrup-033",
				Name:        "Comptes offshore Delmas",
				Type:        models.EvidenceDigital,
				Description: "2.3 millions € sur comptes à Dubaï. Origine: sociétés de Roux.",
				Reliability: 10,
				LinkedEntities: []string{"ent-corr-001"},
			},
			{
				ID:          "ev-corr-003",
				CaseID:      "case-corrup-033",
				Name:        "Appels d'offres truqués",
				Type:        models.EvidenceDocumentary,
				Description: "12 appels d'offres avec spécifications sur mesure pour Roux.",
				Reliability: 9,
				LinkedEntities: []string{"ent-corr-002"},
			},
		},
		Timeline: []models.Event{
			{ID: "evt-corr-01", CaseID: "case-corrup-033", Title: "Début enquête Laurent", Timestamp: parseDateTime("2025-07-01T00:00:00"), Importance: "high", Verified: true},
			{ID: "evt-corr-02", CaseID: "case-corrup-033", Title: "Premier article Laurent", Timestamp: parseDateTime("2025-09-10T00:00:00"), Importance: "high", Verified: true},
			{ID: "evt-corr-03", CaseID: "case-corrup-033", Title: "Disparition Laurent", Timestamp: parseDateTime("2025-09-15T19:45:00"), Importance: "high", Verified: true},
			{ID: "evt-corr-04", CaseID: "case-corrup-033", Title: "Ouverture enquête judiciaire", Timestamp: parseDateTime("2025-09-20T00:00:00"), Importance: "high", Verified: true},
		},
		Hypotheses: []models.Hypothesis{
			{
				ID:              "hyp-corr-01",
				CaseID:          "case-corrup-033",
				Title:           "Disparition liée à l'enquête",
				Description:     "Sophie Laurent a été enlevée pour stopper ses révélations.",
				Status:          models.HypothesisPending,
				ConfidenceLevel: 85,
				SupportingEvidence: []string{"ev-corr-001"},
				GeneratedBy:     "user",
			},
			{
				ID:              "hyp-corr-02",
				CaseID:          "case-corrup-033",
				Title:           "Implication du maire",
				Description:     "Castex savait et couvrait les agissements de Delmas.",
				Status:          models.HypothesisPending,
				ConfidenceLevel: 55,
				GeneratedBy:     "ai",
			},
			{
				ID:              "hyp-corr-03",
				CaseID:          "case-corrup-033",
				Title:           "Réseau plus large",
				Description:     "La corruption s'étend au-delà de Toulouse. Connexions régionales.",
				Status:          models.HypothesisPending,
				ConfidenceLevel: 50,
				GeneratedBy:     "ai",
			},
		},
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
