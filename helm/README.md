## Chart Documentation
Versioned chart `README.md` documentation is generated via the following command from the project root:

```
docker run --rm --volume "$(pwd):/helm-docs" -u $(id -u) jnorwood/helm-docs:latest
```

Or just run the `./generate_docs.sh` script

source: https://github.com/norwoodj/helm-docs