// Authentication JavaScript functions for login and register forms

document.addEventListener('DOMContentLoaded', function() {
    // Handle login form submission
    const loginForm = document.getElementById('loginForm');
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
                const response = await fetch('/api/auth/login', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify(data)
                });

                if (response.ok) {
                    // Redirect to home page after successful login
                    window.location.href = '/';
                } else {
                    const error = await response.json();
                    // SECURITY: Escape error message to prevent XSS
                    if (formErrors) formErrors.innerHTML = `<p class="error">${window.escapeHtml(error.error || 'Login failed')}</p>`;
                }
            } catch (error) {
                console.error('Login error:', error);
                if (formErrors) formErrors.innerHTML = '<p class="error">An error occurred during login</p>';
            }
        });
    }

    // Handle register form submission
    const registerForm = document.getElementById('registerForm');
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
                const response = await fetch('/api/auth/register', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify(data)
                });

                if (response.ok) {
                    // Redirect to home page after successful registration
                    window.location.href = '/';
                } else {
                    const error = await response.json();
                    // SECURITY: Escape error message to prevent XSS
                    if (formErrors) formErrors.innerHTML = `<p class="error">${window.escapeHtml(error.error || 'Registration failed')}</p>`;
                }
            } catch (error) {
                console.error('Registration error:', error);
                if (formErrors) formErrors.innerHTML = '<p class="error">An error occurred during registration</p>';
            }
        });
    }
});