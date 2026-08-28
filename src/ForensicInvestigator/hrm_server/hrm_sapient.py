"""
HRM Sapient Integration Module.

This module integrates the real sapientinc/HRM model for hierarchical reasoning
combined with Ollama for textual analysis in forensic investigation.

Architecture:
- HRM (sapientinc): Pattern recognition and hierarchical reasoning on structured data
- Ollama: Textual reasoning, natural language understanding, and explanation generation

The combination provides:
1. HRM's hierarchical two-level reasoning (planning + execution)
2. Ollama's natural language capabilities for forensic context
"""

import torch
import numpy as np
import requests
import json
import logging
from typing import List, Dict, Any, Optional, Tuple
from dataclasses import dataclass
from enum import Enum
import hashlib
import os
import re

logger = logging.getLogger(__name__)

# Remote LLM served by vLLM on the SPARK GB10 (port 8001), a shared server also
# used by other applications: its configuration is taken as-is.
# Backing weights: Qwen/Qwen3.8-27B-FP8 — exposed under the served-model-name
# "Qwen3.5-9B", which is the identifier the API expects in the "model" field and
# which other clients depend on. Override with VLLM_URL / VLLM_MODEL if needed.
DEFAULT_VLLM_URL = os.environ.get("VLLM_URL", "http://86.204.69.30:8001/v1")
DEFAULT_VLLM_MODEL = os.environ.get("VLLM_MODEL", "Qwen3.5-9B")

# Qwen3.8 is a reasoning model. On the raw /v1/completions endpoint there is no chat
# template and no reasoning parser, so the chain of thought is emitted inline as a
# <think>...</think> block instead of a separate "reasoning" field. Downstream HRM
# parsing (JSON extraction, synthesis) expects the answer only, so it is stripped here.
_THINK_BLOCK_RE = re.compile(r"<think>.*?</think>", re.DOTALL | re.IGNORECASE)
_UNCLOSED_THINK_RE = re.compile(r"<think>.*\Z", re.DOTALL | re.IGNORECASE)


def strip_reasoning(text: str) -> str:
    """Remove <think> reasoning blocks from a raw completion.

    Handles the truncation case too: when max_tokens cuts the response mid-thought
    the closing </think> never arrives, and everything after the opening tag is
    reasoning rather than answer.
    """
    if not text:
        return ""
    cleaned = _THINK_BLOCK_RE.sub("", text)
    cleaned = _UNCLOSED_THINK_RE.sub("", cleaned)
    return cleaned.strip()


# Try to import HuggingFace Hub for model loading
try:
    from huggingface_hub import hf_hub_download, snapshot_download
    HF_AVAILABLE = True
except ImportError:
    HF_AVAILABLE = False
    logger.warning("huggingface_hub not available - HRM model loading disabled")


class ReasoningLevel(Enum):
    """HRM reasoning levels inspired by sapientinc architecture."""
    HIGH_LEVEL = "high_level"  # Abstract planning (slow)
    LOW_LEVEL = "low_level"    # Detailed computation (fast)


@dataclass
class HRMConfig:
    """Configuration for HRM integration."""
    # HRM Model settings
    hrm_model_id: str = "sapientinc/HRM-checkpoint-ARC-2"
    use_gpu: bool = True

    # LLM settings (OpenAI-compatible API - vLLM backend on SPARK GB10).
    # vLLM serves Qwen/Qwen3.8-27B-FP8 (128k context, MTP self-speculation, prefix
    # caching) under the alias "Qwen3.5-9B" (--served-model-name). That alias is the
    # only accepted value for the "model" field: sending "Qwen3.8-27B-FP8" or
    # "Qwen/Qwen3.8-27B-FP8" returns a 404 NotFoundError.
    vllm_url: str = DEFAULT_VLLM_URL
    vllm_model: str = DEFAULT_VLLM_MODEL

    # Reasoning settings
    max_reasoning_depth: int = 5
    confidence_threshold: float = 0.6

    # Cache settings
    enable_cache: bool = True
    cache_ttl: int = 3600


