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
    " (dernier_contact) Appel à Victor Moreau le 29/08 à 18h45
    " (latitude) 48.8396
    " (longitude) 2.3876

// Relations de Jean Moreau
Jean Moreau (hérite de:+L) Victor Moreau
Jean Moreau (doit de l'argent à:-C) Casino de Deauville
Jean Moreau (connaît le code d'accès:N) Bibliothèque du Manoir
Jean Moreau (vu au:N) Bar Le Diplomate
Jean Moreau (alibi à:N) UGC Bercy

@exassociee Élodie Dubois (type) personne
    " (rôle) suspect
    " (description) Ex-associée de Victor, 42 ans. Perte financière de 2 millions en 2024. Procédure judiciaire en cours. Menaces proférées: 'Il paiera pour ce qu'il m'a fait'.
    " (âge) 42 ans
    " (profession) Femme d'affaires
    " (mobile) Vengeance - perte 2 millions EUR
    " (procédure) Procès civil en cours contre Victor Moreau
    " (menaces) Il paiera pour ce qu'il m'a fait
    " (alibi) Dîner charité Hôtel Crillon 19h-23h
    " (avocat) Maître Lefebvre
    " (latitude) 48.8686
    " (longitude) 2.3215

// Relations d'Élodie Dubois
Élodie Dubois (en conflit avec:N) Victor Moreau
Élodie Dubois (a menacé:+L) Victor Moreau
Élodie Dubois (représentée par:N) Maître Lefebvre
Élodie Dubois (alibi à:N) Hôtel de Crillon

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
Antoine Mercier (rival de:N) Victor Moreau
Antoine Mercier (a perdu aux enchères contre:N) Victor Moreau

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
Madame Chen (employée par:N) Victor Moreau

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
    " (client) Élodie Dubois
    " (présent) Confrontation tribunal 25/08

// Relations Maître Lefebvre
Maître Lefebvre (représente:N) Élodie Dubois
Maître Lefebvre (a attaqué:N) Victor Moreau

@notaire Maître Durand (type) personne
    " (rôle) temoin
    " (profession) Notaire
    " (cabinet) Étude notariale Paris 7e
    " (action) Rencontre avec Victor Moreau le 28/08

// Relations Maître Durand
Maître Durand (a reçu:N) Victor Moreau
Maître Durand (a rédigé:N) Testament actuel

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
Homme au téléphone inconnu (a appelé:+L) Victor Moreau

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
Bibliothèque du Manoir (propriété de:N) Victor Moreau

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
Jardin du Manoir (donne accès à:+C) Bibliothèque du Manoir

@portillon Portillon du jardin (type) lieu
    " (type) Accès secondaire
    " (sécurité) Fermé à clé
    " (clé) Détenue par Robert Duval
    " (état_29_08) Fermé à 18h par Robert Duval

// Relations Portillon
Portillon du jardin (donne accès à:+C) Jardin du Manoir

@casino Casino de Deauville (type) lieu
    " (description) Casino où Jean Moreau a contracté des dettes importantes.
    " (adresse) 2 Rue Edmond Blanc, 14800 Deauville
    " (type_lieu) Casino
    " (lien) Créancier de Jean Moreau - 150k EUR
    " (latitude) 49.3565
    " (longitude) -0.0742

@ugc UGC Bercy (type) lieu
    " (description) Cinéma où Jean Moreau prétend avoir passé la soirée du crime (19h-22h).
    " (adresse) 2 Cour Saint-Émilion, Paris 12e
    " (type_lieu) Cinéma
    " (alibi) Lieu déclaré par Jean Moreau le soir du crime
    " (latitude) 48.8335
    " (longitude) 2.3867

// Relations UGC
UGC Bercy (alibi de:N) Jean Moreau

@crillon Hôtel de Crillon (type) lieu
    " (description) Palace parisien où Élodie Dubois assistait à un dîner de charité le soir du crime.
    " (adresse) 10 Place de la Concorde, Paris 8e
    " (type_lieu) Hôtel de luxe
    " (alibi) Lieu déclaré par Élodie Dubois le soir du crime
    " (latitude) 48.8677
    " (longitude) 2.3216

// Relations Crillon
Hôtel de Crillon (alibi de:N) Élodie Dubois

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
Galerie Moreau Antiquités (propriété de:N) Victor Moreau

@cercle Cercle des Bibliophiles Parisiens (type) organisation
    " (description) Club exclusif de collectionneurs de livres anciens. Victor en était membre depuis 20 ans.

// Relations Cercle
Cercle des Bibliophiles Parisiens (membre:N) Victor Moreau

// =============================================================
// SECTION PREUVES - Indices matériels et numériques
// =============================================================

:: preuves, indices ::

// Groupement des preuves par type
Preuves physiques => {Tasse de thé, Livre des poisons, Traces de boue}
Preuves numériques => {Téléphone Victor, Vidéosurveillance UGC}
Preuves documentaires => {Testament actuel, Brouillon testament}

@tasse Tasse de thé (type) preuve physique
    " (localisation) Table basse Bibliothèque du Manoir
    " (contenu) Résidus Earl Grey + substance inconnue
    " (empreintes) Victor uniquement
    " (fiabilité) 9/10
    " (concerne) Victor Moreau

@livre Livre "Traité des Poisons Exotiques" (type) preuve physique
    " (édition) 1923
    " (localisation) Bureau de Victor Moreau
    " (page) 247 - Alcaloïdes végétaux
    " (annotations) *Récentes - encre fraîche
    " (empreintes) Partielles non identifiées
    " (fiabilité) 7/10

@traces Traces de boue (type) preuve forensique
    " (localisation) Tapis persan Bibliothèque du Manoir
    " (composition) Terre argileuse des rosiers
    " (pointure) *42 - correspond à Jean Moreau
    " (direction) Fenêtre vers fauteuil
    " (fraîcheur) < 24h
    " (fiabilité) 8/10

@telephone Téléphone de Victor (type) preuve numérique
    " (localisation) Sur la victime
    " (appels) Jean Moreau 18h45, inconnu 19h05
    " (sms) Menaces de Élodie Dubois
    " (recherches) "poisons" le 28/08
    " (fiabilité) 9/10
    " (concerne) Victor Moreau, Jean Moreau, Élodie Dubois

@testament Testament actuel (type) preuve documentaire
    " (date) 15/03/2025
    " (notaire) Maître Durand
    " (bénéficiaire) Jean Moreau 80%
    " (statut) Valide et officiel

@brouillon Brouillon nouveau testament (type) preuve documentaire
    " (date) 28/08/2025
    " (statut) *Non signé - trouvé corbeille
    " (nouveau_bénéficiaire) Fondation Moreau 90%
    " (impact) Réduit Jean Moreau à 10%
    " (mobile) =déshéritage Renforce le mobile de Jean Moreau

@sms SMS menaçants (type) preuve numérique
    " (source) Téléphone de Victor Moreau
    " (expéditeur) Élodie Dubois
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
    " (médecin) Dr. Sarah Martin
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
    " (description) Victor Moreau remporte un manuscrit convoité par Antoine Mercier pour 450 000€
    " (importance) medium
    " (vérifié) oui
    " (implique) Victor Moreau, Antoine Mercier

// Événements du 25 août 2025
@evt_00b 25/08/2025 10:00 Confrontation tribunal Victor Moreau/Élodie Dubois (lieu) Tribunal de Commerce
    " (description) Élodie Dubois menace Victor Moreau: 'Il paiera pour ce qu'il m'a fait'
    " (importance) high
    " (vérifié) oui
    " (implique) Victor Moreau, Élodie Dubois, Maître Lefebvre
    " (preuve) SMS menaçants

// ==========================================
// Semaine du crime (27-28 août 2025)
// ==========================================

// Événements du 27 août 2025
@evt_01 27/08/2025 14:00 Visite houleuse de Jean Moreau (lieu) Manoir
    " (description) Discussion avec Victor Moreau sur l'argent. Claque la porte en partant.
    " (importance) high
    " (vérifié) oui
    " (implique) Victor Moreau, Jean Moreau

@evt_01b 27/08/2025 22:00 Jean Moreau vu au Bar Le Diplomate (lieu) Bar Le Diplomate
    " (description) Rencontre avec individu non identifié
    " (importance) medium
    " (vérifié) non
    " (implique) Jean Moreau, Homme non identifié

// Événements du 28 août 2025
@evt_02 28/08/2025 10:00 Victor Moreau chez le notaire (lieu) Étude notariale
    " (description) Évoque une modification du testament - veut déshériter Jean Moreau
    " (importance) high
    " (vérifié) oui
    " (implique) Victor Moreau, Maître Durand
    " (preuve) Brouillon nouveau testament

@evt_02b 28/08/2025 18:00 Câble caméra sectionné (lieu) Manoir
    " (description) Système de surveillance neutralisé
    " (importance) high
    " (vérifié) oui
    " (preuve) Système vidéosurveillance

// ==========================================
// Jour du crime (29 août 2025)
// ==========================================

// Événements du 29 août 2025
@evt_03 29/08/2025 15:00 Élodie Dubois vue près du manoir (lieu) Aux abords du manoir
    " (description) Aperçue par voisin M. Bertrand à 15h
    " (importance) high
    " (vérifié) non
    " (implique) Élodie Dubois

@evt_04 29/08/2025 17:30 Fenêtre bibliothèque ouverte (lieu) Bibliothèque
    " (description) Observée par Robert Duval - traces de boue découvertes plus tard
    " (importance) medium
    " (vérifié) oui
    " (implique) Robert Duval, Bibliothèque du Manoir
    " (preuve) Traces de boue

@evt_05 29/08/2025 18:00 Départ du jardinier (lieu) Jardin
    " (description) Robert Duval ferme le portillon à clé
    " (importance) medium
    " (vérifié) oui
    " (implique) Robert Duval, Portillon du jardin

@evt_06 29/08/2025 18:45 Appel Jean Moreau vers Victor Moreau (lieu) Téléphone
    " (description) Durée 3 minutes - contenu inconnu
    " (importance) high
    " (vérifié) oui
    " (implique) Jean Moreau, Victor Moreau
    " (preuve) Téléphone de Victor

@evt_07 29/08/2025 19:00 Départ Madame Chen (lieu) Manoir
    " (description) Laisse Victor Moreau seul après avoir servi le thé
    " (importance) high
    " (vérifié) oui
    " (implique) Madame Chen, Victor Moreau
    " (preuve) Tasse de thé

@evt_08 29/08/2025 19:05 Appel numéro inconnu (lieu) Téléphone
    " (description) Appel entrant sur téléphone Victor Moreau - durée 2min
    " (importance) high
    " (vérifié) oui
    " (implique) Victor Moreau, Homme au téléphone inconnu
    " (preuve) Téléphone de Victor

@evt_09 29/08/2025 19:15 Victor Moreau boit son thé (lieu) Bibliothèque
    " (description) Dernière activité connue - thé possiblement empoisonné
    " (importance) high
    " (vérifié) non
    " (implique) Victor Moreau, Bibliothèque du Manoir
    " (preuve) Tasse de thé

@evt_09b 29/08/2025 19:25 Entrée cinéma Jean Moreau (alibi) (lieu) UGC Bercy
    " (description) Ticket acheté pour séance 19h30
    " (importance) high
    " (vérifié) oui
    " (implique) Jean Moreau

@evt_10 29/08/2025 20:30 Heure estimée du décès (lieu) Bibliothèque
    " (description) Selon rapport médecin légiste Dr. Martin
    " (importance) high
    " (vérifié) oui
    " (implique) Victor Moreau, Dr. Sarah Martin
    " (preuve) Tasse de thé, Rapport d'autopsie

@evt_11 29/08/2025 21:45 Découverte du corps (lieu) Bibliothèque
    " (description) Par Madame Chen revenue au manoir
    " (importance) high
    " (vérifié) oui
    " (implique) Victor Moreau, Madame Chen, Bibliothèque du Manoir
    " (preuve) Tasse de thé, Traces de boue

@evt_12 29/08/2025 22:00 Arrivée police (lieu) Manoir
    " (description) Début de l'enquête officielle
    " (importance) medium
    " (vérifié) oui

// ==========================================
// Chaînes causales
// ==========================================

// Chaîne causale principale: empoisonnement
Victor Moreau boit son thé (mène à:+L) Empoisonnement
Empoisonnement (mène à:+L) Heure estimée du décès
Heure estimée du décès (découvert par:+L) Découverte du corps

// Chaîne causale: mobile financier
Jean Moreau (dettes:+L) Casino de Deauville
Casino de Deauville (pression:+L) Jean Moreau
Jean Moreau (besoin urgent:+L) Héritage
Victor Moreau chez le notaire (menace:+L) Jean Moreau

// Chaîne causale: sabotage
Câble caméra sectionné (permet:+L) Intrusion non détectée
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
    " (pour) Brouillon nouveau testament, Traces de boue, Téléphone de Victor
    " (contre) Alibi partiel cinéma UGC Bercy
    " (questions) Alibi vérifié par caméras?; Accès au poison?; Complice possible?
    " (suspect) Jean Moreau

@hyp_m_02 Crime passionnel - Vengeance Élodie (type) hypothèse
    " (statut) en_attente
    " (confiance) 65%
    " (source) user
    " (description) Élodie Dubois aurait empoisonné Victor pour se venger de la perte de 2 millions d'euros. Elle connaissait les habitudes de la victime et avait accès à la cuisine lors de visites antérieures. Sa présence près du manoir le jour du crime est troublante.
    " (mobile) Vengeance - perte 2 millions EUR
    " (pour) Téléphone de Victor (SMS menaces), Connaissance des habitudes
    " (contre) Alibi dîner charité Hôtel Crillon 19h-23h
    " (questions) Témoins au Crillon?; Connaissance des poisons?; Présence près du manoir confirmée?
    " (suspect) Élodie Dubois

@hyp_m_03 Complot commercial - Piste Mercier (type) hypothèse
    " (statut) en_attente
    " (confiance) 35%
    " (source) ai
    " (description) Antoine Mercier, concurrent de Victor et rival au Cercle des Bibliophiles, aurait pu commanditer le crime pour éliminer un concurrent et récupérer sa clientèle. La rivalité lors de la vente aux enchères de juillet était intense.
    " (mobile) Rivalité commerciale, éliminer concurrent
    " (pour) Conflit vente aux enchères juillet 2025, Rivalité intense
    " (contre) Pas de preuves directes, Mobile insuffisant seul
    " (questions) Alibi vérifié?; Mobile suffisant pour meurtre?; Contacts avec tueur à gages?
    " (suspect) Antoine Mercier

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
    " (suspect) Madame Chen

// =============================================================
// RÉSEAU DE RELATIONS - Graphe sémantique
// =============================================================

:: réseau de relations ::

# Légende STTypes: N=proximité, +L=causalité, +C=containment, +E=expression

// Relations familiales et financières
Victor Moreau (oncle de:N) Jean Moreau
Victor Moreau (transfère de l'argent à:+L) Jean Moreau
Jean Moreau (doit de l'argent à:-C) Casino de Deauville
Casino de Deauville (menace pour recouvrement:+L) Jean Moreau

// Relations conflictuelles
Victor Moreau (en conflit avec:N) Élodie Dubois
Élodie Dubois (a menacé:+L) Victor Moreau
Élodie Dubois (accuse de fraude:+E) Victor Moreau

// Relations professionnelles
Victor Moreau (emploie:-C) Madame Chen
Victor Moreau (emploie:-C) Robert Duval
Victor Moreau (en rivalité avec:N) Antoine Mercier

// Relations de propriété
Victor Moreau (propriétaire de:+C) Bibliothèque du Manoir
Victor Moreau (propriétaire de:+C) Galerie Moreau Antiquités
Victor Moreau (membre de:-C) Cercle des Bibliophiles Parisiens

// Chaîne causale du crime (hypothèse principale)
// Jean Moreau (a tué:+L) Victor Moreau

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
// RÉFÉRENCES CROISÉES - Pour utilisation dans les recherches
// =============================================================

:: références croisées ::

# Alias pour références rapides
victimes => {Victor Moreau}
suspects => {Jean Moreau, Élodie Dubois, Antoine Mercier}
temoins => {Madame Chen, Robert Duval, Dr. Sarah Martin}
lieux => {Bibliothèque du Manoir, Jardin du Manoir, Casino de Deauville, UGC Bercy, Hôtel de Crillon}
preuves => {Tasse de thé, Livre des poisons, Traces de boue, Téléphone Victor, Testament actuel, Brouillon testament}

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

15/08/2025 18h00 Marie Lefèvre quitte le travail
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
