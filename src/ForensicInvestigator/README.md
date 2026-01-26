# **FORENSIC INVESTIGATOR**

## Plateforme d'Investigation Intelligente

**Présentation à destination du Ministère de l'Intérieur et de la Police Nationale**

---

## **1. Contexte et Objectif**

Forensic Investigator est une plateforme d'aide à l'investigation qui centralise la gestion des enquêtes, l'analyse de preuves et le raisonnement déductif assisté par l'intelligence artificielle. Elle s'appuie sur plusieurs technologies : **SSTorytime** (moteur d'analyse de graphes), **N4L** (langage de description narrative) **HRM** (moteur de raisonnement à deux niveaux) et les graphe de connaissances

### **Limites des Méthodes Traditionnelles**

Les investigations classiques reposent sur :

* Des dossiers papier ou fichiers dispersés difficiles à croiser
* La mémoire et l'intuition des enquêteurs pour identifier les connexions
* Des tableaux et organigrammes statiques qui deviennent inexploitables au-delà de quelques dizaines d'entités
* Un travail manuel de comparaison entre dossiers chronophage et incomplet
* L'absence d'outil pour tester des hypothèses sans contaminer l'enquête

### **Apports de la Plateforme**

| Technologie                       | Problème Résolu                               | Bénéfice Opérationnel                                                                                                                 |
| --------------------------------- | ----------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------- |
| **Graphe de connaissances** | Données dispersées et non structurées        | Centralisation de toutes les entités, relations et preuves dans une structure interrogeable et visualisable                             |
| **SSTorytime**              | Connexions cachées invisibles à l'œil humain | Algorithmes de détection de chemins, clusters, nœuds-ponts et patterns dans des réseaux de centaines d'entités                       |
| **N4L**                     | Saisie longue et ambiguïté des descriptions   | Langage formel permettant de décrire précisément les relations et de générer automatiquement des graphes depuis des documents texte |
| **HRM**                     | Raisonnement limité par la charge cognitive    | Moteur de déduction qui analyse l'ensemble des preuves, génère des hypothèses alternatives et identifie les contradictions           |
| **Analyse inter-dossiers**  | Cloisonnement des enquêtes                     | Détection automatique de correspondances entre affaires distinctes (entités, modus operandi, lieux)                                    |
| **Simulation What-If**      | Impossibilité de tester sans risque            | Exploration d'hypothèses sans modifier les données réelles, avec calcul des implications                                              |
| **Détection d'anomalies**  | Valeurs aberrantes noyées dans la masse        | Identification statistique des incohérences temporelles, comportementales ou relationnelles                                             |
| **Recherche sémantique**   | Recherche par mots-clés trop restrictive       | Compréhension du sens des requêtes, pas seulement des termes exacts                                                                    |

### **Résultat**

L'enquêteur dispose d'un assistant qui :

* Mémorise et structure l'intégralité du dossier
* Identifie des connexions que l'analyse manuelle ne peut pas détecter
* Propose des pistes de raisonnement traçables et vérifiables
* Permet de tester des scénarios avant de les poursuivre sur le terrain
* Alerte sur les incohérences et les correspondances avec d'autres affaires

La plateforme ne remplace pas l'enquêteur. Elle amplifie sa capacité d'analyse sur des dossiers complexes comportant de nombreuses entités et relations.

---

## **2. Gestion des Enquêtes**

**Dossiers d'investigation**

* Création et suivi de dossiers (homicides, vols, fraudes, etc.)
* Statuts : en cours, résolu, classé
* Traçabilité des modifications

**Entités**

* Gestion des personnes, lieux, objets, événements, organisations, documents
* Classification par rôle : victime, suspect, témoin, enquêteur
* Relations multiples entre entités avec typage

**Preuves**

* Types : physiques, testimoniales, documentaires, numériques, médico-légales
* Chaîne de custody intégrée
* Score de fiabilité (échelle 1-10)
* Métadonnées de collecte (date, lieu, collecteur)

**Chronologie**

* Organisation temporelle des événements
* Niveau d'importance (élevé, moyen, faible)
* Statut de vérification et source

**Hypothèses**

* Création et gestion d'hypothèses d'investigation
* Classification par statut : en cours d'évaluation, confirmée, réfutée
* Niveau de confiance associé
* Liaison avec les preuves à charge et à décharge
* Vérification automatique par le module HRM

**Connexions Inter-Dossiers**

* Détection automatique de liens entre enquêtes distinctes
* Correspondance d'entités (noms, attributs, rôles)
* Identification de modus operandi similaires
* Corrélation de lieux et périodes
* Alertes sur correspondances fortes (>85% confiance)
* Visualisation du réseau de dossiers liés

