package main

import (
	"crypto/rand"
	"crypto/rsa"
	"flag"
	"log"

	"github.com/kelseyhightower/kubernetes-letsencrypt-tutorial/kube-cert-manager/provider/dns/googlecloud"
	"github.com/xenolf/lego/acme"
)

var (
	email          string
	domain         string
	project        string
	serviceAccount string
)

func main() {
	flag.StringVar(&email, "email", "", "User email address.")
	flag.StringVar(&domain, "domain", "", "Domain to request certs for.")
	flag.StringVar(&project, "project", "", "Google Cloud project name.")
	flag.StringVar(&serviceAccount, "service-account", "/etc/googlecloud/service-account.json", "Google Cloud service account path.")
	flag.Parse()

	const rsaKeySize = 2048
	privateKey, err := rsa.GenerateKey(rand.Reader, rsaKeySize)
	if err != nil {
		log.Fatal(err)
	}
	user := User{
		Email: email,
		key:   privateKey,
	}
	client, err := acme.NewClient("https://acme-staging.api.letsencrypt.org/directory",
		&user, acme.RSA2048)
	if err != nil {
		log.Fatal(err)
	}

	provider, err := googlecloud.NewDNSProvider(project, serviceAccount)
	if err != nil {
		log.Fatal(err)
	}

	client.SetChallengeProvider(acme.DNS01, provider)
	client.ExcludeChallenges([]acme.Challenge{
		acme.HTTP01,
		acme.TLSSNI01,
	})

	reg, err := client.Register()
	if err != nil {
		log.Fatal(err)
	}
	user.Registration = reg

	err = client.AgreeToTOS()
	if err != nil {
		log.Fatal(err)
	}
	bundle := false
	certificates, failures := client.ObtainCertificate([]string{domain}, bundle, nil)
	if len(failures) > 0 {
		log.Fatal(failures)
	}

	err = createSecret(domain, certificates.Certificate, certificates.PrivateKey)
	if err != nil {
		log.Fatal(err)
	}
}
