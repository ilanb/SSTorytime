#!/bin/bash
# ForensicInvestigator - Script de démarrage des services
# Usage: ./start_services.sh

set -e

# Activer le mode vLLM pour HRM (raisonnement LLM avancé)
export USE_SAPIENT=true

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PID_DIR="$SCRIPT_DIR/.pids"
LOG_DIR="$SCRIPT_DIR/logs"

# Couleurs
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

# Créer les répertoires nécessaires
mkdir -p "$PID_DIR" "$LOG_DIR"

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  ForensicInvestigator - Démarrage     ${NC}"
echo -e "${GREEN}========================================${NC}"

# Fonction pour vérifier si un port est utilisé
check_port() {
    lsof -ti:$1 > /dev/null 2>&1
}

# Fonction pour tuer un processus sur un port
kill_port() {
    if check_port $1; then
        echo -e "${YELLOW}Port $1 occupé, arrêt du processus...${NC}"
        lsof -ti:$1 | xargs kill -9 2>/dev/null || true
        sleep 1
    fi
}

# Le service Model2vec local (port 8085) n'est plus démarré: la recherche
# sémantique utilise désormais le service d'embeddings du SPARK
# (multilingual-e5-base, port 8002). Le répertoire embedding_service/ est conservé
# comme solution de repli hors-ligne mais n'est plus sollicité par l'application.

# 1. Service HRM (Python) - Port 8081
echo -e "\n${GREEN}[1/2] Démarrage du service HRM...${NC}"
kill_port 8081
cd "$SCRIPT_DIR/hrm_server"
python main.py > "$LOG_DIR/hrm.log" 2>&1 &
HRM_PID=$!
echo $HRM_PID > "$PID_DIR/hrm.pid"
echo -e "  → PID: $HRM_PID"
echo -e "  → Port: 8081"
echo -e "  → Log: $LOG_DIR/hrm.log"

# Attendre que le service soit prêt
sleep 2

# 2. Application Go principale - Port 8082
echo -e "\n${GREEN}[2/2] Démarrage de l'application Go...${NC}"
kill_port 8082
cd "$SCRIPT_DIR"
# Compiler si nécessaire
if [ ! -f "./forensicinvestigator" ] || [ "main.go" -nt "./forensicinvestigator" ]; then
    echo -e "  → Compilation en cours..."
    go build -o forensicinvestigator .
fi
./forensicinvestigator > "$LOG_DIR/go.log" 2>&1 &
GO_PID=$!
echo $GO_PID > "$PID_DIR/go.pid"
echo -e "  → PID: $GO_PID"
echo -e "  → Port: 8082"
echo -e "  → Log: $LOG_DIR/go.log"

# Attendre que le service soit prêt
sleep 2

echo -e "\n${GREEN}========================================${NC}"
echo -e "${GREEN}  Services démarrés avec succès !       ${NC}"
echo -e "${GREEN}========================================${NC}"
echo -e "\n  📊 Application:     http://localhost:8082"
echo -e "  🧠 HRM Server:      http://localhost:8081 (USE_SAPIENT=true)"
echo -e "  🤖 LLM (distant):   http://86.204.69.30:8001/v1 (Qwen3.8-27B-FP8 / vLLM)"
echo -e "  🔍 Embeddings:      http://86.204.69.30:8002/v1 (multilingual-e5-base, 768d)"
echo -e "\n  Pour arrêter: ./stop_services.sh"
echo -e "  Pour les logs: tail -f $LOG_DIR/*.log"
