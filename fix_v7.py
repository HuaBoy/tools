import subprocess
import os
import sys

os.chdir(r'E:\tools20260623')
sys.stdout.reconfigure(encoding='utf-8', errors='replace')

def run(cmd, env=None, check=True):
    print(f"  > {' '.join(cmd[:5])}...")
    final_env = {**os.environ, **(env or {})}
    r = subprocess.run(cmd, capture_output=True, text=True, encoding='utf-8', errors='replace', env=final_env)
    if r.stdout.strip():
        for line in r.stdout.strip().split('\n')[:3]:
            print(f"    {line[:150]}")
    if r.stderr.strip():
        for line in r.stderr.strip().split('\n')[:3]:
            print(f"    ERR: {line[:150]}")
    if check and r.returncode != 0:
        raise Exception(f"Command failed: {' '.join(cmd)}")
    return r

# Step 1: Switch to main and delete all temp branches
print("=== Step 1: Cleanup ===")
run(['git', 'checkout', 'main'], check=False)
run(['git', 'branch', '-D', 'fix-temp'], check=False)
run(['git', 'branch', '-D', 'fix-temp2'], check=False)

# Step 2: Create fix-temp2 from correct parent
print("\n=== Step 2: Create fix-temp2 ===")
run(['git', 'checkout', '-b', 'fix-temp2', '3b39c09726f088df60400d123f89f81ce659b54a'])

# Step 3: Cherry-pick each commit with forced author via env vars
orig_commits = [
    "4a8f82ea84fc52a7ab957a7768be7f2fa1dacec9",
    "74febc94985d796a25617161ff78a5efe760d435",
    "ec198b0389139578dad1a3e4e32e4993600926e9",
    "b3e9a172777a64b95426c142fa89c59fb6185a43",
    "f5bf9e886097a5efb6851e35d2f4b452328d3aa3",
    "face78ac7467e19dbf30fb9fc2989808b03b25aa",
]

# Set forced author env vars
forced_env = {
    'GIT_AUTHOR_NAME': 'makersoft',
    'GIT_AUTHOR_EMAIL': '18072729393@163.com',
    'GIT_COMMITTER_NAME': 'makersoft',
    'GIT_COMMITTER_EMAIL': '18072729393@163.com',
}

for i, commit_hash in enumerate(orig_commits):
    print(f"\n=== Step 3.{i+1}: Cherry-pick {commit_hash[:8]} with forced author ===")
    
    # Use git commit-tree to create a new commit with correct author
    # First, get the tree of the original commit
    r = run(['git', 'rev-parse', commit_hash + '^{tree}'], check=False)
    tree_hash = r.stdout.strip()
    
    # Get the parent (which is HEAD after cherry-picking previous commits)
    r = run(['git', 'rev-parse', 'HEAD'], check=False)
    parent_hash = r.stdout.strip()
    
    # Get the commit message
    r = subprocess.run(['git', 'log', '-1', '--format=%B', commit_hash], 
                      capture_output=True)
    msg = r.stdout.decode('utf-8', errors='replace').strip()
    
    # Create new commit using commit-tree with correct author
    env_str = f"AUTHOR_NAME=makersoft\nAUTHOR_EMAIL=18072729393@163.com\nAUTHOR_DATE={subprocess.run(['git', 'log', '-1', '--format=%aD', commit_hash], capture_output=True).stdout.decode('utf-8', errors='replace').strip()}"
    
    # Actually, simpler: use GIT_AUTHOR_NAME/EMAIL env vars with commit-tree
    final_env = {**os.environ, **forced_env}
    
    # Write commit message to temp file
    msg_file = r'E:\tools20260623\.commit_msg'
    with open(msg_file, 'w', encoding='utf-8') as f:
        f.write(msg)
    
    # Use commit-tree to create new commit
    r = subprocess.run(
        ['git', 'commit-tree', tree_hash, '-p', parent_hash, '-F', msg_file],
        capture_output=True, env=final_env
    )
    new_commit = r.stdout.decode('utf-8', errors='replace').strip()
    
    if new_commit:
        # Update HEAD to point to new commit
        subprocess.run(['git', 'reset', '--hard', new_commit], capture_output=True)
        print(f"  Created new commit {new_commit[:8]} with correct author")
    else:
        print(f"  ERROR: commit-tree failed")
        print(f"    stderr: {r.stderr.decode('utf-8', errors='replace')}")
        sys.exit(1)

# Clean up temp file
try:
    os.remove(msg_file)
except:
    pass

# Verify
print("\n=== Step 4: Verify ===")
r = subprocess.run(['git', 'log', '--format=%ae', 'fix-temp2'], capture_output=True)
email_text = r.stdout.decode('utf-8', errors='replace')
if 'makersoft@example.com' in email_text:
    print("ERROR: Still has wrong emails!")
    r2 = subprocess.run(['git', 'log', '--format=%H %ae %s', 'fix-temp2'], capture_output=True)
    print(r2.stdout.decode('utf-8', errors='replace'))
    sys.exit(1)
print("  All emails correct!")

# Switch main
print("\n=== Step 5: Switch main ===")
run(['git', 'checkout', 'main'])
run(['git', 'reset', '--hard', 'fix-temp2'])
run(['git', 'branch', '-D', 'fix-temp2'], check=False)

# Final verify
print("\n=== Final verify ===")
r = subprocess.run(['git', 'log', '-8', '--format=%H %ae %s'], capture_output=True)
print(r.stdout.decode('utf-8', errors='replace'))
print("\nALL DONE!")
