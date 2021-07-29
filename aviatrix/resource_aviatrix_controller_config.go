package aviatrix

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/AviatrixSystems/terraform-provider-aviatrix/v2/goaviatrix"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const defaultAwsGuardDutyScanningInterval = 60

var validAwsGuardDutyScanningIntervals = []int{5, 10, 15, 30, 60}

func resourceAviatrixControllerConfig() *schema.Resource {
	return &schema.Resource{
		Create: resourceAviatrixControllerConfigCreate,
		Read:   resourceAviatrixControllerConfigRead,
		Update: resourceAviatrixControllerConfigUpdate,
		Delete: resourceAviatrixControllerConfigDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"sg_management_account_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Cloud account name of user.",
			},
			"security_group_management": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Used to manage the Controller instance’s inbound rules from gateways.",
			},
			"http_access": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Switch for http access. Default: false.",
			},
			"fqdn_exception_rule": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "A system-wide mode. Default: true.",
			},
			"target_version": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The release version number to which the controller will be upgraded to.",
			},
			"manage_gateway_upgrades": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
				Description: "If true, aviatrix_controller_config will upgrade all gateways when target_version " +
					"is set. If false, only the controller will be upgraded when target_version is set. In that " +
					"case gateway upgrades should be handled in each gateway resource individually using the " +
					"software_version and image_version attributes.",
			},
			"backup_configuration": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Switch to enable/disable controller cloudn backup config.",
			},
			"backup_cloud_type": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     0,
				Description: "Type of cloud service provider, requires an integer value. Use 1 for AWS.",
			},
			"backup_account_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "This parameter represents the name of a Cloud-Account in Aviatrix controller.",
			},
			"backup_bucket_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "Bucket name. Required for AWS, AWSGov, GCP and OCI.",
			},
			"backup_storage_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "Storage name. Required for Azure.",
			},
			"backup_container_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "Container name. Required for Azure.",
			},
			"backup_region": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "Name of region. Required for Azure and OCI.",
			},
			"multiple_backups": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Switch to enable the controller to backup up to a maximum of 3 rotating backups.",
			},
			"version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Current version of the controller without the build number.",
			},
			"current_version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Current version of the controller.",
			},
			"previous_version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Previous version of the controller.",
			},
			"enable_vpc_dns_server": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Enable VPC/VNET DNS Server.",
			},
			"ca_certificate_file_path": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "File path to the CA Certificate.",
				RequiredWith: []string{"server_public_certificate_file_path", "server_private_key_file_path"},
			},
			"server_public_certificate_file_path": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "File path to the Server public certificate.",
				RequiredWith: []string{"ca_certificate_file_path", "server_private_key_file_path"},
			},
			"server_private_key_file_path": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "File path to server private key.",
				RequiredWith: []string{"server_public_certificate_file_path", "ca_certificate_file_path"},
			},
			"aws_guard_duty_scanning_interval": {
				Type:         schema.TypeInt,
				Optional:     true,
				Description:  "Scanning Interval for AWS Guard Duty.",
				Default:      defaultAwsGuardDutyScanningInterval,
				ValidateFunc: validation.IntInSlice(validAwsGuardDutyScanningIntervals),
			},
		},
	}
}

