from cryptography.fernet import Fernet
import os
import logging
from dotenv import load_dotenv

load_dotenv()
logger = logging.getLogger(__name__)

class QuantumResistantEncryptor:
    def __init__(self):
        self.key = os.getenv('ENCRYPTION_KEY', Fernet.generate_key())
        self.cipher_suite = Fernet(self.key)
        self.logger = logging.getLogger(__name__)

    def encrypt_data(self, data):
        """
        Encrypt data with role-specific messages on success/failure.
        
        Args:
            data (str): Data to encrypt
        
        Returns:
            bytes: Encrypted data
        """
        try:
            encrypted = self.cipher_suite.encrypt(data.encode())
            role = "patient"  # Simulated role, update from request context
            name = "Patient"  # Default, update with actual name
            message = self.get_role_message(role, "data encryption", True, name)
            self.logger.info(f"Data encrypted successfully - {message}")
            return encrypted
        except Exception as e:
            role = "patient"  # Simulated role
            name = "Patient"  # Default
            self.logger.error(f"Encryption error: {e} - {self.get_role_message(role, 'data encryption', False, name)}")
            raise Exception(self.get_role_message(role, "data encryption", False, name))

    def get_role_message(self, role, feature, success=True, name=None):
        """
        Generate role-specific, user-friendly messages.
        
        Args:
            role (str): User role (admin, doctor, patient)
            feature (str): Feature or action being performed
            success (bool): Whether the action succeeded
            name (str, optional): User name, defaults to role
        
        Returns:
            str: Formatted message
        """
        name = name or role.capitalize()
        messages = {
            'admin': {
                True: f"Thank you, Admin! Your action on {feature} has been completed successfully.",
                False: f"Sorry, Admin, we couldn’t process your {feature} request. Please try again or contact support."
            },
            'doctor': {
                True: f"Great job, Doctor! Your update to {feature} was successful.",
                False: f"Oops, Doctor, we encountered an issue with your {feature}. Please try again later or reach out to support."
            },
            'patient': {
                True: f"Thank you, {name}! Your {feature} has been updated successfully.",
                False: f"Sorry, {name}, we couldn’t complete your {feature} request. Please try again or contact our support team."
            }
        }
        return messages[role][success]

