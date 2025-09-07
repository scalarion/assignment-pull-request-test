#!/usr/bin/env python3
"""
Integration tests for Assignment Pull Request Creator with comprehensive mocking.

This module provides integration tests that test the complete workflow
with proper mocking of all external dependencies.
"""

import os
import sys
import unittest
import tempfile
import shutil
from unittest.mock import patch, Mock

# Add parent directory to path
sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from create_assignment_prs import AssignmentPRCreator
from tests.test_config import TestConfig, GitCommandMocker, GitHubAPIMocker, TestScenarios, create_temp_workspace


class TestAssignmentPRCreatorIntegration(unittest.TestCase):
    """Integration tests with comprehensive mocking."""
    
    def setUp(self):
        """Set up test environment."""
        self.temp_workspace = None
        self.git_mocker = GitCommandMocker()
        self.github_mocker = GitHubAPIMocker()
        
        # Start patches
        self.env_patcher = patch.dict(os.environ, TestConfig.DEFAULT_ENV)
        self.env_patcher.start()
        
        self.github_patcher = patch('create_assignment_prs.Github')
        self.mock_github_class = self.github_patcher.start()
        self.mock_github_class.return_value = self.github_mocker.get_github_mock()
        
        self.subprocess_patcher = patch('create_assignment_prs.subprocess')
        self.mock_subprocess = self.subprocess_patcher.start()
        self.mock_subprocess.run.side_effect = self.git_mocker.mock_subprocess_run
        
    def tearDown(self):
        """Clean up test environment."""
        self.env_patcher.stop()
        self.github_patcher.stop()
        self.subprocess_patcher.stop()
        
        if self.temp_workspace and os.path.exists(self.temp_workspace):
            shutil.rmtree(self.temp_workspace, ignore_errors=True)
    
    @patch('create_assignment_prs.os.getcwd')
    @patch('create_assignment_prs.os.walk')
    def test_clean_repository_workflow(self, mock_walk, mock_getcwd):
        """Test complete workflow on a clean repository."""
        # Set up workspace
        self.temp_workspace = create_temp_workspace('simple')
        mock_getcwd.return_value = self.temp_workspace
        
        # Configure os.walk mock to return the assignment structure
        mock_walk.return_value = [
            (self.temp_workspace, ['assignments'], []),
            (os.path.join(self.temp_workspace, 'assignments'), ['assignment-1', 'assignment-2'], []),
            (os.path.join(self.temp_workspace, 'assignments', 'assignment-1'), [], ['instructions.md']),
            (os.path.join(self.temp_workspace, 'assignments', 'assignment-2'), [], ['instructions.md']),
        ]
        
        # Test scenario
        scenario = TestScenarios.clean_repository()
        
        # Create and run the tool
        creator = AssignmentPRCreator()
        creator.process_assignments()
        
        # Verify results
        self.assertEqual(len(creator.created_branches), scenario['expected_new_branches'])
        self.assertEqual(len(creator.created_pull_requests), scenario['expected_new_prs'])
        
        # Verify git commands were called
        self.assertGreater(self.git_mocker.get_call_count('git fetch --all'), 0)
        self.assertGreater(self.git_mocker.get_call_count('git push --all origin'), 0)
        
        print(f"✅ {scenario['name']}: Created {len(creator.created_branches)} branches and {len(creator.created_pull_requests)} PRs")
    
    @patch('create_assignment_prs.os.getcwd')
    @patch('create_assignment_prs.os.walk')
    def test_existing_branches_workflow(self, mock_walk, mock_getcwd):
        """Test workflow with existing branches."""
        # Set up workspace
        self.temp_workspace = create_temp_workspace('simple')
        mock_getcwd.return_value = self.temp_workspace
        
        # Configure os.walk mock
        mock_walk.return_value = [
            (self.temp_workspace, ['assignments'], []),
            (os.path.join(self.temp_workspace, 'assignments'), ['assignment-1', 'assignment-2'], []),
            (os.path.join(self.temp_workspace, 'assignments', 'assignment-1'), [], ['instructions.md']),
            (os.path.join(self.temp_workspace, 'assignments', 'assignment-2'), [], ['instructions.md']),
        ]
        
        # Test scenario
        scenario = TestScenarios.existing_branches()
        
        # Mock existing branches
        def mock_get_existing_branches():
            return scenario['existing_branches']
        
        # Create and configure the tool
        creator = AssignmentPRCreator()
        creator.get_existing_branches = mock_get_existing_branches
        creator.process_assignments()
        
        # Verify results
        self.assertEqual(len(creator.created_branches), scenario['expected_new_branches'])
        self.assertEqual(len(creator.created_pull_requests), scenario['expected_new_prs'])
        
        print(f"✅ {scenario['name']}: Created {len(creator.created_branches)} branches and {len(creator.created_pull_requests)} PRs")
    
    @patch('create_assignment_prs.os.getcwd')
    @patch('create_assignment_prs.os.getcwd')
    @patch('create_assignment_prs.os.walk')
    def test_existing_prs_workflow(self, mock_walk, mock_getcwd):
        """Test workflow with existing PRs."""
        # Set up workspace
        self.temp_workspace = create_temp_workspace('simple')
        mock_getcwd.return_value = self.temp_workspace
        
        # Configure os.walk mock
        mock_walk.return_value = [
            (self.temp_workspace, ['assignments'], []),
            (os.path.join(self.temp_workspace, 'assignments'), ['assignment-1', 'assignment-2'], []),
            (os.path.join(self.temp_workspace, 'assignments', 'assignment-1'), [], ['instructions.md']),
            (os.path.join(self.temp_workspace, 'assignments', 'assignment-2'), [], ['instructions.md']),
        ]
        
        # Test scenario
        scenario = TestScenarios.existing_prs()
        
        # Add existing PR to GitHub mocker
        for branch, state in scenario['existing_prs'].items():
            self.github_mocker.add_existing_pr(branch, state)
        
        # Create and run the tool
        creator = AssignmentPRCreator()
        creator.process_assignments()
        
        # Verify results
        self.assertEqual(len(creator.created_branches), scenario['expected_new_branches'])
        self.assertEqual(len(creator.created_pull_requests), scenario['expected_new_prs'])
        
        print(f"✅ {scenario['name']}: Created {len(creator.created_branches)} branches and {len(creator.created_pull_requests)} PRs")
    
    @patch('create_assignment_prs.os.getcwd')
    @patch('create_assignment_prs.os.walk')
    def test_complex_structure_workflow(self, mock_walk, mock_getcwd):
        """Test workflow with complex assignment structure."""
        # Set up workspace
        self.temp_workspace = create_temp_workspace('complex')
        mock_getcwd.return_value = self.temp_workspace
        
        # Configure os.walk mock for complex structure
        mock_walk.return_value = [
            (self.temp_workspace, ['assignments', 'src'], []),
            (os.path.join(self.temp_workspace, 'assignments'), ['assignment-1', 'assignment-2', 'week-3', 'final'], []),
            (os.path.join(self.temp_workspace, 'assignments', 'assignment-1'), [], ['instructions.md']),
            (os.path.join(self.temp_workspace, 'assignments', 'assignment-2'), [], ['instructions.md', 'template.py']),
            (os.path.join(self.temp_workspace, 'assignments', 'week-3'), ['assignment-3'], []),
            (os.path.join(self.temp_workspace, 'assignments', 'week-3', 'assignment-3'), [], ['instructions.md']),
            (os.path.join(self.temp_workspace, 'assignments', 'final'), ['assignment-4'], []),
            (os.path.join(self.temp_workspace, 'assignments', 'final', 'assignment-4'), [], ['instructions.md']),
        ]
        
        # Create and run the tool
        creator = AssignmentPRCreator()
        creator.process_assignments()
        
        # Verify all assignments were found
        assignments = creator.find_assignments()
        expected = TestConfig.EXPECTED_ASSIGNMENTS['complex']
        self.assertEqual(sorted(assignments), sorted(expected))
        
        # Verify branches and PRs were created
        self.assertEqual(len(creator.created_branches), len(expected))
        self.assertEqual(len(creator.created_pull_requests), len(expected))
        
        print(f"✅ Complex structure: Found {len(assignments)} assignments, created {len(creator.created_branches)} branches")
    
    def test_dry_run_mode(self):
        """Test that dry-run mode doesn't call external commands."""
        with patch.dict(os.environ, {'DRY_RUN': 'true'}):
            creator = AssignmentPRCreator()
            
            # These should all return True in dry-run mode without calling external commands
            self.assertTrue(creator.run_git_command('git status', 'Test'))
            self.assertEqual(creator.run_git_command_with_output('git branch', 'Test'), '')
            self.assertTrue(creator.fetch_all_remote_branches())
            self.assertEqual(creator.get_existing_branches(), set())
            self.assertTrue(creator.create_branch('test-branch'))
            self.assertTrue(creator.create_readme('test/assignment', 'test-branch'))
            self.assertTrue(creator.push_branches_to_remote())
            
            # Verify no subprocess calls were made
            self.assertEqual(self.git_mocker.get_call_count('git status'), 0)
            
            print("✅ Dry-run mode: No external commands executed")
    
    def test_branch_name_sanitization(self):
        """Test branch name sanitization with various inputs."""
        creator = AssignmentPRCreator()
        
        test_cases = [
            ('assignments/assignment-1', 'assignments-assignment-1'),
            ('assignments/week-3/assignment-3', 'assignments-week-3-assignment-3'),
            ('Complex Path / With Spaces', 'complex-path-with-spaces'),
            ('--multiple--hyphens--', 'multiple-hyphens'),
            ('UPPERCASE/path', 'uppercase-path'),
            ('assignments/Final Project/assignment-10', 'assignments-final-project-assignment-10'),
        ]
        
        for input_path, expected in test_cases:
            with self.subTest(input_path=input_path):
                result = creator.sanitize_branch_name(input_path)
                self.assertEqual(result, expected)
        
        print("✅ Branch name sanitization: All test cases passed")
    
    def test_error_handling(self):
        """Test error handling in git operations."""
        # Test with real subprocess to trigger actual error handling
        with patch.dict(os.environ, {'DRY_RUN': 'false'}):
            creator = AssignmentPRCreator()
            
            # Mock a failing git command
            import subprocess as subprocess_module
            error = subprocess_module.CalledProcessError(1, 'git command', stderr='Test error')
            self.mock_subprocess.run.side_effect = error
            self.mock_subprocess.CalledProcessError = subprocess_module.CalledProcessError
            
            # This should raise SystemExit due to the error
            with self.assertRaises(SystemExit):
                creator.run_git_command('git status', 'Test command')
            
            print("✅ Error handling: Git command failures properly handled")


def run_integration_tests():
    """Run integration tests with detailed output."""
    print("Assignment Pull Request Creator - Integration Tests")
    print("=" * 55)
    
    # Create test suite
    loader = unittest.TestLoader()
    suite = loader.loadTestsFromTestCase(TestAssignmentPRCreatorIntegration)
    
    # Run tests with detailed output
    runner = unittest.TextTestRunner(verbosity=2, stream=sys.stdout)
    result = runner.run(suite)
    
    print("\n" + "=" * 55)
    print(f"Integration Tests Summary:")
    print(f"  Tests run: {result.testsRun}")
    print(f"  Failures: {len(result.failures)}")
    print(f"  Errors: {len(result.errors)}")
    
    if result.wasSuccessful():
        print("🎉 All integration tests passed!")
    else:
        print("❌ Some integration tests failed:")
        for test, error in result.failures + result.errors:
            print(f"  - {test}: {error.split(chr(10))[0]}")
    
    return result.wasSuccessful()


if __name__ == '__main__':
    success = run_integration_tests()
    sys.exit(0 if success else 1)
