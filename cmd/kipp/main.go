package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"log"
	"math/big"
	"mime"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/6f7262/kipp"
	"github.com/alecthomas/units"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

type worker time.Duration

func (w worker) Do(ctx context.Context, f func() error) {
loop:
	if err := f(); err != nil {
		log.Fatal(err)
	}
	t := time.After(time.Duration(w))
	select {
	case <-ctx.Done():
		return
	case <-t:
		goto loop
	}
}

func loadMimeTypes(path string) error {
	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	m := make(map[string][]string)
	if err := json.NewDecoder(f).Decode(&m); err != nil {
		return err
	}
	for k, v := range m {
		for _, vv := range v {
			mime.AddExtensionType(vv, k)
		}
	}
	return nil
}

func CertificateGetter() func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	var cached *tls.Certificate
	return func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
		if cached != nil {
			return cached, nil
		}
		// Generate a self-signed certificate
		sn, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
		if err != nil {
			return nil, err
		}
		now := time.Now()
		t := &x509.Certificate{
			SerialNumber:          sn,
			NotBefore:             now,
			NotAfter:              now,
			KeyUsage:              x509.KeyUsageCertSign,
			ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
			BasicConstraintsValid: true,
			IsCA: true,
		}
		k, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return nil, err
		}
		c, err := x509.CreateCertificate(rand.Reader, t, t, &k.PublicKey, k)
		if err != nil {
			return nil, err
		}
		cert, err := tls.X509KeyPair(
			pem.EncodeToMemory(&pem.Block{
				Type:  "CERTIFICATE",
				Bytes: c,
			}),
			pem.EncodeToMemory(&pem.Block{
				Type:  "RSA PRIVATE KEY",
				Bytes: x509.MarshalPKCS1PrivateKey(k),
			}),
		)
		cached = &cert
		return cached, err
	}
}

func main() {
	var (
		d kipp.Driver
		s kipp.Server
	)
	servecmd := kingpin.Command("serve", "Start a kipp server.").Default()

	addr := servecmd.
		Flag("addr", "Server listen address.").
		Default("127.0.0.1:443").
		String()
	cert := servecmd.
		Flag("cert", "TLS certificate path.").
		String()
	key := servecmd.
		Flag("key", "TLS key path.").
		String()
	cleanupInterval := servecmd.
		Flag("cleanup-interval", "Cleanup interval for deleting expired files.").
		Default("5m").
		Duration()
	mime := servecmd.
		Flag("mime", "A json formatted collection of extensions and mime types.").
		PlaceHolder("PATH").
		String()
	servecmd.
		Flag("driver", "Available database drivers: mysql, postgres, sqlite3 and mssql.").
		Default("sqlite3").
		StringVar(&d.Dialect)
	servecmd.
		Flag("driver-username", "Database driver username.").
		Default("kipp").
		StringVar(&d.Username)
	servecmd.
		Flag("driver-password", "Database driver password.").
		PlaceHolder("PASSWORD").
		StringVar(&d.Password)
	servecmd.
		Flag("driver-path", "Database driver path. ex: localhost:8080").
		Default("kipp.db").
		StringVar(&d.Path)
	servecmd.
		Flag("expiration", "File expiration time.").
		Default("24h").
		DurationVar(&s.Expiration)
	servecmd.
		Flag("max", "The maximum file size  for uploads.").
		Default("150MB").
		BytesVar((*units.Base2Bytes)(&s.Max))
	servecmd.
		Flag("files", "File path.").
		Default("files").
		StringVar(&s.FilePath)
	servecmd.
		Flag("tmp", "Temp path for in-progress uploads.").
		Default("files/tmp").
		StringVar(&s.TempPath)
	servecmd.
		Flag("public", "Public path for web resources.").
		Default("public").
		StringVar(&s.PublicPath)

	var u UploadCommand
	{
		uploadcmd := kingpin.Command("upload", "Upload a file.")
		uploadcmd.
			Arg("file", "File to be uploaded").
			Required().
			FileVar(&u.File)
		uploadcmd.
			Flag("insecure", "Don't verify SSL certificates.").
			BoolVar(&u.Insecure)
		uploadcmd.
			Flag("private", "Encrypt the uploaded file").
			BoolVar(&u.Private)
		uploadcmd.
			Flag("url", "Source URL").
			Envar("kipp-upload-url").
			Default("https://kipp.6f.io").
			URLVar(&u.URL)
	}

	t := kingpin.Parse()

	// kipp upload
	if t == "upload" {
		u.Do()
		return
	}

	// Connect to database
	db, err := d.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	s.DB = db

	// Load mime types
	if m := *mime; m != "" {
		if err := loadMimeTypes(m); err != nil {
			log.Fatal(err)
		}
	}

	// Make paths for files and temp files
	if err := os.MkdirAll(s.FilePath, 0755); err != nil {
		log.Fatal(err)
	}
	if err := os.MkdirAll(s.TempPath, 0755); err != nil {
		log.Fatal(err)
	}

	// Start cleanup worker
	if s.Expiration > 0 {
		w := worker(*cleanupInterval)
		go w.Do(context.Background(), s.Cleanup)
	}

	// Start HTTP server
	hs := &http.Server{
		Addr:    *addr,
		Handler: s,
		TLSConfig: &tls.Config{
			GetCertificate: CertificateGetter(),
			CipherSuites: []uint16{
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			},
			PreferServerCipherSuites: true,
			MinVersion:               tls.VersionTLS12,
			CurvePreferences:         []tls.CurveID{tls.CurveP256, tls.X25519},
		},
		// ReadTimeout:  5 * time.Second,
		// WriteTimeout: 10 * time.Second,
		IdleTimeout: 120 * time.Second,
	}
	// Use multiple cores for concurrent serving
	runtime.GOMAXPROCS(runtime.NumCPU())
	// Output a message so users know when the server has been started.
	log.Printf("Listening on %s", *addr)
	log.Fatal(hs.ListenAndServeTLS(*cert, *key))
}
