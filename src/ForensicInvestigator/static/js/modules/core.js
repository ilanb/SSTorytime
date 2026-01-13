// ForensicInvestigator - Module Core
// Utilitaires de base, navigation, modals et API

const CoreModule = {
    // ============================================
    // Recent Cases Management (localStorage)
    // ============================================
    loadRecentCases() {
        try {
            const stored = localStorage.getItem('forensic_recent_cases');
            return stored ? JSON.parse(stored) : [];
        } catch (e) {
            return [];
        }
    },

    saveRecentCase(caseId) {
        this.recentCases = this.recentCases.filter(id => id !== caseId);
        this.recentCases.unshift(caseId);
        this.recentCases = this.recentCases.slice(0, 50);
        try {
            localStorage.setItem('forensic_recent_cases', JSON.stringify(this.recentCases));
        } catch (e) {
            console.warn('Unable to save recent cases to localStorage');
        }
    },

    // ============================================
    // Navigation
    // ============================================
    setupNavigation() {
        const navBtns = document.querySelectorAll('.nav-btn');
        navBtns.forEach(btn => {
            btn.addEventListener('click', (e) => {
                e.preventDefault();
                const view = btn.dataset.view;
                this.switchView(view);
                navBtns.forEach(b => b.classList.remove('active'));
                btn.classList.add('active');
            });
        });
    },

    switchView(viewName) {
        const views = document.querySelectorAll('[id^="view-"]');
        views.forEach(v => v.classList.add('hidden'));

        const targetView = document.getElementById(`view-${viewName}`);
        if (targetView) {
            targetView.classList.remove('hidden');

            // Refresh view-specific content
            if (viewName === 'dashboard' && this.currentCase) {
                if (typeof this.renderGraph === 'function') this.renderGraph();
            } else if (viewName === 'n4l' && this.currentCase) {
                if (typeof this.loadN4LContent === 'function') this.loadN4LContent();
            } else if (viewName === 'config') {
                if (typeof this.loadConfig === 'function') this.loadConfig();
            } else if (viewName === 'notebook') {
                if (typeof this.loadNotebook === 'function') this.loadNotebook();
            } else if (viewName === 'geo-map') {
                // Render geo map view
                const content = document.getElementById('geo-map-content');
                if (content && typeof this.renderGeoMap === 'function') {
                    content.innerHTML = this.renderGeoMap();
                    // Initialize Leaflet map after DOM is updated
                    if (this.currentCase) {
                        setTimeout(() => {
                            if (typeof this.initLeafletMap === 'function') {
                                this.initLeafletMap();
                            }
                        }, 150);
                    }
                }
            } else if (viewName === 'social-network') {
                // Render social network view
                const content = document.getElementById('social-network-content');
                if (content && typeof this.renderSocialNetwork === 'function') {
                    content.innerHTML = this.renderSocialNetwork();
                    if (this.currentCase) {
                        setTimeout(() => {
                            if (typeof this.initSocialNetworkGraph === 'function') {
                                this.initSocialNetworkGraph();
                            }
                        }, 150);
                    }
                }
            } else if (viewName === 'cross-case') {
                // Auto-scan cross-case connections when entering the view
                if (typeof this.scanCrossConnections === 'function') {
                    this.scanCrossConnections();
                }
            } else if (viewName === 'graph-analysis') {
                // Auto-launch graph analysis when entering the view
                if (this.currentCase && typeof this.analyzeGraphComplete === 'function') {
                    setTimeout(() => {
                        this.analyzeGraphComplete();
                    }, 150);
                }
            }
        }

        // Reshow "Noter" button for new analyses
        const noteBtn = document.getElementById('btn-save-to-notebook');
        if (noteBtn) noteBtn.style.display = '';
    },

    // ============================================
    // Modals
    // ============================================
    setupModals() {
        const overlay = document.getElementById('modal-overlay');
        const closeBtn = document.getElementById('modal-close');
        const cancelBtn = document.getElementById('modal-cancel');

        closeBtn.addEventListener('click', () => this.closeModal());
        cancelBtn.addEventListener('click', () => this.closeModal());
        overlay.addEventListener('click', (e) => {
            if (e.target === overlay) this.closeModal();
        });
    },

    showModal(title, content, onConfirm) {
        document.getElementById('modal-title').textContent = title;
        document.getElementById('modal-body').innerHTML = content;
        document.getElementById('modal-overlay').classList.add('active');

        const confirmBtn = document.getElementById('modal-confirm');
        confirmBtn.onclick = () => {
            if (onConfirm) onConfirm();
            this.closeModal();
        };
    },

    closeModal() {
        document.getElementById('modal-overlay').classList.remove('active');
    },

    showAnalysisModal(content, title = 'Analyse IA', type = 'graph_analysis', context = '') {
        this.setAnalysisContext(type, title, context);

        const modalTitle = document.getElementById('analysis-modal-title');
        if (modalTitle) modalTitle.textContent = title;

        const noteBtn = document.getElementById('btn-save-to-notebook');
        if (noteBtn) noteBtn.style.display = '';

        document.getElementById('analysis-content').innerHTML = marked.parse(content);
        document.getElementById('analysis-modal').classList.add('active');
    },

    toggleHelpSidebar(show) {
        const sidebar = document.getElementById('help-sidebar');
        const overlay = document.getElementById('help-overlay');
        if (show) {
            sidebar.classList.add('active');
            overlay.classList.add('active');
            document.body.style.overflow = 'hidden';
        } else {
            sidebar.classList.remove('active');
            overlay.classList.remove('active');
            document.body.style.overflow = '';
        }
    },

    // ============================================
    // Toast Notifications
    // ============================================
    showToast(message, type = 'info') {
        const container = document.getElementById('toast-container') || this.createToastContainer();
        const toast = document.createElement('div');
        toast.className = `toast toast-${type}`;

        const icons = {
            success: 'check_circle',
            error: 'error',
            warning: 'warning',
            info: 'info'
        };

        toast.innerHTML = `
            <span class="material-icons">${icons[type] || 'info'}</span>
            <span>${message}</span>
        `;

        container.appendChild(toast);
        setTimeout(() => toast.classList.add('show'), 10);
        setTimeout(() => {
            toast.classList.remove('show');
            setTimeout(() => toast.remove(), 300);
        }, 3000);
    },

    createToastContainer() {
        const container = document.createElement('div');
        container.id = 'toast-container';
        document.body.appendChild(container);
        return container;
    },

    // ============================================
    // API Calls
    // ============================================
    async apiCall(endpoint, method = 'GET', body = null) {
        const options = {
            method,
            headers: { 'Content-Type': 'application/json' }
        };
        if (body) options.body = JSON.stringify(body);

        const response = await fetch(endpoint, options);
        if (!response.ok) {
            throw new Error(await response.text());
        }

        const contentType = response.headers.get('content-type');
        if (contentType && contentType.includes('application/json')) {
            return response.json();
        }
        return response.text();
    },

    async streamAIResponse(endpoint, body, targetElement, options = {}) {
        const {
            onStart = null,
            onChunk = null,
            onComplete = null,
            showCursor = true
        } = options;

        if (onStart) onStart();

        try {
            const response = await fetch(endpoint, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(body)
            });

            let fullResponse = '';
            const reader = response.body.getReader();
            const decoder = new TextDecoder();

            while (true) {
                const { done, value } = await reader.read();
                if (done) break;

                const chunk = decoder.decode(value);
                const lines = chunk.split('\n');

                for (const line of lines) {
                    if (line.startsWith('data: ')) {
                        try {
                            const data = JSON.parse(line.slice(6));
                            if (data.error) {
                                targetElement.innerHTML = `<div class="error-message"><span class="material-icons">error</span> ${data.error}</div>`;
                                return null;
                            }
                            if (data.chunk) {
                                fullResponse += data.chunk;
                                if (onChunk) {
                                    onChunk(fullResponse);
                                } else {
                                    const cursor = showCursor ? '<span class="streaming-cursor">▊</span>' : '';
                                    targetElement.innerHTML = marked.parse(fullResponse) + cursor;
                                }
                            }
                            if (data.done) {
                                targetElement.innerHTML = marked.parse(fullResponse);
                                if (onComplete) onComplete(fullResponse);
                            }
                        } catch (e) {
                            // Ignore parsing errors for incomplete chunks
                        }
                    }
                }
            }

            if (fullResponse) {
                targetElement.innerHTML = marked.parse(fullResponse);
                if (onComplete) onComplete(fullResponse);
            }

            return fullResponse;
        } catch (error) {
            targetElement.innerHTML = `<div class="error-message"><span class="material-icons">error</span> Erreur: ${error.message}</div>`;
            return null;
        }
    },

    // ============================================
    // Table Rendering (for AI responses)
    // ============================================
    parseAndRenderTables(text) {
        const lines = text.split('\n');
        const result = [];
        let inTable = false;
        let tableRows = [];

        for (let i = 0; i < lines.length; i++) {
            const line = lines[i].trim();

            if (line.startsWith('|') && line.endsWith('|')) {
                if (/^\|[\s\-:|]+\|$/.test(line) && line.includes('-')) {
                    continue;
                }

                if (!inTable) {
                    inTable = true;
                    tableRows = [];
                }

                const cells = line.slice(1, -1).split('|').map(c => c.trim());
                tableRows.push(cells);
            } else {
                if (inTable && tableRows.length > 0) {
                    result.push(this.renderTable(tableRows));
                    tableRows = [];
                    inTable = false;
                }
                result.push(line);
            }
        }

        if (inTable && tableRows.length > 0) {
            result.push(this.renderTable(tableRows));
        }

        return result.join('\n');
    },

    renderTable(rows) {
        if (rows.length === 0) return '';

        const tableStyle = 'width: 100%; border-collapse: collapse; margin: 1rem 0; font-size: 0.85rem;';
        const thStyle = 'background: #1e3a5f; color: white; padding: 0.625rem 0.5rem; border: 1px solid #1e3a5f; text-align: left; font-weight: 600;';
        const tdStyle = 'padding: 0.5rem; border: 1px solid #e2e8f0; color: #1a1a2e;';

        let html = `<table style="${tableStyle}">`;

        html += '<thead><tr>';
        rows[0].forEach(cell => {
            html += `<th style="${thStyle}">${cell}</th>`;
        });
        html += '</tr></thead>';

        if (rows.length > 1) {
            html += '<tbody>';
            for (let i = 1; i < rows.length; i++) {
                const rowBg = i % 2 === 0 ? 'background: #f8fafc;' : '';
                html += `<tr style="${rowBg}">`;
                rows[i].forEach(cell => {
                    html += `<td style="${tdStyle}">${cell}</td>`;
                });
                html += '</tr>';
            }
            html += '</tbody>';
        }

        html += '</table>';
        return html;
    },

    // ============================================
    // Entity Icon Helper
    // ============================================
    getEntityIcon(type) {
        const icons = {
            'personne': 'person',
            'lieu': 'place',
            'objet': 'inventory_2',
            'evenement': 'event',
            'organisation': 'business',
            'document': 'description',
            'vehicule': 'directions_car'
        };
        return icons[type] || 'help_outline';
    },

    // ============================================
    // Markdown Formatting
    // ============================================
    formatMarkdownTables(text) {
        const lines = text.split('\n');
        let result = [];
        let inTable = false;
        let tableRows = [];

        for (let i = 0; i < lines.length; i++) {
            const line = lines[i].trim();

            // Détecter une ligne de tableau (commence et finit par |)
            if (line.startsWith('|') && line.endsWith('|')) {
                // Vérifier si c'est une ligne de séparation (|---|---|) - contient uniquement |, -, :, espaces
                if (/^\|[\s\-:|]+\|$/.test(line) && line.includes('-')) {
                    // Ligne de séparation, on la saute
                    continue;
                }

                if (!inTable) {
                    inTable = true;
                    tableRows = [];
                }

                // Extraire les cellules
                const cells = line.slice(1, -1).split('|').map(c => c.trim());
                tableRows.push(cells);
            } else {
                // Fin du tableau
                if (inTable && tableRows.length > 0) {
                    result.push(this.renderTable(tableRows));
                    tableRows = [];
                    inTable = false;
                }
                result.push(line);
            }
        }

        // Gérer un tableau en fin de texte
        if (inTable && tableRows.length > 0) {
            result.push(this.renderTable(tableRows));
        }

        return result.join('\n');
    },

    formatAnalysisText(text) {
        if (!text) return '<p>Aucune analyse disponible</p>';

        // Convert Markdown to HTML
        let html = text;

        // Escape HTML entities first (except for markdown we'll process)
        html = html.replace(/&/g, '&amp;');

        // Process tables first (before other formatting)
        html = this.processMarkdownTables(html);

        // Headers: ## Title -> <h2>Title</h2>
        html = html.replace(/^#{6}\s+(.+)$/gm, '<h6 class="md-h6">$1</h6>');
        html = html.replace(/^#{5}\s+(.+)$/gm, '<h5 class="md-h5">$1</h5>');
        html = html.replace(/^#{4}\s+(.+)$/gm, '<h4 class="md-h4">$1</h4>');
        html = html.replace(/^#{3}\s+(.+)$/gm, '<h3 class="md-h3">$1</h3>');
        html = html.replace(/^#{2}\s+(.+)$/gm, '<h2 class="md-h2">$1</h2>');
        html = html.replace(/^#{1}\s+(.+)$/gm, '<h1 class="md-h1">$1</h1>');

        // Horizontal rules: --- or ***
        html = html.replace(/^[-*]{3,}$/gm, '<hr class="md-hr">');

        // Bold: **text** or __text__
        html = html.replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>');
        html = html.replace(/__(.+?)__/g, '<strong>$1</strong>');

        // Italic: *text* or _text_
        html = html.replace(/\*([^*]+)\*/g, '<em>$1</em>');
        html = html.replace(/_([^_]+)_/g, '<em>$1</em>');

        // Inline code: `code`
        html = html.replace(/`([^`]+)`/g, '<code class="md-code">$1</code>');

        // Process lists
        html = this.processMarkdownLists(html);

        // Wrap remaining lines in paragraphs
        const lines = html.split('\n');
        const processed = [];
        let inParagraph = false;

        for (const line of lines) {
            const trimmed = line.trim();

            // Skip if already wrapped in HTML tags
            if (trimmed.startsWith('<h') || trimmed.startsWith('<ul') ||
                trimmed.startsWith('<ol') || trimmed.startsWith('<li') ||
                trimmed.startsWith('<table') || trimmed.startsWith('<tr') ||
                trimmed.startsWith('<hr') || trimmed.startsWith('</') ||
                trimmed.startsWith('<div') || trimmed === '') {
                if (inParagraph) {
                    processed.push('</p>');
                    inParagraph = false;
                }
                processed.push(line);
            } else {
                if (!inParagraph) {
                    processed.push('<p>');
                    inParagraph = true;
                }
                processed.push(trimmed);
            }
        }
        if (inParagraph) {
            processed.push('</p>');
        }

        return processed.join('\n');
    },

    processMarkdownTables(text) {
        const lines = text.split('\n');
        const result = [];
        let inTable = false;
        let tableRows = [];

        for (let i = 0; i < lines.length; i++) {
            const line = lines[i].trim();

            // Check if line is a table row (starts and ends with |)
            if (line.startsWith('|') && line.endsWith('|')) {
                // Check if this line is a separator (|---|---|)
                const isSeparator = /^\|[\s-:|]+\|$/.test(line);

                if (!inTable && !isSeparator) {
                    inTable = true;
                    tableRows = [];
                }

                if (inTable && !isSeparator) {
                    const cells = line.split('|').slice(1, -1).map(c => c.trim());
                    tableRows.push(cells);
                }
            } else {
                if (inTable && tableRows.length > 0) {
                    // Output table
                    result.push('<div class="md-table-wrapper"><table class="md-table">');
                    tableRows.forEach((row, idx) => {
                        const tag = idx === 0 ? 'th' : 'td';
                        const rowClass = idx === 0 ? 'md-table-header' : '';
                        result.push(`<tr class="${rowClass}">`);
                        row.forEach(cell => {
                            result.push(`<${tag}>${cell}</${tag}>`);
                        });
                        result.push('</tr>');
                    });
                    result.push('</table></div>');
                    tableRows = [];
                }
                inTable = false;
                result.push(lines[i]);
            }
        }

        // Handle table at end of text
        if (inTable && tableRows.length > 0) {
            result.push('<div class="md-table-wrapper"><table class="md-table">');
            tableRows.forEach((row, idx) => {
                const tag = idx === 0 ? 'th' : 'td';
                result.push('<tr>');
                row.forEach(cell => {
                    result.push(`<${tag}>${cell}</${tag}>`);
                });
                result.push('</tr>');
            });
            result.push('</table></div>');
        }

        return result.join('\n');
    },

    processMarkdownLists(text) {
        const lines = text.split('\n');
        const result = [];
        let inUl = false;
        let inOl = false;

        for (const line of lines) {
            const trimmed = line.trim();

            // Unordered list: - item or * item
            const ulMatch = trimmed.match(/^[-*]\s+(.+)$/);
            // Ordered list: 1. item
            const olMatch = trimmed.match(/^\d+\.\s+(.+)$/);

            if (ulMatch) {
                if (!inUl) {
                    if (inOl) { result.push('</ol>'); inOl = false; }
                    result.push('<ul class="md-list">');
                    inUl = true;
                }
                result.push(`<li>${ulMatch[1]}</li>`);
            } else if (olMatch) {
                if (!inOl) {
                    if (inUl) { result.push('</ul>'); inUl = false; }
                    result.push('<ol class="md-list">');
                    inOl = true;
                }
                result.push(`<li>${olMatch[1]}</li>`);
            } else {
                if (inUl) { result.push('</ul>'); inUl = false; }
                if (inOl) { result.push('</ol>'); inOl = false; }
                result.push(line);
            }
        }

        if (inUl) result.push('</ul>');
        if (inOl) result.push('</ol>');

        return result.join('\n');
    },

    updateModalContent(html) {
        const modalBody = document.querySelector('.modal-body');
        if (modalBody) {
            modalBody.innerHTML = html;
        }
    }
};

// Export for use in main app
if (typeof module !== 'undefined' && module.exports) {
    module.exports = CoreModule;
}
