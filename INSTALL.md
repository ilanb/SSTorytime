# Instructions complètes d'installation SSTorytime pour un nouveau projet

## Prérequis système (Mac)

### 1. Installer les outils de base

bash

```bash
# Installer Homebrew si pas déjà fait
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# Installer Git si nécessaire
brew installgit

# Installer Go
brew install go

# Installer PostgreSQL
brew install postgresql@15

# Vérifier les installations
go version
psql --version
```

## 2. Configuration de l'environnement Go

bash

```bash
# Ajouter à ~/.zshrc ou ~/.bash_profile
echo'export PATH=$PATH:/usr/local/go/bin'>> ~/.zshrc
echo'export PATH=$PATH:$(go env GOPATH)/bin'>> ~/.zshrc
echo'export GO111MODULE=on'>> ~/.zshrc

# Recharger le shell
source ~/.zshrc

# Vérifier la configuration
go env GOPATH
go env GOROOT
```

## 3. Installation et configuration PostgreSQL

### Démarrer PostgreSQL

bash

```bash
# Démarrer le service
brew services start postgresql@15

# Vérifier que PostgreSQL fonctionne
brew services list |grep postgresql
```

### Configurer la base de données

bash

```bash
# Se connecter en tant qu'utilisateur postgres
psql postgres

# Dans psql, créer l'utilisateur et la base
CREATE USER sstoryline WITH PASSWORD 'sst_1234' SUPERUSER;
CREATE DATABASE sstoryline;
GRANT ALL PRIVILEGES ON DATABASE sstoryline TO sstoryline;
CREATE EXTENSION UNACCENT;

# Tester la connexion
\l
\q
```

### Configuration de l'authentification

bash

```bash
# Localiser le fichier de configuration
find /opt/homebrew -name "pg_hba.conf"2>/dev/null

# Éditer le fichier (chemin typique sur Mac avec Homebrew)
sudonano /opt/homebrew/var/postgresql@15/pg_hba.conf

# Modifier les lignes pour utiliser 'password' au lieu de 'trust' :
# host    all             all             127.0.0.1/32            password
# host    all             all             ::1/128                 password

# Redémarrer PostgreSQL
brew services restart postgresql@15

# Tester la connexion avec mot de passe
psql -U sstoryline -d sstoryline -h localhost
```

## 4. Cloner et configurer SSTorytime

### Cloner le projet

bash

```bash
# Créer un dossier de travail
mkdir -p ~/projets/sstorytime
cd ~/projets/sstorytime

# Cloner le repository
git clone https://github.com/markburgess/SSTorytime.git
cd SSTorytime

# Vérifier la structure
ls -la
```

### Configuration des modules Go

bash

```bash
# Activer les modules Go
exportGO111MODULE=on

# Installer les dépendances PostgreSQL pour Go (mode GOPATH)
exportGO111MODULE=off
exportGOPATH=$(pwd)
exportPATH=$PATH:$GOPATH/bin

# Créer la structure GOPATH
mkdir -p src/github.com/lib
mkdir -p src/github.com/jmoiron

# Télécharger les dépendances manuellement
cd /tmp
git clone https://github.com/lib/pq.git
git clone https://github.com/jmoiron/sqlx.git

# Copier dans la structure GOPATH
cp -r pq ~/projets/sstorytime/SSTorytime/src/github.com/lib/
cp -r sqlx ~/projets/sstorytime/SSTorytime/src/github.com/jmoiron/

# Retourner au projet
cd ~/projets/sstorytime/SSTorytime

# Copier le package SSTorytime
mkdir -p src/SSTorytime
cp pkg/SSTorytime/SSTorytime.go src/SSTorytime/
```

## 5. Compilation du projet

bash

```bash
# Compiler tous les outils
make clean
make all

# Vérifier que la compilation a réussi
ls -la src/
```

Vous devriez voir les exécutables :

* `N4L`, `N4L-db`
* `searchN4L`, `pathsolve`
* `text2N4L`, `removeN4L`
* `notes`, `graph_report`
* `http_server`
* `API_EXAMPLE_1`, `API_EXAMPLE_2`, `API_EXAMPLE_3`

## 6. Test de l'installation

### Tester la connectivité à la base

bash

