package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	// Global flags
	logFile          string
	detailsToStdout  bool
	verbose          bool
	quiet            bool
	waitForInit      bool
)

// Args struct to pass to Python - matches Python's args structure
type Args struct {
	Action          string `json:"action"`
	Subaction       string `json:"subaction,omitempty"`
	LogFile         string `json:"log,omitempty"`
	DetailsToStdout bool   `json:"details_to_stdout,omitempty"`
	Verbose         bool   `json:"verbose,omitempty"`
	Quiet           bool   `json:"quiet,omitempty"`
	WaitForInit     bool   `json:"wait_for_init,omitempty"`
	
	// Command-specific args
	ExtraArgs       map[string]interface{} `json:"extra_args,omitempty"`
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func callPython(args Args) error {
	// Try to find Python script in multiple locations
	var scriptPath string
	
	// First try current directory
	if _, err := os.Stat("waydroid_runner.py"); err == nil {
		scriptPath = "waydroid_runner.py"
	} else {
		// Try next to executable
		execPath, err := os.Executable()
		if err != nil {
			return fmt.Errorf("failed to get executable path: %v", err)
		}
		scriptPath = filepath.Join(filepath.Dir(execPath), "waydroid_runner.py")
		
		// Check if it exists
		if _, err := os.Stat(scriptPath); err != nil {
			return fmt.Errorf("cannot find waydroid_runner.py script: %v", err)
		}
	}
	
	// Convert args to JSON
	argsJSON, err := json.Marshal(args)
	if err != nil {
		return fmt.Errorf("failed to marshal args: %v", err)
	}
	
	// Call Python script with JSON args
	cmd := exec.Command("python3", scriptPath, string(argsJSON))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	
	return cmd.Run()
}

func createArgs(action, subaction string, extra map[string]interface{}) Args {
	return Args{
		Action:          action,
		Subaction:       subaction,
		LogFile:         logFile,
		DetailsToStdout: detailsToStdout,
		Verbose:         verbose,
		Quiet:           quiet,
		WaitForInit:     waitForInit,
		ExtraArgs:       extra,
	}
}

var rootCmd = &cobra.Command{
	Use:     "waydroid",
	Short:   "Waydroid is a container-based approach to boot a full Android system",
	Long:    "Waydroid is a container-based approach to boot a full Android system on a regular GNU/Linux system",
	Version: "1.4.2",
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVarP(&logFile, "log", "l", "", "path to log file")
	rootCmd.PersistentFlags().BoolVar(&detailsToStdout, "details-to-stdout", false, "print details to stdout instead of log")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "write even more to the logfiles")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "do not output any log messages")
	rootCmd.PersistentFlags().BoolVarP(&waitForInit, "wait", "w", false, "wait for init before running")
	
	// Add basic commands first
	setupBasicCommands()
	setupContainerCommands()
	setupAppCommands()
	setupOtherCommands()
}

func setupBasicCommands() {
	// Status command
	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Quick check for the waydroid",
		RunE: func(cmd *cobra.Command, args []string) error {
			return callPython(createArgs("status", "", nil))
		},
	}
	rootCmd.AddCommand(statusCmd)

	// Log command
	logCmd := &cobra.Command{
		Use:   "log",
		Short: "Follow the waydroid logfile",
		RunE: func(cmd *cobra.Command, args []string) error {
			lines, _ := cmd.Flags().GetString("lines")
			clearLog, _ := cmd.Flags().GetBool("clear")
			extra := map[string]interface{}{
				"lines":     lines,
				"clear_log": clearLog,
			}
			return callPython(createArgs("log", "", extra))
		},
	}
	logCmd.Flags().String("lines", "60", "number of lines to show")
	logCmd.Flags().Bool("clear", false, "clear log")
	rootCmd.AddCommand(logCmd)

	// Init command
	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Set up waydroid specific configs and install images",
		RunE: func(cmd *cobra.Command, args []string) error {
			imagesPath, _ := cmd.Flags().GetString("images-path")
			force, _ := cmd.Flags().GetBool("force")
			systemChannel, _ := cmd.Flags().GetString("system-channel")
			vendorChannel, _ := cmd.Flags().GetString("vendor-channel")
			romType, _ := cmd.Flags().GetString("rom-type")
			systemType, _ := cmd.Flags().GetString("system-type")
			
			extra := map[string]interface{}{
				"images_path":    imagesPath,
				"force":          force,
				"system_channel": systemChannel,
				"vendor_channel": vendorChannel,
				"rom_type":       romType,
				"system_type":    systemType,
			}
			return callPython(createArgs("init", "", extra))
		},
	}
	initCmd.Flags().StringP("images-path", "i", "", "custom path to waydroid images")
	initCmd.Flags().BoolP("force", "f", false, "re-initialize configs and images")
	initCmd.Flags().StringP("system-channel", "c", "", "custom system channel")
	initCmd.Flags().String("vendor-channel", "", "custom vendor channel")
	initCmd.Flags().StringP("rom-type", "r", "", "rom type (lineage, bliss, etc.)")
	initCmd.Flags().StringP("system-type", "s", "", "system type (VANILLA, FOSS, GAPPS)")
	rootCmd.AddCommand(initCmd)

	// Upgrade command
	upgradeCmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade images",
		RunE: func(cmd *cobra.Command, args []string) error {
			return callPython(createArgs("upgrade", "", nil))
		},
	}
	rootCmd.AddCommand(upgradeCmd)
}

