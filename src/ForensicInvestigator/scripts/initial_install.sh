#!/bin/bash
#
# ForensicInvestigator - Script d'installation initiale depuis la machine locale
# Usage: ./initial_install.sh [user@]server [ssh_key] [port]
#
# Ce script:
# 1. Prépare et envoie tous les fichiers nécessaires
# 2. Lance l'installation sur le serveur distant
#

set -e

# Configuration par défaut
DEFAULT_PORT=22
APP_NAME="forensicinvestigator"

# Couleurs
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[OK]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Vérifier les arguments
if [ -z "$1" ]; then
    echo "Usage: $0 [user@]server [ssh_key] [port]"
    echo ""
    echo "Exemples:"
    echo "  $0 ubuntu@51.75.240.95 ~/.ssh/id_ed25519_ovh"
    echo "  $0 root@192.168.1.100 ~/.ssh/id_rsa 2222"
    echo ""
    exit 1
fi

SERVER="$1"
SSH_KEY="${2:-}"
PORT="${3:-$DEFAULT_PORT}"

# Construire les options SSH
SSH_OPTS="-o StrictHostKeyChecking=no -o ConnectTimeout=30"
if [ -n "$SSH_KEY" ]; then
    SSH_OPTS="$SSH_OPTS -i $SSH_KEY"
fi
SSH_OPTS="$SSH_OPTS -p $PORT"

SCP_OPTS="-o StrictHostKeyChecking=no"
if [ -n "$SSH_KEY" ]; then
    SCP_OPTS="$SCP_OPTS -i $SSH_KEY"
fi
SCP_OPTS="$SCP_OPTS -P $PORT"

# Déterminer le répertoire source
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SOURCE_DIR="$(dirname "$SCRIPT_DIR")"

echo ""
echo "=============================================="
echo "  ForensicInvestigator - Installation initiale"
echo "=============================================="
echo ""
echo "Serveur: ${SERVER}"
echo "Port: ${PORT}"
echo "Clé SSH: ${SSH_KEY:-'défaut'}"
echo ""

# Test de connexion
log_info "Test de connexion SSH..."
if ! ssh $SSH_OPTS "${SERVER}" "echo 'Connexion OK'" 2>/dev/null; then
    log_error "Impossible de se connecter au serveur"
    exit 1
fi
log_success "Connexion établie"

# 1. Compiler l'application pour Linux
log_info "Compilation de l'application pour Linux amd64..."
cd "${SOURCE_DIR}"
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o "${APP_NAME}_linux" .
log_success "Application compilée"

# 2. Créer une archive avec tous les fichiers nécessaires
log_info "Préparation de l'archive d'installation..."
TEMP_DIR=$(mktemp -d)
INSTALL_DIR="${TEMP_DIR}/forensicinvestigator"
mkdir -p "${INSTALL_DIR}"

# Copier tous les fichiers nécessaires
cp "${APP_NAME}_linux" "${INSTALL_DIR}/bin_${APP_NAME}"
cp -r static "${INSTALL_DIR}/"
cp -r config "${INSTALL_DIR}/" 2>/dev/null || mkdir -p "${INSTALL_DIR}/config"
cp -r data "${INSTALL_DIR}/" 2>/dev/null || mkdir -p "${INSTALL_DIR}/data"
mkdir -p "${INSTALL_DIR}/data/notebooks"

# Copier hrm_server sans venv
mkdir -p "${INSTALL_DIR}/hrm_server"
find hrm_server -maxdepth 1 -type f \( -name "*.py" -o -name "*.txt" -o -name "*.json" \) -exec cp {} "${INSTALL_DIR}/hrm_server/" \; 2>/dev/null || true

# Copier embedding_service sans venv
mkdir -p "${INSTALL_DIR}/embedding_service"
find embedding_service -maxdepth 1 -type f \( -name "*.py" -o -name "*.txt" -o -name "*.json" \) -exec cp {} "${INSTALL_DIR}/embedding_service/" \; 2>/dev/null || true

# Copier les scripts d'installation
cp -r scripts "${INSTALL_DIR}/"

# Créer l'archive
cd "${TEMP_DIR}"
tar -czf install_package.tar.gz forensicinvestigator
log_success "Archive créée ($(du -h install_package.tar.gz | cut -f1))"

# 3. Envoyer l'archive sur le serveur
log_info "Envoi de l'archive vers le serveur..."
scp $SCP_OPTS "${TEMP_DIR}/install_package.tar.gz" "${SERVER}:/tmp/"
log_success "Archive envoyée"

