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
$intermediaire.1 (complice de:+L) $chef_reseau.1
$intermediaire.1 (fréquente:N) $bar.1
$intermediaire.1 (a accès à:+C) $galerie_moreau.1

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
$chef_reseau.1 (blanchit via:+L) $btp.1
$chef_reseau.1 (commandite:+L) Vols d'œuvres
$chef_reseau.1 (dirige:+C) Réseau international

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
$btp.1 (verse des commissions à:-C) $politicien.1
$btp.1 (blanchit pour:+L) $chef_reseau.1

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
$politicien.1 (reçoit des pots-de-vin de:-C) $btp.1
$politicien.1 (possède:+C) Œuvres volées

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
$expert.1 (travaille pour:+L) $chef_reseau.1
$expert.1 (expertisait pour:N) $galerie_moreau.1

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
$galerie_eclipse.1 (contrôlée par:+C) $chef_reseau.1
$galerie_eclipse.1 (écoule:+L) Œuvres volées

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
$galerie_moreau.1 (héritée par:N) $intermediaire.1
$galerie_moreau.1 (expertisée par:N) $expert.1

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
$temoin_cle.1 (ancienne employée de:N) $galerie_eclipse.1
$temoin_cle.1 (a dénoncé:+L) $chef_reseau.1
$temoin_cle.1 (a identifié:+L) $intermediaire.1, $expert.1

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
    " (concerne) $btp.1, $politicien.1

@liste_interpol Liste d'œuvres volées Interpol (type) preuve documentaire
    " (source) Base de données Interpol
    " (description) Inventaire des 12 œuvres transitées par le réseau
    " (oeuvres_retrouvees) 3 chez $politicien.1
    " (oeuvres_recherchees) 9 en cours de localisation
    " (fiabilité) 10/10

@ecoutes Écoutes téléphoniques (type) preuve numérique
    " (periode) Septembre-Octobre 2025
    " (description) Conversations entre $chef_reseau.1 et $intermediaire.1
    " (contenu) Mentions de 'livraisons', 'l'oncle', 'la galerie'
    " (fiabilité) 8/10
    " (concerne) $intermediaire.1, $chef_reseau.1

@video_bar Vidéosurveillance Bar Le Diplomate (type) preuve numérique
    " (date) 05/10/2025
    " (description) $intermediaire.1 remet une enveloppe à un homme non identifié
    " (localisation) $bar.1
    " (fiabilité) 7/10
    " (concerne) $intermediaire.1

@temoignage Témoignage Claire Fontaine (type) preuve testimoniale
    " (temoin) $temoin_cle.1
    " (description) Décrit le processus de blanchiment et nomme Sokolov, Moreau et Mercier
    " (elements) Schéma du réseau, contacts, méthodes
    " (fiabilité) 8/10
    " (concerne) $chef_reseau.1, $intermediaire.1, $expert.1

@oeuvres_saisies Œuvres saisies chez Delmas (type) preuve physique
    " (localisation) Domicile de $politicien.1
    " (description) 3 œuvres d'art africain figurant sur la liste Interpol
    " (valeur) 450 000 € estimés
    " (provenance) Réseau Sokolov
    " (fiabilité) 10/10
    " (concerne) $politicien.1, $chef_reseau.1

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
    " (implique) $chef_reseau.1

@evt_t_00b 15/03/2024 09:00 Claire Fontaine embauchée à L'Éclipse (lieu) Galerie L'Éclipse
    " (description) $temoin_cle.1 commence à travailler à la galerie
    " (importance) medium
    " (vérifié) oui
    " (implique) $temoin_cle.1, $galerie_eclipse.1

// ==========================================
// Connexion avec Affaire Moreau (août 2025)
// ==========================================

@evt_t_01 27/08/2025 22:00 Jean Moreau vu au Bar Le Diplomate (lieu) Bar Le Diplomate
    " (description) $intermediaire.1 rencontre un individu non identifié - même soir mentionné dans affaire Moreau
    " (importance) high
    " (vérifié) oui
    " (implique) $intermediaire.1, $bar.1
    " (connexion) Affaire Moreau - evt-m-01b

@evt_t_02 29/08/2025 20:30 Décès de Victor Moreau (lieu) Manoir Moreau
    " (description) Victor Moreau décède empoisonné - affaire case-moreau-001
    " (importance) high
    " (vérifié) oui
    " (implique) $intermediaire.1
    " (connexion) Affaire Moreau - $intermediaire.1 héritier et suspect

// ==========================================
// Dénonciation et enquête (août-octobre 2025)
// ==========================================

@evt_t_03 15/08/2025 09:00 Démission et dénonciation Claire Fontaine (lieu) Anonyme
    " (description) $temoin_cle.1 quitte la galerie et alerte anonymement les autorités
    " (importance) high
    " (vérifié) oui
    " (implique) $temoin_cle.1

// CONNEXION AFFAIRE DISPARITION
@evt_t_04 15/09/2025 19:48 Disparition Sophie Laurent (lieu) Parking Toulouse
    " (description) Sophie Laurent disparaît - enquêtait sur Roux Constructions et Delmas
    " (importance) high
    " (vérifié) oui
    " (implique) $politicien.1, $btp.1
    " (connexion) Affaire Disparition - case-disparition-002

