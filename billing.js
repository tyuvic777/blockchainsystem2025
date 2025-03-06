/**
 * Initialize the billing page with role-specific messages and real-time updates.
 * @param {string} token - JWT token for authentication
 * @param {number} userId - User ID
 * @param {string} role - User role (admin, doctor, patient)
 * @param {string} socketUrl - SocketIO URL
 */
export async function initializeBillingPage(token, userId, role, socketUrl) {
    const billingContent = document.getElementById('billingContent');

    async function fetchBilling() {
        try {
            const data = await fetchWithRetry(`/api/patients/billing/history/${userId}`, {
                method: 'GET',
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'X-CSRF-Token': localStorage.getItem('csrf_token')
                }
            });
            updateBillingContent(data, role);
        } catch (error) {
            displayError(role, "billing retrieval");
        }
    }

    function updateBillingContent(data, role) {
        if (!data || data.error) {
            displayError(role, "billing retrieval");
            return;
        }
        billingContent.innerHTML = data.billing.map(bill => `
            <div class="card mb-3" tabindex="0" aria-label="Billing for ${bill.user_id}">
                <div class="card-body">
                    <h5 class="card-title">Bill for ${bill.user_id}</h5>
                    <p>Amount: $${bill.amount}</p>
                    <p>Verified: ${bill.verified ? '(Verified: ✔)' : 'Not Verified'}</p>
                </div>
            </div>
        `).join('');
        displaySuccess(role, "billing retrieval");
    }

   
}