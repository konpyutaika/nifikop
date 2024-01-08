## Chart Dependencies
To download local copy of each chart's dependencies, go inside the chart dir and run
```shell
helm dependency update
```
Or in the root of repository run the following command to update all charts dependencies:
```shell
make helm-dep-update
```
## Chart Documentation
Versioned chart `README.md` documentation is generated via the following command from the project root:

```
docker run --rm --volume "$(pwd):/helm-docs" -u $(id -u) jnorwood/helm-docs:latest
```

Or just run the following `make` command from root:
```shell
make helm-gen-docs
```

source: https://github.com/norwoodj/helm-docs

## Charts Version Match

All the helm charts present in repository needs to have same version and the pipeline runs the following command to 
check that all charts match the version present in `helm/nifikop/Chart.yaml`:
```shell
make helm-chart-version-match
```
The command exits with 1 if there is any mismatch between versions, otherwise it exits with 0.