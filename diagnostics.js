/**
 * Initialize the diagnostics page with role-specific messages and real-time updates.
 * @param {string} token - JWT token for authentication
 * @param {number} userId - User ID
 * @param {string} role - User role (admin, doctor, patient)
 * @param {string} socketUrl - SocketIO URL
 */
export async function initializeDiagnosticsPage(token, userId, role, socketUrl) {
    const diagnosticsContent = document.getElementById('diagnosticsContent');

    async function fetchDiagnostics() {
        try {
            const data = await fetchWithRetry(`/api/patients/medical-history/${userId}`, {
                method: 'GET',
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'X-CSRF-Token': localStorage.getItem('csrf_token')
                }
            });
            updateDiagnosticsContent(data, role);
        } catch (error) {
            displayError(role, "diagnostics retrieval");
        }
    }

    function updateDiagnosticsContent(data, role) {
        if (!data || data.error) {
            displayError(role, "diagnostics retrieval");
            return;
        }
        diagnosticsContent.innerHTML = data.records.map(diagnostic => `
            <div class="card mb-3" tabindex="0" aria-label="Diagnostic for ${diagnostic.name}">
                <div class="card-body">
                    <h5 class="card-title">${sanitizeHTML(diagnostic.name)}</h5>
                    <p>Condition: ${sanitizeHTML(diagnostic.medical_condition)}</p>
                    <p>Date: ${sanitizeHTML(new Date(diagnostic.date_of_admission).toLocaleString())}</p>
                    <p>Verified: ${diagnostic.verified ? '(Verified: ✔)' : 'Not Verified'}</p>
                </div>
            </div>
        `).join('');
        displaySuccess(role, "diagnostics retrieval");
    }

    
}