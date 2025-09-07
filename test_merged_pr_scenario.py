#!/usr/bin/env python3
"""
Test script to demonstrate the new branch recreation logic.
This simulates the scenario where a PR was merged and branch was deleted.
"""

import os
import tempfile
from pathlib import Path
from unittest.mock import patch, MagicMock


def test_merged_pr_scenario():
    """
    Demonstrate that branches are not recreated after PR merge.
    """
    print("🧪 Testing merged PR scenario...")
    
    # Create temporary assignment structure
    temp_dir = tempfile.mkdtemp()
    assignments_dir = Path(temp_dir) / "assignments"
    assignments_dir.mkdir()
    (assignments_dir / "assignment-1").mkdir()
    (assignments_dir / "assignment-1" / "instructions.md").write_text("Test assignment")
    
    # Mock environment
    env_vars = {
        'GITHUB_TOKEN': 'test_token',
        'GITHUB_REPOSITORY': 'test/repo',
        'ASSIGNMENTS_ROOT_REGEX': '^assignments$',
        'ASSIGNMENT_REGEX': '^assignment-\\d+$',
        'DEFAULT_BRANCH': 'main',
        'DRY_RUN': 'false'
    }
    
    with patch.dict(os.environ, env_vars):
        with patch('create_assignment_prs.Github') as mock_github_class:
            # Setup mocks
            mock_github = MagicMock()
            mock_repo = MagicMock()
            mock_github_class.return_value = mock_github
            mock_github.get_repo.return_value = mock_repo
            
            # Scenario: Only main branch exists (assignment branch was deleted)
            mock_main_branch = MagicMock()
            mock_main_branch.name = "main"
            mock_repo.get_branches.return_value = [mock_main_branch]
            
            # Scenario: Closed/merged PR exists for the assignment
            mock_pr = MagicMock()
            mock_pr.head.ref = "assignments-assignment-1"
            mock_pr.state = "closed"
            mock_repo.get_pulls.return_value = [mock_pr]
            
            # Import and test
            from create_assignment_prs import AssignmentPRCreator
            
            original_cwd = os.getcwd()
            try:
                os.chdir(temp_dir)
                creator = AssignmentPRCreator()
                
                # Override methods for testing
                creator.find_assignments = lambda: ["assignments/assignment-1"]
                creator.sanitize_branch_name = lambda path: "assignments-assignment-1"
                
                print(f"📁 Created test structure at: {temp_dir}")
                print(f"🏗️ Assignment found: assignments/assignment-1")
                print(f"🌿 Expected branch name: assignments-assignment-1")
                print(f"📋 Simulated closed PR exists for this branch")
                print()
                
                # Process assignments
                print("⚡ Processing assignments...")
                creator.process_assignments()
                print()
                
                # Verify results
                if mock_repo.create_git_ref.called:
                    print("❌ FAILED: Branch was recreated (should not happen)")
                    return False
                else:
                    print("✅ SUCCESS: Branch was NOT recreated (correct behavior)")
                
                if mock_repo.create_pull.called:
                    print("❌ FAILED: PR was created (should not happen)")
                    return False
                else:
                    print("✅ SUCCESS: PR was NOT created (correct behavior)")
                
                print()
                print("🎉 Test passed! Assignment with merged PR is correctly ignored.")
                return True
                
            finally:
                os.chdir(original_cwd)
                import shutil
                shutil.rmtree(temp_dir, ignore_errors=True)


def test_new_assignment_scenario():
    """
    Demonstrate that branches ARE created for new assignments.
    """
    print("\n🧪 Testing new assignment scenario...")
    
    # Create temporary assignment structure
    temp_dir = tempfile.mkdtemp()
    assignments_dir = Path(temp_dir) / "assignments"
    assignments_dir.mkdir()
    (assignments_dir / "assignment-2").mkdir()
    (assignments_dir / "assignment-2" / "instructions.md").write_text("New assignment")
    
    # Mock environment
    env_vars = {
        'GITHUB_TOKEN': 'test_token',
        'GITHUB_REPOSITORY': 'test/repo',
        'ASSIGNMENTS_ROOT_REGEX': '^assignments$',
        'ASSIGNMENT_REGEX': '^assignment-\\d+$',
        'DEFAULT_BRANCH': 'main',
        'DRY_RUN': 'false'
    }
    
    with patch.dict(os.environ, env_vars):
        with patch('create_assignment_prs.Github') as mock_github_class:
            # Setup mocks
            mock_github = MagicMock()
            mock_repo = MagicMock()
            mock_github_class.return_value = mock_github
            mock_github.get_repo.return_value = mock_repo
            
            # Scenario: Only main branch exists
            mock_main_branch = MagicMock()
            mock_main_branch.name = "main"
            mock_repo.get_branches.return_value = [mock_main_branch]
            
            # Scenario: No PRs exist
            mock_repo.get_pulls.return_value = []
            
            # Mock successful branch creation
            mock_ref = MagicMock()
            mock_ref.object.sha = "abc123"
            mock_repo.get_git_ref.return_value = mock_ref
            
            # Import and test
            from create_assignment_prs import AssignmentPRCreator
            
            original_cwd = os.getcwd()
            try:
                os.chdir(temp_dir)
                creator = AssignmentPRCreator()
                
                # Override methods for testing
                creator.find_assignments = lambda: ["assignments/assignment-2"]
                creator.sanitize_branch_name = lambda path: "assignments-assignment-2"
                
                print(f"📁 Created test structure at: {temp_dir}")
                print(f"🏗️ Assignment found: assignments/assignment-2")
                print(f"🌿 Expected branch name: assignments-assignment-2")
                print(f"📋 No PRs exist for this assignment")
                print()
                
                # Process assignments
                print("⚡ Processing assignments...")
                creator.process_assignments()
                print()
                
                # Verify results
                if mock_repo.create_git_ref.called:
                    print("✅ SUCCESS: Branch was created (correct behavior)")
                    call_args = mock_repo.create_git_ref.call_args
                    expected_ref = 'refs/heads/assignments-assignment-2'
                    if call_args[1]['ref'] == expected_ref:
                        print(f"✅ SUCCESS: Correct branch ref created: {expected_ref}")
                    else:
                        print(f"❌ FAILED: Wrong branch ref: {call_args[1]['ref']}")
                        return False
                else:
                    print("❌ FAILED: Branch was NOT created (should have been created)")
                    return False
                
                print()
                print("🎉 Test passed! New assignment correctly gets branch and PR.")
                return True
                
            finally:
                os.chdir(original_cwd)
                import shutil
                shutil.rmtree(temp_dir, ignore_errors=True)


if __name__ == "__main__":
    print("🚀 Testing Assignment Pull Request Creator - Branch Recreation Logic")
    print("=" * 70)
    
    success1 = test_merged_pr_scenario()
    success2 = test_new_assignment_scenario()
    
    print("\n" + "=" * 70)
    if success1 and success2:
        print("🎊 ALL TESTS PASSED! Branch recreation logic is working correctly.")
    else:
        print("💥 SOME TESTS FAILED! Check the logic above.")
        exit(1)
