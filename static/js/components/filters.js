// Filter dropdown functionality
export function initFilters() {
    // New custom dropdown logic
    const filterDropdownToggle = document.querySelector('.filter-dropdown-toggle');
    const filterDropdownMenu = document.getElementById('filter-dropdown-menu');

    if (filterDropdownToggle && filterDropdownMenu) {
        // Keep the logic to close dropdown if clicked outside
        document.addEventListener('click', (event) => {
            if (!filterDropdownToggle.contains(event.target) && !filterDropdownMenu.contains(event.target)) {
                // Ensure elements exist before trying to access classList or querySelector
                if (filterDropdownMenu) {
                    filterDropdownMenu.classList.remove('show');
                }
                if (filterDropdownToggle && filterDropdownToggle.querySelector('.dropdown-arrow')) {
                    filterDropdownToggle.querySelector('.dropdown-arrow').classList.remove('rotate');
                }
            }
        });
    }

    // The old logic for a <select> element can be removed or kept if you plan to use it elsewhere.
    // For this example, I'm assuming the new custom dropdown is the primary filter mechanism.
    // const filterDropdown = document.getElementById('filter-dropdown'); // This was for a select
    // if (filterDropdown) { ... }
}

// Function to toggle the main filter dropdown (used by onclick in HTML)
// This function is correctly called by the onclick attribute in the HTML
window.toggleFilterDropdown = function(event) {
    event.stopPropagation();
    const menu = document.getElementById('filter-dropdown-menu');
    const toggle = event.currentTarget; // This is the .filter-dropdown-toggle div

    if (menu) {
        menu.classList.toggle('show');
    }
    const arrow = toggle.querySelector('.dropdown-arrow');
    if (arrow) {
        arrow.classList.toggle('rotate');
    }
}
