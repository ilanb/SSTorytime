#!/usr/bin/env python3
"""
Massive Template Generator for N4L Dataset - Target: 10,000+ examples

Generates high-variation template-based examples across 20 domains.
Each domain has multiple sub-templates for maximum diversity.
"""

import json
import random
import hashlib
from pathlib import Path
from typing import List, Dict, Tuple, Callable, Optional
from dataclasses import dataclass
import logging
from tqdm import tqdm
from concurrent.futures import ProcessPoolExecutor, as_completed
import multiprocessing

logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')
logger = logging.getLogger(__name__)


@dataclass
class TrainingExample:
    instruction: str
    input: str
    output: str
    domain: str
    source: str = "template_massive"

    def to_dict(self) -> Dict:
        return {
            "instruction": self.instruction,
            "input": self.input,
            "output": self.output,
            "domain": self.domain,
            "source": self.source
        }


# =====================================================
# EXTENSIVE DATA POOLS FOR HIGH VARIATION
# =====================================================

INSTRUCTION_VARIANTS = [
    "Convertis ce texte en format N4L structuré.",
    "Transforme ce récit en notes N4L.",
    "Analyse ce texte et produis une représentation N4L.",
    "Structure les informations au format N4L.",
    "Génère un fichier N4L à partir de ce contenu.",
    "Extrait les entités et relations en N4L.",
    "Organise ces informations en N4L.",
    "Crée une note N4L structurée.",
    "Produis une représentation N4L de ce texte.",
    "Convertis en N4L avec sections et relations.",
    "Transforme en format de notes structurées N4L.",
    "Analyse et structure en N4L.",
    "Génère la version N4L de ce contenu.",
    "Crée un document N4L à partir de ce texte.",
    "Restructure ce texte au format N4L.",
]

# Extended French names (200+)
FIRST_NAMES = [
    "Jean", "Marie", "Pierre", "Sophie", "Antoine", "Claire", "Thomas", "Julie",
    "Nicolas", "Camille", "François", "Isabelle", "Michel", "Catherine", "Philippe",
    "Anne", "Laurent", "Nathalie", "Éric", "Valérie", "Olivier", "Christine",
    "Patrick", "Sylvie", "Alain", "Martine", "Bruno", "Sandrine", "Christophe", "Céline",
    "David", "Stéphanie", "Sébastien", "Véronique", "Frédéric", "Aurélie", "Guillaume",
    "Émilie", "Julien", "Caroline", "Maxime", "Laure", "Alexandre", "Marine",
    "Romain", "Pauline", "Benjamin", "Charlotte", "Vincent", "Margaux", "Florian",
    "Léa", "Adrien", "Chloé", "Mathieu", "Emma", "Lucas", "Inès", "Hugo", "Manon",
    "Théo", "Sarah", "Louis", "Jade", "Nathan", "Louise", "Raphaël", "Alice",
    "Gabriel", "Anna", "Arthur", "Eva", "Paul", "Lola", "Étienne", "Clara",
    "Victor", "Zoé", "Martin", "Juliette", "Léo", "Rose", "Adam", "Lucie",
    "Charles", "Nina", "Simon", "Élise", "Jules", "Agathe", "Axel", "Victoire",
    "Tom", "Anaïs", "Clément", "Ambre", "Baptiste", "Océane", "Enzo", "Mathilde",
    "Quentin", "Jeanne", "Alexis", "Romane", "Valentin", "Clémence", "Antoine", "Maëlle"
]

LAST_NAMES = [
    "Martin", "Bernard", "Dubois", "Thomas", "Robert", "Richard", "Petit", "Durand",
    "Leroy", "Moreau", "Simon", "Laurent", "Lefebvre", "Michel", "Garcia", "David",
    "Bertrand", "Roux", "Vincent", "Fournier", "Morel", "Girard", "Andre", "Lefevre",
    "Mercier", "Dupont", "Lambert", "Bonnet", "François", "Martinez", "Legrand", "Garnier",
    "Faure", "Rousseau", "Blanc", "Guerin", "Muller", "Henry", "Roussel", "Nicolas",
    "Perrin", "Morin", "Mathieu", "Clement", "Gauthier", "Dumont", "Lopez", "Fontaine",
    "Chevalier", "Robin", "Masson", "Sanchez", "Gerard", "Nguyen", "Boyer", "Denis",
    "Lemaire", "Duval", "Joly", "Gautier", "Roger", "Roche", "Roy", "Noel",
    "Meyer", "Lucas", "Meunier", "Jean", "Perez", "Marchand", "Dufour", "Blanchard",
    "Marie", "Barbier", "Brun", "Dumas", "Brunet", "Schmitt", "Leroux", "Colin",
    "Fernandez", "Pierre", "Renard", "Arnaud", "Rolland", "Caron", "Aubert", "Giraud",
    "Leclerc", "Vidal", "Bourgeois", "Renaud", "Lemoine", "Picard", "Gaillard", "Philippe"
]

CITIES_FR = [
    "Paris", "Lyon", "Marseille", "Bordeaux", "Toulouse", "Nantes", "Strasbourg", "Lille",
    "Nice", "Rennes", "Montpellier", "Grenoble", "Dijon", "Angers", "Reims", "Le Havre",
    "Saint-Étienne", "Toulon", "Clermont-Ferrand", "Orléans", "Rouen", "Metz", "Caen",
    "Nancy", "Tours", "Limoges", "Amiens", "Perpignan", "Besançon", "Brest", "Le Mans",
    "Aix-en-Provence", "Villeurbanne", "Nîmes", "Clermont", "Saint-Denis", "Argenteuil",
    "Montreuil", "Boulogne-Billancourt", "Versailles", "Cannes", "Antibes", "La Rochelle"
]

CITIES_WORLD = [
    "New York", "Londres", "Tokyo", "Berlin", "Rome", "Madrid", "Amsterdam", "Bruxelles",
    "Zurich", "Genève", "Vienne", "Prague", "Varsovie", "Moscou", "Stockholm", "Oslo",
    "Copenhague", "Dublin", "Lisbonne", "Athènes", "Istanbul", "Dubai", "Singapour",
    "Hong Kong", "Sydney", "Melbourne", "Toronto", "Vancouver", "Los Angeles", "San Francisco",
    "Chicago", "Miami", "São Paulo", "Buenos Aires", "Mexico", "Shanghai", "Pékin", "Séoul"
]

MONTHS_FR = ["janvier", "février", "mars", "avril", "mai", "juin",
             "juillet", "août", "septembre", "octobre", "novembre", "décembre"]


def random_name() -> str:
    return f"{random.choice(FIRST_NAMES)} {random.choice(LAST_NAMES)}"


def random_date(year_range=(2020, 2025)) -> str:
    year = random.randint(*year_range)
    month = random.randint(1, 12)
    day = random.randint(1, 28)
    return f"{day:02d}/{month:02d}/{year}"


def random_date_text(year_range=(2020, 2025)) -> str:
    year = random.randint(*year_range)
    month = random.choice(MONTHS_FR)
    day = random.randint(1, 28)
    return f"{day} {month} {year}"


def random_time() -> str:
    hour = random.randint(6, 23)
    minute = random.choice([0, 15, 30, 45])
    return f"{hour}h{minute:02d}"


def random_phone() -> str:
    return f"06 {random.randint(10,99)} {random.randint(10,99)} {random.randint(10,99)} {random.randint(10,99)}"


def random_email(name: str) -> str:
    first, last = name.lower().split()[0], name.lower().split()[1]
    domains = ["gmail.com", "outlook.com", "yahoo.fr", "orange.fr", "free.fr", "entreprise.fr"]
    return f"{first}.{last}@{random.choice(domains)}"


# =====================================================
# DOMAIN TEMPLATES - 20 DOMAINS
# =====================================================

