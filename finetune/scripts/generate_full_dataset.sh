#!/bin/bash
# Script de génération du dataset complet pour fine-tuning N4L
# Combine: templates améliorés + données Claude + fichiers existants
#
# Objectif: 1500+ exemples
# - ~1000 templates (10 domaines x 100)
# - ~500 synthétiques Claude (10 domaines x 50)
# - ~60 depuis fichiers N4L existants
#
# Usage: ./generate_full_dataset.sh [--skip-claude] [--templates-only]

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

echo "========================================"
echo "  N4L Full Dataset Generation"
echo "========================================"
echo ""

# Configuration
TEMPLATES_PER_DOMAIN=100
CLAUDE_PER_DOMAIN=50
OUTPUT_DIR="${PROJECT_DIR}/data/full_dataset"
SKIP_CLAUDE=false
TEMPLATES_ONLY=false

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --skip-claude)
            SKIP_CLAUDE=true
            shift
            ;;
        --templates-only)
            TEMPLATES_ONLY=true
            SKIP_CLAUDE=true
            shift
            ;;
        --templates-per-domain)
            TEMPLATES_PER_DOMAIN="$2"
            shift 2
            ;;
        --claude-per-domain)
            CLAUDE_PER_DOMAIN="$2"
            shift 2
            ;;
        --output)
            OUTPUT_DIR="$2"
            shift 2
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Calculate expected totals
EXPECTED_TEMPLATES=$((10 * TEMPLATES_PER_DOMAIN))
EXPECTED_CLAUDE=$((10 * CLAUDE_PER_DOMAIN + 60))
EXPECTED_TOTAL=$((EXPECTED_TEMPLATES + EXPECTED_CLAUDE))

if [ "$SKIP_CLAUDE" = true ]; then
    EXPECTED_TOTAL=$EXPECTED_TEMPLATES
fi

echo "Configuration:"
echo "  Templates per domain: $TEMPLATES_PER_DOMAIN"
echo "  Expected templates: ~$EXPECTED_TEMPLATES"
if [ "$SKIP_CLAUDE" = false ]; then
    echo "  Claude per domain: $CLAUDE_PER_DOMAIN"
    echo "  Expected Claude: ~$EXPECTED_CLAUDE"
fi
echo "  Expected total: ~$EXPECTED_TOTAL"
echo "  Output: $OUTPUT_DIR"
echo ""

# Activate venv
if [ -d "${PROJECT_DIR}/venv" ]; then
    source "${PROJECT_DIR}/venv/bin/activate"
else
    echo "Creating virtual environment..."
    python3 -m venv "${PROJECT_DIR}/venv"
    source "${PROJECT_DIR}/venv/bin/activate"
    pip install -r "${PROJECT_DIR}/requirements.txt"
fi

# Create output directories
mkdir -p "$OUTPUT_DIR"
mkdir -p "$OUTPUT_DIR/splits"

cd "$PROJECT_DIR"

# ============================================
# Step 1: Generate enhanced templates
# ============================================
echo ""
echo "========================================"
echo "Step 1: Generating enhanced templates..."
echo "========================================"

python src/enhanced_template_generator.py \
    --num-per-domain "$TEMPLATES_PER_DOMAIN" \
    --output "$OUTPUT_DIR/templates.jsonl"

TEMPLATE_COUNT=$(wc -l < "$OUTPUT_DIR/templates.jsonl")
echo "Generated $TEMPLATE_COUNT template examples"

# ============================================
# Step 2: Generate Claude synthetic data (if not skipped)
# ============================================
if [ "$SKIP_CLAUDE" = false ]; then
    if [ -z "$ANTHROPIC_API_KEY" ]; then
        echo ""
        echo "⚠️  ANTHROPIC_API_KEY not set - skipping Claude generation"
        echo "   Set it with: export ANTHROPIC_API_KEY=your_key"
        SKIP_CLAUDE=true
    else
        echo ""
        echo "========================================"
        echo "Step 2: Generating Claude synthetic data..."
        echo "========================================"

        # Install anthropic if needed
        pip show anthropic > /dev/null 2>&1 || pip install anthropic

        python src/claude_data_generator.py \
            --api-key "$ANTHROPIC_API_KEY" \
            --output-dir "$OUTPUT_DIR/claude" \
            --n4l-dir "../examples" \
            --num-per-domain "$CLAUDE_PER_DOMAIN" \
            --model "claude-sonnet-4-20250514"

        if [ -f "$OUTPUT_DIR/claude/all_examples.jsonl" ]; then
            CLAUDE_COUNT=$(wc -l < "$OUTPUT_DIR/claude/all_examples.jsonl")
            echo "Generated $CLAUDE_COUNT Claude examples"
        fi
    fi
fi

# ============================================
# Step 3: Merge all data
# ============================================
echo ""
echo "========================================"
echo "Step 3: Merging datasets..."
echo "========================================"

# Combine all JSONL files
cat "$OUTPUT_DIR/templates.jsonl" > "$OUTPUT_DIR/all_combined.jsonl"

if [ "$SKIP_CLAUDE" = false ] && [ -f "$OUTPUT_DIR/claude/all_examples.jsonl" ]; then
    cat "$OUTPUT_DIR/claude/all_examples.jsonl" >> "$OUTPUT_DIR/all_combined.jsonl"
fi

TOTAL_COUNT=$(wc -l < "$OUTPUT_DIR/all_combined.jsonl")
echo "Combined dataset: $TOTAL_COUNT examples"

# ============================================
# Step 4: Create train/val/test splits
# ============================================
echo ""
echo "========================================"
echo "Step 4: Creating splits..."
echo "========================================"

python -c "
import json
import random

# Load all examples
examples = []
with open('$OUTPUT_DIR/all_combined.jsonl', 'r') as f:
    for line in f:
        examples.append(json.loads(line))

# Shuffle
random.seed(42)
random.shuffle(examples)

# Split 80/10/10
n = len(examples)
train_end = int(n * 0.8)
val_end = train_end + int(n * 0.1)

train = examples[:train_end]
val = examples[train_end:val_end]
test = examples[val_end:]

# Save splits
for split_name, split_data in [('train', train), ('val', val), ('test', test)]:
    with open(f'$OUTPUT_DIR/splits/{split_name}.jsonl', 'w') as f:
        for ex in split_data:
            f.write(json.dumps(ex, ensure_ascii=False) + '\n')
    print(f'{split_name}: {len(split_data)} examples')
"

# ============================================
# Final statistics
# ============================================
echo ""
echo "========================================"
echo "  Generation Complete!"
echo "========================================"
echo ""
echo "=== Dataset Statistics ==="
wc -l "$OUTPUT_DIR/splits"/*.jsonl
echo ""

# Domain distribution
echo "=== Domain Distribution ==="
python -c "
import json
from collections import Counter

domain_counts = Counter()
source_counts = Counter()

with open('$OUTPUT_DIR/all_combined.jsonl', 'r') as f:
    for line in f:
        ex = json.loads(line)
        domain_counts[ex.get('domain', 'unknown')] += 1
        source_counts[ex.get('source', 'unknown')] += 1

print('By domain:')
for domain, count in sorted(domain_counts.items()):
    print(f'  {domain}: {count}')

print()
print('By source:')
for source, count in sorted(source_counts.items()):
    print(f'  {source}: {count}')
"

echo ""
echo "Output directory: $OUTPUT_DIR"
echo ""
echo "Next steps:"
echo "  1. Review data quality: head -5 $OUTPUT_DIR/splits/train.jsonl | jq ."
echo "  2. Update config to use new data"
echo "  3. Run training: ./scripts/train.sh"
