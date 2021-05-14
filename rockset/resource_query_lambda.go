package rockset

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceQueryLambda() *schema.Resource {
	return &schema.Resource{
		Description: "Sample resource in the Terraform provider QueryLambda.",

		CreateContext: resourceQueryLambdaCreate,
		ReadContext:   resourceQueryLambdaRead,
		UpdateContext: resourceQueryLambdaUpdate,
		DeleteContext: resourceQueryLambdaDelete,

		Schema: map[string]*schema.Schema{
			"sample_attribute": {
				Description: "Sample attribute.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"last_updated": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourceQueryLambdaCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	//rc := meta.(*rockset.RockClient)
	//var diags diag.Diagnostics

	idFromAPI := "my-id"
	d.SetId(idFromAPI)

	return diag.Errorf("not implemented")
}

func resourceQueryLambdaRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	//rc := meta.(*rockset.RockClient)
	//var diags diag.Diagnostics

	return diag.Errorf("not implemented")
}

func resourceQueryLambdaUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	//rc := meta.(*rockset.RockClient)
	//var diags diag.Diagnostics

	return diag.Errorf("not implemented")
}

func resourceQueryLambdaDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	//rc := meta.(*rockset.RockClient)
	//var diags diag.Diagnostics

	return diag.Errorf("not implemented")
}