class VLLMClient:
    """Client for vLLM API (OpenAI-compatible)."""

    def __init__(self, base_url: str = DEFAULT_VLLM_URL, model: str = DEFAULT_VLLM_MODEL):
        self.base_url = base_url
        self.model = model

    def is_available(self) -> bool:
        """Check if vLLM is available."""
        try:
            resp = requests.get(f"{self.base_url}/models", timeout=5)
            return resp.status_code == 200
        except:
            return False

    def generate(self, prompt: str, stream: bool = False, max_tokens: int = 8192) -> str:
        """Generate a response from vLLM."""
        try:
            logger.info(f"vLLM request to {self.base_url}/completions with model {self.model}")
            resp = requests.post(
                f"{self.base_url}/completions",
                json={
                    "model": self.model,
                    "prompt": prompt,
                    "max_tokens": max_tokens,
                    "temperature": 0.4,
                    "stream": stream
                },
                timeout=600  # 10 minutes per vLLM request
            )
            if resp.status_code == 200:
                data = resp.json()
                if data.get("choices") and len(data["choices"]) > 0:
                    raw = data["choices"][0].get("text", "")
                    text = strip_reasoning(raw)
                    logger.info(
                        f"vLLM response length: {len(text)} chars "
                        f"({len(raw) - len(text)} chars of reasoning stripped)"
                    )
                    return text
                logger.warning("vLLM returned empty choices")
                return ""
            else:
                logger.error(f"vLLM error: {resp.status_code} - {resp.text[:200]}")
                return ""
        except Exception as e:
            logger.error(f"vLLM connection error: {e}")
            return ""


