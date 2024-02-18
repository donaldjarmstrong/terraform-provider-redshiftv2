package redshift

import (
	"context"
	"fmt"
	"terraform-provider-redshift/internal/helpers"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/jackc/pgx/v5"
)

type pg_group struct {
	GroupName string  `db:"groname"`
	Id        string  `db:"grosysid"`
	Username  *string `db:"usename"`
}

type Group struct {
	Id    string
	Name  string
	Users *[]string
}

type GroupService struct {
	cfg     *pgx.ConnConfig
	ctx     context.Context
	timeout time.Duration
}

func NewGroupService(ctx context.Context, cfg *pgx.ConnConfig) (*GroupService, error) {
	timeout := cfg.ConnectTimeout
	if cfg.ConnectTimeout <= 0 {
		tflog.Info(ctx, "No timeout provided, using 5 minutes")
		timeout = time.Minute * 5
	}

	return &GroupService{
		cfg:     cfg,
		ctx:     ctx,
		timeout: timeout,
	}, nil
}

func (s *GroupService) FindGroup(id string) (*Group, error) {
	// SQL return a group even if no users are in the group
	sql := `
		WITH groups_no_users
			 AS (SELECT groname,
					    grosysid,
					    NULL AS usename
				   FROM pg_group),
			groups_with_users
			AS (SELECT pg.grosysid,
					   pu.usename
				  FROM pg_user pu
					   LEFT JOIN pg_group pg
							  ON pu.usesysid = ANY ( pg.grolist ))
		SELECT t1.groname,
			   t1.grosysid,
			   t2.usename
		  FROM groups_no_users t1
			   LEFT JOIN groups_with_users t2
					  ON t1.grosysid = t2.grosysid
		 WHERE t1.grosysid = @GroupId
		 ORDER BY t1.groname,
				  t2.usename 
	`
	args := pgx.NamedArgs{"GroupId": id}

	ctx, cancel := context.WithTimeout(s.ctx, s.timeout)
	defer cancel()

	conn, err := pgx.ConnectConfig(ctx, s.cfg)
	if err != nil {
		return nil, fmt.Errorf("FindGroup: Unable to connect %w", err)
	}

	opts := pgx.TxOptions{
		BeginQuery: "SET enable_case_sensitive_identifier TO true",
	}
	tx, err := conn.BeginTx(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("FindGroup: Failed to begin transaction: %w", err)
	}

	group, err := buildGroup(sql, args, ctx, tx)
	if err != nil {
		return nil, fmt.Errorf("FindGroup: Failed to build group: %w", err)
	}
	if group == nil {
		return nil, fmt.Errorf("FindGroup: Could not find group with id '%s'", id)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("FindGroup: Failed to commit: %w", err)
	}

	return group, nil
}

