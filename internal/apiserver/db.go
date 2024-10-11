package apiserver

import (
	"encoding/json"
	"time"

	bolt "go.etcd.io/bbolt"
)

type Database struct {
	engine *bolt.DB
}

var DB *Database

func InitDB() error {
	boltDB, err := bolt.Open("./eagleeye.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return err
	}

	DB = &Database{engine: boltDB}

	err = DB.engine.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("plans"))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte("results"))
		if err != nil {
			return err
		}

		return nil
	})

	return err
}

func ReleaseDB() {
	DB.engine.Close()
}

func (db *Database) StorePlan(planID string, plan *CreatePlanRequest) error {
	return WithCaller(db.engine.Update(func(tx *bolt.Tx) error {
		b, err := json.Marshal(plan)
		if err != nil {
			return WithCaller(err)
		}
		return WithCaller(tx.Bucket([]byte("plans")).Put([]byte(planID), b))
	}))
}

func (db *Database) GetPlan(planID string) (*CreatePlanRequest, error) {
	var options CreatePlanRequest
	err := db.engine.View(func(tx *bolt.Tx) error {
		v := tx.Bucket([]byte("plans")).Get([]byte(planID))
		if v == nil {
			return WithCaller(ErrPlanNotFound)
		}
		return WithCaller(json.Unmarshal(v, &options))
	})
	return &options, WithCaller(err)
}

func (db *Database) RestorePlan(oPlanID string, nPlanID string, options *CreatePlanRequest) error {
	return WithCaller(db.engine.Update(func(tx *bolt.Tx) error {
		err := tx.Bucket([]byte("plans")).Delete([]byte(oPlanID))
		if err != nil {
			return WithCaller(err)
		}

		b, err := json.Marshal(options)
		if err != nil {
			return WithCaller(err)
		}
		return WithCaller(tx.Bucket([]byte("plans")).Put([]byte(nPlanID), b))
	}))
}

func (db *Database) DeletePlan(planID string) error {
	return WithCaller(db.engine.Update(func(tx *bolt.Tx) error {
		return WithCaller(tx.Bucket([]byte("plans")).Delete([]byte(planID)))
	}))
}

func (db *Database) StoreResults(planID string, results *GetPlanResultsReplay) error {
	return WithCaller(db.engine.Update(func(tx *bolt.Tx) error {
		b, err := json.Marshal(results)
		if err != nil {
			return WithCaller(err)
		}
		err = tx.Bucket([]byte("results")).Put([]byte(planID), b)
		if err != nil {
			return WithCaller(err)
		}
		return WithCaller(tx.Bucket([]byte("plans")).Delete([]byte(planID)))
	}))
}

func (db *Database) GetResults(planID string) (*GetPlanResultsReplay, error) {
	var results GetPlanResultsReplay
	err := db.engine.View(func(tx *bolt.Tx) error {
		v := tx.Bucket([]byte("results")).Get([]byte(planID))
		if v == nil {
			return WithCaller(ErrPlanResultsNotFound)
		}
		return WithCaller(json.Unmarshal(v, &results))
	})
	return &results, WithCaller(err)
}

func (db *Database) RunningPlans() ([]string, error) {
	plans := make([]string, 0)
	err := db.engine.View(func(tx *bolt.Tx) error {
		return WithCaller(tx.Bucket([]byte("plans")).ForEach(func(k, v []byte) error {
			plans = append(plans, string(k))
			return nil
		}))
	})
	return plans, WithCaller(err)
}

func (db *Database) StoppedPlans() ([]string, error) {
	plans := make([]string, 0)
	err := db.engine.View(func(tx *bolt.Tx) error {
		return WithCaller(tx.Bucket([]byte("results")).ForEach(func(k, v []byte) error {
			plans = append(plans, string(k))
			return nil
		}))
	})
	return plans, WithCaller(err)
}
