#!/usr/bin/env python3
"""
N4L Dataset Generator for Fine-tuning

This module generates training data pairs (narrative text, N4L) for fine-tuning
a language model to convert text to N4L format.

Usage:
    python data_generator.py --n4l-dir ../examples --output-dir data/processed
    python data_generator.py --create-splits --input data/processed --output data/splits
"""

import json
import os
import re
import random
import hashlib
from pathlib import Path
from typing import List, Dict, Tuple, Optional, Any
from dataclasses import dataclass, field
from enum import Enum
import logging
import yaml
import requests
from tqdm import tqdm

logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)


class Domain(Enum):
    """Domains for synthetic data generation"""
    INVESTIGATION = "investigation"
    BIOGRAPHY = "biography"
    PROJECT = "project"
    DOCUMENTATION = "documentation"
    RECIPE = "recipe"
    GENERAL = "general"


@dataclass
class N4LStructure:
    """Parsed structure of an N4L file"""
    title: str = ""
    sections: List[str] = field(default_factory=list)
    contexts: List[str] = field(default_factory=list)
    entities: List[str] = field(default_factory=list)
    relations: List[Dict[str, str]] = field(default_factory=list)
    timeline_events: List[str] = field(default_factory=list)
    aliases: Dict[str, str] = field(default_factory=dict)
    comments: List[str] = field(default_factory=list)


@dataclass
class TrainingExample:
    """A training example for fine-tuning"""
    instruction: str
    input: str  # Narrative text
    output: str  # N4L format
    domain: str = "general"
    source: str = "unknown"  # "real", "synthetic", "template"

    def to_dict(self) -> Dict:
        return {
            "instruction": self.instruction,
            "input": self.input,
            "output": self.output,
            "domain": self.domain,
            "source": self.source
        }


class N4LParser:
    """Parser for N4L files to extract structure"""

    def parse(self, content: str) -> N4LStructure:
        """Parse N4L content and extract structure"""
        structure = N4LStructure()

        # Extract title (lines starting with -)
        title_match = re.search(r'^-\s*(.+)$', content, re.MULTILINE)
        if title_match:
            structure.title = title_match.group(1).strip()

        # Extract contexts (:: ... ::)
        contexts = re.findall(r'::\s*([^:]+)\s*::', content)
        structure.contexts = [c.strip() for c in contexts if c.strip()]

        # Extract sections (--- or multiple -)
        sections = re.findall(r'^-{3,}$|^-\s+([^-\n].*)$', content, re.MULTILINE)
        structure.sections = [s for s in sections if s]

        # Extract relations (Subject (relation) Object)
        relations = re.findall(
            r'^([^#"\n(][^(\n]*?)\s*\(([^)]+)\)\s*(.+?)(?:\s*\(|$)',
            content,
            re.MULTILINE
        )
        for rel in relations:
            subj, pred, obj = rel
            if subj.strip() and pred.strip() and obj.strip():
                structure.relations.append({
                    "subject": subj.strip(),
                    "predicate": pred.strip(),
                    "object": obj.strip().split('(')[0].strip()
                })

        # Extract entities from relations
        entities = set()
        for rel in structure.relations:
            entities.add(rel["subject"])
            entities.add(rel["object"])
        structure.entities = list(entities)

        # Extract aliases (@name)
        aliases = re.findall(r'^@(\w+)\s+(.+)$', content, re.MULTILINE)
        structure.aliases = {a[0]: a[1].strip() for a in aliases}

        # Extract timeline events
        timeline_match = re.search(
            r'\+::\s*_timeline_\s*::(.+?)-::\s*_timeline_\s*::',
            content,
            re.DOTALL
        )
        if timeline_match:
            timeline_content = timeline_match.group(1)
            events = re.findall(r'^(.+?->.*?)$', timeline_content, re.MULTILINE)
            structure.timeline_events = [e.strip() for e in events]

        return structure

    def validate(self, content: str) -> Dict[str, Any]:
        """Validate N4L syntax"""
        errors = []
        warnings = []

        # Check for unclosed contexts
        ctx_count = content.count('::')
        if ctx_count % 2 != 0:
            errors.append("Unclosed context (:: without pair)")

        # Check parentheses balance in relations
        for i, line in enumerate(content.split('\n'), 1):
            if '(' in line and not line.strip().startswith('#'):
                if line.count('(') != line.count(')'):
                    errors.append(f"Line {i}: Unbalanced parentheses")

        # Check ditto usage
        lines = content.split('\n')
        prev_subject = None
        for i, line in enumerate(lines, 1):
            stripped = line.strip()
            if stripped.startswith('"') and '(' in stripped:
                if prev_subject is None:
                    warnings.append(f"Line {i}: Ditto without previous subject")
            elif re.match(r'^[^#"(\s]', stripped):
                match = re.match(r'^([^(]+)', stripped)
                if match:
                    prev_subject = match.group(1).strip()

        return {
            "valid": len(errors) == 0,
            "errors": errors,
            "warnings": warnings,
            "score": max(0, 1.0 - len(errors) * 0.2 - len(warnings) * 0.05)
        }


