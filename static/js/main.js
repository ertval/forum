// Forum Application JavaScript
'use strict';

// User Menu Dropdown Toggle
// The script is loaded at the end of body, so DOM is already ready
(function() {
    const userMenuBtn = document.getElementById('user-menu-btn');
    const userMenuDropdown = document.getElementById('user-menu-dropdown');
    const userMenuContainer = document.querySelector('.user-menu-container');

    if (userMenuBtn && userMenuDropdown) {
        // Toggle dropdown on button click and set clicked state
        userMenuBtn.addEventListener('click', function(event) {
            event.stopPropagation();
            const opened = userMenuDropdown.classList.toggle('show');
            userMenuBtn.classList.toggle('active', opened);
            userMenuBtn.setAttribute('aria-expanded', opened ? 'true' : 'false');
        });

        // Close dropdown when clicking outside
        document.addEventListener('click', function(event) {
            if (userMenuContainer && !userMenuContainer.contains(event.target)) {
                userMenuDropdown.classList.remove('show');
                userMenuBtn.classList.remove('active');
                userMenuBtn.setAttribute('aria-expanded', 'false');
            }
        });

        // Close dropdown when pressing Escape
        document.addEventListener('keydown', function(event) {
            if (event.key === 'Escape') {
                userMenuDropdown.classList.remove('show');
                userMenuBtn.classList.remove('active');
                userMenuBtn.setAttribute('aria-expanded', 'false');
            }
        });
    }
})();

// Notification Badge - Fetch unread count and update badge in user card/dropdown
(function() {
    const notificationBadges = document.querySelectorAll('[data-notification-badge]');
    if (!notificationBadges.length) {
        return;
    }

    function updateBadges(unreadCount) {
        const safeCount = Number.isFinite(unreadCount) && unreadCount > 0 ? unreadCount : 0;
        const displayCount = safeCount > 99 ? '99+' : String(safeCount);

        notificationBadges.forEach(function(badge) {
            if (safeCount > 0) {
                badge.textContent = displayCount;
                badge.hidden = false;
                badge.setAttribute('aria-label', 'Unread notifications: ' + displayCount);
                return;
            }

            badge.textContent = '0';
            badge.hidden = true;
            badge.removeAttribute('aria-label');
        });
    }

    fetch('/api/notifications', {
        headers: {
            Accept: 'application/json'
        }
    })
        .then(function(response) {
            if (!response.ok) {
                throw new Error('Failed to load notifications');
            }
            return response.json();
        })
        .then(function(payload) {
            if (typeof payload.unread_count === 'number') {
                updateBadges(payload.unread_count);
                return;
            }

            if (!Array.isArray(payload.notifications)) {
                updateBadges(0);
                return;
            }

            const unreadCount = payload.notifications.reduce(function(total, notification) {
                if (notification && notification.is_read === false) {
                    return total + 1;
                }
                return total;
            }, 0);

            updateBadges(unreadCount);
        })
        .catch(function() {
            updateBadges(0);
        });
})();

// Clickable Card - Makes entire post card clickable (except links and buttons)
// Using event delegation for reliability with dynamically loaded content
(function() {
    document.addEventListener('click', function(event) {
        // Find the closest clickable-card ancestor
        const card = event.target.closest('.clickable-card');
        if (!card) return;
        
        // Don't navigate if clicking on a link, button, or category tag
        const target = event.target;
        if (target.tagName === 'A' || 
            target.tagName === 'BUTTON' || 
            target.closest('a') || 
            target.closest('button') ||
            target.classList.contains('category-tag') ||
            target.classList.contains('category-tag-compact')) {
            return;
        }
        
        // Navigate to the post detail page
        const href = card.getAttribute('data-href');
        if (href) {
            window.location.href = href;
        }
    });
})();
