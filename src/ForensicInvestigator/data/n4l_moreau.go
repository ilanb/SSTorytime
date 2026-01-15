package data

// N4L Content pour les affaires de démonstration
// Ce fichier contient le contenu N4L authentique utilisant toutes les fonctionnalités du langage SSTorytime

// GetMoreauN4LContent retourne le contenu N4L complet pour l'Affaire Moreau
// Cet exemple démontre toutes les fonctionnalités avancées de N4L
func GetMoreauN4LContent() string {
	return `-affaire/moreau-001

# Affaire: Homicide Victor Moreau
# Type: Empoisonnement
# Investigation: Brigade Criminelle Paris
# Syntaxe: SSTorytime N4L v2

// =============================================================
// SECTION VICTIMES - Définitions avec alias et attributs
// =============================================================

:: victimes ::

@victime Victor Moreau (type) personne
    " (rôle) victime
    " (description) Antiquaire renommé, 67 ans, veuf depuis 2019. Fortune estimée à 12 millions d'euros. Décédé le 29/08/2025 vers 20h30 par empoisonnement.
    " (âge) 67 ans
    " (profession) Antiquaire renommé
    " (domicile) Manoir rue de Varenne, Paris 7e
    " (fortune) 12 millions d'euros
    " (cause_décès) Empoisonnement par alcaloïde
    " (date_décès) 29/08/2025
    " (heure_décès) 20h30
    " (état_civil) Veuf depuis 2019
    " (membre_de) Cercle des Bibliophiles Parisiens
    " (spécialité) Livres rares XVIIIe siècle
    " (latitude) 48.8566
    " (longitude) 2.3175

// =============================================================
// SECTION SUSPECTS - Avec mobiles et relations
// =============================================================

:: suspects ::

@neveu Jean Moreau (type) personne
    " (rôle) suspect
    " (description) Neveu de la victime, 35 ans, sans emploi fixe. Dettes de jeu importantes (150 000€). Héritier principal (8 millions). Alibi: cinéma UGC Bercy 19h-22h.
    " (âge) 35 ans
    " (profession) Sans emploi
    " (mobile) Héritage 8 millions EUR
    " (dettes) 150 000 EUR - Casino de Deauville
    " (alibi) Cinéma UGC Bercy 19h-22h
    " (alibi_vérifié) Partiellement - ticket trouvé mais pas vu par caméras
    " (véhicule) BMW série 3 noire AB-123-CD
    " (pointure) 42
    " (comportement) Nerveux lors interrogatoire
    " (dernier_contact) Appel à $victime.1 le 29/08 à 18h45
    " (latitude) 48.8396
    " (longitude) 2.3876

// Relations de Jean Moreau
$neveu.1 (hérite de:+L) $victime.1
$neveu.1 (doit de l'argent à:-C) $casino.1
$neveu.1 (connaît le code d'accès:N) $bibliotheque.1
$neveu.1 (vu au:N) Bar Le Diplomate
$neveu.1 (alibi à:N) $ugc.1

@exassociee Élodie Dubois (type) personne
    " (rôle) suspect
    " (description) Ex-associée de Victor, 42 ans. Perte financière de 2 millions en 2024. Procédure judiciaire en cours. Menaces proférées: 'Il paiera pour ce qu'il m'a fait'.
    " (âge) 42 ans
    " (profession) Femme d'affaires
    " (mobile) Vengeance - perte 2 millions EUR
    " (procédure) Procès civil en cours contre $victime.1
    " (menaces) Il paiera pour ce qu'il m'a fait
    " (alibi) Dîner charité Hôtel Crillon 19h-23h
    " (avocat) Maître Lefebvre
    " (latitude) 48.8686
    " (longitude) 2.3215

// Relations d'Élodie Dubois
$exassociee.1 (en conflit avec:N) $victime.1
$exassociee.1 (a menacé:+L) $victime.1
$exassociee.1 (représentée par:N) $avocat.1
$exassociee.1 (alibi à:N) $crillon.1

@concurrent Antoine Mercier (type) personne
    " (rôle) suspect
    " (description) Expert en livres anciens et concurrent de Victor. Conflit sur vente aux enchères en juillet 2025.
    " (âge) 55 ans
    " (profession) Expert en livres anciens
    " (galerie) Mercier & Fils
    " (mobile) Rivalité commerciale
    " (conflit) Vente aux enchères juillet 2025
    " (alibi) Non vérifié

// Relations d'Antoine Mercier
$concurrent.1 (rival de:N) $victime.1
$concurrent.1 (a perdu aux enchères contre:N) $victime.1

// =============================================================
// SECTION TÉMOINS - Observations et dépositions
// =============================================================

:: témoins ::

@gouvernante Madame Chen (type) personne
    " (rôle) temoin
    " (description) Gouvernante du manoir depuis 15 ans. Présente le soir du crime jusqu'à 19h. A servi le thé à 19h15. A découvert le corps à 21h45.
    " (âge) 45 ans
    " (profession) Gouvernante
    " (ancienneté) 15 ans au manoir
    " (présence) Jusqu'à 19h le 29/08
    " (action) A servi le thé à 19h15
    " (découverte) Corps découvert à 21h45
    " (observation) Victor semblait nerveux ce soir-là
    " (latitude) 48.8566
    " (longitude) 2.3175

// Relations de Madame Chen
$gouvernante.1 (employée par:N) $victime.1

@jardinier Robert Duval (type) personne
    " (rôle) temoin
    " (description) Jardinier. Présent jusqu'à 18h. A observé la fenêtre bibliothèque ouverte à 17h30 et des traces de pas inhabituelles près des rosiers.
    " (âge) 58 ans
    " (profession) Jardinier
    " (présence) Jusqu'à 18h le 29/08
    " (observation) Fenêtre bibliothèque ouverte à 17h30
    " (indices) Traces de pas inhabituelles près des rosiers
    " (latitude) 48.8570
    " (longitude) 2.3180

@legiste Dr. Sarah Martin (type) personne
    " (rôle) temoin
    " (description) Médecin légiste. Rapport préliminaire: décès entre 20h et 21h par alcaloïde non identifié.
    " (profession) Médecin légiste
    " (rapport) Décès entre 20h et 21h
    " (cause) Alcaloïde végétal non identifié
    " (remarque) Aucun signe de lutte

@avocat Maître Lefebvre (type) personne
    " (rôle) temoin
    " (description) Avocat d'Élodie Dubois. Agressif mais efficace. A demandé une saisie conservatoire.
    " (profession) Avocat
    " (spécialité) Droit des affaires
    " (réputation) Agressif mais efficace
    " (client) $exassociee.1
    " (présent) Confrontation tribunal 25/08

// Relations Maître Lefebvre
$avocat.1 (représente:N) $exassociee.1
$avocat.1 (a attaqué:N) $victime.1

@notaire Maître Durand (type) personne
    " (rôle) temoin
    " (profession) Notaire
    " (cabinet) Étude notariale Paris 7e
    " (action) Rencontre avec $victime.1 le 28/08

// Relations Maître Durand
$notaire.1 (a reçu:N) $victime.1
$notaire.1 (a rédigé:N) $testament.1

@individu_inconnu Homme non identifié (type) personne
    " (rôle) suspect
    " (description) Personne non identifiée rencontrée par Jean au Bar Le Diplomate le 27/08.
    " (date) 27/08/2025
    " (statut) À identifier

@appelant_inconnu Homme au téléphone inconnu (type) personne
    " (rôle) suspect
    " (description) Personne non identifiée ayant appelé Victor à 19h05 le soir du crime. Durée: 2 minutes.
    " (heure_appel) 19h05
    " (durée) 2 minutes
    " (numéro) Numéro masqué
    " (identification) En cours

// Relations appelant inconnu
$appelant_inconnu.1 (a appelé:+L) $victime.1

// =============================================================
// SECTION LIEUX - Scène de crime et localisations
// =============================================================

:: lieux, scène de crime ::

@bibliotheque Bibliothèque du Manoir (type) lieu
    " (description) Scène de crime principale. RDC, 8x6m. Fenêtre ouest ouverte. Corps trouvé dans fauteuil près de la cheminée.
    " (adresse) Manoir Moreau, Rue de Varenne, Paris 7e
    " (étage) Rez-de-chaussée
    " (dimensions) 8m x 6m
    " (accès) Porte principale + porte-fenêtre jardin
    " (état_fenêtre) Ouverte, empreintes essuyées
    " (latitude) 48.8556
    " (longitude) 2.3177

// Relations Bibliothèque
$bibliotheque.1 (propriété de:N) $victime.1

@jardin Jardin du Manoir (type) lieu
    " (description) Jardin de 500m² avec haie haute côté rue. Traces de pas près des rosiers.
    " (adresse) Manoir Moreau, Rue de Varenne, Paris 7e
    " (superficie) 500m²
    " (particularité) Haie haute côté rue
    " (accès) Portillon avec clé
    " (indices) Traces de pas, terre argileuse
    " (latitude) 48.8557
    " (longitude) 2.3175

// Relations Jardin
$jardin.1 (donne accès à:+C) $bibliotheque.1

@portillon Portillon du jardin (type) lieu
    " (type) Accès secondaire
    " (sécurité) Fermé à clé
    " (clé) Détenue par $jardinier.1
    " (état_29_08) Fermé à 18h par $jardinier.1

// Relations Portillon
$portillon.1 (donne accès à:+C) $jardin.1

@casino Casino de Deauville (type) lieu
    " (description) Casino où Jean Moreau a contracté des dettes importantes.
    " (adresse) 2 Rue Edmond Blanc, 14800 Deauville
    " (type_lieu) Casino
    " (lien) Créancier de $neveu.1 - 150k EUR
    " (latitude) 49.3565
    " (longitude) -0.0742

@ugc UGC Bercy (type) lieu
    " (description) Cinéma où Jean Moreau prétend avoir passé la soirée du crime (19h-22h).
    " (adresse) 2 Cour Saint-Émilion, Paris 12e
    " (type_lieu) Cinéma
    " (alibi) Lieu déclaré par $neveu.1 le soir du crime
    " (latitude) 48.8335
    " (longitude) 2.3867

// Relations UGC
$ugc.1 (alibi de:N) $neveu.1

@crillon Hôtel de Crillon (type) lieu
    " (description) Palace parisien où Élodie Dubois assistait à un dîner de charité le soir du crime.
    " (adresse) 10 Place de la Concorde, Paris 8e
    " (type_lieu) Hôtel de luxe
    " (alibi) Lieu déclaré par $exassociee.1 le soir du crime
    " (latitude) 48.8677
    " (longitude) 2.3216

// Relations Crillon
$crillon.1 (alibi de:N) $exassociee.1

@bar Bar Le Diplomate (type) lieu
    " (description) Bar fréquenté par Jean Moreau. Lieu de rencontres suspects.
    " (adresse) 45 rue de Bercy, Paris 12e
    " (type_lieu) Bar de nuit
    " (latitude) 48.8387
    " (longitude) 2.3826

@galerie Galerie Moreau Antiquités (type) lieu
    " (description) Galerie d'antiquités appartenant à Victor Moreau, située rue du Faubourg Saint-Honoré.
    " (adresse) 12 rue du Faubourg Saint-Honoré, Paris 8e
    " (spécialité) Livres rares et manuscrits
    " (valeur_stock) 3.5 millions d'euros
    " (latitude) 48.8699
    " (longitude) 2.3189

// Relations Galerie
$galerie.1 (propriété de:N) $victime.1

@cercle Cercle des Bibliophiles Parisiens (type) organisation
    " (description) Club exclusif de collectionneurs de livres anciens. Victor en était membre depuis 20 ans.

// Relations Cercle
$cercle.1 (membre:N) $victime.1

// =============================================================
// SECTION PREUVES - Indices matériels et numériques
// =============================================================

:: preuves, indices ::

// Groupement des preuves par type
Preuves physiques => {Tasse de thé, Livre des poisons, Traces de boue}
Preuves numériques => {Téléphone Victor, Vidéosurveillance UGC}
Preuves documentaires => {Testament actuel, Brouillon testament}

@tasse Tasse de thé (type) preuve physique
    " (localisation) Table basse $bibliotheque.1
    " (contenu) Résidus Earl Grey + substance inconnue
    " (empreintes) Victor uniquement
    " (fiabilité) 9/10
    " (concerne) $victime.1

@livre Livre "Traité des Poisons Exotiques" (type) preuve physique
    " (édition) 1923
    " (localisation) Bureau de $victime.1
    " (page) 247 - Alcaloïdes végétaux
    " (annotations) *Récentes - encre fraîche
    " (empreintes) Partielles non identifiées
    " (fiabilité) 7/10

@traces Traces de boue (type) preuve forensique
    " (localisation) Tapis persan $bibliotheque.1
    " (composition) Terre argileuse des rosiers
    " (pointure) *42 - correspond à $neveu.1
    " (direction) Fenêtre vers fauteuil
    " (fraîcheur) < 24h
    " (fiabilité) 8/10

@telephone Téléphone de Victor (type) preuve numérique
    " (localisation) Sur la victime
    " (appels) $neveu.1 18h45, inconnu 19h05
    " (sms) Menaces de $exassociee.1
    " (recherches) "poisons" le 28/08
    " (fiabilité) 9/10
    " (concerne) $victime.1, $neveu.1, $exassociee.1

@testament Testament actuel (type) preuve documentaire
    " (date) 15/03/2025
    " (notaire) Maître Durand
    " (bénéficiaire) $neveu.1 80%
    " (statut) Valide et officiel

@brouillon Brouillon nouveau testament (type) preuve documentaire
    " (date) 28/08/2025
    " (statut) *Non signé - trouvé corbeille
    " (nouveau_bénéficiaire) Fondation Moreau 90%
    " (impact) Réduit $neveu.1 à 10%
    " (mobile) =déshéritage Renforce le mobile de $neveu.1

@sms SMS menaçants (type) preuve numérique
    " (source) Téléphone de $victime.1
    " (expéditeur) $exassociee.1
    " (contenu) Menaces de vengeance
    " (date) Août 2025
    " (fiabilité) 9/10

@camera Système vidéosurveillance (type) preuve technique
    " (localisation) Manoir
    " (état) Câble sectionné le 28/08
    " (découverte) Sabotage intentionnel
    " (heure_coupure) 18h00
    " (fiabilité) 8/10

@autopsie Rapport d'autopsie (type) preuve médicale
    " (médecin) $legiste.1
    " (date) 30/08/2025
    " (cause_décès) Empoisonnement par alcaloïde végétal
    " (heure_décès) Entre 20h et 21h
    " (signes) Aucune trace de lutte
    " (fiabilité) 10/10

// =============================================================
// SECTION CHRONOLOGIE - Séquence temporelle complète
// =============================================================

:: chronologie ::

+:: _timeline_, _sequence_ ::

// ==========================================
// Événements antérieurs (contexte)
// ==========================================

// Événements du 15 juillet 2025
@evt_00a 15/07/2025 14:00 Vente aux enchères contestée (lieu) Drouot
    " (description) $victime.1 remporte un manuscrit convoité par $concurrent.1 pour 450 000€
    " (importance) medium
    " (vérifié) oui
    " (implique) $victime.1, $concurrent.1

// Événements du 25 août 2025
@evt_00b 25/08/2025 10:00 Confrontation tribunal $victime.1/$exassociee.1 (lieu) Tribunal de Commerce
    " (description) $exassociee.1 menace $victime.1: 'Il paiera pour ce qu'il m'a fait'
    " (importance) high
    " (vérifié) oui
    " (implique) $victime.1, $exassociee.1, $avocat.1
    " (preuve) $sms.1

// ==========================================
// Semaine du crime (27-28 août 2025)
// ==========================================

// Événements du 27 août 2025
@evt_01 27/08/2025 14:00 Visite houleuse de $neveu.1 (lieu) Manoir
    " (description) Discussion avec $victime.1 sur l'argent. Claque la porte en partant.
    " (importance) high
    " (vérifié) oui
    " (implique) $victime.1, $neveu.1

@evt_01b 27/08/2025 22:00 $neveu.1 vu au Bar Le Diplomate (lieu) Bar Le Diplomate
    " (description) Rencontre avec individu non identifié
    " (importance) medium
    " (vérifié) non
    " (implique) $neveu.1, $individu_inconnu.1

// Événements du 28 août 2025
@evt_02 28/08/2025 10:00 $victime.1 chez le notaire (lieu) Étude notariale
    " (description) Évoque une modification du testament - veut déshériter $neveu.1
    " (importance) high
    " (vérifié) oui
    " (implique) $victime.1, $notaire.1
    " (preuve) $brouillon.1

@evt_02b 28/08/2025 18:00 Câble caméra sectionné (lieu) Manoir
    " (description) Système de surveillance neutralisé
    " (importance) high
    " (vérifié) oui
    " (preuve) $camera.1

// ==========================================
// Jour du crime (29 août 2025)
// ==========================================

// Événements du 29 août 2025
@evt_03 29/08/2025 15:00 $exassociee.1 vue près du manoir (lieu) Aux abords du manoir
    " (description) Aperçue par voisin M. Bertrand à 15h
    " (importance) high
    " (vérifié) non
    " (implique) $exassociee.1

@evt_04 29/08/2025 17:30 Fenêtre bibliothèque ouverte (lieu) Bibliothèque
    " (description) Observée par $jardinier.1 - traces de boue découvertes plus tard
    " (importance) medium
    " (vérifié) oui
    " (implique) $jardinier.1, $bibliotheque.1
    " (preuve) $traces.1

@evt_05 29/08/2025 18:00 Départ du jardinier (lieu) Jardin
    " (description) $jardinier.1 ferme le portillon à clé
    " (importance) medium
    " (vérifié) oui
    " (implique) $jardinier.1, $portillon.1

@evt_06 29/08/2025 18:45 Appel $neveu.1 vers $victime.1 (lieu) Téléphone
    " (description) Durée 3 minutes - contenu inconnu
    " (importance) high
    " (vérifié) oui
    " (implique) $neveu.1, $victime.1
    " (preuve) $telephone.1

@evt_07 29/08/2025 19:00 Départ $gouvernante.1 (lieu) Manoir
    " (description) Laisse $victime.1 seul après avoir servi le thé
    " (importance) high
    " (vérifié) oui
    " (implique) $gouvernante.1, $victime.1
    " (preuve) $tasse.1

@evt_08 29/08/2025 19:05 Appel numéro inconnu (lieu) Téléphone
    " (description) Appel entrant sur téléphone $victime.1 - durée 2min
    " (importance) high
    " (vérifié) oui
    " (implique) $victime.1, $appelant_inconnu.1
    " (preuve) $telephone.1

@evt_09 29/08/2025 19:15 $victime.1 boit son thé (lieu) Bibliothèque
    " (description) Dernière activité connue - thé possiblement empoisonné
    " (importance) high
    " (vérifié) non
    " (implique) $victime.1, $bibliotheque.1
    " (preuve) $tasse.1

@evt_09b 29/08/2025 19:25 Entrée cinéma $neveu.1 (alibi) (lieu) UGC Bercy
    " (description) Ticket acheté pour séance 19h30
    " (importance) high
    " (vérifié) oui
    " (implique) $neveu.1

@evt_10 29/08/2025 20:30 Heure estimée du décès (lieu) Bibliothèque
    " (description) Selon rapport médecin légiste Dr. Martin
    " (importance) high
    " (vérifié) oui
    " (implique) $victime.1, $medecin.1
    " (preuve) $tasse.1, $autopsie.1

@evt_11 29/08/2025 21:45 Découverte du corps (lieu) Bibliothèque
    " (description) Par $gouvernante.1 revenue au manoir
    " (importance) high
    " (vérifié) oui
    " (implique) $victime.1, $gouvernante.1, $bibliotheque.1
    " (preuve) $tasse.1, $traces.1

@evt_12 29/08/2025 22:00 Arrivée police (lieu) Manoir
    " (description) Début de l'enquête officielle
    " (importance) medium
    " (vérifié) oui

// ==========================================
// Chaînes causales
// ==========================================

// Chaîne causale principale: empoisonnement
$evt_09 (mène à:+L) Empoisonnement
Empoisonnement (mène à:+L) $evt_10
$evt_10 (découvert par:+L) $evt_11

// Chaîne causale: mobile financier
$neveu.1 (dettes:+L) $casino.1
$casino.1 (pression:+L) $neveu.1
$neveu.1 (besoin urgent:+L) Héritage
$evt_02 (menace:+L) $neveu.1

// Chaîne causale: sabotage
$evt_02b (permet:+L) Intrusion non détectée
Intrusion non détectée (permet:+L) Crime

-:: _timeline_, _sequence_ ::

// =============================================================
// SECTION HYPOTHÈSES - Pistes d'enquête
// =============================================================

:: hypothèses, pistes ::

@hyp_m_01 Crime d'intérêt - Héritage Jean (type) hypothèse
    " (statut) en_attente
    " (confiance) 75%
    " (source) user
    " (description) Jean Moreau aurait agi pour sécuriser son héritage avant la modification du testament. Ses dettes de jeu (150 000€) créent une urgence financière. Il connaît le code d'accès du manoir et sa pointure (42) correspond aux traces de boue.
    " (mobile) Héritage 8 millions + dettes de jeu 150 000€
    " (pour) $brouillon.1, $traces.1, $telephone.1
    " (contre) Alibi partiel cinéma UGC Bercy
    " (questions) Alibi vérifié par caméras?; Accès au poison?; Complice possible?
    " (suspect) $neveu.1

@hyp_m_02 Crime passionnel - Vengeance Élodie (type) hypothèse
    " (statut) en_attente
    " (confiance) 65%
    " (source) user
    " (description) Élodie Dubois aurait empoisonné Victor pour se venger de la perte de 2 millions d'euros. Elle connaissait les habitudes de la victime et avait accès à la cuisine lors de visites antérieures. Sa présence près du manoir le jour du crime est troublante.
    " (mobile) Vengeance - perte 2 millions EUR
    " (pour) $telephone.1 (SMS menaces), Connaissance des habitudes
    " (contre) Alibi dîner charité Hôtel Crillon 19h-23h
    " (questions) Témoins au Crillon?; Connaissance des poisons?; Présence près du manoir confirmée?
    " (suspect) $exassociee.1

@hyp_m_03 Complot commercial - Piste Mercier (type) hypothèse
    " (statut) en_attente
    " (confiance) 35%
    " (source) ai
    " (description) Antoine Mercier, concurrent de Victor et rival au Cercle des Bibliophiles, aurait pu commanditer le crime pour éliminer un concurrent et récupérer sa clientèle. La rivalité lors de la vente aux enchères de juillet était intense.
    " (mobile) Rivalité commerciale, éliminer concurrent
    " (pour) Conflit vente aux enchères juillet 2025, Rivalité intense
    " (contre) Pas de preuves directes, Mobile insuffisant seul
    " (questions) Alibi vérifié?; Mobile suffisant pour meurtre?; Contacts avec tueur à gages?
    " (suspect) $concurrent.1

@hyp_m_04 Suicide maquillé (type) hypothèse
    " (statut) partielle
    " (confiance) 25%
    " (source) user
    " (description) Victor aurait orchestré sa propre mort pour incriminer ses héritiers. Les recherches sur les poisons (historique navigateur) et les indices trop évidents (livre ouvert page des alcaloïdes) suggèrent une mise en scène.
    " (pour) Recherches "poisons" sur téléphone, Livre ouvert page alcaloïdes, Indices trop évidents
    " (contre) Pas de lettre d'adieu, Circonstances suspectes, Aucun motif apparent
    " (questions) État psychologique récent?; Raison de vouloir incriminer héritiers?; Assurance vie récente?

@hyp_m_05 Complice interne - Madame Chen (type) hypothèse
    " (statut) en_attente
    " (confiance) 20%
    " (source) ai
    " (description) La gouvernante avait accès total à la cuisine et connaissait les habitudes de Victor. Son comportement 'étrangement calme' après la découverte du corps et le dîner privé avec Victor le 26/08 soulèvent des questions.
    " (mobile) Inconnu - possible complice rémunérée
    " (pour) Accès total cuisine, Comportement calme après découverte, Dîner privé 26/08
    " (contre) Loyauté 15 ans, Pas de mobile apparent
    " (questions) Identifier l'appelant mystérieux; Analyse empreintes non identifiées; Mouvements bancaires suspects?
    " (suspect) $gouvernante.1

// =============================================================
// RÉSEAU DE RELATIONS - Graphe sémantique
// =============================================================

:: réseau de relations ::

# Légende STTypes: N=proximité, +L=causalité, +C=containment, +E=expression

// Relations familiales et financières
$victime.1 (oncle de:N) $neveu.1
$victime.1 (transfère de l'argent à:+L) $neveu.1
$neveu.1 (doit de l'argent à:-C) $casino.1
$casino.1 (menace pour recouvrement:+L) $neveu.1

// Relations conflictuelles
$victime.1 (en conflit avec:N) $exassociee.1
$exassociee.1 (a menacé:+L) $victime.1
$exassociee.1 (accuse de fraude:+E) $victime.1

// Relations professionnelles
$victime.1 (emploie:-C) $gouvernante.1
$victime.1 (emploie:-C) $jardinier.1
$victime.1 (en rivalité avec:N) $concurrent.1

// Relations de propriété
$victime.1 (propriétaire de:+C) $bibliotheque.1
$victime.1 (propriétaire de:+C) Galerie Moreau Antiquités
$victime.1 (membre de:-C) Cercle des Bibliophiles Parisiens

// Chaîne causale du crime (hypothèse principale)
// $neveu.1 (a tué:+L) $victime.1

// =============================================================
// CHAÎNES CAUSALES DÉTECTÉES
// =============================================================

:: chaînes causales ::

+:: _sequence_ ::

# Chaîne 1: Mobile financier
@chain_mobile Dettes de jeu (mène à) Besoin d'argent (mène à) Motivation héritier (mène à) Mobile du crime

# Chaîne 2: Opportunité
@chain_opp Connaissance du code d'accès (mène à) Accès au manoir (mène à) Accès à la victime (mène à) Opportunité

# Chaîne 3: Chronologie du crime
@chain_crime Préparation poison (puis) Entrée par jardin (puis) Ajout dans thé (puis) Empoisonnement (puis) Fuite

-:: _sequence_ ::

// =============================================================
// RÉFÉRENCES CROISÉES - Pour utilisation $alias.n
// =============================================================

:: références croisées ::

# Alias pour références rapides
victimes => {Victor Moreau}
suspects => {Jean Moreau, Élodie Dubois, Antoine Mercier}
temoins => {Madame Chen, Robert Duval, Dr. Sarah Martin}
lieux => {Bibliothèque du Manoir, Jardin du Manoir, Casino de Deauville, UGC Bercy, Hôtel de Crillon}
preuves => {Tasse de thé, Livre des poisons, Traces de boue, Téléphone Victor, Testament actuel, Brouillon testament}

# Exemples d'utilisation:
# $victimes.1 = Victor Moreau
# $suspects.1 = Jean Moreau (suspect principal)
# $suspects.2 = Élodie Dubois
# $preuves.3 = Traces de boue

// =============================================================
// NOTES D'ENQUÊTE - TODO items
// =============================================================

:: notes, TODO ::

VÉRIFIER ALIBI JEAN MOREAU AU CINÉMA UGC
IDENTIFIER APPELANT INCONNU 19H05
ANALYSE TOXICOLOGIQUE COMPLÈTE DU THÉ
RECHERCHER ACHATS DE POISON RÉCENTS
INTERROGER LES MEMBRES DU CERCLE DES BIBLIOPHILES
VÉRIFIER LES CAMÉRAS DE SURVEILLANCE DU QUARTIER
`
}

