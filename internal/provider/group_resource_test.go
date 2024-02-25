package provider

import (
	"fmt"
	"strings"
	"terraform-provider-redshift/internal/helpers"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccGroup_basic(t *testing.T) {
	group1 := "tst_group1" + strings.ToUpper(acctest.RandStringFromCharSet(10, helpers.CharSetAlpha))
	group2 := "tst_group2" + strings.ToUpper(acctest.RandStringFromCharSet(10, helpers.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + fmt.Sprintf(`
				resource "redshift_group" "under_test" {
					name      = "%s"
					usernames = []
				}
				`, group1),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("redshift_group.under_test", "id"),
					resource.TestCheckResourceAttr("redshift_group.under_test", "name", group1),
					resource.TestCheckResourceAttr("redshift_group.under_test", "usernames.#", "0"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "redshift_group.under_test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: providerConfig + fmt.Sprintf(`
				resource "redshift_group" "under_test" {
					name      = "%s"
					usernames = []
				}
				`, group2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redshift_group.under_test", "name", group2),
				),
			},
		},
	})
}

func TestAccGroup_users(t *testing.T) {
	group1 := "tst_group1" + strings.ToUpper(acctest.RandStringFromCharSet(10, helpers.CharSetAlpha))
	user1 := "tst-user1" + strings.ToUpper(acctest.RandStringFromCharSet(10, helpers.CharSetAlpha))
	user2 := "tst-user2" + strings.ToUpper(acctest.RandStringFromCharSet(10, helpers.CharSetAlpha))
	user3 := "tst-user3" + strings.ToUpper(acctest.RandStringFromCharSet(10, helpers.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
				resource "redshift_user" "test1" {
					name = "%s"
				}
				resource "redshift_user" "test2" {
					name = "%s"
				}
				resource "redshift_group" "under_test" {
					name      = "%s"
					usernames = [ redshift_user.test1.name, redshift_user.test2.name ]
				}
				`, user1, user2, group1),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("redshift_group.under_test", "id"),
					resource.TestCheckResourceAttr("redshift_group.under_test", "name", group1),
					resource.TestCheckResourceAttr("redshift_group.under_test", "usernames.#", "2"),
					resource.TestCheckResourceAttr("redshift_group.under_test", "usernames.0", user1),
					resource.TestCheckResourceAttr("redshift_group.under_test", "usernames.1", user2),
				),
			},
			{
				ResourceName:      "redshift_group.under_test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: providerConfig + fmt.Sprintf(`
				resource "redshift_user" "test1" {
					name = "%s"
				}				
				resource "redshift_user" "test2" {
					name = "%s"
				}
				resource "redshift_user" "test3" {
					name = "%s"
				}
				resource "redshift_group" "under_test" {
					name      = "%s"
					usernames = [ redshift_user.test3.name, redshift_user.test2.name ]
				}
				`, user1, user2, user3, group1),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redshift_group.under_test", "name", group1),
					resource.TestCheckResourceAttr("redshift_group.under_test", "usernames.#", "2"),
					resource.TestCheckResourceAttr("redshift_group.under_test", "usernames.0", user2),
					resource.TestCheckResourceAttr("redshift_group.under_test", "usernames.1", user3),
				),
			},
			{
				ResourceName:      "redshift_group.under_test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
