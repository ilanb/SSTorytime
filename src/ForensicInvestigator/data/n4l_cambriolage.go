package data

// N4L Content pour l'affaire Cambriolage Musée des Arts Premiers
// Ce fichier contient le contenu N4L authentique utilisant toutes les fonctionnalités du langage SSTorytime

// GetCambriolageN4LContent retourne le contenu N4L complet pour l'Affaire Cambriolage Musée
func GetCambriolageN4LContent() string {
	return `-affaire/cambriolage-004

# Affaire: Cambriolage Musée des Arts Premiers
# Type: Vol qualifié
# Investigation: Brigade de Répression du Banditisme
# Syntaxe: SSTorytime N4L v2

// =============================================================
// SECTION LIEUX - Scène de crime
// =============================================================

:: lieux, scène de crime ::

@musee Musée des Arts Premiers (type) lieu
    " (description) Musée municipal spécialisé dans l'art africain. Victime du cambriolage. Sécurité niveau 3 neutralisée.
    " (adresse) 15 rue des Beaux-Arts, Paris 6e
    " (superficie) 2500 m²
    " (securite) Niveau 3 - neutralisée
    " (alarme) Système Securitas - neutralisé
    " (acces) 3 entrées - principale, service, urgence
    " (latitude) 48.8566
    " (longitude) 2.3375

@salle_africaine Salle d'Art Africain (type) lieu
    " (description) Salle où étaient exposées les statuettes Dogon. 2e étage du musée.
    " (etage) 2e étage
    " (superficie) 150 m²
    " (vitrines) 12 vitrines sécurisées
    " (vitrines_forcees) 3 vitrines

// Relations Salle
$salle_africaine.1 (située dans:+C) $musee.1

@sortie_secours Sortie de secours (type) lieu
    " (description) Issue utilisée par les cambrioleurs. Donne sur l'allée de service.
    " (acces) Allée de service
    " (serrure) Forcée
    " (indices) Gants en latex retrouvés

// =============================================================
// SECTION SUSPECTS
// =============================================================

:: suspects ::

@agent Pierre Lafont (type) personne
    " (rôle) suspect principal
    " (description) Ancien agent de sécurité du musée, licencié il y a 6 mois. Connaît parfaitement le système d'alarme et les procédures.
    " (âge) 41 ans
    " (ancien_emploi) Agent sécurité musée
    " (licenciement) Avril 2025 - faute professionnelle
    " (connaissance) Système alarme, codes, rondes
    " (mobile) Revanche et argent
    " (alibi) Prétend être chez lui
    " (latitude) 48.8490
    " (longitude) 2.3520

// Relations de Pierre Lafont
$agent.1 (ancien employé de:N) $musee.1
$agent.1 (aurait collaboré avec:N) $reseau.1
$agent.1 (connaît les codes de:+L) $musee.1

@reseau Collectionneurs suspects (type) organisation
    " (rôle) suspect
    " (description) Réseau international de collectionneurs d'art africain identifié par Interpol. Commanditaires présumés.
    " (membres) 5-10 personnes identifiées
    " (nationalites) Russe, Belge, Américain
    " (specialite) Art africain ancien
    " (interpol) Fichés - trafic d'œuvres

// Relations du réseau
$reseau.1 (intéressé par:+L) $statuettes.1
$reseau.1 (a ciblé:+L) $musee.1
$reseau.1 (aurait recruté:N) $agent.1

@receleur Galerie Brunel (type) organisation
    " (rôle) suspect
    " (description) Galerie d'art parisienne soupçonnée de recel. Spécialisée dans l'art africain.
    " (adresse) 42 rue de Seine, Paris 6e
    " (proprietaire) Jean-Marc Brunel
    " (reputation) Douteuse - déjà enquêtée
    " (specialite) Art africain et océanien

// Relations du receleur
$receleur.1 (pourrait recevoir:N) $statuettes.1
$receleur.1 (en contact avec:N) $reseau.1

@complice Individu non identifié (type) personne
    " (rôle) suspect
    " (description) Complice vu sur les images de vidéosurveillance avant la coupure. Silhouette masculine, 1m75-1m80.
    " (taille) 1m75-1m80
    " (vetements) Noir - cagoule
    " (identification) En cours

// =============================================================
// SECTION TÉMOINS
// =============================================================

:: témoins ::

@gardien Robert Martinez (type) personne
    " (rôle) temoin
    " (description) Gardien de nuit du musée. A découvert le vol à 6h. N'a rien entendu pendant sa ronde de 3h.
    " (âge) 55 ans
    " (profession) Gardien de nuit
    " (decouverte) 6h00 le 02/10/2025
    " (ronde) 3h00 - rien remarqué
    " (observation) Système d'alarme semblait normal

// Relations du gardien
$gardien.1 (employé de:N) $musee.1
$gardien.1 (a découvert:+L) Vol

@conservateur Dr. Émilie Durand (type) personne
    " (rôle) temoin
    " (description) Conservatrice en chef du musée. Experte en art africain. A évalué les pièces volées.
    " (profession) Conservatrice - Docteur en histoire de l'art
    " (specialite) Art africain
    " (evaluation) 2.8 millions € pour les 5 statuettes

// Relations de la conservatrice
$conservateur.1 (responsable de:N) $salle_africaine.1
$conservateur.1 (a évalué:N) $statuettes.1

@voisin Claude Bernard (type) personne
    " (rôle) temoin
    " (description) Commerçant voisin. A vu un fourgon blanc suspect garé dans l'allée de service vers 2h30.
    " (profession) Commerçant
    " (observation) Fourgon blanc, 2h30
    " (description_vehicule) Renault Master blanc, immatriculation partielle

// =============================================================
// SECTION OBJETS VOLÉS
// =============================================================

:: objets volés ::

@statuettes Statuettes Dogon (type) objet
    " (description) 5 statuettes rituelles du Mali, XIVe siècle. Pièces maîtresses de la collection africaine.
    " (origine) Mali - Culture Dogon
    " (epoque) XIVe siècle
    " (nombre) 5 pièces
    " (valeur) 2.8 millions €
    " (provenance) Acquisition 1985 - Mission ethnographique
    " (statut) Volées

// Détail des statuettes
@statuette_1 Statuette Nommo n°1 (type) objet
    " (description) Représentation du génie de l'eau. Bois iroko. 45cm.
    " (hauteur) 45 cm
    " (materiau) Bois iroko
    " (valeur) 800 000 €

@statuette_2 Statuette Ancêtre féminin (type) objet
    " (description) Figure d'ancêtre féminine. Bois et métal. 38cm.
    " (hauteur) 38 cm
    " (materiau) Bois et métal
    " (valeur) 650 000 €

@statuette_3 Statuette Cavalier (type) objet
    " (description) Cavalier Dogon. Exceptionnelle par sa taille. 52cm.
    " (hauteur) 52 cm
    " (materiau) Bois
    " (valeur) 550 000 €

@statuette_4 Statuette Couple primordial (type) objet
    " (description) Couple primordial enlacé. Grande rareté. 35cm.
    " (hauteur) 35 cm
    " (materiau) Bois noirci
    " (valeur) 500 000 €

@statuette_5 Statuette Hogon (type) objet
    " (description) Chef spirituel Hogon. Traces de libations rituelles. 40cm.
    " (hauteur) 40 cm
    " (materiau) Bois avec patine
    " (valeur) 300 000 €

// Relations des statuettes
$statuettes.1 (exposées dans:+C) $salle_africaine.1
$statuettes.1 (convoitées par:N) $reseau.1

// =============================================================
// SECTION PREUVES - Indices matériels et numériques
// =============================================================

:: preuves, indices ::

// Groupement des preuves par type
Preuves physiques => {Gants en latex, Traces de fourgon, Vitrine forcée}
Preuves numériques => {Vidéosurveillance, Signal d'alarme}

@gants Gants en latex (type) preuve physique
    " (localisation) Près $sortie_secours.1
    " (description) Paire de gants chirurgicaux noirs
    " (analyse) ADN en cours d'analyse
    " (fiabilité) 8/10

@traces Traces de fourgon (type) preuve forensique
    " (localisation) Allée de service
    " (description) Empreintes de pneus dans la boue
    " (modele) Renault Master
    " (fiabilité) 6/10
    " (correspondance) Témoignage $voisin.1

@video Vidéosurveillance neutralisée (type) preuve numérique
    " (localisation) Système central $musee.1
    " (description) Signal coupé de 2h15 à 3h45. Image de boucle détectée.
    " (methode) Boucle vidéo - professionnel
    " (derniere_image) 2h14 - silhouette suspecte
    " (fiabilité) 7/10

@vitrine Vitrine forcée (type) preuve physique
    " (localisation) $salle_africaine.1
    " (description) 3 vitrines ouvertes sans bruit - technique professionnelle
    " (methode) Découpe diamant sur verre blindé
    " (empreintes) Aucune - gants utilisés
    " (fiabilité) 9/10

@alarme Signal d'alarme (type) preuve numérique
    " (description) Aucune alerte déclenchée - système neutralisé de l'intérieur
    " (methode) Code d'accès utilisé
    " (code) Ancien code - changé après licenciement Lafont?
    " (fiabilité) 8/10
    " (concerne) $agent.1

// =============================================================
// SECTION CHRONOLOGIE - Séquence temporelle complète
// =============================================================

:: chronologie ::

+:: _timeline_, _sequence_ ::

// ==========================================
// Contexte (avant le vol)
// ==========================================

@evt_c_00 01/04/2025 09:00 Licenciement Pierre Lafont (lieu) Musée
    " (description) $agent.1 licencié pour faute professionnelle (négligence)
    " (importance) high
    " (vérifié) oui
    " (implique) $agent.1, $musee.1

@evt_c_00b 15/04/2025 09:00 Codes d'accès censés être changés (lieu) Musée
    " (description) Procédure de changement des codes après licenciement
    " (importance) high
    " (vérifié) non - à vérifier
    " (implique) $musee.1

@evt_c_00c 15/09/2025 14:00 Visite suspecte au musée (lieu) Salle africaine
    " (description) Homme photographiant longuement les vitrines - identifié?
    " (importance) medium
    " (vérifié) non
    " (implique) $salle_africaine.1

// ==========================================
// Nuit du cambriolage (1-2 octobre 2025)
// ==========================================

@evt_c_01 01/10/2025 18:00 Fermeture musée (lieu) Musée
    " (description) Fermeture normale du musée au public
    " (importance) medium
    " (vérifié) oui
    " (implique) $musee.1

@evt_c_01b 01/10/2025 22:00 Début service gardien (lieu) Musée
    " (description) $gardien.1 prend son service de nuit
    " (importance) medium
    " (vérifié) oui
    " (implique) $gardien.1

@evt_c_02 02/10/2025 02:14 Dernière image vidéosurveillance (lieu) Entrée de service
    " (description) Silhouette suspecte captée avant la coupure
    " (importance) high
    " (vérifié) oui
    " (implique) $complice.1
    " (preuve) $video.1

@evt_c_03 02/10/2025 02:15 Coupure vidéosurveillance (lieu) Musée
    " (description) Système vidéo neutralisé - boucle insérée
    " (importance) high
    " (vérifié) oui
    " (preuve) $video.1

@evt_c_03b 02/10/2025 02:30 Fourgon blanc observé (lieu) Allée de service
    " (description) Fourgon blanc aperçu par $voisin.1
    " (importance) high
    " (vérifié) oui
    " (implique) $voisin.1
    " (preuve) $traces.1

@evt_c_04 02/10/2025 02:30 Intrusion estimée (lieu) Musée
    " (description) Entrée des cambrioleurs - durée estimée 1h15
    " (importance) high
    " (vérifié) non - estimation
    " (implique) $sortie_secours.1

@evt_c_04b 02/10/2025 03:00 Ronde gardien - RAS (lieu) Musée
    " (description) $gardien.1 effectue sa ronde - ne remarque rien
    " (importance) medium
    " (vérifié) oui
    " (implique) $gardien.1

@evt_c_05 02/10/2025 03:45 Retour vidéosurveillance (lieu) Musée
    " (description) Système vidéo revient à la normale - cambrioleurs partis
    " (importance) high
    " (vérifié) oui
    " (preuve) $video.1

@evt_c_06 02/10/2025 06:00 Découverte vol (lieu) Salle africaine
    " (description) $gardien.1 découvre les vitrines vides lors de sa ronde matinale
    " (importance) high
    " (vérifié) oui
    " (implique) $gardien.1, $salle_africaine.1

@evt_c_07 02/10/2025 06:15 Alerte police (lieu) Musée
    " (description) Appel au commissariat - arrivée police 6h30
    " (importance) high
    " (vérifié) oui

// ==========================================
// Chaînes causales
// ==========================================

// Chaîne causale principale: préparation
$evt_c_00 (mène à:+L) Connaissance système
Connaissance système (permet:+L) Neutralisation alarme
$evt_c_00c (mène à:+L) Repérage cibles

// Chaîne causale: exécution
$evt_c_03 (permet:+L) Intrusion non détectée
Intrusion non détectée (permet:+L) Vol des statuettes
$evt_c_03b (confirme:N) Véhicule de fuite

-:: _timeline_, _sequence_ ::

// =============================================================
// SECTION HYPOTHÈSES - Pistes d'enquête
// =============================================================

:: hypothèses, pistes ::

@hyp_c_01 Complicité interne - Lafont (type) hypothèse
    " (statut) en_attente
    " (confiance) 70%
    " (source) user
    " (description) L'ancien agent de sécurité Pierre Lafont aurait fourni les codes d'accès et les plans du système de sécurité à une équipe de professionnels. Son licenciement récent lui donne un mobile de revanche.
    " (mobile) Revanche + argent
    " (pour) $alarme.1, Connaissance du système, Licenciement récent
    " (contre) Alibi à vérifier
    " (questions) Les codes ont-ils été changés après son départ?; Contacts avec le réseau de collectionneurs?; Où était-il la nuit du vol?
    " (suspect) $agent.1

@hyp_c_02 Commande du réseau international (type) hypothèse
    " (statut) en_attente
    " (confiance) 65%
    " (source) user
    " (description) Le vol aurait été commandité par le réseau de collectionneurs identifié par Interpol. Les statuettes Dogon étaient spécifiquement ciblées pour un acheteur.
    " (pour) $reseau.1 fiché Interpol, Visite suspecte septembre, Valeur des pièces
    " (contre) Pas de preuves directes
    " (questions) Qui est le commanditaire final?; Les pièces ont-elles déjà quitté la France?; Lien avec autres vols similaires?
    " (suspect) $reseau.1

@hyp_c_03 Recel via galerie Brunel (type) hypothèse
    " (statut) en_attente
    " (confiance) 45%
    " (source) ai
    " (description) La galerie Brunel, déjà soupçonnée de recel, pourrait servir d'intermédiaire pour écouler les œuvres volées vers des collectionneurs privés.
    " (pour) Spécialisation art africain, Réputation douteuse
    " (contre) Sous surveillance depuis enquête précédente
    " (questions) Contacts récents avec le réseau?; Mouvements financiers suspects?
    " (suspect) $receleur.1

// =============================================================
// RÉSEAU DE RELATIONS - Graphe sémantique
// =============================================================

:: réseau de relations ::

# Légende STTypes: N=proximité, +L=causalité, +C=containment, +E=expression

// Relations avec le musée
$agent.1 (ancien employé de:N) $musee.1
$gardien.1 (employé de:N) $musee.1
$conservateur.1 (responsable de:+C) $salle_africaine.1

// Relations de complicité présumée
$agent.1 (aurait informé:+L) $reseau.1
$reseau.1 (aurait commandité:+L) Vol
$reseau.1 (utilise:N) $receleur.1

// Relations avec les objets volés
$statuettes.1 (volées dans:+C) $salle_africaine.1
$statuettes.1 (convoitées par:N) $reseau.1

// =============================================================
// CHAÎNES CAUSALES DÉTECTÉES
// =============================================================

:: chaînes causales ::

+:: _sequence_ ::

# Chaîne 1: Préparation
@chain_prep Licenciement Lafont (mène à) Accès aux codes (mène à) Planification vol (mène à) Repérage

# Chaîne 2: Exécution
@chain_exec Neutralisation vidéo (puis) Neutralisation alarme (puis) Entrée service (puis) Découpe vitrines (puis) Vol statuettes (puis) Fuite fourgon

# Chaîne 3: Recel
@chain_recel Vol (mène à) Transport (mène à) Recel galerie (mène à) Vente collectionneurs

-:: _sequence_ ::

// =============================================================
// RÉFÉRENCES CROISÉES - Pour utilisation $alias.n
// =============================================================

:: références croisées ::

# Alias pour références rapides
suspects => {Pierre Lafont, Collectionneurs suspects, Galerie Brunel, Individu non identifié}
temoins => {Robert Martinez, Dr. Émilie Durand, Claude Bernard}
lieux => {Musée des Arts Premiers, Salle d'Art Africain, Sortie de secours}
objets => {Statuettes Dogon}
preuves => {Gants en latex, Traces de fourgon, Vidéosurveillance, Vitrine forcée}

// =============================================================
// NOTES D'ENQUÊTE - TODO items
// =============================================================

:: notes, TODO ::

VÉRIFIER SI LES CODES ONT ÉTÉ CHANGÉS APRÈS LICENCIEMENT LAFONT
ANALYSER ADN SUR GANTS EN LATEX
IDENTIFIER LE FOURGON BLANC RENAULT MASTER
CONTACTER INTERPOL POUR RÉSEAU COLLECTIONNEURS
SURVEILLER GALERIE BRUNEL ET MARCHÉS DE L'ART
VÉRIFIER ALIBI DE PIERRE LAFONT
IDENTIFIER L'HOMME DE LA VISITE DU 15/09
ALERTER DOUANES ET PORTS/AÉROPORTS
`
}
