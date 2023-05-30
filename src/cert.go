// Copyright 2018 The mkcert Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package misc

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"net"
	"net/mail"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var userAndHostname string

//	func init() {
//		u, err := user.Current()
//		if err == nil {
//			userAndHostname = u.Username + "@"
//		}
//		if h, err := os.Hostname(); err == nil {
//			userAndHostname += h
//		}
//		if err == nil && u.Name != "" && u.Name != u.Username {
//			userAndHostname += " (" + u.Name + ")"
//		}
//	}
func FatalIfErr(err error, msg string) {
	if err != nil {
		log.Fatalf("ERROR: %s: %s", msg, err)
	}
}

const rootName = "rootCA.pem"
const rootKeyName = "rootCA-key.pem"

type Mkcert struct {
	caCert *x509.Certificate
	caKey  crypto.PrivateKey
	// The system cert pool is only loaded once. After installing the root, checks
	// will keep failing until the next execution. TODO: maybe execve?
	// https://github.com/golang/go/issues/24540 (thanks, myself)
	ignoreCheckFailure bool
}

var CAROOT string

func getCAROOT() string {
	if env := os.Getenv("CAROOT"); env != "" {
		return env
	}
	if !IsMobile {
		AppDirectory = filepath.Join(os.Getenv("LocalAppData"), "sillydokkan")
	}
	return AppDirectory
}

//func storeEnabled(name string) bool {
//	stores := os.Getenv("TRUST_STORES")
//	if stores == "" {
//		return true
//	}
//	for _, store := range strings.Split(stores, ",") {
//		if store == name {
//			return true
//		}
//	}
//	return false
//}

func (m *Mkcert) Run() {
	fmt.Println("HIIIII FROM GOLANG CERT")
	CAROOT = getCAROOT()
	if CAROOT == "" {
		log.Fatalln("ERROR: failed to find the default CA location, set one as the CAROOT env var")
	}
	FatalIfErr(os.MkdirAll(CAROOT, 0755), "failed to create the CAROOT")
	m.loadCA()
	if !IsMobile {
		//m.install()
	}
	fake := []string{"localhost", GetLocalIP(), "::1", "127.0.0.1"}
	m.makeCert(fake)
}

//func (m *Mkcert) install() {
//	if storeEnabled("system") {
//		if m.checkPlatform() {
//			//log.Print("The local CA is already installed in the system trust store! 👍")
//		} else {
//			if truststore.InstallPlatform() {
//				log.Print("The local CA is now installed in the system trust store! ⚡️")
//			}
//			m.ignoreCheckFailure = true // TODO: replace with a check for a successful install
//		}
//	}
//	log.Print("")
//}

//func (m *Mkcert) uninstall() {
//	if storeEnabled("system") && uninstallPlatform() {
//		log.Print("The local CA is now uninstalled from the system trust store(s)! 👋")
//		log.Print("")
//	}
//}

func (m *Mkcert) checkPlatform() bool {
	if m.ignoreCheckFailure {
		return true
	}

	_, err := m.caCert.Verify(x509.VerifyOptions{})
	return err == nil
}

