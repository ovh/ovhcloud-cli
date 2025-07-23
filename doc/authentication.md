## Authenticating the CLI

OVHcloud CLI requires authentication to be able to make API calls.
There are three ways to define your credentials:
- Using a configuration file
- Using environment variables
- Running the `ovhcloud login` command

The following sections explain how to define the authentication parameters and the
available authentication flows.

### Configuration file

The CLI will successively attempt to locate a configuration file in:

1. Current working directory: ``./ovh.conf``
2. Current user's home directory: ``~/.ovh.conf``
3. System wide configuration: ``/etc/ovh.conf``

It is an `ini` file having the following structure:

```ini
[default]
; general configuration: default endpoint
endpoint=ovh-eu

[ovh-eu]
; configuration specific to 'ovh-eu' endpoint
; see the following section of the documentation to check the available keys
client_id=my_client_id
client_secret=my_client_secret

[ovh-cli]
default_cloud_project=my_cloud_project
```

### Authentication means

`ovhcloud` supports two forms of authentication:
- Application key, application secret & consumer key
- OAuth2, using scoped service accounts

#### Application Key/Application Secret/Consumer Key

This is the authentication mean that can be defined interactively using command `ovhcloud login`. Once
you visited the credentials creation page and filled the fields in the CLI, the credentials will be saved in
your configuration file.

Alternatively, you can define the credentials manually.
The CLI will first look for `OVH_ENDPOINT`, `OVH_APPLICATION_KEY`, `OVH_APPLICATION_SECRET` and
`OVH_CONSUMER_KEY` environment variables. If some of these parameters are not
provided, it will look for a configuration file of the form:

```ini
[default]
endpoint=ovh-eu

[ovh-eu]
application_key=my_app_key
application_secret=my_application_secret
consumer_key=my_consumer_key
```

Depending on the API you want to use, you may set the `endpoint` to:

* `ovh-eu` for OVHcloud Europe API
* `ovh-us` for OVHcloud US API
* `ovh-ca` for OVHcloud Canada API
* `soyoustart-eu` for So you Start Europe API
* `soyoustart-ca` for So you Start Canada API
* `kimsufi-eu` for Kimsufi Europe API
* `kimsufi-ca` for Kimsufi Canada API
* Or any arbitrary URL to use in a test for example

#### OAuth2

First, you need to generate a pair of valid `client_id` and `client_secret`: you
can proceed by [following this documentation](https://help.ovhcloud.com/csm/en-manage-service-account?id=kb_article_view&sysparm_article=KB0059343).

Once you have retrieved your `client_id` and `client_secret`, you can create or edit the configuration file of the CLI:

```ini
[default]
endpoint=ovh-eu

[ovh-eu]
client_id=my_client_id
client_secret=my_client_secret
```

It is also possible to use the environment variables `OVH_ENDPOINT`, `OVH_CLIENT_ID` and `OVH_CLIENT_SECRET`.

Depending on the API you want to use, you may set the `endpoint` to:

* `ovh-eu` for OVHcloud Europe API
* `ovh-us` for OVHcloud US API
* `ovh-ca` for OVHcloud Canada API

### Region/company limitations

~> **WARNING**: some products are not available for `soyoustart` and `kimsufi`, or for some endpoints. If you try to use a product that is not available, you will encounter the following error: `Client::NotFound: "Got an invalid (or empty) URL"`.