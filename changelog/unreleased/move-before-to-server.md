Change: Move Before hook from root to server command

This is needed because he before hook which initializes the config is not executed correctly in the root cmd if
the service is called from the ocis single-binary.

With this commit the --config-file argument needs to be passed to the server sub-command.

https://github.com/owncloud/ocis/issues/139
