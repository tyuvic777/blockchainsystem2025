import os
import requests
import logging
from dotenv import load_dotenv

load_dotenv()
logger = logging.getLogger(__name__)

class WearableService:
    def __init__(self):
        self.logger = logging.getLogger(__name__)

    def process_wearable_data(self, user_id, data, role='patient', name='Patient'):
        """
        Process wearable data with role-specific messages.
        
        Args:
            user_id (int): User ID
            data (dict): Wearable data (e.g., heart rate, steps)
            role (str): User role (default: patient)
            name (str): User name (default: Patient)
        
        Returns:
            dict: Processed data with role-specific message
        """
        try:
            response = requests.post(f'{os.getenv("BACKEND_URL", "http://localhost:3000")}/patients/wearable/{user_id}', json=data)
            response.raise_for_status()
            blockchain_client = BlockchainClient()  # Assume initialized elsewhere
            blockchain_client.add_wearable_data(user_id, json.dumps(data))
            message = self.get_role_message(role, "wearable data processing", True, name)
            self.logger.info(f"Wearable data processed for {user_id} - {message}")
            return {"message": message, "data": response.json()}
        except Exception as e:
            self.logger.error(f"Wearable data error: {e} - {self.get_role_message(role, 'wearable data processing', False, name)}")
            raise Exception(self.get_role_message(role, "wearable data processing", False, name))

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

