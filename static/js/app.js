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
