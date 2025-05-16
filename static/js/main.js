document.addEventListener('DOMContentLoaded', () => {
    const themeToggle = document.getElementById('theme-toggle');
    const htmlElement = document.documentElement;
    
    // Toggle theme on button click
    if (themeToggle) {
        themeToggle.addEventListener('click', () => {
            const currentTheme = htmlElement.getAttribute('data-theme');
            const newTheme = currentTheme === 'light' ? 'dark' : 'light';
            
            htmlElement.setAttribute('data-theme', newTheme);
            localStorage.setItem('theme', newTheme);
        });
    }
    
    // Mobile sidebar toggle
    const sidebarToggle = document.getElementById('sidebar-toggle');
    const sidebar = document.getElementById('sidebar');
    
    if (sidebarToggle && sidebar) { // Added null check for sidebar as well
        sidebarToggle.addEventListener('click', () => {
            sidebar.classList.toggle('open');
        });
    }
    
    // Close sidebar when clicking outside (for mobile)
    document.addEventListener('click', (e) => {
        if (sidebar && sidebarToggle && // Ensure sidebar and sidebarToggle exist
            window.innerWidth <= 768 && 
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
                if (globalToggleReadButton) { // Null check
                    globalToggleReadButton.style.display = 'inline-flex'; // Show button
                }
                updateGlobalToggleButtonText();
            } else {
                selectedArticle = null;
                if (globalToggleReadButton) { // Null check
                    globalToggleReadButton.style.display = 'none'; // Hide button
                }
            }
        });
    });

    function updateGlobalToggleButtonText() {
        if (selectedArticle && globalToggleReadButton) { // Null check
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
