# This will be for resuable helper functions

import pandas
import os
from dotenv import load_dotenv 
from pathlib import Path

dotenv_path = Path(__file__).resolve().parents[3] / ".env"
load_dotenv()

prof_threshold_str = os.getenv("PROFIT_THRESHOLD")
prof_threshold = float(prof_threshold_str)

# First we will define a function to collect the sliding window info at a specified candle
def get_sliding_window(df, idx, size):
    # Need a checker for when there is no sliding window available
    if idx < size - 1:
        return None
    return df.iloc[idx - size + 1 : idx + 1]        # Note that the +1 is due to this being exclusive bounds

# Next function is a simple function to find the exit price of an entered positoin
def find_exit(df, idx):
    # Logic for this function is to find the position entered (short vs long) and then find the next occurance of that signal in sig_exit
    entry_signal = df.iloc[idx]['sig_entry']
    for i in range(idx+1, len(df)):
        exit_signal = df.iloc[i]['sig_exit']
        # Will add a check to ensure that we dont have the first occurance in the exit signal column be for the opposite signal
        # Can do this since we know the "simple" approach the bot takes, it only inputs one trade at a time
        # Thus we know that after, for example, a long entry(1), the next non-zero in sig_exit should also be 1 and vice-versa.
        if exit_signal == 0:
            continue
        elif exit_signal != entry_signal:
            return f"""Error, next exit position doesnt correlate to this entry: {idx}"""
        else:
            return i
        
# Will also define a quick function to check that an supposed entry index is actually an entry index
def is_entry(df, idx):
    return df.iloc[idx]['sig_entry'] != 0

# The next function necessary is a helper function to determine if a trade was "profitable"
# I have deemed a trade to be "profitable" if the price trended >X% in the correct direction, this could be adjusted depending on performance
def is_profitable(df, idx, threshold=prof_threshold):
    # First we need the signal (so was it a short or long position)
    signal      = df.iloc[idx]['sig_entry']

    # Will add a checker to ensure that the entry signal is not 0 (no trade)
    if signal == 0:
        return f"""Error, no position entered at: {idx}"""
    
    # Now find the price at entry
    entry_price = df.iloc[idx]['close']

    # To find the exit price, I call the helper function find_exit to get the idx and then can extract the price
    exit_idx    = find_exit(df, idx)
    if exit_idx is None:
        return f"""missing exit at: {idx}"""            # Checker for if we have entries with no exit position within dataset
    exit_price  = df.iloc[exit_idx]['close']

    if entry_price == 0:
        return f"""Error, entry price at {idx} is 0"""  # Check for divide by zero
    pct_change = (exit_price - entry_price) / entry_price      

    # Can now return either 1 or 0 depending on profitable or not, will first assess if the signal is short or long and then assess the profit margin
    if (signal == 1 and pct_change >= threshold) or (signal == -1 and pct_change <= -threshold):
        return 1
    else:
        return 0

# Quick function to check the profitability of the bot WITHOUT ML component, in other words % of labels that are 1
def assess_prof(df):
    good = 0.0
    total = len(df)
    for i in range(len(df)):
        if df[i] == 1:
            good += 1.0
    pct = good / total
    return pct

