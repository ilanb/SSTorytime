// ForensicInvestigator - Module N4L
// Gestion du contenu N4L, import, export, graphe N4L

const N4LModule = {
    // ============================================
    // Load N4L Content
    // ============================================
    async loadN4LContent() {
        if (!this.currentCase) return;

        try {
            // Charger le contenu N4L depuis l'API (retourne du texte brut)
            const response = await fetch(`/api/n4l/export?case_id=${this.currentCase.id}`);
            const n4l = await response.text();

            document.getElementById('n4l-editor').value = n4l;

            // Mettre à jour le DataProvider si disponible
            if (typeof DataProvider !== 'undefined') {
                DataProvider.n4lContent = n4l;
            }

            // Parser et afficher le graphe N4L
            await this.parseN4L();
        } catch (error) {
            console.error('Error loading N4L:', error);
        }
    },

    // ============================================
    // Parse N4L
    // ============================================
    async parseN4L() {
        const content = document.getElementById('n4l-editor').value;
        if (!content) return;

        try {
            // Parser le N4L via l'API
            const result = await fetch('/api/n4l/parse', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    content: content,
                    case_id: this.currentCase?.id
                })
            }).then(r => r.json());

            this.lastN4LParse = result;

            console.log('N4LModule: Résultat parsing complet', result);
            console.log('N4LModule: Graphe', result.graph);

            // Utiliser le graphe du résultat parsé - c'est la même source que /api/graph
            // car les deux utilisent ParseForensicN4L
            if (result.graph && result.graph.nodes && result.graph.nodes.length > 0) {
                this.renderN4LGraph(result.graph);
            } else {
                console.warn('N4LModule: Graphe vide ou invalide, affichage message');
                const container = document.getElementById('n4l-graph-container');
                if (container) {
                    container.innerHTML = `
                        <div class="empty-state" style="height: 100%; display: flex; flex-direction: column; justify-content: center; align-items: center;">
                            <span class="material-icons empty-state-icon">hub</span>
                            <p class="empty-state-title">Graphe N4L vide</p>
                            <p class="empty-state-description">Le contenu N4L ne contient pas de relations ou entités</p>
                        </div>
                    `;
                }
            }
            this.showN4LMetadata(result);

            console.log('N4LModule: Graphe parsé', {
                nodes: result.graph?.nodes?.length || 0,
                edges: result.graph?.edges?.length || 0
            });
        } catch (error) {
            console.error('Error parsing N4L:', error);
        }
    },

    // ============================================
    // Render N4L Graph (utilise le même style que le graphe principal)
    // ============================================
    renderN4LGraph(graphData) {
        console.log('N4LModule.renderN4LGraph appelé avec:', graphData);

        const container = document.getElementById('n4l-graph-container');
        console.log('N4LModule: Container trouvé:', !!container);

        if (!container) {
            console.error('N4LModule: Container n4l-graph-container non trouvé!');
            return;
        }
        if (!graphData || !graphData.nodes) {
            console.error('N4LModule: graphData invalide:', graphData);
            return;
        }
        if (graphData.nodes.length === 0) {
            console.warn('N4LModule: graphData.nodes est vide');
            container.innerHTML = '<div class="empty-state"><p>Aucun nœud dans le graphe</p></div>';
            return;
        }

        console.log('N4LModule: Création du graphe avec', graphData.nodes.length, 'nœuds et', graphData.edges?.length || 0, 'arêtes');

        container.innerHTML = '';

        this.n4lGraphNodesData = graphData.nodes;
        this.n4lGraphEdgesData = graphData.edges;

        // Utiliser getEdgeColor du GraphModule si disponible
        const getEdgeColor = (edge) => {
            if (typeof this.getEdgeColor === 'function') {
                return this.getEdgeColor(edge);
            }
            switch (edge.type) {
                case 'never': return '#dc2626';
                case 'new': return '#059669';
                case 'sequence': return '#f59e0b';
                case 'equivalence': return '#8b5cf6';
                case 'group': return '#06b6d4';
                default: return '#1e3a5f';
            }
        };

        const nodes = new vis.DataSet(graphData.nodes.map(n => ({
            id: n.id,
            label: n.label,
            color: this.getNodeColor(n),
            shape: this.getNodeShape(n),
            title: this.getNodeTooltip(n),
            originalColor: this.getNodeColor(n),
            context: n.context,
            nodeType: n.type,
            role: n.role
        })));

        const edges = new vis.DataSet(graphData.edges.map((e, i) => ({
            id: `n4l-edge-${i}`,
            from: e.from,
            to: e.to,
            label: e.label,
            arrows: e.type === 'equivalence' ? '' : 'to',
            dashes: e.type === 'new',
            color: { color: getEdgeColor(e) },
            title: e.context ? `Contexte: ${e.context}` : '',
            originalColor: getEdgeColor(e),
            edgeType: e.type,
            context: e.context
        })));

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
                stabilization: { iterations: 100 },
                barnesHut: {
                    gravitationalConstant: -2000,
                    springLength: 200
                }
            },
            interaction: {
                hover: true,
                tooltipDelay: 200
            }
        };

        this.n4lGraph = new vis.Network(container, { nodes, edges }, options);
        this.n4lGraphNodes = nodes;
        this.n4lGraphEdges = edges;
        this.selectedN4LNode = null;

        this.n4lGraph.on('click', (params) => {
            if (params.nodes.length > 0) {
                this.focusN4LGraphNode(params.nodes[0]);
            } else {
                this.resetN4LGraphFocus();
            }
        });

        this.n4lGraph.on('oncontext', (params) => {
            params.event.preventDefault();
            this.handleN4LGraphRightClick(params);
        });

        // Setup context menu actions (if not already done by graph.js)
        if (typeof this.setupContextMenuActions === 'function') {
            this.setupContextMenuActions();
        }

        this.addN4LLegend(container);
    },

    // ============================================
    // N4L Metadata Display
    // ============================================
    showN4LMetadata(result) {
        const metadataContainer = document.getElementById('n4l-metadata-container');
        if (!metadataContainer) return;

        const nodeCount = result.graph?.nodes?.length || 0;
        const edgeCount = result.graph?.edges?.length || 0;
        const contextCount = result.contexts?.length || 0;
        const aliasCount = result.aliases ? Object.keys(result.aliases).length : 0;
        const sequenceCount = result.sequences?.length || 0;

        // Store current filter state
        this.n4lActiveContextFilter = this.n4lActiveContextFilter || null;

        let metadataHtml = `
            <!-- Info Banner -->
            <div class="n4l-info-banner">
                <span class="material-icons">info</span>
                <div>
                    <strong>Graphe N4L interactif</strong> - Visualisation de la structure sémantique de l'affaire.
                    Cliquez sur un <em>contexte</em> pour filtrer le graphe, sur un <em>nœud</em> pour voir ses connexions,
                    ou utilisez le clic droit pour plus d'options.
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
        `;

        // Contexts Section with Filters
        if (result.contexts && result.contexts.length > 0) {
            metadataHtml += `
                <div class="n4l-section">
                    <div class="n4l-section-header">
                        <span class="material-icons">filter_list</span>
                        <span>Filtrer par contexte</span>
                        <button class="n4l-reset-btn ${this.n4lActiveContextFilter ? '' : 'hidden'}" onclick="app.resetN4LFilter()" title="Réinitialiser le filtre">
                            <span class="material-icons">refresh</span> Reset
                        </button>
                    </div>
                    <div class="n4l-context-grid">
                        ${result.contexts.map(ctx => {
                            const isActive = this.n4lActiveContextFilter === ctx;
                            const icon = this.getContextIcon(ctx);
                            return `
                                <button class="n4l-context-btn ${isActive ? 'active' : ''}" onclick="app.filterByContext('${ctx}')">
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
                                <div class="n4l-sequence-item" onclick="app.highlightN4LSequence(${i})">
                                    <div class="n4l-sequence-header">
                                        <span class="n4l-sequence-number">${i + 1}</span>
                                        <span class="n4l-sequence-count">${seq.length} étapes</span>
                                    </div>
                                    <div class="n4l-sequence-preview">
                                        ${seq.slice(0, 4).join(' → ')}${seq.length > 4 ? ' → ...' : ''}
                                    </div>
                                </div>
                                <div class="n4l-sequence-nav" id="seq-nav-${i}" style="display: none;">
                                    <div class="n4l-sequence-nav-header">
                                        <button class="n4l-nav-btn" onclick="event.stopPropagation(); app.navigateSequence(${i}, -1)" title="Étape précédente">
                                            <span class="material-icons">skip_previous</span>
                                        </button>
                                        <span class="n4l-nav-indicator">
                                            <span class="n4l-nav-current" id="seq-current-${i}">1</span> / ${seq.length}
                                        </span>
                                        <button class="n4l-nav-btn" onclick="event.stopPropagation(); app.navigateSequence(${i}, 1)" title="Étape suivante">
                                            <span class="material-icons">skip_next</span>
                                        </button>
                                        <button class="n4l-nav-btn n4l-nav-play" onclick="event.stopPropagation(); app.playSequence(${i})" title="Lecture automatique" id="seq-play-${i}">
                                            <span class="material-icons">play_arrow</span>
                                        </button>
                                        <button class="n4l-nav-btn" onclick="event.stopPropagation(); app.stopSequenceNav(${i})" title="Fermer">
                                            <span class="material-icons">close</span>
                                        </button>
                                    </div>
                                    <div class="n4l-sequence-current-label" id="seq-label-${i}"></div>
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
            // Helper to extract display name from alias value
            const getDisplayName = (value) => {
                let v = Array.isArray(value) ? value[0] : value;
                if (!v) return '';
                // Remove type info in parentheses, e.g., "Casino de Deauville (type) lieu" -> "Casino de Deauville"
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
                            ${aliases.slice(0, 20).map(([key, value]) => {
                                const displayName = getDisplayName(value);
                                return `
                                <div class="n4l-alias-item" data-alias-key="${escapeHtml(key)}" title="@${key}">
                                    <span class="n4l-alias-key">${escapeHtml(displayName)}</span>
                                </div>
                            `;
                            }).join('')}
                            ${aliases.length > 20 ? `<div class="n4l-alias-more">+${aliases.length - 20} autres...</div>` : ''}
                        </div>
                    </div>
                </div>
            `;
        }

        // TODOs Section
        if (result.todo_items && result.todo_items.length > 0) {
            metadataHtml += `
                <div class="n4l-section n4l-todo-section">
                    <div class="n4l-section-header">
                        <span class="material-icons">warning</span>
                        <span>Points à vérifier (${result.todo_items.length})</span>
                    </div>
                    <div class="n4l-todo-list">
                        ${result.todo_items.map(item => `
                            <div class="n4l-todo-item">
                                <span class="material-icons">radio_button_unchecked</span>
                                <span>${item}</span>
                            </div>
                        `).join('')}
                    </div>
                </div>
            `;
        }

        metadataContainer.innerHTML = metadataHtml;

        // Add click handler with event delegation for alias items
        const aliasGrid = metadataContainer.querySelector('.n4l-alias-grid');
        if (aliasGrid) {
            aliasGrid.onclick = (e) => {
                const aliasItem = e.target.closest('.n4l-alias-item');
                if (aliasItem) {
                    const aliasKey = aliasItem.getAttribute('data-alias-key');
                    console.log('Alias clicked:', aliasKey);
                    if (aliasKey) {
                        this.focusN4LAlias(aliasKey);
                    }
                }
            };
        }
    },

    // Get icon for context type
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
            'relations': 'people'
        };

        const lowerCtx = context.toLowerCase();
        for (const [key, icon] of Object.entries(icons)) {
            if (lowerCtx.includes(key)) return icon;
        }
        return 'label';
    },

    // Toggle collapsible section
    toggleN4LSection(header) {
        const content = header.nextElementSibling;
        const icon = header.querySelector('.n4l-expand-icon');

        if (content.classList.contains('n4l-collapsed')) {
            content.classList.remove('n4l-collapsed');
            icon.textContent = 'expand_less';
        } else {
            content.classList.add('n4l-collapsed');
            icon.textContent = 'expand_more';
        }
    },

    // Highlight sequence on graph and show navigation
    highlightN4LSequence(seqIndex) {
        if (!this.lastN4LParse?.sequences || !this.n4lGraph) return;

        const sequence = this.lastN4LParse.sequences[seqIndex];
        if (!sequence) return;

        // Find nodes matching sequence labels
        const seqNodeIds = [];
        const allNodes = this.n4lGraphNodes?.get() || [];

        sequence.forEach(label => {
            const node = allNodes.find(n => n.label === label || n.label.includes(label));
            if (node) seqNodeIds.push(node.id);
        });

        if (seqNodeIds.length === 0) {
            this.showToast('Séquence non trouvée sur le graphe');
            return;
        }

        // Store sequence navigation state
        this.currentSequenceIndex = seqIndex;
        this.currentSequenceNodeIds = seqNodeIds;
        this.currentSequenceStep = 0;
        this.sequencePlayInterval = null;

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
        this.n4lGraphNodes?.update(nodeUpdates);

        // Show navigation controls
        const navEl = document.getElementById(`seq-nav-${seqIndex}`);
        if (navEl) {
            navEl.style.display = 'block';
        }

        // Navigate to first node
        this.navigateToSequenceStep(seqIndex, 0);

        this.showToast(`Chronologie: ${seqNodeIds.length} étapes`);
    },

    // Navigate within sequence
    navigateSequence(seqIndex, direction) {
        if (!this.currentSequenceNodeIds || this.currentSequenceIndex !== seqIndex) return;

        const newStep = this.currentSequenceStep + direction;
        if (newStep < 0 || newStep >= this.currentSequenceNodeIds.length) {
            // Loop around
            const loopedStep = newStep < 0 ? this.currentSequenceNodeIds.length - 1 : 0;
            this.navigateToSequenceStep(seqIndex, loopedStep);
        } else {
            this.navigateToSequenceStep(seqIndex, newStep);
        }
    },

    // Navigate to specific step
    navigateToSequenceStep(seqIndex, step) {
        if (!this.currentSequenceNodeIds) return;

        this.currentSequenceStep = step;
        const nodeId = this.currentSequenceNodeIds[step];

        // Update all nodes - highlight current step, show sequence nodes, hide others
        const allNodes = this.n4lGraphNodes?.get() || [];
        const seqNodeSet = new Set(this.currentSequenceNodeIds);

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
        this.n4lGraphNodes?.update(nodeUpdates);

        // Also hide edges not connected to sequence nodes
        const allEdges = this.n4lGraphEdges?.get() || [];
        const edgeUpdates = allEdges.map(edge => {
            const fromInSeq = seqNodeSet.has(edge.from);
            const toInSeq = seqNodeSet.has(edge.to);
            return {
                id: edge.id,
                hidden: !(fromInSeq && toInSeq)
            };
        });
        this.n4lGraphEdges?.update(edgeUpdates);

        // Focus on current node
        this.n4lGraph.focus(nodeId, { scale: 1.2, animation: { duration: 300 } });

        // Update UI
        const currentEl = document.getElementById(`seq-current-${seqIndex}`);
        if (currentEl) currentEl.textContent = step + 1;

        const labelEl = document.getElementById(`seq-label-${seqIndex}`);
        if (labelEl) {
            const node = allNodes.find(n => n.id === nodeId);
            labelEl.textContent = node ? node.label : nodeId;
        }
    },

    // Auto-play sequence
    playSequence(seqIndex) {
        if (this.sequencePlayInterval) {
            this.pauseSequence(seqIndex);
            return;
        }

        const playBtn = document.getElementById(`seq-play-${seqIndex}`);
        if (playBtn) {
            playBtn.querySelector('.material-icons').textContent = 'pause';
        }

        this.sequencePlayInterval = setInterval(() => {
            const nextStep = this.currentSequenceStep + 1;
            if (nextStep >= this.currentSequenceNodeIds.length) {
                this.pauseSequence(seqIndex);
                this.navigateToSequenceStep(seqIndex, 0);
            } else {
                this.navigateToSequenceStep(seqIndex, nextStep);
            }
        }, 1500);
    },

    // Pause sequence playback
    pauseSequence(seqIndex) {
        if (this.sequencePlayInterval) {
            clearInterval(this.sequencePlayInterval);
            this.sequencePlayInterval = null;
        }

        const playBtn = document.getElementById(`seq-play-${seqIndex}`);
        if (playBtn) {
            playBtn.querySelector('.material-icons').textContent = 'play_arrow';
        }
    },

    // Stop and close sequence navigation
    stopSequenceNav(seqIndex) {
        this.pauseSequence(seqIndex);

        // Hide navigation
        const navEl = document.getElementById(`seq-nav-${seqIndex}`);
        if (navEl) {
            navEl.style.display = 'none';
        }

        // Show all nodes again
        const allNodes = this.n4lGraphNodes?.get() || [];
        const nodeUpdates = allNodes.map(node => ({
            id: node.id,
            hidden: false
        }));
        this.n4lGraphNodes?.update(nodeUpdates);

        // Show all edges again
        const allEdges = this.n4lGraphEdges?.get() || [];
        const edgeUpdates = allEdges.map(edge => ({
            id: edge.id,
            hidden: false
        }));
        this.n4lGraphEdges?.update(edgeUpdates);

        // Reset graph colors
        this.resetN4LGraphFocus();

        // Clear state
        this.currentSequenceIndex = null;
        this.currentSequenceNodeIds = null;
        this.currentSequenceStep = 0;
    },

    // Focus on alias node
    focusN4LAlias(aliasKey) {
        if (!this.lastN4LParse?.aliases || !this.n4lGraph) {
            this.showToast('Graphe non disponible', 'warning');
            return;
        }

        let aliasValue = this.lastN4LParse.aliases[aliasKey];

        // Handle array values (alias can be an array)
        if (Array.isArray(aliasValue)) {
            aliasValue = aliasValue[0];
        }

        if (!aliasValue) {
            this.showToast(`Alias @${aliasKey} non défini`, 'warning');
            return;
        }

        // Extract the main label (before any parentheses with type info)
        // e.g., "Casino de Deauville (type) lieu" -> "Casino de Deauville"
        const mainLabel = aliasValue.split('(')[0].trim();

        const allNodes = this.n4lGraphNodes?.get() || [];

        // Try multiple matching strategies
        let node = allNodes.find(n => n.id === aliasKey);
        if (!node) node = allNodes.find(n => n.id === aliasValue);
        if (!node) node = allNodes.find(n => n.label === aliasValue);
        if (!node) node = allNodes.find(n => n.label === mainLabel);
        if (!node) node = allNodes.find(n => n.label && n.label.includes(mainLabel));
        if (!node) node = allNodes.find(n => n.id && n.id.includes(aliasKey));
        if (!node) node = allNodes.find(n => n.id === `@${aliasKey}`);

        if (node) {
            this.focusN4LGraphNode(node.id);
            this.n4lGraph.focus(node.id, { scale: 1.5, animation: true });
            this.showToast(`Focus: ${node.label || aliasKey}`);
        } else {
            this.showToast(`Alias @${aliasKey} non trouvé`, 'warning');
        }
    },

    // ============================================
    // N4L Filter and Reset
    // ============================================
    resetN4LFilter() {
        if (!this.n4lGraph) return;

        // Reset filter state
        this.n4lActiveContextFilter = null;

        // Reset all nodes
        const allNodes = this.n4lGraph.body.data.nodes.getIds();
        const nodeUpdates = allNodes.map(id => ({
            id,
            hidden: false,
            opacity: 1,
            font: { color: '#1a1a2e' }
        }));
        this.n4lGraph.body.data.nodes.update(nodeUpdates);

        // Reset all edges
        const allEdges = this.n4lGraph.body.data.edges.getIds();
        const edgeUpdates = allEdges.map(id => ({
            id,
            hidden: false,
            color: undefined
        }));
        this.n4lGraph.body.data.edges.update(edgeUpdates);

        // Also reset node colors
        this.resetN4LGraphFocus();

        // Refresh metadata display to update button states
        if (this.lastN4LParse) {
            this.showN4LMetadata(this.lastN4LParse);
        }

        // Recentrer le graphe
        setTimeout(() => {
            if (this.n4lGraph) {
                this.n4lGraph.fit({ animation: { duration: 300, easingFunction: 'easeInOutQuad' } });
            }
        }, 100);

        this.showToast('Filtre réinitialisé');
    },

    filterByContext(context) {
        if (!this.lastN4LParse || !this.n4lGraph) return;

        // Toggle filter if clicking same context
        if (this.n4lActiveContextFilter === context) {
            this.resetN4LFilter();
            return;
        }

        // Set active filter
        this.n4lActiveContextFilter = context;

        const result = this.lastN4LParse;
        const filteredEdges = result.graph.edges.filter(e =>
            !e.context || e.context === context || e.context.includes(context)
        );

        const involvedNodes = new Set();
        filteredEdges.forEach(e => {
            involvedNodes.add(e.from);
            involvedNodes.add(e.to);
        });

        // Ne pas cacher les nœuds, juste les rendre très transparents pour que fit() fonctionne
        const allNodes = this.n4lGraph.body.data.nodes.getIds();
        const visibleNodeIds = Array.from(involvedNodes);

        const nodeUpdates = allNodes.map(id => {
            const isVisible = involvedNodes.has(id);
            return {
                id,
                hidden: false, // Ne pas cacher pour que fit() fonctionne
                color: isVisible ? undefined : { background: 'rgba(200,200,200,0.1)', border: 'rgba(200,200,200,0.1)' },
                font: { color: isVisible ? '#1a1a2e' : 'rgba(0,0,0,0.05)' },
                opacity: isVisible ? 1 : 0.05
            };
        });
        this.n4lGraph.body.data.nodes.update(nodeUpdates);

        // Cacher aussi les arêtes non concernées
        const allEdges = this.n4lGraph.body.data.edges.getIds();
        const edgeUpdates = allEdges.map(edgeId => {
            const edge = this.n4lGraph.body.data.edges.get(edgeId);
            const isVisible = involvedNodes.has(edge.from) && involvedNodes.has(edge.to);
            return {
                id: edgeId,
                hidden: !isVisible,
                color: isVisible ? undefined : { color: 'rgba(200,200,200,0.05)' }
            };
        });
        this.n4lGraph.body.data.edges.update(edgeUpdates);

        // Refresh metadata display to update button states
        this.showN4LMetadata(this.lastN4LParse);

        // Recentrer le graphe sur les nœuds visibles
        setTimeout(() => {
            if (this.n4lGraph && visibleNodeIds.length > 0) {
                this.n4lGraph.fit({
                    nodes: visibleNodeIds,
                    animation: { duration: 400, easingFunction: 'easeInOutQuad' }
                });
            }
        }, 100);

        this.showToast(`Filtré: ${context} (${involvedNodes.size} entités)`);
    },

    // ============================================
    // N4L Graph Focus
    // ============================================
    focusN4LGraphNode(nodeId) {
        if (!this.n4lGraph || !this.n4lGraphNodes || !this.n4lGraphEdges) return;

        if (this.selectedN4LNode === nodeId) {
            this.resetN4LGraphFocus();
            return;
        }

        this.selectedN4LNode = nodeId;

        const connectedNodes = new Set([nodeId]);
        const connectedEdgeIds = new Set();

        this.n4lGraphEdges.forEach(edge => {
            if (edge.from === nodeId || edge.to === nodeId) {
                connectedNodes.add(edge.from);
                connectedNodes.add(edge.to);
                connectedEdgeIds.add(edge.id);
            }
        });

        const nodeUpdates = this.n4lGraphNodes.map(node => {
            const isConnected = connectedNodes.has(node.id);
            const originalColor = node.originalColor || this.getNodeColor(node);

            if (isConnected) {
                return {
                    id: node.id,
                    color: originalColor,
                    opacity: 1,
                    borderWidth: node.id === nodeId ? 4 : 2,
                    font: { color: '#1a1a2e', size: node.id === nodeId ? 14 : 12 }
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

        this.n4lGraphNodes.update(nodeUpdates);

        const edgeUpdates = this.n4lGraphEdges.map(edge => {
            const isConnected = connectedEdgeIds.has(edge.id);
            return {
                id: edge.id,
                color: isConnected ? { color: edge.originalColor || '#1e3a5f', opacity: 1 } : { color: '#e2e8f0', opacity: 0.2 },
                width: isConnected ? 2 : 1,
                font: { color: isConnected ? '#4a5568' : '#cbd5e0', size: isConnected ? 10 : 8 }
            };
        });

        this.n4lGraphEdges.update(edgeUpdates);

        const nodeData = this.n4lGraphNodesData.find(n => n.id === nodeId);
        if (nodeData) {
            this.showToast(`${nodeData.label} - ${connectedNodes.size - 1} connexion(s)`);
        }
    },

    resetN4LGraphFocus() {
        if (!this.n4lGraph || !this.n4lGraphNodes || !this.n4lGraphEdges) return;

        this.selectedN4LNode = null;

        const nodeUpdates = this.n4lGraphNodes.map(node => {
            const originalColor = node.originalColor || this.getNodeColor(node);
            return {
                id: node.id,
                color: originalColor,
                opacity: 1,
                borderWidth: 2,
                font: { color: '#1a1a2e', size: 12 }
            };
        });

        this.n4lGraphNodes.update(nodeUpdates);

        const edgeUpdates = this.n4lGraphEdges.map(edge => ({
            id: edge.id,
            color: { color: edge.originalColor || '#1e3a5f', opacity: 1 },
            width: 1,
            font: { color: '#4a5568', size: 10 }
        }));

        this.n4lGraphEdges.update(edgeUpdates);
    },

    // ============================================
    // N4L Context Menu
    // ============================================
    handleN4LGraphRightClick(params) {
        const nodeId = this.n4lGraph.getNodeAt(params.pointer.DOM);
        const menu = document.getElementById('graph-context-menu');

        if (!nodeId) {
            this.hideContextMenu();
            return;
        }

        this.contextMenuNodeId = nodeId;
        this.contextMenuGraphType = 'n4l';

        const x = params.event.clientX || params.pointer.DOM.x;
        const y = params.event.clientY || params.pointer.DOM.y;

        menu.style.left = `${x}px`;
        menu.style.top = `${y}px`;
        menu.classList.remove('hidden');

        const rect = menu.getBoundingClientRect();
        if (rect.right > window.innerWidth) {
            menu.style.left = `${x - rect.width}px`;
        }
        if (rect.bottom > window.innerHeight) {
            menu.style.top = `${y - rect.height}px`;
        }
    },

    // ============================================
    // N4L Legend and Reset Button
    // ============================================
    addN4LLegend(container) {
        const legend = document.createElement('div');
        legend.className = 'n4l-legend';
        legend.style.cssText = 'position: absolute; top: 10px; right: 10px; background: white; padding: 0.6rem; border-radius: 6px; font-size: 0.85rem; box-shadow: 0 2px 8px rgba(0,0,0,0.1);';
        legend.innerHTML = `
            <div style="font-weight: 600; margin-bottom: 0.35rem;">Types de relations</div>
            <div style="display: flex; align-items: center; gap: 0.35rem; margin-bottom: 0.15rem;"><span style="width: 25px; height: 3px; background: #1e3a5f;"></span> Standard</div>
            <div style="display: flex; align-items: center; gap: 0.35rem; margin-bottom: 0.15rem;"><span style="width: 25px; height: 3px; background: #059669;"></span> Nouveau (\\new)</div>
            <div style="display: flex; align-items: center; gap: 0.35rem; margin-bottom: 0.15rem;"><span style="width: 25px; height: 3px; background: #dc2626;"></span> Jamais (\\never)</div>
            <div style="display: flex; align-items: center; gap: 0.35rem; margin-bottom: 0.15rem;"><span style="width: 25px; height: 3px; background: #f59e0b;"></span> Séquence</div>
            <div style="display: flex; align-items: center; gap: 0.35rem;"><span style="width: 25px; height: 3px; background: #8b5cf6;"></span> Équivalence</div>
        `;
        container.style.position = 'relative';
        container.appendChild(legend);
    },

    // ============================================
    // Toggle Panel
    // ============================================
    togglePanel(panelId) {
        const panel = document.getElementById(panelId);
        if (!panel) return;

        const isCollapsed = panel.classList.contains('panel-collapsed');
        const container = panel.parentElement;
        const isEditor = panelId.includes('editor');

        if (isCollapsed) {
            panel.classList.remove('panel-collapsed');
            container.classList.remove('has-collapsed', 'has-collapsed-right');
            const icon = panel.querySelector('.btn-icon .material-icons');
            if (icon) icon.textContent = isEditor ? 'chevron_left' : 'chevron_right';
        } else {
            panel.classList.add('panel-collapsed');
            if (isEditor) {
                container.classList.add('has-collapsed');
                container.classList.remove('has-collapsed-right');
            } else {
                container.classList.add('has-collapsed-right');
                container.classList.remove('has-collapsed');
            }
            const icon = panel.querySelector('.btn-icon .material-icons');
            if (icon) icon.textContent = isEditor ? 'chevron_right' : 'chevron_left';
        }

        setTimeout(() => {
            if (this.n4lGraph) {
                this.n4lGraph.redraw();
                this.n4lGraph.fit();
            }
        }, 350);
    },

    // ============================================
    // Export N4L
    // ============================================
    exportN4L() {
        const content = document.getElementById('n4l-editor').value;
        if (!content) return;

        const blob = new Blob([content], { type: 'text/plain' });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `${this.currentCase?.name || 'affaire'}.n4l`;
        a.click();
        URL.revokeObjectURL(url);
    },

    // ============================================
    // Import N4L Modal
    // ============================================
    showImportN4LModal() {
        if (!this.currentCase) {
            this.showToast('Sélectionnez une affaire d\'abord', 'warning');
            return;
        }

        const content = `
            <div class="modal-explanation">
                <span class="material-icons">info</span>
                <p><strong>Importer N4L</strong> - Importez un fichier N4L existant pour enrichir l'affaire avec des données structurées.</p>
            </div>
            <div class="form-group">
                <label class="form-label">Fichier N4L</label>
                <input type="file" class="form-input" id="n4l-file" accept=".n4l,.txt">
            </div>
            <div class="form-group">
                <label class="form-label">Ou collez le contenu N4L</label>
                <textarea class="form-textarea" id="n4l-content" rows="10" placeholder="Collez votre contenu N4L ici..."></textarea>
            </div>
        `;

        this.showModal('Importer N4L', content, async () => {
            const fileInput = document.getElementById('n4l-file');
            const textContent = document.getElementById('n4l-content').value;

            let n4lContent = textContent;

            if (fileInput.files.length > 0) {
                n4lContent = await fileInput.files[0].text();
            }

            if (!n4lContent) {
                this.showToast('Aucun contenu N4L fourni', 'warning');
                return;
            }

            try {
                const response = await fetch(`/api/n4l/import?case_id=${this.currentCase.id}`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'text/plain' },
                    body: n4lContent
                });

                if (!response.ok) throw new Error('Erreur import N4L');

                await this.selectCase(this.currentCase.id);
                this.showToast('N4L importé avec succès', 'success');
            } catch (error) {
                console.error('Error importing N4L:', error);
                this.showToast('Erreur lors de l\'import: ' + error.message, 'error');
            }
        });
    },

    // ============================================
    // N4L Expansion Cone
    // ============================================
    showN4LExpansionConeModal(nodeId) {
        const node = this.n4lGraphNodes?.get(nodeId);
        if (!node) return;

        const content = `
            <div class="modal-explanation">
                <span class="material-icons">info</span>
                <p><strong>Cône d'expansion N4L</strong> - Visualisez tous les noeuds connectés à "${node.label}".</p>
            </div>
            <div class="form-group">
                <label class="form-label">Profondeur :</label>
                <select id="n4l-cone-depth" class="form-select">
                    <option value="1">1 niveau</option>
                    <option value="2" selected>2 niveaux</option>
                    <option value="3">3 niveaux</option>
                </select>
            </div>
        `;

        this.showModal('Explorer le Voisinage N4L', content, () => {
            const depth = parseInt(document.getElementById('n4l-cone-depth').value);
            this.showN4LExpansionCone(nodeId, depth);
        });
    },

    showN4LExpansionCone(nodeId, depth) {
        if (!this.n4lGraph || !this.n4lGraphNodes || !this.n4lGraphEdges) return;

        const coneNodes = new Set([nodeId]);
        const coneEdges = new Set();
        let currentLevel = new Set([nodeId]);

        const allEdges = this.n4lGraphEdges.get();

        for (let d = 0; d < depth; d++) {
            const nextLevel = new Set();
            allEdges.forEach(edge => {
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

        const allNodes = this.n4lGraphNodes.get();
        const nodeUpdates = allNodes.map(node => {
            const inCone = coneNodes.has(node.id);
            const isCenter = node.id === nodeId;

            if (isCenter) {
                return { id: node.id, color: { background: '#1e3a5f', border: '#152a45' }, borderWidth: 4 };
            } else if (inCone) {
                return { id: node.id, color: { background: '#3b82f6', border: '#1e3a5f' }, borderWidth: 3, opacity: 1 };
            } else {
                return { id: node.id, color: { background: '#e2e8f0', border: '#cbd5e0' }, opacity: 0.3, borderWidth: 1 };
            }
        });

        this.n4lGraphNodes.update(nodeUpdates);

        const edgeUpdates = allEdges.map(edge => ({
            id: edge.id,
            color: coneEdges.has(edge.id) ? { color: '#1e3a5f', opacity: 1 } : { color: '#cbd5e0', opacity: 0.2 },
            width: coneEdges.has(edge.id) ? 2 : 1
        }));

        this.n4lGraphEdges.update(edgeUpdates);
        this.n4lGraph.focus(nodeId, { scale: 1.2, animation: true });
        this.showToast(`Cône N4L: ${coneNodes.size} noeuds, profondeur ${depth}`);
    },

    async analyzeN4LExpansionCone(nodeId) {
        const node = this.n4lGraphNodes?.get(nodeId);
        if (!node) return;

        const coneNodes = new Set([nodeId]);
        let currentLevel = new Set([nodeId]);

        for (let d = 0; d < 2; d++) {
            const nextLevel = new Set();
            this.n4lGraphEdges?.forEach(edge => {
                if (currentLevel.has(edge.from)) { nextLevel.add(edge.to); coneNodes.add(edge.to); }
                if (currentLevel.has(edge.to)) { nextLevel.add(edge.from); coneNodes.add(edge.from); }
            });
            currentLevel = nextLevel;
        }

        const nodeLabels = Array.from(coneNodes).map(id => {
            const n = this.n4lGraphNodes?.get(id);
            return n ? n.label : id;
        });

        this.setAnalysisContext('n4l_cone_analysis', `Analyse N4L - ${node.label}`, `Centre: ${node.label}`);

        const analysisModal = document.getElementById('analysis-modal');
        const analysisContent = document.getElementById('analysis-content');
        const modalTitle = document.getElementById('analysis-modal-title');
        if (modalTitle) modalTitle.textContent = 'Analyse du voisinage N4L';

        analysisContent.innerHTML = `
            <div class="modal-explanation">
                <span class="material-icons">hub</span>
                <p><strong>Voisinage de ${node.label}:</strong> ${nodeLabels.join(', ')}</p>
            </div>
            <div id="n4l-cone-analysis"><span class="streaming-cursor">▊</span></div>
        `;
        analysisModal.classList.add('active');

        await this.streamAIResponse(
            '/api/chat/stream',
            { case_id: this.currentCase?.id, message: `Analyse le voisinage N4L de ${node.label}: ${nodeLabels.join(', ')}` },
            document.getElementById('n4l-cone-analysis')
        );
    },

    // ============================================
    // N4L Path Finding
    // ============================================
    showFindN4LPathsFromNode(fromNodeId) {
        const fromNode = this.n4lGraphNodes?.get(fromNodeId);
        if (!fromNode) return;

        const otherNodes = [];
        this.n4lGraphNodes?.forEach(node => {
            if (node.id !== fromNodeId) {
                otherNodes.push(`<option value="${node.id}">${node.label}</option>`);
            }
        });

        const content = `
            <div class="modal-explanation">
                <span class="material-icons">info</span>
                <p><strong>Trouver les chemins N4L</strong> - Découvrez les chemins entre "${fromNode.label}" et une autre entité.</p>
            </div>
            <div class="form-group">
                <label class="form-label">De: ${fromNode.label}</label>
            </div>
            <div class="form-group">
                <label class="form-label">Vers:</label>
                <select id="n4l-path-to-node" class="form-select">
                    <option value="">-- Sélectionner --</option>
                    ${otherNodes.join('')}
                </select>
            </div>
            <div class="form-group">
                <label class="form-label">Profondeur max:</label>
                <select id="n4l-path-max-depth" class="form-select">
                    <option value="3">3 niveaux</option>
                    <option value="4" selected>4 niveaux</option>
                    <option value="5">5 niveaux</option>
                </select>
            </div>
        `;

        this.showModal('Trouver les Chemins N4L', content, async () => {
            const toNodeId = document.getElementById('n4l-path-to-node').value;
            const maxDepth = parseInt(document.getElementById('n4l-path-max-depth').value);

            if (!toNodeId) {
                alert('Veuillez sélectionner une destination');
                return;
            }

            this.closeModal();
            await this.findAndDisplayN4LPaths(fromNodeId, toNodeId, maxDepth);
        });
    },

    async findAndDisplayN4LPaths(fromId, toId, maxDepth) {
        const paths = this.findN4LPaths(fromId, toId, maxDepth);
        const fromNode = this.n4lGraphNodes?.get(fromId);
        const toNode = this.n4lGraphNodes?.get(toId);

        if (paths.length === 0) {
            this.showToast(`Aucun chemin N4L trouvé entre ${fromNode?.label} et ${toNode?.label}`);
            return;
        }

        const pathsHtml = paths.map((path, index) => {
            const pathNodes = path.map(id => {
                const n = this.n4lGraphNodes?.get(id);
                return n ? n.label : id;
            }).join(' → ');

            return `
                <div class="path-item" data-path-index="${index}">
                    <div class="path-nodes">${pathNodes}</div>
                    <div class="path-actions">
                        <button class="btn btn-sm btn-secondary btn-icon" onclick="app.highlightN4LPath(${index})" data-tooltip="Afficher ce chemin sur le graphe">
                            <span class="material-icons">visibility</span>
                        </button>
                        <button class="btn btn-sm btn-secondary btn-icon" onclick="app.analyzeN4LPath(${index})" data-tooltip="Analyser ce chemin avec l'IA">
                            <span class="material-icons">psychology</span>
                        </button>
                    </div>
                </div>
            `;
        }).join('');

        this.discoveredN4LPaths = paths;

        this.showModal(`${paths.length} Chemin(s) N4L Trouvé(s)`, `
            <div class="modal-explanation">
                <span class="material-icons">info</span>
                <p>Chemins entre <strong>${fromNode?.label}</strong> et <strong>${toNode?.label}</strong>.</p>
            </div>
            <div class="paths-list">${pathsHtml}</div>
        `);
    },

    findN4LPaths(fromId, toId, maxDepth) {
        const paths = [];
        const visited = new Set();

        const dfs = (current, target, path, depth) => {
            if (depth > maxDepth) return;
            if (current === target) {
                paths.push([...path]);
                return;
            }

            visited.add(current);

            this.n4lGraphEdges?.forEach(edge => {
                let next = null;
                if (edge.from === current && !visited.has(edge.to)) next = edge.to;
                else if (edge.to === current && !visited.has(edge.from)) next = edge.from;

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

    highlightN4LPath(pathIndex) {
        if (!this.discoveredN4LPaths || !this.discoveredN4LPaths[pathIndex]) return;

        const path = this.discoveredN4LPaths[pathIndex];
        const pathSet = new Set(path);

        const pathEdges = new Set();
        for (let i = 0; i < path.length - 1; i++) {
            this.n4lGraphEdges?.forEach(edge => {
                if ((edge.from === path[i] && edge.to === path[i + 1]) ||
                    (edge.to === path[i] && edge.from === path[i + 1])) {
                    pathEdges.add(edge.id);
                }
            });
        }

        const nodeUpdates = this.n4lGraphNodes?.get().map(node => {
            const inPath = pathSet.has(node.id);
            if (inPath) {
                return { id: node.id, color: { background: '#f59e0b', border: '#d97706' }, borderWidth: 4 };
            } else {
                return { id: node.id, color: { background: '#e2e8f0', border: '#cbd5e0' }, opacity: 0.3, borderWidth: 1 };
            }
        });

        this.n4lGraphNodes?.update(nodeUpdates);

        const edgeUpdates = this.n4lGraphEdges?.get().map(edge => ({
            id: edge.id,
            color: pathEdges.has(edge.id) ? { color: '#f59e0b', opacity: 1 } : { color: '#cbd5e0', opacity: 0.2 },
            width: pathEdges.has(edge.id) ? 3 : 1
        }));

        this.n4lGraphEdges?.update(edgeUpdates);
        this.showToast(`Chemin N4L ${pathIndex + 1} affiché`);
    },

    async analyzeN4LPath(pathIndex) {
        if (!this.discoveredN4LPaths || !this.discoveredN4LPaths[pathIndex]) return;

        const path = this.discoveredN4LPaths[pathIndex];
        const pathLabels = path.map(id => {
            const n = this.n4lGraphNodes?.get(id);
            return n ? n.label : id;
        });

        this.setAnalysisContext('n4l_path_analysis', 'Analyse du Chemin N4L', `Chemin: ${pathLabels.join(' → ')}`);

        const analysisModal = document.getElementById('analysis-modal');
        const analysisContent = document.getElementById('analysis-content');
        const modalTitle = document.getElementById('analysis-modal-title');
        if (modalTitle) modalTitle.textContent = 'Analyse du Chemin N4L';

        analysisContent.innerHTML = `
            <div class="modal-explanation">
                <span class="material-icons">route</span>
                <p><strong>Chemin:</strong> ${pathLabels.join(' → ')}</p>
            </div>
            <div id="n4l-path-analysis"><span class="streaming-cursor">▊</span></div>
        `;
        analysisModal.classList.add('active');

        await this.streamAIResponse(
            '/api/chat/stream',
            { case_id: this.currentCase?.id, message: `Analyse ce chemin N4L: ${pathLabels.join(' → ')}` },
            document.getElementById('n4l-path-analysis')
        );
    },

    // ============================================
    // Toggle N4L Fullscreen
    // ============================================
    toggleN4LFullscreen() {
        const panel = document.getElementById('panel-n4l-graph');
        const btn = document.getElementById('btn-fullscreen-n4l');
        const icon = btn?.querySelector('.material-icons');

        if (!panel) return;

        if (panel.classList.contains('fullscreen-panel')) {
            // Exit fullscreen
            panel.classList.remove('fullscreen-panel');
            if (icon) icon.textContent = 'fullscreen';
            if (btn) btn.setAttribute('data-tooltip', 'Plein écran');
        } else {
            // Enter fullscreen
            panel.classList.add('fullscreen-panel');
            if (icon) icon.textContent = 'fullscreen_exit';
            if (btn) btn.setAttribute('data-tooltip', 'Quitter le plein écran');
        }

        // Redraw graph after transition
        setTimeout(() => {
            if (this.n4lGraph) {
                this.n4lGraph.redraw();
                this.n4lGraph.fit();
            }
        }, 350);
    }
};

// Export for use in main app
if (typeof module !== 'undefined' && module.exports) {
    module.exports = N4LModule;
}