def template_investigation_murder(seed: int) -> Tuple[str, str]:
    """Murder investigation"""
    random.seed(seed)

    victim = random_name()
    victim_age = random.randint(25, 75)
    suspects = [random_name() for _ in range(random.randint(2, 5))]
    detective = random_name()
    forensic = random_name()
    location = random.choice(["appartement", "villa", "hôtel", "bureau", "entrepôt", "parking souterrain"])
    city = random.choice(CITIES_FR)
    weapon = random.choice(["arme blanche", "arme à feu", "poison", "strangulation", "objet contondant"])
    date = random_date()
    time = random_time()
    mobiles = ["héritage", "jalousie", "vengeance", "dette", "secret compromettant"]

    text = f"""Homicide à {city} : l'affaire {victim.split()[1]}

Le {date} à {time}, le corps de {victim}, {victim_age} ans, a été découvert dans son {location}
situé au cœur de {city}. L'inspecteur {detective} de la brigade criminelle a été immédiatement
dépêché sur les lieux, accompagné du médecin légiste {forensic}.

Les premiers éléments de l'enquête révèlent que la victime a succombé à ses blessures causées par
{weapon}. Aucun signe d'effraction n'a été constaté, ce qui suggère que la victime connaissait
son agresseur.

L'enquête a permis d'identifier plusieurs personnes d'intérêt :
{chr(10).join(f"- {s}, mobile potentiel : {random.choice(mobiles)}" for s in suspects[:3])}

Des prélèvements ADN et des empreintes ont été collectés sur la scène de crime. L'autopsie
complète est prévue dans les prochaines heures. L'enquête se poursuit."""

    n4l = f"""- Affaire {victim.split()[1]}

:: Métadonnées ::

Affaire (type) Homicide
    "   (statut) En cours d'investigation
    "   (lieu) {city}
    "   (date) {date}
    "   (heure découverte) {time}

---

:: Victime ::

@victim {victim}

$victim.1 (âge) {victim_age} ans
    "     (lieu découverte) {location}
    "     (cause décès) {weapon}
    "     (date décès) {date}

---

:: Enquêteurs ::

{detective} (fonction) Inspecteur brigade criminelle
    "       (rôle) Enquêteur principal

{forensic} (fonction) Médecin légiste
    "      (mission) Autopsie

---

:: Suspects ::

{chr(10).join(f'''{s} (statut) Personne d'intérêt
    "  (mobile potentiel) {random.choice(mobiles)}''' for s in suspects[:3])}

---

:: Preuves ::

Scène crime (prélèvements) ADN
    "       (prélèvements) Empreintes digitales
    "       (constat) Pas d'effraction

---

:: Chronologie ::

+:: _timeline_ ::
{date} {time} -> Découverte corps -> {location}
{date} -> Arrivée enquêteurs -> {detective}
{date} -> Prélèvements -> Scène de crime
-:: _timeline_ ::
"""
    return text, n4l


def template_investigation_theft(seed: int) -> Tuple[str, str]:
    """Theft investigation"""
    random.seed(seed)

    victim = random_name()
    location = random.choice(["bijouterie", "banque", "musée", "galerie d'art", "coffre-fort", "domicile"])
    city = random.choice(CITIES_FR)
    stolen_items = random.choice([
        ("bijoux", f"{random.randint(50, 500)}000€"),
        ("œuvres d'art", f"{random.randint(1, 10)} millions €"),
        ("espèces", f"{random.randint(100, 900)}000€"),
        ("documents confidentiels", "valeur inestimable")
    ])
    detective = random_name()
    date = random_date()
    time = random_time()
    method = random.choice(["effraction nocturne", "ruse et déguisement", "complicité interne", "tunnel souterrain"])

    text = f"""Vol spectaculaire à {city}

Un cambriolage audacieux a eu lieu le {date} vers {time} dans une {location} de {city}.
Les malfaiteurs ont dérobé {stolen_items[0]} d'une valeur estimée à {stolen_items[1]}.

L'inspecteur {detective} est chargé de l'enquête. Selon les premiers éléments, les voleurs
ont utilisé la méthode suivante : {method}. Les caméras de surveillance ont été neutralisées
et aucun témoin direct n'a été identifié pour le moment.

La propriétaire, {victim}, a déposé plainte dans la matinée. Les experts en criminalistique
analysent actuellement les indices laissés sur place."""

    n4l = f"""- Vol {location} {city}

:: Métadonnées ::

Affaire (type) Vol qualifié
    "   (lieu) {location}
    "   (ville) {city}
    "   (date) {date}
    "   (heure) {time}

---

:: Butin ::

Vol (nature) {stolen_items[0].capitalize()}
   "(valeur) {stolen_items[1]}
   "(méthode) {method.capitalize()}

---

:: Victime ::

{victim} (statut) Plaignant
    "    (préjudice) {stolen_items[1]}

---

:: Enquête ::

{detective} (fonction) Inspecteur
    "       (mission) Enquête vol

Indices (analyse) En cours
Caméras (statut) Neutralisées
Témoins (statut) Aucun identifié

---

:: Chronologie ::

+:: _timeline_ ::
{date} {time} -> Vol commis -> {location}
{date} -> Dépôt plainte -> {victim}
{date} -> Ouverture enquête -> {detective}
-:: _timeline_ ::
"""
    return text, n4l


def template_biography_artist(seed: int) -> Tuple[str, str]:
    """Artist biography"""
    random.seed(seed)

    name = random_name()
    birth_year = random.randint(1940, 1995)
    birth_city = random.choice(CITIES_FR)
    art_form = random.choice(["peinture", "sculpture", "photographie", "musique", "cinéma", "littérature"])
    style = random.choice(["contemporain", "abstrait", "réaliste", "impressionniste", "expressionniste", "minimaliste"])
    awards = random.sample(["César", "Prix Goncourt", "Victoire de la Musique", "Palme d'Or",
                           "Prix Femina", "Grand Prix de Rome", "Molière"], k=random.randint(1, 3))
    galleries = random.sample(["Louvre", "Centre Pompidou", "Musée d'Orsay", "Grand Palais",
                              "Fondation Louis Vuitton", "Palais de Tokyo"], k=random.randint(1, 2))

    death_year = birth_year + random.randint(60, 90) if random.random() > 0.6 else None

    text = f"""{name} : figure majeure de la {art_form} {style}

Né(e) le {random.randint(1, 28)} {random.choice(MONTHS_FR)} {birth_year} à {birth_city},
{name} s'est imposé(e) comme l'un(e) des artistes les plus influent(e)s de sa génération
dans le domaine de la {art_form} {style}.

Formé(e) aux Beaux-Arts de Paris, {name.split()[0]} développe rapidement un style unique
qui lui vaut une reconnaissance internationale. Ses œuvres ont été exposées dans les plus
grandes institutions, notamment {' et '.join(galleries)}.

Au cours de sa carrière, {name} a reçu de nombreuses distinctions : {', '.join(awards)}.
{"Décédé(e) en " + str(death_year) + ", son" if death_year else "Son"} œuvre continue
d'influencer les nouvelles générations d'artistes."""

    n4l = f"""- Biographie {name}

:: Identité ::

@artist {name}

$artist.1 (naissance) {birth_year}
    "     (lieu naissance) {birth_city}
    "     (domaine) {art_form.capitalize()}
    "     (style) {style.capitalize()}
    "     (formation) Beaux-Arts de Paris
{f'    "     (décès) {death_year}' if death_year else '    "     (statut) En activité'}

---

:: Carrière ::

$artist.1 (expositions) {', '.join(galleries)}
{chr(10).join(f'    "     (distinction) {a}' for a in awards)}

---

:: Œuvre ::

$artist.1 (influence) Nouvelles générations
    "     (reconnaissance) Internationale

---

:: Chronologie ::

+:: _timeline_ ::
{birth_year} -> Naissance -> {birth_city}
{birth_year + 20} -> Formation -> Beaux-Arts de Paris
{birth_year + 30} -> Premières expositions -> {galleries[0]}
{f"{death_year} -> Décès ->" if death_year else ""}
-:: _timeline_ ::
"""
    return text, n4l


