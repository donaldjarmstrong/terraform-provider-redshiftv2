package redshift

import (
	"context"
	"fmt"
	"terraform-provider-redshift/internal/helpers"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/jackc/pgx/v5"
)

type svv_roles struct {
	Id         string  `db:"role_id"`
	RoleName   string  `db:"role_name"`
	ExternalId *string `db:"external_id"`
	RoleOwner  string  `db:"role_owner"`
}

type Role struct {
	svv_roles
}

type RoleService struct {
	cfg     *pgx.ConnConfig
	ctx     context.Context
	timeout time.Duration
}

func NewRoleService(ctx context.Context, cfg *pgx.ConnConfig) (*RoleService, error) {
	timeout := cfg.ConnectTimeout
	if cfg.ConnectTimeout <= 0 {
		tflog.Info(ctx, "No timeout provided, using 5 minutes")
		timeout = time.Minute * 5
	}

	return &RoleService{
		cfg:     cfg,
		ctx:     ctx,
		timeout: timeout,
	}, nil
}

func (s *RoleService) FindRole(id string) (*Role, error) {
	sql := `
	SELECT svv.external_id,
		   svv.role_id::varchar,
		   svv.role_owner,
		   svv.role_name
	  FROM svv_roles svv
	 WHERE svv.role_id = @RoleId
	`
	args := pgx.NamedArgs{"RoleId": id}

	ctx, cancel := context.WithTimeout(s.ctx, s.timeout)
	defer cancel()

	conn, err := pgx.ConnectConfig(ctx, s.cfg)
	if err != nil {
		return nil, fmt.Errorf("FindRole: Unable to connect %w", err)
	}

	tx, err := conn.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("FindRole: Failed to begin transaction: %w", err)
	}

	role, err := buildRole(sql, args, ctx, tx)
	if err != nil {
		return nil, fmt.Errorf("FindRole: Failed to build role: %w", err)
	}

	if role == nil {
		return nil, fmt.Errorf("FindRole: Could not find role with id '%s'", id)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("FindRole: Failed to commit: %w", err)
	}

	return role, nil
}

func (s *RoleService) DropRole(name string) error {
	sql := fmt.Sprintf("DROP ROLE %s", pgx.Identifier{name}.Sanitize())

	ctx, cancel := context.WithTimeout(s.ctx, s.timeout)
	defer cancel()

	conn, err := pgx.ConnectConfig(ctx, s.cfg)
	if err != nil {
		return fmt.Errorf("DropRole: Unable to connect %w", err)
	}

	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("DropRole: Failed to begin transaction: %w", err)
	}

	_, err = tx.Exec(ctx, sql)
	if err != nil {
		return fmt.Errorf("DropRole: Failed to execute: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("DropRole: Failed to commit: %w", err)
	}

	return nil
}

type CreateRoleDDLParams struct {
	Name       string
	ExternalId *string
}

func (s *RoleService) CreateRole(args CreateRoleDDLParams) (*Role, error) {
	t := `
		CREATE ROLE {{.Name}}
			{{if .ExternalId}}EXTERNALID '{{.ExternalId}}' {{end}}
	`
	name := args.Name // save this unsanitized for lookup later
	args.Name = pgx.Identifier{args.Name}.Sanitize()

	ctx, cancel := context.WithTimeout(s.ctx, s.timeout)
	defer cancel()

	conn, err := pgx.ConnectConfig(ctx, s.cfg)
	if err != nil {
		return nil, fmt.Errorf("CreateRole: Unable to connect %w", err)
	}

	tx, err := conn.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("CreateRole: Failed to begin transaction: %w", err)
	}

	sql, err := helpers.Merge(t, args)
	if err != nil {
		return nil, fmt.Errorf("CreateRole: Failed to merge template: %w", err)
	}

	_, err = tx.Exec(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("CreateRole: Failed to execute: %w", err)
	}

	role, err := getRoleByName(name, ctx, tx)
	if err != nil {
		return nil, fmt.Errorf("CreateRole: Failed to getRoleByName: %w", err)
	}
	if role == nil {
		return nil, fmt.Errorf("CreateRole: Could not find role with name '%s'", name)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("CreateRole: Failed to commit: %w", err)
	}

	return role, nil
}

type AlterRoleDDLParams struct {
	Name       string
	RenameTo   *string
	ExternalId *string
}

func (s *RoleService) AlterRole(args AlterRoleDDLParams) error {
	args.Name = pgx.Identifier{args.Name}.Sanitize()

	ctx, cancel := context.WithTimeout(s.ctx, s.timeout)
	defer cancel()

	conn, err := pgx.ConnectConfig(ctx, s.cfg)
	if err != nil {
		return fmt.Errorf("AlterRole: Unable to connect %w", err)
	}

	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("AlterRole: Failed to begin transaction: %w", err)
	}

	if args.RenameTo != nil {
		rn := pgx.Identifier{*args.RenameTo}.Sanitize()
		args.RenameTo = &rn
	}

	t := `
		ALTER ROLE {{.Name}}
			{{if .RenameTo}}RENAME TO {{.RenameTo}}{{end}}
			{{if .ExternalId}}EXTERNALID '{{.ExternalId}}' {{end}}
		`

	sql, err := helpers.Merge(t, args)
	if err != nil {
		return fmt.Errorf("AlterRole: failed to merge template: %w", err)
	}

	_, err = tx.Exec(ctx, sql)
	if err != nil {
		return fmt.Errorf("AlterRole: failed to execute: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("AlterRole: Failed to commit: %w", err)
	}

	return nil
}

// hidden from outside the package, expect that callers use the ById variant.
func getRoleByName(name string, ctx context.Context, tx pgx.Tx) (*Role, error) {
	sql := `
	SELECT svv.external_id,
		   svv.role_id::varchar,
		   svv.role_owner,
		   svv.role_name
	  FROM svv_roles svv
	 WHERE svv.role_name = @RoleName
	`
	args := pgx.NamedArgs{"RoleName": name}

	return buildRole(sql, args, ctx, tx)
}

func buildRole(sql string, args pgx.NamedArgs, ctx context.Context, tx pgx.Tx) (*Role, error) {
	rows, err := tx.Query(ctx, sql, args)
	if err != nil {
		return nil, fmt.Errorf("buildRole: Failed query execute: %w", err)
	}

	svv_roles, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[svv_roles])
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}

		return nil, fmt.Errorf("buildRole: Failed to collect row: %w", err)
	}

	role := Role{
		svv_roles: svv_roles,
	}

	return &role, nil
}