# 4. Lancer l'installation sur le serveur
log_info "Lancement de l'installation sur le serveur..."
ssh $SSH_OPTS "${SERVER}" << 'REMOTE_INSTALL'
set -e

echo ""
echo "=============================================="
echo "  Installation sur le serveur"
echo "=============================================="
echo ""

# Vérifier root
if [ "$EUID" -ne 0 ] && [ "$(id -u)" -ne 0 ]; then
    SUDO="sudo"
else
    SUDO=""
fi

# Configuration
APP_NAME="forensicinvestigator"
APP_DIR="/opt/forensicinvestigator"
APP_USER="forensic"
APP_GROUP="forensic"

# Couleurs
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[OK]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[WARN]${NC} $1"; }

# Extraire l'archive
log_info "Extraction de l'archive..."
cd /tmp
tar -xzf install_package.tar.gz
cd forensicinvestigator

# Installer les dépendances système
log_info "Installation des dépendances système..."
$SUDO apt-get update -qq
$SUDO apt-get install -y -qq python3 python3-pip python3-venv nginx certbot python3-certbot-nginx jq > /dev/null

# Créer l'utilisateur système
log_info "Création de l'utilisateur ${APP_USER}..."
if ! id "$APP_USER" &>/dev/null; then
    $SUDO useradd -r -s /bin/false -m -d /home/${APP_USER} ${APP_USER}
    log_success "Utilisateur ${APP_USER} créé"
else
    log_info "L'utilisateur ${APP_USER} existe déjà"
fi

# Créer les répertoires
log_info "Création des répertoires..."
$SUDO mkdir -p ${APP_DIR}/{bin,config,data,logs,static}
$SUDO mkdir -p ${APP_DIR}/data/notebooks
$SUDO mkdir -p ${APP_DIR}/hrm_server
$SUDO mkdir -p ${APP_DIR}/embedding_service

