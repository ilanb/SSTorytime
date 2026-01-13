#!/usr/bin/env python3
"""
Enhanced Template Generator for N4L Dataset

Generates ~1000+ template-based examples across 10 domains with high variation.
This is meant to complement Claude-generated synthetic data.
"""

import json
import random
from pathlib import Path
from typing import List, Dict, Tuple
from dataclasses import dataclass
import logging
from tqdm import tqdm

logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')
logger = logging.getLogger(__name__)


@dataclass
class TrainingExample:
    instruction: str
    input: str
    output: str
    domain: str
    source: str = "template"

    def to_dict(self) -> Dict:
        return {
            "instruction": self.instruction,
            "input": self.input,
            "output": self.output,
            "domain": self.domain,
            "source": self.source
        }


INSTRUCTION_VARIANTS = [
    "Convertis ce texte en format N4L structuré.",
    "Transforme ce récit en notes N4L.",
    "Analyse ce texte et produis une représentation N4L.",
    "Structure les informations au format N4L.",
    "Génère un fichier N4L à partir de ce contenu.",
    "Extrait les entités et relations en N4L.",
    "Organise ces informations en N4L.",
    "Crée une note N4L structurée.",
]

# Extended name pools
FIRST_NAMES_FR = ["Jean", "Marie", "Pierre", "Sophie", "Antoine", "Claire", "Thomas", "Julie",
                   "Nicolas", "Camille", "François", "Isabelle", "Michel", "Catherine", "Philippe",
                   "Anne", "Laurent", "Nathalie", "Éric", "Valérie", "Olivier", "Christine",
                   "Patrick", "Sylvie", "Alain", "Martine", "Bruno", "Sandrine", "Christophe", "Céline"]

LAST_NAMES_FR = ["Martin", "Dubois", "Bernard", "Petit", "Robert", "Richard", "Durand", "Moreau",
                  "Laurent", "Simon", "Michel", "Lefebvre", "Leroy", "Roux", "David", "Bertrand",
                  "Morel", "Fournier", "Girard", "Bonnet", "Dupont", "Lambert", "Fontaine", "Rousseau",
                  "Vincent", "Muller", "Lefevre", "Faure", "Andre", "Mercier"]

CITIES_FR = ["Paris", "Lyon", "Marseille", "Bordeaux", "Toulouse", "Nantes", "Strasbourg", "Lille",
             "Nice", "Rennes", "Montpellier", "Grenoble", "Dijon", "Angers", "Reims", "Le Havre",
             "Saint-Étienne", "Toulon", "Clermont-Ferrand", "Orléans"]


def random_name() -> str:
    return f"{random.choice(FIRST_NAMES_FR)} {random.choice(LAST_NAMES_FR)}"


def random_date(year_range=(2020, 2025)) -> str:
    year = random.randint(*year_range)
    month = random.randint(1, 12)
    day = random.randint(1, 28)
    return f"{day:02d}/{month:02d}/{year}"


def random_time() -> str:
    hour = random.randint(6, 23)
    minute = random.choice([0, 15, 30, 45])
    return f"{hour}h{minute:02d}"


# ============== DOMAIN TEMPLATES ==============

