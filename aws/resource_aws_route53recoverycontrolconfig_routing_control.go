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

func resourceAwsRoute53RecoveryControlConfigRoutingControl() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsRoute53RecoveryControlConfigRoutingControlCreate,
		Read:   resourceAwsRoute53RecoveryControlConfigRoutingControlRead,
		Update: resourceAwsRoute53RecoveryControlConfigRoutingControlUpdate,
		Delete: resourceAwsRoute53RecoveryControlConfigRoutingControlDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"routing_control_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cluster_arn": {
				Type:     schema.TypeString,
				Required: true,
			},
			"control_panel_arn": {
				Type:     schema.TypeString,
				Optional: true,
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

func resourceAwsRoute53RecoveryControlConfigRoutingControlCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).route53recoverycontrolconfigconn

	input := &route53recoverycontrolconfig.CreateRoutingControlInput{
		ClientToken:        aws.String(resource.UniqueId()),
		ClusterArn:         aws.String(d.Get("cluster_arn").(string)),
		RoutingControlName: aws.String(d.Get("name").(string)),
	}

	if v, ok := d.GetOk("control_panel_arn"); ok {
		input.ControlPanelArn = aws.String(v.(string))
	}

	output, err := conn.CreateRoutingControl(input)
	result := output.RoutingControl

	if err != nil {
		return fmt.Errorf("Error creating Route53 Recovery Control Config Routing Control: %w", err)
	}

	if result == nil {
		return fmt.Errorf("Error creating Route53 Recovery Control Config Routing Control empty response")
	}

	d.SetId(aws.StringValue(result.RoutingControlArn))

	if _, err := waiter.Route53RecoveryControlConfigRoutingControlCreated(conn, d.Id()); err != nil {
		return fmt.Errorf("Error waiting for Route53 Recovery Control Config Routing Control (%s) to be Deployed: %w", d.Id(), err)
	}

	return resourceAwsRoute53RecoveryControlConfigRoutingControlRead(d, meta)
}

func resourceAwsRoute53RecoveryControlConfigRoutingControlRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).route53recoverycontrolconfigconn

	input := &route53recoverycontrolconfig.DescribeRoutingControlInput{
		RoutingControlArn: aws.String(d.Id()),
	}

	output, err := conn.DescribeRoutingControl(input)
	result := output.RoutingControl

	if err != nil {
		return fmt.Errorf("Error describing Route53 Recovery Control Config Routing Control: %s", err)
	}

	if !d.IsNewResource() && result == nil {
		log.Printf("[WARN] Route53 Recovery Control Config Routing Control (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("routing_control_arn", result.RoutingControlArn)
	d.Set("control_panel_arn", result.ControlPanelArn)
	d.Set("name", result.Name)
	d.Set("status", result.Status)

	return nil
}

func resourceAwsRoute53RecoveryControlConfigRoutingControlUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).route53recoverycontrolconfigconn

	input := &route53recoverycontrolconfig.UpdateRoutingControlInput{
		RoutingControlName: aws.String(d.Get("name").(string)),
		RoutingControlArn:  aws.String(d.Get("routing_control_arn").(string)),
	}

	_, err := conn.UpdateRoutingControl(input)
	if err != nil {
		return fmt.Errorf("error updating Route53 Recovery Control Config Routing Control: %s", err)
	}

	return resourceAwsRoute53RecoveryControlConfigRoutingControlRead(d, meta)
}

func resourceAwsRoute53RecoveryControlConfigRoutingControlDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).route53recoverycontrolconfigconn

	input := &route53recoverycontrolconfig.DeleteRoutingControlInput{
		RoutingControlArn: aws.String(d.Id()),
	}

	_, err := conn.DeleteRoutingControl(input)

	if err != nil {
		if isAWSErr(err, route53recoverycontrolconfig.ErrCodeResourceNotFoundException, "") {
			return nil
		}
		return fmt.Errorf("error deleting Route53 Recovery Control Config Routing Control: %s", err)
	}

	if _, err := waiter.Route53RecoveryControlConfigRoutingControlDeleted(conn, d.Id()); err != nil {
		if isResourceNotFoundError(err) {
			return nil
		}
		return fmt.Errorf("Error waiting for Route53 Recovery Control Config  Routing Control (%s) to be deleted: %w", d.Id(), err)
	}

	return nil
}
