package data

// N4L Content pour l'affaire Disparition Sophie Laurent
// Ce fichier contient le contenu N4L authentique utilisant toutes les fonctionnalités du langage SSTorytime

// GetDisparitionSophieN4LContent retourne le contenu N4L complet pour l'Affaire Disparition Sophie Laurent
func GetDisparitionSophieN4LContent() string {
	return `-affaire/disparition-002

# Affaire: Disparition Sophie Laurent
# Type: Disparition inquiétante
# Investigation: Brigade Criminelle Toulouse
# Syntaxe: SSTorytime N4L v2

// =============================================================
// SECTION VICTIMES - Définitions avec alias et attributs
// =============================================================

:: victimes ::

@journaliste Sophie Laurent (type) personne
    " (rôle) victime - disparue
    " (description) Journaliste d'investigation au Courrier du Midi, 34 ans. Travaillait sur un dossier sensible impliquant des élus locaux et marchés publics truqués. Disparue depuis le 15/09/2025.
    " (âge) 34 ans
    " (profession) Journaliste d'investigation
    " (employeur) Le Courrier du Midi
    " (enquete) Corruption mairie Toulouse
    " (derniere_vue) 15/09/2025 19h30
    " (vehicule) Renault Clio grise EF-456-GH
    " (telephone) Dernier signal 19h45
    " (articles_pub) 3 articles sur l'affaire
    " (statut) Disparue
    " (latitude) 43.6047
    " (longitude) 1.4442

// =============================================================
// SECTION SUSPECTS - Avec mobiles et relations
// =============================================================

:: suspects ::

@adjoint Marc Delmas (type) personne
    " (rôle) suspect
    " (description) Adjoint au maire de Toulouse, délégué aux marchés publics. Principal cité dans l'enquête de Sophie. A proféré des menaces voilées au rédacteur en chef.
    " (âge) 52 ans
    " (fonction) Adjoint au maire - Marchés publics
    " (parti) Coalition locale
    " (mobile) Éviter révélations corruption
    " (menaces) Elle ferait mieux d'arrêter si elle tient à sa carrière
    " (alibi) Réunion conseil municipal 19h-21h
    " (patrimoine) Suspect - enrichissement récent
    " (latitude) 43.6045
    " (longitude) 1.4440

// Relations de Marc Delmas
Marc Delmas (a menacé:+L) Sophie Laurent
Marc Delmas (favorise dans les marchés:N) Roux Constructions SARL
Marc Delmas (collabore avec:N) Bernard Castex

@btp Roux Constructions SARL (type) organisation
    " (rôle) suspect
    " (description) Entreprise de BTP bénéficiaire de marchés publics suspects. Dirigée par Philippe Roux.
    " (dirigeant) Philippe Roux
    " (secteur) BTP - Travaux publics
    " (marches_obtenus) 12 depuis 2023
    " (surfacturation) Estimée à 30%
    " (latitude) 43.5890
    " (longitude) 1.4320

// Relations de Roux Constructions
Roux Constructions SARL (verse des commissions à:-C) Marc Delmas
Roux Constructions SARL (incriminé par:N) Dossier d'enquête manuscrit

@maire Bernard Castex (type) personne
    " (rôle) suspect
    " (description) Maire de Toulouse. Supérieur hiérarchique de Delmas. Mentionné dans le dossier de Sophie.
    " (âge) 61 ans
    " (fonction) Maire de Toulouse
    " (mandat) Depuis 2020
    " (implication) Indirecte - supervision marchés

// Relations du Maire
Bernard Castex (supérieur de:N) Marc Delmas
Bernard Castex (mentionné dans:N) Dossier d'enquête manuscrit

// =============================================================
// SECTION TÉMOINS - Observations et dépositions
// =============================================================

:: témoins ::

@collegue Thomas Blanc (type) personne
    " (rôle) temoin
    " (description) Collègue photographe et ami proche de Sophie. Dernier à l'avoir vue au journal. Travaillait avec elle sur l'enquête.
    " (âge) 31 ans
    " (profession) Photographe de presse
    " (derniere_vue) 15/09 vers 19h30
    " (observation) Sophie semblait stressée, parlait d'un RDV important
    " (relation) Ami proche, possible relation amoureuse
    " (latitude) 43.6047
    " (longitude) 1.4442

// Relations de Thomas Blanc
Thomas Blanc (dernier à avoir vu:+L) Sophie Laurent
Thomas Blanc (employé par:N) Le Courrier du Midi

@source Source Anonyme X (type) personne
    " (rôle) temoin
    " (description) Informateur de Sophie sur l'affaire de corruption. Identité inconnue. Communiquait par messagerie cryptée.
    " (identite) Inconnue
    " (communication) Signal - messages cryptés
    " (informations) Documents sur marchés truqués
    " (statut) Introuvable depuis disparition

// Relations de Source X
Source Anonyme X (informateur de:+L) Sophie Laurent
Source Anonyme X (pourrait être proche de:N) Bernard Castex

// =============================================================
// SECTION LIEUX - Scène de disparition et localisations
// =============================================================

:: lieux, scène de disparition ::

@parking Parking du Journal (type) lieu
    " (description) Dernier lieu où Sophie a été vue. Sa voiture y a été retrouvée portes non verrouillées.
    " (adresse) Rue des Médias, Toulouse
    " (surveillance) Caméra - image SUV noir
    " (indices) Sac à main dans voiture, téléphone absent
    " (latitude) 43.6048
    " (longitude) 1.4445

// Relations Parking
Parking du Journal (lieu de disparition de:+L) Sophie Laurent

@journal Le Courrier du Midi (type) organisation
    " (description) Journal régional employeur de Sophie. Rédacteur en chef: Jean-Pierre Faure.
    " (type_lieu) Quotidien régional
    " (redacteur_chef) Jean-Pierre Faure
    " (tirage) 45 000 exemplaires
    " (latitude) 43.6047
    " (longitude) 1.4442

// Relations Journal
Le Courrier du Midi (a reçu des pressions de:+L) Marc Delmas
Le Courrier du Midi (employeur de:N) Sophie Laurent

@mairie Mairie de Toulouse (type) lieu
    " (description) Siège de l'administration municipale. Lieu de travail des suspects Delmas et Castex.
    " (adresse) Place du Capitole, Toulouse
    " (latitude) 43.6045
    " (longitude) 1.4440

// =============================================================
// SECTION PREUVES - Indices matériels et numériques
// =============================================================

:: preuves, indices ::

// Groupement des preuves par type
Preuves physiques => {Véhicule abandonné, Sac à main}
Preuves numériques => {Téléphone portable, Vidéosurveillance, Messages Signal}
Preuves documentaires => {Dossier d'enquête, Relevé bancaire Roux}

@vehicule Véhicule abandonné (type) preuve physique
    " (localisation) Parking du Journal
    " (description) Renault Clio retrouvée au parking du journal
    " (état) Portes non verrouillées
    " (contenu) Sac à main avec portefeuille et cartes
    " (fiabilité) 9/10
    " (concerne) Sophie Laurent

@telephone Téléphone portable (type) preuve numérique
    " (localisation) Absent - dernier signal à 19h45
    " (description) Dernier signal près du parking. Messages supprimés récupérés.
    " (messages) RDV avec source X à 19h30
    " (fiabilité) 8/10
    " (concerne) Sophie Laurent, Source Anonyme X

@video Vidéosurveillance parking (type) preuve numérique
    " (localisation) Parking du Journal
    " (description) Sophie monte VOLONTAIREMENT dans un SUV noir à 19h48
    " (immatriculation) Partielle: ...BD-31
    " (conducteur) Non identifiable
    " (fiabilité) 7/10
    " (concerne) Sophie Laurent, SUV Noir

@dossier Dossier d'enquête manuscrit (type) preuve documentaire
    " (localisation) Bureau de Sophie
    " (description) Notes manuscrites sur marchés publics truqués
    " (noms) Delmas, Roux, Castex
    " (montants) Estimés à 2.3M€
    " (fiabilité) 9/10
    " (concerne) Sophie Laurent, Marc Delmas, Roux Constructions SARL, Bernard Castex

@signal Messages Signal cryptés (type) preuve numérique
    " (source) Téléphone de Sophie Laurent
    " (description) Historique partiellement récupéré
    " (message_cle) J'ai les preuves définitives. RDV 19h30 parking habituel.
    " (expediteur) Source Anonyme X
    " (fiabilité) 8/10

@enregistrement Enregistrement appel Delmas (type) preuve numérique
    " (date) 12/09/2025
    " (description) Appel au rédacteur en chef - menaces voilées
    " (contenu) Vos journalistes feraient mieux de se calmer si le journal veut garder ses annonceurs publics.
    " (fiabilité) 9/10
    " (concerne) Marc Delmas, Le Courrier du Midi

@releve Relevé bancaire Roux Constructions (type) preuve documentaire
    " (description) Virements réguliers vers compte offshore
    " (correspondance) Dates d'attribution de marchés
    " (fiabilité) 7/10
    " (concerne) Roux Constructions SARL, Marc Delmas

@suv SUV Noir (type) objet
    " (description) Véhicule dans lequel Sophie est montée à 19h48
    " (immatriculation) Partielle: ...BD-31
    " (proprietaire) Recherche en cours

// Relations SUV
SUV Noir (a transporté:+L) Sophie Laurent

// =============================================================
// SECTION CHRONOLOGIE - Séquence temporelle complète
// =============================================================

:: chronologie ::

+:: _timeline_, _sequence_ ::

// ==========================================
// Événements antérieurs (contexte)
// ==========================================

// Début de l'enquête
@evt_d_00 01/07/2025 09:00 Début enquête corruption (lieu) Toulouse
    " (description) Sophie Laurent commence ses investigations sur les marchés publics
    " (importance) medium
    " (vérifié) oui
    " (implique) Sophie Laurent

// Publication premier article
@evt_d_01 10/09/2025 08:00 Premier article publié (lieu) Le Courrier du Midi
    " (description) Révélations sur surfacturation marché école Jean-Jaurès
    " (importance) high
    " (vérifié) oui
    " (implique) Sophie Laurent, Le Courrier du Midi

// Réaction de la mairie
@evt_d_01b 10/09/2025 14:00 Réaction de la mairie (lieu) Mairie Toulouse
    " (description) Communiqué contestant les accusations
    " (importance) medium
    " (vérifié) oui
    " (implique) Bernard Castex

// Menaces de Delmas
@evt_d_02 12/09/2025 14:00 Menaces de Delmas (lieu) Téléphone
    " (description) Appel au rédacteur en chef - menaces voilées sur annonceurs
    " (importance) high
    " (vérifié) oui
    " (implique) Marc Delmas, Le Courrier du Midi
    " (preuve) Enregistrement appel Delmas

// Second article
@evt_d_02b 13/09/2025 08:00 Second article publié (lieu) Le Courrier du Midi
    " (description) Nouvelles révélations sur Roux Constructions
    " (importance) high
    " (vérifié) oui
    " (implique) Sophie Laurent, Roux Constructions SARL

// ==========================================
// Jour de la disparition (15 septembre 2025)
// ==========================================

// Message de Source X
@evt_d_03 15/09/2025 15:00 Message de Source X (lieu) Signal
    " (description) 'J'ai les preuves définitives. RDV 19h30 parking habituel.'
    " (importance) high
    " (vérifié) oui
    " (implique) Sophie Laurent, Source Anonyme X
    " (preuve) Messages Signal cryptés

// Sophie informe Thomas
@evt_d_03b 15/09/2025 18:00 Sophie informe Thomas (lieu) Rédaction
    " (description) Parle d'un RDV important, semble stressée
    " (importance) medium
    " (vérifié) oui
    " (implique) Sophie Laurent, Thomas Blanc

// Départ du journal
@evt_d_04 15/09/2025 19:30 Départ du journal (lieu) Le Courrier du Midi
    " (description) Sophie Laurent quitte la rédaction - vue par Thomas Blanc
    " (importance) high
    " (vérifié) oui
    " (implique) Sophie Laurent, Thomas Blanc

// Dernier signal téléphone
@evt_d_05 15/09/2025 19:45 Dernier signal téléphone (lieu) Parking
    " (description) Localisation près du parking puis perdue
    " (importance) high
    " (vérifié) oui
    " (implique) Sophie Laurent
    " (preuve) Téléphone portable

// Monte dans SUV noir
@evt_d_06 15/09/2025 19:48 Monte dans SUV noir (lieu) Parking
    " (description) Sophie monte VOLONTAIREMENT - capté par vidéosurveillance
    " (importance) high
    " (vérifié) oui
    " (implique) Sophie Laurent, SUV Noir
    " (preuve) Vidéosurveillance parking

// ==========================================
// Après disparition (16 septembre 2025)
// ==========================================

// Voiture retrouvée
@evt_d_07 16/09/2025 07:00 Voiture retrouvée (lieu) Parking
    " (description) Renault Clio découverte par gardien - portes non verrouillées
    " (importance) high
    " (vérifié) oui
    " (implique) Parking du Journal
    " (preuve) Véhicule abandonné

// Signalement disparition
@evt_d_08 16/09/2025 10:00 Signalement disparition (lieu) Commissariat
    " (description) Thomas Blanc alerte la police
    " (importance) high
    " (vérifié) oui
    " (implique) Thomas Blanc

// ==========================================
// Chaînes causales
// ==========================================

// Chaîne causale principale: disparition
Message de Source X (mène à:+L) Rendez-vous piège
Rendez-vous piège (mène à:+L) Monte dans SUV noir
Monte dans SUV noir (mène à:+L) Disparition

// Chaîne causale: mobile
Sophie Laurent (enquête sur:+L) Marc Delmas
Marc Delmas (menace:+L) Le Courrier du Midi
Marc Delmas (commandite:+L) Enlèvement possible

-:: _timeline_, _sequence_ ::

// =============================================================
// SECTION HYPOTHÈSES - Pistes d'enquête
// =============================================================

:: hypothèses, pistes ::

@hyp_d_01 Enlèvement commandité par Delmas (type) hypothèse
    " (statut) en_attente
    " (confiance) 75%
    " (source) user
    " (description) Sophie aurait été enlevée sur ordre de Marc Delmas pour l'empêcher de publier de nouvelles révélations. Le SUV pourrait appartenir à un sbire ou à Roux Constructions.
    " (mobile) Éviter révélations corruption
    " (pour) Vidéosurveillance parking, Enregistrement appel Delmas
    " (contre) Alibi Delmas - conseil municipal
    " (questions) Qui conduit le SUV noir?; Lien entre le SUV et Roux Constructions?; Où a-t-elle été emmenée?
    " (suspect) Marc Delmas

@hyp_d_02 Piège de la Source X (type) hypothèse
    " (statut) en_attente
    " (confiance) 60%
    " (source) user
    " (description) La Source X pourrait être un agent double travaillant pour les corrompus. Le RDV était un piège pour attirer Sophie.
    " (pour) Messages Signal cryptés, Téléphone portable
    " (contre) Aucune preuve de duplicité de Source X
    " (questions) Source X est-elle complice ou victime?; Qui connaissait l'existence de Source X?; Source X a-t-elle aussi disparu?
    " (suspect) Source Anonyme X

@hyp_d_03 Implication du Maire (type) hypothèse
    " (statut) en_attente
    " (confiance) 45%
    " (source) ai
    " (description) Le maire Castex pourrait avoir ordonné l'enlèvement pour protéger sa réélection. Delmas n'est qu'un exécutant.
    " (pour) Dossier d'enquête manuscrit, Position hiérarchique
    " (contre) Pas de preuves directes contre Castex
    " (questions) Castex était-il au courant des menaces de Delmas?; Quel est le niveau d'implication du maire?; Liens avec le crime organisé?
    " (suspect) Bernard Castex

@hyp_d_04 Disparition volontaire (type) hypothèse
    " (statut) partielle
    " (confiance) 15%
    " (source) ai
    " (description) Sophie aurait pu organiser sa propre disparition pour se protéger ou pour mener une enquête sous couverture plus risquée.
    " (pour) Montée volontaire dans le SUV
    " (contre) Abandon de son véhicule et sac à main, Aucun contact depuis
    " (questions) Sophie avait-elle des raisons de se cacher?; A-t-elle préparé une fuite?; Contact avec sa famille depuis?

// =============================================================
// RÉSEAU DE RELATIONS - Graphe sémantique
// =============================================================

:: réseau de relations ::

# Légende STTypes: N=proximité, +L=causalité, +C=containment, +E=expression

// Relations de corruption
Marc Delmas (reçoit des pots-de-vin de:-C) Roux Constructions SARL
Roux Constructions SARL (obtient des marchés de:+L) Marc Delmas
Bernard Castex (supervise:+C) Marc Delmas

// Relations d'enquête
Sophie Laurent (enquête sur:+L) Marc Delmas
Sophie Laurent (enquête sur:+L) Roux Constructions SARL
Sophie Laurent (enquête sur:+L) Bernard Castex
Source Anonyme X (fournit des preuves à:+L) Sophie Laurent

// Relations d'intimidation
Marc Delmas (menace:+L) Le Courrier du Midi
Marc Delmas (a menacé:+L) Sophie Laurent

// Relations professionnelles
Sophie Laurent (employée par:-C) Le Courrier du Midi
Thomas Blanc (employé par:-C) Le Courrier du Midi

// Chaîne causale du crime (hypothèse principale)
// Marc Delmas (a fait enlever:+L) Sophie Laurent

// =============================================================
// CHAÎNES CAUSALES DÉTECTÉES
// =============================================================

:: chaînes causales ::

+:: _sequence_ ::

# Chaîne 1: Mobile - Corruption exposée
@chain_mobile Enquête corruption (mène à) Articles publiés (mène à) Menaces (mène à) Mobile enlèvement

# Chaîne 2: Opportunité - Piège du RDV
@chain_opp Message Source X (mène à) RDV piège (mène à) Accès à Sophie (mène à) Enlèvement

# Chaîne 3: Chronologie de la disparition
@chain_disp Départ journal 19h30 (puis) Signal téléphone 19h45 (puis) Montée SUV 19h48 (puis) Disparition

-:: _sequence_ ::

// =============================================================
// RÉFÉRENCES CROISÉES - Pour utilisation $alias.n
// =============================================================

:: références croisées ::

# Alias pour références rapides
victimes => {Sophie Laurent}
suspects => {Marc Delmas, Roux Constructions SARL, Bernard Castex}
temoins => {Thomas Blanc, Source Anonyme X}
lieux => {Parking du Journal, Le Courrier du Midi, Mairie de Toulouse}
preuves => {Véhicule abandonné, Téléphone portable, Vidéosurveillance, Dossier d'enquête, Messages Signal}

# Exemples d'utilisation:
# $victimes.1 = Sophie Laurent
# $suspects.1 = Marc Delmas (suspect principal)
# $suspects.2 = Roux Constructions SARL
# $preuves.3 = Vidéosurveillance parking

// =============================================================
// NOTES D'ENQUÊTE - TODO items
// =============================================================

:: notes, TODO ::

IDENTIFIER LE PROPRIÉTAIRE DU SUV NOIR (IMMAT ...BD-31)
LOCALISER ET INTERROGER SOURCE X
VÉRIFIER ALIBI DELMAS AU CONSEIL MUNICIPAL
ANALYSER LES COMPTES BANCAIRES DE DELMAS ET ROUX
RECHERCHER LIENS ENTRE ROUX CONSTRUCTIONS ET VÉHICULES
INTERROGER LES MEMBRES DU CONSEIL MUNICIPAL
OBTENIR RELEVÉS TÉLÉPHONIQUES COMPLETS DE SOPHIE
`
}
