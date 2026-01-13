// ForensicInvestigator - Module Chat
// Chat IA et streaming des réponses

const ChatModule = {
    // ============================================
    // Init Chat
    // ============================================
    initChat() {
        document.getElementById('btn-send-chat')?.addEventListener('click', () => this.sendChatMessage());
        document.getElementById('chat-input')?.addEventListener('keydown', (e) => {
            if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault();
                this.sendChatMessage();
            }
        });

        // Setup AI Assistant panel
        this.setupAIAssistant();
    },

    // ============================================
    // AI Assistant Panel Setup
    // ============================================
    setupAIAssistant() {
        const toggle = document.getElementById('ai-assistant-toggle');
        const assistant = document.getElementById('ai-assistant');
        const input = document.getElementById('ai-input');
        const sendBtn = document.getElementById('btn-ai-send');

        if (toggle && assistant) {
            // Draggable functionality
            let isDragging = false;
            let hasMoved = false;
            let startX, startY, initialX, initialY;

            toggle.addEventListener('mousedown', (e) => {
                // Start potential drag
                isDragging = true;
                hasMoved = false;

                const rect = assistant.getBoundingClientRect();
                startX = e.clientX;
                startY = e.clientY;
                initialX = rect.left;
                initialY = rect.top;

                e.preventDefault();
            });

            document.addEventListener('mousemove', (e) => {
                if (!isDragging) return;

                const dx = e.clientX - startX;
                const dy = e.clientY - startY;

                // Consider it a drag if moved more than 5 pixels
                if (Math.abs(dx) > 5 || Math.abs(dy) > 5) {
                    hasMoved = true;

                    // Apply fixed positioning on first move
                    if (!assistant.classList.contains('dragging')) {
                        assistant.classList.add('dragging');
                        assistant.style.position = 'fixed';
                        assistant.style.bottom = 'auto';
                        assistant.style.right = 'auto';
                    }

                    assistant.style.left = (initialX + dx) + 'px';
                    assistant.style.top = (initialY + dy) + 'px';
                }
            });

            document.addEventListener('mouseup', () => {
                if (!isDragging) return;

                // If it was a click (not a drag), toggle collapsed state
                if (!hasMoved) {
                    assistant.classList.toggle('collapsed');
                }

                assistant.classList.remove('dragging');
                isDragging = false;
                hasMoved = false;
            });
        }

        if (sendBtn) {
            sendBtn.addEventListener('click', () => this.sendAIMessage());
        }

        if (input) {
            input.addEventListener('keypress', (e) => {
                if (e.key === 'Enter') this.sendAIMessage();
            });
        }
    },

    // ============================================
    // Send AI Message (Assistant Panel)
    // ============================================
    async sendAIMessage() {
        const input = document.getElementById('ai-input');
        const message = input?.value.trim();
        if (!message) return;

        const messagesContainer = document.getElementById('ai-messages');
        if (!messagesContainer) return;

        // Add user message
        messagesContainer.innerHTML += `
            <div class="ai-message ai-message-user">
                <strong>Vous:</strong> ${this.escapeHtml(message)}
            </div>
        `;
        input.value = '';
        messagesContainer.scrollTop = messagesContainer.scrollHeight;

        // Add response container
        const responseId = 'ai-response-' + Date.now();
        messagesContainer.innerHTML += `
            <div class="ai-message ai-message-bot" id="${responseId}">
                <span class="streaming-cursor">▊</span>
            </div>
        `;
        messagesContainer.scrollTop = messagesContainer.scrollHeight;

        try {
            const response = await fetch('/api/chat/stream', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    case_id: this.currentCase?.id || '',
                    message: message
                })
            });

            const responseDiv = document.getElementById(responseId);
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
                                responseDiv.innerHTML = `<strong>Erreur:</strong> ${data.error}`;
                                return;
                            }
                            if (data.chunk) {
                                fullResponse += data.chunk;
                                responseDiv.innerHTML = marked.parse(fullResponse) + '<span class="streaming-cursor">▊</span>';
                                messagesContainer.scrollTop = messagesContainer.scrollHeight;
                            }
                            if (data.done) {
                                responseDiv.innerHTML = marked.parse(fullResponse);
                            }
                        } catch (e) {
                            // Ignore parsing errors
                        }
                    }
                }
            }

            if (fullResponse) {
                responseDiv.innerHTML = marked.parse(fullResponse);
            }
        } catch (error) {
            document.getElementById(responseId).innerHTML = `<strong>Erreur:</strong> Désolé, une erreur s'est produite.`;
        }
        messagesContainer.scrollTop = messagesContainer.scrollHeight;
    },

    // ============================================
    // Send Chat Message
    // ============================================
    async sendChatMessage() {
        const input = document.getElementById('chat-input');
        const message = input.value.trim();

        if (!message) return;

        if (!this.currentCase) {
            this.showToast('Veuillez sélectionner une affaire', 'warning');
            return;
        }

        const chatHistory = document.getElementById('chat-history');

        // Add user message
        chatHistory.innerHTML += `
            <div class="chat-message user">
                <div class="chat-avatar"><span class="material-icons">person</span></div>
                <div class="chat-content">${this.escapeHtml(message)}</div>
            </div>
        `;

        // Clear input
        input.value = '';

        // Add AI response placeholder
        const responseId = `ai-response-${Date.now()}`;
        chatHistory.innerHTML += `
            <div class="chat-message ai">
                <div class="chat-avatar"><span class="material-icons">psychology</span></div>
                <div class="chat-content" id="${responseId}">
                    <span class="streaming-cursor">▊</span>
                </div>
            </div>
        `;

        // Scroll to bottom
        chatHistory.scrollTop = chatHistory.scrollHeight;

        // Stream response
        const responseDiv = document.getElementById(responseId);
        await this.streamAIResponse(
            '/api/chat/stream',
            { case_id: this.currentCase.id, message },
            responseDiv
        );

        // Scroll to bottom after response
        chatHistory.scrollTop = chatHistory.scrollHeight;
    },

    // ============================================
    // Streaming AI Response
    // ============================================
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

            // Final render
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
    // AI Analysis Methods
    // ============================================
    async analyzeCase() {
        if (!this.currentCase) {
            this.showToast('Sélectionnez une affaire d\'abord', 'warning');
            return;
        }

        this.setAnalysisContext('graph_analysis', `Analyse de l'affaire - ${this.currentCase.name}`, 'Analyse complète');

        const analysisContent = document.getElementById('analysis-content');
        const analysisModal = document.getElementById('analysis-modal');
        const modalTitle = document.getElementById('analysis-modal-title');
        if (modalTitle) modalTitle.textContent = 'Analyse IA - Affaire';

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

    async generateQuestions() {
        if (!this.currentCase) {
            this.showToast('Sélectionnez une affaire d\'abord', 'warning');
            return;
        }

        this.setAnalysisContext('question', `Questions d'investigation - ${this.currentCase.name}`, 'Génération automatique');

        const analysisContent = document.getElementById('analysis-content');
        const analysisModal = document.getElementById('analysis-modal');
        const modalTitle = document.getElementById('analysis-modal-title');
        if (modalTitle) modalTitle.textContent = 'Questions d\'investigation';

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

    async detectContradictions() {
        if (!this.currentCase) {
            this.showToast('Sélectionnez une affaire d\'abord', 'warning');
            return;
        }

        this.setAnalysisContext('contradiction', `Contradictions - ${this.currentCase.name}`, 'Détection automatique');

        const analysisContent = document.getElementById('analysis-content');
        const analysisModal = document.getElementById('analysis-modal');
        const modalTitle = document.getElementById('analysis-modal-title');
        if (modalTitle) modalTitle.textContent = 'Détection de contradictions';

        const noteBtn = document.getElementById('btn-save-to-notebook');
        if (noteBtn) noteBtn.style.display = '';

        analysisContent.innerHTML = '<span class="streaming-cursor">▊</span>';
        analysisModal.classList.add('active');

        await this.streamAIResponse(
            '/api/contradictions/detect/stream',
            { case_id: this.currentCase.id },
            analysisContent
        );
    },

    // ============================================
    // Escape HTML Helper
    // ============================================
    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    },

    // ============================================
    // Format Markdown
    // ============================================
    formatMarkdown(text) {
        if (!text) return '';
        return marked.parse(text);
    },

    // ============================================
    // Process Markdown Tables
    // ============================================
    processMarkdownTables(text) {
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
    }
};

// Export for use in main app
if (typeof module !== 'undefined' && module.exports) {
    module.exports = ChatModule;
}
