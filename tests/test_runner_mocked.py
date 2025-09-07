#!/usr/bin/env python3
"""
Test runner for mocked tests of Assignment Pull Request Creator.

This module provides a simple way to run the mocked tests and verify
that the git command mocking is working correctly.
"""

import os
import sys
import unittest
from unittest.mock import patch, Mock

# Add the parent directory to the path
sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from tests.test_mocked import TestAssignmentPRCreatorMocked, TestGitCommandMocking


def run_specific_tests():
    """Run specific tests to verify mocking functionality."""
    
    print("Testing Assignment Pull Request Creator with Mocking")
    print("=" * 60)
    
    # Create test suite
    loader = unittest.TestLoader()
    suite = unittest.TestSuite()
    
    # Add specific test methods
    test_methods = [
        'test_initialization',
        'test_sanitize_branch_name',
        'test_run_git_command_dry_run',
        'test_run_git_command_with_output_dry_run',
        'test_fetch_all_remote_branches_dry_run',
        'test_get_existing_branches_dry_run',
        'test_create_branch_dry_run',
        'test_push_branches_to_remote_dry_run',
    ]
    
    for method in test_methods:
        suite.addTest(TestAssignmentPRCreatorMocked(method))
    
    # Add git command mocking tests
    git_test_methods = [
        'test_git_command_failure_handling',
        'test_complex_git_workflow_mocking',
    ]
    
    for method in git_test_methods:
        suite.addTest(TestGitCommandMocking(method))
    
    # Run tests
    runner = unittest.TextTestRunner(verbosity=2)
    result = runner.run(suite)
    
    print(f"\nRan {result.testsRun} tests")
    if result.wasSuccessful():
        print("✅ All tests passed!")
    else:
        print(f"❌ {len(result.failures)} failures, {len(result.errors)} errors")
        for test, error in result.failures + result.errors:
            print(f"  - {test}: {error.split(chr(10))[0]}")
    
    return result.wasSuccessful()


def test_git_mocking_specifically():
    """Test git command mocking in isolation."""
    
    print("\nTesting Git Command Mocking in Isolation")
    print("=" * 45)
    
    # Import the class directly
    from create_assignment_prs import AssignmentPRCreator
    
    # Test dry-run mode (no actual git commands)
    with patch.dict(os.environ, {
        'GITHUB_TOKEN': 'test_token',
        'GITHUB_REPOSITORY': 'test/repo',
        'DRY_RUN': 'true'
    }):
        with patch('create_assignment_prs.Github'):
            creator = AssignmentPRCreator()
            
            print("✓ Dry-run initialization successful")
            
            # Test git commands in dry-run mode
            result = creator.run_git_command('git status', 'Check status')
            print(f"✓ Git command dry-run: {result}")
            
            output = creator.run_git_command_with_output('git branch', 'List branches')
            print(f"✓ Git command with output dry-run: '{output}'")
            
            # Test branch creation
            result = creator.create_branch('test-branch')
            print(f"✓ Branch creation dry-run: {result}")
            
            # Test README creation
            result = creator.create_readme('assignments/test', 'test-branch')
            print(f"✓ README creation dry-run: {result}")
    
    # Test with subprocess mocking
    with patch.dict(os.environ, {
        'GITHUB_TOKEN': 'test_token',
        'GITHUB_REPOSITORY': 'test/repo',
        'DRY_RUN': 'false'
    }):
        with patch('create_assignment_prs.Github'):
            with patch('create_assignment_prs.subprocess') as mock_subprocess:
                # Mock successful git command
                mock_result = Mock()
                mock_result.stdout = 'mocked output'
                mock_result.stderr = ''
                mock_subprocess.run.return_value = mock_result
                
                creator = AssignmentPRCreator()
                
                result = creator.run_git_command('git status', 'Check status')
                print(f"✓ Mocked git command: {result}")
                
                output = creator.run_git_command_with_output('git branch', 'List branches')
                print(f"✓ Mocked git output: '{output}'")
                
                # Verify subprocess was called
                assert mock_subprocess.run.called
                print("✓ Subprocess mocking verified")
    
    print("✅ Git mocking tests completed successfully!")


if __name__ == '__main__':
    print("Assignment Pull Request Creator - Test Suite")
    print("=" * 50)
    
    # Test git mocking specifically first
    test_git_mocking_specifically()
    
    print("\n" + "=" * 50)
    
    # Run the full test suite
    success = run_specific_tests()
    
    sys.exit(0 if success else 1)
