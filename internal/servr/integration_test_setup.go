package servr

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	dbUser     = "root"
	dbPassword = "testpass123"
	dbName     = "todotest"
	dbPort     = "3306"
)

type TestDB struct {
	DB        *sql.DB
	Container testcontainers.Container
	Ctx       context.Context
}

func setupTestDB(ctx context.Context, t *testing.T) *TestDB {
	req := testcontainers.ContainerRequest{
		Image:        "mysql:5.7",
		ExposedPorts: []string{dbPort + "/tcp"},
		Env: map[string]string{
			"MYSQL_ROOT_PASSWORD": dbPassword,
			"MYSQL_DATABASE":      dbName,
		},
		WaitingFor: wait.ForListeningPort("3306/tcp"),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

	if err != nil {
		t.Fatalf("failed to start container: %v", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("failed to get container host: %v", err)
	}

	port, err := container.MappedPort(ctx, "3306")
	if err != nil {
		t.Fatalf("failed to get mapped port: %v", err)
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", dbUser, dbPassword, host, port.Port(), dbName)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}

	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(3 * time.Minute)

	// Retry ping with backoff
	var lastErr error
	for i := 0; i < 10; i++ {
		if err := db.PingContext(ctx); err == nil {
			break
		} else {
			lastErr = err
			time.Sleep(500 * time.Millisecond)
		}
	}
	if lastErr != nil {
		t.Fatalf("failed to ping database: %v", lastErr)
	}

	// Initialize schema
	if err := initializeSchema(ctx, db); err != nil {
		t.Fatalf("failed to initialize schema: %v", err)
	}

	return &TestDB{
		DB:        db,
		Container: container,
		Ctx:       ctx,
	}
}

func (tdb *TestDB) Cleanup(t *testing.T) {
	if tdb.DB != nil {
		if err := tdb.DB.Close(); err != nil {
			t.Logf("failed to close database: %v", err)
		}
	}

	if tdb.Container != nil {
		if err := tdb.Container.Terminate(tdb.Ctx); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	}
}

func initializeSchema(ctx context.Context, db *sql.DB) error {
	statements := []string{
		`create table if not exists person
		(
		    person_id   varchar(64)  not null primary key,
		    provided_id varchar(128) not null unique,
		    person_name varchar(255) not null,
		    login       varchar(64)  not null,
		    password_salt varchar(255),
		    password_payload varchar(4096)
		)`,
		`create table if not exists session
		(
		    session_id  varchar(64) not null primary key,
		    updated_dt  datetime    not null,
		    token       varchar(64) not null,
		    expired_ind varchar(1)  not null
		)`,
		`create table if not exists session_person
		(
		    session_person_id varchar(64) not null primary key,
		    session_id        varchar(64) not null,
		    person_id         varchar(64) not null,
		    foreign key (session_id) references session (session_id),
		    foreign key (person_id) references person (person_id)
		)`,
		`create table if not exists todo_item
		(
		    todo_item_id  varchar(64) not null primary key,
		    person_id     varchar(64) not null,
		    comments      mediumtext  not null,
		    project_cd    varchar(255),
		    context_cd    varchar(255),
		    priority      int(11) not null,
		    added_dt      datetime    not null,
		    due_dt        datetime,
		    completed_ind varchar(1),
		    updated_dt    datetime not null,
		    foreign key (person_id) references person (person_id)
		)`,
		`create table if not exists item
		(
		    item_id     varchar(192) not null primary key,
		    person_id   varchar(192),
		    item_name   varchar(765),
		    description mediumtext,
		    category    varchar(765),
		    flag_ind    varchar(3),
		    added_dt    datetime,
		    updated_dt  datetime,
		    foreign key (person_id) references person (person_id)
		)`,
		`create table if not exists attachment
		(
		    attachment_id varchar(192) not null primary key,
		    item_id       varchar(192) not null,
		    description   varchar(765),
		    file_name     varchar(192),
		    content_type  varchar(384),
		    seq_no        bigint(20),
		    added_dt      datetime,
		    updated_dt    datetime,
		    foreign key (item_id) references item (item_id)
		)`,
	}

	for _, stmt := range statements {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return err
		}
	}
	return nil
}

func (tdb *TestDB) InsertPerson(ctx context.Context, personID, providedID, name, login string) error {
	query := `insert into person (person_id, provided_id, person_name, login) values (?, ?, ?, ?)`
	_, err := tdb.DB.ExecContext(ctx, query, personID, providedID, name, login)
	return err
}

func (tdb *TestDB) InsertSession(ctx context.Context, sessionID, token string, persons []string) error {
	tx, err := tdb.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `insert into session (session_id, updated_dt, token, expired_ind) values (?, now(), ?, ?)`
	if _, err := tx.ExecContext(ctx, query, sessionID, token, "N"); err != nil {
		return err
	}

	for _, personID := range persons {
		spQuery := `insert into session_person (session_person_id, session_id, person_id) values (uuid(), ?, ?)`
		if _, err := tx.ExecContext(ctx, spQuery, sessionID, personID); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (tdb *TestDB) InsertTodoItem(ctx context.Context, todoID, personID, comments string, priority int) error {
	query := `insert into todo_item (todo_item_id, person_id, comments, priority, added_dt, updated_dt, completed_ind)
	          values (?, ?, ?, ?, now(), now(), 'N')`
	_, err := tdb.DB.ExecContext(ctx, query, todoID, personID, comments, priority)
	return err
}

func (tdb *TestDB) InsertItem(ctx context.Context, itemID, personID, name string) error {
	query := `insert into item (item_id, person_id, item_name, added_dt, updated_dt) values (?, ?, ?, now(), now())`
	_, err := tdb.DB.ExecContext(ctx, query, itemID, personID, name)
	return err
}

func (tdb *TestDB) ClearTables(ctx context.Context) error {
	tables := []string{"attachment", "session_person", "session", "todo_item", "item", "person"}
	for _, table := range tables {
		if _, err := tdb.DB.ExecContext(ctx, fmt.Sprintf("delete from %s", table)); err != nil {
			return err
		}
	}
	return nil
}
