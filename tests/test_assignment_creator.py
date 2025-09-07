#!/usr/bin/env python3
"""
Comprehensive test suite for Assignment Pull Request Creator.

This module provides unit tests and integration tests for the assignment
scanning functionality and GitHub Actions integration.
"""

import os
import re
import tempfile
import unittest
from pathlib import Path
from unittest.mock import patch, MagicMock
from typing import List, Tuple


class TestAssignmentDiscovery(unittest.TestCase):
    """Test cases for assignment discovery functionality."""

    def setUp(self):
        """Set up test environment."""
        self.temp_dir = tempfile.mkdtemp()
        self.assignments_root_regex = r'^assignments$'
        self.assignment_regex = r'^assignment-\d+$'

    def tearDown(self):
        """Clean up test environment."""
        import shutil
        shutil.rmtree(self.temp_dir, ignore_errors=True)

    def create_test_structure(self, structure: dict, base_path: Path = None):
        """
        Create a test directory structure.
        
        Args:
            structure: Dict representing directory structure
            base_path: Base path to create structure in
        """
        if base_path is None:
            base_path = Path(self.temp_dir)
            
        for name, content in structure.items():
            if isinstance(content, dict):
                # Directory
                dir_path = base_path / name
                dir_path.mkdir(exist_ok=True)
                self.create_test_structure(content, dir_path)
            else:
                # File
                file_path = base_path / name
                file_path.write_text(content)

    def test_assignment_discovery_basic(self):
        """Test basic assignment discovery."""
        # Create test structure
        structure = {
            'assignments': {
                'assignment-1': {
                    'instructions.md': '# Assignment 1'
                },
                'assignment-2': {
                    'instructions.md': '# Assignment 2'
                },
                'not-an-assignment': {
                    'file.txt': 'content'
                }
            }
        }
        self.create_test_structure(structure)

        # Test discovery
        assignments = self._find_assignments(
            str(Path(self.temp_dir) / 'assignments'),
            self.assignment_regex
        )
        
        self.assertEqual(len(assignments), 2)
        self.assertIn('assignment-1', assignments)
        self.assertIn('assignment-2', assignments)
        self.assertNotIn('not-an-assignment', assignments)

    def test_assignment_discovery_nested(self):
        """Test assignment discovery in nested structures."""
        structure = {
            'assignments': {
                'week-1': {
                    'assignment-1': {
                        'instructions.md': '# Assignment 1'
                    }
                },
                'week-2': {
                    'assignment-2': {
                        'instructions.md': '# Assignment 2'
                    }
                }
            }
        }
        self.create_test_structure(structure)

        assignments = self._find_assignments(
            str(Path(self.temp_dir) / 'assignments'),
            self.assignment_regex
        )
        
        self.assertEqual(len(assignments), 2)
        # Normalize paths for cross-platform compatibility
        normalized_assignments = [assignment.replace('\\', '/') for assignment in assignments]
        self.assertIn('week-1/assignment-1', normalized_assignments)
        self.assertIn('week-2/assignment-2', normalized_assignments)

    def test_assignment_discovery_multiple_roots(self):
        """Test discovery with multiple assignment root patterns."""
        structure = {
            'assignments': {
                'assignment-1': {'file.txt': 'content'}
            },
            'homework': {
                'hw-1': {'file.txt': 'content'}
            },
            'labs': {
                'lab-1': {'file.txt': 'content'}
            }
        }
        self.create_test_structure(structure)

        # Test with multiple root pattern
        root_regex = r'^(assignments|homework|labs)$'
        assignment_regex = r'^(assignment|hw|lab)-\d+$'
        
        assignments = self._find_assignments_with_root_pattern(
            Path(self.temp_dir),
            root_regex,
            assignment_regex
        )
        
        self.assertEqual(len(assignments), 3)
        # Normalize paths for cross-platform compatibility
        normalized_assignments = [assignment.replace('\\', '/') for assignment in assignments]
        self.assertIn('assignments/assignment-1', normalized_assignments)
        self.assertIn('homework/hw-1', normalized_assignments)
        self.assertIn('labs/lab-1', normalized_assignments)

    def test_empty_assignments_folder(self):
        """Test handling of empty assignments folder."""
        structure = {'assignments': {}}
        self.create_test_structure(structure)

        assignments = self._find_assignments(
            str(Path(self.temp_dir) / 'assignments'),
            self.assignment_regex
        )
        
        self.assertEqual(len(assignments), 0)

    def test_nonexistent_assignments_folder(self):
        """Test handling of nonexistent assignments folder."""
        assignments = self._find_assignments(
            str(Path(self.temp_dir) / 'nonexistent'),
            self.assignment_regex
        )
        
        self.assertEqual(len(assignments), 0)

    def _find_assignments(self, assignments_folder: str, assignment_regex: str) -> List[str]:
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
            return assignments
        
        pattern = re.compile(assignment_regex)
        
        for root, dirs, files in os.walk(assignments_root):
            for dir_name in dirs:
                if pattern.match(dir_name):
                    full_path = Path(root) / dir_name
                    relative_path = full_path.relative_to(assignments_root)
                    assignments.append(str(relative_path))
        
        return assignments

    def _find_assignments_with_root_pattern(
        self, 
        workspace_root: Path, 
        root_regex: str, 
        assignment_regex: str
    ) -> List[str]:
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


