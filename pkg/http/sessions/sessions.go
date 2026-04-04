package sessions

import (
	"fmt"
	"time"
	"errors"
	"net/http"

	"github.com/gomodule/redigo/redis"
	"github.com/boj/redistore/v2"
	"github.com/gorilla/sessions"

	"github.com/gczuczy/ed-survey-tools/pkg/config"
	"github.com/gczuczy/ed-survey-tools/pkg/misc"
)

var (
	Store *redistore.RediStore
)

const (
	SessionName = "edst-auth"
)

func Init(cfg *config.SessionsConfig) error {

	if len(cfg.Key) == 0 {
		return fmt.Errorf("Session key not specified")
	}

	if cfg.Store == "redis" {

		maxidle := misc.Coalesce(cfg.Redis.MaxIdle, 16)
		idletimeout := misc.Coalesce(cfg.Redis.IdleTimeout, 5*time.Minute)
		port := misc.Coalesce(cfg.Redis.Port, 6379)
		host := misc.Coalesce(cfg.Redis.Host, "localhost")

		addr := fmt.Sprintf("%s:%d", host, port)

		opts := make([]redis.DialOption, 0)
		if cfg.Redis.DB != nil {
			opts = append(opts, redis.DialDatabase(*cfg.Redis.DB))
		}
		if cfg.Redis.User != nil {
			opts = append(opts, redis.DialUsername(*cfg.Redis.User))
		}
		if cfg.Redis.Pass != nil {
			opts = append(opts, redis.DialPassword(*cfg.Redis.Pass))
		}

		pool := &redis.Pool{
			MaxIdle: maxidle,
			IdleTimeout: idletimeout,
			Dial: func() (redis.Conn, error) {
				return redis.Dial("tcp", addr, opts...)
			},
		}
		maxage := cfg.MaxAge
		if maxage <= 0 {
			maxage = 7200
		}
		cookieOpts := &sessions.Options{
			Path:     "/",
			HttpOnly: true,
			Secure:   cfg.Secure,
			SameSite: http.SameSiteStrictMode,
			MaxAge:   maxage,
		}
		var err error
		Store, err = redistore.NewStore(
			redistore.KeysFromStrings(cfg.Key),
			redistore.WithPool(pool),
			redistore.WithSessionOptions(cookieOpts),
		)
		if err != nil {
			return errors.Join(err, fmt.Errorf("Unable to init redis pool"))
		}
	} else {
		return fmt.Errorf("Session store %s not supported", cfg.Store)
	}
	return nil
}

func Close() error {
	if Store != nil {
		return Store.Close()
	}
	return nil
}

func Get(r *http.Request) (*sessions.Session, error) {
	return Store.Get(r, SessionName)
}

func Save(r *http.Request, w http.ResponseWriter, session *sessions.Session) error {
	return Store.Save(r, w, session)
}
