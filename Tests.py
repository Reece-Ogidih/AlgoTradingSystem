from CoinmarketcapData import DataParse
import pandas as pd
class Tests(DataParse):
    
    def __init__(self) -> None:
        super().__init__()
if __name__ == '__main__':
    Tests().get_crypto_name('BTC')
    Tests().get_crypto_price('ETH')
    Tests().get_crypto_volume('BTC')
    Tests().get_crypto_market_cap('BTC')
    Tests().crypto_change_hour('BTC')
    Tests().crypto_change_day('ETH')
    Tests().crypto_change_week('BTC')
    Tests().crypto_change_over_time('BTC')
    Tests().get_crypto_circ_supply('BTC')
    Tests().crypto_info('BNB')


if __name__ == '__main__':
    data_parser = DataParse()

    symbols = ['BTC', 'ETH', 'USDT']  
    crypto_data = {}

    for symbol in symbols:
        crypto_data[symbol] = data_parser.crypto_info(symbol)

    df = pd.DataFrame.from_dict(crypto_data, orient='index')
    print(df)