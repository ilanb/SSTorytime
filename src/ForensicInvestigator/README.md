# Forensic Investigator - Système d'Aide à l'Enquête Criminalistique

## 🎯 Vision

**Forensic Investigator** est une application d'aide à l'investigation criminalistique qui combine plusieurs technologies avancées :

| Technologie | Description |
|-------------|-------------|
| **PEACE** | Méthodologie d'interrogatoire britannique |
| **PROGREAI** | Méthode française de recueil d'auditions |
| **Model2vec** | Embeddings sémantiques pour la recherche intelligente |
| **LLM (vLLM)** | Analyse par IA via vLLM (Qwen2.5-7B-Instruct) |
| **SSTorytime** | Narration sémantique et découverte de chemins |
| **HRM** | Hypothetical Reasoning Model - Raisonnement logique |
| **Graphe de Connaissance** | Visualisation des relations entre entités |
| **N4L** | Notes for Linking - Structuration des données |

## 🚀 Fonctionnalités Implémentées

### ✅ Gestion d'Affaires
- Création/édition d'affaires avec métadonnées
- Classification (homicide, vol, fraude, espionnage, etc.)
- Statuts (en cours, résolu, classé)
- Affaires de démonstration pré-chargées

### ✅ Graphe de Connaissance (vis.js)
- Visualisation interactive des relations
- Nœuds colorés par type (personne, lieu, organisation, document, objet)
- Nœuds colorés par rôle (suspect, victime, témoin)
- Zoom, pan, sélection de nœuds
- Menu contextuel au clic droit

### ✅ Gestion des Entités
- Types : Personne, Lieu, Organisation, Document, Objet
- Rôles : Suspect, Victime, Témoin, Enquêteur, Autre
- Ajout, modification, suppression
- Relations entre entités

### ✅ Gestion des Preuves
- Types : Physique, Testimoniale, Documentaire, Numérique, Forensique
- Score de fiabilité (1-10)
- Chaîne de possession
- Liens avec entités

### ✅ Timeline Interactive
- Chronologie des événements
- Ajout d'événements avec horodatage
- Visualisation temporelle
- Détection d'incohérences potentielles

### ✅ Hypothèses d'Investigation
- Création manuelle d'hypothèses
- Niveau de confiance (0-100%)
- Preuves à l'appui / contradictoires
- Analyse IA des hypothèses

### ✅ Recherche Hybride (BM25 + Model2vec)
- **BM25** : Algorithme de recherche lexicale (Best Matching 25)
- **Model2vec** : Recherche sémantique par embeddings
- Pondération configurable (slider 0-100%)
- Recherche sur entités, preuves et événements
- Affichage des scores détaillés

### ✅ Inférences Sémantiques
- **Fermeture transitive** : Si A→B→C, suggère A→C
- **Détection de siblings** : Entités avec parents communs
- **Liaison d'orphelins** : Connexions pour nœuds isolés
- Bouton d'explication pour chaque suggestion
- Prévisualisation sur le graphe
- Application/rejet des inférences

### ✅ Assistant IA (Ollama)
- Chat conversationnel contextuel
- Analyse d'affaire complète
- Génération d'hypothèses
- Détection de contradictions
- Questions d'investigation suggérées

### ✅ Analyse Inter-Affaires (Cross-Case)
- Scan de connexions entre affaires
- Détection d'entités communes
- Correspondances de lieux/modus operandi
- Graphe multi-affaires
- Analyse IA des patterns

### ✅ HRM - Hypothetical Reasoning Model
- Raisonnement déductif/inductif/abductif
- Vérification formelle d'hypothèses
- Détection de contradictions logiques
- Analyse inter-affaires avancée

### ✅ Import/Export N4L
- Parsing de fichiers N4L
- Export d'affaires au format N4L
- Support des modificateurs (\new, \never)
- Contextes temporels

### ✅ Conversion Texte → N4L (IA)

- Upload de fichiers .txt dans le modal de création d'affaire
- Conversion automatique via modèle fine-tuné `n4l-qwen:latest`
- Aperçu N4L généré avant création
- Import automatique des entités et timeline parsées
- Drag & drop ou sélection de fichier

### ✅ Gestion des Affaires (Améliorée)

- Bouton de suppression sur chaque affaire dans la sidebar
- Confirmation modale avant suppression
- Nettoyage automatique de l'historique récent
- Tri par consultation récente, nom, date de création/modification

