#!/usr/bin/env python3
"""
N4L Dataset Generator using Claude API

This module generates high-quality training data pairs (narrative text, N4L)
using Claude for both forward (text→N4L) and inverse (N4L→text) generation.

Usage:
    python claude_data_generator.py --api-key $ANTHROPIC_API_KEY --output-dir data/claude_generated
"""

import json
import os
import re
import random
import time
from pathlib import Path
from typing import List, Dict, Tuple, Optional, Any
from dataclasses import dataclass, field
from enum import Enum
import logging
from tqdm import tqdm
import anthropic

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
    SCIENCE = "science"
    HISTORY = "history"
    LITERATURE = "literature"
    TECHNOLOGY = "technology"
    MEDICINE = "medicine"
    LAW = "law"
    BUSINESS = "business"


@dataclass
class TrainingExample:
    """A training example for fine-tuning"""
    instruction: str
    input: str  # Narrative text
    output: str  # N4L format
    domain: str = "general"
    source: str = "claude"

    def to_dict(self) -> Dict:
        return {
            "instruction": self.instruction,
            "input": self.input,
            "output": self.output,
            "domain": self.domain,
            "source": self.source
        }


# N4L format documentation for Claude prompts
N4L_DOCUMENTATION = """
# N4L (Notes for Learning) Format Specification

N4L is a structured text format for organizing knowledge. Key syntax elements:

## 1. Title (required)
- Document title: `- Titre du document`

## 2. Sections
- Section dividers: `---`
- Named sections: `- Nom de section`

## 3. Contexts (grouping)
- Open/close context: `:: Nom du contexte ::`
- Contexts group related information

## 4. Relations (core element)
```
Sujet (relation) Objet
```
Examples:
- `Marie (est) médecin`
- `Paris (capitale de) France`
- `Projet Alpha (deadline) 15/03/2025`

## 5. Ditto " (subject continuation)
Use `"` to continue with same subject:
```
Jean Dupont (né en) 1985
    "       (profession) ingénieur
    "       (ville) Paris
```

## 6. Timeline blocks
```
+:: _timeline_ ::
2020-01 -> Événement A -> Résultat A
2020-06 -> Événement B -> Résultat B
-:: _timeline_ ::
```

## 7. Aliases (references)
```
@jd Jean Dupont
...
$jd.1 (rencontre) Marie  # Reference to Jean Dupont
```

## 8. Comments
```
# Ceci est un commentaire
```

## Example N4L document:
```
- Enquête Dupont

:: Métadonnées ::

Affaire (type) Homicide
    "   (statut) En cours
    "   (date) 15/01/2025

---

:: Victime ::

@victim Marie Dupont

$victim.1 (âge) 45 ans
    "     (profession) Avocate
    "     (lieu décès) Appartement personnel

---

:: Chronologie ::

+:: _timeline_ ::
14/01/2025 20h00 -> Dernier contact téléphonique -> Témoin A
15/01/2025 08h30 -> Découverte du corps -> Voisin
15/01/2025 09h00 -> Arrivée police -> Début enquête
-:: _timeline_ ::
```
"""


