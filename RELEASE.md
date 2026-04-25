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
- Release Please opens or updates a release PR on pushes to `main` and `release/**`.
- Merging the release PR creates the next version tag from Conventional Commits.
- GoReleaser publishes checksums and binaries when Release Please creates a release.
- Pushing a tag matching `v*` still runs GoReleaser for manual release recovery.
- The built binary embeds the release version, commit, and build date. Check it with `follower version`.

## Version Bumps

Release Please calculates the next version from Conventional Commits:

- `fix:` creates a patch release, such as `v0.1.0` to `v0.1.1`.
- `feat:` creates a minor release, such as `v0.1.1` to `v0.2.0`.
- `!` or `BREAKING CHANGE:` creates a major release, such as `v1.2.3` to `v2.0.0`.
- `chore:`, `docs:`, `ci:`, and `test:` do not create releases by themselves.

The current bootstrap version is tracked in `.release-please-manifest.json`. Release Please updates that manifest and `CHANGELOG.md` in its release PRs.

## Release Flow

Normal release flow:

```sh
git checkout main
git pull
```

Merge work into `main` using Conventional Commit PR titles. Release Please will open or update a release PR. Review and merge that release PR when you want to ship. After it merges, Release Please creates the version tag and GoReleaser publishes the binaries.

For maintenance releases, cherry-pick fixes onto the relevant `release/vX.Y` branch and push that branch. Release Please will open or update a release PR against that same release branch.

## Manual Release Recovery

Create a release branch when starting a version line:

```sh
git checkout main
git pull
git checkout -b release/v0.1
git push origin release/v0.1
```

Tag the exact commit to ship manually only if the automated release flow is unavailable:

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
