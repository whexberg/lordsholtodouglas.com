// Theme handler — runs immediately to prevent FOUC
(function() {
    var theme = localStorage.getItem('theme');
    if (theme === 'dark' || (!theme && window.matchMedia('(prefers-color-scheme: dark)').matches)) {
        document.documentElement.classList.add('dark');
    }
})();

document.addEventListener('DOMContentLoaded', function() {
    // Theme toggle
    function updateIcons() {
        var isDark = document.documentElement.classList.contains('dark');
        document.querySelectorAll('.theme-icon-sun').forEach(function(el) {
            el.classList.toggle('hidden', !isDark);
        });
        document.querySelectorAll('.theme-icon-moon').forEach(function(el) {
            el.classList.toggle('hidden', isDark);
        });
    }

    function toggleTheme() {
        document.documentElement.classList.toggle('dark');
        var isDark = document.documentElement.classList.contains('dark');
        localStorage.setItem('theme', isDark ? 'dark' : 'light');
        updateIcons();
    }

    document.querySelectorAll('[data-theme-toggle]').forEach(function(btn) {
        btn.addEventListener('click', toggleTheme);
    });

    updateIcons();

    // Mobile menu
    var toggle = document.getElementById('mobile-menu-toggle');
    var menu = document.getElementById('mobile-menu');
    if (toggle && menu) {
        toggle.addEventListener('click', function() {
            menu.classList.toggle('hidden');
        });
    }
});
