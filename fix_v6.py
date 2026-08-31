import subprocess
import os
import sys

os.chdir(r'E:\tools20260623')
sys.stdout.reconfigure(encoding='utf-8', errors='replace')

def run(cmd, check=True):
    print(f"  > {' '.join(cmd)}")
    r = subprocess.run(cmd, capture_output=True, text=True, encoding='utf-8', errors='replace')
    if r.stdout.strip():
        print(f"  OUT: {r.stdout.strip()[:200]}")
    if r.stderr.strip():
        print(f"  ERR: {r.stderr.strip()[:200]}")
    if check and r.returncode != 0:
        raise Exception(f"Command failed: {' '.join(cmd)}")
    return r

# The 4 commits with wrong email on fix-temp
bad_commits = [
    "dc20a7fb2cb7b99346c2f6e77feec6c1f0971483",  # fix: add __pycache__ to .gitignore
    "0fcbd65c7531d7987e980f7cac33b5b964e0a5c2",  # feat: delete conversation API + home page cleanup
    "34fa9d8d06303b58d863bf98bb82cbbec21e3859",  # 20260822-0001: update README with latest updates
    "0100b5253d6cb53f592304034abd0c296b94923e",  # chore: remove dist/index.html from tracking
]

for bad_commit in bad_commits:
    print(f"\n=== Amending commit {bad_commit[:8]} ===")
    r = run(['git', 'rebase', '--onto', 'HEAD', bad_commit, bad_commit + '~1'], check=False)
    
    # Use git replace + rebase approach:
    # Simpler: use filter-branch on the specific commit range
    pass

# Actually, simplest approach: use git commit --amend in interactive rebase
# Let's use GIT_SEQUENCE_EDITOR to auto-amend each bad commit

print("\n=== Using interactive rebase with exec ===")
# Get the parent of the first bad commit (on fix-temp)
first_bad = bad_commits[-1]  # oldest first
print(f"  First bad: {first_bad[:8]}")

# Use git rebase -i with --exec to amend each commit
r = run(['git', 'log', '--oneline', 'fix-temp', '--format=%H'], check=False)
all_commits = r.stdout.strip().split('\n')
print(f"  Total commits: {len(all_commits)}")

# Find the index of the first bad commit
# We need to rebase from first bad commit's parent
# Actually, use rebase --exec approach

# Simpler: just do a new temp branch
print("\n=== Step A: Create fix-temp2 ===")
run(['git', 'checkout', '-b', 'fix-temp2', '3b39c09726f088df60400d123f89f81ce659b54a'], check=False)

# Get all commits from main that need to be replayed (only the ones from the 6 we're fixing)
# Original 6 commits in chronological order
orig_commits = [
    "4a8f82ea84fc52a7ab957a7768be7f2fa1dacec9",
    "74febc94985d796a25617161ff78a5efe760d435",
    "ec198b0389139578dad1a3e4e32e4993600926e9",
    "b3e9a172777a64b95426c142fa89c59fb6185a43",
    "f5bf9e886097a5efb6851e35d2f4b452328d3aa3",
    "face78ac7467e19dbf30fb9fc2989808b03b25aa",
]

for i, orig_commit in enumerate(orig_commits):
    print(f"\n=== Replay {i+1}: {orig_commit[:8]} ===")
    
    # Get the tree of this commit
    # Use git diff-tree to get files changed
    r = run(['git', 'diff-tree', '--no-commit-id', '-r', '--name-status', orig_commit], check=False)
    if not r.stdout.strip():
        # Empty commit
        r2 = run(['git', 'log', '-1', '--format=%B', orig_commit], check=False)
        msg = r2.stdout.strip() or f"empty commit from {orig_commit[:8]}"
        run(['git', 'commit', '--allow-empty', '--no-verify',
             '--author=makersoft <18072729393@163.com>',
             '-m', msg])
        continue
    
    files_changed = r.stdout.strip().split('\n')
    added_or_modified = []
    deleted = []
    
    for line in files_changed:
        parts = line.split('\t')
        if len(parts) >= 2:
            status = parts[0]
            filepath = parts[-1]
            if status.startswith('D'):
                deleted.append(filepath)
            else:
                added_or_modified.append(filepath)
    
    # Checkout added/modified files from original commit (skip problematic ones)
    for f in added_or_modified:
        if 'app.log' in f or 'server.exe' in f:
            continue
        run(['git', 'checkout', orig_commit, '--', f], check=False)
    
    # Delete files that were deleted in original
    for f in deleted:
        if 'app.log' in f or 'server.exe' in f:
            continue
        run(['git', 'rm', '--cached', '-f', f], check=False)
    
    # Get commit message
    r2 = subprocess.run(['git', 'log', '-1', '--format=%B', orig_commit], 
                       capture_output=True)
    msg = r2.stdout.decode('utf-8', errors='replace').strip()
    if not msg:
        msg = f"commit from {orig_commit[:8]}"
    
    # Stage everything and commit
    run(['git', 'add', '-A'], check=False)
    run(['git', 'commit', '--allow-empty', '--no-verify',
         '--author=makersoft <18072729393@163.com>',
         '-m', msg])

# Verify fix-temp2
print("\n=== Verify fix-temp2 ===")
r = subprocess.run(['git', 'log', '--format=%ae', 'fix-temp2'], capture_output=True)
email_text = r.stdout.decode('utf-8', errors='replace')
if 'makersoft@example.com' in email_text:
    print("ERROR: Still has wrong emails!")
    # Show which ones
    r2 = subprocess.run(['git', 'log', '--format=%H %ae %s', 'fix-temp2'], capture_output=True)
    print(r2.stdout.decode('utf-8', errors='replace'))
    sys.exit(1)

print("  All emails correct!")

# Switch main
print("\n=== Switch main ===")
run(['git', 'checkout', 'main'])
run(['git', 'reset', '--hard', 'fix-temp2'])
run(['git', 'branch', '-D', 'fix-temp'], check=False)
run(['git', 'branch', '-D', 'fix-temp2'], check=False)

# Final verify
print("\n=== Final verify ===")
r = subprocess.run(['git', 'log', '-8', '--format=%H %ae %s'], capture_output=True)
print(r.stdout.decode('utf-8', errors='replace'))
print("\nALL DONE!")
