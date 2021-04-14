# GitHub Workflows

## Using the `setup-terraform` action

By default, the [`setup-terraform` action](https://github.com/hashicorp/setup-terraform) adds a wrapper for the `terraform` command that allows passing results to subsequent steps. This will prevent using the output of a `terraform` command as the input to another command in the same step.

The wrapper can be turned off by using

```yaml
steps:
- uses: hashicorp/setup-terraform@v1
  with:
    terraform_wrapper: false
```

## Testing workflows locally

The tool [`act`](https://github.com/nektos/act) can be used to test GitHub workflows locally. The default container [intentionally does not have feature parity](https://github.com/nektos/act#default-runners-are-intentionally-incomplete) with the containers used in GitHub due to the size of a full container.

The file `./actrc` configures `act` to use a fully-featured container.
