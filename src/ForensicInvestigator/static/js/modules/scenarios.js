// ForensicInvestigator - Module Scenarios
// Simulation de Scénarios "What-If"

const ScenariosModule = {
    // State
    scenarios: [],
    selectedScenario: null,
    comparisonScenario1: null,
    comparisonScenario2: null,
    scenarioGraph: null,

    // ============================================
    // Load Scenarios
    // ============================================
    async loadScenarios() {
        if (!this.currentCase) return;

        try {
            const scenarios = await this.apiCall(`/api/scenarios?case_id=${this.currentCase.id}`);
            this.scenarios = scenarios || [];
            this.renderScenariosList();
        } catch (error) {
            console.error('Error loading scenarios:', error);
            this.scenarios = [];
            this.renderScenariosList();
        }
    },

    renderScenariosList() {
        const container = document.getElementById('scenarios-list');
        if (!container) return;

        if (this.scenarios.length === 0) {
            container.innerHTML = `
                <div class="empty-state">
                    <span class="material-icons empty-state-icon">psychology_alt</span>
                    <p class="empty-state-title">Aucun scénario</p>
                    <p class="empty-state-description">Créez un scénario "What-If" pour explorer différentes hypothèses</p>
                </div>
            `;
            return;
        }

        container.innerHTML = this.scenarios.map(scenario => {
            const plausibilityClass = this.getPlausibilityClass(scenario.plausibility_score);
            const typeLabel = this.getAssumptionTypeLabel(scenario.assumption_type);
            const typeIcon = this.getAssumptionTypeIcon(scenario.assumption_type);
            const statusClass = scenario.status === 'active' ? 'status-active' : 'status-archived';
            const implicationsCount = scenario.implications?.length || 0;
            const supportingCount = scenario.supporting_facts?.length || 0;
            const contradictingCount = scenario.contradicting_facts?.length || 0;

            return `
                <div class="scenario-card-new ${statusClass} ${this.selectedScenario?.id === scenario.id ? 'selected' : ''}"
                     data-scenario-id="${scenario.id}" onclick="app.selectScenario('${scenario.id}')">
                    <div class="scenario-card-header">
                        <div class="scenario-type-icon ${scenario.assumption_type}">
                            <span class="material-icons">${typeIcon}</span>
                        </div>
                        <div class="scenario-card-info">
                            <span class="scenario-type-label">${typeLabel}</span>
                            <h4>${scenario.name}</h4>
                        </div>
                        <div class="scenario-plausibility-badge ${plausibilityClass}">
                            <span class="plausibility-value">${scenario.plausibility_score}%</span>
                            <div class="plausibility-bar-mini">
                                <div class="plausibility-fill-mini" style="width: ${scenario.plausibility_score}%"></div>
                            </div>
                        </div>
                    </div>

                    <div class="scenario-assumption-box">
                        <span class="material-icons">format_quote</span>
                        <p>${scenario.assumption}</p>
                    </div>

                    <div class="scenario-metrics-row">
                        <div class="scenario-metric">
                            <span class="material-icons">account_tree</span>
                            <span class="metric-number">${implicationsCount}</span>
                            <span class="metric-text">implications</span>
                        </div>
                        <div class="scenario-metric supporting">
                            <span class="material-icons">check_circle</span>
                            <span class="metric-number">${supportingCount}</span>
                            <span class="metric-text">pour</span>
                        </div>
                        <div class="scenario-metric contradicting">
                            <span class="material-icons">cancel</span>
                            <span class="metric-number">${contradictingCount}</span>
                            <span class="metric-text">contre</span>
                        </div>
                    </div>

                    <div class="scenario-actions-grid">
                        <button class="btn btn-action" onclick="event.stopPropagation(); app.simulateScenario('${scenario.id}')" data-tooltip="Analyser avec l'IA">
                            <span class="material-icons">psychology</span>
                        </button>
                        <button class="btn btn-action" onclick="event.stopPropagation(); app.propagateScenario('${scenario.id}')" data-tooltip="Propager les implications">
                            <span class="material-icons">account_tree</span>
                        </button>
                        <button class="btn btn-action" onclick="event.stopPropagation(); app.selectScenarioForComparison('${scenario.id}')" data-tooltip="Comparer">
                            <span class="material-icons">compare_arrows</span>
                        </button>
                        <button class="btn btn-action danger" onclick="event.stopPropagation(); app.deleteScenario('${scenario.id}')" data-tooltip="Supprimer">
                            <span class="material-icons">delete</span>
                        </button>
                    </div>
                </div>
            `;
        }).join('');
    },

    getAssumptionTypeIcon(type) {
        const icons = {
            'guilt': 'gavel',
            'presence': 'place',
            'motive': 'psychology',
            'timeline': 'schedule',
            'relation': 'people'
        };
        return icons[type] || 'help_outline';
    },

    // ============================================
    // Auto-Generate Scenarios with AI
    // ============================================
    async generateScenariosAI() {
        if (!this.currentCase) {
            this.showToast('Sélectionnez une affaire d\'abord', 'warning');
            return;
        }

        const btn = document.getElementById('btn-auto-scenarios');
        if (btn) {
            btn.disabled = true;
            btn.innerHTML = '<span class="material-icons rotating">sync</span> Génération...';
        }

        try {
            this.showToast('Analyse de l\'affaire en cours...', 'info');

            const result = await this.apiCall(`/api/scenarios/generate?case_id=${this.currentCase.id}`, 'POST', {});

            if (result && result.scenarios && result.scenarios.length > 0) {
                this.showToast(`${result.scenarios.length} scénario(s) généré(s) avec succès`, 'success');
                await this.loadScenarios();

                // Sélectionner le premier scénario généré
                if (result.scenarios[0]?.id) {
                    this.selectScenario(result.scenarios[0].id);
                }
            } else {
                this.showToast('Aucun scénario pertinent trouvé', 'warning');
            }
        } catch (error) {
            console.error('Error generating scenarios:', error);
            this.showToast('Erreur lors de la génération: ' + (error.message || 'Erreur inconnue'), 'error');
        } finally {
            if (btn) {
                btn.disabled = false;
                btn.innerHTML = '<span class="material-icons">auto_awesome</span> Auto';
            }
        }
    },

    // ============================================
    // Create Scenario Modal
    // ============================================
    showCreateScenarioModal() {
        if (!this.currentCase) {
            this.showToast('Selectionnez une affaire d\'abord', 'warning');
            return;
        }

        const entities = this.currentCase.entities || [];
        const entityOptions = entities.map(e =>
            `<option value="${e.id}">${e.name} (${e.type} - ${e.role})</option>`
        ).join('');

        const content = `
            <div class="modal-explanation">
                <span class="material-icons">psychology_alt</span>
                <p>Creez un scenario "What-If" pour explorer une hypothese alternative.
                Le systeme analysera automatiquement les implications et calculera un score de plausibilite.</p>
            </div>
            <form id="scenario-form">
                <div class="form-group">
                    <label class="form-label">Type d'hypothese</label>
                    <select class="form-select" id="scenario-type" required>
                        <option value="guilt">Culpabilite - "X est coupable"</option>
                        <option value="presence">Presence - "X etait present a Y"</option>
                        <option value="motive">Mobile - "X avait le mobile Y"</option>
                        <option value="timeline">Chronologie - "L'evenement X s'est produit a Y"</option>
                        <option value="relation">Relation - "X et Y sont lies par Z"</option>
                    </select>
                </div>
                <div class="form-group">
                    <label class="form-label">Entite cible (optionnel)</label>
                    <select class="form-select" id="scenario-entity">
                        <option value="">-- Aucune entite specifique --</option>
                        ${entityOptions}
                    </select>
                </div>
                <div class="form-group">
                    <label class="form-label">Hypothese</label>
                    <textarea class="form-textarea" id="scenario-assumption" required
                        placeholder="Ex: 'Jean Dupont est le coupable du meurtre'"></textarea>
                </div>
            </form>
        `;

        this.showModal('Creer un Scenario What-If', content, async () => {
            const assumptionEl = document.getElementById('scenario-assumption');
            const typeEl = document.getElementById('scenario-type');
            const entityEl = document.getElementById('scenario-entity');

            if (!assumptionEl || !typeEl) {
                console.error('Form elements not found');
                return false;
            }

            const assumption = assumptionEl.value;
            const assumptionType = typeEl.value;
            const targetEntityId = entityEl ? entityEl.value : '';

            if (!assumption.trim()) {
                this.showToast('Veuillez saisir une hypothese', 'warning');
                return false;
            }

            try {
                const scenario = await this.apiCall(`/api/scenarios?case_id=${this.currentCase.id}`, 'POST', {
                    assumption,
                    assumption_type: assumptionType,
                    target_entity_id: targetEntityId || undefined
                });

                this.showToast('Scenario cree avec succes');
                await this.loadScenarios();
                this.selectScenario(scenario.id);
                return true;
            } catch (error) {
                console.error('Error creating scenario:', error);
                this.showToast('Erreur lors de la creation du scenario', 'error');
                return false;
            }
        });
    },

    // ============================================
    // Select Scenario
    // ============================================
    selectScenario(scenarioId) {
        this.selectedScenario = this.scenarios.find(s => s.id === scenarioId);

        // Update selection UI
        document.querySelectorAll('.scenario-card-new').forEach(card => {
            card.classList.toggle('selected', card.dataset.scenarioId === scenarioId);
        });

        // Update detail panel
        this.renderScenarioDetail();

        // Update graph if available
        if (this.selectedScenario?.modified_graph) {
            this.renderScenarioGraph();
        }
    },

    renderScenarioDetail() {
        const container = document.getElementById('scenario-detail');
        if (!container) return;

        if (!this.selectedScenario) {
            container.innerHTML = `
                <div class="empty-state">
                    <span class="material-icons empty-state-icon">touch_app</span>
                    <p class="empty-state-description">Selectionnez un scenario pour voir les details</p>
                </div>
            `;
            return;
        }

        const scenario = this.selectedScenario;
        const plausibilityClass = this.getPlausibilityClass(scenario.plausibility_score);

        container.innerHTML = `
            <div class="scenario-detail-header">
                <h3>${scenario.name}</h3>
                <div class="scenario-plausibility-large ${plausibilityClass}">
                    <div class="plausibility-score">${scenario.plausibility_score}%</div>
                    <div class="plausibility-label">Plausibilite</div>
                </div>
            </div>

            <div class="scenario-detail-section">
                <h4><span class="material-icons">lightbulb</span> Hypothese</h4>
                <blockquote class="scenario-assumption-quote">"${scenario.assumption}"</blockquote>
            </div>

            ${scenario.implications?.length > 0 ? `
                <div class="scenario-detail-section">
                    <h4><span class="material-icons">share</span> Implications detectees</h4>
                    <div class="implications-list">
                        ${scenario.implications.map(impl => `
                            <div class="implication-item ${impl.impact}">
                                <span class="implication-type-badge ${impl.type}">${this.getImplicationTypeLabel(impl.type)}</span>
                                <span class="implication-description">${this.resolveEntityIds(impl.description)}</span>
                                <span class="implication-confidence">${impl.confidence}%</span>
                            </div>
                        `).join('')}
                    </div>
                </div>
            ` : ''}

            ${scenario.supporting_facts?.length > 0 ? `
                <div class="scenario-detail-section">
                    <h4><span class="material-icons" style="color: var(--success);">check_circle</span> Faits a l'appui</h4>
                    <ul class="facts-list supporting">
                        ${scenario.supporting_facts.map(fact => `<li>${fact}</li>`).join('')}
                    </ul>
                </div>
            ` : ''}

            ${scenario.contradicting_facts?.length > 0 ? `
                <div class="scenario-detail-section">
                    <h4><span class="material-icons" style="color: var(--danger);">cancel</span> Faits contradictoires</h4>
                    <ul class="facts-list contradicting">
                        ${scenario.contradicting_facts.map(fact => `<li>${fact}</li>`).join('')}
                    </ul>
                </div>
            ` : ''}

            ${scenario.ai_analysis ? `
                <div class="scenario-detail-section">
                    <h4><span class="material-icons">psychology</span> Analyse IA</h4>
                    <div class="ai-analysis-content">${this.formatMarkdown ? this.formatMarkdown(scenario.ai_analysis) : scenario.ai_analysis}</div>
                </div>
            ` : ''}
        `;
    },

    // ============================================
    // Scenario Graph Visualization
    // ============================================
    renderScenarioGraph() {
        const container = document.getElementById('scenario-graph-container');
        if (!container || !this.selectedScenario?.modified_graph) return;

        const graphData = this.selectedScenario.modified_graph;
        console.log('Scenario Graph Data:', JSON.stringify(graphData, null, 2));
        console.log('Target Entity ID:', this.selectedScenario.target_entity_id);
        console.log('Implications:', this.selectedScenario.implications);

        // Stocker les données originales du graphe pour la réinitialisation
        this.scenarioGraphOriginalData = {
            nodes: graphData.nodes || [],
            edges: graphData.edges || []
        };

        this.renderScenarioGraphWithData(this.scenarioGraphOriginalData.nodes, this.scenarioGraphOriginalData.edges, false);
    },

    renderScenarioGraphWithData(nodesData, edgesData, isFiltered = false) {
        const container = document.getElementById('scenario-graph-container');
        if (!container) return;

        const nodes = nodesData.map(node => ({
            id: node.id,
            label: node.label,
            color: this.getScenarioNodeColor(node),
            shape: this.getNodeShape ? this.getNodeShape(node.type) : 'ellipse',
            font: { color: node.data?.scenario_modified === 'true' ? '#fff' : '#333' },
            borderWidth: node.data?.scenario_modified === 'true' ? 3 : 1,
            borderWidthSelected: 4
        }));

        const edges = edgesData.map(edge => ({
            from: edge.from,
            to: edge.to,
            label: edge.label,
            color: edge.context?.includes('scenario:') ? '#f97316' : '#94a3b8',
            dashes: edge.type === 'implication',
            arrows: 'to'
        }));

        if (this.scenarioGraph) {
            this.scenarioGraph.destroy();
        }

        const data = {
            nodes: new vis.DataSet(nodes),
            edges: new vis.DataSet(edges)
        };

        const options = {
            physics: { enabled: true, stabilization: { iterations: 100 } },
            edges: { smooth: { type: 'curvedCW', roundness: 0.2 } },
            nodes: { font: { size: 12 } },
            interaction: { hover: true, tooltipDelay: 100 }
        };

        this.scenarioGraph = new vis.Network(container, data, options);
        this.scenarioGraphFiltered = isFiltered;

        // Gérer le clic sur un nœud pour filtrer, ou en dehors pour réinitialiser
        this.scenarioGraph.on('click', (params) => {
            if (params.nodes.length > 0) {
                const clickedNodeId = params.nodes[0];
                this.filterScenarioGraphByNode(clickedNodeId);
            } else if (this.scenarioGraphFiltered) {
                // Clic en dehors d'un nœud et le graphe est filtré -> réinitialiser
                this.resetScenarioGraph();
            }
        });

        // Afficher/masquer le bouton reset
        this.updateScenarioGraphResetButton(isFiltered);
    },

    filterScenarioGraphByNode(nodeId) {
        if (!this.scenarioGraphOriginalData) return;

        const originalNodes = this.scenarioGraphOriginalData.nodes;
        const originalEdges = this.scenarioGraphOriginalData.edges;

        // Trouver toutes les relations du nœud cliqué
        const relatedEdges = originalEdges.filter(edge =>
            edge.from === nodeId || edge.to === nodeId
        );

        // Trouver tous les nœuds connectés
        const connectedNodeIds = new Set([nodeId]);
        relatedEdges.forEach(edge => {
            connectedNodeIds.add(edge.from);
            connectedNodeIds.add(edge.to);
        });

        // Filtrer les nœuds pour ne garder que le nœud cliqué et ses relations
        const filteredNodes = originalNodes.filter(node => connectedNodeIds.has(node.id));

        // Recréer le graphe avec les données filtrées
        this.renderScenarioGraphWithData(filteredNodes, relatedEdges, true);

        // Afficher le nom du nœud filtré
        const clickedNode = originalNodes.find(n => n.id === nodeId);
        if (clickedNode) {
            this.showToast(`Filtré sur: ${clickedNode.label}`, 'info');
        }
    },

    resetScenarioGraph() {
        if (!this.scenarioGraphOriginalData) return;
        this.renderScenarioGraphWithData(
            this.scenarioGraphOriginalData.nodes,
            this.scenarioGraphOriginalData.edges,
            false
        );
        this.showToast('Graphe réinitialisé', 'info');
    },

    updateScenarioGraphResetButton(show) {
        let resetBtn = document.getElementById('btn-reset-scenario-graph');

        if (show) {
            if (!resetBtn) {
                // Créer le bouton s'il n'existe pas
                const headerActions = document.querySelector('#view-scenarios .scenarios-right-column .panel-header-actions');
                if (headerActions) {
                    resetBtn = document.createElement('button');
                    resetBtn.id = 'btn-reset-scenario-graph';
                    resetBtn.className = 'btn btn-sm btn-ghost';
                    resetBtn.setAttribute('data-tooltip', 'Réinitialiser le graphe');
                    resetBtn.innerHTML = '<span class="material-icons">refresh</span>';
                    resetBtn.onclick = () => this.resetScenarioGraph();
                    headerActions.insertBefore(resetBtn, headerActions.firstChild);
                }
            }
            if (resetBtn) resetBtn.style.display = '';
        } else {
            if (resetBtn) resetBtn.style.display = 'none';
        }
    },

    getScenarioNodeColor(node) {
        console.log('getScenarioNodeColor for node:', node.id, 'data:', node.data);
        if (node.data?.scenario_modified === 'true') {
            console.log('Node', node.id, 'is ORANGE (scenario_modified)');
            return '#f97316'; // Orange for modified nodes
        }

        const roleColors = {
            'suspect': '#ef4444',
            'victime': '#3b82f6',
            'temoin': '#22c55e',
            'enqueteur': '#8b5cf6',
            'autre': '#6b7280'
        };

        return roleColors[node.role] || '#6b7280';
    },

    // ============================================
    // Simulate Scenario with AI
    // ============================================
    async simulateScenario(scenarioId) {
        console.log('simulateScenario called with:', scenarioId);
        const scenario = this.scenarios.find(s => s.id === scenarioId);
        if (!scenario) {
            console.error('Scenario not found:', scenarioId);
            return;
        }

        this.setAnalysisContext('hypothesis', `Simulation: ${scenario.name}`, `Scenario What-If`);

        const analysisModal = document.getElementById('analysis-modal');
        const analysisContent = document.getElementById('analysis-content');
        const modalTitle = document.getElementById('analysis-modal-title');
        if (modalTitle) modalTitle.textContent = `Simulation: ${scenario.name}`;

        const noteBtn = document.getElementById('btn-save-to-notebook');
        if (noteBtn) noteBtn.style.display = '';

        analysisContent.innerHTML = `
            <div class="modal-explanation">
                <span class="material-icons">psychology</span>
                <p>Analyse du scenario en cours...</p>
            </div>
            <div id="simulation-content"><span class="streaming-cursor"></span></div>
        `;
        analysisModal.classList.add('active');

        try {
            const result = await this.apiCall('/api/scenario/simulate', 'POST', {
                case_id: this.currentCase.id,
                scenario_id: scenarioId
            });

            document.getElementById('simulation-content').innerHTML =
                this.formatMarkdown ? this.formatMarkdown(result.analysis) : result.analysis;

            // Update local scenario with analysis
            if (result.scenario) {
                const idx = this.scenarios.findIndex(s => s.id === scenarioId);
                if (idx !== -1) {
                    this.scenarios[idx] = result.scenario;
                    if (this.selectedScenario?.id === scenarioId) {
                        this.selectedScenario = result.scenario;
                        this.renderScenarioDetail();
                    }
                }
            }
        } catch (error) {
            document.getElementById('simulation-content').innerHTML =
                `<p class="error">Erreur: ${error.message}</p>`;
        }
    },

    // ============================================
    // Propagate Scenario Implications
    // ============================================
    async propagateScenario(scenarioId) {
        console.log('propagateScenario called with:', scenarioId);
        try {
            const result = await this.apiCall('/api/scenario/propagate', 'POST', {
                case_id: this.currentCase.id,
                scenario_id: scenarioId
            });

            this.showToast('Implications propagees');

            // Update local scenario
            if (result.scenario) {
                const idx = this.scenarios.findIndex(s => s.id === scenarioId);
                if (idx !== -1) {
                    this.scenarios[idx] = result.scenario;
                    if (this.selectedScenario?.id === scenarioId) {
                        this.selectedScenario = result.scenario;
                        this.renderScenarioDetail();
                        this.renderScenarioGraph();
                    }
                }
            }
        } catch (error) {
            console.error('Error propagating scenario:', error);
            this.showToast('Erreur lors de la propagation', 'error');
        }
    },

    // ============================================
    // Compare Scenarios
    // ============================================
    selectScenarioForComparison(scenarioId) {
        console.log('selectScenarioForComparison called with:', scenarioId);
        const scenario = this.scenarios.find(s => s.id === scenarioId);
        if (!scenario) return;

        if (!this.comparisonScenario1) {
            this.comparisonScenario1 = scenario;
            document.querySelector(`[data-scenario-id="${scenarioId}"]`)?.classList.add('selected-for-compare');
            this.showToast('Scenario selectionne. Cliquez sur "Comparer" sur un autre scenario.');
        } else if (this.comparisonScenario1.id === scenarioId) {
            this.comparisonScenario1 = null;
            document.querySelector(`[data-scenario-id="${scenarioId}"]`)?.classList.remove('selected-for-compare');
        } else {
            this.comparisonScenario2 = scenario;
            this.showScenarioComparison();
        }
    },

    async showScenarioComparison() {
        if (!this.comparisonScenario1 || !this.comparisonScenario2) return;

        try {
            const comparison = await this.apiCall('/api/scenario/compare', 'POST', {
                case_id: this.currentCase.id,
                scenario1_id: this.comparisonScenario1.id,
                scenario2_id: this.comparisonScenario2.id
            });

            const content = `
                <div class="modal-explanation">
                    <span class="material-icons">compare_arrows</span>
                    <p>Comparaison cote a cote des deux scenarios pour evaluer leurs forces respectives.</p>
                </div>
                <div class="scenario-comparison">
                    <div class="comparison-column">
                        <h3>${comparison.scenario1_name}</h3>
                        <div class="comparison-score ${this.getPlausibilityClass(this.comparisonScenario1.plausibility_score)}">
                            ${this.comparisonScenario1.plausibility_score}%
                        </div>
                        <div class="comparison-facts">
                            <div class="facts-supporting">
                                <span class="material-icons">check_circle</span>
                                ${this.comparisonScenario1.supporting_facts?.length || 0} faits a l'appui
                            </div>
                            <div class="facts-contradicting">
                                <span class="material-icons">cancel</span>
                                ${this.comparisonScenario1.contradicting_facts?.length || 0} faits contre
                            </div>
                        </div>
                    </div>
                    <div class="comparison-vs">
                        <div class="delta ${comparison.plausibility_delta > 0 ? 'positive' : comparison.plausibility_delta < 0 ? 'negative' : ''}">
                            ${comparison.plausibility_delta > 0 ? '+' : ''}${comparison.plausibility_delta}%
                        </div>
                        VS
                    </div>
                    <div class="comparison-column">
                        <h3>${comparison.scenario2_name}</h3>
                        <div class="comparison-score ${this.getPlausibilityClass(this.comparisonScenario2.plausibility_score)}">
                            ${this.comparisonScenario2.plausibility_score}%
                        </div>
                        <div class="comparison-facts">
                            <div class="facts-supporting">
                                <span class="material-icons">check_circle</span>
                                ${this.comparisonScenario2.supporting_facts?.length || 0} faits a l'appui
                            </div>
                            <div class="facts-contradicting">
                                <span class="material-icons">cancel</span>
                                ${this.comparisonScenario2.contradicting_facts?.length || 0} faits contre
                            </div>
                        </div>
                    </div>
                </div>

                ${comparison.common_facts?.length > 0 ? `
                    <div class="comparison-section">
                        <h4><span class="material-icons">handshake</span> Faits communs</h4>
                        <ul class="common-facts">
                            ${comparison.common_facts.map(f => `<li>${f}</li>`).join('')}
                        </ul>
                    </div>
                ` : ''}

                ${comparison.different_facts?.length > 0 ? `
                    <div class="comparison-section">
                        <h4><span class="material-icons">difference</span> Differences</h4>
                        <table class="differences-table">
                            <thead>
                                <tr><th>Aspect</th><th>Scenario 1</th><th>Scenario 2</th><th>Importance</th></tr>
                            </thead>
                            <tbody>
                                ${comparison.different_facts.map(diff => `
                                    <tr class="${diff.significance}">
                                        <td>${diff.aspect}</td>
                                        <td>${diff.scenario1_value}</td>
                                        <td>${diff.scenario2_value}</td>
                                        <td><span class="significance-badge ${diff.significance}">${diff.significance}</span></td>
                                    </tr>
                                `).join('')}
                            </tbody>
                        </table>
                    </div>
                ` : ''}

                ${comparison.recommendation ? `
                    <div class="comparison-section recommendation">
                        <h4><span class="material-icons">recommend</span> Recommandation</h4>
                        <p>${comparison.recommendation}</p>
                    </div>
                ` : ''}
            `;

            this.showModal('Comparaison des Scenarios', content);

        } catch (error) {
            console.error('Error comparing scenarios:', error);
            this.showToast('Erreur lors de la comparaison', 'error');
        } finally {
            // Reset comparison state
            document.querySelectorAll('.selected-for-compare').forEach(el =>
                el.classList.remove('selected-for-compare'));
            this.comparisonScenario1 = null;
            this.comparisonScenario2 = null;
        }
    },

    // ============================================
    // Delete Scenario
    // ============================================
    async deleteScenario(scenarioId) {
        console.log('deleteScenario called with:', scenarioId);
        if (!confirm('Supprimer ce scenario ?')) return;

        try {
            await this.apiCall(`/api/scenario?case_id=${this.currentCase.id}&scenario_id=${scenarioId}`, 'DELETE');
            this.showToast('Scenario supprime');

            if (this.selectedScenario?.id === scenarioId) {
                this.selectedScenario = null;
            }

            await this.loadScenarios();
            this.renderScenarioDetail();
        } catch (error) {
            console.error('Error deleting scenario:', error);
            this.showToast('Erreur lors de la suppression', 'error');
        }
    },

    // ============================================
    // Helper Methods
    // ============================================
    getPlausibilityClass(score) {
        if (score >= 70) return 'plausibility-high';
        if (score >= 40) return 'plausibility-medium';
        return 'plausibility-low';
    },

    getAssumptionTypeLabel(type) {
        const labels = {
            'guilt': 'Culpabilite',
            'presence': 'Presence',
            'motive': 'Mobile',
            'timeline': 'Chronologie',
            'relation': 'Relation'
        };
        return labels[type] || type;
    },

    getImplicationTypeLabel(type) {
        const labels = {
            'change_role': 'Changement de role',
            'add_relation': 'Nouvelle relation',
            'remove_relation': 'Suppression relation',
            'add_motive': 'Nouveau mobile',
            'timeline_conflict': 'Conflit temporel',
            'relation_review': 'Relation a examiner',
            'presence_verification': 'Verification presence',
            'evidence_review': 'Preuve a reexaminer',
            'temporal_proximity': 'Proximite temporelle',
            'relation_impact': 'Impact relationnel'
        };
        return labels[type] || type;
    },

    // Resolve entity IDs (ent-xxx-xxx) to entity names in a description string
    resolveEntityIds(description) {
        if (!description || !this.currentCase?.entities) return description;

        // Pattern to match entity IDs like 'ent-moreau-002'
        const entityIdPattern = /'(ent-[a-z]+-\d{3})'/gi;

        return description.replace(entityIdPattern, (match, entityId) => {
            const entity = this.currentCase.entities.find(e => e.id === entityId);
            if (entity) {
                return `'${entity.name}'`;
            }
            return match; // Return original if not found
        });
    }
};

// Export for use in main app
if (typeof module !== 'undefined' && module.exports) {
    module.exports = ScenariosModule;
}