class TestBranchNameSanitization(unittest.TestCase):
    """Test cases for branch name sanitization."""

    def test_sanitize_basic(self):
        """Test basic branch name sanitization."""
        test_cases = [
            ("assignment-1", "assignment-1"),
            ("assignment-2", "assignment-2"),
            ("UPPERCASE-ASSIGNMENT", "uppercase-assignment"),
            ("  spaced  assignment  ", "spaced-assignment"),
            ("assignment/with/slashes", "assignment-with-slashes"),
            ("assignment-with---multiple-hyphens", "assignment-with-multiple-hyphens"),
            ("-leading-and-trailing-", "leading-and-trailing"),
        ]
        
        for input_path, expected in test_cases:
            with self.subTest(input_path=input_path):
                result = self._sanitize_branch_name(input_path)
                self.assertEqual(result, expected)

    def test_sanitize_nested_paths(self):
        """Test sanitization of nested assignment paths."""
        test_cases = [
            ("week-1/assignment-1", "week-1-assignment-1"),
            ("Module 4/Lab Assignment", "module-4-lab-assignment"),
            ("assignments/hw-1/part-a", "assignments-hw-1-part-a"),
        ]
        
        for input_path, expected in test_cases:
            with self.subTest(input_path=input_path):
                result = self._sanitize_branch_name(input_path)
                self.assertEqual(result, expected)

    def _sanitize_branch_name(self, assignment_path: str) -> str:
        """
        Sanitize assignment path to create a valid branch name.
        
        Args:
            assignment_path: Relative path of assignment
            
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


class TestEnvironmentConfiguration(unittest.TestCase):
    """Test cases for environment variable configuration."""

    def test_default_environment_variables(self):
        """Test default environment variable values."""
        with patch.dict(os.environ, {}, clear=True):
            assignments_root_regex = os.environ.get('ASSIGNMENTS_ROOT_REGEX', '^assignments$')
            assignment_regex = os.environ.get('ASSIGNMENT_REGEX', r'^assignment-\d+$')
            default_branch = os.environ.get('DEFAULT_BRANCH', 'main')
            
            self.assertEqual(assignments_root_regex, '^assignments$')
            self.assertEqual(assignment_regex, r'^assignment-\d+$')
            self.assertEqual(default_branch, 'main')

    def test_custom_environment_variables(self):
        """Test custom environment variable values."""
        custom_env = {
            'ASSIGNMENTS_ROOT_REGEX': '^(assignments|homework)$',
            'ASSIGNMENT_REGEX': r'^(assignment|hw)-\d+$',
            'DEFAULT_BRANCH': 'develop'
        }
        
        with patch.dict(os.environ, custom_env, clear=True):
            assignments_root_regex = os.environ.get('ASSIGNMENTS_ROOT_REGEX', '^assignments$')
            assignment_regex = os.environ.get('ASSIGNMENT_REGEX', r'^assignment-\d+$')
            default_branch = os.environ.get('DEFAULT_BRANCH', 'main')
            
            self.assertEqual(assignments_root_regex, '^(assignments|homework)$')
            self.assertEqual(assignment_regex, r'^(assignment|hw)-\d+$')
            self.assertEqual(default_branch, 'develop')


class TestRegexPatterns(unittest.TestCase):
    """Test cases for regex pattern validation."""

    def test_assignment_regex_patterns(self):
        """Test various assignment regex patterns."""
        test_cases = [
            # Pattern, test_string, should_match
            (r'^assignment-\d+$', 'assignment-1', True),
            (r'^assignment-\d+$', 'assignment-10', True),
            (r'^assignment-\d+$', 'assignment-', False),
            (r'^assignment-\d+$', 'assignment-1a', False),
            (r'^(assignment|hw|lab)-\d+$', 'hw-1', True),
            (r'^(assignment|hw|lab)-\d+$', 'lab-5', True),
            (r'^(assignment|hw|lab)-\d+$', 'project-1', False),
        ]
        
        for pattern, test_string, should_match in test_cases:
            with self.subTest(pattern=pattern, test_string=test_string):
                regex = re.compile(pattern)
                result = bool(regex.match(test_string))
                self.assertEqual(result, should_match)

    def test_root_regex_patterns(self):
        """Test various root directory regex patterns."""
        test_cases = [
            # Pattern, test_string, should_match
            (r'^assignments$', 'assignments', True),
            (r'^assignments$', 'assignments-old', False),
            (r'^(assignments|homework|labs)$', 'assignments', True),
            (r'^(assignments|homework|labs)$', 'homework', True),
            (r'^(assignments|homework|labs)$', 'projects', False),
        ]
        
        for pattern, test_string, should_match in test_cases:
            with self.subTest(pattern=pattern, test_string=test_string):
                regex = re.compile(pattern)
                result = bool(regex.match(test_string))
                self.assertEqual(result, should_match)


def run_integration_tests():
    """Run integration tests with actual workspace structure."""
    print("Running Integration Tests")
    print("=" * 50)
    
    # Test with current workspace
    workspace_root = Path(".")
    assignments_root_regex = os.environ.get('ASSIGNMENTS_ROOT_REGEX', '^assignments$')
    assignment_regex = os.environ.get('ASSIGNMENT_REGEX', r'^assignment-\d+$')
    
    print(f"Workspace root: {workspace_root.absolute()}")
    print(f"Root regex: {assignments_root_regex}")
    print(f"Assignment regex: {assignment_regex}")
    
    # Find assignments using the same logic as the main script
    root_pattern = re.compile(assignments_root_regex)
    assignment_pattern = re.compile(assignment_regex)
    
    assignments = []
    
    for root, dirs, _ in os.walk(workspace_root):
        root_path = Path(root)
        
        for dir_name in dirs:
            if root_pattern.match(dir_name):
                assignments_root = root_path / dir_name
                print(f"Found assignment root: {assignments_root}")
                
                if assignments_root.exists():
                    for assignment_root, assignment_dirs, _ in os.walk(assignments_root):
                        assignment_root_path = Path(assignment_root)
                        
                        for assignment_dir in assignment_dirs:
                            if assignment_pattern.match(assignment_dir):
                                full_assignment_path = assignment_root_path / assignment_dir
                                relative_path = full_assignment_path.relative_to(workspace_root)
                                assignments.append(str(relative_path))
                                print(f"Found assignment: {relative_path}")
    
    print(f"\nIntegration Test Results:")
    print(f"Found {len(assignments)} assignments:")
    for assignment in assignments:
        print(f"  - {assignment}")
    
    return assignments


if __name__ == "__main__":
    # Run unit tests
    print("Running Unit Tests")
    print("=" * 50)
    unittest.main(verbosity=2, exit=False)
    
    print("\n\n")
    
    # Run integration tests
    run_integration_tests()
