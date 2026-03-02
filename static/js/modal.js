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

            // Build modal structure with DOM methods (avoids innerHTML with user data)
            const modal = document.createElement('div');
            modal.className = 'modal';
            modal.setAttribute('role', 'dialog');
            modal.setAttribute('aria-modal', 'true');
            modal.setAttribute('aria-labelledby', 'modal-title');

            const header = document.createElement('div');
            header.className = 'modal-header';
            const iconSpan = document.createElement('span');
            iconSpan.className = `modal-icon ${type}`;
            iconSpan.textContent = icon;
            const titleEl = document.createElement('h3');
            titleEl.id = 'modal-title';
            titleEl.textContent = options.title || 'Confirm Action';
            header.appendChild(iconSpan);
            header.appendChild(titleEl);

            const body = document.createElement('div');
            body.className = 'modal-body';
            const msgP = document.createElement('p');
            msgP.textContent = options.message || 'Are you sure you want to proceed?';
            body.appendChild(msgP);

            const actionsEl = document.createElement('div');
            actionsEl.className = 'modal-actions';
            const cancelBtn = document.createElement('button');
            cancelBtn.className = 'btn btn-cancel';
            cancelBtn.id = 'modal-cancel';
            cancelBtn.type = 'button';
            cancelBtn.textContent = options.cancelText || 'Cancel';
            const confirmBtn = document.createElement('button');
            confirmBtn.className = 'btn btn-confirm-danger';
            confirmBtn.id = 'modal-confirm';
            confirmBtn.type = 'button';
            confirmBtn.textContent = options.confirmText || 'Confirm';
            actionsEl.appendChild(cancelBtn);
            actionsEl.appendChild(confirmBtn);

            modal.appendChild(header);
            modal.appendChild(body);
            modal.appendChild(actionsEl);
            overlay.appendChild(modal);

            // Add to document
            document.body.appendChild(overlay);
            currentModal = overlay;

            // Trigger reflow then show
            overlay.offsetHeight;
            overlay.classList.add('show');

            // Focus confirm button (already created as DOM node above)
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
