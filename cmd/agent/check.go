package agent

// PreflightCheck 这里定义了对配置的检查
func PreflightCheck() error {
	checkA2sConfigPath()
	checkPterodactylDirExists()
	return nil
}

func checkPterodactylDirExists() {

}

// checkA2sConfigPath 检查 A2S 配置目录是否存在
func checkA2sConfigPath() error {
	return nil
}
