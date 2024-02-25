package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRole_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
				resource "redshift_role" "test" {
					name = "test"
				}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redshift_role.test", "name", "test"),
					resource.TestCheckNoResourceAttr("redshift_role.test", "externalid"),
					resource.TestCheckResourceAttrSet("redshift_role.test", "id"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "redshift_role.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: providerConfig + `
				resource "redshift_role" "test" {
					name = "renamed"
				}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redshift_role.test", "name", "renamed"),
				),
			},
		},
	})
}
