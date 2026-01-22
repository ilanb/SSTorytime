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
from fastapi.responses import StreamingResponse
from contextlib import asynccontextmanager
import logging
import os
import json
import asyncio

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
            vllm_url=os.environ.get("VLLM_URL", "http://86.204.69.30:8001/v1"),
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


@app.post("/reason/stream")
async def reason_stream(request: ReasoningRequest):
    """
    Perform reasoning analysis with streaming response.

    Streams each phase of the hierarchical reasoning process:
    1. Planning phase
    2. Each reasoning step
    3. Final synthesis
    """
    if not hrm_engine:
        raise HTTPException(status_code=503, detail="HRM Engine not initialized")

    async def generate():
        try:
            logger.info(f"Streaming reasoning request: type={request.reasoning_type}")

            # Convert evidence to dict format
            evidence_dicts = [e.model_dump() for e in request.evidence]

            # Phase 1: Planning
            yield f"data: {json.dumps({'phase': 'planning', 'message': '🧠 Phase 1: Planification stratégique...'})}\n\n"
            await asyncio.sleep(0.1)

            plan = await asyncio.to_thread(
                hrm_engine._high_level_planning,
                request.context,
                request.question,
                evidence_dicts
            )
            plan["strategy"] = request.reasoning_type.value

            strategy = plan.get('strategy', 'déductif')
            steps_count = len(plan.get('reasoning_steps', []))
            planning_msg = f"Stratégie: {strategy} - {steps_count} étapes planifiées"
            yield f"data: {json.dumps({'phase': 'planning_complete', 'plan': plan, 'message': planning_msg}, ensure_ascii=False)}\n\n"
            await asyncio.sleep(0.1)

            # Phase 2: Execute each reasoning step
            reasoning_chain = []
            reasoning_steps = plan.get("reasoning_steps", ["analyze_evidence"])

            for i, step in enumerate(reasoning_steps[:hrm_engine.config.max_reasoning_depth]):
                step_name = step if isinstance(step, str) else str(step)
                step_msg = f"Étape {i+1}/{len(reasoning_steps)}: {step_name}..."
                yield f"data: {json.dumps({'phase': 'step_start', 'step': i+1, 'total': len(reasoning_steps), 'message': step_msg}, ensure_ascii=False)}\n\n"
                await asyncio.sleep(0.1)

                # Execute step
                prompt = f"""Tu es un système de raisonnement hiérarchique - NIVEAU INFÉRIEUR (exécution détaillée).

## Étape {i+1}: {step}

## Stratégie globale: {plan.get('strategy', 'déductif')}

## Contexte
{request.context}

## Preuves à analyser
{hrm_engine._format_evidence_detail(evidence_dicts)}

## TÂCHE
Exécute l'étape "{step}" de façon détaillée. Réponds en JSON avec:
1. "premise": La prémisse ou point de départ de cette étape
2. "analysis": L'analyse détaillée effectuée
3. "inference": La conclusion/inférence de cette étape
4. "evidence_used": IDs des preuves utilisées
5. "confidence": Score de confiance (0.0 à 1.0)

JSON:"""

                response = await asyncio.to_thread(hrm_engine.vllm.generate, prompt)
                logger.info(f"Step {i+1} response length: {len(response)} chars")

                step_result = None
                try:
                    # Try to find and parse JSON in response
                    json_start = response.find('{')
                    json_end = response.rfind('}') + 1
                    if json_start >= 0 and json_end > json_start:
                        json_str = response[json_start:json_end]
                        step_result = json.loads(json_str)
                        logger.info(f"Step {i+1} JSON parsed successfully")
                except json.JSONDecodeError as e:
                    logger.warning(f"Step {i+1} JSON parse failed: {e}")
                    step_result = None

                # If JSON parsing failed, extract useful text from response
                if step_result is None or not step_result.get("inference"):
                    # Clean the response - remove JSON artifacts and extract text
                    clean_response = response.strip()
                    # Remove common JSON prefixes/suffixes
                    for prefix in ["```json", "```", "JSON:", "json:"]:
                        if clean_response.startswith(prefix):
                            clean_response = clean_response[len(prefix):].strip()
                    for suffix in ["```"]:
                        if clean_response.endswith(suffix):
                            clean_response = clean_response[:-len(suffix)].strip()

                    # Extract inference from response text
                    inference_text = clean_response[:800] if clean_response else "Analyse complétée"
                    # If it looks like JSON, try to extract just the inference field
                    if step_result and isinstance(step_result.get("inference"), str):
                        inference_text = step_result["inference"]
                    elif step_result and isinstance(step_result.get("analysis"), str):
                        inference_text = step_result["analysis"]

                    step_result = {
                        "premise": step_result.get("premise", step_name) if step_result else step_name,
                        "analysis": step_result.get("analysis", inference_text[:400]) if step_result else inference_text[:400],
                        "inference": inference_text,
                        "evidence_used": step_result.get("evidence_used", []) if step_result else [],
                        "confidence": step_result.get("confidence", 0.65) if step_result else 0.65
                    }

                step_result["step_number"] = i + 1

                # Ensure inference is a string
                inference = step_result.get("inference", "")
                if isinstance(inference, (list, dict)):
                    inference = json.dumps(inference, ensure_ascii=False)

                reasoning_chain.append({
                    "step_number": i + 1,
                    "premise": str(step_result.get("premise", step_name)),
                    "inference": str(inference),
                    "confidence": float(step_result.get("confidence", 0.5)),
                    "evidence_used": step_result.get("evidence_used", [])
                })

                conf_pct = int(reasoning_chain[-1]['confidence'] * 100)
                complete_msg = f"Étape {i+1} terminée (confiance: {conf_pct}%)"
                yield f"data: {json.dumps({'phase': 'step_complete', 'step': i+1, 'result': reasoning_chain[-1], 'message': complete_msg}, ensure_ascii=False)}\n\n"
                await asyncio.sleep(0.1)

            # Phase 3: Synthesis
            yield f"data: {json.dumps({'phase': 'synthesis', 'message': 'Phase 3: Synthèse des conclusions...'}, ensure_ascii=False)}\n\n"
            await asyncio.sleep(0.1)

            synthesis = await asyncio.to_thread(
                hrm_engine._synthesize_conclusion,
                plan,
                reasoning_chain,
                request.question
            )

            # Build final result
            avg_confidence = sum(r.get("confidence", 0.5) for r in reasoning_chain) / len(reasoning_chain) if reasoning_chain else 0.5

            final_result = {
                "conclusion": synthesis.get("conclusion", ""),
                "confidence": synthesis.get("confidence", avg_confidence),
                "reasoning_chain": reasoning_chain,
                "alternative_conclusions": [
                    {"conclusion": c, "confidence": 0.4, "reason": "Alternative identifiée"}
                    for c in synthesis.get("alternative_conclusions", [])[:3]
                ],
                "warnings": synthesis.get("warnings", [])
            }

            yield f"data: {json.dumps({'phase': 'complete', 'result': final_result, 'message': 'Raisonnement terminé'}, ensure_ascii=False)}\n\n"

        except Exception as e:
            logger.error(f"Streaming reasoning error: {str(e)}")
            error_msg = f"Erreur: {str(e)}"
            yield f"data: {json.dumps({'phase': 'error', 'error': str(e), 'message': error_msg}, ensure_ascii=False)}\n\n"

    return StreamingResponse(
        generate(),
        media_type="text/event-stream",
        headers={
            "Cache-Control": "no-cache",
            "Connection": "keep-alive",
            "X-Accel-Buffering": "no"
        }
    )


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
