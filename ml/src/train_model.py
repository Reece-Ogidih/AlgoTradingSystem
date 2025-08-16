from xgboost import XGBClassifier                                          # Going with an XGBoost model
import sklearn                                                             # Will also need some scikit-learn functions
from ml.src.process_data import train_X, train_y, val_X, val_y             # Import the training and validation datasets.

# Will use the scikit-learn wrapped interface model XGBClassifier
# Can now define the model
model = XGBClassifier(
    n_estimators = 1000,                    # This is the upper bound on num of trees. Since will be using early stopping, this shouldn't be too impactful
    early_stopping_rounds = 50,             # Stop if no improvement for 50 rounds
    learning_rate = 0.05,                   # "eta"
    max_depth = 6,                          # max depth of each tree, will start with 6, could go smaller if overfitting
    subsample = 0.8,                        # the fraction of rows randomly sampled for each tree
    colsample_bytree = 1.0,                 # fraction of columns samples, I picked for all rows to be considered due to nature of data
    reg_alpha = 0.0,                        # For now have set L1 and L2 to small values as a default
    reg_lambda = 1.0,                       # Could increase these if signs of overfitting
    eval_metric = "logloss",                # Decided to use logloss since I am interested in probabilities
    use_label_encoder = False,              # Avoids XGBoost’s older internal label encoder
    random_state = 42,                      # Seed for randomisation
    n_jobs = -1                             # Want it to use all available cores
)  

# Next need to actually fit the model to the training data
# Will also use the validation data set for early stopping
model.fit(
    train_X, train_y,                       # Fit the model on the training data
    eval_set=[(val_X, val_y)],              # Validation set for early stopping and monitoring
    verbose = 10                            # Print the eval metric (logloss) every 10 rounds
)

# Finally we want to save the trained model
model.save_model("ml/models/xgb_model.json")