**Simulation What-If**

* Création de scénarios hypothétiques
* Modification virtuelle d'entités, preuves ou relations
* Calcul des implications en cascade
* Score de plausibilité
* Comparaison côte à côte de scénarios
* Aucune modification des données réelles

**Détection d'Anomalies**

* Analyse statistique (Z-score, MAD)
* Identification d'anomalies temporelles, comportementales, relationnelles
* Système d'alertes avec classification par gravité
* Tableau de bord statistique
* Historique et acquittement des alertes

**Carte Géographique**

* Visualisation des lieux sur carte interactive (Leaflet)
* Positionnement des entités géolocalisées
* Affichage des événements par localisation
* Analyse spatiale des patterns

**Notebook**

* Centralisation des analyses IA générées
* Prise de notes sur l'investigation
* Marquage et organisation par tags
* Épinglage des analyses importantes
* Historique des décisions et raisonnements
* Tableau de bord des statistiques d'analyse

---

## **3. N4L - Langage de Description Narrative**

N4L (Narrative 4 Logic) permet de décrire formellement les relations et les récits d'une enquête.

**Syntaxe**

```
:: contexte_enquête ::
Suspect_A -> possède -> Arme_1
Témoin_B (a_vu) Suspect_A (quitter) Lieu_crime
Victime <-> connaissait -> Suspect_A
```

**Fonctionnalités**

* Relations directionnelles et bidirectionnelles
* Chaînes causales automatiques (A → B → C)
* Contextes sémantiques modulables
* Alias et références croisées
* Conversion bidirectionnelle N4L ↔ graphe visuel
* Import/export de fichiers N4L

**Génération Automatique depuis Documents Texte**

Un modèle de langage fine-tuné spécifiquement pour N4L permet de :

* Analyser des documents textuels (procès-verbaux, rapports, témoignages)
* Extraire automatiquement les entités (personnes, lieux, objets, événements)
* Identifier les relations entre entités
* Générer un fichier N4L structuré à partir du texte brut
* Accélérer la saisie initiale d'un dossier d'enquête

Cette fonctionnalité permet de transformer rapidement des documents d'investigation existants en graphe de connaissances exploitable par la plateforme.

---

## **4. Visualisation et Analyse de Graphe**

**Visualisation interactive**

* Réseau de nœuds et d'arêtes en temps réel
* Coloration par type d'entité et rôle
* Manipulation par glisser-déposer
* Zoom, défilement, réorganisation dynamique

**Algorithmes d'analyse (hérités de SSTorytime)**

| Fonction                       | Description                                           |
| ------------------------------ | ----------------------------------------------------- |
| Détection de clusters         | Identification de communautés dans le réseau        |
| Recherche de chemins           | Plus court chemin entre entités, connexions cachées |
| Graphe en couches              | Organisation hiérarchique, analyse de flux           |
| Cône d'expansion              | Exploration radiale depuis un nœud central           |
| Cartographie de densité       | Zones explorées, partielles, inexplorées            |
| Motifs temporels               | Séquences, cycles, intervalles                       |
| Centralité eigenvector        | Score d'influence des entités                        |
| Centralité d'intermédiarité | Identification des nœuds-ponts                       |
| Détection de super-nœuds     | Entités à forte connectivité                       |
| Recherche contrawave           | Propagation de contradictions                         |

---

## **5. HRM - Module de Raisonnement Hiérarchique**

### **Architecture**

Le HRM (Hierarchical Reasoning Model) est un moteur de raisonnement à deux niveaux :

* **Niveau 1** : Modèle HRM spécialisé (sapientinc/HRM-checkpoint) pour la reconnaissance de patterns hiérarchiques
* **Niveau 2** : Modèle de langage Qwen2.5-7B-Instruct (via vLLM) pour le raisonnement en langage naturel

### **Processus de Raisonnement Déductif**

Le HRM traite les questions complexes en trois phases :

**Phase 1 - Planification**

* Décomposition de la question en sous-tâches
* Identification des axes d'analyse : analyse des preuves, identification des acteurs, construction de la chronologie, évaluation des hypothèses
* Génération d'un plan d'exécution structuré

**Phase 2 - Exécution**

* Traitement séquentiel de chaque sous-tâche
* Collecte des inférences intermédiaires
* Suivi des prémisses utilisées à chaque étape
* Accumulation des résultats partiels

**Phase 3 - Synthèse**

