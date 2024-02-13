package redshift

import (
	"context"
	"fmt"
	"terraform-provider-redshift/internal/static"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type svv_user_info struct {
	Id              string  `db:"user_id"`
	CreateDb        bool    `db:"createdb"`
	CreateUser      bool    `db:"superuser"`
	ConnectionLimit string  `db:"connection_limit"`
	SyslogAccess    string  `db:"syslog_access"`
	SessionTimeout  int64   `db:"session_timeout"`
	ExternalId      *string `db:"external_user_id"`
	UserName        string  `db:"user_name"`
}

type pg_user_info struct {
	ValidUntil pgtype.Timestamp `db:"valid_until"`
}

type User struct {
	svv_user_info
	ValidUntil string
}

type UserService struct {
	cfg     *pgx.ConnConfig
	ctx     context.Context
	timeout time.Duration
}

func NewUserService(ctx context.Context, cfg *pgx.ConnConfig) (*UserService, error) {
	timeout := cfg.ConnectTimeout
	if cfg.ConnectTimeout <= 0 {
		tflog.Info(ctx, "No timeout provided, using 5 minutes")
		timeout = time.Minute * 5
	}

	return &UserService{
		cfg:     cfg,
		ctx:     ctx,
		timeout: timeout,
	}, nil
}

func (s *UserService) FindUser(id string) (*User, error) {
	sql := `
	SELECT svv.connection_limit,
		   svv.createdb,
		   svv.superuser,
		   svv.external_user_id,
		   svv.user_id::varchar,
		   svv.session_timeout,
		   svv.syslog_access,
		   svv.user_name
	  FROM svv_user_info svv
	 WHERE svv.user_id = @UserId
	`
	args := pgx.NamedArgs{"UserId": id}

	ctx, cancel := context.WithTimeout(s.ctx, s.timeout)
	defer cancel()

	conn, err := pgx.ConnectConfig(ctx, s.cfg)
	if err != nil {
		return nil, fmt.Errorf("FindUser: Unable to connect %w", err)
	}

	tx, err := conn.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("FindUser: Failed to begin transaction: %w", err)
	}

	user, err := buildUser(sql, args, ctx, tx)
	if err != nil {
		return nil, fmt.Errorf("FindUser: Failed to build user: %w", err)
	}

	if user == nil {
		return nil, fmt.Errorf("FindUser: Could not find user with id '%s'", id)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("FindUser: Failed to commit: %w", err)
	}

	return user, nil
}

func (s *UserService) DropUser(name string) error {
	sql := fmt.Sprintf("DROP USER %s", pgx.Identifier{name}.Sanitize())

	ctx, cancel := context.WithTimeout(s.ctx, s.timeout)
	defer cancel()

	conn, err := pgx.ConnectConfig(ctx, s.cfg)
	if err != nil {
		return fmt.Errorf("DropUser: Unable to connect %w", err)
	}

	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("DropUser: Failed to begin transaction: %w", err)
	}

	_, err = tx.Exec(ctx, sql)
	if err != nil {
		return fmt.Errorf("DropUser: Failed to execute: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("DropUser: Failed to commit: %w", err)
	}

	return nil
}

type CreateUserDDLParams struct {
	Name            string
	Password        *string
	CreateDb        bool
	CreateUser      bool
	SyslogAccess    string
	ValidUntil      string
	ConnectionLimit string
	SessionTimeout  int64
	ExternalId      *string
}

func (s *UserService) CreateUser(ddl CreateUserDDLParams) (*User, error) {
	t := `
		CREATE USER {{.Name}}
		    PASSWORD {{if .Password}}'{{.Password}}'{{else}}DISABLE{{end}}
			{{if .CreateDb}}CREATEDB{{else}}NOCREATEDB{{end}}
			{{if .CreateUser}}CREATEUSER{{else}}NOCREATEUSER{{end}}
			SYSLOG ACCESS {{.SyslogAccess}}
			VALID UNTIL '{{.ValidUntil}}'
			CONNECTION LIMIT {{.ConnectionLimit}}
			{{if gt .SessionTimeout 0}}SESSION TIMEOUT {{.SessionTimeout}}{{end}}
			{{if .ExternalId}}EXTERNALID '{{.ExternalId}}' {{end}}
	`
	args := CreateUserDDLParams(ddl)
	args.Name = pgx.Identifier{ddl.Name}.Sanitize()

	ctx, cancel := context.WithTimeout(s.ctx, s.timeout)
	defer cancel()

	conn, err := pgx.ConnectConfig(ctx, s.cfg)
	if err != nil {
		return nil, fmt.Errorf("CreateUser: Unable to connect %w", err)
	}

	tx, err := conn.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("CreateUser: Failed to begin transaction: %w", err)
	}

	sql, err := static.Merge(t, args)
	if err != nil {
		return nil, fmt.Errorf("CreateUser: Failed to merge template: %w", err)
	}

	_, err = tx.Exec(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("CreateUser: Failed to execute: %w", err)
	}

	user, err := getUserByName(ddl.Name, ctx, tx)
	if err != nil {
		return nil, fmt.Errorf("CreateUser: Failed to getUserByName: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("CreateUser: Failed to commit: %w", err)
	}

	return user, nil
}

