---
id: 1_circle_ci_build_pipeline
title: CircleCI build pipeline
sidebar_label: CircleCI build pipeline
---

We use CircleCI to build and test the operator code on each commit.

## CircleCI config validation hook

To discover errors in the CircleCI earlier, we can uses the [CircleCI cli](https://circleci.com/docs/2.0/local-cli/)
to validate the config file on pre-commit git hook.

Fisrt you must install the cli, then to install the hook, runs:<

```console
cp tools/pre-commit .git/hooks/pre-commit
```

The Pipeline uses some envirenment variables that you need to set-up if you want your fork to build

- DOCKER_REPO_BASE -- name of your docker base reposirory (ex: orangeopensource)
- DOCKERHUB_PASSWORD
- DOCKERHUB_USER
- SONAR_PROJECT
- SONAR_TOKEN

If not set in CircleCI environment, according steps will be ignored.

## CircleCI on PR

When you submit a Pull Request, then CircleCI will trigger build pipeline.
Since this is pushed from a fork, for security reason the pipeline won't have access to the environment secrets, and not all steps could be executed.