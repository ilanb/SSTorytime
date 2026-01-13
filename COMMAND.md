Voici une analyse complète des fichiers d'aide pour les applications Go dans SSTorytime :

## Applications Go Principales et leurs Fonctions

### 1. **N4L et N4L-db** - Processeurs de langage N4L

* **N4L** : Version standalone pour valider et tester les fichiers
* **N4L-db** : Version avec upload en base de données
* **Syntaxe** : `N4L [-v] [-u] [-s] [file].n4l`
* **Options clés** :
  * `-v` : mode verbose
  * `-u` : upload vers PostgreSQL
  * `-s` : résumé (nœuds, liens...)
  * `-adj="pe,he"` : filtrer par types de relations

### 2. **searchN4L** - Moteur de recherche principal

Commandes naturelles supportées :

bash

```bash
./searchN4L Mark                      # Recherche simple
./searchN4L please in chinese         # Recherche contextuelle
./searchN4L notes about chinese context restaurant
./searchN4L from start to "target 1"# Recherche de chemins
./searchN4L chapter brain             # Recherche par chapitre
./searchN4L context restaurant        # Recherche par contexte
./searchN4L arrow ph,pe               # Recherche par flèches
./searchN4L paths from a1 to s1 depth 16
```

### 3. **pathsolve** - Résolveur de chemins

bash

```bash
./pathsolve -begin A1 -end B6
./pathsolve "<B6|A1>"# Notation Dirac
./pathsolve "<target|start>"
```

Analyse les supernodes et la centralité de betweenness.

### 4. **notes** - Navigateur de notes

bash

```bash
./notes fox and crow            # Parcourir par chapitre
./notes -page 2 brain          # Navigation page par page
```

### 5. **text2N4L** - Convertisseur de texte

bash

```bash
./text2N4L filename.txt         # Conversion automatique
./text2N4L -% 77 MobyDick.dat  # Contrôle du pourcentage d'échantillonnage
```

### 6. **graph_report** - Analyseur de graphe

bash

```bash
./graph_report -chapter maze -sttype L
```

Génère des rapports sur :

* Boucles et cycles
* Sources et puits
* Nœuds désignés
* Centralité par vecteur propre

### 7. **removeN4L** - Gestionnaire de suppression

bash

```bash
./removeN4L reminders.n4l
```

### 8. **http_server** - Interface web

Lance un serveur web sur `localhost:8080` pour une interface graphique.

## Langage N4L - Syntaxe Principale

### Structure de base :

n4l

```n4l
# Commentaire
- chapitre                          # Déclaration de chapitre

:: contexte, mots-clés ::           # Tags de contexte
+:: étendre-contexte ::             # Étendre le contexte
-:: supprimer-mots ::               # Supprimer du contexte

# Relations de base
A (relation) B                      # Relation simple
A (relation) B (relation) C         # Chaîne de relations
" (relation) D                      # Continuation
$1 (relation) D                     # Référence au premier élément précédent

@alias                              # Alias pour référence
$alias.1                            # Référence à l'alias

"texte avec espaces"               # Texte quoté
'texte avec "guillemets"'          # Citation alternative
```

### Quatre types de relations sémantiques :

1. **SIMILARITY (0)** - Proximité/similitude

   n4l

   ```n4l
   A (sounds like) B
   A (similar to) B
   ```
2. **LEADSTO (1)** - Causalité/ordre

   n4l

   ```n4l
   A (causes) B
   A (leads to) B
   A (then) B
   ```
3. **CONTAINS (2)** - Contenance/appartenance

   n4l

   ```n4l
   A (contains) B
   A (has part) B
   ```
4. **PROPERTIES (3)** - Attributs/expression

   n4l

   ```n4l
   A (has property) B
   A (means) B
   ```

### Mode séquence :

n4l

```n4l
:: contexte ::
+:: _sequence_ ::                   # Début de séquence
Élément 1
Élément 2                          # Liés automatiquement par "then"
Élément 3
-:: _sequence_ ::                   # Fin de séquence
```

## Configuration - N4Lconfig.in

Structure du fichier de configuration :

n4l

```n4l
- leadsto
 + leads to (lt) - arriving from (af)
 + causes (cf) - is caused by (cb)

- contains
 + contains (c) - is within (in)
 + has component (has) - is component of (part)

- properties
 + means (means) - is meant by (meansb)
 + has property (prop) - is a property of (propof)

- similarity
 similar to (sim)
 looks like (ll)

- annotations
 % (discusses)
 = (depends on)
 * (is a special case of)
 > (has subject)
```

## Flux de travail recommandé

### 1. Création de notes

bash

```bash
# Éditer vos fichiers .n4l
vim mes_notes.n4l

# Valider la syntaxe
./N4L -v mes_notes.n4l

# Tester avec résumé
./N4L -s mes_notes.n4l
```

### 2. Upload en base

bash

```bash
# Premier upload (efface la base)
./N4L-db -wipe -u fichier1.n4l fichier2.n4l

# Ajouts ultérieurs
./N4L-db -u nouveau_fichier.n4l
```

### 3. Recherche et analyse

bash

```bash
# Interface en ligne de commande
./searchN4L "terme de recherche"

# Interface web
./http_server &
# Puis naviguer vers localhost:8080
```

## Exemples pratiques

### Apprentissage de langues :

n4l

```n4l
- notes chinoises

:: nourriture ::

肉 (hp) ròu (pe) meat
牛肉 (hp) niúròu (pe) beef
羊肉 (hp) yángròu (pe) lamb

:: phrases, à l'hôtel ::

@robot Je attends de la nourriture du robot (eh) 我在等机器人送来的食物 (hp) Wǒ zài děng jīqìrén sòng lái de shíwù
```

### Gestion des connaissances :

n4l

```n4l
- projet kubernetes

:: concepts de base ::

pod (represents) la plus petite unité déployable dans Kubernetes
 "  (contains) un ou plusieurs conteneurs
 "  (managed by) contrôleurs

conteneur (managed by) processus containerd
kubectl (use for) interaction utilisateur avec kubernetes via CLI
```
