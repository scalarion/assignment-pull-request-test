#!/usr/bin/env python3
"""
Test configuration and utilities for Assignment Pull Request Creator tests.

This module provides test configurations, fixtures, and utilities for
comprehensive testing with proper mocking.
"""

import os
import tempfile
from typing import Dict, List, Any
from unittest.mock import Mock


class TestConfig:
    """Configuration class for test scenarios."""
    
    # Default environment variables for testing
    DEFAULT_ENV = {
        'GITHUB_TOKEN': 'test_token_12345',
        'GITHUB_REPOSITORY': 'test-org/test-repo',
        'DEFAULT_BRANCH': 'main',
        'ASSIGNMENTS_ROOT_REGEX': '^assignments$',
        'ASSIGNMENT_REGEX': r'^assignment-\d+$',
        'DRY_RUN': 'false'
    }
    
    # Test assignment directory structures
    ASSIGNMENT_STRUCTURES = {
        'simple': [
            ('/workspace', ['assignments'], []),
            ('/workspace/assignments', ['assignment-1', 'assignment-2'], []),
            ('/workspace/assignments/assignment-1', [], ['instructions.md']),
            ('/workspace/assignments/assignment-2', [], ['instructions.md']),
        ],
        'nested': [
            ('/workspace', ['assignments'], []),
            ('/workspace/assignments', ['assignment-1', 'week-3'], []),
            ('/workspace/assignments/assignment-1', [], ['instructions.md']),
            ('/workspace/assignments/week-3', ['assignment-3'], []),
            ('/workspace/assignments/week-3/assignment-3', [], ['instructions.md']),
        ],
        'complex': [
            ('/workspace', ['assignments', 'src'], []),
            ('/workspace/assignments', ['assignment-1', 'assignment-2', 'week-3', 'final'], []),
            ('/workspace/assignments/assignment-1', [], ['instructions.md']),
            ('/workspace/assignments/assignment-2', [], ['instructions.md', 'template.py']),
            ('/workspace/assignments/week-3', ['assignment-3'], []),
            ('/workspace/assignments/week-3/assignment-3', [], ['instructions.md']),
            ('/workspace/assignments/final', ['assignment-4'], []),
            ('/workspace/assignments/final/assignment-4', [], ['instructions.md']),
        ]
    }
    
    # Expected results for each structure
    EXPECTED_ASSIGNMENTS = {
        'simple': [
            'assignments/assignment-1',
            'assignments/assignment-2'
        ],
        'nested': [
            'assignments/assignment-1',
            'assignments/week-3/assignment-3'
        ],
        'complex': [
            'assignments/assignment-1',
            'assignments/assignment-2',
            'assignments/week-3/assignment-3',
            'assignments/final/assignment-4'
        ]
    }


class GitCommandMocker:
    """Mock git commands with realistic responses."""
    
    def __init__(self):
        self.command_responses = {
            'git fetch --all': {'stdout': 'Fetching origin\n', 'stderr': ''},
            'git branch -r': {'stdout': '  origin/main\n  origin/feature-1\n', 'stderr': ''},
            'git branch': {'stdout': '* main\n  feature-1\n', 'stderr': ''},
            'git checkout main': {'stdout': 'Switched to branch \'main\'\n', 'stderr': ''},
            'git status': {'stdout': 'On branch main\nnothing to commit\n', 'stderr': ''},
            'git push --all origin': {'stdout': 'Everything up-to-date\n', 'stderr': ''},
        }
        self.command_call_count = {}
    
    def mock_subprocess_run(self, command, **kwargs):
        """Mock subprocess.run for git commands."""
        # Track command calls
        self.command_call_count[command] = self.command_call_count.get(command, 0) + 1
        
        # Return mock result
        result = Mock()
        
        # Handle specific commands
        if command in self.command_responses:
            response = self.command_responses[command]
            result.stdout = response['stdout']
            result.stderr = response['stderr']
            result.returncode = 0
        elif 'git checkout -b' in command:
            branch_name = command.split()[-1]
            result.stdout = f'Switched to a new branch \'{branch_name}\'\n'
            result.stderr = ''
            result.returncode = 0
        elif 'git add' in command:
            result.stdout = ''
            result.stderr = ''
            result.returncode = 0
        elif 'git commit' in command:
            result.stdout = '[main 1234567] Add README\n 1 file changed, 10 insertions(+)\n'
            result.stderr = ''
            result.returncode = 0
        else:
            # Default successful response
            result.stdout = ''
            result.stderr = ''
            result.returncode = 0
        
        return result
    
    def get_call_count(self, command: str) -> int:
        """Get the number of times a command was called."""
        return self.command_call_count.get(command, 0)
    
    def reset_counts(self):
        """Reset command call counts."""
        self.command_call_count.clear()