def template_biography_scientist(seed: int) -> Tuple[str, str]:
    """Scientist biography"""
    random.seed(seed)

    name = f"Dr. {random_name()}"
    birth_year = random.randint(1930, 1980)
    birth_city = random.choice(CITIES_FR + CITIES_WORLD[:10])
    field = random.choice(["physique quantique", "biologie moléculaire", "neurosciences",
                          "astrophysique", "chimie organique", "génétique", "informatique théorique"])
    discovery = random.choice(["nouvelle particule", "mécanisme cellulaire", "algorithme révolutionnaire",
                              "molécule thérapeutique", "théorie unifiée", "gène responsable"])
    institutions = random.sample(["CNRS", "MIT", "Stanford", "Cambridge", "Max Planck", "CERN",
                                 "Institut Pasteur", "INSERM"], k=2)
    awards = random.choice(["Prix Nobel", "Médaille Fields", "Prix Abel", "Prix Turing", "Médaille CNRS"])
    publications = random.randint(50, 300)

    text = f"""{name} : pionnier de la {field}

{name}, né(e) en {birth_year} à {birth_city}, est un(e) scientifique de renommée mondiale
spécialisé(e) en {field}. Après des études brillantes, {name.split()[-1]} rejoint {institutions[0]}
où il/elle mène des recherches fondamentales.

Sa découverte majeure, la mise en évidence d'un(e) {discovery}, a révolutionné le domaine en {birth_year + 35}.
Ces travaux lui ont valu le prestigieux {awards} en {birth_year + 40}.

Actuellement rattaché(e) à {institutions[1]}, {name.split()[-1]} a publié plus de {publications} articles
dans des revues internationales à comité de lecture. Son influence sur la communauté scientifique
est considérable."""

    n4l = f"""- Biographie {name}

:: Identité ::

@scientist {name}

$scientist.1 (naissance) {birth_year}
    "        (lieu naissance) {birth_city}
    "        (spécialité) {field.capitalize()}
    "        (affiliation) {institutions[1]}

---

:: Recherche ::

$scientist.1 (découverte majeure) {discovery.capitalize()}
    "        (année découverte) {birth_year + 35}
    "        (publications) {publications}+ articles

---

:: Distinctions ::

$scientist.1 (prix) {awards}
    "        (année prix) {birth_year + 40}

---

:: Parcours ::

+:: _timeline_ ::
{birth_year} -> Naissance -> {birth_city}
{birth_year + 25} -> Début carrière -> {institutions[0]}
{birth_year + 35} -> Découverte majeure -> {discovery}
{birth_year + 40} -> {awards} -> Reconnaissance internationale
-:: _timeline_ ::
"""
    return text, n4l


def template_project_software(seed: int) -> Tuple[str, str]:
    """Software project"""
    random.seed(seed)

    project = f"Projet {random.choice(['Phoenix', 'Atlas', 'Nexus', 'Horizon', 'Quantum', 'Nova', 'Zenith', 'Aurora', 'Titan', 'Orion'])}"
    project_type = random.choice(["application mobile", "plateforme SaaS", "API microservices",
                                  "système IA", "application web", "infrastructure cloud"])
    company = f"{random.choice(['Tech', 'Data', 'Cloud', 'Digital', 'Smart', 'Cyber'])}{random.choice(['Vision', 'Lab', 'Core', 'Hub', 'Works', 'Soft'])}"
    manager = random_name()
    team_size = random.randint(5, 25)
    tech_stack = random.sample(["Python", "Go", "Rust", "TypeScript", "React", "Vue.js", "Node.js",
                                "PostgreSQL", "MongoDB", "Redis", "Kubernetes", "Docker", "AWS", "GCP"], k=4)
    budget = random.randint(100, 2000) * 1000
    start_date = random_date((2024, 2024))
    deadline = random_date((2025, 2026))
    status = random.choice(["planification", "développement", "test", "beta", "production"])

    text = f"""{project} : développement d'une {project_type}

{company} lance {project}, une initiative ambitieuse visant à créer une {project_type} innovante.
Sous la direction de {manager}, une équipe de {team_size} développeurs travaille sur ce projet
depuis le {start_date}.

Stack technique retenue : {', '.join(tech_stack[:3])} avec {tech_stack[3]} pour l'infrastructure.
Le budget alloué s'élève à {budget:,}€ et la mise en production est prévue pour le {deadline}.

Actuellement en phase de {status}, le projet avance conformément au planning. Les objectifs
principaux sont l'amélioration de l'expérience utilisateur et l'optimisation des performances."""

    n4l = f"""- {project}

:: Informations générales ::

{project} (type) {project_type.capitalize()}
    "     (entreprise) {company}
    "     (statut) {status.capitalize()}
    "     (date début) {start_date}
    "     (deadline) {deadline}
    "     (budget) {budget:,}€

---

:: Équipe ::

@pm {manager}

$pm.1 (rôle) Chef de projet
    " (équipe) {team_size} développeurs

---

:: Stack technique ::

{project} (backend) {tech_stack[0]}
    "     (backend) {tech_stack[1]}
    "     (frontend) {tech_stack[2]}
    "     (infrastructure) {tech_stack[3]}

---

:: Objectifs ::

Objectif 1 (description) Amélioration UX
Objectif 2 (description) Optimisation performances

---

:: Chronologie ::

+:: _timeline_ ::
{start_date} -> Lancement -> {manager}
{deadline} -> Production prévue -> Déploiement
-:: _timeline_ ::
"""
    return text, n4l


def template_project_construction(seed: int) -> Tuple[str, str]:
    """Construction project"""
    random.seed(seed)

    project = f"{random.choice(['Tour', 'Résidence', 'Centre', 'Complexe', 'Parc'])} {random.choice(['Horizon', 'Lumière', 'Avenir', 'Harmonie', 'Prestige'])}"
    project_type = random.choice(["immeuble de bureaux", "résidence de standing", "centre commercial",
                                  "hôpital", "école", "infrastructure sportive"])
    city = random.choice(CITIES_FR)
    architect = random_name()
    company = f"{random.choice(['Bouygues', 'Vinci', 'Eiffage', 'Nexity', 'Groupe'])} Construction"
    surface = random.randint(5, 50) * 1000
    floors = random.randint(3, 40)
    budget = random.randint(10, 500) * 1000000
    start_date = random_date((2023, 2024))
    end_date = random_date((2026, 2028))
    workers = random.randint(50, 500)

    text = f"""{project} : nouveau {project_type} à {city}

Le projet {project} prévoit la construction d'un {project_type} de {surface:,}m² sur {floors} étages
dans le quartier en développement de {city}. L'architecte {architect} a conçu ce bâtiment innovant
pour le compte de {company}.

Les travaux, débutés le {start_date}, mobilisent actuellement {workers} ouvriers sur le chantier.
La livraison est prévue pour le {end_date}, pour un budget total de {budget/1000000:.1f} millions d'euros.

Ce projet emblématique intègre les dernières normes environnementales HQE et vise la certification
BREEAM Excellent."""

    n4l = f"""- {project}

:: Description ::

{project} (type) {project_type.capitalize()}
    "     (ville) {city}
    "     (surface) {surface:,}m²
    "     (étages) {floors}
    "     (budget) {budget/1000000:.1f}M€

---

:: Intervenants ::

{architect} (rôle) Architecte principal
{company} (rôle) Maître d'œuvre
Chantier (effectif) {workers} ouvriers

---

:: Certifications ::

{project} (norme) HQE
    "     (certification visée) BREEAM Excellent

---

:: Planning ::

+:: _timeline_ ::
{start_date} -> Début travaux -> Terrassement
{end_date} -> Livraison prévue -> Réception
-:: _timeline_ ::
"""
    return text, n4l


