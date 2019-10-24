package main

import (
	"flag"
	"log"
	"strings"

	"github.com/go-acme/lego/providers/dns/route53"
	"github.com/mholt/certmagic"
	"github.com/securityclippy/magicstorage"

	"github.com/erdii/headless-certmagic/internal"
)

func main() {
	log.Println("Welcome to headless-certmagic")

	flagDomains := flag.String("domains", "", "comma-separated list of domains")
	flagStaging := flag.Bool("staging", true, "use LE staging environment")
	flagEmail := flag.String("email", "", "email for LE account")
	flagBucketName := flag.String("bucket-name", "", "s3 bucket for storage/locking")
	flagBucketRegion := flag.String("bucket-region", "", "s3 bucket region")
	flagOutPath := flag.String("out-path", "", "ssl output folder")
	flagHook := flag.String("hook", "", "command(s) to run after certificates changed")
	flagHelp := flag.Bool("h", false, "show this help text")
	flag.Parse()

	if *flagHelp || len(*flagEmail) < 1 || len(*flagBucketName) < 1 || len(*flagBucketRegion) < 1 || len(*flagOutPath) < 1 || len(*flagDomains) < 1 {
		flag.PrintDefaults()
		return
	}

	conf := internal.Config{
		DomainNames:  strings.Split(*flagDomains, ","),
		Path:         *flagOutPath,
		Staging:      *flagStaging,
		Email:        *flagEmail,
		BucketName:   *flagBucketName,
		BucketRegion: *flagBucketRegion,
		Hook:         *flagHook,
	}

	log.Printf("conf: %+v", conf)

	certmagic.Default.DisableHTTPChallenge = true
	certmagic.Default.DisableTLSALPNChallenge = true
	certmagic.Default.Agreed = true
	certmagic.Default.MustStaple = true
	certmagic.Default.Email = conf.Email

	if conf.Staging {
		certmagic.Default.CA = certmagic.LetsEncryptStagingCA
	} else {
		certmagic.Default.CA = certmagic.LetsEncryptProductionCA
	}

	provider, err := route53.NewDNSProvider()
	if err != nil {
		log.Panicf("could not initialize dns provider: %s\n", err)
	}

	certmagic.Default.DNSProvider = provider
	certmagic.Default.Storage = magicstorage.NewS3Storage(conf.BucketName, conf.BucketRegion)

	log.Printf("Obtaining certificate(s) from certmagic: %+v\n", conf.DomainNames)

	cfg := certmagic.NewDefault()
	err = cfg.ManageSync(conf.DomainNames)
	if err != nil {
		log.Panicf("certmagic failed: %s\n", err)
	}

	var changed = false

	for _, domainName := range conf.DomainNames {
		cert, err := cfg.CacheManagedCertificate(domainName)
		if err != nil {
			log.Panicf("could not load obtained certificate for domain: %s\n", domainName)
		}

		log.Printf("domain: %s\n", domainName)
		log.Printf("cert domains: %+v\n", cert.Names)
		log.Printf("expires: %s", cert.NotAfter)

		certHash, err := internal.HashCertificates(cert.Certificate.Certificate)
		if err != nil {
			log.Panicf("could not hash cert for domain %s: %s\n", domainName, err)
		}

		oldCertFileHash, err := internal.HashFile(&conf, domainName, true)
		if err != nil {
			log.Panicf("could not hash existing cert file for domain %s: %s\n", domainName, err)
		}

		log.Printf("certHash: %s\n oldCertFileHash: %s\n", certHash, oldCertFileHash)
		if certHash == oldCertFileHash {
			log.Printf("cert did not change, skipping writes for domain: %s\n", domainName)
			continue
		}

		changed = true

		err = internal.ExportCertificateFile(&conf, domainName, &cert)
		if err != nil {
			log.Panicf("could not export certificate for domain %s: %s\n", domainName, err)
		}

		err = internal.ExportKeyFile(&conf, domainName, &cert)
		if err != nil {
			log.Panicf("could not export key for domain %s: %s\n", domainName, err)
		}
	}

	if changed {
		log.Printf("files changed - you should reload your systems manually if you did not set up a renewal hook")

		if len(conf.Hook) > 0 {
			err = internal.ExecCmd(conf.Hook)
			if err != nil {
				log.Panicf("could not run hook command: %s\n", err)
			}
		}
	}
}
