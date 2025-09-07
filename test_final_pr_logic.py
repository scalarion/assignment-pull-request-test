#!/usr/bin/env python3
"""
Comprehensive test to demonstrate the final PR creation logic.
Tests that PRs are only created when NO PR has ever existed for a branch.
"""

import os
import tempfile
from pathlib import Path
from unittest.mock import patch, MagicMock


def test_no_pr_creation_for_closed_pr():
    """Test that no PR is created when a closed PR exists for a branch."""
    print("🧪 Testing: Branch exists + Closed PR exists = No new PR created")
    
    temp_dir = tempfile.mkdtemp()
    assignments_dir = Path(temp_dir) / "assignments"
    assignments_dir.mkdir()
    (assignments_dir / "assignment-1").mkdir()
    (assignments_dir / "assignment-1" / "instructions.md").write_text("Test")
    
    env_vars = {
        'GITHUB_TOKEN': 'test_token',
        'GITHUB_REPOSITORY': 'test/repo',
        'DRY_RUN': 'false'
    }
    
    with patch.dict(os.environ, env_vars):
        with patch('create_assignment_prs.Github') as mock_github_class:
            mock_github = MagicMock()
            mock_repo = MagicMock()
            mock_github_class.return_value = mock_github
            mock_github.get_repo.return_value = mock_repo
            
            # Branch exists
            mock_main = MagicMock()
            mock_main.name = "main"
            mock_branch = MagicMock()
            mock_branch.name = "assignments-assignment-1"
            mock_repo.get_branches.return_value = [mock_main, mock_branch]
            
            # Closed PR exists
            mock_pr = MagicMock()
            mock_pr.head.ref = "assignments-assignment-1"
            mock_pr.state = "closed"
            mock_repo.get_pulls.return_value = [mock_pr]
            
            from create_assignment_prs import AssignmentPRCreator
            
            original_cwd = os.getcwd()
            try:
                os.chdir(temp_dir)
                creator = AssignmentPRCreator()
                creator.find_assignments = lambda: ["assignments/assignment-1"]
                creator.sanitize_branch_name = lambda path: "assignments-assignment-1"
                
                print("  📁 Branch exists: assignments-assignment-1")
                print("  📋 Closed PR exists for this branch")
                print("  ⚡ Processing...")
                
                creator.process_assignments()
                
                # Verify no PR created
                if not mock_repo.create_pull.called:
                    print("  ✅ SUCCESS: No new PR created (correct behavior)")
                    return True
                else:
                    print("  ❌ FAILED: PR was created (should not happen)")
                    return False
                    
            finally:
                os.chdir(original_cwd)
                import shutil
                shutil.rmtree(temp_dir, ignore_errors=True)


def test_pr_creation_for_branch_without_pr_history():
    """Test that PR IS created when branch exists but no PR has ever existed."""
    print("\n🧪 Testing: Branch exists + No PR history = New PR created")
    
    temp_dir = tempfile.mkdtemp()
    assignments_dir = Path(temp_dir) / "assignments"
    assignments_dir.mkdir()
    (assignments_dir / "assignment-2").mkdir()
    (assignments_dir / "assignment-2" / "instructions.md").write_text("Test")
    
    env_vars = {
        'GITHUB_TOKEN': 'test_token',
        'GITHUB_REPOSITORY': 'test/repo',
        'DRY_RUN': 'false'
    }
    
    with patch.dict(os.environ, env_vars):
        with patch('create_assignment_prs.Github') as mock_github_class:
            mock_github = MagicMock()
            mock_repo = MagicMock()
            mock_github_class.return_value = mock_github
            mock_github.get_repo.return_value = mock_repo
            
            # Branch exists
            mock_main = MagicMock()
            mock_main.name = "main"
            mock_branch = MagicMock()
            mock_branch.name = "assignments-assignment-2"
            mock_repo.get_branches.return_value = [mock_main, mock_branch]
            
            # No PRs exist
            mock_repo.get_pulls.return_value = []
            
            from create_assignment_prs import AssignmentPRCreator
            
            original_cwd = os.getcwd()
            try:
                os.chdir(temp_dir)
                creator = AssignmentPRCreator()
                creator.find_assignments = lambda: ["assignments/assignment-2"]
                creator.sanitize_branch_name = lambda path: "assignments-assignment-2"
                
                print("  📁 Branch exists: assignments-assignment-2")
                print("  📋 No PR history for this branch")
                print("  ⚡ Processing...")
                
                creator.process_assignments()
                
                # Verify PR created
                if mock_repo.create_pull.called:
                    print("  ✅ SUCCESS: New PR created (correct behavior)")
                    return True
                else:
                    print("  ❌ FAILED: No PR created (should have been created)")
                    return False
                    
            finally:
                os.chdir(original_cwd)
                import shutil
                shutil.rmtree(temp_dir, ignore_errors=True)


