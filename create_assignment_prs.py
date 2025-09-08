#!/usr/bin/env python3
"""
Assignment Pull Request Creator

This script scans for assignment folders and creates pull requests with README
files.
"""

import os
import re
import sys
import json
import subprocess
from pathlib import Path
from typing import Dict, List, Set
from github import Github
from github.GithubException import GithubException


class AssignmentPRCreator:
    """
    A GitHub automation tool for creating assignment branches and pull
    requests.

    This class automatically scans a repository for assignment directories,
    creates branches for each assignment, adds README files, and creates
    pull requests for assignment management and student submissions.

    The tool is designed to work with GitHub Actions and follows configurable
    regex patterns to identify assignment directories. It helps educators
    automate the setup of assignment repositories by:

    1. Scanning for assignment directories matching specified patterns
    2. Creating dedicated branches for each assignment locally
    3. Adding template README files to assignment directories locally
    4. Pushing all changes atomically to the remote repository
    5. Creating pull requests via GitHub API for assignment review and management

    IMPORTANT: All file and branch operations are performed locally first,
    then pushed atomically to the remote. This ensures repository consistency
    even if the process fails partway through. Only pull request creation
    uses the GitHub API directly.

    Environment Variables Required:
        GITHUB_TOKEN: GitHub personal access token with repository permissions
        GITHUB_REPOSITORY: Repository name in format "owner/repo"

    Environment Variables Optional:
        ASSIGNMENTS_ROOT_REGEX: Regex pattern for assignment root directories
            (default: "^assignments$")
        ASSIGNMENT_REGEX: Regex pattern for individual assignments
            (default: r"^assignment-\\d+$")
        DEFAULT_BRANCH: Default branch name (default: "main")
        DRY_RUN: Enable simulation mode without making actual changes
            (default: "false", accepts: "true", "1", "yes")

    Attributes:
        github_token (str): GitHub authentication token
        assignments_root_regex (str): Compiled regex for assignment root
            directories
        assignment_regex (str): Compiled regex for individual assignment
            directories
        repository_name (str): Full repository name
        default_branch (str): Default branch name for PR base
        dry_run (bool): Whether to simulate operations without making changes
        github (Github): PyGithub instance for API interactions
        repo (Repository): Repository object for operations
        root_pattern (Pattern): Compiled regex pattern for root directories
        assignment_pattern (Pattern): Compiled regex pattern for assignments
        created_branches (List[str]): List of branches created during execution
        created_pull_requests (List[str]): List of PRs created during execution
        pending_pushes (List[str]): List of branches pending push to remote

    Raises:
        ValueError: If required environment variables are missing
        GithubException: If GitHub API operations fail

    Example:
        Basic usage in a GitHub Action:

        creator = AssignmentPRCreator()
        creator.run()

        The tool will automatically process all assignments, create the
        necessary branches and files locally, push them atomically to remote,
        and create pull requests.
    """

    def __init__(self):
        """Initialize the Assignment PR Creator with environment variables."""
        self.default_branch = os.environ.get("DEFAULT_BRANCH", "main")
        self.github_token = os.environ.get("GITHUB_TOKEN")
        self.repository_name = os.environ.get("GITHUB_REPOSITORY")
        self.assignments_root_regex = os.environ.get(
            "ASSIGNMENTS_ROOT_REGEX", "^assignments$"
        )
        self.assignment_regex = os.environ.get(
            "ASSIGNMENT_REGEX", r"^assignment-\d+$"
        )
        self.dry_run = os.environ.get("DRY_RUN", "false").lower() in ("true", "1", "yes")

        if not self.github_token:
            raise ValueError("GITHUB_TOKEN environment variable is required")
        if not self.repository_name:
            raise ValueError(
                "GITHUB_REPOSITORY environment variable is required"
            )

        # Only initialize GitHub API connection for PR operations (not in dry-run)
        if not self.dry_run:
            self.github = Github(self.github_token)
            self.repo = self.github.get_repo(self.repository_name)
        else:
            # In dry-run mode, we don't need actual GitHub API access
            self.github = None
            self.repo = None

        # Compile regex patterns
        self.root_pattern = re.compile(self.assignments_root_regex)
        self.assignment_pattern = re.compile(self.assignment_regex)

        # Track created items
        self.created_branches: List[str] = []
        self.created_pull_requests: List[str] = []
        self.pending_pushes: List[str] = []

    def sanitize_branch_name(self, assignment_path: str) -> str:
        """
        Sanitize assignment path to create a valid branch name.

        Args:
            assignment_path: Relative path of assignment from workspace root

        Returns:
            Sanitized branch name
        """
        # Remove leading/trailing whitespace
        branch_name = assignment_path.strip()

        # Replace spaces with hyphens
        branch_name = re.sub(r"\s+", "-", branch_name)

        # Remove slashes
        branch_name = branch_name.replace("/", "-")

        # Remove consecutive hyphens
        branch_name = re.sub(r"-+", "-", branch_name)

        # Convert to lowercase
        branch_name = branch_name.lower()

        # Remove leading/trailing hyphens
        branch_name = branch_name.strip("-")

        return branch_name

    def run_git_command(self, command: str, description: str = "") -> bool:
        """
        Run a git command, either for real or simulate in dry-run mode.
        
        Args:
            command: Full command string (e.g., 'git checkout -b branch-name')
            description: Optional description of what the command does
            
        Returns:
            True if command succeeded (or was simulated), False otherwise
        """
        if self.dry_run:
            print(f"[DRY RUN] {description}: {command}")
            return True
        
        try:
            if description:
                print(f"{description}: {command}")
            
            result = subprocess.run(
                command,
                shell=True,
                capture_output=True,
                text=True,
                check=True
            )
            
            if result.stdout.strip():
                print(f"  Output: {result.stdout.strip()}")
            
            return True
            
        except subprocess.CalledProcessError as e:
            print(f"Error running command '{command}': {e}")
            if e.stderr:
                print(f"  Stderr: {e.stderr}")
            sys.exit(1)

    def run_git_command_with_output(self, command: str, description: str = "") -> str:
        """
        Run a git command and return its output, either for real or simulate in dry-run mode.
        
        Args:
            command: Full command string (e.g., 'git rev-list --count HEAD')
            description: Optional description of what the command does
            
        Returns:
            Command output as string
        """
        if self.dry_run:
            print(f"[DRY RUN] {description}: {command}")
            return ""  # Return empty string for dry-run
        
        try:
            if description:
                print(f"{description}: {command}")
            
            result = subprocess.run(
                command,
                shell=True,
                capture_output=True,
                text=True,
                check=True
            )
            
            return result.stdout.strip()
            
        except subprocess.CalledProcessError as e:
            print(f"Error running command '{command}': {e}")
            if e.stderr:
                print(f"  Stderr: {e.stderr}")
            sys.exit(1)

    def create_branch(self, branch_name: str) -> bool:
        """
        Create a new branch from the default branch locally.

        Args:
            branch_name: Name of the branch to create

        Returns:
            True if branch was created, False otherwise
        """
        # First, ensure we're on the default branch
        if not self.run_git_command(
            f'git checkout {self.default_branch}',
            f"Switch to default branch '{self.default_branch}'"
        ):
            return False
            
        # Create and switch to new branch
        if not self.run_git_command(
            f'git checkout -b {branch_name}',
            f"Create and switch to branch '{branch_name}'"
        ):
            return False

        print(f"✅ Created branch: {branch_name} (local)")
        self.created_branches.append(branch_name)
        self.pending_pushes.append(branch_name)
        return True

    def create_readme(self, assignment_path: str, branch_name: str) -> bool:
        """
        Create or augment README.md file in the assignment folder locally.

        Args:
            assignment_path: Relative path to assignment folder
            branch_name: Branch name to commit to

        Returns:
            True if README was created/augmented, False otherwise
        """
        readme_path = Path(assignment_path) / "README.md"
        assignment_title = assignment_path.replace("/", " - ").title()

        # Create assignment directory if it doesn't exist
        assignment_dir = Path(assignment_path)
        if not self.dry_run:
            assignment_dir.mkdir(parents=True, exist_ok=True)
        else:
            print(f"[DRY RUN] Would create directory: mkdir -p {assignment_path}")

        # Check if README already exists
        if readme_path.exists():
            print(f"README already exists at {readme_path}, augmenting...")
            
            # Read existing content
            existing_content = readme_path.read_text(encoding='utf-8')
            
            # Add workflow augmentation comment
            augmentation_comment = f"""

---

*This README was augmented by the Assignment Pull Request Creator action.*
"""
            
            augmented_content = existing_content.rstrip() + augmentation_comment
            
            # Write augmented content
            if not self.dry_run:
                readme_path.write_text(augmented_content, encoding='utf-8')
                print(f"✅ Augmented README.md at {readme_path} (local)")
            else:
                print(f"[DRY RUN] Would augment README at {readme_path}")
                print("[DRY RUN] Augmentation content:")
                print(augmentation_comment)
        else:
            # Create new README content
            readme_content = f"""# {assignment_title}

This is the README for the assignment located at `{assignment_path}`.

## Instructions

Please add your assignment instructions and requirements here.

## Submission

Please add your submission guidelines here.

---

*This README was automatically generated by the Assignment Pull Request*
*Creator action.*
"""

            # Write new README
            if not self.dry_run:
                readme_path.write_text(readme_content, encoding='utf-8')
                print(f"✅ Created README.md at {readme_path} (local)")
            else:
                print(f"[DRY RUN] Would create README at {readme_path}")
                print("[DRY RUN] README content:")
                print(readme_content)

        # Add and commit the README
        if not self.run_git_command(
            f'git add {readme_path}',
            f"Stage README file"
        ):
            return False
            
        commit_message = f"Add README for assignment {assignment_path}" if not readme_path.exists() or self.dry_run else f"Augment README for assignment {assignment_path}"
        
        if not self.run_git_command(
            f'git commit -m "{commit_message}"',
            f"Commit README changes"
        ):
            return False

        return True

    def push_branches_to_remote(self) -> bool:
        """
        Push all created branches atomically to the remote repository.
        Uses git push --all for maximum simplicity and true atomicity.
        
        In GitHub Actions, the local repository is ephemeral and destroyed after
        the action completes, so tracking setup is unnecessary.
        
        Returns:
            True if all branches were pushed successfully, False otherwise
        """
        if not self.pending_pushes:
            print("No branches to push to remote")
            return True
            
        print(f"Pushing all local branches (including {len(self.pending_pushes)} new branches) to remote atomically...")
        
        # Use git push --all for true atomic push of all local branches
        # This is safe in GitHub Actions environment where we start with clean checkout
        if not self.run_git_command(
            'git push --all origin',
            f"Atomically push all local branches to remote"
        ):
            return False
        
        print(f"✅ Successfully pushed all local branches to remote atomically")
        self.pending_pushes.clear()
        return True
                    
    def create_pull_request(self, assignment_path: str, branch_name: str) -> bool:
        """
        Create a pull request for the assignment branch using GitHub API.

        Args:
            assignment_path: Relative path to assignment folder
            branch_name: Branch name to create PR from

        Returns:
            True if PR was created, False otherwise
        """
        title = f"Assignment: {assignment_path.replace('/', ' - ').title()}"
        body = f"""## Assignment Pull Request

This pull request contains the setup for the assignment located at
`{assignment_path}`.

### Changes included:
- ✅ Created README.md with assignment template
- ✅ Set up branch structure for assignment submission

### Next steps:
1. Review the assignment requirements in the README.md
2. Add any additional assignment materials
3. Students can fork this repository and work on their submissions

---

*This pull request was automatically created by the Assignment Pull*
*Request Creator action.*
"""

        if self.dry_run:
            print(f"[DRY RUN] Would create pull request:")
            print(f"  Title: {title}")
            print(f"  Head: {branch_name}")
            print(f"  Base: {self.default_branch}")
            print(f"  Body: {body[:100]}...")
            
            # Simulate PR number
            simulated_pr_number = len(self.created_pull_requests) + 1
            print(f"[DRY RUN] Simulated pull request #{simulated_pr_number}")
            self.created_pull_requests.append(f"#{simulated_pr_number}")
            return True

        try:
            # Create the pull request via GitHub API
            pr = self.repo.create_pull(
                title=title,
                body=body,
                head=branch_name,
                base=self.default_branch,
            )

            print(f"✅ Created pull request #{pr.number}: {title}")
            self.created_pull_requests.append(f"#{pr.number}")
            return True

        except GithubException as e:
            print(f"Error creating pull request for '{assignment_path}': {e}")
            sys.exit(1)

    def find_assignments(self) -> List[str]:
        """
        Find all assignment folders that match the regex patterns.

        Returns:
            List of relative paths to assignment folders
        """
        assignments = []
        workspace_root = Path(os.getcwd())

        print(
            f"Scanning workspace for assignment roots matching "
            f"'{self.assignments_root_regex}'"
        )
        print(f"Looking for assignments matching '{self.assignment_regex}'")

        # First, find all directories that match the root pattern
        for root, dirs, _ in os.walk(workspace_root):
            root_path = Path(root)

            # Check each directory against the root pattern
            for dir_name in dirs:
                if self.root_pattern.match(dir_name):
                    assignments_root = root_path / dir_name
                    print(f"Found assignment root: {assignments_root}")

                    # Now scan for individual assignments within this root
                    if assignments_root.exists():
                        for assignment_root, _, _ in os.walk(
                            assignments_root
                        ):
                            assignment_root_path = Path(assignment_root)

                            # Check if the current directory itself matches the assignment pattern
                            # But skip the root assignments directory itself
                            current_dir_name = assignment_root_path.name
                            if (self.assignment_pattern.match(current_dir_name) and 
                                assignment_root_path != assignments_root):
                                # Get relative path from workspace root
                                relative_path = assignment_root_path.relative_to(
                                    workspace_root
                                )
                                assignments.append(str(relative_path))
                                print(f"Found assignment: {relative_path}")

        return assignments

    def fetch_all_remote_branches(self) -> bool:
        """
        Fetch all remote branches to ensure local repository has complete state.
        This is essential in GitHub Actions to get all existing branches locally.
        
        Returns:
            True if fetch succeeded, False otherwise
        """
        print("Fetching all remote branches to local repository...")
        
        # Fetch all remote branches and tags
        if not self.run_git_command(
            'git fetch --all',
            "Fetch all remote branches and tags"
        ):
            return False
            
        # Create local tracking branches for all remote branches (except HEAD)
        if self.dry_run:
            print("[DRY RUN] Would create local tracking branches for all remote branches")
            return True
            
        try:
            # Get list of remote branches
            output = self.run_git_command_with_output(
                'git branch -r',
                "List remote branches"
            )
            
            for line in output.split('\n'):
                if line.strip() and not line.strip().endswith('/HEAD'):
                    # Format: "  origin/branch-name"
                    remote_branch = line.strip()
                    if remote_branch.startswith('origin/'):
                        branch_name = remote_branch.replace('origin/', '')
                        # Skip default branch as it already exists locally
                        if branch_name != self.default_branch:
                            self.run_git_command(
                                f'git checkout -b {branch_name} {remote_branch}',
                                f"Create local tracking branch for {branch_name}"
                            )
            
            # Return to default branch
            self.run_git_command(
                f'git checkout {self.default_branch}',
                f"Return to default branch {self.default_branch}"
            )
            
            return True
            
        except Exception as e:
            print(f"Error setting up local tracking branches: {e}")
            return False

    def get_existing_branches(self) -> Set[str]:
        """
        Get all existing branches in the local repository.
        This should be called after fetch_all_remote_branches().

        Returns:
            Set of branch names
        """
        branches = set()
        
        if self.dry_run:
            print("[DRY RUN] Would check local branches with command:")
            print("  git branch")
            # Return empty set for dry-run to simulate clean repository
            return set()
        
        try:
            # Get local branches
            output = self.run_git_command_with_output(
                'git branch',
                "Get local branches"
            )
            
            for line in output.split('\n'):
                if line.strip():
                    # Format: "* main" or "  branch-name"
                    branch_name = line.strip().replace('* ', '').strip()
                    if branch_name:
                        branches.add(branch_name)
            
            print(f"Found {len(branches)} local branches")
            return branches
            
        except Exception as e:
            print(f"Error getting local branches: {e}")
            sys.exit(1)

    def get_existing_pull_requests(self) -> Dict[str, str]:
        """
        Get all existing pull request head branch names and their states.

        Returns:
            Dictionary mapping branch names to their PR states (open/closed)
        """
        if self.dry_run:
            print("[DRY RUN] Would check existing pull requests with GitHub API")
            # Return empty dict for dry-run to simulate no existing PRs
            return {}
            
        try:
            pulls = self.repo.get_pulls(state="all")
            return {pr.head.ref: pr.state for pr in pulls}
        except GithubException as e:
            print(f"Error getting pull requests: {e}")
            sys.exit(1)

    def has_branch_changes(self, branch_name: str) -> bool:
        """
        Check if a branch has changes compared to the default branch.
        
        Args:
            branch_name: Name of the branch to compare
            
        Returns:
            True if branch has changes, False otherwise
        """
        if self.dry_run:
            print(f"[DRY RUN] Would check for changes between '{branch_name}' and '{self.default_branch}' with git diff")
            # In dry-run, assume there are changes
            return True
        
        try:
            # Use git to compare the branches locally
            output = self.run_git_command_with_output(
                f'git rev-list --count {self.default_branch}..{branch_name}',
                f"Count commits ahead of {self.default_branch}"
            )
            
            ahead_by = int(output)
            
            print(f"Branch '{branch_name}' has {ahead_by} commits ahead of '{self.default_branch}'")
            if ahead_by == 0:
                print(f"No commits found. Branch '{branch_name}' is up to date with '{self.default_branch}'")
            
            return ahead_by > 0
            
        except (ValueError, Exception) as e:
            print(f"Error comparing branches '{self.default_branch}' and '{branch_name}': {e}")
            sys.exit(1)

    def process_assignments(self) -> None:
        """
        Process all found assignments and create branches/PRs as needed.
        
        NEW APPROACH: Complete local git operations with atomic remote push
        
        1. Fetch all remote branches to ensure local repository has complete state
        2. Process all assignments locally (branches, commits)
        3. Push all changes atomically to remote 
        4. Create pull requests via GitHub API
        
        This ensures repository consistency - the local repo is synced with remote
        first, then all work is done locally, and only successful work is pushed
        atomically to remote.
        
        Branch Creation Logic:
        - Creates branch locally only if no local branch exists AND no PR has ever existed
        - Prevents recreating branches for completed assignments (merged PRs)
        
        Pull Request Creation Logic:  
        - Creates PR only if NO PR has ever existed for the branch
        - All branches are pushed before PR creation to ensure remote consistency
        """
        assignments = self.find_assignments()

        if not assignments:
            print("No assignments found matching the criteria")
            return

        # Phase 0: Fetch all remote branches to ensure complete local state
        print("\n=== Phase 0: Syncing with remote ===")
        if not self.fetch_all_remote_branches():
            print("❌ Failed to fetch remote branches, aborting")
            return

        # Phase 1: Get current state after sync
        existing_branches = self.get_existing_branches()
        existing_prs = self.get_existing_pull_requests()

        print(f"Found {len(assignments)} assignments to process")
        print(f"Existing local branches: {len(existing_branches)}")
        print(f"Existing PRs: {len(existing_prs)}")

        # Phase 2: Process all assignments locally
        print("\n=== Phase 2: Local processing ===")
        branches_to_process = []
        
        for assignment_path in assignments:
            branch_name = self.sanitize_branch_name(assignment_path)

            print(f"\nProcessing assignment: {assignment_path}")
            print(f"Branch name: {branch_name}")

            # Check if branch exists locally and if PR exists (or has ever existed)
            branch_exists = branch_name in existing_branches
            pr_has_existed = branch_name in existing_prs

            # Only create branch if:
            # 1. Branch doesn't exist locally AND
            # 2. No PR has ever existed for this branch name
            if not branch_exists and not pr_has_existed:
                print(f"Branch '{branch_name}' does not exist locally and no PR has ever existed, creating locally...")
                
                # Create branch locally
                if self.create_branch(branch_name):
                    # Create README content locally
                    print(f"Creating README content for assignment '{assignment_path}'...")
                    if self.create_readme(assignment_path, branch_name):
                        branches_to_process.append((assignment_path, branch_name))
                    else:
                        print(f"❌ Failed to create README for '{assignment_path}', skipping")
                else:
                    print(f"❌ Failed to create branch '{branch_name}', skipping")
                    
            elif not branch_exists and pr_has_existed:
                print(f"Branch '{branch_name}' does not exist but PR has existed before (likely merged and branch deleted), skipping")
                continue
            elif branch_exists and not pr_has_existed:
                print(f"Branch '{branch_name}' already exists locally but no PR has ever existed, will create PR")
                branches_to_process.append((assignment_path, branch_name))
            elif branch_exists and pr_has_existed:
                print(f"Branch '{branch_name}' already exists locally and PR has existed before, skipping")

        # Phase 3: Push all changes atomically to remote
        if branches_to_process:
            print(f"\n=== Phase 3: Atomic push to remote ===")
            print(f"Pushing {len(branches_to_process)} branches to remote...")
            
            if not self.push_branches_to_remote():
                print("❌ Failed to push branches to remote, aborting PR creation")
                return
                
        # Phase 4: Create pull requests
        if branches_to_process:
            print(f"\n=== Phase 4: Pull request creation ===")
            
            for assignment_path, branch_name in branches_to_process:
                pr_has_existed = branch_name in existing_prs
                
                # Double-check PR status (should still be false)
                if not pr_has_existed:
                    print(f"Creating pull request for branch '{branch_name}'...")
                    self.create_pull_request(assignment_path, branch_name)
                else:
                    print(f"PR has existed for branch '{branch_name}', skipping PR creation")
        
        if not branches_to_process:
            print("\n=== No new assignments to process ===")
            print("All assignments either already have branches or have had PRs created previously")

    def set_outputs(self) -> None:
        """Set GitHub Actions outputs."""
        # Set outputs for GitHub Actions
        if "GITHUB_OUTPUT" in os.environ:
            with open(os.environ["GITHUB_OUTPUT"], "a", encoding="utf-8") as f:
                f.write(
                    f"created-branches={json.dumps(self.created_branches)}\n"
                )
                f.write(
                    "created-pull-requests="
                    f"{json.dumps(self.created_pull_requests)}\n"
                )

        print("\nSummary:")
        print(f"Created branches: {self.created_branches}")
        print(f"Created pull requests: {self.created_pull_requests}")

    def run(self) -> None:
        """Main execution method using local git with atomic remote operations."""
        try:
            print("Starting Assignment Pull Request Creator")
            if self.dry_run:
                print("🏃 DRY RUN MODE: Simulating local git operations without making actual changes")
            else:
                print("🔄 LIVE MODE: Using local git operations with atomic remote push")
            print(f"Repository: {self.repository_name}")
            print(f"Assignments root regex: {self.assignments_root_regex}")
            print(f"Assignment regex: {self.assignment_regex}")
            print(f"Default branch: {self.default_branch}")
            print(f"Dry run mode: {self.dry_run}")

            self.process_assignments()
            self.set_outputs()

            if self.dry_run:
                print("\n🏃 DRY RUN MODE: Assignment Pull Request Creator simulation completed")
                print("In real mode, all local changes would be pushed atomically to remote")
            else:
                print("\nAssignment Pull Request Creator completed successfully")
                print("All changes have been pushed to remote repository")

        except (ValueError, GithubException, subprocess.CalledProcessError) as e:
            print(f"Error: {e}")
            sys.exit(1)


if __name__ == "__main__":
    creator = AssignmentPRCreator()
    creator.run()