def template_medical_diagnosis(seed: int) -> Tuple[str, str]:
    """Medical diagnosis"""
    random.seed(seed)

    patient = random_name()
    patient_age = random.randint(20, 85)
    doctor = f"Dr. {random_name()}"
    hospital = random.choice(["CHU de Paris", "Hôpital Necker", "CHU de Lyon", "Hôpital Saint-Louis",
                             "CHU de Bordeaux", "Hôpital Européen", "Clinique du Parc"])
    condition = random.choice(["diabète de type 2", "hypertension artérielle", "insuffisance cardiaque",
                              "asthme sévère", "maladie de Crohn", "polyarthrite rhumatoïde"])
    symptoms = random.sample(["fatigue chronique", "douleurs articulaires", "essoufflement",
                             "perte de poids", "fièvre récurrente", "troubles digestifs"], k=3)
    treatment = random.choice(["traitement médicamenteux", "intervention chirurgicale",
                              "rééducation", "immunothérapie", "changement mode de vie"])
    date = random_date()

    text = f"""Dossier médical - {patient}

Patient(e) : {patient}, {patient_age} ans
Date de consultation : {date}
Établissement : {hospital}
Médecin traitant : {doctor}

Motif de consultation : {', '.join(symptoms[:2])}

Après examen clinique approfondi et analyses complémentaires, le diagnostic de {condition}
a été établi. Le/La patient(e) présente également des symptômes de {symptoms[2]}.

Traitement prescrit : {treatment}
Suivi recommandé : consultations mensuelles avec {doctor}

Pronostic : favorable avec bonne observance du traitement."""

    n4l = f"""- Dossier médical {patient.split()[1]}

:: Patient ::

@patient {patient}

$patient.1 (âge) {patient_age} ans
    "      (date consultation) {date}
    "      (établissement) {hospital}

---

:: Symptômes ::

$patient.1 (symptôme) {symptoms[0].capitalize()}
    "      (symptôme) {symptoms[1].capitalize()}
    "      (symptôme) {symptoms[2].capitalize()}

---

:: Diagnostic ::

$patient.1 (diagnostic) {condition.capitalize()}
    "      (traitement) {treatment.capitalize()}
    "      (pronostic) Favorable

---

:: Suivi ::

{doctor} (rôle) Médecin traitant
    "    (suivi) Consultations mensuelles
"""
    return text, n4l


def template_medical_surgery(seed: int) -> Tuple[str, str]:
    """Surgical procedure"""
    random.seed(seed)

    patient = random_name()
    patient_age = random.randint(30, 75)
    surgeon = f"Pr. {random_name()}"
    anesthetist = f"Dr. {random_name()}"
    hospital = random.choice(["CHU Pitié-Salpêtrière", "Hôpital Georges Pompidou",
                             "CHU de Marseille", "Institut Curie", "Hôpital Cochin"])
    procedure = random.choice(["pontage coronarien", "remplacement valve cardiaque",
                              "résection tumorale", "arthroscopie du genou", "cholécystectomie"])
    duration = random.choice(["2h30", "3h45", "4h15", "5h00", "6h30"])
    date = random_date()
    complications = random.choice(["Aucune", "Mineures", "Hémorragie contrôlée"])
    recovery = random.randint(3, 14)

    text = f"""Compte-rendu opératoire

Patient : {patient}, {patient_age} ans
Date intervention : {date}
Établissement : {hospital}

Chirurgien : {surgeon}
Anesthésiste : {anesthetist}

Intervention : {procedure}
Durée : {duration}

L'intervention s'est déroulée sans incident majeur. Complications per-opératoires : {complications}.
Le/La patient(e) a été transféré(e) en salle de réveil dans un état stable.

Durée d'hospitalisation prévue : {recovery} jours
Suivi post-opératoire : consultations à J+15 et J+30"""

    n4l = f"""- Intervention chirurgicale {patient.split()[1]}

:: Patient ::

@patient {patient}

$patient.1 (âge) {patient_age} ans
    "      (date intervention) {date}
    "      (établissement) {hospital}

---

:: Intervention ::

Opération (type) {procedure.capitalize()}
    "     (durée) {duration}
    "     (complications) {complications}

---

:: Équipe médicale ::

{surgeon} (rôle) Chirurgien principal
{anesthetist} (rôle) Anesthésiste

---

:: Suites ::

$patient.1 (hospitalisation) {recovery} jours
    "      (suivi) J+15 et J+30
    "      (état) Stable
"""
    return text, n4l


def template_legal_contract(seed: int) -> Tuple[str, str]:
    """Contract/Agreement"""
    random.seed(seed)

    party1 = f"{random.choice(['Groupe', 'Société', 'Entreprise'])} {random_name().split()[1]}"
    party2 = f"{random.choice(['SAS', 'SARL', 'SA'])} {random_name().split()[1]}"
    contract_type = random.choice(["prestation de services", "licence logicielle", "partenariat commercial",
                                   "distribution exclusive", "joint-venture", "cession de droits"])
    lawyer1 = f"Me {random_name()}"
    lawyer2 = f"Me {random_name()}"
    amount = random.randint(50, 5000) * 1000
    duration = random.randint(1, 5)
    date = random_date()
    city = random.choice(CITIES_FR)

    text = f"""Contrat de {contract_type}

Entre les parties :
- {party1}, représentée par {lawyer1}
- {party2}, représentée par {lawyer2}

Objet du contrat : {contract_type}
Lieu de signature : {city}
Date : {date}

Durée : {duration} an(s) renouvelable
Montant : {amount:,}€

Les parties conviennent des modalités détaillées dans les annexes ci-jointes.
Juridiction compétente : Tribunal de Commerce de {city}"""

    n4l = f"""- Contrat {contract_type}

:: Parties ::

@party1 {party1}
@party2 {party2}

$party1.1 (représentant) {lawyer1}
$party2.1 (représentant) {lawyer2}

---

:: Termes ::

Contrat (type) {contract_type.capitalize()}
    "   (date signature) {date}
    "   (lieu) {city}
    "   (durée) {duration} an(s)
    "   (montant) {amount:,}€

---

:: Juridiction ::

Contrat (tribunal compétent) Tribunal de Commerce de {city}
"""
    return text, n4l


def template_legal_trial(seed: int) -> Tuple[str, str]:
    """Court trial"""
    random.seed(seed)

    plaintiff = random_name()
    defendant = random_name()
    judge = f"Juge {random_name()}"
    lawyer_p = f"Me {random_name()}"
    lawyer_d = f"Me {random_name()}"
    case_type = random.choice(["litige commercial", "licenciement abusif", "contrefaçon",
                               "diffamation", "rupture de contrat", "concurrence déloyale"])
    court = random.choice(["Tribunal de Commerce", "Conseil de Prud'hommes",
                          "Tribunal Judiciaire", "Cour d'Appel"])
    city = random.choice(CITIES_FR)
    damages = random.randint(10, 500) * 1000
    date_filing = random_date((2023, 2024))
    date_hearing = random_date((2024, 2025))
    verdict = random.choice(["en délibéré", "favorable au demandeur", "favorable au défendeur", "rejet"])

    text = f"""Affaire {plaintiff.split()[1]} c/ {defendant.split()[1]}

Juridiction : {court} de {city}
Type : {case_type}

Demandeur : {plaintiff}, représenté par {lawyer_p}
Défendeur : {defendant}, représenté par {lawyer_d}
Magistrat : {judge}

Date de saisine : {date_filing}
Audience : {date_hearing}

Le demandeur réclame {damages:,}€ de dommages et intérêts.
Décision : {verdict}"""

    n4l = f"""- Affaire {plaintiff.split()[1]} c/ {defendant.split()[1]}

:: Procédure ::

Affaire (type) {case_type.capitalize()}
    "   (juridiction) {court} de {city}
    "   (date saisine) {date_filing}
    "   (audience) {date_hearing}

---

:: Parties ::

@demandeur {plaintiff}
@defendeur {defendant}

$demandeur.1 (avocat) {lawyer_p}
    "        (demande) {damages:,}€

$defendeur.1 (avocat) {lawyer_d}

---

:: Magistrat ::

{judge} (fonction) Juge
    "   (affaire) {plaintiff.split()[1]} c/ {defendant.split()[1]}

---

:: Décision ::

Affaire (verdict) {verdict.capitalize()}
"""
    return text, n4l