# Copier les fichiers
log_info "Copie des fichiers..."
$SUDO cp bin_${APP_NAME} ${APP_DIR}/bin/${APP_NAME}
$SUDO chmod +x ${APP_DIR}/bin/${APP_NAME}
$SUDO cp -r static/* ${APP_DIR}/static/
$SUDO cp -r config/* ${APP_DIR}/config/ 2>/dev/null || true
$SUDO cp -r data/* ${APP_DIR}/data/ 2>/dev/null || true
$SUDO cp -r hrm_server/* ${APP_DIR}/hrm_server/ 2>/dev/null || true
$SUDO cp -r embedding_service/* ${APP_DIR}/embedding_service/ 2>/dev/null || true

# Configuration du serveur HRM
log_info "Configuration du serveur HRM (Python)..."
cd ${APP_DIR}/hrm_server
$SUDO python3 -m venv venv
$SUDO ${APP_DIR}/hrm_server/venv/bin/pip install --upgrade pip -q
$SUDO ${APP_DIR}/hrm_server/venv/bin/pip install fastapi uvicorn pydantic httpx aiohttp -q
log_success "Serveur HRM configuré"

# Configuration du service Embedding
log_info "Configuration du service Embedding (Model2vec)..."
cd ${APP_DIR}/embedding_service
$SUDO python3 -m venv venv
$SUDO ${APP_DIR}/embedding_service/venv/bin/pip install --upgrade pip -q
if [ -f requirements.txt ]; then
    $SUDO ${APP_DIR}/embedding_service/venv/bin/pip install -r requirements.txt -q
else
    $SUDO ${APP_DIR}/embedding_service/venv/bin/pip install fastapi uvicorn model2vec numpy pydantic -q
fi
log_success "Service Embedding configuré"

# Créer le fichier de configuration
log_info "Création de la configuration..."
$SUDO tee ${APP_DIR}/config/environment > /dev/null << 'EOF'
# ForensicInvestigator - Configuration
VLLM_URL=http://86.204.69.30:8001
VLLM_MODEL=Qwen/Qwen2.5-7B-Instruct
OLLAMA_URL=http://localhost:11434
OLLAMA_MODEL=n4l-qwen:latest
FORENSIC_PORT=8082
HRM_PORT=8081
EMBEDDING_PORT=8085
LOG_LEVEL=info
EOF

# Créer les services systemd
log_info "Création des services systemd..."

# Service principal
$SUDO tee /etc/systemd/system/forensicinvestigator.service > /dev/null << EOF
[Unit]
Description=ForensicInvestigator - Application d'investigation forensique
After=network.target
Wants=forensicinvestigator-hrm.service forensicinvestigator-embedding.service

[Service]
Type=simple
User=${APP_USER}
Group=${APP_GROUP}
WorkingDirectory=${APP_DIR}
EnvironmentFile=${APP_DIR}/config/environment
ExecStart=${APP_DIR}/bin/${APP_NAME}
Restart=always
RestartSec=5
StandardOutput=append:${APP_DIR}/logs/forensic.log
StandardError=append:${APP_DIR}/logs/forensic-error.log
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=${APP_DIR}/data ${APP_DIR}/logs
PrivateTmp=true

[Install]
WantedBy=multi-user.target
EOF

# Service HRM
$SUDO tee /etc/systemd/system/forensicinvestigator-hrm.service > /dev/null << EOF
[Unit]
Description=ForensicInvestigator HRM Server
After=network.target

[Service]
Type=simple
User=${APP_USER}
Group=${APP_GROUP}
WorkingDirectory=${APP_DIR}/hrm_server
EnvironmentFile=${APP_DIR}/config/environment
ExecStart=${APP_DIR}/hrm_server/venv/bin/python -m uvicorn main:app --host 0.0.0.0 --port 8081
Restart=always
RestartSec=5
StandardOutput=append:${APP_DIR}/logs/hrm.log
StandardError=append:${APP_DIR}/logs/hrm-error.log
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=${APP_DIR}/logs
PrivateTmp=true

[Install]
WantedBy=multi-user.target
EOF

# Service Embedding
$SUDO tee /etc/systemd/system/forensicinvestigator-embedding.service > /dev/null << EOF
[Unit]
Description=ForensicInvestigator Embedding Server - Model2vec
After=network.target

[Service]
Type=simple
User=${APP_USER}
Group=${APP_GROUP}
WorkingDirectory=${APP_DIR}/embedding_service
EnvironmentFile=${APP_DIR}/config/environment
ExecStart=${APP_DIR}/embedding_service/venv/bin/python main.py
Restart=always
RestartSec=5
StandardOutput=append:${APP_DIR}/logs/embedding.log
StandardError=append:${APP_DIR}/logs/embedding-error.log
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=${APP_DIR}/logs
PrivateTmp=true

[Install]
WantedBy=multi-user.target
EOF

# Configuration Nginx
log_info "Configuration de Nginx..."
$SUDO tee /etc/nginx/sites-available/forensicinvestigator > /dev/null << 'EOF'
server {
    listen 80;
    server_name _;

    access_log /var/log/nginx/forensic-access.log;
    error_log /var/log/nginx/forensic-error.log;

    location / {
        proxy_pass http://127.0.0.1:8082;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
        proxy_read_timeout 300s;
        proxy_connect_timeout 75s;
    }

    location /api/hrm/ {
        proxy_pass http://127.0.0.1:8081/;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_read_timeout 300s;
    }

    location ~ ^/api/(analyze|chat|contradictions)/stream {
        proxy_pass http://127.0.0.1:8082;
        proxy_http_version 1.1;
        proxy_set_header Connection '';
        proxy_set_header Host $host;
        proxy_buffering off;
        proxy_cache off;
        chunked_transfer_encoding off;
        proxy_read_timeout 600s;
    }

    location /static/ {
        proxy_pass http://127.0.0.1:8082/static/;
        proxy_cache_valid 200 1d;
        expires 1d;
    }

    client_max_body_size 50M;
}
EOF

$SUDO ln -sf /etc/nginx/sites-available/forensicinvestigator /etc/nginx/sites-enabled/
$SUDO rm -f /etc/nginx/sites-enabled/default
$SUDO nginx -t

# Créer le script de gestion
log_info "Création du script de gestion..."
$SUDO tee /usr/local/bin/forensic > /dev/null << 'MGMT_EOF'
#!/bin/bash
APP_DIR="/opt/forensicinvestigator"
case "$1" in
    start)
        echo "Démarrage de ForensicInvestigator..."
        systemctl start forensicinvestigator-embedding
        systemctl start forensicinvestigator-hrm
        systemctl start forensicinvestigator
        systemctl start nginx
        echo "Services démarrés"
        echo ""
        echo "  📊 Application:     http://localhost:8082"
        echo "  🧠 HRM Server:      http://localhost:8081"
        echo "  🔍 Embedding:       http://localhost:8085"
        ;;
    stop)
        echo "Arrêt..."
        systemctl stop forensicinvestigator
        systemctl stop forensicinvestigator-hrm
        systemctl stop forensicinvestigator-embedding
        echo "Services arrêtés"
        ;;
    restart)
        echo "Redémarrage..."
        systemctl restart forensicinvestigator-embedding
        systemctl restart forensicinvestigator-hrm
        systemctl restart forensicinvestigator
        systemctl restart nginx
        echo "Services redémarrés"
        ;;
    status)
        echo "=== ForensicInvestigator Status ==="
        echo ""
        echo "--- Service Principal (8082) ---"
        systemctl status forensicinvestigator --no-pager -l | head -10
        echo ""
        echo "--- Service HRM (8081) ---"
        systemctl status forensicinvestigator-hrm --no-pager -l | head -10
        echo ""
        echo "--- Service Embedding (8085) ---"
        systemctl status forensicinvestigator-embedding --no-pager -l | head -10
        echo ""
        echo "--- Nginx ---"
        systemctl status nginx --no-pager -l | head -8
        ;;
    logs)
        tail -f ${APP_DIR}/logs/forensic.log
        ;;
    logs-hrm)
        tail -f ${APP_DIR}/logs/hrm.log
        ;;
    logs-embedding)
        tail -f ${APP_DIR}/logs/embedding.log
        ;;
    logs-all)
        tail -f ${APP_DIR}/logs/*.log
        ;;
    backup)
        BACKUP_FILE="/tmp/forensic-backup-$(date +%Y%m%d_%H%M%S).tar.gz"
        echo "Création de la sauvegarde..."
        tar -czf ${BACKUP_FILE} -C ${APP_DIR} data config
        echo "Sauvegarde créée: ${BACKUP_FILE}"
        ;;
    *)
        echo "Usage: forensic {start|stop|restart|status|logs|logs-hrm|logs-embedding|logs-all|backup}"
        exit 1
        ;;
esac
MGMT_EOF
$SUDO chmod +x /usr/local/bin/forensic

# Permissions
log_info "Configuration des permissions..."
$SUDO chown -R ${APP_USER}:${APP_GROUP} ${APP_DIR}
$SUDO chmod -R 755 ${APP_DIR}
$SUDO chmod 600 ${APP_DIR}/config/environment

# Créer les fichiers de log
$SUDO touch ${APP_DIR}/logs/forensic.log ${APP_DIR}/logs/hrm.log ${APP_DIR}/logs/embedding.log
$SUDO chown ${APP_USER}:${APP_GROUP} ${APP_DIR}/logs/*.log

# Activer et démarrer les services
log_info "Activation des services..."
$SUDO systemctl daemon-reload
$SUDO systemctl enable forensicinvestigator forensicinvestigator-hrm forensicinvestigator-embedding nginx

log_info "Démarrage des services..."
$SUDO systemctl start forensicinvestigator-embedding
sleep 2
$SUDO systemctl start forensicinvestigator-hrm
sleep 2
$SUDO systemctl start forensicinvestigator
$SUDO systemctl reload nginx

# Nettoyer
rm -rf /tmp/forensicinvestigator /tmp/install_package.tar.gz

# Vérification finale
sleep 3
echo ""
echo "=============================================="
echo -e "${GREEN}Installation terminée avec succès!${NC}"
echo "=============================================="
echo ""
echo "Services installés:"
echo "  📊 Application Go       - Port 8082"
echo "  🧠 HRM Server (Python)  - Port 8081"
echo "  🔍 Embedding (Model2vec) - Port 8085"
echo "  🤖 vLLM (distant)       - http://86.204.69.30:8001"
echo ""
echo "Commandes disponibles:"
echo "  sudo forensic start    - Démarrer"
echo "  sudo forensic stop     - Arrêter"
echo "  sudo forensic status   - Statut"
echo "  sudo forensic logs     - Logs"
echo ""

# Afficher le statut
$SUDO forensic status

REMOTE_INSTALL

# 5. Nettoyer localement
rm -rf "${TEMP_DIR}"
rm -f "${SOURCE_DIR}/${APP_NAME}_linux"

echo ""
log_success "Installation terminée!"
echo ""

# Extraire l'IP du serveur
SERVER_IP="${SERVER#*@}"
echo "L'application est accessible sur: http://${SERVER_IP}/"
echo ""
