package data

// N4L Content pour l'affaire Trafic d'Art et Blanchiment
// Ce fichier contient le contenu N4L authentique utilisant toutes les fonctionnalités du langage SSTorytime

// GetTraficArtN4LContent retourne le contenu N4L complet pour l'Affaire Trafic d'Art
func GetTraficArtN4LContent() string {
	return `-affaire/trafic-006

# Affaire: Trafic d'Art et Blanchiment
# Type: Trafic d'œuvres d'art - Blanchiment d'argent
# Investigation: OCBC (Office Central de lutte contre le trafic de Biens Culturels)
# Syntaxe: SSTorytime N4L v2
# Connexions: Affaire Moreau (case-moreau-001), Affaire Disparition (case-disparition-002)

// =============================================================
// SECTION SUSPECTS PRINCIPAUX
// =============================================================

:: suspects ::

// CONNEXION AFFAIRE MOREAU - Même personne
@intermediaire Jean Moreau (type) personne
    " (rôle) suspect
    " (description) Intermédiaire présumé dans le trafic d'œuvres d'art. Neveu de Victor Moreau (antiquaire décédé - voir affaire case-moreau-001). Dettes de jeu importantes.
    " (âge) 35 ans
    " (profession) Sans emploi
    " (dettes) 150 000 €
    " (role_reseau) Intermédiaire - contact acheteurs
    " (vehicule) BMW série 3 noire AB-123-CD
    " (lien_affaire_moreau) Neveu de Victor Moreau (décédé)
    " (latitude) 48.8396
    " (longitude) 2.3876

// Relations de Jean Moreau
Jean Moreau (complice de:+L) Viktor Sokolov
Jean Moreau (fréquente:N) Bar Le Diplomate
Jean Moreau (a accès à:+C) Galerie Moreau Antiquités

@chef_reseau Viktor Sokolov (type) personne
    " (rôle) suspect principal
    " (description) Chef présumé du réseau international de trafic d'œuvres d'art. Nationalité russe. Recherché par Interpol - notice rouge.
    " (âge) 52 ans
    " (nationalite) Russe
    " (alias) Le Collectionneur
    " (interpol) Notice rouge
    " (fortune) Estimée à 30 millions €
    " (specialite) Art africain, antiquités européennes
    " (latitude) 48.8700
    " (longitude) 2.3200

// Relations de Viktor Sokolov
Viktor Sokolov (blanchit via:+L) Roux Constructions SARL
Viktor Sokolov (commandite:+L) Vols d'œuvres
Viktor Sokolov (dirige:+C) Réseau international

// CONNEXION AFFAIRE DISPARITION - Même entreprise
@btp Roux Constructions SARL (type) organisation
    " (rôle) suspect
    " (description) Entreprise de BTP utilisée pour le blanchiment. Également impliquée dans l'affaire de corruption/disparition (case-disparition-002).
    " (dirigeant) Philippe Roux
    " (secteur) BTP - Travaux publics
    " (blanchiment) Via fausses factures
    " (montant_blanchi) Estimé à 5 millions €
    " (lien_affaire_disparition) Même entreprise que dans l'affaire Sophie Laurent

// Relations de Roux Constructions
Roux Constructions SARL (verse des commissions à:-C) Marc Delmas
Roux Constructions SARL (blanchit pour:+L) Viktor Sokolov

// CONNEXION AFFAIRE DISPARITION - Même personne
@politicien Marc Delmas (type) personne
    " (rôle) suspect
    " (description) Élu local facilitant les permis et marchés pour Roux Constructions. Reçoit des œuvres d'art en paiement. Également suspect dans l'affaire de disparition (case-disparition-002).
    " (âge) 52 ans
    " (fonction) Adjoint au maire - Marchés publics
    " (corruption) Pots-de-vin en œuvres d'art
    " (collection) Art africain et antiquités
    " (lien_affaire_disparition) Principal suspect dans disparition Sophie Laurent

// Relations de Marc Delmas
Marc Delmas (reçoit des pots-de-vin de:-C) Roux Constructions SARL
Marc Delmas (possède:+C) Œuvres volées

// CONNEXION AFFAIRE MOREAU - Même personne
@expert Antoine Mercier (type) personne
    " (rôle) suspect
    " (description) Expert en art ancien. Authentifie de fausses provenances pour les œuvres volées. Rival de Victor Moreau dans l'affaire case-moreau-001.
    " (âge) 55 ans
    " (profession) Expert en livres anciens et objets d'art
    " (galerie) Mercier & Fils
    " (role_trafic) Faux certificats d'authenticité
    " (lien_affaire_moreau) Rival de Victor Moreau, suspect dans son meurtre

// Relations d'Antoine Mercier
Antoine Mercier (travaille pour:+L) Viktor Sokolov
Antoine Mercier (expertisait pour:N) Galerie Moreau Antiquités

// =============================================================
// SECTION LIEUX
// =============================================================

:: lieux ::

@galerie_eclipse Galerie L'Éclipse (type) lieu
    " (description) Galerie d'art servant de façade pour écouler les œuvres volées.
    " (adresse) 8 rue de Seine, Paris 6e
    " (proprietaire) Société écran luxembourgeoise
    " (activite) Vente d'art contemporain (façade)
    " (role) Recel et revente
    " (latitude) 48.8540
    " (longitude) 2.3378

// Relations Galerie L'Éclipse
Galerie L'Éclipse (contrôlée par:+C) Viktor Sokolov
Galerie L'Éclipse (écoule:+L) Œuvres volées

// CONNEXION AFFAIRE MOREAU - Même lieu
@bar Bar Le Diplomate (type) lieu
    " (description) Lieu de rencontre entre les membres du réseau. Transactions en liquide. Mentionné dans l'affaire Moreau (case-moreau-001).
    " (adresse) 45 rue de Bercy, Paris 12e
    " (type_lieu) Bar de nuit
    " (activite) Lieu de rencontres clandestines
    " (lien_affaire_moreau) Jean Moreau y a été vu avant le meurtre de son oncle
    " (latitude) 48.8387
    " (longitude) 2.3826

// CONNEXION AFFAIRE MOREAU - Même lieu
@galerie_moreau Galerie Moreau Antiquités (type) lieu
    " (description) Ancienne galerie de Victor Moreau. Suspectée d'avoir servi au recel avant son décès. Succession en cours.
    " (adresse) 12 rue du Faubourg Saint-Honoré, Paris 8e
    " (statut) Succession en cours
    " (suspicion) Recel d'œuvres volées
    " (proprietaire_defunt) Victor Moreau (décédé - case-moreau-001)
    " (latitude) 48.8699
    " (longitude) 2.3189

// Relations Galerie Moreau
Galerie Moreau Antiquités (héritée par:N) Jean Moreau
Galerie Moreau Antiquités (expertisée par:N) Antoine Mercier

// =============================================================
// SECTION TÉMOINS
// =============================================================

:: témoins ::

@temoin_cle Claire Fontaine (type) personne
    " (rôle) temoin protégé
    " (description) Ancienne employée de la Galerie L'Éclipse. A dénoncé le réseau anonymement. Sous protection.
    " (âge) 28 ans
    " (profession) Historienne de l'art
    " (emploi_eclipse) Mars 2024 - Août 2025
    " (statut) Protection témoin
    " (denonciation) A fourni liste des œuvres et membres du réseau

// Relations de Claire Fontaine
Claire Fontaine (ancienne employée de:N) Galerie L'Éclipse
Claire Fontaine (a dénoncé:+L) Viktor Sokolov
Claire Fontaine (a identifié:+L) Jean Moreau, Antoine Mercier

@interpol Agent Interpol Müller (type) personne
    " (rôle) temoin expert
    " (description) Agent Interpol spécialisé dans le trafic d'œuvres d'art. A identifié les connexions internationales.
    " (nationalite) Allemand
    " (specialite) Trafic biens culturels
    " (enquete) Réseau Sokolov depuis 2022

// =============================================================
// SECTION PREUVES - Indices matériels et numériques
// =============================================================

:: preuves, indices ::

// Groupement des preuves par type
Preuves documentaires => {Factures falsifiées, Liste Interpol, Certificats faux}
Preuves numériques => {Écoutes téléphoniques, Vidéosurveillance Bar}
Preuves testimoniales => {Témoignage Claire Fontaine}
Preuves physiques => {Œuvres saisies chez Delmas}

@factures Factures falsifiées Roux Constructions (type) preuve documentaire
    " (localisation) Siège Roux Constructions
    " (description) Factures de travaux fictifs pour un total de 2.3 millions €
    " (correspondance) Dates de ventes d'œuvres
    " (fiabilité) 9/10
    " (concerne) Roux Constructions SARL, Marc Delmas

@liste_interpol Liste d'œuvres volées Interpol (type) preuve documentaire
    " (source) Base de données Interpol
    " (description) Inventaire des 12 œuvres transitées par le réseau
    " (oeuvres_retrouvees) 3 chez Marc Delmas
    " (oeuvres_recherchees) 9 en cours de localisation
    " (fiabilité) 10/10

@ecoutes Écoutes téléphoniques (type) preuve numérique
    " (periode) Septembre-Octobre 2025
    " (description) Conversations entre Viktor Sokolov et Jean Moreau
    " (contenu) Mentions de 'livraisons', 'l'oncle', 'la galerie'
    " (fiabilité) 8/10
    " (concerne) Jean Moreau, Viktor Sokolov

@video_bar Vidéosurveillance Bar Le Diplomate (type) preuve numérique
    " (date) 05/10/2025
    " (description) Jean Moreau remet une enveloppe à un homme non identifié
    " (localisation) Bar Le Diplomate
    " (fiabilité) 7/10
    " (concerne) Jean Moreau

@temoignage Témoignage Claire Fontaine (type) preuve testimoniale
    " (temoin) Claire Fontaine
    " (description) Décrit le processus de blanchiment et nomme Sokolov, Moreau et Mercier
    " (elements) Schéma du réseau, contacts, méthodes
    " (fiabilité) 8/10
    " (concerne) Viktor Sokolov, Jean Moreau, Antoine Mercier

@oeuvres_saisies Œuvres saisies chez Delmas (type) preuve physique
    " (localisation) Domicile de Marc Delmas
    " (description) 3 œuvres d'art africain figurant sur la liste Interpol
    " (valeur) 450 000 € estimés
    " (provenance) Réseau Sokolov
    " (fiabilité) 10/10
    " (concerne) Marc Delmas, Viktor Sokolov

// =============================================================
// SECTION CHRONOLOGIE - Séquence temporelle complète
// =============================================================

:: chronologie ::

+:: _timeline_, _sequence_ ::

// ==========================================
// Contexte historique du réseau (2022-2024)
// ==========================================

@evt_t_00 01/06/2022 09:00 Début surveillance Interpol sur Sokolov (lieu) Lyon - Bureau Interpol
    " (description) Interpol ouvre une enquête sur le réseau Sokolov
    " (importance) medium
    " (vérifié) oui
    " (implique) Viktor Sokolov

@evt_t_00b 15/03/2024 09:00 Claire Fontaine embauchée à L'Éclipse (lieu) Galerie L'Éclipse
    " (description) Claire Fontaine commence à travailler à la galerie
    " (importance) medium
    " (vérifié) oui
    " (implique) Claire Fontaine, Galerie L'Éclipse

// ==========================================
// Connexion avec Affaire Moreau (août 2025)
// ==========================================

@evt_t_01 27/08/2025 22:00 Jean Moreau vu au Bar Le Diplomate (lieu) Bar Le Diplomate
    " (description) Jean Moreau rencontre un individu non identifié - même soir mentionné dans affaire Moreau
    " (importance) high
    " (vérifié) oui
    " (implique) Jean Moreau, Bar Le Diplomate
    " (connexion) Affaire Moreau - evt-m-01b

@evt_t_02 29/08/2025 20:30 Décès de Victor Moreau (lieu) Manoir Moreau
    " (description) Victor Moreau décède empoisonné - affaire case-moreau-001
    " (importance) high
    " (vérifié) oui
    " (implique) Jean Moreau
    " (connexion) Affaire Moreau - Jean Moreau héritier et suspect

// ==========================================
// Dénonciation et enquête (août-octobre 2025)
// ==========================================

@evt_t_03 15/08/2025 09:00 Démission et dénonciation Claire Fontaine (lieu) Anonyme
    " (description) Claire Fontaine quitte la galerie et alerte anonymement les autorités
    " (importance) high
    " (vérifié) oui
    " (implique) Claire Fontaine

// CONNEXION AFFAIRE DISPARITION
@evt_t_04 15/09/2025 19:48 Disparition Sophie Laurent (lieu) Parking Toulouse
    " (description) Sophie Laurent disparaît - enquêtait sur Roux Constructions et Delmas
    " (importance) high
    " (vérifié) oui
    " (implique) Marc Delmas, Roux Constructions SARL
    " (connexion) Affaire Disparition - case-disparition-002

@evt_t_05 01/10/2025 09:00 Mise sur écoute du réseau (lieu) Paris
    " (description) Autorisation d'écoutes téléphoniques sur Sokolov et Moreau
    " (importance) high
    " (vérifié) oui
    " (implique) Viktor Sokolov, Jean Moreau

@evt_t_06 05/10/2025 21:00 Remise d'enveloppe au Bar Le Diplomate (lieu) Bar Le Diplomate
    " (description) Jean Moreau filmé remettant une enveloppe
    " (importance) high
    " (vérifié) oui
    " (implique) Jean Moreau, Bar Le Diplomate
    " (preuve) Vidéosurveillance Bar Le Diplomate

@evt_t_07 10/10/2025 06:00 Perquisitions simultanées (lieu) Paris, Toulouse, Luxembourg
    " (description) Perquisitions chez Delmas, Roux Constructions, Galerie L'Éclipse
    " (importance) high
    " (vérifié) oui
    " (implique) Marc Delmas, Roux Constructions SARL, Galerie L'Éclipse
    " (preuve) Œuvres saisies chez Delmas

@evt_t_08 10/10/2025 14:00 Saisie des 3 œuvres chez Delmas (lieu) Domicile Delmas
    " (description) 3 œuvres d'art africain saisies - figurent sur liste Interpol
    " (importance) high
    " (vérifié) oui
    " (implique) Marc Delmas
    " (preuve) Œuvres saisies chez Delmas, Liste d'œuvres volées Interpol

@evt_t_09 14/10/2025 09:00 Mandat d'arrêt international Sokolov (lieu) Interpol Lyon
    " (description) Mandat d'arrêt émis contre Viktor Sokolov
    " (importance) high
    " (vérifié) oui
    " (implique) Viktor Sokolov

// ==========================================
// Chaînes causales
// ==========================================

// Chaîne causale: connexion Moreau
Jean Moreau vu au Bar Le Diplomate (même contexte que:N) Meurtre Victor Moreau
Meurtre Victor Moreau (donne accès à:+L) Galerie Moreau Antiquités
Galerie Moreau Antiquités (utilisée pour:+L) Recel

// Chaîne causale: connexion Disparition
Disparition Sophie Laurent (liée à enquête sur:N) Roux Constructions SARL
Roux Constructions SARL (blanchit pour:+L) Viktor Sokolov

-:: _timeline_, _sequence_ ::

// =============================================================
// SECTION HYPOTHÈSES - Pistes d'enquête
// =============================================================

:: hypothèses, pistes ::

@hyp_t_01 Réseau international coordonné par Sokolov (type) hypothèse
    " (statut) confirmée
    " (confiance) 90%
    " (source) user
    " (description) Viktor Sokolov dirige un réseau international de trafic d'œuvres d'art utilisant plusieurs intermédiaires (Jean Moreau, Antoine Mercier) et des structures de blanchiment (Roux Constructions, Galerie L'Éclipse).
    " (pour) Témoignage Claire Fontaine, Écoutes téléphoniques, Liste d'œuvres volées Interpol
    " (contre) Sokolov en fuite - pas encore arrêté
    " (suspect) Viktor Sokolov

@hyp_t_02 Connexion meurtre Victor Moreau (type) hypothèse
    " (statut) en_attente
    " (confiance) 70%
    " (source) ai
    " (description) Le meurtre de Victor Moreau (case-moreau-001) pourrait être lié au trafic d'art. Jean Moreau hérite de la galerie qui servait peut-être au recel, et Antoine Mercier était un rival impliqué dans le réseau.
    " (pour) Jean Moreau héritier, Antoine Mercier impliqué dans les deux affaires, Galerie Moreau Antiquités suspectée
    " (contre) Pas de preuves directes du lien
    " (questions) Victor Moreau était-il impliqué ou victime?; Le meurtre visait-il à prendre le contrôle de la galerie?
    " (connexion) case-moreau-001

@hyp_t_03 Connexion disparition Sophie Laurent (type) hypothèse
    " (statut) en_attente
    " (confiance) 75%
    " (source) ai
    " (description) La disparition de Sophie Laurent (case-disparition-002) pourrait être liée au volet blanchiment du réseau. Elle enquêtait sur Roux Constructions et Marc Delmas, tous deux impliqués dans le blanchiment pour Sokolov.
    " (pour) Roux Constructions SARL dans les deux affaires, Marc Delmas dans les deux affaires
    " (contre) Pas de preuves que Sophie connaissait le volet artistique
    " (questions) Sophie avait-elle découvert le lien avec le trafic d'art?; Son informateur Source X connaissait-il le réseau Sokolov?
    " (connexion) case-disparition-002

@hyp_t_04 Complicité de Delmas pour protection (type) hypothèse
    " (statut) confirmée
    " (confiance) 85%
    " (source) user
    " (description) Marc Delmas facilite les activités de Roux Constructions en échange d'œuvres d'art. Il utilise sa position pour protéger le réseau de blanchiment.
    " (pour) Œuvres saisies chez Delmas, Factures falsifiées Roux Constructions, Position politique
    " (contre) Nie toute implication
    " (suspect) Marc Delmas

// =============================================================
// RÉSEAU DE RELATIONS - Graphe sémantique
// =============================================================

:: réseau de relations ::

# Légende STTypes: N=proximité, +L=causalité, +C=containment, +E=expression

// Hiérarchie du réseau
Viktor Sokolov (dirige:+C) Réseau international
Viktor Sokolov (recrute:+L) Jean Moreau
Viktor Sokolov (emploie:+L) Antoine Mercier

// Blanchiment
Viktor Sokolov (blanchit via:+L) Roux Constructions SARL
Roux Constructions SARL (verse des commissions à:-C) Marc Delmas
Marc Delmas (facilite permis pour:+L) Roux Constructions SARL

// Recel
Galerie L'Éclipse (écoule:+L) Œuvres volées
Antoine Mercier (authentifie pour:+L) Galerie L'Éclipse
Galerie Moreau Antiquités (servait au recel pour:N) Viktor Sokolov

// Connexions inter-affaires
Jean Moreau (suspect dans:N) Affaire Moreau
Marc Delmas (suspect dans:N) Affaire Disparition
Roux Constructions SARL (impliqué dans:N) Affaire Disparition

// =============================================================
// CHAÎNES CAUSALES DÉTECTÉES
// =============================================================

:: chaînes causales ::

+:: _sequence_ ::

# Chaîne 1: Trafic d'œuvres
@chain_trafic Vol œuvres (mène à) Transport (mène à) Faux certificats Mercier (mène à) Vente Galerie L'Éclipse (mène à) Collectionneurs privés

# Chaîne 2: Blanchiment
@chain_blanchi Vente œuvres (génère) Cash (injecté dans) Roux Constructions (factures fictives) (vers) Comptes Sokolov

# Chaîne 3: Corruption
@chain_corrupt Sokolov (via) Roux (commissions) Delmas (facilite) Permis et marchés (protège) Réseau

# Chaîne 4: Connexion Moreau
@chain_moreau Mort Victor Moreau (donne) Héritage galerie (à) Jean Moreau (pour) Réseau Sokolov

-:: _sequence_ ::

// =============================================================
// RÉFÉRENCES CROISÉES - Pour utilisation $alias.n
// =============================================================

:: références croisées ::

# Alias pour références rapides
suspects => {Viktor Sokolov, Jean Moreau, Roux Constructions, Marc Delmas, Antoine Mercier}
temoins => {Claire Fontaine, Agent Interpol Müller}
lieux => {Galerie L'Éclipse, Bar Le Diplomate, Galerie Moreau Antiquités}
preuves => {Factures falsifiées, Liste Interpol, Écoutes téléphoniques, Témoignage Fontaine, Œuvres saisies}

# Connexions avec autres affaires
affaire_moreau => case-moreau-001
affaire_disparition => case-disparition-002

// =============================================================
// NOTES D'ENQUÊTE - TODO items
// =============================================================

:: notes, TODO ::

LOCALISER VIKTOR SOKOLOV (EN FUITE)
ANALYSER CONNEXION AVEC MEURTRE VICTOR MOREAU
VÉRIFIER SI SOPHIE LAURENT AVAIT DÉCOUVERT LE RÉSEAU
TRACER LES 9 ŒUVRES ENCORE MANQUANTES
INTERROGER JEAN MOREAU SUR SON RÔLE DANS LES DEUX AFFAIRES
AUDITIONNER ANTOINE MERCIER - LIEN MOREAU ET TRAFIC
IDENTIFIER AUTRES COLLECTIONNEURS-ACHETEURS
COOPÉRATION INTERNATIONALE POUR ARRESTATION SOKOLOV
`
}