### ✅ Recherche Avancée (Filtres)
- Filtrage par type d'entité
- Filtrage par rôle
- Filtrage par type de relation
- Recherche textuelle rapide
- Exclusion de nœuds spécifiques

### ✅ Menu Contextuel (Clic droit)
- Explorer le voisinage
- Cône d'expansion
- Supprimer entité/relation
- Exclure des filtres
- Analyser le chemin

### ✅ Mode Investigation (PEACE/PROGREAI)

- **6 étapes guidées** : Identification des Acteurs, Analyse des Lieux, Reconstitution Chronologique, Analyse des Mobiles, Évaluation des Preuves, Synthèse et Hypothèses
- Questions d'exploration pour chaque étape
- Suggestions automatiques basées sur le graphe
- Analyse IA contextuelle par étape
- Notes d'enquêteur
- Insights et recommandations

### ✅ Analyse de Graphe Avancée

- **Clusters** : Détection automatique de groupes d'entités connectées
  - Mini-graphe interactif par cluster
  - Nœud central mis en évidence
  - Double-clic pour voir dans le graphe principal
- **Centralité** : Classement des nœuds par importance
  - Degree centrality (nombre de connexions)
  - Betweenness centrality (intermédiarité)
  - Closeness centrality (proximité)
  - Top 10 avec médailles or/argent/bronze
- **Scores de Suspicion** : Évaluation automatique des suspects
  - Facteurs : Mobile financier, Conflit connu, Accès aux lieux, Alibi non vérifié, Preuves liées
  - Score de 0 à 100%
  - Classification : high (rouge), medium (orange), low (vert)
- **Timeline des Alibis** : Visualisation temporelle
  - Barre verticale marquant l'heure du crime
  - Blocs verts = alibi vérifié, orange = non vérifié
  - Indicateur de fenêtre d'opportunité
  - Axe horaire interactif
- **Densité** : Zones explorées vs inexplorées
- **Cohérence** : Détection des contradictions et cycles
- **Patterns Temporels** : Séquences d'événements automatiquement détectées

## 📋 Fondements Méthodologiques

### Méthode PEACE (UK/International)
- **P**lanification et préparation
- **E**ngagement et explication
- **A**ccount (Récit libre)
- **C**losure (Clôture)
- **E**valuation

### Méthode PROGREAI (Gendarmerie Française)
- Processus Général de Recueil des Entretiens, Auditions et Interrogatoires
- Accent sur l'écoute active (80-90% du temps pour le témoin)
- Mise en confiance progressive
- Questions ouvertes privilégiées

## 🏗️ Architecture

```
ForensicInvestigator/
├── main.go                          # Point d'entrée
├── go.mod                           # Dépendances Go
├── data/
│   └── demo.go                      # Données de démonstration
├── internal/
│   ├── models/
│   │   └── models.go                # Structures de données
│   ├── services/
│   │   ├── ollama.go                # Service LLM
│   │   ├── case.go                  # Service gestion affaires
│   │   ├── n4l.go                   # Service N4L
│   │   ├── hrm.go                   # Service HRM
│   │   └── search.go                # Service recherche hybride
│   └── handlers/
│       └── handlers.go              # API REST
├── embedding_service/               # Service Model2vec (Python)
│   ├── main.py                      # API FastAPI
│   └── requirements.txt             # Dépendances Python
└── static/
    ├── index.html                   # Interface principale
    ├── css/
    │   └── styles.css               # Styles
    └── js/
        ├── app.js                   # Application principale
        └── inference.js             # Moteur d'inférences
```

## 🔧 Technologies

| Composant | Technologie |
|-----------|-------------|
| Backend | Go 1.21+ |
| Frontend | HTML5, CSS3, JavaScript ES6+ |
| Graphes | vis.js |
| Markdown | marked.js |
| LLM | vLLM (Qwen2.5-7B-Instruct) |
| Embeddings | Model2vec (Python/FastAPI) |
| API | REST JSON |

## 🚀 Installation et Démarrage

### Prérequis

- Go 1.21+
- Python 3.9+ (pour Model2vec)
- Accès au serveur vLLM : `http://86.204.69.30:8001`

### Démarrage

