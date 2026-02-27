// Modal utility for in-app confirmation dialogs
// This replaces browser's native confirm() with styled modals

(function() {
    'use strict';

    // Modal state
    let currentModal = null;
    let currentResolve = null;

    /**
     * Show a confirmation modal dialog
     * @param {Object} options - Modal configuration
     * @param {string} options.title - Modal title
     * @param {string} options.message - Modal message/body
     * @param {string} options.confirmText - Confirm button text (default: "Confirm")
     * @param {string} options.cancelText - Cancel button text (default: "Cancel")
     * @param {string} options.type - Modal type: "warning", "danger", "info" (default: "warning")
     * @returns {Promise<boolean>} - Resolves to true if confirmed, false if cancelled
     */
    window.showConfirmModal = function(options) {
        return new Promise((resolve) => {
            // Close any existing modal
            closeModal();

            // Store resolve function
            currentResolve = resolve;

            // Create modal elements
            const overlay = document.createElement('div');
            overlay.className = 'modal-overlay';
            overlay.id = 'confirm-modal-overlay';

            const iconMap = {
                warning: '⚠️',
                danger: '🗑️',
                info: 'ℹ️'
            };

            const type = options.type || 'warning';
            const icon = iconMap[type] || iconMap.warning;

            overlay.innerHTML = `
                <div class="modal" role="dialog" aria-modal="true" aria-labelledby="modal-title">
                    <div class="modal-header">
                        <span class="modal-icon ${type}">${icon}</span>
                        <h3 id="modal-title">${escapeHtml(options.title || 'Confirm Action')}</h3>
                    </div>
                    <div class="modal-body">
                        <p>${escapeHtml(options.message || 'Are you sure you want to proceed?')}</p>
                    </div>
                    <div class="modal-actions">
                        <button class="btn btn-cancel" id="modal-cancel" type="button">
                            ${escapeHtml(options.cancelText || 'Cancel')}
                        </button>
                        <button class="btn btn-confirm-danger" id="modal-confirm" type="button">
                            ${escapeHtml(options.confirmText || 'Confirm')}
                        </button>
                    </div>
                </div>
            `;

            // Add to document
            document.body.appendChild(overlay);
            currentModal = overlay;

            // Trigger reflow then show
            overlay.offsetHeight;
            overlay.classList.add('show');

            // Focus confirm button
            const confirmBtn = overlay.querySelector('#modal-confirm');
            const cancelBtn = overlay.querySelector('#modal-cancel');
            confirmBtn.focus();

            // Event handlers
            confirmBtn.addEventListener('click', handleConfirm);
            cancelBtn.addEventListener('click', handleCancel);
            overlay.addEventListener('click', handleOverlayClick);
            document.addEventListener('keydown', handleKeydown);
        });
    };

    /**
     * Close the current modal
     * IMPORTANT: Captures the modal reference before clearing to prevent race conditions
     * when modals are opened in rapid succession (e.g., double-click scenarios).
     */
    function closeModal() {
        if (currentModal) {
            const modalToRemove = currentModal; // Capture reference before clearing
            currentModal = null; // Clear immediately to prevent race conditions
            
            modalToRemove.classList.remove('show');
            // Wait for transition then remove from DOM
            setTimeout(() => {
                if (modalToRemove.parentNode) {
                    modalToRemove.parentNode.removeChild(modalToRemove);
                }
            }, window.FORUM_CONSTANTS?.MODAL_TRANSITION_MS || 200);
            
            // Remove event listeners
            document.removeEventListener('keydown', handleKeydown);
        }
    }

    /**
     * Handle confirm button click
     */
    function handleConfirm() {
        closeModal();
        if (currentResolve) {
            currentResolve(true);
            currentResolve = null;
        }
    }

    /**
     * Handle cancel button click
     */
    function handleCancel() {
        closeModal();
        if (currentResolve) {
            currentResolve(false);
            currentResolve = null;
        }
    }

    /**
     * Handle overlay click (close on background click)
     */
    function handleOverlayClick(e) {
        if (e.target.classList.contains('modal-overlay')) {
            handleCancel();
        }
    }

    /**
     * Handle keyboard events
     */
    function handleKeydown(e) {
        if (!currentModal) return;

        if (e.key === 'Escape') {
            e.preventDefault();
            handleCancel();
        } else if (e.key === 'Enter') {
            e.preventDefault();
            handleConfirm();
        }
    }

    /**
     * Escape HTML to prevent XSS
     */
    function escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    // Convenience functions for common dialogs
    window.confirmDelete = function(itemType) {
        return window.showConfirmModal({
            title: 'Delete ' + itemType,
            message: 'Are you sure you want to delete this ' + itemType.toLowerCase() + '? This action cannot be undone.',
            confirmText: 'Delete',
            cancelText: 'Cancel',
            type: 'danger'
        });
    };

    window.confirmAction = function(message) {
        return window.showConfirmModal({
            title: 'Confirm Action',
            message: message,
            confirmText: 'Yes',
            cancelText: 'No',
            type: 'warning'
        });
    };
})();
