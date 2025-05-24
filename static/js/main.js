// Import modular components
import { initThemeToggle } from './components/theme.js';
import { initSidebar } from './components/sidebar.js';
import { initArticles } from './components/articles.js';
import { initFilters } from './components/filters.js';

// Initialize all components when the DOM is ready
document.addEventListener('DOMContentLoaded', () => {
    initThemeToggle();
    initSidebar();
    initArticles();
    initFilters();
});
