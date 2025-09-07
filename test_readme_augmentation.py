#!/usr/bin/env python3
"""
Test script to verify README augmentation logic in dry-run mode.
"""

import os
import tempfile
from pathlib import Path

# Create a temporary assignment with existing README
def test_readme_augmentation():
    print("Testing README augmentation in dry-run mode...")
    
    # Create temporary test assignment with existing README
    test_assignment = "tests/fixtures/assignments/assignment-1"
    readme_path = Path(test_assignment) / "README.md"
    
    # Ensure directory exists
    readme_path.parent.mkdir(parents=True, exist_ok=True)
    
    # Create existing README
    existing_content = """# Assignment 1 - Existing README

This README already exists and has some content.

## Existing Instructions

These are the existing instructions for the assignment.
"""
    
    readme_path.write_text(existing_content, encoding='utf-8')
    print(f"Created existing README at {readme_path}")
    print("Existing content:")
    print(existing_content)
    
    # Now run the script in dry-run mode to see augmentation
    print("\n" + "="*50)
    print("Running script in dry-run mode...")
    os.system("DRY_RUN=true GITHUB_TOKEN=test GITHUB_REPOSITORY=test/repo python create_assignment_prs.py")
    
    # Clean up
    readme_path.unlink()
    print(f"\nCleaned up test file: {readme_path}")

if __name__ == "__main__":
    test_readme_augmentation()
