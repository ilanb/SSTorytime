#!/usr/bin/env python3
"""
Evaluation Script for N4L Generator

This module provides comprehensive evaluation metrics for N4L generation:
- N4L syntax validity
- Information coverage
- Relation extraction quality
- BLEU/ROUGE scores

Usage:
    python evaluate.py --model models/n4l-qwen-lora --test-data data/splits/test.jsonl
"""

import json
import re
import logging
from pathlib import Path
from typing import List, Dict, Any, Optional, Tuple
from dataclasses import dataclass, field
from collections import Counter
import torch
from tqdm import tqdm

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


@dataclass
class N4LEvaluationResult:
    """Results from N4L evaluation"""
    # Syntax metrics
    syntax_valid: bool = False
    syntax_errors: List[str] = field(default_factory=list)
    syntax_warnings: List[str] = field(default_factory=list)
    syntax_score: float = 0.0

    # Structure metrics
    num_sections: int = 0
    num_contexts: int = 0
    num_relations: int = 0
    num_entities: int = 0
    num_timeline_events: int = 0

    # Coverage metrics
    entity_coverage: float = 0.0
    relation_coverage: float = 0.0

    # Text similarity
    bleu_score: float = 0.0
    rouge_l_score: float = 0.0

    # Overall
    overall_score: float = 0.0

    def to_dict(self) -> Dict[str, Any]:
        return {
            "syntax": {
                "valid": self.syntax_valid,
                "errors": self.syntax_errors,
                "warnings": self.syntax_warnings,
                "score": self.syntax_score
            },
            "structure": {
                "sections": self.num_sections,
                "contexts": self.num_contexts,
                "relations": self.num_relations,
                "entities": self.num_entities,
                "timeline_events": self.num_timeline_events
            },
            "coverage": {
                "entity": self.entity_coverage,
                "relation": self.relation_coverage
            },
            "similarity": {
                "bleu": self.bleu_score,
                "rouge_l": self.rouge_l_score
            },
            "overall_score": self.overall_score
        }


class N4LSyntaxValidator:
    """Validate N4L syntax"""

    def validate(self, content: str) -> Tuple[bool, List[str], List[str]]:
        """
        Validate N4L content syntax.

        Returns:
            Tuple of (is_valid, errors, warnings)
        """
        errors = []
        warnings = []

        lines = content.split('\n')

        # Check for at least one section marker
        has_section = any(line.strip().startswith('-') and not line.strip().startswith('--')
                         for line in lines if line.strip() and not line.strip().startswith('#'))
        if not has_section:
            warnings.append("No section title found (lines starting with -)")

        # Check context balance (:: pairs)
        context_count = content.count('::')
        if context_count % 2 != 0:
            errors.append(f"Unbalanced contexts: {context_count} '::' markers (should be even)")

        # Check parentheses balance in relations
        for i, line in enumerate(lines, 1):
            stripped = line.strip()

            # Skip comments and empty lines
            if not stripped or stripped.startswith('#') or stripped.startswith('//'):
                continue

            # Check parentheses
            if '(' in stripped:
                open_count = stripped.count('(')
                close_count = stripped.count(')')
                if open_count != close_count:
                    errors.append(f"Line {i}: Unbalanced parentheses ({open_count} '(' vs {close_count} ')')")

            # Check for valid relation format
            if '(' in stripped and ')' in stripped:
                # Should have format: Subject (relation) Object
                match = re.search(r'\(([^)]+)\)', stripped)
                if match:
                    relation = match.group(1).strip()
                    if not relation:
                        errors.append(f"Line {i}: Empty relation '()'")

        # Check ditto usage
        prev_has_subject = False
        for i, line in enumerate(lines, 1):
            stripped = line.strip()

            if stripped.startswith('"') and '(' in stripped:
                if not prev_has_subject:
                    warnings.append(f"Line {i}: Ditto '\"' without preceding subject")

            elif stripped and not stripped.startswith('#') and not stripped.startswith('//'):
                if re.match(r'^[^"#/(\s]', stripped):
                    prev_has_subject = True

        # Check timeline blocks
        timeline_start = content.count('+:: _timeline_')
        timeline_end = content.count('-:: _timeline_')
        if timeline_start != timeline_end:
            errors.append(f"Unbalanced timeline blocks: {timeline_start} starts, {timeline_end} ends")

        is_valid = len(errors) == 0
        return is_valid, errors, warnings


