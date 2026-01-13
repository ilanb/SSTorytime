#!/bin/bash
# ForensicInvestigator - Script d'arrêt des services
# Usage: ./stop_services.sh

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PID_DIR="$SCRIPT_DIR/.pids"

# Couleurs
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${RED}========================================${NC}"
echo -e "${RED}  ForensicInvestigator - Arrêt         ${NC}"
echo -e "${RED}========================================${NC}"

# Fonction pour arrêter un service
stop_service() {
    local name=$1
    local pid_file="$PID_DIR/$2.pid"
    local port=$3

    echo -e "\n${YELLOW}Arrêt de $name...${NC}"

    # Arrêter via PID file si disponible
    if [ -f "$pid_file" ]; then
        pid=$(cat "$pid_file")
        if ps -p $pid > /dev/null 2>&1; then
            kill $pid 2>/dev/null || true
            echo -e "  → Processus $pid arrêté"
        fi
        rm -f "$pid_file"
    fi

    # Arrêter via port au cas où
    if lsof -ti:$port > /dev/null 2>&1; then
        lsof -ti:$port | xargs kill -9 2>/dev/null || true
        echo -e "  → Port $port libéré"
    else
        echo -e "  → Port $port déjà libre"
    fi
}

# Arrêter les services
stop_service "Application Go" "go" 8082
stop_service "Service HRM" "hrm" 8081
stop_service "Service Embedding" "embedding" 8085

echo -e "\n${GREEN}========================================${NC}"
echo -e "${GREEN}  Tous les services sont arrêtés       ${NC}"
echo -e "${GREEN}========================================${NC}"
