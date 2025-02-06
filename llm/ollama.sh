#!/bin/bash

# Create logs directory if it doesn't exist
mkdir -p logs

# Function to check if Ollama is installed
check_ollama() {
    if ! command -v ollama &> /dev/null; then
        echo "Ollama is not installed. Installing now..."
        curl -fsSL https://ollama.com/install.sh | sh
        if [ $? -ne 0 ]; then
            echo "Failed to install Ollama"
            exit 1
        fi
        echo "Ollama installed successfully"
    else
        echo "Ollama is already installed"
    fi
}

# Function to check if Ollama service is running
check_ollama_running() {
    if pgrep -x "ollama" > /dev/null; then
        echo "Ollama is already running"
        return 0
    else
        echo "Starting Ollama service..."
        # Redirect both stdout and stderr to log file
        ollama serve > logs/ollama.log 2>&1 &
        sleep 2  # Give it a moment to start
        if pgrep -x "ollama" > /dev/null; then
            echo "Ollama service started successfully"
            echo "Logs are being written to logs/ollama.log"
            return 0
        else
            echo "Failed to start Ollama service"
            return 1
        fi
    fi
}

# Main execution
echo "Setting up Ollama..."
check_ollama
check_ollama_running

# Pull the default model (optional - uncomment if you want to pull a specific model)
# echo "Pulling default model..."
# ollama pull llama2

echo "Ollama is ready to use!"