def template_investigation(seed: int) -> Tuple[str, str]:
    """Criminal investigation template"""
    random.seed(seed)

    victim = random_name()
    suspects = [random_name() for _ in range(random.randint(2, 4))]
    detective = random_name()
    location_types = ["appartement", "bureau", "restaurant", "parc", "parking", "hôtel",
                      "entrepôt", "villa", "studio", "garage"]
    location = random.choice(location_types)
    city = random.choice(CITIES_FR)
    crime_types = ["homicide", "vol", "agression", "empoisonnement", "enlèvement"]
    crime = random.choice(crime_types)
    mobiles = ["financier", "passionnel", "vengeance", "jalousie", "héritage", "professionnel"]
    evidence_types = ["ADN", "empreintes digitales", "vidéosurveillance", "témoignages", "documents"]
    date = random_date()
    time = random_time()
    age = random.randint(25, 75)

    # Narrative
    text = f"""L'affaire {victim.split()[1]} : un {crime} à {city}

Le {date}, les autorités ont découvert le corps de {victim}, {age} ans, dans un {location} situé à {city}.
L'inspecteur {detective} a été chargé de l'enquête.

Les premiers éléments indiquent que la victime a été retrouvée vers {time}. L'autopsie préliminaire suggère
que le décès remonte à quelques heures avant la découverte.

Plusieurs personnes d'intérêt ont été identifiées :
- {suspects[0]}, qui avait un mobile {random.choice(mobiles)}
- {suspects[1]}, présent(e) dans les environs au moment des faits
{f"- {suspects[2]}, en conflit avec la victime depuis plusieurs mois" if len(suspects) > 2 else ""}

Les enquêteurs ont collecté plusieurs preuves : {random.choice(evidence_types)} et {random.choice(evidence_types)}.
L'enquête se poursuit activement."""

    # N4L
    n4l = f"""- Affaire {victim.split()[1]}

:: Métadonnées ::

Affaire (type) {crime.capitalize()}
    "   (statut) En cours
    "   (lieu) {city}
    "   (date ouverture) {date}

---

:: Victime ::

@victim {victim}

$victim.1 (âge) {age} ans
    "     (lieu découverte) {location}
    "     (heure découverte) {time}
    "     (date décès) {date}

---

:: Enquêteur ::

{detective} (rôle) Inspecteur principal
    "       (affecté à) Affaire {victim.split()[1]}

---

:: Suspects ::

{suspects[0]} (statut) Personne d'intérêt
    "        (mobile) {random.choice(mobiles)}

{suspects[1]} (statut) Personne d'intérêt
    "        (présence) Environs au moment des faits
{f'''
{suspects[2]} (statut) Personne d'intérêt
    "        (relation victime) Conflit''' if len(suspects) > 2 else ""}

---

:: Chronologie ::

+:: _timeline_ ::
{date} {time} -> Découverte du corps -> {location}, {city}
{date} -> Ouverture enquête -> Inspecteur {detective}
{date} -> Identification suspects -> {len(suspects)} personnes d'intérêt
-:: _timeline_ ::
"""
    return text, n4l


def template_biography(seed: int) -> Tuple[str, str]:
    """Biography template"""
    random.seed(seed)

    name = random_name()
    birth_city = random.choice(CITIES_FR)
    birth_year = random.randint(1920, 1995)
    professions = ["écrivain", "scientifique", "artiste peintre", "musicien", "architecte",
                   "médecin", "avocat", "ingénieur", "journaliste", "réalisateur"]
    profession = random.choice(professions)
    awards = ["Prix Nobel", "Légion d'honneur", "Prix Goncourt", "César", "Palme d'Or",
              "Médaille Fields", "Prix Femina", "Grand Prix de Rome"]
    award = random.choice(awards)
    universities = ["Sorbonne", "École Polytechnique", "ENS", "Sciences Po", "HEC"]
    university = random.choice(universities)
    works = ["plusieurs ouvrages majeurs", "des contributions révolutionnaires",
             "de nombreuses innovations", "des œuvres marquantes"]
    contributions = random.choice(works)

    death_year = birth_year + random.randint(60, 95) if random.random() > 0.5 else None
    death_info = f"décédé(e) en {death_year}" if death_year else "toujours en activité"

    text = f"""{name} : une vie dédiée à l'excellence

{name} est né(e) le {random.randint(1,28)} {random.choice(['janvier', 'février', 'mars', 'avril', 'mai', 'juin', 'juillet', 'août', 'septembre', 'octobre', 'novembre', 'décembre'])} {birth_year} à {birth_city}.

Après des études à {university}, {name.split()[0]} s'est orienté(e) vers une carrière de {profession}.
Au fil des années, cette personnalité remarquable a réalisé {contributions} dans son domaine.

En {birth_year + random.randint(30, 50)}, {name} a reçu le prestigieux {award} en reconnaissance de son travail exceptionnel.

{f"Décédé(e) en {death_year}," if death_year else "Aujourd'hui,"} {name.split()[0]} laisse un héritage considérable
qui continue d'influencer les générations actuelles."""

    n4l = f"""- Biographie de {name}

:: Identité ::

@person {name}

$person.1 (né en) {birth_year}
    "     (lieu naissance) {birth_city}
    "     (profession) {profession}
    "     (formation) {university}
{f'    "     (décès) {death_year}' if death_year else '    "     (statut) En activité'}

---

:: Carrière ::

$person.1 (contribution) {contributions.replace("de nombreuses ", "").replace("des ", "").capitalize()}
    "     (distinction) {award}
    "     (année distinction) {birth_year + random.randint(30, 50)}

---

:: Chronologie ::

+:: _timeline_ ::
{birth_year} -> Naissance -> {birth_city}
{birth_year + 20} -> Formation -> {university}
{birth_year + 25} -> Début carrière -> {profession}
{birth_year + random.randint(30, 50)} -> Distinction -> {award}
{f"{death_year} -> Décès ->" if death_year else ""}
-:: _timeline_ ::
"""
    return text, n4l


