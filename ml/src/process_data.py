import os
from dotenv import load_dotenv 
import pandas
from sqlalchemy import create_engine

# Load the .env variables
load_dotenv()

user = os.getenv("DB_USER")
password = os.getenv("DB_PASS")
host = os.getenv("DB_HOST")
port = os.getenv("DB_PORT")
database = os.getenv("DB_NAME")

# Use it to generate the string to connect to the MySQL database and then connect
connection_string = f"mysql+pymysql://{user}:{password}@{host}:{port}/{database}"
engine = create_engine(connection_string)

# Can now read from the table
dataframe = pandas.read_sql("SELECT * FROM train_ml", con=engine)

# # Check that the connection is established and data is retrieved (passed)
# print(df.head())

# Can now split data into training and test data sets
# Due to the nature of the data (time series), chronological order is very important to preserve and so will not opt for random split
# I will use a 70% Training -- 15% Validation -- 15% Testing split for my data
total_len = len(dataframe)
train_end = int(total_len * 0.7)
val_end = int(total_len * 0.85)

# .iloc is wau to do indexing for Pandas dataframe
# Must reset the indexing within each split of the data for the next function to actually work
train_df = dataframe.iloc[:train_end].reset_index(drop=True)
val_df = dataframe.iloc[train_end:val_end].reset_index(drop=True)
test_df = dataframe.iloc[val_end:].reset_index(drop=True)

# Now there could be some open positions at the end of these datasets. In other words, the bot enters a trade but the exit position is not within the dataset.
# Will build a small function to check for this first
def has_open_trades(df):
    # First collect the entry and exit indexs
    entry_idxs = df.index[df['sig_entry'] != 0]
    exit_idxs = df.index[df['sig_entry'] != 0]
    
    # Will add some checks to ensure that there are entry and exits within all datasets
    if len(entry_idxs) == 0:
        return "No entry positions"
    if len(exit_idxs) == 0:
        return "No exit positions"
    
    # Can now check for if there are any open positions
    return entry_idxs.max() > exit_idxs.max()

# Now can check every set, seems like everything is working well so far
# Will add an if__name__ == "__main__" block here to stop the debug code from being called when importing the data in other files
if __name__ == "__main__":
    print("Training", has_open_trades(train_df))
    print("Validation", has_open_trades(val_df))
    print("Testing", has_open_trades(test_df))