def template_event_conference(seed: int) -> Tuple[str, str]:
    """Conference/Event"""
    random.seed(seed)

    event = f"{random.choice(['Forum', 'Conférence', 'Sommet', 'Symposium'])} {random.choice(['Innovation', 'Tech', 'Digital', 'Avenir', 'Leaders'])}"
    theme = random.choice(["intelligence artificielle", "transition écologique", "transformation digitale",
                          "cybersécurité", "finance durable", "santé connectée"])
    venue = random.choice(["Palais des Congrès", "Paris Expo", "Centre de Conventions", "La Défense Arena"])
    city = random.choice(CITIES_FR)
    organizer = f"{random.choice(['Association', 'Fédération', 'Institut'])} {random.choice(['Tech France', 'Digital', 'Innovation'])}"
    speakers = [random_name() for _ in range(random.randint(4, 8))]
    date = random_date((2025, 2025))
    attendees = random.randint(500, 5000)
    price = random.choice([0, 150, 350, 500, 800])

    text = f"""{event} - {theme.capitalize()}

Date : {date}
Lieu : {venue}, {city}
Organisateur : {organizer}

Cet événement majeur réunira {attendees} participants autour du thème "{theme}".

Intervenants confirmés :
{chr(10).join(f"- {s}" for s in speakers[:5])}

{"Entrée gratuite sur inscription" if price == 0 else f"Tarif : {price}€"}

Programme : conférences plénières, ateliers thématiques, networking et exposition."""

    n4l = f"""- {event}

:: Informations ::

{event} (thème) {theme.capitalize()}
    "   (date) {date}
    "   (lieu) {venue}
    "   (ville) {city}
    "   (organisateur) {organizer}

---

:: Participation ::

{event} (participants) {attendees}
    "   (tarif) {"Gratuit" if price == 0 else f"{price}€"}

---

:: Intervenants ::

{chr(10).join(f'{s} (rôle) Speaker' for s in speakers[:5])}

---

:: Programme ::

Session 1 (type) Conférences plénières
Session 2 (type) Ateliers thématiques
Session 3 (type) Networking
"""
    return text, n4l


def template_event_exhibition(seed: int) -> Tuple[str, str]:
    """Art exhibition"""
    random.seed(seed)

    exhibition = f"Exposition {random.choice(['Lumières', 'Horizons', 'Visions', 'Métamorphoses', 'Résonances'])}"
    artist = random_name()
    museum = random.choice(["Musée d'Art Moderne", "Centre Pompidou", "Grand Palais",
                           "Fondation Louis Vuitton", "Palais de Tokyo", "Musée d'Orsay"])
    city = random.choice(CITIES_FR)
    curator = random_name()
    works = random.randint(30, 150)
    start_date = random_date((2025, 2025))
    end_date = random_date((2025, 2026))
    style = random.choice(["contemporain", "impressionniste", "surréaliste", "minimaliste"])

    text = f"""{exhibition} - Rétrospective {artist}

Du {start_date} au {end_date}
{museum}, {city}

Cette exposition majeure présente {works} œuvres de l'artiste {artist}, figure emblématique
de l'art {style}. Le parcours, conçu par le/la commissaire {curator}, retrace l'évolution
artistique sur plus de trois décennies.

Vernissage : {start_date}
Commissaire : {curator}

Horaires : 10h-18h, fermé le mardi
Tarif : {random.choice([12, 14, 16])}€"""

    n4l = f"""- {exhibition}

:: Exposition ::

{exhibition} (artiste) {artist}
    "        (style) {style.capitalize()}
    "        (œuvres) {works}
    "        (commissaire) {curator}

---

:: Lieu et dates ::

{exhibition} (musée) {museum}
    "        (ville) {city}
    "        (début) {start_date}
    "        (fin) {end_date}

---

:: Informations pratiques ::

{exhibition} (horaires) 10h-18h
    "        (fermeture) Mardi
    "        (tarif) {random.choice([12, 14, 16])}€
"""
    return text, n4l


def template_recipe_main(seed: int) -> Tuple[str, str]:
    """Main dish recipe"""
    random.seed(seed)

    dishes = ["bœuf bourguignon", "coq au vin", "blanquette de veau", "cassoulet",
              "bouillabaisse", "pot-au-feu", "navarin d'agneau", "daube provençale"]
    dish = random.choice(dishes)
    chef = random_name()
    prep_time = random.choice([20, 30, 45, 60])
    cook_time = random.choice([90, 120, 150, 180])
    servings = random.randint(4, 8)
    difficulty = random.choice(["facile", "moyen", "difficile"])
    ingredients = random.sample([
        ("viande", f"{random.randint(6, 12)*100}g"),
        ("oignons", f"{random.randint(2, 4)}"),
        ("carottes", f"{random.randint(3, 6)}"),
        ("vin rouge", f"{random.randint(25, 50)}cl"),
        ("bouillon", f"{random.randint(25, 50)}cl"),
        ("champignons", f"{random.randint(200, 400)}g"),
        ("lardons", f"{random.randint(100, 200)}g"),
        ("beurre", f"{random.randint(30, 60)}g"),
        ("farine", f"{random.randint(2, 4)} c.à.s"),
        ("thym", "1 branche"),
        ("laurier", "2 feuilles"),
        ("ail", f"{random.randint(2, 4)} gousses")
    ], k=8)

    text = f"""Recette du {dish}

Par Chef {chef}
Pour {servings} personnes
Préparation : {prep_time} min | Cuisson : {cook_time} min
Difficulté : {difficulty}

Ingrédients :
{chr(10).join(f"- {qty} de {ing}" for ing, qty in ingredients)}

Cette recette traditionnelle française demande un peu de patience mais le résultat
est incomparable. Servir bien chaud avec des pommes de terre vapeur ou du riz."""

    n4l = f"""- Recette {dish}

:: Informations ::

Recette (nom) {dish.capitalize()}
    "   (chef) {chef}
    "   (portions) {servings}
    "   (difficulté) {difficulty.capitalize()}

---

:: Temps ::

Préparation (durée) {prep_time} min
Cuisson (durée) {cook_time} min
Total (durée) {prep_time + cook_time} min

---

:: Ingrédients ::

{chr(10).join(f'{ing.capitalize()} (quantité) {qty}' for ing, qty in ingredients)}

---

:: Accompagnement ::

Recette (suggestion) Pommes de terre vapeur
    "   (suggestion) Riz
"""
    return text, n4l


def template_recipe_dessert(seed: int) -> Tuple[str, str]:
    """Dessert recipe"""
    random.seed(seed)

    desserts = ["tarte tatin", "crème brûlée", "mousse au chocolat", "île flottante",
                "profiteroles", "mille-feuille", "clafoutis", "far breton"]
    dessert = random.choice(desserts)
    chef = random_name()
    prep_time = random.choice([15, 20, 30, 45])
    cook_time = random.choice([20, 30, 45, 60])
    servings = random.randint(4, 8)
    ingredients = random.sample([
        ("œufs", f"{random.randint(3, 6)}"),
        ("sucre", f"{random.randint(100, 200)}g"),
        ("farine", f"{random.randint(100, 200)}g"),
        ("beurre", f"{random.randint(50, 150)}g"),
        ("lait", f"{random.randint(25, 50)}cl"),
        ("crème fraîche", f"{random.randint(20, 30)}cl"),
        ("chocolat noir", f"{random.randint(150, 250)}g"),
        ("vanille", "1 gousse"),
        ("pommes", f"{random.randint(4, 6)}"),
        ("rhum", "2 c.à.s")
    ], k=6)

    text = f"""Recette : {dessert}

Par {chef}
{servings} portions | Préparation {prep_time} min | Cuisson {cook_time} min

Ingrédients :
{chr(10).join(f"- {qty} de {ing}" for ing, qty in ingredients)}

Un dessert classique de la pâtisserie française, à servir tiède ou froid
selon les goûts. Peut se préparer la veille."""

    n4l = f"""- Recette {dessert}

:: Informations ::

Recette (nom) {dessert.capitalize()}
    "   (type) Dessert
    "   (chef) {chef}
    "   (portions) {servings}

---

:: Temps ::

Préparation (durée) {prep_time} min
Cuisson (durée) {cook_time} min

---

:: Ingrédients ::

{chr(10).join(f'{ing.capitalize()} (quantité) {qty}' for ing, qty in ingredients)}

---

:: Service ::

Recette (température) Tiède ou froid
    "   (conservation) Possible la veille
"""
    return text, n4l


