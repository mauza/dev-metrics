# dev-metrics

A demonstration of why common developer productivity metrics are flawed and easily gameable. This project illustrates why leaders should focus on meaningful outcomes rather than vanity metrics.

## Features

### Git Activity Generator
A CLI tool that demonstrates how git metrics can be artificially inflated:
- Generates realistic-looking code changes across multiple repositories
- Creates convincing commit messages using LLM
- Configurable commit frequency and patterns
- Supports multiple file types and coding styles

### Local LLM Integration
- Uses local LLM (like llama.cpp) to generate realistic changes
- Customizable prompts for different types of commits
- Maintains consistent coding style with existing codebase
- Generates contextually appropriate commit messages

## Setup

1. Install dependencies using uv:
```bash
# Install uv if you haven't already
curl -LsSf https://astral.sh/uv/install.sh | sh

# Install dependencies
uv pip install -e .
```

2. Download the LLM model:
```bash
# Download the recommended model (4.31GB)
python setup_model.py

# Or choose a different size
python setup_model.py --model-size tiny   # Smaller, faster (2.87GB)
python setup_model.py --model-size medium # Better quality (5.04GB)
```

3. Configure your repositories in `config.yaml`

## Usage

Basic usage:
```bash
python dev_metrics.py generate --repo-path /path/to/repo --frequency daily
```

See `python dev_metrics.py --help` for more options.

## Configuration

Create a `config.yaml` file:
```yaml
repositories:
  - path: /path/to/repo1
    patterns:
      - "*.py"
      - "*.js"
  - path: /path/to/repo2
    patterns:
      - "*.go"

llm:
  model_path: ./models/llama-7b.gguf
  temperature: 0.7
```

## Disclaimer

This tool is meant for educational purposes to demonstrate the flaws in using commit metrics for performance evaluation. Use responsibly and in accordance with your workplace policies.
