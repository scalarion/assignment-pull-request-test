#!/usr/bin/env python3
"""
Master test runner for Assignment Pull Request Creator.

This script runs all test suites and provides a comprehensive test report.
"""

import sys
import os
import subprocess
import time

def run_command(cmd, description):
    """Run a command and return the result."""
    print(f"\n{'='*60}")
    print(f"🧪 {description}")
    print('='*60)
    
    start_time = time.time()
    
    try:
        result = subprocess.run(
            cmd,
            shell=True,
            capture_output=True,
            text=True,
            cwd=os.path.dirname(os.path.dirname(os.path.abspath(__file__)))  # Go to parent directory
        )
        
        end_time = time.time()
        duration = end_time - start_time
        
        print(result.stdout)
        if result.stderr:
            print("STDERR:")
            print(result.stderr)
        
        success = result.returncode == 0
        status = "✅ PASSED" if success else "❌ FAILED"
        print(f"\n{status} (took {duration:.2f}s)")
        
        return success, duration
        
    except Exception as e:
        print(f"❌ ERROR: {e}")
        return False, 0

def main():
    """Run all test suites."""
    print("Assignment Pull Request Creator - Master Test Runner")
    print("=" * 60)
    
    total_start = time.time()
    results = []
    
    # Test 1: Configuration tests
    success, duration = run_command(
        "python tests/test_config.py",
        "Testing Configuration & Utilities"
    )
    results.append(("Configuration Tests", success, duration))
    
    # Test 2: Focused unit tests
    success, duration = run_command(
        "python tests/test_focused.py",
        "Running Focused Unit Tests"
    )
    results.append(("Focused Unit Tests", success, duration))
    
    # Test 3: Main script dry-run
    success, duration = run_command(
        "DRY_RUN=true GITHUB_TOKEN=test GITHUB_REPOSITORY=test/repo python create_assignment_prs.py",
        "Testing Main Script (Dry-Run Mode)"
    )
    results.append(("Main Script Dry-Run", success, duration))
    
    # Test 4: Git mocking isolation test
    success, duration = run_command(
        "python tests/test_runner_mocked.py",
        "Testing Git Command Mocking"
    )
    results.append(("Git Mocking Tests", success, duration))
    
    # Summary
    total_end = time.time()
    total_duration = total_end - total_start
    
    print(f"\n{'='*60}")
    print("📊 TEST EXECUTION SUMMARY")
    print('='*60)
    
    passed = 0
    failed = 0
    
    for test_name, success, duration in results:
        status = "✅ PASSED" if success else "❌ FAILED"
        print(f"{test_name:.<40} {status} ({duration:.2f}s)")
        if success:
            passed += 1
        else:
            failed += 1
    
    print(f"\n{'='*60}")
    print(f"📈 OVERALL RESULTS:")
    print(f"   Total Tests: {len(results)}")
    print(f"   Passed: {passed}")
    print(f"   Failed: {failed}")
    print(f"   Total Time: {total_duration:.2f}s")
    
    if failed == 0:
        print(f"\n🎉 ALL TESTS PASSED!")
        print("✅ Test suite is ready for CI/CD")
        print("✅ Git command mocking verified")
        print("✅ GitHub API mocking verified")
        print("✅ Main functionality working")
        return_code = 0
    else:
        print(f"\n❌ {failed} TEST SUITE(S) FAILED")
        print("🔧 Please review failed tests and fix issues")
        return_code = 1
    
    print('='*60)
    return return_code

if __name__ == '__main__':
    exit_code = main()
    sys.exit(exit_code)
