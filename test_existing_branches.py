#!/usr/bin/env python3

import os
import subprocess
import tempfile
import shutil
import sys

def run_command(cmd, cwd=None):
    """Run a command and return its output."""
    result = subprocess.run(cmd, shell=True, capture_output=True, text=True, cwd=cwd)
    if result.returncode != 0:
        print(f"Command failed: {cmd}")
        print(f"Error: {result.stderr}")
        return None
    return result.stdout.strip()

def test_existing_branches_logic():
    """Test the fetch and local branch handling logic."""
    
    # Create a temporary git repository
    with tempfile.TemporaryDirectory() as temp_dir:
        print(f"Creating test repository in: {temp_dir}")
        
        # Initialize git repo
        run_command("git init", cwd=temp_dir)
        run_command("git config user.name 'Test User'", cwd=temp_dir)
        run_command("git config user.email 'test@example.com'", cwd=temp_dir)
        
        # Create initial commit
        run_command("echo '# Test Repository' > README.md", cwd=temp_dir)
        run_command("git add README.md", cwd=temp_dir)
        run_command("git commit -m 'Initial commit'", cwd=temp_dir)
        
        # Create some branches to simulate existing assignment branches
        for i in range(1, 4):
            branch_name = f"assignment-{i}"
            run_command(f"git checkout -b {branch_name}", cwd=temp_dir)
            run_command(f"echo 'Assignment {i}' > assignment-{i}.md", cwd=temp_dir)
            run_command(f"git add assignment-{i}.md", cwd=temp_dir)
            run_command(f"git commit -m 'Add assignment {i}'", cwd=temp_dir)
        
        # Go back to main
        run_command("git checkout main", cwd=temp_dir)
        
        # Show branches
        print("\nLocal branches:")
        branches = run_command("git branch", cwd=temp_dir)
        print(branches)
        
        # Test fetching all branches (simulating what our script would do)
        print("\nTesting branch detection logic:")
        
        # Get local branches like our script does
        output = run_command("git branch", cwd=temp_dir)
        local_branches = set()
        
        for line in output.split('\n'):
            if line.strip():
                # Format: "* main" or "  branch-name"
                branch_name = line.strip().replace('* ', '').strip()
                if branch_name:
                    local_branches.add(branch_name)
        
        print(f"Detected local branches: {local_branches}")
        
        # Test what happens when we create a new assignment dir
        assignments_dir = os.path.join(temp_dir, "assignments")
        os.makedirs(assignments_dir, exist_ok=True)
        
        # Create new assignment directories
        for i in range(1, 6):
            assignment_dir = os.path.join(assignments_dir, f"assignment-{i}")
            os.makedirs(assignment_dir, exist_ok=True)
            with open(os.path.join(assignment_dir, "instructions.md"), "w") as f:
                f.write(f"# Assignment {i}\nInstructions for assignment {i}")
        
        print(f"\nCreated assignment directories: assignment-1 through assignment-5")
        print(f"Existing branches: {sorted(local_branches)}")
        
        # Simulate our script's logic
        for i in range(1, 6):
            assignment_path = f"assignments/assignment-{i}"
            branch_name = f"assignment-{i}"  # Simplified for this test
            
            if branch_name in local_branches:
                print(f"✓ {assignment_path} -> Branch '{branch_name}' already exists locally, would skip")
            else:
                print(f"• {assignment_path} -> Branch '{branch_name}' does not exist, would create")

if __name__ == "__main__":
    test_existing_branches_logic()
