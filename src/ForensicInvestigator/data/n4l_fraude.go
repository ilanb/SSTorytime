package data

// N4L Content pour l'affaire Fraude Pyramidale FinanceMax
// Ce fichier contient le contenu N4L authentique utilisant toutes les fonctionnalités du langage SSTorytime

// GetFraudeN4LContent retourne le contenu N4L complet pour l'Affaire Fraude Pyramidale FinanceMax
func GetFraudeN4LContent() string {
	return `-affaire/fraude-003

# Affaire: Fraude Pyramidale FinanceMax
# Type: Escroquerie de type Ponzi
# Investigation: Brigade Financière Paris
# Syntaxe: SSTorytime N4L v2

// =============================================================
// SECTION SUSPECTS PRINCIPAUX
// =============================================================

:: suspects ::

@pdg Philippe Martin (type) personne
    " (rôle) suspect principal
    " (description) Fondateur et PDG de FinanceMax. 48 ans. Promettait des rendements de 15% par mois. Cerveau de l'escroquerie.
    " (âge) 48 ans
    " (profession) Financier
    " (societe) FinanceMax SARL
    " (prejudice) 23 millions €
    " (victimes) 847 personnes
    " (mode_operatoire) Schéma de Ponzi - paiements des anciens avec l'argent des nouveaux
    " (train_de_vie) Luxueux - villa Cannes, Ferrari, yacht
    " (latitude) 48.8698
    " (longitude) 2.3075

// Relations de Philippe Martin
$pdg.1 (supérieur hiérarchique de:+C) $directrice.1
$pdg.1 (a escroqué:+L) $victimes.1
$pdg.1 (a transféré vers:-C) $compte_suisse.1
$pdg.1 (propriétaire de:+C) $societe.1

@directrice Céline Roux (type) personne
    " (rôle) suspect
    " (description) Directrice commerciale de FinanceMax. 39 ans. Recrutait les investisseurs avec des promesses mensongères.
    " (âge) 39 ans
    " (profession) Directrice commerciale
    " (role) Recrutement investisseurs
    " (methode) Promesses de rendements garantis
    " (commission) 5% sur chaque investissement recruté

// Relations de Céline Roux
$directrice.1 (complice de:+L) $pdg.1
$directrice.1 (a recruté:+L) $victimes.1
$directrice.1 (subordonnée de:-C) $pdg.1

@comptable Marc Dubois (type) personne
    " (rôle) suspect
    " (description) Comptable de FinanceMax. Falsifiait les bilans et rapports financiers.
    " (âge) 45 ans
    " (profession) Expert-comptable
    " (role) Falsification des comptes
    " (methode) Faux bilans, rapports mensongers

// Relations du comptable
$comptable.1 (complice de:+L) $pdg.1
$comptable.1 (a falsifié:+L) $bilans.1

// =============================================================
// SECTION VICTIMES ET ORGANISATIONS
// =============================================================

:: victimes ::

@victimes Association Victimes FinanceMax (type) organisation
    " (rôle) partie civile
    " (description) Regroupement de 612 victimes pour action collective. Préjudice total de 23 millions d'euros.
    " (membres) 612 personnes
    " (prejudice_total) 23 millions €
    " (action) Plainte collective

// Relations de l'association
$victimes.1 (porte plainte contre:+L) $pdg.1
$victimes.1 (porte plainte contre:+L) $directrice.1

@retraite Marcel Lefebvre (type) personne
    " (rôle) victime
    " (description) Retraité de 72 ans. A investi toute son épargne retraite (180 000€) dans FinanceMax.
    " (âge) 72 ans
    " (profession) Retraité (ancien cadre)
    " (investissement) 180 000 €
    " (perte) 180 000 € - totalité
    " (statut) Représentant des victimes

// Relations de Marcel Lefebvre
$retraite.1 (a été escroqué par:+L) $pdg.1
$retraite.1 (recruté par:N) $directrice.1
$retraite.1 (président de:+C) $victimes.1

@medecin Dr. Catherine Moreau (type) personne
    " (rôle) victime
    " (description) Médecin libérale de 58 ans. A investi 320 000€ pour sa retraite.
    " (âge) 58 ans
    " (profession) Médecin libérale
    " (investissement) 320 000 €
    " (perte) 320 000 € - totalité

// =============================================================
// SECTION LIEUX ET ORGANISATIONS
// =============================================================

:: lieux, organisations ::

@societe FinanceMax SARL (type) organisation
    " (description) Société écran pour l'escroquerie. Siège à Paris 8e. Radiée depuis.
    " (adresse) 25 avenue Hoche, Paris 8e
    " (capital) 50 000 €
    " (creation) Mars 2022
    " (radiation) Octobre 2025
    " (latitude) 48.8756
    " (longitude) 2.3012

@bureau Bureaux FinanceMax (type) lieu
    " (description) Bureaux luxueux pour impressionner les clients. Loyer: 15 000€/mois.
    " (adresse) 25 avenue Hoche, Paris 8e
    " (superficie) 200 m²
    " (loyer) 15 000 €/mois
    " (decoration) Luxueuse - tableaux, mobilier design

@compte_suisse Compte HSBC Suisse (type) compte
    " (description) Compte bancaire offshore où transitaient les fonds détournés.
    " (banque) HSBC Private Bank Genève
    " (numero) CH** **** **** 4521
    " (solde_actuel) 4.2 millions € gelés
    " (flux) 18 millions € en 3 ans

@compte_panama Société écran Panama (type) organisation
    " (description) Société offshore pour blanchir les fonds.
    " (nom) Golden Investments Corp
    " (juridiction) Panama
    " (beneficiaire_reel) $pdg.1

// Relations des comptes
$compte_suisse.1 (alimenté par:+L) $societe.1
$compte_suisse.1 (transfère vers:+L) $compte_panama.1
$compte_panama.1 (bénéficiaire réel:N) $pdg.1

// =============================================================
// SECTION TÉMOINS
// =============================================================

:: témoins ::

@lanceur Antoine Mercier (type) personne
    " (rôle) temoin - lanceur d'alerte
    " (description) Ancien commercial de FinanceMax. A démissionné et alerté l'AMF après avoir découvert la fraude.
    " (âge) 35 ans
    " (profession) Commercial financier
    " (emploi_financemax) Janvier 2024 - Juin 2024
    " (decouverte) Schéma de Ponzi
    " (action) Alerte AMF juillet 2024

// Relations du lanceur d'alerte
$lanceur.1 (ancien employé de:N) $societe.1
$lanceur.1 (a dénoncé:+L) $pdg.1
$lanceur.1 (a alerté:+L) $amf.1

@amf AMF - Autorité des Marchés Financiers (type) organisation
    " (rôle) temoin - autorité
    " (description) Autorité de régulation ayant reçu l'alerte et transmis au parquet.
    " (action) Enquête préliminaire
    " (transmission) Parquet de Paris

// Relations AMF
$amf.1 (enquête sur:+L) $societe.1
$amf.1 (a saisi:+L) Parquet de Paris

@banquier Laurent Petit (type) personne
    " (rôle) temoin
    " (description) Ancien conseiller bancaire de FinanceMax. A remarqué des mouvements suspects.
    " (profession) Conseiller bancaire
    " (observations) Virements internationaux suspects
    " (signalement) TRACFIN

// =============================================================
// SECTION PREUVES - Indices matériels et numériques
// =============================================================

:: preuves, indices ::

// Groupement des preuves par type
Preuves documentaires => {Contrats d'investissement, Bilans falsifiés, Relevés bancaires}
Preuves numériques => {Emails internes, Base de données clients}
Preuves testimoniales => {Témoignage lanceur d'alerte}

@contrats Contrats d'investissement (type) preuve documentaire
    " (quantité) 847 contrats
    " (description) Contrats promettant 15% de rendement mensuel garanti
    " (clauses) Abusives - aucune mention des risques
    " (fiabilité) 10/10
    " (concerne) $pdg.1, $directrice.1, $victimes.1

@bilans Bilans falsifiés (type) preuve documentaire
    " (periode) 2022-2025
    " (description) Bilans comptables montrant des actifs fictifs
    " (actifs_declares) 45 millions €
    " (actifs_reels) < 2 millions €
    " (fiabilité) 10/10
    " (concerne) $pdg.1, $comptable.1

@emails Emails internes (type) preuve numérique
    " (source) Serveur FinanceMax
    " (description) Communications entre dirigeants sur la gestion du schéma
    " (extraits) Discussions sur 'tenir encore 6 mois'
    " (fiabilité) 9/10
    " (concerne) $pdg.1, $directrice.1

@releves Relevés bancaires (type) preuve documentaire
    " (comptes) FinanceMax, HSBC Suisse, Panama
    " (description) Traçabilité des flux financiers
    " (montant_trace) 18 millions €
    " (destination) Comptes offshore
    " (fiabilité) 10/10
    " (concerne) $pdg.1, $compte_suisse.1, $compte_panama.1

@temoignage Témoignage lanceur d'alerte (type) preuve testimoniale
    " (temoin) $lanceur.1
    " (description) Décrit le fonctionnement interne et le schéma de Ponzi
    " (elements) Recrutement, promesses, versements
    " (fiabilité) 9/10

@biens_saisis Biens saisis (type) preuve physique
    " (description) Villa Cannes, Ferrari 488, Yacht 15m, Montres de luxe
    " (valeur_totale) 3.5 millions €
    " (origine) Produit de l'escroquerie

// =============================================================
// SECTION CHRONOLOGIE - Séquence temporelle complète
// =============================================================

:: chronologie ::

+:: _timeline_, _sequence_ ::

// ==========================================
// Création et développement (2022-2023)
// ==========================================

@evt_f_01 15/03/2022 09:00 Création de FinanceMax SARL (lieu) Paris
    " (description) Immatriculation de la société au RCS Paris
    " (importance) high
    " (vérifié) oui
    " (implique) $pdg.1

@evt_f_02 01/06/2022 09:00 Premiers investisseurs (lieu) Bureaux FinanceMax
    " (description) Début du recrutement de clients. 50 premiers investisseurs.
    " (importance) high
    " (vérifié) oui
    " (implique) $pdg.1, $directrice.1
    " (montant) 500 000 €

@evt_f_03 15/12/2022 09:00 Premiers paiements de rendements (lieu) Paris
    " (description) Versement des premiers 'rendements' aux investisseurs initiaux
    " (importance) medium
    " (vérifié) oui
    " (implique) $societe.1, $victimes.1

// ==========================================
// Expansion (2023-2024)
// ==========================================

@evt_f_04 01/01/2023 09:00 Embauche de commerciaux (lieu) Bureaux FinanceMax
    " (description) Recrutement de 5 commerciaux pour accélérer le démarchage
    " (importance) medium
    " (vérifié) oui
    " (implique) $pdg.1, $directrice.1

@evt_f_05 01/07/2023 09:00 Cap des 500 clients (lieu) Paris
    " (description) FinanceMax compte désormais 500 investisseurs
    " (importance) medium
    " (vérifié) oui
    " (implique) $societe.1
    " (montant) 12 millions €

@evt_f_06 15/01/2024 09:00 Embauche Antoine Mercier (lieu) Bureaux FinanceMax
    " (description) $lanceur.1 rejoint l'équipe commerciale
    " (importance) medium
    " (vérifié) oui
    " (implique) $lanceur.1

@evt_f_07 15/06/2024 09:00 Démission Mercier (lieu) Bureaux FinanceMax
    " (description) $lanceur.1 découvre la fraude et démissionne
    " (importance) high
    " (vérifié) oui
    " (implique) $lanceur.1, $pdg.1

// ==========================================
// Alerte et effondrement (2024-2025)
// ==========================================

@evt_f_08 01/07/2024 09:00 Alerte AMF (lieu) AMF Paris
    " (description) $lanceur.1 alerte l'AMF sur la fraude
    " (importance) high
    " (vérifié) oui
    " (implique) $lanceur.1, $amf.1

@evt_f_09 15/09/2024 09:00 Enquête préliminaire (lieu) Paris
    " (description) L'AMF ouvre une enquête préliminaire
    " (importance) high
    " (vérifié) oui
    " (implique) $amf.1, $societe.1

@evt_f_10 01/03/2025 09:00 Premiers retards de paiement (lieu) Paris
    " (description) FinanceMax commence à retarder les versements de rendements
    " (importance) high
    " (vérifié) oui
    " (implique) $societe.1, $victimes.1

@evt_f_11 15/07/2025 09:00 Perquisitions (lieu) Bureaux FinanceMax et domiciles
    " (description) Perquisitions simultanées aux bureaux et chez les dirigeants
    " (importance) high
    " (vérifié) oui
    " (implique) $pdg.1, $directrice.1, $comptable.1

@evt_f_12 16/07/2025 09:00 Mises en examen (lieu) Tribunal de Paris
    " (description) Martin, Roux et Dubois mis en examen pour escroquerie en bande organisée
    " (importance) high
    " (vérifié) oui
    " (implique) $pdg.1, $directrice.1, $comptable.1

@evt_f_13 01/08/2025 09:00 Gel des avoirs (lieu) Suisse et Panama
    " (description) Gel des comptes offshore sur commission rogatoire internationale
    " (importance) high
    " (vérifié) oui
    " (implique) $compte_suisse.1, $compte_panama.1
    " (montant) 4.2 millions €

// ==========================================
// Chaînes causales
// ==========================================

// Chaîne causale principale: schéma de Ponzi
$evt_f_02 (alimente:+L) $evt_f_03
$evt_f_03 (attire:+L) Nouveaux investisseurs
Nouveaux investisseurs (alimente:+L) Anciens rendements

// Chaîne causale: effondrement
$evt_f_07 (mène à:+L) $evt_f_08
$evt_f_08 (mène à:+L) $evt_f_09
$evt_f_09 (mène à:+L) $evt_f_11

-:: _timeline_, _sequence_ ::

// =============================================================
// SECTION HYPOTHÈSES - Pistes d'enquête
// =============================================================

:: hypothèses, pistes ::

@hyp_f_01 Escroquerie en bande organisée (type) hypothèse
    " (statut) confirmée
    " (confiance) 95%
    " (source) user
    " (description) Philippe Martin a mis en place un schéma de Ponzi classique avec l'aide de Céline Roux et Marc Dubois. Les rendements versés aux anciens clients provenaient uniquement des apports des nouveaux.
    " (pour) $contrats.1, $bilans.1, $releves.1, $temoignage.1
    " (contre) Aucun - preuves accablantes
    " (suspect) $pdg.1, $directrice.1, $comptable.1

@hyp_f_02 Blanchiment international (type) hypothèse
    " (statut) en_attente
    " (confiance) 85%
    " (source) user
    " (description) Une partie des fonds aurait été blanchie via des sociétés offshore au Panama et en Suisse. D'autres complices pourraient être impliqués.
    " (pour) $releves.1, Société écran Panama
    " (contre) Traçabilité complexe
    " (questions) Autres sociétés écran?; Complices à l'étranger?; Montant total blanchi?

@hyp_f_03 Réseau de démarcheurs complices (type) hypothèse
    " (statut) en_attente
    " (confiance) 60%
    " (source) ai
    " (description) Certains des 5 commerciaux recrutés pourraient être des complices conscients du caractère frauduleux de l'opération.
    " (pour) Commissions élevées, Méthodes de recrutement agressives
    " (contre) Possibles victimes eux-mêmes
    " (questions) Les commerciaux savaient-ils?; Ont-ils investi eux-mêmes?

// =============================================================
// RÉSEAU DE RELATIONS - Graphe sémantique
// =============================================================

:: réseau de relations ::

# Légende STTypes: N=proximité, +L=causalité, +C=containment, +E=expression

// Relations hiérarchiques
$pdg.1 (dirige:+C) $societe.1
$pdg.1 (supérieur de:+C) $directrice.1
$pdg.1 (supérieur de:+C) $comptable.1

// Relations de fraude
$pdg.1 (a escroqué:+L) $victimes.1
$directrice.1 (a recruté:+L) $victimes.1
$comptable.1 (a falsifié:+L) $bilans.1

// Relations financières
$societe.1 (transfère vers:-C) $compte_suisse.1
$compte_suisse.1 (transfère vers:-C) $compte_panama.1
$compte_panama.1 (bénéficiaire:N) $pdg.1

// Relations d'alerte
$lanceur.1 (a dénoncé:+L) $pdg.1
$lanceur.1 (a alerté:+L) $amf.1
$amf.1 (a saisi:+L) Parquet de Paris

// =============================================================
// CHAÎNES CAUSALES DÉTECTÉES
// =============================================================

:: chaînes causales ::

+:: _sequence_ ::

# Chaîne 1: Schéma de Ponzi
@chain_ponzi Recrutement investisseurs (mène à) Collecte fonds (mène à) Paiement anciens avec nouveaux (mène à) Besoin permanent de nouveaux

# Chaîne 2: Effondrement
@chain_collapse Ralentissement recrutement (mène à) Retards paiement (mène à) Plaintes (mène à) Enquête (mène à) Effondrement

# Chaîne 3: Blanchiment
@chain_blanchi Fonds collectés (vers) Compte société (vers) Compte Suisse (vers) Panama (vers) Biens personnels

-:: _sequence_ ::

// =============================================================
// RÉFÉRENCES CROISÉES - Pour utilisation $alias.n
// =============================================================

:: références croisées ::

# Alias pour références rapides
suspects => {Philippe Martin, Céline Roux, Marc Dubois}
victimes => {Association Victimes FinanceMax, Marcel Lefebvre, Dr. Catherine Moreau}
temoins => {Antoine Mercier, AMF, Laurent Petit}
organisations => {FinanceMax SARL, Compte HSBC Suisse, Société écran Panama}
preuves => {Contrats d'investissement, Bilans falsifiés, Emails internes, Relevés bancaires}

// =============================================================
// NOTES D'ENQUÊTE - TODO items
// =============================================================

:: notes, TODO ::

TRACER TOUS LES FLUX VERS COMPTES OFFSHORE
IDENTIFIER D'ÉVENTUELS COMPLICES BANCAIRES
AUDITIONNER LES 5 COMMERCIAUX
ÉVALUER LE PRÉJUDICE EXACT DE CHAQUE VICTIME
RECHERCHER D'AUTRES SOCIÉTÉS ÉCRAN
COOPÉRATION INTERNATIONALE SUISSE/PANAMA
SAISIR LES BIENS PERSONNELS DES DIRIGEANTS
`
}
