#!/usr/bin/env python3
"""
Test script for the Assignment Pull Request Creator.

This script helps test the assignment scanning functionality locally without GitHub dependencies.
"""

import os
import re
from pathlib import Path


def sanitize_branch_name(assignment_path: str) -> str:
    """
    Sanitize assignment path to create a valid branch name.
    
    Args:
        assignment_path: Relative path of assignment from assignments folder
        
    Returns:
        Sanitized branch name
    """
    # Remove leading/trailing whitespace
    branch_name = assignment_path.strip()
    
    # Replace spaces with hyphens
    branch_name = re.sub(r'\s+', '-', branch_name)
    
    # Remove slashes
    branch_name = branch_name.replace('/', '-')
    
    # Remove consecutive hyphens
    branch_name = re.sub(r'-+', '-', branch_name)
    
    # Convert to lowercase
    branch_name = branch_name.lower()
    
    # Remove leading/trailing hyphens
    branch_name = branch_name.strip('-')
    
    return branch_name


def find_assignments(assignments_folder: str, assignment_regex: str) -> list:
    """
    Find all assignment folders that match the regex pattern.
    
    Args:
        assignments_folder: Root folder to scan
        assignment_regex: Regex pattern to match
        
    Returns:
        List of relative paths to assignment folders
    """
    assignments = []
    assignments_root = Path(assignments_folder)
    
    if not assignments_root.exists():
        print(f"Assignments folder '{assignments_folder}' does not exist")
        return assignments
    
    print(f"Scanning for assignments in '{assignments_folder}' with pattern '{assignment_regex}'")
    
    # Compile regex pattern
    pattern = re.compile(assignment_regex)
    
    # Walk through all subdirectories
    for root, dirs, files in os.walk(assignments_root):
        for dir_name in dirs:
            if pattern.match(dir_name):
                # Get relative path from assignments folder
                full_path = Path(root) / dir_name
                relative_path = full_path.relative_to(assignments_root)
                assignments.append(str(relative_path))
                print(f"Found assignment: {relative_path}")
    
    return assignments


def test_assignment_scanning():
    """Test the assignment scanning functionality without GitHub API calls."""
    
    assignments_folder = 'assignments'
    assignment_regex = r'^assignment-\d+$'
    
    print("Testing Assignment Pull Request Creator")
    print("=" * 50)
    
    # Test assignment scanning
    assignments = find_assignments(assignments_folder, assignment_regex)
    print(f"\nFound {len(assignments)} assignments:")
    for assignment in assignments:
        print(f"  - {assignment}")
    
    # Test branch name sanitization
    print(f"\nBranch name sanitization:")
    test_paths = [
        "assignment-1",
        "assignment-2", 
        "week-3/assignment-3",
        "Module 4/Lab Assignment",
        "  spaced  assignment  ",
        "assignment/with/many/slashes",
        "UPPERCASE-assignment"
    ]
    
    for path in test_paths:
        sanitized = sanitize_branch_name(path)
        print(f"  '{path}' -> '{sanitized}'")
    
    print(f"\nTest completed successfully!")


if __name__ == "__main__":
    test_assignment_scanning()