* Agrégation des résultats des sous-tâches
* Génération de la conclusion principale avec niveau de confiance
* Production de conclusions alternatives avec leurs probabilités respectives
* Identification des avertissements (limites des preuves, interprétations incertaines)
* Extraction des découvertes clés et recommandations

### **Capacités de Raisonnement**

| Fonction                     | Description                                                                                   |
| ---------------------------- | --------------------------------------------------------------------------------------------- |
| Raisonnement déductif       | Analyse par chaîne de pensée avec traçabilité complète des preuves utilisées            |
| Inférence multi-étapes     | Suivi explicite des prémisses à chaque étape du raisonnement                               |
| Conclusions alternatives     | Génération de plusieurs hypothèses avec score de confiance (ex: 75%, 60%, 45%)             |
| Avertissements               | Identification des limites : preuves circonstancielles, témoignages non corroborés, lacunes |
| Vérification d'hypothèses  | Validation d'énoncés contre les preuves, éléments à charge et à décharge               |
| Détection de contradictions | Identification de conflits entre déclarations, classification par gravité                   |
| Raisonnement inter-dossiers  | Détection de patterns et connexions entre plusieurs enquêtes                                |

### **Interface de Visualisation**

L'interface HRM affiche en temps réel :

* **Progression** : Barre de progression par phase (planification → exécution → synthèse)
* **Sous-tâches** : Liste des axes d'analyse avec statut (en attente, en cours, terminé)
* **Conclusion principale** : Résultat avec niveau de confiance
* **Alternatives** : Autres hypothèses classées par probabilité
* **Avertissements** : Limites et points d'attention identifiés
* **Découvertes clés** : Éléments factuels importants extraits
* **Recommandations** : Pistes d'investigation suggérées

### **Streaming**

Le HRM fonctionne en mode streaming :

* Affichage progressif des résultats pendant le traitement
* Mise à jour en temps réel de l'interface
* Possibilité d'annuler une requête longue
* Timeout configurable (défaut : 1 heure pour les raisonnements complexes)

### **Exemple de Sortie HRM**

```
Question : "Qui est le principal suspect dans le meurtre de Victor Moreau ?"

[Phase: Planification] ████████████ 100%
  ✓ analyze_evidence
  ✓ identify_actors  
  ✓ build_timeline
  ✓ evaluate_hypotheses

[Conclusion Principale] (Confiance: 78%)
Jean Dupont présente le profil le plus cohérent avec les preuves 
disponibles : présence sur les lieux confirmée par vidéosurveillance,
mobile financier établi, absence d'alibi vérifiable.

[Alternatives]
• Marie Laurent (22%) - Accès aux lieux mais mobile non établi
• Inconnu (15%) - Traces ADN non identifiées sur la scène

[Avertissements]
⚠ Témoignage de Pierre Martin non corroboré
⚠ Heure du décès estimée avec marge de ±2 heures
⚠ Preuves principalement circonstancielles

[Découvertes Clés]
• Transaction bancaire suspecte 48h avant les faits
• Appel téléphonique entre suspect et victime à 22h47

[Recommandations]
• Vérifier l'alibi de Jean Dupont auprès de son employeur
• Analyser les relevés téléphoniques complets
• Identifier la source des traces ADN inconnues
```

### **Intégration avec les Autres Modules**

* **Graphe** : Les entités et relations identifiées par le HRM peuvent être ajoutées au graphe
* **Hypothèses** : Les conclusions générées alimentent le module de gestion d'hypothèses
* **Notebook** : Les analyses HRM sont centralisées dans le carnet d'investigation
* **Anomalies** : Le HRM peut être utilisé pour analyser les anomalies détectées
* **Scénarios** : Les hypothèses HRM peuvent servir de base aux simulations

## **6. Détection d'Anomalies**

**Méthodes statistiques**

* Analyse Z-score et Z-score modifié (MAD)
* Calcul moyenne/médiane/écart-type
* Identification de valeurs aberrantes

**Types d'anomalies détectées**

* Anomalies temporelles
* Déviations comportementales
* Valeurs aberrantes dans les preuves
* Anomalies relationnelles

**Système d'alertes**

* Classification par gravité
* Historique et acquittement
* Tableau de bord statistique

---

## **7. Simulation de Scénarios**

**Analyse "Et si..."**

* Modification d'hypothèses sur les entités
* Simulation de changements de preuves
* Hypothèses relationnelles
* Altérations de chronologie

**Résultats**

* Calcul des implications
* Identification des faits en support ou en contradiction
* Score de plausibilité
* Visualisation du graphe modifié
* Comparaison côte à côte de scénarios

**Propagation**

* Analyse en cascade des hypothèses
* Détection des réactions en chaîne

