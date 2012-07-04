// go-camo daemon (go-camod)
package main

import (
	"encoding/json"
	"flag"
	"github.com/cactus/gologit"
	"go-camo/camoproxy"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"sync"
	"syscall"
)

func main() {
	var gmx int
	if gmxEnv := os.Getenv("GOMAXPROCS"); gmxEnv != "" {
		gmx, _ = strconv.Atoi(gmxEnv)
	} else {
		gmx = runtime.NumCPU()
	}
	runtime.GOMAXPROCS(gmx)

	// command line flags
	debug := flag.Bool("debug", false, "Enable Debug Logging")
	hmacKey := flag.String("hmac-key", "", "HMAC Key")
	configFile := flag.String("config-file", "", "JSON Config File")
	maxSize := flag.Int64("max-size", 5120, "Max size in KB to allow")
	reqTimeout := flag.Uint("timeout", 4, "Upstream request timeout in seconds")
	follow := flag.Bool("follow-redirects", false,
		"Enable following upstream redirects")
	bindAddress := flag.String("bind-address", "0.0.0.0:8080",
		"Address:Port to bind to for HTTP")
	bindAddressSSL := flag.String("bind-address-ssl", "",
		"Address:Port to bind to for HTTPS/SSL/TLS")
	sslKey := flag.String("ssl-key", "",
		"Path to ssl private key (key.pem). Required if bind-address-ssl is specified.")
	sslCert := flag.String("ssl-cert", "",
		"Path to ssl cert (cert.pem). Required if bind-address-ssl is specified.")
	// parse said flags
	flag.Parse()

	// Anonymous struct Container for holding configuration parameters parsed
	// from JSON config file.
	config := camoproxy.ProxyConfig{}

	if *configFile != "" {
		b, err := ioutil.ReadFile(*configFile)
		if err != nil {
			log.Fatal("Could not read configFile", err)
		}
		err = json.Unmarshal(b, &config)
		if err != nil {
			log.Fatal("Could not parse configFile", err)
		}
	}

	// flags override config file
	if *hmacKey != "" {
		config.HmacKey = *hmacKey
	}

	if config.MaxSize == 0 {
		config.MaxSize = *maxSize
	}

	if *bindAddress == "" && *bindAddressSSL == "" {
		log.Fatal("One of bind-address or bind-ssl-address required")
	}

	if *bindAddressSSL != "" && *sslKey == "" {
		log.Fatal("ssl-key is required when specifying bind-ssl-address")
	}
	if *bindAddressSSL != "" && *sslCert == "" {
		log.Fatal("ssl-cert is required when specifying bind-ssl-address")
	}

	// convert from KB to Bytes
	config.MaxSize = config.MaxSize * 1024

	config.RequestTimeout = *reqTimeout
	config.FollowRedirects = *follow

	// create logger and start toggle on signal handler
	logger := gologit.New(*debug)
	logger.Debugln("Debug logging enabled")
	logger.ToggleOnSignal(syscall.SIGUSR1)

	proxy := camoproxy.New(config, logger)

	http.Handle("/favicon.ico", http.NotFoundHandler())
	http.Handle("/", proxy)

	wg := &sync.WaitGroup{}

	if *bindAddress != "" {
		log.Println("Starting server on", *bindAddress)
		wg.Add(1)
		go func() {
			err := http.ListenAndServe(*bindAddress, nil)
			if err != nil {
				log.Fatal("ListenAndServe: ", err)
			}
			wg.Done()
		}()
	}
	if *bindAddressSSL != "" {
		log.Println("Starting TLS server on", *bindAddressSSL)
		wg.Add(1)
		go func() {
			err := http.ListenAndServeTLS(*bindAddressSSL, *sslCert, *sslKey, nil)
			if err != nil {
				log.Fatal("ListenAndServeTLS: ", err)
			}
			wg.Done()
		}()
	}
	wg.Wait()
}
