import os
import logging
from dotenv import load_dotenv
from transformers import pipeline

load_dotenv()
logger = logging.getLogger(__name__)

class NLPChatbot:
    def __init__(self):
        self.chatbot = pipeline("conversational", model="facebook/blenderbot-400M-distill")
        self.logger = logging.getLogger(__name__)

    def analyze_intent(self, message, user_id):
        """
        Analyze user intent with role-specific messages.
        
        Args:
            message (str): User message
            user_id (int): User ID
        
        Returns:
            str: Intent or response
        """
        try:
            response = self.chatbot(message)
            role = "patient"  # Simulated role
            name = "Patient"  # Default
            message = self.get_role_message(role, "intent analysis", True, name)
            self.logger.info(f"Intent analyzed for {user_id}: {response} - {message}")
            return response
        except Exception as e:
            role = "patient"  # Simulated role
            name = "Patient"  # Default
            self.logger.error(f"Intent analysis error: {e} - {self.get_role_message(role, 'intent analysis', False, name)}")
            return "Sorry, I couldn’t understand your message."

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

 