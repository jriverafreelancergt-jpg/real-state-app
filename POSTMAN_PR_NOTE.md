Postman artifacts PR helper

This file was added to create a trivial diff on branch `feat/postman-qa` so a Pull Request can be opened.

Changes in this branch relate to adding Postman collections, environment, a happy-path suite, CI examples and QA docs. Please review the following files:

- real-state-backend/real-state.postman_collection.json
- real-state-backend/real-state.postman_environment.json
- real-state-backend/real-state.happy_path.postman_collection.json
- real-state-backend/README.postman.md
- real-state-backend/QA_CHECKLIST.md
- real-state-backend/.github/workflows/postman-newman.yml
- real-state-backend/.gitlab-ci.yml
- real-state-backend/scripts/run-newman.sh

If this PR should not include this helper file, remove `POSTMAN_PR_NOTE.md` in the PR review.
