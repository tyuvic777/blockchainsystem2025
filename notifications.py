import os
import requests
import smtplib
from email.mime.text import MIMEText
from flask_socketio import SocketIO
import logging
from dotenv import load_dotenv

load_dotenv()
logger = logging.getLogger(__name__)

class NotificationService:
    def __init__(self, socketio=None):
        self.socketio = socketio
        self.twilio_sid = os.getenv('TWILIO_ACCOUNT_SID')
        self.twilio_token = os.getenv('TWILIO_AUTH_TOKEN')
        self.twilio_phone = os.getenv('TWILIO_PHONE_NUMBER')
        self.email_user = os.getenv('EMAIL_USER')
        self.email_password = os.getenv('EMAIL_PASSWORD')
        self.logger = logging.getLogger(__name__)

    def send_sms(self, to, message):
        """
        Send SMS notification with role-specific messages.
        
        Args:
            to (str): Recipient phone number
            message (str): Message content
        """
        try:
            response = requests.post(
                'https://api.twilio.com/2010-04-01/Accounts/{}/Messages.json'.format(self.twilio_sid),
                auth=(self.twilio_sid, self.twilio_token),
                data={'To': to, 'From': self.twilio_phone, 'Body': message}
            )
            response.raise_for_status()
            role = "patient"  
            name = "Patient"  
            message = self.get_role_message(role, "SMS notification", True, name)
            self.logger.info(f"SMS sent to {to} - {message}")
        except Exception as e:
            role = "patient"  # Simulated role
            name = "Patient"  # Default
            self.logger.error(f"SMS error: {e} - {self.get_role_message(role, 'SMS notification', False, name)}")

    def send_email(self, to, subject, body):
        """
        Send email notification with role-specific messages.
        
        Args:
            to (str): Recipient email
            subject (str): Email subject
            body (str): Email body
        """
        try:
            msg = MIMEText(body)
            msg['Subject'] = subject
            msg['From'] = self.email_user
            msg['To'] = to
            with smtplib.SMTP('smtp.gmail.com', 587) as server:
                server.starttls()
                server.login(self.email_user, self.email_password)
                server.send_message(msg)
            role = "patient"  
            name = "Patient"  
            message = self.get_role_message(role, "email notification", True, name)
            self.logger.info(f"Email sent to {to} - {message}")
        except Exception as e:
            role = "patient"  
            name = "Patient"  
            self.logger.error(f"Email error: {e} - {self.get_role_message(role, 'email notification', False, name)}")

    def send_socketio(self, event, data, room):
        """
        Send SocketIO notification with role-specific messages.
        
        Args:
            event (str): SocketIO event name
            data (dict): Event data
            room (str): Room to broadcast to
        """
        try:
            role = "patient"  
            name = "Patient"  
            message = self.get_role_message(role, "SocketIO notification", True, name)
            data['message'] = message
            self.socketio.emit(event, data, room=room)
            self.logger.info(f"SocketIO event {event} sent to room {room} - {message}")
        except Exception as e:
            role = "patient"  
            name = "Patient"  
            self.logger.error(f"SocketIO error: {e} - {self.get_role_message(role, 'SocketIO notification', False, name)}")

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