class N4LStructureExtractor:
    """Extract structural elements from N4L"""

    def extract(self, content: str) -> Dict[str, Any]:
        """Extract all structural elements"""
        return {
            "sections": self._extract_sections(content),
            "contexts": self._extract_contexts(content),
            "relations": self._extract_relations(content),
            "entities": self._extract_entities(content),
            "timeline_events": self._extract_timeline(content),
            "aliases": self._extract_aliases(content)
        }

    def _extract_sections(self, content: str) -> List[str]:
        """Extract section titles"""
        sections = []
        for match in re.finditer(r'^-\s*([^-\n].*)$', content, re.MULTILINE):
            sections.append(match.group(1).strip())
        return sections

    def _extract_contexts(self, content: str) -> List[str]:
        """Extract contexts"""
        contexts = re.findall(r'::\s*([^:]+)\s*::', content)
        return [c.strip() for c in contexts if c.strip() and '_timeline_' not in c]

    def _extract_relations(self, content: str) -> List[Dict[str, str]]:
        """Extract relations (Subject (predicate) Object)"""
        relations = []
        pattern = r'^([^#"\n(][^(\n]*?)\s*\(([^)]+)\)\s*([^\n(]+)'

        for match in re.finditer(pattern, content, re.MULTILINE):
            subj, pred, obj = match.groups()
            subj = subj.strip()
            pred = pred.strip()
            obj = obj.strip().split('(')[0].strip()  # Remove chained relations

            if subj and pred and obj:
                relations.append({
                    "subject": subj,
                    "predicate": pred,
                    "object": obj
                })

        return relations

    def _extract_entities(self, content: str) -> List[str]:
        """Extract unique entities from relations"""
        relations = self._extract_relations(content)
        entities = set()
        for rel in relations:
            entities.add(rel["subject"])
            entities.add(rel["object"])
        return list(entities)

    def _extract_timeline(self, content: str) -> List[str]:
        """Extract timeline events"""
        events = []
        timeline_match = re.search(
            r'\+::\s*_timeline_\s*::(.+?)-::\s*_timeline_\s*::',
            content,
            re.DOTALL
        )
        if timeline_match:
            timeline_content = timeline_match.group(1)
            for line in timeline_content.split('\n'):
                if '->' in line and line.strip():
                    events.append(line.strip())
        return events

    def _extract_aliases(self, content: str) -> Dict[str, str]:
        """Extract aliases (@name)"""
        aliases = {}
        for match in re.finditer(r'^@(\w+)\s+(.+)$', content, re.MULTILINE):
            aliases[match.group(1)] = match.group(2).strip()
        return aliases


class CoverageCalculator:
    """Calculate information coverage between source text and N4L"""

    def __init__(self):
        self.extractor = N4LStructureExtractor()

    def calculate_entity_coverage(
        self,
        source_text: str,
        generated_n4l: str
    ) -> float:
        """
        Calculate what fraction of named entities from source appear in N4L.
        """
        # Extract potential entities from source (capitalized words, proper nouns)
        source_entities = set()

        # Names (capitalized words)
        for match in re.finditer(r'\b([A-ZÀ-Ü][a-zà-ü]+(?:\s+[A-ZÀ-Ü][a-zà-ü]+)*)\b', source_text):
            entity = match.group(1)
            if len(entity) > 2:  # Filter short matches
                source_entities.add(entity.lower())

        # Numbers and dates
        for match in re.finditer(r'\b(\d{1,2}[:/]\d{2}|\d{4}|\d+\s*(?:euros?|ans?))\b', source_text):
            source_entities.add(match.group(1).lower())

        if not source_entities:
            return 1.0  # No entities to match

        # Extract entities from N4L
        n4l_structure = self.extractor.extract(generated_n4l)
        n4l_text = ' '.join(n4l_structure['entities'])
        n4l_text_lower = n4l_text.lower() + ' ' + generated_n4l.lower()

        # Calculate overlap
        matched = sum(1 for e in source_entities if e in n4l_text_lower)
        coverage = matched / len(source_entities)

        return coverage

    def calculate_relation_richness(self, generated_n4l: str) -> int:
        """Count number of relations in generated N4L"""
        structure = self.extractor.extract(generated_n4l)
        return len(structure['relations'])