class OllamaClient:
    """Client for Ollama API"""

    def __init__(self, base_url: str = "http://localhost:11434", model: str = "mistral"):
        self.base_url = base_url
        self.model = model

    def generate(self, prompt: str, temperature: float = 0.7) -> str:
        """Generate text using Ollama"""
        try:
            response = requests.post(
                f"{self.base_url}/api/generate",
                json={
                    "model": self.model,
                    "prompt": prompt,
                    "stream": False,
                    "options": {"temperature": temperature}
                },
                timeout=120
            )
            if response.status_code == 200:
                return response.json().get("response", "")
            else:
                logger.error(f"Ollama error: {response.status_code}")
                return ""
        except Exception as e:
            logger.error(f"Ollama connection error: {e}")
            return ""

    def is_available(self) -> bool:
        """Check if Ollama is available"""
        try:
            response = requests.get(f"{self.base_url}/api/tags", timeout=5)
            return response.status_code == 200
        except:
            return False


class N4LDatasetGenerator:
    """Generator for N4L training datasets"""

    INSTRUCTION_VARIANTS = [
        "Convertis ce texte en format N4L structuré.",
        "Transforme ce récit en notes N4L pour une base de connaissances.",
        "Analyse ce texte et produis une représentation N4L.",
        "Structure les informations de ce texte au format N4L.",
        "Génère un fichier N4L à partir de ce contenu narratif.",
        "Extrait les entités et relations de ce texte en N4L.",
    ]

    def __init__(
        self,
        n4l_examples_dir: str,
        output_dir: str,
        config_path: Optional[str] = None
    ):
        self.n4l_dir = Path(n4l_examples_dir)
        self.output_dir = Path(output_dir)
        self.output_dir.mkdir(parents=True, exist_ok=True)

        self.parser = N4LParser()
        self.ollama = None

        # Load config if provided
        self.config = {}
        if config_path and Path(config_path).exists():
            with open(config_path, 'r') as f:
                self.config = yaml.safe_load(f)

    def setup_llm(self, provider: str = "ollama", model: str = "mistral"):
        """Setup LLM client for narrative generation"""
        if provider == "ollama":
            self.ollama = OllamaClient(model=model)
            if not self.ollama.is_available():
                logger.warning("Ollama not available - synthetic generation disabled")
                self.ollama = None

    def load_n4l_files(self) -> List[Tuple[str, str]]:
        """Load all N4L files from the examples directory"""
        n4l_files = []

        for pattern in ["*.n4l", "**/*.n4l"]:
            for filepath in self.n4l_dir.glob(pattern):
                try:
                    content = filepath.read_text(encoding='utf-8')
                    n4l_files.append((str(filepath), content))
                    logger.info(f"Loaded: {filepath.name}")
                except Exception as e:
                    logger.error(f"Error loading {filepath}: {e}")

        return n4l_files

    def generate_narrative_from_n4l(self, n4l_content: str, structure: N4LStructure) -> str:
        """Generate narrative text from N4L using LLM"""
        if not self.ollama:
            return ""

        prompt = f"""Tu es un expert en rédaction narrative.

À partir de ces notes structurées en format N4L, génère un texte narratif fluide
et cohérent qui raconte l'histoire de manière naturelle.

Notes N4L:
```
{n4l_content[:4000]}
```

Règles:
1. Écris un récit en français, fluide et engageant
2. Inclus TOUTES les informations importantes des notes
3. Ne mentionne PAS le format N4L ni la structure des notes
4. Utilise des transitions naturelles entre les idées
5. Longueur: 500-1500 mots selon la complexité

Texte narratif:"""

        return self.ollama.generate(prompt, temperature=0.7)

    def create_example(
        self,
        text: str,
        n4l: str,
        domain: str = "general",
        source: str = "unknown"
    ) -> TrainingExample:
        """Create a training example"""
        instruction = random.choice(self.INSTRUCTION_VARIANTS)

        return TrainingExample(
            instruction=instruction,
            input=text.strip(),
            output=n4l.strip(),
            domain=domain,
            source=source
        )

    def generate_from_existing_n4l(self) -> List[TrainingExample]:
        """Generate examples from existing N4L files"""
        examples = []
        n4l_files = self.load_n4l_files()

        logger.info(f"Generating narratives for {len(n4l_files)} N4L files...")

        for filepath, n4l_content in tqdm(n4l_files, desc="Processing N4L files"):
            # Parse structure
            structure = self.parser.parse(n4l_content)

            # Validate
            validation = self.parser.validate(n4l_content)
            if not validation["valid"]:
                logger.warning(f"Skipping invalid N4L: {filepath}")
                continue

            # Determine domain from path/content
            domain = self._detect_domain(filepath, n4l_content)

            # Generate narrative
            if self.ollama:
                narrative = self.generate_narrative_from_n4l(n4l_content, structure)
                if narrative:
                    example = self.create_example(
                        text=narrative,
                        n4l=n4l_content,
                        domain=domain,
                        source="real"
                    )
                    examples.append(example)
                    logger.info(f"Generated example from {Path(filepath).name}")

        return examples

    def _detect_domain(self, filepath: str, content: str) -> str:
        """Detect domain from filepath or content"""
        filepath_lower = filepath.lower()
        content_lower = content.lower()

        if any(k in filepath_lower or k in content_lower
               for k in ["murder", "crime", "enquête", "investigation", "forensic", "cluedo"]):
            return Domain.INVESTIGATION.value
        elif any(k in content_lower for k in ["biography", "born", "né", "career"]):
            return Domain.BIOGRAPHY.value
        elif any(k in content_lower for k in ["project", "task", "milestone", "deadline"]):
            return Domain.PROJECT.value
        elif any(k in filepath_lower for k in ["tutorial", "doc", "readme"]):
            return Domain.DOCUMENTATION.value

        return Domain.GENERAL.value

    def generate_template_examples(self, num_per_domain: int = 50) -> List[TrainingExample]:
        """Generate synthetic examples from templates"""
        examples = []

        template_generators = {
            Domain.INVESTIGATION: self._template_investigation,
            Domain.BIOGRAPHY: self._template_biography,
            Domain.PROJECT: self._template_project,
        }

        for domain, generator in template_generators.items():
            logger.info(f"Generating {num_per_domain} {domain.value} templates...")
            for i in tqdm(range(num_per_domain), desc=f"{domain.value}"):
                try:
                    text, n4l = generator(seed=i)
                    example = self.create_example(
                        text=text,
                        n4l=n4l,
                        domain=domain.value,
                        source="template"
                    )
                    examples.append(example)
                except Exception as e:
                    logger.error(f"Template generation error: {e}")

        return examples

    def _template_investigation(self, seed: int) -> Tuple[str, str]:
        """Generate investigation domain example"""
        random.seed(seed)

        victims = ["Marie Dupont", "Jean Martin", "Pierre Leblanc", "Sophie Bernard", "Lucas Petit"]
        suspects = ["le voisin", "l'associé", "l'ex-conjoint", "le collègue", "l'ami d'enfance"]
        locations = ["appartement", "bureau", "restaurant", "parc", "parking"]
        mobiles = ["financier", "passionnel", "vengeance", "jalousie", "héritage"]
        times = [f"{h}h{m:02d}" for h in range(18, 24) for m in [0, 15, 30, 45]]
        dates = [f"{d:02d}/0{m}/2025" for m in range(1, 10) for d in range(1, 29)]

        victim = random.choice(victims)
        suspect = random.choice(suspects)
        location = random.choice(locations)
        mobile = random.choice(mobiles)
        time = random.choice(times)
        date = random.choice(dates)

        # Generate narrative text
        text = f"""L'enquête sur la mort de {victim} a débuté le {date}.
La victime, âgée de {random.randint(30, 70)} ans, a été retrouvée dans son {location} vers {time}.
Les premiers éléments de l'enquête indiquent qu'il s'agirait d'un homicide.

Les enquêteurs ont rapidement identifié {suspect} comme personne d'intérêt.
Celui-ci avait un mobile {mobile} et était présent dans les environs au moment des faits.
Plusieurs témoins ont été interrogés et des preuves matérielles ont été collectées sur la scène de crime.

L'autopsie a révélé que la victime est décédée des suites de blessures multiples.
L'enquête se poursuit pour déterminer les circonstances exactes du décès."""

        # Generate N4L
        n4l = f"""- Enquête {victim.split()[0]}

:: Métadonnées ::

Affaire (type) Homicide
    "   (statut) En cours
    "   (date ouverture) {date}

---

:: Victime ::

{victim} (lieu découverte) {location}
    "    (heure) {time}
    "    (date décès) {date}
    "    (âge) {random.randint(30, 70)} ans
    "    (cause décès) Blessures multiples

---

:: Suspects ::

{suspect.capitalize()} (mobile) {mobile}
    "                   (statut) Personne d'intérêt
    "                   (présence) Environs au moment des faits

---

:: Preuves ::

Preuves matérielles (source) Scène de crime
Témoignages (statut) En cours de collecte
Autopsie (résultat) Blessures multiples

---

:: Chronologie ::

+:: _timeline_ ::
{date} {time} -> Découverte du corps -> {location}
{date} -> Début enquête -> Identification suspect
-:: _timeline_ ::
"""
        return text, n4l

    def _template_biography(self, seed: int) -> Tuple[str, str]:
        """Generate biography domain example"""
        random.seed(seed)

        first_names = ["Jean", "Marie", "Pierre", "Sophie", "Antoine", "Claire"]
        last_names = ["Martin", "Dubois", "Bernard", "Petit", "Robert", "Richard"]
        professions = ["écrivain", "scientifique", "artiste", "entrepreneur", "médecin", "avocat"]
        cities = ["Paris", "Lyon", "Marseille", "Bordeaux", "Toulouse", "Nantes"]
        achievements = ["Prix Nobel", "Légion d'honneur", "Prix Goncourt", "Oscar", "César"]

        name = f"{random.choice(first_names)} {random.choice(last_names)}"
        profession = random.choice(professions)
        birth_city = random.choice(cities)
        birth_year = random.randint(1920, 1990)
        achievement = random.choice(achievements)

        text = f"""{name} est né(e) en {birth_year} à {birth_city}.
Dès son plus jeune âge, {name.split()[0]} a montré un intérêt particulier pour les arts et les sciences.
Après des études brillantes, il/elle s'est orienté(e) vers une carrière de {profession}.

Au cours de sa carrière, {name} a réalisé de nombreuses contributions significatives dans son domaine.
Son travail a été reconnu internationalement, lui valant notamment le {achievement}.
{name.split()[0]} continue d'influencer sa discipline à travers ses publications et son enseignement."""

        n4l = f"""- Biographie de {name}

:: Informations personnelles ::

{name} (né en) {birth_year}
    "  (lieu naissance) {birth_city}
    "  (profession) {profession}

---

:: Carrière ::

{name} (distinction) {achievement}
    "  (domaine) {profession}
    "  (influence) Publications et enseignement

---

:: Chronologie ::

+:: _timeline_ ::
{birth_year} -> Naissance -> {birth_city}
{birth_year + 20} -> Début carrière -> {profession}
{birth_year + 40} -> Distinction -> {achievement}
-:: _timeline_ ::
"""
        return text, n4l

    def _template_project(self, seed: int) -> Tuple[str, str]:
        """Generate project domain example"""
        random.seed(seed)

        project_types = ["application mobile", "site web", "API", "système embarqué", "IA"]
        teams = ["équipe Alpha", "équipe Beta", "équipe Gamma", "équipe Delta"]
        statuses = ["planification", "développement", "test", "déploiement"]
        technologies = ["Python", "Go", "React", "Node.js", "Kubernetes", "Docker"]

        project = f"Projet {random.choice(['Phoenix', 'Atlas', 'Nexus', 'Horizon', 'Zenith'])}"
        project_type = random.choice(project_types)
        team = random.choice(teams)
        status = random.choice(statuses)
        tech = random.sample(technologies, 3)
        deadline = f"{random.randint(1, 28):02d}/{random.randint(1, 12):02d}/2025"

        text = f"""Le {project} est une initiative visant à développer une {project_type}.
Ce projet est mené par l'{team} et utilise les technologies {', '.join(tech)}.
Actuellement en phase de {status}, le projet doit être livré pour le {deadline}.

L'objectif principal est d'améliorer l'efficacité des processus existants.
Le budget alloué permet une équipe de {random.randint(3, 10)} développeurs."""

        n4l = f"""- {project}

:: Métadonnées projet ::

{project} (type) {project_type}
    "     (équipe) {team}
    "     (statut) {status}
    "     (deadline) {deadline}

---

:: Technologies ::

{project} (utilise) {tech[0]}
    "     (utilise) {tech[1]}
    "     (utilise) {tech[2]}

---

:: Objectifs ::

Objectif principal (description) Améliorer efficacité des processus
Équipe (taille) {random.randint(3, 10)} développeurs
"""
        return text, n4l

    def save_examples(self, examples: List[TrainingExample], filename: str):
        """Save examples to JSONL file"""
        output_path = self.output_dir / filename

        with open(output_path, 'w', encoding='utf-8') as f:
            for example in examples:
                f.write(json.dumps(example.to_dict(), ensure_ascii=False) + '\n')

        logger.info(f"Saved {len(examples)} examples to {output_path}")

    def create_splits(
        self,
        examples: List[TrainingExample],
        train_ratio: float = 0.8,
        val_ratio: float = 0.1,
        test_ratio: float = 0.1,
        seed: int = 42
    ) -> Tuple[List, List, List]:
        """Split examples into train/val/test sets"""
        random.seed(seed)
        shuffled = examples.copy()
        random.shuffle(shuffled)

        n = len(shuffled)
        train_end = int(n * train_ratio)
        val_end = train_end + int(n * val_ratio)

        train = shuffled[:train_end]
        val = shuffled[train_end:val_end]
        test = shuffled[val_end:]

        return train, val, test

    def generate_full_dataset(
        self,
        use_existing: bool = True,
        num_templates: int = 50,
        create_splits: bool = True
    ) -> Dict[str, int]:
        """Generate complete dataset"""
        all_examples = []

        # Generate from existing N4L files
        if use_existing:
            existing_examples = self.generate_from_existing_n4l()
            all_examples.extend(existing_examples)
            logger.info(f"Generated {len(existing_examples)} examples from existing N4L")

        # Generate template examples
        template_examples = self.generate_template_examples(num_per_domain=num_templates)
        all_examples.extend(template_examples)
        logger.info(f"Generated {len(template_examples)} template examples")

        # Save all examples
        self.save_examples(all_examples, "all_examples.jsonl")

        # Create splits
        if create_splits:
            train, val, test = self.create_splits(all_examples)

            splits_dir = self.output_dir.parent / "splits"
            splits_dir.mkdir(exist_ok=True)

            self.output_dir = splits_dir
            self.save_examples(train, "train.jsonl")
            self.save_examples(val, "val.jsonl")
            self.save_examples(test, "test.jsonl")

        return {
            "total": len(all_examples),
            "train": len(train) if create_splits else 0,
            "val": len(val) if create_splits else 0,
            "test": len(test) if create_splits else 0
        }


