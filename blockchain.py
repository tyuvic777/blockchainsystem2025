import os
import json
import requests
import time
import numpy as np
from functools import wraps
from dotenv import load_dotenv
from cryptography.fernet import Fernet
import logging
import threading

load_dotenv()
logger = logging.getLogger(__name__)

BLOCKCHAIN_URL = os.getenv('PYTHON_BLOCKCHAIN_URL', 'https://localhost:7051')
ETH_URL = os.getenv('ETH_URL', 'http://localhost:8545')
IPFS_URL = os.getenv('IPFS_URL', 'http://localhost:5001')

key = Fernet.generate_key()
cipher_suite = Fernet(key)

_thread_local = threading.local()

def retry(max_attempts=5, initial_delay=2000, jitter_range=(0, 100)):
    # [Unchanged retry decorator]

def get_role_message(role, feature, success=True, name=None):
    # [Unchanged role-specific message function]

class BlockchainClient:
    def __init__(self, token=None):
        self.token = token
        _thread_local.token = token
        self.headers = {'Authorization': f'Bearer {token}'} if token else {}
        self.lock = threading.Lock()
        self.min_batch_size = 100
        self.max_batch_size = 1000

    def _get_headers(self):
        # [Unchanged]

    def _dynamic_batch_size(self, data_size):
        """
        Calculate dynamic batch size based on dataset size.
        
        Args:
            data_size (int): Number of items to batch
        
        Returns:
            int: Optimal batch size
        """
        if data_size <= 1000:
            return self.min_batch_size
        elif data_size <= 10000:
            return 500
        else:
            return self.max_batch_size

    @retry()
    def batch_register_patients(self, patients, role='patient', name='Patient'):
        """
        Batch register patients on the blockchain with dynamic batch sizing and role-specific messages.
        """
        with self.lock:
            results = []
            data_size = len(patients)
            batch_size = self._dynamic_batch_size(data_size)
            for i in range(0, data_size, batch_size):
                batch = patients[i:i + batch_size]
                for patient in batch:
                    result = self.register_patient(
                        patient['patient_id'], patient['name'], patient['medical_condition'],
                        patient['medication'], patient['date_of_admission'], patient['discharge_date'], patient['doctor'],
                        role, name
                    )
                    results.append(result)
                time.sleep(0.1 * np.random.random())  # Jitter for network stability
            return results

