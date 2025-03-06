import os
import logging
import stripe
import json
from dotenv import load_dotenv
from blockchain import BlockchainClient
from datetime import datetime
import re

load_dotenv()
logger = logging.getLogger(__name__)

# Stripe setup
stripe.api_key = os.getenv('STRIPE_API_KEY')  

class PaymentService:
    def __init__(self):
        self.logger = logging.getLogger(__name__)
        self.blockchain_client = BlockchainClient()

    def upload_insurance(self, patient_id, insurance_data, role='patient', name=None):
        """
        Upload insurance information for a patient to the blockchain.
        """
        try:
            insurance_json = json.dumps(insurance_data)
            result = self.blockchain_client.register_patient(
                patient_id=patient_id,
                name=insurance_data.get('name', 'Unknown'),
                condition='N/A',
                medication='N/A',
                admission='N/A',
                discharge='N/A',
                doctor=insurance_data.get('doctor', 'N/A'),
                role=role,
                name_user=name
            )
            if result['status'] != 'success':
                raise Exception("Blockchain registration failed")
            message = self.get_role_message(role, "insurance upload", True, name)
            self.logger.info(f"Insurance uploaded for patient {patient_id} - {message}")
            return {"message": message, "data": result}
        except Exception as e:
            message = self.get_role_message(role, "insurance upload", False, name)
            self.logger.error(f"Insurance upload error for {patient_id}: {e} - {message}")
            raise Exception(message)

    def get_payment_history(self, patient_id, role='patient', name=None):
        """
        Retrieve past payment history for a patient from the blockchain.
        """
        try:
            response = self.blockchain_client.channel.invoke_chaincode('paymentHistory', 'getPaymentHistory', [str(patient_id)])
            if not response:
                raise Exception("No payment history found")
            history = json.loads(response)
            message = self.get_role_message(role, "payment history retrieval", True, name)
            self.logger.info(f"Payment history retrieved for {patient_id} - {message}")
            return {"message": message, "data": history}
        except Exception as e:
            message = self.get_role_message(role, "payment history retrieval", False, name)
            self.logger.error(f"Payment history error for {patient_id}: {e} - {message}")
            raise Exception(message)

    def process_payment(self, patient_id, amount, card_details, role='patient', name=None):
        """
        Process a payment via credit/debit card and record it on the blockchain.
        """
        try:
            # Validate card details
            card_number = card_details.get('number', '')
            expiry = card_details.get('expiry', '')
            cvv = card_details.get('cvv', '')

            if not re.match(r'^\d{13,19}$', card_number):
                raise ValueError("Card number must be 13-19 digits.")
            if not re.match(r'^\d{2}/\d{2}$', expiry):
                raise ValueError("Expiry must be in MM/YY format (e.g., 12/25).")
            month, year = map(int, expiry.split('/'))
            if month < 1 or month > 12:
                raise ValueError("Month must be 01-12.")
            current_year = datetime.now().year % 100
            if year < current_year or (year == current_year and month < datetime.now().month):
                raise ValueError("Card is expired.")
            if not re.match(r'^\d{3,4}$', cvv):
                raise ValueError("CVV must be 3 or 4 digits.")

            # Process payment with Stripe
            charge = stripe.Charge.create(
                amount=int(amount * 100),
                currency='usd',
                source={
                    'object': 'card',
                    'number': card_number,
                    'exp_month': str(month),
                    'exp_year': f"20{year}",
                    'cvc': cvv
                },
                description=f"Payment for patient {patient_id}"
            )

            # Record payment on blockchain
            timestamp = datetime.now().isoformat()
            result = self.blockchain_client.channel.invoke_chaincode(
                'paymentHistory',
                'recordPayment',
                [str(patient_id), str(amount), timestamp, "completed"]
            )
            if not result or "success" not in result.lower():
                raise Exception("Blockchain payment recording failed")

            message = self.get_role_message(role, "payment processing", True, name)
            self.logger.info(f"Payment processed for {patient_id} - {message}")
            return {"message": message, "data": charge}
        except (ValueError, stripe.error.StripeError) as e:
            message = self.get_role_message(role, "payment processing", False, name)
            self.logger.error(f"Payment error for {patient_id}: {e} - {message}")
            raise Exception(f"{message} Error: {str(e)}")
        except Exception as e:
            message = self.get_role_message(role, "payment processing", False, name)
            self.logger.error(f"Payment error for {patient_id}: {e} - {message}")
            raise Exception(message)

    def get_role_message(self, role, feature, success=True, name=None):
        """
        Generate formal, professional messages for website users.
        
        Args:
            role (str): User role (admin, doctor, patient)
            feature (str): Feature or action being performed
            success (bool): Whether the action succeeded
            name (str, optional): User name, defaults to role
        
        Returns:
            str: Formatted message
        """
        name = name or "User"  # Default to "User" if no name provided
        messages = {
            'admin': {
                True: f"Dear Administrator, your {feature} has been successfully completed. Thank you for your diligent oversight.",
                False: f"Dear Administrator, we regret that an issue occurred during your {feature}. Please retry or contact our support team for assistance."
            },
            'doctor': {
                True: f"Dear Dr., your {feature} has been processed successfully. We appreciate your dedicated service.",
                False: f"Dear Dr., we apologize for the inconvenience; an error occurred with your {feature}. Please attempt again or reach out to our support team."
            },
            'patient': {
                True: f"Dear {name}, your {feature} has been completed successfully. Thank you for choosing our services.",
                False: f"Dear {name}, we’re sorry, but an issue prevented your {feature} from completing. Please try again or contact our support team for help."
            }
        }
        return messages[role][success]
   