def template_project(seed: int) -> Tuple[str, str]:
    """Tech project template"""
    random.seed(seed)

    project_names = ["Phoenix", "Atlas", "Nexus", "Horizon", "Zenith", "Titan", "Aurora",
                     "Quantum", "Neptune", "Orion", "Falcon", "Matrix"]
    project = f"Projet {random.choice(project_names)}"
    types = ["application mobile", "plateforme web", "API REST", "système IA", "infrastructure cloud",
             "solution IoT", "application desktop", "microservices", "data pipeline"]
    project_type = random.choice(types)
    teams = ["Alpha", "Beta", "Gamma", "Delta", "Omega", "Innovation", "R&D"]
    team = f"Équipe {random.choice(teams)}"
    manager = random_name()
    tech_stacks = [
        ["Python", "FastAPI", "PostgreSQL"],
        ["Go", "gRPC", "Redis"],
        ["React", "Node.js", "MongoDB"],
        ["Rust", "Kubernetes", "Kafka"],
        ["Java", "Spring Boot", "MySQL"],
    ]
    tech = random.choice(tech_stacks)
    statuses = ["planification", "développement", "test", "pré-production", "production"]
    status = random.choice(statuses)
    budget = random.randint(50, 500) * 1000
    team_size = random.randint(3, 15)
    start_date = random_date((2024, 2024))
    deadline = random_date((2025, 2026))

    text = f"""{project} : développement d'une {project_type}

Le {project} est une initiative stratégique visant à créer une {project_type} innovante.
Lancé le {start_date}, ce projet est piloté par {manager} à la tête de l'{team}.

L'équipe de {team_size} personnes utilise une stack technique moderne comprenant {tech[0]}, {tech[1]} et {tech[2]}.
Le budget alloué s'élève à {budget:,}€ et la date de livraison est fixée au {deadline}.

Actuellement en phase de {status}, le projet avance conformément au planning établi.
Les objectifs principaux incluent l'amélioration de l'expérience utilisateur et l'optimisation des performances."""

    n4l = f"""- {project}

:: Métadonnées ::

{project} (type) {project_type.capitalize()}
    "     (statut) {status.capitalize()}
    "     (date début) {start_date}
    "     (deadline) {deadline}
    "     (budget) {budget:,}€

---

:: Équipe ::

@pm {manager}

$pm.1 (rôle) Chef de projet
    " (équipe) {team}

{team} (taille) {team_size} personnes
    " (responsable) $pm.1

---

:: Stack technique ::

{project} (utilise) {tech[0]}
    "     (utilise) {tech[1]}
    "     (utilise) {tech[2]}

---

:: Objectifs ::

Objectif 1 (description) Amélioration expérience utilisateur
Objectif 2 (description) Optimisation performances

---

:: Chronologie ::

+:: _timeline_ ::
{start_date} -> Lancement projet -> {manager}
{deadline} -> Livraison prévue -> Production
-:: _timeline_ ::
"""
    return text, n4l


