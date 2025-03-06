/**
 * Initialize the appointments page with role-specific messages and real-time updates.
 * @param {string} token - JWT token for authentication
 * @param {number} userId - User ID
 * @param {string} role - User role (admin, doctor, patient)
 * @param {string} socketUrl - SocketIO URL
 */
export async function initializeAppointmentsPage(token, userId, role, socketUrl) {
    const appointmentsList = document.getElementById('appointmentsList');

    async function fetchAppointments() {
        try {
            const data = await fetchWithRetry(`/api/appointments/patient/${userId}`, {
                method: 'GET',
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'X-CSRF-Token': localStorage.getItem('csrf_token')
                }
            });
            updateAppointmentsList(data, role);
        } catch (error) {
            displayError(role, "appointments retrieval");
        }
    }

    function updateAppointmentsList(data, role) {
        if (!data || data.error) {
            displayError(role, "appointments retrieval");
            return;
        }
        appointmentsList.innerHTML = '';
        data.appointments.forEach(appointment => {
            const card = document.createElement('div');
            card.className = 'card mb-3';
            card.tabIndex = '0';
            card.setAttribute('aria-label', `Appointment for ${sanitizeHTML(appointment.patient_name)} with ${sanitizeHTML(appointment.doctor_name)}`);
            card.innerHTML = `
                <div class="card-body">
                    <h5 class="card-title">${sanitizeHTML(appointment.patient_name)} - ${sanitizeHTML(appointment.doctor_name)}</h5>
                    <p>Date: ${sanitizeHTML(new Date(appointment.date).toLocaleString())}</p>
                    <p>Status: ${sanitizeHTML(appointment.status)} ${appointment.verified ? '(Verified: ✔)' : ''}</p>
                    <p>Comment: ${sanitizeHTML(appointment.comment || 'N/A')}</p>
                </div>
            `;
            appointmentsList.appendChild(card);
        });
        displaySuccess(role, "appointments list retrieval");
    }

    function displayError(role, feature) {
        const messages = {
            'admin': `Sorry, Admin, we couldn’t process your ${feature} request. Please try again or contact support.`,
            'doctor': `Oops, Doctor, we encountered an issue with your ${feature}. Please try again later or reach out to support.`,
            'patient': `Sorry, ${localStorage.getItem('name') || 'Patient'}, we couldn’t complete your ${feature} request. Please try again or contact our support team.`
        };
        const alert = document.createElement('div');
        alert.className = 'alert alert-danger mt-3';
        alert.setAttribute('role', 'alert');
        alert.setAttribute('aria-live', 'polite');
        alert.textContent = messages[role];
        document.querySelector('main').appendChild(alert);
    }

    function displaySuccess(role, feature) {
        const messages = {
            'admin': `Thank you, Admin! Your action on ${feature} has been completed successfully.`,
            'doctor': `Great job, Doctor! Your update to ${feature} was successful.`,
            'patient': `Thank you, ${localStorage.getItem('name') || 'Patient'}! Your ${feature} has been updated successfully.`
        };
        const alert = document.createElement('div');
        alert.className = 'alert alert-success mt-3';
        alert.setAttribute('role', 'alert');
        alert.setAttribute('aria-live', 'polite');
        alert.textContent = messages[role];
        document.querySelector('main').appendChild(alert);
        setTimeout(() => alert.remove(), 5000); // Auto-dismiss after 5 seconds
    }

    // SocketIO for real-time updates
    const socket = io(socketUrl);
    socket.on('connect', () => console.log('Connected to MediNet SocketIO'));
    socket.on('appointmentUpdate', (data) => {
        fetchAppointments().then(() => displaySuccess(role, "appointment list update"));
    });

    fetchAppointments();
}

/**
 * Fetch with retry logic for robust API calls.
 * @param {string} url - API endpoint
 * @param {Object} options - Fetch options
 * @param {number} retries - Number of retry attempts
 * @returns {Promise} - JSON response
 */
async function fetchWithRetry(url, options, retries = 3) {
    for (let i = 0; i < retries; i++) {
        try {
            const response = await fetch(url, options);
            if (!response.ok) {
                if (response.status === 429) {
                    console.warn(`Rate limit exceeded for ${url}, retrying in ${1000 * Math.pow(2, i)}ms`);
                }
                throw new Error((await response.json()).error || `Oops! We couldn’t complete your request. Please try again later, or contact support for help.`);
            }
            return response.json();
        } catch (e) {
            if (i === retries - 1) throw e;
            await new Promise(resolve => setTimeout(resolve, 1000 * Math.pow(2, i)));
        }
    }
}

/**
 * Sanitize HTML to prevent XSS.
 * @param {string} html - HTML input
 * @returns {string} - Sanitized string
 */
function sanitizeHTML(html) {
    return sanitizeHtml(html, { allowedTags: [], allowedAttributes: {} });
}