func (m *Mkcert) makeCert(hosts []string) {
	if m.caKey == nil {
		log.Fatalln("ERROR: can't create new certificates because the CA key (rootCA-key.pem) is missing")
	}

	priv, err := m.generateKey(false)
	FatalIfErr(err, "failed to generate certificate key")
	pub := priv.(crypto.Signer).Public()

	// Certificates last for 2 years and 3 months, which is always less than
	// 825 days, the limit that macOS/iOS apply to all certificates,
	// including custom roots. See https://support.apple.com/en-us/HT210176.

	IssuedOn := time.Now().AddDate(0, 0, -5)
	expiration := time.Now().AddDate(2, 3, 0)
	//fmt.Println(IssuedOn)
	tpl := &x509.Certificate{
		SerialNumber: randomSerialNumber(),
		Subject: pkix.Name{
			Organization:       []string{"mkcert development certificate"},
			OrganizationalUnit: []string{userAndHostname},
		},

		NotBefore: IssuedOn, NotAfter: expiration,

		KeyUsage: x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
	}

	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			tpl.IPAddresses = append(tpl.IPAddresses, ip)
		} else if email, err := mail.ParseAddress(h); err == nil && email.Address == h {
			tpl.EmailAddresses = append(tpl.EmailAddresses, h)
		} else if uriName, err := url.Parse(h); err == nil && uriName.Scheme != "" && uriName.Host != "" {
			tpl.URIs = append(tpl.URIs, uriName)
		} else {
			tpl.DNSNames = append(tpl.DNSNames, h)
		}
	}

	if len(tpl.IPAddresses) > 0 || len(tpl.DNSNames) > 0 || len(tpl.URIs) > 0 {
		tpl.ExtKeyUsage = append(tpl.ExtKeyUsage, x509.ExtKeyUsageServerAuth)
	}
	if len(tpl.EmailAddresses) > 0 {
		tpl.ExtKeyUsage = append(tpl.ExtKeyUsage, x509.ExtKeyUsageEmailProtection)
	}

	// IIS (the main target of PKCS #12 files), only shows the deprecated
	// Common Name in the UI. See issue #115.

	cert, err := x509.CreateCertificate(rand.Reader, tpl, m.caCert, pub, m.caKey)
	FatalIfErr(err, "failed to generate certificate")

	certFile, keyFile, _ := m.fileNames(hosts)

	certFile = filepath.Join(AppDirectory, "server.crt")
	keyFile = filepath.Join(AppDirectory, "server.key")
	//println(certFile, keyFile, p12File)
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert})
	privDER, err2 := x509.MarshalPKCS8PrivateKey(priv)
	FatalIfErr(err2, "failed to encode certificate key")
	privPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privDER})
	if certFile == keyFile {
		err2 = ioutil.WriteFile(keyFile, append(certPEM, privPEM...), 0600)
		FatalIfErr(err2, "failed to save certificate and key")
	} else {
		err2 = ioutil.WriteFile(certFile, certPEM, 0644)
		FatalIfErr(err2, "failed to save certificate")
		err2 = ioutil.WriteFile(keyFile, privPEM, 0600)
		FatalIfErr(err2, "failed to save certificate key")
	}

	//m.printHosts(hosts)

	if certFile == keyFile {
		//log.Printf("\nThe certificate and key are at \"%s\" ✅\n\n", certFile)
	} else {
		//log.Printf("\nThe certificate is at \"%s\" and the key at \"%s\" ✅\n\n", certFile, keyFile)
	}

	//log.Printf("It will expire on %s 🗓\n\n", expiration.Format("2 January 2006"))
}

//func (m *Mkcert) printHosts(hosts []string) {
//	secondLvlWildcardRegexp := regexp.MustCompile(`(?i)^\*\.[\da-z_-]+$`)
//	log.Printf("\nCreated a new certificate valid for the following names 📜")
//	for _, h := range hosts {
//		log.Printf(" - %q", h)
//		if secondLvlWildcardRegexp.MatchString(h) {
//			log.Printf("   Warning: many browsers don't support second-level wildcards like %q ⚠️", h)
//		}
//	}
//
//	for _, h := range hosts {
//		if strings.HasPrefix(h, "*.") {
//			log.Printf("\nReminder: X.509 wildcards only go one level deep, so this won't match a.b.%s ℹ️", h[2:])
//			break
//		}
//	}
//}

func (m *Mkcert) generateKey(rootCA bool) (crypto.PrivateKey, error) {

	if rootCA {
		return rsa.GenerateKey(rand.Reader, 3072)
	}
	return rsa.GenerateKey(rand.Reader, 2048)
}