func resourceAviatrixControllerConfigCreate(d *schema.ResourceData, meta interface{}) error {
	var err error

	account := d.Get("sg_management_account_name").(string)

	client := meta.(*goaviatrix.Client)

	log.Printf("[INFO] Configuring Aviatrix controller : %#v", d)

	httpAccess := d.Get("http_access").(bool)
	if httpAccess {
		curStatus, _ := client.GetHttpAccessEnabled()
		if curStatus == "True" {
			log.Printf("[INFO] Http Access is already enabled")
		} else {
			err = client.EnableHttpAccess()
			time.Sleep(10 * time.Second)
		}
	} else {
		curStatus, _ := client.GetHttpAccessEnabled()
		if curStatus == "False" {
			log.Printf("[INFO] Http Access is already disabled")
		} else {
			err = client.DisableHttpAccess()
			time.Sleep(10 * time.Second)
		}
	}
	if err != nil {
		return fmt.Errorf("failed to configure controller http access: %s", err)
	}

	fqdnExceptionRule := d.Get("fqdn_exception_rule").(bool)
	if fqdnExceptionRule {
		curStatus, _ := client.GetExceptionRuleStatus()
		if curStatus {
			log.Printf("[INFO] FQDN Exception Rule is already enabled")
		} else {
			err = client.EnableExceptionRule()
		}
	} else {
		curStatus, _ := client.GetExceptionRuleStatus()
		if !curStatus {
			log.Printf("[INFO] FQDN Exception Rule is already disabled")
		} else {
			err = client.DisableExceptionRule()
		}
	}
	if err != nil {
		return fmt.Errorf("failed to configure controller exception rule: %s", err)
	}

	securityGroupManagement := d.Get("security_group_management").(bool)
	if securityGroupManagement {
		curStatus, _ := client.GetSecurityGroupManagementStatus()
		if curStatus.State == "Enabled" {
			log.Printf("[INFO] Security Group Management is already enabled")
		} else {
			err = client.EnableSecurityGroupManagement(account)
		}
	} else {
		curStatus, _ := client.GetSecurityGroupManagementStatus()
		if curStatus.State == "Disabled" {
			log.Printf("[INFO] Security Group Management is already disabled")
		} else {
			err = client.DisableSecurityGroupManagement()
		}
	}
	if err != nil {
		return fmt.Errorf("failed to configure controller Security Group Management: %s", err)
	}

	version := &goaviatrix.Version{
		Version: d.Get("target_version").(string),
	}
	if version.Version != "" {
		manageGatewayUpgrades := d.Get("manage_gateway_upgrades").(bool)
		err = client.AsyncUpgrade(version, manageGatewayUpgrades)
		if err != nil {
			return fmt.Errorf("failed to upgrade Aviatrix Controller: %s", err)
		}
		newCurrent, _, _ := client.GetCurrentVersion()
		log.Printf("Upgrade complete (now %s)", newCurrent)
	}

	backupConfiguration := d.Get("backup_configuration").(bool)
	backupCloudType := d.Get("backup_cloud_type").(int)
	backupAccountName := d.Get("backup_account_name").(string)
	backupBucketName := d.Get("backup_bucket_name").(string)
	backupStorageName := d.Get("backup_storage_name").(string)
	backupContainerName := d.Get("backup_container_name").(string)
	backupRegion := d.Get("backup_region").(string)
	multipleBackups := d.Get("multiple_backups").(bool)

	if backupConfiguration {
		err = validateBackupConfig(d)
		if err != nil {
			return err
		}

		cloudnBackupConfiguration := &goaviatrix.CloudnBackupConfiguration{
			BackupCloudType:     backupCloudType,
			BackupAccountName:   backupAccountName,
			BackupBucketName:    backupBucketName,
			BackupStorageName:   backupStorageName,
			BackupContainerName: backupContainerName,
			BackupRegion:        backupRegion,
		}
		if multipleBackups {
			cloudnBackupConfiguration.MultipleBackups = "true"
		}

		err = client.EnableCloudnBackupConfig(cloudnBackupConfiguration)
		if err != nil {
			return fmt.Errorf("failed to enable backup configuration: %s", err)
		}
	} else {
		if backupCloudType != 0 || backupAccountName != "" || backupBucketName != "" || backupStorageName != "" ||
			backupContainerName != "" || backupRegion != "" || multipleBackups {
			return fmt.Errorf("'backup_cloud_type', 'backup_account_name', 'backup_bucket_name'," +
				" 'backup_storage_name', 'backup_container_name' and 'backup_region' should all be empty," +
				" 'multiple_backups' should be empty or false for not enabling backup configuration")
		}
	}

	enableVpcDnsServer := d.Get("enable_vpc_dns_server").(bool)
	err = client.SetControllerVpcDnsServer(enableVpcDnsServer)
	if err != nil {
		return fmt.Errorf("could not toggle controller vpc dns server: %v", err)
	}

	certConfig := &goaviatrix.HTTPSCertConfig{
		CACertificateFilePath:     d.Get("ca_certificate_file_path").(string),
		ServerCertificateFilePath: d.Get("server_public_certificate_file_path").(string),
		ServerPrivateKeyFilePath:  d.Get("server_private_key_file_path").(string),
	}
	if !(certConfig.CACertificateFilePath != "" && certConfig.ServerCertificateFilePath != "" && certConfig.ServerPrivateKeyFilePath != "") &&
		!(certConfig.CACertificateFilePath == "" && certConfig.ServerCertificateFilePath == "" && certConfig.ServerPrivateKeyFilePath == "") {
		return fmt.Errorf("ca_certificate_file_path, server_public_certificate_file_path, and server_private_key_file_path must either all be set or all unset. Please update your configuaration")
	}
	if certConfig.CACertificateFilePath != "" && certConfig.ServerCertificateFilePath != "" && certConfig.ServerPrivateKeyFilePath != "" {
		err = client.ImportNewHTTPSCerts(certConfig)
		if err != nil {
			return fmt.Errorf("could not import HTTPS certs: %v", err)
		}
	}

	scanningInterval := d.Get("aws_guard_duty_scanning_interval")
	err = client.UpdateAwsGuardDutyPollInterval(scanningInterval.(int))
	if err != nil {
		return fmt.Errorf("could not update scanning interval: %v", err)
	}

	d.SetId(strings.Replace(client.ControllerIP, ".", "-", -1))
	return resourceAviatrixControllerConfigRead(d, meta)
}

func resourceAviatrixControllerConfigRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*goaviatrix.Client)

	log.Printf("[INFO] Getting controller %s configuration", d.Id())
	result, err := client.GetHttpAccessEnabled()
	if err != nil {
		if err == goaviatrix.ErrNotFound {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("could not read Aviatrix Controller Config: %s", err)
	}

	if result[1:5] == "True" {
		d.Set("http_access", true)
	} else {
		d.Set("http_access", false)
	}

	res, err := client.GetExceptionRuleStatus()
	if err != nil {
		return fmt.Errorf("could not read Aviatrix Controller Exception Rule Status: %s", err)
	}
	if res {
		d.Set("fqdn_exception_rule", true)
	} else {
		d.Set("fqdn_exception_rule", false)
	}

	sgm, err := client.GetSecurityGroupManagementStatus()
	if err != nil {
		return fmt.Errorf("could not read Aviatrix Controller Security Group Management Status: %s", err)
	}
	if sgm != nil {
		if sgm.State == "Enabled" {
			d.Set("security_group_management", true)
		} else {
			d.Set("security_group_management", false)
		}
		d.Set("sg_management_account_name", sgm.AccountName)
	} else {
		return fmt.Errorf("could not read Aviatrix Controller Security Group Management Status")
	}

	versionInfo, err := client.GetVersionInfo()
	if err != nil {
		return fmt.Errorf("unable to read Controller version information: %s", err)
	}

	current := versionInfo.Current.String(true)
	currentWithoutBuild := versionInfo.Current.String(false)
	previous := versionInfo.Previous.String(true)
	targetVersion := d.Get("target_version")
	if targetVersion == "latest" {
		d.Set("target_version", currentWithoutBuild)
	} else {
		d.Set("target_version", targetVersion)
	}

	d.Set("version", currentWithoutBuild)
	d.Set("previous_version", previous)
	d.Set("current_version", current)

	cloudnBackupConfig, err := client.GetCloudnBackupConfig()
	if err != nil {
		return fmt.Errorf("unable to read current controller cloudn backup config: %s", err)
	}
	if cloudnBackupConfig != nil && cloudnBackupConfig.BackupConfiguration == "yes" {
		d.Set("backup_configuration", true)
		d.Set("backup_cloud_type", cloudnBackupConfig.BackupCloudType)
		d.Set("backup_account_name", cloudnBackupConfig.BackupAccountName)
		d.Set("backup_bucket_name", cloudnBackupConfig.BackupBucketName)
		d.Set("backup_storage_name", cloudnBackupConfig.BackupStorageName)
		d.Set("backup_container_name", cloudnBackupConfig.BackupContainerName)
		d.Set("backup_region", cloudnBackupConfig.BackupRegion)
		if cloudnBackupConfig.MultipleBackups == "yes" {
			d.Set("multiple_backups", true)
		} else {
			d.Set("multiple_backups", false)
		}
	}

	vpcDnsServerEnabled, err := client.GetControllerVpcDnsServerStatus()
	if err != nil {
		return fmt.Errorf("could not get controller vpc dns server status: %v", err)
	}

	d.Set("enable_vpc_dns_server", vpcDnsServerEnabled)

	httpsCertsImported, err := client.GetHTTPSCertsStatus()
	if err != nil {
		return fmt.Errorf("could not get HTTPS Certificate status: %v", err)
	}
	if !httpsCertsImported {
		d.Set("ca_certificate_file_path", "")
		d.Set("server_public_certificate_file_path", "")
		d.Set("server_private_key_file_path", "")
	}

	guardDuty, err := client.GetAwsGuardDuty()
	if err != nil {
		return fmt.Errorf("could not get aws guard duty scanning interval: %v", err)
	}
	d.Set("aws_guard_duty_scanning_interval", guardDuty.ScanningInterval)

	d.SetId(strings.Replace(client.ControllerIP, ".", "-", -1))
	return nil
}

func resourceAviatrixControllerConfigUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*goaviatrix.Client)
	account := d.Get("sg_management_account_name").(string)

	log.Printf("[INFO] Updating Controller configuration: %#v", d)
	d.Partial(true)

	if d.HasChange("http_access") {
		httpAccess := d.Get("http_access").(bool)
		if httpAccess {
			err := client.EnableHttpAccess()
			time.Sleep(10 * time.Second)
			if err != nil {
				log.Printf("[ERROR] Failed to enable http access on controller %s", d.Id())
				return err
			}
		} else {
			err := client.DisableHttpAccess()
			time.Sleep(10 * time.Second)
			if err != nil {
				log.Printf("[ERROR] Failed to disable http access on controller %s", d.Id())
				return err
			}
		}
	}

	if d.HasChange("fqdn_exception_rule") {
		fqdnExceptionRule := d.Get("fqdn_exception_rule").(bool)
		if fqdnExceptionRule {
			err := client.EnableExceptionRule()
			if err != nil {
				log.Printf("[ERROR] Failed to enable exception rule on controller %s", d.Id())
				return err
			}
		} else {
			err := client.DisableExceptionRule()
			if err != nil {
				log.Printf("[ERROR] Failed to disable exception rule on controller %s", d.Id())
				return err
			}
		}
	}

	if d.HasChange("security_group_management") {
		securityGroupManagement := d.Get("security_group_management").(bool)
		if securityGroupManagement {
			err := client.EnableSecurityGroupManagement(account)
			if err != nil {
				log.Printf("[ERROR] Failed to enable Security Group Management on controller %s", d.Id())
				return err
			}
		} else {
			err := client.DisableSecurityGroupManagement()
			if err != nil {
				log.Printf("[ERROR] Failed to disable Security Group Management on controller %s", d.Id())
				return err
			}
		}
	}

	if d.HasChange("target_version") {
		curVersion := d.Get("version").(string)
		cur := strings.Split(curVersion, ".")
		latestVersion, _ := client.GetLatestVersion()
		latest := strings.Split(latestVersion, ".")
		version := &goaviatrix.Version{
			Version: d.Get("target_version").(string),
		}
		if version.Version != "" {
			manageGatewayUpgrades := d.Get("manage_gateway_upgrades").(bool)
			targetVersion := d.Get("target_version").(string)
			if targetVersion == "latest" {
				if latestVersion != "" {
					for i := range cur {
						if cur[i] != latest[i] {
							err := client.AsyncUpgrade(version, manageGatewayUpgrades)
							if err != nil {
								return fmt.Errorf("failed to upgrade Aviatrix Controller: %s", err)
							}
							break
						}
					}
				}
			} else {
				err := client.AsyncUpgrade(version, manageGatewayUpgrades)
				if err != nil {
					return fmt.Errorf("failed to upgrade Aviatrix Controller: %s", err)
				}
			}
		}
	}

	backupConfiguration := d.Get("backup_configuration").(bool)
	backupCloudType := d.Get("backup_cloud_type").(int)
	backupAccountName := d.Get("backup_account_name").(string)
	backupBucketName := d.Get("backup_bucket_name").(string)
	backupStorageName := d.Get("backup_storage_name").(string)
	backupContainerName := d.Get("backup_container_name").(string)
	backupRegion := d.Get("backup_region").(string)
	multipleBackups := d.Get("multiple_backups").(bool)

	if d.HasChange("backup_configuration") {
		if backupConfiguration {
			err := validateBackupConfig(d)
			if err != nil {
				return err
			}

			cloudnBackupConfiguration := &goaviatrix.CloudnBackupConfiguration{
				BackupCloudType:     backupCloudType,
				BackupAccountName:   backupAccountName,
				BackupBucketName:    backupBucketName,
				BackupStorageName:   backupStorageName,
				BackupContainerName: backupContainerName,
				BackupRegion:        backupRegion,
			}
			if multipleBackups {
				cloudnBackupConfiguration.MultipleBackups = "true"
			}

			err = client.EnableCloudnBackupConfig(cloudnBackupConfiguration)
			if err != nil {
				return fmt.Errorf("failed to enable backup configuration: %s", err)
			}
		} else {
			if backupCloudType != 0 || backupAccountName != "" || backupBucketName != "" || backupStorageName != "" ||
				backupContainerName != "" || backupRegion != "" || multipleBackups {
				return fmt.Errorf("'backup_cloud_type', 'backup_account_name', 'backup_bucket_name'," +
					" 'backup_storage_name', 'backup_container_name' and 'backup_region' should all be empty," +
					" 'multiple_backups' should be empty or false for not enabling backup configuration")
			}

			err := client.DisableCloudnBackupConfig()
			if err != nil {
				return fmt.Errorf("failed to disable backup configuration: %s", err)
			}
		}
	} else {
		if d.HasChange("backup_cloud_type") || d.HasChange("backup_account_name") ||
			d.HasChange("backup_bucket_name") || d.HasChange("backup_storage_name") ||
			d.HasChange("backup_container_name") || d.HasChange("backup_region") ||
			d.HasChange("multiple_backups") {

			if backupConfiguration {
				err := validateBackupConfig(d)
				if err != nil {
					return err
				}

				err = client.DisableCloudnBackupConfig()
				if err != nil {
					return fmt.Errorf("failed to disable backup configuration: %s", err)
				}

				cloudnBackupConfiguration := &goaviatrix.CloudnBackupConfiguration{
					BackupCloudType:     backupCloudType,
					BackupAccountName:   backupAccountName,
					BackupBucketName:    backupBucketName,
					BackupStorageName:   backupStorageName,
					BackupContainerName: backupContainerName,
					BackupRegion:        backupRegion,
				}
				if multipleBackups {
					cloudnBackupConfiguration.MultipleBackups = "true"
				}

				err = client.EnableCloudnBackupConfig(cloudnBackupConfiguration)
				if err != nil {
					return fmt.Errorf("failed to enable backup configuration: %s", err)
				}
			} else {
				if backupCloudType != 0 || backupAccountName != "" || backupBucketName != "" || backupStorageName != "" ||
					backupContainerName != "" || backupRegion != "" || multipleBackups {
					return fmt.Errorf("'backup_cloud_type', 'backup_account_name', 'backup_bucket_name'," +
						" 'backup_storage_name', 'backup_container_name' and 'backup_region' should all be empty," +
						" 'multiple_backups' should be empty or false for not enabling backup configuration")
				}
			}
		}
	}

	if d.HasChange("enable_vpc_dns_server") {
		enableVpcDnsServer := d.Get("enable_vpc_dns_server").(bool)
		err := client.SetControllerVpcDnsServer(enableVpcDnsServer)
		if err != nil {
			return fmt.Errorf("could not toggle controller vpc dns server: %v", err)
		}
	}

	if d.HasChange("ca_certificate_file_path") || d.HasChange("server_public_certificate_file_path") || d.HasChange("server_private_key_file_path") {
		certConfig := &goaviatrix.HTTPSCertConfig{
			CACertificateFilePath:     d.Get("ca_certificate_file_path").(string),
			ServerCertificateFilePath: d.Get("server_public_certificate_file_path").(string),
			ServerPrivateKeyFilePath:  d.Get("server_private_key_file_path").(string),
		}
		if !(certConfig.CACertificateFilePath != "" && certConfig.ServerCertificateFilePath != "" && certConfig.ServerPrivateKeyFilePath != "") &&
			!(certConfig.CACertificateFilePath == "" && certConfig.ServerCertificateFilePath == "" && certConfig.ServerPrivateKeyFilePath == "") {
			return fmt.Errorf("ca_certificate_file_path, server_public_certificate_file_path, and server_private_key_file_path must either all be set or all unset. Please update your configuaration")
		}
		err := client.DisableImportedHTTPSCerts()
		if err != nil {
			return fmt.Errorf("problem trying to disable imported certs to prepare for importing new certs: %v", err)
		}
		if certConfig.CACertificateFilePath != "" && certConfig.ServerCertificateFilePath != "" && certConfig.ServerPrivateKeyFilePath != "" {
			err = client.ImportNewHTTPSCerts(certConfig)
			if err != nil {
				return fmt.Errorf("could not import new HTTPS certs: %v", err)
			}
		}
	}

	if d.HasChange("aws_guard_duty_scanning_interval") {
		scanningInterval := d.Get("aws_guard_duty_scanning_interval").(int)
		err := client.UpdateAwsGuardDutyPollInterval(scanningInterval)
		if err != nil {
			return fmt.Errorf("could not update scanning interval: %v", err)
		}
	}

	d.Partial(false)
	return resourceAviatrixControllerConfigRead(d, meta)
}

func resourceAviatrixControllerConfigDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*goaviatrix.Client)
	d.Set("http_access", false)
	curStatusHttp, _ := client.GetHttpAccessEnabled()
	if curStatusHttp != "Disabled" {
		err := client.DisableHttpAccess()
		time.Sleep(10 * time.Second)
		if err != nil {
			log.Printf("[ERROR] Failed to disable http access on controller %s", d.Id())
			return err
		}
	}

	d.Set("fqdn_exception_rule", true)
	curStatusException, _ := client.GetExceptionRuleStatus()
	if !curStatusException {
		err := client.EnableExceptionRule()
		if err != nil {
			log.Printf("[ERROR] Failed to enable exception rule on controller %s", d.Id())
			return err
		}
	}

	d.Set("security_group_management", false)
	curStatusSG, _ := client.GetSecurityGroupManagementStatus()
	if curStatusSG.State != "Disabled" {
		err := client.DisableSecurityGroupManagement()
		if err != nil {
			log.Printf("[ERROR] Failed to disable security group management on controller %s", d.Id())
			return err
		}
	}

	d.Set("backup_configuration", false)
	cloudnBackupConfig, _ := client.GetCloudnBackupConfig()
	if cloudnBackupConfig.BackupConfiguration == "yes" {
		err := client.DisableCloudnBackupConfig()
		if err != nil {
			log.Printf("[ERROR] Failed to disable cloudn backup config on controller %s", d.Id())
			return err
		}
	}

	err := client.SetControllerVpcDnsServer(false)
	if err != nil {
		return fmt.Errorf("could not disable controller vpc dns server: %v", err)
	}

	err = client.DisableImportedHTTPSCerts()
	if err != nil {
		return fmt.Errorf("could not disable imported certs: %v", err)
	}

	err = client.UpdateAwsGuardDutyPollInterval(defaultAwsGuardDutyScanningInterval)
	if err != nil {
		return fmt.Errorf("could not update scanning interval: %v", err)
	}

	return nil
}

