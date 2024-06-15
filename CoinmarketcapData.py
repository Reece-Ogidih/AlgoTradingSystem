import requests
import pandas as pd

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

def get_crypto_name(symbol):
    data = get_crypto_data(symbol)
    name = data['data'][symbol]['name']
    return name 

def get_crypto_price(symbol):
    data = get_crypto_data(symbol)
    price = data['data'][symbol]['quote']['GBP']['price']
    return price

def get_crypto_circ_supply(symbol):
    data = get_crypto_data(symbol)
    circ_supply = data['data'][symbol]['circulating_supply']
    return circ_supply

def get_crypto_volume(symbol):
    data = get_crypto_data(symbol)
    volume = data['data'][symbol]['quote']['GBP']['volume_24h']
    return volume

def get_crypto_market_cap(symbol):
    data = get_crypto_data(symbol)
    market_cap = data['data'][symbol]['quote']['GBP']['market_cap']
    return market_cap

def crypto_change_hour(symbol):
    data = get_crypto_data(symbol)
    hour = data['data'][symbol]['quote']['GBP']['percent_change_1h']
    return hour

def crypto_change_day(symbol):
    data = get_crypto_data(symbol)
    day = data['data'][symbol]['quote']['GBP']['percent_change_24h']
    return day 

def crypto_change_week(symbol):
    data = get_crypto_data(symbol)
    week = data['data'][symbol]['quote']['GBP']['percent_change_7d']
    return week

def crypto_change_over_time(symbol):
    hour = crypto_change_hour(symbol)
    day = crypto_change_day(symbol)
    week = crypto_change_week(symbol)
    output = f"1hr: {hour:.2f}%, 24hr: {day:.2f}%, 7d: {week:.2f}%"
    return output

def crypto_info(symbol):
    info = {
        'name' : get_crypto_name(symbol),
        'price' : get_crypto_price(symbol),
        '%1hr' : crypto_change_hour(symbol),
        '%24hr' : crypto_change_day(symbol),
        '%7d' : crypto_change_week(symbol),
        'market_cap' : get_crypto_market_cap(symbol),
        'volume' : get_crypto_volume(symbol),
        'circ_supply' : get_crypto_circ_supply(symbol)
    }
    return info 

#Example: For Bitcoin
get_crypto_name('BTC')
get_crypto_price('BTC')
get_crypto_volume('BTC')
get_crypto_market_cap('BTC')
crypto_change_hour('BTC')
crypto_change_day('ETH')
crypto_change_week('BTC')
crypto_change_over_time('BTC')
get_crypto_circ_supply('BTC')
crypto_info('BNB')

#Testing and manipulating

symbols = ['BTC', 'ETH', 'USDT']  # List of cryptocurrency symbols
crypto_data = {}

for symbol in symbols:
    crypto_data[symbol] = crypto_info(symbol)

# Displaying the collected data
for symbol, info in crypto_data.items():
    print(f"{symbol} info: {info}")

# Convert the collected data into a pandas DataFrame for easier manipulation
df = pd.DataFrame.from_dict(crypto_data, orient='index')
print(df)