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

            // Le graphe est maintenant affiché uniquement dans le Dashboard
            // Mettre à jour le graphe du dashboard si disponible
            if (result.graph && result.graph.nodes && result.graph.nodes.length > 0) {
                // Rendre dans le conteneur du dashboard (graphe principal)
                this.renderN4LGraphInContainer(result.graph, 'dashboard-graph-container');
                this.showN4LMetadataInContainer(result, 'dashboard-metadata-container');
            }

            console.log('N4LModule: Graphe parsé', {
                nodes: result.graph?.nodes?.length || 0,
                edges: result.graph?.edges?.length || 0
            });

            this.showToast('N4L parsé - graphe mis à jour dans le Tableau de bord', 'success');
        } catch (error) {
            console.error('Error parsing N4L:', error);
        }
    },

    // ============================================
    // Render N4L Graph (utilise le même style que le graphe principal)
    // ============================================
    // ============================================
    // N4L Metadata Display (legacy - redirige vers dashboard)
    // ============================================
    showN4LMetadata(result) {
        // Le panneau N4L a été supprimé - utiliser le conteneur du dashboard
        this.showN4LMetadataInContainer(result, 'dashboard-metadata-container');
    },

    // Fonction legacy pour compatibilité - redirige vers dashboard
    renderN4LGraph(graphData) {
        // Le panneau N4L a été supprimé - utiliser le conteneur du dashboard
        this.renderN4LGraphInContainer(graphData, 'dashboard-graph-container');
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

    // Toggle alias "voir plus" to show all aliases
    toggleN4LAliasMore(element) {
        const hiddenContainer = document.getElementById('n4l-alias-hidden');
        const isExpanded = element.dataset.expanded === 'true';
        const icon = element.querySelector('.material-icons');
        const textSpan = element.querySelector('.n4l-alias-more-text');

        if (isExpanded) {
            // Collapse - hide extra aliases
            if (hiddenContainer) hiddenContainer.style.display = 'none';
            element.dataset.expanded = 'false';
            if (icon) icon.textContent = 'expand_more';
            // Restore original text
            const totalHidden = hiddenContainer?.querySelectorAll('.n4l-alias-item').length || 0;
            if (textSpan) textSpan.textContent = `+${totalHidden} autres...`;
        } else {
            // Expand - show all aliases
            if (hiddenContainer) hiddenContainer.style.display = 'contents';
            element.dataset.expanded = 'true';
            if (icon) icon.textContent = 'expand_less';
            if (textSpan) textSpan.textContent = 'Réduire';
        }
    },

    // Toggle cross-references "voir plus" to show all references
    toggleN4LCrossRefsMore(element) {
        const hiddenContainer = element.previousElementSibling;
        const icon = element.querySelector('.material-icons');
        const textSpan = element.querySelector('.n4l-crossref-more-text');

        if (element.dataset.expanded === 'true') {
            // Collapse - hide extra refs
            if (hiddenContainer) hiddenContainer.style.display = 'none';
            element.dataset.expanded = 'false';
            if (icon) icon.textContent = 'expand_more';
            // Restore original text
            const totalHidden = hiddenContainer?.querySelectorAll('.n4l-crossref-item').length || 0;
            if (textSpan) textSpan.textContent = `+${totalHidden} autres...`;
        } else {
            // Expand - show all refs
            if (hiddenContainer) hiddenContainer.style.display = 'contents';
            element.dataset.expanded = 'true';
            if (icon) icon.textContent = 'expand_less';
            if (textSpan) textSpan.textContent = 'Réduire';
        }
    },

    // Highlight sequence on graph and show navigation
    highlightN4LSequence(seqIndex) {
        if (!this.lastN4LParse?.sequences || !this.n4lGraph) return;

        const sequence = this.lastN4LParse.sequences[seqIndex];
        if (!sequence) return;

        // Always reset the graph first to ensure all nodes are visible
        if (this.isInSpecialView) {
            this.isInSpecialView = false;
            this.restoreFullGraph();
        } else {
            this.resetGraphStateQuiet();
        }
        this.n4lActiveContextFilter = null;

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

    // Silent reset of graph state (no toast, no fit animation)
    // Used before applying a new filter to clear previous styling
    resetGraphStateQuiet() {
        if (!this.n4lGraph) return;

        // Reset all nodes to original colors
        const allNodes = this.n4lGraph.body.data.nodes.getIds();
        const nodeUpdates = allNodes.map(id => ({
            id,
            hidden: false,
            opacity: 1,
            color: this.getOriginalNodeColor(id),
            font: { color: '#1a1a2e' },
            borderWidth: 2
        }));
        this.n4lGraph.body.data.nodes.update(nodeUpdates);

        // Reset all edges
        const allEdges = this.n4lGraph.body.data.edges.getIds();
        const edgeUpdates = allEdges.map(id => ({
            id,
            hidden: false,
            color: this.getOriginalEdgeColor(id),
            width: 1
        }));
        this.n4lGraph.body.data.edges.update(edgeUpdates);
    },

    // Get original node color based on context
    getOriginalNodeColor(nodeId) {
        if (!this.n4lGraphNodesOriginal) {
            // If no original cache, use context-based colors
            const node = this.n4lGraphNodes?.get(nodeId);
            if (node && node.context) {
                return this.getContextColor(node.context);
            }
            return { background: '#667eea', border: '#5a67d8' };
        }
        const original = this.n4lGraphNodesOriginal.get(nodeId);
        return original?.color || { background: '#667eea', border: '#5a67d8' };
    },

    // Get original edge color
    getOriginalEdgeColor(edgeId) {
        if (!this.n4lGraphEdgesOriginal) {
            return { color: '#94a3b8', opacity: 0.6 };
        }
        const original = this.n4lGraphEdgesOriginal.get(edgeId);
        return original?.color || { color: '#94a3b8', opacity: 0.6 };
    },

    // Get context-based color for nodes
    getContextColor(context) {
        const contextColors = {
            'victimes': { background: '#ef4444', border: '#dc2626' },
            'suspects': { background: '#f97316', border: '#ea580c' },
            'témoins': { background: '#eab308', border: '#ca8a04' },
            'temoins': { background: '#eab308', border: '#ca8a04' },
            'lieux': { background: '#22c55e', border: '#16a34a' },
            'preuves': { background: '#3b82f6', border: '#2563eb' },
            'chronologie': { background: '#8b5cf6', border: '#7c3aed' },
            'hypothèses': { background: '#ec4899', border: '#db2777' },
            'hypotheses': { background: '#ec4899', border: '#db2777' },
            'general': { background: '#667eea', border: '#5a67d8' }
        };
        // Check for partial match
        const lowerContext = (context || '').toLowerCase();
        for (const [key, color] of Object.entries(contextColors)) {
            if (lowerContext.includes(key)) {
                return color;
            }
        }
        return contextColors['general'];
    },

    resetN4LFilter() {
        if (!this.n4lGraph) return;

        // Check if we're in special mode (causal chains, hypotheses, or TODO) - need to restore full graph
        const specialFilters = ['chaînes causales', 'hypothèses, pistes', 'pistes, hypothèses', 'TODO, notes', 'notes, TODO'];
        if (this.isInSpecialView || this.currentCausalChain || specialFilters.includes(this.n4lActiveContextFilter)) {
            this.isInSpecialView = false;
            this.restoreFullGraph();
            this.n4lActiveContextFilter = null;
            // Refresh metadata display to update button states
            if (this.lastN4LParse) {
                this.showN4LMetadata(this.lastN4LParse);
            }
            return;
        }

        // Reset filter state
        this.n4lActiveContextFilter = null;
        this.isInSpecialView = false;

        // Reset graph state using the quiet method
        this.resetGraphStateQuiet();

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

        // Special case: "chaînes causales" - show all causal chains combined
        if (context === 'chaînes causales') {
            this.isInSpecialView = true;
            this.showAllCausalChains();
            return;
        }

        // Special case: "hypothèses" or "pistes" - show hypothesis aliases as graph
        if (context.includes('hypothèse') || context.includes('piste')) {
            this.isInSpecialView = true;
            this.showHypothesesGraph();
            return;
        }

        // Special case: "TODO" or "notes" - show TODO items as graph
        if (context.toLowerCase().includes('todo') || context.toLowerCase().includes('note')) {
            this.isInSpecialView = true;
            this.showTodoGraph();
            return;
        }

        // If coming from a special view, do a full reset first
        if (this.isInSpecialView) {
            this.isInSpecialView = false;
            this.resetN4LFilter();
        }

        // IMPORTANT: Reset graph state before applying new filter
        // This ensures previous filter's styling is cleared
        this.resetGraphStateQuiet();

        // Set active filter
        this.n4lActiveContextFilter = context;

        const result = this.lastN4LParse;

        // Find nodes that match the context directly
        const involvedNodes = new Set();

        // First, add nodes that have this context
        result.graph.nodes.forEach(n => {
            if (n.context && (n.context === context || n.context.includes(context) || context.includes(n.context))) {
                involvedNodes.add(n.id);
            }
        });

        // Also check subjects for matching context
        if (result.subjects) {
            result.subjects.forEach(s => {
                if (s.context && (s.context === context || s.context.includes(context) || context.includes(s.context))) {
                    involvedNodes.add(s.name);
                }
            });
        }

        // Find edges that match the context
        const filteredEdges = result.graph.edges.filter(e =>
            e.context && (e.context === context || e.context.includes(context) || context.includes(e.context))
        );

        // Add nodes connected by filtered edges
        filteredEdges.forEach(e => {
            involvedNodes.add(e.from);
            involvedNodes.add(e.to);
        });

        // Check if we found anything
        if (involvedNodes.size === 0) {
            this.showToast(`Aucune entité trouvée pour le contexte "${context}"`, 'warning');
            this.n4lActiveContextFilter = null;
            return;
        }

        // Ne pas cacher les nœuds, juste les rendre très transparents pour que fit() fonctionne
        const allNodes = this.n4lGraph.body.data.nodes.getIds();
        const visibleNodeIds = Array.from(involvedNodes);

        const nodeUpdates = allNodes.map(id => {
            const isVisible = involvedNodes.has(id);
            return {
                id,
                hidden: false, // Ne pas cacher pour que fit() fonctionne
                color: isVisible ? this.getOriginalNodeColor(id) : { background: 'rgba(200,200,200,0.1)', border: 'rgba(200,200,200,0.1)' },
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
        if (!this.n4lGraph || !this.n4lGraphNodes || !this.n4lGraphEdges || !this.n4lGraphNodesData) return;

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

        // Centrer la vue sur le nœud sélectionné (sans changer le zoom)
        const nodePosition = this.n4lGraph.getPositions([nodeId])[nodeId];
        if (nodePosition) {
            this.n4lGraph.moveTo({
                position: { x: nodePosition.x, y: nodePosition.y },
                animation: false
            });
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

    // Focus on a marker node by its label text
    focusMarkerNode(markerText) {
        // Try dashboard graph first (since Graph Preview was removed)
        if (this.dashboardGraph && this.dashboardGraphNodes) {
            const allNodes = this.dashboardGraphNodes.get();
            let node = allNodes.find(n => n.label === markerText);
            if (!node) node = allNodes.find(n => n.label && n.label.toLowerCase() === markerText.toLowerCase());
            if (!node) node = allNodes.find(n => n.label && n.label.includes(markerText));

            if (node) {
                this.dashboardGraph.selectNodes([node.id]);
                this.dashboardGraph.focus(node.id, { scale: 1.5, animation: { duration: 300, easingFunction: 'easeInOutQuad' } });
                this.showToast(`Nœud "${node.label}" sélectionné`);
                return;
            }
        }

        // Fallback to N4L graph if available
        if (this.n4lGraph && this.n4lGraphNodes) {
            const allNodes = this.n4lGraphNodes.get();
            let node = allNodes.find(n => n.label === markerText);
            if (!node) node = allNodes.find(n => n.label && n.label.toLowerCase() === markerText.toLowerCase());
            if (!node) node = allNodes.find(n => n.label && n.label.includes(markerText));

            if (node) {
                this.focusN4LGraphNode(node.id);
                this.showToast(`Nœud "${node.label}" sélectionné`);
                return;
            }
        }

        this.showToast(`Nœud "${markerText}" non trouvé dans le graphe`, 'info');
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
        // Legend is now displayed in the metadata panel (n4l-section)
        // This function is kept for compatibility but does nothing
        // Remove any existing floating legend if present
        const existingLegend = container.querySelector('.n4l-legend');
        if (existingLegend) existingLegend.remove();
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
            this.showToast(`Aucun chemin trouvé entre ${fromNode?.label || fromId} et ${toNode?.label || toId}`);
            return;
        }

        // Store paths for later use
        this.discoveredN4LPaths = paths;

        // Highlight the first/shortest path directly on the graph
        this.highlightN4LPath(0);

        // Show toast with path info
        const pathLabels = paths[0].map(id => {
            const n = this.n4lGraphNodes?.get(id);
            return n ? n.label : id;
        }).join(' → ');

        this.showToast(`${paths.length} chemin(s) trouvé(s): ${pathLabels}`, 'success');
    },

    findN4LPaths(fromId, toId, maxDepth) {
        const paths = [];
        const visited = new Set();

        // Get edges as array from DataSet
        const edges = this.n4lGraphEdges?.get() || [];

        const dfs = (current, target, path, depth) => {
            if (depth > maxDepth) return;
            if (current === target) {
                paths.push([...path]);
                return;
            }

            visited.add(current);

            edges.forEach(edge => {
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

        // Get edges as array from DataSet
        const edges = this.n4lGraphEdges?.get() || [];

        const pathEdges = new Set();
        for (let i = 0; i < path.length - 1; i++) {
            edges.forEach(edge => {
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

        const edgeUpdates = edges.map(edge => ({
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
    // N4L Advanced Features - Causal Chains, Cross Refs, Markers
    // ============================================

    // Get STType label for display
    getSTTypeLabel(stType) {
        const labels = {
            0: 'Proximité',    // STNear
            1: 'Causalité',    // STLeadsTo
            2: 'Contient',     // STContains
            3: 'Expression'    // STExpresses
        };
        return labels[stType] || 'Relation';
    },

    // Highlight a causal chain on the graph - creates a dedicated chain visualization
    highlightCausalChain(chainIndex) {
        if (!this.lastN4LParse?.causal_chains) {
            console.warn('highlightCausalChain: No causal chains data');
            return;
        }

        const chain = this.lastN4LParse.causal_chains[chainIndex];
        if (!chain || !chain.steps || chain.steps.length === 0) {
            console.warn('highlightCausalChain: Chain not found or empty', chainIndex);
            return;
        }

        // Check if graph is initialized
        if (!this.n4lGraph || !this.n4lGraphNodes || !this.n4lGraphEdges) {
            console.log('highlightCausalChain: Graph not initialized, rendering first');
            // Try to render the graph first
            if (this.lastN4LParse?.graph) {
                this.renderN4LGraph(this.lastN4LParse.graph);
            } else {
                this.showToast('Graphe non initialisé, cliquez d\'abord sur "Afficher le graphe"', 'warning');
                return;
            }
        }

        // Create nodes for each step in the chain
        const chainNodes = chain.steps.map((step, i) => {
            const hue = 30 + (i / chain.steps.length) * 30; // Orange to yellow gradient
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

        // Create edges between consecutive steps with relation labels
        const chainEdges = [];
        for (let i = 0; i < chain.steps.length - 1; i++) {
            const step = chain.steps[i];
            chainEdges.push({
                id: `chain_edge_${i}`,
                from: `chain_step_${i}`,
                to: `chain_step_${i + 1}`,
                label: step.relation || '',
                arrows: 'to',
                color: { color: '#f59e0b', opacity: 1 },
                width: 3,
                font: { size: 12, color: '#666', strokeWidth: 2, strokeColor: '#fff' },
                smooth: { type: 'curvedCW', roundness: 0.1 }
            });
        }

        console.log('highlightCausalChain: Creating chain with', chainNodes.length, 'nodes and', chainEdges.length, 'edges');
        console.log('highlightCausalChain: Nodes:', chainNodes);
        console.log('highlightCausalChain: Edges:', chainEdges);

        // Update the graph with chain nodes and edges
        try {
            this.n4lGraphNodes.clear();
            this.n4lGraphEdges.clear();
            this.n4lGraphNodes.add(chainNodes);
            this.n4lGraphEdges.add(chainEdges);
            console.log('highlightCausalChain: Graph updated successfully');
        } catch (err) {
            console.error('highlightCausalChain: Error updating graph:', err);
            return;
        }

        // Store current chain state
        this.currentCausalChain = {
            index: chainIndex,
            nodeIds: chainNodes.map(n => n.id),
            step: 0
        };

        // Fit to show all chain nodes
        setTimeout(() => {
            if (this.n4lGraph) {
                this.n4lGraph.fit({
                    animation: { duration: 400, easingFunction: 'easeInOutQuad' }
                });
                console.log('highlightCausalChain: Graph fitted');
            }
        }, 100);

        // Show restore button
        const restoreBtn = document.getElementById('n4l-restore-graph-btn');
        if (restoreBtn) restoreBtn.style.display = 'flex';

        this.showToast(`Chaîne causale: ${chain.id || chainIndex + 1} (${chain.steps.length} étapes)`);
    },

    // Restore full graph after viewing a causal chain
    restoreFullGraph() {
        if (!this.lastN4LParse?.graph) return;

        this.currentCausalChain = null;

        // Hide restore button
        const restoreBtn = document.getElementById('n4l-restore-graph-btn');
        if (restoreBtn) restoreBtn.style.display = 'none';

        // Re-render the full graph with proper node/edge formatting
        this.renderN4LGraph(this.lastN4LParse.graph);

        this.showToast('Graphe complet restauré');
    },

    // Show all causal chains combined in the graph
    showAllCausalChains() {
        if (!this.lastN4LParse?.causal_chains || this.lastN4LParse.causal_chains.length === 0) {
            this.showToast('Aucune chaîne causale trouvée', 'warning');
            return;
        }

        // Check if graph is initialized
        if (!this.n4lGraph || !this.n4lGraphNodes || !this.n4lGraphEdges) {
            if (this.lastN4LParse?.graph) {
                this.renderN4LGraph(this.lastN4LParse.graph);
            } else {
                this.showToast('Graphe non initialisé', 'warning');
                return;
            }
        }

        const chains = this.lastN4LParse.causal_chains;
        const allNodes = [];
        const allEdges = [];

        // Color palette for different chains
        const chainColors = [
            { bg: '#f97316', border: '#ea580c' },  // Orange
            { bg: '#3b82f6', border: '#2563eb' },  // Blue
            { bg: '#10b981', border: '#059669' },  // Green
            { bg: '#8b5cf6', border: '#7c3aed' },  // Purple
            { bg: '#ec4899', border: '#db2777' },  // Pink
            { bg: '#06b6d4', border: '#0891b2' }   // Cyan
        ];

        chains.forEach((chain, chainIndex) => {
            const color = chainColors[chainIndex % chainColors.length];
            const yOffset = chainIndex * 150; // Vertical spacing between chains

            chain.steps.forEach((step, stepIndex) => {
                const nodeId = `chain_${chainIndex}_step_${stepIndex}`;
                allNodes.push({
                    id: nodeId,
                    label: step.item,
                    color: { background: color.bg, border: color.border },
                    borderWidth: 3,
                    font: { size: 13, color: '#fff' },
                    shape: 'box',
                    margin: 8,
                    x: stepIndex * 200,
                    y: yOffset,
                    fixed: { x: false, y: false }
                });

                // Create edge to next step
                if (stepIndex < chain.steps.length - 1) {
                    allEdges.push({
                        id: `chain_${chainIndex}_edge_${stepIndex}`,
                        from: nodeId,
                        to: `chain_${chainIndex}_step_${stepIndex + 1}`,
                        label: step.relation || '',
                        arrows: 'to',
                        color: { color: color.bg, opacity: 1 },
                        width: 3,
                        font: { size: 11, color: '#666', strokeWidth: 2, strokeColor: '#fff' },
                        smooth: { type: 'curvedCW', roundness: 0.1 }
                    });
                }
            });

            // Add chain label node
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

        // Update graph
        this.n4lGraphNodes.clear();
        this.n4lGraphEdges.clear();
        this.n4lGraphNodes.add(allNodes);
        this.n4lGraphEdges.add(allEdges);

        // Set active filter
        this.n4lActiveContextFilter = 'chaînes causales';

        // Refresh metadata to show active button
        this.showN4LMetadata(this.lastN4LParse);

        // Show restore button
        const restoreBtn = document.getElementById('n4l-restore-graph-btn');
        if (restoreBtn) restoreBtn.style.display = 'flex';

        // Fit graph
        setTimeout(() => {
            if (this.n4lGraph) {
                this.n4lGraph.fit({
                    animation: { duration: 400, easingFunction: 'easeInOutQuad' }
                });
            }
        }, 100);

        this.showToast(`${chains.length} chaînes causales affichées`);
    },

    // Show hypotheses as a graph
    showHypothesesGraph() {
        if (!this.lastN4LParse?.aliases) {
            this.showToast('Aucune hypothèse trouvée', 'warning');
            return;
        }

        // Find hypothesis aliases (starting with "hyp")
        const hypAliases = Object.entries(this.lastN4LParse.aliases)
            .filter(([key]) => key.startsWith('hyp'));

        if (hypAliases.length === 0) {
            this.showToast('Aucune hypothèse trouvée', 'warning');
            return;
        }

        // Check if graph is initialized
        if (!this.n4lGraph || !this.n4lGraphNodes || !this.n4lGraphEdges) {
            if (this.lastN4LParse?.graph) {
                this.renderN4LGraph(this.lastN4LParse.graph);
            } else {
                this.showToast('Graphe non initialisé', 'warning');
                return;
            }
        }

        const allNodes = [];
        const allEdges = [];

        // Color based on confidence level
        const getConfidenceColor = (content) => {
            const match = content.match(/(\d+)%/);
            if (match) {
                const conf = parseInt(match[1]);
                if (conf >= 70) return { bg: '#10b981', border: '#059669' }; // Green - high
                if (conf >= 40) return { bg: '#f59e0b', border: '#d97706' }; // Orange - medium
                return { bg: '#ef4444', border: '#dc2626' }; // Red - low
            }
            return { bg: '#6b7280', border: '#4b5563' }; // Gray - unknown
        };

        // Parse hypothesis content to extract name
        const getHypName = (content) => {
            // Extract text before first attribute (before "(")
            const parenIdx = content.indexOf('(');
            if (parenIdx > 0) {
                return content.substring(0, parenIdx).trim();
            }
            return content.trim();
        };

        // Central node for hypotheses
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
                font: { size: 12, color: '#fff', multi: true },
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

        // Update graph
        this.n4lGraphNodes.clear();
        this.n4lGraphEdges.clear();
        this.n4lGraphNodes.add(allNodes);
        this.n4lGraphEdges.add(allEdges);

        // Set active filter
        this.n4lActiveContextFilter = 'hypothèses, pistes';

        // Refresh metadata to show active button
        this.showN4LMetadata(this.lastN4LParse);

        // Show restore button
        const restoreBtn = document.getElementById('n4l-restore-graph-btn');
        if (restoreBtn) restoreBtn.style.display = 'flex';

        // Fit graph
        setTimeout(() => {
            if (this.n4lGraph) {
                this.n4lGraph.fit({
                    animation: { duration: 400, easingFunction: 'easeInOutQuad' }
                });
            }
        }, 100);

        this.showToast(`${hypAliases.length} hypothèses affichées`);
    },

    // Show TODO/notes as a graph
    showTodoGraph() {
        if (!this.lastN4LParse?.aliases) {
            this.showToast('Aucune note/TODO trouvée', 'warning');
            return;
        }

        // Find TODO aliases (starting with "todo" or "note")
        const todoAliases = Object.entries(this.lastN4LParse.aliases)
            .filter(([key]) => key.toLowerCase().startsWith('todo') || key.toLowerCase().startsWith('note'));

        // If no todo aliases, try to parse from notes section
        let todoItems = [];
        if (todoAliases.length === 0) {
            // Try to get TODO items from notes
            const notesSection = this.lastN4LParse.notes?.['notes, TODO'] || this.lastN4LParse.notes?.['TODO, notes'] || [];
            if (notesSection.length > 0) {
                todoItems = notesSection;
            }
        } else {
            todoItems = todoAliases.map(([key, values]) => values[0] || key);
        }

        if (todoItems.length === 0) {
            this.showToast('Aucune note/TODO trouvée dans cette affaire', 'info');
            return;
        }

        // Check if graph is initialized
        if (!this.n4lGraph || !this.n4lGraphNodes || !this.n4lGraphEdges) {
            if (this.lastN4LParse?.graph) {
                this.renderN4LGraph(this.lastN4LParse.graph);
            } else {
                this.showToast('Graphe non initialisé', 'warning');
                return;
            }
        }

        const allNodes = [];
        const allEdges = [];

        // Central node for TODOs
        allNodes.push({
            id: 'todo_center',
            label: 'À FAIRE',
            color: { background: '#dc2626', border: '#b91c1c' },
            font: { size: 16, color: '#fff', bold: true },
            shape: 'box',
            margin: 12,
            x: 0,
            y: 0
        });

        todoItems.forEach((item, i) => {
            const angle = (2 * Math.PI * i) / todoItems.length;
            const radius = 250;

            // Truncate long items
            const label = item.length > 40 ? item.substring(0, 37) + '...' : item;

            allNodes.push({
                id: `todo_${i}`,
                label: label,
                title: item, // Full text on hover
                color: { background: '#fef3c7', border: '#f59e0b' },
                borderWidth: 2,
                font: { size: 11, color: '#92400e' },
                shape: 'box',
                margin: 8,
                x: Math.cos(angle) * radius,
                y: Math.sin(angle) * radius
            });

            allEdges.push({
                id: `todo_edge_${i}`,
                from: 'todo_center',
                to: `todo_${i}`,
                color: { color: '#f59e0b', opacity: 0.5 },
                width: 1,
                dashes: true
            });
        });

        // Update graph
        this.n4lGraphNodes.clear();
        this.n4lGraphEdges.clear();
        this.n4lGraphNodes.add(allNodes);
        this.n4lGraphEdges.add(allEdges);

        // Set active filter
        this.n4lActiveContextFilter = 'TODO, notes';

        // Refresh metadata to show active button
        this.showN4LMetadata(this.lastN4LParse);

        // Show restore button
        const restoreBtn = document.getElementById('n4l-restore-graph-btn');
        if (restoreBtn) restoreBtn.style.display = 'flex';

        // Fit graph
        setTimeout(() => {
            if (this.n4lGraph) {
                this.n4lGraph.fit({
                    animation: { duration: 400, easingFunction: 'easeInOutQuad' }
                });
            }
        }, 100);

        this.showToast(`${todoItems.length} notes/TODO affichées`);
    },

    // Focus on a cross reference
    focusN4LCrossRef(alias, index) {
        if (!this.lastN4LParse?.cross_refs || !this.n4lGraph) return;

        // Find the cross reference
        const ref = this.lastN4LParse.cross_refs.find(r => r.alias === alias && r.index === index);
        if (!ref) {
            this.showToast(`Référence $${alias}.${index} non trouvée`, 'warning');
            return;
        }

        // Try to find node matching resolved value
        const allNodes = this.n4lGraphNodes?.get() || [];
        const resolved = ref.resolved;

        let node = allNodes.find(n => n.label === resolved);
        if (!node) node = allNodes.find(n => n.label.includes(resolved));
        if (!node) node = allNodes.find(n => n.id === resolved);

        if (node) {
            this.focusN4LGraphNode(node.id);
            this.n4lGraph.focus(node.id, { scale: 1.5, animation: true });
            this.showToast(`$${alias}.${index} → ${node.label}`);
        } else {
            this.showToast(`Référence $${alias}.${index} = ${resolved} (non visible)`, 'info');
        }
    },

    // ============================================
    // Dashboard Graph (reuses N4L functionality)
    // ============================================
    dashboardGraph: null,
    dashboardGraphNodes: null,
    dashboardGraphEdges: null,
    lastDashboardParse: null,
    dashboardActiveContextFilter: null,

    async loadDashboardGraph() {
        console.log('[Dashboard] loadDashboardGraph called, currentCase:', this.currentCase?.id);
        if (!this.currentCase?.id) {
            console.warn('[Dashboard] No currentCase, aborting');
            return;
        }

        try {
            // Parse N4L to get graph data
            const result = await fetch('/api/n4l/parse', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ case_id: this.currentCase.id })
            }).then(r => r.json());

            console.log('[Dashboard] Parse result:', result ? 'received' : 'empty', 'graph nodes:', result?.graph?.nodes?.length);

            // Store as N4L parse result so all N4L functions work
            this.lastN4LParse = result;

            if (result.graph) {
                // Render N4L graph in the dashboard container
                this.renderN4LGraphInContainer(result.graph, 'dashboard-graph-container');
                // Show metadata in dashboard container
                this.showN4LMetadataInContainer(result, 'dashboard-metadata-container');
            } else {
                console.warn('[Dashboard] No graph in result');
            }
        } catch (error) {
            console.error('[Dashboard] Error loading dashboard graph:', error);
        }
    },

    // Render N4L graph in a specific container
    renderN4LGraphInContainer(graphData, containerId) {
        const container = document.getElementById(containerId);
        if (!container) {
            console.error(`[N4L] Container #${containerId} not found!`);
            return;
        }

        const getEdgeColor = (edge) => {
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

        // Store as the main N4L graph so all N4L functions work
        if (this.n4lGraph) {
            this.n4lGraph.destroy();
        }
        this.n4lGraph = new vis.Network(container, { nodes, edges }, options);
        this.n4lGraphNodes = nodes;
        this.n4lGraphEdges = edges;
        this.n4lGraphNodesData = nodes.get();  // Store array copy for lookups
        this.selectedN4LNode = null;

        // Désactiver la physique après stabilisation pour éviter les mouvements erratiques
        this.n4lGraph.once('stabilized', () => {
            this.n4lGraph.setOptions({ physics: { enabled: false } });
        });

        // Add click handler
        this.n4lGraph.on('click', (params) => {
            if (params.nodes.length > 0) {
                this.focusN4LGraphNode(params.nodes[0]);
            } else {
                this.resetN4LGraphFocus();
            }
        });

        // Add right-click context menu handler
        this.n4lGraph.on('oncontext', (params) => {
            params.event.preventDefault();
            this.handleN4LGraphRightClick(params);
        });

        // Setup context menu action handlers
        if (typeof this.setupContextMenuActions === 'function') {
            this.setupContextMenuActions();
        }
    },

    // Show N4L metadata in a specific container
    showN4LMetadataInContainer(result, containerId) {
        const metadataContainer = document.getElementById(containerId);
        if (!metadataContainer) return;

        const nodeCount = result.graph?.nodes?.length || 0;
        const edgeCount = result.graph?.edges?.length || 0;
        const contextCount = result.contexts?.length || 0;
        const aliasCount = result.aliases ? Object.keys(result.aliases).length : 0;
        const sequenceCount = result.sequences?.length || 0;

        // Store current filter state
        this.n4lActiveContextFilter = this.n4lActiveContextFilter || null;

        // Helper function for escaping HTML in attributes
        const escapeHtml = (str) => String(str).replace(/"/g, '&quot;').replace(/'/g, '&#39;');

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

        // Contexts Section with Filters (excluding special views)
        const isSpecialContext = (ctx) => {
            const lower = ctx.toLowerCase();
            return lower.includes('hypothèse') || lower.includes('piste') || lower.includes('chaînes causales') || lower.includes('chaines causales');
        };
        const regularContexts = result.contexts ? result.contexts.filter(ctx => !isSpecialContext(ctx)) : [];
        const availableSpecialContexts = result.contexts ? result.contexts.filter(ctx => isSpecialContext(ctx)) : [];

        if (regularContexts.length > 0) {
            metadataHtml += `
                <div class="n4l-section">
                    <div class="n4l-section-header">
                        <span class="material-icons">filter_list</span>
                        <span>Filtrer par contexte</span>
                    </div>
                    <div class="n4l-context-grid">
                        ${regularContexts.map(ctx => {
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

        // Special Views Section (hypothèses, chaînes causales)
        if (availableSpecialContexts.length > 0) {
            metadataHtml += `
                <div class="n4l-section">
                    <div class="n4l-section-header">
                        <span class="material-icons">analytics</span>
                        <span>Vues spéciales</span>
                    </div>
                    <div class="n4l-context-grid">
                        ${availableSpecialContexts.map(ctx => {
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
                        <div class="n4l-alias-grid" id="n4l-alias-grid">
                            ${aliases.slice(0, 20).map(([key, value]) => {
                                const displayName = getDisplayName(value);
                                return `
                                <div class="n4l-alias-item" data-alias-key="${escapeHtml(key)}" title="@${key}">
                                    <span class="n4l-alias-key">${escapeHtml(displayName)}</span>
                                </div>
                            `;
                            }).join('')}
                            ${aliases.length > 20 ? `
                                <div class="n4l-alias-hidden" id="n4l-alias-hidden" style="display: none;">
                                    ${aliases.slice(20).map(([key, value]) => {
                                        const displayName = getDisplayName(value);
                                        return `
                                        <div class="n4l-alias-item" data-alias-key="${escapeHtml(key)}" title="@${key}">
                                            <span class="n4l-alias-key">${escapeHtml(displayName)}</span>
                                        </div>
                                        `;
                                    }).join('')}
                                </div>
                                <div class="n4l-alias-more" onclick="app.toggleN4LAliasMore(this)" data-expanded="false">
                                    <span class="material-icons">expand_more</span>
                                    <span class="n4l-alias-more-text">+${aliases.length - 20} autres...</span>
                                </div>
                            ` : ''}
                        </div>
                    </div>
                </div>
            `;
        }

        // Causal Chains Section
        if (result.causal_chains && result.causal_chains.length > 0) {
            metadataHtml += `
                <div class="n4l-section n4l-causal-section">
                    <div class="n4l-section-header n4l-collapsible" onclick="app.toggleN4LSection(this)">
                        <span class="material-icons">route</span>
                        <span>Chaînes Causales (${result.causal_chains.length})</span>
                        <span class="material-icons n4l-expand-icon">expand_more</span>
                    </div>
                    <div class="n4l-section-content n4l-collapsed">
                        <button class="n4l-restore-btn" onclick="event.stopPropagation(); app.restoreFullGraph();" style="display:none;" id="n4l-restore-graph-btn">
                            <span class="material-icons">restore</span> Restaurer le graphe complet
                        </button>
                        <div class="n4l-chains-list">
                            ${result.causal_chains.map((chain, i) => `
                                <div class="n4l-chain-item" onclick="app.highlightCausalChain(${i})" title="Cliquer pour visualiser">
                                    <div class="n4l-chain-header">
                                        <span class="n4l-chain-number">${i + 1}</span>
                                        <span class="n4l-chain-id">${chain.id || ''}</span>
                                        <span class="n4l-chain-type">${this.getSTTypeLabel(chain.st_type)}</span>
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

        // Cross References Section
        if (result.cross_refs && result.cross_refs.length > 0) {
            const uniqueRefs = new Map();
            result.cross_refs.forEach(ref => {
                const key = `${ref.alias}.${ref.index}`;
                if (!uniqueRefs.has(key)) {
                    uniqueRefs.set(key, ref);
                }
            });
            metadataHtml += `
                <div class="n4l-section n4l-crossrefs-section">
                    <div class="n4l-section-header n4l-collapsible" onclick="app.toggleN4LSection(this)">
                        <span class="material-icons">link</span>
                        <span>Références Croisées (${uniqueRefs.size})</span>
                        <span class="material-icons n4l-expand-icon">expand_more</span>
                    </div>
                    <div class="n4l-section-content n4l-collapsed">
                        <div class="n4l-crossrefs-grid">
                            ${Array.from(uniqueRefs.values()).slice(0, 15).map(ref => `
                                <div class="n4l-crossref-item" onclick="app.focusN4LCrossRef('${ref.alias}', ${ref.index})" title="$${ref.alias}.${ref.index}">
                                    <span class="n4l-ref-key">$${ref.alias}.${ref.index}</span>
                                    <span class="n4l-ref-value">${ref.resolved || '?'}</span>
                                </div>
                            `).join('')}
                            ${uniqueRefs.size > 15 ? `
                                <div class="n4l-crossrefs-hidden" style="display: none;">
                                    ${Array.from(uniqueRefs.values()).slice(15).map(ref => `
                                        <div class="n4l-crossref-item" onclick="app.focusN4LCrossRef('${ref.alias}', ${ref.index})" title="$${ref.alias}.${ref.index}">
                                            <span class="n4l-ref-key">$${ref.alias}.${ref.index}</span>
                                            <span class="n4l-ref-value">${ref.resolved || '?'}</span>
                                        </div>
                                    `).join('')}
                                </div>
                                <div class="n4l-crossref-more" onclick="app.toggleN4LCrossRefsMore(this)" style="cursor: pointer;">
                                    <span class="material-icons" style="font-size: 14px; vertical-align: middle;">expand_more</span>
                                    <span class="n4l-crossref-more-text">+${uniqueRefs.size - 15} autres...</span>
                                </div>
                            ` : ''}
                        </div>
                    </div>
                </div>
            `;
        }

        // Implicit Markers Section
        if (result.implicit_markers && Object.keys(result.implicit_markers).length > 0) {
            const markers = result.implicit_markers;
            const definitions = Object.entries(markers).filter(([k]) => k.startsWith('definition:')).flatMap(([, v]) => v);
            const importants = Object.entries(markers).filter(([k]) => k.startsWith('important:')).flatMap(([, v]) => v);
            const references = Object.entries(markers).filter(([k]) => k.startsWith('reference:')).flatMap(([, v]) => v);

            if (definitions.length > 0 || importants.length > 0 || references.length > 0) {
                metadataHtml += `
                    <div class="n4l-section n4l-markers-section">
                        <div class="n4l-section-header n4l-collapsible" onclick="app.toggleN4LSection(this)">
                            <span class="material-icons">bookmark</span>
                            <span>Marqueurs N4L</span>
                            <span class="material-icons n4l-expand-icon">expand_more</span>
                        </div>
                        <div class="n4l-section-content n4l-collapsed">
                            ${definitions.length > 0 ? `
                                <div class="n4l-marker-group">
                                    <div class="n4l-marker-label"><span class="n4l-marker-icon">=</span> Définitions</div>
                                    <div class="n4l-marker-tags">${definitions.slice(0, 10).map(d => `<span class="n4l-tag n4l-tag-def n4l-marker-clickable" data-marker="${escapeHtml(d)}">${d}</span>`).join(' ')}</div>
                                </div>
                            ` : ''}
                            ${importants.length > 0 ? `
                                <div class="n4l-marker-group">
                                    <div class="n4l-marker-label"><span class="n4l-marker-icon">*</span> Importants</div>
                                    <div class="n4l-marker-tags">${importants.slice(0, 10).map(i => `<span class="n4l-tag n4l-tag-important n4l-marker-clickable" data-marker="${escapeHtml(i)}">${i}</span>`).join(' ')}</div>
                                </div>
                            ` : ''}
                            ${references.length > 0 ? `
                                <div class="n4l-marker-group">
                                    <div class="n4l-marker-label"><span class="n4l-marker-icon">.</span> Références</div>
                                    <div class="n4l-marker-tags">${references.slice(0, 10).map(r => `<span class="n4l-tag n4l-tag-ref n4l-marker-clickable" data-marker="${escapeHtml(r)}">${r}</span>`).join(' ')}</div>
                                </div>
                            ` : ''}
                        </div>
                    </div>
                `;
            }
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
                    if (aliasKey) {
                        this.focusN4LAlias(aliasKey);
                    }
                }
            };
        }

        // Add click handler for marker tags
        const markersSection = metadataContainer.querySelector('.n4l-markers-section');
        if (markersSection) {
            markersSection.onclick = (e) => {
                const markerTag = e.target.closest('.n4l-marker-clickable');
                if (markerTag) {
                    const markerText = markerTag.getAttribute('data-marker');
                    if (markerText) {
                        this.focusMarkerNode(markerText);
                    }
                }
            };
        }
    },

    renderDashboardGraph(graphData) {
        console.log('[Dashboard] renderDashboardGraph called, nodes:', graphData?.nodes?.length);
        const container = document.getElementById('dashboard-graph-container');
        if (!container) {
            console.error('[Dashboard] Container #dashboard-graph-container not found!');
            return;
        }
        console.log('[Dashboard] Container found, dimensions:', container.offsetWidth, 'x', container.offsetHeight);

        // Reuse the same node/edge formatting as N4L
        const getEdgeColor = (edge) => {
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
            id: `dashboard-edge-${i}`,
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
                solver: 'forceAtlas2Based',
                forceAtlas2Based: {
                    gravitationalConstant: -50,
                    centralGravity: 0.01,
                    springLength: 100,
                    springConstant: 0.08
                },
                stabilization: { iterations: 150 }
            },
            interaction: {
                hover: true,
                tooltipDelay: 200,
                navigationButtons: true
            }
        };

        this.dashboardGraphNodes = nodes;
        this.dashboardGraphEdges = edges;

        if (this.dashboardGraph) {
            this.dashboardGraph.destroy();
        }

        this.dashboardGraph = new vis.Network(container, { nodes, edges }, options);
    },

    showDashboardMetadata(result) {
        const container = document.getElementById('dashboard-metadata-container');
        if (!container) return;

        const nodeCount = result.graph?.nodes?.length || 0;
        const edgeCount = result.graph?.edges?.length || 0;
        const contextCount = result.contexts?.length || 0;
        const aliasCount = result.aliases ? Object.keys(result.aliases).length : 0;
        const sequenceCount = result.sequences?.length || 0;

        // Helper function for escaping HTML in attributes
        const escapeHtml = (str) => String(str).replace(/"/g, '&quot;').replace(/'/g, '&#39;');

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

        // Contexts Section with Filters (excluding special views)
        const isSpecialContext = (ctx) => {
            const lower = ctx.toLowerCase();
            return lower.includes('hypothèse') || lower.includes('piste') || lower.includes('chaînes causales') || lower.includes('chaines causales');
        };
        const regularContexts = result.contexts ? result.contexts.filter(ctx => !isSpecialContext(ctx)) : [];
        const availableSpecialContexts = result.contexts ? result.contexts.filter(ctx => isSpecialContext(ctx)) : [];

        if (regularContexts.length > 0) {
            metadataHtml += `
                <div class="n4l-section">
                    <div class="n4l-section-header">
                        <span class="material-icons">filter_list</span>
                        <span>Filtrer par contexte</span>
                    </div>
                    <div class="n4l-context-grid">
                        ${regularContexts.map(ctx => {
                            const isActive = this.dashboardActiveContextFilter === ctx;
                            const icon = this.getContextIcon(ctx);
                            return `
                                <button class="n4l-context-btn ${isActive ? 'active' : ''}" onclick="app.filterDashboardByContext('${ctx}')">
                                    <span class="material-icons">${icon}</span>
                                    <span>${ctx}</span>
                                </button>
                            `;
                        }).join('')}
                    </div>
                </div>
            `;
        }

        // Special Views Section (hypothèses, chaînes causales)
        if (availableSpecialContexts.length > 0) {
            metadataHtml += `
                <div class="n4l-section">
                    <div class="n4l-section-header">
                        <span class="material-icons">analytics</span>
                        <span>Vues spéciales</span>
                    </div>
                    <div class="n4l-context-grid">
                        ${availableSpecialContexts.map(ctx => {
                            const isActive = this.dashboardActiveContextFilter === ctx;
                            const icon = this.getContextIcon(ctx);
                            return `
                                <button class="n4l-context-btn ${isActive ? 'active' : ''}" onclick="app.filterDashboardByContext('${ctx}')">
                                    <span class="material-icons">${icon}</span>
                                    <span>${ctx}</span>
                                </button>
                            `;
                        }).join('')}
                    </div>
                </div>
            `;
        }

        // Sequences Section (collapsible) - same as N4L with navigation
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
                        <div class="n4l-alias-grid" id="dashboard-alias-grid">
                            ${aliases.slice(0, 20).map(([key, value]) => {
                                const displayName = getDisplayName(value);
                                return `
                                <div class="n4l-alias-item" data-alias-key="${escapeHtml(key)}" title="@${key}" onclick="app.focusN4LAlias('${escapeHtml(key)}')">
                                    <span class="n4l-alias-key">${escapeHtml(displayName)}</span>
                                </div>
                            `;
                            }).join('')}
                            ${aliases.length > 20 ? `
                                <div class="n4l-alias-hidden" id="dashboard-alias-hidden" style="display: none;">
                                    ${aliases.slice(20).map(([key, value]) => {
                                        const displayName = getDisplayName(value);
                                        return `
                                        <div class="n4l-alias-item" data-alias-key="${escapeHtml(key)}" title="@${key}" onclick="app.focusN4LAlias('${escapeHtml(key)}')">
                                            <span class="n4l-alias-key">${escapeHtml(displayName)}</span>
                                        </div>
                                        `;
                                    }).join('')}
                                </div>
                                <div class="n4l-alias-more" onclick="app.toggleN4LAliasMore(this)" data-expanded="false">
                                    <span class="material-icons">expand_more</span>
                                    <span class="n4l-alias-more-text">+${aliases.length - 20} autres...</span>
                                </div>
                            ` : ''}
                        </div>
                    </div>
                </div>
            `;
        }

        // Causal Chains Section
        if (result.causal_chains && result.causal_chains.length > 0) {
            metadataHtml += `
                <div class="n4l-section n4l-causal-section">
                    <div class="n4l-section-header n4l-collapsible" onclick="app.toggleN4LSection(this)">
                        <span class="material-icons">route</span>
                        <span>Chaînes Causales (${result.causal_chains.length})</span>
                        <span class="material-icons n4l-expand-icon">expand_more</span>
                    </div>
                    <div class="n4l-section-content n4l-collapsed">
                        <button class="n4l-restore-btn" onclick="event.stopPropagation(); app.restoreDashboardFullGraph();" style="display:none;" id="dashboard-restore-graph-btn">
                            <span class="material-icons">restore</span> Restaurer le graphe complet
                        </button>
                        <div class="n4l-chains-list">
                            ${result.causal_chains.map((chain, i) => `
                                <div class="n4l-chain-item" onclick="app.highlightDashboardCausalChain(${i})" title="Cliquer pour visualiser">
                                    <div class="n4l-chain-header">
                                        <span class="n4l-chain-number">${i + 1}</span>
                                        <span class="n4l-chain-id">${chain.id || ''}</span>
                                        <span class="n4l-chain-type">${this.getSTTypeLabel ? this.getSTTypeLabel(chain.st_type) : ''}</span>
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

        // Cross References Section
        if (result.cross_refs && result.cross_refs.length > 0) {
            const uniqueRefs = new Map();
            result.cross_refs.forEach(ref => {
                const key = `${ref.alias}.${ref.index}`;
                if (!uniqueRefs.has(key)) {
                    uniqueRefs.set(key, ref);
                }
            });
            metadataHtml += `
                <div class="n4l-section n4l-crossrefs-section">
                    <div class="n4l-section-header n4l-collapsible" onclick="app.toggleN4LSection(this)">
                        <span class="material-icons">link</span>
                        <span>Références Croisées (${uniqueRefs.size})</span>
                        <span class="material-icons n4l-expand-icon">expand_more</span>
                    </div>
                    <div class="n4l-section-content n4l-collapsed">
                        <div class="n4l-crossrefs-grid">
                            ${Array.from(uniqueRefs.values()).slice(0, 15).map(ref => `
                                <div class="n4l-crossref-item" onclick="app.focusDashboardCrossRef('${ref.alias}', ${ref.index})" title="$${ref.alias}.${ref.index}">
                                    <span class="n4l-ref-key">$${ref.alias}.${ref.index}</span>
                                    <span class="n4l-ref-value">${ref.resolved || '?'}</span>
                                </div>
                            `).join('')}
                            ${uniqueRefs.size > 15 ? `
                                <div class="n4l-crossrefs-hidden" style="display: none;">
                                    ${Array.from(uniqueRefs.values()).slice(15).map(ref => `
                                        <div class="n4l-crossref-item" onclick="app.focusDashboardCrossRef('${ref.alias}', ${ref.index})" title="$${ref.alias}.${ref.index}">
                                            <span class="n4l-ref-key">$${ref.alias}.${ref.index}</span>
                                            <span class="n4l-ref-value">${ref.resolved || '?'}</span>
                                        </div>
                                    `).join('')}
                                </div>
                                <div class="n4l-crossref-more" onclick="app.toggleN4LCrossRefsMore(this)" style="cursor: pointer;">
                                    <span class="material-icons" style="font-size: 14px; vertical-align: middle;">expand_more</span>
                                    <span class="n4l-crossref-more-text">+${uniqueRefs.size - 15} autres...</span>
                                </div>
                            ` : ''}
                        </div>
                    </div>
                </div>
            `;
        }

        // Implicit Markers Section
        if (result.implicit_markers && Object.keys(result.implicit_markers).length > 0) {
            const markers = result.implicit_markers;
            const definitions = Object.entries(markers).filter(([k]) => k.startsWith('definition:')).flatMap(([, v]) => v);
            const importants = Object.entries(markers).filter(([k]) => k.startsWith('important:')).flatMap(([, v]) => v);
            const references = Object.entries(markers).filter(([k]) => k.startsWith('reference:')).flatMap(([, v]) => v);

            if (definitions.length > 0 || importants.length > 0 || references.length > 0) {
                metadataHtml += `
                    <div class="n4l-section n4l-markers-section">
                        <div class="n4l-section-header n4l-collapsible" onclick="app.toggleN4LSection(this)">
                            <span class="material-icons">bookmark</span>
                            <span>Marqueurs N4L</span>
                            <span class="material-icons n4l-expand-icon">expand_more</span>
                        </div>
                        <div class="n4l-section-content n4l-collapsed">
                            ${definitions.length > 0 ? `
                                <div class="n4l-marker-group">
                                    <div class="n4l-marker-label"><span class="n4l-marker-icon">=</span> Définitions</div>
                                    <div class="n4l-marker-tags">${definitions.slice(0, 10).map(d => `<span class="n4l-tag n4l-tag-def n4l-marker-clickable" data-marker="${escapeHtml(d)}">${d}</span>`).join(' ')}</div>
                                </div>
                            ` : ''}
                            ${importants.length > 0 ? `
                                <div class="n4l-marker-group">
                                    <div class="n4l-marker-label"><span class="n4l-marker-icon">*</span> Importants</div>
                                    <div class="n4l-marker-tags">${importants.slice(0, 10).map(i => `<span class="n4l-tag n4l-tag-important n4l-marker-clickable" data-marker="${escapeHtml(i)}">${i}</span>`).join(' ')}</div>
                                </div>
                            ` : ''}
                            ${references.length > 0 ? `
                                <div class="n4l-marker-group">
                                    <div class="n4l-marker-label"><span class="n4l-marker-icon">.</span> Références</div>
                                    <div class="n4l-marker-tags">${references.slice(0, 10).map(r => `<span class="n4l-tag n4l-tag-ref n4l-marker-clickable" data-marker="${escapeHtml(r)}">${r}</span>`).join(' ')}</div>
                                </div>
                            ` : ''}
                        </div>
                    </div>
                `;
            }
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

        container.innerHTML = metadataHtml;

        // Add click handler for marker tags
        const markersSection = container.querySelector('.n4l-markers-section');
        if (markersSection) {
            markersSection.onclick = (e) => {
                const markerTag = e.target.closest('.n4l-marker-clickable');
                if (markerTag) {
                    const markerText = markerTag.getAttribute('data-marker');
                    if (markerText) {
                        this.focusMarkerNode(markerText);
                    }
                }
            };
        }
    },

    // Dashboard filter by context (same logic as N4L)
    filterDashboardByContext(context) {
        if (!this.lastDashboardParse || !this.dashboardGraph) return;

        if (this.dashboardActiveContextFilter === context) {
            this.resetDashboardFilter();
            return;
        }

        // Special cases
        if (context === 'chaînes causales') {
            this.showAllDashboardCausalChains();
            return;
        }
        if (context.includes('hypothèse') || context.includes('piste')) {
            this.showDashboardHypothesesGraph();
            return;
        }
        if (context.toLowerCase().includes('todo') || context.toLowerCase().includes('note')) {
            this.showToast('Aucune note/TODO trouvée dans cette affaire', 'info');
            return;
        }

        this.dashboardActiveContextFilter = context;

        const result = this.lastDashboardParse;
        const involvedNodes = new Set();

        result.graph.nodes.forEach(n => {
            if (n.context && (n.context === context || n.context.includes(context) || context.includes(n.context))) {
                involvedNodes.add(n.id);
            }
        });

        result.graph.edges.forEach(e => {
            if (e.context && (e.context === context || e.context.includes(context) || context.includes(e.context))) {
                involvedNodes.add(e.from);
                involvedNodes.add(e.to);
            }
        });

        if (involvedNodes.size === 0) {
            this.showToast(`Aucune entité trouvée pour "${context}"`, 'warning');
            this.dashboardActiveContextFilter = null;
            return;
        }

        const allNodes = this.dashboardGraphNodes.getIds();
        const visibleNodeIds = Array.from(involvedNodes);

        const nodeUpdates = allNodes.map(id => {
            const isVisible = involvedNodes.has(id);
            return {
                id,
                hidden: false,
                color: isVisible ? this.getDashboardOriginalNodeColor(id) : { background: 'rgba(200,200,200,0.1)', border: 'rgba(200,200,200,0.1)' },
                font: { color: isVisible ? '#1a1a2e' : 'rgba(0,0,0,0.05)' },
                opacity: isVisible ? 1 : 0.05
            };
        });
        this.dashboardGraphNodes.update(nodeUpdates);

        const allEdges = this.dashboardGraphEdges.getIds();
        const edgeUpdates = allEdges.map(edgeId => {
            const edge = this.dashboardGraphEdges.get(edgeId);
            const isVisible = involvedNodes.has(edge.from) && involvedNodes.has(edge.to);
            return {
                id: edgeId,
                hidden: !isVisible,
                color: isVisible ? undefined : { color: 'rgba(200,200,200,0.05)' }
            };
        });
        this.dashboardGraphEdges.update(edgeUpdates);

        this.showDashboardMetadata(this.lastDashboardParse);

        setTimeout(() => {
            if (this.dashboardGraph && visibleNodeIds.length > 0) {
                this.dashboardGraph.fit({
                    nodes: visibleNodeIds,
                    animation: { duration: 400, easingFunction: 'easeInOutQuad' }
                });
            }
        }, 100);

        this.showToast(`Filtré: ${context} (${involvedNodes.size} entités)`);
    },

    getDashboardOriginalNodeColor(nodeId) {
        const node = this.dashboardGraphNodes.get(nodeId);
        return node?.originalColor || { background: '#e2e8f0', border: '#cbd5e0' };
    },

    resetDashboardFilter() {
        if (!this.dashboardGraph) return;

        const specialFilters = ['chaînes causales', 'hypothèses, pistes', 'pistes, hypothèses'];
        if (this.currentDashboardCausalChain || specialFilters.includes(this.dashboardActiveContextFilter)) {
            this.restoreDashboardFullGraph();
            this.dashboardActiveContextFilter = null;
            if (this.lastDashboardParse) {
                this.showDashboardMetadata(this.lastDashboardParse);
            }
            return;
        }

        this.dashboardActiveContextFilter = null;

        const allNodes = this.dashboardGraphNodes.getIds();
        const nodeUpdates = allNodes.map(id => {
            const node = this.dashboardGraphNodes.get(id);
            return {
                id,
                hidden: false,
                color: node.originalColor || undefined,
                font: { color: '#1a1a2e' },
                opacity: 1
            };
        });
        this.dashboardGraphNodes.update(nodeUpdates);

        const allEdges = this.dashboardGraphEdges.getIds();
        const edgeUpdates = allEdges.map(id => ({
            id,
            hidden: false,
            color: undefined
        }));
        this.dashboardGraphEdges.update(edgeUpdates);

        if (this.lastDashboardParse) {
            this.showDashboardMetadata(this.lastDashboardParse);
        }

        setTimeout(() => {
            if (this.dashboardGraph) {
                this.dashboardGraph.fit({ animation: { duration: 300, easingFunction: 'easeInOutQuad' } });
            }
        }, 100);

        this.showToast('Filtre réinitialisé');
    },

    highlightDashboardCausalChain(chainIndex) {
        if (!this.lastDashboardParse?.causal_chains) return;

        const chain = this.lastDashboardParse.causal_chains[chainIndex];
        if (!chain || !chain.steps || chain.steps.length === 0) return;

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

        const chainEdges = [];
        for (let i = 0; i < chain.steps.length - 1; i++) {
            chainEdges.push({
                id: `chain_edge_${i}`,
                from: `chain_step_${i}`,
                to: `chain_step_${i + 1}`,
                label: chain.steps[i].relation || '',
                arrows: 'to',
                color: { color: '#f59e0b', opacity: 1 },
                width: 3,
                font: { size: 12, color: '#666', strokeWidth: 2, strokeColor: '#fff' },
                smooth: { type: 'curvedCW', roundness: 0.1 }
            });
        }

        this.dashboardGraphNodes.clear();
        this.dashboardGraphEdges.clear();
        this.dashboardGraphNodes.add(chainNodes);
        this.dashboardGraphEdges.add(chainEdges);

        this.currentDashboardCausalChain = { index: chainIndex, nodeIds: chainNodes.map(n => n.id), step: 0 };

        const restoreBtn = document.getElementById('dashboard-restore-graph-btn');
        if (restoreBtn) restoreBtn.style.display = 'flex';

        setTimeout(() => {
            this.dashboardGraph.fit({ animation: { duration: 400, easingFunction: 'easeInOutQuad' } });
        }, 100);

        this.showToast(`Chaîne causale: ${chain.id || chainIndex + 1} (${chain.steps.length} étapes)`);
    },

    restoreDashboardFullGraph() {
        if (!this.lastDashboardParse?.graph) return;

        this.currentDashboardCausalChain = null;

        const restoreBtn = document.getElementById('dashboard-restore-graph-btn');
        if (restoreBtn) restoreBtn.style.display = 'none';

        this.renderDashboardGraph(this.lastDashboardParse.graph);

        this.showToast('Graphe complet restauré');
    },

    showAllDashboardCausalChains() {
        if (!this.lastDashboardParse?.causal_chains || this.lastDashboardParse.causal_chains.length === 0) {
            this.showToast('Aucune chaîne causale trouvée', 'warning');
            return;
        }

        const chains = this.lastDashboardParse.causal_chains;
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
                    y: yOffset,
                    fixed: { x: false, y: false }
                });

                if (stepIndex < chain.steps.length - 1) {
                    allEdges.push({
                        id: `chain_${chainIndex}_edge_${stepIndex}`,
                        from: `chain_${chainIndex}_step_${stepIndex}`,
                        to: `chain_${chainIndex}_step_${stepIndex + 1}`,
                        label: step.relation || '',
                        arrows: 'to',
                        color: { color: color.bg, opacity: 1 },
                        width: 3,
                        font: { size: 11, color: '#666', strokeWidth: 2, strokeColor: '#fff' },
                        smooth: { type: 'curvedCW', roundness: 0.1 }
                    });
                }
            });

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

        this.dashboardGraphNodes.clear();
        this.dashboardGraphEdges.clear();
        this.dashboardGraphNodes.add(allNodes);
        this.dashboardGraphEdges.add(allEdges);

        this.dashboardActiveContextFilter = 'chaînes causales';
        this.showDashboardMetadata(this.lastDashboardParse);

        const restoreBtn = document.getElementById('dashboard-restore-graph-btn');
        if (restoreBtn) restoreBtn.style.display = 'flex';

        setTimeout(() => {
            if (this.dashboardGraph) {
                this.dashboardGraph.fit({ animation: { duration: 400, easingFunction: 'easeInOutQuad' } });
            }
        }, 100);

        this.showToast(`${chains.length} chaînes causales affichées`);
    },

    showDashboardHypothesesGraph() {
        if (!this.lastDashboardParse?.aliases) {
            this.showToast('Aucune hypothèse trouvée', 'warning');
            return;
        }

        const hypAliases = Object.entries(this.lastDashboardParse.aliases)
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
                font: { size: 12, color: '#fff', multi: true },
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

        this.dashboardGraphNodes.clear();
        this.dashboardGraphEdges.clear();
        this.dashboardGraphNodes.add(allNodes);
        this.dashboardGraphEdges.add(allEdges);

        this.dashboardActiveContextFilter = 'hypothèses, pistes';
        this.showDashboardMetadata(this.lastDashboardParse);

        const restoreBtn = document.getElementById('dashboard-restore-graph-btn');
        if (restoreBtn) restoreBtn.style.display = 'flex';

        setTimeout(() => {
            if (this.dashboardGraph) {
                this.dashboardGraph.fit({ animation: { duration: 400, easingFunction: 'easeInOutQuad' } });
            }
        }, 100);

        this.showToast(`${hypAliases.length} hypothèses affichées`);
    },

    highlightDashboardSequence(seqIndex) {
        if (!this.lastDashboardParse?.sequences) return;

        const seq = this.lastDashboardParse.sequences[seqIndex];
        if (!seq || seq.length === 0) return;

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

        this.dashboardGraphNodes.clear();
        this.dashboardGraphEdges.clear();
        this.dashboardGraphNodes.add(seqNodes);
        this.dashboardGraphEdges.add(seqEdges);

        this.currentDashboardCausalChain = { index: seqIndex };

        const restoreBtn = document.getElementById('dashboard-restore-graph-btn');
        if (restoreBtn) restoreBtn.style.display = 'flex';

        setTimeout(() => {
            if (this.dashboardGraph) this.dashboardGraph.fit({ animation: true });
        }, 100);

        this.showToast(`Séquence ${seqIndex + 1}: ${seq.length} étapes`);
    },

    focusDashboardCrossRef(alias, index) {
        if (!this.lastDashboardParse?.cross_refs || !this.dashboardGraph) return;

        const ref = this.lastDashboardParse.cross_refs.find(r => r.alias === alias && r.index === index);
        if (!ref) {
            this.showToast(`Référence $${alias}.${index} non trouvée`, 'warning');
            return;
        }

        const resolved = ref.resolved;
        const allNodes = this.dashboardGraphNodes.get();
        let node = allNodes.find(n => n.label === resolved);
        if (!node) node = allNodes.find(n => n.label.includes(resolved));
        if (!node) node = allNodes.find(n => n.id === resolved);

        if (node) {
            this.dashboardGraph.focus(node.id, { scale: 1.5, animation: true });
            this.showToast(`$${alias}.${index} → ${node.label}`);
        } else {
            this.showToast(`Référence $${alias}.${index} = ${resolved} (non visible)`, 'info');
        }
    },

    toggleDashboardGraphFullscreen() {
        const panel = document.getElementById('panel-dashboard-graph');
        const btn = document.getElementById('btn-fullscreen-dashboard-graph');
        const icon = btn?.querySelector('.material-icons');

        if (!panel) return;

        if (panel.classList.contains('fullscreen-panel')) {
            panel.classList.remove('fullscreen-panel');
            document.body.classList.remove('has-fullscreen-panel');
            if (icon) icon.textContent = 'fullscreen';
            if (btn) btn.setAttribute('data-tooltip', 'Plein écran');
        } else {
            panel.classList.add('fullscreen-panel');
            document.body.classList.add('has-fullscreen-panel');
            if (icon) icon.textContent = 'fullscreen_exit';
            if (btn) btn.setAttribute('data-tooltip', 'Quitter le plein écran');
        }

        setTimeout(() => {
            if (this.dashboardGraph) {
                this.dashboardGraph.redraw();
                this.dashboardGraph.fit({ animation: { duration: 300 } });
            }
        }, 100);
    },

    resetDashboardGraphFocus() {
        // Reset the N4L graph and re-render in Dashboard container
        if (!this.n4lGraph || !this.lastN4LParse) return;

        // Reset filter state
        this.n4lActiveContextFilter = null;
        this.currentCausalChain = null;

        // Re-render the full graph in dashboard container
        if (this.lastN4LParse.graph) {
            this.renderN4LGraphInContainer(this.lastN4LParse.graph, 'dashboard-graph-container');
            this.showN4LMetadataInContainer(this.lastN4LParse, 'dashboard-metadata-container');
        }

        this.showToast('Graphe réinitialisé');
    }
};

// Export for use in main app
if (typeof module !== 'undefined' && module.exports) {
    module.exports = N4LModule;
}
