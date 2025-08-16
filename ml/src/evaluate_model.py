# Will use to evaluate the trained model on the test data
from xgboost import XGBClassifier                                          # Going with an XGBoost model
from sklearn.metrics import accuracy_score                                 # Will also need this scikit-learn function to check accuracy easily
from ml.src.process_data import test_X, test_y                            # Import the testing datasets.

# Load the model
model = XGBClassifier()
model.load_model("ml/models/xgb_model.json")

# Can now get the predictions for the test dataset as well as the accuracy
y_pred = model.predict(test_X)
accuracy = accuracy_score(test_y, y_pred)
print("Test Accuracy:", accuracy)

