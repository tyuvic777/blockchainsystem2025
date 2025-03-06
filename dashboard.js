// JavaScript for MediNet dashboard, handling AR, real-time updates, and voice input
/**
 * Initialize the dashboard page with role-specific messages, real-time updates, and AR.
 * @param {string} token - JWT token for authentication
 * @param {number} userId - User ID
 * @param {string} role - User role (admin, doctor, patient)
 * @param {string} socketUrl - SocketIO URL
 */
export function initializeDashboardPage(token, userId, role, socketUrl) {
    // Initialize Socket.IO connection
    const socket = io(socketUrl);

    // Ensure DOM is fully loaded before interacting with it
    document.addEventListener('DOMContentLoaded', () => {
        const scene = document.querySelector('a-scene');

        // Log Socket.IO connection and messages
        socket.on('connect', () => console.log('Connected to SocketIO'));
        socket.on('message', (data) => console.log(data));

        // Voice input functionality with enhanced browser compatibility and error handling
        const voiceButton = document.getElementById('voiceButton');
        if ('SpeechRecognition' in window || 'webkitSpeechRecognition' in window) {
            const SpeechRecognition = window.SpeechRecognition || window.webkitSpeechRecognition;
            const recognition = new SpeechRecognition();

            // Configure recognition for better reliability
            recognition.continuous = false; // Stop after first result
            recognition.lang = 'en-US'; // Set language (adjust as needed)
            recognition.interimResults = false; // Only final results

            recognition.onresult = (event) => {
                // Check if results exist and handle safely
                if (event.results && event.results[0] && event.results[0][0]) {
                    const transcript = event.results[0][0].transcript;
                    const searchInput = document.getElementById('searchInput');
                    if (searchInput) {
                        searchInput.value = transcript;
                    } else {
                        console.error('Search input element not found');
                        alert('Search input not available. Please check the HTML.');
                    }
                } else {
                    console.error('No speech recognition results received');
                    alert('No speech recognized. Please try again.');
                }
            };

            recognition.onerror = (event) => {
                console.error('Speech recognition error:', event.error);
                alert(`Voice recognition failed: ${event.error}. Please try again.`);
            };

            recognition.onend = () => {
                console.log('Speech recognition ended');
                recognition.stop(); // Ensure recognition stops after use
            };

            voiceButton.addEventListener('click', () => {
                recognition.start();
                console.log('Starting speech recognition...');
            });
        } else {
            console.error('SpeechRecognition API not supported in this browser');
            alert('Voice recognition is not supported in this browser. Please use Chrome, Safari, or Edge.');
        }

        // Fetch dashboard data from the API
        fetch(`/api/patients/analytics/${userId}`, {
            headers: { 'Authorization': `Bearer ${token}` }
        })
            .then(response => {
                if (!response.ok) throw new Error('Network response was not ok');
                return response.json();
            })
            .then(data => {
                const dashboardContent = document.getElementById('dashboardContent');
                if (dashboardContent) {
                    dashboardContent.innerHTML = `
                        <p>Records: ${data.data.records || 'No records'}</p>
                        <p>Appointments: ${data.data.appointments || 'No appointments'}</p>
                    `;

                    // Optional: Update AR scene with dashboard data for visualization
                    if (scene && data.data.records) {
                        data.data.records.forEach((record, index) => {
                            const box = document.createElement('a-box');
                            box.setAttribute('position', `${index * 2} 1.6 -2`);
                            box.setAttribute('color', '#007bff');
                            scene.appendChild(box);
                        });
                    }
                } else {
                    console.error('Dashboard content element not found');
                    alert('Dashboard content not available. Please check the HTML.');
                }
            })
            .catch(error => {
                console.error('Dashboard fetch error:', error);
                alert(`Failed to load dashboard data: ${error.message || 'Unknown error'}`);
            });

        // Simulate FullCalendar placeholder 
        const calendar = document.getElementById('calendar');
        if (calendar) {
            calendar.innerHTML = '<p>Calendar placeholder (FullCalendar not loaded in this demo)</p>';
        } else {
            console.error('Calendar element not found');
            alert('Calendar not available. Please check the HTML.');
        }

        // Mock Socket.IO data if no server is available, now in scope
        setTimeout(() => {
            socket.emit('message', { message: 'Dashboard updated!' });
        }, 2000);
    });
}