// ForensicInvestigator - Module Anomalies
// Detection d'Anomalies

const AnomaliesModule = {
    // State
    anomalies: [],
    alerts: [],
    statistics: null,
    config: null,
    selectedAnomaly: null,
    filterType: 'all',
    filterSeverity: 'all',

    // ============================================
    // Load Anomalies
    // ============================================
    async loadAnomalies() {
        if (!this.currentCase) return;

        try {
            const [anomalies, stats, config, alerts] = await Promise.all([
                this.apiCall(`/api/anomalies?case_id=${this.currentCase.id}`),
                this.apiCall(`/api/anomaly/statistics?case_id=${this.currentCase.id}`),
                this.apiCall(`/api/anomaly/config?case_id=${this.currentCase.id}`),
                this.apiCall(`/api/anomaly/alerts?case_id=${this.currentCase.id}&unread_only=true`)
            ]);

            this.anomalies = anomalies || [];
            this.statistics = stats;
            this.config = config;
            this.alerts = alerts || [];

            this.renderAnomaliesView();
            this.updateAnomalyBadge();
        } catch (error) {
            console.error('Error loading anomalies:', error);
            this.anomalies = [];
            this.renderAnomaliesView();
        }
    },

    renderAnomaliesView() {
        this.renderAnomaliesStats();
        this.renderAnomaliesList();
        this.renderAlertsBanner();
    },

    // ============================================
    // Render Statistics Dashboard
    // ============================================
    renderAnomaliesStats() {
        const container = document.getElementById('anomalies-stats');
        if (!container || !this.statistics) return;

        const stats = this.statistics;
        const criticalCount = stats.by_severity?.critical || 0;
        const highCount = stats.by_severity?.high || 0;
        const mediumCount = stats.by_severity?.medium || 0;
        const lowCount = stats.by_severity?.low || 0;

        container.innerHTML = `
            <div class="anomaly-stats-dashboard">
                <div class="stats-main-row">
                    <div class="stat-box total">
                        <div class="stat-icon"><span class="material-icons">warning_amber</span></div>
                        <div class="stat-info">
                            <div class="stat-number">${stats.total_detected || 0}</div>
                            <div class="stat-label">Total détecté</div>
                        </div>
                    </div>
                    <div class="stat-box pending">
                        <div class="stat-icon"><span class="material-icons">hourglass_empty</span></div>
                        <div class="stat-info">
                            <div class="stat-number">${stats.pending || 0}</div>
                            <div class="stat-label">En attente</div>
                        </div>
                    </div>
                    <div class="stat-box acknowledged">
                        <div class="stat-icon"><span class="material-icons">task_alt</span></div>
                        <div class="stat-info">
                            <div class="stat-number">${stats.acknowledged || 0}</div>
                            <div class="stat-label">Acquittées</div>
                        </div>
                    </div>
                </div>
                <div class="stats-severity-row">
                    <div class="severity-item critical">
                        <span class="severity-dot"></span>
                        <span class="severity-count">${criticalCount}</span>
                        <span class="severity-label">Critiques</span>
                    </div>
                    <div class="severity-item high">
                        <span class="severity-dot"></span>
                        <span class="severity-count">${highCount}</span>
                        <span class="severity-label">Élevées</span>
                    </div>
                    <div class="severity-item medium">
                        <span class="severity-dot"></span>
                        <span class="severity-count">${mediumCount}</span>
                        <span class="severity-label">Moyennes</span>
                    </div>
                    <div class="severity-item low">
                        <span class="severity-dot"></span>
                        <span class="severity-count">${lowCount}</span>
                        <span class="severity-label">Faibles</span>
                    </div>
                </div>
                <div class="stats-confidence-row">
                    <div class="confidence-meter">
                        <span class="material-icons">speed</span>
                        <span class="confidence-label">Confiance moyenne</span>
                        <div class="confidence-bar">
                            <div class="confidence-fill" style="width: ${Math.round(stats.avg_confidence || 0)}%"></div>
                        </div>
                        <span class="confidence-value">${Math.round(stats.avg_confidence || 0)}%</span>
                    </div>
                </div>
            </div>
        `;
    },

    // ============================================
    // Render Anomalies List
    // ============================================
    renderAnomaliesList() {
        const container = document.getElementById('anomalies-list');
        if (!container) return;

        let filteredAnomalies = this.anomalies;

        // Apply filters
        if (this.filterType !== 'all') {
            filteredAnomalies = filteredAnomalies.filter(a => a.type === this.filterType);
        }
        if (this.filterSeverity !== 'all') {
            filteredAnomalies = filteredAnomalies.filter(a => a.severity === this.filterSeverity);
        }

        if (filteredAnomalies.length === 0) {
            container.innerHTML = `
                <div class="empty-state">
                    <span class="material-icons empty-state-icon">check_circle</span>
                    <p class="empty-state-title">Aucune anomalie</p>
                    <p class="empty-state-description">Lancez une detection pour identifier des comportements suspects</p>
                </div>
            `;
            return;
        }

        container.innerHTML = filteredAnomalies.map(anomaly => {
            const severityClass = `severity-${anomaly.severity}`;
            const typeIcon = this.getAnomalyTypeIcon(anomaly.type);
            const isNew = anomaly.is_new ? 'new' : '';
            const isAcknowledged = anomaly.is_acknowledged ? 'acknowledged' : '';

            // Extract advanced detection info
            const details = anomaly.details || {};
            const detectionMethod = details.detection_method || '';
            const zScore = details.z_score;
            const bayesianAdjust = details.bayesian_adjustment;
            const methodBadge = this.getDetectionMethodBadge(detectionMethod);

            return `
                <div class="anomaly-card ${severityClass} ${isNew} ${isAcknowledged}"
                     data-anomaly-id="${anomaly.id}"
                     onclick="app.selectAnomaly('${anomaly.id}')">
                    <div class="anomaly-header">
                        <div class="anomaly-header-left">
                            <span class="material-icons anomaly-type-icon">${typeIcon}</span>
                            <span class="anomaly-type-badge ${anomaly.type}">${this.getAnomalyTypeLabel(anomaly.type)}</span>
                            ${anomaly.is_new ? '<span class="new-badge">NEW</span>' : ''}
                        </div>
                        <div class="anomaly-severity-badge ${severityClass}">
                            ${this.getSeverityLabel(anomaly.severity)}
                        </div>
                    </div>
                    <h4 class="anomaly-title">${anomaly.title}</h4>
                    <p class="anomaly-description">${anomaly.description}</p>
                    <div class="anomaly-meta">
                        <span class="anomaly-confidence" data-tooltip="Confiance bayésienne${bayesianAdjust ? ` (${bayesianAdjust > 0 ? '+' : ''}${bayesianAdjust}%)` : ''}">
                            <span class="material-icons">insights</span>
                            ${anomaly.confidence}%
                            ${bayesianAdjust ? `<span class="bayesian-indicator ${bayesianAdjust > 0 ? 'positive' : 'negative'}">${bayesianAdjust > 0 ? '↑' : '↓'}</span>` : ''}
                        </span>
                        ${zScore !== undefined ? `
                            <span class="anomaly-zscore" data-tooltip="Z-score statistique">
                                <span class="material-icons">analytics</span>
                                Z: ${typeof zScore === 'number' ? zScore.toFixed(2) : zScore}
                            </span>
                        ` : ''}
                        ${methodBadge ? `<span class="detection-method-badge" data-tooltip="${this.getMethodTooltip(detectionMethod)}">${methodBadge}</span>` : ''}
                    </div>
                    <div class="anomaly-actions">
                        <button class="btn btn-ghost btn-sm" onclick="event.stopPropagation(); app.explainAnomaly('${anomaly.id}')" data-tooltip="Expliquer avec l'IA">
                            <span class="material-icons">psychology</span>
                        </button>
                        ${!anomaly.is_acknowledged ? `
                            <button class="btn btn-ghost btn-sm" onclick="event.stopPropagation(); app.acknowledgeAnomaly('${anomaly.id}')" data-tooltip="Acquitter">
                                <span class="material-icons">check</span>
                            </button>
                        ` : ''}
                        <button class="btn btn-ghost btn-sm" onclick="event.stopPropagation(); app.showAnomalyDetails('${anomaly.id}')" data-tooltip="Details">
                            <span class="material-icons">info</span>
                        </button>
                    </div>
                </div>
            `;
        }).join('');
    },

    // ============================================
    // Render Alerts Banner
    // ============================================
    renderAlertsBanner() {
        const container = document.getElementById('anomaly-alerts-banner');
        if (!container) return;

        if (this.alerts.length === 0) {
            container.style.display = 'none';
            return;
        }

        // Get the most severe alert
        const latestAlert = this.alerts[0];
        const severityClass = latestAlert?.severity || 'high';
        const alertMessage = latestAlert?.message || '';

        // Truncate message if too long
        const shortMessage = alertMessage.length > 80 ? alertMessage.substring(0, 80) + '...' : alertMessage;

        container.style.display = 'block';
        container.innerHTML = `
            <div class="alerts-banner-new ${severityClass}">
                <div class="alert-indicator">
                    <span class="material-icons pulse">notifications_active</span>
                </div>
                <div class="alert-main">
                    <div class="alert-header">
                        <span class="alert-badge">${this.alerts.length}</span>
                        <span class="alert-title">${this.alerts.length} alerte${this.alerts.length > 1 ? 's' : ''} active${this.alerts.length > 1 ? 's' : ''}</span>
                    </div>
                    <div class="alert-preview">
                        <span class="severity-tag ${severityClass}">${this.getSeverityLabel(severityClass)}</span>
                        <span class="alert-text">${shortMessage}</span>
                    </div>
                </div>
                <div class="alert-actions-bar">
                    <button class="btn btn-sm btn-outline" onclick="app.showAllAlerts()">
                        <span class="material-icons">visibility</span>
                        Voir tout
                    </button>
                    <button class="btn btn-sm btn-ghost" onclick="app.dismissAlerts()" title="Fermer">
                        <span class="material-icons">close</span>
                    </button>
                </div>
            </div>
        `;
    },

    // ============================================
    // Detect Anomalies
    // ============================================
    async detectAnomalies() {
        if (!this.currentCase) {
            this.showToast('Selectionnez une affaire d\'abord', 'warning');
            return;
        }

        const detectBtn = document.getElementById('btn-detect-anomalies');
        if (detectBtn) {
            detectBtn.disabled = true;
            detectBtn.innerHTML = '<span class="material-icons rotating">sync</span> Detection...';
        }

        try {
            const result = await this.apiCall('/api/anomaly/detect', 'POST', {
                case_id: this.currentCase.id
            });

            this.showToast(result.summary || `${result.total_anomalies} anomalies detectees`);
            await this.loadAnomalies();

            // Show summary modal if new anomalies found
            if (result.new_anomalies > 0) {
                this.showDetectionSummary(result);
            }
        } catch (error) {
            console.error('Error detecting anomalies:', error);
            this.showToast('Erreur lors de la detection', 'error');
        } finally {
            if (detectBtn) {
                detectBtn.disabled = false;
                detectBtn.innerHTML = '<span class="material-icons">radar</span> Detecter';
            }
        }
    },

    showDetectionSummary(result) {
        // Grouper les anomalies par type
        const groupedByType = {};
        const groupedBySeverity = { critical: [], high: [], medium: [], low: [] };
        const groupedByMethod = {};

        // Statistics for advanced algorithms
        let bayesianAdjustedCount = 0;
        let highZScoreCount = 0;
        let avgZScore = 0;
        let zScoreSum = 0;
        let zScoreCount = 0;

        (result.anomalies || []).forEach(a => {
            // Par type
            if (!groupedByType[a.type]) {
                groupedByType[a.type] = [];
            }
            groupedByType[a.type].push(a);

            // Par severite
            if (groupedBySeverity[a.severity]) {
                groupedBySeverity[a.severity].push(a);
            }

            // Par méthode de détection
            const method = a.details?.detection_method || 'heuristic';
            if (!groupedByMethod[method]) {
                groupedByMethod[method] = [];
            }
            groupedByMethod[method].push(a);

            // Stats avancées
            if (a.details?.bayesian_adjustment) {
                bayesianAdjustedCount++;
            }
            if (a.details?.z_score !== undefined) {
                const zScore = Math.abs(a.details.z_score);
                zScoreSum += zScore;
                zScoreCount++;
                if (zScore > 3) {
                    highZScoreCount++;
                }
            }
        });

        if (zScoreCount > 0) {
            avgZScore = zScoreSum / zScoreCount;
        }

        const typeLabels = {
            'timeline': 'Chronologie',
            'evidence': 'Preuves',
            'relationship': 'Relations',
            'alibi': 'Alibis',
            'financial': 'Finances',
            'behavioral': 'Comportement',
            'communication': 'Communication',
            'behavior': 'Comportement',
            'location': 'Localisation',
            'relation': 'Relations',
            'pattern': 'Patterns'
        };

        // Generer le contenu par type
        const typeBreakdownHtml = Object.entries(groupedByType).map(([type, anomalies]) => `
            <div class="detection-type-group">
                <div class="type-header">
                    <span class="material-icons">${this.getAnomalyTypeIcon(type)}</span>
                    <span class="type-name">${typeLabels[type] || type}</span>
                    <span class="type-count">${anomalies.length}</span>
                </div>
                <div class="type-items">
                    ${anomalies.slice(0, 3).map(a => `
                        <div class="type-item severity-${a.severity}">
                            <span class="severity-dot"></span>
                            <span class="item-title">${a.title}</span>
                            ${a.details?.z_score !== undefined ? `<span class="item-zscore">Z:${Math.abs(a.details.z_score).toFixed(1)}</span>` : ''}
                        </div>
                    `).join('')}
                    ${anomalies.length > 3 ? `<div class="type-item more">+${anomalies.length - 3} autres</div>` : ''}
                </div>
            </div>
        `).join('');

        // Générer le contenu par méthode de détection
        const methodBreakdownHtml = Object.entries(groupedByMethod).map(([method, anomalies]) => `
            <div class="detection-method-item">
                <span class="method-badge-summary">${this.getDetectionMethodBadge(method)}</span>
                <span class="method-count">${anomalies.length}</span>
            </div>
        `).join('');

        const content = `
            <div class="detection-summary-enhanced">
                <div class="summary-header">
                    <div class="summary-icon ${result.critical_count > 0 ? 'critical' : result.high_count > 0 ? 'warning' : 'success'}">
                        <span class="material-icons">${result.critical_count > 0 ? 'warning' : result.total_anomalies > 0 ? 'radar' : 'check_circle'}</span>
                    </div>
                    <div class="summary-title">
                        <h3>${result.total_anomalies} anomalie${result.total_anomalies > 1 ? 's' : ''} détectée${result.total_anomalies > 1 ? 's' : ''}</h3>
                        <p>${result.new_anomalies > 0 ? `${result.new_anomalies} nouvelle${result.new_anomalies > 1 ? 's' : ''}` : 'Aucune nouvelle anomalie'}</p>
                    </div>
                </div>

                <div class="severity-breakdown">
                    <div class="severity-bar">
                        ${result.critical_count > 0 ? `<div class="severity-segment critical" style="flex: ${result.critical_count}" data-tooltip="${result.critical_count} critique${result.critical_count > 1 ? 's' : ''}"></div>` : ''}
                        ${result.high_count > 0 ? `<div class="severity-segment high" style="flex: ${result.high_count}" data-tooltip="${result.high_count} élevée${result.high_count > 1 ? 's' : ''}"></div>` : ''}
                        ${(result.total_anomalies - result.critical_count - result.high_count) > 0 ? `<div class="severity-segment other" style="flex: ${result.total_anomalies - result.critical_count - result.high_count}" data-tooltip="${result.total_anomalies - result.critical_count - result.high_count} autre${(result.total_anomalies - result.critical_count - result.high_count) > 1 ? 's' : ''}"></div>` : ''}
                    </div>
                    <div class="severity-legend">
                        ${result.critical_count > 0 ? `<span class="legend-item critical"><span class="dot"></span>${result.critical_count} Critique${result.critical_count > 1 ? 's' : ''}</span>` : ''}
                        ${result.high_count > 0 ? `<span class="legend-item high"><span class="dot"></span>${result.high_count} Élevée${result.high_count > 1 ? 's' : ''}</span>` : ''}
                        ${(result.total_anomalies - result.critical_count - result.high_count) > 0 ? `<span class="legend-item other"><span class="dot"></span>${result.total_anomalies - result.critical_count - result.high_count} Autre${(result.total_anomalies - result.critical_count - result.high_count) > 1 ? 's' : ''}</span>` : ''}
                    </div>
                </div>

                ${zScoreCount > 0 || bayesianAdjustedCount > 0 ? `
                    <div class="algorithm-stats">
                        <h4><span class="material-icons">analytics</span> Analyse Statistique</h4>
                        <div class="stats-grid-summary">
                            ${zScoreCount > 0 ? `
                                <div class="stat-item">
                                    <span class="stat-value">${avgZScore.toFixed(2)}</span>
                                    <span class="stat-label">Z-Score Moyen</span>
                                </div>
                                <div class="stat-item ${highZScoreCount > 0 ? 'highlight' : ''}">
                                    <span class="stat-value">${highZScoreCount}</span>
                                    <span class="stat-label">Z-Score > 3σ</span>
                                </div>
                            ` : ''}
                            ${bayesianAdjustedCount > 0 ? `
                                <div class="stat-item">
                                    <span class="stat-value">${bayesianAdjustedCount}</span>
                                    <span class="stat-label">Ajust. Bayésien</span>
                                </div>
                            ` : ''}
                        </div>
                    </div>
                ` : ''}

                ${Object.keys(groupedByMethod).length > 1 ? `
                    <div class="detection-methods-summary">
                        <h4><span class="material-icons">science</span> Méthodes de Détection</h4>
                        <div class="methods-grid">
                            ${methodBreakdownHtml}
                        </div>
                    </div>
                ` : ''}

                ${Object.keys(groupedByType).length > 0 ? `
                    <div class="detection-types">
                        <h4>Par catégorie</h4>
                        ${typeBreakdownHtml}
                    </div>
                ` : ''}

                ${result.critical_count > 0 || result.high_count > 0 ? `
                    <div class="detection-alert">
                        <span class="material-icons">priority_high</span>
                        <p>Des anomalies importantes nécessitent votre attention. Consultez le détail pour investiguer.</p>
                    </div>
                ` : ''}
            </div>
        `;

        this.showModal('Analyse Terminée', content);
    },

    // ============================================
    // Select Anomaly
    // ============================================
    selectAnomaly(anomalyId) {
        this.selectedAnomaly = this.anomalies.find(a => a.id === anomalyId);

        document.querySelectorAll('.anomaly-card').forEach(card => {
            card.classList.toggle('selected', card.dataset.anomalyId === anomalyId);
        });

        this.renderAnomalyDetail();
    },

    renderAnomalyDetail() {
        const container = document.getElementById('anomaly-detail');
        if (!container) return;

        if (!this.selectedAnomaly) {
            container.innerHTML = `
                <div class="empty-state">
                    <span class="material-icons empty-state-icon">touch_app</span>
                    <p class="empty-state-description">Selectionnez une anomalie pour voir les details</p>
                </div>
            `;
            return;
        }

        const anomaly = this.selectedAnomaly;
        const severityClass = `severity-${anomaly.severity}`;
        const details = anomaly.details || {};

        // Extract advanced detection info
        const detectionMethod = details.detection_method || '';
        const zScore = details.z_score;
        const bayesianAdjustment = details.bayesian_adjustment;

        // Categorize details for better display
        const statisticalDetails = {};
        const contextDetails = {};
        const otherDetails = {};

        const statisticalKeys = ['z_score', 'modified_z_score', 'standard_z_score', 'mean', 'std_dev', 'median', 'mad', 'variance_ratio', 'centrality_score', 'betweenness', 'pagerank', 'degree'];
        const contextKeys = ['detection_method', 'original_confidence', 'bayesian_confidence', 'bayesian_adjustment', 'risk_score', 'pattern_type'];

        for (const [key, value] of Object.entries(details)) {
            if (statisticalKeys.some(k => key.includes(k))) {
                statisticalDetails[key] = value;
            } else if (contextKeys.includes(key)) {
                contextDetails[key] = value;
            } else {
                otherDetails[key] = value;
            }
        }

        // Format technical details nicely
        const formatDetails = (detailsObj) => {
            if (!detailsObj || Object.keys(detailsObj).length === 0) return '';
            const items = [];
            for (const [key, value] of Object.entries(detailsObj)) {
                const label = key.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase());
                let displayValue = value;
                if (typeof value === 'number') {
                    displayValue = Number.isInteger(value) ? value : value.toFixed(3);
                } else if (typeof value === 'boolean') {
                    displayValue = value ? 'Oui' : 'Non';
                }
                items.push(`<div class="detail-row"><span class="detail-key">${label}</span><span class="detail-value">${displayValue}</span></div>`);
            }
            return items.join('');
        };

        container.innerHTML = `
            <div class="anomaly-detail-card">
                <div class="anomaly-detail-header-new ${severityClass}">
                    <div class="header-icon">
                        <span class="material-icons">${this.getAnomalyTypeIcon(anomaly.type)}</span>
                    </div>
                    <div class="header-content">
                        <h3>${anomaly.title}</h3>
                        <div class="header-badges">
                            <span class="type-badge">${this.getAnomalyTypeLabel(anomaly.type)}</span>
                            <span class="severity-badge-new ${severityClass}">${this.getSeverityLabel(anomaly.severity)}</span>
                            ${detectionMethod ? `<span class="method-badge">${this.getDetectionMethodBadge(detectionMethod)}</span>` : ''}
                        </div>
                    </div>
                </div>

                <div class="anomaly-description-box">
                    <p>${anomaly.description}</p>
                </div>

                <div class="anomaly-metrics">
                    <div class="metric-item">
                        <span class="material-icons">speed</span>
                        <div class="metric-content">
                            <span class="metric-value">${anomaly.confidence}%</span>
                            <span class="metric-label">Confiance${bayesianAdjustment ? ' (Bayésienne)' : ''}</span>
                        </div>
                        <div class="confidence-bar-small">
                            <div class="confidence-fill-small" style="width: ${anomaly.confidence}%"></div>
                        </div>
                        ${bayesianAdjustment ? `
                            <span class="bayesian-adjust ${bayesianAdjustment > 0 ? 'positive' : 'negative'}">
                                ${bayesianAdjustment > 0 ? '+' : ''}${bayesianAdjustment}%
                            </span>
                        ` : ''}
                    </div>
                    ${zScore !== undefined ? `
                        <div class="metric-item">
                            <span class="material-icons">analytics</span>
                            <div class="metric-content">
                                <span class="metric-value">${typeof zScore === 'number' ? zScore.toFixed(2) : zScore}</span>
                                <span class="metric-label">Z-Score</span>
                            </div>
                        </div>
                    ` : ''}
                    <div class="metric-item">
                        <span class="material-icons">schedule</span>
                        <div class="metric-content">
                            <span class="metric-value">${new Date(anomaly.detected_at).toLocaleDateString('fr-FR')}</span>
                            <span class="metric-label">Détecté le</span>
                        </div>
                    </div>
                </div>

                ${Object.keys(statisticalDetails).length > 0 ? `
                    <div class="anomaly-section-new stats-section">
                        <div class="section-header">
                            <span class="material-icons">analytics</span>
                            <span>Analyse statistique</span>
                        </div>
                        <div class="details-grid stats-grid">
                            ${formatDetails(statisticalDetails)}
                        </div>
                    </div>
                ` : ''}

                ${Object.keys(otherDetails).length > 0 ? `
                    <div class="anomaly-section-new">
                        <div class="section-header">
                            <span class="material-icons">data_object</span>
                            <span>Détails contextuels</span>
                        </div>
                        <div class="details-grid">
                            ${formatDetails(otherDetails)}
                        </div>
                    </div>
                ` : ''}

                ${anomaly.entity_ids?.length > 0 ? `
                    <div class="anomaly-section-new">
                        <div class="section-header">
                            <span class="material-icons">people</span>
                            <span>Entités impliquées</span>
                        </div>
                        <div class="related-tags">
                            ${anomaly.entity_ids.map(id => {
                                const entity = this.currentCase?.entities?.find(e => e.id === id);
                                return entity ? `<span class="related-tag entity" onclick="app.goToSearchResult('entity', '${id}')"><span class="material-icons">person</span>${entity.name}</span>` : '';
                            }).join('')}
                        </div>
                    </div>
                ` : ''}

                ${anomaly.event_ids?.length > 0 ? `
                    <div class="anomaly-section-new">
                        <div class="section-header">
                            <span class="material-icons">event</span>
                            <span>Événements liés</span>
                        </div>
                        <div class="related-tags">
                            ${anomaly.event_ids.map(id => {
                                const event = this.currentCase?.timeline?.find(e => e.id === id);
                                return event ? `<span class="related-tag event"><span class="material-icons">schedule</span>${event.title}</span>` : '';
                            }).join('')}
                        </div>
                    </div>
                ` : ''}

                ${anomaly.evidence_ids?.length > 0 ? `
                    <div class="anomaly-section-new">
                        <div class="section-header">
                            <span class="material-icons">find_in_page</span>
                            <span>Preuves liées</span>
                        </div>
                        <div class="related-tags">
                            ${anomaly.evidence_ids.map(id => {
                                const evidence = this.currentCase?.evidence?.find(e => e.id === id);
                                return evidence ? `<span class="related-tag evidence" onclick="app.goToSearchResult('evidence', '${id}')"><span class="material-icons">description</span>${evidence.name}</span>` : '';
                            }).join('')}
                        </div>
                    </div>
                ` : ''}

                ${anomaly.ai_explanation ? `
                    <div class="anomaly-section-new ai-section">
                        <div class="section-header">
                            <span class="material-icons">psychology</span>
                            <span>Explication IA</span>
                        </div>
                        <div class="ai-explanation-content">${this.formatMarkdown ? this.formatMarkdown(anomaly.ai_explanation) : anomaly.ai_explanation}</div>
                    </div>
                ` : ''}

                ${anomaly.recommendations?.length > 0 ? `
                    <div class="anomaly-section-new">
                        <div class="section-header">
                            <span class="material-icons">lightbulb</span>
                            <span>Recommandations</span>
                        </div>
                        <ul class="recommendations-list-new">
                            ${anomaly.recommendations.map(r => `<li><span class="material-icons">arrow_right</span>${r}</li>`).join('')}
                        </ul>
                    </div>
                ` : ''}

                <div class="anomaly-actions-footer">
                    <button class="btn btn-primary btn-sm" onclick="app.explainAnomaly('${anomaly.id}')">
                        <span class="material-icons">psychology</span>
                        Analyser avec l'IA
                    </button>
                    ${!anomaly.is_acknowledged ? `
                        <button class="btn btn-secondary btn-sm" onclick="app.acknowledgeAnomaly('${anomaly.id}')">
                            <span class="material-icons">check_circle</span>
                            Acquitter
                        </button>
                    ` : `
                        <span class="acknowledged-badge">
                            <span class="material-icons">verified</span>
                            Acquittée
                        </span>
                    `}
                </div>
            </div>
        `;
    },

    // ============================================
    // Explain Anomaly with AI
    // ============================================
    async explainAnomaly(anomalyId) {
        const anomaly = this.anomalies.find(a => a.id === anomalyId);
        if (!anomaly) return;

        this.setAnalysisContext('entity_analysis', `Anomalie: ${anomaly.title}`, 'Analyse d\'anomalie');

        const analysisModal = document.getElementById('analysis-modal');
        const analysisContent = document.getElementById('analysis-content');
        const modalTitle = document.getElementById('analysis-modal-title');
        if (modalTitle) modalTitle.textContent = `Analyse: ${anomaly.title}`;

        const noteBtn = document.getElementById('btn-save-to-notebook');
        if (noteBtn) noteBtn.style.display = '';

        analysisContent.innerHTML = `
            <div class="modal-explanation">
                <span class="material-icons">psychology</span>
                <p>Analyse de l'anomalie en cours...</p>
            </div>
            <div id="anomaly-explanation"><span class="streaming-cursor"></span></div>
        `;
        analysisModal.classList.add('active');

        try {
            const result = await this.apiCall('/api/anomaly/explain', 'POST', {
                case_id: this.currentCase.id,
                anomaly_id: anomalyId
            });

            document.getElementById('anomaly-explanation').innerHTML =
                this.formatMarkdown ? this.formatMarkdown(result.explanation) : result.explanation;

            // Update local anomaly
            const idx = this.anomalies.findIndex(a => a.id === anomalyId);
            if (idx !== -1) {
                this.anomalies[idx].ai_explanation = result.explanation;
                if (this.selectedAnomaly?.id === anomalyId) {
                    this.selectedAnomaly.ai_explanation = result.explanation;
                    this.renderAnomalyDetail();
                }
            }
        } catch (error) {
            document.getElementById('anomaly-explanation').innerHTML =
                `<p class="error">Erreur: ${error.message}</p>`;
        }
    },

    // ============================================
    // Acknowledge Anomaly
    // ============================================
    async acknowledgeAnomaly(anomalyId) {
        try {
            await this.apiCall('/api/anomaly/acknowledge', 'POST', {
                case_id: this.currentCase.id,
                anomaly_id: anomalyId
            });

            this.showToast('Anomalie acquittee');

            // Update local state
            const idx = this.anomalies.findIndex(a => a.id === anomalyId);
            if (idx !== -1) {
                this.anomalies[idx].is_acknowledged = true;
                this.anomalies[idx].is_new = false;
            }

            this.renderAnomaliesList();
            this.updateAnomalyBadge();

            if (this.selectedAnomaly?.id === anomalyId) {
                this.selectedAnomaly.is_acknowledged = true;
                this.renderAnomalyDetail();
            }
        } catch (error) {
            console.error('Error acknowledging anomaly:', error);
            this.showToast('Erreur lors de l\'acquittement', 'error');
        }
    },

    // ============================================
    // Show Anomaly Details Modal
    // ============================================
    showAnomalyDetails(anomalyId) {
        const anomaly = this.anomalies.find(a => a.id === anomalyId);
        if (!anomaly) return;

        const severityClass = `severity-${anomaly.severity}`;

        const content = `
            <div class="anomaly-detail-modal">
                <div class="anomaly-detail-header ${severityClass}">
                    <span class="material-icons">${this.getAnomalyTypeIcon(anomaly.type)}</span>
                    <h3>${anomaly.title}</h3>
                </div>

                <div class="detail-row">
                    <strong>Type:</strong> ${this.getAnomalyTypeLabel(anomaly.type)}
                </div>
                <div class="detail-row">
                    <strong>Severite:</strong>
                    <span class="${severityClass}">${this.getSeverityLabel(anomaly.severity)}</span>
                </div>
                <div class="detail-row">
                    <strong>Confiance:</strong> ${anomaly.confidence}%
                </div>
                <div class="detail-row">
                    <strong>Detecte le:</strong> ${new Date(anomaly.detected_at).toLocaleString()}
                </div>

                <div class="detail-section">
                    <strong>Description:</strong>
                    <p>${anomaly.description}</p>
                </div>

                ${anomaly.details ? `
                    <div class="detail-section">
                        <strong>Details techniques:</strong>
                        <pre>${JSON.stringify(anomaly.details, null, 2)}</pre>
                    </div>
                ` : ''}
            </div>
        `;

        this.showModal(`Anomalie: ${anomaly.title}`, content);
    },

    // ============================================
    // Alerts Management
    // ============================================
    showAllAlerts() {
        const content = `
            <div class="alerts-list">
                ${this.alerts.map(alert => `
                    <div class="alert-item severity-${alert.priority} ${alert.is_read ? 'read' : ''}">
                        <div class="alert-header">
                            <span class="material-icons">notification_important</span>
                            <span class="alert-time">${new Date(alert.created_at).toLocaleString()}</span>
                        </div>
                        <p class="alert-message">${alert.message}</p>
                        ${!alert.is_read ? `
                            <button class="btn btn-ghost btn-sm" onclick="app.markAlertRead('${alert.id}')">
                                Marquer comme lu
                            </button>
                        ` : ''}
                    </div>
                `).join('')}
            </div>
        `;

        this.showModal('Alertes d\'anomalies', content);
    },

    async markAlertRead(alertId) {
        try {
            await this.apiCall('/api/anomaly/alert/read', 'POST', {
                case_id: this.currentCase.id,
                alert_id: alertId
            });

            // Update local state
            const idx = this.alerts.findIndex(a => a.id === alertId);
            if (idx !== -1) {
                this.alerts[idx].is_read = true;
            }

            this.renderAlertsBanner();
            this.updateAnomalyBadge();
        } catch (error) {
            console.error('Error marking alert read:', error);
        }
    },

    dismissAlerts() {
        const container = document.getElementById('anomaly-alerts-banner');
        if (container) container.style.display = 'none';
    },

    // ============================================
    // Configuration
    // ============================================
    showAnomalyConfig() {
        if (!this.config) return;

        const content = `
            <div class="modal-explanation">
                <span class="material-icons">settings</span>
                <p>Configurez les parametres de detection d'anomalies pour cette affaire.</p>
            </div>
            <form id="anomaly-config-form">
                <div class="config-section">
                    <h4>Types de detection actifs</h4>
                    <div class="config-checkboxes">
                        <label class="config-checkbox">
                            <input type="checkbox" name="enable_timeline" ${this.config.enable_timeline ? 'checked' : ''}>
                            <span class="material-icons">timeline</span>
                            Timeline
                        </label>
                        <label class="config-checkbox">
                            <input type="checkbox" name="enable_financial" ${this.config.enable_financial ? 'checked' : ''}>
                            <span class="material-icons">attach_money</span>
                            Financier
                        </label>
                        <label class="config-checkbox">
                            <input type="checkbox" name="enable_communication" ${this.config.enable_communication ? 'checked' : ''}>
                            <span class="material-icons">chat</span>
                            Communication
                        </label>
                        <label class="config-checkbox">
                            <input type="checkbox" name="enable_behavior" ${this.config.enable_behavior ? 'checked' : ''}>
                            <span class="material-icons">psychology</span>
                            Comportement
                        </label>
                        <label class="config-checkbox">
                            <input type="checkbox" name="enable_location" ${this.config.enable_location ? 'checked' : ''}>
                            <span class="material-icons">place</span>
                            Localisation
                        </label>
                        <label class="config-checkbox">
                            <input type="checkbox" name="enable_relation" ${this.config.enable_relation ? 'checked' : ''}>
                            <span class="material-icons">people</span>
                            Relations
                        </label>
                        <label class="config-checkbox">
                            <input type="checkbox" name="enable_pattern" ${this.config.enable_pattern ? 'checked' : ''}>
                            <span class="material-icons">pattern</span>
                            Patterns
                        </label>
                    </div>
                </div>

                <div class="config-section">
                    <h4>Seuils</h4>
                    <div class="form-group">
                        <label class="form-label">Confiance minimum (%)</label>
                        <input type="range" class="form-range" name="min_confidence"
                               min="0" max="100" value="${this.config.min_confidence}"
                               oninput="document.getElementById('min-conf-value').textContent = this.value + '%'">
                        <span id="min-conf-value">${this.config.min_confidence}%</span>
                    </div>
                </div>

                <div class="config-section">
                    <h4>Alertes</h4>
                    <label class="config-checkbox">
                        <input type="checkbox" name="auto_alert" ${this.config.auto_alert ? 'checked' : ''}>
                        Alertes automatiques
                    </label>
                    <div class="form-group">
                        <label class="form-label">Seuil de severite pour alertes</label>
                        <select class="form-select" name="alert_severity_threshold">
                            <option value="critical" ${this.config.alert_severity_threshold === 'critical' ? 'selected' : ''}>Critique uniquement</option>
                            <option value="high" ${this.config.alert_severity_threshold === 'high' ? 'selected' : ''}>Eleve et plus</option>
                            <option value="medium" ${this.config.alert_severity_threshold === 'medium' ? 'selected' : ''}>Moyen et plus</option>
                            <option value="low" ${this.config.alert_severity_threshold === 'low' ? 'selected' : ''}>Tout</option>
                        </select>
                    </div>
                </div>
            </form>
        `;

        this.showModal('Configuration Detection', content, async () => {
            const form = document.getElementById('anomaly-config-form');
            const updatedConfig = {
                enable_timeline: form.querySelector('[name="enable_timeline"]').checked,
                enable_financial: form.querySelector('[name="enable_financial"]').checked,
                enable_communication: form.querySelector('[name="enable_communication"]').checked,
                enable_behavior: form.querySelector('[name="enable_behavior"]').checked,
                enable_location: form.querySelector('[name="enable_location"]').checked,
                enable_relation: form.querySelector('[name="enable_relation"]').checked,
                enable_pattern: form.querySelector('[name="enable_pattern"]').checked,
                min_confidence: parseInt(form.querySelector('[name="min_confidence"]').value),
                auto_alert: form.querySelector('[name="auto_alert"]').checked,
                alert_severity_threshold: form.querySelector('[name="alert_severity_threshold"]').value
            };

            try {
                await this.apiCall(`/api/anomaly/config?case_id=${this.currentCase.id}`, 'PUT', updatedConfig);
                this.config = { ...this.config, ...updatedConfig };
                this.showToast('Configuration sauvegardee');
            } catch (error) {
                console.error('Error saving config:', error);
                this.showToast('Erreur lors de la sauvegarde', 'error');
            }
        });
    },

    // ============================================
    // Filter Methods
    // ============================================
    setAnomalyTypeFilter(type) {
        this.filterType = type;
        this.renderAnomaliesList();
    },

    setAnomalySeverityFilter(severity) {
        this.filterSeverity = severity;
        this.renderAnomaliesList();
    },

    // ============================================
    // Update Badge in Navigation
    // ============================================
    updateAnomalyBadge() {
        const badge = document.getElementById('anomaly-nav-badge');
        const newCount = this.anomalies.filter(a => a.is_new && !a.is_acknowledged).length;

        if (badge) {
            if (newCount > 0) {
                badge.textContent = newCount;
                badge.style.display = 'inline-flex';
            } else {
                badge.style.display = 'none';
            }
        }
    },

    // ============================================
    // Helper Methods
    // ============================================
    getAnomalyTypeIcon(type) {
        const icons = {
            'timeline': 'timeline',
            'financial': 'attach_money',
            'communication': 'chat',
            'behavior': 'psychology',
            'location': 'place',
            'relation': 'people',
            'pattern': 'pattern'
        };
        return icons[type] || 'warning';
    },

    getAnomalyTypeLabel(type) {
        const labels = {
            'timeline': 'Timeline',
            'financial': 'Financier',
            'communication': 'Communication',
            'behavior': 'Comportement',
            'location': 'Localisation',
            'relation': 'Relation',
            'pattern': 'Pattern'
        };
        return labels[type] || type;
    },

    getSeverityLabel(severity) {
        const labels = {
            'critical': 'Critique',
            'high': 'Eleve',
            'medium': 'Moyen',
            'low': 'Faible',
            'info': 'Info'
        };
        return labels[severity] || severity;
    },

    // ============================================
    // Detection Method Helpers
    // ============================================
    getDetectionMethodBadge(method) {
        const badges = {
            'adaptive_zscore': '<span class="material-icons">bar_chart</span> Z-Score Adaptatif',
            'modified_zscore': '<span class="material-icons">trending_up</span> Z-Score Modifié',
            'standard_zscore': '<span class="material-icons">show_chart</span> Z-Score Standard',
            'network_centrality': '<span class="material-icons">hub</span> Centralité Réseau',
            'pagerank': '<span class="material-icons">public</span> PageRank',
            'betweenness': '<span class="material-icons">mediation</span> Betweenness',
            'cusum': '<span class="material-icons">stacked_line_chart</span> CUSUM',
            'rupture_detection': '<span class="material-icons">flash_on</span> Rupture',
            'cross_correlation': '<span class="material-icons">sync_alt</span> Corrélation',
            'bayesian': '<span class="material-icons">casino</span> Bayésien',
            'pattern_analysis': '<span class="material-icons">search</span> Pattern',
            'statistical': '<span class="material-icons">analytics</span> Statistique',
            'heuristic': '<span class="material-icons">lightbulb</span> Heuristique'
        };
        return badges[method] || (method ? `<span class="material-icons">science</span> ${method}` : '');
    },

    getMethodTooltip(method) {
        const tooltips = {
            'adaptive_zscore': 'Détection par Z-Score adaptatif avec seuils dynamiques basés sur la distribution des données',
            'modified_zscore': 'Z-Score modifié utilisant la médiane et MAD pour plus de robustesse aux outliers',
            'standard_zscore': 'Z-Score classique basé sur la moyenne et l\'écart-type',
            'network_centrality': 'Analyse de centralité du réseau pour identifier les acteurs clés suspects',
            'pagerank': 'Algorithme PageRank pour évaluer l\'importance relative des entités',
            'betweenness': 'Centralité d\'intermédiarité pour détecter les points de contrôle critiques',
            'cusum': 'Détection de rupture CUSUM (Cumulative Sum) pour les changements de tendance',
            'rupture_detection': 'Détection automatique de points de rupture dans les séries temporelles',
            'cross_correlation': 'Corrélation croisée entre différents types d\'anomalies',
            'bayesian': 'Ajustement bayésien de la confiance avec probabilités a priori',
            'pattern_analysis': 'Analyse de patterns comportementaux récurrents',
            'statistical': 'Méthodes statistiques générales',
            'heuristic': 'Règles heuristiques basées sur l\'expertise métier'
        };
        return tooltips[method] || `Méthode de détection: ${method}`;
    }
};

// Export for use in main app
if (typeof module !== 'undefined' && module.exports) {
    module.exports = AnomaliesModule;
}