class GitHubAPIMocker:
    """Mock GitHub API responses."""
    
    def __init__(self):
        self.mock_prs = []
        self.mock_repo = Mock()
        self.mock_github = Mock()
        
        # Set up the mock chain
        self.mock_github.get_repo.return_value = self.mock_repo
        self.mock_repo.get_pulls.return_value = self.mock_prs
        self.mock_repo.create_pull.side_effect = self._create_pr_mock
        
        self.created_prs = []
    
    def add_existing_pr(self, branch_name: str, state: str = 'open'):
        """Add an existing PR to the mock."""
        mock_pr = Mock()
        mock_pr.head.ref = branch_name
        mock_pr.state = state
        mock_pr.number = len(self.mock_prs) + 1
        self.mock_prs.append(mock_pr)
    
    def _create_pr_mock(self, title, body, head, base):
        """Mock PR creation."""
        mock_pr = Mock()
        mock_pr.number = len(self.created_prs) + 100
        mock_pr.html_url = f'https://github.com/test/repo/pull/{mock_pr.number}'
        mock_pr.title = title
        mock_pr.body = body
        mock_pr.head.ref = head
        mock_pr.base.ref = base
        
        self.created_prs.append(mock_pr)
        return mock_pr
    
    def get_github_mock(self):
        """Get the GitHub mock object."""
        return self.mock_github
    
    def get_repo_mock(self):
        """Get the repository mock object."""
        return self.mock_repo


class TestScenarios:
    """Pre-defined test scenarios."""
    
    @staticmethod
    def clean_repository():
        """Scenario: Clean repository with no existing branches or PRs."""
        return {
            'name': 'clean_repository',
            'description': 'Clean repository with no existing branches or PRs',
            'existing_branches': set(),
            'existing_prs': {},
            'expected_new_branches': 2,  # Based on simple structure
            'expected_new_prs': 2
        }
    
    @staticmethod
    def existing_branches():
        """Scenario: Repository with some existing assignment branches."""
        return {
            'name': 'existing_branches',
            'description': 'Repository with existing assignment branches',
            'existing_branches': {'main', 'assignments-assignment-1'},
            'existing_prs': {},
            'expected_new_branches': 1,  # Only assignment-2 should be created
            'expected_new_prs': 2       # PRs for both assignment-1 (existing branch) and assignment-2 (new branch)
        }
    
    @staticmethod
    def existing_prs():
        """Scenario: Repository with existing PRs (some closed)."""
        return {
            'name': 'existing_prs',
            'description': 'Repository with existing PRs',
            'existing_branches': set(),
            'existing_prs': {'assignments-assignment-1': 'closed'},
            'expected_new_branches': 1,  # Only assignment-2 should be created
            'expected_new_prs': 1
        }
    
    @staticmethod
    def mixed_existing():
        """Scenario: Repository with both existing branches and PRs."""
        return {
            'name': 'mixed_existing',
            'description': 'Repository with mixed existing branches and PRs',
            'existing_branches': {'main', 'assignments-assignment-1'},
            'existing_prs': {'assignments-assignment-2': 'open'},
            'expected_new_branches': 0,  # No new branches should be created
            'expected_new_prs': 0
        }


def create_temp_workspace(structure_name: str = 'simple') -> str:
    """Create a temporary workspace with the specified structure."""
    temp_dir = tempfile.mkdtemp()
    
    # Create the directory structure
    structure = TestConfig.ASSIGNMENT_STRUCTURES[structure_name]
    
    for path, dirs, files in structure:
        # Adjust path to use temp_dir
        actual_path = path.replace('/workspace', temp_dir)
        
        # Create directories
        os.makedirs(actual_path, exist_ok=True)
        for dir_name in dirs:
            os.makedirs(os.path.join(actual_path, dir_name), exist_ok=True)
        
        # Create files
        for file_name in files:
            file_path = os.path.join(actual_path, file_name)
            with open(file_path, 'w') as f:
                f.write(f'# {file_name}\nTest content for {file_name}')
    
    return temp_dir


if __name__ == '__main__':
    # Test the configuration
    print("Testing configuration...")
    
    # Test temp workspace creation
    workspace = create_temp_workspace('simple')
    print(f"Created temp workspace: {workspace}")
    
    # Test git mocker
    git_mocker = GitCommandMocker()
    result = git_mocker.mock_subprocess_run('git status')
    print(f"Git mock result: {result.stdout}")
    
    # Test GitHub mocker
    github_mocker = GitHubAPIMocker()
    github_mocker.add_existing_pr('test-branch', 'open')
    repo = github_mocker.get_repo_mock()
    prs = repo.get_pulls()
    print(f"GitHub mock PRs: {[pr.head.ref for pr in prs]}")
    
    print("✅ Configuration tests passed!")
    
    # Clean up
    import shutil
    shutil.rmtree(workspace)
