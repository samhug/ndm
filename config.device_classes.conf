device_class "cisco_isr" {
    backup_target "startup_config" {
        macro = <<-MACRO
            expect("#")
            sendLine("copy startup-config tftp://" + ctx.TFTPHost + "/" + ctx.TFTPFilename)
            expect("["+ctx.TFTPHost+"]?")
            sendLine("")
            expect("["+ctx.TFTPFilename+"]?")
            sendLine("")
            expect("#")
        MACRO
    }
    backup_target "running_config" {
        macro = <<-MACRO
            expect("#")
            sendLine("copy running-config tftp://" + ctx.TFTPHost + "/" + ctx.TFTPFilename)
            expect("["+ctx.TFTPHost+"]?")
            sendLine("")
            expect("["+ctx.TFTPFilename+"]?")
            sendLine("")
            expect("#")
        MACRO
    }
}

device_class "hp_switch" {
    backup_target "startup_config" {
        macro = <<-MACRO
            expect(">")
            sendLine("backup startup-configuration to " + ctx.TFTPHost + " " + ctx.TFTPFilename + ".cfg")
            expect(">")
        MACRO
	}
}

device_class "adtran" {
    backup_target "startup_config" {
        macro = <<-MACRO
            expect(">")
            sendLine("enable")
            expect("Password:")
            sendLine(getAuthAttr("enable_password"))
            expect("#")
            sendLine("copy startup-config tftp")
            expect("?")
            sendLine(ctx.TFTPHost)
            expect("?")
            sendLine(ctx.TFTPFilename)
            expect("#")
        MACRO
    }
    backup_target "running_config" {
        macro = <<-MACRO
            expect(">")
            sendLine("enable")
            expect("Password:")
            sendLine(getAuthAttr("enable_password"))
            expect("#")
            sendLine("copy running-config tftp")
            expect("?")
            sendLine(ctx.TFTPHost)
            expect("?")
            sendLine(ctx.TFTPFilename)
            expect("#")
        MACRO
    }
}

device_class "watchguard" {
    backup_target "config" {
        macro = <<-MACRO
            expect("#")
            sendLine("export config to tftp://"+ctx.TFTPHost+"/"+ctx.TFTPFilename)
            expect("#")
        MACRO
    }
}

device_class "aruba" {
    backup_target "config" {
        macro = <<-MACRO
            expect("# ")
            sendLine("copy config tftp " + ctx.TFTPHost + " " + ctx.TFTPFilename)
            expect("# ")
        MACRO
    }
}

device_class "ubiquiti" {
	backup_target "running_config" {
        macro = <<-MACRO
            expect("$ ")
            sendLine("copy file running://config/ to tftp://"+ctx.TFTPHost+"/"+ctx.TFTPFilename)
            expect("$ ")
        MACRO
	}
}
