# ownCloud Infinite Scale: Bridge

[![Build Status](https://cloud.drone.io/api/badges/owncloud/ocis-bridge/status.svg)](https://cloud.drone.io/owncloud/ocis-bridge)
[![Gitter chat](https://badges.gitter.im/cs3org/reva.svg)](https://gitter.im/cs3org/reva)
[![Codacy Badge](https://api.codacy.com/project/badge/Grade/d005a4722c1b463b9b95060479018e99)](https://www.codacy.com/gh/owncloud/ocis-bridge?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=owncloud/ocis-bridge&amp;utm_campaign=Badge_Grade)
[![Go Doc](https://godoc.org/github.com/owncloud/ocis-bridge?status.svg)](http://godoc.org/github.com/owncloud/ocis-bridge)
[![Go Report](http://goreportcard.com/badge/github.com/owncloud/ocis-bridge)](http://goreportcard.com/report/github.com/owncloud/ocis-bridge)
[![](https://images.microbadger.com/badges/image/owncloud/ocis-bridge.svg)](http://microbadger.com/images/owncloud/ocis-bridge "Get your own image badge on microbadger.com")

**This project is under heavy development, it's not in a working state yet!**

## Install

You can download prebuilt binaries from the GitHub releases or from our [download mirrors](http://download.owncloud.com/ocis/bridge/). For instructions how to install this on your platform you should take a look at our [documentation](https://owncloud.github.io/ocis-bridge/)
****
## Development

Make sure you have a working Go environment, for further reference or a guide take a look at the [install instructions](http://golang.org/doc/install.html). This project requires Go >= v1.13.

```console
git clone https://github.com/owncloud/ocis-bridge.git
cd ocis-bridge

make generate build

./bin/ocis-bridge -h
```

## Security

If you find a security issue please contact security@owncloud.com first.

## Contributing

Fork -> Patch -> Push -> Pull Request

## License

Apache-2.0

## Copyright

```console
Copyright (c) 2019 ownCloud GmbH <https://owncloud.com>
```