def template_company_startup(seed: int) -> Tuple[str, str]:
    """Startup profile"""
    random.seed(seed)

    name = f"{random.choice(['Neo', 'Next', 'Smart', 'Data', 'Cloud', 'AI', 'Tech', 'Digital'])}{random.choice(['Lab', 'Hub', 'Works', 'Soft', 'Mind', 'Flow', 'Pulse'])}"
    sector = random.choice(["fintech", "healthtech", "edtech", "greentech", "proptech", "legaltech"])
    founder = random_name()
    cofounder = random_name()
    city = random.choice(CITIES_FR)
    founded = random.randint(2018, 2023)
    employees = random.randint(10, 150)
    funding = random.choice(["seed", "série A", "série B", "série C"])
    amount = random.randint(1, 50) * 1000000
    investors = random.sample(["BPI France", "Partech", "Kima Ventures", "Alven", "Index Ventures"], k=2)

    text = f"""{name} - Startup {sector}

Fondée en {founded} à {city} par {founder} et {cofounder}, {name} est une startup
spécialisée dans le secteur {sector}. L'entreprise compte aujourd'hui {employees} collaborateurs.

Dernière levée de fonds : {funding} de {amount/1000000:.1f}M€ auprès de {' et '.join(investors)}.

{name} développe des solutions innovantes qui ont déjà séduit plus de {random.randint(100, 10000)} clients.
L'ambition : devenir le leader européen du {sector}."""

    n4l = f"""- Fiche {name}

:: Identité ::

{name} (type) Startup
    " (secteur) {sector.capitalize()}
    " (fondation) {founded}
    " (siège) {city}

---

:: Fondateurs ::

{founder} (rôle) CEO / Co-fondateur
{cofounder} (rôle) CTO / Co-fondateur

---

:: Financement ::

{name} (dernière levée) {funding.capitalize()}
    " (montant) {amount/1000000:.1f}M€
    " (investisseurs) {', '.join(investors)}

---

:: Données ::

{name} (effectif) {employees} personnes
    " (clients) {random.randint(100, 10000)}+
"""
    return text, n4l


def template_company_corporate(seed: int) -> Tuple[str, str]:
    """Large corporation profile"""
    random.seed(seed)

    name = f"{random.choice(['Groupe', 'Société Générale', 'Compagnie', 'Holding'])} {random_name().split()[1]}"
    sector = random.choice(["industrie", "énergie", "télécommunications", "distribution", "banque", "assurance"])
    ceo = random_name()
    city = random.choice(CITIES_FR)
    founded = random.randint(1850, 1990)
    employees = random.randint(5000, 100000)
    revenue = random.randint(1, 50) * 1000000000
    market_cap = random.randint(5, 100) * 1000000000
    listed = random.choice(["CAC 40", "SBF 120", "Euronext Paris"])

    text = f"""{name}

Fondée en {founded}, {name} est un acteur majeur du secteur {sector} en France.
Présidée par {ceo}, l'entreprise emploie {employees:,} collaborateurs et réalise
un chiffre d'affaires de {revenue/1000000000:.1f} milliards d'euros.

Cotée au {listed}, sa capitalisation boursière s'élève à {market_cap/1000000000:.1f} milliards d'euros.
Le siège social est situé à {city}.

{name} poursuit sa stratégie de croissance et de transformation digitale."""

    n4l = f"""- Fiche {name}

:: Identité ::

{name} (fondation) {founded}
    " (siège) {city}
    " (secteur) {sector.capitalize()}
    " (cotation) {listed}

---

:: Direction ::

{ceo} (fonction) Président-Directeur Général

---

:: Données financières ::

{name} (effectif) {employees:,}
    " (CA) {revenue/1000000000:.1f}Md€
    " (capitalisation) {market_cap/1000000000:.1f}Md€
"""
    return text, n4l


def template_scientific_discovery(seed: int) -> Tuple[str, str]:
    """Scientific discovery"""
    random.seed(seed)

    researcher = f"Pr. {random_name()}"
    team = [random_name() for _ in range(random.randint(2, 5))]
    institution = random.choice(["CNRS", "INSERM", "CEA", "INRIA", "Institut Curie", "Institut Pasteur"])
    field = random.choice(["génétique", "physique quantique", "neurosciences", "immunologie", "astrophysique"])
    discovery = random.choice(["nouveau gène", "particule inconnue", "mécanisme neuronal",
                              "anticorps révolutionnaire", "exoplanète habitable"])
    journal = random.choice(["Nature", "Science", "Cell", "The Lancet", "Physical Review Letters"])
    date = random_date()
    impact = random.choice(["traitement maladies génétiques", "informatique quantique",
                           "thérapies neurologiques", "vaccins nouvelle génération"])

    text = f"""Découverte majeure en {field}

Le {date}, l'équipe du {researcher} ({institution}) a annoncé la découverte d'un {discovery},
une avancée qui pourrait révolutionner le domaine de la {field}.

Équipe de recherche : {', '.join(team[:3])}

Les résultats, publiés dans {journal}, ouvrent des perspectives prometteuses pour
le développement de {impact}. Cette découverte est le fruit de {random.randint(3, 10)} années de recherche."""

    n4l = f"""- Découverte {field}

:: Recherche ::

Découverte (domaine) {field.capitalize()}
    "      (nature) {discovery.capitalize()}
    "      (date) {date}
    "      (publication) {journal}

---

:: Équipe ::

{researcher} (rôle) Chercheur principal
    "        (institution) {institution}

{chr(10).join(f'{t} (rôle) Chercheur' for t in team[:3])}

---

:: Impact ::

Découverte (applications) {impact.capitalize()}
    "      (durée recherche) {random.randint(3, 10)} ans
"""
    return text, n4l


def template_scientific_study(seed: int) -> Tuple[str, str]:
    """Scientific study/trial"""
    random.seed(seed)

    lead = f"Dr. {random_name()}"
    institution = random.choice(["CHU de Paris", "Institut Gustave Roussy", "CHU de Lyon",
                                "Hôpital Necker", "Institut Curie"])
    study_type = random.choice(["étude clinique", "essai de phase III", "méta-analyse",
                               "étude épidémiologique", "essai randomisé"])
    subject = random.choice(["nouveau traitement cancer", "vaccin expérimental",
                            "thérapie génique", "médicament anti-inflammatoire"])
    participants = random.randint(100, 5000)
    duration = random.randint(1, 5)
    results = random.choice(["positifs", "prometteurs", "significatifs", "encourageants"])
    efficacy = random.randint(60, 95)

    text = f"""{study_type.capitalize()} : {subject}

Responsable : {lead}
Institution : {institution}

Cette {study_type} portant sur {participants} participants a évalué l'efficacité d'un {subject}.
Durée de l'étude : {duration} an(s).

Résultats : {results} avec un taux d'efficacité de {efficacy}%.

Ces données seront présentées lors du prochain congrès international."""

    n4l = f"""- {study_type.capitalize()} {subject}

:: Étude ::

Étude (type) {study_type.capitalize()}
    " (sujet) {subject.capitalize()}
    " (participants) {participants}
    " (durée) {duration} an(s)

---

:: Direction ::

{lead} (rôle) Investigateur principal
    " (institution) {institution}

---

:: Résultats ::

Étude (résultats) {results.capitalize()}
    " (efficacité) {efficacy}%
"""
    return text, n4l


def template_real_estate_sale(seed: int) -> Tuple[str, str]:
    """Real estate sale listing"""
    random.seed(seed)

    prop_type = random.choice(["appartement", "maison", "villa", "loft", "duplex", "penthouse"])
    city = random.choice(CITIES_FR)
    neighborhood = random.choice(["centre-ville", "quartier résidentiel", "proche gare",
                                 "bord de mer", "quartier historique"])
    surface = random.randint(30, 300)
    rooms = random.randint(1, 8)
    bedrooms = max(1, rooms - 2)
    price = random.randint(100, 2000) * 1000
    year = random.randint(1900, 2023)
    agent = random_name()
    agency = f"Immobilier {random.choice(['Premium', 'Conseil', 'Expert', 'Select'])}"
    features = random.sample(["balcon", "terrasse", "jardin", "parking", "cave", "ascenseur",
                             "gardien", "piscine", "vue mer", "lumineux"], k=4)
    energy = random.choice(["A", "B", "C", "D"])

    text = f"""{prop_type.capitalize()} à vendre - {city}

Localisation : {neighborhood}, {city}
Surface : {surface}m² | {rooms} pièces dont {bedrooms} chambres
Année construction : {year}
DPE : {energy}

Prix : {price:,}€

Caractéristiques : {', '.join(features)}

Contact : {agent} - {agency}
Tél : {random_phone()}"""

    n4l = f"""- Bien immobilier {city}

:: Description ::

Bien (type) {prop_type.capitalize()}
   "(surface) {surface}m²
   "(pièces) {rooms}
   "(chambres) {bedrooms}
   "(année) {year}
   "(DPE) {energy}
   "(prix) {price:,}€

---

:: Localisation ::

Bien (ville) {city}
   "(quartier) {neighborhood.capitalize()}

---

:: Caractéristiques ::

{chr(10).join(f'Bien (équipement) {f.capitalize()}' for f in features)}

---

:: Contact ::

{agent} (fonction) Agent immobilier
    "  (agence) {agency}
"""
    return text, n4l


