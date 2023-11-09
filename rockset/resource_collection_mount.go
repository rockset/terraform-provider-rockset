package rockset

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/rockset/rockset-go-client"
	"github.com/rockset/rockset-go-client/openapi"
	"regexp"
	"strings"
)

func resourceCollectionMount() *schema.Resource {
	return &schema.Resource{
		Description: `Manages a collection mount.`,

		CreateContext: resourceCollectionMountCreate,
		ReadContext:   resourceCollectionMountRead,
		DeleteContext: resourceCollectionMountDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "Unique ID of this mount, collection path and virtual instance id, joined by a `:`.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"rrn": {
				Description: "RRN of this mount.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"virtual_instance_id": {
				Description: "Virtual Instance id",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"virtual_instance_rrn": {
				Description: "Virtual Instance RRN",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"created_at": {
				Description: "ISO 8601 date when the mount was created.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"last_refresh_time": {
				Description: "UNIX timestamp in milliseconds for most recent refresh. Not applicable for live mounts.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"snapshot_expiration_time": {
				Description: "UNIX timestamp in milliseconds when the snapshot expires.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"state": {
				Description: "Mount state.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"path": {
				Description: "Collection path to be mounted, in the form workspace.collection",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				ValidateFunc: validation.StringMatch(
					pathRegexp,
					"must be of the form workspace.collection"),
			},
		},
	}
}

var pathRegexp = regexp.MustCompile(fmt.Sprintf(`^%s\.%s$`, nameRe, nameRe))

func resourceCollectionMountCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics

	vid := d.Get("virtual_instance_id").(string)
	path := d.Get("path").(string)

	mounts, err := rc.MountCollections(ctx, vid, []string{path})
	if err != nil {
		return diag.FromErr(err)
	}
	if len(mounts) != 1 {
		return diag.Errorf("expected exactly one mount response, got %d", len(mounts))
	}
	tflog.Info(ctx, "mounted")

	id := mountToID(path, vid)
	d.SetId(id)

	fields := strings.SplitN(path, ".", 2)
	workspace := fields[0]
	collection := fields[1]

	// TODO make it possible to skip waiting, and then parse the fields from the created vi
	err = rc.Wait.UntilMountActive(ctx, vid, workspace, collection)
	if err != nil {
		return diag.FromErr(err)
	}

	// get the vi info, so we have updated value for current_size
	m, err := rc.GetCollectionMount(ctx, vid, path)
	if err != nil {
		return diag.FromErr(err)
	}
	if err = parseCollectionMountFields(m, d); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceCollectionMountRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics

	path, vid, err := idToMount(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	m, err := rc.GetCollectionMount(ctx, vid, path)
	if err != nil {
		return diag.FromErr(err)
	}

	if err = parseCollectionMountFields(m, d); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceCollectionMountDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rc := meta.(*rockset.RockClient)
	var diags diag.Diagnostics

	path, vid, err := idToMount(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	fields := strings.Split(path, ".")
	if len(fields) != 2 {
		return diag.Errorf("path couldn't be split into workspace and collection: %s", path)
	}

	_, err = rc.UnmountCollection(ctx, vid, path)
	if err != nil {
		return diag.FromErr(err)
	}

	err = rc.Wait.UntilMountGone(ctx, vid, fields[0], fields[1])
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func parseCollectionMountFields(m openapi.CollectionMount, d *schema.ResourceData) error {
	if err := setValue(d, "id", m.GetIdOk); err != nil {
		return err
	}
	if err := setValue(d, "rrn", m.GetRrnOk); err != nil {
		return err
	}
	if err := setValue(d, "virtual_instance_id", m.GetVirtualInstanceIdOk); err != nil {
		return err
	}
	if err := setValue(d, "virtual_instance_rrn", m.GetVirtualInstanceRrnOk); err != nil {
		return err
	}
	if err := setValue(d, "created_at", m.GetCreatedAtOk); err != nil {
		return err
	}
	if err := setValue(d, "last_refresh_time", m.GetLastRefreshTimeMillisOk); err != nil {
		return err
	}
	if err := setValue(d, "snapshot_expiration_time", m.GetSnapshotExpirationTimeMillisOk); err != nil {
		return err
	}
	if err := setValue(d, "state", m.GetStateOk); err != nil {
		return err
	}
	if err := setValue(d, "path", m.GetCollectionPathOk); err != nil {
		return err
	}

	return nil
}