class ClaudeN4LGenerator:
    """Generate N4L training data using Claude API"""

    INSTRUCTION_VARIANTS = [
        "Convertis ce texte en format N4L structuré.",
        "Transforme ce récit en notes N4L pour une base de connaissances.",
        "Analyse ce texte et produis une représentation N4L.",
        "Structure les informations de ce texte au format N4L.",
        "Génère un fichier N4L à partir de ce contenu narratif.",
        "Extrait les entités et relations de ce texte en N4L.",
        "Organise ces informations en format N4L avec sections et relations.",
        "Crée une note N4L structurée à partir de ce texte.",
    ]

    DOMAIN_TOPICS = {
        Domain.INVESTIGATION: [
            "enquête criminelle sur un meurtre dans un hôtel de luxe",
            "vol de bijoux dans une galerie d'art",
            "disparition mystérieuse d'un scientifique",
            "fraude financière dans une grande entreprise",
            "affaire de chantage impliquant un politicien",
            "cambriolage d'une banque avec des complices internes",
            "empoisonnement lors d'un dîner de famille",
            "incendie criminel dans un entrepôt",
            "kidnapping d'un enfant de millionnaire",
            "cyberattaque contre une infrastructure critique",
        ],
        Domain.BIOGRAPHY: [
            "vie d'un inventeur révolutionnaire du 20ème siècle",
            "parcours d'une artiste peintre célèbre",
            "biographie d'un explorateur qui a découvert des terres inconnues",
            "histoire d'un chef cuisinier étoilé",
            "vie d'un compositeur de musique classique",
            "parcours d'un athlète olympique",
            "biographie d'un écrivain prix Nobel",
            "vie d'un architecte visionnaire",
            "histoire d'un médecin humanitaire",
            "parcours d'un entrepreneur tech à succès",
        ],
        Domain.PROJECT: [
            "développement d'une application mobile de santé",
            "construction d'un pont suspendu innovant",
            "mise en place d'un système de gestion hospitalier",
            "création d'une plateforme e-commerce",
            "projet de rénovation urbaine",
            "développement d'un jeu vidéo AAA",
            "implémentation d'un système IoT industriel",
            "migration vers le cloud d'une entreprise",
            "création d'une startup fintech",
            "projet de recherche en intelligence artificielle",
        ],
        Domain.SCIENCE: [
            "découverte d'une nouvelle particule subatomique",
            "étude sur le changement climatique en Arctique",
            "recherche sur les cellules souches",
            "exploration d'une exoplanète habitable",
            "développement d'un nouveau vaccin",
            "étude génétique sur les maladies héréditaires",
            "recherche sur la fusion nucléaire",
            "découverte archéologique majeure",
            "étude océanographique des abysses",
            "recherche sur les matériaux supraconducteurs",
        ],
        Domain.HISTORY: [
            "la révolution industrielle en Angleterre",
            "l'expansion de l'Empire romain",
            "la Renaissance italienne",
            "la guerre de Cent Ans",
            "la conquête spatiale du 20ème siècle",
            "les grandes découvertes maritimes",
            "la Révolution française",
            "l'ère des samouraïs au Japon",
            "l'âge d'or de l'Empire ottoman",
            "la construction des pyramides d'Égypte",
        ],
        Domain.LITERATURE: [
            "analyse d'un roman de Victor Hugo",
            "étude des œuvres de Shakespeare",
            "le mouvement surréaliste en littérature",
            "l'évolution du roman policier",
            "la poésie romantique française",
            "le théâtre de l'absurde",
            "la littérature existentialiste",
            "les contes des Mille et Une Nuits",
            "l'œuvre de Dostoïevski",
            "la science-fiction classique",
        ],
        Domain.TECHNOLOGY: [
            "évolution des processeurs informatiques",
            "histoire de l'Internet",
            "développement de l'intelligence artificielle",
            "l'essor des smartphones",
            "la blockchain et les cryptomonnaies",
            "les véhicules autonomes",
            "la réalité virtuelle et augmentée",
            "l'informatique quantique",
            "les énergies renouvelables",
            "l'impression 3D industrielle",
        ],
        Domain.MEDICINE: [
            "traitement d'une maladie auto-immune rare",
            "chirurgie cardiaque innovante",
            "développement d'une thérapie génique",
            "étude épidémiologique d'une pandémie",
            "traitement du cancer par immunothérapie",
            "recherche sur Alzheimer",
            "médecine régénérative et organes artificiels",
            "psychiatrie et nouvelles thérapies",
            "médecine personnalisée basée sur l'ADN",
            "robots chirurgicaux de précision",
        ],
        Domain.LAW: [
            "procès criminel pour meurtre prémédité",
            "affaire de droit des brevets",
            "contentieux commercial international",
            "procès pour discrimination au travail",
            "affaire de droit de l'environnement",
            "procès pour violation de données personnelles",
            "contentieux en droit de la famille",
            "affaire antitrust contre une Big Tech",
            "procès pour diffamation médiatique",
            "arbitrage commercial international",
        ],
        Domain.BUSINESS: [
            "fusion-acquisition de deux géants industriels",
            "stratégie de croissance d'une startup",
            "restructuration d'une entreprise en difficulté",
            "lancement d'un produit sur un nouveau marché",
            "négociation d'un partenariat stratégique",
            "transformation digitale d'une PME",
            "gestion de crise en entreprise",
            "développement à l'international",
            "stratégie de marque et repositionnement",
            "levée de fonds série A",
        ],
    }

    def __init__(
        self,
        api_key: str,
        output_dir: str,
        model: str = "claude-sonnet-4-20250514",
        existing_n4l_dir: Optional[str] = None
    ):
        self.client = anthropic.Anthropic(api_key=api_key)
        self.model = model
        self.output_dir = Path(output_dir)
        self.output_dir.mkdir(parents=True, exist_ok=True)
        self.existing_n4l_dir = Path(existing_n4l_dir) if existing_n4l_dir else None

        # Rate limiting
        self.requests_per_minute = 50
        self.last_request_time = 0

    def _rate_limit(self):
        """Simple rate limiting"""
        elapsed = time.time() - self.last_request_time
        min_interval = 60.0 / self.requests_per_minute
        if elapsed < min_interval:
            time.sleep(min_interval - elapsed)
        self.last_request_time = time.time()

    def _call_claude(self, system: str, user: str, max_tokens: int = 4000) -> str:
        """Call Claude API with rate limiting"""
        self._rate_limit()

        try:
            response = self.client.messages.create(
                model=self.model,
                max_tokens=max_tokens,
                system=system,
                messages=[{"role": "user", "content": user}]
            )
            return response.content[0].text
        except Exception as e:
            logger.error(f"Claude API error: {e}")
            return ""

    def generate_narrative(self, domain: Domain, topic: str) -> str:
        """Generate a rich narrative text on a given topic"""
        system = """Tu es un expert en rédaction. Génère des textes narratifs riches,
détaillés et informatifs en français. Tes textes doivent contenir des dates,
des noms de personnes, des lieux, des relations entre entités, et des événements chronologiques.
Ne mentionne jamais le format N4L ni aucun format structuré."""

        user = f"""Écris un texte narratif détaillé (400-800 mots) sur le sujet suivant:
{topic}

Le texte doit:
1. Contenir des noms de personnes spécifiques avec leurs rôles
2. Inclure des dates et des lieux précis
3. Décrire des relations entre les personnes/entités
4. Présenter une chronologie d'événements
5. Être écrit comme un article ou un rapport professionnel
6. Contenir des faits vérifiables et des détails concrets

Écris UNIQUEMENT le texte narratif, sans introduction ni commentaire."""

        return self._call_claude(system, user)

    def generate_n4l_from_narrative(self, narrative: str) -> str:
        """Convert narrative text to N4L format"""
        system = f"""Tu es un expert en structuration de l'information au format N4L.
{N4L_DOCUMENTATION}

Convertis les textes en format N4L en respectant strictement la syntaxe."""

        user = f"""Convertis ce texte en format N4L structuré:

{narrative}

Règles importantes:
1. Utilise des sections thématiques avec `:: Nom ::`
2. Utilise le ditto `"` pour les attributs multiples d'une même entité
3. Inclus une timeline si le texte contient des événements chronologiques
4. Utilise des alias (@) pour les entités fréquemment référencées
5. Structure clairement avec des séparateurs `---`

Génère UNIQUEMENT le code N4L, sans explication."""

        return self._call_claude(system, user)

    def generate_narrative_from_n4l(self, n4l_content: str) -> str:
        """Generate narrative from existing N4L (inverse generation)"""
        system = """Tu es un expert en rédaction narrative. À partir de notes structurées,
tu génères des textes fluides et engageants qui contiennent toutes les informations
des notes sans jamais mentionner le format source."""

        user = f"""À partir de ces notes structurées, génère un texte narratif complet:

{n4l_content}

Le texte doit:
1. Être fluide et bien écrit en français
2. Inclure TOUTES les informations des notes
3. Ne JAMAIS mentionner le format des notes ni leur structure
4. Utiliser des transitions naturelles
5. Faire 300-600 mots

Écris UNIQUEMENT le texte narratif."""

        return self._call_claude(system, user)

    def generate_synthetic_pair(self, domain: Domain) -> Optional[TrainingExample]:
        """Generate a complete synthetic training pair"""
        # Pick a random topic for this domain
        topics = self.DOMAIN_TOPICS.get(domain, ["sujet général"])
        topic = random.choice(topics)

        # Generate narrative
        narrative = self.generate_narrative(domain, topic)
        if not narrative or len(narrative) < 200:
            logger.warning(f"Failed to generate narrative for {domain.value}: {topic}")
            return None

        # Generate N4L from narrative
        n4l = self.generate_n4l_from_narrative(narrative)
        if not n4l or len(n4l) < 100:
            logger.warning(f"Failed to generate N4L for {domain.value}")
            return None

        # Clean up N4L (remove markdown code blocks if present)
        n4l = re.sub(r'^```\w*\n?', '', n4l)
        n4l = re.sub(r'\n?```$', '', n4l)

        return TrainingExample(
            instruction=random.choice(self.INSTRUCTION_VARIANTS),
            input=narrative.strip(),
            output=n4l.strip(),
            domain=domain.value,
            source="claude_synthetic"
        )

    def generate_from_existing_n4l(self) -> List[TrainingExample]:
        """Generate training pairs from existing N4L files"""
        if not self.existing_n4l_dir or not self.existing_n4l_dir.exists():
            return []

        examples = []
        n4l_files = list(self.existing_n4l_dir.glob("**/*.n4l"))

        logger.info(f"Processing {len(n4l_files)} existing N4L files...")

        for filepath in tqdm(n4l_files, desc="Processing existing N4L"):
            try:
                n4l_content = filepath.read_text(encoding='utf-8')

                # Skip very short or invalid files
                if len(n4l_content) < 100:
                    continue

                # Check for basic validity (has title and some structure)
                if not re.search(r'^-\s+.+', n4l_content, re.MULTILINE):
                    continue

                # Generate narrative from N4L
                narrative = self.generate_narrative_from_n4l(n4l_content)
                if narrative and len(narrative) > 150:
                    example = TrainingExample(
                        instruction=random.choice(self.INSTRUCTION_VARIANTS),
                        input=narrative.strip(),
                        output=n4l_content.strip(),
                        domain=self._detect_domain(filepath.name, n4l_content),
                        source="claude_from_existing"
                    )
                    examples.append(example)
                    logger.info(f"Generated from {filepath.name}")

            except Exception as e:
                logger.error(f"Error processing {filepath}: {e}")

        return examples

    def _detect_domain(self, filename: str, content: str) -> str:
        """Detect domain from filename or content"""
        filename_lower = filename.lower()
        content_lower = content.lower()

        keywords = {
            "investigation": ["murder", "crime", "enquête", "investigation", "suspect", "victime"],
            "biography": ["né", "born", "biograph", "career", "life"],
            "project": ["project", "milestone", "deadline", "équipe", "développement"],
            "science": ["research", "study", "découverte", "expériment", "scientif"],
            "history": ["siècle", "century", "guerre", "empire", "révolution"],
            "medicine": ["patient", "traitement", "médecin", "maladie", "symptôme"],
            "technology": ["software", "hardware", "algorithme", "système", "application"],
        }

        for domain, kw_list in keywords.items():
            if any(kw in filename_lower or kw in content_lower for kw in kw_list):
                return domain

        return "general"

    def generate_dataset(
        self,
        num_synthetic_per_domain: int = 50,
        use_existing: bool = True,
        progress_callback=None
    ) -> List[TrainingExample]:
        """Generate complete dataset"""
        all_examples = []

        # Generate from existing N4L files first
        if use_existing and self.existing_n4l_dir:
            existing_examples = self.generate_from_existing_n4l()
            all_examples.extend(existing_examples)
            logger.info(f"Generated {len(existing_examples)} examples from existing N4L files")

        # Generate synthetic examples for each domain
        domains = list(Domain)
        total_synthetic = num_synthetic_per_domain * len(domains)

        logger.info(f"Generating {total_synthetic} synthetic examples across {len(domains)} domains...")

        with tqdm(total=total_synthetic, desc="Generating synthetic data") as pbar:
            for domain in domains:
                logger.info(f"Generating {num_synthetic_per_domain} examples for {domain.value}...")

                for i in range(num_synthetic_per_domain):
                    example = self.generate_synthetic_pair(domain)
                    if example:
                        all_examples.append(example)

                    pbar.update(1)

                    if progress_callback:
                        progress_callback(len(all_examples), total_synthetic)

        return all_examples

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
        seed: int = 42
    ) -> Tuple[List, List, List]:
        """Split examples into train/val/test"""
        random.seed(seed)
        shuffled = examples.copy()
        random.shuffle(shuffled)

        n = len(shuffled)
        train_end = int(n * train_ratio)
        val_end = train_end + int(n * val_ratio)

        return shuffled[:train_end], shuffled[train_end:val_end], shuffled[val_end:]


