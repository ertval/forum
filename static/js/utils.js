// Shared JavaScript Utilities
// This module provides common utilities used across the forum application
'use strict';

/**
 * Escape HTML entities to prevent XSS attacks.
 * This function should be used whenever displaying user-generated content.
 * @param {string|number|null|undefined} text - The text to escape
 * @returns {string} - The escaped HTML-safe string
 */
window.escapeHtml = function(text) {
    if (text === null || text === undefined) {
        return '';
    }
    const div = document.createElement('div');
    div.textContent = String(text);
    return div.innerHTML;
};

/**
 * Constants used across the application
 */
window.FORUM_CONSTANTS = Object.freeze({
    BATCH_SIZE: 20,
    MODAL_TRANSITION_MS: 200
});

/**
 * Centralized API request utility.
 * Provides consistent error handling and JSON parsing.
 * @param {string} url - The API endpoint URL
 * @param {Object} options - Fetch options (method, body, etc.)
 * @returns {Promise<Object>} - Parsed JSON response
 * @throws {Error} - Throws with error message from server or generic message
 */
window.api = {
    async request(url, options = {}) {
        // Set default headers for JSON requests (unless FormData is being sent)
        if (!(options.body instanceof FormData)) {
            options.headers = {
                'Content-Type': 'application/json',
                ...options.headers
            };
        }
        
        // Include credentials for cookie-based auth
        options.credentials = options.credentials || 'include';
        
        const response = await fetch(url, options);
        
        // Handle empty responses (e.g., 204 No Content)
        const contentType = response.headers.get('content-type');
        let data = null;
        if (contentType && contentType.includes('application/json')) {
            try {
                data = await response.json();
            } catch (parseError) {
                console.error('JSON parse error:', parseError);
                throw new Error('Invalid response from server');
            }
        }
        
        if (!response.ok) {
            const err = new Error(data?.error || `Server error (${response.status})`);
            err.status = response.status;
            throw err;
        }
        
        return data;
    }
};

/**
 * Show an error message in a container element.
 * @param {string} message - The error message to display
 * @param {string} containerId - The ID of the container element (default: 'form-errors')
 */
window.showError = function(message, containerId = 'form-errors') {
    const container = document.getElementById(containerId);
    if (container) {
        container.innerHTML = `<p class="error">${window.escapeHtml(message)}</p>`;
        container.scrollIntoView({ behavior: 'smooth', block: 'center' });
    }
};

/**
 * Clear error messages from a container.
 * @param {string} containerId - The ID of the container element (default: 'form-errors')
 */
window.clearError = function(containerId = 'form-errors') {
    const container = document.getElementById(containerId);
    if (container) {
        container.innerHTML = '';
    }
};
