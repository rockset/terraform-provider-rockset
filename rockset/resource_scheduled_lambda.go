package rockset

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/rockset/rockset-go-client"
	"github.com/rockset/rockset-go-client/openapi"
	"github.com/rockset/rockset-go-client/option"
)

func resourceScheduledLambda() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a Rockset scheduled lambda, a query lambda that is automatically executed on a schedule.",

		CreateContext: resourceScheduledLambdaCreate,
		ReadContext:   resourceScheduledLambdaRead,
		UpdateContext: resourceScheduledLambdaUpdate,
		DeleteContext: resourceScheduledLambdaDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"rrn": {
				Description: "RRN of this Scheduled Lambda.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"workspace": {
				Description:  "Workspace name.",
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: rocksetNameValidator,
			},
			"apikey": {
				Description:  "The apikey to use when triggering execution of the associated query lambda.",
				Type:         schema.TypeString,
				Required: true,
				Sensitive: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// apikey is not returned for security reasons
					return new == ""
				},
			},
			"cron_string": {
				Description: "The UNIX-formatted cron string for this scheduled query lambda.",
				Type:        schema.TypeString,
				ForceNew:     true,
				Required:    true,
			},
			"ql_name": {
				Description: "The name of the QL to use for scheduled execution.",
				Type:        schema.TypeString,
				ForceNew:     true,
				Required:    true,
			},
			"tag": {
				Description:  "The QL tag to use for scheduled execution.",
				Type:         schema.TypeString,
				ForceNew:     true,
				Optional: true,
			},
			"version": {
				Description:  "The version of the QL to use for scheduled execution.",
				Type:         schema.TypeString,
				ForceNew:     true,
				Optional: true,
			},
			"total_times_to_execute": {
				Description:  "The number of times to execute this scheduled query lambda. Once this scheduled query lambda has been executed this many times, it will no longer be executed.",
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntAtLeast(0),
			},
			"webhook_auth_header": {
				Description:  "The value to use as the authorization header when hitting the webhook.",
				Type:         schema.TypeString,
				Optional: true,
				Sensitive: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// auth header is not returned for security reasons
					return new == ""
				},
			},
			"webhook_payload": {
				Description:  "The payload that should be sent to the webhook. JSON format.",
				Type:         schema.TypeString,
				Optional: true,
			},
			"webhook_url": {
				Description:  "The URL of the webhook that should be triggered after this scheduled query lambda completes.",
				Type:         schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceScheduledLambdaCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics

	workspace := d.Get("workspace").(string)
	apikey := d.Get("apikey").(string)
	cronString := d.Get("cron_string").(string)
	qlName := d.Get("ql_name").(string)
	
	options := getScheduledLambdaOptions(d)

	scheduledLambda, err := rc.CreateScheduledLambda(ctx, workspace, apikey, cronString, qlName, options...)
	if err != nil {
		return DiagFromErr(err)
	}

	scheduledLambdaRrn := *scheduledLambda.Rrn
	d.SetId(toID(workspace, scheduledLambdaRrn))

	err = rc.Wait.UntilScheduledLambdaAvailable(ctx, workspace, scheduledLambdaRrn)
	if err != nil {
		return DiagFromErr(err)
	}

	if err = parseScheduledLambdaFields(scheduledLambda, d); err != nil {
		return DiagFromErr(err)
	}

	return diags
}

func resourceScheduledLambdaUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics

	workspace, scheduledLambdaRRN := workspaceAndNameFromID(d.Id())

	options := getScheduledLambdaOptions(d)
	_, err := rc.UpdateScheduledLambda(ctx, workspace, scheduledLambdaRRN, options...)
	if err != nil {
		return DiagFromErr(err)
	}

	return diags
}

func resourceScheduledLambdaRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics

	workspace, scheduledLambdaRRN := workspaceAndNameFromID(d.Id())

	scheduledLambda, err := rc.GetScheduledLambda(ctx, workspace, scheduledLambdaRRN)
	if err != nil {
		return DiagFromErr(err)
	}

	if err = parseScheduledLambdaFields(scheduledLambda, d); err != nil {
		return DiagFromErr(err)
	}

	return diags
}

func resourceScheduledLambdaDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics

	workspace, scheduledLambdaRRN := workspaceAndNameFromID(d.Id())

	err := rc.DeleteScheduledLambda(ctx, workspace, scheduledLambdaRRN)
	if err != nil {
		return DiagFromErr(err)
	}

	err = rc.Wait.UntilScheduledLambdaGone(ctx, workspace, scheduledLambdaRRN)
	if err != nil {
		return DiagFromErr(err)
	}

	return diags
}

func getScheduledLambdaOptions(d *schema.ResourceData) []option.ScheduledLambdaOption {
	var options []option.ScheduledLambdaOption
    addOptionIfChanged(d, "tag", &options, func(a any) option.ScheduledLambdaOption {
		return option.WithScheduledLambdaTag(a.(string))
	})
	addOptionIfChanged(d, "version", &options, func(a any) option.ScheduledLambdaOption {
		return option.WithScheduledLambdaVersion(a.(string))
	})
	addOptionIfChanged(d, "apikey", &options, func(a any) option.ScheduledLambdaOption {
		return option.WithScheduledLambdaApikey(a.(string))
	})
	addOptionIfChanged(d, "total_times_to_execute", &options, func(a any) option.ScheduledLambdaOption {
		return option.WithScheduledLambdaTotalTimesToExecute(int64(a.(int)))
	})
	addOptionIfChanged(d, "webhook_auth_header", &options, func(a any) option.ScheduledLambdaOption {
		return option.WithScheduledLambdaWebhookAuthHeader(a.(string))
	})
	addOptionIfChanged(d, "webhook_payload", &options, func(a any) option.ScheduledLambdaOption {
		return option.WithScheduledLambdaWebhookPayload(a.(string))
	})
	addOptionIfChanged(d, "webhook_url", &options, func(a any) option.ScheduledLambdaOption {
		return option.WithScheduledLambdaWebhookUrl(a.(string))
	})

	return options
}

func parseScheduledLambdaFields(scheduledLambda openapi.ScheduledLambda, d *schema.ResourceData) error {
	if err := setValue(d, "rrn", scheduledLambda.GetRrnOk); err != nil {
		return err
	}
	if err := setValue(d, "workspace", scheduledLambda.GetWorkspaceOk); err != nil {
		return err
	}
	if err := setValue(d, "cron_string", scheduledLambda.GetCronStringOk); err != nil {
		return err
	}
	if err := setValue(d, "ql_name", scheduledLambda.GetQlNameOk); err != nil {
		return err
	}
	if err := setValue(d, "tag", scheduledLambda.GetTagOk); err != nil {
		return err
	}
	if err := setValue(d, "version", scheduledLambda.GetVersionOk); err != nil {
		return err
	}
	if err := setValue(d, "total_times_to_execute", scheduledLambda.GetTotalTimesToExecuteOk); err != nil {
		return err
	}
	if err := setValue(d, "webhook_payload", scheduledLambda.GetWebhookPayloadOk); err != nil {
		return err
	}
	if err := setValue(d, "webhook_url", scheduledLambda.GetWebhookUrlOk); err != nil {
		return err
	}

	return nil
}
