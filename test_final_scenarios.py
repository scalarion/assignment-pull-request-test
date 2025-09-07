#!/usr/bin/env python3
"""
Comprehensive demonstration of the final branch and PR creation logic.
Shows all scenarios including the new branch changes check.
"""

import os
import tempfile
from pathlib import Path
from unittest.mock import patch, MagicMock


def test_scenario_1_new_assignment():
    """Test scenario 1: New assignment (creates branch and PR)"""
    print("🧪 Scenario 1: New assignment with changes")
    
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
            
            # Only main branch exists
            mock_main = MagicMock()
            mock_main.name = "main"
            mock_repo.get_branches.return_value = [mock_main]
            
            # No PRs exist
            mock_repo.get_pulls.return_value = []
            
            # Mock branch creation
            mock_ref = MagicMock()
            mock_ref.object.sha = "abc123"
            mock_repo.get_git_ref.return_value = mock_ref
            
            # Mock comparison showing changes
            mock_comparison = MagicMock()
            mock_comparison.ahead_by = 1
            mock_repo.compare.return_value = mock_comparison
            
            from create_assignment_prs import AssignmentPRCreator
            
            original_cwd = os.getcwd()
            try:
                os.chdir(temp_dir)
                creator = AssignmentPRCreator()
                creator.find_assignments = lambda: ["assignments/assignment-1"]
                creator.sanitize_branch_name = lambda path: "assignments-assignment-1"
                
                print("  📁 No branch exists")
                print("  📋 No PR history")
                print("  ⚡ Processing...")
                
                creator.process_assignments()
                
                # Verify branch created and PR created
                branch_created = mock_repo.create_git_ref.called
                pr_created = mock_repo.create_pull.called
                comparison_checked = mock_repo.compare.called
                
                if branch_created and pr_created and comparison_checked:
                    print("  ✅ SUCCESS: Branch created, changes checked, PR created")
                    return True
                else:
                    print(f"  ❌ FAILED: Branch={branch_created}, PR={pr_created}, Compare={comparison_checked}")
                    return False
                    
            finally:
                os.chdir(original_cwd)
                import shutil
                shutil.rmtree(temp_dir, ignore_errors=True)


def test_scenario_2_existing_branch_no_changes():
    """Test scenario 2: Existing branch with no changes (no PR created)"""
    print("\n🧪 Scenario 2: Existing branch, no PR history, no changes")
    
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
            
            # Mock comparison showing NO changes
            mock_comparison = MagicMock()
            mock_comparison.ahead_by = 0
            mock_repo.compare.return_value = mock_comparison
            
            from create_assignment_prs import AssignmentPRCreator
            
            original_cwd = os.getcwd()
            try:
                os.chdir(temp_dir)
                creator = AssignmentPRCreator()
                creator.find_assignments = lambda: ["assignments/assignment-2"]
                creator.sanitize_branch_name = lambda path: "assignments-assignment-2"
                
                print("  📁 Branch exists")
                print("  📋 No PR history")
                print("  🔍 No changes detected")
                print("  ⚡ Processing...")
                
                creator.process_assignments()
                
                # Verify no branch creation, comparison checked, no PR created
                branch_created = mock_repo.create_git_ref.called
                pr_created = mock_repo.create_pull.called
                comparison_checked = mock_repo.compare.called
                
                if not branch_created and not pr_created and comparison_checked:
                    print("  ✅ SUCCESS: No branch created, changes checked, no PR created")
                    return True
                else:
                    print(f"  ❌ FAILED: Branch={branch_created}, PR={pr_created}, Compare={comparison_checked}")
                    return False
                    
            finally:
                os.chdir(original_cwd)
                import shutil
                shutil.rmtree(temp_dir, ignore_errors=True)


def test_scenario_3_existing_pr_history():
    """Test scenario 3: Branch with closed PR (no action taken)"""
    print("\n🧪 Scenario 3: Branch exists, closed PR exists")
    
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
            
            # Closed PR exists
            mock_pr = MagicMock()
            mock_pr.head.ref = "assignments-assignment-3"
            mock_pr.state = "closed"
            mock_repo.get_pulls.return_value = [mock_pr]
            
            from create_assignment_prs import AssignmentPRCreator
            
            original_cwd = os.getcwd()
            try:
                os.chdir(temp_dir)
                creator = AssignmentPRCreator()
                creator.find_assignments = lambda: ["assignments/assignment-3"]
                creator.sanitize_branch_name = lambda path: "assignments-assignment-3"
                
                print("  📁 Branch exists")
                print("  📋 Closed PR exists")
                print("  ⚡ Processing...")
                
                creator.process_assignments()
                
                # Verify no branch creation, no comparison (skipped), no PR created
                branch_created = mock_repo.create_git_ref.called
                pr_created = mock_repo.create_pull.called
                comparison_checked = mock_repo.compare.called
                
                if not branch_created and not pr_created and not comparison_checked:
                    print("  ✅ SUCCESS: No action taken (PR history exists)")
                    return True
                else:
                    print(f"  ❌ FAILED: Branch={branch_created}, PR={pr_created}, Compare={comparison_checked}")
                    return False
                    
            finally:
                os.chdir(original_cwd)
                import shutil
                shutil.rmtree(temp_dir, ignore_errors=True)


if __name__ == "__main__":
    print("🚀 Testing Final Branch and PR Creation Logic")
    print("=" * 70)
    print("Includes branch changes validation to prevent empty PRs")
    print("=" * 70)
    
    success1 = test_scenario_1_new_assignment()
    success2 = test_scenario_2_existing_branch_no_changes() 
    success3 = test_scenario_3_existing_pr_history()
    
    print("\n" + "=" * 70)
    if success1 and success2 and success3:
        print("🎊 ALL SCENARIOS PASSED! Final logic is working correctly.")
        print("\n📋 Summary of behaviors:")
        print("  ✅ New assignment → Creates branch + PR (if changes exist)")
        print("  ❌ Existing branch, no changes → No PR created")
        print("  ❌ Any PR history → No action taken")
        print("  ✅ Branch changes validated before PR creation")
    else:
        print("💥 SOME SCENARIOS FAILED! Check the logic above.")
        exit(1)
