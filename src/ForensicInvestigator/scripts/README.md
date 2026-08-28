# ForensicInvestigator - Scripts de déploiement

Scripts pour installer, déployer et gérer ForensicInvestigator sur un serveur Ubuntu.

## Prérequis

### Serveur cible
- Ubuntu 20.04 LTS ou supérieur
- Accès root (sudo)
- 2 GB RAM minimum (4 GB recommandé)
- 10 GB espace disque

### Services externes
- **vLLM** (SPARK, port 8001) : inférence LLM. Serveur partagé, utilisé tel quel.
- **Embeddings** (SPARK, port 8002) : `multilingual-e5-base`, recherche sémantique.
- **Ollama** (optionnel) : Pour la conversion N4L (`http://localhost:11434`)

## Installation

### 1. Première installation

Sur le serveur Ubuntu:

```bash
# Cloner le dépôt ou copier les fichiers
git clone <repo> /tmp/forensicinvestigator
cd /tmp/forensicinvestigator/src/ForensicInvestigator/scripts

# Rendre le script exécutable
chmod +x install.sh

# Lancer l'installation
sudo ./install.sh
```

### 2. Configuration

Après l'installation, éditez la configuration:

```bash
sudo nano /opt/forensicinvestigator/config/environment
```

Variables importantes:
```bash
# URL du serveur vLLM (SPARK GB10)
# Poids réellement chargés : Qwen/Qwen3.8-27B-FP8 (128k ctx, MTP self-speculation,
# prefix caching), exposés sous l'alias "Qwen3.5-9B" (--served-model-name).
# C'est cet alias qu'il faut envoyer dans le champ "model" : "Qwen3.8-27B-FP8" -> 404.
# Le SPARK est partagé avec d'autres applications qui dépendent de cet alias :
# ne pas le renommer. L'app Go le découvre de toute façon via /v1/models.
VLLM_URL=http://86.204.69.30:8001
VLLM_MODEL=Qwen3.5-9B

# URL Ollama (pour conversion N4L)
OLLAMA_URL=http://localhost:11434

# Service d'embeddings (SPARK) — recherche hybride
EMBEDDING_BASE_URL=http://86.204.69.30:8002/v1
EMBEDDING_MODEL=multilingual-e5-base
```

### 3. Démarrer l'application

```bash
sudo forensic start
```

## Commandes de gestion

Le script `forensic` est installé dans `/usr/local/bin/`:

```bash
# Démarrer
sudo forensic start

# Arrêter
sudo forensic stop

# Redémarrer
sudo forensic restart

# Voir le statut
sudo forensic status

# Voir les logs en temps réel
sudo forensic logs           # Logs application
sudo forensic logs-hrm       # Logs serveur HRM
sudo forensic logs-all       # Tous les logs

# Sauvegarder les données
sudo forensic backup
```

## Déploiement de mises à jour

Depuis votre machine de développement:

```bash
# Déployer vers le serveur
./deploy.sh user@server

# Avec un port SSH personnalisé
./deploy.sh user@server 2222
```

Le script:
1. Compile l'application pour Linux
2. Envoie les fichiers sur le serveur
3. Sauvegarde les données existantes
4. Met à jour l'application
5. Redémarre les services

## Désinstallation

```bash
# Désinstaller (conserve les données)
sudo ./uninstall.sh

# Désinstaller complètement (supprime tout)
sudo ./uninstall.sh --purge
```

## Structure des fichiers

Après installation:

```
/opt/forensicinvestigator/
├── bin/
│   └── forensicinvestigator    # Binaire principal
├── config/
│   ├── environment             # Variables d'environnement
│   └── prompts.json            # Configuration des prompts IA
├── data/
│   └── notebooks/              # Données persistantes
├── embedding_service/          # Obsolète — repli hors-ligne, non démarré
├── hrm_server/
│   ├── venv/                   # Environnement Python
│   └── *.py                    # Code serveur HRM
├── logs/
│   ├── forensic.log
│   ├── forensic-error.log
│   ├── hrm.log
│   └── hrm-error.log
└── static/
    ├── css/
    ├── js/
    └── index.html
```

## Services systemd

Deux services sont créés:

| Service | Port | Description |
|---------|------|-------------|
| `forensicinvestigator` | 8082 | Application principale (Go) |
| `forensicinvestigator-hrm` | 8081 | Serveur HRM - Hypothetical Reasoning Model (Python) |

Le LLM (8001) et les embeddings (8002) sont fournis par le SPARK, hors de ce serveur.

```bash
# Gérer les services directement
sudo systemctl status forensicinvestigator
sudo systemctl restart forensicinvestigator-hrm
sudo journalctl -u forensicinvestigator -f
```

## Configuration Nginx

Le fichier de configuration Nginx est dans:
```
/etc/nginx/sites-available/forensicinvestigator
```

### Configurer HTTPS

```bash
# Installer un certificat Let's Encrypt
sudo certbot --nginx -d votre-domaine.com
```

### Personnaliser le domaine

Éditez `/etc/nginx/sites-available/forensicinvestigator`:
```nginx
server {
    server_name votre-domaine.com;
    ...
}
```

Puis rechargez Nginx:
```bash
sudo nginx -t && sudo systemctl reload nginx
```

## Dépannage

### L'application ne démarre pas

```bash
# Vérifier les logs
sudo forensic logs
sudo journalctl -u forensicinvestigator -n 50

# Vérifier la configuration
cat /opt/forensicinvestigator/config/environment

# Tester le binaire manuellement
cd /opt/forensicinvestigator
sudo -u forensic ./bin/forensicinvestigator
```

### Erreur de connexion vLLM

```bash
# Tester la connexion vLLM
curl -X POST http://86.204.69.30:8001/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model":"Qwen3.5-9B","messages":[{"role":"user","content":"test"}]}'
```

### Serveur HRM ne répond pas

```bash
# Vérifier le statut
sudo systemctl status forensicinvestigator-hrm

# Vérifier les logs
sudo forensic logs-hrm

# Tester manuellement
curl http://localhost:8081/health
```

### Recherche sémantique dégradée (BM25 seul)

L'application se replie silencieusement sur BM25 si les embeddings sont injoignables.

```bash
# Tester le service d'embeddings du SPARK
curl http://86.204.69.30:8002/v1/models

# Vérifier l'indexation au démarrage
sudo forensic logs | grep REINDEX
```

### Problèmes de permissions

```bash
# Réappliquer les permissions
sudo chown -R forensic:forensic /opt/forensicinvestigator
sudo chmod -R 755 /opt/forensicinvestigator
sudo chmod 600 /opt/forensicinvestigator/config/environment
```

## Sauvegarde et restauration

### Sauvegarde manuelle

```bash
sudo forensic backup
# Crée: /tmp/forensic-backup-YYYYMMDD_HHMMSS.tar.gz
```

### Restauration

```bash
# Arrêter les services
sudo forensic stop

# Extraire la sauvegarde
cd /opt/forensicinvestigator
sudo tar -xzf /tmp/forensic-backup-*.tar.gz

# Réappliquer les permissions
sudo chown -R forensic:forensic data config

# Redémarrer
sudo forensic start
```

## Mise à jour de Go ou Python

Si vous devez mettre à jour Go ou Python sur le serveur:

```bash
# Mettre à jour Go
GO_VERSION="1.22.0"
wget https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz

# Mettre à jour les dépendances Python HRM
cd /opt/forensicinvestigator/hrm_server
source venv/bin/activate
pip install --upgrade fastapi uvicorn
deactivate
```
