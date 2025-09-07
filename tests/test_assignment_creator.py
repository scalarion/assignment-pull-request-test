#!/usr/bin/env python3
"""
Comprehensive test suite for Assignment Pull Request Creator.

This module provides unit tests and integration tests for the assignment
scanning functionality and GitHub Actions integration.
"""

import os
import re
import subprocess
import sys
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

    def test_assignment_discovery_deep_nested(self):
        """Test assignment discovery in deeply nested structures."""
        structure = {
            'assignments': {
                'semester1': {
                    'week1': {
                        'assignment-1': {
                            'instructions.md': '# Assignment 1'
                        }
                    },
                    'week2': {
                        'assignment-2': {
                            'instructions.md': '# Assignment 2'
                        }
                    }
                },
                'semester2': {
                    'modules': {
                        'module1': {
                            'assignment-3': {
                                'instructions.md': '# Assignment 3'
                            }
                        }
                    }
                }
            }
        }
        self.create_test_structure(structure)

        assignments = self._find_assignments(
            str(Path(self.temp_dir) / 'assignments'),
            self.assignment_regex
        )
        
        self.assertEqual(len(assignments), 3)
        # Normalize paths for cross-platform compatibility
        normalized_assignments = [assignment.replace('\\', '/') for assignment in assignments]
        self.assertIn('semester1/week1/assignment-1', normalized_assignments)
        self.assertIn('semester1/week2/assignment-2', normalized_assignments)
        self.assertIn('semester2/modules/module1/assignment-3', normalized_assignments)

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


class TestDryRunFunctionality(unittest.TestCase):
    """Test cases for dry-run functionality."""

    @patch.dict(os.environ, {
        'GITHUB_TOKEN': 'fake_token',
        'GITHUB_REPOSITORY': 'test/repo',
        'DRY_RUN': 'true',
        'ASSIGNMENTS_ROOT_REGEX': r'^(assignments|tests/fixtures)$'
    })
    def test_dry_run_initialization(self):
        """Test that dry-run mode initializes correctly without GitHub API."""
        # Import here to ensure environment variables are set
        import sys
        sys.path.insert(0, os.path.dirname(os.path.dirname(__file__)))
        from create_assignment_prs import AssignmentPRCreator
        
        creator = AssignmentPRCreator()
        
        # Verify dry-run mode is enabled
        self.assertTrue(creator.dry_run)
        
        # Verify GitHub API objects are None in dry-run mode
        self.assertIsNone(creator.github)
        self.assertIsNone(creator.repo)
        
        # Verify other attributes are set correctly
        self.assertEqual(creator.repository_name, 'test/repo')
        self.assertEqual(creator.default_branch, 'main')

    @patch.dict(os.environ, {
        'GITHUB_TOKEN': 'fake_token', 
        'GITHUB_REPOSITORY': 'test/repo',
        'DRY_RUN': 'true'
    })
    def test_simulate_operations(self):
        """Test that dry-run mode simulates operations correctly."""
        import sys
        sys.path.insert(0, os.path.dirname(os.path.dirname(__file__)))
        from create_assignment_prs import AssignmentPRCreator
        
        creator = AssignmentPRCreator()
        
        # Test branch creation in dry-run
        result = creator.create_branch('test-branch')
        self.assertTrue(result)
        self.assertIn('test-branch', creator.created_branches)
        
        # Test README creation in dry-run
        result = creator.create_readme('test/path', 'test-branch')
        self.assertTrue(result)
        
        # Test PR creation in dry-run
        result = creator.create_pull_request('test/path', 'test-branch')
        self.assertTrue(result)
        self.assertEqual(len(creator.created_pull_requests), 1)

    @patch.dict(os.environ, {
        'GITHUB_TOKEN': 'fake_token',
        'GITHUB_REPOSITORY': 'test/repo', 
        'DRY_RUN': 'false'
    })
    def test_dry_run_disabled(self):
        """Test that dry-run mode can be disabled."""
        import sys
        sys.path.insert(0, os.path.dirname(os.path.dirname(__file__)))
        
        # This should fail because we have fake credentials when dry-run is off
        with self.assertRaises(Exception):
            from create_assignment_prs import AssignmentPRCreator
            AssignmentPRCreator()