---

## **8. Analyse Inter-Dossiers**

**Correspondance d'entités**

* Similarité de noms (correspondance floue, distance de Levenshtein)
* Comparaison d'attributs
* Correspondance par rôle
* Analyse de similarité sémantique

**Corrélations**

* Lieux communs entre enquêtes
* Modus operandi similaires
* Chevauchements temporels
* Preuves similaires

**Détection de patterns**

* Séquences d'événements similaires (3+ événements)
* Reconnaissance de patterns comportementaux
* Correspondance d'attributs professionnels

**Alertes**

* Correspondances fortes (>85% confiance)
* Alertes de connexion (10+ correspondances avec un dossier)

---

## **9. Outils d'Analyse par IA**

**Chat conversationnel**

* Dialogue en temps réel avec streaming
* Questions contextuelles sur l'enquête
* Réponses fondées sur les preuves

**Génération automatique**

* Hypothèses à partir des preuves
* Questions d'investigation clés
* Identification des lacunes

**Recherche hybride**

* Recherche plein texte (BM25)
* Recherche sémantique (embeddings Model2vec)
* Classement combiné

---

## **10. Interface Utilisateur**

**22 modules fonctionnels**

| Module         | Fonction                          |
| -------------- | --------------------------------- |
| Dashboard      | Vue d'ensemble et métriques      |
| Graph          | Visualisation réseau interactive |
| Timeline       | Chronologie des événements      |
| Entities       | Gestion personnes/lieux/objets    |
| Evidence       | Suivi des preuves                 |
| Hypotheses     | Gestion et test d'hypothèses     |
| Chat           | Interface conversationnelle IA    |
| Graph-Analysis | Algorithmes avancés              |
| N4L            | Éditeur et visualisation N4L     |
| Anomalies      | Tableau de bord anomalies         |
| Scenarios      | Simulation "Et si..."             |
| Cross-Case     | Analyse multi-dossiers            |
| HRM            | Interface de raisonnement         |
| Geo-Map        | Visualisation géographique       |
| Social-Network | Cartographie relationnelle        |
| Notebook       | Centralisation des analyses IA    |
| Search         | Recherche globale                 |
| Investigation  | Mode investigation guidée        |

---

## **11. Architecture Technique**

**Services**

* Application principale Go (port 8082)
* Serveur HRM Python (port 8081)
* Service d'embeddings Model2vec (port 8085)
* Inférence vLLM distante

**Déploiement**

* Services systemd pour serveurs Linux
* Support proxy Nginx
* Compatible HTTPS/Let's Encrypt
* Prêt pour conteneurisation

**Données**

* Export/import N4L, JSON, CSV
* Sauvegarde des dossiers
* Migration de données

---

## **12. Relation SSTorytime / N4L / Forensic Investigator**

```
┌─────────────────────────────────────────────────────┐
│                    SSTorytime                        │
│         (Moteur d'analyse de graphes)               │
│  - Algorithmes de centralité                        │
│  - Détection de patterns                            │
│  - Recherche de chemins contraints                  │
└───────────────────────┬─────────────────────────────┘
                        │
                        ▼
┌─────────────────────────────────────────────────────┐
│                       N4L                            │
│        (Langage de description narrative)           │
│  - Syntaxe formelle pour relations                  │
│  - Interopérabilité sémantique                      │
│  - Format d'échange standard                        │
└───────────────────────┬─────────────────────────────┘
                        │
                        ▼
┌─────────────────────────────────────────────────────┐
│              Forensic Investigator                   │
│          (Plateforme d'investigation)               │
│  - Interface utilisateur métier                     │
│  - Gestion des enquêtes                             │
│  - Intégration IA (HRM)                             │
└─────────────────────────────────────────────────────┘
```

---

## **13. Points Clés**

1. **Raisonnement structuré** - Analyse multi-niveaux avec traçabilité des preuves
2. **Intelligence inter-dossiers** - Détection automatique de patterns entre enquêtes
3. **IA fondée sur les preuves** - Conclusions ancrées dans les éléments factuels
4. **Test d'hypothèses** - Simulation sans contamination des données
5. **Détection de contradictions** - Identification automatique d'incohérences
6. **Analyse de réseau** - Capacités SSTorytime intégrées
7. **Langage formel** - N4L pour description précise des relations
8. **Recherche sémantique** - Compréhension du sens, pas seulement des mots-clés

---

## **14. Dimensionnement**

* **80+ points d'API**
* **~26 000 lignes** JavaScript frontend (modulaire)
* **~5 000 lignes** Go backend
* **~25 000 lignes** Python services IA
