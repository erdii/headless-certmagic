# headless-certmagic
Idea: Use [mholt/certmagic](https://github.com/mholt/certmagic) like [go-acme/lego](https://github.com/go-acme/lego),
but leverage certmagic's storage backends to provision certificates on fleets of webservers.

### Features
* use LetsEncrypt's DNS challenge to obtain certificates for one or more domains (one domain per cert)
* sync certificates between multiple machines via a storage backend
* renewal hooks are executed if certificate changes
* flag help: `headless-certmagic -h`
* dns provider support:
	* route53: supply aws credentials like usual (envvars, ~/.aws, ...)
	* uses [go-acme/lego](https://github.com/go-acme/lego) and its providers, so it should be possible to support them all
* storage provider support:
	* s3 ([securityclippy/magicstorage](https://github.com/securityclippy/magicstorage)): supply aws credentials like usual (envvars, ~/.aws, ...)
	* uses [mholt/certmagic](https://github.com/mholt/certmagic) and its storage backends
* forced certificate renewal

### Requirements (for now)
* private s3 bucket
* hosted domain/zone in route53
* aws credentials with rw-access to the bucket and rw-access to the zone, they can be supplied via many ways, eg:
	* aws instanceroles
	* aws-cli: `aws configure`
	* envvars: AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY

### Usage
* set up cronjob:
	```bash
	cat > /etc/cron.daily/headless-certmagic.sh <<EOF
	#!/bin/sh -e
	# run headless-certmagic to ensure our certificates are freshhhhh

	BUCKET_REGION="<your-s3-bucket-region"
	BUCKET_NAME="<your-s3-bucket-name>"
	EMAIL="<your-email-address>"
	HOOK="echo 'lol done - this is only executed it the certificate files were rewritten'"
	OUT_PATH="/etc/certmagic-headless"
	DOMAINS="test.example.com,*.test.example.com"
	STAGING="true" # turn off with "false" to use letsencrypt production environment

	/opt/bin/headless-certmagic -bucket-name="$BUCKET_NAME" \
								-bucket-region="$BUCKET_REGION" \
								-email="$EMAIL" \
								-hook="$HOOK" \
								-out-path="$OUT_PATH" \
								-staging="$STAGING" \
								-domains="$DOMAINS"
	EOF

	# create certificate folder
	mkdir /etc/headless-certmagic

	# fix permissions
	chmod 0700 /etc/cron.daily/headless-certmagic.sh
	chmod 0700 /etc/headless-certmagic

	# first run - this could take longer
	/etc/cron.daily/headless-certmagic.sh
	```

### Todo

If this gains traction:
* remove certmagic dependency (maybe we can still use the same storage-interface and logics?) to be able to provide SAN certs with multiple domains
* implement all teh dns providers!