```bash
# 1. Service Model2vec (terminal 1)
cd embedding_service
pip install -r requirements.txt
python main.py
# → Écoute sur http://localhost:8085

# 2. Application principale (terminal 2)
cd ..
go run main.go
# → Écoute sur http://localhost:8082
# → Connecté au vLLM sur http://86.204.69.30:8001
```

### Accès
Ouvrir http://localhost:8082 dans un navigateur

## 📡 API Endpoints

### Affaires
| Méthode | Endpoint | Description |
|---------|----------|-------------|
| GET | `/api/cases` | Liste des affaires |
| POST | `/api/cases` | Créer une affaire |
| GET | `/api/cases/{id}` | Détails d'une affaire |
| PUT | `/api/cases/{id}` | Modifier une affaire |
| DELETE | `/api/cases/{id}` | Supprimer une affaire |

### Entités
| Méthode | Endpoint | Description |
|---------|----------|-------------|
| GET | `/api/entities?case_id=` | Liste des entités |
| POST | `/api/entities?case_id=` | Ajouter une entité |
| PUT | `/api/entities/update?case_id=` | Modifier une entité |
| DELETE | `/api/entities/delete?case_id=&entity_id=` | Supprimer |

### Preuves
| Méthode | Endpoint | Description |
|---------|----------|-------------|
| GET | `/api/evidence?case_id=` | Liste des preuves |
| POST | `/api/evidence?case_id=` | Ajouter une preuve |
| PUT | `/api/evidence/update?case_id=` | Modifier |
| DELETE | `/api/evidence/delete?case_id=&evidence_id=` | Supprimer |

### Timeline
| Méthode | Endpoint | Description |
|---------|----------|-------------|
| GET | `/api/timeline?case_id=` | Liste des événements |
| POST | `/api/timeline?case_id=` | Ajouter un événement |
| DELETE | `/api/timeline/delete?case_id=&event_id=` | Supprimer |

### Hypothèses
| Méthode | Endpoint | Description |
|---------|----------|-------------|
| GET | `/api/hypotheses?case_id=` | Liste des hypothèses |
| POST | `/api/hypotheses?case_id=` | Ajouter |
| PUT | `/api/hypotheses/update?case_id=` | Modifier |
| DELETE | `/api/hypotheses/delete?case_id=&hypothesis_id=` | Supprimer |
| POST | `/api/hypotheses/analyze` | Analyser une hypothèse |

### Analyse IA
| Méthode | Endpoint | Description |
|---------|----------|-------------|
| POST | `/api/analyze` | Analyse complète |
| POST | `/api/analyze/contradictions` | Détecter contradictions |
| POST | `/api/analyze/questions` | Générer questions |
| POST | `/api/analyze/path` | Analyser un chemin |
| POST | `/api/chat` | Chat avec l'assistant |

### Recherche Hybride
| Méthode | Endpoint | Description |
|---------|----------|-------------|
| POST | `/api/search/hybrid` | Recherche BM25 + sémantique |
| GET | `/api/search/quick?case_id=&q=` | Recherche BM25 rapide |

### HRM (Hypothetical Reasoning Model)
| Méthode | Endpoint | Description |
|---------|----------|-------------|
| GET | `/api/hrm/status` | Statut du service |
| POST | `/api/hrm/reason` | Raisonnement |
| POST | `/api/hrm/verify-hypothesis` | Vérifier hypothèse |
| POST | `/api/hrm/contradictions` | Détecter contradictions |
| POST | `/api/hrm/cross-case` | Analyse inter-affaires |

### Inter-Affaires
| Méthode | Endpoint | Description |
|---------|----------|-------------|
| POST | `/api/cross-case/scan` | Scanner connexions |
| POST | `/api/cross-case/analyze` | Analyser patterns |
| POST | `/api/cross-case/graph` | Graphe multi-affaires |

### N4L
| Méthode | Endpoint | Description |
|---------|----------|-------------|
| POST | `/api/n4l/parse` | Parser du N4L |
| GET | `/api/n4l/export?case_id=` | Exporter en N4L |
| POST | `/api/n4l/convert` | Convertir texte → N4L via IA (n4l-qwen) |

### Graphe
| Méthode | Endpoint | Description |
|---------|----------|-------------|
| GET | `/api/graph?case_id=` | Données du graphe |

### Analyse de Graphe

