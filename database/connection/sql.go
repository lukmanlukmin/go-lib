package connection

import (
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

// Driver list
const (
	DriverMySQL    = "mysql"
	DriverPostgres = "postgres"
)

// Db object
var (
	Master *DB
	Slave  *DB
)

type (
	//DSNConfig for database source name
	DSNConfig struct {
		DSN string
	}

	//DBConfig for databases configuration
	DBConfig struct {
		MasterDSN     string `json:"master_dsn" mapstructure:"master_dsn"`
		SlaveDSN      string `json:"slave_dsn" mapstructure:"slave_dsn"`
		EnableSlave   bool   `json:"enable_slave" mapstructure:"enable_slave"`
		RetryInterval int    `json:"retry_interval" mapstructure:"retry_interval"`
		MaxIdleConn   int    `json:"max_idle" mapstructure:"max_idle"`
		MaxConn       int    `json:"max_con" mapstructure:"max_con"`
	}

	//DB configuration
	DB struct {
		DBConnection  *sqlx.DB
		DBString      string
		RetryInterval int
		MaxIdleConn   int
		MaxConn       int
		doneChannel   chan bool
	}

	Store struct {
		Master *sqlx.DB
		Slave  *sqlx.DB
	}
)

func (s *Store) GetMaster() *sqlx.DB {
	return s.Master
}

func (s *Store) GetSlave() *sqlx.DB {
	return s.Slave
}

func New(cfg DBConfig, dbDriver string) *Store {
	masterDSN := cfg.MasterDSN
	slaveDSN := cfg.SlaveDSN
	store := &Store{}
	Master = &DB{
		DBString:      masterDSN,
		RetryInterval: cfg.RetryInterval,
		MaxIdleConn:   cfg.MaxIdleConn,
		MaxConn:       cfg.MaxConn,
		doneChannel:   make(chan bool),
	}
	err := Master.ConnectAndMonitor(dbDriver)
	if err != nil {
		log.Fatal("Could not initiate Master DB connection: " + err.Error())
		return &Store{}
	}
	store.Master = Master.DBConnection

	if cfg.EnableSlave {
		Slave = &DB{
			DBString:      slaveDSN,
			RetryInterval: cfg.RetryInterval,
			MaxIdleConn:   cfg.MaxIdleConn,
			MaxConn:       cfg.MaxConn,
			doneChannel:   make(chan bool),
		}
		err = Slave.ConnectAndMonitor(dbDriver)
		if err != nil {
			log.Fatal("Could not initiate Slave DB connection: " + err.Error())
			return &Store{}
		}
		store.Slave = Slave.DBConnection
	}

	time.NewTicker(time.Second * 2)
	return store
}

// Connect to database
func (d *DB) Connect(driver string) error {

	var db *sqlx.DB
	var err error
	db, err = sqlx.Open(driver, d.DBString)

	if err != nil {
		log.Println("[Error]: DB open connection error", err.Error())
	} else {
		d.DBConnection = db
		err = db.Ping()
		if err != nil {
			log.Println("[Error]: DB connection error", err.Error())
		}
		return err
	}

	db.SetMaxOpenConns(d.MaxConn)
	db.SetMaxIdleConns(d.MaxIdleConn)

	return err
}

// ConnectAndMonitor to database
func (d *DB) ConnectAndMonitor(driver string) error {
	if err := d.Connect(driver); err != nil {
		log.Printf("Not connected to database %s, trying", d.DBString)
		return err
	}

	ticker := time.NewTicker(time.Duration(d.RetryInterval) * time.Second)
	go d.monitorConnection(driver, ticker)

	return nil
}

func (d *DB) monitorConnection(driver string, ticker *time.Ticker) {
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := d.checkAndReconnect(driver); err != nil {
				log.Println("[Error]: DB reconnect error", err.Error())
			}
		case <-d.doneChannel:
			return
		}
	}
}

func (d *DB) checkAndReconnect(driver string) error {
	if d.DBConnection == nil {
		return d.Connect(driver)
	}
	return d.DBConnection.Ping()
}
