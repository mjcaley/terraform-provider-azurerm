package azurerm

import (
	"fmt"
	"log"

	"github.com/Azure/azure-sdk-for-go/services/web/mgmt/2018-02-01/web"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceArmWebSourceControls() *schema.Resource {
	return &schema.Resource{
		Create: resourceArmWebSourceControlsCreate,
		Read:   resourceArmWebSourceControlsRead,
		Update: resourceArmWebSourceControlsUpdate,
		Delete: resourceArmWebSourceControlsDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"resource_group_name": resourceGroupNameSchema(),

			"app_service_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateAppServiceName,
			},

			"repo_url": {
				Type:     schema.TypeString,
				Required: true,
			},

			"branch": {
				Type:     schema.TypeString,
				Default:  "master",
				Optional: true,
			},
		},
	}
}

func resourceArmWebSourceControlsCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).appServicesClient
	ctx := meta.(*ArmClient).StopContext

	log.Printf("[INFO] preparing arguments for AzureRM Web App source control creation.")

	resGroup := d.Get("resource_group_name").(string)
	appName := d.Get("app_service_name").(string)
	repoURL := d.Get("repo_url").(string)
	branch := d.Get("branch").(string)

	sourceControlsProperties := web.SiteSourceControl{
		SiteSourceControlProperties: &web.SiteSourceControlProperties{
			RepoURL: &repoURL,
			Branch:  &branch,
		},
	}

	createFuture, err := client.CreateOrUpdateSourceControl(ctx, resGroup, appName, sourceControlsProperties)
	if err != nil {
		return err
	}

	err = createFuture.WaitForCompletionRef(ctx, client.Client)
	if err != nil {
		return err
	}

	read, err := client.GetSourceControl(ctx, resGroup, appName)
	if err != nil {
		return err
	}

	d.SetId(*read.ID)

	return nil
}

func resourceArmWebSourceControlsRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).appServicesClient
	ctx := meta.(*ArmClient).StopContext

	id, err := parseAzureResourceID(d.Id())
	if err != nil {
		return err
	}

	resGroup := id.ResourceGroup
	appServiceName := id.Path["sites"]

	resp, err := client.GetSourceControl(ctx, resGroup, appServiceName)
	if err != nil {
		return fmt.Errorf("Error making Read request on AzureRM Web Source Controls Web Site %q in resource group %q", appServiceName, resGroup)
	}

	d.Set("resource_group_name", resGroup)
	d.Set("app_service_name", appServiceName)
	if resp.RepoURL != nil {
		d.Set("repo_url", *resp.RepoURL)
	}
	if resp.Branch != nil {
		d.Set("branch", *resp.Branch)
	}

	return nil
}

func resourceArmWebSourceControlsDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).appServicesClient

	id, Iderr := parseAzureResourceID(d.Id())
	if Iderr != nil {
		return Iderr
	}

	resGroup := id.ResourceGroup
	appServiceName := id.Path["sites"]

	log.Printf("[DEBUG] Deleting Source Controls on Web App %q (resource group %q)", appServiceName, resGroup)

	ctx := meta.(*ArmClient).StopContext
	_, err := client.DeleteSourceControl(ctx, resGroup, appServiceName)
	if err != nil {
		return err
	}

	return nil
}

func resourceArmWebSourceControlsUpdate(d *schema.ResourceData, meta interface{}) error {
	err := resourceArmWebSourceControlsDelete(d, meta)
	if err != nil {
		return err
	}
	return resourceArmWebSourceControlsCreate(d, meta)
}
