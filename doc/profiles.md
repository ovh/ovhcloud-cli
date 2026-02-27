# Profiles

Profiles allow you to store credentials for multiple OVHcloud accounts in a single configuration file (`~/.ovh.conf`) and switch between them easily.

> **Retro-compatibility:** If you don't use profiles, the CLI works exactly as before. No changes are required to your existing configuration.

## Creating a profile

Use the `--profile` flag with the `login` command to store credentials under a named profile:

```bash
ovhcloud login --profile work
ovhcloud login --profile personal
```

Each profile is stored as a `[profile:<name>]` section in your `~/.ovh.conf` file:

```ini
[default]
endpoint=ovh-eu
profile=work

[ovh-eu]
application_key=xxx
application_secret=xxx
consumer_key=xxx

[profile:work]
endpoint=ovh-eu
application_key=aaa
application_secret=aaa
consumer_key=aaa

[profile:personal]
endpoint=ovh-eu
application_key=bbb
application_secret=bbb
consumer_key=bbb
```

## The default profile

The `default` profile is a virtual profile that represents the legacy configuration (credentials stored in `[ovh-<region>]` sections). It is always listed and cannot be created or deleted.

## Managing profiles

```bash
# List all profiles (the active one is marked with *)
ovhcloud config profile list

# Show details of the active profile
ovhcloud config profile show

# Show details of a specific profile
ovhcloud config profile show work

# Switch the active profile
ovhcloud config profile switch personal

# Revert to legacy mode (no profile)
ovhcloud config profile switch default

# Delete a profile
ovhcloud config profile delete work
```

## Using a profile for a single command

Use the `--profile` flag on any command to override the active profile for that invocation only:

```bash
ovhcloud cloud project list --profile personal
ovhcloud cloud instance list --profile work
```

This does not change the active profile in the configuration.

## Environment variable

You can also set the active profile via the `OVH_PROFILE` environment variable:

```bash
export OVH_PROFILE=work
ovhcloud cloud project list
```

## Precedence

When resolving which profile to use, the CLI follows this order (first match wins):

1. `--profile` flag
2. `OVH_PROFILE` environment variable
3. `profile` key in `[default]` section of the configuration file
4. Legacy mode (no profile)

## Profile-bound settings

Settings like `default_cloud_project` are scoped to the active profile. Each profile maintains its own value, and there is no cross-profile fallback.

```bash
# Set default_cloud_project for the "work" profile
ovhcloud config profile switch work
ovhcloud config set default_cloud_project xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx

# Set a different one for "personal"
ovhcloud config profile switch personal
ovhcloud config set default_cloud_project yyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyy
```