def template_medical(seed: int) -> Tuple[str, str]:
    """Medical case template"""
    random.seed(seed)

    patient = random_name()
    doctor = f"Dr. {random_name()}"
    age = random.randint(20, 80)
    conditions = ["diabète de type 2", "hypertension artérielle", "insuffisance cardiaque",
                  "asthme chronique", "arthrite rhumatoïde", "maladie de Crohn",
                  "sclérose en plaques", "fibromyalgie"]
    condition = random.choice(conditions)
    hospitals = ["CHU de Paris", "Hôpital Saint-Louis", "Hôpital Necker", "CHU de Lyon",
                 "Hôpital Européen", "Clinique du Parc"]
    hospital = random.choice(hospitals)
    treatments = ["traitement médicamenteux", "intervention chirurgicale", "rééducation",
                  "immunothérapie", "chimiothérapie", "thérapie ciblée"]
    treatment = random.choice(treatments)
    admission_date = random_date()
    symptoms = ["fatigue chronique", "douleurs articulaires", "essoufflement",
                "perte de poids", "troubles du sommeil"]

    text = f"""Dossier médical : {patient}

Patient(e) : {patient}, {age} ans
Établissement : {hospital}
Médecin traitant : {doctor}

Motif de consultation : {random.choice(symptoms)} et {random.choice(symptoms)}

Diagnostic établi le {admission_date} : {condition}

Le/La patient(e) présente un tableau clinique typique de {condition}.
Après examens complémentaires, un {treatment} a été prescrit.

Le suivi sera assuré par {doctor} avec des consultations mensuelles.
Pronostic : favorable avec une bonne observance du traitement."""

    n4l = f"""- Dossier médical {patient.split()[1]}

:: Patient ::

@patient {patient}

$patient.1 (âge) {age} ans
    "      (diagnostic) {condition}
    "      (date diagnostic) {admission_date}
    "      (établissement) {hospital}

---

:: Équipe médicale ::

{doctor} (rôle) Médecin traitant
    "    (spécialité) Médecine interne
    "    (patient) $patient.1

---

:: Traitement ::

$patient.1 (traitement) {treatment}
    "      (suivi) Consultations mensuelles
    "      (pronostic) Favorable

---

:: Chronologie ::

+:: _timeline_ ::
{admission_date} -> Diagnostic -> {condition}
{admission_date} -> Début traitement -> {treatment}
-:: _timeline_ ::
"""
    return text, n4l


def template_legal(seed: int) -> Tuple[str, str]:
    """Legal case template"""
    random.seed(seed)

    plaintiff = random_name()
    defendant = random_name()
    judge = f"Juge {random_name()}"
    lawyer1 = f"Me {random_name()}"
    lawyer2 = f"Me {random_name()}"
    case_types = ["litige commercial", "conflit de travail", "affaire de propriété intellectuelle",
                  "contentieux fiscal", "divorce", "succession"]
    case_type = random.choice(case_types)
    courts = ["Tribunal de Commerce", "Conseil de Prud'hommes", "Tribunal Judiciaire",
              "Cour d'Appel", "Tribunal Administratif"]
    court = random.choice(courts)
    cities = random.choice(CITIES_FR)
    filing_date = random_date((2023, 2024))
    amounts = [random.randint(10, 500) * 1000 for _ in range(3)]
    status = random.choice(["en cours", "jugement rendu", "appel en cours"])

    text = f"""Affaire {plaintiff.split()[1]} c. {defendant.split()[1]}

Juridiction : {court} de {cities}
Type d'affaire : {case_type}
Date de saisine : {filing_date}

Demandeur : {plaintiff}, représenté(e) par {lawyer1}
Défendeur : {defendant}, représenté(e) par {lawyer2}

Le/La demandeur(esse) réclame {amounts[0]:,}€ de dommages et intérêts.
L'affaire porte sur un {case_type} opposant les deux parties depuis plusieurs mois.

Le {judge} est en charge du dossier. Statut actuel : {status}.
Une audience est prévue pour examiner les pièces du dossier."""

    n4l = f"""- Affaire {plaintiff.split()[1]} c. {defendant.split()[1]}

:: Informations générales ::

Affaire (type) {case_type.capitalize()}
    "   (juridiction) {court} de {cities}
    "   (date saisine) {filing_date}
    "   (statut) {status.capitalize()}

---

:: Parties ::

@demandeur {plaintiff}
@defendeur {defendant}

$demandeur.1 (rôle) Demandeur
    "        (avocat) {lawyer1}
    "        (demande) {amounts[0]:,}€

$defendeur.1 (rôle) Défendeur
    "        (avocat) {lawyer2}

---

:: Magistrat ::

{judge} (fonction) Juge
    "   (juridiction) {court}
    "   (affaire) {plaintiff.split()[1]} c. {defendant.split()[1]}

---

:: Chronologie ::

+:: _timeline_ ::
{filing_date} -> Saisine tribunal -> {court}
{filing_date} -> Dépôt demande -> {amounts[0]:,}€
-:: _timeline_ ::
"""
    return text, n4l


