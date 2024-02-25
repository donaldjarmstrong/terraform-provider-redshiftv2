package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

var (

	// testAccProtoV6ProviderFactories are used to instantiate a provider during
	// acceptance testing. The factory function will be invoked for every Terraform
	// CLI command executed to create a provider server to which the CLI can
	// reattach.
	testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
		"redshift": providerserver.NewProtocol6WithError(New("test")()),
	}
	providerConfig = `
		variable "host" {}
		variable "port" {}
		variable "username" {}
		variable "password" {}
		variable "dbname" {}
		variable "sslmode" {
			type    = string
			default = "require"
		}
		variable "timeout" {
			type    = number
			default = 60
		}

		provider "redshift" {
			host     = var.host
			port     = var.port
			username = var.username
			password = var.password
			dbname   = var.dbname
			sslmode  = var.sslmode
			timeout  = var.timeout
		}
	`
)

func testAccPreCheck(t *testing.T) {
	// You can add code here to run prior to any test case execution, for example assertions
	// about the appropriate environment variables being set are common to see in a pre-check
	// function.
	if _, found := os.LookupEnv("TF_VAR_host"); !found {
		t.Fatal("TF_VAR_host must be set for acceptance tests")
	}

	if _, found := os.LookupEnv("TF_VAR_port"); !found {
		t.Fatal("TF_VAR_port must be set for acceptance tests")
	}

	if _, found := os.LookupEnv("TF_VAR_username"); !found {
		t.Fatal("TF_VAR_username must be set for acceptance tests")
	}

	if _, found := os.LookupEnv("TF_VAR_password"); !found {
		t.Fatal("TF_VAR_password must be set for acceptance tests")
	}

	if _, found := os.LookupEnv("TF_VAR_dbname"); !found {
		t.Fatal("TF_VAR_dbname must be set for acceptance tests")
	}
}

func TestProvider_Config(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				ResourceName: "Test Config Provider",
				Config: providerConfig + `
				resource redshift_user provider_test {
					name = "provider_test"
				}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redshift_user.provider_test", "name", "provider_test"),
				),
			},
		},
	})
}

func TestProvider_impl(t *testing.T) {
	_ = New("1")
}
