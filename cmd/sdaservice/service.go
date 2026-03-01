package sdaservice

import (
	"os"
	"fmt"
	"syscall"
	"os/signal"

	"github.com/knadh/koanf/v2"
	"github.com/knadh/koanf/providers/posflag"
	flag "github.com/spf13/pflag"

	"github.com/gczuczy/ed-survey-tools/pkg/gcp"
	"github.com/gczuczy/ed-survey-tools/pkg/config"
	"github.com/gczuczy/ed-survey-tools/pkg/http"
	"github.com/gczuczy/ed-survey-tools/pkg/db"
	"github.com/gczuczy/ed-survey-tools/pkg/log"
)

func parseArgs(k *koanf.Koanf) error {
	f := flag.NewFlagSet("config", flag.ContinueOnError)
	f.Usage = func() {
		f.PrintDefaults()
		os.Exit(0)
	}

	f.StringP("sa-creds", "s", "credentials.json", "The Google Service Account credentials json")
	f.StringP("config", "c", "~/.edsda.yaml", "Path to the configuration file")
	if err := f.Parse(os.Args[1:]); err != nil {
		return err
	}

	return k.Load(posflag.Provider(f, ".", k), nil)
}

func Run() {
	var (
		cfg *config.Config
		hs *http.HTTPService
	)

	k := koanf.New(".")
	k.Print()

	err := parseArgs(k)
	if err != nil {
		fmt.Fprintf(os.Stderr, "err: %v\n", err)
		os.Exit(1)
	}

	if cfg, err = config.ParseConfig(k); err != nil {
		fmt.Fprintf(os.Stderr, "err: %v\n", err)
		os.Exit(1)
	}

	if err = log.Init(&cfg.Logging); err != nil {
		fmt.Fprintf(os.Stderr, "err: %v\n", err)
		os.Exit(1)
	}

	creds := k.String(`sa-creds`)
	if err = gcp.Init(creds); err != nil {
		fmt.Printf("GCP init error: %s", creds)
		os.Exit(1)
	}

	if err = db.Init(&cfg.DB); err != nil {
		fmt.Fprintf(os.Stderr, "err: %v\n", err)
		os.Exit(1)
	}
	defer db.Pool.Close()

	if hs, err = http.New(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "err: %v\n", err)
		os.Exit(1)
	}
	defer hs.Close()

	if err = hs.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "err: %v\n", err)
		os.Exit(1)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	fmt.Printf("Received %s, bailing out\n", sig)
	if err = hs.Shutdown(); err != nil {
		fmt.Fprintf(os.Stderr, "err: %v\n", err)
		os.Exit(1)
	}
}
