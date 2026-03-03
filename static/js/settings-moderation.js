// Moderation controls for settings page
'use strict';

(function() {
    function text(value) {
        return window.escapeHtml(value || '');
    }

    function renderMessage(container, message, isError) {
        if (!container) return;
        container.innerHTML = `<p><strong>${isError ? 'Error' : 'Success'}:</strong> ${text(message)}</p>`;
    }

    function decisionControl(kind, id) {
        return `
            <div class="form-group">
                <select data-${kind}-status="${id}">
                    <option value="">Choose decision</option>
                    <option value="approved">Approve</option>
                    <option value="denied">Deny</option>
                </select>
                <input type="text" data-${kind}-response="${id}" placeholder="Optional response">
                <button type="button" class="btn btn-primary" data-${kind}-submit="${id}">Submit</button>
            </div>
        `;
    }

    async function loadReports(listEl, role) {
        if (!listEl || (role !== 'moderator' && role !== 'admin')) return;

        try {
            const data = await window.api.request('/api/moderation/reports');
            const reports = Array.isArray(data?.reports) ? data.reports : [];
            if (!reports.length) {
                listEl.innerHTML = '<p>No moderation reports found.</p>';
                return;
            }

            listEl.innerHTML = reports.map((report) => {
                const controls = report.status === 'pending'
                    ? `
                        <div class="form-group">
                            <select data-report-status="${report.id}">
                                <option value="">Choose decision</option>
                                <option value="reviewed">Reviewed</option>
                                <option value="resolved">Resolved</option>
                            </select>
                            <input type="text" data-report-response="${report.id}" placeholder="Optional response">
                            <button type="button" class="btn btn-primary" data-report-submit="${report.id}">Submit</button>
                        </div>
                      `
                    : '<p><em>Already reviewed.</em></p>';

                return `
                    <article class="comment">
                        <div class="comment-content">
                            <p><strong>ID:</strong> ${text(report.id)}</p>
                            <p><strong>Target:</strong> ${text(report.target_type)} / ${text(report.target_id)}</p>
                            <p><strong>Reason:</strong> ${text(report.reason)}</p>
                            <p><strong>Status:</strong> ${text(report.status)}</p>
                            ${controls}
                        </div>
                    </article>
                `;
            }).join('');
        } catch (err) {
            listEl.innerHTML = `<p><strong>Error:</strong> ${text(err.message || 'Failed to load reports')}</p>`;
        }
    }

    async function loadModeratorRequests(listEl) {
        if (!listEl) return;

        try {
            const data = await window.api.request('/api/moderation/requests');
            const requests = Array.isArray(data?.requests) ? data.requests : [];
            if (!requests.length) {
                listEl.innerHTML = '<p>No moderator role requests found.</p>';
                return;
            }

            listEl.innerHTML = requests.map((request) => {
                const controls = request.status === 'pending'
                    ? decisionControl('request', request.id)
                    : '<p><em>Already reviewed.</em></p>';

                return `
                    <article class="comment">
                        <div class="comment-content">
                            <p><strong>ID:</strong> ${text(request.id)}</p>
                            <p><strong>Requester:</strong> ${text(request.requester_id)}</p>
                            <p><strong>Message:</strong> ${text(request.message)}</p>
                            <p><strong>Status:</strong> ${text(request.status)}</p>
                            ${controls}
                        </div>
                    </article>
                `;
            }).join('');
        } catch (err) {
            listEl.innerHTML = `<p><strong>Error:</strong> ${text(err.message || 'Failed to load requests')}</p>`;
        }
    }

    async function loadUsers(listEl) {
        if (!listEl) return;

        try {
            const data = await window.api.request('/api/users');
            const users = Array.isArray(data?.users) ? data.users : [];
            if (!users.length) {
                listEl.innerHTML = '<p>No users found.</p>';
                return;
            }

            listEl.innerHTML = users.map((user) => {
                return `
                    <article class="comment">
                        <div class="comment-content">
                            <p><strong>User:</strong> ${text(user.username)} (${text(user.id)})</p>
                            <p><strong>Current role:</strong> ${text(user.role)}</p>
                            <div class="form-group">
                                <select data-user-role="${user.id}">
                                    <option value="user">user</option>
                                    <option value="moderator">moderator</option>
                                    <option value="admin">admin</option>
                                </select>
                                <button type="button" class="btn btn-primary" data-user-role-submit="${user.id}">Apply Role</button>
                            </div>
                        </div>
                    </article>
                `;
            }).join('');

            users.forEach((user) => {
                const select = listEl.querySelector(`select[data-user-role="${user.id}"]`);
                if (select) {
                    select.value = user.role;
                }
            });
        } catch (err) {
            listEl.innerHTML = `<p><strong>Error:</strong> ${text(err.message || 'Failed to load users')}</p>`;
        }
    }

    document.addEventListener('DOMContentLoaded', function() {
        const root = document.getElementById('moderation-settings');
        if (!root) return;

        const role = root.dataset.userRole || '';

        const requestForm = document.getElementById('moderator-request-form');
        const requestFeedback = document.getElementById('moderator-request-feedback');
        if (requestForm) {
            requestForm.addEventListener('submit', async function(event) {
                event.preventDefault();
                try {
                    const message = (document.getElementById('moderator-request-message') || {}).value || '';
                    await window.api.request('/api/moderation/requests', {
                        method: 'POST',
                        body: JSON.stringify({ message })
                    });
                    renderMessage(requestFeedback, 'Moderator role request submitted.', false);
                } catch (err) {
                    renderMessage(requestFeedback, err.message || 'Failed to submit request.', true);
                }
            });
        }

        const reportsList = document.getElementById('moderation-reports-list');
        const refreshReportsBtn = document.getElementById('refresh-reports-btn');
        if (refreshReportsBtn) {
            refreshReportsBtn.addEventListener('click', function() {
                loadReports(reportsList, role);
            });
            loadReports(reportsList, role);
        }

        const requestsList = document.getElementById('moderator-requests-list');
        const refreshRequestsBtn = document.getElementById('refresh-role-requests-btn');
        if (refreshRequestsBtn) {
            refreshRequestsBtn.addEventListener('click', function() {
                loadModeratorRequests(requestsList);
            });
            loadModeratorRequests(requestsList);
        }

        const usersList = document.getElementById('role-management-list');
        const refreshUsersBtn = document.getElementById('refresh-users-btn');
        if (refreshUsersBtn) {
            refreshUsersBtn.addEventListener('click', function() {
                loadUsers(usersList);
            });
            loadUsers(usersList);
        }

        document.body.addEventListener('click', async function(event) {
            const reportBtn = event.target.closest('[data-report-submit]');
            if (reportBtn) {
                const reportID = reportBtn.getAttribute('data-report-submit');
                const statusEl = document.querySelector(`select[data-report-status="${reportID}"]`);
                const responseEl = document.querySelector(`input[data-report-response="${reportID}"]`);
                const status = statusEl ? statusEl.value : '';
                const response = responseEl ? responseEl.value : '';
                if (!status) return;

                try {
                    await window.api.request(`/api/moderation/reports/${encodeURIComponent(reportID)}`, {
                        method: 'PUT',
                        body: JSON.stringify({ status, response })
                    });
                    loadReports(reportsList, role);
                } catch (err) {
                    alert(err.message || 'Failed to review report');
                }
                return;
            }

            const requestBtn = event.target.closest('[data-request-submit]');
            if (requestBtn) {
                const requestID = requestBtn.getAttribute('data-request-submit');
                const statusEl = document.querySelector(`select[data-request-status="${requestID}"]`);
                const responseEl = document.querySelector(`input[data-request-response="${requestID}"]`);
                const status = statusEl ? statusEl.value : '';
                const response = responseEl ? responseEl.value : '';
                if (!status) return;

                try {
                    await window.api.request(`/api/moderation/requests/${encodeURIComponent(requestID)}`, {
                        method: 'PUT',
                        body: JSON.stringify({ status, response })
                    });
                    loadModeratorRequests(requestsList);
                    loadUsers(usersList);
                } catch (err) {
                    alert(err.message || 'Failed to review moderator request');
                }
                return;
            }

            const userRoleBtn = event.target.closest('[data-user-role-submit]');
            if (userRoleBtn) {
                const userID = userRoleBtn.getAttribute('data-user-role-submit');
                const roleEl = document.querySelector(`select[data-user-role="${userID}"]`);
                const newRole = roleEl ? roleEl.value : '';
                if (!newRole) return;

                try {
                    await window.api.request(`/api/users/${encodeURIComponent(userID)}/role`, {
                        method: 'PUT',
                        body: JSON.stringify({ role: newRole })
                    });
                    loadUsers(usersList);
                } catch (err) {
                    alert(err.message || 'Failed to update user role');
                }
            }
        });
    });
})();
