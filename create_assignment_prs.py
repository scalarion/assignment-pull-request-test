#!/usr/bin/env python3
"""
Assignment Pull Request Creator

This script scans for assignment folders and creates pull requests with README files.
"""

import os
import re
import sys
import json
from pathlib import Path
from typing import List, Tuple, Dict, Set
from github import Github
from github.GithubException import GithubException


class AssignmentPRCreator:
    def __init__(self):
        """Initialize the Assignment PR Creator with environment variables."""
        self.github_token = os.environ.get('GITHUB_TOKEN')
        self.assignments_folder = os.environ.get('ASSIGNMENTS_FOLDER', 'assignments')
        self.assignment_regex = os.environ.get('ASSIGNMENT_REGEX', r'^assignment-\d+$')
        self.repository_name = os.environ.get('GITHUB_REPOSITORY')
        self.default_branch = os.environ.get('GITHUB_REF_NAME', 'main')
        
        if not self.github_token:
            raise ValueError("GITHUB_TOKEN environment variable is required")
        if not self.repository_name:
            raise ValueError("GITHUB_REPOSITORY environment variable is required")
        
        self.github = Github(self.github_token)
        self.repo = self.github.get_repo(self.repository_name)
        
        # Compile regex pattern
        self.pattern = re.compile(self.assignment_regex)
        
        # Track created items
        self.created_branches: List[str] = []
        self.created_pull_requests: List[str] = []

    def sanitize_branch_name(self, assignment_path: str) -> str:
        """
        Sanitize assignment path to create a valid branch name.
        
        Args:
            assignment_path: Relative path of assignment from assignments folder
            
        Returns:
            Sanitized branch name
        """
        # Remove leading/trailing whitespace
        branch_name = assignment_path.strip()
        
        # Replace spaces with hyphens
        branch_name = re.sub(r'\s+', '-', branch_name)
        
        # Remove slashes
        branch_name = branch_name.replace('/', '-')
        
        # Remove consecutive hyphens
        branch_name = re.sub(r'-+', '-', branch_name)
        
        # Convert to lowercase
        branch_name = branch_name.lower()
        
        # Remove leading/trailing hyphens
        branch_name = branch_name.strip('-')
        
        return branch_name

    def find_assignments(self) -> List[str]:
        """
        Find all assignment folders that match the regex pattern.
        
        Returns:
            List of relative paths to assignment folders
        """
        assignments = []
        assignments_root = Path(self.assignments_folder)
        
        if not assignments_root.exists():
            print(f"Assignments folder '{self.assignments_folder}' does not exist")
            return assignments
        
        print(f"Scanning for assignments in '{self.assignments_folder}' with pattern '{self.assignment_regex}'")
        
        # Walk through all subdirectories
        for root, dirs, files in os.walk(assignments_root):
            for dir_name in dirs:
                if self.pattern.match(dir_name):
                    # Get relative path from assignments folder
                    full_path = Path(root) / dir_name
                    relative_path = full_path.relative_to(assignments_root)
                    assignments.append(str(relative_path))
                    print(f"Found assignment: {relative_path}")
        
        return assignments

    def get_existing_branches(self) -> Set[str]:
        """
        Get all existing branches in the repository.
        
        Returns:
            Set of branch names
        """
        try:
            branches = self.repo.get_branches()
            return {branch.name for branch in branches}
        except GithubException as e:
            print(f"Error getting branches: {e}")
            return set()

    def get_existing_pull_requests(self) -> Set[str]:
        """
        Get all existing pull request head branch names.
        
        Returns:
            Set of branch names that have pull requests
        """
        try:
            pulls = self.repo.get_pulls(state='all')
            return {pr.head.ref for pr in pulls}
        except GithubException as e:
            print(f"Error getting pull requests: {e}")
            return set()

    def create_branch(self, branch_name: str) -> bool:
        """
        Create a new branch from the default branch.
        
        Args:
            branch_name: Name of the branch to create
            
        Returns:
            True if branch was created, False otherwise
        """
        try:
            # Get the default branch reference
            default_ref = self.repo.get_git_ref(f"heads/{self.default_branch}")
            
            # Create new branch
            self.repo.create_git_ref(
                ref=f"refs/heads/{branch_name}",
                sha=default_ref.object.sha
            )
            
            print(f"Created branch: {branch_name}")
            self.created_branches.append(branch_name)
            return True
            
        except GithubException as e:
            print(f"Error creating branch '{branch_name}': {e}")
            return False

    def create_readme(self, assignment_path: str, branch_name: str) -> bool:
        """
        Create README.md file in the assignment folder.
        
        Args:
            assignment_path: Relative path to assignment folder
            branch_name: Branch name to commit to
            
        Returns:
            True if README was created, False otherwise
        """
        try:
            readme_path = f"{self.assignments_folder}/{assignment_path}/README.md"
            
            # Create README content
            readme_content = f"""# {assignment_path.replace('/', ' - ').title()}

This is the README for the assignment located at `{self.assignments_folder}/{assignment_path}`.

## Instructions

Please add your assignment instructions and requirements here.

## Submission

Please add your submission guidelines here.

---

*This README was automatically generated by the Assignment Pull Request Creator action.*
"""
            
            # Check if README already exists
            try:
                existing_file = self.repo.get_contents(readme_path, ref=branch_name)
                print(f"README already exists at {readme_path} in branch {branch_name}")
                return True
            except GithubException:
                # File doesn't exist, create it
                pass
            
            # Create the README file
            self.repo.create_file(
                path=readme_path,
                message=f"Add README for assignment {assignment_path}",
                content=readme_content,
                branch=branch_name
            )
            
            print(f"Created README.md at {readme_path} in branch {branch_name}")
            return True
            
        except GithubException as e:
            print(f"Error creating README for '{assignment_path}': {e}")
            return False

    def create_pull_request(self, assignment_path: str, branch_name: str) -> bool:
        """
        Create a pull request for the assignment branch.
        
        Args:
            assignment_path: Relative path to assignment folder
            branch_name: Branch name to create PR from
            
        Returns:
            True if PR was created, False otherwise
        """
        try:
            title = f"Assignment: {assignment_path.replace('/', ' - ').title()}"
            body = f"""## Assignment Pull Request

This pull request contains the setup for the assignment located at `{self.assignments_folder}/{assignment_path}`.

### Changes included:
- ✅ Created README.md with assignment template
- ✅ Set up branch structure for assignment submission

### Next steps:
1. Review the assignment requirements in the README.md
2. Add any additional assignment materials
3. Students can fork this repository and work on their submissions

---

*This pull request was automatically created by the Assignment Pull Request Creator action.*
"""
            
            # Create the pull request
            pr = self.repo.create_pull(
                title=title,
                body=body,
                head=branch_name,
                base=self.default_branch
            )
            
            print(f"Created pull request #{pr.number}: {title}")
            self.created_pull_requests.append(f"#{pr.number}")
            return True
            
        except GithubException as e:
            print(f"Error creating pull request for '{assignment_path}': {e}")
            return False

    def process_assignments(self) -> None:
        """Process all found assignments and create branches/PRs as needed."""
        assignments = self.find_assignments()
        
        if not assignments:
            print("No assignments found matching the criteria")
            return
        
        existing_branches = self.get_existing_branches()
        existing_prs = self.get_existing_pull_requests()
        
        print(f"Found {len(assignments)} assignments to process")
        print(f"Existing branches: {len(existing_branches)}")
        print(f"Existing PRs: {len(existing_prs)}")
        
        for assignment_path in assignments:
            branch_name = self.sanitize_branch_name(assignment_path)
            
            print(f"\nProcessing assignment: {assignment_path}")
            print(f"Branch name: {branch_name}")
            
            # Check if branch exists and if PR exists
            branch_exists = branch_name in existing_branches
            pr_exists = branch_name in existing_prs
            
            if not branch_exists:
                print(f"Branch '{branch_name}' does not exist, creating...")
                if not self.create_branch(branch_name):
                    continue
            else:
                print(f"Branch '{branch_name}' already exists")
            
            if not pr_exists:
                print(f"No PR exists for branch '{branch_name}', creating README and PR...")
                
                # Create README in the assignment folder
                if self.create_readme(assignment_path, branch_name):
                    # Create pull request
                    self.create_pull_request(assignment_path, branch_name)
                else:
                    print(f"Skipping PR creation due to README creation failure")
            else:
                print(f"PR already exists for branch '{branch_name}', skipping")

    def set_outputs(self) -> None:
        """Set GitHub Actions outputs."""
        # Set outputs for GitHub Actions
        if 'GITHUB_OUTPUT' in os.environ:
            with open(os.environ['GITHUB_OUTPUT'], 'a') as f:
                f.write(f"created-branches={json.dumps(self.created_branches)}\n")
                f.write(f"created-pull-requests={json.dumps(self.created_pull_requests)}\n")
        
        print(f"\nSummary:")
        print(f"Created branches: {self.created_branches}")
        print(f"Created pull requests: {self.created_pull_requests}")

    def run(self) -> None:
        """Main execution method."""
        try:
            print("Starting Assignment Pull Request Creator")
            print(f"Repository: {self.repository_name}")
            print(f"Assignments folder: {self.assignments_folder}")
            print(f"Assignment regex: {self.assignment_regex}")
            print(f"Default branch: {self.default_branch}")
            
            self.process_assignments()
            self.set_outputs()
            
            print("\nAssignment Pull Request Creator completed successfully")
            
        except Exception as e:
            print(f"Error: {e}")
            sys.exit(1)


if __name__ == "__main__":
    creator = AssignmentPRCreator()
    creator.run()