def template_event(seed: int) -> Tuple[str, str]:
    """Event/Conference template"""
    random.seed(seed)

    event_types = ["conférence", "salon professionnel", "séminaire", "workshop",
                   "forum", "symposium", "hackathon"]
    event_type = random.choice(event_types)
    themes = ["intelligence artificielle", "développement durable", "innovation technologique",
              "transformation digitale", "cybersécurité", "santé connectée"]
    theme = random.choice(themes)
    organizers = ["Association Tech France", "Chambre de Commerce", "Institut Innovation",
                  "Fondation Numérique", "Cluster Digital"]
    organizer = random.choice(organizers)
    venues = ["Palais des Congrès", "Centre de Conventions", "Parc des Expositions",
              "Campus Innovation", "Hôtel Marriott"]
    venue = random.choice(venues)
    city = random.choice(CITIES_FR)
    speakers = [random_name() for _ in range(random.randint(3, 5))]
    date = random_date((2025, 2025))
    attendees = random.randint(100, 2000)
    price = random.choice([0, 50, 150, 300, 500])

    text = f"""{event_type.capitalize()} : {theme}

Date : {date}
Lieu : {venue}, {city}
Organisateur : {organizer}

Cet événement réunira {attendees} participants autour du thème "{theme}".

Intervenants confirmés :
{chr(10).join(f"- {s}" for s in speakers)}

{"Entrée gratuite" if price == 0 else f"Tarif d'inscription : {price}€"}

Le programme comprendra des conférences plénières, des ateliers pratiques et des sessions de networking.
Une occasion unique de découvrir les dernières tendances et d'échanger avec des experts du domaine."""

    n4l = f"""- {event_type.capitalize()} {theme}

:: Informations ::

Événement (type) {event_type.capitalize()}
    "     (thème) {theme}
    "     (date) {date}
    "     (lieu) {venue}
    "     (ville) {city}
    "     (organisateur) {organizer}

---

:: Participants ::

Événement (participants attendus) {attendees}
    "     (tarif) {"Gratuit" if price == 0 else f"{price}€"}

---

:: Intervenants ::

{chr(10).join(f'{s} (rôle) Intervenant' for s in speakers)}

---

:: Programme ::

Session 1 (type) Conférences plénières
Session 2 (type) Ateliers pratiques
Session 3 (type) Networking
"""
    return text, n4l


def template_recipe(seed: int) -> Tuple[str, str]:
    """Recipe template"""
    random.seed(seed)

    dishes = ["boeuf bourguignon", "quiche lorraine", "ratatouille", "coq au vin",
              "tarte tatin", "blanquette de veau", "gratin dauphinois", "cassoulet"]
    dish = random.choice(dishes)
    prep_time = random.choice([15, 20, 30, 45, 60])
    cook_time = random.choice([30, 45, 60, 90, 120])
    servings = random.randint(4, 8)
    difficulty = random.choice(["facile", "moyen", "difficile"])
    ingredients_pool = [
        ("beurre", "50g"), ("farine", "200g"), ("œufs", "3"), ("lait", "25cl"),
        ("crème fraîche", "20cl"), ("oignon", "2"), ("ail", "3 gousses"),
        ("sel", "1 c.à.c"), ("poivre", "1 pincée"), ("huile d'olive", "3 c.à.s")
    ]
    ingredients = random.sample(ingredients_pool, random.randint(5, 8))
    chef = random_name()

    text = f"""Recette du {dish}

Par {chef}
Temps de préparation : {prep_time} minutes
Temps de cuisson : {cook_time} minutes
Pour {servings} personnes
Difficulté : {difficulty}

Ingrédients :
{chr(10).join(f"- {qty} de {ing}" for ing, qty in ingredients)}

Cette recette traditionnelle française est un classique de la gastronomie.
Suivez les étapes avec attention pour obtenir un plat savoureux et authentique.
Servir chaud, accompagné d'un bon vin rouge."""

    n4l = f"""- Recette {dish}

:: Informations ::

Recette (nom) {dish.capitalize()}
    "   (auteur) {chef}
    "   (difficulté) {difficulty.capitalize()}
    "   (portions) {servings} personnes

---

:: Temps ::

Préparation (durée) {prep_time} minutes
Cuisson (durée) {cook_time} minutes
Total (durée) {prep_time + cook_time} minutes

---

:: Ingrédients ::

{chr(10).join(f'{ing.capitalize()} (quantité) {qty}' for ing, qty in ingredients)}

---

:: Service ::

Recette (accompagnement) Vin rouge
    "   (température) Chaud
"""
    return text, n4l


