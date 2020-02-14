package rockset

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/rockset/rockset-go-client"
)

func dataSourceRocksetAccount() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceReadRocksetAccount,

		Schema: map[string]*schema.Schema{
			//"external_id": &schema.Schema{
			//	Type:     schema.TypeString,
			//	Computed: true,
			//},
			"account_id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		}}
}

// TODO: this is hardcoded as we don't have the API endpoint to pull this info yet
const (
	accountID  = "318212636800"
)

func dataSourceReadRocksetAccount(d *schema.ResourceData, m interface{}) error {
	_ = m.(*rockset.RockClient)
	d.SetId(accountID)

	err := d.Set("account_id", accountID)
	if err != nil {
		return err
	}

	//err = d.Set("external_id", externalID)
	//if err != nil {
	//	return err
	//}

	return nil
}
