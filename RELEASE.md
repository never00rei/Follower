# Release Process

Follower uses release branches for maintenance lines and Git tags for immutable shipped versions.

## Branches

- `main` is the active development branch.
- `feat/*`, `fix/*`, `chore/*`, and similar branches are short-lived work branches.
- `release/vX.Y` branches are long-lived maintenance branches for a minor version line.

## Conventional Commits

Pull request titles must use Conventional Commit format because GitHub squash merges can use the PR title as the commit message on `main`.

Examples:

```text
feat(sync): push prepared checkpoints
fix(jira): handle missing issue fields
chore(build): update release workflow
feat(session)!: change checkpoint storage format
```

Prefer Jira issue IDs in the body instead of the scope:

```text
Refs: DEV-7
```

## Automation

- CI runs on `main`, `release/**`, and pull requests into those branches.
- PR title linting enforces Conventional Commit titles.
- Pushing a tag matching `v*` runs GoReleaser and creates a GitHub release with checksums and binaries.
- The built binary embeds the release version, commit, and build date. Check it with `follower version`.

## Cutting A Release

Create a release branch when starting a version line:

```sh
git checkout main
git pull
git checkout -b release/v0.1
git push origin release/v0.1
```

Tag the exact commit to ship:

```sh
git tag -a v0.1.0 -m "Release v0.1.0"
git push origin v0.1.0
```

Patch an existing version line:

```sh
git checkout release/v0.1
git cherry-pick <fix-commit>
git tag -a v0.1.1 -m "Release v0.1.1"
git push origin release/v0.1 v0.1.1
```

Use semantic versioning for tags:

- `feat:` normally means minor version bump.
- `fix:` normally means patch version bump.
- `!` or `BREAKING CHANGE:` means major version bump.
