# Releasing

## Process

For version v0.x.y:

1. Create the annotated tag
    > NOTE: To use your GPG signature when pushing the tag, use `SIGN_TAG=1 ./contrib/tag-release.sh v0.x.y` instead)
    - `./contrib/tag-release.sh v0.x.y`
1. Push the tag to the GitHub repository. This will automatically trigger a [Github Action](https://github.com/tinkerbell/tink/actions) to create a release.
    > NOTE: `origin` should be the name of the remote pointing to `github.com/tinkerbell/tink`
    - `git push origin v0.x.y`
1. Review the release on GitHub.

### Permissions

Releasing requires a particular set of permissions.

-   Tag push access to the GitHub repository
