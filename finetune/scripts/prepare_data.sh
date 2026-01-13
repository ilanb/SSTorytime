#!/bin/bash
# Script de préparation des données pour le fine-tuning N4L
# Usage: ./prepare_data.sh

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

echo "========================================"
echo "  N4L Data Preparation Script"
echo "========================================"

# Configuration
N4L_EXAMPLES_DIR="${PROJECT_DIR}/../examples"
OUTPUT_DIR="${PROJECT_DIR}/data/processed"
SPLITS_DIR="${PROJECT_DIR}/data/splits"
NUM_TEMPLATES=50
LLM_MODEL="${OLLAMA_MODEL:-gpt-oss:20b}"

# Create directories
mkdir -p "$OUTPUT_DIR"
mkdir -p "$SPLITS_DIR"

# Check if venv exists
if [ ! -d "${PROJECT_DIR}/venv" ]; then
    echo "Creating virtual environment..."
    python3 -m venv "${PROJECT_DIR}/venv"
fi

# Activate venv
source "${PROJECT_DIR}/venv/bin/activate"

# Install dependencies if needed
if ! python -c "import torch" 2>/dev/null; then
    echo "Installing dependencies..."
    pip install -r "${PROJECT_DIR}/requirements.txt"
fi

# Check Ollama availability
echo "Checking Ollama availability..."
if curl -s http://localhost:11434/api/tags > /dev/null 2>&1; then
    echo "✓ Ollama is available"
    USE_LLM=true
else
    echo "⚠ Ollama not available - using template-only generation"
    USE_LLM=false
fi

# Generate data
echo ""
echo "Generating training data..."
echo "  N4L examples dir: $N4L_EXAMPLES_DIR"
echo "  Output dir: $OUTPUT_DIR"
echo "  Templates per domain: $NUM_TEMPLATES"
echo "  LLM model: $LLM_MODEL"
echo ""

cd "$PROJECT_DIR"

if [ "$USE_LLM" = true ]; then
    python src/data_generator.py \
        --n4l-dir "$N4L_EXAMPLES_DIR" \
        --output-dir "$OUTPUT_DIR" \
        --num-templates "$NUM_TEMPLATES" \
        --llm-model "$LLM_MODEL" \
        --create-splits
else
    python src/data_generator.py \
        --n4l-dir "$N4L_EXAMPLES_DIR" \
        --output-dir "$OUTPUT_DIR" \
        --num-templates "$NUM_TEMPLATES" \
        --no-llm \
        --create-splits
fi

# Count generated examples
echo ""
echo "Generated data statistics:"
if [ -f "$SPLITS_DIR/train.jsonl" ]; then
    TRAIN_COUNT=$(wc -l < "$SPLITS_DIR/train.jsonl")
    VAL_COUNT=$(wc -l < "$SPLITS_DIR/val.jsonl")
    TEST_COUNT=$(wc -l < "$SPLITS_DIR/test.jsonl")
    echo "  Train: $TRAIN_COUNT examples"
    echo "  Val: $VAL_COUNT examples"
    echo "  Test: $TEST_COUNT examples"
fi

echo ""
echo "========================================"
echo "  Data preparation complete!"
echo "========================================"
echo ""
echo "Next steps:"
echo "  1. Review generated data in $SPLITS_DIR"
echo "  2. Run training: ./scripts/train.sh"