func setupContainerCommands() {
	// Session command
	sessionCmd := &cobra.Command{
		Use:   "session",
		Short: "Session controller",
	}
	
	sessionStartCmd := &cobra.Command{
		Use:   "start",
		Short: "Start session",
		RunE: func(cmd *cobra.Command, args []string) error {
			return callPython(createArgs("session", "start", nil))
		},
	}
	
	sessionStopCmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop session",
		RunE: func(cmd *cobra.Command, args []string) error {
			return callPython(createArgs("session", "stop", nil))
		},
	}
	
	sessionCmd.AddCommand(sessionStartCmd, sessionStopCmd)
	rootCmd.AddCommand(sessionCmd)

	// Container command
	containerCmd := &cobra.Command{
		Use:   "container",
		Short: "Container controller",
	}
	
	for _, subcmd := range []string{"start", "stop", "restart", "freeze", "unfreeze"} {
		subCmd := &cobra.Command{
			Use:   subcmd,
			Short: fmt.Sprintf("%s container", subcmd),
			RunE: func(cmd *cobra.Command, args []string) error {
				return callPython(createArgs("container", cmd.Use, nil))
			},
		}
		containerCmd.AddCommand(subCmd)
	}
	
	rootCmd.AddCommand(containerCmd)
}

func setupAppCommands() {
	appCmd := &cobra.Command{
		Use:   "app",
		Short: "Applications controller",
	}
	
	// App install
	appInstallCmd := &cobra.Command{
		Use:   "install <package>",
		Short: "Install application package",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			extra := map[string]interface{}{"package": args[0]}
			return callPython(createArgs("app", "install", extra))
		},
	}
	
	// App remove
	appRemoveCmd := &cobra.Command{
		Use:   "remove <package>",
		Short: "Remove application package",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			extra := map[string]interface{}{"package": args[0]}
			return callPython(createArgs("app", "remove", extra))
		},
	}
	
	// App launch
	appLaunchCmd := &cobra.Command{
		Use:   "launch <package>",
		Short: "Launch application",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			extra := map[string]interface{}{"package": args[0]}
			return callPython(createArgs("app", "launch", extra))
		},
	}
	
	// App intent
	appIntentCmd := &cobra.Command{
		Use:   "intent <action> <uri>",
		Short: "Send intent to application",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			extra := map[string]interface{}{
				"action": args[0],
				"uri":    args[1],
			}
			return callPython(createArgs("app", "intent", extra))
		},
	}
	
	// App list
	appListCmd := &cobra.Command{
		Use:   "list",
		Short: "List installed applications",
		RunE: func(cmd *cobra.Command, args []string) error {
			return callPython(createArgs("app", "list", nil))
		},
	}
	
	appCmd.AddCommand(appInstallCmd, appRemoveCmd, appLaunchCmd, appIntentCmd, appListCmd)
	rootCmd.AddCommand(appCmd)
}

func setupOtherCommands() {
	// Prop command
	propCmd := &cobra.Command{
		Use:   "prop",
		Short: "Android properties controller",
	}
	
	propGetCmd := &cobra.Command{
		Use:   "get <key>",
		Short: "Get property value",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			extra := map[string]interface{}{"key": args[0]}
			return callPython(createArgs("prop", "get", extra))
		},
	}
	
	propSetCmd := &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set property value",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			extra := map[string]interface{}{
				"key":   args[0],
				"value": args[1],
			}
			return callPython(createArgs("prop", "set", extra))
		},
	}
	
	propCmd.AddCommand(propGetCmd, propSetCmd)
	rootCmd.AddCommand(propCmd)
	
	// Other simple commands
	simpleCommands := map[string]string{
		"show-full-ui": "Show full UI",
		"first-launch": "First launch setup",
		"shell":        "Launch shell in container",
		"logcat":       "Show Android logcat",
	}
	
	for cmdName, description := range simpleCommands {
		cmd := &cobra.Command{
			Use:   cmdName,
			Short: description,
			RunE: func(cmd *cobra.Command, args []string) error {
				return callPython(createArgs(cmd.Use, "", nil))
			},
		}
		rootCmd.AddCommand(cmd)
	}
	
	// ADB command
	adbCmd := &cobra.Command{
		Use:   "adb",
		Short: "ADB bridge controller",
	}
	
	for _, subcmd := range []string{"connect", "disconnect"} {
		subCmd := &cobra.Command{
			Use:   subcmd,
			Short: fmt.Sprintf("%s ADB", subcmd),
			RunE: func(cmd *cobra.Command, args []string) error {
				return callPython(createArgs("adb", cmd.Use, nil))
			},
		}
		adbCmd.AddCommand(subCmd)
	}
	
	rootCmd.AddCommand(adbCmd)
}