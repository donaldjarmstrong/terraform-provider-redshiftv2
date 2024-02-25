package provider

import (
	"fmt"
	"strings"
	"terraform-provider-redshift/internal/helpers"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRole_basic(t *testing.T) {
	role1 := "tst-terraform" + strings.ToUpper(acctest.RandStringFromCharSet(10, helpers.CharSetAlpha))
	role2 := "tst-terraform" + strings.ToUpper(acctest.RandStringFromCharSet(10, helpers.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
				resource "redshift_role" "under_test" {
					name = "%s"
				}
				`, role1),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("redshift_role.under_test", "id"),
					resource.TestCheckResourceAttr("redshift_role.under_test", "name", role1),
					resource.TestCheckNoResourceAttr("redshift_role.under_test", "externalid"),
				),
			},
			{
				ResourceName:      "redshift_role.under_test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: providerConfig + fmt.Sprintf(`
				resource "redshift_role" "under_test" {
					name = "%s"
				}
				`, role2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redshift_role.under_test", "name", role2),
				),
			},
		},
	})
}