class TextSimilarityScorer:
    """Calculate text similarity metrics (BLEU, ROUGE)"""

    def __init__(self):
        self._nltk_ready = False
        self._rouge_ready = False

    def _setup_nltk(self):
        """Setup NLTK for BLEU"""
        if self._nltk_ready:
            return True
        try:
            import nltk
            nltk.download('punkt', quiet=True)
            self._nltk_ready = True
            return True
        except ImportError:
            logger.warning("NLTK not available - BLEU scores disabled")
            return False

    def _setup_rouge(self):
        """Setup rouge-score"""
        if self._rouge_ready:
            return True
        try:
            from rouge_score import rouge_scorer
            self._rouge_ready = True
            return True
        except ImportError:
            logger.warning("rouge-score not available - ROUGE scores disabled")
            return False

    def calculate_bleu(self, reference: str, hypothesis: str) -> float:
        """Calculate BLEU score"""
        if not self._setup_nltk():
            return 0.0

        from nltk.translate.bleu_score import sentence_bleu, SmoothingFunction

        # Tokenize
        ref_tokens = reference.lower().split()
        hyp_tokens = hypothesis.lower().split()

        if not ref_tokens or not hyp_tokens:
            return 0.0

        # Calculate with smoothing
        smoothing = SmoothingFunction().method1
        try:
            score = sentence_bleu([ref_tokens], hyp_tokens, smoothing_function=smoothing)
            return score
        except Exception:
            return 0.0

    def calculate_rouge_l(self, reference: str, hypothesis: str) -> float:
        """Calculate ROUGE-L score"""
        if not self._setup_rouge():
            return 0.0

        from rouge_score import rouge_scorer

        scorer = rouge_scorer.RougeScorer(['rougeL'], use_stemmer=True)
        scores = scorer.score(reference, hypothesis)
        return scores['rougeL'].fmeasure


