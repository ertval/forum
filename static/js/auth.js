// Authentication JavaScript functions for login and register forms
'use strict';

document.addEventListener('DOMContentLoaded', function() {
    // Handle login form submission
    const loginForm = document.getElementById('login-form');
    if (loginForm) {
        loginForm.addEventListener('submit', async function(e) {
            e.preventDefault();

            const formErrors = document.getElementById('form-errors');
            if (formErrors) formErrors.innerHTML = '';

            const formData = new FormData(e.target);
            const data = {
                email: formData.get('email'),
                password: formData.get('password')
            };

            try {
                await window.api.request('/api/auth/login', {
                    method: 'POST',
                    body: JSON.stringify(data)
                });
                // Redirect to home page after successful login
                window.location.href = '/';
            } catch (error) {
                // SECURITY: Escape error message to prevent XSS
                if (formErrors) formErrors.innerHTML = `<p class="error">${window.escapeHtml(error.message || 'Login failed')}</p>`;
            }
        });
    }

    // Handle register form submission
    const registerForm = document.getElementById('register-form');
    if (registerForm) {
        registerForm.addEventListener('submit', async function(e) {
            e.preventDefault();

            const formErrors = document.getElementById('form-errors');
            if (formErrors) formErrors.innerHTML = '';

            const formData = new FormData(e.target);
            const data = {
                username: formData.get('username'),
                email: formData.get('email'),
                password: formData.get('password')
            };

            try {
                await window.api.request('/api/auth/register', {
                    method: 'POST',
                    body: JSON.stringify(data)
                });
                // Redirect to home page after successful registration
                window.location.href = '/';
            } catch (error) {
                // SECURITY: Escape error message to prevent XSS
                if (formErrors) formErrors.innerHTML = `<p class="error">${window.escapeHtml(error.message || 'Registration failed')}</p>`;
            }
        });
    }
});