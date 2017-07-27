include "config.device_classes.hcl" {}

preferences {
    // The destination directory for all configurations
	backup_dir = "./config-backups/"

	// The IP address devices will use to upload their configs
	host_ip = "192.168.100.10"
}

device_class "sample_device_class" {

    backup_target "config_target" {
        // The macro to trigger the config upload on the device
        macro = <<-MACRO
            expect(">")
            sendLine("enable")
            expect("Password: ")
            sendLine(getAuthAttr("enable_password"))
            expect("# ")
            sendLine("configure terminal")
            expect("# ")
            sendLine("copy startup-config tftp")
            expect("?")
            sendLine(ctx.TFTPHost)
            expect("?")
            sendLine(ctx.TFTPFilename)
            expect("# ")
        MACRO
    }
}

auth_provider "static" "basic_auth" {
    auth "secretA" {
        username = "jdoe"
    	password = "supersecret"
    }
    auth "secretB" {
        username = "jdoe"
        password = "supersecret"
        attributes {
            enable_password = "supersecret2"
        }
    }
}

device "sample_device_class" "sample_device" {
	address = "192.168.100.1:22"
	auth = "basic_auth:secretA"
}