def template_company(seed: int) -> Tuple[str, str]:
    """Company profile template"""
    random.seed(seed)

    company_names = ["TechVision", "InnoSoft", "DataPulse", "CloudNine", "NexGen",
                     "SmartLab", "DigitalEdge", "FutureTech", "CyberCore", "AIpex"]
    company = f"{random.choice(company_names)} SAS"
    sectors = ["technologie", "finance", "santé", "énergie", "retail", "industrie"]
    sector = random.choice(sectors)
    ceo = random_name()
    city = random.choice(CITIES_FR)
    founded = random.randint(2000, 2020)
    employees = random.randint(10, 500)
    revenue = random.randint(1, 50) * 1000000
    products = ["solutions cloud", "applications mobiles", "services IA",
                "plateformes SaaS", "consulting IT"]
    product = random.choice(products)

    text = f"""{company} - Profil entreprise

{company} est une entreprise française fondée en {founded} à {city}.
Dirigée par {ceo} (PDG), l'entreprise compte aujourd'hui {employees} collaborateurs.

Secteur d'activité : {sector}
Chiffre d'affaires : {revenue/1000000:.1f}M€

L'entreprise est spécialisée dans les {product} et accompagne ses clients
dans leur transformation digitale. Sa croissance régulière témoigne de
son expertise reconnue sur le marché."""

    n4l = f"""- Fiche {company}

:: Identité ::

{company} (fondation) {founded}
    "     (siège) {city}
    "     (secteur) {sector.capitalize()}
    "     (forme juridique) SAS

---

:: Direction ::

{ceo} (fonction) PDG
    " (entreprise) {company}

---

:: Données financières ::

{company} (effectif) {employees} personnes
    "     (CA) {revenue/1000000:.1f}M€
    "     (activité) {product.capitalize()}

---

:: Chronologie ::

+:: _timeline_ ::
{founded} -> Création -> {city}
-:: _timeline_ ::
"""
    return text, n4l


def template_scientific(seed: int) -> Tuple[str, str]:
    """Scientific research template"""
    random.seed(seed)

    researcher = f"Dr. {random_name()}"
    fields = ["physique quantique", "biologie moléculaire", "intelligence artificielle",
              "neurosciences", "climatologie", "génétique", "astronomie"]
    field = random.choice(fields)
    institutions = ["CNRS", "INSERM", "CEA", "INRIA", "Institut Pasteur", "INRAE"]
    institution = random.choice(institutions)
    discoveries = ["nouvelle molécule", "algorithme révolutionnaire", "mécanisme cellulaire",
                   "particule subatomique", "méthode d'analyse", "modèle prédictif"]
    discovery = random.choice(discoveries)
    journals = ["Nature", "Science", "Cell", "The Lancet", "Physical Review"]
    journal = random.choice(journals)
    date = random_date((2023, 2025))
    funding = random.randint(100, 2000) * 1000

    text = f"""Recherche en {field} : une avancée majeure

{researcher}, chercheur(se) à {institution}, a publié une étude révolutionnaire
sur {discovery} dans le domaine de la {field}.

Les résultats, publiés le {date} dans {journal}, ouvrent de nouvelles perspectives
pour la recherche fondamentale et appliquée.

Ce projet, financé à hauteur de {funding:,}€, a mobilisé une équipe de {random.randint(5, 20)} chercheurs
pendant {random.randint(2, 5)} ans. Les applications potentielles sont nombreuses et prometteuses."""

    n4l = f"""- Recherche {field}

:: Chercheur principal ::

{researcher} (affiliation) {institution}
    "        (domaine) {field.capitalize()}
    "        (découverte) {discovery.capitalize()}

---

:: Publication ::

Article (journal) {journal}
    "   (date) {date}
    "   (sujet) {discovery.capitalize()}

---

:: Financement ::

Projet (budget) {funding:,}€
    " (source) {institution}
    " (durée) {random.randint(2, 5)} ans

---

:: Chronologie ::

+:: _timeline_ ::
{date} -> Publication -> {journal}
-:: _timeline_ ::
"""
    return text, n4l


