from cryptography.hazmat.primitives import hashes
from cryptography.hazmat.primitives.asymmetric import ec
from cryptography.hazmat.backends import default_backend
import random

# Simplified Schnorr ZKP 
def generate_zkp(user_id):
    # Private key (simulated, should be stored securely)
    private_key = ec.generate_private_key(ec.SECP256R1(), default_backend())
    public_key = private_key.public_key()
    
    # Random nonce
    k = random.randint(1, ec.SECP256R1().curve.order - 1)
    R = k * ec.SECP256R1().generator
    
    # Challenge 
    digest = hashes.Hash(hashes.SHA256(), backend=default_backend())
    digest.update(user_id.encode() + R.x.to_bytes(32, 'big'))
    c = int.from_bytes(digest.finalize(), 'big') % ec.SECP256R1().curve.order
    
    # Response
    s = (k - private_key.private_numbers().private_value * c) % ec.SECP256R1().curve.order
    return {'R_x': R.x, 's': s, 'public_key': public_key.public_numbers().x}

def verify_zkp(user_id, proof):
    try:
        R_x = proof['R_x']
        s = proof['s']
        public_key_x = proof['public_key']
        
        # Reconstruct public key 
        public_key = ec.EllipticCurvePublicNumbers(public_key_x, 0, ec.SECP256R1()).public_key(default_backend())
        
        # Recompute R
        digest = hashes.Hash(hashes.SHA256(), backend=default_backend())
        digest.update(user_id.encode() + R_x.to_bytes(32, 'big'))
        c = int.from_bytes(digest.finalize(), 'big') % ec.SECP256R1().curve.order
        
        R_computed = s * ec.SECP256R1().generator + c * public_key
        return R_computed.x == R_x
    except Exception:
        return False

# Client-side helper 
if __name__ == "__main__":
    user_id = "1"
    proof = generate_zkp(user_id)
    print("Proof:", proof)
    print("Verified:", verify_zkp(user_id, proof))