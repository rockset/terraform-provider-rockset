package rockset

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/rockset/rockset-go-client"
)

func dataSourceRocksetAccount() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceReadRocksetAccount,

		Schema: map[string]*schema.Schema{
			"external_id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
				Sensitive: true,
			},
			"account_id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		}}
}

const accountID = "318212636800"

func dataSourceReadRocksetAccount(d *schema.ResourceData, m interface{}) error {
	rc := m.(*rockset.RockClient)

	err := d.Set("account_id", accountID)
	if err != nil {
		return err
	}

	org, _, err := rc.Organization()
	if err != nil {
		return err
	}

	err = d.Set("external_id", org.ExternalId)
	if err != nil {
		return err
	}

	d.SetId(accountID)

	return nil
}
