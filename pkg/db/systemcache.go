package db

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/gczuczy/ed-survey-tools/pkg/edsm"
)

// SystemCache provides a mid-term in-memory cache for common.systems
// rows, backed by the database and EDSM as fallback. Intended to be
// instantiated per batch processing job and discarded afterwards.
type SystemCache struct {
	cache map[string]System
	txn   *Transaction
	edsm  *edsm.EDSM
}

// NewSystemCache creates a SystemCache that uses txn for all
// database operations.
func NewSystemCache(txn *Transaction, edsm *edsm.EDSM) *SystemCache {
	return &SystemCache{
		cache: make(map[string]System),
		txn:   txn,
		edsm:  edsm,
	}
}

// Lookup resolves a batch of system names, returning a System for
// each. Lookup order: in-memory cache → database → EDSM (with DB
// insert). Returns an error if any name cannot be resolved.
func (sc *SystemCache) Lookup(names []string) ([]System, error) {
	result := make([]System, 0, len(names))
	missing := make([]string, 0)

	for _, name := range names {
		if sys, ok := sc.cache[name]; ok {
			result = append(result, sys)
		} else {
			missing = append(missing, name)
		}
	}

	if len(missing) == 0 {
		return result, nil
	}

	dbHits, err := sc.lookupDB(missing)
	if err != nil {
		return nil, err
	}

	dbFound := make(map[string]bool, len(dbHits))
	for _, sys := range dbHits {
		sc.cache[sys.Name] = sys
		result = append(result, sys)
		dbFound[sys.Name] = true
	}

	stillMissing := make([]string, 0)
	for _, name := range missing {
		if !dbFound[name] {
			stillMissing = append(stillMissing, name)
		}
	}

	if len(stillMissing) == 0 {
		return result, nil
	}

	edsmData, err := sc.edsm.Systems(stillMissing)
	if err != nil {
		return nil, err
	}

	edsmFound := make(map[string]bool, len(edsmData))
	for _, d := range edsmData {
		edsmFound[d.Name] = true
	}

	var errs []error
	for _, name := range stillMissing {
		if !edsmFound[name] {
			errs = append(errs, fmt.Errorf(
				"system %q not found in EDSM", name,
			))
		}
	}

	for _, d := range edsmData {
		if d.Coords == nil {
			errs = append(errs, fmt.Errorf(
				"system %q has no coordinates in EDSM",
				d.Name,
			))
			continue
		}
		sys, err := sc.insertDB(d)
		if err != nil {
			return nil, err
		}
		sc.cache[sys.Name] = sys
		result = append(result, sys)
	}

	return result, errors.Join(errs...)
}

func (sc *SystemCache) lookupDB(
	names []string,
) ([]System, error) {
	rows, err := sc.txn.tx.Query(
		sc.txn.ctx, "lookupsystemsbyname", names,
	)
	if err != nil {
		logger.Error().Err(err).Caller().
			Str("query", "lookupsystemsbyname").
			Msg("Error while executing query")
		return nil, err
	}
	defer rows.Close()

	systems, err := pgx.CollectRows(
		rows, pgx.RowToStructByName[System],
	)
	if err != nil {
		logger.Error().Err(err).Caller().
			Msg("Error while reading results")
		return nil, err
	}

	return systems, nil
}

func (sc *SystemCache) insertDB(
	d edsm.SystemData,
) (System, error) {
	rows, err := sc.txn.tx.Query(
		sc.txn.ctx, "setsystem",
		d.ID, d.Name,
		d.Coords.X, d.Coords.Y, d.Coords.Z,
	)
	if err != nil {
		logger.Error().Err(err).Caller().
			Str("query", "setsystem").
			Str("name", d.Name).
			Msg("Error while executing query")
		return System{}, err
	}
	defer rows.Close()

	systems, err := pgx.CollectRows(
		rows, pgx.RowToStructByName[System],
	)
	if err != nil {
		logger.Error().Err(err).Caller().
			Msg("Error while reading results")
		return System{}, err
	}

	if len(systems) == 0 {
		return System{}, fmt.Errorf(
			"no system returned after insert for %q", d.Name,
		)
	}

	return systems[0], nil
}