def template_real_estate_rental(seed: int) -> Tuple[str, str]:
    """Rental listing"""
    random.seed(seed)

    prop_type = random.choice(["studio", "T2", "T3", "T4", "maison"])
    city = random.choice(CITIES_FR)
    surface = random.randint(20, 150)
    rent = random.randint(500, 3000)
    charges = random.randint(50, 200)
    deposit = rent * 2
    available = random_date((2025, 2025))
    furnished = random.choice([True, False])
    features = random.sample(["meublé", "parking", "cave", "balcon", "ascenseur", "digicode"], k=3)

    text = f"""{prop_type} à louer - {city}

Surface : {surface}m²
Loyer : {rent}€/mois (charges : {charges}€)
Dépôt de garantie : {deposit}€
{"Meublé" if furnished else "Non meublé"}

Disponible : {available}

Équipements : {', '.join(features)}"""

    n4l = f"""- Location {prop_type} {city}

:: Bien ::

Location (type) {prop_type}
    "   (ville) {city}
    "   (surface) {surface}m²
    "   (meublé) {"Oui" if furnished else "Non"}

---

:: Finances ::

Location (loyer) {rent}€/mois
    "   (charges) {charges}€
    "   (dépôt) {deposit}€

---

:: Disponibilité ::

Location (disponible) {available}

---

:: Équipements ::

{chr(10).join(f'Location (équipement) {f.capitalize()}' for f in features)}
"""
    return text, n4l


def template_education_course(seed: int) -> Tuple[str, str]:
    """Educational course"""
    random.seed(seed)

    course = f"{random.choice(['Formation', 'Certificat', 'Master', 'MBA'])} {random.choice(['Data Science', 'Management', 'Marketing Digital', 'Finance', 'IA'])}"
    institution = random.choice(["HEC Paris", "ESSEC", "Sciences Po", "Polytechnique",
                                "CentraleSupélec", "INSEAD"])
    director = f"Pr. {random_name()}"
    duration = random.choice(["6 mois", "1 an", "2 ans", "18 mois"])
    format_type = random.choice(["présentiel", "hybride", "100% en ligne"])
    price = random.randint(5, 50) * 1000
    start_date = random_date((2025, 2025))
    places = random.randint(20, 100)

    text = f"""{course}

Institution : {institution}
Directeur pédagogique : {director}

Durée : {duration}
Format : {format_type}
Prochaine session : {start_date}
Places disponibles : {places}

Tarif : {price:,}€

Programme complet disponible sur demande. Financement CPF possible."""

    n4l = f"""- {course}

:: Formation ::

{course} (institution) {institution}
    "   (durée) {duration}
    "   (format) {format_type}
    "   (tarif) {price:,}€

---

:: Organisation ::

{director} (rôle) Directeur pédagogique

{course} (début) {start_date}
    "   (places) {places}

---

:: Financement ::

{course} (CPF) Éligible
"""
    return text, n4l


def template_education_school(seed: int) -> Tuple[str, str]:
    """School/University profile"""
    random.seed(seed)

    school = f"{random.choice(['École', 'Institut', 'Université'])} {random.choice(['Supérieure', 'Nationale', 'Internationale'])} de {random.choice(['Commerce', 'Sciences', 'Technologie', 'Arts'])}"
    city = random.choice(CITIES_FR)
    president = f"Pr. {random_name()}"
    founded = random.randint(1850, 2000)
    students = random.randint(1000, 20000)
    programs = random.randint(10, 50)
    ranking = random.randint(1, 50)
    accreditations = random.sample(["AACSB", "EQUIS", "AMBA", "CTI", "EESPIG"], k=2)

    text = f"""{school}

Fondée en {founded} à {city}
Président : {president}

Effectif étudiant : {students:,}
Programmes : {programs}
Classement national : #{ranking}

Accréditations : {', '.join(accreditations)}

{school} forme les leaders de demain dans un environnement international d'excellence."""

    n4l = f"""- {school}

:: Identité ::

{school} (fondation) {founded}
    "    (ville) {city}
    "    (président) {president}

---

:: Données ::

{school} (étudiants) {students:,}
    "    (programmes) {programs}
    "    (classement) #{ranking} national

---

:: Accréditations ::

{chr(10).join(f'{school} (accréditation) {a}' for a in accreditations)}
"""
    return text, n4l


def template_sports_match(seed: int) -> Tuple[str, str]:
    """Sports match report"""
    random.seed(seed)

    sport = random.choice(["football", "rugby", "basketball", "handball", "tennis"])
    team1 = f"{random.choice(['Paris', 'Lyon', 'Marseille', 'Bordeaux', 'Lille'])} {random.choice(['FC', 'SG', 'United', 'Racing'])}"
    team2 = f"{random.choice(['Monaco', 'Nice', 'Nantes', 'Rennes', 'Strasbourg'])} {random.choice(['AS', 'OGC', 'FC', 'Stade'])}"
    score1 = random.randint(0, 5)
    score2 = random.randint(0, 5)
    stadium = f"Stade {random.choice(['Municipal', 'Olympique', 'de France', 'Vélodrome'])}"
    date = random_date()
    attendance = random.randint(10000, 80000)
    scorers = [random_name().split()[1] for _ in range(score1 + score2)]
    competition = random.choice(["Ligue 1", "Coupe de France", "Champions League", "Championnat"])

    text = f"""{team1} vs {team2} - {competition}

Date : {date}
Stade : {stadium}
Affluence : {attendance:,} spectateurs

Score final : {team1} {score1} - {score2} {team2}

{"Buteurs : " + ', '.join(scorers[:3]) if scorers else "Match nul et vierge"}

Résumé : {"Victoire de " + (team1 if score1 > score2 else team2) if score1 != score2 else "Match nul"}"""

    n4l = f"""- Match {team1} vs {team2}

:: Rencontre ::

Match (compétition) {competition}
    " (date) {date}
    " (stade) {stadium}
    " (affluence) {attendance:,}

---

:: Équipes ::

{team1} (score) {score1}
{team2} (score) {score2}

---

:: Buteurs ::

{chr(10).join(f'{s} (but) 1' for s in scorers[:3]) if scorers else "Aucun but"}

---

:: Résultat ::

Match (issue) {"Victoire " + (team1 if score1 > score2 else team2) if score1 != score2 else "Match nul"}
"""
    return text, n4l


def template_sports_athlete(seed: int) -> Tuple[str, str]:
    """Athlete profile"""
    random.seed(seed)

    athlete = random_name()
    sport = random.choice(["tennis", "natation", "athlétisme", "cyclisme", "judo", "escrime"])
    birth_year = random.randint(1985, 2005)
    birth_city = random.choice(CITIES_FR)
    club = f"{random.choice(['Racing', 'Paris', 'Lyon', 'Nice'])} {random.choice(['Club', 'Sport', 'Athlétique'])}"
    coach = random_name()
    titles = random.sample(["Champion de France", "Médaille olympique", "Champion du Monde",
                           "Vainqueur Grand Chelem", "Record national"], k=random.randint(1, 3))
    ranking = random.randint(1, 100)

    text = f"""{athlete} - {sport.capitalize()}

Né(e) en {birth_year} à {birth_city}
Club : {club}
Entraîneur : {coach}
Classement mondial : #{ranking}

Palmarès :
{chr(10).join(f"- {t}" for t in titles)}

{athlete.split()[0]} est considéré(e) comme l'un(e) des meilleur(e)s athlètes français(es) de {sport}."""

    n4l = f"""- Fiche {athlete}

:: Identité ::

@athlete {athlete}

$athlete.1 (sport) {sport.capitalize()}
    "      (naissance) {birth_year}
    "      (ville) {birth_city}
    "      (club) {club}

---

:: Encadrement ::

{coach} (rôle) Entraîneur
    "  (athlète) $athlete.1

---

:: Palmarès ::

$athlete.1 (classement) #{ranking} mondial
{chr(10).join(f'    "      (titre) {t}' for t in titles)}
"""
    return text, n4l


