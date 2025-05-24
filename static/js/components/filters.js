// Filter dropdown functionality
export function initFilters() {
    const filterDropdown = document.getElementById('filter-dropdown');
    
    if (filterDropdown) {
        filterDropdown.addEventListener('change', (event) => {
            const selectedFilter = event.target.value;
            const currentURL = new URL(window.location);
            
            // Update the filter parameter
            currentURL.searchParams.set('filter', selectedFilter);
            
            // Navigate to the new URL
            window.location.href = currentURL.toString();
        });
    }
}
