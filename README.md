# Go Implementation of some Git commands

A custom, zero-dependency implementation of the some of the core Git version control system commands written entirely in Go. This project rebuilds Git's local object model and network transfer protocols from the ground up to get a deep understanding of git's internals.

## Features
- Create, hash, read, and compress Git blobs, trees, and commits.
- Implements Git's exact SHA-1 cryptographic hashing and zlib compression.
- Full support for `git-upload-pack` remote repository cloning.
- Decompresses raw multiplexed binary streams.
- Reconstructs optimized `REF_DELTA` patch chains into loose base objects.

## Installation & Build
```bash
# Clone the repository
git clone [https://github.com/yourusername/mygit.git](https://github.com/yourusername/mygit.git)
cd mygit

# Build the binary
go build -o git_cmd_impl_in_go ./cmd/mygit

# Call the commnds in test directory
<path_to_binary>/git_cmd_impl_in_go <command> <args>
```

# Supported Commands
- init: Initialize an empty Git repository.
    example: `<path_to_binary>/git_cmd_impl_in_go init`
- cat-file -p <sha>: Read and decompress an object's contents.
    example: `<path_to_binary>/git_cmd_impl_in_go cat-file -p <sha>`
- hash-object -w <file>: Compute a file's SHA-1 hash and save it to .git/objects.
    example: `<path_to_binary>/git_cmd_impl_in_go hash-object -w <file>`
- write-tree: Traverse the working directory and write tree objects.
    example: `<path_to_binary>/git_cmd_impl_in_go write-tree`
- ls-tree --name-only <tree-sha>: List contents of a tree object.
    example: `<path_to_binary>/git_cmd_impl_in_go ls-tree --name-only <tree-sha>`
- commit-tree <tree-sha> -m "Message": Create a commit object.
    example: `<path_to_binary>/git_cmd_impl_in_go commit-tree <tree-sha> -m "Message"`
- clone <url> <directory>: Clone a remote repository over HTTPS.
    example: `<path_to_binary>/git_cmd_impl_in_go clone <url> <directory>`