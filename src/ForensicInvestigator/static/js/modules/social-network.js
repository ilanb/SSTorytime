// ForensicInvestigator - Module Analyse de Réseaux Sociaux
// Détection de communautés, brokers, flux et évolution temporelle

const SocialNetworkModule = {
    // Configuration
    socialNetworkConfig: {
        minCommunitySize: 2,
        maxCommunities: 10,
        brokerThreshold: 0.3,
        flowTypes: ['information', 'argent', 'influence', 'preuve'],
        temporalGranularity: 'day' // day, week, month
    },

    // État du module
    socialNetworkGraph: null,
    communities: [],
    brokers: [],
    flowAnalysis: null,
    temporalSnapshots: [],
    selectedCommunity: null,

    // ============================================
    // Initialisation du module
    // ============================================
    initSocialNetwork() {
        this.setupSocialNetworkListeners();
    },

    setupSocialNetworkListeners() {
        // Bouton d'analyse
        document.getElementById('btn-social-network')?.addEventListener('click', () => {
            this.showSocialNetworkPanel();
        });

        // Fermer le panneau
        document.getElementById('close-social-network-panel')?.addEventListener('click', () => {
            this.hideSocialNetworkPanel();
        });

        // Actions du panneau
        document.getElementById('btn-detect-communities')?.addEventListener('click', () => {
            this.detectCommunities();
        });

        document.getElementById('btn-identify-brokers')?.addEventListener('click', () => {
            this.identifyBrokers();
        });

        document.getElementById('btn-analyze-flows')?.addEventListener('click', () => {
            this.analyzeFlows();
        });

        document.getElementById('btn-temporal-evolution')?.addEventListener('click', () => {
            this.showTemporalEvolution();
        });

        // Sélecteur de type de flux
        document.getElementById('flow-type-select')?.addEventListener('change', (e) => {
            this.updateFlowVisualization(e.target.value);
        });

        // Slider de temps pour l'évolution
        document.getElementById('temporal-slider')?.addEventListener('input', (e) => {
            this.showTemporalSnapshot(parseInt(e.target.value));
        });

        // Animation temporelle
        document.getElementById('btn-play-temporal')?.addEventListener('click', () => {
            this.playTemporalAnimation();
        });
    },

    // ============================================
    // Affichage du panneau
    // ============================================
    showSocialNetworkPanel() {
        if (!this.currentCase) {
            this.showToast('Sélectionnez une affaire d\'abord', 'warning');
            return;
        }

        const panel = document.getElementById('social-network-panel');
        if (panel) {
            panel.classList.add('active');
            this.renderSocialNetworkOverview();
        }
    },

    hideSocialNetworkPanel() {
        const panel = document.getElementById('social-network-panel');
        if (panel) {
            panel.classList.remove('active');
            panel.classList.remove('fullscreen-panel');
            document.body.classList.remove('has-fullscreen-panel');
        }
    },

    toggleSocialNetworkFullscreen() {
        const panel = document.getElementById('social-network-panel');
        const btn = document.getElementById('btn-fullscreen-sn');

        if (!panel) return;

        if (panel.classList.contains('fullscreen-panel')) {
            // Exit fullscreen
            panel.classList.remove('fullscreen-panel');
            document.body.classList.remove('has-fullscreen-panel');
            if (btn) {
                btn.innerHTML = '<span class="material-icons">fullscreen</span>';
            }
        } else {
            // Enter fullscreen
            panel.classList.add('fullscreen-panel');
            document.body.classList.add('has-fullscreen-panel');
            if (btn) {
                btn.innerHTML = '<span class="material-icons">fullscreen_exit</span>';
            }
        }

        // Redraw graph to fit new size
        setTimeout(() => {
            if (this.socialNetworkGraph) {
                this.socialNetworkGraph.fit({ animation: true });
            }
        }, 100);
    },

    // ============================================
    // Vue d'ensemble
    // ============================================
    async renderSocialNetworkOverview() {
        const container = document.getElementById('social-network-content');
        if (!container || !this.currentCase) return;

        container.innerHTML = `
            <div class="social-network-loading">
                <div class="spinner"></div>
                <p>Analyse du réseau social en cours...</p>
            </div>
        `;

        // Construire le graphe
        const graphData = await this.getGraphData();
        if (!graphData.nodes || graphData.nodes.length === 0) {
            container.innerHTML = `
                <div class="empty-state">
                    <span class="material-icons empty-state-icon">hub</span>
                    <p class="empty-state-title">Aucune donnée</p>
                    <p class="empty-state-description">Ajoutez des entités et relations pour analyser le réseau social</p>
                </div>
            `;
            return;
        }

        // Calculer les métriques de base
        const metrics = this.calculateNetworkMetrics(graphData);

        container.innerHTML = `
            <div class="sn-overview">
                <div class="sn-metrics-grid">
                    <div class="sn-metric-card">
                        <span class="material-icons">people</span>
                        <div class="sn-metric-value">${metrics.nodeCount}</div>
                        <div class="sn-metric-label">Noeuds</div>
                    </div>
                    <div class="sn-metric-card">
                        <span class="material-icons">link</span>
                        <div class="sn-metric-value">${metrics.edgeCount}</div>
                        <div class="sn-metric-label">Relations</div>
                    </div>
                    <div class="sn-metric-card">
                        <span class="material-icons">hub</span>
                        <div class="sn-metric-value">${metrics.density.toFixed(2)}</div>
                        <div class="sn-metric-label">Densité</div>
                    </div>
                    <div class="sn-metric-card">
                        <span class="material-icons">account_tree</span>
                        <div class="sn-metric-value">${metrics.avgDegree.toFixed(1)}</div>
                        <div class="sn-metric-label">Degré moyen</div>
                    </div>
                </div>

                <div class="sn-actions-grid">
                    <button class="sn-action-btn" id="btn-detect-communities">
                        <span class="material-icons">groups</span>
                        <span>Détecter les communautés</span>
                    </button>
                    <button class="sn-action-btn" id="btn-identify-brokers">
                        <span class="material-icons">mediation</span>
                        <span>Identifier les brokers</span>
                    </button>
                    <button class="sn-action-btn" id="btn-analyze-flows">
                        <span class="material-icons">swap_horiz</span>
                        <span>Analyser les flux</span>
                    </button>
                    <button class="sn-action-btn" id="btn-temporal-evolution">
                        <span class="material-icons">auto_graph</span>
                        <span>Évolution temporelle</span>
                    </button>
                </div>

                <div class="sn-two-columns">
                    <div class="sn-column-left">
                        <div class="sn-graph-container" id="sn-graph-container"></div>
                    </div>
                    <div class="sn-column-right">
                        <div class="sn-results-container" id="sn-results-container">
                            <div class="sn-results-placeholder">
                                <span class="material-icons">analytics</span>
                                <p>Cliquez sur une analyse ci-dessus pour afficher les résultats</p>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        `;

        // Réattacher les event listeners
        this.setupSocialNetworkListeners();

        // Rendre le graphe initial
        this.renderSocialNetworkGraph(graphData);

        // Détecter automatiquement les communautés
        setTimeout(() => this.detectCommunities(), 300);
    },

    // ============================================
    // Métriques de réseau
    // ============================================
    calculateNetworkMetrics(graphData) {
        const nodeCount = graphData.nodes.length;
        const edgeCount = graphData.edges.length;
        const maxPossibleEdges = (nodeCount * (nodeCount - 1)) / 2;
        const density = maxPossibleEdges > 0 ? edgeCount / maxPossibleEdges : 0;

        // Calculer le degré de chaque noeud
        const degrees = {};
        graphData.nodes.forEach(n => { degrees[n.id] = 0; });
        graphData.edges.forEach(e => {
            degrees[e.from] = (degrees[e.from] || 0) + 1;
            degrees[e.to] = (degrees[e.to] || 0) + 1;
        });

        const totalDegree = Object.values(degrees).reduce((a, b) => a + b, 0);
        const avgDegree = nodeCount > 0 ? totalDegree / nodeCount : 0;

        return { nodeCount, edgeCount, density, avgDegree, degrees };
    },

    // ============================================
    // Rendu du graphe réseau social
    // ============================================
    renderSocialNetworkGraph(graphData) {
        const container = document.getElementById('sn-graph-container');
        if (!container) return;

        const nodes = new vis.DataSet(graphData.nodes.map(n => ({
            id: n.id,
            label: n.label,
            color: this.getNodeColor(n),
            shape: this.getNodeShape(n.type),
            title: `${n.label} (${n.type || 'entité'})`
        })));

        const edges = new vis.DataSet(graphData.edges.map((e, i) => ({
            id: `edge-${i}`,
            from: e.from,
            to: e.to,
            label: e.label || '',
            arrows: 'to',
            color: { color: '#1e3a5f', opacity: 0.6 }
        })));

        const options = {
            nodes: {
                font: { color: '#1a1a2e', size: 11 },
                borderWidth: 2
            },
            edges: {
                font: { size: 9, color: '#4a5568' },
                smooth: { type: 'curvedCW', roundness: 0.2 }
            },
            physics: {
                stabilization: { iterations: 150 },
                barnesHut: { gravitationalConstant: -3000, springLength: 150 }
            },
            interaction: { hover: true }
        };

        this.socialNetworkGraph = new vis.Network(container, { nodes, edges }, options);
        this.snNodes = nodes;
        this.snEdges = edges;
    },

    // ============================================
    // Détection de communautés (Louvain simplifié)
    // ============================================
    async detectCommunities() {
        const graphData = await this.getGraphData();
        if (!graphData.nodes || graphData.nodes.length < 2) {
            this.showToast('Pas assez de données pour détecter des communautés', 'warning');
            return;
        }

        this.showToast('Détection des communautés...', 'info');

        // Construire la matrice d'adjacence
        const adjacency = this.buildAdjacencyList(graphData);

        // Algorithme de détection de communautés (Label Propagation simplifié)
        const communities = this.labelPropagation(graphData.nodes, adjacency);

        this.communities = communities;

        // Afficher les résultats
        this.displayCommunities(communities, graphData);

        // Colorer le graphe par communauté
        this.colorGraphByCommunity(communities);
    },

    buildAdjacencyList(graphData) {
        const adj = {};
        graphData.nodes.forEach(n => { adj[n.id] = []; });
        graphData.edges.forEach(e => {
            if (!adj[e.from]) adj[e.from] = [];
            if (!adj[e.to]) adj[e.to] = [];
            adj[e.from].push(e.to);
            adj[e.to].push(e.from);
        });
        return adj;
    },

    labelPropagation(nodes, adjacency) {
        // Algorithme de détection de communautés par suppression d'arêtes de forte intermédiarité
        // (Girvan-Newman simplifié) - Déterministe et stable

        const nodeIds = nodes.map(n => n.id).sort();
        const n = nodeIds.length;

        if (n === 0) return [];
        if (n <= 2) return [nodeIds];

        // Créer une copie modifiable de l'adjacence
        const edges = [];
        const edgeSet = new Set();

        nodeIds.forEach(id => {
            (adjacency[id] || []).forEach(nId => {
                const edgeKey = [id, nId].sort().join('|');
                if (!edgeSet.has(edgeKey)) {
                    edgeSet.add(edgeKey);
                    edges.push({ from: id, to: nId, key: edgeKey });
                }
            });
        });

        if (edges.length === 0) {
            return nodeIds.map(id => [id]);
        }

        // Fonction pour trouver les composantes connexes
        const findComponents = (nodeList, edgeList) => {
            const adj = {};
            nodeList.forEach(id => { adj[id] = []; });
            edgeList.forEach(e => {
                if (adj[e.from]) adj[e.from].push(e.to);
                if (adj[e.to]) adj[e.to].push(e.from);
            });

            const visited = new Set();
            const components = [];

            for (const startNode of nodeList) {
                if (visited.has(startNode)) continue;

                const component = [];
                const queue = [startNode];

                while (queue.length > 0) {
                    const node = queue.shift();
                    if (visited.has(node)) continue;

                    visited.add(node);
                    component.push(node);

                    (adj[node] || []).forEach(neighbor => {
                        if (!visited.has(neighbor)) {
                            queue.push(neighbor);
                        }
                    });
                }

                if (component.length > 0) {
                    components.push(component.sort());
                }
            }

            return components;
        };

        // Calculer l'intermédiarité des arêtes (edge betweenness)
        const calculateEdgeBetweenness = (nodeList, edgeList) => {
            const adj = {};
            nodeList.forEach(id => { adj[id] = []; });
            edgeList.forEach(e => {
                adj[e.from].push(e.to);
                adj[e.to].push(e.from);
            });

            const betweenness = {};
            edgeList.forEach(e => { betweenness[e.key] = 0; });

            // Pour chaque noeud source, faire un BFS et compter les chemins
            for (const source of nodeList) {
                const dist = {};
                const paths = {};
                const sigma = {}; // Nombre de plus courts chemins

                nodeList.forEach(id => {
                    dist[id] = -1;
                    paths[id] = [];
                    sigma[id] = 0;
                });

                dist[source] = 0;
                sigma[source] = 1;
                const queue = [source];
                const stack = [];

                while (queue.length > 0) {
                    const v = queue.shift();
                    stack.push(v);

                    for (const w of (adj[v] || [])) {
                        if (dist[w] < 0) {
                            dist[w] = dist[v] + 1;
                            queue.push(w);
                        }
                        if (dist[w] === dist[v] + 1) {
                            sigma[w] += sigma[v];
                            paths[w].push(v);
                        }
                    }
                }

                // Accumuler les contributions
                const delta = {};
                nodeList.forEach(id => { delta[id] = 0; });

                while (stack.length > 0) {
                    const w = stack.pop();
                    for (const v of paths[w]) {
                        const edgeKey = [v, w].sort().join('|');
                        const contrib = (sigma[v] / sigma[w]) * (1 + delta[w]);
                        if (betweenness[edgeKey] !== undefined) {
                            betweenness[edgeKey] += contrib;
                        }
                        delta[v] += contrib;
                    }
                }
            }

            return betweenness;
        };

        // Algorithme principal: supprimer les arêtes à forte intermédiarité
        let currentEdges = [...edges];
        let bestCommunities = [nodeIds];
        let targetCommunities = Math.min(6, Math.ceil(n / 4)); // Cibler ~4-6 communautés

        const maxIterations = Math.min(edges.length, 20);

        for (let iter = 0; iter < maxIterations; iter++) {
            const components = findComponents(nodeIds, currentEdges);

            // Si on a atteint un bon nombre de communautés, s'arrêter
            if (components.length >= targetCommunities) {
                bestCommunities = components;
                break;
            }

            // Garder la meilleure partition jusqu'ici
            if (components.length > bestCommunities.length) {
                bestCommunities = components;
            }

            // Calculer l'intermédiarité et supprimer l'arête avec la plus haute valeur
            const betweenness = calculateEdgeBetweenness(nodeIds, currentEdges);

            let edgeToRemove = null;

            // Trier les arêtes par intermédiarité décroissante, puis par clé pour déterminisme
            const sortedEdges = currentEdges
                .map(e => ({ ...e, bet: betweenness[e.key] || 0 }))
                .sort((a, b) => {
                    if (b.bet !== a.bet) return b.bet - a.bet;
                    return a.key.localeCompare(b.key);
                });

            if (sortedEdges.length > 0 && sortedEdges[0].bet > 0) {
                edgeToRemove = sortedEdges[0].key;
            }

            if (!edgeToRemove) break;

            // Supprimer l'arête
            currentEdges = currentEdges.filter(e => e.key !== edgeToRemove);

            if (currentEdges.length === 0) break;
        }

        // Vérification finale
        const finalComponents = findComponents(nodeIds, currentEdges);
        if (finalComponents.length > bestCommunities.length) {
            bestCommunities = finalComponents;
        }

        // Filtrer et trier les communautés
        const minSize = this.socialNetworkConfig.minCommunitySize;

        return bestCommunities
            .filter(c => c.length >= minSize)
            .sort((a, b) => {
                if (b.length !== a.length) return b.length - a.length;
                return a[0].localeCompare(b[0]);
            })
            .slice(0, this.socialNetworkConfig.maxCommunities);
    },

    displayCommunities(communities, graphData) {
        const container = document.getElementById('sn-results-container');
        if (!container) return;

        const nodeMap = {};
        graphData.nodes.forEach(n => { nodeMap[n.id] = n; });

        let html = `
            <div class="sn-section">
                <h3><span class="material-icons">groups</span> Communautés détectées (${communities.length})</h3>
                <div class="sn-communities-list">
        `;

        const colors = this.getCommunityColors();

        communities.forEach((community, idx) => {
            const color = colors[idx % colors.length];
            const members = community.map(id => nodeMap[id]?.label || id).join(', ');

            html += `
                <div class="sn-community-card" data-community="${idx}" style="border-left: 4px solid ${color}">
                    <div class="sn-community-header">
                        <span class="sn-community-badge" style="background: ${color}">${idx + 1}</span>
                        <span class="sn-community-size">${community.length} membres</span>
                    </div>
                    <div class="sn-community-members">${members}</div>
                    <div class="sn-community-actions">
                        <button class="btn btn-sm btn-secondary" onclick="app.focusCommunity(${idx})">
                            <span class="material-icons">visibility</span> Voir
                        </button>
                        <button class="btn btn-sm btn-secondary" onclick="app.analyzeCommunity(${idx})">
                            <span class="material-icons">psychology</span> Analyser
                        </button>
                    </div>
                </div>
            `;
        });

        html += `
                </div>
            </div>
        `;

        container.innerHTML = html;
    },

    getCommunityColors() {
        return [
            '#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6',
            '#ec4899', '#06b6d4', '#84cc16', '#f97316', '#6366f1'
        ];
    },

    colorGraphByCommunity(communities) {
        if (!this.snNodes) return;

        const colors = this.getCommunityColors();
        const nodeToComm = {};

        communities.forEach((comm, idx) => {
            comm.forEach(nodeId => {
                nodeToComm[nodeId] = idx;
            });
        });

        const updates = this.snNodes.get().map(node => {
            const commIdx = nodeToComm[node.id];
            if (commIdx !== undefined) {
                return {
                    id: node.id,
                    color: {
                        background: colors[commIdx % colors.length],
                        border: colors[commIdx % colors.length]
                    }
                };
            }
            return {
                id: node.id,
                color: { background: '#e2e8f0', border: '#cbd5e0' }
            };
        });

        this.snNodes.update(updates);
    },

    focusCommunity(communityIdx) {
        if (!this.communities[communityIdx] || !this.socialNetworkGraph) return;

        const community = this.communities[communityIdx];

        // Mettre en évidence la communauté
        const updates = this.snNodes.get().map(node => {
            const inCommunity = community.includes(node.id);
            return {
                id: node.id,
                opacity: inCommunity ? 1 : 0.2,
                borderWidth: inCommunity ? 4 : 1
            };
        });

        this.snNodes.update(updates);

        // Focus sur les noeuds
        this.socialNetworkGraph.fit({
            nodes: community,
            animation: true
        });

        this.selectedCommunity = communityIdx;
    },

    async analyzeCommunity(communityIdx) {
        if (!this.communities[communityIdx]) return;

        const community = this.communities[communityIdx];
        const graphData = await this.getGraphData();
        const nodeMap = {};
        graphData.nodes.forEach(n => { nodeMap[n.id] = n; });

        const members = community.map(id => nodeMap[id]?.label || id);

        this.setAnalysisContext(
            'community_analysis',
            `Analyse de la communauté ${communityIdx + 1}`,
            `Membres: ${members.join(', ')}`
        );

        const analysisModal = document.getElementById('analysis-modal');
        const analysisContent = document.getElementById('analysis-content');
        const modalTitle = document.getElementById('analysis-modal-title');

        if (modalTitle) modalTitle.textContent = `Analyse de la Communauté ${communityIdx + 1}`;

        const noteBtn = document.getElementById('btn-save-to-notebook');
        if (noteBtn) noteBtn.style.display = '';

        analysisContent.innerHTML = `
            <div class="modal-explanation">
                <span class="material-icons">groups</span>
                <p><strong>Communauté ${communityIdx + 1}:</strong> ${members.join(', ')}</p>
            </div>
            <div id="community-analysis" class="ai-analysis"><span class="streaming-cursor">▊</span></div>
        `;
        analysisModal.classList.add('active');

        const analysisDiv = document.getElementById('community-analysis');
        await this.streamAIResponse(
            '/api/chat/stream',
            {
                case_id: this.currentCase.id,
                message: `Analyse cette communauté dans le réseau social de l'affaire "${this.currentCase.name}":

Membres: ${members.join(', ')}

Questions à analyser:
1. Quel est le lien qui unit ces personnes?
2. Quel rôle cette communauté joue-t-elle dans l'affaire?
3. Y a-t-il un leader apparent?
4. Quelles sont les implications pour l'enquête?`
            },
            analysisDiv
        );
    },

    // ============================================
    // Identification des brokers
    // ============================================
    async identifyBrokers() {
        const graphData = await this.getGraphData();
        if (!graphData.nodes || graphData.nodes.length < 3) {
            this.showToast('Pas assez de données pour identifier des brokers', 'warning');
            return;
        }

        this.showToast('Identification des brokers...', 'info');

        // Calculer la betweenness centrality
        const betweenness = this.calculateBetweennessCentrality(graphData);

        // Trier et identifier les top brokers
        const sortedNodes = Object.entries(betweenness)
            .sort((a, b) => b[1] - a[1])
            .filter(([_, score]) => score > this.socialNetworkConfig.brokerThreshold);

        this.brokers = sortedNodes.map(([nodeId, score]) => ({ nodeId, score }));

        // Afficher les résultats
        this.displayBrokers(this.brokers, graphData);

        // Mettre en évidence les brokers dans le graphe
        this.highlightBrokers(this.brokers);
    },

    calculateBetweennessCentrality(graphData) {
        const nodes = graphData.nodes.map(n => n.id);
        const adj = this.buildAdjacencyList(graphData);
        const betweenness = {};

        nodes.forEach(n => { betweenness[n] = 0; });

        // Pour chaque paire de noeuds, trouver les chemins les plus courts
        for (const source of nodes) {
            const { distances, paths } = this.bfsShortestPaths(source, adj, nodes);

            for (const target of nodes) {
                if (source === target || distances[target] === Infinity) continue;

                // Reconstruire les chemins et compter les passages
                const allPaths = this.reconstructPaths(source, target, paths);
                const pathCount = allPaths.length;

                if (pathCount === 0) continue;

                // Compter combien de fois chaque noeud intermédiaire apparaît
                const intermediates = {};
                allPaths.forEach(path => {
                    for (let i = 1; i < path.length - 1; i++) {
                        intermediates[path[i]] = (intermediates[path[i]] || 0) + 1;
                    }
                });

                for (const [nodeId, count] of Object.entries(intermediates)) {
                    betweenness[nodeId] += count / pathCount;
                }
            }
        }

        // Normaliser
        const n = nodes.length;
        const normFactor = (n - 1) * (n - 2) / 2;
        if (normFactor > 0) {
            for (const nodeId of nodes) {
                betweenness[nodeId] /= normFactor;
            }
        }

        return betweenness;
    },

    bfsShortestPaths(source, adj, nodes) {
        const distances = {};
        const paths = {};

        nodes.forEach(n => {
            distances[n] = Infinity;
            paths[n] = [];
        });

        distances[source] = 0;
        const queue = [source];

        while (queue.length > 0) {
            const current = queue.shift();
            const neighbors = adj[current] || [];

            for (const neighbor of neighbors) {
                if (distances[neighbor] === Infinity) {
                    distances[neighbor] = distances[current] + 1;
                    queue.push(neighbor);
                }

                if (distances[neighbor] === distances[current] + 1) {
                    paths[neighbor].push(current);
                }
            }
        }

        return { distances, paths };
    },

    reconstructPaths(source, target, paths) {
        if (source === target) return [[source]];

        const result = [];
        const stack = [[target, [target]]];

        while (stack.length > 0) {
            const [current, path] = stack.pop();

            if (current === source) {
                result.push(path.reverse());
                continue;
            }

            for (const prev of (paths[current] || [])) {
                stack.push([prev, [...path, prev]]);
            }
        }

        return result;
    },

    displayBrokers(brokers, graphData) {
        const container = document.getElementById('sn-results-container');
        if (!container) return;

        const nodeMap = {};
        graphData.nodes.forEach(n => { nodeMap[n.id] = n; });

        let html = `
            <div class="sn-section">
                <h3><span class="material-icons">mediation</span> Brokers identifiés (${brokers.length})</h3>
                <p class="sn-section-desc">Les brokers sont des personnes-ponts qui connectent différents groupes.</p>
                <div class="sn-brokers-list">
        `;

        if (brokers.length === 0) {
            html += `<p class="sn-no-results">Aucun broker significatif détecté</p>`;
        } else {
            brokers.forEach((broker, idx) => {
                const node = nodeMap[broker.nodeId];
                const label = node?.label || broker.nodeId;
                const scorePercent = Math.round(broker.score * 100);

                html += `
                    <div class="sn-broker-card">
                        <div class="sn-broker-rank">#${idx + 1}</div>
                        <div class="sn-broker-info">
                            <div class="sn-broker-name">${label}</div>
                            <div class="sn-broker-type">${node?.type || 'entité'}</div>
                        </div>
                        <div class="sn-broker-score">
                            <div class="sn-score-bar" style="width: ${scorePercent}%"></div>
                            <span class="sn-score-value">${scorePercent}%</span>
                        </div>
                        <div class="sn-broker-actions">
                            <button class="btn btn-sm btn-icon" onclick="app.focusBroker('${broker.nodeId}')" data-tooltip="Voir ce broker sur le graphe">
                                <span class="material-icons">visibility</span>
                            </button>
                            <button class="btn btn-sm btn-icon" onclick="app.analyzeBroker('${broker.nodeId}')" data-tooltip="Analyser ce broker avec l'IA">
                                <span class="material-icons">psychology</span>
                            </button>
                        </div>
                    </div>
                `;
            });
        }

        html += `
                </div>
            </div>
        `;

        container.innerHTML = html;
    },

    highlightBrokers(brokers) {
        if (!this.snNodes) return;

        const brokerIds = new Set(brokers.map(b => b.nodeId));

        const updates = this.snNodes.get().map(node => {
            const isBroker = brokerIds.has(node.id);
            return {
                id: node.id,
                size: isBroker ? 30 : 15,
                borderWidth: isBroker ? 4 : 2,
                color: isBroker ? { background: '#ef4444', border: '#dc2626' } : undefined
            };
        });

        this.snNodes.update(updates);
    },

    focusBroker(nodeId) {
        if (!this.socialNetworkGraph) return;

        // Mettre en évidence le broker et ses connexions
        const graphData = { nodes: this.snNodes.get(), edges: this.snEdges.get() };
        const neighbors = new Set([nodeId]);

        graphData.edges.forEach(e => {
            if (e.from === nodeId) neighbors.add(e.to);
            if (e.to === nodeId) neighbors.add(e.from);
        });

        const updates = this.snNodes.get().map(node => ({
            id: node.id,
            opacity: neighbors.has(node.id) ? 1 : 0.2,
            borderWidth: node.id === nodeId ? 6 : (neighbors.has(node.id) ? 3 : 1)
        }));

        this.snNodes.update(updates);
        this.socialNetworkGraph.focus(nodeId, { scale: 1.5, animation: true });
    },

    async analyzeBroker(nodeId) {
        const graphData = await this.getGraphData();
        const node = graphData.nodes.find(n => n.id === nodeId);
        if (!node) return;

        // Trouver les groupes que ce broker connecte
        const neighbors = [];
        graphData.edges.forEach(e => {
            if (e.from === nodeId) neighbors.push(e.to);
            if (e.to === nodeId) neighbors.push(e.from);
        });

        const nodeMap = {};
        graphData.nodes.forEach(n => { nodeMap[n.id] = n; });
        const neighborNames = neighbors.map(id => nodeMap[id]?.label || id);

        this.setAnalysisContext(
            'broker_analysis',
            `Analyse du broker: ${node.label}`,
            `Connecte: ${neighborNames.join(', ')}`
        );

        const analysisModal = document.getElementById('analysis-modal');
        const analysisContent = document.getElementById('analysis-content');
        const modalTitle = document.getElementById('analysis-modal-title');

        if (modalTitle) modalTitle.textContent = `Analyse du Broker: ${node.label}`;

        const noteBtn = document.getElementById('btn-save-to-notebook');
        if (noteBtn) noteBtn.style.display = '';

        analysisContent.innerHTML = `
            <div class="modal-explanation">
                <span class="material-icons">mediation</span>
                <p><strong>${node.label}</strong> connecte ${neighbors.length} entités</p>
            </div>
            <div id="broker-analysis" class="ai-analysis"><span class="streaming-cursor">▊</span></div>
        `;
        analysisModal.classList.add('active');

        const analysisDiv = document.getElementById('broker-analysis');
        await this.streamAIResponse(
            '/api/chat/stream',
            {
                case_id: this.currentCase.id,
                message: `Analyse le rôle de "${node.label}" comme broker (personne-pont) dans l'affaire "${this.currentCase.name}":

Ce broker connecte les entités suivantes: ${neighborNames.join(', ')}

Questions à analyser:
1. Pourquoi cette personne est-elle au centre de ces connexions?
2. Quel pouvoir ou influence cela lui confère-t-il?
3. Si on la "supprime" du réseau, quels groupes seraient déconnectés?
4. Implications pour l'enquête et la surveillance?`
            },
            analysisDiv
        );
    },

    // ============================================
    // Analyse des flux
    // ============================================
    async analyzeFlows() {
        const graphData = await this.getGraphData();
        if (!graphData.edges || graphData.edges.length === 0) {
            this.showToast('Pas de relations pour analyser les flux', 'warning');
            return;
        }

        this.showToast('Analyse des flux...', 'info');

        // Classifier les relations par type de flux
        const flowTypes = {
            information: [],
            argent: [],
            influence: [],
            preuve: []
        };

        // Mots-clés pour classifier les relations
        const keywords = {
            information: ['téléphone', 'message', 'email', 'dit', 'informe', 'communique', 'rencontre', 'appel'],
            argent: ['paie', 'donne', 'transfert', 'transaction', 'argent', 'virement', 'achète', 'vend'],
            influence: ['ordonne', 'contrôle', 'dirige', 'emploie', 'supervise', 'menace', 'influence'],
            preuve: ['possède', 'détient', 'trouvé', 'vu', 'témoin', 'preuve', 'indice']
        };

        graphData.edges.forEach(edge => {
            const label = (edge.label || '').toLowerCase();
            for (const [type, words] of Object.entries(keywords)) {
                if (words.some(w => label.includes(w))) {
                    flowTypes[type].push(edge);
                    break;
                }
            }
        });

        this.flowAnalysis = flowTypes;
        this.displayFlowAnalysis(flowTypes, graphData);
    },

    displayFlowAnalysis(flowTypes, graphData) {
        const container = document.getElementById('sn-results-container');
        if (!container) return;

        const nodeMap = {};
        graphData.nodes.forEach(n => { nodeMap[n.id] = n; });

        const flowIcons = {
            information: 'chat',
            argent: 'payments',
            influence: 'admin_panel_settings',
            preuve: 'fingerprint'
        };

        const flowColors = {
            information: '#3b82f6',
            argent: '#10b981',
            influence: '#f59e0b',
            preuve: '#8b5cf6'
        };

        let html = `
            <div class="sn-section">
                <h3><span class="material-icons">swap_horiz</span> Analyse des Flux</h3>
                <div class="sn-flow-selector">
                    <label>Type de flux:</label>
                    <select id="flow-type-select" class="form-select">
                        <option value="all">Tous les flux</option>
                        ${Object.keys(flowTypes).map(t =>
                            `<option value="${t}">${t.charAt(0).toUpperCase() + t.slice(1)} (${flowTypes[t].length})</option>`
                        ).join('')}
                    </select>
                </div>
                <div class="sn-flow-summary">
        `;

        for (const [type, edges] of Object.entries(flowTypes)) {
            html += `
                <div class="sn-flow-card" style="border-color: ${flowColors[type]}">
                    <span class="material-icons" style="color: ${flowColors[type]}">${flowIcons[type]}</span>
                    <div class="sn-flow-info">
                        <div class="sn-flow-type">${type.charAt(0).toUpperCase() + type.slice(1)}</div>
                        <div class="sn-flow-count">${edges.length} relations</div>
                    </div>
                </div>
            `;
        }

        html += `
                </div>
                <div class="sn-flow-details" id="sn-flow-details">
                    <p class="sn-hint">Sélectionnez un type de flux pour voir les détails</p>
                </div>
                <button class="btn btn-primary" id="btn-ai-flow-analysis" onclick="app.aiFlowAnalysis()">
                    <span class="material-icons">psychology</span> Analyse IA des flux
                </button>
            </div>
        `;

        container.innerHTML = html;

        // Réattacher le listener
        document.getElementById('flow-type-select')?.addEventListener('change', (e) => {
            this.updateFlowVisualization(e.target.value);
        });
    },

    updateFlowVisualization(flowType) {
        if (!this.flowAnalysis || !this.snEdges) return;

        const flowColors = {
            information: '#3b82f6',
            argent: '#10b981',
            influence: '#f59e0b',
            preuve: '#8b5cf6'
        };

        const detailsContainer = document.getElementById('sn-flow-details');

        if (flowType === 'all') {
            // Réinitialiser les couleurs
            const updates = this.snEdges.get().map(e => ({
                id: e.id,
                color: { color: '#1e3a5f', opacity: 0.6 },
                width: 1
            }));
            this.snEdges.update(updates);

            if (detailsContainer) {
                detailsContainer.innerHTML = '<p class="sn-hint">Sélectionnez un type de flux pour voir les détails</p>';
            }
            return;
        }

        const selectedEdges = this.flowAnalysis[flowType] || [];
        const selectedIds = new Set(selectedEdges.map((e, i) => `edge-${i}`));

        // Mettre à jour la visualisation
        const updates = this.snEdges.get().map(e => {
            // Trouver si cet edge correspond au flux sélectionné
            const matchingEdge = selectedEdges.find(se => se.from === e.from && se.to === e.to);

            return {
                id: e.id,
                color: matchingEdge ? { color: flowColors[flowType], opacity: 1 } : { color: '#e2e8f0', opacity: 0.2 },
                width: matchingEdge ? 3 : 1
            };
        });

        this.snEdges.update(updates);

        // Afficher les détails
        if (detailsContainer && selectedEdges.length > 0) {
            const graphData = { nodes: this.snNodes.get() };
            const nodeMap = {};
            graphData.nodes.forEach(n => { nodeMap[n.id] = n; });

            let detailsHtml = `<div class="sn-flow-list">`;
            selectedEdges.forEach(edge => {
                const from = nodeMap[edge.from]?.label || edge.from;
                const to = nodeMap[edge.to]?.label || edge.to;
                detailsHtml += `
                    <div class="sn-flow-item">
                        <span>${from}</span>
                        <span class="sn-flow-arrow" style="color: ${flowColors[flowType]}">→ ${edge.label || ''} →</span>
                        <span>${to}</span>
                    </div>
                `;
            });
            detailsHtml += `</div>`;
            detailsContainer.innerHTML = detailsHtml;
        }
    },

    async aiFlowAnalysis() {
        if (!this.flowAnalysis) return;

        const summary = Object.entries(this.flowAnalysis)
            .map(([type, edges]) => `${type}: ${edges.length} relations`)
            .join('\n');

        this.setAnalysisContext('flow_analysis', 'Analyse des flux', summary);

        const analysisModal = document.getElementById('analysis-modal');
        const analysisContent = document.getElementById('analysis-content');
        const modalTitle = document.getElementById('analysis-modal-title');

        if (modalTitle) modalTitle.textContent = 'Analyse IA des Flux';

        const noteBtn = document.getElementById('btn-save-to-notebook');
        if (noteBtn) noteBtn.style.display = '';

        analysisContent.innerHTML = `
            <div class="modal-explanation">
                <span class="material-icons">swap_horiz</span>
                <p>Analyse des flux dans le réseau</p>
            </div>
            <div id="flow-ai-analysis" class="ai-analysis"><span class="streaming-cursor">▊</span></div>
        `;
        analysisModal.classList.add('active');

        const analysisDiv = document.getElementById('flow-ai-analysis');
        await this.streamAIResponse(
            '/api/chat/stream',
            {
                case_id: this.currentCase.id,
                message: `Analyse les flux dans le réseau social de l'affaire "${this.currentCase.name}":

Résumé des flux:
${summary}

Questions à analyser:
1. Quels sont les patterns de communication/information?
2. Y a-t-il des flux d'argent suspects?
3. Qui exerce de l'influence sur qui?
4. Comment les preuves circulent-elles?
5. Y a-t-il des schémas inhabituels ou suspects?`
            },
            analysisDiv
        );
    },

    // ============================================
    // Évolution temporelle
    // ============================================
    async showTemporalEvolution() {
        if (!this.currentCase?.timeline || this.currentCase.timeline.length === 0) {
            this.showToast('Pas d\'événements dans la timeline pour l\'évolution temporelle', 'warning');
            return;
        }

        this.showToast('Calcul de l\'évolution temporelle...', 'info');

        // Grouper les événements par période
        const events = [...this.currentCase.timeline].sort((a, b) =>
            new Date(a.timestamp) - new Date(b.timestamp)
        );

        // Créer des snapshots du réseau à différents moments
        const snapshots = this.buildTemporalSnapshots(events);
        this.temporalSnapshots = snapshots;

        this.displayTemporalEvolution(snapshots);
    },

    buildTemporalSnapshots(events) {
        const snapshots = [];
        const graphData = { nodes: this.snNodes?.get() || [], edges: this.snEdges?.get() || [] };

        // Créer un snapshot initial (vide)
        const activeNodes = new Set();
        const activeEdges = new Set();

        // Pour chaque événement, ajouter les entités impliquées
        events.forEach((event, idx) => {
            const entities = event.entities || [];
            entities.forEach(e => activeNodes.add(e));

            // Créer des liens entre les entités de l'événement
            for (let i = 0; i < entities.length; i++) {
                for (let j = i + 1; j < entities.length; j++) {
                    activeEdges.add(`${entities[i]}-${entities[j]}`);
                }
            }

            snapshots.push({
                timestamp: event.timestamp,
                title: event.title,
                nodes: new Set(activeNodes),
                edges: new Set(activeEdges),
                eventIdx: idx
            });
        });

        return snapshots;
    },

    displayTemporalEvolution(snapshots) {
        const container = document.getElementById('sn-results-container');
        if (!container) return;

        let html = `
            <div class="sn-section">
                <h3><span class="material-icons">auto_graph</span> Évolution Temporelle du Réseau</h3>
                <p class="sn-section-desc">Visualisez comment le réseau social évolue au fil des événements.</p>

                <div class="sn-temporal-controls">
                    <input type="range" id="temporal-slider" min="0" max="${snapshots.length - 1}" value="0" class="sn-temporal-slider">
                    <div class="sn-temporal-info" id="temporal-info">
                        <span class="sn-temporal-date">${this.formatDate(snapshots[0]?.timestamp)}</span>
                        <span class="sn-temporal-event">${snapshots[0]?.title || ''}</span>
                    </div>
                    <div class="sn-temporal-buttons">
                        <button class="btn btn-sm btn-secondary" id="btn-temporal-start" onclick="app.temporalToStart()">
                            <span class="material-icons">skip_previous</span>
                        </button>
                        <button class="btn btn-sm btn-primary" id="btn-play-temporal">
                            <span class="material-icons">play_arrow</span>
                        </button>
                        <button class="btn btn-sm btn-secondary" id="btn-temporal-end" onclick="app.temporalToEnd()">
                            <span class="material-icons">skip_next</span>
                        </button>
                    </div>
                </div>

                <div class="sn-temporal-stats">
                    <div class="sn-temporal-stat">
                        <span class="sn-stat-value" id="temporal-nodes-count">0</span>
                        <span class="sn-stat-label">Noeuds actifs</span>
                    </div>
                    <div class="sn-temporal-stat">
                        <span class="sn-stat-value" id="temporal-edges-count">0</span>
                        <span class="sn-stat-label">Relations</span>
                    </div>
                    <div class="sn-temporal-stat">
                        <span class="sn-stat-value" id="temporal-event-idx">1/${snapshots.length}</span>
                        <span class="sn-stat-label">Événement</span>
                    </div>
                </div>
            </div>
        `;

        container.innerHTML = html;

        // Réattacher les listeners
        document.getElementById('temporal-slider')?.addEventListener('input', (e) => {
            this.showTemporalSnapshot(parseInt(e.target.value));
        });

        document.getElementById('btn-play-temporal')?.addEventListener('click', () => {
            this.playTemporalAnimation();
        });

        // Afficher le premier snapshot
        this.showTemporalSnapshot(0);
    },

    showTemporalSnapshot(index) {
        if (!this.temporalSnapshots || !this.temporalSnapshots[index]) return;

        const snapshot = this.temporalSnapshots[index];

        // Mettre à jour les infos
        const infoEl = document.getElementById('temporal-info');
        if (infoEl) {
            infoEl.innerHTML = `
                <span class="sn-temporal-date">${this.formatDate(snapshot.timestamp)}</span>
                <span class="sn-temporal-event">${snapshot.title || ''}</span>
            `;
        }

        const nodesCountEl = document.getElementById('temporal-nodes-count');
        if (nodesCountEl) nodesCountEl.textContent = snapshot.nodes.size;

        const edgesCountEl = document.getElementById('temporal-edges-count');
        if (edgesCountEl) edgesCountEl.textContent = snapshot.edges.size;

        const eventIdxEl = document.getElementById('temporal-event-idx');
        if (eventIdxEl) eventIdxEl.textContent = `${index + 1}/${this.temporalSnapshots.length}`;

        // Mettre à jour le graphe
        if (this.snNodes) {
            const updates = this.snNodes.get().map(node => ({
                id: node.id,
                opacity: snapshot.nodes.has(node.id) ? 1 : 0.1,
                borderWidth: snapshot.nodes.has(node.id) ? 3 : 1
            }));
            this.snNodes.update(updates);
        }

        if (this.snEdges) {
            const updates = this.snEdges.get().map(edge => {
                const edgeKey1 = `${edge.from}-${edge.to}`;
                const edgeKey2 = `${edge.to}-${edge.from}`;
                const isActive = snapshot.edges.has(edgeKey1) || snapshot.edges.has(edgeKey2);
                return {
                    id: edge.id,
                    color: isActive ? { color: '#3b82f6', opacity: 1 } : { color: '#e2e8f0', opacity: 0.1 },
                    width: isActive ? 2 : 1
                };
            });
            this.snEdges.update(updates);
        }
    },

    temporalToStart() {
        const slider = document.getElementById('temporal-slider');
        if (slider) {
            slider.value = 0;
            this.showTemporalSnapshot(0);
        }
    },

    temporalToEnd() {
        const slider = document.getElementById('temporal-slider');
        if (slider && this.temporalSnapshots) {
            slider.value = this.temporalSnapshots.length - 1;
            this.showTemporalSnapshot(this.temporalSnapshots.length - 1);
        }
    },

    isTemporalPlaying: false,
    temporalAnimationInterval: null,

    playTemporalAnimation() {
        const playBtn = document.getElementById('btn-play-temporal');
        const slider = document.getElementById('temporal-slider');

        if (this.isTemporalPlaying) {
            // Stop
            this.isTemporalPlaying = false;
            if (this.temporalAnimationInterval) {
                clearInterval(this.temporalAnimationInterval);
            }
            if (playBtn) playBtn.innerHTML = '<span class="material-icons">play_arrow</span>';
        } else {
            // Play
            this.isTemporalPlaying = true;
            if (playBtn) playBtn.innerHTML = '<span class="material-icons">pause</span>';

            let currentIdx = parseInt(slider?.value || 0);

            this.temporalAnimationInterval = setInterval(() => {
                if (!this.isTemporalPlaying) return;

                currentIdx++;
                if (currentIdx >= this.temporalSnapshots.length) {
                    currentIdx = 0;
                }

                if (slider) slider.value = currentIdx;
                this.showTemporalSnapshot(currentIdx);
            }, 1500);
        }
    },

    formatDate(dateStr) {
        if (!dateStr) return 'Date inconnue';
        try {
            const date = new Date(dateStr);
            return date.toLocaleDateString('fr-FR', {
                day: 'numeric',
                month: 'short',
                year: 'numeric',
                hour: '2-digit',
                minute: '2-digit'
            });
        } catch {
            return dateStr;
        }
    },

    // ============================================
    // Export du module
    // ============================================
    async exportSocialNetworkAnalysis() {
        const report = {
            case: this.currentCase?.name,
            date: new Date().toISOString(),
            communities: this.communities,
            brokers: this.brokers,
            flowAnalysis: this.flowAnalysis,
            temporalSnapshots: this.temporalSnapshots?.length || 0
        };

        const blob = new Blob([JSON.stringify(report, null, 2)], { type: 'application/json' });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `social-network-analysis-${this.currentCase?.id || 'export'}.json`;
        a.click();
        URL.revokeObjectURL(url);

        this.showToast('Analyse exportée', 'success');
    }
};

// Export for use in main app
if (typeof module !== 'undefined' && module.exports) {
    module.exports = SocialNetworkModule;
}