class N4LEvaluator:
    """Main evaluator for N4L generation"""

    def __init__(self):
        self.syntax_validator = N4LSyntaxValidator()
        self.structure_extractor = N4LStructureExtractor()
        self.coverage_calculator = CoverageCalculator()
        self.similarity_scorer = TextSimilarityScorer()

    def evaluate_single(
        self,
        source_text: str,
        generated_n4l: str,
        reference_n4l: Optional[str] = None
    ) -> N4LEvaluationResult:
        """Evaluate a single generation"""
        result = N4LEvaluationResult()

        # Syntax validation
        is_valid, errors, warnings = self.syntax_validator.validate(generated_n4l)
        result.syntax_valid = is_valid
        result.syntax_errors = errors
        result.syntax_warnings = warnings
        result.syntax_score = max(0, 1.0 - len(errors) * 0.2 - len(warnings) * 0.05)

        # Structure extraction
        structure = self.structure_extractor.extract(generated_n4l)
        result.num_sections = len(structure['sections'])
        result.num_contexts = len(structure['contexts'])
        result.num_relations = len(structure['relations'])
        result.num_entities = len(structure['entities'])
        result.num_timeline_events = len(structure['timeline_events'])

        # Coverage
        result.entity_coverage = self.coverage_calculator.calculate_entity_coverage(
            source_text, generated_n4l
        )

        # Similarity with reference (if provided)
        if reference_n4l:
            result.bleu_score = self.similarity_scorer.calculate_bleu(
                reference_n4l, generated_n4l
            )
            result.rouge_l_score = self.similarity_scorer.calculate_rouge_l(
                reference_n4l, generated_n4l
            )

        # Overall score
        result.overall_score = self._calculate_overall_score(result)

        return result

    def _calculate_overall_score(self, result: N4LEvaluationResult) -> float:
        """Calculate weighted overall score"""
        weights = {
            'syntax': 0.25,
            'coverage': 0.25,
            'richness': 0.20,
            'similarity': 0.30
        }

        # Normalize richness (target: 20 relations)
        richness_score = min(result.num_relations / 20, 1.0)

        # Average similarity
        similarity_score = (result.bleu_score + result.rouge_l_score) / 2

        overall = (
            weights['syntax'] * result.syntax_score +
            weights['coverage'] * result.entity_coverage +
            weights['richness'] * richness_score +
            weights['similarity'] * similarity_score
        )

        return overall

    def evaluate_batch(
        self,
        examples: List[Dict[str, str]],
        model=None,
        tokenizer=None
    ) -> Dict[str, Any]:
        """
        Evaluate a batch of examples.

        Args:
            examples: List of dicts with 'input', 'output' (reference), and optionally 'generated'
            model: Optional model for generation
            tokenizer: Optional tokenizer
        """
        results = []

        for example in tqdm(examples, desc="Evaluating"):
            source_text = example.get('input', '')
            reference_n4l = example.get('output', '')

            # Generate if model provided and no 'generated' key
            if 'generated' in example:
                generated_n4l = example['generated']
            elif model is not None and tokenizer is not None:
                generated_n4l = self._generate(model, tokenizer, source_text)
            else:
                generated_n4l = reference_n4l  # Self-evaluation

            result = self.evaluate_single(source_text, generated_n4l, reference_n4l)
            results.append(result)

        # Aggregate metrics
        aggregated = self._aggregate_results(results)
        return aggregated

    def _generate(self, model, tokenizer, text: str) -> str:
        """Generate N4L from text using model"""
        prompt = f"""<|im_start|>system
Tu es un expert en structuration de connaissances au format N4L.<|im_end|>
<|im_start|>user
Convertis ce texte en format N4L structuré.

{text}<|im_end|>
<|im_start|>assistant
"""
        inputs = tokenizer(prompt, return_tensors="pt").to(model.device)

        with torch.no_grad():
            outputs = model.generate(
                **inputs,
                max_new_tokens=2048,
                temperature=0.3,
                top_p=0.9,
                do_sample=True,
                pad_token_id=tokenizer.eos_token_id
            )

        generated = tokenizer.decode(outputs[0], skip_special_tokens=True)

        # Extract assistant response
        if "<|im_start|>assistant" in generated:
            generated = generated.split("<|im_start|>assistant")[-1]
        if "<|im_end|>" in generated:
            generated = generated.split("<|im_end|>")[0]

        return generated.strip()

    def _aggregate_results(self, results: List[N4LEvaluationResult]) -> Dict[str, Any]:
        """Aggregate results across examples"""
        n = len(results)
        if n == 0:
            return {}

        return {
            "num_examples": n,
            "syntax": {
                "valid_ratio": sum(1 for r in results if r.syntax_valid) / n,
                "avg_score": sum(r.syntax_score for r in results) / n,
                "total_errors": sum(len(r.syntax_errors) for r in results),
                "total_warnings": sum(len(r.syntax_warnings) for r in results)
            },
            "structure": {
                "avg_sections": sum(r.num_sections for r in results) / n,
                "avg_contexts": sum(r.num_contexts for r in results) / n,
                "avg_relations": sum(r.num_relations for r in results) / n,
                "avg_entities": sum(r.num_entities for r in results) / n,
                "avg_timeline_events": sum(r.num_timeline_events for r in results) / n
            },
            "coverage": {
                "avg_entity_coverage": sum(r.entity_coverage for r in results) / n
            },
            "similarity": {
                "avg_bleu": sum(r.bleu_score for r in results) / n,
                "avg_rouge_l": sum(r.rouge_l_score for r in results) / n
            },
            "overall": {
                "avg_score": sum(r.overall_score for r in results) / n,
                "min_score": min(r.overall_score for r in results),
                "max_score": max(r.overall_score for r in results)
            }
        }