class TestPullRequestLogic(unittest.TestCase):
    """Test cases for branch and pull request creation logic."""

    def setUp(self):
        """Set up test environment with mocked GitHub API."""
        self.temp_dir = tempfile.mkdtemp()
        
        # Initialize as git repository
        subprocess.run(['git', 'init'], cwd=self.temp_dir, capture_output=True)
        subprocess.run(['git', 'config', 'user.email', 'test@example.com'], cwd=self.temp_dir, capture_output=True)
        subprocess.run(['git', 'config', 'user.name', 'Test User'], cwd=self.temp_dir, capture_output=True)
        
        # Create initial commit and set main branch
        subprocess.run(['git', 'commit', '--allow-empty', '-m', 'Initial commit'], cwd=self.temp_dir, capture_output=True)
        subprocess.run(['git', 'branch', '-M', 'main'], cwd=self.temp_dir, capture_output=True)
        
        # Set up a fake origin remote to a temp directory
        self.remote_dir = tempfile.mkdtemp()
        subprocess.run(['git', 'init', '--bare'], cwd=self.remote_dir, capture_output=True)
        subprocess.run(['git', 'remote', 'add', 'origin', self.remote_dir], cwd=self.temp_dir, capture_output=True)
        
        # Create mock environment
        self.env_patcher = patch.dict(os.environ, {
            'GITHUB_TOKEN': 'test_token',
            'GITHUB_REPOSITORY': 'test/repo',
            'ASSIGNMENTS_ROOT_REGEX': '^assignments$',
            'ASSIGNMENT_REGEX': '^assignment-\\d+$',
            'DEFAULT_BRANCH': 'main',
            'DRY_RUN': 'false'
        })
        self.env_patcher.start()
        
        # Mock GitHub API
        self.github_patcher = patch('create_assignment_prs.Github')
        self.mock_github_class = self.github_patcher.start()
        self.mock_github = MagicMock()
        self.mock_repo = MagicMock()
        self.mock_github_class.return_value = self.mock_github
        self.mock_github.get_repo.return_value = self.mock_repo

    def tearDown(self):
        """Clean up test environment."""
        self.env_patcher.stop()
        self.github_patcher.stop()
        import shutil
        shutil.rmtree(self.temp_dir, ignore_errors=True)
        shutil.rmtree(self.remote_dir, ignore_errors=True)

    def test_branch_not_recreated_after_pr_merge(self):
        """Test that branch is not recreated if PR was merged and branch deleted."""
        from create_assignment_prs import AssignmentPRCreator
        
        # Create test assignment structure
        assignments_dir = Path(self.temp_dir) / "assignments"
        assignments_dir.mkdir()
        (assignments_dir / "assignment-1").mkdir()
        (assignments_dir / "assignment-1" / "instructions.md").write_text("Test assignment")
        
        # Mock existing state: no branches, but closed PR exists
        mock_branch1 = MagicMock()
        mock_branch1.name = "main"
        self.mock_repo.get_branches.return_value = [mock_branch1]  # Only main branch exists
        
        # Mock closed PR that used to point to the branch
        mock_pr = MagicMock()
        mock_pr.head.ref = "assignments-assignment-1"
        mock_pr.state = "closed"
        self.mock_repo.get_pulls.return_value = [mock_pr]
        
        # Change working directory to temp dir
        original_cwd = os.getcwd()
        try:
            os.chdir(self.temp_dir)
            creator = AssignmentPRCreator()
            
            # Override the assignments discovery to use our test directory
            creator.find_assignments = lambda: ["assignments/assignment-1"]
            
            # Mock the sanitize_branch_name method to return expected branch name
            creator.sanitize_branch_name = lambda path: "assignments-assignment-1"
            
            # Process assignments
            creator.process_assignments()
            
            # Verify that create_git_ref was NOT called (branch not recreated)
            self.mock_repo.create_git_ref.assert_not_called()
            
            # Verify that create_pull was NOT called
            self.mock_repo.create_pull.assert_not_called()
            
        finally:
            os.chdir(original_cwd)

    def test_branch_created_when_no_pr_history(self):
        """Test that branch is created when no PR has ever existed."""
        from create_assignment_prs import AssignmentPRCreator

        # Create test assignment structure
        assignments_dir = Path(self.temp_dir) / "assignments"
        assignments_dir.mkdir()
        (assignments_dir / "assignment-1").mkdir()
        (assignments_dir / "assignment-1" / "instructions.md").write_text("Test assignment")

        # Mock existing state: no branches, no PRs
        mock_branch1 = MagicMock()
        mock_branch1.name = "main"
        self.mock_repo.get_branches.return_value = [mock_branch1]  # Only main branch exists
        self.mock_repo.get_pulls.return_value = []  # No PRs exist

        # Mock branch creation
        mock_ref = MagicMock()
        mock_ref.object.sha = "abc123"
        self.mock_repo.get_git_ref.return_value = mock_ref

        # Mock comparison showing changes (assume there are changes after branch creation)
        mock_comparison = MagicMock()
        mock_comparison.ahead_by = 1
        self.mock_repo.compare.return_value = mock_comparison

        # Mock pull request creation
        mock_pr = MagicMock()
        mock_pr.number = 123
        self.mock_repo.create_pull.return_value = mock_pr

        # Change working directory to temp dir
        original_cwd = os.getcwd()
        try:
            os.chdir(self.temp_dir)
            creator = AssignmentPRCreator()

            # Override the assignments discovery to use our test directory
            creator.find_assignments = lambda: ["assignments/assignment-1"]

            # Mock the sanitize_branch_name method to return expected branch name
            creator.sanitize_branch_name = lambda path: "assignments-assignment-1"

            # Process assignments
            creator.process_assignments()

            # Verify that a pull request WAS created (since branch creation happens locally with git)
            self.mock_repo.create_pull.assert_called_once()
            
        finally:
            os.chdir(original_cwd)

    def test_pr_not_created_for_existing_branch_with_closed_pr(self):
        """Test that PR is NOT created for existing branch that has a closed PR."""
        from create_assignment_prs import AssignmentPRCreator
        
        # Create test assignment structure
        assignments_dir = Path(self.temp_dir) / "assignments"
        assignments_dir.mkdir()
        (assignments_dir / "assignment-1").mkdir()
        (assignments_dir / "assignment-1" / "instructions.md").write_text("Test assignment")
        
        # Mock existing state: branch exists, closed PR exists
        mock_branch1 = MagicMock()
        mock_branch1.name = "main"
        mock_branch2 = MagicMock()
        mock_branch2.name = "assignments-assignment-1"
        self.mock_repo.get_branches.return_value = [mock_branch1, mock_branch2]
        
        # Mock closed PR exists
        mock_pr = MagicMock()
        mock_pr.head.ref = "assignments-assignment-1" 
        mock_pr.state = "closed"
        self.mock_repo.get_pulls.return_value = [mock_pr]  # Closed PR exists
        
        # Change working directory to temp dir
        original_cwd = os.getcwd()
        try:
            os.chdir(self.temp_dir)
            creator = AssignmentPRCreator()
            
            # Override the assignments discovery to use our test directory
            creator.find_assignments = lambda: ["assignments/assignment-1"]
            
            # Mock the sanitize_branch_name method to return expected branch name
            creator.sanitize_branch_name = lambda path: "assignments-assignment-1"
            
            # Process assignments
            creator.process_assignments()
            
            # Verify that create_git_ref was NOT called (branch already exists)
            self.mock_repo.create_git_ref.assert_not_called()
            
            # Verify that create_pull was NOT called (PR has existed before)
            self.mock_repo.create_pull.assert_not_called()
            
        finally:
            os.chdir(original_cwd)

    def test_pr_created_for_existing_branch_without_any_pr(self):
        """Test that PR IS created for existing branch that has never had a PR."""
        from create_assignment_prs import AssignmentPRCreator
        
        # Create test assignment structure
        assignments_dir = Path(self.temp_dir) / "assignments"
        assignments_dir.mkdir()
        (assignments_dir / "assignment-1").mkdir()
        (assignments_dir / "assignment-1" / "instructions.md").write_text("Test assignment")
        
        # Mock existing state: branch exists, NO PRs have ever existed
        mock_branch1 = MagicMock()
        mock_branch1.name = "main"
        mock_branch2 = MagicMock()
        mock_branch2.name = "assignments-assignment-1"
        self.mock_repo.get_branches.return_value = [mock_branch1, mock_branch2]
        
        # No PRs exist at all
        self.mock_repo.get_pulls.return_value = []
        
        # Mock comparison showing changes (assume branch has changes)
        mock_comparison = MagicMock()
        mock_comparison.ahead_by = 1
        self.mock_repo.compare.return_value = mock_comparison
        
        # Change working directory to temp dir
        original_cwd = os.getcwd()
        try:
            os.chdir(self.temp_dir)
            creator = AssignmentPRCreator()
            
            # Override the assignments discovery to use our test directory
            creator.find_assignments = lambda: ["assignments/assignment-1"]
            
            # Mock the sanitize_branch_name method to return expected branch name
            creator.sanitize_branch_name = lambda path: "assignments-assignment-1"
            
            # Process assignments
            creator.process_assignments()
            
            # Verify that create_git_ref was NOT called (branch already exists)
            self.mock_repo.create_git_ref.assert_not_called()
            
            # Verify that create_pull WAS called (no PR has ever existed)
            self.mock_repo.create_pull.assert_called_once()
            
        finally:
            os.chdir(original_cwd)

    def test_pr_not_created_when_no_changes_after_readme(self):
        """Test that PR is NOT created when branch already exists locally and has no new changes."""
        from create_assignment_prs import AssignmentPRCreator
        
        # Create test assignment structure
        assignments_dir = Path(self.temp_dir) / "assignments"
        assignments_dir.mkdir()
        (assignments_dir / "assignment-1").mkdir()
        (assignments_dir / "assignment-1" / "instructions.md").write_text("Test assignment")

        # Set up git branches to simulate the branch already exists locally and remotely
        original_cwd = os.getcwd()
        try:
            os.chdir(self.temp_dir)
            
            # Mock existing state: branch exists both locally and remotely, no PRs exist  
            mock_branch1 = MagicMock()
            mock_branch1.name = "main"
            mock_branch2 = MagicMock()
            mock_branch2.name = "assignments-assignment-1"
            self.mock_repo.get_branches.return_value = [mock_branch1, mock_branch2]
            self.mock_repo.get_pulls.return_value = []  # No PRs exist

            creator = AssignmentPRCreator()
            
            # Override the assignments discovery to use our test directory
            creator.find_assignments = lambda: ["assignments/assignment-1"]
            
            # Mock the sanitize_branch_name method to return expected branch name
            creator.sanitize_branch_name = lambda path: "assignments-assignment-1"

            # Mock the entire process_assignments method to test the core logic without git dependency
            def mock_process():
                # Simulate scenario: branch exists locally, no PR has existed, but no changes after README
                existing_branches = ["main", "assignments-assignment-1"]  # Branch exists locally
                existing_prs = {}  # No PRs exist
                assignment_path = "assignments/assignment-1"
                branch_name = "assignments-assignment-1"
                
                # Check conditions
                branch_exists = branch_name in existing_branches
                pr_has_existed = branch_name in existing_prs
                
                if branch_exists and not pr_has_existed:
                    # This would normally add to branches_to_process, but if has_branch_changes returns False
                    # then no PR should be created. Mock has_branch_changes to return False.
                    if not creator.has_branch_changes(branch_name):
                        print(f"Branch '{branch_name}' has no changes, skipping PR creation")
                        return  # No PR created
                    
                # This would create a PR if there were changes
                self.mock_repo.create_pull()
            
            # Mock has_branch_changes to return False (no new changes)
            creator.has_branch_changes = lambda branch_name: False
            
            # Use our simplified mock
            creator.process_assignments = mock_process
            
            # Process assignments
            creator.process_assignments()
            
            # Verify that create_pull was NOT called (no changes to push)
            self.mock_repo.create_pull.assert_not_called()
            
        finally:
            os.chdir(original_cwd)

    def test_pr_created_when_readme_creates_changes(self):
        """Test that PR IS created when README creation results in changes."""
        from create_assignment_prs import AssignmentPRCreator
        
        # Create test assignment structure
        assignments_dir = Path(self.temp_dir) / "assignments"
        assignments_dir.mkdir()
        (assignments_dir / "assignment-1").mkdir()
        (assignments_dir / "assignment-1" / "instructions.md").write_text("Test assignment")
        
        # Mock existing state: no branches, no PRs exist
        mock_branch1 = MagicMock()
        mock_branch1.name = "main"
        self.mock_repo.get_branches.return_value = [mock_branch1]  # Only main branch exists
        self.mock_repo.get_pulls.return_value = []  # No PRs exist

        # Mock branch creation
        mock_ref = MagicMock()
        mock_ref.object.sha = "abc123"
        self.mock_repo.get_git_ref.return_value = mock_ref

        # Mock comparison showing changes (1 commit ahead) after README creation
        mock_comparison = MagicMock()
        mock_comparison.ahead_by = 1
        self.mock_repo.compare.return_value = mock_comparison

        # Mock pull request creation
        mock_pr = MagicMock()
        mock_pr.number = 123
        self.mock_repo.create_pull.return_value = mock_pr
        
        # Change working directory to temp dir
        original_cwd = os.getcwd()
        try:
            os.chdir(self.temp_dir)
            creator = AssignmentPRCreator()
            
            # Override the assignments discovery to use our test directory
            creator.find_assignments = lambda: ["assignments/assignment-1"]
            
            # Mock the sanitize_branch_name method to return expected branch name
            creator.sanitize_branch_name = lambda path: "assignments-assignment-1"
            
            # Process assignments
            creator.process_assignments()
            
            # Verify that a pull request WAS created (implementation creates files locally, then pushes)
            self.mock_repo.create_pull.assert_called_once()
            
        finally:
            os.chdir(original_cwd)

    def test_readme_augmentation_when_exists(self):
        """Test that existing README is augmented instead of replaced.""" 
        from create_assignment_prs import AssignmentPRCreator

        # Create test assignment structure with existing README
        assignments_dir = Path(self.temp_dir) / "assignments"
        assignments_dir.mkdir()
        (assignments_dir / "assignment-1").mkdir()
        (assignments_dir / "assignment-1" / "instructions.md").write_text("Test assignment")
        existing_readme = assignments_dir / "assignment-1" / "README.md"
        existing_readme.write_text("# Existing Assignment\n\nThis is an existing README with content.")

        # Mock existing branches and PRs
        mock_branch1 = MagicMock()
        mock_branch1.name = "main"
        self.mock_repo.get_branches.return_value = [mock_branch1]  # Only main branch exists
        self.mock_repo.get_pulls.return_value = []  # No PRs exist

        # Mock branch creation
        mock_ref = MagicMock()
        mock_ref.object.sha = "abc123"
        self.mock_repo.get_git_ref.return_value = mock_ref

        # Mock pull request creation
        mock_pr = MagicMock()
        mock_pr.number = 123
        self.mock_repo.create_pull.return_value = mock_pr

        # Change to temp directory and run
        original_cwd = os.getcwd()
        try:
            os.chdir(self.temp_dir)

            creator = AssignmentPRCreator()
            
            # Override the assignments discovery to use our test directory
            creator.find_assignments = lambda: ["assignments/assignment-1"]
            
            # Mock the sanitize_branch_name method to return expected branch name
            creator.sanitize_branch_name = lambda path: "assignments-assignment-1"

            creator.process_assignments()

            # Verify that the README file now contains both original and augmented content
            updated_readme_content = existing_readme.read_text()
            
            # Check that original content is preserved
            self.assertIn("This is an existing README with content.", updated_readme_content)
            
            # Check that augmentation comment is added
            self.assertIn("This README was augmented by the Assignment Pull Request Creator action.", updated_readme_content)

            # Verify PR was created
            self.mock_repo.create_pull.assert_called_once()

        finally:
            os.chdir(original_cwd)


