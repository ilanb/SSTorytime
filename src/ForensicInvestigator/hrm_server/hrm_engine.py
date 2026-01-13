"""
HRM (Hypothetical Reasoning Model) Engine.

This module implements the core reasoning logic inspired by the Hierarchical Reasoning Model.
It uses a two-level architecture:
1. High-level planning: Strategic reasoning about what needs to be analyzed
2. Low-level computation: Detailed inference and pattern matching

For production use, this can be extended to use actual neural network checkpoints.
"""
import json
import hashlib
from typing import List, Dict, Any, Optional, Tuple
from dataclasses import dataclass
from enum import Enum
import re


class InferenceType(Enum):
    DIRECT = "direct"
    INDIRECT = "indirect"
    ANALOGICAL = "analogical"
    TEMPORAL = "temporal"
    CAUSAL = "causal"


@dataclass
class InferenceResult:
    """Result of a single inference."""
    conclusion: str
    confidence: float
    inference_type: InferenceType
    premises: List[str]
    evidence_ids: List[str]


class HRMEngine:
    """
    Hypothetical Reasoning Model Engine.

    Implements hierarchical reasoning for forensic investigation:
    - Evidence analysis and pattern detection
    - Hypothesis generation and verification
    - Contradiction detection
    - Cross-case pattern matching
    """

    def __init__(self):
        self.knowledge_base: Dict[str, Any] = {}
        self.inference_cache: Dict[str, InferenceResult] = {}

        # Forensic domain knowledge patterns
        self.domain_patterns = {
            "temporal": [
                r"(\d{1,2}[/:]\d{2})",  # Time patterns
                r"(\d{1,2}/\d{1,2}/\d{2,4})",  # Date patterns
                r"(avant|après|pendant|lors de)",  # French temporal markers
                r"(before|after|during|when)",  # English temporal markers
            ],
            "causal": [
                r"(parce que|car|donc|ainsi)",  # French causal markers
                r"(because|therefore|thus|hence)",  # English causal markers
                r"(causé par|résulte de|provoque)",  # French causal verbs
                r"(caused by|results in|leads to)",  # English causal verbs
            ],
            "evidential": [
                r"(ADN|empreinte|trace|résidu)",  # French evidence types
                r"(DNA|fingerprint|trace|residue)",  # English evidence types
                r"(témoin|témoignage|déclaration)",  # Witness-related
                r"(witness|testimony|statement)",
            ],
            "suspect": [
                r"(suspect|accusé|inculpé)",  # French suspect terms
                r"(suspect|accused|defendant)",  # English suspect terms
                r"(alibi|mobile|opportunité)",  # Motive/opportunity
                r"(motive|opportunity|means)",
            ],
        }

    def _compute_hash(self, *args) -> str:
        """Compute hash for caching."""
        content = json.dumps(args, sort_keys=True, default=str)
        return hashlib.md5(content.encode()).hexdigest()

    def _extract_patterns(self, text: str) -> Dict[str, List[str]]:
        """Extract forensic patterns from text."""
        found_patterns = {}
        for pattern_type, patterns in self.domain_patterns.items():
            matches = []
            for pattern in patterns:
                matches.extend(re.findall(pattern, text, re.IGNORECASE))
            if matches:
                found_patterns[pattern_type] = list(set(matches))
        return found_patterns

    def _high_level_plan(self, context: str, question: str, evidence: List[Dict]) -> List[str]:
        """
        High-level planning phase.
        Determines the reasoning strategy and steps needed.
        """
        plan = []

        # Analyze question type
        question_lower = question.lower()

        if any(w in question_lower for w in ["qui", "who", "suspect"]):
            plan.append("identify_actors")
            plan.append("analyze_relationships")
            plan.append("evaluate_motives")

        if any(w in question_lower for w in ["quand", "when", "heure", "time"]):
            plan.append("build_timeline")
            plan.append("identify_temporal_gaps")

        if any(w in question_lower for w in ["comment", "how", "méthode", "method"]):
            plan.append("analyze_method")
            plan.append("identify_tools")
            plan.append("trace_sequence")

        if any(w in question_lower for w in ["pourquoi", "why", "motif", "motive"]):
            plan.append("identify_motives")
            plan.append("analyze_benefits")
            plan.append("evaluate_psychology")

        if any(w in question_lower for w in ["où", "where", "lieu", "location"]):
            plan.append("map_locations")
            plan.append("analyze_movement")

        # Default reasoning steps
        if not plan:
            plan = ["analyze_evidence", "identify_patterns", "generate_hypotheses"]

        plan.append("synthesize_conclusions")
        return plan

    def _low_level_compute(self, step: str, context: str, evidence: List[Dict]) -> Dict[str, Any]:
        """
        Low-level computation phase.
        Executes specific reasoning operations.
        Returns human-readable findings in French.
        """
        result = {
            "step": step,
            "findings": "",
            "details": [],
            "confidence": 0.5
        }

        if step == "identify_actors":
            # Extract person mentions from context and evidence
            suspects = []
            witnesses = []
            victims = []
            other_actors = []

            for ev in evidence:
                desc = ev.get("description", "")
                ev_type = ev.get("type", "").lower()

                # Extract names (simple heuristic: capitalized words of 3+ chars)
                words = desc.split()
                for i, word in enumerate(words):
                    clean_word = word.strip(".,;:()\"'")
                    if clean_word and clean_word[0].isupper() and len(clean_word) > 2:
                        # Check context for role
                        context_window = " ".join(words[max(0,i-3):min(len(words),i+4)]).lower()
                        if any(s in context_window for s in ["suspect", "accusé", "inculpé", "coupable"]):
                            if clean_word not in suspects:
                                suspects.append(clean_word)
                        elif any(s in context_window for s in ["témoin", "witness", "déclare", "affirme"]):
                            if clean_word not in witnesses:
                                witnesses.append(clean_word)
                        elif any(s in context_window for s in ["victime", "victim", "décédé"]):
                            if clean_word not in victims:
                                victims.append(clean_word)
                        elif clean_word not in other_actors and clean_word not in ["Le", "La", "Les", "Un", "Une", "Des"]:
                            other_actors.append(clean_word)

            # Build human-readable finding
            findings_parts = []
            if suspects:
                findings_parts.append(f"Suspects identifiés : {', '.join(suspects[:5])}")
            if witnesses:
                findings_parts.append(f"Témoins mentionnés : {', '.join(witnesses[:5])}")
            if victims:
                findings_parts.append(f"Victimes : {', '.join(victims[:3])}")
            if other_actors and not suspects:
                findings_parts.append(f"Personnes d'intérêt : {', '.join(other_actors[:5])}")

            if findings_parts:
                result["findings"] = " | ".join(findings_parts)
                result["details"] = suspects + witnesses + victims
                result["confidence"] = 0.7
            else:
                result["findings"] = "Aucun acteur clairement identifié dans les preuves."
                result["confidence"] = 0.3

        elif step == "analyze_relationships":
            # Analyze relationships between actors
            relationships = []
            keywords_relations = {
                "complice": "complicité",
                "associé": "association",
                "employé": "emploi",
                "famille": "lien familial",
                "ami": "amitié",
                "contact": "contact",
                "transaction": "transaction financière",
                "transfert": "transfert",
                "rencontre": "rencontre"
            }

            for ev in evidence:
                desc = ev.get("description", "").lower()
                for keyword, relation_type in keywords_relations.items():
                    if keyword in desc:
                        relationships.append(relation_type)

            if relationships:
                unique_relations = list(set(relationships))
                result["findings"] = f"Relations détectées : {', '.join(unique_relations[:5])}"
                result["details"] = unique_relations
                result["confidence"] = 0.65
            else:
                result["findings"] = "Pas de relations explicites détectées entre les acteurs."
                result["confidence"] = 0.4

        elif step == "evaluate_motives":
            # Look for motives in evidence
            motives = []
            motive_keywords = {
                "argent": "motivation financière",
                "dette": "difficultés financières",
                "héritage": "intérêt successoral",
                "vengeance": "vengeance",
                "jalousie": "jalousie",
                "conflit": "conflit personnel",
                "fraude": "fraude",
                "assurance": "fraude à l'assurance",
                "million": "enjeu financier important"
            }

            for ev in evidence:
                desc = ev.get("description", "").lower()
                for keyword, motive in motive_keywords.items():
                    if keyword in desc and motive not in motives:
                        motives.append(motive)

            if motives:
                result["findings"] = f"Mobiles potentiels identifiés : {', '.join(motives)}"
                result["details"] = motives
                result["confidence"] = 0.7
            else:
                result["findings"] = "Aucun mobile évident détecté dans les preuves disponibles."
                result["confidence"] = 0.4

        elif step == "build_timeline":
            # Extract temporal information
            events = []
            for ev in evidence:
                desc = ev.get("description", "")
                patterns = self._extract_patterns(desc)
                if "temporal" in patterns:
                    ev_name = desc[:50] + "..." if len(desc) > 50 else desc
                    events.append(f"« {ev_name} » ({', '.join(patterns['temporal'][:2])})")

            if events:
                result["findings"] = f"{len(events)} événements temporels identifiés"
                result["details"] = events[:5]
                result["confidence"] = 0.75
            else:
                result["findings"] = "Pas de marqueurs temporels clairs dans les preuves."
                result["confidence"] = 0.4

        elif step == "analyze_evidence":
            # Categorize evidence by type
            categorized = {}
            for ev in evidence:
                ev_type = ev.get("type", "autre")
                if ev_type not in categorized:
                    categorized[ev_type] = 0
                categorized[ev_type] += 1

            summary_parts = [f"{count} preuve(s) {etype}" for etype, count in categorized.items()]
            result["findings"] = f"Preuves analysées : {', '.join(summary_parts)}"
            result["details"] = list(categorized.keys())
            result["confidence"] = 0.6

        elif step == "analyze_method":
            # Look for method/modus operandi
            methods = []
            method_keywords = {
                "poison": "empoisonnement",
                "arme": "arme",
                "incendie": "incendie criminel",
                "effraction": "effraction",
                "vol": "vol",
                "faux": "falsification",
                "piratage": "cybercriminalité",
                "transfert": "transfert frauduleux"
            }

            for ev in evidence:
                desc = ev.get("description", "").lower()
                for keyword, method in method_keywords.items():
                    if keyword in desc and method not in methods:
                        methods.append(method)

            if methods:
                result["findings"] = f"Méthodes/modes opératoires détectés : {', '.join(methods)}"
                result["details"] = methods
                result["confidence"] = 0.7
            else:
                result["findings"] = "Mode opératoire non clairement identifié."
                result["confidence"] = 0.4

        elif step == "identify_patterns":
            # Find patterns across evidence
            pattern_summary = []
            all_patterns = {}
            for ev in evidence:
                patterns = self._extract_patterns(ev.get("description", ""))
                for ptype, pmatches in patterns.items():
                    if ptype not in all_patterns:
                        all_patterns[ptype] = []
                    all_patterns[ptype].extend(pmatches)

            pattern_names = {
                "temporal": "temporels",
                "causal": "causaux",
                "evidential": "probatoires",
                "suspect": "suspects"
            }

            for ptype, matches in all_patterns.items():
                if matches:
                    pname = pattern_names.get(ptype, ptype)
                    pattern_summary.append(f"{len(set(matches))} éléments {pname}")

            if pattern_summary:
                result["findings"] = f"Patterns détectés : {', '.join(pattern_summary)}"
                result["details"] = list(all_patterns.keys())
                result["confidence"] = 0.65
            else:
                result["findings"] = "Aucun pattern récurrent détecté."
                result["confidence"] = 0.4

        elif step == "synthesize_conclusions":
            result["findings"] = "Synthèse des analyses effectuée."
            result["confidence"] = 0.7

        else:
            step_names = {
                "identify_tools": "Identification des outils/moyens",
                "trace_sequence": "Reconstitution de la séquence",
                "map_locations": "Cartographie des lieux",
                "analyze_movement": "Analyse des déplacements",
                "identify_motives": "Recherche des mobiles",
                "analyze_benefits": "Analyse des bénéficiaires",
                "evaluate_psychology": "Évaluation psychologique",
                "generate_hypotheses": "Génération d'hypothèses"
            }
            step_name = step_names.get(step, step)
            result["findings"] = f"Étape « {step_name} » exécutée - analyse complémentaire recommandée."
            result["confidence"] = 0.5

        return result

    def reason(
        self,
        context: str,
        question: str,
        evidence: List[Dict],
        reasoning_type: str = "deductive",
        max_depth: int = 3
    ) -> Dict[str, Any]:
        """
        Main reasoning function using hierarchical approach.
        """
        # Check cache
        cache_key = self._compute_hash(context, question, evidence, reasoning_type)
        if cache_key in self.inference_cache:
            return self.inference_cache[cache_key]

        # Phase 1: High-level planning
        plan = self._high_level_plan(context, question, evidence)

        # Phase 2: Low-level computation for each step
        reasoning_chain = []
        all_findings = []

        # French step names for display
        step_display_names = {
            "identify_actors": "Identification des acteurs",
            "analyze_relationships": "Analyse des relations",
            "evaluate_motives": "Évaluation des mobiles",
            "build_timeline": "Construction de la chronologie",
            "identify_temporal_gaps": "Identification des lacunes temporelles",
            "analyze_method": "Analyse du mode opératoire",
            "identify_tools": "Identification des outils",
            "trace_sequence": "Reconstitution de la séquence",
            "map_locations": "Cartographie des lieux",
            "analyze_movement": "Analyse des déplacements",
            "identify_motives": "Identification des mobiles",
            "analyze_benefits": "Analyse des bénéficiaires",
            "evaluate_psychology": "Profil psychologique",
            "analyze_evidence": "Analyse des preuves",
            "identify_patterns": "Identification des patterns",
            "generate_hypotheses": "Génération d'hypothèses",
            "synthesize_conclusions": "Synthèse des conclusions"
        }

        for i, step in enumerate(plan[:max_depth]):
            computation = self._low_level_compute(step, context, evidence)

            # Get display name for step
            step_name = step_display_names.get(step, step.replace("_", " ").capitalize())

            reasoning_step = {
                "step_number": i + 1,
                "premise": step_name,
                "inference": computation["findings"],
                "confidence": computation["confidence"],
                "evidence_used": [e.get("description", e.get("id", ""))[:40] for e in evidence[:3]]
            }
            reasoning_chain.append(reasoning_step)
            all_findings.append(computation)

        # Generate conclusion
        avg_confidence = sum(f["confidence"] for f in all_findings) / len(all_findings) if all_findings else 0.5

        conclusion = self._generate_conclusion(question, all_findings, reasoning_type)

        result = {
            "conclusion": conclusion,
            "confidence": avg_confidence,
            "reasoning_chain": reasoning_chain,
            "alternative_conclusions": [],
            "warnings": []
        }

        # Add warnings for low confidence
        if avg_confidence < 0.5:
            result["warnings"].append("Confiance faible - preuves supplémentaires recommandées")

        # Cache result
        self.inference_cache[cache_key] = result
        return result

    def _generate_conclusion(
        self,
        question: str,
        findings: List[Dict],
        reasoning_type: str
    ) -> str:
        """Generate a conclusion based on findings."""
        if not findings:
            return "Preuves insuffisantes pour établir une conclusion."

        # Collect all findings text
        findings_texts = []
        details = []
        for f in findings:
            if f.get("findings"):
                findings_texts.append(f["findings"])
            if f.get("details"):
                details.extend(f["details"])

        if not findings_texts:
            return "Analyse terminée mais aucun pattern significatif détecté."

        # Analyze question type for targeted conclusion
        question_lower = question.lower()

        # Build conclusion based on question type and findings
        conclusion_parts = []

        # If asking about suspect/who
        if any(w in question_lower for w in ["qui", "who", "suspect", "responsable", "coupable"]):
            # Look for actors in findings
            actors_finding = None
            motives_finding = None
            for f in findings:
                if "acteur" in f.get("findings", "").lower() or "suspect" in f.get("findings", "").lower():
                    actors_finding = f.get("findings", "")
                if "mobile" in f.get("findings", "").lower() or "motif" in f.get("findings", "").lower():
                    motives_finding = f.get("findings", "")

            if actors_finding:
                conclusion_parts.append(f"**Acteurs identifiés** : {actors_finding}")
            if motives_finding:
                conclusion_parts.append(f"**Mobiles détectés** : {motives_finding}")

            if not conclusion_parts:
                conclusion_parts.append("L'analyse n'a pas permis d'identifier clairement un suspect principal.")
                conclusion_parts.append("Recommandation : approfondir l'analyse des relations et des mobiles.")

        # If asking about when/timeline
        elif any(w in question_lower for w in ["quand", "when", "heure", "moment", "chronologie"]):
            for f in findings:
                if "temporel" in f.get("findings", "").lower() or "chronologie" in f.get("findings", "").lower():
                    conclusion_parts.append(f.get("findings", ""))

        # If asking about how/method
        elif any(w in question_lower for w in ["comment", "how", "méthode", "moyen"]):
            for f in findings:
                if "méthode" in f.get("findings", "").lower() or "opératoire" in f.get("findings", "").lower():
                    conclusion_parts.append(f.get("findings", ""))

        # If asking about why/motive
        elif any(w in question_lower for w in ["pourquoi", "why", "motif", "raison"]):
            for f in findings:
                if "mobile" in f.get("findings", "").lower() or "motif" in f.get("findings", "").lower():
                    conclusion_parts.append(f.get("findings", ""))

        # Default: summarize all findings
        if not conclusion_parts:
            conclusion_parts = [f"• {text}" for text in findings_texts if text]

        # Add reasoning type context
        reasoning_intro = {
            "deductive": "**Conclusion (raisonnement déductif)** :",
            "inductive": "**Conclusion (raisonnement inductif)** :",
            "abductive": "**Conclusion (meilleure explication)** :",
            "analogical": "**Conclusion (raisonnement analogique)** :"
        }

        intro = reasoning_intro.get(reasoning_type, "**Conclusion** :")

        if conclusion_parts:
            return f"{intro}\n\n" + "\n\n".join(conclusion_parts)
        else:
            return f"{intro}\n\nL'analyse de {len(findings_texts)} étapes n'a pas permis d'établir une conclusion définitive. Des preuves supplémentaires sont recommandées."

    def verify_hypothesis(
        self,
        hypothesis: Dict,
        evidence: List[Dict],
        case_context: str,
        strict_mode: bool = False
    ) -> Dict[str, Any]:
        """
        Verify a hypothesis against available evidence.
        """
        supporting_reasons = []
        contradicting_reasons = []
        missing_evidence = []

        hypothesis_statement = hypothesis.get("statement", "").lower()
        hypothesis_keywords = set(hypothesis_statement.split())

        # Analyze each piece of evidence
        evidence_coverage = 0
        for ev in evidence:
            ev_desc = ev.get("description", "").lower()
            ev_name = ev.get("description", ev.get("id", "Preuve inconnue"))[:80]  # Use description as name, truncate
            ev_keywords = set(ev_desc.split())

            # Check keyword overlap
            overlap = hypothesis_keywords.intersection(ev_keywords)
            # Filter out common words
            significant_overlap = {w for w in overlap if len(w) > 3}

            if significant_overlap:
                evidence_coverage += 1
                # Check sentiment/direction
                if any(neg in ev_desc for neg in ["non", "pas", "jamais", "aucun", "impossible"]):
                    contradicting_reasons.append(f"« {ev_name} » contient des éléments contradictoires")
                else:
                    keywords_str = ", ".join(list(significant_overlap)[:3])
                    supporting_reasons.append(f"« {ev_name} » - mots-clés communs: {keywords_str}")

        # Calculate confidence
        if not evidence:
            confidence = 0.0
        else:
            support_ratio = len(supporting_reasons) / len(evidence)
            contradict_ratio = len(contradicting_reasons) / len(evidence)
            confidence = max(0, min(1, support_ratio - contradict_ratio + 0.3))

        # Determine support status
        if strict_mode:
            is_supported = confidence > 0.7 and len(contradicting_reasons) == 0
        else:
            is_supported = confidence > 0.5 and len(supporting_reasons) > len(contradicting_reasons)

        # Identify missing evidence
        if confidence < 0.7:
            missing_evidence.append("Des preuves physiques supplémentaires sont nécessaires")
        if "témoin" not in case_context.lower() and "witness" not in case_context.lower():
            missing_evidence.append("Aucun témoignage trouvé dans le contexte")

        # Generate recommendation
        if is_supported and confidence > 0.7:
            recommendation = "L'hypothèse est bien soutenue. Poursuivez l'enquête dans cette direction."
        elif is_supported:
            recommendation = "L'hypothèse est prometteuse mais nécessite des preuves supplémentaires pour confirmation."
        elif confidence > 0.3:
            recommendation = "L'hypothèse est non concluante. Envisagez des explications alternatives."
        else:
            recommendation = "L'hypothèse n'est pas soutenue par les preuves actuelles. Réévaluez les hypothèses de départ."

        return {
            "hypothesis_id": hypothesis.get("id", "unknown"),
            "is_supported": is_supported,
            "confidence": confidence,
            "supporting_reasons": supporting_reasons,
            "contradicting_reasons": contradicting_reasons,
            "missing_evidence": missing_evidence,
            "recommendation": recommendation
        }

    def find_contradictions(
        self,
        statements: List[Dict[str, str]],
        evidence: List[Dict],
        case_context: str = ""
    ) -> Dict[str, Any]:
        """
        Detect contradictions between statements and evidence.
        """
        contradictions = []

        # Compare statements pairwise
        for i, stmt1 in enumerate(statements):
            for j, stmt2 in enumerate(statements[i+1:], i+1):
                contradiction = self._check_contradiction(stmt1, stmt2)
                if contradiction:
                    contradictions.append(contradiction)

        # Check statements against evidence
        for stmt in statements:
            for ev in evidence:
                contradiction = self._check_statement_evidence_conflict(stmt, ev)
                if contradiction:
                    contradictions.append(contradiction)

        # Calculate consistency score
        total_comparisons = (len(statements) * (len(statements) - 1)) / 2 + len(statements) * len(evidence)
        if total_comparisons > 0:
            consistency_score = 1.0 - (len(contradictions) / total_comparisons)
        else:
            consistency_score = 1.0

        # Generate summary
        if not contradictions:
            summary = "Aucune contradiction détectée. Toutes les déclarations semblent cohérentes avec les preuves."
        elif len(contradictions) <= 2:
            summary = f"Incohérences mineures trouvées ({len(contradictions)} contradiction(s)). Révision recommandée."
        else:
            summary = f"Contradictions significatives détectées ({len(contradictions)}). Investigation détaillée requise."

        return {
            "contradictions": contradictions,
            "consistency_score": max(0, consistency_score),
            "analysis_summary": summary
        }

    def _check_contradiction(self, stmt1: Dict, stmt2: Dict) -> Optional[Dict]:
        """Check if two statements contradict each other."""
        text1 = stmt1.get("content", stmt1.get("text", "")).lower()
        text2 = stmt2.get("content", stmt2.get("text", "")).lower()

        # Simple contradiction detection based on negation patterns
        negation_pairs = [
            ("oui", "non"), ("yes", "no"),
            ("présent", "absent"), ("present", "absent"),
            ("avant", "après"), ("before", "after"),
            ("gauche", "droite"), ("left", "right"),
        ]

        for pos, neg in negation_pairs:
            if (pos in text1 and neg in text2) or (neg in text1 and pos in text2):
                return {
                    "statement_ids": [stmt1.get("id", "s1"), stmt2.get("id", "s2")],
                    "description": f"Assertions contradictoires : « {pos} » vs « {neg} »",
                    "severity": "medium",
                    "resolution_suggestions": [
                        "Vérifier la fiabilité de la source",
                        "Vérifier le contexte temporel",
                        "Interroger les témoins séparément"
                    ]
                }

        return None

    def _check_statement_evidence_conflict(self, stmt: Dict, evidence: Dict) -> Optional[Dict]:
        """Check if a statement conflicts with evidence."""
        stmt_text = stmt.get("content", stmt.get("text", "")).lower()
        ev_text = evidence.get("description", "").lower()

        # Check for direct conflicts (simplified)
        if "aucun" in stmt_text or "no " in stmt_text:
            # Statement claims absence
            ev_type = evidence.get("type", "").lower()
            if ev_type in stmt_text:
                return {
                    "statement_ids": [stmt.get("id", "stmt"), evidence.get("id", "ev")],
                    "description": f"La déclaration affirme l'absence de {ev_type} mais une preuve existe",
                    "severity": "high",
                    "resolution_suggestions": [
                        "Réexaminer les preuves physiques",
                        "Vérifier la source de la déclaration",
                        "Considérer le timing de la déclaration par rapport à la collecte des preuves"
                    ]
                }

        return None

    def cross_case_reasoning(
        self,
        primary_case: Dict[str, Any],
        comparison_cases: List[Dict[str, Any]],
        focus_areas: List[str] = None
    ) -> Dict[str, Any]:
        """
        Analyze patterns and connections across multiple cases.
        """
        patterns = []
        connections = []
        investigative_leads = []

        primary_id = primary_case.get("id", "primary")
        primary_type = primary_case.get("type", "unknown")

        # Extract features from primary case
        primary_features = self._extract_case_features(primary_case)

        for comp_case in comparison_cases:
            comp_id = comp_case.get("id", "comparison")
            comp_features = self._extract_case_features(comp_case)

            # Find matching patterns
            common_features = self._find_common_features(primary_features, comp_features)

            if common_features:
                # Create pattern
                pattern = {
                    "pattern_type": "common_features",
                    "description": f"Shared characteristics between cases: {', '.join(common_features[:3])}",
                    "cases_involved": [primary_id, comp_id],
                    "confidence": min(0.9, 0.3 + 0.1 * len(common_features)),
                    "significance": "high" if len(common_features) > 3 else "medium"
                }
                patterns.append(pattern)

                # Create connection
                connection = {
                    "source_case": primary_id,
                    "target_case": comp_id,
                    "connection_type": "pattern_match",
                    "strength": len(common_features) / 10,
                    "details": common_features
                }
                connections.append(connection)

                # Generate leads
                if len(common_features) > 2:
                    investigative_leads.append(
                        f"Enquêter sur la connexion entre {primary_id} et {comp_id} - {len(common_features)} caractéristiques communes"
                    )

        # Risk assessment
        risk_level = "low"
        if len(patterns) > 3:
            risk_level = "high"
            investigative_leads.append("Plusieurs connexions entre affaires suggèrent une activité organisée - recommandation: groupe d'enquête")
        elif len(patterns) > 1:
            risk_level = "medium"

        risk_assessment = {
            "level": risk_level,
            "pattern_count": len(patterns),
            "connection_count": len(connections),
            "recommended_priority": "urgent" if risk_level == "high" else "normal"
        }

        # Generate summary
        if not patterns:
            summary = "Aucun pattern significatif trouvé dans les affaires analysées."
        else:
            summary = f"L'analyse a identifié {len(patterns)} patterns et {len(connections)} connexions sur {len(comparison_cases) + 1} affaires."

        return {
            "patterns": patterns,
            "connections": connections,
            "investigative_leads": investigative_leads,
            "risk_assessment": risk_assessment,
            "summary": summary
        }

    def _extract_case_features(self, case: Dict) -> Dict[str, List[str]]:
        """Extract analyzable features from a case."""
        features = {
            "type": [case.get("type", "unknown")],
            "keywords": [],
            "entities": [],
            "patterns": []
        }

        # Extract from description
        description = case.get("description", "")
        features["keywords"] = [w.lower() for w in description.split() if len(w) > 4][:20]

        # Extract from timeline
        timeline = case.get("timeline", [])
        for event in timeline:
            event_desc = event.get("description", "")
            features["patterns"].extend(list(self._extract_patterns(event_desc).keys()))

        # Extract from evidence
        evidence = case.get("evidence", [])
        for ev in evidence:
            features["entities"].append(ev.get("type", "unknown"))

        return features

    def _find_common_features(self, features1: Dict, features2: Dict) -> List[str]:
        """Find common features between two cases."""
        common = []

        # Check type match
        if features1.get("type") == features2.get("type"):
            common.append(f"same_type:{features1.get('type', [''])[0]}")

        # Check keyword overlap
        keywords1 = set(features1.get("keywords", []))
        keywords2 = set(features2.get("keywords", []))
        keyword_overlap = keywords1.intersection(keywords2)
        if keyword_overlap:
            common.extend([f"keyword:{kw}" for kw in list(keyword_overlap)[:5]])

        # Check entity overlap
        entities1 = set(features1.get("entities", []))
        entities2 = set(features2.get("entities", []))
        entity_overlap = entities1.intersection(entities2)
        if entity_overlap:
            common.extend([f"entity:{e}" for e in entity_overlap])

        # Check pattern overlap
        patterns1 = set(features1.get("patterns", []))
        patterns2 = set(features2.get("patterns", []))
        pattern_overlap = patterns1.intersection(patterns2)
        if pattern_overlap:
            common.extend([f"pattern:{p}" for p in pattern_overlap])

        return common
