# ForensicInvestigator - Document Produit

## Vue d'ensemble

**ForensicInvestigator** est une plateforme d'aide à l'investigation criminelle augmentée par l'intelligence artificielle. Elle centralise la gestion des affaires, l'analyse des preuves, la visualisation des réseaux relationnels et le raisonnement déductif automatisé pour accompagner les enquêteurs dans la résolution d'affaires complexes.

**Domaine :** Investigation criminelle / Analyse forensique
**Utilisateurs cibles :** Enquêteurs de police judiciaire, analystes criminels, magistrats instructeurs
**Langue de l'interface :** Français

---

## Proposition de valeur

| Problème | Solution ForensicInvestigator |
|----------|-------------------------------|
| Les affaires complexes impliquent des centaines d'entités et relations difficiles à appréhender | Graphe de connaissances interactif avec algorithmes d'analyse avancés |
| Les contradictions dans les témoignages passent inaperçues | Détection automatique de contradictions par IA |
| Les liens entre affaires distinctes sont rarement identifiés | Analyse cross-affaires avec matching sémantique |
| Le raisonnement déductif est long et sujet aux biais cognitifs | Raisonnement hiérarchique multi-étapes (HRM) avec scoring de confiance |
| Les anomalies temporelles et comportementales sont noyées dans la masse | Détection statistique d'anomalies (Z-score, MAD) |
| Les hypothèses sont rarement testées formellement | Simulation "What-If" avec propagation des implications |

---

## Architecture technique

### Stack technologique

| Composant | Technologie | Port |
|-----------|-------------|------|
| Backend API | Go 1.24 (stdlib net/http) | 8082 |
| Moteur de raisonnement HRM | Python / FastAPI / vLLM | 8081 |
| Service d'embeddings | multilingual-e5-base sur SPARK (768 dim) | 8002 |
| Frontend | JavaScript vanilla (modulaire) | - |
| Visualisation graphe | vis-network | - |
| Cartographie | Leaflet | - |
| LLM distant | vLLM sur SPARK GB10 (Qwen3.8-27B-FP8, 128k ctx) | 8001 |
| Analyse de graphe | SSTorytime (package Go interne) | - |

### Diagramme de services

```
┌─────────────────────────────────────────────────┐
│                   NAVIGATEUR                     │
│  ┌──────────────────────────────────────────┐   │
│  │         SPA JavaScript (22 modules)       │   │
│  │  Graphe · Timeline · Chat · HRM · N4L    │   │
│  └──────────────────┬───────────────────────┘   │
└─────────────────────┼───────────────────────────┘
                      │ HTTP/SSE
┌─────────────────────┼───────────────────────────┐
│              Backend Go (:8082)                   │
│  ┌─────────┐ ┌──────────┐ ┌──────────────────┐  │
│  │  Cases   │ │  N4L     │ │ Graph Analyzer   │  │
│  │ Service  │ │ Service  │ │ (SSTorytime)     │  │
│  ├─────────┤ ├──────────┤ ├──────────────────┤  │
│  │ Search  │ │ Anomaly  │ │ Scenario         │  │
│  │ Service │ │ Service  │ │ Service          │  │
│  ├─────────┤ ├──────────┤ ├──────────────────┤  │
│  │ Notebook│ │ Ollama   │ │ HRM Client       │  │
│  │ Service │ │ Service  │ │ Service          │  │
│  └─────────┘ └──────────┘ └──────────────────┘  │
└───────┬─────────────────────────────┬───────────────┘
        │                             │
   ┌────▼────┐                        │
   │ HRM     │                        │
   │ Server  │                        │
   │ (:8081) │                        │
   │ FastAPI │                        │
   └────┬────┘                        │
        │                             │
   ┌────▼───────────────────┐  ┌──────▼────────────────┐
   │  vLLM — SPARK GB10     │  │ Embeddings — SPARK    │
   │  Qwen3.8-27B-FP8       │  │ multilingual-e5-base  │
   │  (:8001, ctx 128k)     │  │ (:8002, 768 dim)      │
   └────────────────────────┘  └───────────────────────┘
```

---

## Modules fonctionnels

### 1. Gestion des affaires

