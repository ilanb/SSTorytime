#!/bin/bash
# HRM Server Startup Script

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Engine mode:
# - USE_SAPIENT=false (default): Fast algorithmic engine (instantaneous)
# - USE_SAPIENT=true: vLLM-powered reasoning (slower but more sophisticated)
export USE_SAPIENT="${USE_SAPIENT:-false}"

# Configuration vLLM (only used if USE_SAPIENT=true)
export VLLM_URL="${VLLM_URL:-http://86.204.69.30:8001}"
export VLLM_MODEL="${VLLM_MODEL:-Qwen/Qwen2.5-7B-Instruct}"

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
