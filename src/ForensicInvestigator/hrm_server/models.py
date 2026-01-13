"""
Pydantic models for HRM API requests and responses.
"""
from pydantic import BaseModel, Field
from typing import List, Optional, Dict, Any
from enum import Enum


class ReasoningType(str, Enum):
    """Types of reasoning supported by HRM."""
    DEDUCTIVE = "deductive"
    INDUCTIVE = "inductive"
    ABDUCTIVE = "abductive"
    ANALOGICAL = "analogical"


class Evidence(BaseModel):
    """Evidence item for reasoning."""
    id: str
    type: str
    description: str
    confidence: float = Field(ge=0.0, le=1.0, default=0.5)
    metadata: Optional[Dict[str, Any]] = None


class Hypothesis(BaseModel):
    """Hypothesis to verify."""
    id: str
    statement: str
    supporting_evidence: List[str] = []
    contradicting_evidence: List[str] = []
    confidence: float = Field(ge=0.0, le=1.0, default=0.5)


class ReasoningRequest(BaseModel):
    """Request for general reasoning."""
    context: str
    question: str
    evidence: List[Evidence] = []
    reasoning_type: ReasoningType = ReasoningType.DEDUCTIVE
    max_depth: int = Field(ge=1, le=10, default=3)


class ReasoningStep(BaseModel):
    """A single step in the reasoning chain."""
    step_number: int
    premise: str
    inference: str
    confidence: float = Field(ge=0.0, le=1.0)
    evidence_used: List[str] = []


class ReasoningResponse(BaseModel):
    """Response from reasoning endpoint."""
    conclusion: str
    confidence: float = Field(ge=0.0, le=1.0)
    reasoning_chain: List[ReasoningStep]
    alternative_conclusions: List[Dict[str, Any]] = []
    warnings: List[str] = []


class HypothesisVerificationRequest(BaseModel):
    """Request for hypothesis verification."""
    hypothesis: Hypothesis
    evidence: List[Evidence]
    case_context: str
    strict_mode: bool = False


class HypothesisVerificationResponse(BaseModel):
    """Response from hypothesis verification."""
    hypothesis_id: str
    is_supported: bool
    confidence: float = Field(ge=0.0, le=1.0)
    supporting_reasons: List[str]
    contradicting_reasons: List[str]
    missing_evidence: List[str]
    recommendation: str


class ContradictionRequest(BaseModel):
    """Request for contradiction detection."""
    statements: List[Dict[str, str]]
    evidence: List[Evidence] = []
    case_context: str = ""


class Contradiction(BaseModel):
    """A detected contradiction."""
    statement_ids: List[str]
    description: str
    severity: str = Field(pattern="^(low|medium|high|critical)$")
    resolution_suggestions: List[str]


class ContradictionResponse(BaseModel):
    """Response from contradiction detection."""
    contradictions: List[Contradiction]
    consistency_score: float = Field(ge=0.0, le=1.0)
    analysis_summary: str


class CasePattern(BaseModel):
    """Pattern found across cases."""
    pattern_type: str
    description: str
    cases_involved: List[str]
    confidence: float = Field(ge=0.0, le=1.0)
    significance: str


class CrossCaseRequest(BaseModel):
    """Request for cross-case reasoning."""
    primary_case: Dict[str, Any]
    comparison_cases: List[Dict[str, Any]]
    focus_areas: List[str] = []


class CrossCaseResponse(BaseModel):
    """Response from cross-case reasoning."""
    patterns: List[CasePattern]
    connections: List[Dict[str, Any]]
    investigative_leads: List[str]
    risk_assessment: Dict[str, Any]
    summary: str