func (m *Mkcert) fileNames(hosts []string) (certFile, keyFile, p12File string) {
	defaultName := strings.Replace(hosts[0], ":", "_", -1)
	defaultName = strings.Replace(defaultName, "*", "_wildcard", -1)
	if len(hosts) > 1 {
		defaultName += "+" + strconv.Itoa(len(hosts)-1)
	}

	certFile = "./" + defaultName + ".pem"

	keyFile = "./" + defaultName + "-key.pem"

	p12File = "./" + defaultName + ".p12"

	return
}

func randomSerialNumber() *big.Int {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	FatalIfErr(err, "failed to generate serial number")
	return serialNumber
}

//func (m *Mkcert) makeCertFromCSR() {
//	if m.caKey == nil {
//		log.Fatalln("ERROR: can't create new certificates because the CA key (rootCA-key.pem) is missing")
//	}
//
//	csrPEMBytes, err := ioutil.ReadFile(m.CsrPath)
//	fatalIfErr(err, "failed to read the CSR")
//	csrPEM, _ := pem.Decode(csrPEMBytes)
//	if csrPEM == nil {
//		log.Fatalln("ERROR: failed to read the CSR: unexpected content")
//	}
//	if csrPEM.Type != "CERTIFICATE REQUEST" &&
//		csrPEM.Type != "NEW CERTIFICATE REQUEST" {
//		log.Fatalln("ERROR: failed to read the CSR: expected CERTIFICATE REQUEST, got " + csrPEM.Type)
//	}
//	csr, err := x509.ParseCertificateRequest(csrPEM.Bytes)
//	fatalIfErr(err, "failed to parse the CSR")
//	fatalIfErr(csr.CheckSignature(), "invalid CSR signature")
//
//	IssuedOn := time.Now().AddDate(4, 0, 5)
//	expiration := time.Now().AddDate(2, 3, 0)
//	tpl := &x509.Certificate{
//		SerialNumber:    randomSerialNumber(),
//		Subject:         csr.Subject,
//		ExtraExtensions: csr.Extensions, // includes requested SANs, KUs and EKUs
//
//		NotBefore: IssuedOn, NotAfter: expiration,
//
//		// If the CSR does not request a SAN extension, fix it up for them as
//		// the Common Name field does not work in modern browsers. Otherwise,
//		// this will get overridden.
//		DNSNames: []string{csr.Subject.CommonName},
//
//		// Likewise, if the CSR does not set KUs and EKUs, fix it up as Apple
//		// platforms require serverAuth for TLS.
//		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
//		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
//	}
//
//	if MClient {
//		tpl.ExtKeyUsage = append(tpl.ExtKeyUsage, x509.ExtKeyUsageClientAuth)
//	}
//	if len(csr.EmailAddresses) > 0 {
//		tpl.ExtKeyUsage = append(tpl.ExtKeyUsage, x509.ExtKeyUsageEmailProtection)
//	}
//
//	cert, err := x509.CreateCertificate(rand.Reader, tpl, m.caCert, csr.PublicKey, m.caKey)
//	fatalIfErr(err, "failed to generate certificate")
//	c, err := x509.ParseCertificate(cert)
//	fatalIfErr(err, "failed to parse generated certificate")
//
//	var hosts []string
//	hosts = append(hosts, c.DNSNames...)
//	hosts = append(hosts, c.EmailAddresses...)
//	for _, ip := range c.IPAddresses {
//		hosts = append(hosts, ip.String())
//	}
//	for _, uri := range c.URIs {
//		hosts = append(hosts, uri.String())
//	}
//	certFile, _, _ := m.fileNames(hosts)
//
//	err = ioutil.WriteFile(certFile, pem.EncodeToMemory(
//		&pem.Block{Type: "CERTIFICATE", Bytes: cert}), 0644)
//	fatalIfErr(err, "failed to save certificate")
//
//	m.printHosts(hosts)
//
//	log.Printf("\nThe certificate is at \"%s\" ✅\n\n", certFile)
//
//	log.Printf("It will expire on %s 🗓\n\n", expiration.Format("2 January 2006"))
//}

