terraform {
  required_providers {
    redshift = {
      source = "registry.terraform.io/donaldjarmstrong/redshift"
    }
  }
}

provider "redshift" {
  # host     = var.host
  # port     = var.port
  # username = var.username
  password = var.password
  dbname   = var.dbname
  sslmode  = var.sslmode
  timeout  = 30
}

# resource "redshift_user" "donnie1" {
#   name        = "donnie1234"
#   # password    = "FredGarvin12"
#     session_timeout  = 623
#   valid_until = "2024-10-01 12:00:00"
#   # valid_until = "infinity"
#   createdb    = true
#   createuser = false
#   syslog_access = "UNRESTRICTED"
# }

# resource "redshift_user" "donnie2" {
#   name     = "donnie2"
#   password = null
# }

resource "redshift_user" "donnie3" {
  name     = "donnie32"
  password = "Password6"
}

resource "redshift_user" "donnie4" {
  name     = "donnie457"
  password = "Password567"
}

resource "redshift_user" "donnie5" {
  name             = "IAM:donnie5"
  password         = "md5|Password5"
  createdb         = true
  createuser       = true
  syslog_access    = "UNRESTRICTED"
  connection_limit = "3"
  valid_until      = "2037-01-19 03:14:04"
  session_timeout  = 61
  # external_id      = "dome_id"
}

# resource "redshift_role" "role1" {
#   name = "someRole1"
# }

# resource "redshift_role" "role2" {
#   name = "someRole22"
# }

resource "redshift_group" "group1" {
  name = "someGroup1"
  usernames = []
}

resource "redshift_group" "group2" {
  name = "someGroup2"
  # usernames = [ redshift_user.donnie5.name, redshift_user.donnie4.name ]
  usernames = []
}