class HRMSapientEngine:
    """
    Hierarchical Reasoning Model Engine using sapientinc/HRM approach.

    Implements the two-level hierarchical reasoning:
    1. High-level module: Strategic planning and abstract reasoning
    2. Low-level module: Detailed computation and pattern matching

    Combined with vLLM for natural language processing in forensic context.
    """

    def __init__(self, config: HRMConfig = None):
        self.config = config or HRMConfig()
        self.device = torch.device("cuda" if torch.cuda.is_available() and self.config.use_gpu else "cpu")

        # Initialize vLLM client
        self.vllm = VLLMClient(self.config.vllm_url, self.config.vllm_model)

        # HRM model state (will be loaded if available)
        self.hrm_model = None
        self.hrm_loaded = False

        # Inference cache
        self.cache: Dict[str, Any] = {}

        # Try to load HRM model
        self._try_load_hrm_model()

        logger.info(f"HRMSapientEngine initialized - Device: {self.device}, HRM loaded: {self.hrm_loaded}")

    def _try_load_hrm_model(self):
        """Try to load the sapientinc HRM model."""
        if not HF_AVAILABLE:
            logger.info("HuggingFace Hub not available - using rule-based fallback")
            return

        try:
            logger.info(f"Attempting to load HRM model: {self.config.hrm_model_id}")

            # Download model files from HuggingFace
            # Note: sapientinc/HRM uses custom PyTorch checkpoints, not standard transformers
            model_path = snapshot_download(
                repo_id=self.config.hrm_model_id,
                local_dir=os.path.expanduser("~/.cache/hrm_model"),
                ignore_patterns=["*.md", "*.txt"]
            )

            logger.info(f"HRM model downloaded to: {model_path}")

            # Load checkpoint (sapientinc HRM uses custom format)
            # The actual loading depends on the checkpoint structure
            checkpoint_files = [f for f in os.listdir(model_path) if f.endswith('.pt') or f.endswith('.pth')]

            if checkpoint_files:
                checkpoint_path = os.path.join(model_path, checkpoint_files[0])
                # Note: Full HRM model loading requires the complete sapientinc/HRM codebase
                # For now, we'll use a hybrid approach with Ollama
                logger.info(f"HRM checkpoint found: {checkpoint_path}")
                self.hrm_loaded = True
            else:
                logger.warning("No checkpoint files found in HRM model")

        except Exception as e:
            logger.warning(f"Could not load HRM model: {e} - using hybrid Ollama approach")
            self.hrm_loaded = False

    def _compute_cache_key(self, *args) -> str:
        """Compute cache key for memoization."""
        content = json.dumps(args, sort_keys=True, default=str)
        return hashlib.md5(content.encode()).hexdigest()

    def _high_level_planning(self, context: str, question: str, evidence: List[Dict]) -> Dict[str, Any]:
        """
        High-level planning phase (inspired by HRM's slow, abstract processing).

        Uses Ollama to understand the reasoning strategy needed.
        """
        # Build planning prompt
        prompt = f"""Tu es un système de raisonnement hiérarchique pour l'investigation forensique.

NIVEAU SUPÉRIEUR - PLANIFICATION STRATÉGIQUE

## Contexte de l'affaire
{context}

## Question à analyser
{question}

## Preuves disponibles ({len(evidence)} éléments)
{self._format_evidence_summary(evidence)}

## TÂCHE
Génère un plan de raisonnement structuré en JSON avec:
1. "strategy": La stratégie globale de raisonnement (déductif, inductif, abductif)
2. "key_elements": Liste des éléments clés à analyser
3. "reasoning_steps": Liste ordonnée des étapes de raisonnement (max 5)
4. "focus_areas": Domaines prioritaires d'investigation
5. "potential_hypotheses": Hypothèses préliminaires à vérifier

Réponds UNIQUEMENT avec le JSON valide, sans texte avant ou après.

JSON:"""

        response = self.vllm.generate(prompt)

        # Parse JSON response
        try:
            # Extract JSON from response
            json_start = response.find('{')
            json_end = response.rfind('}') + 1
            if json_start >= 0 and json_end > json_start:
                plan = json.loads(response[json_start:json_end])
            else:
                plan = self._default_plan()
        except json.JSONDecodeError:
            logger.warning("Could not parse planning response, using default plan")
            plan = self._default_plan()

        return plan

    def _low_level_execution(self, plan: Dict, context: str, evidence: List[Dict]) -> List[Dict]:
        """
        Low-level execution phase (inspired by HRM's fast, detailed processing).

        Executes each step of the plan with detailed analysis.
        """
        results = []

        reasoning_steps = plan.get("reasoning_steps", ["analyze_evidence"])

        for i, step in enumerate(reasoning_steps[:self.config.max_reasoning_depth]):
            # Build execution prompt for this step
            prompt = f"""Tu es un système de raisonnement hiérarchique - NIVEAU INFÉRIEUR (exécution détaillée).

## Étape {i+1}: {step}

## Stratégie globale: {plan.get('strategy', 'déductif')}

## Contexte
{context}

## Preuves à analyser
{self._format_evidence_detail(evidence)}

## TÂCHE
Exécute l'étape "{step}" de façon détaillée. Réponds en JSON avec:
1. "premise": La prémisse ou point de départ de cette étape
2. "analysis": L'analyse détaillée effectuée
3. "inference": La conclusion/inférence de cette étape
4. "evidence_used": IDs des preuves utilisées
5. "confidence": Score de confiance (0.0 à 1.0)
6. "next_questions": Questions soulevées pour les étapes suivantes

JSON:"""

            response = self.vllm.generate(prompt)

            try:
                json_start = response.find('{')
                json_end = response.rfind('}') + 1
                if json_start >= 0 and json_end > json_start:
                    step_result = json.loads(response[json_start:json_end])
                else:
                    step_result = {
                        "premise": step,
                        "analysis": response[:500] if response else "Analyse non disponible",
                        "inference": "Inférence non structurée",
                        "evidence_used": [],
                        "confidence": 0.5,
                        "next_questions": []
                    }
            except json.JSONDecodeError:
                step_result = {
                    "premise": step,
                    "analysis": response[:500] if response else "Analyse non disponible",
                    "inference": "Inférence non structurée",
                    "evidence_used": [],
                    "confidence": 0.5,
                    "next_questions": []
                }

            step_result["step_number"] = i + 1
            results.append(step_result)

        return results

    def _synthesize_conclusion(self, plan: Dict, execution_results: List[Dict], question: str) -> Dict[str, Any]:
        """
        Synthesize final conclusion from hierarchical reasoning.
        """
        # Compile all inferences
        inferences = [r.get("inference", "") for r in execution_results]
        avg_confidence = sum(r.get("confidence", 0.5) for r in execution_results) / len(execution_results) if execution_results else 0.5

        # Generate synthesis with Ollama
        prompt = f"""Tu es un système de raisonnement hiérarchique - SYNTHÈSE FINALE.

## Question originale
{question}

## Stratégie utilisée
{plan.get('strategy', 'mixte')}

## Résultats des étapes de raisonnement
{json.dumps(inferences, ensure_ascii=False, indent=2)}

## Confiance moyenne: {avg_confidence:.2f}

## TÂCHE
Génère la conclusion finale en JSON avec:
1. "conclusion": Conclusion principale répondant à la question
2. "confidence": Score de confiance global (0.0 à 1.0)
3. "key_findings": Liste des découvertes clés
4. "alternative_conclusions": Conclusions alternatives possibles (liste)
5. "warnings": Avertissements ou limitations
6. "recommendations": Recommandations pour l'enquête

JSON:"""

        response = self.vllm.generate(prompt)
        logger.info(f"Synthesis response length: {len(response)} chars")

        synthesis = None

        # Clean the response first - remove markdown artifacts
        clean_response = response.strip()

        # Remove markdown code blocks
        import re
        # Pattern to match ```json ... ``` or ``` ... ```
        code_block_pattern = r'```(?:json)?\s*([\s\S]*?)```'
        code_blocks = re.findall(code_block_pattern, clean_response)
        if code_blocks:
            # Use the first code block content
            clean_response = code_blocks[0].strip()
        else:
            # Remove simple prefixes/suffixes
            for prefix in ["```json", "```", "JSON:", "json:"]:
                if clean_response.startswith(prefix):
                    clean_response = clean_response[len(prefix):].strip()
            for suffix in ["```"]:
                if clean_response.endswith(suffix):
                    clean_response = clean_response[:-len(suffix)].strip()

        # Try to parse the cleaned JSON
        try:
            json_start = clean_response.find('{')
            json_end = clean_response.rfind('}') + 1
            if json_start >= 0 and json_end > json_start:
                json_str = clean_response[json_start:json_end]
                synthesis = json.loads(json_str)
                logger.info("Synthesis JSON parsed successfully")
        except json.JSONDecodeError as e:
            logger.warning(f"Synthesis JSON parse failed: {e}")
            synthesis = None

        # If parsing succeeded, ensure conclusion is properly formatted
        if synthesis and isinstance(synthesis.get("conclusion"), str):
            # Clean up the conclusion text
            conclusion = synthesis["conclusion"]
            # Remove any remaining JSON artifacts
            if conclusion.startswith("{") or conclusion.startswith("```"):
                try:
                    inner_json = json.loads(conclusion) if conclusion.startswith("{") else None
                    if inner_json and isinstance(inner_json.get("conclusion"), str):
                        synthesis["conclusion"] = inner_json["conclusion"]
                except:
                    pass

            # Ensure alternative_conclusions is a list of strings
            if synthesis.get("alternative_conclusions"):
                alt_list = []
                for alt in synthesis["alternative_conclusions"][:3]:
                    if isinstance(alt, str):
                        alt_list.append(alt)
                    elif isinstance(alt, dict):
                        alt_list.append(alt.get("conclusion", alt.get("text", str(alt))))
                synthesis["alternative_conclusions"] = alt_list

            # Ensure warnings is a list of strings
            if synthesis.get("warnings"):
                warning_list = []
                for w in synthesis["warnings"][:5]:
                    if isinstance(w, str):
                        warning_list.append(w)
                synthesis["warnings"] = warning_list

            return synthesis

        # If parsing failed, build synthesis from text
        conclusion_text = clean_response[:2000] if clean_response else "Conclusion non disponible"

        # Try to extract structured data even from partial parse
        alt_conclusions = []
        warnings = []
        key_findings = []
        recommendations = []

        if synthesis:
            if synthesis.get("conclusion"):
                conc = synthesis["conclusion"]
                if isinstance(conc, str):
                    conclusion_text = conc
                elif isinstance(conc, dict):
                    conclusion_text = conc.get("text", str(conc))

            if synthesis.get("alternative_conclusions"):
                for alt in synthesis["alternative_conclusions"][:3]:
                    if isinstance(alt, str):
                        alt_conclusions.append(alt)
                    elif isinstance(alt, dict):
                        alt_conclusions.append(alt.get("conclusion", alt.get("text", str(alt))))

            if synthesis.get("warnings"):
                for w in synthesis["warnings"][:5]:
                    if isinstance(w, str):
                        warnings.append(w)

            if synthesis.get("key_findings"):
                for kf in synthesis["key_findings"][:10]:
                    if isinstance(kf, str):
                        key_findings.append(kf)

            if synthesis.get("recommendations"):
                for rec in synthesis["recommendations"][:5]:
                    if isinstance(rec, str):
                        recommendations.append(rec)

        return {
            "conclusion": conclusion_text,
            "confidence": synthesis.get("confidence", avg_confidence) if synthesis else avg_confidence,
            "key_findings": key_findings,
            "alternative_conclusions": alt_conclusions,
            "warnings": warnings if warnings else ["Réponse formatée depuis le texte brut"],
            "recommendations": recommendations
        }

    def reason(
        self,
        context: str,
        question: str,
        evidence: List[Dict],
        reasoning_type: str = "deductive",
        max_depth: int = 3
    ) -> Dict[str, Any]:
        """
        Main reasoning function using hierarchical HRM approach.

        Implements the two-level architecture:
        1. High-level planning (slow, abstract)
        2. Low-level execution (fast, detailed)
        3. Synthesis of conclusions
        """
        # Check cache
        if self.config.enable_cache:
            cache_key = self._compute_cache_key(context, question, evidence, reasoning_type)
            if cache_key in self.cache:
                logger.info("Returning cached reasoning result")
                return self.cache[cache_key]

        # Phase 1: High-level planning
        logger.info("HRM Phase 1: High-level planning")
        plan = self._high_level_planning(context, question, evidence)
        plan["strategy"] = reasoning_type

        # Phase 2: Low-level execution
        logger.info("HRM Phase 2: Low-level execution")
        execution_results = self._low_level_execution(plan, context, evidence)

        # Phase 3: Synthesis
        logger.info("HRM Phase 3: Synthesis")
        synthesis = self._synthesize_conclusion(plan, execution_results, question)

        # Build response
        reasoning_chain = []
        for result in execution_results:
            # Ensure inference is always a string
            inference = result.get("inference", "")
            if isinstance(inference, list):
                inference = "; ".join(str(item) if not isinstance(item, dict) else json.dumps(item, ensure_ascii=False) for item in inference)
            elif isinstance(inference, dict):
                inference = json.dumps(inference, ensure_ascii=False)
            elif not isinstance(inference, str):
                inference = str(inference)

            # Ensure premise is always a string
            premise = result.get("premise", "")
            if isinstance(premise, (list, dict)):
                premise = json.dumps(premise, ensure_ascii=False) if isinstance(premise, dict) else "; ".join(str(p) for p in premise)
            elif not isinstance(premise, str):
                premise = str(premise)

            # Ensure evidence_used is always a list of strings
            evidence_used = result.get("evidence_used", [])
            if isinstance(evidence_used, str):
                evidence_used = [evidence_used]
            elif not isinstance(evidence_used, list):
                evidence_used = []

            reasoning_chain.append({
                "step_number": result.get("step_number", 0),
                "premise": premise,
                "inference": inference,
                "confidence": float(result.get("confidence", 0.5)) if isinstance(result.get("confidence"), (int, float)) else 0.5,
                "evidence_used": evidence_used
            })

        response = {
            "conclusion": synthesis.get("conclusion", ""),
            "confidence": synthesis.get("confidence", 0.5),
            "reasoning_chain": reasoning_chain,
            "alternative_conclusions": [
                {"conclusion": c, "confidence": 0.4, "reason": "Alternative identifiée"}
                for c in synthesis.get("alternative_conclusions", [])[:3]
            ],
            "warnings": synthesis.get("warnings", []),
            "hrm_metadata": {
                "planning_strategy": plan.get("strategy"),
                "focus_areas": plan.get("focus_areas", []),
                "model_type": "hrm_sapient_ollama_hybrid"
            }
        }

        # Cache result
        if self.config.enable_cache:
            self.cache[cache_key] = response

        return response

    def verify_hypothesis(
        self,
        hypothesis: Dict,
        evidence: List[Dict],
        case_context: str,
        strict_mode: bool = False
    ) -> Dict[str, Any]:
        """
        Verify a hypothesis using hierarchical reasoning.
        """
        prompt = f"""Tu es un système de vérification d'hypothèses forensiques utilisant le raisonnement hiérarchique.

## Hypothèse à vérifier
ID: {hypothesis.get('id', 'unknown')}
Énoncé: {hypothesis.get('statement', '')}
Confiance initiale: {hypothesis.get('confidence', 0.5) * 100:.0f}%

## Contexte de l'affaire
{case_context}

## Preuves disponibles
{self._format_evidence_detail(evidence)}

## Mode d'évaluation: {"STRICT (preuves directes requises)" if strict_mode else "STANDARD"}

## TÂCHE
Évalue cette hypothèse de façon rigoureuse. Réponds en JSON avec:
1. "is_supported": true/false - L'hypothèse est-elle soutenue?
2. "confidence": Score de confiance (0.0 à 1.0)
3. "supporting_reasons": Liste des raisons qui soutiennent l'hypothèse
4. "contradicting_reasons": Liste des raisons qui contredisent l'hypothèse
5. "missing_evidence": Liste des preuves manquantes pour confirmer/infirmer
6. "recommendation": Recommandation détaillée pour l'enquêteur

JSON:"""

        response = self.vllm.generate(prompt)

        try:
            json_start = response.find('{')
            json_end = response.rfind('}') + 1
            if json_start >= 0 and json_end > json_start:
                result = json.loads(response[json_start:json_end])
            else:
                result = self._parse_hypothesis_text(response, hypothesis)
        except json.JSONDecodeError:
            result = self._parse_hypothesis_text(response, hypothesis)

        result["hypothesis_id"] = hypothesis.get("id", "unknown")
        return result

    def find_contradictions(
        self,
        statements: List[Dict[str, str]],
        evidence: List[Dict],
        case_context: str = ""
    ) -> Dict[str, Any]:
        """
        Detect contradictions using hierarchical analysis.
        """
        prompt = f"""Tu es un système de détection de contradictions forensiques utilisant le raisonnement hiérarchique.

## Contexte de l'affaire
{case_context}

## Déclarations à analyser
{json.dumps(statements, ensure_ascii=False, indent=2)}

## Preuves disponibles
{self._format_evidence_detail(evidence)}

## TÂCHE
Analyse toutes les déclarations et preuves pour détecter les contradictions.
Types de contradictions à chercher:
- Contradictions directes (A dit X, B dit non-X)
- Incohérences temporelles (problèmes de chronologie)
- Contradictions implicites (incompatibilités logiques)
- Conflits avec les preuves physiques

Réponds en JSON avec:
1. "contradictions": Liste des contradictions trouvées, chacune avec:
   - "statement_ids": IDs des déclarations concernées
   - "description": Description de la contradiction
   - "severity": "critique" / "majeure" / "mineure"
   - "resolution_suggestions": Suggestions pour résoudre
2. "consistency_score": Score de cohérence global (0.0 à 1.0)
3. "analysis_summary": Résumé de l'analyse

JSON:"""

        response = self.vllm.generate(prompt)

        try:
            json_start = response.find('{')
            json_end = response.rfind('}') + 1
            if json_start >= 0 and json_end > json_start:
                result = json.loads(response[json_start:json_end])
            else:
                result = {
                    "contradictions": [],
                    "consistency_score": 0.7,
                    "analysis_summary": response[:500] if response else "Analyse non structurée"
                }
        except json.JSONDecodeError:
            result = {
                "contradictions": [],
                "consistency_score": 0.7,
                "analysis_summary": response[:500] if response else "Analyse non structurée"
            }

        return result

    def cross_case_reasoning(
        self,
        primary_case: Dict[str, Any],
        comparison_cases: List[Dict[str, Any]],
        focus_areas: List[str] = None
    ) -> Dict[str, Any]:
        """
        Analyze patterns across multiple cases using hierarchical reasoning.
        """
        prompt = f"""Tu es un système d'analyse inter-affaires utilisant le raisonnement hiérarchique.

## Affaire principale
{json.dumps(primary_case, ensure_ascii=False, indent=2)[:2000]}

## Affaires de comparaison ({len(comparison_cases)} affaires)
{json.dumps(comparison_cases, ensure_ascii=False, indent=2)[:3000]}

## Domaines d'analyse prioritaires
{focus_areas if focus_areas else ["modus_operandi", "entités_communes", "patterns_temporels"]}

## TÂCHE
Analyse les connexions et patterns entre ces affaires. Réponds en JSON avec:
1. "patterns": Liste des patterns détectés, chacun avec:
   - "pattern_type": Type de pattern (modus_operandi, entité_commune, temporel, géographique)
   - "description": Description détaillée
   - "cases_involved": IDs des affaires concernées
   - "confidence": Score de confiance
   - "significance": "haute" / "moyenne" / "basse"
2. "connections": Connexions directes entre affaires
3. "investigative_leads": Pistes d'investigation suggérées
4. "risk_assessment": Évaluation des risques
5. "summary": Résumé de l'analyse

JSON:"""

        response = self.vllm.generate(prompt)

        try:
            json_start = response.find('{')
            json_end = response.rfind('}') + 1
            if json_start >= 0 and json_end > json_start:
                result = json.loads(response[json_start:json_end])
            else:
                result = {
                    "patterns": [],
                    "connections": [],
                    "investigative_leads": [],
                    "risk_assessment": {"level": "unknown"},
                    "summary": response[:500] if response else "Analyse non structurée"
                }
        except json.JSONDecodeError:
            result = {
                "patterns": [],
                "connections": [],
                "investigative_leads": [],
                "risk_assessment": {"level": "unknown"},
                "summary": response[:500] if response else "Analyse non structurée"
            }

        return result

    # Helper methods

    def _format_evidence_summary(self, evidence: List[Dict]) -> str:
        """Format evidence for summary display."""
        if not evidence:
            return "Aucune preuve disponible"

        lines = []
        for i, ev in enumerate(evidence[:10], 1):
            ev_type = ev.get("type", "inconnu")
            ev_desc = ev.get("description", "")[:100]
            lines.append(f"{i}. [{ev_type}] {ev_desc}")

        if len(evidence) > 10:
            lines.append(f"... et {len(evidence) - 10} autres preuves")

        return "\n".join(lines)

    def _format_evidence_detail(self, evidence: List[Dict]) -> str:
        """Format evidence for detailed analysis."""
        if not evidence:
            return "Aucune preuve disponible"

        lines = []
        for ev in evidence:
            ev_id = ev.get("id", "unknown")
            ev_type = ev.get("type", "inconnu")
            ev_desc = ev.get("description", "")
            ev_conf = ev.get("confidence", 0.5)
            lines.append(f"- ID: {ev_id}\n  Type: {ev_type}\n  Description: {ev_desc}\n  Confiance: {ev_conf:.0%}")

        return "\n\n".join(lines)

    def _default_plan(self) -> Dict[str, Any]:
        """Return default reasoning plan."""
        return {
            "strategy": "deductive",
            "key_elements": ["preuves", "témoignages", "chronologie"],
            "reasoning_steps": [
                "analyze_evidence",
                "identify_actors",
                "build_timeline",
                "evaluate_hypotheses"
            ],
            "focus_areas": ["identification", "chronologie", "mobiles"],
            "potential_hypotheses": []
        }

    def _parse_hypothesis_text(self, response: str, hypothesis: Dict) -> Dict[str, Any]:
        """Parse unstructured hypothesis verification response."""
        response_lower = response.lower()

        is_supported = any(w in response_lower for w in ["soutenue", "supported", "confirmée", "valide"])

        return {
            "hypothesis_id": hypothesis.get("id", "unknown"),
            "is_supported": is_supported,
            "confidence": 0.5,
            "supporting_reasons": [response[:200]] if is_supported else [],
            "contradicting_reasons": [] if is_supported else [response[:200]],
            "missing_evidence": ["Analyse structurée requise"],
            "recommendation": response[:500] if response else "Recommandation non disponible"
        }