def main():
    import argparse

    parser = argparse.ArgumentParser(description="Generate N4L training data using Claude")
    parser.add_argument("--api-key", type=str, default=os.environ.get("ANTHROPIC_API_KEY"),
                       help="Anthropic API key")
    parser.add_argument("--output-dir", type=str, default="data/claude_generated",
                       help="Output directory")
    parser.add_argument("--n4l-dir", type=str, default="../examples",
                       help="Directory with existing N4L files")
    parser.add_argument("--num-per-domain", type=int, default=50,
                       help="Number of synthetic examples per domain")
    parser.add_argument("--model", type=str, default="claude-sonnet-4-20250514",
                       help="Claude model to use")
    parser.add_argument("--create-splits", action="store_true",
                       help="Create train/val/test splits")
    parser.add_argument("--skip-existing", action="store_true",
                       help="Skip processing existing N4L files")

    args = parser.parse_args()

    if not args.api_key:
        print("Error: ANTHROPIC_API_KEY not set")
        print("Usage: export ANTHROPIC_API_KEY=your_key")
        return

    generator = ClaudeN4LGenerator(
        api_key=args.api_key,
        output_dir=args.output_dir,
        model=args.model,
        existing_n4l_dir=args.n4l_dir if not args.skip_existing else None
    )

    # Generate dataset
    examples = generator.generate_dataset(
        num_synthetic_per_domain=args.num_per_domain,
        use_existing=not args.skip_existing
    )

    # Save all examples
    generator.save_examples(examples, "all_examples.jsonl")

    # Create splits
    if args.create_splits:
        train, val, test = generator.create_splits(examples)

        splits_dir = generator.output_dir / "splits"
        splits_dir.mkdir(exist_ok=True)

        generator.output_dir = splits_dir
        generator.save_examples(train, "train.jsonl")
        generator.save_examples(val, "val.jsonl")
        generator.save_examples(test, "test.jsonl")

        print(f"\n=== Statistiques du dataset ===")
        print(f"Total: {len(examples)} exemples")
        print(f"Train: {len(train)} exemples")
        print(f"Val: {len(val)} exemples")
        print(f"Test: {len(test)} exemples")

        # Domain distribution
        domain_counts = {}
        for ex in examples:
            domain_counts[ex.domain] = domain_counts.get(ex.domain, 0) + 1

        print(f"\n=== Distribution par domaine ===")
        for domain, count in sorted(domain_counts.items()):
            print(f"  {domain}: {count}")


if __name__ == "__main__":
    main()
