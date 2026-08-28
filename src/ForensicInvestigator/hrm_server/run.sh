#!/bin/bash
# HRM Server Startup Script

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Engine mode:
# - USE_SAPIENT=false (default): Fast algorithmic engine (instantaneous)
# - USE_SAPIENT=true: vLLM-powered reasoning (slower but more sophisticated)
export USE_SAPIENT="${USE_SAPIENT:-true}"

# Configuration LLM (only used if USE_SAPIENT=true) - vLLM backend on SPARK GB10.
# Weights: Qwen/Qwen3.8-27B-FP8 (128k ctx, MTP self-speculation, prefix caching),
# served under the alias "Qwen3.5-9B" (--served-model-name). Only that alias is
# accepted by the API; "Qwen3.8-27B-FP8" returns a 404. The SPARK is shared with
# other applications: use its configuration as-is, do not rename the alias.
export VLLM_URL="${VLLM_URL:-http://86.204.69.30:8001/v1}"
export VLLM_MODEL="${VLLM_MODEL:-Qwen3.5-9B}"

# Check for virtual environment
if [ ! -d "venv" ]; then
    echo "Creating virtual environment..."
    python3 -m venv venv
fi

# Activate virtual environment
source venv/bin/activate

# Install dependencies
echo "Installing dependencies..."
pip install -q -r requirements.txt

# Start server
echo "Starting HRM Server on port 8081 (USE_SAPIENT=$USE_SAPIENT)..."
uvicorn main:app --host 0.0.0.0 --port 8081 --reload