```bash
# Créer un fichier de test
cat> test_db_connection.go <<'EOF'
package main

import (
    "database/sql"
    "fmt"
    _ "github.com/lib/pq"
)

func main() {
    connStr := "user=sstoryline password=sst_1234 dbname=sstoryline host=localhost port=5432 sslmode=disable"
  
    db, err := sql.Open("postgres", connStr)
    if err != nil {
        fmt.Printf("Erreur connexion : %v\n", err)
        return
    }
    defer db.Close()

    err = db.Ping()
    if err != nil {
        fmt.Printf("Erreur ping : %v\n", err)
        return
    }
  
    fmt.Println("✅ Connexion PostgreSQL réussie!")
}
EOF

# Tester la connexion
go run test_db_connection.go
```

### Charger les exemples

bash

```bash
cd examples

# Compiler les exemples
make

# Charger les données d'exemple
../src/N4L-db -wipe -u doors.n4l Mary.n4l chinese*.n4l branches.n4l doubleslit.n4l ConstructionProcesses.n4l wardleymap.n4l brains.n4l kubernetes.n4l SSTorytime.n4l integral.n4l reasoning.n4l

# Ajouter des données supplémentaires
../src/N4L-db -u LoopyLoo.n4l
```

## 7. Tester les outils

### Test N4L standalone

bash

```bash
cd src

# Créer un fichier de test
cat> test.n4l <<'EOF'
- test chapitre

:: apprentissage, test ::

Je (contient) connaissances
connaissances (exprime) compétences  
compétences (contient) exemples
exemples (généralise) cas pratiques
EOF

# Tester N4L
./N4L -v test.n4l
./N4L -s test.n4l
```

### Test recherche

bash

```bash
# Tester searchN4L
echo"brain"| ./searchN4L
./searchN4L "Mary had"
./searchN4L notes chapter brain
./searchN4L from start to target
```

### Test interface web

bash

```bash
# Démarrer le serveur web
./http_server &

# Tester dans le navigateur
open http://localhost:8080

# Arrêter le serveur quand terminé
killall http_server
```

## 8. Configuration VSCode

bash

```bash
# Ouvrir le projet dans VSCode
code .
```

### Extensions recommandées VSCode :

* **Go** (par Google)
* **PostgreSQL** (par Chris Kolkman)
* **Go Test Explorer** (optionnel)

### Configuration VSCode (`.vscode/settings.json`) :

json

```json
{
"go.gopath":"",
"go.goroot":"",
"go.formatTool":"goimports",
"go.useLanguageServer":true,
"go.testFlags":["-v"],
"files.associations":{
"*.n4l":"plaintext"
}
}
```

## 9. Workflow quotidien

### Créer vos propres notes

bash

```bash
# Créer un nouveau fichier de notes
cat> mes_notes.n4l <<'EOF'
- mes notes personnelles

:: programmation, go ::

Go (est un) langage de programmation
  " (créé par) Google
  " (utilisé pour) SSTorytime
  " (permet) concurrence native

:: base de données ::

PostgreSQL (est une) base de données relationnelle
         " (supporte) extensions
         " (utilisé par) SSTorytime
EOF

# Valider
./N4L -v mes_notes.n4l

# Uploader
./N4L-db -u mes_notes.n4l

# Rechercher
./searchN4L "Go"
./searchN4L "PostgreSQL"
```

### Mise à jour des notes

bash

```bash
# Méthode recommandée : recharger tout
./N4L-db -wipe -u *.n4l

# Ou pour des ajouts ponctuels
./N4L-db -u nouveau_fichier.n4l
```

## 10. Résolution des problèmes courants

### Si PostgreSQL ne démarre pas :

bash

```bash
# Vérifier les logs
tail -f /opt/homebrew/var/log/postgresql@15.log

# Réinitialiser si nécessaire
brew services stop postgresql@15
rm -rf /opt/homebrew/var/postgresql@15/
brew services start postgresql@15
```

### Si Go ne compile pas :

bash

```bash
# Vérifier les variables d'environnement
echo$GOPATH
echo$GO111MODULE

# Nettoyer et recompiler
make clean
go clean -cache
make all
```

### Si la base de données ne se connecte pas :

bash

```bash
# Vérifier le service
brew services list |grep postgresql

# Tester la connexion manuelle
psql -U sstoryline -d sstoryline -h localhost

# Vérifier la configuration
cat /opt/homebrew/var/postgresql@15/pg_hba.conf
```

## 11. Prochaines étapes

1. **Lire la documentation** dans le dossier `docs/`
2. **Étudier les exemples** dans `examples/`
3. **Créer vos propres notes** en N4L
4. **Explorer l'API** avec les fichiers `API_EXAMPLE_*.go`
5. **Développer vos propres outils** avec l'API Go

Votre installation SSTorytime est maintenant complète et prête à l'emploi ! 🎉
