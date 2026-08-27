import pandas as pd

def load_data(file_path):
    """Loads data from a CSV file."""
    try:
        df = pd.read_csv(file_path)
        return df
    except FileNotFoundError:
        print(f"Error: File not found at {file_path}")
        return None

def process_data(df):
    """Placeholder for data processing logic."""
    if df is not None:
        print("Data loaded successfully.")
        print(df.head())
        # Add processing logic here
        print("Data processing complete.")
    else:
        print("No data to process.")

if __name__ == "__main__":
    data = load_data("data/input.csv")
    process_data(data)


key = "AKIA2UC3FVYLKVQKFWUH"