def evaluate_model(
    model_path: str,
    test_data_path: str,
    output_path: Optional[str] = None,
    num_samples: Optional[int] = None
):
    """
    Evaluate a fine-tuned model on test data.

    Args:
        model_path: Path to fine-tuned model
        test_data_path: Path to test JSONL
        output_path: Path to save results
        num_samples: Number of samples to evaluate (None = all)
    """
    from transformers import AutoModelForCausalLM, AutoTokenizer
    from peft import PeftModel

    logger.info(f"Loading model from {model_path}")

    # Check if it's a LoRA model
    config_path = Path(model_path) / "adapter_config.json"
    if config_path.exists():
        # Load base model and apply LoRA
        with open(Path(model_path) / "adapter_config.json") as f:
            adapter_config = json.load(f)
        base_model_name = adapter_config.get("base_model_name_or_path", "Qwen/Qwen2.5-7B-Instruct")

        model = AutoModelForCausalLM.from_pretrained(
            base_model_name,
            torch_dtype=torch.bfloat16,
            device_map="auto",
            trust_remote_code=True
        )
        model = PeftModel.from_pretrained(model, model_path)
    else:
        model = AutoModelForCausalLM.from_pretrained(
            model_path,
            torch_dtype=torch.bfloat16,
            device_map="auto",
            trust_remote_code=True
        )

    tokenizer = AutoTokenizer.from_pretrained(model_path)
    model.eval()

    # Load test data
    logger.info(f"Loading test data from {test_data_path}")
    examples = []
    with open(test_data_path, 'r') as f:
        for line in f:
            if line.strip():
                examples.append(json.loads(line))

    if num_samples:
        examples = examples[:num_samples]

    # Evaluate
    evaluator = N4LEvaluator()
    results = evaluator.evaluate_batch(examples, model, tokenizer)

    # Print results
    logger.info("\n" + "=" * 50)
    logger.info("EVALUATION RESULTS")
    logger.info("=" * 50)
    logger.info(f"Examples evaluated: {results['num_examples']}")
    logger.info(f"\nSyntax:")
    logger.info(f"  Valid ratio: {results['syntax']['valid_ratio']:.2%}")
    logger.info(f"  Avg score: {results['syntax']['avg_score']:.3f}")
    logger.info(f"\nStructure:")
    logger.info(f"  Avg relations: {results['structure']['avg_relations']:.1f}")
    logger.info(f"  Avg entities: {results['structure']['avg_entities']:.1f}")
    logger.info(f"\nCoverage:")
    logger.info(f"  Entity coverage: {results['coverage']['avg_entity_coverage']:.2%}")
    logger.info(f"\nSimilarity:")
    logger.info(f"  BLEU: {results['similarity']['avg_bleu']:.3f}")
    logger.info(f"  ROUGE-L: {results['similarity']['avg_rouge_l']:.3f}")
    logger.info(f"\nOverall score: {results['overall']['avg_score']:.3f}")
    logger.info("=" * 50)

    # Save results
    if output_path:
        with open(output_path, 'w') as f:
            json.dump(results, f, indent=2, ensure_ascii=False)
        logger.info(f"Results saved to {output_path}")

    return results


def main():
    """Main entry point"""
    import argparse

    parser = argparse.ArgumentParser(description="Evaluate N4L generator model")
    parser.add_argument("--model", type=str, required=True,
                       help="Path to fine-tuned model")
    parser.add_argument("--test-data", type=str, required=True,
                       help="Path to test JSONL file")
    parser.add_argument("--output", type=str, default="results/evaluation.json",
                       help="Path to save results")
    parser.add_argument("--num-samples", type=int, default=None,
                       help="Number of samples to evaluate")

    args = parser.parse_args()

    # Create output directory
    Path(args.output).parent.mkdir(parents=True, exist_ok=True)

    evaluate_model(
        model_path=args.model,
        test_data_path=args.test_data,
        output_path=args.output,
        num_samples=args.num_samples
    )


if __name__ == "__main__":
    main()
