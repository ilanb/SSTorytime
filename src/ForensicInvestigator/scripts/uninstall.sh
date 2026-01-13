#!/bin/bash
#
# ForensicInvestigator - Script de désinstallation
# Usage: sudo ./uninstall.sh [--purge]
#
# Options:
#   --purge   Supprime également les données et configurations
#

set -e

# Couleurs
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

APP_DIR="/opt/forensicinvestigator"
APP_USER="forensic"

log_info() { echo -e "${YELLOW}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[OK]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Vérifier root
if [[ $EUID -ne 0 ]]; then
    log_error "Ce script doit être exécuté en tant que root (sudo)"
    exit 1
fi

PURGE=false
if [ "$1" == "--purge" ]; then
    PURGE=true
fi

echo ""
echo "=============================================="
echo "  ForensicInvestigator - Désinstallation"
echo "=============================================="
echo ""

if [ "$PURGE" = true ]; then
    log_info "Mode PURGE: Toutes les données seront supprimées!"
    echo ""
    read -p "Êtes-vous sûr de vouloir tout supprimer? (oui/non): " confirm
    if [ "$confirm" != "oui" ]; then
        echo "Annulé."
        exit 0
    fi
else
    log_info "Les données seront conservées dans ${APP_DIR}/data"
fi

echo ""

# 1. Arrêter les services
log_info "Arrêt des services..."
systemctl stop forensicinvestigator 2>/dev/null || true
systemctl stop forensicinvestigator-hrm 2>/dev/null || true
systemctl stop forensicinvestigator-embedding 2>/dev/null || true

# 2. Désactiver les services
log_info "Désactivation des services..."
systemctl disable forensicinvestigator 2>/dev/null || true
systemctl disable forensicinvestigator-hrm 2>/dev/null || true
systemctl disable forensicinvestigator-embedding 2>/dev/null || true

# 3. Supprimer les fichiers systemd
log_info "Suppression des services systemd..."
rm -f /etc/systemd/system/forensicinvestigator.service
rm -f /etc/systemd/system/forensicinvestigator-hrm.service
rm -f /etc/systemd/system/forensicinvestigator-embedding.service
systemctl daemon-reload

# 4. Supprimer la configuration Nginx
log_info "Suppression de la configuration Nginx..."
rm -f /etc/nginx/sites-enabled/forensicinvestigator
rm -f /etc/nginx/sites-available/forensicinvestigator
systemctl reload nginx 2>/dev/null || true

# 5. Supprimer le script de gestion
log_info "Suppression du script de gestion..."
rm -f /usr/local/bin/forensic

# 6. Sauvegarder les données (si pas purge)
if [ "$PURGE" = false ] && [ -d "${APP_DIR}/data" ]; then
    BACKUP_FILE="/tmp/forensic-data-backup-$(date +%Y%m%d_%H%M%S).tar.gz"
    log_info "Sauvegarde des données dans ${BACKUP_FILE}..."
    tar -czf "${BACKUP_FILE}" -C "${APP_DIR}" data config 2>/dev/null || true
    log_success "Données sauvegardées"
fi

# 7. Supprimer les fichiers
if [ "$PURGE" = true ]; then
    log_info "Suppression complète de ${APP_DIR}..."
    rm -rf ${APP_DIR}
else
    log_info "Suppression des binaires (données conservées)..."
    rm -rf ${APP_DIR}/bin
    rm -rf ${APP_DIR}/static
    rm -rf ${APP_DIR}/logs
    rm -rf ${APP_DIR}/hrm_server
    rm -rf ${APP_DIR}/embedding_service
fi

# 8. Supprimer l'utilisateur (optionnel)
if [ "$PURGE" = true ]; then
    log_info "Suppression de l'utilisateur ${APP_USER}..."
    userdel -r ${APP_USER} 2>/dev/null || true
fi

echo ""
log_success "Désinstallation terminée!"
echo ""

if [ "$PURGE" = false ]; then
    echo "Les données ont été conservées dans: ${APP_DIR}/data"
    echo "Sauvegarde créée: ${BACKUP_FILE}"
    echo ""
    echo "Pour supprimer complètement, relancez avec: sudo $0 --purge"
fi

echo ""
