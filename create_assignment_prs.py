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
from pathlib import Path
from typing import List, Set
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
    2. Creating dedicated branches for each assignment
    3. Adding template README files to assignment directories
    4. Creating pull requests for assignment review and management

    IMPORTANT: All operations are performed directly on the remote repository
    via GitHub API. No local git operations are involved. Branches, commits,
    and pull requests are immediately available on the remote repository.

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

    Raises:
        ValueError: If required environment variables are missing
        GithubException: If GitHub API operations fail

    Example:
        Basic usage in a GitHub Action:

        creator = AssignmentPRCreator()
        creator.run()

        The tool will automatically process all assignments and create
        the necessary branches and pull requests.
    """

    def __init__(self):
        """Initialize the Assignment PR Creator with environment variables."""
        self.github_token = os.environ.get("GITHUB_TOKEN")
        self.assignments_root_regex = os.environ.get(
            "ASSIGNMENTS_ROOT_REGEX", "^assignments$"
        )
        self.assignment_regex = os.environ.get(
            "ASSIGNMENT_REGEX", r"^assignment-\d+$"
        )
        self.repository_name = os.environ.get("GITHUB_REPOSITORY")
        self.default_branch = os.environ.get("DEFAULT_BRANCH", "main")
        self.dry_run = os.environ.get("DRY_RUN", "false").lower() in ("true", "1", "yes")

        if not self.github_token:
            raise ValueError("GITHUB_TOKEN environment variable is required")
        if not self.repository_name:
            raise ValueError(
                "GITHUB_REPOSITORY environment variable is required"
            )

        # Only initialize GitHub API connection if not in dry-run mode
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

    def simulate_branch_creation(self, branch_name: str) -> bool:
        """
        Simulate branch creation by outputting equivalent git commands.
        
        NOTE: The actual implementation uses GitHub API, not git commands.
        These commands show what the equivalent git operations would be.

        Args:
            branch_name: Name of the branch to simulate creating

        Returns:
            True (simulation always succeeds)
        """
        print(f"[DRY RUN] Would create branch with command:")
        print(f"  git checkout -b {branch_name} {self.default_branch}")
        print(f"  git push -u origin {branch_name}")
        print(f"  # Note: Actual implementation uses GitHub API directly")
        self.created_branches.append(branch_name)
        return True

    def simulate_readme_creation(self, assignment_path: str, branch_name: str) -> bool:
        """
        Simulate README creation by outputting equivalent git commands.
        
        NOTE: The actual implementation uses GitHub API, not git commands.
        These commands show what the equivalent git operations would be.

        Args:
            assignment_path: Relative path to assignment folder
            branch_name: Branch name to simulate committing to

        Returns:
            True (simulation always succeeds)
        """
        readme_path = f"{assignment_path}/README.md"
        assignment_title = assignment_path.replace("/", " - ").title()
        
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

        print(f"[DRY RUN] Would create README.md at {readme_path} with content:")
        print("--- README.md content ---")
        print(readme_content)
        print("--- End README.md content ---")
        print(f"[DRY RUN] Would commit with commands:")
        print(f"  # Check if README exists first")
        print(f"  if [ -f {readme_path} ]; then")
        print(f"    # README exists - augment it")
        print(f"    echo '' >> {readme_path}")
        print(f"    echo '---' >> {readme_path}")
        print(f"    echo '' >> {readme_path}")
        print(f"    echo '*This README was augmented by the Assignment Pull Request Creator action.*' >> {readme_path}")
        print(f"    git add {readme_path}")
        print(f"    git commit -m 'Augment README for assignment {assignment_path}'")
        print(f"  else")
        print(f"    # README doesn't exist - create it")
        print(f"    git checkout {branch_name}")
        print(f"    mkdir -p {assignment_path}")
        print(f"    echo '[content]' > {readme_path}")
        print(f"    git add {readme_path}")
        print(f"    git commit -m 'Add README for assignment {assignment_path}'")
        print(f"  fi")
        print(f"  git push origin {branch_name}")
        print(f"  # Note: Actual implementation uses GitHub API directly")
        return True

    def simulate_pull_request_creation(self, assignment_path: str, branch_name: str) -> bool:
        """
        Simulate pull request creation by outputting the GitHub CLI command.
        Assumes changes have already been validated before calling this method.

        Args:
            assignment_path: Relative path to assignment folder
            branch_name: Branch name to simulate creating PR from

        Returns:
            True (simulation always succeeds)
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

        print(f"[DRY RUN] Would create pull request with command:")
        print(f"  gh pr create \\")
        print(f"    --title '{title}' \\")
        print(f"    --body '{body.replace(chr(10), chr(10) + '    ')}' \\")
        print(f"    --head {branch_name} \\")
        print(f"    --base {self.default_branch}")
        
        # Simulate PR number
        simulated_pr_number = len(self.created_pull_requests) + 1
        print(f"[DRY RUN] Simulated pull request #{simulated_pr_number}: {title}")
        self.created_pull_requests.append(f"#{simulated_pr_number}")
        return True

    def find_assignments(self) -> List[str]:
        """
        Find all assignment folders that match the regex patterns.

        Returns:
            List of relative paths to assignment folders
        """
        assignments = []
        workspace_root = Path(".")

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

    def get_existing_branches(self) -> Set[str]:
        """
        Get all existing branches in the repository.

        Returns:
            Set of branch names
        """
        if self.dry_run:
            print("[DRY RUN] Would check existing branches with command:")
            print("  gh api repos/:owner/:repo/branches --jq '.[].name'")
            # Return empty set for dry-run to simulate no existing branches
            return set()
            
        try:
            branches = self.repo.get_branches()
            return {branch.name for branch in branches}
        except GithubException as e:
            print(f"Error getting branches: {e}")
            sys.exit(1)

    def get_existing_pull_requests(self) -> Set[str]:
        """
        Get all existing pull request head branch names (open and closed).

        Returns:
            Set of branch names that have or have had pull requests
        """
        if self.dry_run:
            print("[DRY RUN] Would check existing pull requests with command:")
            print("  gh api repos/:owner/:repo/pulls --jq '.[].head.ref'")
            # Return empty set for dry-run to simulate no existing PRs
            return set()
            
        try:
            pulls = self.repo.get_pulls(state="all")
            return {pr.head.ref for pr in pulls}
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
            print(f"[DRY RUN] Would check for changes between '{branch_name}' and '{self.default_branch}' with command:")
            print(f"  gh api repos/:owner/:repo/compare/{self.default_branch}...{branch_name} --jq '.ahead_by'")
            # In dry-run, assume there are changes
            return True
            
        try:
            # Compare the branch with the default branch
            comparison = self.repo.compare(self.default_branch, branch_name)
            
            # Check if there are commits ahead (changes in the branch)
            has_changes = comparison.ahead_by > 0
            
            print(f"Branch '{branch_name}' has {comparison.ahead_by} commits ahead of '{self.default_branch}'")
            if comparison.ahead_by == 0:
                print(f"No commits found. Branch '{branch_name}' is up to date with '{self.default_branch}'")
            
            return has_changes
            
        except GithubException as e:
            print(f"Error comparing branches '{self.default_branch}' and '{branch_name}': {e}")
            sys.exit(1)

    def create_branch(self, branch_name: str) -> bool:
        """
        Create a new branch from the default branch.

        Args:
            branch_name: Name of the branch to create

        Returns:
            True if branch was created, False otherwise
        """
        if self.dry_run:
            return self.simulate_branch_creation(branch_name)
            
        try:
            # Get the default branch reference
            default_ref = self.repo.get_git_ref(f"heads/{self.default_branch}")

            # Create new branch directly on remote repository via GitHub API
            # (This immediately creates the branch on the remote, no local git operations)
            self.repo.create_git_ref(
                ref=f"refs/heads/{branch_name}", sha=default_ref.object.sha
            )

            print(f"✅ Created branch: {branch_name} (pushed to remote)")
            self.created_branches.append(branch_name)
            return True

        except GithubException as e:
            print(f"Error creating branch '{branch_name}': {e}")
            sys.exit(1)

    def create_readme(self, assignment_path: str, branch_name: str) -> bool:
        """
        Create README.md file in the assignment folder.

        Args:
            assignment_path: Relative path to assignment folder
            branch_name: Branch name to commit to

        Returns:
            True if README was created, False otherwise
        """
        if self.dry_run:
            return self.simulate_readme_creation(assignment_path, branch_name)
            
        try:
            readme_path = f"{assignment_path}/README.md"

            # Create README content
            assignment_title = assignment_path.replace("/", " - ").title()
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

            # Check if README already exists
            try:
                existing_file = self.repo.get_contents(readme_path, ref=branch_name)
                print(
                    f"README already exists at {readme_path} "
                    f"in branch {branch_name} (SHA: {existing_file.sha})"
                )
                
                # Augment existing README by appending workflow comment
                existing_content = existing_file.content
                if isinstance(existing_content, str):
                    # Content from GitHub API is base64 encoded string
                    import base64
                    existing_content = base64.b64decode(existing_content).decode('utf-8')
                elif isinstance(existing_content, bytes):
                    existing_content = existing_content.decode('utf-8')
                else:
                    existing_content = str(existing_content)
                
                # Add workflow augmentation comment
                augmentation_comment = f"""

---

*This README was augmented by the Assignment Pull Request Creator action.*
"""
                
                augmented_content = existing_content.rstrip() + augmentation_comment
                
                # Update the existing file with augmented content
                commit_info = self.repo.update_file(
                    path=readme_path,
                    message=f"Augment README for assignment {assignment_path}",
                    content=augmented_content,
                    sha=existing_file.sha,
                    branch=branch_name,
                )
                
                print(
                    f"✅ Augmented existing README.md at {readme_path} in branch {branch_name} (pushed to remote)"
                )
                print(f"   Commit SHA: {commit_info['commit'].sha}")
                return True
                
            except GithubException:
                # File doesn't exist, create it
                print(f"README does not exist at {readme_path}, creating...")
                pass

            # Create the README file directly on remote repository via GitHub API
            # (This immediately creates the file and commit on the remote, no local git operations)
            commit_info = self.repo.create_file(
                path=readme_path,
                message=f"Add README for assignment {assignment_path}",
                content=readme_content,
                branch=branch_name,
            )

            print(
                f"✅ Created README.md at {readme_path} in branch {branch_name} (pushed to remote)"
            )
            print(f"   Commit SHA: {commit_info['commit'].sha}")
            return True

        except GithubException as e:
            print(f"Error creating README for '{assignment_path}': {e}")
            sys.exit(1)

    def create_pull_request(
        self, assignment_path: str, branch_name: str
    ) -> bool:
        """
        Create a pull request for the assignment branch.
        Assumes changes have already been validated before calling this method.

        Args:
            assignment_path: Relative path to assignment folder
            branch_name: Branch name to create PR from

        Returns:
            True if PR was created, False otherwise
        """
        if self.dry_run:
            return self.simulate_pull_request_creation(assignment_path, branch_name)
            
        try:
            title = (
                f"Assignment: {assignment_path.replace('/', ' - ').title()}"
            )
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

            # Create the pull request directly on remote repository via GitHub API
            # (This immediately creates the PR on the remote, no local git operations)
            pr = self.repo.create_pull(
                title=title,
                body=body,
                head=branch_name,
                base=self.default_branch,
            )

            print(f"✅ Created pull request #{pr.number}: {title} (available on remote)")
            self.created_pull_requests.append(f"#{pr.number}")
            return True

        except GithubException as e:
            print(f"Error creating pull request for '{assignment_path}': {e}")
            sys.exit(1)

    def process_assignments(self) -> None:
        """
        Process all found assignments and create branches/PRs as needed.
        
        Implements smart logic to handle assignment lifecycle:
        
        Branch Creation:
        - Creates branch only if no branch exists AND no PR has ever existed
        - Prevents recreating branches for completed assignments (merged PRs)
        
        Pull Request Creation:  
        - Creates PR only if NO PR has ever existed for the branch AND branch has changes
        - Prevents creating duplicate PRs and invalid PRs with no changes
        """
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

            # Check if branch exists and if PR exists (or has ever existed)
            branch_exists = branch_name in existing_branches
            pr_has_existed = branch_name in existing_prs

            # Only create branch if:
            # 1. Branch doesn't exist AND
            # 2. No PR has ever existed for this branch name
            if not branch_exists and not pr_has_existed:
                print(f"Branch '{branch_name}' does not exist and no PR has ever existed, creating...")
                self.create_branch(branch_name)
                branch_exists = True  # Branch now exists (or simulated)
            elif not branch_exists and pr_has_existed:
                print(f"Branch '{branch_name}' does not exist but PR has existed before (likely merged and branch deleted), skipping branch creation")
                continue
            elif branch_exists:
                print(f"Branch '{branch_name}' already exists")

            # Only create PR if NO PR has ever existed for this branch name
            if not pr_has_existed and branch_exists:
                print(
                    f"No PR has ever existed for branch '{branch_name}', "
                    f"preparing assignment content and PR..."
                )

                # First, create README in the assignment folder to ensure we have changes
                print(f"Creating README content for assignment '{assignment_path}'...")
                self.create_readme(assignment_path, branch_name)

                # Add a small delay to ensure GitHub API consistency (only in live mode)
                # (the commit might need a moment to be reflected in branch comparison)
                if not self.dry_run:
                    print("Waiting for GitHub API consistency...")
                    import time
                    time.sleep(1)

                # Then check if there are changes (there should be after README creation)
                print(f"Checking for changes in branch '{branch_name}'...")
                if not self.has_branch_changes(branch_name):
                    print(f"❌ No changes detected in branch '{branch_name}' after content creation, skipping PR creation")
                    continue

                # Finally, create pull request
                print(f"✅ Changes detected, creating pull request...")
                self.create_pull_request(assignment_path, branch_name)
            elif pr_has_existed:
                print(
                    f"PR has existed before for branch '{branch_name}', skipping PR creation"
                )

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
        """Main execution method."""
        try:
            print("Starting Assignment Pull Request Creator")
            if self.dry_run:
                print("🏃 DRY RUN MODE: Simulating operations without making actual changes")
            print(f"Repository: {self.repository_name}")
            print(f"Assignments root regex: {self.assignments_root_regex}")
            print(f"Assignment regex: {self.assignment_regex}")
            print(f"Default branch: {self.default_branch}")
            print(f"Dry run mode: {self.dry_run}")

            self.process_assignments()
            self.set_outputs()

            if self.dry_run:
                print("\n🏃 DRY RUN MODE: Assignment Pull Request Creator simulation completed")
            else:
                print("\nAssignment Pull Request Creator completed successfully")

        except (ValueError, GithubException) as e:
            print(f"Error: {e}")
            sys.exit(1)


if __name__ == "__main__":
    creator = AssignmentPRCreator()
    creator.run()
