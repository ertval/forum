// Forum Application JavaScript

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
