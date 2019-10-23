# headless-certmagic
Idea: Use [mholt/certmagic](https://github.com/mholt/certmagic) like [go-acme/lego](https://github.com/go-acme/lego),
but leverage certmagic's storage backends to provision certificates on fleets of webservers.

### Features
* use LetsEncrypt's DNS challenge to obtain certificates for one or more domains (one domain per cert)
* abuse [mholt/certmagic](https://github.com/mholt/certmagic) and [securityclippy/magicstorage](https://github.com/securityclippy/magicstorage) to sync certificates between multiple machines via s3
* renewal hooks are executed if certificate changes
* `headless-certmagic -h`
* dns provider support:
	* route53: supply aws credentials like usual (envvars, ~/.aws, ...)
