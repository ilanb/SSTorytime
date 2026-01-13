"""
HRM (Hypothetical Reasoning Model) FastAPI Server.

This server provides endpoints for forensic reasoning capabilities:
- General reasoning with evidence analysis
- Hypothesis verification
- Contradiction detection
- Cross-case pattern analysis

Uses sapientinc/HRM hierarchical reasoning approach combined with vLLM.
"""
from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from contextlib import asynccontextmanager
import logging
import os

from models import (
    ReasoningRequest, ReasoningResponse,
    HypothesisVerificationRequest, HypothesisVerificationResponse,
    ContradictionRequest, ContradictionResponse,
    CrossCaseRequest, CrossCaseResponse,
    ReasoningStep, Contradiction, CasePattern
)

# Choose engine via USE_SAPIENT environment variable:
# - USE_SAPIENT=false (default): Fast local/algorithmic engine (instantaneous)
# - USE_SAPIENT=true: vLLM-powered reasoning (slower but more sophisticated)
USE_SAPIENT = os.environ.get("USE_SAPIENT", "false").lower() == "true"

if USE_SAPIENT:
    try:
        from hrm_sapient import HRMSapientEngine, HRMConfig
    except ImportError:
        from hrm_engine import HRMEngine
        USE_SAPIENT = False
else:
    from hrm_engine import HRMEngine

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

