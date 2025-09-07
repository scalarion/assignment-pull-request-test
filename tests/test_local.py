#!/usr/bin/env python3
"""
Quick local test runner for Assignment Pull Request Creator.

This script provides a simple way to test assignment discovery and
branch name sanitization without running the full test suite.
"""

import os
import re
import sys
from pathlib import Path


def sanitize_branch_name(assignment_path: str) -> str:
    """
    Sanitize assignment path to create a valid branch name.
    
    Args:
        assignment_path: Relative path of assignment from assignments folder
        
    Returns:
        Sanitized branch name
    """
    branch_name = assignment_path.strip()
    branch_name = re.sub(r'\s+', '-', branch_name)
    branch_name = branch_name.replace('/', '-')
    branch_name = re.sub(r'-+', '-', branch_name)
    branch_name = branch_name.lower()
    branch_name = branch_name.strip('-')
    return branch_name


def find_assignments_with_root_pattern(workspace_root: Path, root_regex: str, assignment_regex: str) -> list:
    """
    Find assignments using both root and assignment patterns.
    
    Args:
        workspace_root: Root workspace directory
        root_regex: Pattern for assignment root directories
        assignment_regex: Pattern for assignment directories
        
    Returns:
        List of relative paths to assignment folders
    """
    assignments = []
    root_pattern = re.compile(root_regex)
    assignment_pattern = re.compile(assignment_regex)
    
    for root, dirs, _ in os.walk(workspace_root):
        root_path = Path(root)
        
        for dir_name in dirs:
            if root_pattern.match(dir_name):
                assignments_root = root_path / dir_name
                
                if assignments_root.exists():
                    for assignment_root, assignment_dirs, _ in os.walk(assignments_root):
                        assignment_root_path = Path(assignment_root)
                        
                        for assignment_dir in assignment_dirs:
                            if assignment_pattern.match(assignment_dir):
                                full_assignment_path = assignment_root_path / assignment_dir
                                relative_path = full_assignment_path.relative_to(workspace_root)
                                assignments.append(str(relative_path))
    
    return assignments


def test_assignment_scanning():
    """Test the assignment scanning functionality without GitHub API calls."""
    
    # Get configuration from environment or use defaults
    assignments_root_regex = os.environ.get('ASSIGNMENTS_ROOT_REGEX', '^assignments$')
    assignment_regex = os.environ.get('ASSIGNMENT_REGEX', r'^assignment-\d+$')
    
    print("Testing Assignment Pull Request Creator")
    print("=" * 50)
    print(f"Root regex: {assignments_root_regex}")
    print(f"Assignment regex: {assignment_regex}")
    
    # Test assignment scanning using fixtures only
    fixtures_root = Path("fixtures")
    assignments = find_assignments_with_root_pattern(
        fixtures_root, 
        assignments_root_regex, 
        assignment_regex
    )
    
    print(f"\nFound {len(assignments)} assignments:")
    for assignment in assignments:
        print(f"  - {assignment}")
        
        # Show what the branch name would be
        branch_name = sanitize_branch_name(assignment)
        print(f"    -> branch: {branch_name}")
    
    # Test with fixtures using relative paths (for display purposes)
    if fixtures_root.exists():
        print(f"\nTesting with fixtures:")
        fixture_assignments = find_assignments_with_root_pattern(
            fixtures_root,
            assignments_root_regex,
            assignment_regex
        )
        for assignment in fixture_assignments:
            print(f"  - fixtures/{assignment}")
    
    # Test branch name sanitization with various examples
    print(f"\nBranch name sanitization examples:")
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
    return assignments


def main():
    """Main function to run tests or specific operations."""
    if len(sys.argv) > 1:
        command = sys.argv[1].lower()
        
        if command == "discover":
            # Just run discovery using fixtures
            assignments = find_assignments_with_root_pattern(
                Path("fixtures"),
                os.environ.get('ASSIGNMENTS_ROOT_REGEX', '^assignments$'),
                os.environ.get('ASSIGNMENT_REGEX', r'^assignment-\d+$')
            )
            for assignment in assignments:
                print(assignment)
        elif command == "sanitize":
            # Test branch name sanitization
            if len(sys.argv) > 2:
                test_path = sys.argv[2]
                print(sanitize_branch_name(test_path))
            else:
                print("Usage: test_local.py sanitize <path>")
        else:
            print("Usage: test_local.py [discover|sanitize <path>]")
    else:
        # Run full test
        test_assignment_scanning()


if __name__ == "__main__":
    main()
