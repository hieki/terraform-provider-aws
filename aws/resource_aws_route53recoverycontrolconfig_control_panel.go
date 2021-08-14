package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53recoverycontrolconfig"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/route53recoverycontrolconfig/waiter"
)

func resourceAwsRoute53RecoveryControlConfigControlPanel() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsRoute53RecoveryControlConfigControlPanelCreate,
		Read:   resourceAwsRoute53RecoveryControlConfigControlPanelRead,
		Update: resourceAwsRoute53RecoveryControlConfigControlPanelUpdate,
		Delete: resourceAwsRoute53RecoveryControlConfigControlPanelDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"control_panel_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cluster_arn": {
				Type:     schema.TypeString,
				Required: true,
			},
			"default_control_panel": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"routing_control_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsRoute53RecoveryControlConfigControlPanelCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).route53recoverycontrolconfigconn

	input := &route53recoverycontrolconfig.CreateControlPanelInput{
		ClientToken:      aws.String(resource.UniqueId()),
		ClusterArn:       aws.String(d.Get("cluster_arn").(string)),
		ControlPanelName: aws.String(d.Get("name").(string)),
	}

	output, err := conn.CreateControlPanel(input)
	result := output.ControlPanel

	if err != nil {
		return fmt.Errorf("Error creating Route53 Recovery Control Config Control Panel: %w", err)
	}

	if result == nil {
		return fmt.Errorf("Error creating Route53 Recovery Control Config Control Panel empty response")
	}

	d.SetId(aws.StringValue(result.ControlPanelArn))

	if _, err := waiter.Route53RecoveryControlConfigControlPanelCreated(conn, d.Id()); err != nil {
		return fmt.Errorf("Error waiting for Route53 Recovery Control Config Control Panel (%s) to be Deployed: %w", d.Id(), err)
	}

	return resourceAwsRoute53RecoveryControlConfigControlPanelRead(d, meta)
}

func resourceAwsRoute53RecoveryControlConfigControlPanelRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).route53recoverycontrolconfigconn

	input := &route53recoverycontrolconfig.DescribeControlPanelInput{
		ControlPanelArn: aws.String(d.Id()),
	}

	output, err := conn.DescribeControlPanel(input)
	result := output.ControlPanel

	if err != nil {
		return fmt.Errorf("Error describing Route53 Recovery Control Config Control Panel: %s", err)
	}

	if !d.IsNewResource() && result == nil {
		log.Printf("[WARN] Route53 Recovery Control Config Control Panel (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("control_panel_arn", result.ControlPanelArn)
	d.Set("cluster_arn", result.ClusterArn)
	d.Set("default_control_panel", result.DefaultControlPanel)
	d.Set("routing_control_count", result.RoutingControlCount)
	d.Set("name", result.Name)
	d.Set("status", result.Status)

	return nil
}

func resourceAwsRoute53RecoveryControlConfigControlPanelUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).route53recoverycontrolconfigconn

	input := &route53recoverycontrolconfig.UpdateControlPanelInput{
		ControlPanelName: aws.String(d.Get("name").(string)),
		ControlPanelArn:  aws.String(d.Get("control_panel_arn").(string)),
	}

	_, err := conn.UpdateControlPanel(input)
	if err != nil {
		return fmt.Errorf("error updating Route53 Recovery Control Config Control Panel: %s", err)
	}

	return resourceAwsRoute53RecoveryControlConfigControlPanelRead(d, meta)
}

func resourceAwsRoute53RecoveryControlConfigControlPanelDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).route53recoverycontrolconfigconn

	input := &route53recoverycontrolconfig.DeleteControlPanelInput{
		ControlPanelArn: aws.String(d.Id()),
	}

	_, err := conn.DeleteControlPanel(input)

	if err != nil {
		if isAWSErr(err, route53recoverycontrolconfig.ErrCodeResourceNotFoundException, "") {
			return nil
		}
		return fmt.Errorf("error deleting Route53 Recovery Control Config Control Panel: %s", err)
	}

	if _, err := waiter.Route53RecoveryControlConfigControlPanelDeleted(conn, d.Id()); err != nil {
		if isResourceNotFoundError(err) {
			return nil
		}
		return fmt.Errorf("Error waiting for Route53 Recovery Control Config Control Panel (%s) to be deleted: %w", d.Id(), err)
	}

	return nil
}
