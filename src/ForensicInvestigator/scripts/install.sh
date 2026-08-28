#!/bin/bash
#
# ForensicInvestigator - Script d'installation pour Ubuntu Server
# Usage: sudo ./install.sh
#

set -e

# Couleurs pour l'affichage
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
APP_NAME="forensicinvestigator"
APP_DIR="/opt/forensicinvestigator"
APP_USER="forensic"
APP_GROUP="forensic"
GO_VERSION="1.21.5"
PYTHON_VERSION="3.11"

# URLs des serveurs (à modifier selon votre configuration)
# vLLM sur SPARK GB10 (serveur partagé) : Qwen/Qwen3.8-27B-FP8 (128k ctx)
# servi sous l'alias "Qwen3.5-9B", identifiant attendu par l'API.
VLLM_URL="${VLLM_URL:-http://86.204.69.30:8001}"
VLLM_MODEL="${VLLM_MODEL:-Qwen3.5-9B}"
OLLAMA_URL="${OLLAMA_URL:-http://localhost:11434}"

# Embeddings sur SPARK GB10 : multilingual-e5-base (768 dimensions)
EMBEDDING_BASE_URL="${EMBEDDING_BASE_URL:-http://86.204.69.30:8002/v1}"
EMBEDDING_MODEL="${EMBEDDING_MODEL:-multilingual-e5-base}"

# Ports des services
FORENSIC_PORT=8082
HRM_PORT=8081

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[OK]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

check_root() {
    if [[ $EUID -ne 0 ]]; then
        log_error "Ce script doit être exécuté en tant que root (sudo)"
        exit 1
    fi
}

check_ubuntu() {
    if ! grep -q "Ubuntu" /etc/os-release 2>/dev/null; then
        log_warning "Ce script est optimisé pour Ubuntu. Continuez à vos risques..."
    fi
}

install_dependencies() {
    log_info "Installation des dépendances système..."

    apt-get update
    apt-get install -y \
        build-essential \
        git \
        curl \
        wget \
        unzip \
        python3 \
        python3-pip \
        python3-venv \
        nginx \
        certbot \
        python3-certbot-nginx \
        jq

    log_success "Dépendances système installées"
}

install_go() {
    log_info "Installation de Go ${GO_VERSION}..."

    if command -v go &> /dev/null; then
        CURRENT_GO=$(go version | awk '{print $3}' | sed 's/go//')
        log_info "Go déjà installé (version $CURRENT_GO)"
        return
    fi

    cd /tmp
    wget -q "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz"
    rm -rf /usr/local/go
    tar -C /usr/local -xzf "go${GO_VERSION}.linux-amd64.tar.gz"
    rm "go${GO_VERSION}.linux-amd64.tar.gz"

    # Ajouter Go au PATH système
    if ! grep -q "/usr/local/go/bin" /etc/profile; then
        echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile
    fi

    export PATH=$PATH:/usr/local/go/bin

    log_success "Go ${GO_VERSION} installé"
}

create_user() {
    log_info "Création de l'utilisateur ${APP_USER}..."

    if id "$APP_USER" &>/dev/null; then
        log_info "L'utilisateur ${APP_USER} existe déjà"
    else
        useradd -r -s /bin/false -m -d /home/${APP_USER} ${APP_USER}
        log_success "Utilisateur ${APP_USER} créé"
    fi
}

create_directories() {
    log_info "Création des répertoires..."

    mkdir -p ${APP_DIR}/{bin,config,data,logs,static}
    mkdir -p ${APP_DIR}/data/notebooks
    mkdir -p ${APP_DIR}/hrm_server
    mkdir -p ${APP_DIR}/embedding_service

    log_success "Répertoires créés"
}