def main():
    """Main entry point"""
    import argparse

    parser = argparse.ArgumentParser(description="Generate N4L training dataset")
    parser.add_argument("--n4l-dir", type=str, default="../examples",
                       help="Directory containing N4L example files")
    parser.add_argument("--output-dir", type=str, default="data/processed",
                       help="Output directory for generated data")
    parser.add_argument("--config", type=str, default="config.yaml",
                       help="Configuration file path")
    parser.add_argument("--num-templates", type=int, default=50,
                       help="Number of template examples per domain")
    parser.add_argument("--llm-model", type=str, default="gpt-oss:20b",
                       help="Ollama model for narrative generation")
    parser.add_argument("--no-llm", action="store_true",
                       help="Disable LLM-based narrative generation")
    parser.add_argument("--create-splits", action="store_true",
                       help="Create train/val/test splits")

    args = parser.parse_args()

    # Initialize generator
    generator = N4LDatasetGenerator(
        n4l_examples_dir=args.n4l_dir,
        output_dir=args.output_dir,
        config_path=args.config
    )

    # Setup LLM if not disabled
    if not args.no_llm:
        generator.setup_llm(provider="ollama", model=args.llm_model)

    # Generate dataset
    stats = generator.generate_full_dataset(
        use_existing=not args.no_llm,
        num_templates=args.num_templates,
        create_splits=args.create_splits
    )

    logger.info(f"Dataset generation complete: {stats}")


if __name__ == "__main__":
    main()