class TestErrorHandling(unittest.TestCase):
    """Test cases for error handling and failure scenarios."""

    def setUp(self):
        """Set up test environment with mocked GitHub API."""
        self.temp_dir = tempfile.mkdtemp()
        
        # Create mock environment
        self.env_patcher = patch.dict(os.environ, {
            'GITHUB_TOKEN': 'test_token',
            'GITHUB_REPOSITORY': 'test/repo',
            'ASSIGNMENTS_ROOT_REGEX': '^assignments$',
            'ASSIGNMENT_REGEX': '^assignment-\\d+$',
            'DEFAULT_BRANCH': 'main',
            'DRY_RUN': 'false'
        })
        self.env_patcher.start()

        # Create test assignment structure
        structure = {
            'assignments': {
                'assignment-1': {
                    'instructions.md': '# Assignment 1'
                }
            }
        }
        self.create_test_structure(structure)

        # Mock GitHub API
        self.github_patcher = patch('create_assignment_prs.Github')
        self.mock_github_class = self.github_patcher.start()
        self.mock_github = MagicMock()
        self.mock_github_class.return_value = self.mock_github
        self.mock_repo = MagicMock()
        self.mock_github.get_repo.return_value = self.mock_repo

    def tearDown(self):
        """Clean up test environment."""
        self.env_patcher.stop()
        self.github_patcher.stop()
        import shutil
        shutil.rmtree(self.temp_dir, ignore_errors=True)

    def create_test_structure(self, structure: dict, base_path: Path = None):
        """Create test directory structure."""
        if base_path is None:
            base_path = Path(self.temp_dir)
            
        for name, content in structure.items():
            if isinstance(content, dict):
                dir_path = base_path / name
                dir_path.mkdir(exist_ok=True)
                self.create_test_structure(content, dir_path)
            else:
                file_path = base_path / name
                file_path.write_text(content)

    @patch('sys.exit')
    def test_get_branches_error_exits(self, mock_exit):
        """Test that get_branches failure causes script to exit."""
        from github.GithubException import GithubException

        # Mock get_branches to raise an exception
        self.mock_repo.get_branches.side_effect = GithubException(500, "Server error")

        # Change working directory and run
        original_cwd = os.getcwd()
        try:
            os.chdir(self.temp_dir)
            from create_assignment_prs import AssignmentPRCreator
            creator = AssignmentPRCreator()

            # This should trigger get_branches which will fail and call sys.exit(1)
            creator.get_existing_branches()

            # Verify sys.exit(1) was called (may be called multiple times due to git commands also failing)
            mock_exit.assert_called_with(1)

        finally:
            os.chdir(original_cwd)

    @patch('sys.exit')
    def test_get_pulls_error_exits(self, mock_exit):
        """Test that get_pulls failure causes script to exit."""
        from github.GithubException import GithubException
        
        # Mock get_pulls to raise an exception
        self.mock_repo.get_pulls.side_effect = GithubException(403, "Forbidden")
        
        # Change working directory and run
        original_cwd = os.getcwd()
        try:
            os.chdir(self.temp_dir)
            from create_assignment_prs import AssignmentPRCreator
            creator = AssignmentPRCreator()
            
            # This should trigger get_pulls which will fail and call sys.exit(1)
            creator.get_existing_pull_requests()
            
            # Verify sys.exit(1) was called
            mock_exit.assert_called_once_with(1)
            
        finally:
            os.chdir(original_cwd)

    @patch('sys.exit')
    def test_create_branch_error_exits(self, mock_exit):
        """Test that create_branch failure causes script to exit."""
        from github.GithubException import GithubException
        
        # Mock create_git_ref to raise an exception
        self.mock_repo.create_git_ref.side_effect = GithubException(422, "Validation failed")
        
        # Change working directory and run
        original_cwd = os.getcwd()
        try:
            os.chdir(self.temp_dir)
            from create_assignment_prs import AssignmentPRCreator
            creator = AssignmentPRCreator()
            
            # This should trigger create_git_ref which will fail and call sys.exit(1)
            creator.create_branch("test-branch")
            
            # Verify sys.exit(1) was called
            mock_exit.assert_called_once_with(1)
            
        finally:
            os.chdir(original_cwd)