Création, consultation et mise à jour des dossiers d'enquête. Chaque affaire centralise :
- **Entités** : personnes, lieux, objets, organisations, documents, événements
- **Preuves** : physiques, testimoniales, documentaires, numériques, forensiques
- **Relations typées** entre entités (connaît, possède, était_présent, etc.)
- **Chronologie** des événements avec horodatage précis
- **Hypothèses** d'investigation avec scoring de confiance

### 2. Graphe de connaissances interactif

Visualisation réseau des entités et relations d'une affaire via vis-network :
- Navigation interactive (zoom, drag, filtrage par type)
- Menu contextuel sur les nœuds
- Coloration par type d'entité
- Mise en évidence des chemins et clusters

### 3. Analyse de graphe avancée (SSTorytime)

Algorithmes de théorie des graphes appliqués à l'investigation :

| Algorithme | Usage investigatif |
|------------|-------------------|
| **Détection de clusters** | Identifier les groupes organisés |
| **Plus court chemin** | Tracer les chaînes relationnelles |
| **Centralité eigenvector** | Repérer les acteurs influents |
| **Centralité betweenness** | Identifier les intermédiaires clés |
| **Super-nœuds** | Détecter les entités à haute connectivité |
| **Cône d'expansion** | Explorer le voisinage d'une entité |
| **Analyse de densité** | Évaluer la couverture de l'enquête |
| **Patterns temporels** | Détecter les rythmes d'activité |
| **Propagation contra** | Tracer la diffusion des contradictions |
| **Organisation en couches** | Hiérarchiser le réseau |

### 4. Raisonnement hiérarchique (HRM)

Moteur de raisonnement déductif multi-étapes alimenté par un LLM :

**Processus en 3 phases :**
1. **Planification** — Décomposition de la question en sous-tâches (analyser les preuves, identifier les acteurs, construire la chronologie, évaluer les hypothèses)
2. **Exécution** — Résolution séquentielle de chaque sous-tâche avec le contexte de l'affaire
3. **Synthèse** — Consolidation en conclusion structurée avec niveau de confiance

**Sorties :**
- Conclusion principale avec pourcentage de confiance
- Conclusions alternatives classées
- Avertissements sur les limites des preuves
- Findings clés et recommandations

**Interface temps réel :** Streaming SSE avec progression visuelle par phase.

### 5. Assistant conversationnel (Chat IA)

Interface de dialogue naturel pour interroger les données d'une affaire :
- Streaming des réponses en temps réel
- Contexte automatique (entités, preuves, relations)
- Historique de conversation
- Rendu Markdown

### 6. N4L (Narrative 4 Logic)

Langage formel de description narrative pour encoder les relations d'une affaire :

```
Pierre Dupont -[connaît]-> Marie Martin
Couteau -[trouvé_à]-> Scène de crime
```

- **Éditeur** intégré avec coloration syntaxique
- **Parseur** bidirectionnel : N4L ↔ Graphe
- **Générateur** automatique depuis le texte libre
- Support des contextes, alias, groupes, relations chaînées

### 7. Détection d'anomalies

Analyse statistique automatisée pour identifier les points atypiques :
- **Z-score** et **Z-score modifié (MAD)** pour la robustesse
- Anomalies temporelles (horaires inhabituels, fréquences anormales)
- Anomalies comportementales (changements de pattern)
- Anomalies relationnelles (connexions inattendues)
- Système d'alertes avec niveaux de sévérité
- Tableau de bord statistique

### 8. Simulation What-If (Scénarios)

Tester des hypothèses sans altérer les données de l'affaire :
- Création de scénarios hypothétiques
- Propagation automatique des implications
- Scoring de plausibilité
- Comparaison de scénarios alternatifs

### 9. Analyse cross-affaires

Détection de liens entre affaires distinctes :
- Matching d'entités par nom et attributs
- Détection de patterns récurrents (modus operandi)
- Analyse des correspondances avec déduplication stricte
- Scoring de similarité

### 10. Recherche hybride

Moteur de recherche combinant deux approches :
- **BM25** : recherche plein texte classique (mots-clés)
- **Sémantique** : embeddings multilingual-e5-base (sens du texte), scores normalisés min-max
- Classement fusionné pour des résultats pertinents

