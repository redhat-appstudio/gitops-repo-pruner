# gitops-repo-pruner

## Usage:

The tool requires that the `GITHUB_TOKEN` environment variable be exported before using it. The token that you use, must have sufficient permissions to delete repositories from the specified org. 

To delete invalid repositories in the GitOps org (up to 1000 at a time) run:
```
./gitops-repo-gc --operation delete-invalid
```

To delete repositories by a given keyword run:
```
./gitops-repo-gc --operation delete-by-keyword --keyword some-keyword
```

## Build

To build, run:

```
make build
```

## Deploy

To deploy the CronJob for the Pruner, run:

```
make deploy
```