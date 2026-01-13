#!/bin/bash
# Script d'entraînement pour le fine-tuning N4L
# Usage: ./train.sh [--model MODEL] [--epochs N]

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

echo "========================================"
echo "  N4L Fine-tuning Script"
echo "========================================"

# Default configuration
MODEL="${MODEL:-Qwen/Qwen2.5-7B-Instruct}"
EPOCHS="${EPOCHS:-3}"
BATCH_SIZE="${BATCH_SIZE:-2}"
LEARNING_RATE="${LEARNING_RATE:-2e-4}"
OUTPUT_DIR="${PROJECT_DIR}/models/n4l-qwen-lora"
TRAIN_DATA="${PROJECT_DIR}/data/massive/splits/train.jsonl"
VAL_DATA="${PROJECT_DIR}/data/massive/splits/val.jsonl"
USE_WANDB="${USE_WANDB:-true}"

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --model)
            MODEL="$2"
            shift 2
            ;;
        --epochs)
            EPOCHS="$2"
            shift 2
            ;;
        --batch-size)
            BATCH_SIZE="$2"
            shift 2
            ;;
        --lr)
            LEARNING_RATE="$2"
            shift 2
            ;;
        --output)
            OUTPUT_DIR="$2"
            shift 2
            ;;
        --no-wandb)
            USE_WANDB=false
            shift
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Activate venv
if [ -d "${PROJECT_DIR}/venv" ]; then
    source "${PROJECT_DIR}/venv/bin/activate"
else
    echo "Error: Virtual environment not found. Run prepare_data.sh first."
    exit 1
fi

# Check training data
if [ ! -f "$TRAIN_DATA" ]; then
    echo "Error: Training data not found at $TRAIN_DATA"
    echo "Run prepare_data.sh first."
    exit 1
fi

# Check GPU
echo "Checking GPU availability..."
python -c "import torch; print(f'CUDA available: {torch.cuda.is_available()}')"
python -c "import torch; print(f'GPU: {torch.cuda.get_device_name(0) if torch.cuda.is_available() else \"None\"}')"

# Print configuration
echo ""
echo "Configuration:"
echo "  Model: $MODEL"
echo "  Epochs: $EPOCHS"
echo "  Batch size: $BATCH_SIZE"
echo "  Learning rate: $LEARNING_RATE"
echo "  Output: $OUTPUT_DIR"
echo "  Train data: $TRAIN_DATA"
echo "  Val data: $VAL_DATA"
echo "  W&B logging: $USE_WANDB"
echo ""

# Confirm (skip if --yes or -y flag or non-interactive)
if [[ "$AUTO_CONFIRM" != "true" ]] && [[ -t 0 ]]; then
    read -p "Start training? (y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Training cancelled."
        exit 0
    fi
else
    echo "Auto-confirming (non-interactive mode)..."
fi

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Build command
CMD="python src/train.py"
CMD="$CMD --model $MODEL"
CMD="$CMD --dataset $TRAIN_DATA"
CMD="$CMD --val-dataset $VAL_DATA"
CMD="$CMD --output $OUTPUT_DIR"
CMD="$CMD --epochs $EPOCHS"
CMD="$CMD --batch-size $BATCH_SIZE"
CMD="$CMD --learning-rate $LEARNING_RATE"

if [ "$USE_WANDB" = false ]; then
    CMD="$CMD --no-wandb"
fi

# Run training
echo ""
echo "Starting training..."
echo "Command: $CMD"
echo ""

cd "$PROJECT_DIR"
$CMD

echo ""
echo "========================================"
echo "  Training complete!"
echo "========================================"
echo ""
echo "Model saved to: $OUTPUT_DIR"
echo ""
echo "Next steps:"
echo "  1. Evaluate: ./scripts/evaluate.sh"
echo "  2. Deploy: ./scripts/deploy.sh"
