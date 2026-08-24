
# Jit

Jit is a Git-like version control system implemented from scratch in Go (Golang).

The main goal of this project is to understand how Git works internally by implementing its core concepts and functionality rather than simply using Git as a tool.

## Why Jit?

Git is a tool that I use every day, but I wanted to understand what is actually happening underneath commands such as add, commit, and checkout.

This project is also an opportunity to strengthen my understanding of:

* Go
* Data structures and algorithms
* File systems
* Concurrency
* Version control systems
* Software design


## Implementation
The current version of Jit includes several of Git's core concepts and internal mechanisms.

### Repository & Object Storage
* Repository initialization
* Content-addressed object storage
* Blob objects
* Tree objects
* Commit objects
* Object hashing and persistence

### Staging Area
* Staging/index implementation
* Tracking changes between the working tree and the index
* File comparison for determining changes

### Diff
* File comparison
* Myers diff algorithm

### References
* Repository reference structure

### Concurrency
* File locking and concurrency control for repository operations


## Future Work

The project will continue to evolve as I explore more of Git's internal architecture and identify additional areas that are useful to implement such as:
* Conflict resolution
* Merging branches
* Supporting more of git's commands

## Learning Approach

One of the main goals of this project is not simply to reproduce an existing implementation.

For each major concept, my approach is:
1. Understand the problem and the underlying concept.
2. Attempt an implementation independently.
3. Compare the implementation with other approaches.
4. Understand the differences and the reasoning behind them.
5. Refactor or improve my implementation where appropriate.

I am using Building Git by James Coglan as one of the main references while developing the project.
