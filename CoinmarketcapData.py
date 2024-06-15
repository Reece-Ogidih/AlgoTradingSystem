import requests

API_KEY = '628a69b0-ac6b-4ed6-9a44-9e552b8f260f'

# Base URL for CoinMarketCap API
BASE_URL = 'https://pro-api.coinmarketcap.com/v1'

# Function to get cryptocurrency market data
def get_crypto_data(symbol):
    url = f'{BASE_URL}/cryptocurrency/quotes/latest'
    parameters = {
        'symbol': symbol,
        'convert': 'GBP'  # Convert the data to GBP
    }
    headers = {
        'Accepts': 'application/json',
        'X-CMC_PRO_API_KEY': API_KEY
    }
    response = requests.get(url, headers=headers, params=parameters).json()
    return response

# Example: Get market data for Bitcoin (BTC)
btc_data = get_crypto_data('BTC')

btc_price = btc_data['data']['BTC']['quote']['GBP']['price']
print(f"BTC Price: ${btc_price}")