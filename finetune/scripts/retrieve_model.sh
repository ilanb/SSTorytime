#!/bin/bash
# =============================================================================
# Script pour récupérer le modèle fine-tuné depuis le serveur spark-84a8
# =============================================================================

set -e

# Configuration
SSH_KEY="/Users/ilan/Library/Application Support/NVIDIA/Sync/config/nvsync.key"
SERVER="infostrates@10.0.0.92"
CONTAINER="unsloth-training"
REMOTE_MODEL_DIR="/workspace/n4l-finetune/models/n4l-qwen-unsloth"
LOCAL_DIR="./models"

# Couleurs
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo "========================================"
echo "  Récupération du modèle N4L fine-tuné"
echo "========================================"

# 1. Vérifier le statut du training
echo -e "\n${YELLOW}Vérification du statut du training...${NC}"
TRAINING_STATUS=$(ssh -o 'IdentitiesOnly=yes' -i "$SSH_KEY" "$SERVER" \
    "docker exec $CONTAINER ps aux | grep 'train_unsloth.py' | grep -v grep" 2>/dev/null || echo "")

if [ -n "$TRAINING_STATUS" ]; then
    echo -e "${RED}⚠ Le training est encore en cours !${NC}"
    echo "Progression actuelle :"
    ssh -o 'IdentitiesOnly=yes' -i "$SSH_KEY" "$SERVER" \
        "docker exec $CONTAINER tail -5 /workspace/n4l-finetune/training.log 2>/dev/null"
    echo ""
    read -p "Voulez-vous quand même continuer ? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
else
    echo -e "${GREEN}✓ Training terminé${NC}"
fi

# 2. Vérifier les fichiers disponibles
echo -e "\n${YELLOW}Vérification des modèles disponibles...${NC}"
ssh -o 'IdentitiesOnly=yes' -i "$SSH_KEY" "$SERVER" \
    "docker exec $CONTAINER ls -la $REMOTE_MODEL_DIR/ 2>/dev/null" || {
    echo -e "${RED}Erreur: Répertoire de modèles non trouvé${NC}"
    exit 1
}

# 3. Créer le répertoire local
mkdir -p "$LOCAL_DIR"

# 4. Menu de sélection
echo -e "\n${YELLOW}Quel format voulez-vous récupérer ?${NC}"
echo "1) LoRA adapter seulement (~300MB) - Pour continuer le training ou fusionner"
echo "2) Modèle fusionné complet (~15GB) - Prêt pour l'inférence"
echo "3) Format GGUF (~8GB) - Pour Ollama/llama.cpp"
echo "4) Tout récupérer"
echo "5) Vérifier les logs de training"
read -p "Choix [1-5]: " choice

case $choice in
    1)
        echo -e "\n${YELLOW}Récupération de l'adaptateur LoRA...${NC}"
        # Copier depuis le container vers le serveur
        ssh -o 'IdentitiesOnly=yes' -i "$SSH_KEY" "$SERVER" \
            "docker cp $CONTAINER:$REMOTE_MODEL_DIR/lora_adapter /tmp/lora_adapter"
        # Puis du serveur vers local
        scp -o 'IdentitiesOnly=yes' -i "$SSH_KEY" -r \
            "$SERVER:/tmp/lora_adapter" "$LOCAL_DIR/"
        echo -e "${GREEN}✓ Adaptateur LoRA récupéré dans $LOCAL_DIR/lora_adapter${NC}"
        ;;
    2)
        echo -e "\n${YELLOW}Récupération du modèle fusionné (peut prendre du temps)...${NC}"
        ssh -o 'IdentitiesOnly=yes' -i "$SSH_KEY" "$SERVER" \
            "docker cp $CONTAINER:$REMOTE_MODEL_DIR/merged_model /tmp/merged_model"
        scp -o 'IdentitiesOnly=yes' -i "$SSH_KEY" -r \
            "$SERVER:/tmp/merged_model" "$LOCAL_DIR/"
        echo -e "${GREEN}✓ Modèle fusionné récupéré dans $LOCAL_DIR/merged_model${NC}"
        ;;
    3)
        echo -e "\n${YELLOW}Récupération du modèle GGUF...${NC}"
        ssh -o 'IdentitiesOnly=yes' -i "$SSH_KEY" "$SERVER" \
            "docker cp $CONTAINER:$REMOTE_MODEL_DIR/gguf /tmp/gguf"
        scp -o 'IdentitiesOnly=yes' -i "$SSH_KEY" -r \
            "$SERVER:/tmp/gguf" "$LOCAL_DIR/"
        echo -e "${GREEN}✓ Modèle GGUF récupéré dans $LOCAL_DIR/gguf${NC}"

        echo -e "\n${YELLOW}Pour utiliser avec Ollama:${NC}"
        echo "  ollama create n4l-qwen -f $LOCAL_DIR/gguf/Modelfile"
        ;;
    4)
        echo -e "\n${YELLOW}Récupération de tous les modèles...${NC}"
        ssh -o 'IdentitiesOnly=yes' -i "$SSH_KEY" "$SERVER" \
            "docker cp $CONTAINER:$REMOTE_MODEL_DIR /tmp/n4l-qwen-unsloth"
        scp -o 'IdentitiesOnly=yes' -i "$SSH_KEY" -r \
            "$SERVER:/tmp/n4l-qwen-unsloth" "$LOCAL_DIR/"
        echo -e "${GREEN}✓ Tous les modèles récupérés dans $LOCAL_DIR/n4l-qwen-unsloth${NC}"
        ;;
    5)
        echo -e "\n${YELLOW}Logs de training :${NC}"
        ssh -o 'IdentitiesOnly=yes' -i "$SSH_KEY" "$SERVER" \
            "docker exec $CONTAINER cat /workspace/n4l-finetune/training.log"
        exit 0
        ;;
    *)
        echo "Choix invalide"
        exit 1
        ;;
esac

echo -e "\n${GREEN}========================================"
echo "  Récupération terminée !"
echo "========================================${NC}"
echo ""
echo "Fichiers disponibles :"
ls -la "$LOCAL_DIR/"
