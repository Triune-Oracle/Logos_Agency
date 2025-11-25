'''# Release Checklist

This checklist must be completed before any major or minor version release.

## 1. Code Quality & Testing

- [ ] All new features are covered by unit tests.
- [ ] All existing tests pass (`npm test` or equivalent).
- [ ] Code coverage meets the minimum threshold (e.g., 80%).
- [ ] Code is reviewed and approved by at least one other maintainer.
- [ ] Static analysis (linters, formatters) runs clean.
- [ ] Dependency audit is complete (no known high-severity vulnerabilities).

## 2. Documentation

- [ ] `README.md` is up-to-date with new features and installation instructions.
- [ ] `docs/ROADMAP.md` is updated with the next set of milestones.
- [ ] `docs/ARCHITECTURE.md` reflects any major system changes.
- [ ] `docs/PERFORMANCE.md` includes new benchmarks if applicable.
- [ ] `docs/MONETIZATION.md` reflects any changes to the revenue model.
- [ ] Changelog/Release Notes are drafted.

## 3. Infrastructure & Deployment

- [ ] Build process is successful on the CI/CD pipeline.
- [ ] Staging environment deployment is successful.
- [ ] Final smoke tests are run on the staging environment.
- [ ] Database migrations are tested and verified.
- [ ] Rollback plan is documented and tested.

## 4. Final Steps

- [ ] Version number is updated in all relevant files (e.g., `package.json`, `go.mod`).
- [ ] Final commit is tagged with the new version (e.g., `git tag v1.0.0`).
- [ ] Release is published on GitHub.
- [ ] Announcement is prepared for all relevant channels (social media, community).'''
