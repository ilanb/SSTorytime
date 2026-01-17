// ForensicInvestigator - Module Graph
// Gestion du graphe principal et N4L

const GraphModule = {
    // ============================================
    // Get Graph Data - Source unique: API /api/graph (priorise N4L côté backend)
    // ============================================
    async getGraphData() {
        if (!this.currentCase) return { nodes: [], edges: [] };

        // Utiliser l'API /api/graph qui priorise le contenu N4L côté backend
        // Cela garantit la cohérence entre le graphe principal et le graphe N4L
        try {
            const graphData = await this.apiCall(`/api/graph?case_id=${this.currentCase.id}`);
            console.log('GraphModule: Données depuis /api/graph (N4L prioritaire)', {
                nodes: graphData.nodes?.length || 0,
                edges: graphData.edges?.length || 0
            });

            // Synchroniser avec DataProvider si disponible
            if (typeof DataProvider !== 'undefined' && DataProvider.parsedData) {
                DataProvider._cache.graph = graphData;
            }

            return graphData;
        } catch (error) {
            console.error('GraphModule: Erreur chargement graphe', error);
            return { nodes: [], edges: [] };
        }
    },

    // ============================================
    // Get Edge Color based on N4L type
    // ============================================
    getEdgeColor(edge) {
        switch (edge.type) {
            case 'never': return '#dc2626';      // Rouge - relation interdite
            case 'new': return '#059669';         // Vert - nouvelle relation
            case 'sequence': return '#f59e0b';    // Orange - séquence
            case 'equivalence': return '#8b5cf6'; // Violet - équivalence
            case 'group': return '#06b6d4';       // Cyan - groupe
            case 'contains': return '#3b82f6';    // Bleu - contient
            default: return '#1e3a5f';            // Bleu foncé par défaut
        }
    },

    // ============================================
    // Render Graph - Unifié (DB + N4L)
    // ============================================
    async renderGraph() {
        const container = document.getElementById('graph-container');
        if (!container || !this.currentCase) return;

        const graphData = await this.getGraphData();

        if (!graphData.nodes || graphData.nodes.length === 0) {
            container.innerHTML = `
                <div class="empty-state" style="height: 100%; display: flex; flex-direction: column; justify-content: center;">
                    <span class="material-icons empty-state-icon">hub</span>
                    <p class="empty-state-title">Graphe vide</p>
                    <p class="empty-state-description">Ajoutez des entités pour visualiser les relations</p>
                </div>
            `;
            return;
        }

        container.innerHTML = '';

        // Store original data for reset
        this.graphNodesData = graphData.nodes;
        this.graphEdgesData = graphData.edges;

        // Créer les noeuds avec support des métadonnées N4L
        const nodes = new vis.DataSet(graphData.nodes.map(n => ({
            id: n.id,
            label: n.label,
            color: this.getNodeColor(n),
            shape: this.getNodeShape(n),
            title: this.getNodeTooltip(n),
            originalColor: this.getNodeColor(n),
            // Métadonnées N4L
            context: n.context,
            nodeType: n.type,
            role: n.role
        })));

        // Créer les arêtes avec support des types N4L
        const edges = new vis.DataSet(graphData.edges.map((e, i) => {
            const edgeColor = this.getEdgeColor(e);
            return {
                id: `edge-${i}`,
                from: e.from,
                to: e.to,
                label: e.label,
                arrows: e.type === 'equivalence' ? '' : 'to',
                dashes: e.type === 'new',
                color: { color: edgeColor },
                title: e.context ? `Contexte: ${e.context}` : '',
                originalColor: edgeColor,
                edgeType: e.type,
                context: e.context
            };
        }));

        const options = {
            nodes: {
                font: { color: '#1a1a2e', size: 12 },
                borderWidth: 2
            },
            edges: {
                font: { size: 10, color: '#4a5568' },
                smooth: { type: 'curvedCW', roundness: 0.2 }
            },
            physics: {
                stabilization: { iterations: 150 },
                barnesHut: {
                    gravitationalConstant: -3000,
                    springLength: 200,
                    springConstant: 0.02,
                    damping: 0.3,
                    avoidOverlap: 0.5
                },
                minVelocity: 0.75
            },
            interaction: {
                hover: true,
                tooltipDelay: 200
            }
        };

        this.graph = new vis.Network(container, { nodes, edges }, options);
        this.graphNodes = nodes;
        this.graphEdges = edges;
        this.selectedGraphNode = null;

        // Store original data for reset
        this.originalGraphData = { nodes: nodes.get(), edges: edges.get() };

        // Click event for focus/blur
        this.graph.on('click', (params) => {
            if (params.nodes.length > 0) {
                this.focusGraphNode(params.nodes[0]);
            } else {
                this.resetGraphFocus();
            }
        });

        // Right-click context menu
        this.graph.on('oncontext', (params) => {
            params.event.preventDefault();
            this.handleGraphRightClick(params);
        });

        // Close context menu on click elsewhere
        document.addEventListener('click', () => this.hideContextMenu());

        // Setup context menu actions
        this.setupContextMenuActions();

        // Ajouter la légende N4L si des types spéciaux sont présents
        this.addGraphLegendIfNeeded(container, graphData);

        // Afficher les métadonnées du graphe (comme N4L)
        this.renderGraphMetadata();
    },

    // ============================================
    // Add Legend for N4L edge types
    // ============================================
    addGraphLegendIfNeeded(container, graphData) {
        // Vérifier si des types N4L spéciaux existent
        const hasSpecialTypes = graphData.edges?.some(e =>
            e.type && ['never', 'new', 'sequence', 'equivalence', 'group', 'contains'].includes(e.type)
        );

        if (!hasSpecialTypes) return;

        const legendHtml = `
            <div class="graph-legend" style="position: absolute; bottom: 10px; left: 10px; background: rgba(255,255,255,0.95); padding: 8px 12px; border-radius: 8px; font-size: 11px; box-shadow: 0 2px 8px rgba(0,0,0,0.1); z-index: 100;">
                <div style="font-weight: 600; margin-bottom: 6px; color: #1a1a2e;">Types de relations</div>
                <div style="display: flex; flex-wrap: wrap; gap: 8px;">
                    <span style="display: flex; align-items: center; gap: 4px;"><span style="width: 20px; height: 2px; background: #1e3a5f;"></span> Standard</span>
                    <span style="display: flex; align-items: center; gap: 4px;"><span style="width: 20px; height: 2px; background: #059669;"></span> Nouvelle</span>
                    <span style="display: flex; align-items: center; gap: 4px;"><span style="width: 20px; height: 2px; background: #dc2626;"></span> Interdite</span>
                    <span style="display: flex; align-items: center; gap: 4px;"><span style="width: 20px; height: 2px; background: #f59e0b;"></span> Séquence</span>
                    <span style="display: flex; align-items: center; gap: 4px;"><span style="width: 20px; height: 2px; background: #8b5cf6;"></span> Équivalence</span>
                </div>
            </div>
        `;

        const legendDiv = document.createElement('div');
        legendDiv.innerHTML = legendHtml;
        container.style.position = 'relative';
        container.appendChild(legendDiv.firstElementChild);
    },

    // ============================================
    // Unified Graph Access Methods
    // Ces méthodes permettent d'accéder au graphe de manière transparente
    // qu'il s'agisse du graphe principal ou du graphe N4L
    // ============================================

    /**
     * Retourne le graphe actif (principal ou N4L selon le contexte)
     * @param {boolean} preferN4L - Si true, préfère le graphe N4L si disponible
     * @returns {vis.Network|null} Le graphe vis.js
     */
    getActiveGraph(preferN4L = false) {
        if (preferN4L && this.n4lGraph) {
            return this.n4lGraph;
        }
        return this.graph || this.n4lGraph || null;
    },

    /**
     * Retourne les noeuds du graphe actif
     * @param {boolean} preferN4L - Si true, préfère les noeuds N4L
     * @returns {vis.DataSet|null} DataSet des noeuds
     */
    getActiveGraphNodes(preferN4L = false) {
        if (preferN4L && this.n4lGraphNodes) {
            return this.n4lGraphNodes;
        }
        return this.graphNodes || this.n4lGraphNodes || null;
    },

    /**
     * Retourne les arêtes du graphe actif
     * @param {boolean} preferN4L - Si true, préfère les arêtes N4L
     * @returns {vis.DataSet|null} DataSet des arêtes
     */
    getActiveGraphEdges(preferN4L = false) {
        if (preferN4L && this.n4lGraphEdges) {
            return this.n4lGraphEdges;
        }
        return this.graphEdges || this.n4lGraphEdges || null;
    },

    /**
     * Vérifie si les deux graphes sont synchronisés (mêmes données)
     * @returns {boolean} True si synchronisés
     */
    areGraphsSynchronized() {
        if (!this.graphNodes || !this.n4lGraphNodes) return false;

        const mainNodeIds = new Set(this.graphNodes.getIds());
        const n4lNodeIds = new Set(this.n4lGraphNodes.getIds());

        if (mainNodeIds.size !== n4lNodeIds.size) return false;

        for (const id of mainNodeIds) {
            if (!n4lNodeIds.has(id)) return false;
        }

        return true;
    },

    /**
     * Synchronise le graphe N4L avec le graphe principal
     * Copie les données du graphe principal vers le graphe N4L
     */
    syncGraphToN4L() {
        if (!this.graph || !this.graphNodes || !this.graphEdges) return;

        // Si le conteneur N4L existe, mettre à jour son graphe
        const n4lContainer = document.getElementById('n4l-graph-container');
        if (n4lContainer && this.graphNodesData && this.graphEdgesData) {
            this.renderN4LGraph({
                nodes: this.graphNodesData,
                edges: this.graphEdgesData
            });
        }
    },

    // ============================================
    // Node Colors and Shapes
    // ============================================
    getNodeColor(node) {
        // Couleurs par rôle (prioritaire)
        const roleColors = {
            victime: { background: '#dc2626', border: '#991b1b' },      // Rouge
            suspect: { background: '#f59e0b', border: '#d97706' },      // Orange
            temoin: { background: '#3b82f6', border: '#2563eb' },       // Bleu
            enqueteur: { background: '#10b981', border: '#059669' }     // Vert
        };

        // Couleurs par type (si pas de rôle)
        const typeColors = {
            personne: { background: '#6366f1', border: '#4f46e5' },     // Indigo
            lieu: { background: '#14b8a6', border: '#0d9488' },         // Teal
            objet: { background: '#8b5cf6', border: '#7c3aed' },        // Violet
            preuve: { background: '#ec4899', border: '#db2777' },       // Pink
            evenement: { background: '#06b6d4', border: '#0891b2' },    // Cyan
            document: { background: '#84cc16', border: '#65a30d' },     // Lime
            organisation: { background: '#f97316', border: '#ea580c' }  // Orange
        };

        const role = node.role?.toLowerCase();
        const type = node.type?.toLowerCase();

        if (role && roleColors[role]) {
            return roleColors[role];
        }
        if (type && typeColors[type]) {
            return typeColors[type];
        }
        return { background: '#64748b', border: '#475569' }; // Slate default
    },

    getNodeShape(node) {
        // Forme par rôle (prioritaire)
        const role = node.role?.toLowerCase();
        if (role) {
            const roleShapes = {
                victime: 'star',           // Étoile pour victimes
                suspect: 'triangle',       // Triangle pour suspects
                temoin: 'square',          // Carré pour témoins
                enqueteur: 'hexagon'       // Hexagone pour enquêteurs
            };
            if (roleShapes[role]) return roleShapes[role];
        }

        // Forme par type
        const type = node.type?.toLowerCase();
        const typeShapes = {
            personne: 'dot',           // Cercle pour personnes
            lieu: 'square',            // Carré pour lieux
            objet: 'diamond',          // Losange pour objets
            preuve: 'star',            // Étoile pour preuves
            evenement: 'triangleDown', // Triangle inversé pour événements
            document: 'box',           // Rectangle pour documents
            organisation: 'ellipse'    // Ellipse pour organisations
        };

        return typeShapes[type] || 'dot';
    },

    getNodeTooltip(node) {
        const parts = [node.label];
        if (node.role) parts.push(`Rôle: ${node.role}`);
        if (node.type) parts.push(`Type: ${node.type}`);
        if (node.context) parts.push(`Contexte: ${node.context}`);
        return parts.join('\n');
    },

    // ============================================
    // Graph Focus/Blur
    // ============================================
    focusGraphNode(nodeId) {
        if (!this.graph || !this.graphNodes || !this.graphEdges) return;

        // Find connected nodes
        const connectedNodes = new Set([nodeId]);
        const connectedEdges = new Set();

        this.graphEdges.forEach(edge => {
            if (edge.from === nodeId || edge.to === nodeId) {
                connectedNodes.add(edge.from);
                connectedNodes.add(edge.to);
                connectedEdges.add(edge.id);
            }
        });

        // Update nodes
        const nodeUpdates = this.graphNodes.map(node => {
            const isConnected = connectedNodes.has(node.id);
            if (isConnected) {
                return {
                    id: node.id,
                    opacity: 1,
                    borderWidth: node.id === nodeId ? 4 : 2,
                    font: { color: '#1a1a2e' }
                };
            } else {
                return {
                    id: node.id,
                    opacity: 0.2,
                    borderWidth: 1,
                    font: { color: '#a0aec0' }
                };
            }
        });

        this.graphNodes.update(nodeUpdates);

        // Update edges (including label opacity)
        const edgeUpdates = this.graphEdges.map(edge => {
            const isConnected = connectedEdges.has(edge.id);
            return {
                id: edge.id,
                color: isConnected
                    ? { color: '#1e3a5f', opacity: 1 }
                    : { color: '#e2e8f0', opacity: 0.2 },
                font: {
                    color: isConnected ? '#4a5568' : 'rgba(74, 85, 104, 0.15)'
                }
            };
        });

        this.graphEdges.update(edgeUpdates);

        // Focus camera
        this.graph.focus(nodeId, { scale: 1.5, animation: true });
    },

    resetGraphFocus() {
        if (!this.graph || !this.originalGraphData) return;

        // Restore original nodes with their original colors and visibility
        const nodeUpdates = this.originalGraphData.nodes.map(node => ({
            id: node.id,
            hidden: false,
            color: node.originalColor || this.getNodeColor(node),
            opacity: 1,
            borderWidth: 2,
            font: { color: '#1a1a2e', size: 12 }
        }));

        this.graphNodes.update(nodeUpdates);

        // Restore original edges (including label color and visibility)
        const edgeUpdates = this.originalGraphData.edges.map(edge => ({
            id: edge.id,
            hidden: false,
            color: { color: edge.originalColor || '#1e3a5f', opacity: 1 },
            font: { color: '#4a5568' }
        }));

        this.graphEdges.update(edgeUpdates);

        // Reset view
        this.graph.fit({ animation: true });
    },

    addGraphControlButtons(container) {
        console.log('[Graph] Adding control buttons to container');

        // Remove existing buttons if any
        const existing = container.querySelector('.graph-buttons-container');
        if (existing) existing.remove();

        // Boutons container - positioned absolutely within the graph container
        const buttonsContainer = document.createElement('div');
        buttonsContainer.className = 'graph-buttons-container';
        buttonsContainer.style.cssText = `
            position: absolute;
            bottom: 15px;
            left: 15px;
            z-index: 9999;
            display: flex;
            gap: 0.5rem;
            pointer-events: auto;
        `;

        // Reset button
        const resetBtn = document.createElement('button');
        resetBtn.className = 'graph-control-btn';
        resetBtn.innerHTML = '<span class="material-icons">restart_alt</span> Reset';
        resetBtn.onclick = (e) => { e.stopPropagation(); this.resetGraphFocus(); };
        buttonsContainer.appendChild(resetBtn);

        // Fullscreen button
        const fullscreenBtn = document.createElement('button');
        fullscreenBtn.className = 'graph-control-btn';
        fullscreenBtn.id = 'graph-fullscreen-btn';
        fullscreenBtn.innerHTML = '<span class="material-icons">fullscreen</span> Plein écran';
        fullscreenBtn.onclick = (e) => { e.stopPropagation(); this.toggleGraphFullscreen(); };
        buttonsContainer.appendChild(fullscreenBtn);

        container.appendChild(buttonsContainer);
    },

    toggleGraphFullscreen() {
        const container = document.getElementById('graph-container');
        const panel = container?.closest('.panel');
        const btn = document.getElementById('btn-fullscreen-graph');
        const sidebar = document.getElementById('graph-fullscreen-sidebar');

        if (!panel) {
            console.error('[Graph] Could not find parent panel for fullscreen');
            return;
        }

        if (panel.classList.contains('fullscreen-panel')) {
            // Exit fullscreen
            panel.classList.remove('fullscreen-panel');
            document.body.classList.remove('has-fullscreen-panel');
            if (btn) {
                btn.innerHTML = '<span class="material-icons">fullscreen</span> Plein écran';
            }
            // Hide sidebar
            if (sidebar) {
                sidebar.style.display = 'none';
            }
        } else {
            // Enter fullscreen
            panel.classList.add('fullscreen-panel');
            document.body.classList.add('has-fullscreen-panel');
            if (btn) {
                btn.innerHTML = '<span class="material-icons">fullscreen_exit</span> Quitter';
            }
            // Load and show N4L metadata in sidebar
            if (sidebar) {
                this.loadGraphFullscreenSidebar(sidebar);
            }
        }

        // Redraw graph to fit new size
        setTimeout(() => {
            if (this.graph) {
                this.graph.fit({ animation: true });
            }
        }, 100);
    },

    // Load N4L metadata into the fullscreen sidebar
    async loadGraphFullscreenSidebar(sidebar) {
        if (!this.currentCase?.id) {
            sidebar.innerHTML = '<div class="empty-state"><p>Sélectionnez une affaire</p></div>';
            return;
        }

        try {
            // Parse N4L to get metadata
            const response = await fetch('/api/n4l/parse', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ case_id: this.currentCase.id })
            });
            const result = await response.json();

            // Store the parsed data for use by filter functions
            this.dashboardN4LParse = result;

            // Build sidebar HTML (similar to N4L metadata)
            let html = '';

            // Contexts filter
            if (result.contexts && result.contexts.length > 0) {
                html += `
                    <div class="n4l-section">
                        <div class="n4l-section-header">
                            <span class="material-icons">filter_list</span>
                            <span>Filtrer par contexte</span>
                            <button class="n4l-reset-btn ${this.dashboardActiveFilter ? '' : 'hidden'}" onclick="app.resetDashboardFilter()" title="Réinitialiser">
                                <span class="material-icons">refresh</span> Reset
                            </button>
                        </div>
                        <div class="n4l-context-grid">
                            ${result.contexts.map(ctx => {
                                const icon = this.getContextIcon ? this.getContextIcon(ctx) : 'label';
                                return `
                                    <button class="n4l-context-btn" onclick="app.filterDashboardByContext('${ctx}')">
                                        <span class="material-icons">${icon}</span>
                                        <span>${ctx}</span>
                                    </button>
                                `;
                            }).join('')}
                        </div>
                    </div>
                `;
            }

            // Causal Chains
            if (result.causal_chains && result.causal_chains.length > 0) {
                html += `
                    <div class="n4l-section n4l-causal-section">
                        <div class="n4l-section-header n4l-collapsible" onclick="app.toggleN4LSection(this)">
                            <span class="material-icons">route</span>
                            <span>Chaînes Causales (${result.causal_chains.length})</span>
                            <span class="material-icons n4l-expand-icon">expand_more</span>
                        </div>
                        <div class="n4l-section-content n4l-collapsed">
                            <button class="n4l-restore-btn" onclick="event.stopPropagation(); app.resetDashboardFilter();" style="display:none;" id="dashboard-restore-graph-btn">
                                <span class="material-icons">restore</span> Restaurer le graphe complet
                            </button>
                            <div class="n4l-chains-list">
                                ${result.causal_chains.map((chain, i) => `
                                    <div class="n4l-chain-item" onclick="app.showDashboardCausalChain(${i})" title="Cliquer pour visualiser">
                                        <div class="n4l-chain-header">
                                            <span class="n4l-chain-number">${i + 1}</span>
                                            <span class="n4l-chain-id">${chain.id || ''}</span>
                                        </div>
                                        <div class="n4l-chain-steps">
                                            ${chain.steps.map((step, j) => `
                                                <span class="n4l-chain-step">${step.item}</span>
                                                ${j < chain.steps.length - 1 ? '<span class="n4l-chain-arrow">→</span>' : ''}
                                            `).join('')}
                                        </div>
                                    </div>
                                `).join('')}
                            </div>
                        </div>
                    </div>
                `;
            }

            // Sequences (Chronologie)
            if (result.sequences && result.sequences.length > 0) {
                html += `
                    <div class="n4l-section">
                        <div class="n4l-section-header n4l-collapsible" onclick="app.toggleN4LSection(this)">
                            <span class="material-icons">timeline</span>
                            <span>Chronologie (${result.sequences.length} séquence${result.sequences.length > 1 ? 's' : ''})</span>
                            <span class="material-icons n4l-expand-icon">expand_more</span>
                        </div>
                        <div class="n4l-section-content n4l-collapsed">
                            ${result.sequences.map((seq, i) => `
                                <div class="n4l-sequence-item" onclick="app.showDashboardSequence(${i})" title="Cliquer pour visualiser">
                                    <div class="n4l-sequence-header">
                                        <span class="n4l-sequence-number">${i + 1}</span>
                                        <span class="n4l-sequence-count">${seq.length} étapes</span>
                                    </div>
                                    <div class="n4l-sequence-preview">
                                        ${seq.slice(0, 4).join(' → ')}${seq.length > 4 ? ' → ...' : ''}
                                    </div>
                                </div>
                            `).join('')}
                        </div>
                    </div>
                `;
            }

            // Hypotheses info
            const hypAliases = Object.keys(result.aliases || {}).filter(k => k.startsWith('hyp'));
            if (hypAliases.length > 0) {
                html += `
                    <div class="n4l-section">
                        <div class="n4l-section-header n4l-collapsible" onclick="app.toggleN4LSection(this)">
                            <span class="material-icons">lightbulb</span>
                            <span>Hypothèses (${hypAliases.length})</span>
                            <span class="material-icons n4l-expand-icon">expand_more</span>
                        </div>
                        <div class="n4l-section-content n4l-collapsed">
                            <button class="n4l-context-btn" onclick="app.showDashboardHypotheses()" style="width: 100%;">
                                <span class="material-icons">visibility</span>
                                <span>Afficher les hypothèses</span>
                            </button>
                        </div>
                    </div>
                `;
            }

            // Cross References
            if (result.cross_refs && result.cross_refs.length > 0) {
                const groupedRefs = {};
                result.cross_refs.forEach(ref => {
                    if (!groupedRefs[ref.alias]) groupedRefs[ref.alias] = [];
                    groupedRefs[ref.alias].push(ref);
                });

                html += `
                    <div class="n4l-section">
                        <div class="n4l-section-header n4l-collapsible" onclick="app.toggleN4LSection(this)">
                            <span class="material-icons">link</span>
                            <span>Références Croisées (${result.cross_refs.length})</span>
                            <span class="material-icons n4l-expand-icon">expand_more</span>
                        </div>
                        <div class="n4l-section-content n4l-collapsed">
                            ${Object.entries(groupedRefs).slice(0, 10).map(([alias, refs]) => `
                                <div class="n4l-crossref-group">
                                    <span class="n4l-crossref-alias">$${alias}</span>
                                    <span class="n4l-crossref-count">(${refs.length})</span>
                                </div>
                            `).join('')}
                            ${Object.keys(groupedRefs).length > 10 ? `<div class="n4l-more">+${Object.keys(groupedRefs).length - 10} autres...</div>` : ''}
                        </div>
                    </div>
                `;
            }

            sidebar.innerHTML = html || '<div class="empty-state"><p>Aucune métadonnée N4L disponible</p></div>';
            sidebar.style.display = 'block';

        } catch (error) {
            console.error('[Graph] Error loading sidebar:', error);
            sidebar.innerHTML = '<div class="empty-state"><p>Erreur de chargement</p></div>';
        }
    },

    // ============================================
    // Dashboard Graph Filtering (fullscreen sidebar)
    // ============================================
    filterDashboardByContext(context) {
        if (!this.dashboardN4LParse || !this.graph) return;

        // Special cases
        if (context === 'chaînes causales') {
            this.showAllDashboardCausalChains();
            return;
        }
        if (context.includes('hypothèse') || context.includes('piste')) {
            this.showDashboardHypotheses();
            return;
        }
        if (context.toLowerCase().includes('todo') || context.toLowerCase().includes('note')) {
            this.showToast('Aucune note/TODO dans le graphe', 'info');
            return;
        }

        // Filter nodes by context
        const result = this.dashboardN4LParse;
        const involvedNodes = new Set();

        // Find nodes with matching context
        result.graph.nodes.forEach(n => {
            if (n.context && (n.context === context || n.context.includes(context) || context.includes(n.context))) {
                involvedNodes.add(n.id);
            }
        });

        // Find edges with matching context
        result.graph.edges.forEach(e => {
            if (e.context && (e.context === context || e.context.includes(context) || context.includes(e.context))) {
                involvedNodes.add(e.from);
                involvedNodes.add(e.to);
            }
        });

        if (involvedNodes.size === 0) {
            this.showToast(`Aucune entité pour "${context}"`, 'info');
            return;
        }

        // Update graph visibility
        const allNodes = this.graphNodes.getIds();
        const nodeUpdates = allNodes.map(id => ({
            id,
            hidden: false,
            color: involvedNodes.has(id) ? undefined : { background: 'rgba(200,200,200,0.1)', border: 'rgba(200,200,200,0.1)' },
            font: { color: involvedNodes.has(id) ? '#1a1a2e' : 'rgba(0,0,0,0.05)' },
            opacity: involvedNodes.has(id) ? 1 : 0.05
        }));
        this.graphNodes.update(nodeUpdates);

        // Fit to visible nodes
        setTimeout(() => {
            this.graph.fit({
                nodes: Array.from(involvedNodes),
                animation: { duration: 400 }
            });
        }, 100);

        this.dashboardActiveFilter = context;
        this.showToast(`Filtré: ${context} (${involvedNodes.size} entités)`);
    },

    resetDashboardFilter() {
        if (!this.graph || !this.graphNodes) return;

        // Reset all nodes to original state
        const allNodes = this.graphNodes.getIds();
        const nodeUpdates = allNodes.map(id => {
            const node = this.graphNodes.get(id);
            return {
                id,
                hidden: false,
                color: node.originalColor || undefined,
                font: { color: '#1a1a2e' },
                opacity: 1
            };
        });
        this.graphNodes.update(nodeUpdates);

        // If we replaced the graph with causal chains, reload original
        if (this.dashboardShowingSpecialView) {
            this.loadGraph();
            this.dashboardShowingSpecialView = false;
        }

        // Hide restore button
        const restoreBtn = document.getElementById('dashboard-restore-graph-btn');
        if (restoreBtn) restoreBtn.style.display = 'none';

        this.dashboardActiveFilter = null;

        setTimeout(() => {
            if (this.graph) this.graph.fit({ animation: true });
        }, 100);

        this.showToast('Graphe réinitialisé');
    },

    showDashboardCausalChain(chainIndex) {
        if (!this.dashboardN4LParse?.causal_chains) return;

        const chain = this.dashboardN4LParse.causal_chains[chainIndex];
        if (!chain || !chain.steps) return;

        // Create chain nodes
        const chainNodes = chain.steps.map((step, i) => {
            const hue = 30 + (i / chain.steps.length) * 30;
            return {
                id: `chain_step_${i}`,
                label: step.item,
                color: { background: `hsl(${hue}, 90%, 55%)`, border: `hsl(${hue}, 90%, 40%)` },
                borderWidth: 3,
                font: { size: 14, color: '#1a1a2e' },
                shape: 'box',
                margin: 10
            };
        });

        // Create chain edges
        const chainEdges = [];
        for (let i = 0; i < chain.steps.length - 1; i++) {
            chainEdges.push({
                id: `chain_edge_${i}`,
                from: `chain_step_${i}`,
                to: `chain_step_${i + 1}`,
                label: chain.steps[i].relation || '',
                arrows: 'to',
                color: { color: '#f59e0b' },
                width: 3,
                font: { size: 12, color: '#666' }
            });
        }

        // Replace graph content
        this.graphNodes.clear();
        this.graphEdges.clear();
        this.graphNodes.add(chainNodes);
        this.graphEdges.add(chainEdges);

        this.dashboardShowingSpecialView = true;

        // Show restore button
        const restoreBtn = document.getElementById('dashboard-restore-graph-btn');
        if (restoreBtn) restoreBtn.style.display = 'flex';

        setTimeout(() => {
            if (this.graph) this.graph.fit({ animation: true });
        }, 100);

        this.showToast(`Chaîne: ${chain.id || chainIndex + 1} (${chain.steps.length} étapes)`);
    },

    showAllDashboardCausalChains() {
        if (!this.dashboardN4LParse?.causal_chains || this.dashboardN4LParse.causal_chains.length === 0) {
            this.showToast('Aucune chaîne causale', 'warning');
            return;
        }

        const chains = this.dashboardN4LParse.causal_chains;
        const allNodes = [];
        const allEdges = [];

        const chainColors = [
            { bg: '#f97316', border: '#ea580c' },
            { bg: '#3b82f6', border: '#2563eb' },
            { bg: '#10b981', border: '#059669' },
            { bg: '#8b5cf6', border: '#7c3aed' },
            { bg: '#ec4899', border: '#db2777' },
            { bg: '#06b6d4', border: '#0891b2' }
        ];

        chains.forEach((chain, chainIndex) => {
            const color = chainColors[chainIndex % chainColors.length];
            const yOffset = chainIndex * 150;

            chain.steps.forEach((step, stepIndex) => {
                allNodes.push({
                    id: `chain_${chainIndex}_step_${stepIndex}`,
                    label: step.item,
                    color: { background: color.bg, border: color.border },
                    borderWidth: 3,
                    font: { size: 13, color: '#fff' },
                    shape: 'box',
                    margin: 8,
                    x: stepIndex * 200,
                    y: yOffset
                });

                if (stepIndex < chain.steps.length - 1) {
                    allEdges.push({
                        id: `chain_${chainIndex}_edge_${stepIndex}`,
                        from: `chain_${chainIndex}_step_${stepIndex}`,
                        to: `chain_${chainIndex}_step_${stepIndex + 1}`,
                        label: step.relation || '',
                        arrows: 'to',
                        color: { color: color.bg },
                        width: 3,
                        font: { size: 11, color: '#666' }
                    });
                }
            });

            // Chain label
            allNodes.push({
                id: `chain_${chainIndex}_label`,
                label: chain.id || `Chaîne ${chainIndex + 1}`,
                color: { background: '#1e3a5f', border: '#0f1f33' },
                font: { size: 12, color: '#fff', bold: true },
                shape: 'box',
                margin: 6,
                x: -150,
                y: yOffset
            });

            allEdges.push({
                id: `chain_${chainIndex}_label_edge`,
                from: `chain_${chainIndex}_label`,
                to: `chain_${chainIndex}_step_0`,
                arrows: 'to',
                color: { color: '#1e3a5f', opacity: 0.5 },
                width: 1,
                dashes: true
            });
        });

        this.graphNodes.clear();
        this.graphEdges.clear();
        this.graphNodes.add(allNodes);
        this.graphEdges.add(allEdges);

        this.dashboardShowingSpecialView = true;
        this.dashboardActiveFilter = 'chaînes causales';

        const restoreBtn = document.getElementById('dashboard-restore-graph-btn');
        if (restoreBtn) restoreBtn.style.display = 'flex';

        setTimeout(() => {
            if (this.graph) this.graph.fit({ animation: true });
        }, 100);

        this.showToast(`${chains.length} chaînes causales`);
    },

    showDashboardHypotheses() {
        if (!this.dashboardN4LParse?.aliases) {
            this.showToast('Aucune hypothèse', 'warning');
            return;
        }

        const hypAliases = Object.entries(this.dashboardN4LParse.aliases)
            .filter(([key]) => key.startsWith('hyp'));

        if (hypAliases.length === 0) {
            this.showToast('Aucune hypothèse trouvée', 'warning');
            return;
        }

        const allNodes = [];
        const allEdges = [];

        const getConfidenceColor = (content) => {
            const match = content.match(/(\d+)%/);
            if (match) {
                const conf = parseInt(match[1]);
                if (conf >= 70) return { bg: '#10b981', border: '#059669' };
                if (conf >= 40) return { bg: '#f59e0b', border: '#d97706' };
                return { bg: '#ef4444', border: '#dc2626' };
            }
            return { bg: '#6b7280', border: '#4b5563' };
        };

        const getHypName = (content) => {
            const parenIdx = content.indexOf('(');
            return parenIdx > 0 ? content.substring(0, parenIdx).trim() : content.trim();
        };

        allNodes.push({
            id: 'hyp_center',
            label: 'HYPOTHÈSES',
            color: { background: '#1e3a5f', border: '#0f1f33' },
            font: { size: 16, color: '#fff', bold: true },
            shape: 'box',
            margin: 12,
            x: 0,
            y: 0
        });

        hypAliases.forEach(([aliasKey, values], i) => {
            const content = values[0] || aliasKey;
            const color = getConfidenceColor(content);
            const name = getHypName(content);
            const angle = (2 * Math.PI * i) / hypAliases.length;
            const radius = 200;

            allNodes.push({
                id: `hyp_${i}`,
                label: name,
                color: { background: color.bg, border: color.border },
                borderWidth: 3,
                font: { size: 12, color: '#fff' },
                shape: 'box',
                margin: 10,
                x: Math.cos(angle) * radius,
                y: Math.sin(angle) * radius
            });

            allEdges.push({
                id: `hyp_edge_${i}`,
                from: 'hyp_center',
                to: `hyp_${i}`,
                color: { color: color.bg, opacity: 0.7 },
                width: 2,
                dashes: true
            });
        });

        this.graphNodes.clear();
        this.graphEdges.clear();
        this.graphNodes.add(allNodes);
        this.graphEdges.add(allEdges);

        this.dashboardShowingSpecialView = true;
        this.dashboardActiveFilter = 'hypothèses';

        const restoreBtn = document.getElementById('dashboard-restore-graph-btn');
        if (restoreBtn) restoreBtn.style.display = 'flex';

        setTimeout(() => {
            if (this.graph) this.graph.fit({ animation: true });
        }, 100);

        this.showToast(`${hypAliases.length} hypothèses`);
    },

    showDashboardSequence(seqIndex) {
        if (!this.dashboardN4LParse?.sequences) return;

        const seq = this.dashboardN4LParse.sequences[seqIndex];
        if (!seq || seq.length === 0) return;

        // Create sequence nodes
        const seqNodes = seq.map((item, i) => ({
            id: `seq_step_${i}`,
            label: item,
            color: { background: '#3b82f6', border: '#2563eb' },
            borderWidth: 3,
            font: { size: 13, color: '#fff' },
            shape: 'box',
            margin: 10,
            x: i * 180,
            y: 0
        }));

        // Create sequence edges
        const seqEdges = [];
        for (let i = 0; i < seq.length - 1; i++) {
            seqEdges.push({
                id: `seq_edge_${i}`,
                from: `seq_step_${i}`,
                to: `seq_step_${i + 1}`,
                arrows: 'to',
                color: { color: '#3b82f6' },
                width: 3
            });
        }

        this.graphNodes.clear();
        this.graphEdges.clear();
        this.graphNodes.add(seqNodes);
        this.graphEdges.add(seqEdges);

        this.dashboardShowingSpecialView = true;

        const restoreBtn = document.getElementById('dashboard-restore-graph-btn');
        if (restoreBtn) restoreBtn.style.display = 'flex';

        setTimeout(() => {
            if (this.graph) this.graph.fit({ animation: true });
        }, 100);

        this.showToast(`Séquence ${seqIndex + 1}: ${seq.length} étapes`);
    },

    // Get icon for context (same as N4LModule)
    getContextIcon(context) {
        const icons = {
            'victimes': 'person_off',
            'suspects': 'person_search',
            'témoins': 'record_voice_over',
            'lieux': 'place',
            'objets': 'category',
            'preuves': 'verified',
            'indices': 'search',
            'chronologie': 'schedule',
            'hypothèses': 'lightbulb',
            'pistes': 'explore',
            'réseau': 'share',
            'relations': 'people',
            'chaînes': 'route',
            'todo': 'checklist',
            'notes': 'note'
        };

        const lowerCtx = context.toLowerCase();
        for (const [key, icon] of Object.entries(icons)) {
            if (lowerCtx.includes(key)) return icon;
        }
        return 'label';
    },

    // ============================================
    // Graph Context Menu
    // ============================================
    handleGraphRightClick(params) {
        const nodeId = this.graph.getNodeAt(params.pointer.DOM);
        const menu = document.getElementById('graph-context-menu');

        if (!nodeId) {
            this.hideContextMenu();
            return;
        }

        this.contextMenuNodeId = nodeId;
        this.contextMenuGraphType = 'main';

        const x = params.event.clientX || params.pointer.DOM.x;
        const y = params.event.clientY || params.pointer.DOM.y;

        menu.style.left = `${x}px`;
        menu.style.top = `${y}px`;
        menu.classList.remove('hidden');

        // Keep menu in viewport
        const rect = menu.getBoundingClientRect();
        if (rect.right > window.innerWidth) {
            menu.style.left = `${x - rect.width}px`;
        }
        if (rect.bottom > window.innerHeight) {
            menu.style.top = `${y - rect.height}px`;
        }
    },

    hideContextMenu() {
        const menu = document.getElementById('graph-context-menu');
        if (menu) menu.classList.add('hidden');
    },

    // ============================================
    // Graph Analysis
    // ============================================
    async analyzeGraph() {
        if (!this.currentCase) {
            alert('Sélectionnez une affaire d\'abord');
            return;
        }

        this.setAnalysisContext('graph_analysis', `Analyse du graphe - ${this.currentCase.name}`, 'Analyse complète des relations');

        const analysisContent = document.getElementById('analysis-content');
        const analysisModal = document.getElementById('analysis-modal');
        const modalTitle = document.getElementById('analysis-modal-title');
        if (modalTitle) modalTitle.textContent = 'Analyse IA - Graphe';

        const noteBtn = document.getElementById('btn-save-to-notebook');
        if (noteBtn) noteBtn.style.display = '';

        analysisContent.innerHTML = '<span class="streaming-cursor">▊</span>';
        analysisModal.classList.add('active');

        await this.streamAIResponse(
            '/api/analyze/stream',
            { case_id: this.currentCase.id },
            analysisContent
        );
    },

    // ============================================
    // Path Highlighting
    // ============================================
    highlightPathInGraph(fromId, toId) {
        if (!this.graph || !this.graphNodes) return;

        try {
            // Trouver les noms des entités pour les IDs (N4L utilise les noms comme IDs de nœuds)
            const normalizeId = (id) => id ? id.replace(/-/g, '_') : '';
            const entityMap = {};
            (this.currentCase?.entities || []).forEach(e => {
                entityMap[e.id] = e.name;
                entityMap[normalizeId(e.id)] = e.name;
            });

            const fromName = entityMap[fromId] || entityMap[normalizeId(fromId)] || fromId;
            const toName = entityMap[toId] || entityMap[normalizeId(toId)] || toId;

            // Trouver les nœuds correspondants dans le graphe
            const allNodeIds = this.graphNodes.getIds();
            let fromNodeId = null;
            let toNodeId = null;

            for (const nodeId of allNodeIds) {
                const node = this.graphNodes.get(nodeId);
                if (node && (nodeId === fromName || node.label === fromName)) {
                    fromNodeId = nodeId;
                }
                if (node && (nodeId === toName || node.label === toName)) {
                    toNodeId = nodeId;
                }
            }

            if (!fromNodeId && !toNodeId) return;

            const updates = [];
            if (fromNodeId) {
                updates.push({
                    id: fromNodeId,
                    borderWidth: 4,
                    color: { border: '#dc2626', background: '#fee2e2' }
                });
            }
            if (toNodeId) {
                updates.push({
                    id: toNodeId,
                    borderWidth: 4,
                    color: { border: '#16a34a', background: '#dcfce7' }
                });
            }

            if (updates.length > 0) {
                this.graphNodes.update(updates);
            }

            const fitNodes = [fromNodeId, toNodeId].filter(Boolean);
            if (fitNodes.length > 0) {
                this.graph.fit({
                    nodes: fitNodes,
                    animation: true
                });
            }
        } catch (error) {
            console.warn('[highlightPathInGraph] Error:', error);
        }
    },

    // ============================================
    // Expansion Cone
    // ============================================
    showExpansionCone(nodeId, depth) {
        if (!this.graph || !this.graphNodes || !this.graphEdges) return;

        const coneNodes = new Set([nodeId]);
        const coneEdges = new Set();
        let currentLevel = new Set([nodeId]);

        for (let d = 0; d < depth; d++) {
            const nextLevel = new Set();
            this.graphEdges.forEach(edge => {
                if (currentLevel.has(edge.from)) {
                    nextLevel.add(edge.to);
                    coneNodes.add(edge.to);
                    coneEdges.add(edge.id);
                }
                if (currentLevel.has(edge.to)) {
                    nextLevel.add(edge.from);
                    coneNodes.add(edge.from);
                    coneEdges.add(edge.id);
                }
            });
            currentLevel = nextLevel;
        }

        // Update nodes
        const nodeUpdates = this.graphNodes.map(node => {
            const inCone = coneNodes.has(node.id);
            const isCenter = node.id === nodeId;

            if (isCenter) {
                return {
                    id: node.id,
                    color: { background: '#1e3a5f', border: '#152a45' },
                    borderWidth: 4,
                    font: { color: '#ffffff', size: 14 }
                };
            } else if (inCone) {
                return {
                    id: node.id,
                    color: { background: '#3b82f6', border: '#1e3a5f' },
                    borderWidth: 3,
                    opacity: 1
                };
            } else {
                return {
                    id: node.id,
                    color: { background: '#e2e8f0', border: '#cbd5e0' },
                    opacity: 0.3,
                    borderWidth: 1
                };
            }
        });

        this.graphNodes.update(nodeUpdates);

        // Update edges
        const edgeUpdates = this.graphEdges.map(edge => ({
            id: edge.id,
            color: coneEdges.has(edge.id)
                ? { color: '#1e3a5f', opacity: 1 }
                : { color: '#cbd5e0', opacity: 0.2 },
            width: coneEdges.has(edge.id) ? 2 : 1
        }));

        this.graphEdges.update(edgeUpdates);

        this.graph.focus(nodeId, { scale: 1.2, animation: true });
        this.showToast(`Cône d'expansion: ${coneNodes.size} noeuds, profondeur ${depth}`);
    },

    showExpansionConeModal(nodeId) {
        const node = this.graphNodes?.get(nodeId);
        if (!node) return;

        const content = `
            <div class="modal-explanation">
                <span class="material-icons">info</span>
                <p><strong>Cône d'expansion</strong> - Visualisez tous les noeuds connectés à "${node.label}"
                jusqu'à une certaine profondeur.</p>
            </div>
            <div class="form-group">
                <label class="form-label">Profondeur :</label>
                <select id="cone-depth" class="form-select">
                    <option value="1">1 niveau (voisins directs)</option>
                    <option value="2" selected>2 niveaux</option>
                    <option value="3">3 niveaux</option>
                    <option value="4">4 niveaux (complet)</option>
                </select>
            </div>
        `;

        this.showModal('Explorer le Voisinage', content, () => {
            const depth = parseInt(document.getElementById('cone-depth').value);
            this.showExpansionCone(nodeId, depth);
        });
    },

    // ============================================
    // Entity Icon Helper
    // ============================================
    getEntityIcon(type) {
        const icons = {
            personne: 'person',
            lieu: 'place',
            objet: 'inventory_2',
            evenement: 'event',
            organisation: 'business',
            document: 'description'
        };
        return icons[type] || 'category';
    },

    // ============================================
    // Evidence Type Icon Helper
    // ============================================
    getEvidenceTypeIcon(type) {
        const icons = {
            physique: 'fingerprint',
            testimoniale: 'record_voice_over',
            documentaire: 'description',
            numerique: 'computer',
            forensique: 'biotech'
        };
        return icons[type] || 'find_in_page';
    },

    // ============================================
    // Reliability Class Helper
    // ============================================
    getReliabilityClass(reliability) {
        if (reliability >= 8) return 'high';
        if (reliability >= 5) return 'medium';
        return 'low';
    },

    // ============================================
    // Context Menu Actions
    // ============================================
    setupContextMenuActions() {
        const menu = document.getElementById('graph-context-menu');
        if (!menu) return;

        // Remove old listeners by cloning and replacing
        const items = menu.querySelectorAll('.context-menu-item');
        items.forEach(item => {
            const newItem = item.cloneNode(true);
            item.parentNode.replaceChild(newItem, item);
        });

        // Add fresh listeners
        menu.querySelectorAll('.context-menu-item').forEach(item => {
            item.addEventListener('click', (e) => {
                e.preventDefault();
                e.stopPropagation();
                const action = item.dataset.action;
                console.log('Context menu action:', action); // Debug log
                if (action) {
                    this.handleContextMenuAction(action);
                }
                this.hideContextMenu();
            });
        });
    },

    handleContextMenuAction(action) {
        if (!this.contextMenuNodeId) return;

        const isN4L = this.contextMenuGraphType === 'n4l';

        switch (action) {
            case 'expansion-cone':
                if (isN4L) {
                    this.showN4LExpansionConeModal(this.contextMenuNodeId);
                } else {
                    this.showExpansionConeModal(this.contextMenuNodeId);
                }
                break;
            case 'analyze-cone':
                if (isN4L) {
                    this.analyzeN4LExpansionCone(this.contextMenuNodeId);
                } else {
                    this.analyzeExpansionCone(this.contextMenuNodeId);
                }
                break;
            case 'find-paths':
                if (isN4L) {
                    this.showFindN4LPathsFromNode(this.contextMenuNodeId);
                } else {
                    this.showFindPathsFromNode(this.contextMenuNodeId);
                }
                break;
            case 'focus-node':
                if (isN4L) {
                    this.focusN4LGraphNode(this.contextMenuNodeId);
                } else {
                    this.focusGraphNode(this.contextMenuNodeId);
                }
                break;
            case 'show-details':
                this.showNodeDetails(this.contextMenuNodeId, isN4L);
                break;
            case 'exclude-node':
                this.excludeNodeFromSearch(this.contextMenuNodeId, isN4L);
                break;
            // SSTorytime actions
            case 'orbits-analysis':
                this.launchOrbitsAnalysis(this.contextMenuNodeId, isN4L);
                break;
            case 'dirac-from-node':
                this.launchDiracFromNode(this.contextMenuNodeId, isN4L);
                break;
            case 'contrawave-start':
                this.launchContrawaveFromNode(this.contextMenuNodeId, isN4L);
                break;
            case 'betweenness-highlight':
                this.showNodeBetweenness(this.contextMenuNodeId, isN4L);
                break;
        }
        this.contextMenuGraphType = null;
    },

    // ============================================
    // SSTorytime Quick Actions
    // ============================================
    navigateToGraphAnalysis() {
        // First, navigate to the graph-analysis main view
        const navBtn = document.querySelector('.nav-btn[data-view="graph-analysis"]');
        if (navBtn) {
            navBtn.click();
        }
    },

    launchOrbitsAnalysis(nodeId, isN4L = false) {
        const nodes = isN4L ? this.n4lGraphNodes : this.graphNodes;
        const node = nodes?.get(nodeId);
        const nodeLabel = node?.label || nodeId;

        // Navigate to Graph Analysis view first
        this.navigateToGraphAnalysis();

        setTimeout(() => {
            // Switch to SSTorytime tab in Graph Analysis panel
            document.querySelector('.graph-analysis-tab[data-tab="sstorytime"]')?.click();

            setTimeout(() => {
                // Scroll to Orbits section
                const orbitsSection = document.getElementById('section-orbits');
                if (orbitsSection) {
                    orbitsSection.scrollIntoView({ behavior: 'smooth', block: 'start' });
                }

                // Set the orbit center node - use label for N4L
                const select = document.getElementById('orbit-center-node');
                if (select) {
                    // Try to find matching option by label or id
                    const options = Array.from(select.options);
                    const match = options.find(o => o.value === nodeId || o.textContent.includes(nodeLabel));
                    if (match) {
                        select.value = match.value;
                    } else {
                        // For N4L, add a temporary option
                        if (isN4L) {
                            const opt = document.createElement('option');
                            opt.value = nodeLabel;
                            opt.textContent = `[N4L] ${nodeLabel}`;
                            select.appendChild(opt);
                            select.value = nodeLabel;
                        }
                    }
                }

                // Execute orbits analysis
                setTimeout(() => {
                    if (typeof this.executeOrbitsAnalysis === 'function') {
                        this.executeOrbitsAnalysis();
                    } else {
                        document.getElementById('btn-orbits')?.click();
                    }
                }, 200);
                this.showToast(`Analyse des orbites: ${nodeLabel}`, 'info');
            }, 100);
        }, 100);
    },

    launchDiracFromNode(nodeId, isN4L = false) {
        // Get node label
        const nodes = isN4L ? this.n4lGraphNodes : this.graphNodes;
        const node = nodes?.get(nodeId);
        const nodeLabel = node?.label || nodeId;
        const shortName = nodeLabel.split(' ')[0]; // First word

        // Navigate to Graph Analysis view first
        this.navigateToGraphAnalysis();

        setTimeout(() => {
            // Switch to SSTorytime tab
            document.querySelector('.graph-analysis-tab[data-tab="sstorytime"]')?.click();

            setTimeout(() => {
                // Scroll to Dirac section
                const diracSection = document.getElementById('section-dirac');
                if (diracSection) {
                    diracSection.scrollIntoView({ behavior: 'smooth', block: 'start' });
                }

                const diracInput = document.getElementById('dirac-query');
                if (diracInput) {
                    diracInput.value = `<${shortName}|`;
                    diracInput.focus();
                    diracInput.setSelectionRange(diracInput.value.length, diracInput.value.length);
                }
                this.showToast(`Complétez la requête: <${shortName}|cible>`, 'info');
            }, 100);
        }, 100);
    },

    launchContrawaveFromNode(nodeId, isN4L = false) {
        const nodes = isN4L ? this.n4lGraphNodes : this.graphNodes;
        const node = nodes?.get(nodeId);
        const nodeLabel = node?.label || nodeId;

        // Navigate to Graph Analysis view first
        this.navigateToGraphAnalysis();

        setTimeout(() => {
            // Switch to SSTorytime tab
            document.querySelector('.graph-analysis-tab[data-tab="sstorytime"]')?.click();

            setTimeout(() => {
                // Scroll to Contrawave section
                const contrawaveSection = document.getElementById('section-contrawave');
                if (contrawaveSection) {
                    contrawaveSection.scrollIntoView({ behavior: 'smooth', block: 'start' });
                }

                const startSelect = document.getElementById('contrawave-start-nodes');
                if (startSelect) {
                    // Try to find matching option
                    let found = false;
                    Array.from(startSelect.options).forEach(opt => {
                        if (opt.value === nodeId || opt.textContent.includes(nodeLabel)) {
                            opt.selected = true;
                            found = true;
                        } else {
                            opt.selected = false;
                        }
                    });

                    // For N4L, add a temporary option if not found
                    if (!found && isN4L) {
                        const opt = document.createElement('option');
                        opt.value = nodeLabel;
                        opt.textContent = `[N4L] ${nodeLabel}`;
                        opt.selected = true;
                        startSelect.appendChild(opt);
                    }
                }
                this.showToast(`Contrawave: départ depuis ${nodeLabel}`, 'info');
            }, 100);
        }, 100);
    },

    async showNodeBetweenness(nodeId, isN4L = false) {
        if (isN4L) {
            // For N4L, calculate betweenness locally
            this.showN4LNodeBetweenness(nodeId);
            return;
        }

        if (!this.currentCase) return;

        try {
            const response = await fetch('/api/graph/betweenness-centrality', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    case_id: this.currentCase.id,
                    top_k: 20
                })
            });

            if (!response.ok) throw new Error('Erreur de calcul');

            const data = await response.json();

            // Find this node in the ranking
            const nodeRank = data.ranking?.findIndex(r => r.node_id === nodeId);
            const nodeScore = data.ranking?.find(r => r.node_id === nodeId);

            if (nodeScore) {
                const rank = nodeRank + 1;
                const percentage = (nodeScore.score * 100).toFixed(1);
                this.showToast(
                    `Centralité: ${percentage}% (rang #${rank}/${data.ranking.length})`,
                    'success'
                );

                // Highlight this node in the graph
                this.focusAndHighlightNode(nodeId, '#9333ea');
            } else {
                this.showToast('Ce nœud a une centralité faible (hors top 20)', 'info');
            }
        } catch (error) {
            console.error('Erreur betweenness:', error);
            this.showToast('Erreur lors du calcul de centralité', 'error');
        }
    },

    showN4LNodeBetweenness(nodeId) {
        // Calculate simple degree centrality for N4L graph
        const edges = this.n4lGraphEdges?.get() || [];
        const nodes = this.n4lGraphNodes?.get() || [];
        const node = nodes.find(n => n.id === nodeId);

        if (!node) {
            this.showToast('Nœud non trouvé', 'warning');
            return;
        }

        // Count connections
        const inDegree = edges.filter(e => e.to === nodeId).length;
        const outDegree = edges.filter(e => e.from === nodeId).length;
        const totalDegree = inDegree + outDegree;

        // Calculate relative centrality
        const maxPossible = (nodes.length - 1) * 2;
        const centrality = maxPossible > 0 ? (totalDegree / maxPossible * 100).toFixed(1) : 0;

        // Rank among all nodes
        const degrees = nodes.map(n => {
            const inD = edges.filter(e => e.to === n.id).length;
            const outD = edges.filter(e => e.from === n.id).length;
            return { id: n.id, degree: inD + outD };
        }).sort((a, b) => b.degree - a.degree);

        const rank = degrees.findIndex(d => d.id === nodeId) + 1;

        this.showToast(
            `Centralité: ${centrality}% (${totalDegree} connexions, rang #${rank}/${nodes.length})`,
            'success'
        );

        // Highlight the node
        this.focusAndHighlightNode(nodeId, '#9333ea', true);
    },

    focusAndHighlightNode(nodeId, color = '#9333ea', isN4L = false) {
        const graph = isN4L ? this.n4lGraph : this.graph;
        const graphNodes = isN4L ? this.n4lGraphNodes : this.graphNodes;

        if (!graph || !graphNodes) return;

        // Focus on the node
        graph.focus(nodeId, {
            scale: 1.5,
            animation: { duration: 500, easingFunction: 'easeInOutQuad' }
        });

        // Temporarily change node color
        const node = graphNodes.get(nodeId);
        if (node) {
            const originalColor = node.color;
            graphNodes.update({
                id: nodeId,
                color: { background: color, border: color }
            });

            // Reset after 3 seconds
            setTimeout(() => {
                graphNodes.update({
                    id: nodeId,
                    color: originalColor
                });
            }, 3000);
        }
    },

    // ============================================
    // Node Details
    // ============================================
    showNodeDetails(nodeId, isN4L = false) {
        const nodes = isN4L ? this.n4lGraphNodes : this.graphNodes;
        const node = nodes?.get(nodeId);
        if (!node) return;

        if (isN4L) {
            // For N4L graph, show node info in a modal
            this.showN4LNodeDetails(nodeId, node);
        } else {
            // Chercher l'entité par ID ou par nom (le graphe N4L utilise les noms comme IDs)
            const normalizeId = (id) => id ? id.replace(/-/g, '_') : '';
            let entity = this.currentCase?.entities?.find(e =>
                e.id === nodeId ||
                normalizeId(e.id) === normalizeId(nodeId) ||
                e.name === nodeId ||
                e.name === node.label
            );

            if (entity) {
                this.showEntityDetails(entity);
            } else {
                // Afficher les détails du nœud N4L même si ce n'est pas une entité connue
                this.showN4LNodeDetails(nodeId, node);
            }
        }
    },

    showN4LNodeDetails(nodeId, node) {
        // Get edges connected to this node (use graphEdges if n4lGraphEdges not available)
        const edges = this.graphEdges?.get() || this.n4lGraphEdges?.get() || [];
        const connectedEdges = edges.filter(e => e.from === nodeId || e.to === nodeId);

        const incomingEdges = connectedEdges.filter(e => e.to === nodeId);
        const outgoingEdges = connectedEdges.filter(e => e.from === nodeId);

        const allNodes = this.graphNodes?.get() || this.n4lGraphNodes?.get() || [];
        const nodeMap = {};
        allNodes.forEach(n => { nodeMap[n.id] = n; });

        const incomingHtml = incomingEdges.length > 0
            ? incomingEdges.map(e => {
                const fromNode = nodeMap[e.from];
                return `<div class="relation-item"><span class="relation-target">${fromNode?.label || e.from}</span> <span class="relation-label">→ ${e.label || ''}</span></div>`;
            }).join('')
            : '<div class="empty">Aucune</div>';

        const outgoingHtml = outgoingEdges.length > 0
            ? outgoingEdges.map(e => {
                const toNode = nodeMap[e.to];
                return `<div class="relation-item"><span class="relation-label">${e.label || ''} →</span> <span class="relation-target">${toNode?.label || e.to}</span></div>`;
            }).join('')
            : '<div class="empty">Aucune</div>';

        const content = `
            <div class="entity-details">
                <div class="detail-section">
                    <h4><span class="material-icons">hub</span> Nœud N4L</h4>
                    <div class="detail-row"><span class="detail-label">Label:</span><span class="detail-value">${node.label || nodeId}</span></div>
                    <div class="detail-row"><span class="detail-label">ID:</span><span class="detail-value">${nodeId}</span></div>
                    ${node.title ? `<div class="detail-row"><span class="detail-label">Info:</span><span class="detail-value">${node.title}</span></div>` : ''}
                </div>
                <div class="detail-section">
                    <h4><span class="material-icons">arrow_back</span> Relations entrantes (${incomingEdges.length})</h4>
                    ${incomingHtml}
                </div>
                <div class="detail-section">
                    <h4><span class="material-icons">arrow_forward</span> Relations sortantes (${outgoingEdges.length})</h4>
                    ${outgoingHtml}
                </div>
            </div>
        `;

        this.showModal('Détails du nœud', content);
    },

    showEntityDetails(entity) {
        if (!entity) return;

        const attributes = entity.attributes || {};
        const attributesHtml = Object.keys(attributes).length > 0
            ? Object.entries(attributes).map(([key, value]) => `
                <div class="detail-row">
                    <span class="detail-label">${key}:</span>
                    <span class="detail-value">${value}</span>
                </div>
            `).join('')
            : '<div class="empty">Aucun attribut</div>';

        // Créer une map des entités pour résoudre les noms
        const entityMap = {};
        const entityNameToEntity = {};
        (this.currentCase?.entities || []).forEach(e => {
            entityMap[e.id] = e;
            entityNameToEntity[e.name] = e;
        });

        // Chercher les relations dans entity.relations ET dans les arêtes du graphe
        const relations = [];

        // 1. Relations de l'entité
        if (entity.relations && entity.relations.length > 0) {
            entity.relations.forEach(rel => {
                const targetEntity = entityMap[rel.to_id] || entityNameToEntity[rel.to_id];
                relations.push({
                    label: rel.label || 'lié à',
                    target: targetEntity ? targetEntity.name : rel.to_id,
                    direction: 'outgoing'
                });
            });
        }

        // 2. Relations depuis les arêtes du graphe (le graphe N4L utilise les noms comme IDs)
        const edges = this.graphEdges?.get() || [];
        const entityNodeId = entity.name; // Le graphe N4L utilise les noms comme IDs

        edges.forEach(edge => {
            if (edge.from === entityNodeId) {
                // Relation sortante
                const targetEntity = entityNameToEntity[edge.to];
                if (!relations.find(r => r.target === edge.to && r.label === edge.label)) {
                    relations.push({
                        label: edge.label || 'lié à',
                        target: targetEntity ? targetEntity.name : edge.to,
                        direction: 'outgoing'
                    });
                }
            } else if (edge.to === entityNodeId) {
                // Relation entrante
                const sourceEntity = entityNameToEntity[edge.from];
                if (!relations.find(r => r.target === edge.from && r.label === edge.label)) {
                    relations.push({
                        label: edge.label || 'lié à',
                        target: sourceEntity ? sourceEntity.name : edge.from,
                        direction: 'incoming'
                    });
                }
            }
        });

        const relationsHtml = relations.length > 0
            ? relations.map(rel => {
                const icon = rel.direction === 'incoming' ? '←' : '→';
                return `
                    <div class="relation-item">
                        <span class="relation-direction">${icon}</span>
                        <span class="relation-label">${rel.label}</span>
                        <span class="relation-target">${rel.target}</span>
                    </div>
                `;
            }).join('')
            : '<div class="empty">Aucune relation</div>';

        const content = `
            <div class="entity-details">
                <div class="detail-section">
                    <h4><span class="material-icons">person</span> Informations</h4>
                    <div class="detail-row"><span class="detail-label">Type:</span><span class="detail-value">${entity.type}</span></div>
                    <div class="detail-row"><span class="detail-label">Rôle:</span><span class="detail-value">${entity.role}</span></div>
                    ${entity.description ? `<div class="detail-row"><span class="detail-label">Description:</span><span class="detail-value">${entity.description}</span></div>` : ''}
                </div>
                <div class="detail-section">
                    <h4><span class="material-icons">list</span> Attributs</h4>
                    ${attributesHtml}
                </div>
                <div class="detail-section">
                    <h4><span class="material-icons">hub</span> Relations (${relations.length})</h4>
                    ${relationsHtml}
                </div>
            </div>
        `;

        this.showModal(entity.name, content, null, false);
    },

    // ============================================
    // Path Finding
    // ============================================
    showFindPathsFromNode(fromNodeId) {
        const fromNode = this.graphNodes?.get(fromNodeId);
        if (!fromNode) return;

        const otherNodes = [];
        this.graphNodes?.forEach(node => {
            if (node.id !== fromNodeId) {
                otherNodes.push(`<option value="${node.id}">${node.label}</option>`);
            }
        });

        const content = `
            <div class="modal-explanation">
                <span class="material-icons">info</span>
                <p><strong>Trouver les chemins</strong> - Découvrez tous les chemins possibles entre "${fromNode.label}" et une autre entité.</p>
            </div>
            <div class="form-group">
                <label class="form-label">De: ${fromNode.label}</label>
            </div>
            <div class="form-group">
                <label class="form-label">Vers:</label>
                <select id="path-to-node" class="form-select">
                    <option value="">-- Sélectionner une destination --</option>
                    ${otherNodes.join('')}
                </select>
            </div>
            <div class="form-group">
                <label class="form-label">Profondeur max:</label>
                <select id="path-max-depth" class="form-select">
                    <option value="3">3 niveaux</option>
                    <option value="4" selected>4 niveaux</option>
                    <option value="5">5 niveaux</option>
                </select>
            </div>
        `;

        this.showModal('Trouver les Chemins', content, async () => {
            const toNodeId = document.getElementById('path-to-node').value;
            const maxDepth = parseInt(document.getElementById('path-max-depth').value);

            if (!toNodeId) {
                alert('Veuillez sélectionner une destination');
                return;
            }

            this.closeModal();
            await this.findAndDisplayPaths(fromNodeId, toNodeId, maxDepth);
        });
    },

    async findAndDisplayPaths(fromId, toId, maxDepth) {
        const paths = this.findAllPaths(fromId, toId, maxDepth);
        const fromNode = this.graphNodes?.get(fromId);
        const toNode = this.graphNodes?.get(toId);

        if (paths.length === 0) {
            this.showToast(`Aucun chemin trouvé entre ${fromNode?.label} et ${toNode?.label}`);
            return;
        }

        const pathsHtml = paths.map((path, index) => {
            const pathNodes = path.map(id => {
                const n = this.graphNodes?.get(id);
                return n ? n.label : id;
            }).join(' <span class="path-arrow">→</span> ');

            return `
                <div class="path-item" data-path-index="${index}">
                    <div class="path-nodes">${pathNodes}</div>
                    <div class="path-actions">
                        <button class="btn btn-sm btn-secondary btn-icon" onclick="app.highlightPath(${index})" data-tooltip="Afficher ce chemin sur le graphe">
                            <span class="material-icons">visibility</span>
                        </button>
                        <button class="btn btn-sm btn-secondary btn-icon" onclick="app.analyzePath(${index})" data-tooltip="Analyser ce chemin avec l'IA">
                            <span class="material-icons">psychology</span>
                        </button>
                    </div>
                </div>
            `;
        }).join('');

        this.discoveredPaths = paths;

        this.showModal(`${paths.length} Chemin(s) Trouvé(s)`, `
            <div class="modal-explanation">
                <span class="material-icons">info</span>
                <p>Chemins entre <strong>${fromNode?.label}</strong> et <strong>${toNode?.label}</strong>.</p>
            </div>
            <div class="paths-list">${pathsHtml}</div>
        `);
    },

    findAllPaths(fromId, toId, maxDepth) {
        const paths = [];
        const visited = new Set();

        const dfs = (current, target, path, depth) => {
            if (depth > maxDepth) return;
            if (current === target) {
                paths.push([...path]);
                return;
            }

            visited.add(current);

            this.graphEdges?.forEach(edge => {
                let next = null;
                if (edge.from === current && !visited.has(edge.to)) {
                    next = edge.to;
                } else if (edge.to === current && !visited.has(edge.from)) {
                    next = edge.from;
                }

                if (next) {
                    path.push(next);
                    dfs(next, target, path, depth + 1);
                    path.pop();
                }
            });

            visited.delete(current);
        };

        dfs(fromId, toId, [fromId], 0);
        return paths;
    },

    highlightPath(pathIndex) {
        if (!this.discoveredPaths || !this.discoveredPaths[pathIndex]) return;

        const path = this.discoveredPaths[pathIndex];
        const pathSet = new Set(path);

        const pathEdges = new Set();
        for (let i = 0; i < path.length - 1; i++) {
            this.graphEdges?.forEach(edge => {
                if ((edge.from === path[i] && edge.to === path[i + 1]) ||
                    (edge.to === path[i] && edge.from === path[i + 1])) {
                    pathEdges.add(edge.id);
                }
            });
        }

        const nodeUpdates = this.graphNodes?.get().map(node => {
            const inPath = pathSet.has(node.id);
            if (inPath) {
                return {
                    id: node.id,
                    color: { background: '#f59e0b', border: '#d97706' },
                    borderWidth: 4,
                    font: { color: '#1a1a2e', size: 14 }
                };
            } else {
                return {
                    id: node.id,
                    color: { background: '#e2e8f0', border: '#cbd5e0' },
                    opacity: 0.3,
                    borderWidth: 1,
                    font: { color: '#a0aec0', size: 10 }
                };
            }
        });

        this.graphNodes?.update(nodeUpdates);

        const edgeUpdates = this.graphEdges?.get().map(edge => ({
            id: edge.id,
            color: pathEdges.has(edge.id) ? { color: '#f59e0b', opacity: 1 } : { color: '#cbd5e0', opacity: 0.2 },
            width: pathEdges.has(edge.id) ? 3 : 1
        }));

        this.graphEdges?.update(edgeUpdates);
        this.showToast(`Chemin ${pathIndex + 1} affiché`);
    },

    async analyzePath(pathIndex) {
        if (!this.discoveredPaths || !this.discoveredPaths[pathIndex]) return;

        const path = this.discoveredPaths[pathIndex];
        const pathLabels = path.map(id => {
            const n = this.graphNodes?.get(id);
            return n ? n.label : id;
        });

        this.setAnalysisContext('path_analysis', 'Analyse du Chemin', `Chemin: ${pathLabels.join(' → ')}`);

        const analysisModal = document.getElementById('analysis-modal');
        const analysisContent = document.getElementById('analysis-content');
        const modalTitle = document.getElementById('analysis-modal-title');
        if (modalTitle) modalTitle.textContent = 'Analyse du Chemin';

        const noteBtn = document.getElementById('btn-save-to-notebook');
        if (noteBtn) noteBtn.style.display = '';

        analysisContent.innerHTML = `
            <div class="modal-explanation">
                <span class="material-icons">route</span>
                <p><strong>Chemin:</strong> ${pathLabels.join(' → ')}</p>
            </div>
            <div id="path-analysis" class="ai-analysis"><span class="streaming-cursor">▊</span></div>
        `;
        analysisModal.classList.add('active');

        const analysisDiv = document.getElementById('path-analysis');
        await this.streamAIResponse(
            '/api/chat/stream',
            { case_id: this.currentCase.id, message: `Analyse ce chemin de relations: ${pathLabels.join(' → ')}` },
            analysisDiv
        );
    },

    async analyzeExpansionCone(nodeId) {
        const node = this.graphNodes?.get(nodeId);
        if (!node) return;

        const coneNodes = new Set([nodeId]);
        let currentLevel = new Set([nodeId]);

        for (let d = 0; d < 2; d++) {
            const nextLevel = new Set();
            this.graphEdges?.forEach(edge => {
                if (currentLevel.has(edge.from)) {
                    nextLevel.add(edge.to);
                    coneNodes.add(edge.to);
                }
                if (currentLevel.has(edge.to)) {
                    nextLevel.add(edge.from);
                    coneNodes.add(edge.from);
                }
            });
            currentLevel = nextLevel;
        }

        const nodeLabels = Array.from(coneNodes).map(id => {
            const n = this.graphNodes?.get(id);
            return n ? n.label : id;
        });

        this.setAnalysisContext('cone_analysis', `Analyse du voisinage de ${node.label}`, `Centre: ${node.label}`);

        const analysisModal = document.getElementById('analysis-modal');
        const analysisContent = document.getElementById('analysis-content');
        const modalTitle = document.getElementById('analysis-modal-title');
        if (modalTitle) modalTitle.textContent = `Analyse du voisinage`;

        const noteBtn = document.getElementById('btn-save-to-notebook');
        if (noteBtn) noteBtn.style.display = '';

        analysisContent.innerHTML = `
            <div class="modal-explanation">
                <span class="material-icons">hub</span>
                <p><strong>Voisinage de ${node.label}:</strong> ${nodeLabels.join(', ')}</p>
            </div>
            <div id="cone-analysis" class="ai-analysis"><span class="streaming-cursor">▊</span></div>
        `;
        analysisModal.classList.add('active');

        const analysisDiv = document.getElementById('cone-analysis');
        await this.streamAIResponse(
            '/api/chat/stream',
            { case_id: this.currentCase.id, message: `Analyse le voisinage de ${node.label} qui inclut: ${nodeLabels.join(', ')}` },
            analysisDiv
        );
    },

    // ============================================
    // Exclude Node from Search
    // ============================================
    excludeNodeFromSearch(nodeId, isN4L = false) {
        const nodes = isN4L ? this.n4lGraphNodes : this.graphNodes;
        const node = nodes?.get(nodeId);
        if (!node) return;

        if (isN4L) {
            if (!this.excludedN4LNodes) this.excludedN4LNodes = [];
            if (!this.excludedN4LNodes.includes(nodeId)) {
                this.excludedN4LNodes.push(nodeId);
                // Visually hide the node
                this.n4lGraphNodes?.update({ id: nodeId, hidden: true });
                this.showToast(`${node.label} masqué du graphe N4L`);
            }
        } else {
            if (!this.excludedNodes) this.excludedNodes = [];
            if (!this.excludedNodes.includes(nodeId)) {
                this.excludedNodes.push(nodeId);
                this.showToast(`${node.label} exclu de la recherche`);
            }
        }
    },

    // ============================================
    // Find Path Modal (bouton "Chemins")
    // ============================================
    showFindPathModal() {
        if (!this.currentCase) {
            this.showToast('Sélectionnez une affaire d\'abord', 'warning');
            return;
        }

        const entities = this.currentCase.entities || [];
        if (entities.length < 2) {
            this.showToast('Il faut au moins 2 entités pour rechercher un chemin', 'warning');
            return;
        }

        const entityOptions = entities.map(e =>
            `<option value="${e.id}">${e.name} (${e.type})</option>`
        ).join('');

        const content = `
            <div class="modal-explanation">
                <span class="material-icons">info</span>
                <p><strong>Découverte de chemins</strong> - Recherche toutes les connexions possibles entre deux entités.
                L'IA analyse ensuite les chemins pour déterminer liens causaux, corrélations ou coïncidences.</p>
            </div>
            <form id="path-form">
                <div class="form-group">
                    <label class="form-label">Entité de départ</label>
                    <select class="form-select" id="path-from" required>
                        ${entityOptions}
                    </select>
                </div>
                <div class="form-group">
                    <label class="form-label">Entité d'arrivée</label>
                    <select class="form-select" id="path-to" required>
                        ${entityOptions}
                    </select>
                </div>
                <div class="form-group">
                    <label class="form-label">Profondeur maximale</label>
                    <input type="number" class="form-input" id="path-depth" min="1" max="10" value="5">
                    <p style="font-size: 0.75rem; color: var(--text-muted); margin-top: 0.25rem;">
                        Nombre maximum de relations à traverser
                    </p>
                </div>
            </form>
        `;

        this.showModal('Découvrir les Chemins', content, async () => {
            const fromId = document.getElementById('path-from').value;
            const toId = document.getElementById('path-to').value;
            const maxDepth = parseInt(document.getElementById('path-depth').value);

            if (fromId === toId) {
                this.showToast('Sélectionnez deux entités différentes', 'warning');
                return;
            }

            await this.findPath(fromId, toId, maxDepth);
        });

        // Pre-select second entity if possible
        if (entities.length >= 2) {
            setTimeout(() => {
                const pathTo = document.getElementById('path-to');
                if (pathTo) pathTo.selectedIndex = 1;
            }, 100);
        }
    },

    async findPath(fromId, toId, maxDepth) {
        const analysisContent = document.getElementById('analysis-content');
        const analysisModal = document.getElementById('analysis-modal');
        const modalTitle = document.getElementById('analysis-modal-title');

        const fromEntity = this.currentCase?.entities?.find(e => e.id === fromId);
        const toEntity = this.currentCase?.entities?.find(e => e.id === toId);
        const fromName = fromEntity?.name || fromId;
        const toName = toEntity?.name || toId;

        this.setAnalysisContext('path_analysis', `Chemins: ${fromName} → ${toName}`, `Connexions entre ${fromName} et ${toName}`);

        if (modalTitle) modalTitle.textContent = 'Analyse de chemins';

        const noteBtn = document.getElementById('btn-save-to-notebook');
        if (noteBtn) noteBtn.style.display = '';

        analysisContent.innerHTML = `
            <div class="loading-analysis">
                <div class="spinner"></div>
                <p style="margin-top: 1rem; color: var(--text-muted);">Recherche des chemins en cours...</p>
            </div>
        `;
        analysisModal.classList.add('active');

        try {
            const result = await this.apiCall('/api/analyze/path', 'POST', {
                case_id: this.currentCase.id,
                from_id: fromId,
                to_id: toId,
                max_depth: maxDepth
            });

            console.log('[FindPath] API result:', result);
            console.log('[FindPath] paths:', result?.paths, 'length:', result?.paths?.length);
            console.log('[FindPath] Request params:', { case_id: this.currentCase.id, from_id: fromId, to_id: toId });

            if (result.paths && result.paths.length > 0) {
                let pathsHtml = `<h4 style="color: var(--primary); margin-bottom: 1rem;">Chemins Découverts</h4>`;
                pathsHtml += `<div style="margin-bottom: 1.5rem;">`;
                result.paths.forEach((path, index) => {
                    pathsHtml += `
                        <div class="path-result" style="background: var(--bg-subtle); padding: 0.75rem; border-radius: 6px; margin-bottom: 0.5rem;">
                            <strong>Chemin ${index + 1}:</strong><br>
                            <span style="font-family: monospace; color: var(--primary);">${path.join(' → ')}</span>
                        </div>
                    `;
                });
                pathsHtml += `</div>`;
                pathsHtml += `<h4 style="color: var(--primary); margin-bottom: 1rem;">Analyse IA</h4>`;
                pathsHtml += `<div id="path-analysis-content"><span class="streaming-cursor">▊</span></div>`;

                analysisContent.innerHTML = pathsHtml;

                if (this.graph) {
                    this.highlightPathInGraph(fromId, toId);
                }

                const prompt = `Analyse les connexions entre "${fromName}" et "${toName}" dans l'affaire "${this.currentCase.name}".

Chemins découverts:
${result.paths.map((p, i) => `Chemin ${i + 1}: ${p.join(' → ')}`).join('\n')}

Contexte de l'affaire: ${this.currentCase.description}

Analyse demandée:
1. Quelle est la nature de la connexion entre ces entités?
2. Y a-t-il des implications significatives?
3. Ces connexions suggèrent-elles des pistes d'investigation?`;

                const pathAnalysisDiv = document.getElementById('path-analysis-content');
                await this.streamAIResponse(
                    '/api/chat/stream',
                    { case_id: this.currentCase.id, message: prompt },
                    pathAnalysisDiv
                );
            } else {
                analysisContent.innerHTML = `
                    <div class="empty-state" style="padding: 2rem;">
                        <span class="material-icons empty-state-icon">link_off</span>
                        <p class="empty-state-title">Aucun chemin trouvé</p>
                        <p class="empty-state-description">Ces deux entités ne sont pas connectées dans le graphe actuel.</p>
                    </div>
                `;
            }
        } catch (error) {
            console.error('Error finding path:', error);
            analysisContent.innerHTML = `
                <div class="empty-state">
                    <span class="material-icons empty-state-icon" style="color: var(--danger);">error</span>
                    <p class="empty-state-title">Erreur</p>
                    <p class="empty-state-description">${error.message}</p>
                </div>
            `;
        }
    },

    highlightPathInGraph(fromId, toId) {
        if (!this.graph || !this.graphNodes) return;

        try {
            // Trouver les noms des entités pour les IDs (N4L utilise les noms comme IDs de nœuds)
            const normalizeId = (id) => id ? id.replace(/-/g, '_') : '';
            const entityMap = {};
            (this.currentCase?.entities || []).forEach(e => {
                entityMap[e.id] = e.name;
                entityMap[normalizeId(e.id)] = e.name;
            });

            const fromName = entityMap[fromId] || entityMap[normalizeId(fromId)] || fromId;
            const toName = entityMap[toId] || entityMap[normalizeId(toId)] || toId;

            // Trouver les nœuds correspondants dans le graphe
            const allNodeIds = this.graphNodes.getIds();
            let fromNodeId = null;
            let toNodeId = null;

            for (const nodeId of allNodeIds) {
                const node = this.graphNodes.get(nodeId);
                if (node && (nodeId === fromName || node.label === fromName)) {
                    fromNodeId = nodeId;
                }
                if (node && (nodeId === toName || node.label === toName)) {
                    toNodeId = nodeId;
                }
            }

            if (!fromNodeId && !toNodeId) return;

            const updates = [];
            if (fromNodeId) {
                updates.push({
                    id: fromNodeId,
                    borderWidth: 4,
                    color: { border: '#dc2626', background: '#fee2e2' }
                });
            }
            if (toNodeId) {
                updates.push({
                    id: toNodeId,
                    borderWidth: 4,
                    color: { border: '#16a34a', background: '#dcfce7' }
                });
            }

            if (updates.length > 0) {
                this.graphNodes.update(updates);
            }

            const fitNodes = [fromNodeId, toNodeId].filter(Boolean);
            if (fitNodes.length > 0) {
                this.graph.fit({
                    nodes: fitNodes,
                    animation: true
                });
            }
        } catch (error) {
            console.warn('[highlightPathInGraph] Error:', error);
        }
    },

    // ============================================
    // Graph Metadata Display (like N4L)
    // ============================================
    graphActiveContextFilter: null,
    lastGraphParse: null,

    async renderGraphMetadata() {
        const metadataContainer = document.getElementById('graph-metadata-container');
        if (!metadataContainer || !this.currentCase) return;

        // Get parsed N4L data for metadata
        try {
            const n4lContent = await fetch(`/api/n4l/export?case_id=${this.currentCase.id}`).then(r => r.text());
            if (!n4lContent) {
                metadataContainer.innerHTML = '';
                return;
            }

            const result = await fetch('/api/n4l/parse', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    content: n4lContent,
                    case_id: this.currentCase.id
                })
            }).then(r => r.json());

            this.lastGraphParse = result;
            this.showGraphMetadata(result);
        } catch (error) {
            console.error('Error loading graph metadata:', error);
            metadataContainer.innerHTML = '';
        }
    },

    showGraphMetadata(result) {
        const metadataContainer = document.getElementById('graph-metadata-container');
        if (!metadataContainer) return;

        const nodeCount = result.graph?.nodes?.length || 0;
        const edgeCount = result.graph?.edges?.length || 0;
        const contextCount = result.contexts?.length || 0;
        const aliasCount = result.aliases ? Object.keys(result.aliases).length : 0;
        const sequenceCount = result.sequences?.length || 0;

        let metadataHtml = `
            <!-- Info Banner -->
            <div class="n4l-info-banner">
                <span class="material-icons">info</span>
                <div>
                    <strong>Graphe des Relations</strong> - Visualisation interactive des liens entre entités.
                    Cliquez sur un <em>contexte</em> pour filtrer le graphe, sur un <em>nœud</em> pour voir ses connexions.
                </div>
            </div>

            <!-- Stats Row -->
            <div class="n4l-stats-row">
                <div class="n4l-stat-item">
                    <span class="material-icons">hub</span>
                    <span class="n4l-stat-value">${nodeCount}</span>
                    <span class="n4l-stat-label">entités</span>
                </div>
                <div class="n4l-stat-item">
                    <span class="material-icons">sync_alt</span>
                    <span class="n4l-stat-value">${edgeCount}</span>
                    <span class="n4l-stat-label">relations</span>
                </div>
                <div class="n4l-stat-item">
                    <span class="material-icons">layers</span>
                    <span class="n4l-stat-value">${contextCount}</span>
                    <span class="n4l-stat-label">contextes</span>
                </div>
                ${aliasCount > 0 ? `
                <div class="n4l-stat-item">
                    <span class="material-icons">alternate_email</span>
                    <span class="n4l-stat-value">${aliasCount}</span>
                    <span class="n4l-stat-label">alias</span>
                </div>
                ` : ''}
                ${sequenceCount > 0 ? `
                <div class="n4l-stat-item">
                    <span class="material-icons">timeline</span>
                    <span class="n4l-stat-value">${sequenceCount}</span>
                    <span class="n4l-stat-label">séquences</span>
                </div>
                ` : ''}
            </div>

            <!-- Layout Options -->
            <div class="n4l-section">
                <div class="n4l-section-header">
                    <span class="material-icons">grid_view</span>
                    <span>Disposition</span>
                </div>
                <div class="n4l-layout-grid">
                    <button class="n4l-layout-btn ${this.currentGraphLayout === 'physics' || !this.currentGraphLayout ? 'active' : ''}" onclick="app.setGraphLayout('physics')" data-tooltip="Disposition automatique avec simulation physique">
                        <span class="material-icons">bubble_chart</span>
                        <span>Standard</span>
                    </button>
                    <button class="n4l-layout-btn ${this.currentGraphLayout === 'compact' ? 'active' : ''}" onclick="app.setGraphLayout('compact')" data-tooltip="Disposition compacte et resserrée">
                        <span class="material-icons">compress</span>
                        <span>Compact</span>
                    </button>
                </div>
            </div>
        `;

        // Contexts Section with Filters
        if (result.contexts && result.contexts.length > 0) {
            metadataHtml += `
                <div class="n4l-section">
                    <div class="n4l-section-header">
                        <span class="material-icons">filter_list</span>
                        <span>Filtrer par contexte</span>
                        <button class="n4l-reset-btn ${this.graphActiveContextFilter ? '' : 'hidden'}" onclick="app.resetGraphFilter()" title="Réinitialiser le filtre">
                            <span class="material-icons">refresh</span> Reset
                        </button>
                    </div>
                    <div class="n4l-context-grid">
                        ${result.contexts.map(ctx => {
                            const isActive = this.graphActiveContextFilter === ctx;
                            const icon = this.getGraphContextIcon(ctx);
                            return `
                                <button class="n4l-context-btn ${isActive ? 'active' : ''}" onclick="app.filterGraphByContext('${ctx}')">
                                    <span class="material-icons">${icon}</span>
                                    <span>${ctx}</span>
                                </button>
                            `;
                        }).join('')}
                    </div>
                </div>
            `;
        }

        // Sequences Section (collapsible)
        if (result.sequences && result.sequences.length > 0) {
            metadataHtml += `
                <div class="n4l-section">
                    <div class="n4l-section-header n4l-collapsible" onclick="app.toggleN4LSection(this)">
                        <span class="material-icons">timeline</span>
                        <span>Chronologie (${result.sequences.length} séquence${result.sequences.length > 1 ? 's' : ''})</span>
                        <span class="material-icons n4l-expand-icon">expand_more</span>
                    </div>
                    <div class="n4l-section-content n4l-collapsed">
                        ${result.sequences.map((seq, i) => `
                            <div class="n4l-sequence-block" data-seq-index="${i}">
                                <div class="n4l-sequence-item" onclick="app.highlightGraphSequence(${i})">
                                    <div class="n4l-sequence-header">
                                        <span class="n4l-sequence-number">${i + 1}</span>
                                        <span class="n4l-sequence-count">${seq.length} étapes</span>
                                    </div>
                                    <div class="n4l-sequence-preview">
                                        ${seq.slice(0, 4).join(' → ')}${seq.length > 4 ? ' → ...' : ''}
                                    </div>
                                </div>
                                <div class="n4l-sequence-nav" id="graph-seq-nav-${i}" style="display: none;">
                                    <div class="n4l-sequence-nav-header">
                                        <button class="n4l-nav-btn" onclick="event.stopPropagation(); app.navigateGraphSequence(${i}, -1)" title="Étape précédente">
                                            <span class="material-icons">skip_previous</span>
                                        </button>
                                        <span class="n4l-nav-indicator">
                                            <span class="n4l-nav-current" id="graph-seq-current-${i}">1</span> / ${seq.length}
                                        </span>
                                        <button class="n4l-nav-btn" onclick="event.stopPropagation(); app.navigateGraphSequence(${i}, 1)" title="Étape suivante">
                                            <span class="material-icons">skip_next</span>
                                        </button>
                                        <button class="n4l-nav-btn n4l-nav-play" onclick="event.stopPropagation(); app.playGraphSequence(${i})" title="Lecture automatique" id="graph-seq-play-${i}">
                                            <span class="material-icons">play_arrow</span>
                                        </button>
                                        <button class="n4l-nav-btn" onclick="event.stopPropagation(); app.stopGraphSequenceNav(${i})" title="Fermer">
                                            <span class="material-icons">close</span>
                                        </button>
                                    </div>
                                    <div class="n4l-sequence-current-label" id="graph-seq-label-${i}"></div>
                                </div>
                            </div>
                        `).join('')}
                    </div>
                </div>
            `;
        }

        // Aliases Section (collapsible)
        if (result.aliases && Object.keys(result.aliases).length > 0) {
            const aliases = Object.entries(result.aliases);
            const escapeHtml = (str) => String(str).replace(/"/g, '&quot;').replace(/'/g, '&#39;');
            const getDisplayName = (value) => {
                let v = Array.isArray(value) ? value[0] : value;
                if (!v) return '';
                return v.split('(')[0].trim();
            };
            metadataHtml += `
                <div class="n4l-section">
                    <div class="n4l-section-header n4l-collapsible" onclick="app.toggleN4LSection(this)">
                        <span class="material-icons">alternate_email</span>
                        <span>Alias (${aliases.length})</span>
                        <span class="material-icons n4l-expand-icon">expand_more</span>
                    </div>
                    <div class="n4l-section-content n4l-collapsed">
                        <div class="n4l-alias-grid">
                            ${aliases.map(([key, value]) => `
                                <div class="n4l-alias-item" onclick="app.focusGraphNodeByLabel('${escapeHtml(getDisplayName(value))}')">
                                    <span class="n4l-alias-key">${key}</span>
                                    <span class="material-icons n4l-alias-arrow">arrow_forward</span>
                                    <span class="n4l-alias-value">${getDisplayName(value)}</span>
                                </div>
                            `).join('')}
                        </div>
                    </div>
                </div>
            `;
        }

        metadataContainer.innerHTML = metadataHtml;
    },

    getGraphContextIcon(context) {
        const icons = {
            'victimes': 'person_off',
            'suspects': 'person_search',
            'témoins': 'record_voice_over',
            'lieux': 'place',
            'objets': 'category',
            'preuves': 'verified',
            'indices': 'search',
            'chronologie': 'schedule',
            'hypothèses': 'lightbulb',
            'hypotheses': 'label',
            'pistes': 'explore',
            'réseau': 'share',
            'relations': 'people'
        };

        const lowerCtx = context.toLowerCase();
        for (const [key, icon] of Object.entries(icons)) {
            if (lowerCtx.includes(key)) return icon;
        }
        return 'label';
    },

    // ============================================
    // Graph Context Filtering
    // ============================================
    filterGraphByContext(context) {
        if (!this.lastGraphParse || !this.graph || !this.graphNodes || !this.graphEdges) return;

        // Toggle filter if clicking same context
        if (this.graphActiveContextFilter === context) {
            this.resetGraphFilter();
            return;
        }

        // Set active filter
        this.graphActiveContextFilter = context;

        const result = this.lastGraphParse;
        const filteredEdges = result.graph.edges.filter(e =>
            !e.context || e.context === context || e.context.includes(context)
        );

        const involvedNodes = new Set();
        filteredEdges.forEach(e => {
            involvedNodes.add(e.from);
            involvedNodes.add(e.to);
        });

        // Hide non-matching nodes
        const nodeUpdates = this.graphNodes.get().map(node => ({
            id: node.id,
            hidden: !involvedNodes.has(node.id),
            opacity: involvedNodes.has(node.id) ? 1 : 0.2
        }));
        this.graphNodes.update(nodeUpdates);

        // Hide non-matching edges
        const edgeUpdates = this.graphEdges.get().map(edge => {
            const matches = filteredEdges.some(fe =>
                (fe.from === edge.from || fe.from === this.graphNodes.get(edge.from)?.label) &&
                (fe.to === edge.to || fe.to === this.graphNodes.get(edge.to)?.label)
            );
            return {
                id: edge.id,
                hidden: !matches,
                opacity: matches ? 1 : 0.2
            };
        });
        this.graphEdges.update(edgeUpdates);

        // Refresh metadata display to update button states
        this.showGraphMetadata(this.lastGraphParse);

        this.showToast(`Filtré: ${context} (${involvedNodes.size} entités)`);
    },

    resetGraphFilter() {
        if (!this.graph || !this.graphNodes || !this.graphEdges) return;

        this.graphActiveContextFilter = null;

        // Reset all nodes
        const nodeUpdates = this.graphNodes.get().map(node => ({
            id: node.id,
            hidden: false,
            opacity: 1
        }));
        this.graphNodes.update(nodeUpdates);

        // Reset all edges
        const edgeUpdates = this.graphEdges.get().map(edge => ({
            id: edge.id,
            hidden: false,
            opacity: 1
        }));
        this.graphEdges.update(edgeUpdates);

        // Also reset node colors
        this.resetGraphFocus();

        // Refresh metadata display to update button states
        if (this.lastGraphParse) {
            this.showGraphMetadata(this.lastGraphParse);
        }

        this.showToast('Filtre réinitialisé');
    },

    // ============================================
    // Graph Layout Options
    // ============================================
    currentGraphLayout: 'physics',

    setGraphLayout(layoutType) {
        if (!this.graph || !this.graphNodes || !this.graphEdges) return;

        this.currentGraphLayout = layoutType;

        // Update button states in UI
        document.querySelectorAll('.n4l-layout-btn').forEach(btn => {
            btn.classList.remove('active');
        });
        const activeBtn = document.querySelector(`.n4l-layout-btn[onclick*="'${layoutType}'"]`);
        if (activeBtn) activeBtn.classList.add('active');

        // vis.js doesn't handle hierarchical toggle well, so we recreate the network
        const container = document.getElementById('graph-container');
        if (!container) return;

        // Get current nodes and edges data
        const nodesData = this.graphNodes.get().map(node => ({
            ...node,
            x: undefined,
            y: undefined
        }));
        const edgesData = this.graphEdges.get();

        // Destroy current network
        this.graph.destroy();

        // Create new DataSets
        const nodes = new vis.DataSet(nodesData);
        const edges = new vis.DataSet(edgesData);

        // Build options based on layout type
        let options = {
            nodes: {
                font: { color: '#1a1a2e', size: 12 },
                borderWidth: 2
            },
            edges: {
                font: { size: 10, color: '#4a5568' },
                smooth: { type: 'curvedCW', roundness: 0.2 }
            },
            interaction: {
                hover: true,
                tooltipDelay: 200
            }
        };

        if (layoutType === 'physics') {
            options.physics = {
                enabled: true,
                stabilization: { iterations: 150 },
                barnesHut: {
                    gravitationalConstant: -3000,
                    springLength: 200,
                    springConstant: 0.02,
                    damping: 0.3,
                    avoidOverlap: 0.5
                },
                minVelocity: 0.75
            };
            options.layout = { hierarchical: false };
        } else if (layoutType === 'compact') {
            // Mode compact: physique avec paramètres resserrés
            options.physics = {
                enabled: true,
                stabilization: { iterations: 200 },
                barnesHut: {
                    gravitationalConstant: -8000,
                    springLength: 80,
                    springConstant: 0.08,
                    damping: 0.5,
                    avoidOverlap: 0.8
                },
                minVelocity: 0.5
            };
            options.layout = { hierarchical: false };
        }

        // Create new network
        this.graph = new vis.Network(container, { nodes, edges }, options);
        this.graphNodes = nodes;
        this.graphEdges = edges;

        // Re-attach event handlers
        this.graph.on('click', (params) => {
            if (params.nodes.length > 0) {
                this.focusGraphNode(params.nodes[0]);
            } else {
                this.resetGraphFocus();
            }
        });

        this.graph.on('oncontext', (params) => {
            params.event.preventDefault();
            this.handleGraphRightClick(params);
        });

        // Fit after stabilization (both physics and compact use physics)
        this.graph.once('stabilized', () => {
            this.graph.fit({ animation: { duration: 300, easingFunction: 'easeInOutQuad' } });
        });

        const layoutNames = { physics: 'Standard', compact: 'Compact' };
        this.showToast(`Disposition: ${layoutNames[layoutType] || layoutType}`);
    },

    // ============================================
    // Sequence Highlighting in Main Graph
    // ============================================
    graphCurrentSequenceIndex: null,
    graphCurrentSequenceNodeIds: null,
    graphCurrentSequenceStep: 0,
    graphSequencePlayInterval: null,

    highlightGraphSequence(seqIndex) {
        if (!this.lastGraphParse?.sequences || !this.graph || !this.graphNodes || !this.graphEdges) return;

        const sequence = this.lastGraphParse.sequences[seqIndex];
        if (!sequence || sequence.length === 0) return;

        // Find nodes matching sequence labels
        const seqNodeIds = [];
        const allNodes = this.graphNodes.get();

        sequence.forEach(label => {
            const node = allNodes.find(n => n.label === label || n.id === label || n.label.includes(label));
            if (node) seqNodeIds.push(node.id);
        });

        if (seqNodeIds.length === 0) {
            this.showToast('Séquence non trouvée sur le graphe');
            return;
        }

        // Store sequence navigation state
        this.graphCurrentSequenceIndex = seqIndex;
        this.graphCurrentSequenceNodeIds = seqNodeIds;
        this.graphCurrentSequenceStep = 0;
        this.graphSequencePlayInterval = null;

        // Highlight all nodes in sequence
        const seqNodeSet = new Set(seqNodeIds);
        const nodeUpdates = allNodes.map(node => {
            const inSeq = seqNodeSet.has(node.id);
            return {
                id: node.id,
                color: inSeq ? { background: '#f59e0b', border: '#d97706' } : { background: '#e2e8f0', border: '#cbd5e0' },
                opacity: inSeq ? 1 : 0.3,
                borderWidth: inSeq ? 3 : 1
            };
        });
        this.graphNodes.update(nodeUpdates);

        // Show navigation controls
        const navEl = document.getElementById(`graph-seq-nav-${seqIndex}`);
        if (navEl) {
            navEl.style.display = 'block';
        }

        // Navigate to first node
        this.navigateToGraphSequenceStep(seqIndex, 0);

        this.showToast(`Chronologie: ${seqNodeIds.length} étapes`);
    },

    // Navigate within sequence
    navigateGraphSequence(seqIndex, direction) {
        if (!this.graphCurrentSequenceNodeIds || this.graphCurrentSequenceIndex !== seqIndex) return;

        const newStep = this.graphCurrentSequenceStep + direction;
        if (newStep < 0 || newStep >= this.graphCurrentSequenceNodeIds.length) {
            // Loop around
            const loopedStep = newStep < 0 ? this.graphCurrentSequenceNodeIds.length - 1 : 0;
            this.navigateToGraphSequenceStep(seqIndex, loopedStep);
        } else {
            this.navigateToGraphSequenceStep(seqIndex, newStep);
        }
    },

    // Navigate to specific step
    navigateToGraphSequenceStep(seqIndex, step) {
        if (!this.graphCurrentSequenceNodeIds) return;

        this.graphCurrentSequenceStep = step;
        const nodeId = this.graphCurrentSequenceNodeIds[step];

        // Update all nodes - highlight current step, show sequence nodes, hide others
        const allNodes = this.graphNodes.get();
        const seqNodeSet = new Set(this.graphCurrentSequenceNodeIds);

        const nodeUpdates = allNodes.map(node => {
            const isCurrentStep = node.id === nodeId;
            const inSeq = seqNodeSet.has(node.id);

            if (isCurrentStep) {
                return {
                    id: node.id,
                    hidden: false,
                    color: { background: '#dc2626', border: '#991b1b' },
                    opacity: 1,
                    borderWidth: 5
                };
            } else if (inSeq) {
                return {
                    id: node.id,
                    hidden: false,
                    color: { background: '#f59e0b', border: '#d97706' },
                    opacity: 0.7,
                    borderWidth: 2
                };
            } else {
                return {
                    id: node.id,
                    hidden: true
                };
            }
        });
        this.graphNodes.update(nodeUpdates);

        // Also hide edges not connected to sequence nodes
        const allEdges = this.graphEdges.get();
        const edgeUpdates = allEdges.map(edge => {
            const fromInSeq = seqNodeSet.has(edge.from);
            const toInSeq = seqNodeSet.has(edge.to);
            return {
                id: edge.id,
                hidden: !(fromInSeq && toInSeq)
            };
        });
        this.graphEdges.update(edgeUpdates);

        // Focus on current node
        this.graph.focus(nodeId, { scale: 1.2, animation: { duration: 300 } });

        // Update UI
        const currentEl = document.getElementById(`graph-seq-current-${seqIndex}`);
        if (currentEl) currentEl.textContent = step + 1;

        const labelEl = document.getElementById(`graph-seq-label-${seqIndex}`);
        if (labelEl) {
            const node = allNodes.find(n => n.id === nodeId);
            labelEl.textContent = node ? node.label : nodeId;
        }
    },

    // Auto-play sequence
    playGraphSequence(seqIndex) {
        if (this.graphSequencePlayInterval) {
            this.pauseGraphSequence(seqIndex);
            return;
        }

        const playBtn = document.getElementById(`graph-seq-play-${seqIndex}`);
        if (playBtn) {
            playBtn.querySelector('.material-icons').textContent = 'pause';
        }

        this.graphSequencePlayInterval = setInterval(() => {
            const nextStep = this.graphCurrentSequenceStep + 1;
            if (nextStep >= this.graphCurrentSequenceNodeIds.length) {
                this.pauseGraphSequence(seqIndex);
                this.navigateToGraphSequenceStep(seqIndex, 0);
            } else {
                this.navigateToGraphSequenceStep(seqIndex, nextStep);
            }
        }, 1500);
    },

    // Pause sequence playback
    pauseGraphSequence(seqIndex) {
        if (this.graphSequencePlayInterval) {
            clearInterval(this.graphSequencePlayInterval);
            this.graphSequencePlayInterval = null;
        }

        const playBtn = document.getElementById(`graph-seq-play-${seqIndex}`);
        if (playBtn) {
            playBtn.querySelector('.material-icons').textContent = 'play_arrow';
        }
    },

    // Stop and close sequence navigation
    stopGraphSequenceNav(seqIndex) {
        this.pauseGraphSequence(seqIndex);

        // Hide navigation
        const navEl = document.getElementById(`graph-seq-nav-${seqIndex}`);
        if (navEl) {
            navEl.style.display = 'none';
        }

        // Show all nodes again
        const allNodes = this.graphNodes.get();
        const nodeUpdates = allNodes.map(node => ({
            id: node.id,
            hidden: false
        }));
        this.graphNodes.update(nodeUpdates);

        // Show all edges again
        const allEdges = this.graphEdges.get();
        const edgeUpdates = allEdges.map(edge => ({
            id: edge.id,
            hidden: false
        }));
        this.graphEdges.update(edgeUpdates);

        // Reset graph colors
        this.resetGraphFocus();

        // Clear state
        this.graphCurrentSequenceIndex = null;
        this.graphCurrentSequenceNodeIds = null;
        this.graphCurrentSequenceStep = 0;
    },

    // ============================================
    // Focus Node by Label
    // ============================================
    focusGraphNodeByLabel(label) {
        if (!this.graph || !this.graphNodes) return;

        const nodes = this.graphNodes.get();
        const targetNode = nodes.find(n => n.label === label || n.id === label);

        if (targetNode) {
            this.focusGraphNode(targetNode.id);
        } else {
            this.showToast(`Nœud "${label}" non trouvé`);
        }
    }
};

// Export for use in main app
if (typeof module !== 'undefined' && module.exports) {
    module.exports = GraphModule;
}
