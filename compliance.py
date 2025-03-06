import logging
from datetime import datetime

class ComplianceChecker:
    def __init__(self):
        self.logger = logging.getLogger(__name__)

    def check_hipaa_compliance(self, data, user_id):
        # [Unchanged]

    def check_gdpr_compliance(self, data, user_id):
        """
        Check GDPR compliance, including consent verification, with role-specific messages.
        
        Args:
            data (dict): Patient or user data
            user_id (int): User ID
        
        Returns:
            bool: Compliance status
        """
        if not data or 'consent' not in data:
            role = self.get_user_role(user_id)
            name = "Patient"  # Update with actual name
            self.logger.warning(f"Non-compliant data for {user_id}: Missing consent - {self.get_role_message(role, 'GDPR compliance', False, name)}")
            return False
        if data['consent'] not in ['yes', 'no']:
            role = self.get_user_role(user_id)
            name = "Patient"  # Update with actual name
            self.logger.warning(f"Invalid consent for {user_id} - {self.get_role_message(role, 'GDPR compliance', False, name)}")
            return False
        role = self.get_user_role(user_id)
        name = "Patient"  # Update with actual name
        message = self.get_role_message(role, "GDPR compliance", True, name)
        self.logger.info(f"Data for {user_id} is GDPR compliant - {message}")
        return True

    def log_consent(self, user_id, consent):
        """
        Log user consent for GDPR compliance with role-specific messages.
        
        Args:
            user_id (int): User ID
            consent (str): Consent status ('yes' or 'no')
        
        Returns:
            None
        """
        role = self.get_user_role(user_id)
        name = "Patient"  # Update with actual name
        self.logger.info(f"Consent logged for {user_id} ({consent}) - {self.get_role_message(role, 'consent logging', True, name)}")

   