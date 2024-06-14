import requests
from json.decoder import JSONDecodeError
import time
import base64
import hmac
import hashlib

API_KEY = 'ee9a7cdf-e311-44ec-a0de-e32e5997630c'
API_SECRET = 'MHcCAQEEIM+ort7fAKVb/RjuwntndFLTrZVvNVSzB4wj1aZ9tZQqoAoGCCqGSM49\nAwEHoUQDQgAE9YJI2t5vlM5LDyBDQBJ0UtFLmKREvbgA26sSCBeC6bwh6dP3jvIO\nbFgN60VFb2MT+QRf6IZvVQx1DO7wLdsU7w=='
#API_PASSPHRASE = 'your_passphrase'
BASE_URL = 'https://api.pro.coinbase.com'

def get_request(endpoint):
    url = BASE_URL + endpoint
    timestamp = str(time.time())
    message = timestamp + 'GET' + endpoint
    hmac_key = base64.b64decode(API_SECRET)
    signature = hmac.new(hmac_key, message.encode('utf-8'), hashlib.sha256)
    signature_b64 = base64.b64encode(signature.digest()).decode('utf-8')

    headers = {
        'CB-ACCESS-KEY': API_KEY,
        'CB-ACCESS-SIGN': signature_b64,
        'CB-ACCESS-TIMESTAMP': timestamp
    }

    response = requests.get(url, headers=headers)
    return response.json()

# Get account information
account_info = get_request('/accounts')
print(account_info)