| Méthode | Endpoint | Description |
|---------|----------|-------------|
| GET | `/api/graph/analyze-complete?case_id=` | Analyse complète (clusters, centralité, alibis, etc.) |
| GET | `/api/graph/clusters?case_id=` | Détection de clusters |
| GET | `/api/graph/density?case_id=` | Carte de densité |
| GET | `/api/graph/consistency?case_id=` | Vérification de cohérence |
| GET | `/api/graph/temporal-patterns?case_id=` | Patterns temporels |
| POST | `/api/graph/paths` | Trouver chemins entre nœuds |
| POST | `/api/graph/layered` | Graphe en couches |
| POST | `/api/graph/expansion-cone` | Cône d'expansion |

### Mode Investigation

| Méthode | Endpoint | Description |
|---------|----------|-------------|
| POST | `/api/investigation/start` | Démarrer une session d'investigation |
| POST | `/api/investigation/suggestions` | Obtenir suggestions pour une étape |
| POST | `/api/investigation/analyze` | Analyse IA d'une étape |

## 📚 Sources et Références

### Méthodologies d'Enquête
- [Méthode PROGREAI - Gendarmerie Française](https://consultation.avocat.fr/blog/alexandre-gillioen/article-38924-la-methode-progreai-lors-d-un-interrogatoire-en-gendarmerie.html)
- [Méthodes d'entretien CTI](https://cti2024.org/wp-content/uploads/2021/01/CTI-Training_Tool_1-FRA-FINAL.pdf)
- [UNODC - Enquêtes criminelles](https://www.unodc.org/documents/justice-and-prison-reform/cjat/Enquetes_criminelles.pdf)

### Technologies IA
- [Model2vec - Static Embeddings](https://github.com/MinishLab/model2vec)
- [BM25 Algorithm](https://en.wikipedia.org/wiki/Okapi_BM25)
- [Ollama - Local LLM](https://ollama.ai/)

### Link Analysis
- [Link Analysis Techniques](https://cambridge-intelligence.com/link-analysis-techniques/)
- [Knowledge Graphs in Forensics](https://www.hilarispublisher.com/open-access/advancing-forensic-science-ai-and-knowledge-graphs-unlock-new-insights.pdf)

### Logiciels de Référence
- [Case Closed Software](https://caseclosedsoftware.com/)
- [Kaseware](https://www.kaseware.com/government)
- [i2 Analyst's Notebook](https://www.ibm.com/products/i2-analysts-notebook)

## 📄 Licence

Projet interne - Tous droits réservés

## 🔮 Roadmap

### ✅ Phase 1 : MVP (Complété)
- [x] Structure de données affaires/entités
- [x] Interface de saisie
- [x] Export N4L
- [x] Visualisation graphe

### ✅ Phase 2 : Intelligence (Complété)
- [x] Intégration LLM pour analyse
- [x] Génération automatique d'hypothèses
- [x] Détection d'incohérences
- [x] Timeline interactive
- [x] Inférences sémantiques

### ✅ Phase 3 : Recherche Avancée (Complété)
- [x] Recherche hybride BM25 + Model2vec
- [x] Filtrage multi-critères
- [x] Menu contextuel
- [x] Analyse inter-affaires
- [x] HRM intégration

### ✅ Phase 4 : Intégration N4L-Qwen (Complété)

- [x] Fine-tuning modèle Qwen pour génération N4L
- [x] Conversion texte → N4L via IA
- [x] Upload fichier dans création d'affaire
- [x] Suppression d'affaires avec confirmation
- [x] Amélioration UX sidebar affaires

### ✅ Phase 5 : Investigation Avancée (Complété)

- [x] Mode Investigation guidé (6 étapes PEACE/PROGREAI)
- [x] Analyse de graphe avancée (clusters, centralité, cohérence)
- [x] Scores de suspicion automatiques
- [x] Timeline visuelle des alibis
- [x] Mini-graphe interactif par cluster
- [x] Métriques de centralité (degree, betweenness, closeness)
- [x] Détection automatique de patterns temporels
- [x] Analyse IA contextuelle par étape d'investigation

### 🔄 Phase 6 : Production (En cours)

- [ ] Authentification utilisateurs
- [ ] Chiffrement des données
- [ ] Multi-utilisateurs
- [ ] Export PDF/rapports
- [ ] Import données externes (téléphonie, bancaire)
- [ ] Persistance base de données (PostgreSQL/SQLite)
