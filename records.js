/**
 * Initialize the medical records page with role-specific messages and real-time updates.
 * @param {string} token - JWT token for authentication
 * @param {number} userId - User ID
 * @param {string} role - User role (admin, doctor, patient)
 * @param {string} socketUrl - SocketIO URL
 */
export async function initializeRecordsPage(token, userId, role, socketUrl) {
    const recordsContent = document.getElementById('recordsContent');

    async function fetchRecords() {
        try {
            const data = await fetchWithRetry(`/api/patients/medical-history/${userId}`, {
                method: 'GET',
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'X-CSRF-Token': localStorage.getItem('csrf_token')
                }
            });
            updateRecordsContent(data, role);
        } catch (error) {
            displayError(role, "records retrieval");
        }
    }

    function updateRecordsContent(data, role) {
        if (!data || data.error) {
            displayError(role, "records retrieval");
            return;
        }
        recordsContent.innerHTML = data.records.map(record => `
            <div class="card mb-3" tabindex="0" aria-label="Medical record for ${record.name}">
                <div class="card-body">
                    <h5 class="card-title">${sanitizeHTML(record.name)}</h5>
                    <p>Condition: ${sanitizeHTML(record.medical_condition)}</p>
                    <p>Date: ${sanitizeHTML(new Date(record.date_of_admission).toLocaleString())}</p>
                    <p>Verified: ${record.verified ? '(Verified: ✔)' : 'Not Verified'}</p>
                </div>
            </div>
        `).join('');
        displaySuccess(role, "records retrieval");
    }

    
}