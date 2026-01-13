#!/bin/bash
# Script de déploiement et entraînement sur serveur GPU distant
# Usage: ./deploy_to_gpu.sh [--run]

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
PARENT_DIR="$(dirname "$PROJECT_DIR")"

# Configuration serveur
GPU_SERVER="ubuntu@51.91.166.26"
SSH_KEY="~/.ssh/infostratesIA"
SSH_OPTS="-o IdentitiesOnly=yes -i $SSH_KEY"
REMOTE_DIR="~/n4l-finetune"

# Options
RUN_TRAINING=false
SYNC_ONLY=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --run)
            RUN_TRAINING=true
            shift
            ;;
        --sync)
            SYNC_ONLY=true
            shift
            ;;
        *)
            echo "Unknown option: $1"
            echo "Usage: $0 [--run] [--sync]"
            exit 1
            ;;
    esac
done

echo "========================================"
echo "  N4L GPU Server Deployment"
echo "========================================"
echo ""
echo "Server: $GPU_SERVER"
echo "Remote dir: $REMOTE_DIR"
echo ""

# Test connexion
echo "Testing SSH connection..."
if ! ssh $SSH_OPTS $GPU_SERVER "echo 'Connection OK'" 2>/dev/null; then
    echo "Error: Cannot connect to GPU server"
    echo "Check your SSH key and server address"
    exit 1
fi
echo "✓ Connection successful"
echo ""

# Créer archive (sans venv ni __pycache__)
echo "Creating archive..."
cd "$PARENT_DIR"
tar --exclude='finetune/venv' \
    --exclude='finetune/__pycache__' \
    --exclude='finetune/**/__pycache__' \
    --exclude='finetune/.git' \
    --exclude='finetune/models' \
    --exclude='*.pyc' \
    -czvf /tmp/finetune.tar.gz finetune/

ARCHIVE_SIZE=$(du -h /tmp/finetune.tar.gz | cut -f1)
echo "✓ Archive created: $ARCHIVE_SIZE"
echo ""

# Transférer
echo "Uploading to server..."
scp $SSH_OPTS /tmp/finetune.tar.gz $GPU_SERVER:/tmp/
echo "✓ Upload complete"
echo ""

# Setup sur le serveur
echo "Setting up on server..."
ssh $SSH_OPTS $GPU_SERVER << 'REMOTE_SCRIPT'
set -e

cd ~
rm -rf n4l-finetune
mkdir -p n4l-finetune
cd n4l-finetune
tar -xzf /tmp/finetune.tar.gz --strip-components=1
rm /tmp/finetune.tar.gz

# Créer venv si nécessaire
if [ ! -d "venv" ]; then
    echo "Creating virtual environment..."
    python3 -m venv venv
fi

source venv/bin/activate

# Installer dépendances
echo "Installing dependencies..."
pip install --upgrade pip -q
pip install -r requirements.txt -q

# Vérifier GPU
echo ""
echo "=== GPU Status ==="
nvidia-smi --query-gpu=name,memory.total,memory.free --format=csv
echo ""
python -c "import torch; print(f'PyTorch CUDA: {torch.cuda.is_available()}'); print(f'GPU: {torch.cuda.get_device_name(0) if torch.cuda.is_available() else \"None\"}')"

# Stats dataset
echo ""
echo "=== Dataset Stats ==="
wc -l data/massive/splits/*.jsonl

echo ""
echo "✓ Setup complete!"
REMOTE_SCRIPT

echo "✓ Server setup complete"
echo ""

if [ "$SYNC_ONLY" = true ]; then
    echo "Sync only mode - exiting"
    exit 0
fi

if [ "$RUN_TRAINING" = true ]; then
    echo "========================================"
    echo "  Starting Training on GPU"
    echo "========================================"
    echo ""
    echo "Training will run in background with nohup"
    echo "Logs: ~/n4l-finetune/training.log"
    echo ""

    ssh $SSH_OPTS $GPU_SERVER << 'TRAIN_SCRIPT'
cd ~/n4l-finetune
source venv/bin/activate

# Lancer en background avec nohup
nohup bash -c '
cd ~/n4l-finetune
source venv/bin/activate
export AUTO_CONFIRM=true
./scripts/train.sh --no-wandb 2>&1
' > training.log 2>&1 &

echo "Training started in background (PID: $!)"
echo ""
echo "To monitor:"
echo "  tail -f ~/n4l-finetune/training.log"
echo ""
echo "To check status:"
echo "  ps aux | grep train.py"
TRAIN_SCRIPT

else
    echo "========================================"
    echo "  Deployment Complete!"
    echo "========================================"
    echo ""
    echo "To start training manually:"
    echo "  ssh $SSH_OPTS $GPU_SERVER"
    echo "  cd ~/n4l-finetune"
    echo "  source venv/bin/activate"
    echo "  ./scripts/train.sh --no-wandb"
    echo ""
    echo "Or run this script with --run to start automatically:"
    echo "  ./deploy_to_gpu.sh --run"
fi
