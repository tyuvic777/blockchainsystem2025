/**
 * Initialize the telemedicine page with role-specific messages, real-time updates, and AR.
 * @param {string} token - JWT token for authentication
 * @param {number} userId - User ID
 * @param {string} role - User role (admin, doctor, patient)
 * @param {string} socketUrl - SocketIO URL
 */
export async function initializeTelemedicinePage(token, userId, role, socketUrl) {
    const chatForm = document.getElementById('chatForm');
    const chatMessages = document.getElementById('chatMessages');

    async function sendChat(message) {
        try {
            const data = await fetchWithRetry('/api/telemedicine/chat', {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'X-CSRF-Token': localStorage.getItem('csrf_token'),
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ message: message })
            });
            updateChatMessages(data, role);
        } catch (error) {
            displayError(role, "chat message send");
        }
    }

    chatForm.addEventListener('submit', (e) => {
        e.preventDefault();
        const message = document.getElementById('chatInput').value;
        sendChat(message);
        document.getElementById('chatInput').value = '';
    });

    // AR initialization 
    const arScene = document.getElementById('arScene');
    const webglFallback = document.getElementById('webglFallback');
    if (!AFRAME) {
        arScene.style.display = 'none';
        webglFallback.style.display = 'block';
    }

    // SocketIO for real-time chat
    const socket = io(socketUrl);
    socket.on('connect', () => console.log('Connected to MediNet SocketIO'));
    socket.on('chatUpdate', (data) => {
        updateChatMessages(data, role);
        displaySuccess(role, "chat update");
    });

    function updateChatMessages(data, role) {
        if (!data || data.error) {
            displayError(role, "chat update");
            return;
        }
        chatMessages.innerHTML += `
            <div class="mb-2" tabindex="0" aria-label="${data.from} message: ${sanitizeHTML(data.message)}">
                <strong>${sanitizeHTML(data.from)}:</strong> ${sanitizeHTML(data.message)}
            </div>
        `;
        chatMessages.scrollTop = chatMessages.scrollHeight;
    }

    
}