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

# Le service Model2vec local (8085) est obsolète: on l'arrête s'il subsiste.
sudo systemctl stop forensicinvestigator-embedding 2>/dev/null || true

echo "[1/2] Arrêt du service HRM..."
sudo systemctl stop forensicinvestigator-hrm 2>/dev/null || true

echo "[2/2] Arrêt du service Go..."
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

echo "[REMOTE] Vérification des dépendances Python du HRM..."
# Le moteur Sapient importe requests: sans lui, hrm_sapient est inimportable et le
# serveur retombe sur le moteur de règles sans jamais solliciter le LLM.
# torch et huggingface_hub restent optionnels (checkpoint HRM, inutile ici).
HRM_PY="${APP_DIR}/hrm_server/venv/bin/python"
if [ -x "$HRM_PY" ]; then
    if ! sudo "$HRM_PY" -c "import requests" 2>/dev/null; then
        echo "  → Installation de requests..."
        sudo "${APP_DIR}/hrm_server/venv/bin/pip" install -q requests
    fi
    sudo "$HRM_PY" -c "import requests" 2>/dev/null \
        && echo "  ✓ requests disponible" \
        || echo "  ✗ requests INDISPONIBLE - le HRM restera en mode règles"
else
    echo "  ⚠ venv HRM introuvable: $HRM_PY"
fi

echo "[REMOTE] Configuration de l'environnement..."
sudo mkdir -p ${APP_DIR}/config
ENV_FILE="${APP_DIR}/config/environment"

# Les deux unités systemd (forensicinvestigator et -hrm) chargent config/environment.
# C'est donc le seul fichier qui compte: un config/hrm.env séparé n'était lu par
# personne, ce qui laissait USE_SAPIENT à sa valeur par défaut (false) et faisait
# tourner le HRM en mode règles sans jamais solliciter le LLM.
sudo touch "$ENV_FILE"
sudo cp "$ENV_FILE" "${ENV_FILE}.bak.$(date +%Y%m%d_%H%M%S)"

# set_env_var CLE VALEUR — remplace la ligne existante ou l'ajoute.
set_env_var() {
    local key="$1" value="$2"
    if sudo grep -qE "^${key}=" "$ENV_FILE"; then
        sudo sed -i "s|^${key}=.*|${key}=${value}|" "$ENV_FILE"
    else
        echo "${key}=${value}" | sudo tee -a "$ENV_FILE" > /dev/null
    fi
}

# Moteur HRM: raisonnement adossé au LLM plutôt que le moteur de règles.
set_env_var USE_SAPIENT true

# vLLM sur SPARK GB10 (serveur partagé, configuration prise telle quelle):
# Qwen/Qwen3.8-27B-FP8 servi sous l'alias "Qwen3.5-9B", identifiant attendu par l'API.
set_env_var VLLM_URL http://86.204.69.30:8001/v1
set_env_var VLLM_MODEL Qwen3.5-9B

# Embeddings sur le SPARK (multilingual-e5-base, 768 dimensions).
set_env_var EMBEDDING_BASE_URL http://86.204.69.30:8002/v1
set_env_var EMBEDDING_MODEL multilingual-e5-base

# Le service Model2vec local (8085) n'existe plus.
sudo sed -i '/^EMBEDDING_PORT=/d' "$ENV_FILE"

# Fichier orphelin d'une version antérieure: aucune unité systemd ne le charge.
sudo rm -f ${APP_DIR}/config/hrm.env

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

# Le service Model2vec local (8085) n'est plus utilisé: les embeddings viennent
# du SPARK (port 8002). On s'assure qu'une instance héritée ne tourne plus.
sudo systemctl stop forensicinvestigator-embedding 2>/dev/null || true

echo "[1/2] Démarrage du service HRM..."
sudo systemctl start forensicinvestigator-hrm 2>/dev/null || echo "  → Service HRM non configuré"
sleep 2

echo "[2/2] Démarrage du service Go..."
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

# Vérifier le service d'embeddings distant (SPARK)
if curl -sf -m 5 "${EMBEDDING_BASE_URL:-http://86.204.69.30:8002/v1}/models" > /dev/null 2>&1; then
    echo "  ✓ Embeddings (SPARK)           : joignables (multilingual-e5-base)"
else
    echo "  - Embeddings (SPARK)           : injoignables, repli sur BM25 seul"
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
echo -e "  🔍 Embeddings:      ${BLUE}http://86.204.69.30:8002/v1${NC} (multilingual-e5-base)"
echo ""
