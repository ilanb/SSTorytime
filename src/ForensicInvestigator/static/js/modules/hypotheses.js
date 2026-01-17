// ForensicInvestigator - Module Hypotheses
// Gestion des hypothèses, génération, comparaison

const HypothesesModule = {
    // State
    hypothesisFilters: { status: 'all', confidence: 'all', origin: 'all' },
    selectedHypothesisForCompare: null,

    // ============================================
    // Load Hypotheses
    // ============================================
    async loadHypotheses() {
        if (!this.currentCase) return;

        const container = document.getElementById('hypotheses-list');
        let hypotheses = this.currentCase.hypotheses || [];

        hypotheses = this.filterHypotheses(hypotheses);

        if (hypotheses.length === 0) {
            container.innerHTML = `
                <div class="empty-state">
                    <span class="material-icons empty-state-icon">psychology</span>
                    <p class="empty-state-title">Aucune hypothèse</p>
                    <p class="empty-state-description">Générez des hypothèses avec l'IA ou ajoutez-en manuellement</p>
                </div>
            `;
            return;
        }

        const evidenceMap = {};
        (this.currentCase.evidence || []).forEach(ev => {
            evidenceMap[ev.id] = ev;
        });

        // Get causal chains from N4L parsed data
        const causalChains = this.lastN4LParse?.causal_chains || [];

        container.innerHTML = hypotheses.map(h => {
            const statusClass = this.getHypothesisStatusClass(h.status);
            const statusLabel = this.getHypothesisStatusLabel(h.status);
            const confidenceClass = this.getConfidenceClass(h.confidence_level);
            const originBadge = h.generated_by === 'ai'
                ? '<span class="hypothesis-origin-badge ai"><span class="material-icons">auto_awesome</span>IA</span>'
                : '<span class="hypothesis-origin-badge manual"><span class="material-icons">person</span>Manuel</span>';

            const supportingEvidence = (h.supporting_evidence || []).map(id => evidenceMap[id]).filter(Boolean);
            const contradictingEvidence = (h.contradicting_evidence || []).map(id => evidenceMap[id]).filter(Boolean);

            // Find related causal chains based on hypothesis content
            const relatedChains = this.findRelatedCausalChains(h, causalChains);

            return `
            <div class="hypothesis-card ${statusClass} ${confidenceClass}" data-hypothesis-id="${h.id}" data-status="${h.status}" data-confidence="${h.confidence_level}" data-origin="${h.generated_by || 'user'}">
                <div class="hypothesis-header">
                    <div class="hypothesis-header-left">
                        <span class="hypothesis-title">${h.title}</span>
                        ${originBadge}
                    </div>
                    <div class="hypothesis-status-dropdown">
                        <select class="hypothesis-status-select ${statusClass}" onchange="app.updateHypothesisStatus('${h.id}', this.value)">
                            <option value="en_attente" ${h.status === 'en_attente' ? 'selected' : ''}>En attente</option>
                            <option value="corroboree" ${h.status === 'corroboree' ? 'selected' : ''}>Corroborée</option>
                            <option value="refutee" ${h.status === 'refutee' ? 'selected' : ''}>Réfutée</option>
                            <option value="partielle" ${h.status === 'partielle' ? 'selected' : ''}>Partielle</option>
                        </select>
                    </div>
                </div>
                <p class="hypothesis-description">${h.description}</p>
                <div class="hypothesis-confidence">
                    <div class="confidence-label">Confiance: ${h.confidence_level}%</div>
                    <div class="confidence-bar">
                        <div class="confidence-fill ${confidenceClass}" style="width: ${h.confidence_level}%"></div>
                    </div>
                </div>

                ${supportingEvidence.length > 0 ? `
                    <div class="hypothesis-evidence supporting">
                        <span class="evidence-label"><span class="material-icons">check_circle</span> Preuves à l'appui:</span>
                        <div class="evidence-tags">
                            ${supportingEvidence.map(ev => `<span class="evidence-tag supporting clickable" onclick="app.goToSearchResult('evidence', '${ev.id}')" data-tooltip="Voir cette preuve">${ev.name}</span>`).join('')}
                        </div>
                    </div>
                ` : ''}

                ${contradictingEvidence.length > 0 ? `
                    <div class="hypothesis-evidence contradicting">
                        <span class="evidence-label"><span class="material-icons">cancel</span> Preuves contradictoires:</span>
                        <div class="evidence-tags">
                            ${contradictingEvidence.map(ev => `<span class="evidence-tag contradicting clickable" onclick="app.goToSearchResult('evidence', '${ev.id}')" data-tooltip="Voir cette preuve">${ev.name}</span>`).join('')}
                        </div>
                    </div>
                ` : ''}

                ${h.questions && h.questions.length > 0 ? `
                    <div class="hypothesis-questions" onclick="app.showHypothesisQuestions('${h.id}')" style="cursor: pointer;" data-tooltip="Cliquer pour voir les questions">
                        <span class="material-icons">help_outline</span>
                        <span>${h.questions.length} question(s) à explorer</span>
                    </div>
                ` : ''}

                ${relatedChains.length > 0 ? `
                    <div class="hypothesis-causal-chains">
                        <span class="causal-chains-label">
                            <span class="material-icons">route</span>
                            Chaînes causales liées (${relatedChains.length}):
                        </span>
                        <div class="causal-chains-list">
                            ${relatedChains.map(chain => `
                                <div class="hypothesis-chain-item" onclick="app.showHypothesisCausalChain(${chain.index})" data-tooltip="Cliquer pour visualiser">
                                    <span class="chain-relevance ${chain.relevance > 50 ? 'high' : 'medium'}">${chain.relevance}%</span>
                                    <span class="chain-preview">${chain.preview}</span>
                                </div>
                            `).join('')}
                        </div>
                    </div>
                ` : ''}

                <div class="hypothesis-actions">
                    <button class="btn btn-ghost btn-sm" onclick="app.analyzeHypothesis('${h.id}')" data-tooltip="Analyser avec l'IA">
                        <span class="material-icons">psychology</span>
                    </button>
                    <button class="btn btn-ghost btn-sm" onclick="app.showHypothesisGraph('${h.id}')" data-tooltip="Visualiser sur le graphe">
                        <span class="material-icons">hub</span>
                    </button>
                    <button class="btn btn-ghost btn-sm" onclick="app.manageHypothesisEvidence('${h.id}')" data-tooltip="Gérer les preuves liées">
                        <span class="material-icons">link</span>
                    </button>
                    <button class="btn btn-ghost btn-sm" onclick="app.compareHypothesis('${h.id}')" data-tooltip="Comparer avec d'autres hypothèses">
                        <span class="material-icons">compare_arrows</span>
                    </button>
                    <button class="btn btn-ghost btn-sm" onclick="app.deleteHypothesis('${h.id}')" data-tooltip="Supprimer cette hypothèse">
                        <span class="material-icons">delete</span>
                    </button>
                </div>
            </div>
            `;
        }).join('');
    },

    // ============================================
    // Filter Hypotheses
    // ============================================
    filterHypotheses(hypotheses) {
        return hypotheses.filter(h => {
            if (this.hypothesisFilters.status !== 'all' && h.status !== this.hypothesisFilters.status) return false;
            if (this.hypothesisFilters.confidence !== 'all') {
                const level = h.confidence_level;
                if (this.hypothesisFilters.confidence === 'high' && level < 70) return false;
                if (this.hypothesisFilters.confidence === 'medium' && (level < 40 || level >= 70)) return false;
                if (this.hypothesisFilters.confidence === 'low' && level >= 40) return false;
            }
            if (this.hypothesisFilters.origin !== 'all' && (h.generated_by || 'user') !== this.hypothesisFilters.origin) return false;
            return true;
        });
    },

    toggleHypothesisFilterMenu() {
        const menu = document.getElementById('hypothesis-filter-menu');
        menu.classList.toggle('active');

        // Initialize filter listeners if not already done
        if (!menu._filtersInitialized) {
            menu._filtersInitialized = true;
            const checkboxes = menu.querySelectorAll('input[type="checkbox"]');
            checkboxes.forEach(checkbox => {
                checkbox.addEventListener('change', (e) => {
                    const changed = e.target;
                    const groupName = changed.name;
                    const groupCheckboxes = menu.querySelectorAll(`input[name="${groupName}"]`);

                    if (changed.value === 'all' && changed.checked) {
                        // "Tous" was checked - uncheck all specific options
                        groupCheckboxes.forEach(cb => {
                            if (cb.value !== 'all') cb.checked = false;
                        });
                    } else if (changed.value !== 'all' && changed.checked) {
                        // A specific option was checked - uncheck "Tous"
                        groupCheckboxes.forEach(cb => {
                            if (cb.value === 'all') cb.checked = false;
                        });
                    }

                    // If nothing is checked, re-check "Tous"
                    const anyChecked = Array.from(groupCheckboxes).some(cb => cb.checked);
                    if (!anyChecked) {
                        groupCheckboxes.forEach(cb => {
                            if (cb.value === 'all') cb.checked = true;
                        });
                    }

                    this.applyHypothesisFilter();
                });
            });
        }
    },

    applyHypothesisFilter() {
        const menu = document.getElementById('hypothesis-filter-menu');
        const statusChecked = menu.querySelectorAll('input[name="hyp-status"]:checked');
        const confidenceChecked = menu.querySelectorAll('input[name="hyp-confidence"]:checked');
        const originChecked = menu.querySelectorAll('input[name="hyp-origin"]:checked');

        // Check if "all" is selected or get the first specific value
        const getFilterValue = (checkedItems) => {
            const values = Array.from(checkedItems).map(cb => cb.value);
            if (values.includes('all') || values.length === 0) return 'all';
            return values[0]; // Return first specific selection
        };

        this.hypothesisFilters.status = getFilterValue(statusChecked);
        this.hypothesisFilters.confidence = getFilterValue(confidenceChecked);
        this.hypothesisFilters.origin = getFilterValue(originChecked);

        this.loadHypotheses();
    },

    // ============================================
    // Status Helpers
    // ============================================
    getHypothesisStatusClass(status) {
        const classes = {
            'en_attente': 'status-pending',
            'corroboree': 'status-confirmed',
            'refutee': 'status-refuted',
            'partielle': 'status-partial'
        };
        return classes[status] || 'status-pending';
    },

    getHypothesisStatusLabel(status) {
        const labels = {
            'en_attente': 'En attente',
            'corroboree': 'Corroborée',
            'refutee': 'Réfutée',
            'partielle': 'Partielle'
        };
        return labels[status] || status;
    },

    getConfidenceClass(level) {
        if (level >= 70) return 'confidence-high';
        if (level >= 40) return 'confidence-medium';
        return 'confidence-low';
    },

    async updateHypothesisStatus(hypothesisId, newStatus) {
        try {
            await this.apiCall(`/api/hypotheses/update?case_id=${this.currentCase.id}`, 'PUT', {
                id: hypothesisId,
                status: newStatus
            });
            this.showToast('Statut mis à jour');
        } catch (error) {
            console.error('Error updating hypothesis status:', error);
        }
    },

    // ============================================
    // Add Hypothesis Modal
    // ============================================
    showAddHypothesisModal() {
        if (!this.currentCase) {
            alert('Sélectionnez une affaire d\'abord');
            return;
        }

        const content = `
            <div class="modal-explanation">
                <span class="material-icons">lightbulb</span>
                <p>Créez une nouvelle hypothèse d'investigation. Une hypothèse est une théorie ou piste d'enquête
                que vous souhaitez explorer. Vous pourrez ensuite lui associer des preuves et la faire analyser par l'IA.</p>
            </div>
            <form id="hypothesis-form">
                <div class="form-group">
                    <label class="form-label">Titre</label>
                    <input type="text" class="form-input" id="hypothesis-title" required placeholder="Ex: Mobile financier">
                </div>
                <div class="form-group">
                    <label class="form-label">Niveau de confiance (%)</label>
                    <input type="number" class="form-input" id="hypothesis-confidence" min="0" max="100" value="50">
                </div>
                <div class="form-group">
                    <label class="form-label">Description</label>
                    <textarea class="form-textarea" id="hypothesis-description" required placeholder="Décrivez l'hypothèse..."></textarea>
                </div>
            </form>
        `;

        this.showModal('Ajouter une Hypothèse', content, async () => {
            const hypothesis = {
                title: document.getElementById('hypothesis-title').value,
                confidence_level: parseInt(document.getElementById('hypothesis-confidence').value),
                description: document.getElementById('hypothesis-description').value,
                status: 'en_attente',
                generated_by: 'user'
            };

            if (!hypothesis.title) return;

            try {
                // Utiliser le DataProvider si disponible
                if (typeof DataProvider !== 'undefined' && DataProvider.currentCaseId) {
                    try {
                        await DataProvider.addHypothesis(hypothesis);
                    } catch (dpError) {
                        console.warn('DataProvider.addHypothesis failed, falling back to API:', dpError);
                        await this.apiCall(`/api/hypotheses?case_id=${this.currentCase.id}`, 'POST', hypothesis);
                    }
                } else {
                    await this.apiCall(`/api/hypotheses?case_id=${this.currentCase.id}`, 'POST', hypothesis);
                }
                await this.selectCase(this.currentCase.id);
            } catch (error) {
                console.error('Error adding hypothesis:', error);
            }
        });
    },

    // ============================================
    // Show Hypothesis Questions
    // ============================================
    showHypothesisQuestions(hypothesisId) {
        const hypothesis = this.currentCase?.hypotheses?.find(h => h.id === hypothesisId);
        if (!hypothesis || !hypothesis.questions || hypothesis.questions.length === 0) {
            this.showToast('Aucune question pour cette hypothèse', 'info');
            return;
        }

        const questionsHtml = hypothesis.questions.map((q, i) => `
            <div class="question-item">
                <span class="question-number">${i + 1}</span>
                <span class="question-text">${this.escapeHtml ? this.escapeHtml(q) : q}</span>
            </div>
        `).join('');

        this.showModal(`Questions à explorer: ${hypothesis.title}`, `
            <div class="modal-explanation">
                <span class="material-icons">help_outline</span>
                <p><strong>Questions ouvertes</strong> - Ces questions ont été identifiées pour approfondir l'investigation de cette hypothèse.</p>
            </div>
            <div class="questions-list">
                ${questionsHtml}
            </div>
        `, null, false);
    },

    // ============================================
    // Hypothesis Graph
    // ============================================
    showHypothesisGraph(hypothesisId) {
        const hypothesis = this.currentCase.hypotheses.find(h => h.id === hypothesisId);
        if (!hypothesis) return;

        // Fonction pour normaliser les IDs (tirets et underscores sont équivalents)
        const normalizeId = (id) => id ? id.replace(/-/g, '_') : '';

        // Créer une map des preuves par ID et ID normalisé
        const evidenceMap = {};
        (this.currentCase.evidence || []).forEach(e => {
            evidenceMap[e.id] = e;
            evidenceMap[normalizeId(e.id)] = e;
            // Aussi par nom pour correspondance N4L
            evidenceMap[e.name] = e;
        });

        // Créer une map des entités par ID et nom
        const entityMap = {};
        (this.currentCase.entities || []).forEach(e => {
            entityMap[e.id] = e;
            entityMap[normalizeId(e.id)] = e;
            entityMap[e.name] = e;
        });

        // Fonction pour résoudre les références N4L ($alias.n)
        const resolveN4LRef = (ref) => {
            if (!ref) return null;
            // Si c'est une référence N4L ($alias.n), essayer de la résoudre
            if (ref.startsWith('$') && this.lastN4LParse?.cross_refs) {
                const crossRef = this.lastN4LParse.cross_refs.find(cr =>
                    `$${cr.alias}.${cr.index}` === ref
                );
                if (crossRef?.resolved) {
                    return crossRef.resolved;
                }
            }
            return ref;
        };

        const nodes = [];
        const edges = [];
        const addedNodeIds = new Set();

        // Noeud central: l'hypothèse
        nodes.push({
            id: hypothesis.id,
            label: hypothesis.title,
            color: '#6366f1',
            shape: 'box',
            font: { color: 'white' }
        });
        addedNodeIds.add(hypothesis.id);

        // Ajouter les preuves à l'appui
        (hypothesis.supporting_evidence || []).forEach(evRef => {
            // Résoudre la référence N4L si nécessaire
            const resolvedRef = resolveN4LRef(evRef);

            // Chercher par référence originale, résolue, ID, ID normalisé, ou par nom
            const ev = evidenceMap[evRef] || evidenceMap[resolvedRef] ||
                       evidenceMap[normalizeId(evRef)] || entityMap[evRef] ||
                       entityMap[resolvedRef] || entityMap[normalizeId(evRef)];

            const nodeId = ev?.id || evRef;
            const nodeLabel = ev?.name || resolvedRef || evRef;

            if (!addedNodeIds.has(nodeId)) {
                nodes.push({
                    id: nodeId,
                    label: nodeLabel,
                    color: '#22c55e',
                    shape: 'ellipse'
                });
                addedNodeIds.add(nodeId);
            }
            edges.push({
                from: nodeId,
                to: hypothesis.id,
                label: 'soutient',
                color: '#22c55e'
            });
        });

        // Ajouter les preuves contradictoires
        (hypothesis.contradicting_evidence || []).forEach(evRef => {
            // Résoudre la référence N4L si nécessaire
            const resolvedRef = resolveN4LRef(evRef);

            // Chercher par référence originale, résolue, ID, ID normalisé, ou par nom
            const ev = evidenceMap[evRef] || evidenceMap[resolvedRef] ||
                       evidenceMap[normalizeId(evRef)] || entityMap[evRef] ||
                       entityMap[resolvedRef] || entityMap[normalizeId(evRef)];

            const nodeId = ev?.id || evRef;
            const nodeLabel = ev?.name || resolvedRef || evRef;

            if (!addedNodeIds.has(nodeId)) {
                nodes.push({
                    id: nodeId,
                    label: nodeLabel,
                    color: '#ef4444',
                    shape: 'ellipse'
                });
                addedNodeIds.add(nodeId);
            }
            edges.push({
                from: nodeId,
                to: hypothesis.id,
                label: 'contredit',
                color: '#ef4444'
            });
        });

        const content = `
            <div class="modal-explanation">
                <span class="material-icons">hub</span>
                <p>Ce mini-graphe visualise les relations entre l'hypothèse et ses preuves associées.
                Les <strong style="color: #22c55e;">preuves à l'appui</strong> sont en vert, les <strong style="color: #ef4444;">preuves contradictoires</strong> en rouge.</p>
            </div>
            <div id="hypothesis-mini-graph" style="width: 100%; height: 400px; border: 1px solid var(--border); border-radius: 8px;"></div>
        `;
        this.showModal(`Graphe: ${hypothesis.title}`, content);

        setTimeout(() => {
            const container = document.getElementById('hypothesis-mini-graph');
            if (container && typeof vis !== 'undefined') {
                const data = {
                    nodes: new vis.DataSet(nodes),
                    edges: new vis.DataSet(edges)
                };
                const options = {
                    physics: { enabled: true, stabilization: { iterations: 100 } },
                    edges: { arrows: 'to', smooth: { type: 'curvedCW' } },
                    nodes: { font: { size: 12 } }
                };
                new vis.Network(container, data, options);
            }
        }, 200);
    },

    // ============================================
    // Manage Evidence
    // ============================================
    manageHypothesisEvidence(hypothesisId) {
        const hypothesis = this.currentCase.hypotheses.find(h => h.id === hypothesisId);
        if (!hypothesis) return;

        const allEvidence = this.currentCase.evidence || [];
        const allEntities = this.currentCase.entities || [];
        const supportingIds = hypothesis.supporting_evidence || [];
        const contradictingIds = hypothesis.contradicting_evidence || [];

        // Normalize ID function to handle dash vs underscore differences
        const normalizeId = (id) => id ? id.replace(/-/g, '_') : '';

        // Resolve N4L references like $brouillon.1 to their actual names
        const resolveN4LRef = (ref) => {
            if (!ref) return null;
            if (ref.startsWith('$') && this.lastN4LParse?.cross_refs) {
                const crossRef = this.lastN4LParse.cross_refs.find(cr =>
                    `$${cr.alias}.${cr.index}` === ref
                );
                if (crossRef?.resolved) {
                    return crossRef.resolved;
                }
            }
            return ref;
        };

        // Create evidence and entity maps for lookup
        const evidenceMap = new Map();
        allEvidence.forEach(ev => {
            evidenceMap.set(ev.id, ev);
            evidenceMap.set(normalizeId(ev.id), ev);
            if (ev.name) {
                evidenceMap.set(ev.name, ev);
                evidenceMap.set(normalizeId(ev.name), ev);
            }
        });

        const entityMap = new Map();
        allEntities.forEach(ent => {
            entityMap.set(ent.id, ent);
            entityMap.set(normalizeId(ent.id), ent);
            if (ent.name) {
                entityMap.set(ent.name, ent);
                entityMap.set(normalizeId(ent.name), ent);
            }
        });

        // Helper to check if evidence matches a reference (handles N4L refs)
        const matchesRef = (ev, ref) => {
            if (!ref) return false;

            // Direct ID match
            if (ref === ev.id || normalizeId(ref) === normalizeId(ev.id)) return true;

            // Name match
            if (ev.name && (ref === ev.name || normalizeId(ref) === normalizeId(ev.name))) return true;

            // Resolve N4L reference and check
            const resolved = resolveN4LRef(ref);
            if (resolved && resolved !== ref) {
                if (resolved === ev.id || normalizeId(resolved) === normalizeId(ev.id)) return true;
                if (ev.name && (resolved === ev.name || normalizeId(resolved) === normalizeId(ev.name))) return true;
            }

            return false;
        };

        // Helper to check if evidence is in a list (checks original, normalized, and N4L refs)
        const isInList = (ev, refList) => {
            return refList.some(ref => matchesRef(ev, ref));
        };

        const content = `
            <div class="modal-explanation">
                <span class="material-icons">link</span>
                <p>Associez des preuves à cette hypothèse pour renforcer ou contester sa validité.
                Les <strong style="color: #22c55e;">preuves à l'appui</strong> soutiennent l'hypothèse, tandis que les
                <strong style="color: #ef4444;">preuves contradictoires</strong> la remettent en question.</p>
            </div>
            <div class="evidence-manager">
                <div class="evidence-section">
                    <h4><span class="material-icons" style="color: #22c55e;">check_circle</span> Preuves à l'appui</h4>
                    <div class="evidence-checkboxes">
                        ${allEvidence.map(ev => {
                            const isSupporting = isInList(ev, supportingIds);
                            return `
                            <label class="evidence-checkbox ${isSupporting ? 'selected' : ''}">
                                <input type="checkbox" name="supporting" value="${ev.id}" ${isSupporting ? 'checked' : ''}>
                                ${ev.name}
                            </label>
                        `}).join('')}
                    </div>
                </div>
                <div class="evidence-section">
                    <h4><span class="material-icons" style="color: #ef4444;">cancel</span> Preuves contradictoires</h4>
                    <div class="evidence-checkboxes">
                        ${allEvidence.map(ev => {
                            const isContradicting = isInList(ev, contradictingIds);
                            return `
                            <label class="evidence-checkbox ${isContradicting ? 'selected' : ''}">
                                <input type="checkbox" name="contradicting" value="${ev.id}" ${isContradicting ? 'checked' : ''}>
                                ${ev.name}
                            </label>
                        `}).join('')}
                    </div>
                </div>
            </div>
        `;

        this.showModal('Gérer les preuves liées', content, async () => {
            const supportingChecked = Array.from(document.querySelectorAll('input[name="supporting"]:checked')).map(cb => cb.value);
            const contradictingChecked = Array.from(document.querySelectorAll('input[name="contradicting"]:checked')).map(cb => cb.value);

            try {
                await this.apiCall(`/api/hypotheses/update?case_id=${this.currentCase.id}`, 'PUT', {
                    id: hypothesisId,
                    supporting_evidence: supportingChecked,
                    contradicting_evidence: contradictingChecked
                });
                await this.selectCase(this.currentCase.id);
            } catch (error) {
                console.error('Error updating evidence links:', error);
                alert('Erreur lors de la mise à jour');
            }
        });
    },

    // ============================================
    // Compare Hypotheses
    // ============================================
    compareHypothesis(hypothesisId) {
        const hypothesis = this.currentCase.hypotheses.find(h => h.id === hypothesisId);
        if (!hypothesis) return;

        if (!this.selectedHypothesisForCompare) {
            this.selectedHypothesisForCompare = hypothesis;
            const card = document.querySelector(`[data-hypothesis-id="${hypothesisId}"]`);
            if (card) card.classList.add('selected-for-compare');
            this.showToast('Hypothèse sélectionnée. Cliquez sur "Comparer" sur une autre hypothèse.');
        } else if (this.selectedHypothesisForCompare.id === hypothesisId) {
            this.selectedHypothesisForCompare = null;
            const card = document.querySelector(`[data-hypothesis-id="${hypothesisId}"]`);
            if (card) card.classList.remove('selected-for-compare');
        } else {
            this.showHypothesisComparison(this.selectedHypothesisForCompare, hypothesis);
            document.querySelectorAll('.selected-for-compare').forEach(el => el.classList.remove('selected-for-compare'));
            this.selectedHypothesisForCompare = null;
        }
    },

    showHypothesisComparison(hyp1, hyp2) {
        const content = `
            <div class="modal-explanation">
                <span class="material-icons">compare_arrows</span>
                <p>Comparez côte à côte deux hypothèses pour évaluer leurs forces respectives.</p>
            </div>
            <div class="hypothesis-comparison">
                <div class="comparison-column">
                    <h3>${hyp1.title}</h3>
                    <div class="comparison-status ${this.getHypothesisStatusClass(hyp1.status)}">${this.getHypothesisStatusLabel(hyp1.status)}</div>
                    <div class="comparison-confidence">
                        <div class="confidence-bar"><div class="confidence-fill ${this.getConfidenceClass(hyp1.confidence_level)}" style="width: ${hyp1.confidence_level}%"></div></div>
                        <span>${hyp1.confidence_level}%</span>
                    </div>
                    <p class="comparison-desc">${hyp1.description}</p>
                    <div class="comparison-evidence">
                        <strong>Preuves à l'appui:</strong> ${(hyp1.supporting_evidence || []).length}<br>
                        <strong>Preuves contre:</strong> ${(hyp1.contradicting_evidence || []).length}
                    </div>
                </div>
                <div class="comparison-vs">VS</div>
                <div class="comparison-column">
                    <h3>${hyp2.title}</h3>
                    <div class="comparison-status ${this.getHypothesisStatusClass(hyp2.status)}">${this.getHypothesisStatusLabel(hyp2.status)}</div>
                    <div class="comparison-confidence">
                        <div class="confidence-bar"><div class="confidence-fill ${this.getConfidenceClass(hyp2.confidence_level)}" style="width: ${hyp2.confidence_level}%"></div></div>
                        <span>${hyp2.confidence_level}%</span>
                    </div>
                    <p class="comparison-desc">${hyp2.description}</p>
                    <div class="comparison-evidence">
                        <strong>Preuves à l'appui:</strong> ${(hyp2.supporting_evidence || []).length}<br>
                        <strong>Preuves contre:</strong> ${(hyp2.contradicting_evidence || []).length}
                    </div>
                </div>
            </div>
        `;

        this.showModal('Comparaison des Hypothèses', content);
    },

    // ============================================
    // Generate Hypotheses (avec streaming)
    // ============================================
    async generateHypotheses() {
        if (!this.currentCase) {
            this.showToast('Sélectionnez une affaire d\'abord', 'warning');
            return;
        }

        this.setAnalysisContext('hypothesis', `Hypothèses générées - ${this.currentCase.name}`, 'Génération automatique');

        const analysisContent = document.getElementById('analysis-content');
        const analysisModal = document.getElementById('analysis-modal');
        const modalTitle = document.getElementById('analysis-modal-title');
        if (modalTitle) modalTitle.textContent = 'Génération d\'hypothèses';

        const noteBtn = document.getElementById('btn-save-to-notebook');
        if (noteBtn) noteBtn.style.display = '';

        analysisContent.innerHTML = '<span class="streaming-cursor">▊</span>';
        analysisModal.classList.add('active');

        await this.streamAIResponse(
            '/api/hypotheses/generate/stream',
            { case_id: this.currentCase.id },
            analysisContent
        );
    },

    // ============================================
    // Generate Questions (avec streaming)
    // ============================================
    async generateQuestions() {
        if (!this.currentCase) {
            this.showToast('Sélectionnez une affaire d\'abord', 'warning');
            return;
        }

        this.setAnalysisContext('question', `Questions d'investigation - ${this.currentCase.name}`, 'Génération automatique');

        const analysisContent = document.getElementById('analysis-content');
        const analysisModal = document.getElementById('analysis-modal');
        const modalTitle = document.getElementById('analysis-modal-title');
        if (modalTitle) modalTitle.textContent = 'Questions d\'Investigation';

        const noteBtn = document.getElementById('btn-save-to-notebook');
        if (noteBtn) noteBtn.style.display = '';

        analysisContent.innerHTML = '<span class="streaming-cursor">▊</span>';
        analysisModal.classList.add('active');

        await this.streamAIResponse(
            '/api/questions/generate/stream',
            { case_id: this.currentCase.id },
            analysisContent
        );
    },

    // ============================================
    // Delete Hypothesis
    // ============================================
    async deleteHypothesis(hypothesisId) {
        if (!confirm('Supprimer cette hypothèse ?')) return;

        // Normaliser l'ID: convertir les underscores en tirets pour l'API backend
        const normalizedId = hypothesisId ? hypothesisId.replace(/_/g, '-') : hypothesisId;

        try {
            // Utiliser le DataProvider si disponible
            if (typeof DataProvider !== 'undefined' && DataProvider.currentCaseId) {
                try {
                    await DataProvider.deleteHypothesis(hypothesisId);
                } catch (dpError) {
                    console.warn('DataProvider.deleteHypothesis failed, falling back to API:', dpError);
                    await this.apiCall(`/api/hypotheses/delete?case_id=${this.currentCase.id}&hypothesis_id=${normalizedId}`, 'DELETE');
                }
            } else {
                await this.apiCall(`/api/hypotheses/delete?case_id=${this.currentCase.id}&hypothesis_id=${normalizedId}`, 'DELETE');
            }
            await this.selectCase(this.currentCase.id);
            this.showToast('Hypothèse supprimée');
        } catch (error) {
            console.error('Error deleting hypothesis:', error);
        }
    },

    // ============================================
    // Analyze Hypothesis with AI
    // ============================================
    // ============================================
    // Edit Hypothesis
    // ============================================
    editHypothesis(hypothesisId) {
        const hypothesis = this.currentCase.hypotheses.find(h => h.id === hypothesisId);
        if (!hypothesis) return;

        const content = `
            <div class="modal-explanation">
                <span class="material-icons">info</span>
                <p>Modifiez les propriétés de cette hypothèse. Le statut reflète l'état actuel de validation,
                et le niveau de confiance indique la probabilité estimée que cette hypothèse soit correcte.</p>
            </div>
            <form id="edit-hypothesis-form">
                <div class="form-group">
                    <label class="form-label">Titre</label>
                    <input type="text" class="form-input" id="edit-hyp-title" value="${hypothesis.title}" required>
                </div>
                <div class="form-group">
                    <label class="form-label">Statut</label>
                    <select class="form-input" id="edit-hyp-status">
                        <option value="en_attente" ${hypothesis.status === 'en_attente' ? 'selected' : ''}>En attente</option>
                        <option value="corroboree" ${hypothesis.status === 'corroboree' ? 'selected' : ''}>Corroborée</option>
                        <option value="refutee" ${hypothesis.status === 'refutee' ? 'selected' : ''}>Réfutée</option>
                        <option value="partielle" ${hypothesis.status === 'partielle' ? 'selected' : ''}>Partielle</option>
                    </select>
                </div>
                <div class="form-group">
                    <label class="form-label">Niveau de confiance (%)</label>
                    <input type="range" class="form-range" id="edit-hyp-confidence" min="0" max="100" value="${hypothesis.confidence_level}">
                    <span id="confidence-value">${hypothesis.confidence_level}%</span>
                </div>
                <div class="form-group">
                    <label class="form-label">Description</label>
                    <textarea class="form-textarea" id="edit-hyp-description" required>${hypothesis.description}</textarea>
                </div>
                <div class="form-group">
                    <label class="form-label">Questions (une par ligne)</label>
                    <textarea class="form-textarea" id="edit-hyp-questions">${(hypothesis.questions || []).join('\n')}</textarea>
                </div>
            </form>
        `;

        this.showModal('Modifier l\'Hypothèse', content, async () => {
            const updatedHypothesis = {
                id: hypothesisId,
                title: document.getElementById('edit-hyp-title').value,
                status: document.getElementById('edit-hyp-status').value,
                confidence_level: parseInt(document.getElementById('edit-hyp-confidence').value),
                description: document.getElementById('edit-hyp-description').value,
                questions: document.getElementById('edit-hyp-questions').value.split('\n').filter(q => q.trim())
            };

            try {
                await this.apiCall(`/api/hypotheses/update?case_id=${this.currentCase.id}`, 'PUT', updatedHypothesis);
                await this.selectCase(this.currentCase.id);
            } catch (error) {
                console.error('Error updating hypothesis:', error);
                alert('Erreur lors de la mise à jour');
            }
        });

        // Event listener pour le slider de confiance
        setTimeout(() => {
            const slider = document.getElementById('edit-hyp-confidence');
            const valueSpan = document.getElementById('confidence-value');
            if (slider && valueSpan) {
                slider.addEventListener('input', () => {
                    valueSpan.textContent = slider.value + '%';
                });
            }
        }, 100);
    },

    async analyzeHypothesis(hypothesisId) {
        const hypothesis = this.currentCase.hypotheses.find(h => h.id === hypothesisId);
        if (!hypothesis) return;

        // Normaliser l'ID: convertir les underscores en tirets pour l'API backend
        const normalizedHypothesisId = hypothesisId ? hypothesisId.replace(/_/g, '-') : hypothesisId;

        const container = document.getElementById('hypotheses-list');
        const card = container.querySelector(`[data-hypothesis-id="${hypothesisId}"]`);
        if (card) {
            card.classList.add('analyzing');
        }

        // Build evidence context
        const supportingEvidence = (hypothesis.supporting_evidence || [])
            .map(id => this.currentCase.evidence.find(e => e.id === id))
            .filter(Boolean)
            .map(e => `- ${e.name}: ${e.description}`);

        const contradictingEvidence = (hypothesis.contradicting_evidence || [])
            .map(id => this.currentCase.evidence.find(e => e.id === id))
            .filter(Boolean)
            .map(e => `- ${e.name}: ${e.description}`);

        // Set context for notebook
        this.setAnalysisContext('hypothesis', `Analyse: ${hypothesis.title}`, `Hypothèse: ${hypothesis.title} (confiance: ${hypothesis.confidence_level}%)`);

        // Display modal
        const analysisModal = document.getElementById('analysis-modal');
        const analysisContent = document.getElementById('analysis-content');
        const modalTitle = document.getElementById('analysis-modal-title');
        if (modalTitle) modalTitle.textContent = `Analyse: ${hypothesis.title}`;

        const noteBtn = document.getElementById('btn-save-to-notebook');
        if (noteBtn) noteBtn.style.display = '';

        analysisContent.innerHTML = `
            <div class="modal-explanation">
                <span class="material-icons">psychology</span>
                <p>Analyse de l'hypothèse en cours...</p>
            </div>
            <div id="hypothesis-analysis-content"><span class="streaming-cursor">▊</span></div>
        `;
        analysisModal.classList.add('active');

        // Stream analysis
        const contentDiv = document.getElementById('hypothesis-analysis-content');
        try {
            await this.streamAIResponse(
                '/api/hypotheses/analyze/stream',
                {
                    case_id: this.currentCase.id,
                    hypothesis_id: normalizedHypothesisId,
                    hypothesis_title: hypothesis.title,
                    hypothesis_description: hypothesis.description,
                    confidence: hypothesis.confidence_level,
                    supporting_evidence: supportingEvidence,
                    contradicting_evidence: contradictingEvidence
                },
                contentDiv
            );
        } catch (error) {
            contentDiv.innerHTML = `<p class="error">Erreur: ${error.message}</p>`;
        } finally {
            if (card) card.classList.remove('analyzing');
        }
    },

    // ============================================
    // Causal Chains Integration (N4L Advanced Feature)
    // ============================================
    findRelatedCausalChains(hypothesis, causalChains) {
        if (!hypothesis || !causalChains || causalChains.length === 0) return [];

        const hypothesisText = `${hypothesis.title || ''} ${hypothesis.description || ''}`.toLowerCase();
        const relatedChains = [];

        causalChains.forEach((chain, index) => {
            let relevanceScore = 0;
            const steps = chain.steps || [];
            const matchedSteps = [];

            // Check each step in the chain for relevance to hypothesis
            steps.forEach(step => {
                const stepText = (step.entity || step.label || '').toLowerCase();
                if (stepText && hypothesisText.includes(stepText)) {
                    relevanceScore += 30;
                    matchedSteps.push(stepText);
                }
            });

            // Check chain context relevance
            const context = (chain.context || '').toLowerCase();
            if (context && hypothesisText.includes(context)) {
                relevanceScore += 20;
            }

            // Bonus for chains with mobile/motive keywords
            const mobileKeywords = ['mobile', 'motive', 'héritage', 'argent', 'vengeance', 'jalousie'];
            if (mobileKeywords.some(kw => hypothesisText.includes(kw) && steps.some(s => (s.entity || '').toLowerCase().includes(kw)))) {
                relevanceScore += 25;
            }

            if (relevanceScore > 0) {
                // Create preview of chain
                const preview = steps.slice(0, 3).map(s => s.entity || s.label || '?').join(' → ') +
                    (steps.length > 3 ? ' → ...' : '');

                relatedChains.push({
                    index,
                    chain,
                    relevance: Math.min(relevanceScore, 100),
                    matchedSteps,
                    preview,
                    context: chain.context || 'Général'
                });
            }
        });

        // Sort by relevance and return top 3
        return relatedChains.sort((a, b) => b.relevance - a.relevance).slice(0, 3);
    },

    showHypothesisCausalChain(chainIndex) {
        const causalChains = this.lastN4LParse?.causal_chains || [];
        const chain = causalChains[chainIndex];
        if (!chain) return;

        // Create modal to display causal chain
        const steps = chain.steps || [];
        const modalHtml = `
            <div class="hypothesis-chain-modal">
                <div class="chain-modal-header">
                    <span class="material-icons">route</span>
                    <h3>Chaîne Causale: ${chain.context || 'Chaîne ' + (chainIndex + 1)}</h3>
                </div>
                <div class="chain-modal-content">
                    <div class="chain-visualization">
                        ${steps.map((step, i) => `
                            <div class="chain-step-card">
                                <span class="step-number">${i + 1}</span>
                                <span class="step-entity">${step.entity || step.label || '?'}</span>
                                ${step.relation ? `<span class="step-relation">${step.relation}</span>` : ''}
                            </div>
                            ${i < steps.length - 1 ? '<div class="chain-connector"><span class="material-icons">arrow_downward</span></div>' : ''}
                        `).join('')}
                    </div>
                    <div class="chain-actions">
                        <button class="btn btn-primary" onclick="app.switchView('dashboard'); app.closeModal(); setTimeout(() => app.highlightCausalChain && app.highlightCausalChain(${chainIndex}), 300);">
                            <span class="material-icons">visibility</span>
                            Voir sur le graphe
                        </button>
                    </div>
                </div>
            </div>
        `;

        this.showModal('Chaîne Causale', modalHtml, { width: '500px' });
    }
};

// Export for use in main app
if (typeof module !== 'undefined' && module.exports) {
    module.exports = HypothesesModule;
}
