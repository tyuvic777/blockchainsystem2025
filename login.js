/**
 * Initialize the login page with role-specific message handling.
 * @param {string} socketUrl - SocketIO URL
 */
export function initializeLoginPage(socketUrl) {
    const loginForm = document.getElementById('loginForm');

    loginForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        const email = document.getElementById('email').value;
        const password = document.getElementById('password').value;
        const role = document.getElementById('role').value;

        try {
            const response = await fetchWithRetry('/login', {
                method: 'POST',
                headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
                body: new URLSearchParams({ email, password, role })
            });
            localStorage.setItem('token', response.token);
            localStorage.setItem('user_id', response.user_id);
            localStorage.setItem('role', role);
            localStorage.setItem('name', response.name || email.split('@')[0]);
            window.location.href = '/';
        } catch (error) {
            displayError(role, "login");
        }
    });

    function displayError(role, feature) {
        const messages = {
            'admin': `Sorry, Admin, we couldn’t process your ${feature} request. Please try again or contact support.`,
            'doctor': `Oops, Doctor, we encountered an issue with your ${feature}. Please try again later or reach out to support.`,
            'patient': `Sorry, ${document.getElementById('email').value.split('@')[0] || 'Patient'}, we couldn’t complete your ${feature} request. Please try again or contact our support team.`
        };
        const alert = document.createElement('div');
        alert.className = 'alert alert-danger mt-3 text-center';
        alert.setAttribute('role', 'alert');
        alert.setAttribute('aria-live', 'polite');
        alert.textContent = messages[role] || messages['patient'];
        document.querySelector('main').appendChild(alert);
    }

    
}