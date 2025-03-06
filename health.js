/**
 * Initialize the health monitoring page with role-specific messages and real-time updates.
 * @param {string} token - JWT token for authentication
 * @param {number} userId - User ID
 * @param {string} role - User role (admin, doctor, patient)
 * @param {string} socketUrl - SocketIO URL
 */
export async function initializeHealthPage(token, userId, role, socketUrl) {
    const healthContent = document.getElementById('healthContent');

    async function fetchHealthData() {
        try {
            const data = await fetchWithRetry(`/api/patients/wearable/${userId}`, {
                method: 'GET',
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'X-CSRF-Token': localStorage.getItem('csrf_token')
                }
            });
            updateHealthContent(data, role);
        } catch (error) {
            displayError(role, "health data retrieval");
        }
    }

    function updateHealthContent(data, role) {
        if (!data || data.error) {
            displayError(role, "health data retrieval");
            return;
        }
        healthContent.innerHTML = `
            <div class="card mb-3" tabindex="0" aria-label="Health Data for User ${userId}">
                <div class="card-body">
                    <h5 class="card-title">Health Data</h5>
                    <p>Heart Rate: ${data.wearable_data.heart_rate} bpm</p>
                    <p>Steps: ${data.wearable_data.steps}</p>
                    <p>Verified: ${data.verified ? '(Verified: ✔)' : 'Not Verified'}</p>
                </div>
            </div>
        `;
        displaySuccess(role, "health data retrieval");
    }

    
}