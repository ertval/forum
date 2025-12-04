// Forum Application JavaScript

// Initialize when DOM is ready
document.addEventListener('DOMContentLoaded', function() {
    // User Menu Dropdown Toggle
    const userMenuBtn = document.getElementById('user-menu-btn');
    const userMenuDropdown = document.getElementById('user-menu-dropdown');
    const userMenuContainer = document.querySelector('.user-menu-container');

    if (userMenuBtn && userMenuDropdown) {
        // Toggle dropdown on button click
        userMenuBtn.addEventListener('click', function(event) {
            event.stopPropagation();
            userMenuDropdown.classList.toggle('show');
        });

        // Close dropdown when clicking outside
        document.addEventListener('click', function(event) {
            if (userMenuContainer && !userMenuContainer.contains(event.target)) {
                userMenuDropdown.classList.remove('show');
            }
        });

        // Close dropdown when pressing Escape
        document.addEventListener('keydown', function(event) {
            if (event.key === 'Escape') {
                userMenuDropdown.classList.remove('show');
            }
        });
    }
});