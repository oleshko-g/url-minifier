package main

import (
	"cmp"
	"fmt"
	"runtime/debug"
)

var (
	buildVersion, buildDate, buildCommit string
)

func printBuildInfo() {
	if bi, ok := debug.ReadBuildInfo(); ok {
		buildVersion = bi.Main.Version
		buildCommit = getBuildSetting(bi.Settings, "vcs.revision")
	}

	fmt.Println(fmt.Sprint("Build version: ", cmp.Or(buildVersion, "N/A")))
	fmt.Println(fmt.Sprint("Build date: ", cmp.Or(buildDate, "N/A")))
	fmt.Println(fmt.Sprint("Build commit: ", cmp.Or(buildCommit, "N/A")))
}

func getBuildSetting(settings []debug.BuildSetting, key string) string {
	for _, setting := range settings {
		if setting.Key == key {
			return setting.Value
		}
	}
	return ""
}
