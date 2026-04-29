package i18n

// TranslationSet is a set of localised strings for a given language
type TranslationSet struct {
	NotEnoughSpace                             string
	ProjectTitle                               string
	MainTitle                                  string
	GlobalTitle                                string
	Navigate                                   string
	Menu                                       string
	MenuTitle                                  string
	Execute                                    string
	Scroll                                     string
	Close                                      string
	Quit                                       string
	ErrorTitle                                 string
	NoViewMachingNewLineFocusedSwitchStatement string
	OpenConfig                                 string
	EditConfig                                 string
	ConfirmQuit                                string
	ConfirmUpProject                           string
	ErrorOccurred                              string
	ConnectionFailed                           string
	UnattachableContainerError                 string
	WaitingForContainerInfo                    string
	CannotAttachStoppedContainerError          string
	CannotAccessDockerSocketError              string
	CannotKillChildError                       string

	Donate                      string
	Cancel                      string
	CustomCommandTitle          string
	BulkCommandTitle            string
	Remove                      string
	HideStopped                 string
	ForceRemove                 string
	RemoveWithVolumes           string
	MustForceToRemoveContainer  string
	Confirm                     string
	Return                      string
	FocusMain                   string
	LcFilter                    string
	StopContainer               string
	RestartingStatus            string
	StartingStatus              string
	StoppingStatus              string
	UppingProjectStatus         string
	UppingServiceStatus         string
	PausingStatus               string
	RemovingStatus              string
	DowningStatus               string
	RunningCustomCommandStatus  string
	RunningBulkCommandStatus    string
	RemoveService               string
	UpService                   string
	Stop                        string
	Pause                       string
	Restart                     string
	Down                        string
	DownWithVolumes             string
	Start                       string
	Rebuild                     string
	Recreate                    string
	PreviousContext             string
	NextContext                 string
	Attach                      string
	ViewLogs                    string
	UpProject                   string
	DownProject                 string
	ServicesTitle               string
	ContainersTitle             string
	StandaloneContainersTitle   string
	TopTitle                    string
	ImagesTitle                 string
	VolumesTitle                string
	NetworksTitle               string
	NoContainers                string
	NoContainer                 string
	NoImages                    string
	NoVolumes                   string
	NoNetworks                  string
	NoServices                  string
	RemoveImage                 string
	RemoveVolume                string
	RemoveNetwork               string
	RemoveWithoutPrune          string
	RemoveWithoutPruneWithForce string
	RemoveWithForce             string
	PruneImages                 string
	PruneContainers             string
	PruneVolumes                string
	PruneNetworks               string
	ConfirmPruneContainers      string
	ConfirmStopContainers       string
	ConfirmRemoveContainers     string
	ConfirmPruneImages          string
	ConfirmPruneVolumes         string
	ConfirmPruneNetworks        string
	PruningStatus               string
	StopService                 string
	PressEnterToReturn          string
	DetachFromContainerShortCut string
	StopAllContainers           string
	RemoveAllContainers         string
	ViewRestartOptions          string
	ExecShell                   string
	RunCustomCommand            string
	ViewBulkCommands            string
	FilterList                  string
	OpenInBrowser               string
	SortContainersByState       string

	LogsTitle                   string
	ConfigTitle                 string
	EnvTitle                    string
	DockerComposeConfigTitle    string
	StatsTitle                  string
	CreditsTitle                string
	ContainerConfigTitle        string
	ContainerEnvTitle           string
	NothingToDisplay            string
	NoContainerForService       string
	CannotDisplayEnvVariables   string
	CannotManageNonLocalService string

	No  string
	Yes string

	LcNextScreenMode string
	LcPrevScreenMode string
	FilterPrompt     string

	FocusProjects   string
	FocusServices   string
	FocusContainers string
	FocusImages     string
	FocusVolumes    string
	FocusNetworks   string

	// CLI strings
	AppTitle                   string
	LoadedFiles                string
	DefaultComposeFile         string
	AllServicesList            string
	AllServicesListAll         string
	NotRunningServicesList     string
	RunningServicesList        string
	StatusNotRunning           string
	StatusRunning              string
	StatusStopped              string
	CommonTips                 string
	QuickCommands              string
	Goodbye                    string
	SearchServiceTitle         string
	SearchMenuTitle            string
	SelectActionForService     string
	ExecutingActionOnService   string
	ExternalProjectStatusTip   string
	ExternalProjectNoConfigTip string
	ExternalProjectNoFixTip    string
	ExternalProjectNoBuildTip  string
	RepairingService           string
	DeletingServiceImage       string
	ConfirmCleanService        string
	WarningStopAndRemove       string
	ConfirmContinue            string
	ServiceNotRunningNoLogs    string
	ViewingServiceLogs         string
	EnteringContainer          string
	ActionFailed               string
	ActionSuccess              string
	ActionCompleted            string
	SelectServiceTo            string
	PromptServiceIdxName       string
	PromptServiceAllIdxName    string
	InputServiceNameIdx        string
	ErrorNoMatchingService     string
	ErrorNoStackServiceFound   string
	ExecutingAction            string
	ConfirmOneKeyStartStack    string
	StartingStack              string
	ContainsServices           string
	CleaningDockerBuildCache   string
	CleaningDockerBuildHistory string
	QuickSearchPrompt          string
	MenuFunction               string
	WaitEnterToContinue        string
	Build                      string
	Clean                      string
	Fix                        string
	ForceReconstruct           string
	FzfSelected                string
	SuffixAllSpecified         string
	SuffixSpecified            string
	CommandExecutionError      string
	ExecutionFinishedExitTip   string
	PressExitToReturnTip       string
	ReturningToMainMenu        string
	MustTypeExitToQuit         string

	// Menu Items
	MenuStartService      string
	MenuStopService       string
	MenuRestartService    string
	MenuViewLogs          string
	MenuServiceStatus     string
	MenuServiceConfig     string
	MenuEnterContainer    string
	MenuBuildService      string
	MenuForceReconstruct  string
	MenuCleanService      string
	MenuRemoveImage       string
	MenuLogStack          string
	MenuDBStack           string
	MenuCleanBuildCache   string
	MenuCleanBuildxCache  string
	MenuNetworkManagement string
	MenuVolumeManagement  string
	MenuImageManagement   string
	MenuRepairService     string
	// Image Management
	SearchImageTitle        string
	RunImage                string
	InputContainerName      string
	RunImageSuccess         string
	RunImageFailed          string
	SelectActionForImage    string
	ConfirmDeleteImage      string
	DeletingImage           string
	InvalidIndex            string
	DangerDeleteAllImages   string
	CleaningAllUnusedImages string
	PullImage               string
	MenuRunImage            string
	MenuDeleteImage         string
	DeleteAllImages         string
	InputImageToRun         string
	InputImageToDelete      string
	DetectingShell          string
	PullingImage            string
	PromptSearchKeyword     string
	SearchingRemoteImage    string
}

