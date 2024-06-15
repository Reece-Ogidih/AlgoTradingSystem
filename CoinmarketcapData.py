import requests
import os
from dotenv import load_dotenv
load_dotenv()


class DataParse:
   
    def __init__(self) -> None:
        self.BASE_URL = os.getenv('COINMARKETCAP_URL')
        self.API_KEY = os.getenv('COINMARKETCAP_API_KEY')

    def get_crypto_data(self, symbol):
        url = f'{self.BASE_URL}/cryptocurrency/quotes/latest'
        parameters = {
            'symbol': symbol,
            'convert': 'GBP'  # Convert the data to GBP
        }
        headers = {
            'Accepts': 'application/json',
            'X-CMC_PRO_API_KEY': self.API_KEY
        }
        response = requests.get(url, headers=headers, params=parameters).json()
        return response

    def get_crypto_name(self, symbol):
        data = self.get_crypto_data(symbol)
        name = data['data'][symbol]['name']
        return name 

    def get_crypto_price(self, symbol):
        data = self.get_crypto_data(symbol)
        price = data['data'][symbol]['quote']['GBP']['price']
        return price

    def get_crypto_circ_supply(self, symbol):
        data = self.get_crypto_data(symbol)
        circ_supply = data['data'][symbol]['circulating_supply']
        return circ_supply

    def get_crypto_volume(self, symbol):
        data = self.get_crypto_data(symbol)
        volume = data['data'][symbol]['quote']['GBP']['volume_24h']
        return volume

    def get_crypto_market_cap(self, symbol):
        data = self.get_crypto_data(symbol)
        market_cap = data['data'][symbol]['quote']['GBP']['market_cap']
        return market_cap

    def crypto_change_hour(self, symbol):
        data = self.get_crypto_data(symbol)
        hour = data['data'][symbol]['quote']['GBP']['percent_change_1h']
        return hour

    def crypto_change_day(self, symbol):
        data = self.get_crypto_data(symbol)
        day = data['data'][symbol]['quote']['GBP']['percent_change_24h']
        return day 

    def crypto_change_week(self, symbol):
        data = self.get_crypto_data(symbol)
        week = data['data'][symbol]['quote']['GBP']['percent_change_7d']
        return week

    def crypto_change_over_time(self, symbol):
        hour = self.crypto_change_hour(symbol)
        day = self.crypto_change_day(symbol)
        week = self.crypto_change_week(symbol)
        output = f"1hr: {hour:.2f}%, 24hr: {day:.2f}%, 7d: {week:.2f}%"
        return output

    def crypto_info(self, symbol):
        info = {
            'name' : self.get_crypto_name(symbol),
            'price' : self.get_crypto_price(symbol),
            '%1hr' : self.crypto_change_hour(symbol),
            '%24hr' : self.crypto_change_day(symbol),
            '%7d' : self.crypto_change_week(symbol),
            'market_cap' : self.get_crypto_market_cap(symbol),
            'volume' : self.get_crypto_volume(symbol),
            'circ_supply' : self.get_crypto_circ_supply(symbol)
        }
        return info 
