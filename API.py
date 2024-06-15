import requests
from json.decoder import JSONDecodeError
import time
import base64
import hmac
import hashlib
import os
from dotenv import load_dotenv



API_KEY = os.getenv('COINBASE_API_KEY')
API_SECRET = os.getenv('COINBASE_API_SECRET')
BASE_URL = os.getenv('COINBASE_URL')

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

    response = requests.get(url, headers=headers).json()
    return response

# Get account information
account_info = get_request('/accounts')
print(account_info)