@evt_t_05 01/10/2025 09:00 Mise sur écoute du réseau (lieu) Paris
    " (description) Autorisation d'écoutes téléphoniques sur Sokolov et Moreau
    " (importance) high
    " (vérifié) oui
    " (implique) $chef_reseau.1, $intermediaire.1

@evt_t_06 05/10/2025 21:00 Remise d'enveloppe au Bar Le Diplomate (lieu) Bar Le Diplomate
    " (description) $intermediaire.1 filmé remettant une enveloppe
    " (importance) high
    " (vérifié) oui
    " (implique) $intermediaire.1, $bar.1
    " (preuve) $video_bar.1

@evt_t_07 10/10/2025 06:00 Perquisitions simultanées (lieu) Paris, Toulouse, Luxembourg
    " (description) Perquisitions chez Delmas, Roux Constructions, Galerie L'Éclipse
    " (importance) high
    " (vérifié) oui
    " (implique) $politicien.1, $btp.1, $galerie_eclipse.1
    " (preuve) $oeuvres_saisies.1

@evt_t_08 10/10/2025 14:00 Saisie des 3 œuvres chez Delmas (lieu) Domicile Delmas
    " (description) 3 œuvres d'art africain saisies - figurent sur liste Interpol
    " (importance) high
    " (vérifié) oui
    " (implique) $politicien.1
    " (preuve) $oeuvres_saisies.1, $liste_interpol.1

@evt_t_09 14/10/2025 09:00 Mandat d'arrêt international Sokolov (lieu) Interpol Lyon
    " (description) Mandat d'arrêt émis contre Viktor Sokolov
    " (importance) high
    " (vérifié) oui
    " (implique) $chef_reseau.1

// ==========================================
// Chaînes causales
// ==========================================

// Chaîne causale: connexion Moreau
$evt_t_01 (même contexte que:N) Meurtre Victor Moreau
Meurtre Victor Moreau (donne accès à:+L) $galerie_moreau.1
$galerie_moreau.1 (utilisée pour:+L) Recel

// Chaîne causale: connexion Disparition
$evt_t_04 (liée à enquête sur:N) $btp.1
$btp.1 (blanchit pour:+L) $chef_reseau.1

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
    " (pour) $temoignage.1, $ecoutes.1, $liste_interpol.1
    " (contre) Sokolov en fuite - pas encore arrêté
    " (suspect) $chef_reseau.1

@hyp_t_02 Connexion meurtre Victor Moreau (type) hypothèse
    " (statut) en_attente
    " (confiance) 70%
    " (source) ai
    " (description) Le meurtre de Victor Moreau (case-moreau-001) pourrait être lié au trafic d'art. Jean Moreau hérite de la galerie qui servait peut-être au recel, et Antoine Mercier était un rival impliqué dans le réseau.
    " (pour) $intermediaire.1 héritier, $expert.1 impliqué dans les deux affaires, $galerie_moreau.1 suspectée
    " (contre) Pas de preuves directes du lien
    " (questions) Victor Moreau était-il impliqué ou victime?; Le meurtre visait-il à prendre le contrôle de la galerie?
    " (connexion) case-moreau-001

@hyp_t_03 Connexion disparition Sophie Laurent (type) hypothèse
    " (statut) en_attente
    " (confiance) 75%
    " (source) ai
    " (description) La disparition de Sophie Laurent (case-disparition-002) pourrait être liée au volet blanchiment du réseau. Elle enquêtait sur Roux Constructions et Marc Delmas, tous deux impliqués dans le blanchiment pour Sokolov.
    " (pour) $btp.1 dans les deux affaires, $politicien.1 dans les deux affaires
    " (contre) Pas de preuves que Sophie connaissait le volet artistique
    " (questions) Sophie avait-elle découvert le lien avec le trafic d'art?; Son informateur Source X connaissait-il le réseau Sokolov?
    " (connexion) case-disparition-002

@hyp_t_04 Complicité de Delmas pour protection (type) hypothèse
    " (statut) confirmée
    " (confiance) 85%
    " (source) user
    " (description) Marc Delmas facilite les activités de Roux Constructions en échange d'œuvres d'art. Il utilise sa position pour protéger le réseau de blanchiment.
    " (pour) $oeuvres_saisies.1, $factures.1, Position politique
    " (contre) Nie toute implication
    " (suspect) $politicien.1

// =============================================================
// RÉSEAU DE RELATIONS - Graphe sémantique
// =============================================================

:: réseau de relations ::

# Légende STTypes: N=proximité, +L=causalité, +C=containment, +E=expression

// Hiérarchie du réseau
$chef_reseau.1 (dirige:+C) Réseau international
$chef_reseau.1 (recrute:+L) $intermediaire.1
$chef_reseau.1 (emploie:+L) $expert.1

// Blanchiment
$chef_reseau.1 (blanchit via:+L) $btp.1
$btp.1 (verse des commissions à:-C) $politicien.1
$politicien.1 (facilite permis pour:+L) $btp.1

// Recel
$galerie_eclipse.1 (écoule:+L) Œuvres volées
$expert.1 (authentifie pour:+L) $galerie_eclipse.1
$galerie_moreau.1 (servait au recel pour:N) $chef_reseau.1

// Connexions inter-affaires
$intermediaire.1 (suspect dans:N) Affaire Moreau
$politicien.1 (suspect dans:N) Affaire Disparition
$btp.1 (impliqué dans:N) Affaire Disparition

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
