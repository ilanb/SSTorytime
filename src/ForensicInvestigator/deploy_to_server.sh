#!/bin/bash
#
# ForensicInvestigator - Script de déploiement vers le serveur OVH
# Usage: ./deploy_to_server.sh
#

set -e

# Configuration
SERVER="ubuntu@51.75.240.95"
SSH_KEY="$HOME/.ssh/id_ed25519_ovh"
REMOTE_DIR="/opt/forensicinvestigator"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Couleurs
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[OK]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; exit 1; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }

echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  ForensicInvestigator - Déploiement   ${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Vérifier la clé SSH
if [ ! -f "$SSH_KEY" ]; then
    log_error "Clé SSH non trouvée: $SSH_KEY"
fi

# Test de connexion SSH
log_info "Test de connexion SSH..."
if ! ssh -i "$SSH_KEY" -o ConnectTimeout=10 -o BatchMode=yes "$SERVER" "echo ok" > /dev/null 2>&1; then
    log_error "Impossible de se connecter au serveur. Vérifiez votre connexion SSH."
fi
log_success "Connexion SSH OK"

# 1. Compilation pour Linux
log_info "Compilation pour Linux amd64..."
cd "$SCRIPT_DIR"
GOOS=linux GOARCH=amd64 go build -o ForensicInvestigator-linux .
log_success "Binaire compilé: ForensicInvestigator-linux"

# 2. Envoi du binaire
log_info "Envoi du binaire..."
scp -i "$SSH_KEY" ForensicInvestigator-linux "$SERVER:/tmp/"
log_success "Binaire envoyé"

# 3. Envoi des fichiers statiques
log_info "Envoi des fichiers statiques..."
scp -i "$SSH_KEY" -r static "$SERVER:/tmp/forensic_static"
log_success "Fichiers statiques envoyés"

# 4. Envoi des fichiers HRM (Python)
log_info "Envoi des fichiers HRM..."
scp -i "$SSH_KEY" -r hrm_server/*.py "$SERVER:/tmp/" 2>/dev/null || true
log_success "Fichiers HRM envoyés"

# 5. Déploiement sur le serveur
log_info "Déploiement sur le serveur..."
ssh -i "$SSH_KEY" "$SERVER" << 'REMOTE_SCRIPT'
set -e

APP_DIR="/opt/forensicinvestigator"

echo ""
echo "========================================"
echo "  [REMOTE] Arrêt de tous les services  "
echo "========================================"

echo "[1/3] Arrêt du service Embedding..."
sudo systemctl stop forensicinvestigator-embedding 2>/dev/null || true

echo "[2/3] Arrêt du service HRM..."
sudo systemctl stop forensicinvestigator-hrm 2>/dev/null || true

echo "[3/3] Arrêt du service Go..."
sudo systemctl stop forensicinvestigator 2>/dev/null || true

echo ""
echo "========================================"
echo "  [REMOTE] Mise à jour des fichiers    "
echo "========================================"

echo "[REMOTE] Mise à jour du binaire Go..."
sudo cp /tmp/ForensicInvestigator-linux ${APP_DIR}/bin/forensicinvestigator
sudo chmod +x ${APP_DIR}/bin/forensicinvestigator

echo "[REMOTE] Mise à jour des fichiers statiques..."
sudo rm -rf ${APP_DIR}/static/*
sudo cp -r /tmp/forensic_static/* ${APP_DIR}/static/

echo "[REMOTE] Mise à jour des fichiers HRM..."
for pyfile in /tmp/*.py; do
    if [ -f "$pyfile" ]; then
        sudo cp "$pyfile" ${APP_DIR}/hrm_server/
    fi
done

echo "[REMOTE] Correction des permissions..."
sudo chown -R forensic:forensic ${APP_DIR}
sudo chmod -R 755 ${APP_DIR}/static
sudo chmod 600 ${APP_DIR}/config/environment 2>/dev/null || true

echo "[REMOTE] Nettoyage des fichiers temporaires..."
rm -rf /tmp/ForensicInvestigator-linux /tmp/forensic_static /tmp/*.py 2>/dev/null || true

echo ""
echo "========================================"
echo "  [REMOTE] Redémarrage des services    "
echo "========================================"

echo "[1/3] Démarrage du service Embedding (Model2vec)..."
sudo systemctl start forensicinvestigator-embedding 2>/dev/null || echo "  → Service Embedding non configuré"
sleep 2

echo "[2/3] Démarrage du service HRM..."
sudo systemctl start forensicinvestigator-hrm 2>/dev/null || echo "  → Service HRM non configuré"
sleep 2

echo "[3/3] Démarrage du service Go..."
sudo systemctl start forensicinvestigator
sleep 2

echo ""
echo "========================================"
echo "  [REMOTE] Vérification des services   "
echo "========================================"

# Vérifier le service Go
if sudo systemctl is-active --quiet forensicinvestigator; then
    echo "  ✓ ForensicInvestigator (Go)    : actif sur port 8082"
else
    echo "  ✗ ForensicInvestigator (Go)    : ERREUR"
    sudo journalctl -u forensicinvestigator --no-pager -n 5
fi

# Vérifier le service HRM
if sudo systemctl is-active --quiet forensicinvestigator-hrm 2>/dev/null; then
    echo "  ✓ HRM Server (Python)          : actif sur port 8081"
else
    echo "  - HRM Server (Python)          : non configuré ou inactif"
fi

# Vérifier le service Embedding
if sudo systemctl is-active --quiet forensicinvestigator-embedding 2>/dev/null; then
    echo "  ✓ Embedding Service (Model2vec): actif sur port 8085"
else
    echo "  - Embedding Service (Model2vec): non configuré ou inactif"
fi

echo ""
echo "[REMOTE] Test de l'application Go..."
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8082/ 2>/dev/null || echo "000")
if [ "$HTTP_CODE" = "200" ]; then
    echo "  ✓ Application répond (HTTP $HTTP_CODE)"
else
    echo "  ⚠ Application répond avec HTTP $HTTP_CODE"
fi

REMOTE_SCRIPT

# 6. Nettoyage local
log_info "Nettoyage local..."
rm -f "$SCRIPT_DIR/ForensicInvestigator-linux"
log_success "Fichier temporaire supprimé"

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  Déploiement terminé avec succès !    ${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo -e "  📊 Application:     ${BLUE}http://51.75.240.95:8082${NC}"
echo -e "  🧠 HRM Server:      ${BLUE}http://51.75.240.95:8081${NC}"
echo -e "  🔍 Embedding:       ${BLUE}http://51.75.240.95:8085${NC}"
echo ""
