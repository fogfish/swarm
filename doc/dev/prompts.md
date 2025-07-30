# Prompts for Copilot

Prompts below guides coding agents with the library development and maintanace.

## Update kernel

Your task is to **update dependencies for all brokers** in this Golang library because a new version of the **swarm messaging kernel** (`github.com/fogfish/swarm`) has been released.

Follow these instructions step-by-step **without skipping or altering their order**:

#### **Steps to Complete the Task:**

1. **Create a new Git branch**
   * Name the branch `chore/update-swarm-deps`.
   * Checkout the workspace into this branch.

2. **Identify the latest version of `github.com/fogfish/swarm`**
   * Navigate to the `version.go` file in root of this repository.
   * Extract the latest version string defined there.

3. **Identity all brokers**
   * Brokers are located in the `/broker` directory.
   * Remember these brokers during this session

4. **For each broker module** execute series of instructions defined below
   * Switch current working directory into the broker 

4.1. **Update `go.mod` for each broker**
   * Update **only the version of `github.com/fogfish/swarm`** in its `go.mod` and `go.sum` by running `go get github.com/fogfish/swarm@{set the latest version here}` and then `go mod tidy`
   * Do **not** update or tidy other dependencies.

4.2. **Bump the PATCH version of each broker**

   * In each broker’s `version.go` file, increment the PATCH segment (e.g., `v1.2.3` → `v1.2.4`).

4.3. **Run unit tests for each broker**
   * Execute `go test ./...` in each broker’s directory.
   * Ensure all tests pass before proceeding.

4.4 **Switch current directory** to the repository root
   * Switch current working directory into the repository root

5. **Create a Pull Request (PR)**
   * Commit and push only the files changed by your during this session.
   * Open a PR titled:

     > `chore: update swarm kernel dependency for all brokers`
   * Include a brief description summarizing the update.
   * Upstream origin is `github`, use `git push github ...` to push it. 

#### **Constraints:**

* Only update the `github.com/fogfish/swarm` dependency.
* Do not modify unrelated dependencies or code.
* Ensure all brokers build and pass tests before creating the PR.
* Commit only files changed by this workflow.
* Watch out for the working directory, command might fail if your are executing it in the wrong context