func englishSet() TranslationSet {
	return TranslationSet{
		PruningStatus:              "pruning",
		RemovingStatus:             "removing",
		RestartingStatus:           "restarting",
		StartingStatus:             "starting",
		StoppingStatus:             "stopping",
		UppingServiceStatus:        "upping service",
		UppingProjectStatus:        "upping project",
		DowningStatus:              "downing",
		PausingStatus:              "pausing",
		RunningCustomCommandStatus: "running custom command",
		RunningBulkCommandStatus:   "running bulk command",

		NoViewMachingNewLineFocusedSwitchStatement: "No view matching newLineFocused switch statement",

		ErrorOccurred:                     "An error occurred! Please create an issue at https://github.com/yaogh99123/dcli/issues",
		ConnectionFailed:                  "connection to docker client failed. You may need to restart the docker client",
		UnattachableContainerError:        "Container does not support attaching. You must either run the service with the '-it' flag or use `stdin_open: true, tty: true` in the docker-compose.yml file",
		WaitingForContainerInfo:           "Cannot proceed until docker gives us more information about the container. Please retry in a few moments.",
		CannotAttachStoppedContainerError: "You cannot attach to a stopped container, you need to start it first (which you can actually do with the 'r' key) (yes I'm too lazy to do this automatically for you) (pretty cool that I get to communicate one-on-one with you in the form of an error message though)",
		CannotAccessDockerSocketError:     "Can't access docker socket at: unix:///var/run/docker.sock\nRun dcli as root or read https://docs.docker.com/install/linux/linux-postinstall/",
		CannotKillChildError:              "Waited three seconds for child process to stop. There may be an orphan process that continues to run on your system.",

		Donate:  "Donate",
		Confirm: "Confirm",

		Return:                      "return",
		FocusMain:                   "focus main panel",
		LcFilter:                    "filter list",
		Navigate:                    "navigate",
		Execute:                     "execute",
		Close:                       "close",
		Quit:                        "quit",
		Menu:                        "menu",
		MenuTitle:                   "Menu",
		Scroll:                      "scroll",
		OpenConfig:                  "open dcli config",
		EditConfig:                  "edit dcli config",
		Cancel:                      "cancel",
		Remove:                      "remove",
		HideStopped:                 "hide/show stopped containers",
		ForceRemove:                 "force remove",
		RemoveWithVolumes:           "remove with volumes",
		RemoveService:               "remove containers",
		UpService:                   "up service",
		Stop:                        "stop",
		Pause:                       "pause",
		Restart:                     "restart",
		Down:                        "down project",
		DownWithVolumes:             "down project with volumes",
		Start:                       "start",
		Rebuild:                     "rebuild",
		Recreate:                    "recreate",
		PreviousContext:             "previous tab",
		NextContext:                 "next tab",
		Attach:                      "attach",
		ViewLogs:                    "view logs",
		UpProject:                   "up project",
		DownProject:                 "down project",
		RemoveImage:                 "remove image",
		RemoveVolume:                "remove volume",
		RemoveNetwork:               "remove network",
		RemoveWithoutPrune:          "remove without deleting untagged parents",
		RemoveWithoutPruneWithForce: "remove (forced) without deleting untagged parents",
		RemoveWithForce:             "remove (forced)",
		PruneContainers:             "prune exited containers",
		PruneVolumes:                "prune unused volumes",
		PruneNetworks:               "prune unused networks",
		PruneImages:                 "prune unused images",
		StopAllContainers:           "stop all containers",
		RemoveAllContainers:         "remove all containers (forced)",
		ViewRestartOptions:          "view restart options",
		ExecShell:                   "exec shell",
		RunCustomCommand:            "run predefined custom command",
		ViewBulkCommands:            "view bulk commands",
		FilterList:                  "filter list",
		OpenInBrowser:               "open in browser (first port is http)",
		SortContainersByState:       "sort containers by state",

		GlobalTitle:                 "Global",
		MainTitle:                   "Main",
		ProjectTitle:                "Project",
		ServicesTitle:               "Services",
		ContainersTitle:             "Containers",
		StandaloneContainersTitle:   "Standalone Containers",
		ImagesTitle:                 "Images",
		VolumesTitle:                "Volumes",
		NetworksTitle:               "Networks",
		CustomCommandTitle:          "Custom Command:",
		BulkCommandTitle:            "Bulk Command:",
		ErrorTitle:                  "Error",
		LogsTitle:                   "Logs",
		ConfigTitle:                 "Config",
		EnvTitle:                    "Env",
		DockerComposeConfigTitle:    "Docker-Compose Config",
		TopTitle:                    "Top",
		StatsTitle:                  "Stats",
		CreditsTitle:                "About",
		ContainerConfigTitle:        "Container Config",
		ContainerEnvTitle:           "Container Env",
		NothingToDisplay:            "Nothing to display",
		NoContainerForService:       "No logs to show; service is not associated with a container",
		CannotDisplayEnvVariables:   "Something went wrong while displaying environment variables",
		CannotManageNonLocalService: "This service belongs to a different compose project. Run dcli from that project's directory to manage it.",

		NoContainers: "No containers",
		NoContainer:  "No container",
		NoImages:     "No images",
		NoVolumes:    "No volumes",
		NoNetworks:   "No networks",
		NoServices:   "No services",

		ConfirmQuit:                 "Are you sure you want to quit?",
		ConfirmUpProject:            "Are you sure you want to 'up' your docker compose project?",
		MustForceToRemoveContainer:  "You cannot remove a running container unless you force it. Do you want to force it?",
		NotEnoughSpace:              "Not enough space to render panels",
		ConfirmPruneImages:          "Are you sure you want to prune all unused images?",
		ConfirmPruneContainers:      "Are you sure you want to prune all stopped containers?",
		ConfirmStopContainers:       "Are you sure you want to stop all containers?",
		ConfirmRemoveContainers:     "Are you sure you want to remove all containers?",
		ConfirmPruneVolumes:         "Are you sure you want to prune all unused volumes?",
		ConfirmPruneNetworks:        "Are you sure you want to prune all unused networks?",
		StopService:                 "Are you sure you want to stop this service's containers?",
		StopContainer:               "Are you sure you want to stop this container?",
		PressEnterToReturn:          "Press enter to return to dcli (this prompt can be disabled in your config by setting `gui.returnImmediately: true`)",
		DetachFromContainerShortCut: "By default, to detach from the container press ctrl-p then ctrl-q",

		No:  "no",
		Yes: "yes",

		LcNextScreenMode: "next screen mode (normal/half/fullscreen)",
		LcPrevScreenMode: "prev screen mode",
		FilterPrompt:     "filter",

		FocusProjects:   "focus projects panel",
		FocusServices:   "focus services panel",
		FocusContainers: "focus containers panel",
		FocusImages:     "focus images panel",
		FocusVolumes:    "focus volumes panel",
		FocusNetworks:   "focus networks panel",

		// CLI strings
		AppTitle:                   "Docker Cli Service Management",
		LoadedFiles:                "Loaded files:",
		DefaultComposeFile:         "Default (docker-compose.yml)",
		AllServicesList:            "All services list",
		AllServicesListAll:         "All services list (All)",
		NotRunningServicesList:     "Not running services list",
		RunningServicesList:        "Running services list",
		StatusNotRunning:           "not running",
		StatusRunning:              "running",
		StatusStopped:              "stopped",
		CommonTips:                 "Common tips: 1.Start, 2.Stop, 3.Restart, 4.Logs, 0.Exit",
		QuickCommands:              "Commands: [a]All, [r]Running, [s]Service search, [m]Menu search",
		Goodbye:                    "Goodbye!",
		SearchServiceTitle:         "Search all services (Supports index or name search, Esc to return)",
		SearchMenuTitle:            "Search function menu (Esc to return)",
		SelectActionForService:     "Select action for service [%s] (Esc to return)",
		ExecutingActionOnService:   "Executing action [%s] on service %s...",
		ExternalProjectStatusTip:   "This service belongs to an external project (%s), viewing runtime config via docker inspect...",
		ExternalProjectNoConfigTip: "This service belongs to an external project and container is not running, cannot get config.",
		ExternalProjectNoFixTip:    "This service belongs to an external project, cannot perform full fix here.",
		ExternalProjectNoBuildTip:  "This service belongs to an external project, cannot perform build operation.",
		RepairingService:           "Repairing service: %s...",
		DeletingServiceImage:       "Deleting image for service %s (ID: %s)...",
		ConfirmCleanService:        "Are you sure you want to clean service %s? (y/n): ",
		WarningStopAndRemove:       "Warning: This will stop and remove service: %s",
		ConfirmContinue:            "Are you sure you want to continue? (y/n): ",
		ServiceNotRunningNoLogs:    "Warning: Service %s is not running, there might be no real-time logs.",
		ViewingServiceLogs:         "--- Viewing service logs: %s (type 'exit' to return) ---",
		EnteringContainer:          "--- Entering container: %s (type 'exit' to exit) ---",
		ActionFailed:               "Failed: %v",
		ActionSuccess:              "Success",
		ActionCompleted:            "Action completed, press Enter to continue...",
		SelectServiceTo:            "Select service to %s (Enter or q/0 to return): ",
		PromptServiceIdxName:       "Enter index (e.g. 1) or name (e.g. mysql), multiple space-separated",
		PromptServiceAllIdxName:    "Enter 'all' for all, or index (e.g. 1) or name (e.g. mysql), multiple space-separated",
		InputServiceNameIdx:        "Service Name/Index: ",
		ErrorNoMatchingService:     "Error: No matching service found",
		ErrorNoStackServiceFound:   "Error: No services in stack found in current config",
		ExecutingAction:            "Executing %s: %s...",
		ConfirmOneKeyStartStack:    "Are you sure you want to start %s? (y/n, default n): ",
		StartingStack:              "Starting %s...",
		ContainsServices:           "Contains services: %s",
		CleaningDockerBuildCache:   "Cleaning Docker build cache...",
		CleaningDockerBuildHistory: "Cleaning Docker build history (including buildx)...",
		QuickSearchPrompt:          "Quick search (Built-in mode, enter keywords to filter):",
		MenuFunction:               "Function Menu:",
		WaitEnterToContinue:        "Press Enter to continue...",
		Build:                      "Build",
		Clean:                      "Clean",
		Fix:                        "Fix",
		ForceReconstruct:           "Force Reconstruct",
		FzfSelected:                "Selected",
		SuffixAllSpecified:         " (All/Specified)",
		SuffixSpecified:            " (Specified)",
		CommandExecutionError:      "Command execution error: %v",
		ExecutionFinishedExitTip:   "--- Execution finished, type 'exit' to return to main menu ---",
		PressExitToReturnTip:       "[Tip] Please type 'exit' and press Enter to return to main menu",
		ReturningToMainMenu:        "Returning to main menu...",
		MustTypeExitToQuit:         "[Tip] You must type 'exit' to quit the current screen",

		// Menu Items
		MenuStartService:      "Start Service (All/Specified)",
		MenuStopService:       "Stop Service (All/Specified)",
		MenuRestartService:    "Restart Service (All/Specified)",
		MenuViewLogs:          "View Logs (All/Specified)",
		MenuServiceStatus:     "View Service Status (Specified)",
		MenuServiceConfig:     "View Service Config (Specified)",
		MenuEnterContainer:    "Enter Container (Specified)",
		MenuBuildService:      "Build Service (All/Specified)",
		MenuForceReconstruct:  "Force Reconstruct (All/Specified) - No Cache",
		MenuCleanService:      "Clean Service (All/Specified)",
		MenuRemoveImage:       "Remove Image (All/Specified)",
		MenuLogStack:          "One-key start log monitoring stack (ELK/Graylog etc.)",
		MenuDBStack:           "One-key start database stack (MySQL/Redis/Clickhouse)",
		MenuCleanBuildCache:   "Clean Docker build cache",
		MenuCleanBuildxCache:  "Clean Docker buildx cache",
		MenuNetworkManagement: "Network Management",
		MenuVolumeManagement:  "Volume Management",
		MenuImageManagement:   "Image Management",
		MenuRepairService:     "Repair Service (All/Specified) - Rebuild Image",

		// Image Management
		SearchImageTitle:        "Search all images (Supports name search, Esc to return)",
		RunImage:                "Run Image",
		RunImageSuccess:         "Container %s is running",
		RunImageFailed:          "Run image failed: %v",
		SelectActionForImage:    "Select action for image [%s] (Esc to return)",
		ConfirmDeleteImage:      "Are you sure you want to delete image %s (ID: %s)? (y/n): ",
		DeletingImage:           "Deleting image...",
		InvalidIndex:            "Invalid index",
		DangerDeleteAllImages:   "DANGER: This will delete all unused images! Continue? (y/n): ",
		CleaningAllUnusedImages: "Cleaning all unused images...",
		PullImage:               "Pull Image",
		MenuRunImage:            "Run Image",
		MenuDeleteImage:         "Delete Image",
		DeleteAllImages:         "Prune All Unused Images",
		InputImageToRun:         "Enter the menu index: ",
		InputImageToDelete:      "Enter image index or name to DELETE: ",
		InputContainerName:      "Enter run name: ",
		DetectingShell:          "Detecting available shell...",
		PullingImage:            "Pulling image: %s...",
		PromptSearchKeyword:     "Please enter search keyword: ",
		SearchingRemoteImage:    "Searching Docker Hub...",
	}
}
