from datetime import datetime
import numpy as np
from utils.blockchain import BlockchainClient
import logging

class AnalyticsEngine:
    def __init__(self):
        self.blockchain_client = BlockchainClient()
        self.logger = logging.getLogger(__name__)

    def analyze_patient_data(self, patient_data):
        """
        Analyze patient data with role-specific messages.
        
        Args:
            patient_data (list): List of patient dictionaries
        
        Returns:
            dict: Analysis results
        """
        if not patient_data:
            return {"message": "No data to analyze. Please try again later."}
        conditions = [d['medical_condition'] for d in patient_data if 'medical_condition' in d]
        ages = [d['age'] for d in patient_data if 'age' in d]
        condition_freq = {cond: conditions.count(cond) for cond in set(conditions)}
        mean_age = np.mean(ages) if ages else 0
        role = BlockchainClient().get_user_role()  # Simulated role from blockchain
        name = "Patient"  # Default, update with actual name from user data
        message = self.get_role_message(role, "patient data analysis", True, name)
        self.logger.info(f"Analyzed patient data: {condition_freq}, Mean Age: {mean_age} - {message}")
        return {"condition_frequency": condition_freq, "mean_age": mean_age, "message": message}

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

   