### 11. Cartographie géographique

Visualisation spatiale via Leaflet :
- Positionnement des lieux sur carte
- Analyse des patterns géographiques
- Corrélation spatiale des événements

### 12. Réseau social

Cartographie relationnelle centrée sur les personnes :
- Visualisation des liens interpersonnels
- Analyse de la structure sociale
- Identification des réseaux d'influence

### 13. Investigation guidée (méthode PEACE)

Workflow structuré pour les interrogatoires et la collecte d'informations :
- Génération automatique de questions pertinentes
- Cadre méthodologique PEACE
- Traçabilité des réponses

### 14. Carnet d'analyse (Notebook)

Centralisation des analyses IA :
- Sauvegarde des résultats d'analyse
- Système de tags et favoris
- Épinglage des notes importantes
- Persistance en JSON

### 15. Configuration des prompts

Interface d'administration pour personnaliser les prompts IA :
- 9 catégories de prompts configurables
- Rechargement à chaud
- Adaptation au contexte d'enquête

---

## Données de démonstration

Six affaires pré-chargées couvrant différents types d'infractions :

| Affaire | Type | Complexité |
|---------|------|------------|
| Affaire Victor Moreau | Homicide par empoisonnement | Élevée |
| Affaire Disparition | Disparition de personne | Moyenne |
| Affaire Fraude | Fraude financière | Moyenne |
| Affaire Cambriolage | Vol avec effraction | Standard |
| Affaire Incendie | Incendie criminel | Moyenne |
| Affaire Trafic Art | Trafic d'œuvres d'art | Élevée |

---

## Déploiement

### Développement local

```bash
# Démarrer les 3 services
./start_services.sh

# Arrêter les services
./stop_services.sh
```

### Production (Ubuntu + systemd)

```bash
sudo scripts/install.sh
sudo systemctl start forensicinvestigator          # Go :8082
sudo systemctl start forensicinvestigator-hrm      # Python :8081
```

Reverse proxy Nginx avec support HTTPS (Let's Encrypt).

### Prérequis

- Go 1.24+
- Python 3.12+ avec venv
- Accès au serveur LLM distant vLLM (ou Ollama local en fallback)
- ~4 Go RAM minimum

---

## Métriques du code

| Composant | Volume |
|-----------|--------|
| Frontend JavaScript | ~27 000 lignes |
| Styles CSS | ~21 000 lignes |
| Backend Go | ~15 000 lignes |
| Services Python | ~2 500 lignes |
| Modules JS | 22 modules |
| Modules CSS | 30 fichiers |
| Endpoints API | 80+ routes |
| Services Go | 11 services |
| **Total** | **~88 000 lignes** |

---

## Sécurité

- Authentification par mot de passe à l'entrée de l'application
- Validation des entrées utilisateur côté serveur
- Thread-safety via sync.RWMutex sur les structures partagées
- Pas de stockage de secrets dans le code (variables d'environnement)
- Configuration Nginx pour HTTPS en production

---

## Roadmap potentielle

| Priorité | Fonctionnalité |
|----------|----------------|
| Haute | Persistance en base de données (PostgreSQL) |
| Haute | Authentification multi-utilisateurs avec rôles |
| Moyenne | Export PDF des rapports d'analyse |
| Moyenne | Intégration avec les fichiers de police existants |
| Moyenne | Mode hors-ligne avec LLM local |
| Basse | Application mobile pour le terrain |
| Basse | API publique pour intégrations tierces |

---

## Points forts différenciants

1. **Raisonnement déductif structuré** — Le HRM décompose les questions complexes en sous-tâches, contrairement aux chatbots qui répondent en un seul passage
2. **Langage formel N4L** — Représentation rigoureuse des relations, exportable et réutilisable
3. **Analyse de graphe native** — Algorithmes de théorie des graphes directement intégrés via SSTorytime
4. **Détection d'anomalies statistique** — Approche quantitative complémentaire au raisonnement qualitatif
5. **Simulation sans altération** — Les scénarios What-If n'impactent pas les données réelles
6. **Interface temps réel** — Streaming SSE pour un feedback immédiat pendant les analyses longues
7. **Architecture modulaire** — 22 modules frontend indépendants, 11 services backend découplés
