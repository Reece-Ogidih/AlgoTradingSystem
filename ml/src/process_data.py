import os
from dotenv import load_dotenv 
from pathlib import Path            # Need this to access the .env file due to working directory difference
import pandas as pd
import numpy as np
from sqlalchemy import create_engine
from  utils import is_profitable, get_sliding_window, is_entry, assess_prof

# Load the .env variables
dotenv_path = Path(__file__).resolve().parents[3] / ".env"      # Due to the .env file being at the root of project
load_dotenv(dotenv_path)

os.getcwd()

user = os.getenv("DB_USER")
password = os.getenv("DB_PASS")
host = os.getenv("DB_HOST")
port = os.getenv("DB_PORT")
database = os.getenv("DB_NAME")
window_size_str = (os.getenv("WINDOW_SIZE"))
window_size = int(window_size_str)

# Use it to generate the string to connect to the MySQL database and then connect
connection_string = f"mysql+pymysql://{user}:{password}@{host}:{port}/{database}"
engine = create_engine(connection_string)

# Can now read from the table
dataframe = pd.read_sql("SELECT * FROM train_ml", con=engine)

# # Check that the connection is established and data is retrieved (passed)
#print(dataframe.head())


# There are a few more steps to undertake before our data can be used to train the ML model
# Since the use for the model is to validify proposed trades, each row within the data should be an entry position
# In order to preserve the nature of the candle data and the context of the trades, each row should also contain the sliding window info

# Note: I have defined the corresponding helper functions called within the next function in utils.py
def process_data(df, size=window_size):
    # I am using a window size to match what the bot used when deciding on possible trades
    # First will create empty list to contain the X_samples and y_labels
    X_samples = []
    y_labels  = []

    # Need to define the order of the columns for when we flatten everything
    feature_cols = ["open", "close", "high", "low", "volume", "adx", "sig_entry", "sig_exit"]

    for i in range(len(df)):
        # First we want to pass over rows where there is no new entry
        if not is_entry(df, i):
            continue
        window_df = get_sliding_window(df, i, size)

        # Add a check that the window is not empty
        if window_df is None:
            continue

        # Now we ensure we have the correct feature(columns) order
        window_df = window_df[feature_cols]
        
        # Now can flatten the window into one row
        flattened = window_df.to_numpy().flatten()
        # Can now obtain the label for this row
        label = is_profitable(df, i)

        # Need to check to ensure that there is no problems with getting the label, if there is we will print out the idx and skip the entry for now
        if not isinstance(label, int):
            print(label)
            continue 

        # Next, append the candidate row and the corresponding label to the datasets
        X_samples.append(flattened)
        y_labels.append(label)
    # # Convert to DataFrame/Series (decided against since numpy arrays have slightly less overhead)
    # X_df = pd.DataFrame(X_samples)
    # y_series = pd.Series(y_labels, name="label")
    X_df = np.array(X_samples)
    y_series = np.array(y_labels)
    return X_df, y_series


# Can now simply obtain the modified data here and run some checks
X_samples, y_labels = process_data(dataframe)
len(X_samples)      # Note that these both returned the same size as expected
len(y_labels)
print(X_samples.shape)
# Can now split data into training and test data sets
# Due to the nature of the data (time series), chronological order is very important to preserve and so will not opt for random split
# I will use a 70% Training -- 15% Validation -- 15% Testing split for my data
total_len = len(X_samples)
train_end = int(total_len * 0.7)
val_end = int(total_len * 0.85)

# .iloc is wau to do indexing for Pandas dataframe
# Must reset the indexing within each split of the data for the next function to actually work
train_X = X_samples[:train_end]
train_y = y_labels[:train_end]

val_X = X_samples[train_end:val_end]
val_y = y_labels[train_end:val_end]

test_X = X_samples[val_end:]
test_y = y_labels[val_end:]


# Originally had to check for if there were any open positions left after splitting the dataset
# However, after processing the data into the sliding windows with the corresponding y label, this is no longer necessary
# Instead, just need to ensure that the dimensions are equal
# len(train_X)        # I found here that the dimensions are equal for each set, so no problems
# len(train_y)
# len(val_X)
# len(val_y)
# len(test_y)
# len(test_y)

# Test for how pure trading logic bot performed
#print(assess_prof(y_labels))