def template_real_estate(seed: int) -> Tuple[str, str]:
    """Real estate listing template"""
    random.seed(seed)

    property_types = ["appartement", "maison", "studio", "loft", "villa", "duplex"]
    prop_type = random.choice(property_types)
    city = random.choice(CITIES_FR)
    neighborhoods = ["centre-ville", "quartier historique", "zone résidentielle",
                     "proche gare", "bord de mer", "quartier d'affaires"]
    neighborhood = random.choice(neighborhoods)
    surface = random.randint(25, 200)
    rooms = random.randint(1, 6)
    price = random.randint(100, 800) * 1000
    agent = random_name()
    agency = f"Immobilier {random.choice(['Premium', 'Plus', 'Expert', 'Conseil'])}"
    year = random.randint(1900, 2023)
    features = ["balcon", "parking", "cave", "terrasse", "jardin", "ascenseur"]

    text = f"""{prop_type.capitalize()} à vendre - {city}

Référence : {random.randint(1000, 9999)}
Localisation : {neighborhood}, {city}

Surface : {surface}m²
Nombre de pièces : {rooms}
Année de construction : {year}
Prix : {price:,}€

Ce {prop_type} dispose de nombreux atouts : {random.choice(features)}, {random.choice(features)}.
Idéalement situé en {neighborhood}, proche de toutes commodités.

Contact : {agent}
Agence : {agency}"""

    n4l = f"""- Bien immobilier {city}

:: Description ::

Bien (type) {prop_type.capitalize()}
    "(surface) {surface}m²
    "(pièces) {rooms}
    "(année) {year}
    "(prix) {price:,}€

---

:: Localisation ::

Bien (ville) {city}
    "(quartier) {neighborhood.capitalize()}

---

:: Contact ::

{agent} (fonction) Agent immobilier
    "   (agence) {agency}

---

:: Caractéristiques ::

Bien (équipement) {random.choice(features).capitalize()}
    "(équipement) {random.choice(features).capitalize()}
"""
    return text, n4l


# Main generator function
TEMPLATE_GENERATORS = {
    "investigation": template_investigation,
    "biography": template_biography,
    "project": template_project,
    "medical": template_medical,
    "legal": template_legal,
    "event": template_event,
    "recipe": template_recipe,
    "company": template_company,
    "scientific": template_scientific,
    "real_estate": template_real_estate,
}


def generate_all_templates(num_per_domain: int = 100) -> List[TrainingExample]:
    """Generate templates for all domains"""
    examples = []

    for domain, generator in TEMPLATE_GENERATORS.items():
        logger.info(f"Generating {num_per_domain} {domain} templates...")

        for i in tqdm(range(num_per_domain), desc=domain):
            try:
                text, n4l = generator(seed=i * 1000 + hash(domain) % 1000)
                example = TrainingExample(
                    instruction=random.choice(INSTRUCTION_VARIANTS),
                    input=text.strip(),
                    output=n4l.strip(),
                    domain=domain,
                    source="template_enhanced"
                )
                examples.append(example)
            except Exception as e:
                logger.error(f"Error generating {domain} template {i}: {e}")

    return examples


def save_examples(examples: List[TrainingExample], output_path: Path):
    """Save examples to JSONL"""
    with open(output_path, 'w', encoding='utf-8') as f:
        for ex in examples:
            f.write(json.dumps(ex.to_dict(), ensure_ascii=False) + '\n')
    logger.info(f"Saved {len(examples)} examples to {output_path}")


def main():
    import argparse

    parser = argparse.ArgumentParser(description="Generate enhanced N4L templates")
    parser.add_argument("--num-per-domain", type=int, default=100,
                       help="Number of examples per domain")
    parser.add_argument("--output", type=str, default="data/templates/enhanced_templates.jsonl",
                       help="Output file path")

    args = parser.parse_args()

    output_path = Path(args.output)
    output_path.parent.mkdir(parents=True, exist_ok=True)

    # Generate
    examples = generate_all_templates(num_per_domain=args.num_per_domain)

    # Save
    save_examples(examples, output_path)

    # Stats
    print(f"\n=== Statistics ===")
    print(f"Total examples: {len(examples)}")
    print(f"\nBy domain:")
    domain_counts = {}
    for ex in examples:
        domain_counts[ex.domain] = domain_counts.get(ex.domain, 0) + 1
    for domain, count in sorted(domain_counts.items()):
        print(f"  {domain}: {count}")


if __name__ == "__main__":
    main()
