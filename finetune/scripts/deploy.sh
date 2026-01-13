#!/bin/bash
# Script de déploiement du modèle N4L vers Ollama
# Usage: ./deploy.sh [--model PATH] [--ollama-name NAME]

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

echo "========================================"
echo "  N4L Deployment Script"
echo "========================================"

# Default configuration
MODEL_PATH="${PROJECT_DIR}/models/n4l-qwen-lora"
GGUF_OUTPUT="${PROJECT_DIR}/models/n4l-generator.gguf"
OLLAMA_NAME="n4l-generator"
QUANTIZATION="Q4_K_M"

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --model)
            MODEL_PATH="$2"
            shift 2
            ;;
        --gguf)
            GGUF_OUTPUT="$2"
            shift 2
            ;;
        --ollama-name)
            OLLAMA_NAME="$2"
            shift 2
            ;;
        --quantization)
            QUANTIZATION="$2"
            shift 2
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
    echo "Error: Virtual environment not found."
    exit 1
fi

# Check model exists
if [ ! -d "$MODEL_PATH" ]; then
    echo "Error: Model not found at $MODEL_PATH"
    echo "Run train.sh first."
    exit 1
fi

# Check Ollama
if ! command -v ollama &> /dev/null; then
    echo "Error: Ollama not installed."
    echo "Visit: https://ollama.ai/download"
    exit 1
fi

# Print configuration
echo ""
echo "Configuration:"
echo "  Model path: $MODEL_PATH"
echo "  GGUF output: $GGUF_OUTPUT"
echo "  Ollama name: $OLLAMA_NAME"
echo "  Quantization: $QUANTIZATION"
echo ""

# Confirm
read -p "Start deployment? (y/n) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Deployment cancelled."
    exit 0
fi

# Run conversion and deployment
cd "$PROJECT_DIR"

python src/convert_ollama.py \
    --model "$MODEL_PATH" \
    --output "$GGUF_OUTPUT" \
    --ollama-name "$OLLAMA_NAME" \
    --quantization "$QUANTIZATION" \
    --create-ollama \
    --test

echo ""
echo "========================================"
echo "  Deployment complete!"
echo "========================================"
echo ""
echo "Model ready to use:"
echo "  ollama run $OLLAMA_NAME"
echo ""
echo "Example:"
echo "  ollama run $OLLAMA_NAME 'Convertis ce texte en N4L: Jean est médecin à Paris.'"
