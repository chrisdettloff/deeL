document.addEventListener('DOMContentLoaded', () => {
    const themeToggle = document.getElementById('theme-toggle');
    const htmlElement = document.documentElement;
    
    // Check for saved theme preference or use device preference
    const savedTheme = localStorage.getItem('theme');
    const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
    
    // Set initial theme
    if (savedTheme) {
        htmlElement.setAttribute('data-theme', savedTheme);
    } else if (prefersDark) {
        htmlElement.setAttribute('data-theme', 'dark');
        localStorage.setItem('theme', 'dark');
    }
    
    // Toggle theme on button click
    themeToggle.addEventListener('click', () => {
        const currentTheme = htmlElement.getAttribute('data-theme');
        const newTheme = currentTheme === 'light' ? 'dark' : 'light';
        
        htmlElement.setAttribute('data-theme', newTheme);
        localStorage.setItem('theme', newTheme);
    });
    
    // Mobile sidebar toggle
    const sidebarToggle = document.getElementById('sidebar-toggle');
    const sidebar = document.getElementById('sidebar');
    
    sidebarToggle.addEventListener('click', () => {
        sidebar.classList.toggle('open');
    });
    
    // Close sidebar when clicking outside (for mobile)
    document.addEventListener('click', (e) => {
        if (window.innerWidth <= 768 && 
            !sidebar.contains(e.target) && 
            !sidebarToggle.contains(e.target) &&
            sidebar.classList.contains('open')) {
            sidebar.classList.remove('open');
        }
    });

    const articles = document.querySelectorAll('.article');
    const globalToggleReadButton = document.getElementById('global-toggle-read-button');
    let selectedArticle = null;

    articles.forEach(article => {
        article.addEventListener('click', (event) => {
            // Prevent click from propagating if the click is on the link itself
            if (event.target.closest('a')) {
                return;
            }

            // Deselect previous article
            if (selectedArticle && selectedArticle !== article) {
                selectedArticle.classList.remove('selected');
            }

            // Toggle selection for current article
            article.classList.toggle('selected');

            if (article.classList.contains('selected')) {
                selectedArticle = article;
                globalToggleReadButton.style.display = 'inline-flex'; // Show button
                updateGlobalToggleButtonText();
            } else {
                selectedArticle = null;
                globalToggleReadButton.style.display = 'none'; // Hide button
            }
        });
    });

    function updateGlobalToggleButtonText() {
        if (selectedArticle) {
            const isRead = selectedArticle.dataset.read === 'true' || selectedArticle.classList.contains('read');
            globalToggleReadButton.textContent = isRead ? 'Mark as Unread' : 'Mark as Read';
        }
    }

    if (globalToggleReadButton) {
        globalToggleReadButton.addEventListener('click', () => {
            if (selectedArticle) {
                const itemLink = selectedArticle.dataset.link;
                if (!itemLink) {
                    console.error('Selected article does not have a data-link attribute.');
                    return;
                }

                const form = document.createElement('form');
                form.method = 'post';
                form.action = '/toggle-read';

                const linkInput = document.createElement('input');
                linkInput.type = 'hidden';
                linkInput.name = 'link';
                linkInput.value = itemLink;
                form.appendChild(linkInput);

                document.body.appendChild(form);
                form.submit();
            }
        });
    }
});
