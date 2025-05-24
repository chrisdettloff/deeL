// Article selection and read status functionality
export function initArticles() {
    const articles = document.querySelectorAll('.article');
    const globalToggleReadButton = document.getElementById('global-toggle-read-button');
    let selectedArticle = null;

    // Handle any article link click to mark as read
    document.addEventListener('click', (event) => {
        // Check if clicked element is a link inside an article
        const articleLink = event.target.closest('.article a');
        if (articleLink) {
            const article = articleLink.closest('.article');
            if (article && article.dataset.read === 'false') {
                // Mark as read before opening the link
                markArticleAsRead(article.dataset.link);
            }
            return; // Let the link work normally
        }
    });

    // Use event delegation for article selection
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

    function markArticleAsRead(itemLink) {
        // Send async request to mark as read without redirecting
        fetch('/toggle-read', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/x-www-form-urlencoded',
            },
            body: `link=${encodeURIComponent(itemLink)}`
        }).then(response => {
            if (response.ok) {
                // Update the article's visual state immediately
                const articles = document.querySelectorAll('.article');
                const article = Array.from(articles).find(a => a.dataset.link === itemLink);
                if (article) {
                    article.classList.remove('unread');
                    article.classList.add('read');
                    article.dataset.read = 'true';
                    
                    // Update toggle button text if this article is selected
                    if (article === selectedArticle) {
                        updateGlobalToggleButtonText();
                    }
                }
            }
        }).catch(error => {
            console.error('Error marking article as read:', error);
        });
    }
}
