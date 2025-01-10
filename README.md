# Reeve CI / CD - Pipeline Step: File Revision

This is a [Reeve](https://github.com/reeveci/reeve) step that generates a revision string for a set of files.

This step creates a revision string for the specified set of files and is intended to be used for detecting changes in configuration files.
The files' mode, ownership and contents are included in the hash.

The step sets a runtime variable that changes whenever the files' contents are updated.
This can be used to automatically redeploy services when the files change.

## Configuration

See the environment variables mentioned in [Dockerfile](Dockerfile).
