#!/bin/bash
# setup-devcontainer.sh - Script to set up the development environment

set -e

echo "🚀 Setting up Assignment Pull Request Creator development environment..."

# Check if conda is available
if ! command -v conda &> /dev/null; then
    echo "❌ Conda is not available in this container"
    echo "Installing dependencies with pip instead..."
    pip install -r requirements-dev.txt
    exit 0
fi

# Create conda environment if it doesn't exist
if conda env list | grep -q "assignment-pr-creator"; then
    echo "✅ Conda environment 'assignment-pr-creator' already exists"
else
    echo "📦 Creating conda environment from environment.yml..."
    conda env create -f environment.yml
fi

# Activate environment and install development dependencies
echo "🔧 Activating environment and installing development tools..."
source /opt/conda/etc/profile.d/conda.sh
conda activate assignment-pr-creator

# Install additional development tools if not already installed
pip install -r requirements-dev.txt

echo "✅ Development environment setup complete!"
echo ""
echo "To activate the environment manually, run:"
echo "  source /opt/conda/etc/profile.d/conda.sh"
echo "  conda activate assignment-pr-creator"
echo ""
echo "Or use the activation script:"
echo "  ./activate_env.sh"
