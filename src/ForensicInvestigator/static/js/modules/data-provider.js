/**
 * DataProvider - Couche d'abstraction pour utiliser N4L comme source unique de données
 *
 * Ce module gère :
 * - Le chargement et parsing du contenu N4L
 * - La synchronisation bidirectionnelle UI <-> N4L
 * - L'exposition des données aux modules existants (entities, evidence, timeline, hypotheses)
 * - Le pattern Observer pour notifier les modules des changements
 *
 * @version 1.0.0
 */

const DataProvider = {
    // État interne
    currentCaseId: null,
    n4lContent: '',
    parsedData: null,
    isLoading: false,
    lastSync: null,

    // Cache pour éviter les re-parsings inutiles
    _cache: {
        entities: null,
        evidence: null,
        timeline: null,
        hypotheses: null,
        graph: null
    },

    // Subscribers pour le pattern Observer
    _subscribers: [],

    // ============================================
    // Initialisation et configuration
    // ============================================

    /**
     * Initialise le DataProvider avec un cas
     * @param {string} caseId - ID du cas à charger
     * @returns {Promise<object>} Données parsées du cas
     */
    async init(caseId) {
        this.currentCaseId = caseId;
        this._clearCache();
        return await this.loadCaseData(caseId);
    },

    /**
     * Charge les données d'un cas depuis l'API
     * @param {string} caseId - ID du cas
     * @returns {Promise<object>} Données parsées
     */
    async loadCaseData(caseId) {
        this.isLoading = true;

        try {
            // Récupérer le contenu N4L du cas via /api/n4l/export (retourne du texte)
            const n4lResponse = await fetch(`/api/n4l/export?case_id=${caseId}`);

            if (!n4lResponse.ok) {
                throw new Error(`Erreur chargement N4L: ${n4lResponse.statusText}`);
            }

            this.n4lContent = await n4lResponse.text();

            // Parser le N4L pour obtenir les données structurées
            if (this.n4lContent && this.n4lContent.trim()) {
                await this._parseN4LContent(caseId);
            } else {
                // Pas de contenu N4L, initialiser avec des données vides
                this.parsedData = {
                    entities: [],
                    evidence: [],
                    timeline: [],
                    hypotheses: [],
                    graph: { nodes: [], edges: [] },
                    contexts: [],
                    aliases: {},
                    sequences: []
                };
                this._updateCache();
            }

            this.lastSync = new Date();
            this.isLoading = false;
            this._notifySubscribers('load', this.parsedData);

            return this.parsedData;

        } catch (error) {
            console.error('DataProvider: Erreur chargement:', error);
            this.isLoading = false;
            throw error;
        }
    },

    /**
     * Parse le contenu N4L via l'API backend
     * @param {string} caseId - ID du cas
     * @private
     */
    async _parseN4LContent(caseId) {
        try {
            const response = await fetch('/api/n4l/parse', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    content: this.n4lContent,
                    case_id: caseId
                })
            });

            if (!response.ok) {
                throw new Error(`Erreur parsing N4L: ${response.statusText}`);
            }

            this.parsedData = await response.json();
            this._updateCache();

        } catch (error) {
            console.error('DataProvider: Erreur parsing N4L:', error);
            throw error;
        }
    },

    // ============================================
    // Accesseurs de données (lecture)
    // ============================================

    /**
     * Retourne toutes les entités
     * @returns {Array} Liste des entités
     */
    getEntities() {
        if (this._cache.entities) return this._cache.entities;
        return this.parsedData?.entities || [];
    },

    /**
     * Retourne une entité par son ID
     * @param {string} entityId - ID de l'entité
     * @returns {object|null} Entité ou null
     */
    getEntity(entityId) {
        return this.getEntities().find(e => e.id === entityId) || null;
    },

    /**
     * Retourne les entités filtrées par rôle
     * @param {string} role - Rôle (victime, suspect, temoin, autre)
     * @returns {Array} Entités filtrées
     */
    getEntitiesByRole(role) {
        return this.getEntities().filter(e => e.role === role);
    },

    /**
     * Retourne les entités filtrées par type
     * @param {string} type - Type (personne, lieu, objet, etc.)
     * @returns {Array} Entités filtrées
     */
    getEntitiesByType(type) {
        return this.getEntities().filter(e => e.type === type);
    },

    /**
     * Retourne toutes les preuves
     * @returns {Array} Liste des preuves
     */
    getEvidence() {
        if (this._cache.evidence) return this._cache.evidence;
        return this.parsedData?.evidence || [];
    },

    /**
     * Retourne une preuve par son ID
     * @param {string} evidenceId - ID de la preuve
     * @returns {object|null} Preuve ou null
     */
    getEvidenceById(evidenceId) {
        return this.getEvidence().find(e => e.id === evidenceId) || null;
    },

    /**
     * Retourne les preuves filtrées par type
     * @param {string} type - Type de preuve
     * @returns {Array} Preuves filtrées
     */
    getEvidenceByType(type) {
        return this.getEvidence().filter(e => e.type === type);
    },

    /**
     * Retourne la timeline (événements)
     * @returns {Array} Liste des événements triés par timestamp
     */
    getTimeline() {
        if (this._cache.timeline) return this._cache.timeline;
        const timeline = this.parsedData?.timeline || [];
        // Trier par timestamp
        return timeline.sort((a, b) => new Date(a.timestamp) - new Date(b.timestamp));
    },

    /**
     * Retourne un événement par son ID
     * @param {string} eventId - ID de l'événement
     * @returns {object|null} Événement ou null
     */
    getTimelineEvent(eventId) {
        return this.getTimeline().find(e => e.id === eventId) || null;
    },

    /**
     * Retourne toutes les hypothèses
     * @returns {Array} Liste des hypothèses
     */
    getHypotheses() {
        if (this._cache.hypotheses) return this._cache.hypotheses;
        return this.parsedData?.hypotheses || [];
    },

    /**
     * Retourne une hypothèse par son ID
     * @param {string} hypothesisId - ID de l'hypothèse
     * @returns {object|null} Hypothèse ou null
     */
    getHypothesis(hypothesisId) {
        return this.getHypotheses().find(h => h.id === hypothesisId) || null;
    },

    /**
     * Retourne les hypothèses filtrées par statut
     * @param {string} status - Statut (en_attente, corroboree, refutee, partielle)
     * @returns {Array} Hypothèses filtrées
     */
    getHypothesesByStatus(status) {
        return this.getHypotheses().filter(h => h.status === status);
    },

    /**
     * Retourne les données du graphe
     * @returns {object} {nodes: [], edges: []}
     */
    getGraphData() {
        if (this._cache.graph) return this._cache.graph;
        return this.parsedData?.graph || { nodes: [], edges: [] };
    },

    /**
     * Retourne les relations (edges du graphe)
     * @returns {Array} Liste des relations
     */
    getRelations() {
        return this.parsedData?.relations || this.getGraphData().edges || [];
    },

    /**
     * Retourne les contextes N4L détectés
     * @returns {Array} Liste des contextes
     */
    getContexts() {
        return this.parsedData?.contexts || [];
    },

    /**
     * Retourne les alias N4L
     * @returns {object} Map alias -> valeurs
     */
    getAliases() {
        return this.parsedData?.aliases || {};
    },

    /**
     * Retourne les séquences (chronologies) N4L
     * @returns {Array} Liste des séquences
     */
    getSequences() {
        return this.parsedData?.sequences || [];
    },

    /**
     * Retourne le contenu N4L brut
     * @returns {string} Contenu N4L
     */
    getN4LContent() {
        return this.n4lContent;
    },

    // ============================================
    // Modificateurs de données (écriture)
    // Ces méthodes génèrent automatiquement les modifications N4L
    // ============================================

    /**
     * Ajoute une nouvelle entité
     * @param {object} entity - Données de l'entité
     * @returns {Promise<object>} Entité créée
     */
    async addEntity(entity) {
        // Générer le fragment N4L
        const fragment = await this._generateN4LFragment('entity', entity);

        // Appliquer le patch
        await this._applyPatch({
            operation: 'add',
            entity_type: 'entity',
            entity_id: entity.id || this._generateId('ent'),
            n4l_fragment: fragment,
            context: this._getContextForRole(entity.role)
        });

        return entity;
    },

    /**
     * Met à jour une entité existante
     * @param {object} entity - Données de l'entité
     * @returns {Promise<object>} Entité mise à jour
     */
    async updateEntity(entity) {
        const fragment = await this._generateN4LFragment('entity', entity);

        await this._applyPatch({
            operation: 'update',
            entity_type: 'entity',
            entity_id: entity.id,
            n4l_fragment: fragment,
            context: this._getContextForRole(entity.role)
        });

        return entity;
    },

    /**
     * Supprime une entité
     * @param {string} entityId - ID de l'entité
     * @returns {Promise<boolean>} Succès
     */
    async deleteEntity(entityId) {
        await this._applyPatch({
            operation: 'delete',
            entity_type: 'entity',
            entity_id: entityId
        });

        return true;
    },

    /**
     * Ajoute une nouvelle preuve
     * @param {object} evidence - Données de la preuve
     * @returns {Promise<object>} Preuve créée
     */
    async addEvidence(evidence) {
        const fragment = await this._generateN4LFragment('evidence', evidence);

        await this._applyPatch({
            operation: 'add',
            entity_type: 'evidence',
            entity_id: evidence.id || this._generateId('ev'),
            n4l_fragment: fragment,
            context: 'preuves'
        });

        return evidence;
    },

    /**
     * Met à jour une preuve existante
     * @param {object} evidence - Données de la preuve
     * @returns {Promise<object>} Preuve mise à jour
     */
    async updateEvidence(evidence) {
        const fragment = await this._generateN4LFragment('evidence', evidence);

        await this._applyPatch({
            operation: 'update',
            entity_type: 'evidence',
            entity_id: evidence.id,
            n4l_fragment: fragment,
            context: 'preuves'
        });

        return evidence;
    },

    /**
     * Supprime une preuve
     * @param {string} evidenceId - ID de la preuve
     * @returns {Promise<boolean>} Succès
     */
    async deleteEvidence(evidenceId) {
        await this._applyPatch({
            operation: 'delete',
            entity_type: 'evidence',
            entity_id: evidenceId
        });

        return true;
    },

    /**
     * Ajoute un événement à la timeline
     * @param {object} event - Données de l'événement
     * @returns {Promise<object>} Événement créé
     */
    async addTimelineEvent(event) {
        const fragment = await this._generateN4LFragment('timeline', event);

        await this._applyPatch({
            operation: 'add',
            entity_type: 'timeline',
            entity_id: event.id || this._generateId('evt'),
            n4l_fragment: fragment,
            context: 'chronologie'
        });

        return event;
    },

    /**
     * Met à jour un événement timeline
     * @param {object} event - Données de l'événement
     * @returns {Promise<object>} Événement mis à jour
     */
    async updateTimelineEvent(event) {
        const fragment = await this._generateN4LFragment('timeline', event);

        await this._applyPatch({
            operation: 'update',
            entity_type: 'timeline',
            entity_id: event.id,
            n4l_fragment: fragment,
            context: 'chronologie'
        });

        return event;
    },

    /**
     * Supprime un événement timeline
     * @param {string} eventId - ID de l'événement
     * @returns {Promise<boolean>} Succès
     */
    async deleteTimelineEvent(eventId) {
        await this._applyPatch({
            operation: 'delete',
            entity_type: 'timeline',
            entity_id: eventId
        });

        return true;
    },

    /**
     * Ajoute une nouvelle hypothèse
     * @param {object} hypothesis - Données de l'hypothèse
     * @returns {Promise<object>} Hypothèse créée
     */
    async addHypothesis(hypothesis) {
        const fragment = await this._generateN4LFragment('hypothesis', hypothesis);

        await this._applyPatch({
            operation: 'add',
            entity_type: 'hypothesis',
            entity_id: hypothesis.id || this._generateId('hyp'),
            n4l_fragment: fragment,
            context: 'hypotheses'
        });

        return hypothesis;
    },

    /**
     * Met à jour une hypothèse existante
     * @param {object} hypothesis - Données de l'hypothèse
     * @returns {Promise<object>} Hypothèse mise à jour
     */
    async updateHypothesis(hypothesis) {
        const fragment = await this._generateN4LFragment('hypothesis', hypothesis);

        await this._applyPatch({
            operation: 'update',
            entity_type: 'hypothesis',
            entity_id: hypothesis.id,
            n4l_fragment: fragment,
            context: 'hypotheses'
        });

        return hypothesis;
    },

    /**
     * Supprime une hypothèse
     * @param {string} hypothesisId - ID de l'hypothèse
     * @returns {Promise<boolean>} Succès
     */
    async deleteHypothesis(hypothesisId) {
        await this._applyPatch({
            operation: 'delete',
            entity_type: 'hypothesis',
            entity_id: hypothesisId
        });

        return true;
    },

    /**
     * Ajoute une relation entre deux entités
     * @param {object} relation - Données de la relation
     * @returns {Promise<object>} Relation créée
     */
    async addRelation(relation) {
        const entityNames = {};
        this.getEntities().forEach(e => {
            entityNames[e.id] = e.name;
        });

        const response = await fetch('/api/n4l/generate?type=relation', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ relation, entity_names: entityNames })
        });

        if (!response.ok) {
            throw new Error('Erreur génération relation N4L');
        }

        const { n4l_fragment } = await response.json();

        await this._applyPatch({
            operation: 'add',
            entity_type: 'relation',
            entity_id: `${relation.from_id}_${relation.to_id}`,
            n4l_fragment: n4l_fragment,
            context: relation.context || 'relations'
        });

        return relation;
    },

    // ============================================
    // Synchronisation N4L
    // ============================================

    /**
     * Met à jour le contenu N4L depuis l'éditeur et re-parse
     * @param {string} newContent - Nouveau contenu N4L
     * @returns {Promise<object>} Données parsées
     */
    async updateN4LContent(newContent) {
        this.n4lContent = newContent;
        await this._parseN4LContent(this.currentCaseId);

        // Sauvegarder dans le backend
        await fetch(`/api/n4l/sync?case_id=${this.currentCaseId}`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ n4l_content: newContent })
        });

        this.lastSync = new Date();
        this._notifySubscribers('update', this.parsedData);

        return this.parsedData;
    },

    /**
     * Force un re-parsing du contenu N4L actuel
     * @returns {Promise<object>} Données parsées
     */
    async refresh() {
        await this._parseN4LContent(this.currentCaseId);
        this._notifySubscribers('refresh', this.parsedData);
        return this.parsedData;
    },

    /**
     * Valide le contenu N4L
     * @param {string} content - Contenu à valider (optionnel, utilise le courant sinon)
     * @returns {Promise<object>} {valid: boolean, errors: []}
     */
    async validateN4L(content = null) {
        const response = await fetch('/api/n4l/validate', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ content: content || this.n4lContent })
        });

        return await response.json();
    },

    // ============================================
    // Pattern Observer
    // ============================================

    /**
     * S'abonne aux changements de données
     * @param {function} callback - Fonction appelée lors des changements
     * @returns {function} Fonction pour se désabonner
     */
    subscribe(callback) {
        this._subscribers.push(callback);

        // Retourne une fonction pour se désabonner
        return () => {
            const index = this._subscribers.indexOf(callback);
            if (index > -1) {
                this._subscribers.splice(index, 1);
            }
        };
    },

    /**
     * Notifie tous les subscribers d'un changement
     * @param {string} eventType - Type d'événement (load, update, refresh, add, delete)
     * @param {object} data - Données à transmettre
     * @private
     */
    _notifySubscribers(eventType, data) {
        this._subscribers.forEach(callback => {
            try {
                callback(eventType, data);
            } catch (error) {
                console.error('DataProvider: Erreur subscriber:', error);
            }
        });
    },

    // ============================================
    // Méthodes internes
    // ============================================

    /**
     * Génère un fragment N4L via l'API
     * @param {string} type - Type d'élément
     * @param {object} data - Données de l'élément
     * @returns {Promise<string>} Fragment N4L
     * @private
     */
    async _generateN4LFragment(type, data) {
        const response = await fetch(`/api/n4l/generate?type=${type}`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(data)
        });

        if (!response.ok) {
            throw new Error(`Erreur génération N4L ${type}`);
        }

        const result = await response.json();
        return result.n4l_fragment;
    },

    /**
     * Applique un patch au contenu N4L
     * @param {object} patch - Patch à appliquer
     * @private
     */
    async _applyPatch(patch) {
        const response = await fetch(`/api/n4l/patch?case_id=${this.currentCaseId}`, {
            method: 'PATCH',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(patch)
        });

        if (!response.ok) {
            throw new Error('Erreur application patch N4L');
        }

        const result = await response.json();

        if (result.success) {
            this.n4lContent = result.n4l_content;
            this.parsedData = result.parsed_data;
            this._updateCache();
            this.lastSync = new Date();
            this._notifySubscribers(patch.operation, this.parsedData);
        } else {
            throw new Error(result.message || 'Erreur patch N4L');
        }
    },

    /**
     * Met à jour le cache interne
     * @private
     */
    _updateCache() {
        if (this.parsedData) {
            this._cache.entities = this.parsedData.entities || [];
            this._cache.evidence = this.parsedData.evidence || [];
            this._cache.timeline = this.parsedData.timeline || [];
            this._cache.hypotheses = this.parsedData.hypotheses || [];
            this._cache.graph = this.parsedData.graph || { nodes: [], edges: [] };
        }
    },

    /**
     * Vide le cache
     * @private
     */
    _clearCache() {
        this._cache = {
            entities: null,
            evidence: null,
            timeline: null,
            hypotheses: null,
            graph: null
        };
    },

    /**
     * Génère un ID unique
     * @param {string} prefix - Préfixe de l'ID
     * @returns {string} ID généré
     * @private
     */
    _generateId(prefix) {
        return `${prefix}_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
    },

    /**
     * Retourne le contexte N4L approprié pour un rôle
     * @param {string} role - Rôle de l'entité
     * @returns {string} Nom du contexte
     * @private
     */
    _getContextForRole(role) {
        const roleContextMap = {
            'victime': 'victimes',
            'suspect': 'suspects',
            'temoin': 'témoins',
            'enqueteur': 'enquêteurs',
            'autre': 'entités'
        };
        return roleContextMap[role] || 'entités';
    }
};

// Exporter pour utilisation dans les modules
if (typeof module !== 'undefined' && module.exports) {
    module.exports = DataProvider;
}
