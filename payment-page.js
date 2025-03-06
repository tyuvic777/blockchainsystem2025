/**
 * Initialize the payment page with role-specific messages and real-time updates.
 * @param {string} token - JWT token for authentication
 * @param {number} userId - User ID
 * @param {string} role - User role (admin, doctor, patient)
 * @param {string} socketUrl - SocketIO URL
 */
export function initializePaymentPage(token, userId, role, socketUrl) {
    const socket = io(socketUrl);

    socket.on('connect', () => console.log('Connected to SocketIO'));
    socket.on('message', (data) => console.log(data));

    document.getElementById('paymentForm').addEventListener('submit', async (e) => {
        e.preventDefault();
        const data = {
            patient_id: userId,
            amount: document.getElementById('amount').value,
            card_details: {
                number: document.getElementById('cardNumber').value,
                expiry: document.getElementById('expiry').value,
                cvv: document.getElementById('cvv').value
            }
        };
        try {
            const response = await fetch('/api/patients/payment/process', {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(data)
            });
            const result = await response.json();
            if (response.ok) {
                displaySuccess(role, "payment processing", result.transaction_id);
            } else {
                displayError(role, "payment processing", result.error);
            }
        } catch (error) {
            displayError(role, "payment processing", "Network error");
        }
    });

    function displayError(role, feature, errorMessage) {
        const messages = {
            'admin': `Sorry, Admin, we couldn’t process your ${feature} request.`,
            'doctor': `Oops, Doctor, we encountered an issue with your ${feature}.`,
            'patient': `Sorry, ${localStorage.getItem('name') || 'Patient'}, we couldn’t complete your ${feature} request.`
        };
        const message = `${messages[role]} ${errorMessage}`;
        document.getElementById('paymentMessage').innerHTML = `<div class="alert alert-danger">${message}</div>`;
    }

    function displaySuccess(role, feature, transactionId) {
        const messages = {
            'admin': `Thank you, Admin! Your ${feature} has been completed successfully.`,
            'doctor': `Great job, Doctor! Your ${feature} was successful.`,
            'patient': `Thank you, ${localStorage.getItem('name') || 'Patient'}! Your ${feature} has been updated successfully.`
        };
        const message = `${messages[role]} Transaction ID: ${transactionId}`;
        document.getElementById('paymentMessage').innerHTML = `<div class="alert alert-success">${message}</div>`;
    }
}