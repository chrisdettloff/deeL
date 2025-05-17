// Article selection and read status functionality
export function initArticles() {
    const articles = document.querySelectorAll('.article');
    const globalToggleReadButton = document.getElementById('global-toggle-read-button');
    let selectedArticle = null;

    // Use event delegation for better performance
    document.addEventListener('click', (event) => {
        // Find the closest article if clicked within one
        const article = event.target.closest('.article');
        
        // If not clicking an article or clicking a link inside article, do nothing
        if (!article || event.target.closest('a')) {
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
            if (globalToggleReadButton) {
                globalToggleReadButton.style.display = 'inline-flex'; // Show button
            }
            updateGlobalToggleButtonText();
        } else {
            selectedArticle = null;
            if (globalToggleReadButton) {
                globalToggleReadButton.style.display = 'none'; // Hide button
            }
        }
    });

    // Add the event listener to the toggle button
    if (globalToggleReadButton) {
        globalToggleReadButton.addEventListener('click', () => {
            if (selectedArticle) {
                const itemLink = selectedArticle.dataset.link;
                if (!itemLink) {
                    console.error('Selected article does not have a data-link attribute.');
                    return;
                }

                submitReadStatusForm(itemLink);
            }
        });
    }
    
    function updateGlobalToggleButtonText() {
        if (selectedArticle && globalToggleReadButton) {
            const isRead = selectedArticle.dataset.read === 'true' || selectedArticle.classList.contains('read');
            globalToggleReadButton.textContent = isRead ? 'Mark as Unread' : 'Mark as Read';
        }
    }
    
    function submitReadStatusForm(itemLink) {
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
}