// GetDisparitionN4LContent retourne le contenu N4L pour l'affaire de disparition
func GetDisparitionN4LContent() string {
	return `-affaire/disparition-001

# Affaire: Disparition suspecte - Marie Lefèvre
# Type: Personne disparue
# Statut: En cours

:: victimes ::

@disparue Marie Lefèvre (type) personne
    " (rôle) victime - disparue
    " (âge) 28 ans
    " (profession) Ingénieure aérospatiale
    " (employeur) Aérospace Industries
    " (dernier_contact) 15/08/2025 22h30
    " (lieu_disparition) Domicile Toulouse
    " (téléphone) Retrouvé chez elle
    " (véhicule) Voiture garée devant domicile

:: suspects ::

@ex Julien Martin (type) personne
    " (rôle) suspect
    " (lien) Ex-compagnon depuis 3 mois
    " (comportement) Messages insistants
    " (alibi) Prétend être à Lyon

:: chronologie ::

+:: _timeline_ ::

15/08/2025 18h00 $disparue.1 quitte le travail
15/08/2025 19h30 Dernier achat carte bancaire - supermarché
15/08/2025 22h30 Dernier message WhatsApp à une amie
16/08/2025 08h00 Non-présentation au travail
16/08/2025 20h00 Signalement disparition

-:: _timeline_ ::

:: hypothèses ::

@hyp_d_01 Enlèvement par ex-compagnon (confiance) 55%
@hyp_d_02 Départ volontaire (confiance) 25%
@hyp_d_03 Accident non découvert (confiance) 20%
`
}