copy_application() {
    log_info "Copie des fichiers de l'application..."

    SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    SOURCE_DIR="$(dirname "$SCRIPT_DIR")"

    # Copier les fichiers statiques
    cp -r ${SOURCE_DIR}/static/* ${APP_DIR}/static/

    # Copier la configuration
    cp -r ${SOURCE_DIR}/config/* ${APP_DIR}/config/ 2>/dev/null || true

    # Copier le serveur HRM
    cp -r ${SOURCE_DIR}/hrm_server/* ${APP_DIR}/hrm_server/

    # Copier le service d'embedding
    cp -r ${SOURCE_DIR}/embedding_service/* ${APP_DIR}/embedding_service/

    # Copier les données de démonstration
    cp -r ${SOURCE_DIR}/data/* ${APP_DIR}/data/ 2>/dev/null || true

    log_success "Fichiers copiés"
}

build_application() {
    log_info "Compilation de l'application Go..."

    SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    SOURCE_DIR="$(dirname "$SCRIPT_DIR")"

    cd ${SOURCE_DIR}
    export PATH=$PATH:/usr/local/go/bin

    # Compiler l'application
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ${APP_DIR}/bin/${APP_NAME} .

    log_success "Application compilée"
}

setup_hrm_server() {
    log_info "Configuration du serveur HRM (Python)..."

    cd ${APP_DIR}/hrm_server

    # Créer l'environnement virtuel
    python3 -m venv venv
    source venv/bin/activate

    # Installer les dépendances
    pip install --upgrade pip
    pip install fastapi uvicorn pydantic httpx aiohttp

    deactivate

    log_success "Serveur HRM configuré"
}

setup_embedding_service() {
    log_info "Configuration du service Embedding (Model2vec)..."

    cd ${APP_DIR}/embedding_service

    # Créer l'environnement virtuel
    python3 -m venv venv
    source venv/bin/activate

    # Installer les dépendances
    pip install --upgrade pip
    if [ -f requirements.txt ]; then
        pip install -r requirements.txt
    else
        pip install fastapi uvicorn model2vec numpy pydantic
    fi

    deactivate

    log_success "Service Embedding configuré"
}

create_env_file() {
    log_info "Création du fichier de configuration..."

    cat > ${APP_DIR}/config/environment << EOF
# ForensicInvestigator - Configuration
# Modifiez ces valeurs selon votre environnement

# Serveur LLM distant (llama.cpp, modèle principal)
VLLM_URL=${VLLM_URL}
VLLM_MODEL=${VLLM_MODEL}

# Serveur Ollama (conversion N4L)
OLLAMA_URL=${OLLAMA_URL}
OLLAMA_MODEL=n4l-qwen:latest

# Service d'embeddings (recherche hybride)
EMBEDDING_BASE_URL=${EMBEDDING_BASE_URL}
EMBEDDING_MODEL=${EMBEDDING_MODEL}

# Ports
FORENSIC_PORT=8082
HRM_PORT=8081

# Logging
LOG_LEVEL=info
EOF

    log_success "Fichier de configuration créé"
}

create_systemd_services() {
    log_info "Création des services systemd..."

    # Service principal ForensicInvestigator
    cat > /etc/systemd/system/forensicinvestigator.service << EOF
[Unit]
Description=ForensicInvestigator - Application d'investigation forensique
After=network.target
Wants=forensicinvestigator-hrm.service

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

# Sécurité
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=${APP_DIR}/data ${APP_DIR}/logs
PrivateTmp=true

[Install]
WantedBy=multi-user.target
EOF

    # Service HRM (Python)
    cat > /etc/systemd/system/forensicinvestigator-hrm.service << EOF
[Unit]
Description=ForensicInvestigator HRM Server - Hypothetical Reasoning Model
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

# Sécurité
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=${APP_DIR}/logs
PrivateTmp=true

[Install]
WantedBy=multi-user.target
EOF

    # Service Embedding (Model2vec)
    cat > /etc/systemd/system/forensicinvestigator-embedding.service << EOF
[Unit]
Description=ForensicInvestigator Embedding Server - Model2vec Semantic Search
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

# Sécurité
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=${APP_DIR}/logs
PrivateTmp=true

[Install]
WantedBy=multi-user.target
EOF

    systemctl daemon-reload

    log_success "Services systemd créés"
}

create_nginx_config() {
    log_info "Configuration de Nginx..."

    cat > /etc/nginx/sites-available/forensicinvestigator << 'EOF'
server {
    listen 80;
    server_name _;  # Remplacez par votre domaine

    # Logs
    access_log /var/log/nginx/forensic-access.log;
    error_log /var/log/nginx/forensic-error.log;

    # Application principale
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

    # API HRM
    location /api/hrm/ {
        proxy_pass http://127.0.0.1:8081/;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_read_timeout 300s;
    }

    # Streaming SSE
    location ~ ^/api/(analyze|chat|contradictions)/stream {
        proxy_pass http://127.0.0.1:8082;
        proxy_http_version 1.1;
        proxy_set_header Connection '';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_buffering off;
        proxy_cache off;
        chunked_transfer_encoding off;
        proxy_read_timeout 600s;
    }

    # Fichiers statiques (cache)
    location /static/ {
        proxy_pass http://127.0.0.1:8082/static/;
        proxy_cache_valid 200 1d;
        expires 1d;
        add_header Cache-Control "public, immutable";
    }

    # Limite de taille des requêtes
    client_max_body_size 50M;
}
EOF

    # Activer le site
    ln -sf /etc/nginx/sites-available/forensicinvestigator /etc/nginx/sites-enabled/

    # Désactiver le site par défaut
    rm -f /etc/nginx/sites-enabled/default

    # Tester la configuration
    nginx -t

    log_success "Nginx configuré"
}

set_permissions() {
    log_info "Configuration des permissions..."

    chown -R ${APP_USER}:${APP_GROUP} ${APP_DIR}
    chmod -R 755 ${APP_DIR}
    chmod 600 ${APP_DIR}/config/environment
    chmod +x ${APP_DIR}/bin/${APP_NAME}

    log_success "Permissions configurées"
}

create_management_script() {
    log_info "Création du script de gestion..."

    cat > /usr/local/bin/forensic << 'EOF'
#!/bin/bash
#
# ForensicInvestigator - Script de gestion
#

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
        echo "  🤖 vLLM (distant):  Voir config/environment"
        ;;
    stop)
        echo "Arrêt de ForensicInvestigator..."
        systemctl stop forensicinvestigator
        systemctl stop forensicinvestigator-hrm
        systemctl stop forensicinvestigator-embedding
        echo "Services arrêtés"
        ;;
    restart)
        echo "Redémarrage de ForensicInvestigator..."
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
        echo "=== Logs ForensicInvestigator ==="
        tail -f ${APP_DIR}/logs/forensic.log
        ;;
    logs-hrm)
        echo "=== Logs HRM ==="
        tail -f ${APP_DIR}/logs/hrm.log
        ;;
    logs-embedding)
        echo "=== Logs Embedding ==="
        tail -f ${APP_DIR}/logs/embedding.log
        ;;
    logs-all)
        echo "=== Tous les logs ==="
        tail -f ${APP_DIR}/logs/*.log
        ;;
    update)
        echo "Mise à jour de ForensicInvestigator..."
        cd ${APP_DIR}
        systemctl stop forensicinvestigator forensicinvestigator-hrm forensicinvestigator-embedding
        # Sauvegarde
        cp -r data data.backup.$(date +%Y%m%d_%H%M%S)
        echo "Recompilez et redéployez l'application, puis lancez: forensic start"
        ;;
    backup)
        BACKUP_FILE="/tmp/forensic-backup-$(date +%Y%m%d_%H%M%S).tar.gz"
        echo "Création de la sauvegarde..."
        tar -czf ${BACKUP_FILE} -C ${APP_DIR} data config
        echo "Sauvegarde créée: ${BACKUP_FILE}"
        ;;
    *)
        echo "Usage: forensic {start|stop|restart|status|logs|logs-hrm|logs-embedding|logs-all|update|backup}"
        exit 1
        ;;
esac
EOF

    chmod +x /usr/local/bin/forensic

    log_success "Script de gestion créé (/usr/local/bin/forensic)"
}

enable_services() {
    log_info "Activation des services au démarrage..."

    systemctl enable forensicinvestigator
    systemctl enable forensicinvestigator-hrm
    systemctl enable forensicinvestigator-embedding
    systemctl enable nginx

    log_success "Services activés"
}

configure_firewall() {
    log_info "Configuration du pare-feu..."

    if command -v ufw &> /dev/null; then
        ufw allow 80/tcp
        ufw allow 443/tcp
        ufw allow 22/tcp
        log_success "Pare-feu configuré (UFW)"
    else
        log_warning "UFW non installé, configurez manuellement le pare-feu"
    fi
}

print_summary() {
    echo ""
    echo "=============================================="
    echo -e "${GREEN}Installation terminée avec succès!${NC}"
    echo "=============================================="
    echo ""
    echo "Fichiers installés dans: ${APP_DIR}"
    echo ""
    echo "Services installés:"
    echo "  📊 Application Go      - Port 8082"
    echo "  🧠 HRM Server (Python) - Port 8081"
    echo "  🔍 Embedding (Model2vec) - Port 8085"
    echo "  🤖 vLLM (distant)      - ${VLLM_URL}"
    echo ""
    echo "Commandes disponibles:"
    echo "  forensic start         - Démarrer tous les services"
    echo "  forensic stop          - Arrêter tous les services"
    echo "  forensic restart       - Redémarrer tous les services"
    echo "  forensic status        - Voir le statut de tous les services"
    echo "  forensic logs          - Logs application principale"
    echo "  forensic logs-hrm      - Logs serveur HRM"
    echo "  forensic logs-embedding - Logs service embedding"
    echo "  forensic logs-all      - Tous les logs"
    echo "  forensic backup        - Créer une sauvegarde"
    echo ""
    echo "Configuration:"
    echo "  ${APP_DIR}/config/environment"
    echo ""
    echo "Pour démarrer l'application:"
    echo "  sudo forensic start"
    echo ""
    echo "L'application sera accessible sur:"
    echo "  http://votre-serveur/"
    echo ""
    echo "N'oubliez pas de:"
    echo "  1. Vérifier la configuration vLLM dans ${APP_DIR}/config/environment"
    echo "  2. Configurer le domaine dans /etc/nginx/sites-available/forensicinvestigator"
    echo "  3. Optionnel: Configurer HTTPS avec: sudo certbot --nginx"
    echo ""
}

# Main
main() {
    echo ""
    echo "=============================================="
    echo "  ForensicInvestigator - Installation Ubuntu"
    echo "=============================================="
    echo ""

    check_root
    check_ubuntu

    install_dependencies
    install_go
    create_user
    create_directories
    copy_application
    build_application
    setup_hrm_server
    setup_embedding_service
    create_env_file
    create_systemd_services
    create_nginx_config
    set_permissions
    create_management_script
    enable_services
    configure_firewall

    print_summary
}

main "$@"