def template_travel_destination(seed: int) -> Tuple[str, str]:
    """Travel destination"""
    random.seed(seed)

    destination = random.choice(CITIES_WORLD + CITIES_FR)
    country = random.choice(["France", "Italie", "Espagne", "Japon", "USA", "Grèce", "Portugal"])
    attractions = random.sample(["musées", "plages", "monuments historiques", "gastronomie",
                                "vie nocturne", "nature", "shopping", "architecture"], k=4)
    best_season = random.choice(["printemps", "été", "automne", "hiver"])
    budget = random.choice(["économique", "modéré", "confortable", "luxe"])
    duration = random.choice(["3 jours", "1 semaine", "10 jours", "2 semaines"])

    text = f"""Guide : {destination}

Pays : {country}
Meilleure saison : {best_season}
Durée recommandée : {duration}
Budget : {budget}

À voir/faire :
{chr(10).join(f"- {a.capitalize()}" for a in attractions)}

{destination} est une destination incontournable pour les amateurs de {attractions[0]} et {attractions[1]}."""

    n4l = f"""- Destination {destination}

:: Informations ::

{destination} (pays) {country}
    "        (saison) {best_season.capitalize()}
    "        (durée conseillée) {duration}
    "        (budget) {budget.capitalize()}

---

:: Attractions ::

{chr(10).join(f'{destination} (activité) {a.capitalize()}' for a in attractions)}
"""
    return text, n4l


def template_travel_hotel(seed: int) -> Tuple[str, str]:
    """Hotel listing"""
    random.seed(seed)

    hotel = f"Hôtel {random.choice(['Le', 'Grand', 'Royal', 'Palace'])} {random.choice(['Majestic', 'Splendide', 'Riviera', 'Continental'])}"
    city = random.choice(CITIES_FR + CITIES_WORLD[:10])
    stars = random.randint(3, 5)
    rooms = random.randint(50, 300)
    price = random.randint(80, 500)
    amenities = random.sample(["piscine", "spa", "restaurant gastronomique", "salle de sport",
                              "bar rooftop", "parking", "conciergerie", "room service 24h"], k=4)
    rating = round(random.uniform(7.5, 9.8), 1)

    text = f"""{hotel} - {city}

Catégorie : {stars} étoiles
Chambres : {rooms}
À partir de : {price}€/nuit

Note clients : {rating}/10

Services :
{chr(10).join(f"- {a.capitalize()}" for a in amenities)}

Réservation en ligne disponible."""

    n4l = f"""- {hotel}

:: Établissement ::

{hotel} (ville) {city}
    "   (catégorie) {stars} étoiles
    "   (chambres) {rooms}
    "   (tarif) À partir de {price}€/nuit
    "   (note) {rating}/10

---

:: Services ::

{chr(10).join(f'{hotel} (service) {a.capitalize()}' for a in amenities)}
"""
    return text, n4l


# =====================================================
# MAIN GENERATOR
# =====================================================

TEMPLATE_GENERATORS: Dict[str, List[Callable]] = {
    "investigation": [template_investigation_murder, template_investigation_theft],
    "biography": [template_biography_artist, template_biography_scientist],
    "project": [template_project_software, template_project_construction],
    "medical": [template_medical_diagnosis, template_medical_surgery],
    "legal": [template_legal_contract, template_legal_trial],
    "event": [template_event_conference, template_event_exhibition],
    "recipe": [template_recipe_main, template_recipe_dessert],
    "company": [template_company_startup, template_company_corporate],
    "scientific": [template_scientific_discovery, template_scientific_study],
    "real_estate": [template_real_estate_sale, template_real_estate_rental],
    "education": [template_education_course, template_education_school],
    "sports": [template_sports_match, template_sports_athlete],
    "travel": [template_travel_destination, template_travel_hotel],
}


def generate_example(args: Tuple[str, int]) -> Optional[TrainingExample]:
    """Generate a single example (for parallel processing)"""
    domain, seed = args
    try:
        generators = TEMPLATE_GENERATORS[domain]
        generator = generators[seed % len(generators)]
        text, n4l = generator(seed)

        return TrainingExample(
            instruction=random.choice(INSTRUCTION_VARIANTS),
            input=text.strip(),
            output=n4l.strip(),
            domain=domain,
            source="template_massive"
        )
    except Exception as e:
        logger.error(f"Error generating {domain} example {seed}: {e}")
        return None


def generate_all_examples(num_per_domain: int = 500, parallel: bool = True) -> List[TrainingExample]:
    """Generate all examples across all domains"""
    examples = []
    domains = list(TEMPLATE_GENERATORS.keys())

    # Create list of (domain, seed) pairs
    tasks = []
    for domain in domains:
        for i in range(num_per_domain):
            seed = hash(f"{domain}_{i}") % (2**31)
            tasks.append((domain, seed))

    logger.info(f"Generating {len(tasks)} examples across {len(domains)} domains...")

    if parallel and len(tasks) > 100:
        # Use multiprocessing for large generations
        num_workers = min(multiprocessing.cpu_count(), 8)
        with ProcessPoolExecutor(max_workers=num_workers) as executor:
            results = list(tqdm(executor.map(generate_example, tasks), total=len(tasks), desc="Generating"))
            examples = [r for r in results if r is not None]
    else:
        # Sequential for smaller batches
        for task in tqdm(tasks, desc="Generating"):
            result = generate_example(task)
            if result:
                examples.append(result)

    return examples


def save_examples(examples: List[TrainingExample], output_path: Path):
    """Save examples to JSONL"""
    with open(output_path, 'w', encoding='utf-8') as f:
        for ex in examples:
            f.write(json.dumps(ex.to_dict(), ensure_ascii=False) + '\n')
    logger.info(f"Saved {len(examples)} examples to {output_path}")


def create_splits(examples: List[TrainingExample], output_dir: Path,
                  train_ratio: float = 0.8, val_ratio: float = 0.1):
    """Create train/val/test splits"""
    random.seed(42)
    shuffled = examples.copy()
    random.shuffle(shuffled)

    n = len(shuffled)
    train_end = int(n * train_ratio)
    val_end = train_end + int(n * val_ratio)

    splits = {
        "train": shuffled[:train_end],
        "val": shuffled[train_end:val_end],
        "test": shuffled[val_end:]
    }

    output_dir.mkdir(parents=True, exist_ok=True)

    for name, data in splits.items():
        save_examples(data, output_dir / f"{name}.jsonl")

    return splits


def main():
    import argparse

    parser = argparse.ArgumentParser(description="Generate massive N4L dataset")
    parser.add_argument("--num-per-domain", type=int, default=500,
                       help="Number of examples per domain (default: 500)")
    parser.add_argument("--output", type=str, default="data/massive",
                       help="Output directory")
    parser.add_argument("--no-parallel", action="store_true",
                       help="Disable parallel processing")
    parser.add_argument("--create-splits", action="store_true", default=True,
                       help="Create train/val/test splits")

    args = parser.parse_args()

    output_dir = Path(args.output)
    output_dir.mkdir(parents=True, exist_ok=True)

    # Generate
    examples = generate_all_examples(
        num_per_domain=args.num_per_domain,
        parallel=not args.no_parallel
    )

    # Save all
    save_examples(examples, output_dir / "all_examples.jsonl")

    # Create splits
    if args.create_splits:
        splits = create_splits(examples, output_dir / "splits")

    # Statistics
    print(f"\n{'='*50}")
    print("GENERATION COMPLETE")
    print(f"{'='*50}")
    print(f"Total examples: {len(examples)}")

    if args.create_splits:
        print(f"\nSplits:")
        for name, data in splits.items():
            print(f"  {name}: {len(data)}")

    print(f"\nBy domain:")
    domain_counts = {}
    for ex in examples:
        domain_counts[ex.domain] = domain_counts.get(ex.domain, 0) + 1
    for domain, count in sorted(domain_counts.items()):
        print(f"  {domain}: {count}")

    print(f"\nOutput: {output_dir}")


if __name__ == "__main__":
    main()
