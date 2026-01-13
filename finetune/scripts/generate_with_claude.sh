#!/bin/bash
# Script de génération de données avec Claude API
# Usage: ./generate_with_claude.sh [--num-per-domain N]

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

echo "========================================"
echo "  N4L Data Generation with Claude"
echo "========================================"

# Check API key
if [ -z "$ANTHROPIC_API_KEY" ]; then
    echo "Error: ANTHROPIC_API_KEY not set"
    echo ""
    echo "Please set your API key:"
    echo "  export ANTHROPIC_API_KEY=your_key_here"
    echo ""
    exit 1
fi

# Default configuration
NUM_PER_DOMAIN="${NUM_PER_DOMAIN:-50}"
N4L_EXAMPLES_DIR="${PROJECT_DIR}/../examples"
OUTPUT_DIR="${PROJECT_DIR}/data/claude_generated"
MODEL="${CLAUDE_MODEL:-claude-sonnet-4-20250514}"

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --num-per-domain)
            NUM_PER_DOMAIN="$2"
            shift 2
            ;;
        --model)
            MODEL="$2"
            shift 2
            ;;
        --output)
            OUTPUT_DIR="$2"
            shift 2
            ;;
        --skip-existing)
            SKIP_EXISTING="--skip-existing"
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
    echo "Creating virtual environment..."
    python3 -m venv "${PROJECT_DIR}/venv"
    source "${PROJECT_DIR}/venv/bin/activate"
    pip install -r "${PROJECT_DIR}/requirements.txt"
fi

# Install anthropic if needed
pip show anthropic > /dev/null 2>&1 || pip install anthropic

# Calculate expected examples
# 10 domains x NUM_PER_DOMAIN + ~60 from existing files
EXPECTED_SYNTHETIC=$((10 * NUM_PER_DOMAIN))
echo ""
echo "Configuration:"
echo "  Model: $MODEL"
echo "  Examples per domain: $NUM_PER_DOMAIN"
echo "  Total domains: 10"
echo "  Expected synthetic: ~$EXPECTED_SYNTHETIC"
echo "  Expected from existing: ~60"
echo "  Total expected: ~$((EXPECTED_SYNTHETIC + 60))"
echo "  Output: $OUTPUT_DIR"
echo ""

# Estimate cost (rough: ~0.003$ per input/output token pair for Sonnet)
ESTIMATED_CALLS=$((EXPECTED_SYNTHETIC * 2 + 60))
echo "Estimated API calls: ~$ESTIMATED_CALLS"
echo "Estimated cost: ~\$$(echo "scale=2; $ESTIMATED_CALLS * 0.02" | bc)"
echo ""

# Confirm
read -p "Start generation? (y/n) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Generation cancelled."
    exit 0
fi

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Run generation
cd "$PROJECT_DIR"

python src/claude_data_generator.py \
    --api-key "$ANTHROPIC_API_KEY" \
    --output-dir "$OUTPUT_DIR" \
    --n4l-dir "$N4L_EXAMPLES_DIR" \
    --num-per-domain "$NUM_PER_DOMAIN" \
    --model "$MODEL" \
    --create-splits \
    $SKIP_EXISTING

echo ""
echo "========================================"
echo "  Generation complete!"
echo "========================================"
echo ""
echo "Generated data in: $OUTPUT_DIR"
echo ""

# Show statistics
if [ -d "$OUTPUT_DIR/splits" ]; then
    echo "=== Dataset Statistics ==="
    wc -l "$OUTPUT_DIR/splits"/*.jsonl
fi

echo ""
echo "Next steps:"
echo "  1. Review generated data"
echo "  2. Merge with existing data if needed"
echo "  3. Run training: ./scripts/train.sh"
