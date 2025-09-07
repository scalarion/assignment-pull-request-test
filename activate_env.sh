#!/bin/bash
# Activation script for the Assignment Pull Request Creator conda environment

echo "Activating assignment-pr-creator conda environment..."
source /opt/conda/etc/profile.d/conda.sh
conda activate assignment-pr-creator

echo "Environment activated! 🚀"
echo "Python version: $(python --version)"
echo "Available packages:"
echo "  - PyGithub: $(python -c 'import github; print(github.__version__)')"
echo "  - requests: $(python -c 'import requests; print(requests.__version__)')"
echo ""
echo "Ready to run assignment pull request creator!"
echo ""
echo "To run the script locally for testing:"
echo "  python create_assignment_prs.py"
echo ""
echo "To format code:"
echo "  black create_assignment_prs.py --line-length=79"
echo ""
echo "To check code style:"
echo "  flake8 create_assignment_prs.py --max-line-length=79"
echo ""
echo "To run type checking:"
echo "  mypy create_assignment_prs.py --ignore-missing-imports"