# Global HRM engine instance
hrm_engine = None


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Manage application lifespan - initialize and cleanup."""
    global hrm_engine

    if USE_SAPIENT:
        logger.info("Initializing HRM Sapient Engine (hierarchical reasoning + vLLM)...")
        config = HRMConfig(
            vllm_url=os.environ.get("VLLM_URL", "http://86.204.69.30:8001"),
            vllm_model=os.environ.get("VLLM_MODEL", "Qwen/Qwen2.5-7B-Instruct")
        )
        hrm_engine = HRMSapientEngine(config)
        logger.info("HRM Sapient Engine initialized successfully")
    else:
        logger.info("Initializing basic HRM Engine (rule-based)...")
        from hrm_engine import HRMEngine
        hrm_engine = HRMEngine()
        logger.info("Basic HRM Engine initialized")

    yield
    logger.info("Shutting down HRM Engine...")


app = FastAPI(
    title="HRM - Hypothetical Reasoning Model API",
    description="Forensic investigation reasoning engine providing hypothesis verification, contradiction detection, and cross-case analysis.",
    version="1.0.0",
    lifespan=lifespan
)

# Configure CORS
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],  # In production, restrict to specific origins
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)


@app.get("/health")
@app.get("/status")
async def health_check():
    """Health check endpoint."""
    return {
        "status": "healthy",
        "engine_ready": hrm_engine is not None,
        "available": hrm_engine is not None
    }


@app.get("/info")
async def get_info():
    """Get information about the HRM API."""
    return {
        "name": "HRM - Hypothetical Reasoning Model",
        "version": "1.0.0",
        "capabilities": [
            "reasoning",
            "hypothesis_verification",
            "contradiction_detection",
            "cross_case_analysis"
        ],
        "supported_reasoning_types": [
            "deductive",
            "inductive",
            "abductive",
            "analogical"
        ]
    }


@app.post("/reason", response_model=ReasoningResponse)
async def reason(request: ReasoningRequest):
    """
    Perform reasoning analysis on provided context and evidence.

    This endpoint uses hierarchical reasoning to:
    1. Plan the reasoning strategy (high-level)
    2. Execute reasoning steps (low-level)
    3. Generate conclusions with confidence scores
    """
    if not hrm_engine:
        raise HTTPException(status_code=503, detail="HRM Engine not initialized")

    try:
        logger.info(f"Reasoning request: type={request.reasoning_type}, evidence_count={len(request.evidence)}")

        # Convert evidence to dict format
        evidence_dicts = [e.model_dump() for e in request.evidence]

        # Execute reasoning
        result = hrm_engine.reason(
            context=request.context,
            question=request.question,
            evidence=evidence_dicts,
            reasoning_type=request.reasoning_type.value,
            max_depth=request.max_depth
        )

        # Convert to response model
        reasoning_chain = [
            ReasoningStep(**step) for step in result["reasoning_chain"]
        ]

        return ReasoningResponse(
            conclusion=result["conclusion"],
            confidence=result["confidence"],
            reasoning_chain=reasoning_chain,
            alternative_conclusions=result.get("alternative_conclusions", []),
            warnings=result.get("warnings", [])
        )

    except Exception as e:
        logger.error(f"Reasoning error: {str(e)}")
        raise HTTPException(status_code=500, detail=f"Reasoning failed: {str(e)}")


@app.post("/verify-hypothesis", response_model=HypothesisVerificationResponse)
async def verify_hypothesis(request: HypothesisVerificationRequest):
    """
    Verify a hypothesis against available evidence.

    Returns:
    - Whether the hypothesis is supported
    - Confidence score
    - Supporting and contradicting reasons
    - Missing evidence recommendations
    """
    if not hrm_engine:
        raise HTTPException(status_code=503, detail="HRM Engine not initialized")

    try:
        logger.info(f"Hypothesis verification: id={request.hypothesis.id}")

        # Convert to dict format
        hypothesis_dict = request.hypothesis.model_dump()
        evidence_dicts = [e.model_dump() for e in request.evidence]

        # Verify hypothesis
        result = hrm_engine.verify_hypothesis(
            hypothesis=hypothesis_dict,
            evidence=evidence_dicts,
            case_context=request.case_context,
            strict_mode=request.strict_mode
        )

        return HypothesisVerificationResponse(**result)

    except Exception as e:
        logger.error(f"Hypothesis verification error: {str(e)}")
        raise HTTPException(status_code=500, detail=f"Verification failed: {str(e)}")


@app.post("/find-contradictions", response_model=ContradictionResponse)
@app.post("/contradictions", response_model=ContradictionResponse)
async def find_contradictions(request: ContradictionRequest):
    """
    Detect contradictions in statements and evidence.

    Analyzes:
    - Contradictions between statements
    - Conflicts between statements and evidence
    - Overall consistency score
    """
    if not hrm_engine:
        raise HTTPException(status_code=503, detail="HRM Engine not initialized")

    try:
        logger.info(f"Contradiction detection: statement_count={len(request.statements)}")

        # Convert evidence to dict format
        evidence_dicts = [e.model_dump() for e in request.evidence]

        # Find contradictions
        result = hrm_engine.find_contradictions(
            statements=request.statements,
            evidence=evidence_dicts,
            case_context=request.case_context
        )

        # Convert contradictions to model
        contradictions = [
            Contradiction(**c) for c in result["contradictions"]
        ]

        return ContradictionResponse(
            contradictions=contradictions,
            consistency_score=result["consistency_score"],
            analysis_summary=result["analysis_summary"]
        )

    except Exception as e:
        logger.error(f"Contradiction detection error: {str(e)}")
        raise HTTPException(status_code=500, detail=f"Detection failed: {str(e)}")


@app.post("/cross-case-reasoning", response_model=CrossCaseResponse)
async def cross_case_reasoning(request: CrossCaseRequest):
    """
    Analyze patterns and connections across multiple cases.

    Provides:
    - Pattern detection across cases
    - Connection mapping
    - Investigative leads
    - Risk assessment
    """
    if not hrm_engine:
        raise HTTPException(status_code=503, detail="HRM Engine not initialized")

    try:
        logger.info(f"Cross-case reasoning: comparing {len(request.comparison_cases)} cases")

        # Execute cross-case analysis
        result = hrm_engine.cross_case_reasoning(
            primary_case=request.primary_case,
            comparison_cases=request.comparison_cases,
            focus_areas=request.focus_areas
        )

        # Convert patterns to model
        patterns = [
            CasePattern(**p) for p in result["patterns"]
        ]

        return CrossCaseResponse(
            patterns=patterns,
            connections=result["connections"],
            investigative_leads=result["investigative_leads"],
            risk_assessment=result["risk_assessment"],
            summary=result["summary"]
        )

    except Exception as e:
        logger.error(f"Cross-case reasoning error: {str(e)}")
        raise HTTPException(status_code=500, detail=f"Analysis failed: {str(e)}")


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(
        "main:app",
        host="0.0.0.0",
        port=8081,
        reload=True,
        log_level="info"
    )