// loadCA will load or create the CA at CAROOT.
func (m *Mkcert) loadCA() {
	//fmt.Println(getCAROOT())
	if !PathExists(filepath.Join(CAROOT, rootName)) {
		m.newCA()
	}
	certPEMBlock, err := ioutil.ReadFile(filepath.Join(CAROOT, rootName))
	FatalIfErr(err, "failed to read the CA certificate")
	certDERBlock, _ := pem.Decode(certPEMBlock)
	if certDERBlock == nil || certDERBlock.Type != "CERTIFICATE" {
		log.Fatalln("ERROR: failed to read the CA certificate: unexpected content")
	}
	m.caCert, err = x509.ParseCertificate(certDERBlock.Bytes)
	FatalIfErr(err, "failed to parse the CA certificate")

	if !PathExists(filepath.Join(CAROOT, rootKeyName)) {
		println("asd")
		return // keyless mode, where only -install works
	}
	keyPEMBlock, err := ioutil.ReadFile(filepath.Join(CAROOT, rootKeyName))
	FatalIfErr(err, "failed to read the CA key")
	keyDERBlock, _ := pem.Decode(keyPEMBlock)
	if keyDERBlock == nil || keyDERBlock.Type != "PRIVATE KEY" {
		log.Fatalln("ERROR: failed to read the CA key: unexpected content")
	}
	m.caKey, err = x509.ParsePKCS8PrivateKey(keyDERBlock.Bytes)
	FatalIfErr(err, "failed to parse the CA key")
}

func (m *Mkcert) newCA() {
	priv, err := m.generateKey(true)
	FatalIfErr(err, "failed to generate the CA key")
	pub := priv.(crypto.Signer).Public()

	spkiASN1, err := x509.MarshalPKIXPublicKey(pub)
	FatalIfErr(err, "failed to encode public key")

	var spki struct {
		Algorithm        pkix.AlgorithmIdentifier
		SubjectPublicKey asn1.BitString
	}
	_, err = asn1.Unmarshal(spkiASN1, &spki)
	FatalIfErr(err, "failed to decode public key")

	skid := sha1.Sum(spki.SubjectPublicKey.Bytes)

	tpl := &x509.Certificate{
		SerialNumber: randomSerialNumber(),
		Subject: pkix.Name{
			Organization:       []string{"mkcert development CA"},
			OrganizationalUnit: []string{userAndHostname},

			// The CommonName is required by iOS to show the certificate in the
			// "Certificate Trust Settings" menu.
			// https://github.com/FiloSottile/mkcert/issues/47
			CommonName: "mkcert " + userAndHostname,
		},
		SubjectKeyId: skid[:],

		NotAfter:  time.Now().AddDate(10, 0, 0),
		NotBefore: time.Now(),

		KeyUsage: x509.KeyUsageCertSign,

		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLenZero:        true,
	}

	cert, err := x509.CreateCertificate(rand.Reader, tpl, tpl, pub, priv)
	FatalIfErr(err, "failed to generate CA certificate")

	privDER, err := x509.MarshalPKCS8PrivateKey(priv)
	FatalIfErr(err, "failed to encode CA key")
	err = ioutil.WriteFile(filepath.Join(CAROOT, rootKeyName), pem.EncodeToMemory(
		&pem.Block{Type: "PRIVATE KEY", Bytes: privDER}), 0400)
	FatalIfErr(err, "failed to save CA key")

	err = ioutil.WriteFile(filepath.Join(CAROOT, rootName), pem.EncodeToMemory(
		&pem.Block{Type: "CERTIFICATE", Bytes: cert}), 0644)
	FatalIfErr(err, "failed to save CA certificate")

	log.Printf("Created a new local CA 💥\n")
}

func (m *Mkcert) caUniqueName() string {
	return "mkcert development CA " + m.caCert.SerialNumber.String()
}