type AlterUserDDLParams struct {
	Name            string
	RenameTo        *string
	Password        *string
	CreateDb        *bool
	CreateUser      *bool
	SyslogAccess    *string
	ValidUntil      *string
	ConnectionLimit *string
	SessionTimeout  *int64
	ExternalId      *string
}

func (s *UserService) AlterUser(ddl AlterUserDDLParams) error {
	args := AlterUserDDLParams(ddl)
	args.Name = pgx.Identifier{ddl.Name}.Sanitize()

	ctx, cancel := context.WithTimeout(s.ctx, s.timeout)
	defer cancel()

	conn, err := pgx.ConnectConfig(ctx, s.cfg)
	if err != nil {
		return fmt.Errorf("AlterUser: Unable to connect %w", err)
	}

	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("AlterUser: Failed to begin transaction: %w", err)
	}

	// Undocumented, redshift must perform a rename without other options; it is a syntax error otherwise
	if ddl.RenameTo != nil {
		rn := pgx.Identifier{*ddl.RenameTo}.Sanitize()
		args.RenameTo = &rn

		rename := `
			ALTER USER {{.Name}} RENAME TO {{.RenameTo}}
		`
		sql, err := static.Merge(rename, args)
		if err != nil {
			return fmt.Errorf("AlterUser: failed to merge rename template: %w", err)
		}

		_, err = tx.Exec(ctx, sql)
		if err != nil {
			return fmt.Errorf("AlterUser: failed to execute rename: %w", err)
		}
	}

	t := `
		ALTER USER {{if .RenameTo}}{{.RenameTo}}{{else}}{{.Name}}{{end}}
		    {{if .Password}}PASSWORD {{if (DerefString .Password)}}'{{.Password}}'{{else}}DISABLE{{end}}{{end}}
			{{if .CreateDb}}{{if (DerefBool .CreateDb)}}CREATEDB{{else}}NOCREATEDB{{end}}{{end}}
			{{if .CreateUser}}{{if (DerefBool .CreateUser)}}CREATEUSER{{else}}NOCREATEUSER{{end}}{{end}}
			{{if .SyslogAccess}}SYSLOG ACCESS {{.SyslogAccess}}{{end}}
			{{if .ValidUntil}}VALID UNTIL '{{.ValidUntil}}'{{end}}
			{{if .ConnectionLimit}}CONNECTION LIMIT {{.ConnectionLimit}}{{end}}
			{{if .SessionTimeout}}{{if gt (DerefInt64 .SessionTimeout) 0}}SESSION TIMEOUT {{.SessionTimeout}}{{else}}RESET SESSION TIMEOUT{{end}}{{end}}
			{{if .ExternalId}}EXTERNALID '{{.ExternalId}}' {{end}}
		`

	sql, err := static.Merge(t, args)
	if err != nil {
		return fmt.Errorf("AlterUser: failed to merge template: %w", err)
	}

	_, err = tx.Exec(ctx, sql)
	if err != nil {
		return fmt.Errorf("AlterUser: failed to execute: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("AlterUser: Failed to commit: %w", err)
	}

	return nil
}

func getUserValidUntil(id string, ctx context.Context, tx pgx.Tx) (*pg_user_info, error) {
	sql := `
		select coalesce(valuntil::timestamp, 'infinity') as valid_until
		  from pg_user_info pg
		 where usesysid = @UserId
	`
	args := pgx.NamedArgs{"UserId": id}

	rows, err := tx.Query(ctx, sql, args)
	if err != nil {
		return nil, fmt.Errorf("getUserValidUntil: failed query execute: %w", err)
	}

	pg_user_info, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[pg_user_info])
	if err != nil {
		return nil, fmt.Errorf("getUserValidUntil: failed to collect row: %w", err)
	}

	return &pg_user_info, nil
}

// hidden from outside the package, expect that callers use the ById variant
func getUserByName(name string, ctx context.Context, tx pgx.Tx) (*User, error) {
	sql := `
	SELECT svv.connection_limit,
		   svv.createdb,
		   svv.superuser,
		   svv.external_user_id,
		   svv.user_id::varchar,
		   svv.session_timeout,
		   svv.syslog_access,
		   svv.user_name
	  FROM svv_user_info svv
	 WHERE svv.user_name = @UserName
	`
	args := pgx.NamedArgs{"UserName": name}

	return buildUser(sql, args, ctx, tx)
}

func buildUser(sql string, args pgx.NamedArgs, ctx context.Context, tx pgx.Tx) (*User, error) {
	rows, err := tx.Query(ctx, sql, args)
	if err != nil {
		return nil, fmt.Errorf("buildUser: Failed query execute: %w", err)
	}

	svv_user_info, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[svv_user_info])
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}

		return nil, fmt.Errorf("buildUser: Failed to collect row: %w", err)
	}

	pg_user_info, err := getUserValidUntil(svv_user_info.Id, ctx, tx)
	if err != nil {
		return nil, fmt.Errorf("buildUser: Failed to fetch getUserValidUntil: %w", err)
	}

	validUntil := "infinity"
	if pg_user_info.ValidUntil.InfinityModifier == pgtype.Finite {
		validUntil = pg_user_info.ValidUntil.Time.Format(time.DateTime)
	}

	user := User{
		ValidUntil:    validUntil,
		svv_user_info: svv_user_info,
	}

	return &user, nil
}
