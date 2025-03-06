import redis
import os
import time
from functools import wraps
from dotenv import load_dotenv
import logging

load_dotenv()
logger = logging.getLogger(__name__)

class CacheManager:
    def __init__(self):
        self.redis_client = redis.Redis(host=os.getenv('REDIS_HOST', 'localhost'), port=6379, db=0, decode_responses=True)
        self.default_timeout = 3600  # 1 hour in seconds

    def cache_decorator(self, timeout=None):
        """
        Decorate functions to cache results with role-specific messages on cache misses.
        
        Args:
            timeout (int, optional): Cache timeout in seconds, defaults to 1 hour
        
        Returns:
            function: Decorated function
        """
        def decorator(f):
            @wraps(f)
            def wrapper(*args, **kwargs):
                cache_key = f"{f.__name__}:{args}:{json.dumps(kwargs)}"
                cached = self.redis_client.get(cache_key)
                if cached:
                    return json.loads(cached)
                result = f(*args, **kwargs)
                self.redis_client.setex(cache_key, timeout or self.default_timeout, json.dumps(result))
                role = kwargs.get('role', 'patient')  # Simulated role, update from request context
                name = "Patient"  # Default
                message = self.get_role_message(role, f.__name__, True, name)
                logger.info(f"Cached {f.__name__} for {role}: {message}")
                return result
            return wrapper
        return decorator

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

