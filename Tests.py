from CoinmarketcapData import DataParse
class Tests(DataParse):
    
    def _init_(self) -> None:
        super().__init__()
if _name_ == '_main_':
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