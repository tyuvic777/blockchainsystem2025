/**
 * Initialize the care plan page with role-specific messages, real-time updates, and ZKP.
 * @param {string} token - JWT token for authentication
 * @param {number} userId - User ID
 * @param {string} role - User role (admin, doctor, patient)
 * @param {string} socketUrl - SocketIO URL
 */
import { ec as EC } from 'elliptic';  

export function initializeCareplanPage(token, userId, role, socketUrl) {
    const socket = io(socketUrl);
    const ec = new EC('secp256r1');

    socket.on('connect', () => console.log('Connected to SocketIO'));
    socket.on('careplan_update', (data) => console.log('Care plan updated:', data));

    // Generate ZKP proof 
    function generateZKP(userId) {
        const keyPair = ec.genKeyPair();
        const k = ec.genKeyPair().getPrivate();  
        const R = ec.g.mul(k);  // R = k * G
        const hash = new Uint8Array(sha256.arrayBuffer(userId + R.getX().toString(16)));
        const c = BigInt('0x' + Buffer.from(hash).toString('hex').slice(0, 64)) % ec.curve.n;
        const s = k.add(keyPair.getPrivate().mul(c)).mod(ec.curve.n);
        return {
            R_x: R.getX().toString(10),
            s: s.toString(10),
            public_key: keyPair.getPublic().getX().toString(10)
        };
    }

    document.getElementById('careplanForm').addEventListener('submit', async (e) => {
        e.preventDefault();
        const carePlan = document.getElementById('carePlan').value;
        const zkpProof = generateZKP(userId);
        const data = { carePlan, zkp_proof: zkpProof };
        try {
            const response = await fetch('/api/patients/careplan', {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(data)
            });
            const result = await response.json();
            if (response.ok) {
                displaySuccess(role, "care plan submission");
            } else {
                displayError(role, "care plan submission");
            }
        } catch (error) {
            displayError(role, "care plan submission");
        }
    });

    function displayError(role, feature) {
        const messages = {
            'admin': `Sorry, Admin, we couldn’t process your ${feature} request.`,
            'doctor': `Oops, Doctor, we encountered an issue with your ${feature}.`,
            'patient': `Sorry, ${localStorage.getItem('name') || 'Patient'}, we couldn’t complete your ${feature} request.`
        };
        alert(messages[role]);
    }

    function displaySuccess(role, feature) {
        const messages = {
            'admin': `Thank you, Admin! Your ${feature} has been completed successfully.`,
            'doctor': `Great job, Doctor! Your ${feature} was successful.`,
            'patient': `Thank you, ${localStorage.getItem('name') || 'Patient'}! Your ${feature} has been updated successfully.`
        };
        alert(messages[role]);
    }
}

// SHA-256 implementation
function sha256(message) {
    const encoder = new TextEncoder();
    const data = encoder.encode(message);
    return crypto.subtle.digest('SHA-256', data);
}