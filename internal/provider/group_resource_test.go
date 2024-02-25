package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccGroup_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
				// resource "redshift_user" "test1" {
				// 	name "test1"
				// }
				// resource "redshift_user" "test2" {
				// 	name "test2"
				// }
				resource "redshift_group" "test" {
					name      = "test"
					usernames = []
				}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("redshift_group.test", "id"),
					resource.TestCheckResourceAttr("redshift_group.test", "name", "test"),
					resource.TestCheckResourceAttr("redshift_group.test", "usernames.#", "0"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "redshift_group.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: providerConfig + `
				resource "redshift_group" "test" {
					name      = "renamed"
					usernames = []
				}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redshift_group.test", "name", "renamed"),
				),
			},
		},
	})
}

func TestAccGroup_users(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
				resource "redshift_user" "test1" {
					name = "test1"
				}
				resource "redshift_user" "test2" {
					name = "test2"
				}
				resource "redshift_group" "test" {
					name      = "test"
					usernames = [ redshift_user.test1.name, redshift_user.test2.name ]
				}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("redshift_group.test", "id"),
					resource.TestCheckResourceAttr("redshift_group.test", "name", "test"),
					resource.TestCheckResourceAttr("redshift_group.test", "usernames.#", "2"),
					resource.TestCheckResourceAttr("redshift_group.test", "usernames.0", "test1"),
					resource.TestCheckResourceAttr("redshift_group.test", "usernames.1", "test2"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "redshift_group.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: providerConfig + `
				resource "redshift_user" "test1" {
					name = "test1"
				}				
				resource "redshift_user" "test2" {
					name = "test2"
				}
				resource "redshift_user" "test3" {
					name = "test3"
				}
				resource "redshift_group" "test" {
					name      = "test"
					usernames = [ redshift_user.test3.name, redshift_user.test2.name ]
				}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redshift_group.test", "name", "test"),
					resource.TestCheckResourceAttr("redshift_group.test", "usernames.#", "2"),
					resource.TestCheckResourceAttr("redshift_group.test", "usernames.0", "test3"),
					resource.TestCheckResourceAttr("redshift_group.test", "usernames.1", "test2"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "redshift_group.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