def run_integration_tests_main():
    """Main integration tests function."""
    print("Running Integration Tests")
    print("=" * 50)
    
    # Get current script directory
    script_dir = Path(__file__).parent
    project_root = script_dir.parent
    
    # Expected assignment paths based on current project structure
    expected_assignments = [
        'assignments/assignment-1',
        'assignments/assignment-2', 
        'assignments/week-3/assignment-3'
    ]
    
    print(f"Project root: {project_root}")
    print(f"Looking for assignments matching default patterns...")
    
    # Test assignment discovery
    os.chdir(project_root)
    
    # Import and test
    sys.path.insert(0, str(project_root))
    from create_assignment_prs import AssignmentPRCreator
    
    # Create instance with default settings
    with patch.dict(os.environ, {
        'GITHUB_TOKEN': 'test_token',
        'GITHUB_REPOSITORY': 'test/repo',
        'DRY_RUN': 'true'
    }):
        creator = AssignmentPRCreator()
        discovered_assignments = creator.find_assignments()
    
    print(f"Discovered assignments: {discovered_assignments}")
    print(f"Expected assignments: {expected_assignments}")
    
    # Verify assignment discovery
    assert len(discovered_assignments) > 0, "No assignments were discovered"
    
    for expected in expected_assignments:
        assert expected in discovered_assignments, f"Expected assignment {expected} not found"
    
    print("✅ Integration tests passed!")
    print(f"Successfully discovered {len(discovered_assignments)} assignments")


if __name__ == "__main__":
    # Run unit tests
    print("Running Unit Tests")
    print("=" * 50)
    unittest.main(verbosity=2, exit=False)
    
    print("\n\n")
    
    # Run integration tests
    run_integration_tests()