def test_no_pr_creation_for_merged_pr():
    """Test that no PR is created when a merged PR exists."""
    print("\n🧪 Testing: Branch exists + Merged PR exists = No new PR created")
    
    temp_dir = tempfile.mkdtemp()
    assignments_dir = Path(temp_dir) / "assignments"
    assignments_dir.mkdir()
    (assignments_dir / "assignment-3").mkdir()
    (assignments_dir / "assignment-3" / "instructions.md").write_text("Test")
    
    env_vars = {
        'GITHUB_TOKEN': 'test_token',
        'GITHUB_REPOSITORY': 'test/repo',
        'DRY_RUN': 'false'
    }
    
    with patch.dict(os.environ, env_vars):
        with patch('create_assignment_prs.Github') as mock_github_class:
            mock_github = MagicMock()
            mock_repo = MagicMock()
            mock_github_class.return_value = mock_github
            mock_github.get_repo.return_value = mock_repo
            
            # Branch exists
            mock_main = MagicMock()
            mock_main.name = "main"
            mock_branch = MagicMock()
            mock_branch.name = "assignments-assignment-3"
            mock_repo.get_branches.return_value = [mock_main, mock_branch]
            
            # Merged PR exists
            mock_pr = MagicMock()
            mock_pr.head.ref = "assignments-assignment-3"
            mock_pr.state = "merged"
            mock_repo.get_pulls.return_value = [mock_pr]
            
            from create_assignment_prs import AssignmentPRCreator
            
            original_cwd = os.getcwd()
            try:
                os.chdir(temp_dir)
                creator = AssignmentPRCreator()
                creator.find_assignments = lambda: ["assignments/assignment-3"]
                creator.sanitize_branch_name = lambda path: "assignments-assignment-3"
                
                print("  📁 Branch exists: assignments-assignment-3")
                print("  📋 Merged PR exists for this branch")
                print("  ⚡ Processing...")
                
                creator.process_assignments()
                
                # Verify no PR created
                if not mock_repo.create_pull.called:
                    print("  ✅ SUCCESS: No new PR created (correct behavior)")
                    return True
                else:
                    print("  ❌ FAILED: PR was created (should not happen)")
                    return False
                    
            finally:
                os.chdir(original_cwd)
                import shutil
                shutil.rmtree(temp_dir, ignore_errors=True)


if __name__ == "__main__":
    print("🚀 Testing Final PR Creation Logic")
    print("=" * 60)
    print("Rule: Create PR ONLY when NO PR has ever existed for the branch")
    print("=" * 60)
    
    success1 = test_no_pr_creation_for_closed_pr()
    success2 = test_pr_creation_for_branch_without_pr_history()
    success3 = test_no_pr_creation_for_merged_pr()
    
    print("\n" + "=" * 60)
    if success1 and success2 and success3:
        print("🎊 ALL TESTS PASSED! Final PR logic is working correctly.")
        print("\n📋 Summary of behavior:")
        print("  • No PR created if closed PR exists")
        print("  • No PR created if merged PR exists") 
        print("  • PR created only if no PR has ever existed")
    else:
        print("💥 SOME TESTS FAILED! Check the logic above.")
        exit(1)