func (s *GroupService) DropGroup(name string) error {
	sql := fmt.Sprintf("DROP GROUP %s", pgx.Identifier{name}.Sanitize())

	ctx, cancel := context.WithTimeout(s.ctx, s.timeout)
	defer cancel()

	conn, err := pgx.ConnectConfig(ctx, s.cfg)
	if err != nil {
		return fmt.Errorf("DropGroup: Unable to connect %w", err)
	}

	opts := pgx.TxOptions{
		BeginQuery: "SET enable_case_sensitive_identifier TO true",
	}
	tx, err := conn.BeginTx(ctx, opts)
	if err != nil {
		return fmt.Errorf("DropGroup: Failed to begin transaction: %w", err)
	}

	_, err = tx.Exec(ctx, sql)
	if err != nil {
		return fmt.Errorf("DropGroup: Failed to execute: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("DropGroup: Failed to commit: %w", err)
	}

	return nil
}

type CreateGroupDDLParams struct {
	Name      string
	Usernames *[]string
}

func (s *GroupService) CreateGroup(args CreateGroupDDLParams) (*Group, error) {
	t := `
		CREATE GROUP {{.Name}}
			{{if .Usernames}}{{$length := len .Usernames}}{{if gt $length 0 }}WITH USER {{(StringsJoin .Usernames ", ")}}{{end}}{{end}}
	`
	fmt.Printf("FFFFF %s", *args.Usernames)
	name := args.Name // save this unsanitized for lookup later
	args.Name = pgx.Identifier{args.Name}.Sanitize()

	ctx, cancel := context.WithTimeout(s.ctx, s.timeout)
	defer cancel()

	conn, err := pgx.ConnectConfig(ctx, s.cfg)
	if err != nil {
		return nil, fmt.Errorf("CreateGroup: Unable to connect %w", err)
	}

	opts := pgx.TxOptions{
		BeginQuery: "SET enable_case_sensitive_identifier TO true",
	}
	tx, err := conn.BeginTx(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("CreateGroup: Failed to begin transaction: %w", err)
	}

	sql, err := helpers.Merge(t, args)
	if err != nil {
		return nil, fmt.Errorf("CreateGroup: Failed to merge template: %w", err)
	}

	_, err = tx.Exec(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("CreateGroup: Failed to execute: %w", err)
	}

	group, err := getGroupByName(name, ctx, tx)
	if err != nil {
		return nil, fmt.Errorf("CreateGroup: Failed to getGroupByName: %w", err)
	}
	if group == nil {
		return nil, fmt.Errorf("CreateGroup: Could not find group with name '%s'", name)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("CreateGroup: Failed to commit: %w", err)
	}

	return group, nil
}

type AlterGroupDDLParams struct {
	Name     string
	RenameTo *string
	Add      *[]string
	Drop     *[]string
}

func (s *GroupService) AlterGroup(args AlterGroupDDLParams) error {
	args.Name = pgx.Identifier{args.Name}.Sanitize()

	ctx, cancel := context.WithTimeout(s.ctx, s.timeout)
	defer cancel()

	conn, err := pgx.ConnectConfig(ctx, s.cfg)
	if err != nil {
		return fmt.Errorf("AlterGroup: Unable to connect %w", err)
	}

	opts := pgx.TxOptions{
		BeginQuery: "SET enable_case_sensitive_identifier TO true",
	}
	tx, err := conn.BeginTx(ctx, opts)
	if err != nil {
		return fmt.Errorf("AlterGroup: Failed to begin transaction: %w", err)
	}

	// redshift must perform add, drop and rename separately
	if args.Add != nil && len(*args.Add) > 0 {
		var adds []string
		for _, add := range *args.Add {
			username := pgx.Identifier{add}.Sanitize()
			adds = append(adds, username)
		}
		args.Add = &adds

		rename := `
		ALTER GROUP {{.Name}}
			ADD USER  {{(StringsJoin .Add ", ")}}
		`
		sql, err := helpers.Merge(rename, args)
		if err != nil {
			return fmt.Errorf("AlterGroup: failed to merge adduser template: %w", err)
		}

		_, err = tx.Exec(ctx, sql)
		if err != nil {
			return fmt.Errorf("AlterGroup: failed to execute adduser: %w", err)
		}
	}

	if args.Drop != nil && len(*args.Drop) > 0 {
		var drops []string
		for _, drop := range *args.Drop {
			username := pgx.Identifier{drop}.Sanitize()
			drops = append(drops, username)
		}
		args.Drop = &drops

		rename := `
		ALTER GROUP {{.Name}}
			DROP USER  {{(StringsJoin .Drop ", ")}}
		`
		sql, err := helpers.Merge(rename, args)
		if err != nil {
			return fmt.Errorf("AlterGroup: failed to merge dropusers template: %w", err)
		}

		_, err = tx.Exec(ctx, sql)
		if err != nil {
			return fmt.Errorf("AlterGroup: failed to execute dropuser: %w", err)
		}
	}

	if args.RenameTo != nil {
		rn := pgx.Identifier{*args.RenameTo}.Sanitize()
		args.RenameTo = &rn

		rename := `
			ALTER GROUP {{.Name}} RENAME TO {{.RenameTo}}
		`
		sql, err := helpers.Merge(rename, args)
		if err != nil {
			return fmt.Errorf("AlterGroup: failed to merge rename template: %w", err)
		}

		_, err = tx.Exec(ctx, sql)
		if err != nil {
			return fmt.Errorf("AlterGroup: failed to execute rename: %w", err)
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("AlterGroup: Failed to commit: %w", err)
	}

	return nil
}

// hidden from outside the package, expect that callers use the ById variant.
func getGroupByName(name string, ctx context.Context, tx pgx.Tx) (*Group, error) {
	// SQL return a group even if no users are in the group
	sql := `
		WITH groups_no_users
			 AS (SELECT groname,
					    grosysid,
					    NULL AS usename
				   FROM pg_group),
			groups_with_users
			AS (SELECT pg.grosysid,
					   pu.usename
				  FROM pg_user pu
					   LEFT JOIN pg_group pg
							  ON pu.usesysid = ANY ( pg.grolist ))
		SELECT t1.groname,
			   t1.grosysid,
			   t2.usename
		  FROM groups_no_users t1
			   LEFT JOIN groups_with_users t2
					  ON t1.grosysid = t2.grosysid
		 WHERE t1.groname = @GroupName
		 ORDER BY t1.groname,
				  t2.usename 
	`
	args := pgx.NamedArgs{"GroupName": name}

	return buildGroup(sql, args, ctx, tx)
}

func buildGroup(sql string, args pgx.NamedArgs, ctx context.Context, tx pgx.Tx) (*Group, error) {
	rows, err := tx.Query(ctx, sql, args)
	if err != nil {
		return nil, fmt.Errorf("buildGroup: Failed query execute: %w", err)
	}

	pg_groups, err := pgx.CollectRows(rows, pgx.RowToStructByName[pg_group])
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}

		return nil, fmt.Errorf("buildGroup: Failed to collect row: %w", err)
	}

	if len(pg_groups) == 0 {
		return nil, nil
	}

	usernames := []string{}
	group := Group{
		Id:    pg_groups[0].Id,
		Name:  pg_groups[0].GroupName,
		Users: &usernames,
	}
	for _, group := range pg_groups {
		// if there are no users in group, this will be nil
		if group.Username != nil {
			usernames = append(usernames, *group.Username)
		}
	}

	return &group, nil
}
