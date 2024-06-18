# Contributing to FLUIDOS Node #

First of all, thank you for your time!

This section provides guidelines and suggestions for the local development of FLUIDOS Node components.

## Local development ##

When developing a new feature, it is beneficial to test changes in a local environment and debug the code to identify potential issues. To facilitate this, you can use the provided setup.sh script for the quick start example. This script helps you create two development clusters using KinD and install FLUIDOS Node along with its dependencies on both clusters.

To get started, follow these steps:

1. Run the `/tools/scripts/setup.sh` script.
    When prompted, choose the type of environment:
     - Select option 1 for one consumer and one provider cluster.
     - Select option 2 for an environment with multiple consumers and providers.
2. Confirm with "yes" when asked if you want to use local repositories.

## How to contribute ##

We welcome contributions to the project! To ensure a smooth process, please follow these steps:

1. ### Open a issue ###

    When reporting a bug, requesting a new feature, or suggesting documentation improvements, please use the appropriate labels:
    - **bug** for reporting a bug
    - **enhancement** for a new feature
    - **documentation** for documentation improvements

2. ### Fork the Repository ###

    Create your own copy of the repository.

3. ### Create a New Branch ###

    It is recommended to create a new branch in your fork to work on the requested modifications.

4. ### Merge Modifications ###

    After making the necessary changes, merge them into your fork's main branch. Ensure you have rebased your branch from the remote upstream repository.

5. ### Test and Resolve Conflicts ###

    Test your changes thoroughly and resolve any potential merge conflicts.

6. ### Open a Pull Request (PR) ###

    Once your modifications are tested and conflicts are resolved, open a PR to the main repository. Remember to squash all your commits during the PR.

Thank you for your contributions!
