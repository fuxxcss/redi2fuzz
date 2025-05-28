package utils

// fuxx tools
const (
	AFL string = "afl"
	HonggFuzz string = "honggfuzz"

	CoverageMap string = "COVERAGE_MAP"
	BaseTool string = "BASE_TOOL"
)

// tools features
type ToolsFeature int

const (
	// exe and args
	TOOLS_EXE ToolsFeature = iota
	TOOLS_TIMEOUT
	TOOLS_INPUT
	TOOLS_OUTPUT
	// envs
	TOOLS_ENV_DEBUG
	TOOLS_ENV_DEBUG_SIZE
	TOOLS_ENV_MAX_SIZE
	TOOLS_ENV_SHM_ID
	TOOLS_ENV_CUSTOM_FLAG
	TOOLS_ENV_CUSTOM_PATH
	TOOLS_ENV_SKIP_CPUFREQ
	TOOLS_ENV_SKIP_BIN_CHECK
	TOOLS_ENV_USE_ASAN
	TOOLS_ENV_NO_CAL
)

type ToolsType map[ToolsFeature]string

var Tools = map[string]ToolsType {
	AFL : {
		// exe and args
		TOOLS_EXE : "afl-fuzz",
		TOOLS_TIMEOUT : "-t",
		TOOLS_INPUT : "-i",
		TOOLS_OUTPUT : "-o",
		// envs
		TOOLS_ENV_DEBUG : "AFL_DEBUG",
		TOOLS_ENV_DEBUG_SIZE : "__afl_map_size",
		TOOLS_ENV_MAX_SIZE : "AFL_MAP_SIZE",
		TOOLS_ENV_SHM_ID : "__AFL_SHM_ID",
		TOOLS_ENV_CUSTOM_FLAG : "AFL_CUSTOM_MUTATOR_ONLY",
		TOOLS_ENV_CUSTOM_PATH : "AFL_CUSTOM_MUTATOR_LIBRARY",
		TOOLS_ENV_SKIP_CPUFREQ : "AFL_SKIP_CPUFREQ",
		TOOLS_ENV_SKIP_BIN_CHECK : "AFL_SKIP_BIN_CHECK",
		TOOLS_ENV_USE_ASAN : "AFL_USE_ASAN",
		TOOLS_ENV_NO_CAL : "AFL_NO_STARTUP_CALIBRATION",
		
	},
	HonggFuzz : {

	},
	
}