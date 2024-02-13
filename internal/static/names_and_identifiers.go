package static

import "regexp"

/*
Names identify database objects, including tables and columns, as well as users and passwords.
The terms name and identifier can be used interchangeably. There are two types of identifiers,
standard identifiers and quoted or delimited identifiers. Identifiers must consist of only UTF-8
printable characters. ASCII letters in standard and delimited identifiers are case-insensitive
and are folded to lowercase in the database. In query results, column names are returned as
lowercase by default.

See  https://docs.aws.amazon.com/redshift/latest/dg/r_names.html
*/

// The following PostgreSQL system column names can't be used as column names in user-defined columns.
// For more information, see https://www.postgresql.org/docs/8.0/static/ddl-system-columns.html.
var SystemColumnNames = []string{
	`oid`,
	"tableoid",
	"xmin",
	"cmin",
	"xmax",
	"cmax",
	"ctid",
}

// Begin with an ASCII single-byte alphabetic character or underscore character.
// Subsequent characters can be ASCII single-byte alphanumeric characters, underscores, or dollar signs
var IdentifierValidCharacters = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_\$]*$`)

// The following is a list of Amazon Redshift reserved words. You can use the reserved words with delimited
// identifiers (double quotation marks).  See https://docs.aws.amazon.com/redshift/latest/dg/r_pg_keywords.html
var ReservedWords = []string{
	"aes128",
	"aes256",
	"all",
	"allowoverwrite",
	"analyse",
	"analyze",
	"and",
	"any",
	"array",
	"as",
	"asc",
	"authorization",
	"az64",
	"backup",
	"between",
	"binary",
	"blanksasnull",
	"both",
	"bytedict",
	"bzip2",
	"case",
	"cast",
	"check",
	"collate",
	"column",
	"connect",
	"constraint",
	"create",
	"credentials",
	"cross",
	"current_date",
	"current_time",
	"current_timestamp",
	"current_user",
	"current_user_id",
	"default",
	"deferrable",
	"deflate",
	"defrag",
	"delta",
	"delta32k",
	"desc",
	"disable",
	"distinct",
	"do",
	"else",
	"emptyasnull",
	"enable",
	"encode",
	"encrypt",
	"encryption",
	"end",
	"except",
	"explicit",
	"false",
	"for",
	"foreign",
	"freeze",
	"from",
	"full",
	"globaldict256",
	"globaldict64k",
	"grant",
	"group",
	"gzip",
	"having",
	"identity",
	"ignore",
	"ilike",
	"in",
	"initially",
	"inner",
	"intersect",
	"interval",
	"into",
	"is",
	"isnull",
	"join",
	"leading",
	"left",
	"like",
	"limit",
	"localtime",
	"localtimestamp",
	"lun",
	"luns",
	"lzo",
	"lzop",
	"minus",
	"mostly16",
	"mostly32",
	"mostly8",
	"natural",
	"new",
	"not",
	"notnull",
	"null",
	"nulls",
	"off",
	"offline",
	"offset",
	"oid",
	"old",
	"on",
	"only",
	"open",
	"or",
	"order",
	"outer",
	"overlaps",
	"parallel",
	"partition",
	"percent",
	"permissions",
	"pivot",
	"placing",
	"primary",
	"raw",
	"readratio",
	"recover",
	"references",
	"rejectlog",
	"resort",
	"respect",
	"restore",
	"right",
	"select",
	"session_user",
	"similar",
	"snapshot ",
	"some",
	"sysdate",
	"system",
	"start",
	"table",
	"tag",
	"tdes",
	"text255",
	"text32k",
	"then",
	"timestamp",
	"to",
	"top",
	"trailing",
	"true",
	"truncatecolumns",
	"union",
	"unique",
	"unnest",
	"unpivot",
	"user",
	"using",
	"verbose",
	"wallet",
	"when",
	"where",
	"with",
	"without",
}

// It must contain at least one uppercase letter, one lowercase letter, and one number
// var

// (?=.*\d)(?=.*[a-z])(?=.*[A-Z])((?=.*\W)|(?=.*_))^[^ ]+$
