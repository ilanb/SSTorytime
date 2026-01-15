package data

// N4L Content pour l'affaire Incendie Entrepôt Logistique Nord
// Ce fichier contient le contenu N4L authentique utilisant toutes les fonctionnalités du langage SSTorytime

// GetIncendieN4LContent retourne le contenu N4L complet pour l'Affaire Incendie
func GetIncendieN4LContent() string {
	return `-affaire/incendie-005

# Affaire: Incendie Entrepôt Logistique Nord
# Type: Incendie criminel - Fraude à l'assurance
# Investigation: Brigade des Incendies Criminels
# Syntaxe: SSTorytime N4L v2

// =============================================================
// SECTION LIEUX - Scène de crime
// =============================================================

:: lieux, scène de crime ::

@entrepot Entrepôt Logistique Nord (type) lieu
    " (description) Entrepôt de stockage de 5000m² détruit à 80% par l'incendie. Valeur assurée: 4.5 millions €.
    " (adresse) Zone Industrielle Nord, Villeneuve-d'Ascq
    " (superficie) 5000 m²
    " (destruction) 80%
    " (valeur_assuree) 4.5 millions €
    " (construction) 1995
    " (latitude) 50.6292
    " (longitude) 3.1746

@zone_depart Zone de départ du feu (type) lieu
    " (description) Zone identifiée comme origine de l'incendie - traces d'accélérant
    " (localisation) Secteur nord-est de l'entrepôt
    " (indices) Résidus d'essence en 3 points

// Relations des lieux
Zone de départ du feu (située dans:+C) Entrepôt Logistique Nord

// =============================================================
// SECTION SUSPECTS
// =============================================================

:: suspects ::

@proprietaire André Petit (type) personne
    " (rôle) suspect principal
    " (description) Propriétaire de l'entrepôt via sa société LogiNord SARL. 56 ans. Difficultés financières importantes. A augmenté son assurance 3 mois avant l'incendie.
    " (âge) 56 ans
    " (profession) Chef d'entreprise
    " (societe) LogiNord SARL
    " (situation_financiere) Dettes importantes - 380 000€
    " (assurance) Augmentée en juillet 2025 (de 2M à 4.5M)
    " (mobile) Toucher l'indemnisation pour rembourser ses dettes
    " (alibi) Prétend être chez lui la nuit de l'incendie
    " (latitude) 50.6200
    " (longitude) 3.1500

// Relations d'André Petit
André Petit (propriétaire de:+C) Entrepôt Logistique Nord
André Petit (assuré par:N) Assurance MutualPro
André Petit (endetté envers:-C) Banque Populaire Nord
André Petit (aurait commandité:+L) Incendie

@executant Individu non identifié (type) personne
    " (rôle) suspect
    " (description) Personne ayant potentiellement mis le feu sur commande. Silhouette captée par caméra voisine à 3h10.
    " (description_physique) Homme, 1m75-1m80, vêtu de sombre
    " (vehicule) Scooter sombre
    " (identification) En cours

@employe Michel Garnier (type) personne
    " (rôle) suspect
    " (description) Ancien employé de LogiNord, licencié pour vol il y a 4 mois. Pourrait avoir agi par vengeance ou été recruté par Petit.
    " (âge) 38 ans
    " (profession) Ancien manutentionnaire
    " (licenciement) Mai 2025 - vol de marchandises
    " (mobile) Vengeance
    " (connaissance) Accès et agencement de l'entrepôt

// Relations de l'employé
Michel Garnier (ancien employé de:N) Entrepôt Logistique Nord
Michel Garnier (licencié par:N) André Petit

// =============================================================
// SECTION TÉMOINS
// =============================================================

:: témoins ::

@expert Expert Assurance Durand (type) personne
    " (rôle) temoin expert
    " (description) Expert mandaté par MutualPro. A conclu à une origine criminelle de l'incendie.
    " (profession) Expert en sinistres
    " (specialite) Incendies industriels
    " (conclusion) Origine criminelle - 3 foyers distincts
    " (rapport) Accélérant identifié (essence)

// Relations de l'expert
Expert Assurance Durand (a expertisé:+L) Entrepôt Logistique Nord
Expert Assurance Durand (mandaté par:N) Assurance MutualPro
Expert Assurance Durand (conclut contre:+L) André Petit

@pompier Capitaine Leroy (type) personne
    " (rôle) temoin
    " (description) Capitaine des pompiers ayant dirigé l'intervention. A noté des anomalies dans la propagation du feu.
    " (profession) Capitaine de sapeurs-pompiers
    " (observation) Propagation anormalement rapide
    " (rapport) Trois foyers distincts - origine criminelle probable

@voisin Entreprise Transports Martin (type) organisation
    " (rôle) temoin
    " (description) Entreprise voisine. Leur caméra de surveillance a capté une silhouette suspecte à 3h10.
    " (localisation) 100m de l'entrepôt
    " (video) Silhouette en scooter à 3h10

// =============================================================
// SECTION ORGANISATIONS
// =============================================================

:: organisations ::

@assurance Assurance MutualPro (type) organisation
    " (description) Compagnie d'assurance ayant couvert l'entrepôt. Suspecte une fraude et refuse de payer.
    " (type) Assurance professionnelle
    " (police) 4.5 millions €
    " (position) Refuse indemnisation - fraude présumée

// Relations de l'assurance
Assurance MutualPro (suspecte fraude de:+L) André Petit
Assurance MutualPro (a mandaté:N) Expert Assurance Durand

@banque Banque Populaire Nord (type) organisation
    " (description) Créancier principal d'André Petit. Menace de saisie immobilière.
    " (creance) 380 000 €
    " (action) Mise en demeure envoyée

// Relations de la banque
Banque Populaire Nord (créancier de:-C) André Petit
Banque Populaire Nord (menace de saisie:+L) André Petit

@societe LogiNord SARL (type) organisation
    " (description) Société d'André Petit exploitant l'entrepôt. En difficulté financière.
    " (gerant) André Petit
    " (capital) 100 000 €
    " (chiffre_affaires) En baisse - 2.1M en 2024 vs 3.5M en 2023

// =============================================================
// SECTION PREUVES - Indices matériels et numériques
// =============================================================

:: preuves, indices ::

// Groupement des preuves par type
Preuves forensiques => {Traces d'accélérant, Trois foyers distincts}
Preuves documentaires => {Augmentation assurance, Relevés bancaires, Inventaire stock}
Preuves numériques => {Vidéosurveillance voisine}

@accelerant Traces d'accélérant (type) preuve forensique
    " (localisation) Zone de départ du feu - 3 points distincts
    " (description) Résidus d'essence retrouvés en 3 points distincts de l'entrepôt
    " (analyse) Essence sans plomb 95
    " (conclusion) Origine criminelle confirmée
    " (fiabilité) 10/10

@foyers Trois foyers distincts (type) preuve forensique
    " (localisation) Secteurs nord-est, centre, sud-ouest
    " (description) L'incendie a démarré en 3 endroits simultanément - exclut accident
    " (conclusion) Mise à feu volontaire coordonnée
    " (fiabilité) 10/10

@police_assurance Augmentation assurance (type) preuve documentaire
    " (date) 01/07/2025
    " (description) Police d'assurance modifiée - couverture doublée
    " (avant) 2 millions €
    " (apres) 4.5 millions €
    " (delai) 3 mois avant l'incendie
    " (fiabilité) 9/10
    " (concerne) André Petit, Assurance MutualPro

@releves_bancaires Relevés bancaires Petit (type) preuve documentaire
    " (source) Banque Populaire Nord
    " (description) Compte à découvert important - rejets de prélèvements
    " (decouvert) 380 000 €
    " (rejets) Plusieurs prélèvements impayés
    " (fiabilité) 9/10
    " (concerne) André Petit

@inventaire Inventaire stock (type) preuve documentaire
    " (date) 15/09/2025
    " (description) Dernier inventaire avant incendie - stock déclaré suspect
    " (valeur_declaree) 1.2 million €
    " (suspicion) Surévaluation probable
    " (fiabilité) 7/10

@video_voisin Vidéosurveillance voisine (type) preuve numérique
    " (source) Transports Martin
    " (description) Caméra ayant capté une silhouette suspecte
    " (heure) 3h10 le 28/09/2025
    " (contenu) Homme en scooter, direction entrepôt
    " (fiabilité) 6/10
    " (concerne) Individu non identifié

// =============================================================
// SECTION CHRONOLOGIE - Séquence temporelle complète
// =============================================================

:: chronologie ::

+:: _timeline_, _sequence_ ::

// ==========================================
// Contexte financier (2024-2025)
// ==========================================

@evt_i_00 01/01/2024 09:00 Début difficultés financières LogiNord (lieu) Villeneuve-d'Ascq
    " (description) Baisse du chiffre d'affaires - perte de clients majeurs
    " (importance) medium
    " (vérifié) oui
    " (implique) André Petit, LogiNord SARL

@evt_i_00b 15/03/2025 09:00 Mise en demeure de la banque (lieu) Banque Populaire
    " (description) Banque Populaire Nord envoie une mise en demeure pour impayés
    " (importance) high
    " (vérifié) oui
    " (implique) André Petit, Banque Populaire Nord

@evt_i_00c 01/05/2025 09:00 Licenciement Michel Garnier (lieu) Entrepôt
    " (description) Michel Garnier licencié pour vol de marchandises
    " (importance) medium
    " (vérifié) oui
    " (implique) Michel Garnier, André Petit

@evt_i_01 01/07/2025 09:00 Augmentation assurance (lieu) MutualPro
    " (description) André Petit fait doubler la couverture d'assurance
    " (importance) high
    " (vérifié) oui
    " (implique) André Petit, Assurance MutualPro
    " (preuve) Augmentation assurance

@evt_i_02 15/09/2025 09:00 Dernier inventaire (lieu) Entrepôt
    " (description) Inventaire déclarant un stock de 1.2 million € - surévaluation suspectée
    " (importance) medium
    " (vérifié) oui
    " (implique) André Petit
    " (preuve) Inventaire stock

// ==========================================
// Nuit de l'incendie (28 septembre 2025)
// ==========================================

@evt_i_03 28/09/2025 03:10 Silhouette suspecte captée (lieu) Zone industrielle
    " (description) Caméra voisine capte un individu en scooter
    " (importance) high
    " (vérifié) oui
    " (implique) Individu non identifié
    " (preuve) Vidéosurveillance voisine

@evt_i_04 28/09/2025 03:15 Départ feu (lieu) Entrepôt
    " (description) Incendie déclenché en 3 points simultanés
    " (importance) high
    " (vérifié) oui
    " (preuve) Traces d'accélérant, Trois foyers distincts

@evt_i_05 28/09/2025 03:25 Détection incendie (lieu) Zone industrielle
    " (description) Alarme incendie déclenchée - voisins alertent les secours
    " (importance) medium
    " (vérifié) oui

@evt_i_06 28/09/2025 03:35 Arrivée pompiers (lieu) Entrepôt
    " (description) Premier véhicule de pompiers sur place - feu déjà important
    " (importance) medium
    " (vérifié) oui
    " (implique) Capitaine Leroy

@evt_i_07 28/09/2025 07:00 Feu maîtrisé (lieu) Entrepôt
    " (description) Incendie sous contrôle après 3h30 d'intervention
    " (importance) medium
    " (vérifié) oui
    " (implique) Capitaine Leroy

// ==========================================
// Après incendie
// ==========================================

@evt_i_08 29/09/2025 09:00 Début expertise (lieu) Entrepôt
    " (description) Expert Assurance Durand commence son investigation
    " (importance) high
    " (vérifié) oui
    " (implique) Expert Assurance Durand

@evt_i_09 01/10/2025 09:00 Conclusion origine criminelle (lieu) Entrepôt
    " (description) Expert conclut à un incendie volontaire - 3 foyers, accélérant
    " (importance) high
    " (vérifié) oui
    " (implique) Expert Assurance Durand
    " (preuve) Traces d'accélérant

@evt_i_10 05/10/2025 09:00 Refus indemnisation MutualPro (lieu) MutualPro
    " (description) Assurance MutualPro refuse de payer - suspicion de fraude
    " (importance) high
    " (vérifié) oui
    " (implique) Assurance MutualPro, André Petit

// ==========================================
// Chaînes causales
// ==========================================

// Chaîne causale: mobile financier
$evt_i_00 (mène à:+L) $evt_i_00b
$evt_i_00b (mène à:+L) Besoin urgent d'argent
Besoin urgent d'argent (mène à:+L) $evt_i_01

// Chaîne causale: exécution
$evt_i_01 (prépare:+L) $evt_i_04
$evt_i_04 (mène à:+L) Destruction entrepôt
Destruction entrepôt (permet:+L) Demande indemnisation

-:: _timeline_, _sequence_ ::

// =============================================================
// SECTION HYPOTHÈSES - Pistes d'enquête
// =============================================================

:: hypothèses, pistes ::

@hyp_i_01 Fraude à l'assurance - Petit commanditaire (type) hypothèse
    " (statut) en_attente
    " (confiance) 80%
    " (source) user
    " (description) André Petit aurait commandité l'incendie pour toucher l'indemnisation de 4.5 millions € et rembourser ses dettes de 380 000€. L'augmentation de l'assurance 3 mois avant est très suspecte.
    " (mobile) Rembourser dettes 380 000€ + profit
    " (pour) Augmentation assurance, Relevés bancaires Petit, Traces d'accélérant
    " (contre) Pas de preuves directes de commandite
    " (questions) Qui a exécuté l'incendie?; Liens avec l'ancien employé Garnier?; Communications suspectes?
    " (suspect) André Petit

@hyp_i_02 Exécution par ancien employé (type) hypothèse
    " (statut) en_attente
    " (confiance) 55%
    " (source) ai
    " (description) Michel Garnier, licencié pour vol, pourrait avoir été recruté par Petit pour mettre le feu, ou avoir agi seul par vengeance.
    " (pour) Connaissance des lieux, Mobile de vengeance, Licenciement récent
    " (contre) Pas de preuves de contact avec Petit
    " (questions) Où était Garnier la nuit de l'incendie?; Contacts avec Petit après licenciement?
    " (suspect) Michel Garnier

@hyp_i_03 Stock surévalué (type) hypothèse
    " (statut) en_attente
    " (confiance) 70%
    " (source) user
    " (description) En plus de l'incendie volontaire, le stock détruit aurait été volontairement surévalué dans l'inventaire pour maximiser l'indemnisation.
    " (pour) Inventaire stock, Difficultés financières
    " (contre) Difficile à prouver après destruction
    " (questions) Factures d'achat du stock?; Clients pouvant témoigner du stock réel?

// =============================================================
// RÉSEAU DE RELATIONS - Graphe sémantique
// =============================================================

:: réseau de relations ::

# Légende STTypes: N=proximité, +L=causalité, +C=containment, +E=expression

// Relations de propriété et finance
André Petit (propriétaire de:+C) Entrepôt Logistique Nord
André Petit (gérant de:+C) LogiNord SARL
André Petit (endetté envers:-C) Banque Populaire Nord
André Petit (assuré par:N) Assurance MutualPro

// Relations suspectes
André Petit (aurait commandité:+L) Incendie
Michel Garnier (aurait exécuté:+L) Incendie
Assurance MutualPro (suspecte fraude de:+L) André Petit

// Relations professionnelles passées
Michel Garnier (licencié par:N) André Petit
Michel Garnier (ancien employé de:N) Entrepôt Logistique Nord

// =============================================================
// CHAÎNES CAUSALES DÉTECTÉES
// =============================================================

:: chaînes causales ::

+:: _sequence_ ::

# Chaîne 1: Mobile financier
@chain_mobile Difficultés financières (mène à) Mise en demeure banque (mène à) Besoin urgent argent (mène à) Augmentation assurance (mène à) Plan incendie

# Chaîne 2: Exécution
@chain_exec Recrutement exécutant (puis) Acquisition accélérant (puis) Mise à feu 3 points (puis) Incendie (puis) Destruction

# Chaîne 3: Objectif
@chain_obj Destruction entrepôt (permet) Déclaration sinistre (permet) Demande indemnisation 4.5M€ (permet) Remboursement dettes

-:: _sequence_ ::

// =============================================================
// RÉFÉRENCES CROISÉES - Pour utilisation $alias.n
// =============================================================

:: références croisées ::

# Alias pour références rapides
suspects => {André Petit, Michel Garnier, Individu non identifié}
temoins => {Expert Assurance Durand, Capitaine Leroy, Transports Martin}
lieux => {Entrepôt Logistique Nord, Zone de départ du feu}
organisations => {MutualPro, Banque Populaire Nord, LogiNord SARL}
preuves => {Traces d'accélérant, Trois foyers distincts, Augmentation assurance, Relevés bancaires}

// =============================================================
// NOTES D'ENQUÊTE - TODO items
// =============================================================

:: notes, TODO ::

ANALYSER TÉLÉPHONIE D'ANDRÉ PETIT (CONTACTS SUSPECTS)
INTERROGER MICHEL GARNIER - ALIBI NUIT INCENDIE
IDENTIFIER L'INDIVIDU AU SCOOTER
VÉRIFIER ACHATS D'ESSENCE RÉCENTS DANS LA ZONE
RECONSTITUER LE STOCK RÉEL VIA FACTURES FOURNISSEURS
ANALYSER COMPTES BANCAIRES PETIT - VERSEMENTS SUSPECTS
RECHERCHER D'AUTRES POLICES D'ASSURANCE AU NOM DE PETIT
`
}