func validateBackupConfig(d *schema.ResourceData) error {
	backupCloudType := d.Get("backup_cloud_type").(int)
	backupAccountName := d.Get("backup_account_name").(string)
	backupBucketName := d.Get("backup_bucket_name").(string)
	backupStorageName := d.Get("backup_storage_name").(string)
	backupContainerName := d.Get("backup_container_name").(string)
	backupRegion := d.Get("backup_region").(string)

	if backupCloudType == 0 || backupAccountName == "" {
		return fmt.Errorf("please specify 'backup_cloud_type' and 'backup_account_name'" +
			" to enable backup configuration")
	}

	switch backupCloudType {
	case goaviatrix.AWS, goaviatrix.AWSGov, goaviatrix.GCP:
		if backupBucketName == "" {
			return fmt.Errorf("please specify 'backup_bucket_name' to enable backup configuration for AWS, AWSGov and GCP")
		}
		if backupStorageName != "" || backupContainerName != "" || backupRegion != "" {
			return fmt.Errorf("'backup_storage_name', 'backup_container_name' and 'backup_region'" +
				" should be empty for AWS, AWSGov and GCP")
		}
	case goaviatrix.Azure:
		if backupStorageName == "" || backupContainerName == "" || backupRegion == "" {
			return fmt.Errorf("please specify 'backup_storage_name', 'backup_container_name' and" +
				" 'backup_region' to enable backup configuration for Azure")
		}
		if backupBucketName != "" {
			return fmt.Errorf("'backup_bucket_name' should be empty for Azure")

		}
	case goaviatrix.OCI:
		if backupBucketName == "" || backupRegion == "" {
			return fmt.Errorf("please specify 'backup_bucket_name' and 'backup_region'" +
				" to enable backup configuration for OCI")
		}
		if backupStorageName != "" || backupContainerName != "" {
			return fmt.Errorf("'backup_storage_name' and 'backup_container_name'" +
				" should be empty for OCI")
		}
	}

	return nil
}
